"use client"

import * as React from "react"
import { BanknoteIcon, Loader2, FileSpreadsheetIcon, FileTextIcon, FileBarChartIcon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { advanceSalaryApi, companyApi, departmentApi } from "@/lib/api"
import { format } from "date-fns"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }

export default function AdvanceSalaryReportsPage() {
  const [loadingSheet, setLoadingSheet] = React.useState(false)
  const [loadingBank, setLoadingBank] = React.useState(false)
  const [loadingSummary, setLoadingSummary] = React.useState(false)

  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])

  const [companyId, setCompanyId] = React.useState("")
  const [departmentId, setDepartmentId] = React.useState("")
  const [employeeIdFilter, setEmployeeIdFilter] = React.useState("")
  
  const currentDate = new Date()
  const [month, setMonth] = React.useState(currentDate.getMonth() + 1)
  const [year, setYear] = React.useState(currentDate.getFullYear())

  React.useEffect(() => {
    const init = async () => {
      try {
        const [comp, dept] = await Promise.all([
          companyApi.list(),
          departmentApi.list({ limit: "100" })
        ])
        setCompanies(Array.isArray(comp.data?.data) ? comp.data.data : Array.isArray(comp.data) ? comp.data : [])
        setDepartments(Array.isArray(dept.data?.data) ? dept.data.data : Array.isArray(dept.data) ? dept.data : [])
      } catch (err) {
        console.error("Failed to load initial dropdowns", err)
      }
    }
    init()
  }, [])

  const handleGenerateSheet = async () => {
    if (!companyId) return toast.error("Please select a company")
    setLoadingSheet(true)
    try {
      const params: Record<string, string> = { company_id: companyId, month: String(month), year: String(year), report_type: "sheet" }
      if (departmentId) params.department_id = departmentId
      if (employeeIdFilter) params.employee_id = employeeIdFilter
      
      const res = await advanceSalaryApi.exportPdf(params)
      const url = window.URL.createObjectURL(new Blob([res.data], { type: "application/pdf" }))
      const link = document.createElement("a")
      link.href = url
      link.setAttribute("download", `advance_sheet_${month}_${year}.pdf`)
      document.body.appendChild(link)
      link.click()
      link.remove()
    } catch {
      toast.error("Failed to generate sheet")
    } finally {
      setLoadingSheet(false)
    }
  }

  const handleGenerateBankSheet = async () => {
    if (!companyId) return toast.error("Please select a company")
    setLoadingBank(true)
    try {
      const params: Record<string, string> = { company_id: companyId, month: String(month), year: String(year), report_type: "bank_sheet" }
      if (departmentId) params.department_id = departmentId
      
      const res = await advanceSalaryApi.exportPdf(params)
      const url = window.URL.createObjectURL(new Blob([res.data], { type: "application/pdf" }))
      const link = document.createElement("a")
      link.href = url
      link.setAttribute("download", `advance_bank_sheet_${month}_${year}.pdf`)
      document.body.appendChild(link)
      link.click()
      link.remove()
    } catch {
      toast.error("Failed to generate bank sheet")
    } finally {
      setLoadingBank(false)
    }
  }

  const handleGenerateSummary = async () => {
    if (!companyId) return toast.error("Please select a company")
    setLoadingSummary(true)
    try {
      const params: Record<string, string> = { company_id: companyId, month: String(month), year: String(year), report_type: "summary" }
      
      const res = await advanceSalaryApi.exportPdf(params)
      const url = window.URL.createObjectURL(new Blob([res.data], { type: "application/pdf" }))
      const link = document.createElement("a")
      link.href = url
      link.setAttribute("download", `advance_summary_${month}_${year}.pdf`)
      document.body.appendChild(link)
      link.click()
      link.remove()
    } catch {
      toast.error("Failed to generate summary")
    } finally {
      setLoadingSummary(false)
    }
  }

  const selectCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
  const inputCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
  const labelCls = "text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1.5 block"

  return (
    <div className="flex flex-col gap-6 p-6 max-w-4xl mx-auto">
      <div className="flex items-center gap-2 text-primary shrink-0">
        <div className="rounded-lg bg-primary/10 p-2">
          <BanknoteIcon className="h-6 w-6" />
        </div>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Advance Salary Reports</h1>
          <p className="text-sm text-muted-foreground">Generate comprehensive reports for employee advances</p>
        </div>
      </div>

      <Card className="shadow-sm border-emerald-100 dark:border-emerald-900/50">
        <CardContent className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
            <div>
              <label className={labelCls}>Company</label>
              <select value={companyId} onChange={e => setCompanyId(e.target.value)} className={selectCls}>
                <option value="">Select Company</option>
                {companies.map(c => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
              </select>
            </div>
            <div>
              <label className={labelCls}>Month</label>
              <select value={month} onChange={e => setMonth(Number(e.target.value))} className={selectCls}>
                {Array.from({ length: 12 }, (_, i) => i + 1).map(m => (
                  <option key={m} value={m}>{new Date(2000, m - 1).toLocaleString('default', { month: 'short' })}</option>
                ))}
              </select>
            </div>
            <div>
              <label className={labelCls}>Year</label>
              <select value={year} onChange={e => setYear(Number(e.target.value))} className={selectCls}>
                {Array.from({ length: 5 }, (_, i) => currentDate.getFullYear() - 2 + i).map(y => (
                  <option key={y} value={y}>{y}</option>
                ))}
              </select>
            </div>
            <div>
              <label className={labelCls}>Department (Optional)</label>
              <select value={departmentId} onChange={e => setDepartmentId(e.target.value)} className={selectCls}>
                <option value="">All Departments</option>
                {departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>
            <div>
              <label className={labelCls}>Employee ID (Optional)</label>
              <input 
                type="text" 
                placeholder="Specific employee ID" 
                value={employeeIdFilter} 
                onChange={e => setEmployeeIdFilter(e.target.value)} 
                className={inputCls} 
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 border-t pt-6">
            <Button onClick={handleGenerateSheet} disabled={loadingSheet || !companyId} size="lg" className="w-full flex-col h-auto py-4 bg-emerald-600 hover:bg-emerald-700">
              {loadingSheet ? <Loader2 className="h-6 w-6 mb-2 animate-spin" /> : <FileTextIcon className="h-6 w-6 mb-2" />}
              <span>Advance Sheet</span>
              <span className="text-xs font-normal opacity-80 mt-1">Detailed list of advances</span>
            </Button>
            
            <Button onClick={handleGenerateBankSheet} disabled={loadingBank || !companyId} size="lg" className="w-full flex-col h-auto py-4 bg-blue-600 hover:bg-blue-700">
              {loadingBank ? <Loader2 className="h-6 w-6 mb-2 animate-spin" /> : <FileSpreadsheetIcon className="h-6 w-6 mb-2" />}
              <span>Bank Sheet</span>
              <span className="text-xs font-normal opacity-80 mt-1">For bank disbursements</span>
            </Button>
            
            <Button onClick={handleGenerateSummary} disabled={loadingSummary || !companyId} size="lg" className="w-full flex-col h-auto py-4 bg-purple-600 hover:bg-purple-700">
              {loadingSummary ? <Loader2 className="h-6 w-6 mb-2 animate-spin" /> : <FileBarChartIcon className="h-6 w-6 mb-2" />}
              <span>Advance Summary</span>
              <span className="text-xs font-normal opacity-80 mt-1">Department-wise totals</span>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
