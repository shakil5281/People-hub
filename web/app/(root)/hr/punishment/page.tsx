"use client"

import * as React from "react"
import { SwordsIcon, PlusIcon, Loader2, SearchIcon, FilterIcon, XIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { DatePicker } from "@/components/ui/date-picker"
import { format } from "date-fns"
import { punishmentApi, employeeApi, companyApi } from "@/lib/api"

interface Punishment {
  id: string
  employee_id: string
  punishment_type: string
  reason: string
  amount: number
  overtime_less_hours: number | null
  overtime_rate: number | null
  absent_days: number | null
  per_day_rate: number | null
  date: string
  status: string
  employee?: {
    name_en: string
    department?: { name: string }
    designation_ref?: { name: string }
  }
}

interface EmployeeInfo {
  employee_id: string
  name_en: string
  name_bn: string
  department?: { id: string; name: string }
  designation_ref?: { id: string; name: string }
  company_id: string
  status: string
}

const typeOptions = [
  { value: "ot_less", label: "Overtime Less" },
  { value: "absent", label: "Absent" },
  { value: "fixed", label: "Fixed Amount" },
]

const typeLabels: Record<string, string> = {
  ot_less: "Overtime Less",
  absent: "Absent",
  fixed: "Fixed Amount",
}

const statusBadge = (status: string) => {
  const variant = status === "active" ? "default" : status === "inactive" ? "secondary" : "outline"
  return <Badge variant={variant} className="capitalize">{status}</Badge>
}

const selectClass = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"

export default function PunishmentPage() {
  const [data, setData] = React.useState<Punishment[]>([])
  const [loading, setLoading] = React.useState(true)
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editing, setEditing] = React.useState<Punishment | null>(null)
  const [submitting, setSubmitting] = React.useState(false)
  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)
  const [companies, setCompanies] = React.useState<{ id: string; company_name_en: string }[]>([])

  const [searchCode, setSearchCode] = React.useState("")
  const [searching, setSearching] = React.useState(false)
  const [foundEmployee, setFoundEmployee] = React.useState<EmployeeInfo | null>(null)
  const [employeeError, setEmployeeError] = React.useState("")
  const [calculatedAmount, setCalculatedAmount] = React.useState<number | null>(null)
  const [dateObj, setDateObj] = React.useState<Date | undefined>(undefined)

  const [form, setForm] = React.useState({
    employee_id: "",
    company_id: "",
    punishment_type: "ot_less",
    reason: "",
    amount: "",
    overtime_less_hours: "",
    overtime_rate: "",
    absent_days: "",
    per_day_rate: "",
  })

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then(({ data }) => {
      const list = Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : []
      setCompanies(list)
      if (list.length > 0 && !form.company_id) {
        setForm((p) => ({ ...p, company_id: list[0].id }))
      }
    })
  }, [])

  const columns: ColumnDef<Punishment>[] = React.useMemo(() => [
    { accessorKey: "employee_id", header: "Employee ID" },
    {
      header: "Employee",
      accessorFn: (r) => r.employee?.name_en || "-",
    },
    {
      header: "Department",
      accessorFn: (r) => r.employee?.department?.name || "-",
    },
    {
      header: "Type",
      accessorFn: (r) => typeLabels[r.punishment_type] || r.punishment_type,
    },
    {
      accessorKey: "amount",
      header: "Amount",
      cell: ({ row }) => row.original.amount ? `৳${row.original.amount.toLocaleString()}` : "-",
    },
    { accessorKey: "date", header: "Date" },
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
      const { data: res } = await punishmentApi.list(params)
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setData([])
      toast.error("Failed to load punishments")
    } finally {
      setLoading(false)
    }
  }, [page, limit])

  React.useEffect(() => { fetchData() }, [])
  React.useEffect(() => { fetchData() }, [page, limit])

  const resetForm = () => {
    setForm({
      employee_id: "",
      company_id: companies.length > 0 ? companies[0].id : "",
      punishment_type: "ot_less",
      reason: "",
      amount: "",
      overtime_less_hours: "",
      overtime_rate: "",
      absent_days: "",
      per_day_rate: "",
    })
    setDateObj(undefined)
    setFoundEmployee(null)
    setEmployeeError("")
    setCalculatedAmount(null)
    setEditing(null)
  }

  const handleSearchEmployee = async () => {
    if (!searchCode.trim()) { toast.error("Enter an employee ID"); return }
    setSearching(true)
    setEmployeeError("")
    setFoundEmployee(null)
    try {
      const { data: emp } = await employeeApi.getByCode(searchCode.trim())
      setFoundEmployee(emp)
      setForm((p) => ({
        ...p,
        employee_id: emp.employee_id,
        company_id: emp.company_id,
      }))
    } catch {
      setEmployeeError("Employee not found")
      toast.error("Employee not found")
    } finally {
      setSearching(false)
    }
  }

  const calculateAmount = React.useCallback(async (formData: typeof form) => {
    if (!formData.employee_id) return
    try {
      const payload: Record<string, unknown> = {
        company_id: formData.company_id,
        employee_id: formData.employee_id,
        punishment_type: formData.punishment_type,
        date: dateObj ? format(dateObj, "yyyy-MM-dd") : new Date().toISOString().split("T")[0],
      }
      if (formData.punishment_type === "ot_less") {
        payload.overtime_less_hours = formData.overtime_less_hours ? Number(formData.overtime_less_hours) : 0
        payload.overtime_rate = formData.overtime_rate ? Number(formData.overtime_rate) : 0
      } else if (formData.punishment_type === "absent") {
        payload.absent_days = formData.absent_days ? Number(formData.absent_days) : 0
        payload.per_day_rate = formData.per_day_rate ? Number(formData.per_day_rate) : 0
      } else {
        payload.amount = formData.amount ? Number(formData.amount) : 0
      }
      const { data: res } = await punishmentApi.calculate(payload)
      setCalculatedAmount(res.amount)
    } catch {
      setCalculatedAmount(null)
    }
  }, [])

  React.useEffect(() => {
    const timer = setTimeout(() => {
      if (form.employee_id) calculateAmount(form)
    }, 300)
    return () => clearTimeout(timer)
  }, [form.punishment_type, form.overtime_less_hours, form.overtime_rate, form.absent_days, form.per_day_rate, form.amount, form.employee_id, dateObj, calculateAmount])

  const handleCreate = async () => {
    if (!form.employee_id || !dateObj) { toast.error("Employee ID and date are required"); return }
    setSubmitting(true)
    try {
      const payload: Record<string, unknown> = {
        company_id: form.company_id,
        employee_id: form.employee_id,
        punishment_type: form.punishment_type,
        reason: form.reason,
        date: format(dateObj, "yyyy-MM-dd"),
      }
      if (form.punishment_type === "ot_less") {
        payload.overtime_less_hours = form.overtime_less_hours ? Number(form.overtime_less_hours) : 0
        payload.overtime_rate = form.overtime_rate ? Number(form.overtime_rate) : 0
      } else if (form.punishment_type === "absent") {
        payload.absent_days = form.absent_days ? Number(form.absent_days) : 0
        payload.per_day_rate = form.per_day_rate ? Number(form.per_day_rate) : 0
      } else {
        payload.amount = form.amount ? Number(form.amount) : 0
      }
      await punishmentApi.create(payload)
      toast.success("Punishment created")
      setDialogOpen(false)
      resetForm()
      fetchData(1)
    } catch { toast.error("Failed to create punishment") }
    finally { setSubmitting(false) }
  }

  const handleUpdate = async () => {
    if (!editing) return
    setSubmitting(true)
    try {
      const payload: Record<string, unknown> = {
        company_id: form.company_id,
        employee_id: form.employee_id,
        punishment_type: form.punishment_type,
        reason: form.reason,
        date: dateObj ? format(dateObj, "yyyy-MM-dd") : new Date().toISOString().split("T")[0],
      }
      if (form.punishment_type === "ot_less") {
        payload.overtime_less_hours = form.overtime_less_hours ? Number(form.overtime_less_hours) : 0
        payload.overtime_rate = form.overtime_rate ? Number(form.overtime_rate) : 0
      } else if (form.punishment_type === "absent") {
        payload.absent_days = form.absent_days ? Number(form.absent_days) : 0
        payload.per_day_rate = form.per_day_rate ? Number(form.per_day_rate) : 0
      } else {
        payload.amount = form.amount ? Number(form.amount) : 0
      }
      await punishmentApi.update(editing.id, payload)
      toast.success("Punishment updated")
      setDialogOpen(false)
      resetForm()
      fetchData(page)
    } catch { toast.error("Failed to update punishment") }
    finally { setSubmitting(false) }
  }

  const handleDelete = async (row: Punishment) => {
    try {
      await punishmentApi.delete(row.id)
      toast.success("Punishment deleted")
      fetchData(page)
    } catch { toast.error("Failed to delete punishment") }
  }

  const openEdit = (row: Punishment) => {
    setEditing(row)
    setFoundEmployee({
      employee_id: row.employee_id,
      name_en: row.employee?.name_en || "",
      name_bn: "",
      department: row.employee?.department ? { id: "", name: row.employee.department.name } : undefined,
      designation_ref: row.employee?.designation_ref ? { id: "", name: row.employee.designation_ref.name } : undefined,
      company_id: "",
      status: "active",
    })
    setSearchCode(row.employee_id)
    setForm({
      employee_id: row.employee_id,
      company_id: "",
      punishment_type: row.punishment_type,
      reason: row.reason || "",
      amount: row.punishment_type === "fixed" ? String(row.amount ?? "") : "",
      overtime_less_hours: row.overtime_less_hours ? String(row.overtime_less_hours) : "",
      overtime_rate: row.overtime_rate ? String(row.overtime_rate) : "",
      absent_days: row.absent_days ? String(row.absent_days) : "",
      per_day_rate: row.per_day_rate ? String(row.per_day_rate) : "",
    })
    setDateObj(row.date ? new Date(row.date) : undefined)
    setCalculatedAmount(row.amount)
    setDialogOpen(true)
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <SwordsIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Punishment</h1>
            <p className="text-muted-foreground mt-1">Manage employee punishments by type</p>
          </div>
        </div>
        <Dialog open={dialogOpen} onOpenChange={(open) => { setDialogOpen(open); if (!open) resetForm() }}>
          <DialogTrigger asChild>
            <Button>
              <PlusIcon className="mr-2 h-4 w-4" />
              Add Punishment
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-lg max-h-[85vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>{editing ? "Edit Punishment" : "Add Punishment"}</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div className="flex flex-col gap-1.5">
                <Label>Employee ID *</Label>
                <div className="flex gap-2">
                  <Input
                    value={searchCode}
                    onChange={(e) => { setSearchCode(e.target.value); setFoundEmployee(null); setEmployeeError("") }}
                    placeholder="EMP-001"
                    onKeyDown={(e) => { if (e.key === "Enter") handleSearchEmployee() }}
                    className="flex-1"
                  />
                  <Button variant="secondary" size="sm" onClick={handleSearchEmployee} disabled={searching}>
                    {searching ? <Loader2 className="h-4 w-4 animate-spin" /> : <SearchIcon className="h-4 w-4" />}
                  </Button>
                </div>
                {employeeError && <p className="text-xs text-destructive">{employeeError}</p>}
                {foundEmployee && (
                  <div className="rounded-md border bg-muted/30 p-3 mt-1 space-y-1">
                    <p className="text-sm font-medium">{foundEmployee.name_en}</p>
                    <div className="flex flex-wrap gap-x-4 gap-y-0.5 text-xs text-muted-foreground">
                      {foundEmployee.department?.name && <span>Dept: {foundEmployee.department.name}</span>}
                      {foundEmployee.designation_ref?.name && <span>Desig: {foundEmployee.designation_ref.name}</span>}
                    </div>
                  </div>
                )}
              </div>

              <div className="flex flex-col gap-1.5">
                <Label>Punishment Type *</Label>
                <select
                  value={form.punishment_type}
                  onChange={(e) => setForm((p) => ({ ...p, punishment_type: e.target.value }))}
                  className={selectClass}
                >
                  {typeOptions.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
                </select>
              </div>

              {form.punishment_type === "ot_less" && (
                <div className="grid grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <Label>OT Hours Less</Label>
                    <Input
                      type="number" step="0.5"
                      value={form.overtime_less_hours}
                      onChange={(e) => setForm((p) => ({ ...p, overtime_less_hours: e.target.value }))}
                      placeholder="0"
                    />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label>OT Rate (৳/hr)</Label>
                    <Input
                      type="number" step="0.01"
                      value={form.overtime_rate}
                      onChange={(e) => setForm((p) => ({ ...p, overtime_rate: e.target.value }))}
                      placeholder="0"
                    />
                  </div>
                </div>
              )}

              {form.punishment_type === "absent" && (
                <div className="grid grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <Label>Absent Days</Label>
                    <Input
                      type="number"
                      value={form.absent_days}
                      onChange={(e) => setForm((p) => ({ ...p, absent_days: e.target.value }))}
                      placeholder="0"
                    />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label>Per Day Rate (৳)</Label>
                    <Input
                      type="number" step="0.01"
                      value={form.per_day_rate}
                      onChange={(e) => setForm((p) => ({ ...p, per_day_rate: e.target.value }))}
                      placeholder="0"
                    />
                  </div>
                </div>
              )}

              {form.punishment_type === "fixed" && (
                <div className="flex flex-col gap-1.5">
                  <Label>Amount (৳)</Label>
                  <Input
                    type="number" step="0.01"
                    value={form.amount}
                    onChange={(e) => setForm((p) => ({ ...p, amount: e.target.value }))}
                    placeholder="0"
                  />
                </div>
              )}

              {calculatedAmount !== null && (
                <div className="rounded-md border bg-primary/5 p-3 text-center">
                  <span className="text-sm text-muted-foreground">Calculated Amount: </span>
                  <span className="text-lg font-bold">৳{calculatedAmount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
                </div>
              )}

              <div className="flex flex-col gap-1.5">
                <Label>Reason</Label>
                <Input value={form.reason} onChange={(e) => setForm((p) => ({ ...p, reason: e.target.value }))} placeholder="Reason..." />
              </div>

              <div className="flex flex-col gap-1.5">
                <Label>Date *</Label>
                <DatePicker value={dateObj} onChange={setDateObj} placeholder="dd/mm/yyyy" />
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
