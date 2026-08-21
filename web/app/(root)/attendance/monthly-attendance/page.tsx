"use client"

import * as React from "react"
import { BarChart3Icon, CalendarRangeIcon, Building2Icon, UsersIcon, SearchIcon, FileSpreadsheetIcon, FileTextIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { attendanceApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, shiftApi, groupApi } from "@/lib/api"
import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Shift { id: string; name: string }
interface Group { id: string; name: string }

interface MonthlyRecord {
  id: string
  employee_id: string
  emp_id: string
  employee_name: string
  designation_name: string
  department_name: string
  shift_name: string
  present: number
  absent: number
  late: number
  leave: number
  weekend: number
  half_day: number
  holiday: number
  over_time: number
  total_days: number
}

interface Totals {
  present: number
  absent: number
  late: number
  leave: number
  weekend: number
  half_day: number
  holiday: number
  over_time: number
}

const currentYear = new Date().getFullYear()
const currentMonth = new Date().getMonth()
const MONTHS = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
]
const YEARS = Array.from({ length: 10 }, (_, i) => currentYear - 5 + i)

const numBody = "text-center font-medium tabular-nums"
const columns: ColumnDef<MonthlyRecord>[] = [
  { id: "sl", header: "Sl", cell: ({ row }) => <span className="text-center">{row.index + 1}</span> },
  { accessorKey: "emp_id", header: "Emp. ID", cell: ({ getValue }) => <span className="font-semibold tabular-nums">{getValue() as string}</span> },
  { accessorKey: "employee_name", header: "Name" },
  { accessorKey: "designation_name", header: "Designation" },
  { accessorKey: "department_name", header: "Department", cell: ({ getValue }) => <span className="text-muted-foreground">{getValue() as string}</span> },
  { accessorKey: "shift_name", header: "Shift" },
  { accessorKey: "present", header: "Pr", cell: ({ getValue }) => <span className={`${numBody} text-emerald-600`}>{getValue() as number}</span> },
  { accessorKey: "absent", header: "Ab", cell: ({ getValue }) => <span className={`${numBody} text-red-600`}>{getValue() as number}</span> },
  { accessorKey: "late", header: "Lt", cell: ({ getValue }) => <span className={`${numBody} text-amber-600`}>{getValue() as number}</span> },
  { accessorKey: "leave", header: "Lv", cell: ({ getValue }) => <span className={`${numBody} text-sky-600`}>{getValue() as number}</span> },
  { accessorKey: "weekend", header: "Wk", cell: ({ getValue }) => <span className={`${numBody} text-violet-600`}>{getValue() as number}</span> },
  { accessorKey: "half_day", header: "HD", cell: ({ getValue }) => <span className={`${numBody} text-pink-600`}>{getValue() as number}</span> },
  { accessorKey: "holiday", header: "Ho", cell: ({ getValue }) => <span className={`${numBody} text-teal-600`}>{getValue() as number}</span> },
  { accessorKey: "over_time", header: "OT", cell: ({ getValue }) => <span className={`${numBody} text-indigo-600`}>{getValue() as number}</span> },
  { accessorKey: "total_days", header: "Total", cell: ({ getValue }) => <span className={`${numBody} font-bold`}>{getValue() as number}</span> },
]

