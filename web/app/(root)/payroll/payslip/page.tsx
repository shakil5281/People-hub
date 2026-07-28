"use client"

import * as React from "react"
import { ReceiptIcon, Loader2, SearchIcon, FileDownIcon, ChevronLeftIcon, ChevronRightIcon } from "lucide-react"
import { salaryApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
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
  designation_ref?: { name: string }
  department?: { name: string }
  section_ref?: { name: string }
  line_ref?: { name: string }
  group_ref?: { name: string }
}

interface PayslipRecord {
  id: string
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
  absent_deduction: number
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

const MONTHS = ["January","February","March","April","May","June","July","August","September","October","November","December"]
const currentYear = new Date().getFullYear()
const currentMonth = new Date().getMonth()
const YEARS = Array.from({length:10},(_,i)=>currentYear-5+i)

const selectCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
const labelCls = "text-xs font-medium text-muted-foreground"

const fmt = (n: number) => n.toLocaleString()
const fmt2 = (n: number) => n.toFixed(2)

function PayslipCard({ s, month }: { s: PayslipRecord; month: number }) {
  return (
    <Card className="border-primary/10 shadow-sm">
      <CardHeader className="bg-gradient-to-r from-primary/5 to-primary/10 pb-3">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg">{s.employee?.name_en || "-"}</CardTitle>
            <p className="text-xs text-muted-foreground mt-0.5">
              {s.employee?.employee_id} | {s.employee?.designation_ref?.name || "-"}
            </p>
          </div>
          <span className="text-xs font-semibold text-primary bg-primary/10 px-2 py-1 rounded">
            {MONTHS[month]}
          </span>
        </div>
      </CardHeader>
      <CardContent className="pt-4 space-y-3">
        <div>
          <p className="text-xs font-medium text-muted-foreground mb-1 uppercase tracking-wide">Earnings</p>
          <div className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
            <span className="text-muted-foreground">Basic Salary</span><span className="text-right font-medium">{fmt(s.basic_salary)}</span>
            <span className="text-muted-foreground">House Rent</span><span className="text-right font-medium">{fmt(s.house_rent)}</span>
            <span className="text-muted-foreground">Medical</span><span className="text-right font-medium">{fmt(s.medical_allowance)}</span>
            <span className="text-muted-foreground">Transport</span><span className="text-right font-medium">{fmt(s.transport_allowance)}</span>
            <span className="text-muted-foreground">Food</span><span className="text-right font-medium">{fmt(s.food_allowance)}</span>
            <span className="text-muted-foreground">Other</span><span className="text-right font-medium">{fmt(s.other_allowance)}</span>
            <span className="font-semibold text-primary border-t pt-0.5">Gross</span>
            <span className="text-right font-semibold text-primary border-t pt-0.5">{fmt(s.gross_salary)}</span>
          </div>
        </div>
        <div>
          <p className="text-xs font-medium text-muted-foreground mb-1 uppercase tracking-wide">Overtime & Bonus</p>
          <div className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
            <span className="text-muted-foreground">OT Hours</span><span className="text-right">{fmt2(s.overtime_hours)}</span>
            <span className="text-muted-foreground">OT Rate</span><span className="text-right">{fmt2(s.overtime_rate)}</span>
            <span className="text-muted-foreground">OT Amount</span><span className="text-right font-medium">{fmt(s.overtime_amount)}</span>
            <span className="text-muted-foreground">Att. Bonus</span><span className="text-right font-medium">{fmt(s.attendance_bonus)}</span>
          </div>
        </div>
        <div>
          <p className="text-xs font-medium text-muted-foreground mb-1 uppercase tracking-wide">Deductions</p>
          <div className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
            <span className="text-muted-foreground">PF</span><span className="text-right">{fmt(s.provident_fund)}</span>
            <span className="text-muted-foreground">Tax</span><span className="text-right">{fmt(s.tax)}</span>
            <span className="text-muted-foreground">Absent Deduction</span><span className="text-right">{fmt(s.absent_deduction)}</span>
            <span className="font-semibold text-destructive border-t pt-0.5">Total Deductions</span>
            <span className="text-right font-semibold text-destructive border-t pt-0.5">{fmt(s.total_deductions)}</span>
          </div>
        </div>
        <div className="flex justify-between items-center p-3 rounded-lg bg-primary/5 border border-primary/10">
          <span className="font-bold text-lg">Net Salary</span>
          <span className="font-bold text-lg text-primary">{fmt(s.net_salary)}</span>
        </div>
        <div className="grid grid-cols-4 gap-1 text-xs text-muted-foreground border-t pt-2">
          <span>P: <strong>{s.present_days}</strong></span>
          <span>A: <strong>{s.absent_days}</strong></span>
          <span>L: <strong>{s.late_days}</strong></span>
          <span>LV: <strong>{s.leave_days}</strong></span>
          <span>H: <strong>{s.holiday_days}</strong></span>
          <span>W: <strong>{s.weekend_days}</strong></span>
          <span>T: <strong>{s.total_days}</strong></span>
          <span className="text-right"><span className={`px-1.5 py-0.5 rounded text-[10px] font-medium ${s.status === "processed" ? "bg-green-100 text-green-700" : "bg-yellow-100 text-yellow-700"}`}>{s.status}</span></span>
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
      setPage(data.page || pageNum)
    } catch {
      setPaginated(null)
    } finally {
      setLoading(false)
    }
  }, [buildParams])

  const handleSearch = async () => {
    if (!employeeId) {
      setSearched(false)
      return
    }
    setLoading(true)
    setSearched(true)
    try {
      const params = buildParams()
      const { data } = await salaryApi.payslip(params)
      setSinglePayslip(data)
      setPaginated(null)
    } catch {
      setSinglePayslip(null)
    } finally {
      setLoading(false)
    }
  }

  const handleApply = () => {
    setPage(1)
    loadPage(1)
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
  }

  const handleExport = async () => {
    setLoading(true)
    try {
      const params = buildParams()
      const res = await salaryApi.payslipExport(params)
      const url = window.URL.createObjectURL(new Blob([res.data]))
      const a = document.createElement("a")
      a.href = url
      const suffix = employeeId ? `_${employeeId}` : ""
      a.download = `payslip${suffix}_${MONTHS[month]}_${year}.xlsx`
      a.click()
      window.URL.revokeObjectURL(url)
    } finally {
      setLoading(false)
    }
  }

  const hasResults = singlePayslip || (paginated?.salaries && paginated.salaries.length > 0)

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <ReceiptIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Payslip</h1>
            <p className="text-muted-foreground mt-1">Employee salary payslip</p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" disabled={loading || !searched || !hasResults} onClick={handleExport}>
            <FileDownIcon className="mr-2 h-4 w-4" />
            {loading ? "Exporting..." : "Export Excel"}
          </Button>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <div className="rounded-lg border bg-card p-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Company</label>
              <select value={companyId} onChange={e => setCompanyId(e.target.value)} className={selectCls}>
                <option value="">Select</option>
                {companies.map(c => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Month</label>
              <select value={month} onChange={e => setMonth(Number(e.target.value))} className={selectCls}>
                {MONTHS.map((n, i) => <option key={n} value={i}>{n}</option>)}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Year</label>
              <select value={year} onChange={e => setYear(Number(e.target.value))} className={selectCls}>
                {YEARS.map(y => <option key={y} value={y}>{y}</option>)}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Department</label>
              <select value={departmentId} onChange={e => setDepartmentId(e.target.value)} className={selectCls}>
                <option value="">All</option>
                {departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Section</label>
              <select value={sectionId} onChange={e => setSectionId(e.target.value)} className={selectCls}>
                <option value="">All</option>
                {sections.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Designation</label>
              <select value={designationId} onChange={e => setDesignationId(e.target.value)} className={selectCls}>
                <option value="">All</option>
                {designations.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Line</label>
              <select value={lineId} onChange={e => setLineId(e.target.value)} className={selectCls}>
                <option value="">All</option>
                {lines.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Group</label>
              <select value={groupId} onChange={e => setGroupId(e.target.value)} className={selectCls}>
                <option value="">All</option>
                {groups.map(g => <option key={g.id} value={g.id}>{g.name}</option>)}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Shift</label>
              <select value={shiftId} onChange={e => setShiftId(e.target.value)} className={selectCls}>
                <option value="">All</option>
                {shifts.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Employee ID</label>
              <div className="relative">
                <SearchIcon className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <input
                  value={employeeId}
                  onChange={e => setEmployeeId(e.target.value)}
                  placeholder="Search..."
                  className="flex h-9 w-full rounded-md border border-input bg-transparent pl-8 pr-3 py-1 text-sm"
                />
              </div>
            </div>
          </div>
          <div className="flex gap-2 mt-4">
            <Button disabled={loading || !employeeId} onClick={handleSearch}>
              <SearchIcon className="mr-2 h-4 w-4" />
              {loading ? "Searching..." : "Search"}
            </Button>
            <Button variant="outline" disabled={loading || !companyId} onClick={handleApply}>
              Apply Filters
            </Button>
            <Button variant="outline" onClick={handleReset}>Reset</Button>
          </div>
        </div>
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
        <div className="px-4 lg:px-6 max-w-2xl mx-auto">
          <PayslipCard s={singlePayslip} month={month} />
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
              <PayslipCard key={s.id} s={s} month={month} />
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