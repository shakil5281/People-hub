"use client"

import * as React from "react"
import { ClipboardCheckIcon, FileSpreadsheetIcon, UserXIcon, Loader2, FilterIcon, XIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { format } from "date-fns"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetFooter, SheetClose } from "@/components/ui/sheet"
import { attendanceApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"
import { formatCheck } from "@/lib/utils"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

interface AttendanceRecord {
  id: string
  employee_id: string
  employee_name: string
  designation: string
  company_id: string
  date: string
  check_in: string | null
  check_out: string | null
  total_hours: string | null
  over_time: string | null
  status: string
  late_minutes: number
  shift_name: string
  punch_number: string | null
}

const columns: ColumnDef<AttendanceRecord>[] = [
  { accessorKey: "employee_id", header: "Employee ID" },
  { accessorKey: "employee_name", header: "Name", cell: ({ row }) => row.original.employee_name || "-" },
  { accessorKey: "designation", header: "Designation", cell: ({ row }) => row.original.designation || "-" },
  { accessorKey: "check_in", header: "Check In", cell: ({ row }) => formatCheck(row.original.check_in) },
  { accessorKey: "check_out", header: "Check Out", cell: ({ row }) => formatCheck(row.original.check_out) },
  { accessorKey: "total_hours", header: "Total Hours", cell: ({ row }) => row.original.total_hours || "-" },
  { accessorKey: "over_time", header: "Over Time", cell: ({ row }) => row.original.over_time || "-" },
  { accessorKey: "late_minutes", header: "Late (min)" },
  { accessorKey: "shift_name", header: "Shift", cell: ({ row }) => row.original.shift_name || "-" },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const variant = row.original.status === "present" ? "default" : row.original.status === "late" ? "destructive" : "secondary"
      return <Badge variant={variant} className="capitalize">{row.original.status.replace("_", " ")}</Badge>
    },
  },
]

const today = new Date().toISOString().split("T")[0]

