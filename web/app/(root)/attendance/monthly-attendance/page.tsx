"use client"

import * as React from "react"
import { BarChart3Icon, FilterIcon, XIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { attendanceApi, companyApi, departmentApi } from "@/lib/api"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }

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
  total_days: number
}

interface Totals {
  present: number
  absent: number
  late: number
  leave: number
  weekend: number
  half_day: number
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
  { accessorKey: "total_days", header: "Total" },
]

export default function MonthlyAttendancePage() {
  const [data, setData] = React.useState<MonthlyRecord[]>([])
  const [totals, setTotals] = React.useState<Totals | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [selectedMonth, setSelectedMonth] = React.useState(currentMonth)
  const [selectedYear, setSelectedYear] = React.useState(currentYear)
  const [filters, setFilters] = React.useState<Record<string, string>>({})
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  const filterDefs: FilterDef[] = React.useMemo(() => [
    {
      key: "company_id", label: "Company", type: "select",
      options: companies.map((c) => ({ value: c.id, label: c.company_name_en })),
    },
    {
      key: "department_id", label: "Department", type: "select",
      options: departments.map((d) => ({ value: d.id, label: d.name })),
    },
  ], [companies, departments])

  const fetchData = React.useCallback(async (params: Record<string, string>) => {
    setLoading(true)
    setError("")
    try {
      const apiParams: Record<string, string> = {
        year: String(selectedYear),
        month: String(selectedMonth + 1),
        company_id: params.company_id || "",
        department_id: params.department_id || "",
      }
      const { data: res } = await attendanceApi.monthlyReport(apiParams)
      const records = (res?.records || []).map((r: any, i: number) => ({
        ...r,
        id: r.employee_id || `emp-${i}`,
      }))
      setData(records)
      setTotals(res?.totals || null)
    } catch {
      setError("Failed to load monthly attendance report")
    } finally {
      setLoading(false)
    }
  }, [selectedYear, selectedMonth])

  React.useEffect(() => {
    const init = async () => {
      const [cRes, dRes] = await Promise.all([
        companyApi.list({ limit: "100" }),
        departmentApi.list({ limit: "100" }),
      ])
      let companyId = ""
      if (Array.isArray(cRes.data?.data) && cRes.data.data.length > 0) {
        setCompanies(cRes.data.data)
        companyId = cRes.data.data[0].id
      }
      if (Array.isArray(dRes.data?.data)) setDepartments(dRes.data.data)
      setFilters({ company_id: companyId })
      if (companyId) fetchData({ company_id: companyId })
      else setLoading(false)
    }
    init()
  }, [])

  React.useEffect(() => {
    if (filters.company_id) fetchData(filters)
  }, [selectedYear, selectedMonth])

  const handleChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleApply = () => {
    fetchData(filters)
  }

  const handleReset = () => {
    setFilters({})
    fetchData({})
  }

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
                <div className="flex-1 overflow-y-auto px-4 py-4 space-y-4">
                  <div className="space-y-3">
                    <div className="flex flex-col gap-1.5">
                      <label className="text-xs font-medium text-muted-foreground">Month</label>
                      <select
                        value={selectedMonth}
                        onChange={(e) => { setSelectedMonth(Number(e.target.value)) }}
                        className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
                      >
                        {MONTHS.map((name, idx) => (
                          <option key={name} value={idx}>{name}</option>
                        ))}
                      </select>
                    </div>
                    <div className="flex flex-col gap-1.5">
                      <label className="text-xs font-medium text-muted-foreground">Year</label>
                      <select
                        value={selectedYear}
                        onChange={(e) => { setSelectedYear(Number(e.target.value)) }}
                        className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
                      >
                        {YEARS.map((y) => (
                          <option key={y} value={y}>{y}</option>
                        ))}
                      </select>
                    </div>
                  </div>
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
        <div className="rounded-lg border bg-card p-4">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Month</label>
              <select
                value={selectedMonth}
                onChange={(e) => { setSelectedMonth(Number(e.target.value)) }}
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
              >
                {MONTHS.map((name, idx) => (
                  <option key={name} value={idx}>{name}</option>
                ))}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Year</label>
              <select
                value={selectedYear}
                onChange={(e) => { setSelectedYear(Number(e.target.value)) }}
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
              >
                {YEARS.map((y) => (
                  <option key={y} value={y}>{y}</option>
                ))}
              </select>
            </div>
          </div>
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
          </div>
        )}
      </div>

      <DataTable data={data} columns={columns} loading={loading} />
    </div>
  )
}
