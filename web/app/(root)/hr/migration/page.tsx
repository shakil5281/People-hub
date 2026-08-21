"use client"

import * as React from "react"
import { ArrowRightLeftIcon, Loader2, FilterIcon, XIcon, FileSpreadsheetIcon, FileTextIcon, UserPlusIcon, UserMinusIcon, UserXIcon, UsersIcon, Building2Icon, LayersIcon, AwardIcon, GitCommitIcon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { migrationApi, companyApi, departmentApi, sectionApi, designationApi, lineApi } from "@/lib/api"
import type { Department, Section, Designation, Line } from "@/components/data/organization-data"
import type { Company } from "@/components/data/company-data"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"

export interface OrganizationalRow {
  id: string
  name: string
  new_joining: number
  total_left: number
  total_resign: number
  active_total: number
}

export interface SummaryData {
  total_joining: number
  total_left: number
  total_resign: number
  total_active: number
  by_department: OrganizationalRow[]
  by_section: OrganizationalRow[]
  by_designation: OrganizationalRow[]
  by_line: OrganizationalRow[]
}

const today = new Date().toISOString().split("T")[0]
const firstOfMonth = new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString().split("T")[0]

export default function MigrationSummaryPage() {
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])

  const [summary, setSummary] = React.useState<SummaryData>({
    total_joining: 0,
    total_left: 0,
    total_resign: 0,
    total_active: 0,
    by_department: [],
    by_section: [],
    by_designation: [],
    by_line: [],
  })

  const [loading, setLoading] = React.useState(true)
  const [exporting, setExporting] = React.useState<"excel" | "pdf" | null>(null)
  const [filters, setFilters] = React.useState<Record<string, string>>({
    date_from: firstOfMonth,
    date_to: today,
  })
  const [submitting, setSubmitting] = React.useState(false)
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)
  const [activeTab, setActiveTab] = React.useState("department")

  const fetchSummaryData = React.useCallback(async (f?: Record<string, string>) => {
    setLoading(true)
    try {
      const active: Record<string, string> = {}
      for (const [k, v] of Object.entries(f || filters)) {
        if (v) active[k] = v
      }
      const res = await migrationApi.summary(active)
      if (res.data) {
        setSummary({
          total_joining: res.data.total_joining || 0,
          total_left: res.data.total_left || 0,
          total_resign: res.data.total_resign || 0,
          total_active: res.data.total_active || 0,
          by_department: Array.isArray(res.data.by_department) ? res.data.by_department : [],
          by_section: Array.isArray(res.data.by_section) ? res.data.by_section : [],
          by_designation: Array.isArray(res.data.by_designation) ? res.data.by_designation : [],
          by_line: Array.isArray(res.data.by_line) ? res.data.by_line : [],
        })
      }
    } catch {
      toast.error("Failed to load migration summary data")
    } finally {
      setLoading(false)
    }
  }, [filters])

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
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
    ]).then(([cRes, dRes]) => {
      setCompanies(Array.isArray(cRes.data?.data) ? cRes.data.data : [])
      setDepartments(Array.isArray(dRes.data?.data) ? dRes.data.data : [])
    }).catch(() => {})
    fetchSummaryData(filters)
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
  }, [filters.department_id, loadSections])

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
  }, [filters.section_id, loadDesignations, loadLines])

  const handleApply = async () => {
    setSubmitting(true)
    await fetchSummaryData(filters)
    setSubmitting(false)
  }

  const handleReset = () => {
    setSections([])
    setDesignations([])
    setLines([])
    const def = { date_from: firstOfMonth, date_to: today }
    setFilters(def)
    fetchSummaryData(def)
  }

  const handleChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleExport = async (kind: "excel" | "pdf") => {
    setExporting(kind)
    try {
      const active: Record<string, string> = {}
      for (const [k, v] of Object.entries(filters)) {
        if (v) active[k] = v
      }
      const res = kind === "excel" ? await migrationApi.exportExcel(active) : await migrationApi.exportPdf(active)
      const blob = new Blob([res.data], {
        type: kind === "excel" ? "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" : "application/pdf",
      })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement("a")
      link.href = url
      const ext = kind === "excel" ? "xlsx" : "pdf"
      link.download = `migration_summary_${new Date().toISOString().slice(0, 10)}.${ext}`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
      toast.success(`Migration summary (${kind.toUpperCase()}) exported successfully`)
    } catch {
      toast.error(`Failed to export ${kind.toUpperCase()}`)
    } finally {
      setExporting(null)
    }
  }

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date_from", label: "Starting Date", type: "datepicker" },
    { key: "date_to", label: "End Date", type: "datepicker" },
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en })) },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    {
      key: "section_id",
      label: "Section",
      type: "select",
      options: sections.map((s) => ({ value: s.id, label: s.name })),
      disabled: !filters.department_id,
    },
    {
      key: "designation_id",
      label: "Designation",
      type: "select",
      options: designations.map((d) => ({ value: d.id, label: d.name })),
      disabled: !filters.section_id,
    },
    {
      key: "line_id",
      label: "Line",
      type: "select",
      options: lines.map((l) => ({ value: l.id, label: l.name })),
      disabled: !filters.section_id,
    },
  ], [companies, departments, sections, designations, lines, filters.department_id, filters.section_id])

  const renderBreakdownTable = (title: string, rows: OrganizationalRow[]) => {
    return (
      <Card className="mt-4">
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow className="bg-muted/50">
                <TableHead className="w-[60px] text-center">SL</TableHead>
                <TableHead>{title} Name</TableHead>
                <TableHead className="text-center font-semibold text-emerald-600">New Joining</TableHead>
                <TableHead className="text-center font-semibold text-rose-600">Total Left / Close</TableHead>
                <TableHead className="text-center font-semibold text-amber-600">Total Resign</TableHead>
                <TableHead className="text-center font-semibold text-blue-600">Current Active</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                    No summary data available for this range
                  </TableCell>
                </TableRow>
              ) : (
                rows.map((row, idx) => (
                  <TableRow key={row.id || idx} className="hover:bg-muted/30">
                    <TableCell className="text-center font-medium">{idx + 1}</TableCell>
                    <TableCell className="font-medium text-foreground">{row.name}</TableCell>
                    <TableCell className="text-center">
                      <Badge variant="outline" className="bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400 border-emerald-200">
                        +{row.new_joining}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-center">
                      <Badge variant="outline" className="bg-rose-50 text-rose-700 dark:bg-rose-950/40 dark:text-rose-400 border-rose-200">
                        {row.total_left}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-center">
                      <Badge variant="outline" className="bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400 border-amber-200">
                        {row.total_resign}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-center font-semibold text-blue-600 dark:text-blue-400">
                      {row.active_total}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      {/* Header */}
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <ArrowRightLeftIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Employee Migration Summary</h1>
            <p className="text-muted-foreground mt-1">Starting date to end date joining, left, resign & active summary</p>
          </div>
        </div>
        <ButtonGroup className="hidden md:flex">
          <Button variant="outline" size="sm" onClick={() => handleExport("excel")} disabled={loading || !!exporting}>
            {exporting === "excel" ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileSpreadsheetIcon className="mr-2 h-4 w-4 text-green-600" />}
            Export Excel
          </Button>
          <Button variant="outline" size="sm" onClick={() => handleExport("pdf")} disabled={loading || !!exporting}>
            {exporting === "pdf" ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileTextIcon className="mr-2 h-4 w-4 text-red-600" />}
            Export PDF
          </Button>
        </ButtonGroup>
      </div>

      {/* Top Metric Cards */}
      <div className="px-4 lg:px-6 grid grid-cols-2 sm:grid-cols-4 gap-3 md:gap-4">
        <Card className="border-emerald-200 bg-emerald-50/20 dark:bg-emerald-950/10">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-xs font-medium text-emerald-700 dark:text-emerald-400">Total New Joining</p>
              <h3 className="text-2xl font-bold mt-1 text-emerald-700 dark:text-emerald-300">{summary.total_joining}</h3>
            </div>
            <UserPlusIcon className="h-8 w-8 text-emerald-600/80" />
          </CardContent>
        </Card>
        <Card className="border-rose-200 bg-rose-50/20 dark:bg-rose-950/10">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-xs font-medium text-rose-700 dark:text-rose-400">Total Employee Left</p>
              <h3 className="text-2xl font-bold mt-1 text-rose-700 dark:text-rose-300">{summary.total_left}</h3>
            </div>
            <UserMinusIcon className="h-8 w-8 text-rose-600/80" />
          </CardContent>
        </Card>
        <Card className="border-amber-200 bg-amber-50/20 dark:bg-amber-950/10">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-xs font-medium text-amber-700 dark:text-amber-400">Total Resign</p>
              <h3 className="text-2xl font-bold mt-1 text-amber-700 dark:text-amber-300">{summary.total_resign}</h3>
            </div>
            <UserXIcon className="h-8 w-8 text-amber-600/80" />
          </CardContent>
        </Card>
        <Card className="border-blue-200 bg-blue-50/20 dark:bg-blue-950/10">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-xs font-medium text-blue-700 dark:text-blue-400">Total Active Employees</p>
              <h3 className="text-2xl font-bold mt-1 text-blue-700 dark:text-blue-300">{summary.total_active}</h3>
            </div>
            <UsersIcon className="h-8 w-8 text-blue-600/80" />
          </CardContent>
        </Card>
      </div>

      {/* Mobile Filter */}
      <div className="md:hidden px-4 lg:px-6">
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
                  submitting={submitting}
                  singleColumn
                  noBorder
                />
              </div>
            </SheetContent>
          </Sheet>
        </ButtonGroup>
        <ButtonGroup className="w-full mt-2">
          <Button variant="outline" size="sm" onClick={() => handleExport("excel")} disabled={loading || !!exporting} className="flex-1">
            {exporting === "excel" ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <FileSpreadsheetIcon className="mr-1.5 h-4 w-4 text-green-600" />}
            Excel
          </Button>
          <Button variant="outline" size="sm" onClick={() => handleExport("pdf")} disabled={loading || !!exporting} className="flex-1">
            {exporting === "pdf" ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <FileTextIcon className="mr-1.5 h-4 w-4 text-red-600" />}
            PDF
          </Button>
        </ButtonGroup>
      </div>

      {/* Desktop Relational Advance Filter */}
      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleChange}
          onApply={handleApply}
          onReset={handleReset}
          submitting={submitting}
        />
      </div>

      {/* Organizational Breakdown Tabs */}
      <div className="px-4 lg:px-6">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="grid grid-cols-4 w-full md:w-[600px]">
            <TabsTrigger value="department" className="flex items-center gap-1.5">
              <Building2Icon className="h-4 w-4" />
              Department
            </TabsTrigger>
            <TabsTrigger value="section" className="flex items-center gap-1.5">
              <LayersIcon className="h-4 w-4" />
              Section
            </TabsTrigger>
            <TabsTrigger value="designation" className="flex items-center gap-1.5">
              <AwardIcon className="h-4 w-4" />
              Designation
            </TabsTrigger>
            <TabsTrigger value="line" className="flex items-center gap-1.5">
              <GitCommitIcon className="h-4 w-4" />
              Line
            </TabsTrigger>
          </TabsList>

          <TabsContent value="department">
            {renderBreakdownTable("Department", summary.by_department)}
          </TabsContent>
          <TabsContent value="section">
            {renderBreakdownTable("Section", summary.by_section)}
          </TabsContent>
          <TabsContent value="designation">
            {renderBreakdownTable("Designation", summary.by_designation)}
          </TabsContent>
          <TabsContent value="line">
            {renderBreakdownTable("Line", summary.by_line)}
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
