"use client"

import * as React from "react"
import { ClockIcon, FileSpreadsheetIcon, Loader2, FilterIcon, XIcon, Trash2Icon } from "lucide-react"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { otEarlyExitApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"
import { toast } from "sonner"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"

interface EarlyExitRecord {
  id: string
  employee_id: string
  employee_name: string
  designation: string
  department: string
  date: string
  status: string
  shift_start: string
  shift_end: string
  expected_hours: number
  worked_hours: number
  shortfall_hours: number
  company_id?: string
}

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

const now = new Date()
const currentMonth = now.getMonth() + 1
const currentYear = now.getFullYear()

export default function OtEarlyExitPage() {
  const [data, setData] = React.useState<EarlyExitRecord[]>([])
  const [loading, setLoading] = React.useState(false)
  const [submitting, setSubmitting] = React.useState(false)
  const [processing, setProcessing] = React.useState(false)
  const [exporting, setExporting] = React.useState(false)
  const [recordToRemove, setRecordToRemove] = React.useState<EarlyExitRecord | null>(null)
  const [removing, setRemoving] = React.useState(false)
  const [totalShortfall, setTotalShortfall] = React.useState(0)
  const [affectedEmployees, setAffectedEmployees] = React.useState(0)
  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(1)
  const [filters, setFilters] = React.useState<Record<string, string>>({
    month: String(currentMonth),
    year: String(currentYear),
  })
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)
  const [error, setError] = React.useState("")

  const columns: ColumnDef<EarlyExitRecord>[] = [
    { id: "sl", header: "Sl", cell: ({ row }) => row.index + 1 },
    { accessorKey: "employee_name", header: "Employee Name" },
    { accessorKey: "employee_id", header: "Emp. ID" },
    { accessorKey: "department", header: "Department" },
    { accessorKey: "designation", header: "Designation" },
    { accessorKey: "date", header: "Date", cell: ({ row }) => row.original.date ? row.original.date.split("T")[0].split("-").reverse().join("-") : "-" },
    { accessorKey: "shift_start", header: "Shift", cell: ({ row }) => `${row.original.shift_start} - ${row.original.shift_end}` },
    { accessorKey: "expected_hours", header: "Expected", cell: ({ row }) => `${Math.round(row.original.expected_hours ?? 0)} hrs` },
    { accessorKey: "worked_hours", header: "Worked", cell: ({ row }) => `${Math.round(row.original.worked_hours ?? 0)} hrs` },
    { accessorKey: "shortfall_hours", header: "OT Deducted", cell: ({ row }) => `${Math.round(row.original.shortfall_hours ?? 0)} hrs` },
    {
      id: "actions",
      header: "Actions",
      cell: ({ row }) => {
        const rec = row.original
        return (
          <Button
            variant="ghost"
            size="sm"
            className="h-8 px-2.5 text-destructive hover:text-destructive hover:bg-destructive/10 text-xs font-medium"
            onClick={() => setRecordToRemove(rec)}
            title="Remove OT Early Exit deduction"
          >
            <Trash2Icon className="h-3.5 w-3.5 mr-1" />
            Remove
          </Button>
        )
      },
    },
  ]

  const activeParams = (f?: Record<string, string>, p?: number, l?: number) => {
    const params = f || filters
    const active: Record<string, string> = {
      company_id: params.company_id || "",
      month: params.month || String(currentMonth),
      year: params.year || String(currentYear),
      page: String(p ?? page),
      limit: String(l ?? limit),
    }
    if (params.department_id) active.department_id = params.department_id
    if (params.section_id) active.section_id = params.section_id
    if (params.designation_id) active.designation_id = params.designation_id
    if (params.line_id) active.line_id = params.line_id
    if (params.group_id) active.group_id = params.group_id
    if (params.shift_id) active.shift_id = params.shift_id
    if (params.employee_id) active.employee_id = params.employee_id
    return active
  }

  const fetchData = async (f?: Record<string, string>, p?: number, l?: number) => {
    setLoading(true)
    try {
      const { data: res } = await otEarlyExitApi.list(activeParams(f, p, l))
      setData(res.records || [])
      setTotal(res.total ?? 0)
      setTotalPages(Math.max(1, Math.ceil((res.total ?? 0) / (l ?? limit))))
      setTotalShortfall(res.total_shortfall || 0)
      setAffectedEmployees(res.affected_employees || 0)
    } catch {
      setData([])
      setTotal(0)
      setTotalPages(1)
      setTotalShortfall(0)
      setAffectedEmployees(0)
    } finally {
      setLoading(false)
    }
  }

  const handleApply = () => {
    setSubmitting(true)
    setPage(1)
    fetchData(filters, 1).finally(() => setSubmitting(false))
  }

  const handleReset = () => {
    setFilters({ month: String(currentMonth), year: String(currentYear) })
    setPage(1)
    setData([])
    setTotal(0)
    setTotalPages(1)
    setTotalShortfall(0)
    setAffectedEmployees(0)
  }

  const handleFilterChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleProcess = async () => {
    const companyID = filters.company_id
    if (!companyID) {
      toast.error("Please select a company first")
      return
    }
    setProcessing(true)
    try {
      const { data: res } = await otEarlyExitApi.process({
        company_id: companyID,
        month: Number(filters.month || currentMonth),
        year: Number(filters.year || currentYear),
      })
      toast.success(res.message || "Early-exit deductions computed")
      setPage(1)
      fetchData(filters, 1)
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
      toast.error(msg || "Failed to compute early-exit deductions")
    } finally {
      setProcessing(false)
    }
  }

  const handleExport = async () => {
    setExporting(true)
    try {
      const res = await otEarlyExitApi.exportExcel(activeParams(filters))
      const blob = new Blob([res.data], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `ot_early_exit_${filters.year || currentYear}_${filters.month || currentMonth}.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setError("Failed to export early-exit deductions")
    } finally {
      setExporting(false)
    }
  }

  const handleRemoveDeduction = async () => {
    if (!recordToRemove) return
    setRemoving(true)
    try {
      await otEarlyExitApi.remove(recordToRemove.id, { reason: "Manually removed from OT Early Exit page" })
      toast.success(`Removed OT Early Exit deduction for ${recordToRemove.employee_name}. This will not be deducted in future salary processes.`)
      setRecordToRemove(null)
      fetchData(filters, page)
    } catch (err: any) {
      toast.error(err?.response?.data?.error || "Failed to remove OT Early Exit deduction")
    } finally {
      setRemoving(false)
    }
  }

  const handleExemptFullMonth = async () => {
    if (!recordToRemove) return
    setRemoving(true)
    try {
      await otEarlyExitApi.exemptEmployee({
        company_id: recordToRemove.company_id || filters.company_id,
        employee_id: recordToRemove.employee_id,
        month: Number(filters.month || currentMonth),
        year: Number(filters.year || currentYear),
        reason: "Exempted full month by admin",
      })
      toast.success(`Exempted employee ${recordToRemove.employee_name} (${recordToRemove.employee_id}) from all OT Early Exit deductions for this month.`)
      setRecordToRemove(null)
      fetchData(filters, page)
    } catch (err: any) {
      toast.error(err?.response?.data?.error || "Failed to exempt employee")
    } finally {
      setRemoving(false)
    }
  }

  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      sectionApi.list(undefined, { limit: "100" }),
      designationApi.list(undefined, { limit: "100" }),
      lineApi.list(undefined, { limit: "100" }),
      groupApi.list({ limit: "100" }),
      shiftApi.list({ limit: "100" }),
    ]).then(([cRes, dRes, secRes, desRes, lRes, gRes, sRes]) => {
      const cList = cRes.data?.data || []
      setCompanies(cList)
      setDepartments(dRes.data?.data || [])
      setSections(secRes.data?.data || [])
      setDesignations(desRes.data?.data || [])
      setLines(lRes.data?.data || [])
      setGroups(gRes.data?.data || [])
      setShifts(sRes.data?.data || [])

      let initialCompany = ""
      if (cList.length > 0) {
        initialCompany = cList[0].id
        setFilters((prev) => ({ ...prev, company_id: initialCompany }))
      }
      fetchData({ month: String(currentMonth), year: String(currentYear), company_id: initialCompany })
    }).catch(() => {
      fetchData()
    })
  }, [])

  React.useEffect(() => { fetchData() }, [page, limit]) // eslint-disable-line react-hooks/exhaustive-deps

  const filterDefs: FilterDef[] = React.useMemo(() => [
    {
      key: "company_id", label: "Company", type: "select",
      options: companies.map((c) => ({ value: c.id, label: c.company_name_en })),
    },
    {
      key: "month", label: "Month", type: "select", options: [
        { value: "1", label: "January" }, { value: "2", label: "February" },
        { value: "3", label: "March" }, { value: "4", label: "April" },
        { value: "5", label: "May" }, { value: "6", label: "June" },
        { value: "7", label: "July" }, { value: "8", label: "August" },
        { value: "9", label: "September" }, { value: "10", label: "October" },
        { value: "11", label: "November" }, { value: "12", label: "December" },
      ],
    },
    { key: "year", label: "Year", type: "text" },
    {
      key: "department_id", label: "Department", type: "select",
      options: departments.map((d) => ({ value: d.id, label: d.name })),
    },
    {
      key: "section_id", label: "Section", type: "select",
      options: sections.map((s) => ({ value: s.id, label: s.name })),
    },
    {
      key: "designation_id", label: "Designation", type: "select",
      options: designations.map((d) => ({ value: d.id, label: d.name })),
    },
    {
      key: "line_id", label: "Line", type: "select",
      options: lines.map((l) => ({ value: l.id, label: l.name })),
    },
    {
      key: "group_id", label: "Group", type: "select",
      options: groups.map((g) => ({ value: g.id, label: g.name })),
    },
    {
      key: "shift_id", label: "Shift", type: "select",
      options: shifts.map((s) => ({ value: s.id, label: s.name })),
    },
    { key: "employee_id", label: "Emp. ID", type: "text" },
  ], [companies, departments, sections, designations, lines, groups, shifts])

  const processDisabled = !filters.company_id

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <ClockIcon className="h-6 w-6 text-muted-foreground" />
            <div>
              <h1 className="text-3xl font-bold tracking-tight">OT Early Exit</h1>
              <p className="text-muted-foreground mt-1">Monthly overtime deductions for early departure</p>
            </div>
          </div>
          <div className="hidden md:flex gap-2">
            <Button
              onClick={handleProcess}
              disabled={processing || processDisabled}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {processing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              {processing ? "Computing..." : "Compute Deductions"}
            </Button>
            <Button onClick={handleExport} disabled={exporting} className="bg-primary text-primary-foreground hover:bg-primary/90">
              {exporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileSpreadsheetIcon className="mr-2 h-4 w-4" />}
              {exporting ? "Exporting..." : "Export Excel"}
            </Button>
          </div>
        </div>
        <div className="md:hidden mt-3">
          <ButtonGroup className="w-full">
            <Button onClick={handleProcess} disabled={processing || processDisabled} className="flex-1 bg-primary text-primary-foreground hover:bg-primary/90">
              {processing ? "Computing..." : "Compute"}
            </Button>
            <Button onClick={handleExport} disabled={exporting} variant="outline" className="flex-1">
              {exporting ? "Exporting..." : "Export"}
            </Button>
            <Sheet open={mobileFilterOpen} onOpenChange={setMobileFilterOpen}>
              <SheetTrigger asChild>
                <Button variant="outline" className="flex-1">
                  <FilterIcon className="mr-2 h-4 w-4" />
                  Filters
                </Button>
              </SheetTrigger>
              <SheetContent side="right" className="w-full sm:max-w-md p-0 flex flex-col" showCloseButton={false}>
                <SheetHeader className="px-4 py-3 border-b flex flex-row items-center justify-between">
                  <SheetTitle className="text-base">Filters</SheetTitle>
                  <SheetClose asChild>
                    <Button variant="ghost" size="icon-sm">
                      <XIcon className="h-4 w-4" />
                    </Button>
                  </SheetClose>
                </SheetHeader>
                <div className="flex-1 overflow-y-auto px-4 py-4">
                  <FilterBar
                    filters={filterDefs}
                    values={filters}
                    onChange={handleFilterChange}
                    onApply={() => { handleApply(); setMobileFilterOpen(false) }}
                    onReset={() => { handleReset(); setMobileFilterOpen(false) }}
                    submitting={submitting}
                    singleColumn
                    noBorder
                  />
                </div>
              </SheetContent>
            </Sheet>
          </ButtonGroup>
        </div>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleFilterChange}
          onApply={handleApply}
          onReset={handleReset}
          submitting={submitting}
        />
      </div>

      <div className="px-4 lg:px-6 grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="rounded-lg border bg-card p-4">
          <p className="text-sm text-muted-foreground">Total Deduction Records</p>
          <p className="text-2xl font-bold mt-1">{total}</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <p className="text-sm text-muted-foreground">Affected Employees</p>
          <p className="text-2xl font-bold mt-1">{affectedEmployees}</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <p className="text-sm text-muted-foreground">Total OT Shortfall Hours Deducted</p>
          <p className="text-2xl font-bold mt-1">{Math.round(totalShortfall)} hrs</p>
        </div>
      </div>

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      <DataTable
        data={data}
        columns={columns}
        loading={loading}
        serverSide
        page={page}
        pageSize={limit}
        pageCount={totalPages}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={(size) => { setLimit(size); setPage(1) }}
      />

      <AlertDialog open={!!recordToRemove} onOpenChange={(open) => { if (!open) setRecordToRemove(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove OT Early Exit Deduction</AlertDialogTitle>
            <AlertDialogDescription asChild>
              <div className="space-y-2 text-sm text-muted-foreground">
                <p>
                  Are you sure you want to remove OT Early Exit deduction for <b className="text-foreground">{recordToRemove?.employee_name}</b> (Emp. ID: <b className="text-foreground">{recordToRemove?.employee_id}</b>) on {recordToRemove?.date ? recordToRemove.date.split("T")[0].split("-").reverse().join("-") : ""}?
                </p>
                <div className="rounded-md bg-muted/60 p-2.5 text-xs text-foreground mt-2 border">
                  ✓ <b>Protected in future salary processes</b>: When salary process is run again, this employee will <b>not</b> have OT deducted for this early exit.
                </div>
              </div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter className="flex-col sm:flex-row gap-2 mt-2">
            <AlertDialogCancel disabled={removing}>Cancel</AlertDialogCancel>
            <Button
              variant="outline"
              disabled={removing}
              onClick={handleExemptFullMonth}
              className="text-xs"
            >
              Exempt Full Month
            </Button>
            <AlertDialogAction
              disabled={removing}
              onClick={handleRemoveDeduction}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {removing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              {removing ? "Removing..." : "Remove Deduction"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}