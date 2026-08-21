"use client"

import * as React from "react"
import { CalendarDaysIcon, PlusIcon, Loader2 } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { dailyScheduleApi, companyApi } from "@/lib/api"

interface DailySchedule {
  id: string
  employee_id: string
  date: string
  schedule_type: string
  start_time: string
  end_time: string
  remarks: string
  status: string
  employee?: { name_en: string }
}

const scheduleTypeOptions = [
  { value: "Regular", label: "Regular" },
  { value: "OT", label: "OT" },
  { value: "Day Off", label: "Day Off" },
  { value: "Holiday", label: "Holiday" },
]

const statusBadge = (status: string) => {
  const variant = status === "active" ? "default" : status === "inactive" ? "secondary" : "outline"
  return <Badge variant={variant} className="capitalize">{status}</Badge>
}

const selectClass = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"

export default function DailySchedulePage() {
  const [data, setData] = React.useState<DailySchedule[]>([])
  const [loading, setLoading] = React.useState(true)
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editing, setEditing] = React.useState<DailySchedule | null>(null)
  const [submitting, setSubmitting] = React.useState(false)
  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const [form, setForm] = React.useState({ employee_id: "", date: "", schedule_type: "Regular", start_time: "", end_time: "" })

  const columns: ColumnDef<DailySchedule>[] = React.useMemo(() => [
    { accessorKey: "employee_id", header: "Employee ID" },
    {
      header: "Employee",
      accessorFn: (r) => r.employee?.name_en || "-",
    },
    { accessorKey: "date", header: "Date" },
    { accessorKey: "schedule_type", header: "Schedule Type" },
    { accessorKey: "start_time", header: "Start Time" },
    { accessorKey: "end_time", header: "End Time" },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => statusBadge(row.original.status),
    },
  ], [])

  const fetchData = React.useCallback(async (p?: number, l?: number) => {
    setLoading(true)
    try {
      const params: Record<string, string> = { page: String(p ?? page), limit: String(l ?? limit) }
      const { data: res } = await dailyScheduleApi.list(params)
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setData([])
      toast.error("Failed to load daily schedules")
    } finally {
      setLoading(false)
    }
  }, [page, limit])

  React.useEffect(() => { fetchData() }, [])
  React.useEffect(() => { fetchData() }, [page, limit])

  const resetForm = () => {
    setForm({ employee_id: "", date: "", schedule_type: "Regular", start_time: "", end_time: "" })
    setEditing(null)
  }

  const handleCreate = async () => {
    if (!form.employee_id || !form.date) { toast.error("Employee ID and date are required"); return }
    setSubmitting(true)
    try {
      await dailyScheduleApi.create({
        employee_id: form.employee_id,
        date: form.date,
        schedule_type: form.schedule_type,
        start_time: form.start_time,
        end_time: form.end_time,
      })
      toast.success("Daily schedule created")
      setDialogOpen(false)
      resetForm()
      fetchData(1)
    } catch { toast.error("Failed to create daily schedule") }
    finally { setSubmitting(false) }
  }

  const handleUpdate = async () => {
    if (!editing) return
    setSubmitting(true)
    try {
      await dailyScheduleApi.update(editing.id, {
        employee_id: form.employee_id,
        date: form.date,
        schedule_type: form.schedule_type,
        start_time: form.start_time,
        end_time: form.end_time,
      })
      toast.success("Daily schedule updated")
      setDialogOpen(false)
      resetForm()
      fetchData(page)
    } catch { toast.error("Failed to update daily schedule") }
    finally { setSubmitting(false) }
  }

  const handleDelete = async (row: DailySchedule) => {
    try {
      await dailyScheduleApi.delete(row.id)
      toast.success("Daily schedule deleted")
      fetchData(page)
    } catch { toast.error("Failed to delete daily schedule") }
  }

  const openEdit = (row: DailySchedule) => {
    setEditing(row)
    setForm({
      employee_id: row.employee_id,
      date: row.date,
      schedule_type: row.schedule_type,
      start_time: row.start_time,
      end_time: row.end_time,
    })
    setDialogOpen(true)
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <CalendarDaysIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Daily Schedule</h1>
            <p className="text-muted-foreground mt-1">Manage employee daily schedules</p>
          </div>
        </div>
        <Dialog open={dialogOpen} onOpenChange={(open) => { setDialogOpen(open); if (!open) resetForm() }}>
          <DialogTrigger asChild>
            <Button>
              <PlusIcon className="mr-2 h-4 w-4" />
              Add Schedule
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>{editing ? "Edit Schedule" : "Add Schedule"}</DialogTitle>
            </DialogHeader>
            <div className="space-y-3">
              <div className="flex flex-col gap-1.5">
                <Label>Employee ID *</Label>
                <Input value={form.employee_id} onChange={(e) => setForm((p) => ({ ...p, employee_id: e.target.value }))} placeholder="EMP-001" />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label>Date *</Label>
                <Input type="date" value={form.date} onChange={(e) => setForm((p) => ({ ...p, date: e.target.value }))} />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label>Schedule Type *</Label>
                <select value={form.schedule_type} onChange={(e) => setForm((p) => ({ ...p, schedule_type: e.target.value }))} className={selectClass}>
                  {scheduleTypeOptions.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <Label>Start Time</Label>
                <Input type="time" value={form.start_time} onChange={(e) => setForm((p) => ({ ...p, start_time: e.target.value }))} />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label>End Time</Label>
                <Input type="time" value={form.end_time} onChange={(e) => setForm((p) => ({ ...p, end_time: e.target.value }))} />
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <Button variant="outline" onClick={() => { setDialogOpen(false); resetForm() }}>Cancel</Button>
              <Button onClick={editing ? handleUpdate : handleCreate} disabled={submitting}>
                {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {editing ? "Update" : "Create"}
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      <DataTable
        data={data}
        columns={columns}
        onEdit={openEdit}
        onDelete={handleDelete}
        serverSide={true}
        page={page}
        pageSize={limit}
        pageCount={totalPages}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={(size) => { setLimit(size); setPage(1) }}
        loading={loading}
      />
    </div>
  )
}
