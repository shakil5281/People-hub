"use client"

import * as React from "react"
import { ClockIcon, ChevronDownIcon, ChevronUpIcon, Loader2, FilterIcon, XIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, employeeApi } from "@/lib/api"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card } from "@/components/ui/card"
import { differenceInYears, differenceInMonths, differenceInDays, parseISO } from "date-fns"

interface JobAgeRecord {
  id: string
  employee_id: string
  name_en: string
  designation: string
  department: string
  joining_date: string
  job_years: number
  job_months: number
  job_days: number
  status: string
}

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }

const columns: ColumnDef<JobAgeRecord>[] = [
  { id: "sl", header: "Sl", cell: ({ row }) => row.index + 1 },
  { accessorKey: "employee_id", header: "Code" },
  { accessorKey: "name_en", header: "Name" },
  { accessorKey: "designation", header: "Designation" },
  { accessorKey: "department", header: "Department" },
  {
    accessorKey: "joining_date",
    header: "Joining Date",
    cell: ({ row }) => {
      if (!row.original.joining_date) return "-"
      try {
        return parseISO(row.original.joining_date).toLocaleDateString("en-GB", { day: "2-digit", month: "short", year: "numeric" })
      } catch { return row.original.joining_date }
    },
  },
  {
    id: "job_age",
    header: "Job Age",
    cell: ({ row }) => {
      const r = row.original
      if (r.job_years === 0 && r.job_months === 0) return `${r.job_days} days`
      return `${r.job_years}y ${r.job_months}m ${r.job_days}d`
    },
  },
  { accessorKey: "status", header: "Status" },
]

function calcJobAge(joiningDate: string): { years: number; months: number; days: number } {
  if (!joiningDate) return { years: 0, months: 0, days: 0 }
  try {
    const join = parseISO(joiningDate)
    const now = new Date()
    const years = differenceInYears(now, join)
    const months = differenceInMonths(now, join) % 12
    const days = differenceInDays(now, join) % 30
    return { years, months, days }
  } catch {
    return { years: 0, months: 0, days: 0 }
  }
}

export default function JobAgePage() {
  const [data, setData] = React.useState<JobAgeRecord[]>([])
  const [filteredData, setFilteredData] = React.useState<JobAgeRecord[]>([])
  const [loading, setLoading] = React.useState(false)
  const [showAdvance, setShowAdvance] = React.useState(false)
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)
  const [minAge, setMinAge] = React.useState("")
  const [maxAge, setMaxAge] = React.useState("")
  const [filters, setFilters] = React.useState<Record<string, string>>({})
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])

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
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en })) },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    { key: "section_id", label: "Section", type: "select", options: sections.map((s) => ({ value: s.id, label: s.name })) },
    { key: "designation_id", label: "Designation", type: "select", options: designations.map((d) => ({ value: d.id, label: d.name })) },
    { key: "line_id", label: "Line", type: "select", options: lines.map((l) => ({ value: l.id, label: l.name })) },
    { key: "group_id", label: "Group", type: "select", options: groups.map((g) => ({ value: g.id, label: g.name })) },
  ], [companies, departments, sections, designations, lines, groups])

  const applyAdvanceFilters = React.useCallback((employees: JobAgeRecord[]) => {
    let result = employees
    const min = parseInt(minAge)
    const max = parseInt(maxAge)
    if (!isNaN(min) && min > 0) result = result.filter((e) => e.job_years >= min)
    if (!isNaN(max) && max > 0) result = result.filter((e) => e.job_years <= max)
    return result
  }, [minAge, maxAge])

  const fetchData = async (params: Record<string, string>) => {
    setLoading(true)
    try {
      const active: Record<string, string> = { limit: "1000" }
      if (params.company_id) active.company_id = params.company_id
      if (params.department_id) active.department_id = params.department_id
      if (params.section_id) active.section_id = params.section_id
      if (params.designation_id) active.designation_id = params.designation_id
      if (params.line_id) active.line_id = params.line_id
      if (params.group_id) active.group_id = params.group_id
      const { data: res } = await employeeApi.list(active)
      const employees = res.data?.data || []
      const mapped: JobAgeRecord[] = employees.map((e: any) => {
        const age = calcJobAge(e.joining_date)
        return {
          id: e.id,
          employee_id: e.employee_id,
          name_en: e.name_en || "",
          designation: e.designation_ref?.name || "",
          department: e.department?.name || "",
          joining_date: e.joining_date || "",
          job_years: age.years,
          job_months: age.months,
          job_days: age.days,
          status: e.status || "",
        }
      })
      setData(mapped)
      setFilteredData(applyAdvanceFilters(mapped))
    } catch {
      setData([])
      setFilteredData([])
    } finally {
      setLoading(false)
    }
  }

  const handleApply = () => { fetchData(filters) }
  const handleReset = () => { setFilters({}); setData([]); setFilteredData([]) }
  const handleChange = (key: string, value: string) => setFilters((prev) => ({ ...prev, [key]: value }))

  React.useEffect(() => {
    setFilteredData(applyAdvanceFilters(data))
  }, [minAge, maxAge, data, applyAdvanceFilters])

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <ClockIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Job Age</h1>
            <p className="text-muted-foreground mt-1">Employee length of service / job age report</p>
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

      <div className="px-4 lg:px-6">
        <Card className="overflow-hidden">
          <button
            type="button"
            onClick={() => setShowAdvance(!showAdvance)}
            className="flex items-center justify-between w-full px-4 py-3 text-sm font-medium bg-muted/20 hover:bg-muted/40 transition-colors"
          >
            <span>Advance Filters</span>
            {showAdvance ? <ChevronUpIcon className="h-4 w-4" /> : <ChevronDownIcon className="h-4 w-4" />}
          </button>
          {showAdvance && (
            <div className="px-4 py-3 border-t">
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <div className="space-y-1.5">
                  <Label htmlFor="min-age">Min Job Age (years)</Label>
                  <Input id="min-age" type="number" min="0" placeholder="e.g. 1" value={minAge} onChange={(e) => setMinAge(e.target.value)} />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="max-age">Max Job Age (years)</Label>
                  <Input id="max-age" type="number" min="0" placeholder="e.g. 10" value={maxAge} onChange={(e) => setMaxAge(e.target.value)} />
                </div>
              </div>
              <p className="text-xs text-muted-foreground mt-2">Filters are applied automatically to the loaded data.</p>
            </div>
          )}
        </Card>
      </div>

      <DataTable data={filteredData} columns={columns} loading={loading} enableSelection={false} />
    </div>
  )
}