export default function DailyAttendancePage() {
  const [data, setData] = React.useState<AttendanceRecord[]>([])
  const [loading, setLoading] = React.useState(true)
  const [exporting, setExporting] = React.useState(false)
  const [exportingAbsent, setExportingAbsent] = React.useState(false)
  const [exportingMissing, setExportingMissing] = React.useState(false)
  const [error, setError] = React.useState("")
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])
  const [filters, setFilters] = React.useState<Record<string, string>>({
    date: today,
  })

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date", label: "Date", type: "datepicker" },
    {
      key: "company_id", label: "Company", type: "select",
      options: companies.map((c) => ({ value: c.id, label: c.company_name_en })),
    },
    {
      key: "department_id", label: "Department", type: "select",
      options: departments.map((d) => ({ value: d.id, label: d.name })),
    },
    {
      key: "section_id", label: "Section", type: "select",
      options: sections.map((s) => ({ value: s.id, label: s.name })),
      disabled: !filters.department_id,
    },
    {
      key: "designation_id", label: "Designation", type: "select",
      options: designations.map((d) => ({ value: d.id, label: d.name })),
      disabled: !filters.section_id,
    },
    {
      key: "line_id", label: "Line", type: "select",
      options: lines.map((l) => ({ value: l.id, label: l.name })),
      disabled: !filters.section_id,
    },
    {
      key: "group_id", label: "Group", type: "select",
      options: groups.map((g) => ({ value: g.id, label: g.name })),
    },
    {
      key: "shift_id", label: "Shift", type: "select",
      options: shifts.map((s) => ({ value: s.id, label: s.name })),
    },
    {
      key: "status", label: "Status", type: "select",
      options: [
        { value: "present", label: "Present" },
        { value: "late", label: "Late" },
        { value: "absent", label: "Absent" },
        { value: "half_day", label: "Half Day" },
      ],
    },
    { key: "employee_id", label: "Employee ID", type: "text", placeholder: "Enter employee code..." },
  ], [companies, departments, sections, designations, lines, groups, shifts, filters.department_id, filters.section_id])

  const fetchData = React.useCallback(async (params: Record<string, string>, p?: number, l?: number) => {
    setLoading(true)
    setError("")
    try {
      const reqParams = { ...params, page: String(p ?? page), limit: String(l ?? limit) }
      const { data: res } = await attendanceApi.list(reqParams)
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setError("Failed to load attendance")
    } finally {
      setLoading(false)
    }
  }, [page, limit])

  const loadSections = React.useCallback(async (departmentId: string) => {
    try {
      const res = await sectionApi.list(departmentId, { limit: "100" })
      setSections(Array.isArray(res.data?.data) ? res.data.data : [])
    } catch {
      setSections([])
    }
  }, [])

  const loadDesignations = React.useCallback(async (sectionId: string) => {
    try {
      const res = await designationApi.list(sectionId, { limit: "100" })
      setDesignations(Array.isArray(res.data?.data) ? res.data.data : [])
    } catch {
      setDesignations([])
    }
  }, [])

  const loadLines = React.useCallback(async (sectionId: string) => {
    try {
      const res = await lineApi.list(sectionId, { limit: "100" })
      setLines(Array.isArray(res.data?.data) ? res.data.data : [])
    } catch {
      setLines([])
    }
  }, [])

  React.useEffect(() => {
    const init = async () => {
      const [cRes, dRes, gRes, sRes] = await Promise.all([
        companyApi.list({ limit: "100" }),
        departmentApi.list({ limit: "100" }),
        groupApi.list({ limit: "100" }),
        shiftApi.list({ limit: "100" }),
      ])
      setCompanies(Array.isArray(cRes.data?.data) ? cRes.data.data : [])
      setDepartments(Array.isArray(dRes.data?.data) ? dRes.data.data : [])
      setGroups(Array.isArray(gRes.data?.data) ? gRes.data.data : [])
      setShifts(Array.isArray(sRes.data?.data) ? sRes.data.data : [])
    }
    init()
    fetchData({ date: today }, 1, 20)
  }, [])

  React.useEffect(() => {
    if (!filters.department_id) {
      setSections([])
      setDesignations([])
      setLines([])
      setFilters((prev) => {
        const next = { ...prev }
        delete next.section_id
        delete next.designation_id
        delete next.line_id
        return next
      })
      return
    }
    loadSections(filters.department_id)
  }, [filters.department_id])

  React.useEffect(() => {
    if (!filters.section_id) {
      setDesignations([])
      setLines([])
      setFilters((prev) => {
        const next = { ...prev }
        delete next.designation_id
        delete next.line_id
        return next
      })
      return
    }
    loadDesignations(filters.section_id)
    loadLines(filters.section_id)
  }, [filters.section_id])

  React.useEffect(() => {
    const active: Record<string, string> = {}
    for (const [key, value] of Object.entries(filters)) {
      if (value) active[key] = value
    }
    fetchData(active)
  }, [page, limit])

  const handleChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleApply = () => {
    setPage(1)
    const active: Record<string, string> = {}
    for (const [key, value] of Object.entries(filters)) {
      if (value) active[key] = value
    }
    fetchData(active, 1)
    setMobileFilterOpen(false)
  }

  const handleReset = () => {
    setPage(1)
    setLimit(20)
    setDepartments([])
    setSections([])
    setDesignations([])
    setLines([])
    setFilters({ date: today })
    fetchData({ date: today }, 1, 20)
    setMobileFilterOpen(false)
  }

  const buildExportParams = () => {
    const params: Record<string, string> = { date: filters.date || today }
    const filterKeys = ["company_id", "department_id", "section_id", "designation_id", "line_id", "group_id", "shift_id", "status", "employee_id"]
    for (const key of filterKeys) {
      if (filters[key]) params[key] = filters[key]
    }
    return params
  }

  const handleExport = async () => {
    setExporting(true)
    try {
      const res = await attendanceApi.exportExcel(buildExportParams())
      const blob = new Blob([res.data], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `attendance_${filters.date || today}.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setError("Failed to export attendance")
    } finally {
      setExporting(false)
    }
  }

  const buildAbsentExportParams = () => {
    const date = filters.date || today
    return { start_date: date, end_date: date }
  }

  const handleExportAbsent = async () => {
    setExportingAbsent(true)
    try {
      const res = await attendanceApi.exportAbsentExcel(buildAbsentExportParams())
      const blob = new Blob([res.data], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `absent_report_${filters.date || today}.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setError("Failed to export absent report")
    } finally {
      setExportingAbsent(false)
    }
  }

  const buildMissingExportParams = () => {
    const date = filters.date || today
    return { start_date: date, end_date: date }
  }

  const handleExportMissing = async () => {
    setExportingMissing(true)
    try {
      const res = await attendanceApi.exportMissingExcel(buildMissingExportParams())
      const blob = new Blob([res.data], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `missing_attendance_${filters.date || today}.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setError("Failed to export missing attendance")
    } finally {
      setExportingMissing(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <ClipboardCheckIcon className="h-6 w-6 text-muted-foreground" />
            <div>
              <h1 className="text-3xl font-bold tracking-tight">Daily Attendance</h1>
              <p className="text-muted-foreground mt-1">View and manage daily attendance records</p>
            </div>
          </div>
          <div className="hidden md:flex gap-2">
            <Button onClick={handleExport} disabled={exporting} className="bg-primary text-primary-foreground hover:bg-primary/90">
              {exporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileSpreadsheetIcon className="mr-2 h-4 w-4" />}
              {exporting ? "Exporting..." : "Export Excel"}
            </Button>
            <Button onClick={handleExportAbsent} disabled={exportingAbsent} variant="destructive">
              {exportingAbsent ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <UserXIcon className="mr-2 h-4 w-4" />}
              {exportingAbsent ? "Exporting..." : "Export Absent"}
            </Button>
            <Button onClick={handleExportMissing} disabled={exportingMissing} variant="secondary">
              {exportingMissing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <ClipboardCheckIcon className="mr-2 h-4 w-4" />}
              {exportingMissing ? "Exporting..." : "Export Missing"}
            </Button>
          </div>
        </div>
        <div className="md:hidden mt-3">
          <ButtonGroup className="w-full">
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
                    onChange={handleChange}
                    onApply={handleApply}
                    onReset={handleReset}
                    submitting={loading}
                    singleColumn
                    noBorder
                  />
                </div>
              </SheetContent>
            </Sheet>
            <Button onClick={handleExport} disabled={exporting} className="flex-1 bg-primary text-primary-foreground hover:bg-primary/90">
              {exporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileSpreadsheetIcon className="mr-2 h-4 w-4" />}
              {exporting ? "Exporting..." : "Export"}
            </Button>
            <Button onClick={handleExportAbsent} disabled={exportingAbsent} variant="destructive" className="flex-1">
              {exportingAbsent ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <UserXIcon className="mr-2 h-4 w-4" />}
              {exportingAbsent ? "Exporting..." : "Absent"}
            </Button>
            <Button onClick={handleExportMissing} disabled={exportingMissing} variant="secondary" className="flex-1">
              {exportingMissing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <ClipboardCheckIcon className="mr-2 h-4 w-4" />}
              {exportingMissing ? "Exporting..." : "Missing"}
            </Button>
          </ButtonGroup>
        </div>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleChange}
          onApply={handleApply}
          onReset={handleReset}
          submitting={loading}
        />
      </div>

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      <div className="px-4 lg:px-6">
        <h2 className="text-lg font-semibold mb-2">
          Attendance for{" "}
          {filters.date
            ? (() => {
                const parts = filters.date.split("-")
                return parts.length === 3 ? `${parts[2]}-${parts[1]}-${parts[0]}` : filters.date
              })()
            : "-"}
        </h2>
      </div>

      {loading || data.length > 0 ? (
        <DataTable
          data={data}
          columns={columns}
          enableSelection={false}
          serverSide={true}
          page={page}
          pageSize={limit}
          pageCount={totalPages}
          total={total}
          onPageChange={setPage}
          onPageSizeChange={(size) => { setLimit(size); setPage(1); }}
          loading={loading}
        />
      ) : (
        <div className="px-4 lg:px-6">
          <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-12 text-center">
            <ClipboardCheckIcon className="mb-3 h-10 w-10 text-muted-foreground/60" />
            <h3 className="text-lg font-semibold">No attendance records</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              No attendance data found for the selected date. Try a different date or process data logs.
            </p>
          </div>
        </div>
      )}
    </div>
  )
}
