"use client"

import * as React from "react"
import { TagIcon, PlusIcon } from "lucide-react"
import { toast } from "sonner"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { leaveTypeApi } from "@/lib/api"

import { LeaveTypeForm } from "@/components/form/leave-type-form"
import type { LeaveType } from "@/components/data/leave-type-data"

const columns: ColumnDef<LeaveType>[] = [
  { id: "sl", header: "Id", cell: ({ row }) => row.index + 1 },
  { accessorKey: "name", header: "Leave Type" },
  { accessorKey: "total_days", header: "Balance" },
]

export default function LeaveTypePage() {
  const [data, setData] = React.useState<LeaveType[]>([])
  const [loading, setLoading] = React.useState(true)
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editingItem, setEditingItem] = React.useState<LeaveType | null>(null)

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const fetchData = React.useCallback(async () => {
    setLoading(true)
    try {
      const { data: res } = await leaveTypeApi.list(undefined, { page: String(page), limit: String(limit) })
      setData(res?.data ?? [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setData([])
    } finally {
      setLoading(false)
    }
  }, [page, limit])

  React.useEffect(() => { fetchData() }, [fetchData])

  const handleAdd = () => {
    setEditingItem(null)
    setDialogOpen(true)
  }

  const handleEdit = (item: LeaveType) => {
    setEditingItem(item)
    setDialogOpen(true)
  }

  const handleDelete = async (item: LeaveType) => {
    try {
      await leaveTypeApi.delete(item.id)
      setData((prev) => prev.filter((d) => d.id !== item.id))
      toast.success("Leave type deleted successfully")
    } catch {
      toast.error("Failed to delete leave type")
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <TagIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Leave Type</h1>
            <p className="text-muted-foreground mt-1">Manage leave types and policies</p>
          </div>
        </div>
        <Button onClick={handleAdd}>
          <PlusIcon className="mr-2 h-4 w-4" />
          Add Leave Type
        </Button>
      </div>

      <DataTable
        data={data}
        columns={columns}
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
            <DialogTitle>{editingItem ? "Edit Leave Type" : "Add New Leave Type"}</DialogTitle>
            <DialogDescription>
              {editingItem ? "Update the leave type details below." : "Fill in the leave type details below."}
            </DialogDescription>
          </DialogHeader>
          <LeaveTypeForm
            initialData={editingItem}
            onSuccess={() => {
              setDialogOpen(false)
              setEditingItem(null)
              fetchData()
            }}
            onCancel={() => { setDialogOpen(false); setEditingItem(null) }}
            isEditing={!!editingItem}
            leaveTypeId={editingItem?.id}
          />
        </DialogContent>
      </Dialog>
    </div>
  )
}
