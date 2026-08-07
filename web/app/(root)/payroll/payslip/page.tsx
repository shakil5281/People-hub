"use client"

import * as React from "react"
import { Loader2, FileDownIcon, FileTextIcon, ChevronLeftIcon, ChevronRightIcon, ReceiptIcon } from "lucide-react"
import { toast } from "sonner"
import { salaryApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

interface Emp {
  employee_id: string
  name_en: string
  name_bn?: string
  grade?: string
  employee_type?: string
  account_number?: string
  account_type?: string
  nid?: string
  joining_date?: string
  designation_ref?: { name: string; name_bn?: string }
  department?: { name: string; name_bn?: string }
  section_ref?: { name: string; name_bn?: string }
  shift?: { name: string }
}

interface PayslipRecord {
  id: string
  company?: {
    company_name_en?: string
    company_name_bn?: string
    address_en?: string
    address_bn?: string
    phone?: string
    email?: string
  }
  employee: Emp
  basic_salary: number
  house_rent: number
  medical_allowance: number
  transport_allowance: number
  food_allowance: number
  other_allowance: number
  gross_salary: number
  provident_fund: number
  tax: number
  loan_deduction: number
  advance_deduction: number
  absent_deduction: number
  other_deduction: number
  total_deductions: number
  overtime_hours: number
  overtime_rate: number
  overtime_amount: number
  attendance_bonus: number
  net_salary: number
  present_days: number
  absent_days: number
  late_days: number
  leave_days: number
  holiday_days: number
  weekend_days: number
  total_days: number
  status: string
}

interface PaginatedResponse {
  salaries: PayslipRecord[]
  total: number
  page: number
  limit: number
  total_pages: number
  month: number
  year: number
}

type ExportKind = "excel" | "pdf"

const MONTHS = ["January","February","March","April","May","June","July","August","September","October","November","December"]
const currentYear = new Date().getFullYear()
const currentMonth = new Date().getMonth()
const YEARS = Array.from({length:10},(_,i)=>currentYear-5+i)

const fmt = (n: number) => (n || 0).toLocaleString()

const orDash = (v?: string | null) => (v && v.trim() ? v : "-")

const shortPayrollNo = (id: string) => (id || "").toUpperCase().slice(0, 10) || "-"

const paymentMethodName = (acctType?: string) => {
  switch ((acctType || "").toLowerCase()) {
    case "cash": return "Cash"
    case "mobile": case "mobile_banking": case "bkash": case "nagad": case "rocket": return "Mobile Banking"
    case "cheque": case "check": return "Cheque"
    default: return "Bank Transfer"
  }
}

function formatDate(v?: string) {
  if (!v) return "-"
  const d = new Date(v)
  if (isNaN(d.getTime())) return "-"
  return d.toLocaleDateString("en-GB", { day: "2-digit", month: "short", year: "numeric" })
}

function SectionTitle({ title }: { title: string }) {
  return (
    <div className="flex items-center justify-between border border-slate-300 bg-slate-50 px-2 py-1">
      <span className="text-[11px] font-bold uppercase tracking-wide text-blue-900">{title}</span>
    </div>
  )
}

function FieldGrid({ fields, cols = 2 }: { fields: [string, string][]; cols?: number }) {
  return (
    <div className="grid border-x border-b border-slate-300" style={{ gridTemplateColumns: `repeat(${cols}, minmax(0,1fr))` }}>
      {fields.map(([label, value], i) => (
        <div key={i} className="flex items-start gap-1 border-b border-slate-200 px-1.5 py-1 text-[11px]">
          <span className="w-[46%] shrink-0 font-semibold text-slate-500">{label}</span>
          <span className="min-w-0 flex-1 break-words font-medium text-slate-800">{value}</span>
        </div>
      ))}
    </div>
  )
}

function MoneyTable({ title, rows, total, totalLabel }: { title: string; rows: [string, string][]; total: string; totalLabel: string }) {
  return (
    <div className="min-w-0 flex-1">
      <SectionTitle title={title} />
      <div className="border-x border-b border-slate-300">
        {rows.map(([label, amount], i) => (
          <div key={i} className="flex items-center justify-between gap-2 border-b border-slate-200 px-1.5 py-[3px] text-[11px] last:border-b-0">
            <span className="min-w-0 truncate text-slate-700">{label}</span>
            <span className="shrink-0 font-medium text-slate-800">{amount}</span>
          </div>
        ))}
      </div>
      <div className="flex items-center justify-between gap-2 border border-slate-300 bg-slate-100 px-1.5 py-1 text-[11px]">
        <span className="font-bold text-slate-800">{totalLabel}</span>
        <span className="font-bold text-slate-900">{total}</span>
      </div>
    </div>
  )
}

function PayslipCard({ s, month, year }: { s: PayslipRecord; month: number; year: number }) {
  const emp = s.employee || {} as Emp
  const company = s.company
  const payrollNo = shortPayrollNo(s.id)
  const printDate = new Date().toLocaleDateString("en-GB", { day: "2-digit", month: "short", year: "numeric" })

  const employeeInfo: [string, string][] = [
    ["Employee ID", orDash(emp.employee_id)],
    ["Name", orDash(emp.name_en)],
    ["Department", orDash(emp.department?.name)],
    ["Section", orDash(emp.section_ref?.name)],
    ["Designation", orDash(emp.designation_ref?.name)],
    ["Grade", orDash(emp.grade)],
    ["Shift", orDash(emp.shift?.name)],
    ["Joining Date", formatDate(emp.joining_date)],
    ["Employment Type", orDash(emp.employee_type)],
    ["Account No.", orDash(emp.account_number)],
    ["NID", orDash(emp.nid)],
  ]

  const attendance: [string, string][] = [
    ["Working Days", String(s.total_days ?? 0)],
    ["Weekend", String(s.weekend_days ?? 0)],
    ["Holiday", String(s.holiday_days ?? 0)],
    ["Present", String(s.present_days ?? 0)],
    ["Absent", String(s.absent_days ?? 0)],
    ["Leave", String(s.leave_days ?? 0)],
    ["Late", String(s.late_days ?? 0)],
    ["OT Hours", String(Math.floor(s.overtime_hours))],
  ]

  const earnings: [string, string][] = [
    ["Basic Salary", fmt(s.basic_salary)],
    ["House Rent", fmt(s.house_rent)],
    ["Medical", fmt(s.medical_allowance)],
    ["Transport", fmt(s.transport_allowance)],
    ["Food", fmt(s.food_allowance)],
    ["Other Allowance", fmt(s.other_allowance)],
    ["Attendance Bonus", fmt(s.attendance_bonus)],
    ["Overtime", fmt(s.overtime_amount)],
  ]

  const deductions: [string, string][] = [
    ["Tax", fmt(s.tax)],
    ["PF", fmt(s.provident_fund)],
    ["Loan", fmt(s.loan_deduction)],
    ["Advance", fmt(s.advance_deduction)],
    ["Absent Deduction", fmt(s.absent_deduction)],
    ["Other Deduction", fmt(s.other_deduction)],
  ]

  const summary: [string, string][] = [
    ["Gross Salary", fmt(s.gross_salary)],
    ["Total Earnings", fmt(s.gross_salary)],
    ["Total Deduction", fmt(s.total_deductions)],
  ]

  return (
    <Card className="overflow-hidden border-slate-300 shadow-sm">
      {/* Header band */}
      <div className="bg-slate-900 px-3 py-2 text-white">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <div className="text-sm font-bold leading-tight text-white">{orDash(company?.company_name_en)}</div>
            <div className="mt-0.5 line-clamp-2 text-[10px] text-slate-300">{orDash(company?.address_en)}</div>
            {(company?.phone || company?.email) && (
              <div className="mt-0.5 truncate text-[10px] text-slate-300">
                {[company?.phone, company?.email].filter(Boolean).join("  ")}
              </div>
            )}
          </div>
          <div className="shrink-0 text-right">
            <div className="text-base font-bold leading-tight tracking-wide text-white">PAYSLIP</div>
            <div className="mt-0.5 text-[9px] font-bold tracking-widest text-amber-500">OFFICE COPY</div>
            <div className="mt-1 text-[10px] text-slate-300">
              <span className="text-slate-400">Payroll Month: </span>{MONTHS[month]} {year}
            </div>
            <div className="text-[10px] text-slate-300">
              <span className="text-slate-400">Payroll No: </span>{payrollNo}
            </div>
            <div className="text-[10px] text-slate-300">
              <span className="text-slate-400">Print Date: </span>{printDate}
            </div>
          </div>
        </div>
      </div>

      <CardContent className="space-y-2 p-3">
        {/* Employee Information */}
        <div>
          <SectionTitle title="Employee Information" />
          <FieldGrid fields={employeeInfo} cols={2} />
        </div>

        {/* Attendance */}
        <div>
          <SectionTitle title="Attendance" />
          <FieldGrid fields={attendance} cols={4} />
        </div>

        {/* Earnings + Deductions side by side */}
        <div className="flex gap-2">
          <MoneyTable title="Earnings" rows={earnings} total={fmt(s.gross_salary)} totalLabel="Total Earnings" />
          <MoneyTable title="Deductions" rows={deductions} total={fmt(s.total_deductions)} totalLabel="Total Deduction" />
        </div>

        {/* Summary */}
        <div>
          {summary.map(([label, value], i) => (
            <div key={i} className="flex items-center justify-between border border-slate-300 bg-slate-50 px-2 py-1 text-[11px]">
              <span className="font-semibold text-blue-900">{label}</span>
              <span className="font-semibold text-slate-900">{value}</span>
            </div>
          ))}
        </div>

        {/* Net salary green box */}
        <div className="flex items-center justify-between bg-green-700 px-3 py-1.5 text-white">
          <span className="text-sm font-bold">Net Salary</span>
          <span className="text-sm font-bold">BDT {fmt(s.net_salary)}</span>
        </div>

        {/* Signatures */}
        <div className="grid grid-cols-4 border border-slate-300 mt-2">
          {["Prepared By", "Checked By", "Approved By", "Employee Signature"].map((label, i) => (
            <div key={i} className="flex h-12 items-end justify-center border-r border-slate-300 pb-1 text-[9px] text-slate-600 last:border-r-0">
              {label}
            </div>
          ))}
        </div>

        {/* Generated / confidential */}
        <div className="flex items-center justify-between text-[9px] text-slate-500 mt-1">
          <span>System Generated</span>
          <span className="font-medium uppercase">Confidential Document</span>
        </div>
      </CardContent>
    </Card>
  )
}

export default function PaySlipPage() {
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])

  const [companyId, setCompanyId] = React.useState("")
  const [departmentId, setDepartmentId] = React.useState("")
  const [sectionId, setSectionId] = React.useState("")
  const [designationId, setDesignationId] = React.useState("")
  const [lineId, setLineId] = React.useState("")
  const [groupId, setGroupId] = React.useState("")
  const [shiftId, setShiftId] = React.useState("")
  const [employeeId, setEmployeeId] = React.useState("")
  const [month, setMonth] = React.useState(currentMonth)
  const [year, setYear] = React.useState(currentYear)

  const [singlePayslip, setSinglePayslip] = React.useState<PayslipRecord | null>(null)
  const [paginated, setPaginated] = React.useState<PaginatedResponse | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [exporting, setExporting] = React.useState<ExportKind | null>(null)
  const [lang, setLang] = React.useState<"en" | "bn">("en")
  const [searched, setSearched] = React.useState(false)
  const [page, setPage] = React.useState(1)

  const fetchSections = React.useCallback(async (deptId: string) => {
    try {
      const { data } = await sectionApi.list(deptId || undefined)
      setSections(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setSections([]) }
  }, [])

  const fetchDesignations = React.useCallback(async (secId: string) => {
    try {
      const { data } = await designationApi.list(secId || undefined)
      setDesignations(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setDesignations([]) }
  }, [])

  const fetchLines = React.useCallback(async (secId: string) => {
    try {
      const { data } = await lineApi.list(secId || undefined)
      setLines(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setLines([]) }
  }, [])

  React.useEffect(() => {
    fetchSections(departmentId)
    setSectionId("")
    setDesignationId("")
    setLineId("")
    setDesignations([])
    setLines([])
  }, [departmentId, fetchSections])

  React.useEffect(() => {
    fetchDesignations(sectionId)
    fetchLines(sectionId)
    setDesignationId("")
    setLineId("")
  }, [sectionId, fetchDesignations, fetchLines])

  React.useEffect(() => {
    const init = async () => {
      const [cRes, dRes, gRes, sRes] = await Promise.all([
        companyApi.list({ limit: "100" }),
        departmentApi.list({ limit: "100" }),
        groupApi.list({ limit: "100" }),
        shiftApi.list({ limit: "100" }),
      ])
      if (Array.isArray(cRes.data?.data) && cRes.data.data.length > 0) {
        setCompanies(cRes.data.data)
        setCompanyId(cRes.data.data[0].id)
      }
      if (Array.isArray(dRes.data?.data)) setDepartments(dRes.data.data)
      if (Array.isArray(gRes.data?.data)) setGroups(gRes.data.data)
      else if (Array.isArray(gRes.data)) setGroups(gRes.data)
      if (Array.isArray(sRes.data?.data)) setShifts(sRes.data.data)
      else if (Array.isArray(sRes.data)) setShifts(sRes.data)
    }
    init()
  }, [])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en })) },
    { key: "month", label: "Month", type: "select", options: MONTHS.map((n, i) => ({ value: String(i), label: n })) },
    { key: "year", label: "Year", type: "select", options: YEARS.map((y) => ({ value: String(y), label: String(y) })) },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    { key: "section_id", label: "Section", type: "select", options: sections.map((s) => ({ value: s.id, label: s.name })), disabled: !departmentId },
    { key: "designation_id", label: "Designation", type: "select", options: designations.map((d) => ({ value: d.id, label: d.name })), disabled: !sectionId },
    { key: "line_id", label: "Line", type: "select", options: lines.map((l) => ({ value: l.id, label: l.name })), disabled: !sectionId },
    { key: "group_id", label: "Group", type: "select", options: groups.map((g) => ({ value: g.id, label: g.name })) },
    { key: "shift_id", label: "Shift", type: "select", options: shifts.map((s) => ({ value: s.id, label: s.name })) },
    { key: "employee_id", label: "Employee ID", type: "text", placeholder: "Enter employee code..." },
  ], [companies, departments, sections, designations, lines, groups, shifts, departmentId, sectionId, month, year])

  const filters: Record<string, string> = {
    company_id: companyId,
    month: String(month),
    year: String(year),
    department_id: departmentId,
    section_id: sectionId,
    designation_id: designationId,
    line_id: lineId,
    group_id: groupId,
    shift_id: shiftId,
    employee_id: employeeId,
  }

  const handleFilterChange = React.useCallback((key: string, value: string) => {
    switch (key) {
      case "company_id": setCompanyId(value); break
      case "month": setMonth(Number(value)); break
      case "year": setYear(Number(value)); break
      case "department_id": setDepartmentId(value); break
      case "section_id": setSectionId(value); break
      case "designation_id": setDesignationId(value); break
      case "line_id": setLineId(value); break
      case "group_id": setGroupId(value); break
      case "shift_id": setShiftId(value); break
      case "employee_id": setEmployeeId(value); break
    }
  }, [])

  const buildParams = React.useCallback(() => {
    const params: Record<string, string> = {
      month: String(month + 1),
      year: String(year),
    }
    if (employeeId) params.employee_id = employeeId
    if (companyId) params.company_id = companyId
    if (departmentId) params.department_id = departmentId
    if (sectionId) params.section_id = sectionId
    if (designationId) params.designation_id = designationId
    if (lineId) params.line_id = lineId
    if (groupId) params.group_id = groupId
    if (shiftId) params.shift_id = shiftId
    return params
  }, [companyId, month, year, departmentId, sectionId, designationId, lineId, groupId, shiftId, employeeId])

  const loadPage = React.useCallback(async (pageNum: number) => {
    setPage(pageNum)
    setLoading(true)
    setSearched(true)
    try {
      const bp = buildParams()
      const params: Record<string, string> = { month: bp.month, year: bp.year, page: String(pageNum) }
      if (bp.company_id) params.company_id = bp.company_id
      if (bp.department_id) params.department_id = bp.department_id
      if (bp.section_id) params.section_id = bp.section_id
      if (bp.designation_id) params.designation_id = bp.designation_id
      if (bp.line_id) params.line_id = bp.line_id
      if (bp.group_id) params.group_id = bp.group_id
      if (bp.shift_id) params.shift_id = bp.shift_id
      const { data } = await salaryApi.payslip(params)
      setPaginated(data)
      setSinglePayslip(null)
    } catch {
      setPaginated(null)
    } finally {
      setLoading(false)
    }
  }, [buildParams])

  const handleSearch = async () => {
    setLoading(true)
    setSearched(true)
    try {
      const params = buildParams()
      if (employeeId) {
        const { data } = await salaryApi.payslip(params)
        setSinglePayslip(data)
        setPaginated(null)
      } else {
        params.page = "1"
        const { data } = await salaryApi.payslip(params)
        setPaginated(data)
        setSinglePayslip(null)
        setPage(data.page || 1)
      }
    } catch {
      setSinglePayslip(null)
      setPaginated(null)
    } finally {
      setLoading(false)
    }
  }

  const handleReset = () => {
    setCompanyId(companies[0]?.id || "")
    setDepartmentId("")
    setSectionId("")
    setDesignationId("")
    setLineId("")
    setGroupId("")
    setShiftId("")
    setEmployeeId("")
    setSinglePayslip(null)
    setPaginated(null)
    setSearched(false)
    setPage(1)
  }

  const handleExport = async (kind: ExportKind) => {
    if (exporting) return
    setExporting(kind)
    try {
      const params = { ...buildParams(), lang }
      const res = kind === "pdf" ? await salaryApi.payslipExportPdf(params) : await salaryApi.payslipExport(params)
      const url = window.URL.createObjectURL(new Blob([res.data]))
      const a = document.createElement("a")
      a.href = url
      const suffix = employeeId ? `_${employeeId}` : ""
      const ext = kind === "pdf" ? "pdf" : "xlsx"
      a.download = `payslip${suffix}_${MONTHS[month]}_${year}_${lang}.${ext}`
      a.click()
      window.URL.revokeObjectURL(url)
    } catch {
      toast.error(`Failed to export ${kind.toUpperCase()}`)
    } finally {
      setExporting(null)
    }
  }

  const hasResults = singlePayslip || (paginated?.salaries && paginated.salaries.length > 0)
  const exportDisabled = loading || !searched || !hasResults || !!exporting

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-2">
          <div className="flex items-center gap-2">
            <ReceiptIcon className="h-6 w-6 text-muted-foreground" />
            <div>
              <h1 className="text-lg md:text-3xl font-bold tracking-tight">Payslip</h1>
              <p className="text-muted-foreground mt-1">Employee salary payslip with Office & Employee copy</p>
            </div>
          </div>
          <div className="hidden md:flex items-center gap-2">
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
            <Button variant="outline" size="sm" disabled={exportDisabled} onClick={() => handleExport("excel")}>
              {exporting === "excel" ? <Loader2 className="h-4 w-4 mr-1.5 animate-spin" /> : <FileDownIcon className="h-4 w-4 mr-1.5" />}
              Excel
            </Button>
            <Button variant="outline" size="sm" disabled={exportDisabled} onClick={() => handleExport("pdf")}>
              {exporting === "pdf" ? <Loader2 className="h-4 w-4 mr-1.5 animate-spin" /> : <FileTextIcon className="h-4 w-4 mr-1.5" />}
              PDF
            </Button>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleFilterChange}
          onApply={handleSearch}
          onReset={handleReset}
          submitting={loading}
          applyLabel="Search"
        />
      </div>

      <div className="md:hidden px-4 lg:px-6">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleFilterChange}
          onApply={handleSearch}
          onReset={handleReset}
          submitting={loading}
          singleColumn
          applyLabel="Search"
        />
      </div>

      {searched && !loading && !hasResults && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-muted px-4 py-3 text-sm text-muted-foreground">
            No payslip found. Process salary first or adjust filters.
          </div>
        </div>
      )}

      {loading && (
        <div className="px-4 lg:px-6 flex justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      )}

      {singlePayslip && !loading && (
        <div className="px-4 lg:px-6 max-w-xl mx-auto w-full">
          <PayslipCard s={singlePayslip} month={month} year={year} />
        </div>
      )}

      {paginated && paginated.salaries.length > 0 && !loading && (
        <div className="px-4 lg:px-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold">{MONTHS[month]} {year} - Payslips</h2>
            <span className="text-sm text-muted-foreground">
              Page {paginated.page} of {paginated.total_pages} ({paginated.total} total)
            </span>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {paginated.salaries.map((s) => (
              <PayslipCard key={s.id} s={s} month={month} year={year} />
            ))}
          </div>
          {paginated.total_pages > 1 && (
            <div className="flex items-center justify-center gap-4 mt-6">
              <Button
                variant="outline"
                size="sm"
                disabled={paginated.page <= 1}
                onClick={() => loadPage(paginated.page - 1)}
              >
                <ChevronLeftIcon className="mr-1 h-4 w-4" /> Previous
              </Button>
              <div className="flex gap-1">
                {Array.from({ length: Math.min(paginated.total_pages, 5) }, (_, i) => {
                  const startPage = Math.max(1, paginated.page - 2)
                  const p = startPage + i
                  if (p > paginated.total_pages) return null
                  return (
                    <Button
                      key={p}
                      variant={p === paginated.page ? "default" : "outline"}
                      size="sm"
                      className="min-w-[32px]"
                      onClick={() => loadPage(p)}
                    >
                      {p}
                    </Button>
                  )
                })}
              </div>
              <Button
                variant="outline"
                size="sm"
                disabled={paginated.page >= paginated.total_pages}
                onClick={() => loadPage(paginated.page + 1)}
              >
                Next <ChevronRightIcon className="ml-1 h-4 w-4" />
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}