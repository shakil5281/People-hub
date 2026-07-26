"use client"

import * as React from "react"
import { ClipboardListIcon, Loader2, ChevronLeftIcon, ChevronRightIcon } from "lucide-react"
import { format } from "date-fns"
import { Button } from "@/components/ui/button"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { attendanceApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"
import { formatCheck } from "@/lib/utils"

interface JobCardRecord {
  id: string
  employee_id: string
  date: string
  check_in: string | null
  check_out: string | null
  total_hours: string | null
  over_time: string | null
  status: string
  late_minutes: number
  shift_name: string
}

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

const today = new Date().toISOString().split("T")[0]

const statusMap: Record<string, string> = {
  present: "P", late: "L", absent: "A", half_day: "H", leave: "V", on_leave: "V", weekend: "W",
}

export default function JobCardPage() {
  const [data, setData] = React.useState<JobCardRecord[]>([])
  const [employees, setEmployees] = React.useState<{employee_id: string; name_en: string; designation: string; department: string; phone: string; joining_date: string; company: string}[]>([])
  const [currentIndex, setCurrentIndex] = React.useState(0)
  const [loading, setLoading] = React.useState(false)
  const [fetchingList, setFetchingList] = React.useState(false)
  const [error, setError] = React.useState("")
  const [submitting, setSubmitting] = React.useState(false)
  const [filters, setFilters] = React.useState<Record<string, string>>({
    start_date: today,
    end_date: today,
  })
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])

  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      groupApi.list({ limit: "100" }),
      shiftApi.list({ limit: "100" }),
    ]).then(([cRes, dRes, gRes, sRes]) => {
      setCompanies(cRes.data?.data || [])
      setDepartments(dRes.data?.data || [])
      setGroups(gRes.data?.data || [])
      setShifts(sRes.data?.data || [])
    }).catch(() => {})
    handleApply()
  }, [])

  React.useEffect(() => {
    if (!filters.department_id) { setSections([]); setDesignations([]); setLines([]); return }
    sectionApi.list(filters.department_id, { limit: "100" }).then((r) => setSections(r.data?.data || [])).catch(() => setSections([]))
  }, [filters.department_id])

  React.useEffect(() => {
    if (!filters.section_id) { setDesignations([]); setLines([]); return }
    Promise.all([
      designationApi.list(filters.section_id, { limit: "100" }),
      lineApi.list(filters.section_id, { limit: "100" }),
    ]).then(([dr, lr]) => {
      setDesignations(dr.data?.data || [])
      setLines(lr.data?.data || [])
    }).catch(() => { setDesignations([]); setLines([]) })
  }, [filters.section_id])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date_range", label: "Date Range", type: "daterange-split", dateRangeKeys: { start: "start_date", end: "end_date" } },
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en })) },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    { key: "section_id", label: "Section", type: "select", options: sections.map((s) => ({ value: s.id, label: s.name })), disabled: !filters.department_id },
    { key: "designation_id", label: "Designation", type: "select", options: designations.map((d) => ({ value: d.id, label: d.name })), disabled: !filters.section_id },
    { key: "line_id", label: "Line", type: "select", options: lines.map((l) => ({ value: l.id, label: l.name })), disabled: !filters.section_id },
    { key: "group_id", label: "Group", type: "select", options: groups.map((g) => ({ value: g.id, label: g.name })) },
    { key: "shift_id", label: "Shift", type: "select", options: shifts.map((s) => ({ value: s.id, label: s.name })) },
    { key: "status", label: "Status", type: "select", options: [
      { value: "present", label: "Present" }, { value: "late", label: "Late" },
      { value: "absent", label: "Absent" }, { value: "half_day", label: "Half Day" },
    ] },
    { key: "employee_id", label: "Employee ID", type: "text", placeholder: "Enter employee code..." },
  ], [companies, departments, sections, designations, lines, groups, shifts, filters.department_id, filters.section_id])

  const buildParams = (f?: Record<string, string>) => {
    const params = f || filters
    const active: Record<string, string> = {
      start_date: params.start_date || today,
      end_date: params.end_date || today,
    }
    if (params.company_id) active.company_id = params.company_id
    if (params.department_id) active.department_id = params.department_id
    if (params.section_id) active.section_id = params.section_id
    if (params.designation_id) active.designation_id = params.designation_id
    if (params.line_id) active.line_id = params.line_id
    if (params.group_id) active.group_id = params.group_id
    if (params.shift_id) active.shift_id = params.shift_id
    if (params.status) active.status = params.status
    if (params.employee_id) active.employee_id = params.employee_id
    return active
  }

  const fetchEmployeeList = async (f?: Record<string, string>) => {
    setFetchingList(true)
    try {
      const params = buildParams(f || filters)
      params.list_mode = "true"
      const { data: res } = await attendanceApi.jobCard(params)
      setEmployees(res.data || [])
      setCurrentIndex(0)
      return res.data || []
    } catch {
      setError("Failed to load employee list")
      return []
    } finally {
      setFetchingList(false)
    }
  }

  const fetchEmployeeData = async (empId: string) => {
    setLoading(true)
    setError("")
    try {
      const params = buildParams()
      params.employee_id = empId
      delete params.list_mode
      const { data: res } = await attendanceApi.jobCard(params)
      setData(res.data || [])
    } catch {
      setError("Failed to load attendance data")
    } finally {
      setLoading(false)
    }
  }

  const handleApply = async () => {
    setSubmitting(true)
    const empList = await fetchEmployeeList()
    setSubmitting(false)
    if (empList.length > 0) {
      fetchEmployeeData(empList[0].employee_id)
    } else {
      setData([])
    }
  }

  const handleReset = () => {
    setFilters({ start_date: today, end_date: today })
    setData([])
    setEmployees([])
    setCurrentIndex(0)
    setError("")
  }

  const handlePrev = () => {
    if (currentIndex > 0) {
      const newIdx = currentIndex - 1
      setCurrentIndex(newIdx)
      fetchEmployeeData(employees[newIdx].employee_id)
    }
  }

  const handleNext = () => {
    if (currentIndex < employees.length - 1) {
      const newIdx = currentIndex + 1
      setCurrentIndex(newIdx)
      fetchEmployeeData(employees[newIdx].employee_id)
    }
  }

  const handleFilterChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const currentEmpId = employees[currentIndex]?.employee_id || ""
  const emp = employees[currentIndex] || null
  const totalByStatus = data.reduce<Record<string, number>>((acc, r) => {
    acc[r.status] = (acc[r.status] || 0) + 1
    return acc
  }, {})
  const totalLateMinutes = data.reduce((sum, r) => sum + (r.late_minutes || 0), 0)

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <ClipboardListIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Job Card</h1>
            <p className="text-muted-foreground mt-1">Employee attendance job card report</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleFilterChange}
          onApply={handleApply}
          onReset={handleReset}
          submitting={submitting}
        />
      </div>

      <div className="px-4 lg:px-6">
        <div className="rounded-lg border bg-card overflow-hidden">
          {employees.length > 0 && (
            <div className="border-b px-4 py-3 flex items-center justify-between">
              <Button variant="outline" size="sm" onClick={handlePrev} disabled={currentIndex === 0 || loading}>
                <ChevronLeftIcon className="h-4 w-4 mr-1" />Previous
              </Button>
              <span className="text-sm text-muted-foreground">Employee {currentIndex + 1} of {employees.length}</span>
              <Button variant="outline" size="sm" onClick={handleNext} disabled={currentIndex >= employees.length - 1 || loading}>
                Next<ChevronRightIcon className="h-4 w-4 ml-1" />
              </Button>
            </div>
          )}

          {emp && (
            <div className="border-b bg-muted/30 px-4 py-3">
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-sm">
                <div><span className="text-muted-foreground">Name: </span><span className="font-medium">{emp.name_en}</span></div>
                <div><span className="text-muted-foreground">Code: </span><span className="font-medium">{emp.employee_id}</span></div>
                <div><span className="text-muted-foreground">Designation: </span><span className="font-medium">{emp.designation || "-"}</span></div>
                <div><span className="text-muted-foreground">Department: </span><span className="font-medium">{emp.department || "-"}</span></div>
                <div><span className="text-muted-foreground">Phone: </span><span className="font-medium">{emp.phone || "-"}</span></div>
                <div><span className="text-muted-foreground">Joining: </span><span className="font-medium">{emp.joining_date ? format(new Date(emp.joining_date), "dd MMM yyyy") : "-"}</span></div>
                <div><span className="text-muted-foreground">Period: </span><span className="font-medium">
                  {filters.start_date ? `${filters.start_date.split("-").reverse().join("-")} - ${(filters.end_date || filters.start_date).split("-").reverse().join("-")}` : "-"}
                </span></div>
                <div><span className="text-muted-foreground">Days: </span><span className="font-medium">{data.length}</span></div>
              </div>
            </div>
          )}

          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="px-3 py-2.5 text-left font-semibold w-10">#</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Date</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Shift</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Day</th>
                  <th className="px-3 py-2.5 text-left font-semibold">In Time</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Out Time</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Hours</th>
                  <th className="px-3 py-2.5 text-left font-semibold">OT</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Late (min)</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Status</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr><td colSpan={10} className="px-3 py-12 text-center"><Loader2 className="h-6 w-6 animate-spin mx-auto text-muted-foreground" /></td></tr>
                ) : error ? (
                  <tr><td colSpan={10} className="px-3 py-8 text-center text-destructive">{error}</td></tr>
                ) : data.length === 0 ? (
                  <tr><td colSpan={10} className="px-3 py-8 text-center text-muted-foreground">No records found</td></tr>
                ) : (
                  data.map((row, i) => {
                    const bg = row.status === "present" ? "bg-green-50/40"
                      : row.status === "late" ? "bg-orange-50/40"
                      : row.status === "absent" ? "bg-red-50/40"
                      : row.status === "half_day" ? "bg-yellow-50/40"
                      : row.status === "on_leave" ? "bg-indigo-50/40"
                      : row.status === "weekend" ? "bg-slate-50/40" : ""
                    const tc = row.status === "absent" ? "text-red-600"
                      : row.status === "present" ? "text-green-700"
                      : row.status === "late" ? "text-orange-700"
                      : row.status === "half_day" ? "text-yellow-700"
                      : row.status === "on_leave" ? "text-indigo-700"
                      : "text-muted-foreground"
                    return (
                      <tr key={row.id} className={`border-b last:border-0 ${bg} hover:bg-muted/50 transition-colors`}>
                        <td className="px-3 py-2">{i + 1}</td>
                        <td className="px-3 py-2">{format(new Date(row.date), "dd MMM yyyy")}</td>
                        <td className="px-3 py-2">{row.shift_name || "-"}</td>
                        <td className="px-3 py-2">{format(new Date(row.date), "EEE")}</td>
                        <td className="px-3 py-2">{formatCheck(row.check_in)}</td>
                        <td className="px-3 py-2">{formatCheck(row.check_out)}</td>
                        <td className="px-3 py-2">{row.total_hours || "-"}</td>
                        <td className="px-3 py-2">{row.over_time || "-"}</td>
                        <td className="px-3 py-2">{row.late_minutes > 0 ? row.late_minutes : "-"}</td>
                        <td className={`px-3 py-2 font-semibold ${tc}`}>
                          {statusMap[row.status] || row.status}
                        </td>
                      </tr>
                    )
                  })
                )}
              </tbody>
            </table>
          </div>

          {data.length > 0 && (
            <div className="border-t px-4 py-3">
              <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm">
                <span className="text-muted-foreground">Total Days: <b>{data.length}</b></span>
                {Object.entries(totalByStatus).map(([s, c]) => (
                  <span key={s} className="text-muted-foreground">
                    {s === "on_leave" ? "Leave" : s.charAt(0).toUpperCase() + s.slice(1)}: <b>{c}</b>
                  </span>
                ))}
                <span className="text-muted-foreground">Late Minutes: <b>{totalLateMinutes}</b></span>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
