"use client"

import * as React from "react"
import {
  TrendingUpIcon,
  FilterIcon,
  XIcon,
  UserIcon,
  CalendarIcon,
  ClockIcon,
  CheckCircle2Icon,
  AlertCircleIcon,
  HourglassIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  Loader2,
  FileSpreadsheetIcon,
  BriefcaseIcon,
  Building2Icon,
  PhoneIcon,
  ArrowRightIcon,
  DollarSignIcon,
  LayersIcon,
  AwardIcon,
  ArrowLeftIcon,
} from "lucide-react"
import Link from "next/link"
import { format } from "date-fns"
import { toast } from "sonner"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import {
  salaryIncrementApi,
  companyApi,
  departmentApi,
  sectionApi,
  designationApi,
  employeeApi,
} from "@/lib/api"
import { downloadExport } from "@/lib/utils"

interface EmployeeProfile {
  id: string
  employee_id: string
  punch_number?: string
  name_en: string
  name_bn?: string
  phone?: string
  gender?: string
  status: string
  employee_type?: string
  joining_date?: string
  gross_salary: number
  basic_salary: number
  house_rent: number
  medical_allowance: number
  transport_allowance: number
  food_allowance: number
  company_id?: string
  company_name?: string
  department_id?: string
  department_name?: string
  designation_id?: string
  designation_name?: string
  section_id?: string
  section_name?: string
  line_id?: string
  line_name?: string
  group_id?: string
  group_name?: string
  shift_id?: string
  shift_name?: string
  photo_url?: string
}

interface IncrementRecord {
  id: string
  employee_id: string
  increment_type: string
  calculation_type: string
  calculation_value: number
  previous_gross: number
  previous_basic: number
  increment_amount: number
  new_gross: number
  new_basic: number
  increment_date: string
  effective_date: string
  status: string
  remarks?: string
  rejection_reason?: string
  previous_designation?: { name: string }
  new_designation?: { name: string }
  created_at: string
}

interface IncrementSummary {
  current_gross: number
  current_basic: number
  total_increment_amount: number
  total_increment_count: number
  total_promotions_count: number
  last_increment_date: string
  last_increment_amount: number
  initial_gross: number
  first_time_gross: number
  first_time_basic: number
  first_time_house: number
  first_time_medical: number
  first_time_date: string
  total_records: number
}

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }

const currentYear = new Date().getFullYear()

const monthOptions = [
  { value: "", label: "All Months" },
  { value: "1", label: "January (01)" },
  { value: "2", label: "February (02)" },
  { value: "3", label: "March (03)" },
  { value: "4", label: "April (04)" },
  { value: "5", label: "May (05)" },
  { value: "6", label: "June (06)" },
  { value: "7", label: "July (07)" },
  { value: "8", label: "August (08)" },
  { value: "9", label: "September (09)" },
  { value: "10", label: "October (10)" },
  { value: "11", label: "November (11)" },
  { value: "12", label: "December (12)" },
]

const yearOptions = [
  { value: "", label: "All Years" },
  ...Array.from({ length: 7 }, (_, i) => {
    const y = currentYear - i + 1
    return { value: String(y), label: String(y) }
  }),
]

const renderTypeBadge = (type?: string) => {
  switch (type?.toLowerCase()) {
    case "gov_policy":
    case "gov":
    case "policy":
      return <Badge className="bg-emerald-600 hover:bg-emerald-700 text-white text-xs">Gov Policy</Badge>
    case "promotion":
      return <Badge className="bg-amber-600 hover:bg-amber-700 text-white text-xs">Promotion</Badge>
    case "promotion_with_increment":
      return <Badge className="bg-purple-600 hover:bg-purple-700 text-white text-xs">Promotion + Inc</Badge>
    default:
      return <Badge variant="secondary" className="text-xs capitalize">{type || "Increment"}</Badge>
  }
}

const getStatusBadge = (status: string) => {
  switch (status?.toLowerCase()) {
    case "approved":
      return <Badge className="bg-emerald-600 hover:bg-emerald-700 text-white font-medium">Approved</Badge>
    case "pending":
      return <Badge variant="outline" className="text-amber-600 border-amber-500 bg-amber-50 dark:bg-amber-950/40 font-medium">Pending</Badge>
    case "rejected":
      return <Badge variant="destructive" className="font-medium">Rejected</Badge>
    default:
      return <Badge variant="outline" className="capitalize">{status || "Pending"}</Badge>
  }
}

