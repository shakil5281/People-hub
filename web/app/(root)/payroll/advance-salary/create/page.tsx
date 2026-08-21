"use client"

import * as React from "react"
import { BanknoteIcon, Loader2, SearchIcon, UsersIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { DatePicker } from "@/components/ui/date-picker"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { advanceSalaryApi, employeeApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi } from "@/lib/api"
import { useRouter } from "next/navigation"
import { format } from "date-fns"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }

interface Employee {
  id: string
  employee_id: string
  name_en: string
  gross_salary: number
  designation_ref?: { name: string }
  department?: { name: string }
}

export default function ApplyAdvancePage() {
  const router = useRouter()
  const [employees, setEmployees] = React.useState<Employee[]>([])
  const [loading, setLoading] = React.useState(false)
  const [submitting, setSubmitting] = React.useState(false)
  const [dialogOpen, setDialogOpen] = React.useState(false)

  const [selectedRows, setSelectedRows] = React.useState<Employee[]>([])
  
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])

  const [companyId, setCompanyId] = React.useState("")
  const [departmentId, setDepartmentId] = React.useState("")
  const [sectionId, setSectionId] = React.useState("")
  const [designationId, setDesignationId] = React.useState("")
  const [lineId, setLineId] = React.useState("")
  const [groupId, setGroupId] = React.useState("")

  const currentDate = new Date()
  const [advanceDate, setAdvanceDate] = React.useState(format(currentDate, "yyyy-MM-dd"))
  const [deductionMonth, setDeductionMonth] = React.useState(currentDate.getMonth() + 1)
  const [deductionYear, setDeductionYear] = React.useState(currentDate.getFullYear())
  const [amount, setAmount] = React.useState("")
  const [reason, setReason] = React.useState("")

  const [statusFilter, setStatusFilter] = React.useState("active")
  const [employeeTypeFilter, setEmployeeTypeFilter] = React.useState("Regular")

  // New features requested by user
  const [salaryType, setSalaryType] = React.useState("fixed_salary")
  const [isOverTime, setIsOverTime] = React.useState(false)

  React.useEffect(() => {
    const init = async () => {
      try {
        const [comp, dept, grp] = await Promise.all([
          companyApi.list(),
          departmentApi.list({ limit: "100" }),
          groupApi.list({ limit: "100" })
        ])
        setCompanies(Array.isArray(comp.data?.data) ? comp.data.data : Array.isArray(comp.data) ? comp.data : [])
        setDepartments(Array.isArray(dept.data?.data) ? dept.data.data : Array.isArray(dept.data) ? dept.data : [])
        setGroups(Array.isArray(grp.data?.data) ? grp.data.data : Array.isArray(grp.data) ? grp.data : [])
      } catch (err) {
        console.error("Failed to load initial dropdowns", err)
      }
    }
    init()
  }, [])

  const fetchSections = React.useCallback(async (deptId: string) => {
    try {
      const { data } = await sectionApi.list(deptId || undefined, { limit: "100" })
      setSections(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setSections([]) }
  }, [])

  const fetchDesignations = React.useCallback(async (secId: string) => {
    try {
      const { data } = await designationApi.list(secId || undefined, { limit: "100" })
      setDesignations(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setDesignations([]) }
  }, [])

  const fetchLines = React.useCallback(async (secId: string) => {
    try {
      const { data } = await lineApi.list(secId || undefined, { limit: "100" })
      setLines(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setLines([]) }
  }, [])

  React.useEffect(() => {
    if (!departmentId) {
      setSections([])
      setDesignations([])
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
    setDesignations([])
    setLines([])
  }, [departmentId, fetchSections])

  React.useEffect(() => {
    if (!sectionId) {
      setDesignations([])
      setLines([])
      setDesignationId("")
      setLineId("")
      return
    }
    fetchDesignations(sectionId)
    fetchLines(sectionId)
    setDesignationId("")
    setLineId("")
  }, [sectionId, fetchDesignations, fetchLines])

  const handleSearch = async () => {
    if (!companyId) {
      toast.error("Please select a company")
      return
    }
    setLoading(true)
    try {
      const params: Record<string, string> = { 
        company_id: companyId, 
        limit: "1000",
        status: statusFilter,
        employee_type: employeeTypeFilter
      }
      if (departmentId) params.department_id = departmentId
      if (sectionId) params.section_id = sectionId
      if (designationId) params.designation_id = designationId
      if (lineId) params.line_id = lineId
      if (groupId) params.group_id = groupId

      const { data } = await employeeApi.list(params)
      const list = Array.isArray(data?.data) ? data.data : []
      setEmployees(list.map((e: any) => ({ ...e, id: e.employee_id })))
      setSelectedRows([])
    } catch (err: any) {
      toast.error(err.response?.data?.error || "Failed to load employees")
    } finally {
      setLoading(false)
    }
  }



  const handleSubmit = async () => {
    if (!selectedRows.length) return toast.error("Select at least one employee")
    if (!companyId) return toast.error("Company is missing")
    if (!amount || Number(amount) <= 0) return toast.error("Enter a valid amount")
    if (!advanceDate) return toast.error("Advance date is required")
    if (!deductionMonth || !deductionYear) return toast.error("Target deduction month/year is required")

    setSubmitting(true)
    try {
      const payload: Record<string, any> = {
        company_id: companyId,
        amount: Number(amount),
        advance_date: advanceDate,
        deduction_month: Number(deductionMonth),
        deduction_year: Number(deductionYear),
        reason: reason,
        employee_ids: selectedRows.map(r => r.employee_id),
        // future fields
        salary_type: salaryType,
        is_overtime: isOverTime
      }

      await advanceSalaryApi.bulkApply(payload)
      toast.success("Advance salary applied successfully")
      setDialogOpen(false)
      router.push("/payroll/advance-salary")
    } catch (err: any) {
      toast.error(err.response?.data?.error || "Failed to apply advance")
    } finally {
      setSubmitting(false)
    }
  }

  const columns: ColumnDef<Employee>[] = [
    { accessorKey: "employee_id", header: "Emp ID" },
    { accessorKey: "name_en", header: "Name" },
    { accessorKey: "designation_ref.name", header: "Designation" },
    { accessorKey: "department.name", header: "Department" },
    {
      accessorKey: "gross_salary",
      header: "Gross Salary",
      cell: ({ row }) => <span className="font-medium text-emerald-600">{(row.original.gross_salary || 0).toFixed(2)}</span>
    }
  ]

  const selectCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
  const inputCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
  const labelCls = "text-xs font-semibold text-muted-foreground uppercase tracking-wider"

  return (
    <div className="flex flex-col gap-6 p-6">
      <div className="flex items-center justify-between gap-2 shrink-0">
        <div className="flex items-center gap-2 text-primary">
          <div className="rounded-lg bg-primary/10 p-2">
            <BanknoteIcon className="h-6 w-6" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight">Apply Advance</h1>
            <p className="text-sm text-muted-foreground">Select employees and apply bulk advance salary</p>
          </div>
        </div>
        
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button size="lg" disabled={!selectedRows.length}>
              Continue with {selectedRows.length} Employee{selectedRows.length !== 1 ? 's' : ''}
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>Advance Details</DialogTitle>
              <DialogDescription>
                Configure advance properties for the {selectedRows.length} selected employees.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className={labelCls}>Amount</label>
                  <input type="number" placeholder="Enter amount" value={amount} onChange={e => setAmount(e.target.value)} className={inputCls} min="0" />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className={labelCls}>Advance Date</label>
                  <DatePicker 
                    value={advanceDate ? new Date(advanceDate) : undefined} 
                    onChange={date => setAdvanceDate(date ? format(date, "yyyy-MM-dd") : "")} 
                    className="w-full"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className={labelCls}>Salary Source</label>
                  <select value={salaryType} onChange={e => setSalaryType(e.target.value)} className={selectCls}>
                    <option value="fixed_salary">Fixed Salary</option>
                    <option value="fixed_date">Fixed Date</option>
                  </select>
                </div>
                <div className="flex items-center gap-3 pt-6">
                  <Switch 
                    id="overtime-mode" 
                    checked={isOverTime} 
                    onCheckedChange={setIsOverTime} 
                  />
                  <Label htmlFor="overtime-mode" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                    OverTime Deduction
                  </Label>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className={labelCls}>Ded. Month</label>
                  <select value={deductionMonth} onChange={e => setDeductionMonth(Number(e.target.value))} className={selectCls}>
                    {Array.from({ length: 12 }, (_, i) => i + 1).map(m => (
                      <option key={m} value={m}>{new Date(2000, m - 1).toLocaleString('default', { month: 'short' })}</option>
                    ))}
                  </select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className={labelCls}>Ded. Year</label>
                  <select value={deductionYear} onChange={e => setDeductionYear(Number(e.target.value))} className={selectCls}>
                    {Array.from({ length: 5 }, (_, i) => currentDate.getFullYear() - 1 + i).map(y => (
                      <option key={y} value={y}>{y}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Reason (Optional)</label>
                <input type="text" placeholder="E.g. Medical emergency" value={reason} onChange={e => setReason(e.target.value)} className={inputCls} />
              </div>

              <Button 
                onClick={handleSubmit} 
                disabled={submitting || !amount || !advanceDate} 
                className="w-full mt-4"
                size="lg"
              >
                {submitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                Submit Advance Salary
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      <div className="flex flex-col gap-4">
        <Card className="shrink-0 shadow-sm border-emerald-100 dark:border-emerald-900/50">
          <CardContent className="p-4">
            <div className="grid grid-cols-2 md:grid-cols-6 xl:grid-cols-12 gap-3">
              <div className="flex flex-col gap-1.5 md:col-span-2">
                <label className={labelCls}>Company</label>
                <select value={companyId} onChange={e => setCompanyId(e.target.value)} className={selectCls}>
                  <option value="">Select Company</option>
                  {companies.map(c => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5 md:col-span-2">
                <label className={labelCls}>Department</label>
                <select value={departmentId} onChange={e => setDepartmentId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5 md:col-span-2">
                <label className={labelCls}>Section</label>
                <select value={sectionId} onChange={e => setSectionId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {sections.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5 md:col-span-2">
                <label className={labelCls}>Designation</label>
                <select value={designationId} onChange={e => setDesignationId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {designations.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5 md:col-span-2">
                <label className={labelCls}>Line</label>
                <select value={lineId} onChange={e => setLineId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {lines.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5 md:col-span-2">
                <label className={labelCls}>Status</label>
                <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} className={selectCls}>
                  <option value="">— All —</option>
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                  <option value="terminated">Terminated</option>
                  <option value="resigned">Resigned</option>
                  <option value="left">Left</option>
                </select>
              </div>
              <div className="flex flex-col gap-1.5 md:col-span-2">
                <label className={labelCls}>Employee Type</label>
                <select value={employeeTypeFilter} onChange={e => setEmployeeTypeFilter(e.target.value)} className={selectCls}>
                  <option value="">— All —</option>
                  <option value="Regular">Regular</option>
                  <option value="Lefty">Lefty</option>
                  <option value="Resign">Resign</option>
                  <option value="Close">Close</option>
                </select>
              </div>
              <div className="flex flex-col justify-end md:col-span-2">
                <Button onClick={handleSearch} disabled={loading || !companyId} className="w-full">
                  {loading ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : <SearchIcon className="h-4 w-4 mr-2" />}
                  Search
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        <div className="rounded-xl border bg-card text-card-foreground shadow-sm">
          {employees.length === 0 ? (
            <div className="flex h-full flex-col items-center justify-center text-muted-foreground p-8">
              <UsersIcon className="mb-2 h-10 w-10 opacity-20" />
              <p>Search to select employees</p>
            </div>
          ) : (
            <DataTable 
              columns={columns} 
              data={employees} 
              enableSelection={true}
              onSelectionChange={setSelectedRows}
            />
          )}
        </div>
      </div>
    </div>
  )
}
