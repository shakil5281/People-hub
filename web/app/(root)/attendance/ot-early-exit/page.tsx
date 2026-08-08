"use client"

import * as React from "react"
import { ClockIcon, FileSpreadsheetIcon, Loader2, FilterIcon, XIcon } from "lucide-react"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
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
  const [totalShortfall, setTotalShortfall] = React.useState(0)
  const [affectedEmployees, setAffectedEmployees] = React.useState(0)
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
  ]

  const activeParams = (f?: Record<string, string>) => {
    const params = f || filters
    const active: Record<string, string> = {
      company_id: params.company_id || "",
      month: params.month || String(currentMonth),
      year: params.year || String(currentYear),
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

  const fetchData = async (f?: Record<string, string>) => {
    setLoading(true)
    try {
      const { data: res } = await otEarlyExitApi.list(activeParams(f))
      setData(res.records || [])
      setTotalShortfall(res.total_shortfall || 0)
      setAffectedEmployees(res.affected_employees || 0)
    } catch {
      setData([])
      setTotalShortfall(0)
      setAffectedEmployees(0)
    } finally {
      setLoading(false)
    }
  }

  const handleApply = () => {
    setSubmitting(true)
    fetchData(filters).finally(() => setSubmitting(false))
  }

  const handleReset = () => {
    setFilters({ month: String(currentMonth), year: String(currentYear) })
    setData([])
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
      await fetchData(filters)
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
          <p className="text-2xl font-bold mt-1">{data.length}</p>
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

      <DataTable data={data} columns={columns} loading={loading} />
    </div>
  )
}