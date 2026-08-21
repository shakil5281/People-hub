"use client"

import * as React from "react"
import { CalendarClockIcon, PlusIcon } from "lucide-react"
import { toast } from "sonner"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Shift, statusOptions } from "@/components/data/shift-data"
import { ShiftForm } from "@/components/form/shift-form"
import { shiftApi } from "@/lib/api"

const columns: ColumnDef<Shift>[] = [
  { accessorKey: "name", header: "Shift Name" },
  {
    accessorKey: "shift_type",
    header: "Type",
    cell: ({ row }) => (
      <Badge variant="outline" className="capitalize">{row.original.shift_type || "day"}</Badge>
    ),
  },
  { accessorKey: "start_time", header: "Start Time" },
  { accessorKey: "end_time", header: "End Time" },
  { accessorKey: "late_grace_minutes", header: "Grace (min)" },
  {
    accessorKey: "weekend_days",
    header: "Weekend Days",
    cell: ({ row }) => row.original.weekend_days || "-",
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => (
      <Badge variant={row.original.status === "active" ? "default" : "secondary"} className="capitalize">
        {statusOptions.find((s) => s.value === row.original.status)?.label}
      </Badge>
    ),
  },
]

export default function ShiftPage() {
  const [shifts, setShifts] = React.useState<Shift[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editingShift, setEditingShift] = React.useState<Shift | null>(null)

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const fetchShifts = async () => {
    setLoading(true)
    setError("")
    try {
      const { data: res } = await shiftApi.list({ page: String(page), limit: String(limit) })
      setShifts(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setError("Failed to load shifts")
    } finally {
      setLoading(false)
    }
  }

  React.useEffect(() => {
    fetchShifts()
  }, [page, limit])

  const handleAdd = () => {
    setEditingShift(null)
    setDialogOpen(true)
  }

  const handleEdit = (shift: Shift) => {
    setEditingShift(shift)
    setDialogOpen(true)
  }

  const handleDelete = async (shift: Shift) => {
    try {
      await shiftApi.delete(shift.id)
      setShifts((prev) => prev.filter((s) => s.id !== shift.id))
      toast.success("Shift deleted successfully")
    } catch {
      toast.error("Failed to delete shift")
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <CalendarClockIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Shift</h1>
            <p className="text-muted-foreground mt-1">Manage work shifts and schedules</p>
          </div>
        </div>
        <Button onClick={handleAdd}>
          <PlusIcon className="mr-2 h-4 w-4" />
          Add Shift
        </Button>
      </div>

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      <DataTable
        data={shifts}
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
            <DialogTitle>{editingShift ? "Edit Shift" : "Add New Shift"}</DialogTitle>
            <DialogDescription>
              {editingShift ? "Update the shift details below." : "Fill in the shift details below."}
            </DialogDescription>
          </DialogHeader>
          <ShiftForm
            initialData={editingShift ? {
              company_id: editingShift.company_id,
              name: editingShift.name,
              shift_type: (editingShift.shift_type || "day") as "day" | "night" | "general",
              start_time: editingShift.start_time,
              end_time: editingShift.end_time,
              late_grace_minutes: editingShift.late_grace_minutes,
              weekend_days: editingShift.weekend_days || "",
              status: editingShift.status,
            } : undefined}
            onSuccess={() => {
              setDialogOpen(false)
              setEditingShift(null)
              fetchShifts()
            }}
            onCancel={() => { setDialogOpen(false); setEditingShift(null) }}
            isEditing={!!editingShift}
            shiftId={editingShift?.id}
          />
        </DialogContent>
      </Dialog>
    </div>
  )
}
