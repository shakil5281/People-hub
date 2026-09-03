"use client"

import * as React from "react"
import { RefreshCwIcon, Loader2, BuildingIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { toast } from "sonner"
import { earnedLeaveApi, companyApi } from "@/lib/api"

export default function ELAccrualPage() {
  const [companies, setCompanies] = React.useState<{ id: string; company_name_en: string }[]>([])
  const [companyId, setCompanyId] = React.useState("")
  const [period, setPeriod] = React.useState(new Date().toISOString().slice(0, 7)) // YYYY-MM
  const [processing, setProcessing] = React.useState(false)
  const [result, setResult] = React.useState<any>(null)

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((r) => {
      const list = r.data?.data || r.data || []
      setCompanies(Array.isArray(list) ? list : [])
      if (Array.isArray(list) && list.length > 0) setCompanyId(list[0].id)
    })
  }, [])

  const handleProcess = async () => {
    if (!companyId || !period) { toast.error("Company and period required"); return }
    setProcessing(true)
    setResult(null)
    try {
      const res = await earnedLeaveApi.accrualProcess({ company_id: companyId, period })
      setResult(res.data)
      toast.success(res.data?.message || "Accrual completed")
    } catch (e: any) {
      const msg = e?.response?.data?.error || "Failed to process accrual"
      toast.error(msg)
    } finally { setProcessing(false) }
  }

  return (
    <div className="flex flex-col gap-6 py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <RefreshCwIcon className="h-6 w-6 text-blue-600" />
          <div>
            <h1 className="text-2xl font-bold tracking-tight">EL Accrual Process</h1>
            <p className="text-sm text-muted-foreground">Bulk accrual per period — idempotent (company+employee+period+type unique), transactional, with concurrency lock</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card className="max-w-2xl">
          <CardHeader><CardTitle className="text-base flex items-center gap-2"><BuildingIcon className="h-4 w-4" />Select Period</CardTitle><CardDescription>Period is YYYY-MM. Re-running same period will skip already accrued employees (idempotent).</CardDescription></CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium">Company *</label>
                <select value={companyId} onChange={(e) => setCompanyId(e.target.value)} className="flex h-10 rounded-md border bg-background px-3 text-sm">
                  <option value="">Select Company</option>
                  {companies.map((c) => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium">Period (YYYY-MM) *</label>
                <input type="month" value={period} onChange={(e) => setPeriod(e.target.value)} className="flex h-10 rounded-md border px-3 text-sm" />
              </div>
            </div>
            <div className="flex gap-2">
              <Button onClick={handleProcess} disabled={processing || !companyId || !period}>
                {processing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {processing ? "Processing..." : "Process Accrual"}
              </Button>
            </div>
            {result && (
              <div className="rounded-lg border bg-muted/30 p-4 text-sm space-y-1">
                <div className="font-semibold">{result.message}</div>
                <div className="grid grid-cols-3 gap-2 text-xs">
                  <span>Employees: <b>{result.employees}</b></span>
                  <span>Eligible: <b>{result.eligible}</b></span>
                  <span>Created: <b>{result.created}</b></span>
                  <span>Skipped: <b>{result.skipped}</b></span>
                  <span>Duration: <b>{result.duration_ms}ms</b></span>
                  <span>Period: <b>{result.period}</b></span>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader><CardTitle className="text-sm">Notes</CardTitle></CardHeader>
          <CardContent className="text-xs text-muted-foreground leading-relaxed">
            <ul className="list-disc pl-5 space-y-1">
              <li>Eligibility: joining + minService, status active, separation after period, policy active.</li>
              <li>Accrued = policy.accrual_rate clamped by max_balance - opening, rounded per policy.</li>
              <li>Second process for same period → existing Accrual skipped (idempotent).</li>
              <li>Concurrency: sync.Map + advisory lock prevents double process (409).</li>
            </ul>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
