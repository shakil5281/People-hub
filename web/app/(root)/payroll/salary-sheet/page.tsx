"use client"

import * as React from "react"
import { FileSpreadsheetIcon, LandmarkIcon, SearchIcon, FileDownIcon } from "lucide-react"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { salaryApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

interface EmployeeInfo {
  employee_id: string
  name_en: string
  designation_ref?: { name: string }
  joining_date: string
  department?: { name: string }
}

interface SalaryRecord {
  id: string
  employee: EmployeeInfo
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

export default function SalarySheetPage() {
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

  const [data, setData] = React.useState<SalaryRecord[]>([])
  const [totals, setTotals] = React.useState<Record<string,number> | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [exporting, setExporting] = React.useState(false)
  const [applied, setApplied] = React.useState(false)

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

  const fetchData = React.useCallback(async () => {
    if (!companyId) return
    setLoading(true)
    setApplied(true)
    try {
      const params: Record<string, string> = {
        company_id: companyId,
        month: String(month + 1),
        year: String(year),
      }
      if (departmentId) params.department_id = departmentId
      if (sectionId) params.section_id = sectionId
      if (designationId) params.designation_id = designationId
      if (lineId) params.line_id = lineId
      if (groupId) params.group_id = groupId
      if (shiftId) params.shift_id = shiftId
      if (employeeId) params.employee_id = employeeId

      const { data: res } = await salaryApi.sheet(params)
      setData((res.salaries || []).map((s: any, i: number) => ({ ...s, id: s.id || `s-${i}` })))
      setTotals(res.totals || null)
    } catch {
      setData([])
      setTotals(null)
    } finally {
      setLoading(false)
    }
  }, [companyId, month, year, departmentId, sectionId, designationId, lineId, groupId, shiftId, employeeId])

  const handleApply = () => { fetchData() }

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

  const cols: ColumnDef<SalaryRecord>[] = [
    {id:"sl",header:"Sl",cell:({row}:any)=>row.index+1},
    {id:"emp_code",header:"Employee ID",accessorFn:(r:any)=>r.employee?.employee_id},
    {id:"emp_name",header:"Name",accessorFn:(r:any)=>r.employee?.name_en},
    {id:"designation",header:"Designation",accessorFn:(r:any)=>r.employee?.designation_ref?.name||"-"},
    {id:"joining_date",header:"Joining Date",accessorFn:(r:any)=>r.employee?.joining_date?.split("T")[0]},
    {accessorKey:"total_days",header:"Working Days"},
    {accessorKey:"present_days",header:"Present"},
    {accessorKey:"absent_days",header:"Absent"},
    {accessorKey:"late_days",header:"Late"},
    {accessorKey:"leave_days",header:"Leave"},
    {accessorKey:"holiday_days",header:"Holiday"},
    {accessorKey:"weekend_days",header:"Weekend"},
    {accessorKey:"basic_salary",header:"Basic Salary",cell:({row}:any)=>row.original.basic_salary.toLocaleString()},
    {accessorKey:"house_rent",header:"House Rent",cell:({row}:any)=>row.original.house_rent.toLocaleString()},
    {accessorKey:"medical_allowance",header:"Medical",cell:({row}:any)=>row.original.medical_allowance.toLocaleString()},
    {accessorKey:"transport_allowance",header:"Transport",cell:({row}:any)=>row.original.transport_allowance.toLocaleString()},
    {accessorKey:"food_allowance",header:"Food",cell:({row}:any)=>row.original.food_allowance.toLocaleString()},
    {accessorKey:"other_allowance",header:"Other",cell:({row}:any)=>row.original.other_allowance.toLocaleString()},
    {accessorKey:"gross_salary",header:"Gross",cell:({row}:any)=>row.original.gross_salary.toLocaleString()},
    {accessorKey:"absent_deduction",header:"Absent Ded.",cell:({row}:any)=>row.original.absent_deduction.toLocaleString()},
    {accessorKey:"overtime_hours",header:"OT Hours",cell:({row}:any)=>row.original.overtime_hours.toFixed(2)},
    {accessorKey:"overtime_rate",header:"OT Rate",cell:({row}:any)=>row.original.overtime_rate.toFixed(2)},
    {accessorKey:"overtime_amount",header:"OT Amount",cell:({row}:any)=>row.original.overtime_amount.toLocaleString()},
    {accessorKey:"attendance_bonus",header:"Att. Bonus",cell:({row}:any)=>row.original.attendance_bonus.toLocaleString()},
    {accessorKey:"net_salary",header:"Net Salary",cell:({row}:any)=>row.original.net_salary.toLocaleString()},
    {accessorKey:"status",header:"Status"},
  ]

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <FileSpreadsheetIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Salary Sheet</h1>
            <p className="text-muted-foreground mt-1">Employee salary breakdown</p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={exporting || !applied}
            onClick={async () => {
              setExporting(true)
              try {
                const params: Record<string, string> = {
                  company_id: companyId,
                  month: String(month + 1),
                  year: String(year),
                }
                if (departmentId) params.department_id = departmentId
                if (sectionId) params.section_id = sectionId
                if (designationId) params.designation_id = designationId
                if (lineId) params.line_id = lineId
                if (groupId) params.group_id = groupId
                if (shiftId) params.shift_id = shiftId
                if (employeeId) params.employee_id = employeeId
                const res = await salaryApi.sheetExport(params)
                const url = window.URL.createObjectURL(new Blob([res.data]))
                const a = document.createElement("a")
                a.href = url
                a.download = `salary_sheet_${MONTHS[month]}_${year}.xlsx`
                a.click()
                window.URL.revokeObjectURL(url)
              } finally { setExporting(false) }
            }}
          >
            <FileDownIcon className="mr-2 h-4 w-4" />
            {exporting ? "Exporting..." : "Export Excel"}
          </Button>
          <Link href="/payroll/bank-sheet">
            <Button variant="outline" size="sm">
              <LandmarkIcon className="mr-2 h-4 w-4" />
              Bank Sheet
            </Button>
          </Link>
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
            <Button onClick={handleApply} disabled={loading || !companyId}>
              {loading ? "Loading..." : "Apply Filters"}
            </Button>
            <Button variant="outline" onClick={handleReset}>Reset</Button>
          </div>
        </div>
      </div>

      {applied && (
        <div className="px-4 lg:px-6">
          <h2 className="text-lg font-semibold mb-2">{MONTHS[month]} {year} - Salary Sheet</h2>
        </div>
      )}

      {totals && (
        <div className="px-4 lg:px-6">
          <div className="flex flex-wrap gap-x-6 gap-y-1 text-sm text-muted-foreground mb-2 p-3 rounded-lg border bg-card">
            <span>Gross: <strong className="text-foreground">{totals.gross_salary?.toLocaleString()}</strong></span>
            <span>Absent Ded.: <strong className="text-foreground">{totals.absent_deduction?.toLocaleString()}</strong></span>
            <span>OT Hrs: <strong className="text-foreground">{totals.overtime_hours?.toFixed(2)}</strong></span>
            <span>OT Amt: <strong className="text-foreground">{totals.overtime_amount?.toLocaleString()}</strong></span>
            <span>Att. Bonus: <strong className="text-foreground">{totals.attendance_bonus?.toLocaleString()}</strong></span>
            <span>Deductions: <strong className="text-foreground">{totals.total_deductions?.toLocaleString()}</strong></span>
            <span>Present: <strong className="text-foreground">{totals.present_days}</strong></span>
            <span>Absent: <strong className="text-foreground">{totals.absent_days}</strong></span>
            <span>Late: <strong className="text-foreground">{totals.late_days}</strong></span>
            <span>Leave: <strong className="text-foreground">{totals.leave_days}</strong></span>
            <span>Holiday: <strong className="text-foreground">{totals.holiday_days}</strong></span>
            <span>Weekend: <strong className="text-foreground">{totals.weekend_days}</strong></span>
            <span>Net: <strong className="text-foreground">{totals.net_salary?.toLocaleString()}</strong></span>
          </div>
        </div>
      )}
      <DataTable data={data} columns={cols} loading={loading} />
    </div>
  )
}
