"use client"

import * as React from "react"
import { ClockIcon, ChevronDownIcon, ChevronUpIcon, Loader2, FilterIcon, XIcon, SearchIcon, RotateCcwIcon, UsersIcon, CalendarIcon, AwardIcon } from "lucide-react"
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
import { Badge } from "@/components/ui/badge"
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
  {
    accessorKey: "employee_id",
    header: "Emp. ID",
    cell: ({ row }) => <span className="font-mono font-medium">{row.original.employee_id}</span>,
  },
  { accessorKey: "name_en", header: "Employee Name" },
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
    header: "Job Age (Tenure)",
    cell: ({ row }) => {
      const r = row.original
      if (r.job_years === 0 && r.job_months === 0) return <span className="text-muted-foreground">{r.job_days} days</span>
      return (
        <span className="font-medium text-foreground">
          {r.job_years > 0 ? `${r.job_years}y ` : ""}
          {r.job_months > 0 ? `${r.job_months}m ` : ""}
          {r.job_days > 0 && r.job_years === 0 ? `${r.job_days}d` : ""}
        </span>
      )
    },
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const st = row.original.status?.toLowerCase()
      const isAct = st === "active"
      return (
        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold ${isAct ? "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400" : "bg-destructive/15 text-destructive"}`}>
          {row.original.status || "-"}
        </span>
      )
    },
  },
]

function calcJobAge(joiningDate: string): { years: number; months: number; days: number } {
  if (!joiningDate) return { years: 0, months: 0, days: 0 }
  try {
    const join = parseISO(joiningDate)
    const now = new Date()
    const years = differenceInYears(now, join)
    const months = differenceInMonths(now, join) % 12
    const days = differenceInDays(now, join) % 30
    return { years: Math.max(0, years), months: Math.max(0, months), days: Math.max(0, days) }
  } catch {
    return { years: 0, months: 0, days: 0 }
  }
}

export default function JobAgePage() {
  const [data, setData] = React.useState<JobAgeRecord[]>([])
  const [filteredData, setFilteredData] = React.useState<JobAgeRecord[]>([])
  const [loading, setLoading] = React.useState(false)
  const [page, setPage] = React.useState(1)
  const [limit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)
  const [showAdvance, setShowAdvance] = React.useState(false)
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  // Advance filter state
  const [minAge, setMinAge] = React.useState("")
  const [maxAge, setMaxAge] = React.useState("")
  const [searchEmp, setSearchEmp] = React.useState("")
  const [activePreset, setActivePreset] = React.useState<string>("all")

  // Primary filter state
  const [filters, setFilters] = React.useState<Record<string, string>>({})
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])

  // Initial load for Top-level entities
  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      groupApi.list({ limit: "100" }),
    ]).then(([c, d, g]) => {
      if (Array.isArray(c.data?.data)) setCompanies(c.data.data)
      if (Array.isArray(d.data?.data)) setDepartments(d.data.data)
      if (Array.isArray(g.data?.data)) setGroups(g.data.data)
      fetchData({})
    }).catch(() => {
      fetchData({})
    })
  }, [])

  // Relational cascading: Department -> Section
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
    sectionApi.list(filters.department_id, { limit: "100" })
      .then((r) => setSections(Array.isArray(r.data?.data) ? r.data.data : []))
      .catch(() => setSections([]))
  }, [filters.department_id])

  // Relational cascading: Section -> Designation & Line
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
    Promise.all([
      designationApi.list(filters.section_id, { limit: "100" }),
      lineApi.list(filters.section_id, { limit: "100" }),
    ]).then(([desRes, lineRes]) => {
      setDesignations(Array.isArray(desRes.data?.data) ? desRes.data.data : [])
      setLines(Array.isArray(lineRes.data?.data) ? lineRes.data.data : [])
    }).catch(() => {
      setDesignations([])
      setLines([])
    })
  }, [filters.section_id])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en })) },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    { key: "section_id", label: "Section", type: "select", options: sections.map((s) => ({ value: s.id, label: s.name })), disabled: !filters.department_id },
    { key: "designation_id", label: "Designation", type: "select", options: designations.map((d) => ({ value: d.id, label: d.name })), disabled: !filters.section_id },
    { key: "line_id", label: "Line", type: "select", options: lines.map((l) => ({ value: l.id, label: l.name })), disabled: !filters.section_id },
    { key: "group_id", label: "Group", type: "select", options: groups.map((g) => ({ value: g.id, label: g.name })) },
    { key: "employee_id", label: "Employee ID", type: "text", placeholder: "Search by Employee ID..." },
  ], [companies, departments, sections, designations, lines, groups, filters.department_id, filters.section_id])

  const applyAdvanceFilters = React.useCallback((employees: JobAgeRecord[]) => {
    let result = employees
    const min = parseInt(minAge)
    const max = parseInt(maxAge)
    if (!isNaN(min) && min >= 0) result = result.filter((e) => e.job_years >= min)
    if (!isNaN(max) && max >= 0) result = result.filter((e) => e.job_years <= max)
    if (searchEmp.trim()) {
      const q = searchEmp.trim().toLowerCase()
      result = result.filter((e) => e.employee_id.toLowerCase().includes(q) || e.name_en.toLowerCase().includes(q))
    }
    return result
  }, [minAge, maxAge, searchEmp])

  const fetchData = async (params: Record<string, string>, p = page) => {
    setLoading(true)
    try {
      const active: Record<string, string> = { status: "active", page: String(p), limit: String(limit) }
      if (params.company_id) active.company_id = params.company_id
      if (params.department_id) active.department_id = params.department_id
      if (params.section_id) active.section_id = params.section_id
      if (params.designation_id) active.designation_id = params.designation_id
      if (params.line_id) active.line_id = params.line_id
      if (params.group_id) active.group_id = params.group_id
      if (params.employee_id) active.employee_id = params.employee_id

      const { data: res } = await employeeApi.list(active)
      const employees: any[] = Array.isArray((res as any)?.data) ? (res as any).data : Array.isArray((res as any)?.data?.data) ? (res as any).data.data : []
      const totalRes = (res as any)?.total ?? employees.length
      const totalPagesRes = (res as any)?.total_pages ?? (Math.ceil(totalRes / limit) || 1)
      console.log("[JobAge] fetch", active, "->", employees.length, "employees, total", totalRes)
      const mapped: JobAgeRecord[] = employees.map((e: any) => {
        const age = calcJobAge(e.joining_date)
        return {
          id: e.id,
          employee_id: e.employee_id || "",
          name_en: e.name_en || "",
          designation: e.designation_ref?.name || e.designation || "",
          department: e.department?.name || e.department || "",
          joining_date: e.joining_date || "",
          job_years: age.years,
          job_months: age.months,
          job_days: age.days,
          status: e.status || "",
        }
      })
      setData(mapped)
      setFilteredData(applyAdvanceFilters(mapped))
      setTotal(totalRes)
      setTotalPages(totalPagesRes)
    } catch {
      setData([])
      setFilteredData([])
      setTotal(0)
      setTotalPages(0)
    } finally {
      setLoading(false)
    }
  }

  const handleApply = () => { setPage(1); fetchData(filters, 1) }
  const handleReset = () => {
    setFilters({})
    setMinAge("")
    setMaxAge("")
    setSearchEmp("")
    setActivePreset("all")
    setPage(1)
    fetchData({}, 1)
  }
  const handleChange = (key: string, value: string) => setFilters((prev) => ({ ...prev, [key]: value }))

  const handlePreset = (preset: string, min: string, max: string) => {
    setActivePreset(preset)
    setMinAge(min)
    setMaxAge(max)
  }

  React.useEffect(() => {
    setFilteredData(applyAdvanceFilters(data))
  }, [minAge, maxAge, searchEmp, data, applyAdvanceFilters])

  // Summary Metrics
  const totalCount = data.length
  const filteredCount = filteredData.length
  const avgYears = totalCount > 0 ? (data.reduce((acc, curr) => acc + curr.job_years + (curr.job_months / 12), 0) / totalCount).toFixed(1) : "0"
  const maxYears = totalCount > 0 ? Math.max(...data.map((d) => d.job_years)) : 0

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <ClockIcon className="h-6 w-6 text-muted-foreground" />
            <div>
              <h1 className="text-3xl font-bold tracking-tight">Job Age</h1>
              <p className="text-muted-foreground mt-1">Employee length of service / job age report</p>
            </div>
          </div>
          <div className="hidden sm:flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowAdvance(!showAdvance)}
              className="h-9 gap-1.5"
            >
              <FilterIcon className="h-3.5 w-3.5" />
              <span>{showAdvance ? "Hide Advance Filters" : "Advance Filters"}</span>
              {(minAge || maxAge || searchEmp) && (
                <Badge variant="secondary" className="ml-1 px-1.5 py-0 text-[10px]">Active</Badge>
              )}
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
                    onApply={() => { handleApply(); setMobileFilterOpen(false) }}
                    onReset={() => { handleReset(); setMobileFilterOpen(false) }}
                    submitting={loading}
                    singleColumn
                    noBorder
                  />
                </div>
              </SheetContent>
            </Sheet>
            <Button
              variant="outline"
              className="flex-1"
              onClick={() => setShowAdvance(!showAdvance)}
            >
              Advance
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

      {/* Advance Filter Section */}
      <div className="px-4 lg:px-6">
        <Card className="overflow-hidden border border-border/80 shadow-xs">
          <button
            type="button"
            onClick={() => setShowAdvance(!showAdvance)}
            className="flex items-center justify-between w-full px-4 py-3 text-sm font-medium bg-muted/20 hover:bg-muted/40 transition-colors"
          >
            <div className="flex items-center gap-2">
              <FilterIcon className="h-4 w-4 text-primary" />
              <span className="font-semibold">Advance Search & Tenure Filters</span>
              {(minAge || maxAge || searchEmp) && (
                <span className="text-xs bg-primary/10 text-primary font-medium px-2 py-0.5 rounded-full">
                  Filtered
                </span>
              )}
            </div>
            {showAdvance ? <ChevronUpIcon className="h-4 w-4 text-muted-foreground" /> : <ChevronDownIcon className="h-4 w-4 text-muted-foreground" />}
          </button>
          {showAdvance && (
            <div className="px-4 py-4 border-t bg-card space-y-4">
              {/* Presets */}
              <div className="space-y-1.5">
                <Label className="text-xs text-muted-foreground">Quick Tenure Ranges</Label>
                <div className="flex flex-wrap gap-2">
                  {[
                    { id: "all", label: "All Employees", min: "", max: "" },
                    { id: "less1", label: "< 1 Year", min: "0", max: "0" },
                    { id: "1to3", label: "1 - 3 Years", min: "1", max: "3" },
                    { id: "3to5", label: "3 - 5 Years", min: "3", max: "5" },
                    { id: "5to10", label: "5 - 10 Years", min: "5", max: "10" },
                    { id: "10plus", label: "10+ Years", min: "10", max: "" },
                  ].map((p) => (
                    <Button
                      key={p.id}
                      type="button"
                      size="sm"
                      variant={activePreset === p.id ? "default" : "outline"}
                      className="h-7 text-xs font-normal"
                      onClick={() => handlePreset(p.id, p.min, p.max)}
                    >
                      {p.label}
                    </Button>
                  ))}
                </div>
              </div>

              {/* Input grid */}
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 pt-1">
                <div className="space-y-1.5">
                  <Label htmlFor="search-emp" className="text-xs font-medium">Search by Employee ID / Name</Label>
                  <div className="relative">
                    <SearchIcon className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="search-emp"
                      type="text"
                      placeholder="e.g. 1024 or Shakil..."
                      className="pl-8 h-9 text-sm"
                      value={searchEmp}
                      onChange={(e) => setSearchEmp(e.target.value)}
                    />
                  </div>
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="min-age" className="text-xs font-medium">Min Job Age (years)</Label>
                  <Input
                    id="min-age"
                    type="number"
                    min="0"
                    placeholder="e.g. 1"
                    className="h-9 text-sm"
                    value={minAge}
                    onChange={(e) => {
                      setMinAge(e.target.value)
                      setActivePreset("custom")
                    }}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="max-age" className="text-xs font-medium">Max Job Age (years)</Label>
                  <Input
                    id="max-age"
                    type="number"
                    min="0"
                    placeholder="e.g. 10"
                    className="h-9 text-sm"
                    value={maxAge}
                    onChange={(e) => {
                      setMaxAge(e.target.value)
                      setActivePreset("custom")
                    }}
                  />
                </div>
              </div>

              <div className="flex items-center justify-between pt-2 border-t text-xs text-muted-foreground">
                <span>Showing <b>{filteredCount}</b> of <b>{totalCount}</b> employees</span>
                {(minAge || maxAge || searchEmp || activePreset !== "all") && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      setMinAge("")
                      setMaxAge("")
                      setSearchEmp("")
                      setActivePreset("all")
                    }}
                    className="h-7 text-xs text-muted-foreground hover:text-foreground"
                  >
                    <RotateCcwIcon className="mr-1 h-3 w-3" />
                    Clear Advance Filters
                  </Button>
                )}
              </div>
            </div>
          )}
        </Card>
      </div>

      {/* Summary KPI Cards */}
      <div className="px-4 lg:px-6 grid grid-cols-2 sm:grid-cols-4 gap-3 md:gap-4">
        <div className="rounded-lg border bg-card p-3.5">
          <div className="flex items-center gap-2 text-muted-foreground text-xs font-medium">
            <UsersIcon className="h-4 w-4 text-primary" />
            <span>Total Loaded</span>
          </div>
          <p className="text-xl font-bold mt-1">{totalCount}</p>
        </div>
        <div className="rounded-lg border bg-card p-3.5">
          <div className="flex items-center gap-2 text-muted-foreground text-xs font-medium">
            <FilterIcon className="h-4 w-4 text-emerald-500" />
            <span>Matching Filter</span>
          </div>
          <p className="text-xl font-bold mt-1 text-emerald-600 dark:text-emerald-400">{filteredCount}</p>
        </div>
        <div className="rounded-lg border bg-card p-3.5">
          <div className="flex items-center gap-2 text-muted-foreground text-xs font-medium">
            <CalendarIcon className="h-4 w-4 text-blue-500" />
            <span>Average Tenure</span>
          </div>
          <p className="text-xl font-bold mt-1">{avgYears} yrs</p>
        </div>
        <div className="rounded-lg border bg-card p-3.5">
          <div className="flex items-center gap-2 text-muted-foreground text-xs font-medium">
            <AwardIcon className="h-4 w-4 text-amber-500" />
            <span>Max Tenure</span>
          </div>
          <p className="text-xl font-bold mt-1">{maxYears} yrs</p>
        </div>
      </div>

      <DataTable
        data={filteredData}
        columns={columns}
        loading={loading}
        enableSelection={false}
        enableDnd={false}
        serverSide={true}
        page={page}
        pageSize={limit}
        pageCount={totalPages}
        total={total}
        onPageChange={(p)=> { setPage(p); fetchData(filters, p) }}
        onPageSizeChange={()=>{}}
      />
    </div>
  )
}