export default function IncrementDetailsPage() {
  const [employee, setEmployee] = React.useState<EmployeeProfile | null>(null)
  const [increments, setIncrements] = React.useState<IncrementRecord[]>([])
  const [summary, setSummary] = React.useState<IncrementSummary>({
    current_gross: 0,
    current_basic: 0,
    total_increment_amount: 0,
    total_increment_count: 0,
    total_promotions_count: 0,
    last_increment_date: "",
    last_increment_amount: 0,
    initial_gross: 0,
    first_time_gross: 0,
    first_time_basic: 0,
    first_time_house: 0,
    first_time_medical: 0,
    first_time_date: "",
    total_records: 0,
  })

  const [loading, setLoading] = React.useState(false)
  const [exporting, setExporting] = React.useState(false)
  const [filters, setFilters] = React.useState<Record<string, string>>({
    year: "",
    month: "",
    employee_id: "",
  })

  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  // Employee list for quick browsing / next-prev navigation
  const [employeeList, setEmployeeList] = React.useState<{ id: string; employee_id: string; name_en: string }[]>([])
  const [currentEmpIndex, setCurrentEmpIndex] = React.useState(0)

  // Load dropdown options
  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      employeeApi.list({ limit: "1000", status: "active" }),
    ]).then(([cRes, dRes, eRes]) => {
      setCompanies(cRes.data?.data || [])
      setDepartments(dRes.data?.data || [])
      const empData = eRes.data?.data?.employees || eRes.data?.data || []
      if (Array.isArray(empData)) {
        setEmployeeList(empData)
      }
    }).catch(() => {})

    // Initial load
    fetchIncrementDetails({})
  }, [])

  // Dynamic section loading
  React.useEffect(() => {
    if (!filters.department_id) {
      setSections([])
      setDesignations([])
      return
    }
    sectionApi.list(filters.department_id, { limit: "100" })
      .then((r) => setSections(r.data?.data || []))
      .catch(() => setSections([]))
  }, [filters.department_id])

  // Dynamic designation loading
  React.useEffect(() => {
    if (!filters.section_id) {
      setDesignations([])
      return
    }
    designationApi.list(filters.section_id, { limit: "100" })
      .then((r) => setDesignations(r.data?.data || []))
      .catch(() => setDesignations([]))
  }, [filters.section_id])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    {
      key: "employee_id",
      label: "Employee ID",
      type: "text",
      placeholder: "Enter Employee Code...",
    },
    {
      key: "year",
      label: "Year",
      type: "select",
      options: yearOptions,
    },
    {
      key: "month",
      label: "Month",
      type: "select",
      options: monthOptions,
    },
    {
      key: "company_id",
      label: "Company",
      type: "select",
      options: companies.map((c) => ({ value: c.id, label: c.company_name_en })),
    },
    {
      key: "department_id",
      label: "Department",
      type: "select",
      options: departments.map((d) => ({ value: d.id, label: d.name })),
    },
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
  ], [companies, departments, sections, designations, filters.department_id, filters.section_id])

  const fetchIncrementDetails = React.useCallback(async (params: Record<string, string>) => {
    setLoading(true)
    try {
      const active: Record<string, string> = {}
      for (const [k, v] of Object.entries(params)) {
        if (v !== undefined && v !== null && v !== "") {
          active[k] = v
        }
      }

      const { data: res } = await salaryIncrementApi.getDetails(active)
      if (res) {
        setEmployee(res.employee || null)
        setIncrements(Array.isArray(res.increments) ? res.increments : [])
        if (res.summary) {
          setSummary(res.summary)
        }

        // Sync index in employee list if present
        if (res.employee?.employee_id && employeeList.length > 0) {
          const idx = employeeList.findIndex((e) => e.employee_id === res.employee.employee_id)
          if (idx !== -1) setCurrentEmpIndex(idx)
        }
      }
    } catch (err: any) {
      toast.error("Failed to load increment details")
      setEmployee(null)
      setIncrements([])
    } finally {
      setLoading(false)
    }
  }, [employeeList])

  const handleFilterChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleApply = () => {
    fetchIncrementDetails(filters)
  }

  const handleReset = () => {
    const defaultFilters = { year: "", month: "", employee_id: "" }
    setFilters(defaultFilters)
    fetchIncrementDetails(defaultFilters)
  }

  const handlePrev = () => {
    if (employeeList.length === 0 || currentEmpIndex <= 0) return
    const prevIdx = currentEmpIndex - 1
    const prevEmp = employeeList[prevIdx]
    setCurrentEmpIndex(prevIdx)
    const newFilters = { ...filters, employee_id: prevEmp.employee_id }
    setFilters(newFilters)
    fetchIncrementDetails(newFilters)
  }

  const handleNext = () => {
    if (employeeList.length === 0 || currentEmpIndex >= employeeList.length - 1) return
    const nextIdx = currentEmpIndex + 1
    const nextEmp = employeeList[nextIdx]
    setCurrentEmpIndex(nextIdx)
    const newFilters = { ...filters, employee_id: nextEmp.employee_id }
    setFilters(newFilters)
    fetchIncrementDetails(newFilters)
  }

  const handleExportExcel = async () => {
    if (!employee) return
    setExporting(true)
    try {
      const params: Record<string, string> = {
        company_id: employee.company_id || "",
        employee_id: employee.employee_id,
      }
      if (filters.year) params.year = filters.year
      if (filters.month) params.month = filters.month
      const res = await salaryIncrementApi.exportExcel(params)
      downloadExport(res, `increment_details_${employee.employee_id}.xlsx`)
      toast.success("Increment details exported to Excel")
    } catch {
      toast.error("Failed to export Excel")
    } finally {
      setExporting(false)
    }
  }

  const handleExportPdf = async () => {
    if (!employee) return
    setExporting(true)
    try {
      const params: Record<string, string> = {
        company_id: employee.company_id || "",
        employee_id: employee.employee_id,
      }
      if (filters.year) params.year = filters.year
      if (filters.month) params.month = filters.month
      const res = await salaryIncrementApi.exportPdf(params)
      downloadExport(res, `increment_details_${employee.employee_id}.pdf`)
      toast.success("Increment details exported to PDF")
    } catch {
      toast.error("Failed to export PDF")
    } finally {
      setExporting(false)
    }
  }

  // Calculate service tenure / job age
  const getJobAge = (joiningDate?: string) => {
    if (!joiningDate) return "-"
    try {
      const join = new Date(joiningDate)
      const now = new Date()
      let years = now.getFullYear() - join.getFullYear()
      let months = now.getMonth() - join.getMonth()
      if (months < 0) {
        years--
        months += 12
      }
      if (years <= 0 && months <= 0) return "New Joiner (<1 mo)"
      if (years <= 0) return `${months} month${months > 1 ? "s" : ""}`
      return `${years} yr${years > 1 ? "s" : ""} ${months > 0 ? `${months} mo` : ""}`
    } catch {
      return "-"
    }
  }

  return (
    <div className="flex flex-col gap-5 py-4 md:gap-6 md:py-6">
      {/* Header */}
      <div className="px-4 lg:px-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <TrendingUpIcon className="h-6 w-6" />
            </div>
            <div>
              <h1 className="text-xl md:text-2xl font-bold tracking-tight">Increment Details & Reports</h1>
              <p className="text-sm text-muted-foreground">
                Employee salary increment history, designation progression, and compensation growth breakdown
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Link href="/payroll/increment">
              <Button variant="outline" size="sm">
                <ArrowLeftIcon className="h-4 w-4 mr-1.5" />
                Increment List
              </Button>
            </Link>
            {employee && (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleExportExcel}
                  disabled={exporting || loading}
                >
                  <FileSpreadsheetIcon className="h-4 w-4 mr-1.5 text-emerald-600" />
                  {exporting ? "Exporting..." : "Export Excel"}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleExportPdf}
                  disabled={exporting || loading}
                >
                  <FileSpreadsheetIcon className="h-4 w-4 mr-1.5 text-rose-600" />
                  {exporting ? "Exporting..." : "Export PDF"}
                </Button>
              </>
            )}
            <div className="md:hidden">
              <Sheet open={mobileFilterOpen} onOpenChange={setMobileFilterOpen}>
                <SheetTrigger asChild>
                  <Button variant="outline" size="sm">
                    <FilterIcon className="h-4 w-4 mr-1.5" />
                    Filters
                  </Button>
                </SheetTrigger>
                <SheetContent side="right" className="w-full sm:max-w-md p-0 flex flex-col" showCloseButton={false}>
                  <SheetHeader className="px-4 py-3 border-b flex flex-row items-center justify-between">
                    <SheetTitle className="text-base">Filters & Search</SheetTitle>
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
                      onChange={handleFilterChange}
                      onApply={() => { handleApply(); setMobileFilterOpen(false) }}
                      onReset={() => { handleReset(); setMobileFilterOpen(false) }}
                      submitting={loading}
                      singleColumn
                      noBorder
                    />
                  </div>
                </SheetContent>
              </Sheet>
            </div>
          </div>
        </div>
      </div>

      {/* Desktop Filter Bar */}
      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleFilterChange}
          onApply={handleApply}
          onReset={handleReset}
          submitting={loading}
        />
      </div>

      {/* Main Content Area */}
      <div className="px-4 lg:px-6 flex flex-col gap-6">
        {loading ? (
          <div className="flex flex-col items-center justify-center p-16 rounded-xl border bg-card text-muted-foreground gap-3">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
            <p className="text-sm font-medium">Loading employee increment details...</p>
          </div>
        ) : !employee ? (
          <div className="flex flex-col items-center justify-center p-16 rounded-xl border bg-card text-center text-muted-foreground gap-3">
            <UserIcon className="h-12 w-12 text-muted-foreground/50 stroke-1" />
            <h3 className="text-base font-semibold text-foreground">No Employee Selected</h3>
            <p className="text-sm max-w-sm">
              Search by Employee ID in the filters above to inspect increment records, promotion milestones, and salary history.
            </p>
          </div>
        ) : (
          <>
            {/* Employee Navigation Bar */}
            {employeeList.length > 0 && (
              <div className="flex items-center justify-between bg-card border rounded-lg px-4 py-2 text-sm shadow-xs">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handlePrev}
                  disabled={currentEmpIndex === 0 || loading}
                  className="h-8 text-xs"
                >
                  <ChevronLeftIcon className="h-4 w-4 mr-1" /> Previous Employee
                </Button>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <span>Employee <b>{currentEmpIndex + 1}</b> of <b>{employeeList.length}</b></span>
                  <span className="hidden sm:inline">•</span>
                  <span className="hidden sm:inline font-mono font-medium text-foreground">{employee.employee_id}</span>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleNext}
                  disabled={currentEmpIndex >= employeeList.length - 1 || loading}
                  className="h-8 text-xs"
                >
                  Next Employee <ChevronRightIcon className="h-4 w-4 ml-1" />
                </Button>
              </div>
            )}

            {/* Employee Profile Data Card / Info Table */}
            <Card className="shadow-xs overflow-hidden border-border/80">
              <CardHeader className="bg-muted/30 border-b pb-3 pt-4">
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                  <div className="flex items-center gap-3">
                    <div className="h-12 w-12 rounded-full bg-primary/10 text-primary font-bold flex items-center justify-center text-lg uppercase border border-primary/20">
                      {employee.name_en?.charAt(0) || "E"}
                    </div>
                    <div>
                      <div className="flex items-center gap-2 flex-wrap">
                        <h2 className="text-lg font-bold text-foreground">{employee.name_en}</h2>
                        {employee.name_bn && (
                          <span className="text-sm text-muted-foreground">({employee.name_bn})</span>
                        )}
                        <Badge variant="outline" className="font-mono bg-background">
                          ID: {employee.employee_id}
                        </Badge>
                        <Badge
                          variant={employee.status === "active" ? "default" : "secondary"}
                          className="capitalize text-xs"
                        >
                          {employee.status || "Active"}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted-foreground mt-0.5">
                        {employee.designation_name || "Employee"} • {employee.department_name || "General"}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 self-start sm:self-auto">
                    {filters.year && (
                      <Badge variant="outline" className="text-xs font-normal">
                        Year: <b className="ml-1 text-foreground">{filters.year}</b>
                      </Badge>
                    )}
                    {filters.month && (
                      <Badge variant="outline" className="text-xs font-normal">
                        Month: <b className="ml-1 text-foreground">{monthOptions.find(m => m.value === filters.month)?.label || filters.month}</b>
                      </Badge>
                    )}
                  </div>
                </div>
              </CardHeader>

              <CardContent className="p-0">
                <div className="grid grid-cols-2 md:grid-cols-4 divide-x divide-y md:divide-y-0 border-b text-xs">
                  <div className="p-3.5 space-y-1">
                    <div className="text-muted-foreground flex items-center gap-1.5">
                      <Building2Icon className="h-3.5 w-3.5" /> Company
                    </div>
                    <div className="font-medium text-foreground">{employee.company_name || "-"}</div>
                  </div>
                  <div className="p-3.5 space-y-1">
                    <div className="text-muted-foreground flex items-center gap-1.5">
                      <BriefcaseIcon className="h-3.5 w-3.5" /> Section / Line
                    </div>
                    <div className="font-medium text-foreground">
                      {employee.section_name || "-"} {employee.line_name ? `• ${employee.line_name}` : ""}
                    </div>
                  </div>
                  <div className="p-3.5 space-y-1">
                    <div className="text-muted-foreground flex items-center gap-1.5">
                      <CalendarIcon className="h-3.5 w-3.5" /> Joining Date
                    </div>
                    <div className="font-medium text-foreground">
                      {employee.joining_date ? format(new Date(employee.joining_date), "dd-MM-yyyy") : "-"}
                      <span className="text-muted-foreground ml-1 font-normal">({getJobAge(employee.joining_date)})</span>
                    </div>
                  </div>
                  <div className="p-3.5 space-y-1">
                    <div className="text-muted-foreground flex items-center gap-1.5">
                      <PhoneIcon className="h-3.5 w-3.5" /> Phone / Shift
                    </div>
                    <div className="font-medium text-foreground">
                      {employee.phone || "-"} {employee.shift_name ? `(${employee.shift_name})` : ""}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Increment Summary - 5 KPI Metric Cards with First Time Salary */}
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3.5">
              <Card className="shadow-xs bg-linear-to-br from-slate-50 to-gray-50 dark:from-slate-900/50 dark:to-gray-900/30 border-slate-200 dark:border-slate-800">
                <CardContent className="p-4 flex items-center justify-between">
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-slate-700 dark:text-slate-400">First Time Salary</p>
                    <p className="text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100">
                      ৳{Number(summary.first_time_gross || summary.initial_gross || 0).toLocaleString()}
                    </p>
                    <p className="text-[11px] text-muted-foreground">
                      {summary.first_time_date ? format(new Date(summary.first_time_date), "dd MMM yyyy") : employee.joining_date ? format(new Date(employee.joining_date), "dd MMM yyyy") : "Joining"}
                    </p>
                  </div>
                  <div className="h-10 w-10 rounded-full bg-slate-500/10 text-slate-600 dark:text-slate-400 flex items-center justify-center">
                    <LayersIcon className="h-5 w-5" />
                  </div>
                </CardContent>
              </Card>

              <Card className="shadow-xs bg-linear-to-br from-blue-50/50 to-indigo-50/30 dark:from-blue-950/20 dark:to-indigo-950/10 border-blue-200/60 dark:border-blue-900/40">
                <CardContent className="p-4 flex items-center justify-between">
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-blue-700 dark:text-blue-400">Current Gross Salary</p>
                    <p className="text-2xl font-bold tracking-tight text-blue-950 dark:text-blue-100">
                      ৳{Number(employee.gross_salary || 0).toLocaleString()}
                    </p>
                    <p className="text-[11px] text-muted-foreground">Basic: ৳{Number(employee.basic_salary || 0).toLocaleString()}</p>
                  </div>
                  <div className="h-10 w-10 rounded-full bg-blue-500/10 text-blue-600 dark:text-blue-400 flex items-center justify-center">
                    <DollarSignIcon className="h-5 w-5" />
                  </div>
                </CardContent>
              </Card>

              <Card className="shadow-xs bg-linear-to-br from-emerald-50/50 to-teal-50/30 dark:from-emerald-950/20 dark:to-teal-950/10 border-emerald-200/60 dark:border-emerald-900/40">
                <CardContent className="p-4 flex items-center justify-between">
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-emerald-700 dark:text-emerald-400">Total Increment Given</p>
                    <p className="text-2xl font-bold tracking-tight text-emerald-950 dark:text-emerald-100">
                      +৳{Number(summary.total_increment_amount || 0).toLocaleString()}
                    </p>
                    <p className="text-[11px] text-muted-foreground">
                      Starting gross: ৳{Number(summary.first_time_gross || summary.initial_gross || 0).toLocaleString()}
                    </p>
                  </div>
                  <div className="h-10 w-10 rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center">
                    <TrendingUpIcon className="h-5 w-5" />
                  </div>
                </CardContent>
              </Card>

              <Card className="shadow-xs bg-linear-to-br from-amber-50/50 to-orange-50/30 dark:from-amber-950/20 dark:to-orange-950/10 border-amber-200/60 dark:border-amber-900/40">
                <CardContent className="p-4 flex items-center justify-between">
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-amber-700 dark:text-amber-400">Increments Count</p>
                    <p className="text-2xl font-bold tracking-tight text-amber-950 dark:text-amber-100">
                      {summary.total_increment_count} <span className="text-xs font-normal text-muted-foreground">Times</span>
                    </p>
                    <p className="text-[11px] text-muted-foreground">
                      {summary.total_promotions_count} Promotion{summary.total_promotions_count === 1 ? "" : "s"}
                    </p>
                  </div>
                  <div className="h-10 w-10 rounded-full bg-amber-500/10 text-amber-600 dark:text-amber-400 flex items-center justify-center">
                    <AwardIcon className="h-5 w-5" />
                  </div>
                </CardContent>
              </Card>

              <Card className="shadow-xs bg-linear-to-br from-purple-50/50 to-pink-50/30 dark:from-purple-950/20 dark:to-pink-950/10 border-purple-200/60 dark:border-purple-900/40">
                <CardContent className="p-4 flex items-center justify-between">
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-purple-700 dark:text-purple-400">Last Increment Date</p>
                    <p className="text-lg font-bold tracking-tight text-purple-950 dark:text-purple-100 truncate">
                      {summary.last_increment_date ? format(new Date(summary.last_increment_date), "dd MMM yyyy") : "None"}
                    </p>
                    <p className="text-[11px] text-muted-foreground">
                      {summary.last_increment_amount > 0 ? `+৳${summary.last_increment_amount.toLocaleString()}` : "No increment applied"}
                    </p>
                  </div>
                  <div className="h-10 w-10 rounded-full bg-purple-500/10 text-purple-600 dark:text-purple-400 flex items-center justify-center">
                    <ClockIcon className="h-5 w-5" />
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Current Salary Structure Breakdown */}
            <Card className="shadow-xs">
              <CardHeader className="py-3 px-4 border-b bg-muted/20">
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-sm font-semibold">Current Salary Structure Breakdown</CardTitle>
                    <CardDescription className="text-xs">
                      Official compensation components based on Bangladesh labor law formula
                    </CardDescription>
                  </div>
                  <Badge variant="secondary" className="text-xs font-mono font-medium">
                    Total Gross: ৳{Number(employee.gross_salary || 0).toLocaleString()}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="p-0">
                <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 divide-x divide-y sm:divide-y-0 text-xs">
                  <div className="p-3.5 space-y-1">
                    <span className="text-muted-foreground">Basic Salary</span>
                    <p className="text-sm font-bold text-foreground">৳{Number(employee.basic_salary || 0).toLocaleString()}</p>
                    <span className="text-[10px] text-muted-foreground">~50% of gross</span>
                  </div>
                  <div className="p-3.5 space-y-1">
                    <span className="text-muted-foreground">House Rent</span>
                    <p className="text-sm font-bold text-foreground">৳{Number(employee.house_rent || 0).toLocaleString()}</p>
                    <span className="text-[10px] text-muted-foreground">~50% of basic</span>
                  </div>
                  <div className="p-3.5 space-y-1">
                    <span className="text-muted-foreground">Medical Allowance</span>
                    <p className="text-sm font-bold text-foreground">৳{Number(employee.medical_allowance || 750).toLocaleString()}</p>
                    <span className="text-[10px] text-muted-foreground">Fixed: ৳750</span>
                  </div>
                  <div className="p-3.5 space-y-1">
                    <span className="text-muted-foreground">Food Allowance</span>
                    <p className="text-sm font-bold text-foreground">৳{Number(employee.food_allowance || 1250).toLocaleString()}</p>
                    <span className="text-[10px] text-muted-foreground">Fixed: ৳1,250</span>
                  </div>
                  <div className="p-3.5 space-y-1">
                    <span className="text-muted-foreground">Transport Allowance</span>
                    <p className="text-sm font-bold text-foreground">৳{Number(employee.transport_allowance || 450).toLocaleString()}</p>
                    <span className="text-[10px] text-muted-foreground">Fixed: ৳450</span>
                  </div>
                  <div className="p-3.5 space-y-1 bg-primary/5">
                    <span className="text-primary font-medium">Net Gross Pay</span>
                    <p className="text-sm font-extrabold text-primary">৳{Number(employee.gross_salary || 0).toLocaleString()}</p>
                    <span className="text-[10px] text-muted-foreground">Monthly Total</span>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Increment & Promotion History Data Table */}
            <Card className="shadow-xs">
              <CardHeader className="py-3 px-4 border-b bg-muted/20">
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
                  <div>
                    <CardTitle className="text-sm font-semibold">
                      Salary Increment & Promotion History {filters.year ? `(${filters.year})` : "(All Records)"}
                    </CardTitle>
                    <CardDescription className="text-xs">
                      Chronological progression of salary increments and designation changes for {employee.name_en}
                    </CardDescription>
                  </div>
                  <Badge variant="outline" className="text-xs self-start sm:self-auto">
                    {increments.length} Record{increments.length === 1 ? "" : "s"}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="p-0">
                <div className="overflow-x-auto">
                  <table className="w-full text-xs">
                    <thead>
                      <tr className="border-b bg-muted/40 text-muted-foreground">
                        <th className="px-4 py-2.5 text-left font-semibold w-12">#</th>
                        <th className="px-4 py-2.5 text-left font-semibold">Increment Date</th>
                        <th className="px-4 py-2.5 text-left font-semibold">Effective Date</th>
                        <th className="px-4 py-2.5 text-center font-semibold">Type</th>
                        <th className="px-4 py-2.5 text-left font-semibold">Designation Change</th>
                        <th className="px-4 py-2.5 text-right font-semibold">Previous Gross</th>
                        <th className="px-4 py-2.5 text-right font-semibold">Increment Amount</th>
                        <th className="px-4 py-2.5 text-right font-semibold">New Gross</th>
                        <th className="px-4 py-2.5 text-center font-semibold">Status</th>
                        <th className="px-4 py-2.5 text-left font-semibold">Remarks</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y">
                      {/* First Time Salary - initial joining salary before any increments */}
                      {employee && (summary.first_time_gross > 0 || summary.initial_gross > 0) && (
                        <tr className="bg-slate-50 dark:bg-slate-900/30 font-medium border-b-2 border-slate-200 dark:border-slate-700">
                          <td className="px-4 py-3 text-slate-600 dark:text-slate-400">0</td>
                          <td className="px-4 py-3 whitespace-nowrap text-slate-600">
                            {summary.first_time_date ? format(new Date(summary.first_time_date), "dd MMM yyyy") : employee.joining_date ? format(new Date(employee.joining_date), "dd MMM yyyy") : "-"}
                          </td>
                          <td className="px-4 py-3 whitespace-nowrap font-semibold text-slate-700 dark:text-slate-300">
                            {summary.first_time_date ? format(new Date(summary.first_time_date), "dd MMM yyyy") : "-"}
                          </td>
                          <td className="px-4 py-3 text-center">
                            <Badge className="bg-slate-600 hover:bg-slate-700 text-white text-xs">Initial</Badge>
                          </td>
                          <td className="px-4 py-3 text-xs text-muted-foreground">
                            {employee.designation_name || "-"} <span className="text-[10px]">(Joining)</span>
                          </td>
                          <td className="px-4 py-3 text-right font-mono text-slate-500">—</td>
                          <td className="px-4 py-3 text-right font-mono font-bold text-slate-600">—</td>
                          <td className="px-4 py-3 text-right font-mono font-bold text-slate-900 dark:text-slate-100">
                            ৳{Number(summary.first_time_gross || summary.initial_gross || 0).toLocaleString()}
                          </td>
                          <td className="px-4 py-3 text-center">
                            <Badge variant="outline" className="bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border-slate-300 text-xs">First Time</Badge>
                          </td>
                          <td className="px-4 py-3 max-w-[180px] truncate text-xs text-muted-foreground" title="First time salary at joining">
                            Joining Salary
                          </td>
                        </tr>
                      )}
                      {increments.length === 0 ? (
                        summary.first_time_gross > 0 || summary.initial_gross > 0 ? null : (
                        <tr>
                          <td colSpan={10} className="px-4 py-10 text-center text-muted-foreground">
                            <div className="flex flex-col items-center justify-center gap-1.5">
                              <TrendingUpIcon className="h-7 w-7 text-muted-foreground/40 stroke-1" />
                              <p className="font-medium text-xs">No increment records found for this employee</p>
                              <p className="text-[11px] text-muted-foreground">Applied increments will appear here</p>
                            </div>
                          </td>
                        </tr>
                        )
                      ) : (
                        increments.map((inc, i) => {
                          const prevDesig = inc.previous_designation?.name || "-"
                          const newDesig = inc.new_designation?.name
                          const hasDesigChange = newDesig && newDesig !== prevDesig

                          return (
                            <tr key={inc.id || i} className="hover:bg-muted/40 transition-colors">
                              <td className="px-4 py-3 text-muted-foreground">{i + 1}</td>
                              <td className="px-4 py-3 whitespace-nowrap">
                                {inc.increment_date ? format(new Date(inc.increment_date), "dd MMM yyyy") : "-"}
                              </td>
                              <td className="px-4 py-3 whitespace-nowrap font-medium text-foreground">
                                {inc.effective_date ? format(new Date(inc.effective_date), "dd MMM yyyy") : "-"}
                              </td>
                              <td className="px-4 py-3 text-center">
                                {renderTypeBadge(inc.increment_type)}
                              </td>
                              <td className="px-4 py-3">
                                {hasDesigChange ? (
                                  <div className="flex items-center gap-1.5 text-xs">
                                    <span className="text-muted-foreground">{prevDesig}</span>
                                    <ArrowRightIcon className="h-3 w-3 text-primary shrink-0" />
                                    <span className="font-semibold text-foreground">{newDesig}</span>
                                  </div>
                                ) : (
                                  <span className="text-muted-foreground">{prevDesig}</span>
                                )}
                              </td>
                              <td className="px-4 py-3 text-right font-mono text-muted-foreground">
                                ৳{Number(inc.previous_gross || 0).toLocaleString()}
                              </td>
                              <td className="px-4 py-3 text-right font-mono font-bold text-emerald-600 dark:text-emerald-400">
                                +৳{Number(inc.increment_amount || 0).toLocaleString()}
                              </td>
                              <td className="px-4 py-3 text-right font-mono font-bold text-foreground">
                                ৳{Number(inc.new_gross || 0).toLocaleString()}
                              </td>
                              <td className="px-4 py-3 text-center">
                                {getStatusBadge(inc.status)}
                              </td>
                              <td className="px-4 py-3 max-w-[180px] truncate text-muted-foreground" title={inc.remarks || inc.rejection_reason || ""}>
                                {inc.remarks || inc.rejection_reason || "-"}
                              </td>
                            </tr>
                          )
                        })
                      )}
                    </tbody>
                  </table>
                </div>
              </CardContent>
            </Card>
          </>
        )}
      </div>
    </div>
  )
}