export default function MonthlyAttendancePage() {
  const [data, setData] = React.useState<MonthlyRecord[]>([])
  const [totals, setTotals] = React.useState<Totals | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [exporting, setExporting] = React.useState<string | null>(null)
  const [error, setError] = React.useState("")

  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])

  const [selectedMonth, setSelectedMonth] = React.useState(currentMonth)
  const [selectedYear, setSelectedYear] = React.useState(currentYear)
  const [companyId, setCompanyId] = React.useState("")
  const [departmentId, setDepartmentId] = React.useState("")
  const [sectionId, setSectionId] = React.useState("")
  const [designationId, setDesignationId] = React.useState("")
  const [lineId, setLineId] = React.useState("")
  const [shiftId, setShiftId] = React.useState("")
  const [groupId, setGroupId] = React.useState("")
  const [employeeId, setEmployeeId] = React.useState("")

  const fetchSections = React.useCallback(async (deptId: string) => {
    try {
      const { data } = await sectionApi.list(deptId || undefined)
      setSections(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setSections([]) }
  }, [])

  const fetchDesignations = React.useCallback(async (secId: string) => {
    try {
      const { data } = await designationApi.list(secId || undefined)
      setDesignations(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setDesignations([]) }
  }, [])

  const fetchLines = React.useCallback(async (secId: string) => {
    try {
      const { data } = await lineApi.list(secId || undefined)
      setLines(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setLines([]) }
  }, [])

  React.useEffect(() => {
    fetchSections(departmentId)
    setSectionId("")
    setDesignationId("")
    setLineId("")
    setDesignations([])
    setLines([])
  }, [departmentId, fetchSections])

  React.useEffect(() => {
    fetchDesignations(sectionId)
    fetchLines(sectionId)
    setDesignationId("")
    setLineId("")
  }, [sectionId, fetchDesignations, fetchLines])

  const buildParams = React.useCallback((): Record<string, string> => {
    const params: Record<string, string> = {
      year: String(selectedYear),
      month: String(selectedMonth + 1),
      company_id: companyId,
    }
    if (departmentId) params.department_id = departmentId
    if (sectionId) params.section_id = sectionId
    if (designationId) params.designation_id = designationId
    if (lineId) params.line_id = lineId
    if (shiftId) params.shift_id = shiftId
    if (groupId) params.group_id = groupId
    if (employeeId) params.employee_id = employeeId
    return params
  }, [selectedYear, selectedMonth, companyId, departmentId, sectionId, designationId, lineId, shiftId, groupId, employeeId])

  const fetchData = React.useCallback(async () => {
    if (!companyId) return
    setLoading(true)
    setError("")
    try {
      const { data: res } = await attendanceApi.monthlyReport(buildParams())
      const records = (res?.records || []).map((r: any, i: number) => ({
        ...r, id: r.employee_id || `emp-${i}`,
      }))
      setData(records)
      setTotals(res?.totals || null)
    } catch {
      setError("Failed to load monthly attendance report")
    } finally {
      setLoading(false)
    }
  }, [buildParams])

  React.useEffect(() => {
    const init = async () => {
      const [cRes, dRes, shiftRes, groupRes] = await Promise.all([
        companyApi.list({ limit: "100" }),
        departmentApi.list({ limit: "100" }),
        shiftApi.list({ limit: "100" }),
        groupApi.list({ limit: "100" }),
      ])
      if (Array.isArray(cRes.data?.data) && cRes.data.data.length > 0) {
        setCompanies(cRes.data.data)
        setCompanyId(cRes.data.data[0].id)
      }
      if (Array.isArray(dRes.data?.data)) setDepartments(dRes.data.data)
      if (Array.isArray(shiftRes.data?.data)) setShifts(shiftRes.data.data)
      else if (Array.isArray(shiftRes.data)) setShifts(shiftRes.data)
      if (Array.isArray(groupRes.data?.data)) setGroups(groupRes.data.data)
      else if (Array.isArray(groupRes.data)) setGroups(groupRes.data)
    }
    init()
  }, [])

  React.useEffect(() => {
    if (companyId) fetchData()
  }, [companyId, fetchData])

  const handleReset = () => {
    setCompanyId(companies[0]?.id || "")
    setDepartmentId("")
    setSectionId("")
    setDesignationId("")
    setLineId("")
    setShiftId("")
    setGroupId("")
    setEmployeeId("")
  }

  const handleExport = async (format: "excel" | "pdf") => {
    if (!companyId) return
    setExporting(format)
    try {
      const params = buildParams()
      const res = format === "excel"
        ? await attendanceApi.exportMonthlyReportExcel(params)
        : await attendanceApi.exportMonthlyReportPdf(params)
      const url = window.URL.createObjectURL(new Blob([res.data]))
      const a = document.createElement("a")
      a.href = url
      a.download = `monthly_attendance_report_${selectedMonth + 1}_${selectedYear}.${format === "excel" ? "xlsx" : "pdf"}`
      a.click()
      window.URL.revokeObjectURL(url)
    } catch {
      setError("Failed to export report")
    } finally {
      setExporting(null)
    }
  }

  const selectCls = "flex h-10 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30"
  const labelCls = "text-xs font-medium text-muted-foreground"

  const statCards = totals ? [
    { label: "Present", value: totals.present, cls: "text-emerald-600 border-emerald-200 bg-emerald-50/60" },
    { label: "Absent", value: totals.absent, cls: "text-red-600 border-red-200 bg-red-50/60" },
    { label: "Late", value: totals.late, cls: "text-amber-600 border-amber-200 bg-amber-50/60" },
    { label: "Leave", value: totals.leave, cls: "text-sky-600 border-sky-200 bg-sky-50/60" },
    { label: "Weekend", value: totals.weekend, cls: "text-violet-600 border-violet-200 bg-violet-50/60" },
    { label: "Half Day", value: totals.half_day, cls: "text-pink-600 border-pink-200 bg-pink-50/60" },
    { label: "Holiday", value: totals.holiday, cls: "text-teal-600 border-teal-200 bg-teal-50/60" },
    { label: "OverTime", value: totals.over_time, cls: "text-indigo-600 border-indigo-200 bg-indigo-50/60" },
  ] : []

  return (
    <div className="flex flex-col gap-6 py-6">
      <div className="px-4 lg:px-6">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10">
              <BarChart3Icon className="h-6 w-6 text-primary" />
            </div>
            <div>
              <h1 className="text-3xl font-bold tracking-tight">Monthly Attendance Report</h1>
              <p className="text-muted-foreground mt-1">Per-employee monthly attendance breakdown</p>
            </div>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleExport("excel")}
              disabled={exporting !== null || !companyId || data.length === 0}
            >
              <FileSpreadsheetIcon className="mr-2 h-4 w-4 text-emerald-600" />
              {exporting === "excel" ? "Exporting..." : "Export Excel"}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleExport("pdf")}
              disabled={exporting !== null || !companyId || data.length === 0}
            >
              <FileTextIcon className="mr-2 h-4 w-4 text-red-600" />
              {exporting === "pdf" ? "Exporting..." : "Export PDF"}
            </Button>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card className="overflow-hidden">
          <div className="border-b bg-muted/40 px-5 py-3 flex items-center gap-2">
            <CalendarRangeIcon className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-semibold">Filters</span>
          </div>
          <CardContent className="pt-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Month</label>
                <select value={selectedMonth} onChange={(e) => setSelectedMonth(Number(e.target.value))} className={selectCls}>
                  {MONTHS.map((name, idx) => <option key={name} value={idx}>{name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Year</label>
                <select value={selectedYear} onChange={(e) => setSelectedYear(Number(e.target.value))} className={selectCls}>
                  {YEARS.map((y) => <option key={y} value={y}>{y}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Company</label>
                <select value={companyId} onChange={(e) => setCompanyId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {companies.map((c) => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Department</label>
                <select value={departmentId} onChange={(e) => setDepartmentId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {departments.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Section</label>
                <select value={sectionId} onChange={(e) => setSectionId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {sections.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Designation</label>
                <select value={designationId} onChange={(e) => setDesignationId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {designations.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Line</label>
                <select value={lineId} onChange={(e) => setLineId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {lines.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Shift</label>
                <select value={shiftId} onChange={(e) => setShiftId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {shifts.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Group</label>
                <select value={groupId} onChange={(e) => setGroupId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {groups.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Employee ID</label>
                <div className="relative">
                  <SearchIcon className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <input
                    value={employeeId}
                    onChange={(e) => setEmployeeId(e.target.value)}
                    placeholder="Search..."
                    className="flex h-10 w-full rounded-lg border border-input bg-background pl-9 pr-3 py-1 text-sm shadow-sm outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30"
                  />
                </div>
              </div>
            </div>
            <div className="flex gap-2 mt-5">
              <Button onClick={() => fetchData()} disabled={loading || !companyId} className="min-w-[120px]">
                {loading ? "Loading..." : "Apply Filters"}
              </Button>
              <Button variant="outline" onClick={handleReset}>Reset</Button>
            </div>
          </CardContent>
        </Card>
      </div>

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-lg bg-destructive/10 border border-destructive/20 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      {totals && (
        <div className="px-4 lg:px-6">
          <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-3">
            {statCards.map((s) => (
              <div key={s.label} className={`rounded-xl border p-3 flex flex-col items-center gap-1 ${s.cls}`}>
                <span className="text-[11px] font-medium uppercase tracking-wide opacity-80">{s.label}</span>
                <span className="text-2xl font-bold tabular-nums">{s.value}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="px-4 lg:px-6">
        <div className="flex flex-col sm:flex-row sm:items-center gap-3 mb-4">
          <div className="flex items-center gap-2 bg-muted/50 rounded-lg px-3 py-2">
            <UsersIcon className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-medium">{MONTHS[selectedMonth]} {selectedYear}</span>
            <Badge variant="secondary">{data.length} employees</Badge>
          </div>
        </div>
        <DataTable data={data} columns={columns} loading={loading && !error} enableSelection={false} enableDnd={false} />
      </div>
    </div>
  )
}