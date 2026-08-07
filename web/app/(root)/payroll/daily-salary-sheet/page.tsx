"use client"

import * as React from "react"
import { CalendarRangeIcon, Loader2, FilterIcon, XIcon, FileDownIcon, FileBarChartIcon } from "lucide-react"
import Link from "next/link"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { salaryApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi } from "@/lib/api"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { Card } from "@/components/ui/card"

interface DailySalaryRecord {
  id: string
  employee_id: string
  employee_name: string
  designation: string
  department_name: string
  date: string
  status: string
  check_in: string
  check_out: string
  total_hours: string
  over_time: string
  gross_salary: number
  daily_rate: number
  ot_hours: number
  ot_amount: number
  total_pay: number
}

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }

const today = new Date().toISOString().split("T")[0]

const columns: ColumnDef<DailySalaryRecord>[] = [
  { id: "sl", header: "Sl", cell: ({ row }) => row.index + 1 },
  { accessorKey: "employee_id", header: "Code" },
  { accessorKey: "employee_name", header: "Employee" },
  { accessorKey: "designation", header: "Designation" },
  { accessorKey: "department_name", header: "Department" },
  { accessorKey: "status", header: "Status" },
  { accessorKey: "check_in", header: "In" },
  { accessorKey: "check_out", header: "Out" },
  { accessorKey: "total_hours", header: "Hours" },
  { accessorKey: "over_time", header: "OT" },
  { accessorKey: "daily_rate", header: "Rate", cell: ({ row }) => `TK ${row.original.daily_rate.toFixed(0)}` },
  { accessorKey: "ot_amount", header: "OT Amt", cell: ({ row }) => `TK ${row.original.ot_amount.toFixed(0)}` },
  { accessorKey: "total_pay", header: "Total", cell: ({ row }) => `TK ${row.original.total_pay.toFixed(0)}` },
]

export default function DailySalarySheetPage() {
  const [data, setData] = React.useState<DailySalaryRecord[]>([])
  const [loading, setLoading] = React.useState(false)
  const [exporting, setExporting] = React.useState(false)
  const [totals, setTotals] = React.useState<Record<string, number>>({})
  const [filters, setFilters] = React.useState<Record<string, string>>({ date: today })
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      sectionApi.list(undefined, { limit: "100" }),
      designationApi.list(undefined, { limit: "100" }),
      lineApi.list(undefined, { limit: "100" }),
      groupApi.list({ limit: "100" }),
    ]).then(([c, d, sec, des, l, g]) => {
      if (Array.isArray(c.data?.data)) setCompanies(c.data.data)
      if (Array.isArray(d.data?.data)) setDepartments(d.data.data)
      if (Array.isArray(sec.data?.data)) setSections(sec.data.data)
      if (Array.isArray(des.data?.data)) setDesignations(des.data.data)
      if (Array.isArray(l.data?.data)) setLines(l.data.data)
      if (Array.isArray(g.data?.data)) setGroups(g.data.data)
    })
  }, [])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date", label: "Date", type: "datepicker" },
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en })) },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    { key: "section_id", label: "Section", type: "select", options: sections.map((s) => ({ value: s.id, label: s.name })) },
    { key: "designation_id", label: "Designation", type: "select", options: designations.map((d) => ({ value: d.id, label: d.name })) },
    { key: "line_id", label: "Line", type: "select", options: lines.map((l) => ({ value: l.id, label: l.name })) },
    { key: "group_id", label: "Group", type: "select", options: groups.map((g) => ({ value: g.id, label: g.name })) },
  ], [companies, departments, sections, designations, lines, groups])

  const fetchData = async (params: Record<string, string>) => {
    setLoading(true)
    try {
      const date = params.date || today
      const active: Record<string, string> = { date, company_id: params.company_id || "" }
      if (params.department_id) active.department_id = params.department_id
      if (params.section_id) active.section_id = params.section_id
      if (params.designation_id) active.designation_id = params.designation_id
      if (params.line_id) active.line_id = params.line_id
      if (params.group_id) active.group_id = params.group_id
      const { data: res } = await salaryApi.dailySheet(active)
      setData(res.records || [])
      setTotals(res.totals || {})
    } catch {
      setData([])
      setTotals({})
    } finally {
      setLoading(false)
    }
  }

  const handleApply = () => { fetchData(filters) }
  const handleReset = () => { setFilters({ date: today }); setData([]); setTotals({}) }
  const handleChange = (key: string, value: string) => setFilters((prev) => ({ ...prev, [key]: value }))

  const handleExcelExport = async () => {
    setExporting(true)
    try {
      const date = filters.date || today
      const active: Record<string, string> = { date, company_id: filters.company_id || "" }
      if (filters.department_id) active.department_id = filters.department_id
      if (filters.section_id) active.section_id = filters.section_id
      if (filters.designation_id) active.designation_id = filters.designation_id
      if (filters.line_id) active.line_id = filters.line_id
      if (filters.group_id) active.group_id = filters.group_id

      const res = await salaryApi.dailySheetExport(active)
      const url = window.URL.createObjectURL(new Blob([res.data]))
      const a = document.createElement("a")
      a.href = url
      a.download = `daily_salary_sheet_${date}.xlsx`
      a.click()
      window.URL.revokeObjectURL(url)
    } finally { setExporting(false) }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <CalendarRangeIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Daily Salary Sheet</h1>
            <p className="text-muted-foreground mt-1">Daily salary calculation sheet</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleExcelExport} disabled={exporting || data.length === 0}>
            <FileDownIcon className="mr-2 h-4 w-4" />
            {exporting ? "Exporting..." : "Excel"}
          </Button>
          <Button size="sm" asChild>
            <Link href="/payroll/daily-salary-summary">
              <FileBarChartIcon className="mr-2 h-4 w-4" />
              Summary Report
            </Link>
          </Button>
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
                    onApply={() => { handleApply(); setMobileFilterOpen(false) }}
                    onReset={() => { handleReset(); setMobileFilterOpen(false) }}
                    submitting={loading}
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
          onChange={handleChange}
          onApply={handleApply}
          onReset={handleReset}
          submitting={loading}
        />
      </div>

      {totals.employees > 0 && (
        <div className="px-4 lg:px-6">
          <Card className="p-4">
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 text-sm">
              <div><span className="text-muted-foreground">Employees:</span> <strong>{totals.employees}</strong></div>
              <div><span className="text-muted-foreground">Gross Salary:</span> <strong>TK {totals.gross_salary?.toFixed(0)}</strong></div>
              <div><span className="text-muted-foreground">Daily Rate:</span> <strong>TK {totals.daily_rate?.toFixed(0)}</strong></div>
              <div><span className="text-muted-foreground">OT Hours:</span> <strong>{totals.ot_hours?.toFixed(1)}</strong></div>
              <div><span className="text-muted-foreground">OT Amount:</span> <strong>TK {totals.ot_amount?.toFixed(0)}</strong></div>
              <div><span className="text-muted-foreground">Total Pay:</span> <strong>TK {totals.total_pay?.toFixed(0)}</strong></div>
            </div>
          </Card>
        </div>
      )}

      <DataTable data={data} columns={columns} loading={loading} enableSelection={false} />
    </div>
  )
}
