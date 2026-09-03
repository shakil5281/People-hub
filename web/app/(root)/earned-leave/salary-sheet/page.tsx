"use client"

import * as React from "react"
import { FileSpreadsheetIcon, Loader2, CheckIcon, ShieldIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Badge } from "@/components/ui/badge"
import Link from "next/link"
import { toast } from "sonner"
import { earnedLeaveApi, companyApi } from "@/lib/api"

interface Sheet {
  id: string
  period: string
  period_month: number
  period_year: number
  status: string
  total_employees: number
  total_days: number
  total_amount: number
}

export default function ELSalarySheetPage() {
  const [companies, setCompanies] = React.useState<{ id: string; company_name_en: string }[]>([])
  const [companyId, setCompanyId] = React.useState("")
  const [month, setMonth] = React.useState(new Date().getMonth() + 1)
  const [year, setYear] = React.useState(new Date().getFullYear())
  const [data, setData] = React.useState<Sheet[]>([])
  const [loading, setLoading] = React.useState(false)
  const [generating, setGenerating] = React.useState(false)

  const MONTHS = ["January","February","March","April","May","June","July","August","September","October","November","December"]

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((r) => {
      const list = r.data?.data || r.data || []
      setCompanies(Array.isArray(list) ? list : [])
      if (Array.isArray(list) && list.length > 0) setCompanyId(list[0].id)
    })
  }, [])

  const fetch = async () => {
    if (!companyId) return
    setLoading(true)
    try {
      const res = await earnedLeaveApi.listSalarySheets({ company_id: companyId })
      setData(res.data?.data || res.data || [])
    } catch { toast.error("Failed to load sheets") }
    finally { setLoading(false) }
  }

  React.useEffect(() => { if (companyId) fetch() }, [companyId])

  const handleGenerate = async () => {
    if (!companyId) { toast.error("Select company"); return }
    setGenerating(true)
    try {
      const res = await earnedLeaveApi.generateSalarySheet({ company_id: companyId, month, year })
      toast.success("Salary sheet generated")
      fetch()
    } catch (e: any) {
      toast.error(e?.response?.data?.error || "Failed to generate")
    } finally { setGenerating(false) }
  }

  const cols: ColumnDef<Sheet>[] = [
    { accessorKey: "period", header: "Period" },
    { accessorKey: "status", header: "Status", cell: ({ row }) => <Badge variant={row.original.status === "FINALIZED" ? "default" : "secondary"}>{row.original.status}</Badge> },
    { accessorKey: "total_employees", header: "Employees" },
    { accessorKey: "total_days", header: "Total Days" },
    { accessorKey: "total_amount", header: "Total Amount", cell: ({ row }) => `৳${Number(row.original.total_amount || 0).toLocaleString()}` },
    {
      id: "actions", header: "Actions", cell: ({ row }) => (
        <div className="flex gap-1">
          <Link href={`/earned-leave/salary-sheet/${row.original.id}`}><Button variant="outline" size="sm">View</Button></Link>
        </div>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4 py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <FileSpreadsheetIcon className="h-6 w-6 text-rose-600" />
          <div>
            <h1 className="text-2xl font-bold tracking-tight">EL Salary Sheet</h1>
            <p className="text-sm text-muted-foreground">Generate → Approve → Finalize (reversal via reverse)</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader><CardTitle className="text-sm">Generate</CardTitle></CardHeader>
          <CardContent className="grid grid-cols-1 sm:grid-cols-4 gap-4">
            <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Company *</label><select value={companyId} onChange={(e) => setCompanyId(e.target.value)} className="flex h-10 rounded-md border px-3 text-sm bg-background"><option value="">Select</option>{companies.map((c) => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}</select></div>
            <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Month *</label><select value={month} onChange={(e) => setMonth(Number(e.target.value))} className="flex h-10 rounded-md border px-3 text-sm bg-background">{MONTHS.map((m, i) => <option key={m} value={i + 1}>{m}</option>)}</select></div>
            <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Year *</label><select value={year} onChange={(e) => setYear(Number(e.target.value))} className="flex h-10 rounded-md border px-3 text-sm bg-background">{Array.from({length: 6}, (_,i)=> new Date().getFullYear()-2+i).map((y)=> <option key={y} value={y}>{y}</option>)}</select></div>
            <div className="flex items-end"><Button onClick={handleGenerate} disabled={generating} className="w-full">{generating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}Generate</Button></div>
          </CardContent>
        </Card>
      </div>

      <div className="px-4 lg:px-6">
        <DataTable data={data} columns={cols} loading={loading} enableSelection={false} />
      </div>
    </div>
  )
}
