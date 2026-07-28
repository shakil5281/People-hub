"use client"

import * as React from "react"
import { ReceiptIcon, Loader2, SearchIcon, FileDownIcon } from "lucide-react"
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

interface PayslipData {
  employee?: { employee_id:string; name_en:string; designation_ref?:{name:string}; joining_date:string }
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

const MONTHS = ["January","February","March","April","May","June","July","August","September","October","November","December"]
const currentYear = new Date().getFullYear()
const currentMonth = new Date().getMonth()
const YEARS = Array.from({length:10},(_,i)=>currentYear-5+i)

const selectCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
const labelCls = "text-xs font-medium text-muted-foreground"

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

  const [payslip, setPayslip] = React.useState<PayslipData | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [searched, setSearched] = React.useState(false)
  const [exporting, setExporting] = React.useState(false)

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

  const handleSearch = async () => {
    if (!employeeId) return
    setLoading(true)
    setSearched(true)
    try {
      const { data } = await salaryApi.payslip({ employee_id: employeeId, month: String(month+1), year: String(year) })
      setPayslip(data)
    } catch {
      setPayslip(null)
    }
    finally { setLoading(false) }
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
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <ReceiptIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Payslip</h1>
            <p className="text-muted-foreground mt-1">View employee payslip</p>
          </div>
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
            <Button onClick={handleSearch} disabled={loading || !employeeId}>
              {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <SearchIcon className="mr-2 h-4 w-4" />}
              Search
            </Button>
            <Button variant="outline" onClick={handleReset}>Reset</Button>
          </div>
        </div>
      </div>

      {searched && !payslip && !loading && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-muted px-4 py-3 text-sm text-muted-foreground">
            No payslip found for this employee. Process salary first.
          </div>
        </div>
      )}

      {loading && (
        <div className="px-4 lg:px-6 flex justify-center py-12"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>
      )}

      {payslip && (
        <div className="px-4 lg:px-6">
          <div className="flex justify-end mb-2">
            <Button
              variant="outline"
              size="sm"
              disabled={exporting}
              onClick={async () => {
                setExporting(true)
                try {
                  const res = await salaryApi.payslipExport({ employee_id: employeeId, month: String(month+1), year: String(year) })
                  const url = window.URL.createObjectURL(new Blob([res.data]))
                  const a = document.createElement("a")
                  a.href = url
                  a.download = `payslip_${employeeId}_${MONTHS[month]}_${year}.xlsx`
                  a.click()
                  window.URL.revokeObjectURL(url)
                } finally { setExporting(false) }
              }}
            >
              <FileDownIcon className="mr-2 h-4 w-4" />
              {exporting ? "Exporting..." : "Export Excel"}
            </Button>
          </div>
          <Card>
            <CardHeader>
              <CardTitle className="text-xl">Payslip - {MONTHS[month]} {year}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4 mb-6 p-4 rounded-lg bg-muted/50">
                <div><span className="text-sm text-muted-foreground">Employee:</span> <strong>{payslip.employee?.name_en}</strong></div>
                <div><span className="text-sm text-muted-foreground">Code:</span> <strong>{payslip.employee?.employee_id}</strong></div>
                <div><span className="text-sm text-muted-foreground">Designation:</span> <strong>{payslip.employee?.designation_ref?.name || "-"}</strong></div>
              </div>

              <h3 className="font-semibold mb-2">Earnings</h3>
              <div className="grid grid-cols-2 gap-3 mb-4">
                <div className="flex justify-between p-2 rounded border"><span>Transport Allowance</span><span>{payslip.transport_allowance.toLocaleString()}</span></div>
                <div className="flex justify-between p-2 rounded border"><span>Food Allowance</span><span>{payslip.food_allowance.toLocaleString()}</span></div>
                <div className="flex justify-between p-2 rounded border"><span>Other Allowance</span><span>{payslip.other_allowance.toLocaleString()}</span></div>
                <div className="flex justify-between p-2 rounded border font-semibold bg-primary/5"><span>Gross Salary</span><span>{payslip.gross_salary.toLocaleString()}</span></div>
              </div>

              <h3 className="font-semibold mb-2">Overtime & Bonus</h3>
              <div className="grid grid-cols-2 gap-3 mb-4">
                <div className="flex justify-between p-2 rounded border"><span>OT Hours</span><span>{payslip.overtime_hours?.toFixed(2)}</span></div>
                <div className="flex justify-between p-2 rounded border"><span>OT Rate</span><span>{payslip.overtime_rate?.toFixed(2)}</span></div>
                <div className="flex justify-between p-2 rounded border"><span>OT Amount</span><span>{payslip.overtime_amount?.toLocaleString()}</span></div>
                <div className="flex justify-between p-2 rounded border"><span>Attendance Bonus</span><span>{payslip.attendance_bonus?.toLocaleString()}</span></div>
              </div>

              <h3 className="font-semibold mb-2">Deductions</h3>
              <div className="grid grid-cols-2 gap-3 mb-4">
                <div className="flex justify-between p-2 rounded border"><span>Provident Fund</span><span>{payslip.provident_fund.toLocaleString()}</span></div>
                <div className="flex justify-between p-2 rounded border"><span>Tax</span><span>{payslip.tax.toLocaleString()}</span></div>
                <div className="flex justify-between p-2 rounded border"><span>Absent Deduction</span><span>{payslip.absent_deduction.toLocaleString()}</span></div>
                <div className="flex justify-between p-2 rounded border font-semibold bg-destructive/5"><span>Total Deductions</span><span>{payslip.total_deductions.toLocaleString()}</span></div>
              </div>

              <div className="flex justify-between p-4 rounded-lg border-2 border-primary/20 bg-primary/5 text-lg font-bold mt-2">
                <span>Net Salary</span>
                <span>{payslip.net_salary.toLocaleString()}</span>
              </div>

              <div className="grid grid-cols-3 gap-3 mt-6 text-sm text-muted-foreground">
                <span>Present: <strong>{payslip.present_days}</strong></span>
                <span>Absent: <strong>{payslip.absent_days}</strong></span>
                <span>Late: <strong>{payslip.late_days}</strong></span>
                <span>Leave: <strong>{payslip.leave_days}</strong></span>
                <span>Holiday: <strong>{payslip.holiday_days}</strong></span>
                <span>Weekend: <strong>{payslip.weekend_days}</strong></span>
                <span>Total Days: <strong>{payslip.total_days}</strong></span>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
