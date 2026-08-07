"use client"

import * as React from "react"
import { FileBarChartIcon, Loader2, FileDownIcon, CalendarRangeIcon, ArrowLeftIcon } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { salaryApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi } from "@/lib/api"
import { toast } from "sonner"
import Link from "next/link"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }

interface DailySummaryRow {
  id: string
  group_key: string
  employees: number
  gross_salary: number
  daily_rate: number
  ot_hours: number
  ot_amount: number
  total_pay: number
}

interface DailySummaryResponse {
  summaries: DailySummaryRow[]
  total: number
  totals: Record<string, number>
  date: string
}

const today = new Date().toISOString().split("T")[0]

const fmt = (n: number) => Math.round(n || 0).toLocaleString()

const TABS = [
  { value: "department", label: "Department" },
  { value: "section", label: "Section" },
  { value: "designation", label: "Designation" },
  { value: "line", label: "Line" },
  { value: "custom", label: "Custom Summary" },
]

export default function DailySalarySummaryPage() {
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])

  const [companyId, setCompanyId] = React.useState("")
  const [departmentId, setDepartmentId] = React.useState("")
  const [sectionId, setSectionId] = React.useState("")
  const [designationId, setDesignationId] = React.useState("")
  const [lineId, setLineId] = React.useState("")
  const [groupId, setGroupId] = React.useState("")
  const [date, setDate] = React.useState(today)
  const [tab, setTab] = React.useState("department")

  const [data, setData] = React.useState<DailySummaryResponse | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [exporting, setExporting] = React.useState(false)
  const [lang, setLang] = React.useState<"en" | "bn">("en")

  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      sectionApi.list(undefined, { limit: "100" }),
      designationApi.list(undefined, { limit: "100" }),
      lineApi.list(undefined, { limit: "100" }),
      groupApi.list({ limit: "100" }),
    ]).then(([cRes, dRes, secRes, desRes, lRes, gRes]) => {
      const clist = Array.isArray(cRes.data?.data) ? cRes.data.data : (Array.isArray(cRes.data) ? cRes.data : [])
      if (clist.length > 0) { setCompanies(clist); setCompanyId(clist[0].id) }
      if (Array.isArray(dRes.data?.data)) setDepartments(dRes.data.data)
      if (Array.isArray(secRes.data?.data)) setSections(secRes.data.data)
      if (Array.isArray(desRes.data?.data)) setDesignations(desRes.data.data)
      if (Array.isArray(lRes.data?.data)) setLines(lRes.data.data)
      if (Array.isArray(gRes.data?.data)) setGroups(gRes.data.data)
    }).catch(() => {})
  }, [])

  const buildParams = React.useCallback(() => {
    const p: Record<string, string> = {
      date,
      company_id: companyId,
      group_by: tab,
      lang,
    }
    if (departmentId) p.department_id = departmentId
    if (sectionId) p.section_id = sectionId
    if (designationId) p.designation_id = designationId
    if (lineId) p.line_id = lineId
    if (groupId) p.group_id = groupId
    return p
  }, [companyId, date, tab, lang, departmentId, sectionId, designationId, lineId, groupId])

  const handleLoad = React.useCallback(async () => {
    if (!date) { toast.error("Select a date"); return }
    setLoading(true)
    try {
      const { data: res } = await salaryApi.dailySummary(buildParams())
      const rows = Array.isArray(res.summaries) ? res.summaries.map((s: DailySummaryRow, i: number) => ({ ...s, id: `ds-${i}` })) : []
      setData({ ...res, summaries: rows })
    } catch { toast.error("Failed to load daily summary") }
    finally { setLoading(false) }
  }, [buildParams, date])

  React.useEffect(() => { if (companyId && date) handleLoad() }, [companyId, date, tab])

  const columns: ColumnDef<DailySummaryRow>[] = React.useMemo(() => {
    const labelMap: Record<string, string> = { department: "Department", section: "Section", designation: "Designation", line: "Line", custom: "Category / Line" }
    return [
      { accessorKey: "group_key", header: labelMap[tab] || "Group" },
      { accessorKey: "employees", header: "Employees" },
      { accessorKey: "gross_salary", header: "Gross Total", cell: ({ row }) => fmt(row.original.gross_salary) },
      { accessorKey: "daily_rate", header: "Daily Rate Total", cell: ({ row }) => fmt(row.original.daily_rate) },
      { accessorKey: "ot_hours", header: "OT Hours", cell: ({ row }) => (row.original.ot_hours || 0).toFixed(1) },
      { accessorKey: "ot_amount", header: "OT Amount", cell: ({ row }) => fmt(row.original.ot_amount) },
      { accessorKey: "total_pay", header: "Total Pay", cell: ({ row }) => fmt(row.original.total_pay) },
    ]
  }, [tab])

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <FileBarChartIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Daily Salary Summary</h1>
            <p className="text-muted-foreground mt-1">Daily salary summary grouped by department, section, designation, or line</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" asChild>
            <Link href="/payroll/daily-salary-sheet">
              <ArrowLeftIcon className="mr-2 h-4 w-4" />
              Back to Daily Salary Sheet
            </Link>
          </Button>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader className="pb-3"><CardTitle className="text-base">Filters</CardTitle></CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">Date</label>
                <input
                  type="date"
                  value={date}
                  onChange={e => setDate(e.target.value)}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">Company</label>
                <select value={companyId} onChange={e => setCompanyId(e.target.value)} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                  <option value="">Select</option>
                  {companies.map(c => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">Department</label>
                <select value={departmentId} onChange={e => setDepartmentId(e.target.value)} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                  <option value="">All</option>
                  {departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">Section</label>
                <select value={sectionId} onChange={e => setSectionId(e.target.value)} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                  <option value="">All</option>
                  {sections.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">Designation</label>
                <select value={designationId} onChange={e => setDesignationId(e.target.value)} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                  <option value="">All</option>
                  {designations.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">Line</label>
                <select value={lineId} onChange={e => setLineId(e.target.value)} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                  <option value="">All</option>
                  {lines.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">Group</label>
                <select value={groupId} onChange={e => setGroupId(e.target.value)} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                  <option value="">All</option>
                  {groups.map(g => <option key={g.id} value={g.id}>{g.name}</option>)}
                </select>
              </div>
              <div className="lg:col-span-4 flex justify-end">
                <Button onClick={handleLoad} disabled={loading}>
                  {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileBarChartIcon className="mr-2 h-4 w-4" />}
                  Load Summary
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {data ? (
        <div className="px-4 lg:px-6">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-lg font-semibold">Daily Salary Summary ({data.date})</h2>
            <div className="flex gap-2 items-center">
              <span className="text-sm text-muted-foreground">Total Employees: <b>{data.totals.employees}</b></span>
              <div className="flex gap-2">
                <div className="flex border rounded-md overflow-hidden">
                  <button
                    className={`px-2 py-1 text-xs font-medium ${lang === "en" ? "bg-primary text-primary-foreground" : "bg-muted"}`}
                    onClick={() => setLang("en")}
                  >
                    EN
                  </button>
                  <button
                    className={`px-2 py-1 text-xs font-medium ${lang === "bn" ? "bg-primary text-primary-foreground" : "bg-muted"}`}
                    onClick={() => setLang("bn")}
                  >
                    BN
                  </button>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={exporting}
                  onClick={async () => {
                    setExporting(true)
                    try {
                      const res = await salaryApi.dailySummaryExport(buildParams())
                      const url = window.URL.createObjectURL(new Blob([res.data]))
                      const a = document.createElement("a")
                      a.href = url
                      a.download = `daily_salary_summary_${date}_${lang}.xlsx`
                      a.click()
                      window.URL.revokeObjectURL(url)
                    } finally { setExporting(false) }
                  }}
                >
                  <FileDownIcon className="mr-2 h-4 w-4" />
                  {exporting ? "Exporting..." : "Excel"}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={exporting}
                  onClick={async () => {
                    setExporting(true)
                    try {
                      const res = await salaryApi.dailySummaryExportPdf(buildParams())
                      const url = window.URL.createObjectURL(new Blob([res.data]))
                      const a = document.createElement("a")
                      a.href = url
                      a.download = `daily_salary_summary_${date}_${lang}.pdf`
                      a.click()
                      window.URL.revokeObjectURL(url)
                    } finally { setExporting(false) }
                  }}
                >
                  <FileDownIcon className="mr-2 h-4 w-4" />
                  {exporting ? "Exporting..." : "PDF"}
                </Button>
              </div>
            </div>
          </div>

          <Tabs value={tab} onValueChange={v => { setTab(v); setData(null) }}>
            <TabsList className="mb-4">
              {TABS.map(t => <TabsTrigger key={t.value} value={t.value}>{t.label}</TabsTrigger>)}
            </TabsList>

            {TABS.map(t => (
              <TabsContent key={t.value} value={t.value}>
                <DataTable
                  data={tab === t.value ? data.summaries : []}
                  columns={columns}
                  loading={loading}
                  enableSelection={false}
                />
              </TabsContent>
            ))}
          </Tabs>

          <div className="mt-4 rounded-md border bg-muted/30 p-4">
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 text-sm">
              <div><span className="text-muted-foreground">Total Employees</span><p className="font-semibold">{data.totals.employees}</p></div>
              <div><span className="text-muted-foreground">Gross Salary</span><p className="font-semibold">{fmt(data.totals.gross_salary)}</p></div>
              <div><span className="text-muted-foreground">Daily Rate</span><p className="font-semibold">{fmt(data.totals.daily_rate)}</p></div>
              <div><span className="text-muted-foreground">OT Hours</span><p className="font-semibold">{(data.totals.ot_hours || 0).toFixed(1)}</p></div>
              <div><span className="text-muted-foreground">OT Amount</span><p className="font-semibold">{fmt(data.totals.ot_amount)}</p></div>
              <div><span className="text-muted-foreground">Total Pay</span><p className="font-semibold">{fmt(data.totals.total_pay)}</p></div>
            </div>
          </div>
        </div>
      ) : (
        <div className="px-4 lg:px-6 text-center text-muted-foreground py-12">Select filters and click Load Summary to view daily salary summary</div>
      )}
    </div>
  )
}
