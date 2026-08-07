"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { MoonIcon, PlusIcon, ZapIcon, FileSpreadsheetIcon, FileTextIcon, Loader2, FilterIcon, XIcon, RefreshCwIcon, UsersIcon, CoinsIcon, ClockIcon, SunsetIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { DatePicker } from "@/components/ui/date-picker"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { format } from "date-fns"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { nightBillApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, type NightBillRecord } from "@/lib/api"
import { cn, formatCheck } from "@/lib/utils"

interface DropdownItem {
  id: string
  name?: string
  company_name_en?: string
}

const today = new Date().toISOString().split("T")[0]

const billTypeBadge = (type: string) => {
  const cfg: Record<string, string> = { fixed: "default", hourly: "secondary", manual: "outline" }
  return <Badge variant={cfg[type] as "default" | "secondary" | "outline"} className="capitalize">{type}</Badge>
}

export default function NightBillPage() {
  const router = useRouter()
  const [data, setData] = React.useState<NightBillRecord[]>([])
  const [selectedRows, setSelectedRows] = React.useState<NightBillRecord[]>([])
  const [loading, setLoading] = React.useState(true)
  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const [companies, setCompanies] = React.useState<DropdownItem[]>([])
  const [departments, setDepartments] = React.useState<DropdownItem[]>([])
  const [sections, setSections] = React.useState<DropdownItem[]>([])
  const [designations, setDesignations] = React.useState<DropdownItem[]>([])
  const [lines, setLines] = React.useState<DropdownItem[]>([])
  const [groups, setGroups] = React.useState<DropdownItem[]>([])

  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  // Process dialog
  const [processOpen, setProcessOpen] = React.useState(false)
  const [processing, setProcessing] = React.useState(false)
  const [processCompanyId, setProcessCompanyId] = React.useState("")
  const [processStartDate, setProcessStartDate] = React.useState(today)
  const [processEndDate, setProcessEndDate] = React.useState(today)

  // Add manual dialog
  const [addOpen, setAddOpen] = React.useState(false)
  const [adding, setAdding] = React.useState(false)
  const [addCompanyId, setAddCompanyId] = React.useState("")
  const [addEmployeeId, setAddEmployeeId] = React.useState("")
  const [addDate, setAddDate] = React.useState(today)
  const [addAmount, setAddAmount] = React.useState("")
  const [addRemarks, setAddRemarks] = React.useState("Manual Night Bill")

  // Edit dialog
  const [editRecord, setEditRecord] = React.useState<NightBillRecord | null>(null)
  const [editOpen, setEditOpen] = React.useState(false)
  const [updating, setUpdating] = React.useState(false)
  const [editAmount, setEditAmount] = React.useState("")
  const [editRemarks, setEditRemarks] = React.useState("")

  // Delete
  const [deleting, setDeleting] = React.useState(false)
  const [confirmDeleteOpen, setConfirmDeleteOpen] = React.useState(false)
  const [deleteTarget, setDeleteTarget] = React.useState<NightBillRecord | null>(null)

  const [filters, setFilters] = React.useState<Record<string, string>>({
    start_date: today,
    end_date: today,
  })

  const buildParams = (f: Record<string, string>, p: number, l: number) => {
    const params: Record<string, string> = {
      start_date: f.start_date || today,
      end_date: f.end_date || today,
      page: String(p),
      limit: String(l),
    }
    ;["company_id", "department_id", "section_id", "designation_id", "line_id", "group_id", "bill_type"].forEach((k) => { if (f[k]) params[k] = f[k] })
    return params
  }

  const fetchData = React.useCallback(async (f?: Record<string, string>, p?: number, l?: number) => {
    setLoading(true)
    try {
      const { data: res } = await nightBillApi.list(buildParams(f ?? filters, p ?? page, l ?? limit))
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setData([])
      toast.error("Failed to load night bills")
    } finally {
      setLoading(false)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, limit, filters])

  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "200" }),
      departmentApi.list({ limit: "200" }),
      sectionApi.list(undefined, { limit: "200" }),
      designationApi.list(undefined, { limit: "200" }),
      lineApi.list(undefined, { limit: "200" }),
      groupApi.list({ limit: "200" }),
    ]).then(([c, d, s, de, l, g]) => {
      if (Array.isArray(c.data?.data)) setCompanies(c.data.data)
      if (Array.isArray(d.data?.data)) setDepartments(d.data.data)
      if (Array.isArray(s.data?.data)) setSections(s.data.data)
      if (Array.isArray(de.data?.data)) setDesignations(de.data.data)
      if (Array.isArray(l.data?.data)) setLines(l.data.data)
      if (Array.isArray(g.data?.data)) setGroups(g.data.data)
    }).catch(() => {})
    fetchData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  React.useEffect(() => { fetchData() }, [page, limit]) // eslint-disable-line react-hooks/exhaustive-deps

  const stats = React.useMemo(() => {
    const totalAmt = data.reduce((s, r) => s + (r.amount || 0), 0)
    return {
      totalAmt,
      fixedCount: data.filter((r) => r.bill_type === "fixed").length,
      hourlyCount: data.filter((r) => r.bill_type === "hourly").length,
      manualCount: data.filter((r) => r.bill_type === "manual").length,
    }
  }, [data])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date_range", label: "Date Range", type: "daterange-split", dateRangeKeys: { start: "start_date", end: "end_date" } },
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en || c.id })) },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name || d.id })) },
    { key: "section_id", label: "Section", type: "select", options: sections.map((d) => ({ value: d.id, label: d.name || d.id })) },
    { key: "designation_id", label: "Designation", type: "select", options: designations.map((d) => ({ value: d.id, label: d.name || d.id })) },
    { key: "line_id", label: "Line", type: "select", options: lines.map((d) => ({ value: d.id, label: d.name || d.id })) },
    { key: "group_id", label: "Group", type: "select", options: groups.map((d) => ({ value: d.id, label: d.name || d.id })) },
    { key: "bill_type", label: "Bill Type", type: "select", options: [{ value: "fixed", label: "Fixed" }, { value: "hourly", label: "Hourly" }, { value: "manual", label: "Manual" }] },
  ], [companies, departments, sections, designations, lines, groups])
  const handleChange = (key: string, value: string) => setFilters((prev) => ({ ...prev, [key]: value }))
  const handleApply = () => { setPage(1); fetchData(filters, 1) }
  const handleReset = () => {
    const def = { start_date: today, end_date: today }
    setFilters(def); setPage(1); fetchData(def, 1)
  }

  const handleProcess = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!processCompanyId) { toast.error("Company is required"); return }
    setProcessing(true)
    try {
      const res = await nightBillApi.processConfig({
        companyId: processCompanyId,
        fromDate: processStartDate,
        toDate: processEndDate,
      })
      toast.success(`Processed ${res.data.processed} — Generated ${res.data.generated}, Skipped ${res.data.skipped}, Duplicates ${res.data.duplicates}`)
      setProcessOpen(false)
      fetchData()
    } catch {
      toast.error("Failed to process night bills")
    } finally {
      setProcessing(false)
    }
  }

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!addEmployeeId || !addCompanyId || !addDate) { toast.error("Company, Employee and Date are required"); return }
    setAdding(true)
    try {
      await nightBillApi.create({
        company_id: addCompanyId,
        employee_id: addEmployeeId,
        attendance_date: addDate,
        bill_type: "manual",
        amount: parseFloat(addAmount) || 0,
        remarks: addRemarks,
      })
      toast.success("Night bill added")
      setAddOpen(false)
      fetchData()
    } catch {
      toast.error("Failed to add night bill")
    } finally {
      setAdding(false)
    }
  }

  const openEdit = (rec: NightBillRecord) => {
    setEditRecord(rec)
    setEditAmount(String(rec.amount || 0))
    setEditRemarks(rec.remarks || "")
    setEditOpen(true)
  }

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editRecord) return
    setUpdating(true)
    try {
      await nightBillApi.update(editRecord.id, { amount: parseFloat(editAmount) || 0, remarks: editRemarks })
      toast.success("Night bill updated")
      setEditOpen(false)
      fetchData()
    } catch {
      toast.error("Failed to update")
    } finally {
      setUpdating(false)
    }
  }

  const handleDelete = async () => {
    setDeleting(true)
    try {
      const ids = deleteTarget ? [deleteTarget.id] : selectedRows.map((r) => r.id)
      if (deleteTarget) await nightBillApi.delete(deleteTarget.id)
      else await nightBillApi.deleteBulk(ids)
      toast.success(`Deleted ${ids.length} record(s)`)
      setSelectedRows([])
      fetchData()
    } catch {
      toast.error("Delete failed")
    } finally {
      setDeleting(false)
      setConfirmDeleteOpen(false)
      setDeleteTarget(null)
    }
  }

  const doExport = async (type: "excel" | "pdf") => {
    try {
      const fn = type === "excel" ? nightBillApi.exportExcel : nightBillApi.exportPdf
      const res = await fn(buildParams(filters, page, limit))
      const mime = type === "excel" ? "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" : "application/pdf"
      const blob = new Blob([res.data], { type: mime })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `night_bill_${today}.${type === "excel" ? "xlsx" : "pdf"}`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      toast.error(`Failed to export ${type.toUpperCase()}`)
    }
  }

  const columns: ColumnDef<NightBillRecord>[] = React.useMemo(() => [
    {
      id: "sl",
      header: "Sl",
      size: 40,
      cell: ({ row }) => <span className="text-xs text-muted-foreground tabular-nums">{(page - 1) * limit + row.index + 1}</span>,
    },
    {
      accessorKey: "employee_id",
      header: "Employee ID",
      cell: ({ row }) => <span className="font-mono text-xs font-semibold text-primary">{row.original.employee_id}</span>,
    },
    {
      id: "name",
      header: "Employee",
      cell: ({ row }) => (
        <div className="min-w-0">
          <p className="text-sm font-medium truncate leading-tight">{row.original.employee?.name_en || "—"}</p>
          {row.original.employee?.designationRef?.name && (
            <p className="text-[11px] text-muted-foreground truncate">{row.original.employee.designationRef.name}</p>
          )}
        </div>
      ),
    },
    {
      accessorKey: "attendance_date",
      header: "Date",
      cell: ({ row }) => {
        const d = row.original.attendance_date
        if (!d) return "—"
        const [y, m, day] = d.split("-")
        return <span className="text-xs font-medium">{day}/{m}/{y}</span>
      },
    },
    {
      id: "in_time",
      header: "In Time",
      cell: ({ row }) => <span className="text-xs font-medium text-emerald-600 tabular-nums">{formatCheck(row.original.in_time)}</span>,
    },
    {
      id: "out_time",
      header: "Out Time",
      cell: ({ row }) => <span className="text-xs font-medium text-rose-500 tabular-nums">{formatCheck(row.original.out_time)}</span>,
    },
    { accessorKey: "bill_type", header: "Bill Type", cell: ({ row }) => billTypeBadge(row.original.bill_type) },
    {
      header: "Hours",
      accessorKey: "eligible_hours",
      cell: ({ row }) => {
        const h = row.original.eligible_hours
        return <span className="text-xs tabular-nums">{h > 0 ? h : "—"}</span>
      },
    },
    {
      header: "Amount (Tk)",
      accessorKey: "amount",
      cell: ({ row }) => <span className="text-sm font-bold tabular-nums">৳{row.original.amount.toFixed(2)}</span>,
    },
    {
      accessorKey: "remarks",
      header: "Remarks",
      cell: ({ row }) => <span className="text-xs text-muted-foreground max-w-[140px] truncate block">{row.original.remarks || "—"}</span>,
    },
  ], [page, limit])

  const statCards = [
    { label: "Total Bills", value: total, icon: MoonIcon, color: "text-indigo-600 bg-indigo-500/10" },
    { label: "Total Amount", value: stats.totalAmt.toFixed(0), suffix: "Tk", icon: CoinsIcon, color: "text-emerald-600 bg-emerald-500/10" },
    { label: "Fixed Mode", value: stats.fixedCount, icon: ClockIcon, color: "text-violet-600 bg-violet-500/10" },
    { label: "Hourly Mode", value: stats.hourlyCount, icon: SunsetIcon, color: "text-amber-600 bg-amber-500/10" },
  ]

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      {/* Header */}
      <div className="px-4 lg:px-6 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex items-center gap-2">
          <MoonIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Night Bill</h1>
            <p className="text-muted-foreground mt-1">Daily night allowance for overtime &amp; post-8 PM work</p>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {selectedRows.length > 0 && (
            <Button variant="destructive" size="sm" onClick={() => { setDeleteTarget(null); setConfirmDeleteOpen(true) }} disabled={deleting}>
              <XIcon className="mr-2 h-4 w-4" />Delete ({selectedRows.length})
            </Button>
          )}
          <Button variant="outline" size="sm" onClick={() => router.push("/attendance/night-bill/employees")}>
            <UsersIcon className="mr-2 h-4 w-4" />Employee Night Bill List
          </Button>
          <Button size="sm" onClick={() => setProcessOpen(true)}>
            <ZapIcon className="mr-2 h-4 w-4" />Process
          </Button>
          <Button variant="outline" size="sm" onClick={() => setAddOpen(true)}>
            <PlusIcon className="mr-2 h-4 w-4" />Add Manual
          </Button>
          <Button variant="outline" size="sm" onClick={() => doExport("excel")}>
            <FileSpreadsheetIcon className="mr-2 h-4 w-4" />Excel
          </Button>
          <Button variant="outline" size="sm" onClick={() => doExport("pdf")}>
            <FileTextIcon className="mr-2 h-4 w-4" />PDF
          </Button>
          <Button variant="ghost" size="icon" onClick={() => fetchData()} disabled={loading}>
            <RefreshCwIcon className={cn("h-4 w-4", loading && "animate-spin")} />
          </Button>
        </div>
      </div>

      {/* Stats */}
      <div className="px-4 lg:px-6 grid grid-cols-2 sm:grid-cols-4 gap-3">
        {statCards.map((s) => (
          <Card key={s.label} className="border-0 shadow-sm">
            <CardContent className="p-3 flex items-center gap-2.5">
              <div className={cn("h-9 w-9 rounded-xl flex items-center justify-center shrink-0", s.color)}>
                <s.icon className="h-4 w-4" />
              </div>
              <div>
                <p className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wide">{s.label}</p>
                <p className="text-xl font-bold leading-none mt-0.5 tabular-nums">{s.value}{s.suffix ? <span className="text-xs font-normal ml-0.5">{s.suffix}</span> : null}</p>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Filters */}
      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar filters={filterDefs} values={filters} onChange={handleChange} onApply={handleApply} onReset={handleReset} submitting={loading} />
      </div>

      {/* Mobile filter */}
      <div className="md:hidden px-4">
        <Sheet open={mobileFilterOpen} onOpenChange={setMobileFilterOpen}>
          <SheetTrigger asChild>
            <Button variant="outline" className="w-full h-9 text-sm"><FilterIcon className="mr-2 h-4 w-4" />Filters</Button>
          </SheetTrigger>
          <SheetContent side="right" className="w-full sm:max-w-md p-0 flex flex-col" showCloseButton={false}>
            <SheetHeader className="px-4 py-3 border-b flex flex-row items-center justify-between">
              <SheetTitle className="text-base">Filters</SheetTitle>
              <SheetClose asChild><Button variant="ghost" size="icon"><XIcon className="h-4 w-4" /></Button></SheetClose>
            </SheetHeader>
            <div className="flex-1 overflow-y-auto px-4 py-4">
              <FilterBar filters={filterDefs} values={filters} onChange={handleChange} onApply={() => { handleApply(); setMobileFilterOpen(false) }} onReset={() => { handleReset(); setMobileFilterOpen(false) }} submitting={loading} singleColumn noBorder />
            </div>
          </SheetContent>
        </Sheet>
      </div>

      {/* Table */}
      <div className="px-4 lg:px-6">
        <DataTable
          data={data}
          columns={columns}
          serverSide
          page={page}
          pageSize={limit}
          pageCount={totalPages}
          total={total}
          onPageChange={setPage}
          onPageSizeChange={(size) => { setLimit(size); setPage(1) }}
          loading={loading}
          enableSelection
          onSelectionChange={setSelectedRows}
          onDelete={(row) => { setDeleteTarget(row); setConfirmDeleteOpen(true) }}
        />
      </div>

      {/* Process dialog */}
      <Dialog open={processOpen} onOpenChange={setProcessOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Process Night Bill</DialogTitle>
            <DialogDescription>Auto-calculate night bills from attendance &amp; the Employee Night Bill List</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleProcess} className="grid gap-4 py-2">
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <Label>From Date *</Label>
                <DatePicker
                  value={processStartDate ? new Date(processStartDate) : undefined}
                  onChange={(d) => setProcessStartDate(d ? format(d, "yyyy-MM-dd") : "")}
                  placeholder="From date"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label>To Date *</Label>
                <DatePicker
                  value={processEndDate ? new Date(processEndDate) : undefined}
                  onChange={(d) => setProcessEndDate(d ? format(d, "yyyy-MM-dd") : "")}
                  placeholder="To date"
                />
              </div>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Company *</Label>
              <select
                className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                value={processCompanyId}
                onChange={(e) => setProcessCompanyId(e.target.value)}
                required
              >
                <option value="">Select company</option>
                {companies.map((c) => (
                  <option key={c.id} value={c.id}>{c.company_name_en || c.id}</option>
                ))}
              </select>
            </div>

            <div className="flex justify-end gap-2 mt-2">
              <Button type="button" variant="outline" onClick={() => setProcessOpen(false)} disabled={processing}>Cancel</Button>
              <Button type="submit" disabled={processing}>
                {processing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <ZapIcon className="mr-2 h-4 w-4" />}
                Run Process
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Add manual dialog */}
      <Dialog open={addOpen} onOpenChange={setAddOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Add Night Bill</DialogTitle>
            <DialogDescription>Manually add a night bill for an employee</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleAdd} className="grid gap-4 py-2">
            <div className="flex flex-col gap-1.5">
              <Label>Company *</Label>
              <Select value={addCompanyId} onValueChange={setAddCompanyId} required>
                <SelectTrigger><SelectValue placeholder="Select company" /></SelectTrigger>
                <SelectContent>
                  {companies.map((c) => <SelectItem key={c.id} value={c.id}>{c.company_name_en || c.id}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Employee ID *</Label>
              <Input value={addEmployeeId} onChange={(e) => setAddEmployeeId(e.target.value)} placeholder="e.g. 12" />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5"><Label>Date *</Label><Input type="date" value={addDate} onChange={(e) => setAddDate(e.target.value)} required /></div>
              <div className="flex flex-col gap-1.5"><Label>Amount (Tk) *</Label><Input type="number" step="0.01" min="0" value={addAmount} onChange={(e) => setAddAmount(e.target.value)} required /></div>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Remarks</Label>
              <Input value={addRemarks} onChange={(e) => setAddRemarks(e.target.value)} />
            </div>
            <div className="flex justify-end gap-2 mt-2">
              <Button type="button" variant="outline" onClick={() => setAddOpen(false)} disabled={adding}>Cancel</Button>
              <Button type="submit" disabled={adding}>{adding ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <PlusIcon className="mr-2 h-4 w-4" />}Save</Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Edit dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Edit Night Bill</DialogTitle>
            <DialogDescription>
              {editRecord?.employee?.name_en} &mdash; {editRecord?.employee_id} &mdash; {editRecord?.attendance_date}
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleUpdate} className="grid gap-4 py-2">
            <div className="flex flex-col gap-1.5">
              <Label>Amount (Tk) *</Label>
              <Input type="number" step="0.01" min="0" value={editAmount} onChange={(e) => setEditAmount(e.target.value)} required />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Remarks</Label>
              <Input value={editRemarks} onChange={(e) => setEditRemarks(e.target.value)} />
            </div>
            <div className="flex justify-end gap-2 mt-2">
              <Button type="button" variant="outline" onClick={() => setEditOpen(false)} disabled={updating}>Cancel</Button>
              <Button type="submit" disabled={updating}>{updating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}Update</Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <AlertDialog open={confirmDeleteOpen} onOpenChange={setConfirmDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Night Bill Record{!deleteTarget && "s"}</AlertDialogTitle>
            <AlertDialogDescription>
              {deleteTarget
                ? "Are you sure you want to delete this night bill record? This cannot be undone."
                : `Delete ${selectedRows.length} selected record(s)? This cannot be undone.`}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
              {deleting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}