"use client"

import * as React from "react"
import { BarChart3Icon, SearchIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { attendanceApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, shiftApi, groupApi } from "@/lib/api"
import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"

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

const columns: ColumnDef<MonthlyRecord>[] = [
  { id: "sl", header: "Sl", cell: ({ row }) => row.index + 1 },
  { accessorKey: "emp_id", header: "Emp. ID" },
  { accessorKey: "employee_name", header: "Name" },
  { accessorKey: "designation_name", header: "Designation" },
  { accessorKey: "department_name", header: "Department" },
  { accessorKey: "shift_name", header: "Shift" },
  { accessorKey: "present", header: "Present" },
  { accessorKey: "absent", header: "Absent" },
  { accessorKey: "late", header: "Late" },
  { accessorKey: "leave", header: "Leave" },
  { accessorKey: "weekend", header: "Weekend" },
  { accessorKey: "half_day", header: "Half Day" },
  { accessorKey: "holiday", header: "Holiday" },
  { accessorKey: "over_time", header: "OverTime" },
  { accessorKey: "total_days", header: "Total" },
]

export default function MonthlyAttendancePage() {
  const [data, setData] = React.useState<MonthlyRecord[]>([])
  const [totals, setTotals] = React.useState<Totals | null>(null)
  const [loading, setLoading] = React.useState(true)
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

  const fetchData = React.useCallback(async () => {
    if (!companyId) return
    setLoading(true)
    setError("")
    try {
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

      const { data: res } = await attendanceApi.monthlyReport(params)
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
  }, [selectedYear, selectedMonth, companyId, departmentId, sectionId, designationId, lineId, shiftId, groupId, employeeId])

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
  }, [selectedYear, selectedMonth, companyId, fetchData])

  const handleApply = () => {
    fetchData()
  }

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

  const selectCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
  const labelCls = "text-xs font-medium text-muted-foreground"

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <BarChart3Icon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Monthly Attendance Report</h1>
            <p className="text-muted-foreground mt-1">Per-employee monthly attendance breakdown</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
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
                  <SearchIcon className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <input
                    value={employeeId}
                    onChange={(e) => setEmployeeId(e.target.value)}
                    placeholder="Search..."
                    className="flex h-9 w-full rounded-md border border-input bg-transparent pl-8 pr-3 py-1 text-sm"
                  />
                </div>
              </div>
            </div>
            <div className="flex gap-2 mt-4">
              <Button onClick={handleApply} disabled={loading || !companyId}>
                {loading ? "Loading..." : "Apply Filters"}
              </Button>
              <Button variant="outline" onClick={handleReset}>Reset</Button>
            </div>
          </CardContent>
        </Card>
      </div>

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      <div className="px-4 lg:px-6">
        <h2 className="text-lg font-semibold mb-2">
          Report for {MONTHS[selectedMonth]} {selectedYear}
        </h2>
        {totals && (
          <div className="flex flex-wrap gap-4 text-sm text-muted-foreground mb-2">
            <span>Present: <strong className="text-foreground">{totals.present}</strong></span>
            <span>Absent: <strong className="text-foreground">{totals.absent}</strong></span>
            <span>Late: <strong className="text-foreground">{totals.late}</strong></span>
            <span>Leave: <strong className="text-foreground">{totals.leave}</strong></span>
            <span>Weekend: <strong className="text-foreground">{totals.weekend}</strong></span>
            <span>Half Day: <strong className="text-foreground">{totals.half_day}</strong></span>
            <span>Holiday: <strong className="text-foreground">{totals.holiday}</strong></span>
            <span>OverTime: <strong className="text-foreground">{totals.over_time}</strong></span>
          </div>
        )}
      </div>

      <DataTable data={data} columns={columns} loading={loading && !error} />
    </div>
  )
}
