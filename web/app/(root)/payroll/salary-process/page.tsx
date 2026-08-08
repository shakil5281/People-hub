"use client"

import * as React from "react"
import { ArrowUpDownIcon, Loader2 } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { salaryApi, companyApi } from "@/lib/api"
interface Company { id: string; company_name_en: string }

const MONTHS = ["January","February","March","April","May","June","July","August","September","October","November","December"]
const currentYear = new Date().getFullYear()
const currentMonth = new Date().getMonth()
const YEARS = Array.from({length:10},(_,i)=>currentYear-5+i)

export default function SalaryProcessPage() {
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [companyId, setCompanyId] = React.useState("")
  const [month, setMonth] = React.useState(currentMonth)
  const [year, setYear] = React.useState(currentYear)
  const [deductEarlyExit, setDeductEarlyExit] = React.useState(true)
  const [processing, setProcessing] = React.useState(false)
  const [result, setResult] = React.useState<{processed:number;total:number} | null>(null)

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then(({data})=>{ const list = Array.isArray(data?.data) ? data.data : (Array.isArray(data) ? data : []); if (list.length>0) { setCompanies(list); setCompanyId(list[0].id) } })
  }, [])

  const handleProcess = async () => {
    if (!companyId) { toast.error("Select a company"); return }
    setProcessing(true)
    try {
      const { data } = await salaryApi.process({ company_id: companyId, month: month+1, year, deduct_early_exit: deductEarlyExit })
      toast.success(data.message || "Salary processed")
      setResult({ processed: data.processed, total: data.total })
    } catch {
      toast.error("Failed to process salary")
    } finally { setProcessing(false) }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <ArrowUpDownIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Salary Process</h1>
            <p className="text-muted-foreground mt-1">Process monthly salary for all employees</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader className="pb-3"><CardTitle className="text-base">Process Parameters</CardTitle></CardHeader>
          <CardContent>
            <div className="flex flex-col gap-4">
              <div className="flex flex-wrap gap-4 items-end">
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Company</label>
                  <select value={companyId} onChange={e=>setCompanyId(e.target.value)} className="flex h-10 w-60 rounded-md border border-input bg-background px-3 py-2 text-sm">
                    <option value="">Select</option>
                    {companies.map(c=><option key={c.id} value={c.id}>{c.company_name_en}</option>)}
                  </select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Month</label>
                  <select value={month} onChange={e=>setMonth(Number(e.target.value))} className="flex h-10 w-40 rounded-md border border-input bg-background px-3 py-2 text-sm">
                    {MONTHS.map((n,i)=><option key={n} value={i}>{n}</option>)}
                  </select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Year</label>
                  <select value={year} onChange={e=>setYear(Number(e.target.value))} className="flex h-10 w-28 rounded-md border border-input bg-background px-3 py-2 text-sm">
                    {YEARS.map(y=><option key={y} value={y}>{y}</option>)}
                  </select>
                </div>
                <Button onClick={handleProcess} disabled={processing}>
                  {processing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <ArrowUpDownIcon className="mr-2 h-4 w-4" />}
                  Process Salary
                </Button>
              </div>

              <div className="flex items-center gap-2 pt-2 border-t text-sm">
                <input
                  type="checkbox"
                  id="deductEarlyExit"
                  checked={deductEarlyExit}
                  onChange={(e) => setDeductEarlyExit(e.target.checked)}
                  className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
                />
                <label htmlFor="deductEarlyExit" className="font-medium text-muted-foreground cursor-pointer">
                  Deduct OT Early Exit Shortfall Hours from Monthly Overtime (Formula Enabled)
                </label>
              </div>
            </div>
            {result && (
              <p className="text-sm text-muted-foreground mt-3">
                Processed {result.processed} of {result.total} employees
              </p>
            )}
          </CardContent>
        </Card>
      </div>

    </div>
  )
}
