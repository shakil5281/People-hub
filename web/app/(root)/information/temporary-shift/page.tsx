"use client"

import * as React from "react"
import { TimerIcon, PlusIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { TempShift, getTempShifts, createTempShift, updateTempShift, deleteTempShift, TempShiftFormData } from "@/components/data/temporary-shift-data"
import { TempShiftForm } from "@/components/form/temporary-shift-form"
import { companyApi } from "@/lib/api"

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  active: "default",
  inactive: "secondary",
}

const columns: ColumnDef<TempShift>[] = [
  {
    accessorKey: "employee",
    header: "Employee",
    cell: ({ row }) => row.original.employee?.name_en || "-",
  },
  {
    accessorKey: "employee_id",
    header: "Code",
    cell: ({ row }) => row.original.employee?.employee_id || "-",
  },
  {
    accessorKey: "shift",
    header: "Assigned Shift",
    cell: ({ row }) => row.original.shift?.name || "-",
  },
  { accessorKey: "date", header: "Date" },
  { accessorKey: "reason", header: "Reason" },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const s = row.original.status
      return <Badge variant={statusVariant[s] || "secondary"}>{s}</Badge>
    },
  },
]

export default function TemporaryShiftPage() {
  const [data, setData] = React.useState<TempShift[]>([])
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editing, setEditing] = React.useState<TempShift | null>(null)
  const [companyId, setCompanyId] = React.useState("")

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const refreshData = async (p?: number, l?: number) => {
    const result = await getTempShifts(companyId, { page: String(p ?? page), limit: String(l ?? limit) })
    setData(result.data)
    setTotal(result.total)
    setTotalPages(result.total_pages)
  }

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((res) => {
      const list = Array.isArray(res.data?.data) ? res.data.data : []
      const cid = list[0]?.id || ""
      setCompanyId(cid)
      return getTempShifts(cid, { page: "1", limit: "20" })
    }).then((result) => {
      setData(result.data)
      setTotal(result.total)
      setTotalPages(result.total_pages)
    })
  }, [])

  React.useEffect(() => {
    if (companyId) {
      refreshData()
    }
  }, [page, limit])

  const handleAdd = () => { setEditing(null); setDialogOpen(true) }
  const handleEdit = (item: TempShift) => { setEditing(item); setDialogOpen(true) }
  const handleDelete = async (item: TempShift) => {
    await deleteTempShift(item.id)
    refreshData()
  }
  const handleFormSuccess = async (formData: TempShiftFormData) => {
    if (editing) {
      await updateTempShift(editing.id, formData)
    } else {
      await createTempShift({ ...formData, company_id: companyId })
    }
    refreshData()
    setDialogOpen(false)
    setEditing(null)
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <TimerIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Temporary Shift</h1>
            <p className="text-muted-foreground mt-1">Manage temporary shift assignments</p>
          </div>
        </div>
        <Button onClick={handleAdd}>
          <PlusIcon className="mr-2 h-4 w-4" />
          Add
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
      />
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editing ? "Edit Temporary Shift" : "Add Temporary Shift"}</DialogTitle>
            <DialogDescription>
              {editing ? "Update the temporary shift details." : "Fill in the temporary shift details."}
            </DialogDescription>
          </DialogHeader>
          <TempShiftForm
            initialData={editing || undefined}
            onSuccess={handleFormSuccess}
            onCancel={() => { setDialogOpen(false); setEditing(null) }}
            isEditing={!!editing}
            tempShiftId={editing?.id}
          />
        </DialogContent>
      </Dialog>
    </div>
  )
}
