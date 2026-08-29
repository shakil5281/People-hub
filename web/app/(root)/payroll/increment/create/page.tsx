"use client"

import * as React from "react"
import {
  TrendingUpIcon,
  ArrowLeftIcon,
  Loader2,
  SearchIcon,
  AlertCircle,
  SparklesIcon,
  AwardIcon,
  BuildingIcon,
  CalculatorIcon,
  CalendarIcon,
} from "lucide-react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { getApiBaseUrl } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  salaryIncrementApi,
  companyApi,
  departmentApi,
  sectionApi,
  designationApi,
  lineApi,
  groupApi,
  employeeApi,
} from "@/lib/api"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { DatePicker } from "@/components/ui/date-picker"
import { format } from "date-fns"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }

interface EmployeeRow {
  id: string
  employee_id: string
  punch_number?: string
  name_en: string
  name_bn?: string
  designation?: string
  designation_id?: string
  department?: string
  section?: string
  joining_date?: string
  gross_salary: number
  basic_salary: number
  house_rent?: number
  medical_allowance?: number
  transport_allowance?: number
  food_allowance?: number
  image_url?: string
}

const MONTHS = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
]
const currentYear = new Date().getFullYear()
const YEARS = Array.from({ length: 12 }, (_, i) => 2020 + i) // 2020..2031

const selectCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
const labelCls = "text-xs font-medium text-muted-foreground"

