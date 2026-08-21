"use client"

import * as React from "react"
import { LayersIcon, PlusIcon } from "lucide-react"
import { toast } from "sonner"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Floor } from "@/components/data/floor-data"
import { FloorForm } from "@/components/form/floor-form"
import { floorApi } from "@/lib/api"

const columns: ColumnDef<Floor>[] = [
  {
    id: "sl",
    header: "SL",
    cell: ({ row }) => row.index + 1,
  },
  { accessorKey: "name", header: "Floor Name" },
]

export default function FloorPage() {
  const [floors, setFloors] = React.useState<Floor[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editingFloor, setEditingFloor] = React.useState<Floor | null>(null)

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const fetchFloors = async () => {
    setLoading(true)
    setError("")
    try {
      const { data: res } = await floorApi.list({ page: String(page), limit: String(limit) })
      setFloors(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setError("Failed to load floors")
    } finally {
      setLoading(false)
    }
  }

  React.useEffect(() => {
    fetchFloors()
  }, [page, limit])

  const handleAdd = () => {
    setEditingFloor(null)
    setDialogOpen(true)
  }

  const handleEdit = (floor: Floor) => {
    setEditingFloor(floor)
    setDialogOpen(true)
  }

  const handleDelete = async (floor: Floor) => {
    try {
      await floorApi.delete(floor.id)
      setFloors((prev) => prev.filter((f) => f.id !== floor.id))
      toast.success("Floor deleted successfully")
    } catch {
      toast.error("Failed to delete floor")
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <LayersIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Floor</h1>
            <p className="text-muted-foreground mt-1">Manage building floors</p>
          </div>
        </div>
        <Button onClick={handleAdd}>
          <PlusIcon className="mr-2 h-4 w-4" />
          Add Floor
        </Button>
      </div>

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      <DataTable
        data={floors}
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
            <DialogTitle>{editingFloor ? "Edit Floor" : "Add New Floor"}</DialogTitle>
            <DialogDescription>
              {editingFloor ? "Update the floor name below." : "Fill in the floor details below."}
            </DialogDescription>
          </DialogHeader>
          <FloorForm
            initialData={editingFloor ? { name: editingFloor.name } : undefined}
            onSuccess={() => {
              setDialogOpen(false)
              setEditingFloor(null)
              fetchFloors()
            }}
            onCancel={() => { setDialogOpen(false); setEditingFloor(null) }}
            isEditing={!!editingFloor}
            floorId={editingFloor?.id}
          />
        </DialogContent>
      </Dialog>
    </div>
  )
}
