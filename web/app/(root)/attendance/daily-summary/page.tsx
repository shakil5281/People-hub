"use client"

import * as React from "react"
import { FileTextIcon, DownloadIcon, Loader2, ClipboardCheckIcon } from "lucide-react"
import { attendanceApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { Button } from "@/components/ui/button"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

interface SummaryRecord {
  id: string
  entity_id?: string
  date?: string
  name?: string
  present: number
  late: number
  absent: number
  half_day: number
  on_leave?: number
  weekend?: number
  total: number
}

const today = new Date().toISOString().split("T")[0]

const groupTabs = [
  { value: "department", label: "Department" },
  { value: "section", label: "Section" },
  { value: "designation", label: "Designation" },
  { value: "line", label: "Line" },
  { value: "group", label: "Group" },
]

function SummaryTable({ data, loading, title, activeFilters }: { data: SummaryRecord[]; loading: boolean; title?: string; activeFilters?: Record<string, string> }) {
  if (loading) {
    return (
      <div className="flex items-center justify-center py-16 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin mr-2" />
        Loading...
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-12 text-center mx-4 my-6">
        <ClipboardCheckIcon className="mb-3 h-10 w-10 text-muted-foreground/60" />
        <h3 className="text-lg font-semibold text-foreground">No attendance summary</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          No attendance data found for the selected period. Try a different date range or process data logs.
        </p>
      </div>
    )
  }

  const toOthers = (r: SummaryRecord) => (r.late || 0) + (r.half_day || 0) + (r.weekend || 0)
  const toLeave = (r: SummaryRecord) => r.on_leave || 0

  const grandTotal = data.reduce((s, r) => ({
    present: s.present + r.present,
    absent: s.absent + r.absent,
    leave: s.leave + toLeave(r),
    others: s.others + toOthers(r),
    total: s.total + r.total,
  }), { present: 0, absent: 0, leave: 0, others: 0, total: 0 })

  const displayName = (name: string | undefined) => {
    if (!name) return `Unassigned ${title || ""}`.trim()
    return name
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b bg-muted/30">
            <th className="text-left py-3 px-4 font-semibold text-muted-foreground w-10">#</th>
            <th className="text-left py-3 px-4 font-semibold text-muted-foreground">{title || "Section"}</th>
            <th className="text-center py-3 px-4 font-semibold text-green-700 bg-green-50/50">Present</th>
            <th className="text-center py-3 px-4 font-semibold text-red-700 bg-red-50/50">Absent</th>
            <th className="text-center py-3 px-4 font-semibold text-indigo-700 bg-indigo-50/50">Leave</th>
            <th className="text-center py-3 px-4 font-semibold text-orange-700 bg-orange-50/50">Others</th>
            <th className="text-center py-3 px-4 font-semibold text-muted-foreground">Total</th>
            <th className="text-left py-3 px-4 font-semibold text-muted-foreground">Remarks</th>
          </tr>
        </thead>
        <tbody>
          {data.map((row, i) => {
            const others = toOthers(row)
            const leave = toLeave(row)
            return (
              <tr key={row.id} className="border-b last:border-0 hover:bg-muted/20">
                <td className="py-2.5 px-4 text-muted-foreground text-xs">{i + 1}</td>
                <td className="py-2.5 px-4 font-medium">{displayName(row.name)}</td>
                <td className="py-2.5 px-4 text-center font-semibold text-green-700">{row.present}</td>
                <td className="py-2.5 px-4 text-center font-semibold text-red-700">{row.absent}</td>
                <td className="py-2.5 px-4 text-center font-semibold text-indigo-700">{leave || "-"}</td>
                <td className="py-2.5 px-4 text-center font-semibold text-orange-700">{others || "-"}</td>
                <td className="py-2.5 px-4 text-center font-semibold">{row.total}</td>
                <td className="py-2.5 px-4 text-muted-foreground text-xs"></td>
              </tr>
            )
          })}
        </tbody>
        <tfoot>
          <tr className="border-t-2 border-muted bg-muted/40 font-semibold">
            <td className="py-3 px-4 text-muted-foreground" colSpan={1}></td>
            <td className="py-3 px-4 text-base">Grand Total</td>
            <td className="py-3 px-4 text-center text-green-700 text-base">{grandTotal.present}</td>
            <td className="py-3 px-4 text-center text-red-700 text-base">{grandTotal.absent}</td>
            <td className="py-3 px-4 text-center text-indigo-700 text-base">{grandTotal.leave}</td>
            <td className="py-3 px-4 text-center text-orange-700 text-base">{grandTotal.others}</td>
            <td className="py-3 px-4 text-center text-base">{grandTotal.total}</td>
            <td className="py-3 px-4"></td>
          </tr>
        </tfoot>
      </table>
    </div>
  )
}

export default function DailySummaryPage() {
  const [data, setData] = React.useState<SummaryRecord[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [exporting, setExporting] = React.useState(false)
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])
  const [filters, setFilters] = React.useState<Record<string, string>>({ date: today })
  const [groupBy, setGroupBy] = React.useState("department")
  const [activeFilters, setActiveFilters] = React.useState<Record<string, string>>({ date: today })

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date", label: "Date", type: "datepicker" },
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
      { value: "on_leave", label: "On Leave" }, { value: "weekend", label: "Weekend" },
    ] },
    { key: "employee_id", label: "Employee ID", type: "text", placeholder: "Enter employee code..." },
  ], [companies, departments, sections, designations, lines, groups, shifts, filters.department_id, filters.section_id])

  const loadSections = React.useCallback(async (departmentId: string) => {
    if (!departmentId) { setSections([]); return }
    try {
      const res = await sectionApi.list(departmentId)
      setSections(Array.isArray(res.data?.data) ? res.data.data : [])
    } catch { setSections([]) }
  }, [])

  const loadDesignationsAndLines = React.useCallback(async (sectionId: string) => {
    if (!sectionId) { setDesignations([]); setLines([]); return }
    try {
      const [desRes, lRes] = await Promise.all([
        designationApi.list(sectionId),
        lineApi.list(sectionId),
      ])
      setDesignations(Array.isArray(desRes.data?.data) ? desRes.data.data : [])
      setLines(Array.isArray(lRes.data?.data) ? lRes.data.data : [])
    } catch { setDesignations([]); setLines([]) }
  }, [])

  const fetchData = React.useCallback(async (params: Record<string, string>, gb: string) => {
    setLoading(true)
    setError("")
    try {
      const apiParams: Record<string, string> = { start_date: params.date || today, end_date: params.date || today }

      if (gb) apiParams.group_by = gb
      ;["company_id", "department_id", "section_id", "designation_id", "line_id", "group_id", "shift_id", "status", "employee_id"].forEach((k) => { if (params[k]) apiParams[k] = params[k] })
      const { data: res } = await attendanceApi.summary(apiParams)
      setData((res?.summaries || []).map((r: any, i: number) => ({ ...r, id: r.entity_id || r.id || `row-${i}` })))
      setActiveFilters(params)
    } catch {
      setError("Failed to load summary")
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    const init = async () => {
      const [cRes, dRes, gRes, sRes] = await Promise.all([
        companyApi.list({ limit: "100" }), departmentApi.list({ limit: "100" }),
        groupApi.list({ limit: "100" }), shiftApi.list({ limit: "100" }),
      ])
      if (Array.isArray(cRes.data?.data)) setCompanies(cRes.data.data)
      if (Array.isArray(dRes.data?.data)) setDepartments(dRes.data.data)
      if (Array.isArray(gRes.data?.data)) setGroups(gRes.data.data)
      if (Array.isArray(sRes.data?.data)) setShifts(sRes.data.data)
    }
    init()
    fetchData({ date: today }, "department")
  }, [])

  const handleTabChange = (value: string) => {
    setGroupBy(value)
    const active: Record<string, string> = {}
    for (const [k, v] of Object.entries(filters)) { if (v) active[k] = v }
    fetchData(active, value)
  }
  const handleChange = (key: string, value: string) => {
    setFilters((prev) => {
      const next = { ...prev, [key]: value }
      if (key === "department_id") {
        delete next.section_id
        delete next.designation_id
        delete next.line_id
        loadSections(value)
        setDesignations([])
        setLines([])
      }
      if (key === "section_id") {
        delete next.designation_id
        delete next.line_id
        loadDesignationsAndLines(value)
      }
      return next
    })
  }

  const handleApply = () => {
    const active: Record<string, string> = {}
    for (const [k, v] of Object.entries(filters)) { if (v) active[k] = v }
    fetchData(active, groupBy)
  }

  const handleReset = () => {
    const rf = { date: today }
    setFilters(rf)
    fetchData(rf, groupBy)
  }

  const handleExport = async () => {
    setExporting(true)
    try {
      const active: Record<string, string> = {}
      for (const [k, v] of Object.entries(filters)) { if (v) active[k] = v }
      active.start_date = filters.date || today
      active.end_date = filters.date || today
      const res = await attendanceApi.exportSummaryExcel(active)
      const blob = new Blob([res.data], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `daily_summary_${filters.date || today}.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setError("Failed to export summary")
    } finally {
      setExporting(false)
    }
  }

  const tabLabel = groupTabs.find((t) => t.value === groupBy)?.label || ""

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <FileTextIcon className="h-6 w-6 text-muted-foreground" />
            <div>
              <h1 className="text-3xl font-bold tracking-tight">Daily Summary</h1>
              <p className="text-muted-foreground mt-1">Daily attendance summary reports</p>
            </div>
          </div>
          <Button onClick={handleExport} disabled={exporting} variant="outline">
            {exporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <DownloadIcon className="mr-2 h-4 w-4" />}
            {exporting ? "Exporting..." : "Export Excel"}
          </Button>
        </div>
      </div>

      <div className="px-4 lg:px-6">
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
        <div className="rounded-lg border bg-card overflow-hidden">
          <Tabs value={groupBy} onValueChange={handleTabChange}>
            <div className="flex items-center justify-between px-4 py-3 border-b bg-muted/20">
              <h2 className="text-base font-semibold">{tabLabel} Summary</h2>
              <TabsList className="h-8">
                {groupTabs.map((tab) => (
                  <TabsTrigger key={tab.value} value={tab.value} className="h-7 px-3 text-xs">{tab.label}</TabsTrigger>
                ))}
              </TabsList>
            </div>
            {groupTabs.map((tab) => (
              <TabsContent key={tab.value} value={tab.value} className="mt-0">
                <SummaryTable data={data} loading={loading} title={tab.label} activeFilters={activeFilters} />
              </TabsContent>
            ))}
          </Tabs>
        </div>
      </div>
    </div>
  )
}
