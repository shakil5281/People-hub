"use client"

import * as React from "react"
import { UsersIcon, PlusIcon } from "lucide-react"
import { toast } from "sonner"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Group } from "@/components/data/group-data"
import { GroupForm } from "@/components/form/group-form"
import { groupApi } from "@/lib/api"

const columns: ColumnDef<Group>[] = [
  {
    id: "sl",
    header: "SL",
    cell: ({ row }) => row.index + 1,
  },
  { accessorKey: "name", header: "Group Name" },
]

export default function GroupPage() {
  const [groups, setGroups] = React.useState<Group[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editingGroup, setEditingGroup] = React.useState<Group | null>(null)

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const fetchGroups = async () => {
    setLoading(true)
    setError("")
    try {
      const { data: res } = await groupApi.list({ page: String(page), limit: String(limit) })
      setGroups(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setError("Failed to load groups")
    } finally {
      setLoading(false)
    }
  }

  React.useEffect(() => {
    fetchGroups()
  }, [page, limit])

  const handleAdd = () => {
    setEditingGroup(null)
    setDialogOpen(true)
  }

  const handleEdit = (group: Group) => {
    setEditingGroup(group)
    setDialogOpen(true)
  }

  const handleDelete = async (group: Group) => {
    try {
      await groupApi.delete(group.id)
      setGroups((prev) => prev.filter((g) => g.id !== group.id))
      toast.success("Group deleted successfully")
    } catch {
      toast.error("Failed to delete group")
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <UsersIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Group</h1>
            <p className="text-muted-foreground mt-1">Manage employee groups</p>
          </div>
        </div>
        <Button onClick={handleAdd}>
          <PlusIcon className="mr-2 h-4 w-4" />
          Add Group
        </Button>
      </div>

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      <DataTable
        data={groups}
        columns={columns}
        enableDnd
        onEdit={handleEdit}
        onDelete={handleDelete}
        serverSide={true}
        page={page}
        pageSize={limit}
        pageCount={totalPages}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={(size) => { setLimit(size); setPage(1); }}
        loading={loading}
      />

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingGroup ? "Edit Group" : "Add New Group"}</DialogTitle>
            <DialogDescription>
              {editingGroup ? "Update the group name below." : "Fill in the group details below."}
            </DialogDescription>
          </DialogHeader>
          <GroupForm
            initialData={editingGroup ? { name: editingGroup.name } : undefined}
            onSuccess={() => {
              setDialogOpen(false)
              setEditingGroup(null)
              fetchGroups()
            }}
            onCancel={() => { setDialogOpen(false); setEditingGroup(null) }}
            isEditing={!!editingGroup}
            groupId={editingGroup?.id}
          />
        </DialogContent>
      </Dialog>
    </div>
  )
}