export default function CreateIncrementPage() {
  const router = useRouter()

  // Lookups
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [allDesignations, setAllDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])

  // Advance Filter states
  const [companyId, setCompanyId] = React.useState("")
  const [departmentId, setDepartmentId] = React.useState("")
  const [sectionId, setSectionId] = React.useState("")
  const [designationId, setDesignationId] = React.useState("")
  const [lineId, setLineId] = React.useState("")
  const [groupId, setGroupId] = React.useState("")
  const [incrementType, setIncrementType] = React.useState<string>("gov_policy")
  const [filterMonth, setFilterMonth] = React.useState<number>(new Date().getMonth() + 1) // default 8 for August
  const [filterYear, setFilterYear] = React.useState<number>(currentYear) // default 2026

  // Employee list & selection
  const [employees, setEmployees] = React.useState<EmployeeRow[]>([])
  const [loading, setLoading] = React.useState(false)
  const [selectedRows, setSelectedRows] = React.useState<EmployeeRow[]>([])

  // Modal Increment Configuration states
  const [modalIncrementType, setModalIncrementType] = React.useState<string>("gov_policy")
  const [calcType, setCalcType] = React.useState<string>("percentage")
  const [value, setValue] = React.useState<string>("9")
  const [newDesignationId, setNewDesignationId] = React.useState<string>("")
  const [incrementDate, setIncrementDate] = React.useState<Date | undefined>(new Date())
  const [effectiveDate, setEffectiveDate] = React.useState<Date | undefined>(new Date())
  const [remarks, setRemarks] = React.useState<string>("")

  const [applying, setApplying] = React.useState(false)
  const [dialogOpen, setDialogOpen] = React.useState(false)

  // Initialize master data
  React.useEffect(() => {
    const init = async () => {
      try {
        const [cRes, dRes, gRes, desRes] = await Promise.all([
          companyApi.list({ limit: "100" }),
          departmentApi.list({ limit: "100" }),
          groupApi.list({ limit: "100" }),
          designationApi.list(undefined, { limit: "200" }),
        ])
        const cList = Array.isArray(cRes.data?.data) ? cRes.data.data : (Array.isArray(cRes.data) ? cRes.data : [])
        setCompanies(cList)
        if (cList.length > 0 && !companyId) {
          setCompanyId(cList[0].id)
        }

        if (Array.isArray(dRes.data?.data)) setDepartments(dRes.data.data)
        else if (Array.isArray(dRes.data)) setDepartments(dRes.data)

        if (Array.isArray(gRes.data?.data)) setGroups(gRes.data.data)
        else if (Array.isArray(gRes.data)) setGroups(gRes.data)

        if (Array.isArray(desRes.data?.data)) {
          setAllDesignations(desRes.data.data)
          setDesignations(desRes.data.data)
        } else if (Array.isArray(desRes.data)) {
          setAllDesignations(desRes.data)
          setDesignations(desRes.data)
        }
      } catch {
        toast.error("Failed to load initial form data")
      }
    }
    init()
  }, [])

  const fetchSections = React.useCallback(async (deptId: string) => {
    try {
      const { data } = await sectionApi.list(deptId || undefined, { limit: "100" })
      setSections(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch {
      setSections([])
    }
  }, [])

  const fetchDesignations = React.useCallback(async (secId: string) => {
    try {
      const { data } = await designationApi.list(secId || undefined, { limit: "100" })
      setDesignations(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : allDesignations)
    } catch {
      setDesignations(allDesignations)
    }
  }, [allDesignations])

  const fetchLines = React.useCallback(async (secId: string) => {
    try {
      const { data } = await lineApi.list(secId || undefined, { limit: "100" })
      setLines(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch {
      setLines([])
    }
  }, [])

  React.useEffect(() => {
    if (!departmentId) {
      setSections([])
      setDesignations(allDesignations)
      setLines([])
      setSectionId("")
      setDesignationId("")
      setLineId("")
      return
    }
    fetchSections(departmentId)
    setSectionId("")
    setDesignationId("")
    setLineId("")
  }, [departmentId, fetchSections, allDesignations])

  React.useEffect(() => {
    if (!sectionId) {
      setDesignations(allDesignations)
      setLines([])
      setDesignationId("")
      setLineId("")
      return
    }
    fetchDesignations(sectionId)
    fetchLines(sectionId)
    setDesignationId("")
    setLineId("")
  }, [sectionId, fetchDesignations, fetchLines, allDesignations])

  // Sync modal state when incrementType in filter changes
  React.useEffect(() => {
    setModalIncrementType(incrementType)
    if (incrementType === "gov_policy") {
      setCalcType("percentage")
      setValue("9") // 9% as per policy update
    } else if (incrementType === "promotion") {
      setCalcType("fixed")
      setValue("0")
    } else if (incrementType === "promotion_with_increment") {
      setCalcType("percentage")
      setValue("10")
    } else {
      setCalcType("percentage")
      setValue("5")
    }
  }, [incrementType])

  const handleSearch = async () => {
    if (!companyId) {
      toast.error("Please select a Company first")
      return
    }
    setLoading(true)
    try {
      const params: Record<string, string> = {
        company_id: companyId,
        limit: "1000",
        status: "active",
      }
      if (departmentId) params.department_id = departmentId
      if (sectionId) params.section_id = sectionId
      if (designationId) params.designation_id = designationId
      if (lineId) params.line_id = lineId
      if (groupId) params.group_id = groupId

      // Anniversary Joining Date filter (e.g. Month = Aug, Year = 2026 -> finds joined August 2024, August 2025...)
      if (filterMonth > 0) {
        params.joining_month = String(filterMonth)
        params.joining_year = String(filterYear)
        params.joining_after_year = "2024"
      }

      const res = await employeeApi.list(params)
      const data = res.data?.data || res.data || []
      setEmployees(Array.isArray(data) ? data : [])
      setSelectedRows([])
      if (Array.isArray(data) && data.length === 0) {
        toast.info("No active employees found matching the specified filters")
      }
    } catch {
      toast.error("Failed to fetch employees")
    } finally {
      setLoading(false)
    }
  }

  const handleOpenDialog = () => {
    if (selectedRows.length === 0) {
      toast.error("Please select at least one employee from the list")
      return
    }
    setModalIncrementType(incrementType)
    setDialogOpen(true)
  }

  const handleApply = async () => {
    if (!incrementDate || !effectiveDate) {
      toast.error("Increment Date and Effective Date are required")
      return
    }
    if (selectedRows.length === 0) {
      toast.error("Please select at least one employee")
      return
    }

    const isPromotion = modalIncrementType === "promotion" || modalIncrementType === "promotion_with_increment"
    if (isPromotion && !newDesignationId) {
      toast.error("Please select a New Designation for promotion")
      return
    }

    const numVal = Number(value) || 0
    if (modalIncrementType !== "promotion" && numVal <= 0) {
      toast.error("Please enter a valid increment value greater than 0")
      return
    }

    setApplying(true)
    try {
      const payload: Record<string, unknown> = {
        company_id: companyId,
        increment_type: modalIncrementType,
        calculation_type: calcType,
        increment_date: format(incrementDate, "yyyy-MM-dd"),
        effective_date: format(effectiveDate, "yyyy-MM-dd"),
        value: numVal,
        employee_ids: selectedRows.map((r) => r.employee_id),
        remarks: remarks.trim(),
      }
      if (newDesignationId) payload.new_designation_id = newDesignationId
      if (departmentId) payload.department_id = departmentId
      if (sectionId) payload.section_id = sectionId
      if (designationId) payload.designation_id = designationId
      if (lineId) payload.line_id = lineId
      if (groupId) payload.group_id = groupId

      const { data: res } = await salaryIncrementApi.bulkApply(payload)
      toast.success(`${res.message || "Increment submitted successfully"}`)
      setDialogOpen(false)
      router.push("/payroll/increment")
    } catch (err: unknown) {
      const msg =
        typeof err === "object" && err !== null && "response" in err
          ? (err as any).response?.data?.error || "Failed to apply increment"
          : "Failed to apply increment"
      toast.error(msg)
    } finally {
      setApplying(false)
    }
  }

  // Calculate estimated increment for an individual employee
  const calculateEstimatedInc = (emp: EmployeeRow): number => {
    const numVal = Number(value) || 0
    const gross = emp.gross_salary || 0
    let basic = emp.basic_salary || 0
    let house = emp.house_rent || 0

    if ((basic <= 0 || house <= 0) && gross > 0) {
      const fixedAllowances = (emp.medical_allowance || 750) + (emp.transport_allowance || 450) + (emp.food_allowance || 1250)
      const core = gross > fixedAllowances ? gross - fixedAllowances : 0
      basic = Math.round(core / 1.5)
      house = core - basic
    }

    const base = basic + house

    if (modalIncrementType === "gov_policy") {
      const rate = numVal > 0 ? numVal : 9
      return Math.round((base * rate) / 100)
    }
    if (calcType === "percentage") {
      return Math.round((base * numVal) / 100)
    }
    return numVal
  }

  const columns: ColumnDef<EmployeeRow>[] = [
    {
      accessorKey: "photo",
      header: "Photo",
      cell: ({ row }) => {
        const img = row.original.image_url
        const baseUrl = getApiBaseUrl().replace("/api/v1", "")
        return (
          <Avatar className="h-8 w-8">
            <AvatarImage
              src={img ? (img.startsWith("http") ? img : `${baseUrl}/${img.replace(/^\//, "")}`) : ""}
            />
            <AvatarFallback>{row.original.name_en?.charAt(0)}</AvatarFallback>
          </Avatar>
        )
      },
    },
    {
      accessorKey: "employee_id",
      header: "Emp ID",
      cell: ({ row }) => (
        <span className="font-mono font-medium text-xs sm:text-sm">{row.original.employee_id}</span>
      ),
    },
    {
      accessorKey: "name_en",
      header: "Name",
      cell: ({ row }) => (
        <div>
          <div className="font-medium text-foreground">{row.original.name_en}</div>
          {row.original.name_bn && (
            <div className="text-xs text-muted-foreground">{row.original.name_bn}</div>
          )}
        </div>
      ),
    },
    {
      accessorKey: "joining_date",
      header: "Joining Date",
      cell: ({ row }) => {
        const jd = row.original.joining_date
        return jd ? (
          <span className="text-xs font-mono font-medium text-foreground">
            {format(new Date(jd), "dd-MMM-yyyy")}
          </span>
        ) : (
          "-"
        )
      },
    },
    {
      accessorKey: "designation",
      header: "Current Designation",
      cell: ({ row }) => row.original.designation || "-",
    },
    {
      accessorKey: "department",
      header: "Department / Section",
      cell: ({ row }) => (
        <div className="text-xs text-muted-foreground">
          <div>{row.original.department || "-"}</div>
          {row.original.section && <div>{row.original.section}</div>}
        </div>
      ),
    },
    {
      accessorKey: "gross_salary",
      header: "Gross Salary",
      cell: ({ row }) => (
        <span className="font-mono font-medium">৳{row.original.gross_salary?.toLocaleString() || 0}</span>
      ),
    },
    {
      accessorKey: "projected_new_gross",
      header: "Projected Increment",
      cell: ({ row }) => {
        const estInc = calculateEstimatedInc(row.original)
        const newGross = (row.original.gross_salary || 0) + estInc
        return (
          <div className="text-xs">
            <span className="text-emerald-600 dark:text-emerald-400 font-semibold">
              +৳{estInc.toLocaleString()}
            </span>
            <div className="text-muted-foreground font-mono">
              New: ৳{newGross.toLocaleString()}
            </div>
          </div>
        )
      },
    },
  ]

  const selectedMonthName = MONTHS[filterMonth - 1] || "Selected Month"

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      {/* Header Bar */}
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => router.back()} className="h-9 w-9">
            <ArrowLeftIcon className="h-4 w-4" />
          </Button>
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <TrendingUpIcon className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight md:text-3xl">Apply Bulk Increment</h1>
            <p className="text-muted-foreground text-xs sm:text-sm">
              Filter eligible employees by joining anniversary and apply Regular Increment, Gov Policy (9%), or Promotion
            </p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6 flex flex-col gap-6">
        {/* Advance Filter Section */}
        <div className="rounded-lg border bg-card p-5 md:p-6 space-y-5 shadow-sm">
          <div className="flex items-center justify-between border-b pb-3">
            <div className="flex items-center gap-2 font-semibold text-base">
              <CalculatorIcon className="h-4 w-4 text-primary" />
              Advance Filter Section
            </div>
            <Button onClick={handleSearch} disabled={loading || !companyId} size="sm">
              {loading ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <SearchIcon className="mr-2 h-4 w-4" />
              )}
              Search Employees
            </Button>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {/* Company */}
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Company *</label>
              <select
                value={companyId}
                onChange={(e) => setCompanyId(e.target.value)}
                className={selectCls}
              >
                <option value="">-- Select Company --</option>
                {companies.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.company_name_en}
                  </option>
                ))}
              </select>
            </div>

            {/* Department */}
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Department</label>
              <select
                value={departmentId}
                onChange={(e) => setDepartmentId(e.target.value)}
                className={selectCls}
              >
                <option value="">All Departments</option>
                {departments.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name}
                  </option>
                ))}
              </select>
            </div>

            {/* Section */}
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Section</label>
              <select
                value={sectionId}
                onChange={(e) => setSectionId(e.target.value)}
                className={selectCls}
              >
                <option value="">All Sections</option>
                {sections.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}
                  </option>
                ))}
              </select>
            </div>

            {/* Designation */}
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Designation</label>
              <select
                value={designationId}
                onChange={(e) => setDesignationId(e.target.value)}
                className={selectCls}
              >
                <option value="">All Designations</option>
                {designations.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name}
                  </option>
                ))}
              </select>
            </div>

            {/* Line */}
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Line</label>
              <select
                value={lineId}
                onChange={(e) => setLineId(e.target.value)}
                className={selectCls}
              >
                <option value="">All Lines</option>
                {lines.map((l) => (
                  <option key={l.id} value={l.id}>
                    {l.name}
                  </option>
                ))}
              </select>
            </div>

            {/* Group */}
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Group</label>
              <select
                value={groupId}
                onChange={(e) => setGroupId(e.target.value)}
                className={selectCls}
              >
                <option value="">All Groups</option>
                {groups.map((g) => (
                  <option key={g.id} value={g.id}>
                    {g.name}
                  </option>
                ))}
              </select>
            </div>

            {/* Month Filter */}
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Joining Month *</label>
              <select
                value={filterMonth}
                onChange={(e) => setFilterMonth(Number(e.target.value))}
                className={selectCls}
              >
                {MONTHS.map((mName, idx) => (
                  <option key={mName} value={idx + 1}>
                    {mName}
                  </option>
                ))}
              </select>
            </div>

            {/* Year Filter */}
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Increment Target Year *</label>
              <select
                value={filterYear}
                onChange={(e) => setFilterYear(Number(e.target.value))}
                className={selectCls}
              >
                {YEARS.map((yr) => (
                  <option key={yr} value={yr}>
                    {yr}
                  </option>
                ))}
              </select>
            </div>

            {/* Increment Type Selector */}
            <div className="flex flex-col gap-1.5 sm:col-span-2 lg:col-span-4">
              <label className="text-xs font-semibold text-primary flex items-center gap-1.5">
                <SparklesIcon className="h-3.5 w-3.5" />
                Increment Type *
              </label>
              <select
                value={incrementType}
                onChange={(e) => setIncrementType(e.target.value)}
                className={`${selectCls} font-medium border-primary/40 bg-primary/5`}
              >
                <option value="gov_policy">As per gov policy (9% on Basic + House Rent)</option>
                <option value="increment">Increment (Regular Salary Increment)</option>
                <option value="promotion">Promotion (Designation Change)</option>
                <option value="promotion_with_increment">Promotion with increment (Designation + Salary Increment)</option>
              </select>
            </div>
          </div>

          {/* Context Banner */}
          {incrementType === "gov_policy" && (
            <div className="rounded-md bg-emerald-500/10 border border-emerald-500/20 p-3 text-xs text-emerald-800 dark:text-emerald-300 flex items-start gap-2.5">
              <BuildingIcon className="h-4 w-4 shrink-0 text-emerald-600 mt-0.5" />
              <div>
                <div className="font-semibold text-sm">
                  Government Policy Increment Rule (9% on Basic + House Rent)
                </div>
                <div className="mt-1 text-emerald-700 dark:text-emerald-300/90">
                  Filtering employees who joined in <b>{selectedMonthName}</b> after <b>2024</b> (e.g. {filterYear > 2024 ? `01-${String(filterMonth).padStart(2, '0')}-2024 to 31-${String(filterMonth).padStart(2, '0')}-${filterYear - 1}` : `August 2024 onwards`}).
                </div>
              </div>
            </div>
          )}

          {(incrementType === "promotion" || incrementType === "promotion_with_increment") && (
            <div className="rounded-md bg-amber-500/10 border border-amber-500/20 p-3 text-xs text-amber-800 dark:text-amber-300 flex items-center gap-2">
              <AwardIcon className="h-4 w-4 shrink-0 text-amber-600" />
              <span>
                <b>Promotion Workflow</b>: Allows designating a new promotional title. When approved, employee profile designation will update automatically.
              </span>
            </div>
          )}
        </div>

        {/* Data Table Section */}
        <div className="flex flex-col gap-4">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
            <div>
              <h2 className="text-lg font-semibold">Matched Employees ({employees.length})</h2>
              <p className="text-xs text-muted-foreground">
                Showing active employees joined in {selectedMonthName} after 2024
              </p>
            </div>
            {selectedRows.length > 0 && (
              <Button onClick={handleOpenDialog} className="bg-primary hover:bg-primary/90">
                <SparklesIcon className="mr-2 h-4 w-4" />
                Apply {incrementType.replace(/_/g, " ")} to {selectedRows.length} Selected
              </Button>
            )}
          </div>

          {employees.length === 0 && !loading ? (
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                No active employees found matching the filters (Joined in {selectedMonthName} between 2024 and {filterYear - 1}). Please adjust filters and click <b>Search Employees</b>.
              </AlertDescription>
            </Alert>
          ) : (
            <DataTable
              data={employees}
              columns={columns}
              loading={loading}
              enableSelection={true}
              onSelectionChange={setSelectedRows}
            />
          )}
        </div>
      </div>

      {/* Increment Configuration Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <TrendingUpIcon className="h-5 w-5 text-primary" />
              Confirm & Apply Bulk Increment
            </DialogTitle>
            <DialogDescription>
              Applying to <b>{selectedRows.length}</b> selected employee(s). Review calculation rules and effective dates.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-2">
            {/* Increment Type in Dialog */}
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Increment Type</label>
              <select
                value={modalIncrementType}
                onChange={(e) => setModalIncrementType(e.target.value)}
                className={selectCls}
              >
                <option value="gov_policy">As per gov policy (9% on Basic + House Rent)</option>
                <option value="increment">Increment (Regular Salary Increment)</option>
                <option value="promotion">Promotion</option>
                <option value="promotion_with_increment">Promotion with increment</option>
              </select>
            </div>

            {/* Target Designation for Promotion types */}
            {(modalIncrementType === "promotion" || modalIncrementType === "promotion_with_increment") && (
              <div className="flex flex-col gap-1.5 rounded-lg border border-amber-500/30 bg-amber-500/5 p-3">
                <label className="text-xs font-semibold text-amber-900 dark:text-amber-300">
                  Target Promotional Designation *
                </label>
                <select
                  value={newDesignationId}
                  onChange={(e) => setNewDesignationId(e.target.value)}
                  className={selectCls}
                >
                  <option value="">-- Select New Designation --</option>
                  {allDesignations.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.name}
                    </option>
                  ))}
                </select>
                <p className="text-[11px] text-muted-foreground mt-1">
                  Upon manager approval, this designation will replace the employee&apos;s current designation.
                </p>
              </div>
            )}

            {/* Calculation Type & Value */}
            {modalIncrementType === "gov_policy" ? (
              <div className="rounded-lg bg-muted/40 p-3 border space-y-1.5">
                <div className="flex justify-between text-xs">
                  <span className="font-medium text-muted-foreground">Government Formula:</span>
                  <span className="font-semibold text-emerald-600">(Basic + House Rent) × 9%</span>
                </div>
                <div className="flex items-center gap-2 pt-1">
                  <label className="text-xs text-muted-foreground">Policy Increment Rate (%):</label>
                  <input
                    type="number"
                    value={value}
                    onChange={(e) => setValue(e.target.value)}
                    placeholder="9"
                    className="flex h-8 w-24 rounded-md border border-input bg-background px-2 text-xs font-mono"
                  />
                  <span className="text-xs text-muted-foreground">%</span>
                </div>
              </div>
            ) : modalIncrementType === "promotion" ? (
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Optional Salary Adjustment (BDT)</label>
                <input
                  type="number"
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  placeholder="0 (Leave 0 for no salary change)"
                  className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm font-mono"
                />
              </div>
            ) : (
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1.5">
                  <label className={labelCls}>Calculation Method</label>
                  <select
                    value={calcType}
                    onChange={(e) => setCalcType(e.target.value)}
                    className={selectCls}
                  >
                    <option value="percentage">Percentage (%)</option>
                    <option value="fixed">Fixed Amount (BDT)</option>
                  </select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className={labelCls}>
                    {calcType === "percentage" ? "Percentage Rate (%)" : "Amount (BDT)"}
                  </label>
                  <input
                    type="number"
                    value={value}
                    onChange={(e) => setValue(e.target.value)}
                    placeholder={calcType === "percentage" ? "e.g. 10" : "e.g. 2000"}
                    className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm font-mono"
                  />
                </div>
              </div>
            )}

            {/* Date Pickers */}
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Increment Date *</label>
                <DatePicker value={incrementDate} onChange={setIncrementDate} placeholder="Select date" />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Effective Date *</label>
                <DatePicker value={effectiveDate} onChange={setEffectiveDate} placeholder="Select date" />
              </div>
            </div>

            {/* Remarks */}
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Remarks / Justification</label>
              <input
                type="text"
                value={remarks}
                onChange={(e) => setRemarks(e.target.value)}
                placeholder="Optional notes..."
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleApply} disabled={applying}>
              {applying && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Submit Increment Request
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
