"use client"

import * as React from "react"
import { FileTextIcon, SearchIcon, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Badge } from "@/components/ui/badge"
import { toast } from "sonner"
import { earnedLeaveApi, companyApi } from "@/lib/api"

interface LedgerRow {
  id: string
  employee_id: string
  transaction_date: string
  transaction_type: string
  opening_balance: number
  accrued: number
  used: number
  adjusted: number
  encashed: number
  closing_balance: number
  remarks: string
}

export default function ELLedgerPage() {
  const [companies, setCompanies] = React.useState<{ id: string; company_name_en: string }[]>([])
  const [companyId, setCompanyId] = React.useState("")
  const [employeeId, setEmployeeId] = React.useState("")
  const [type, setType] = React.useState("")
  const [data, setData] = React.useState<LedgerRow[]>([])
  const [total, setTotal] = React.useState(0)
  const [page, setPage] = React.useState(1)
  const [limit] = React.useState(20)
  const [loading, setLoading] = React.useState(false)

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((r) => {
      const list = r.data?.data || r.data || []
      setCompanies(Array.isArray(list) ? list : [])
      if (Array.isArray(list) && list.length > 0) setCompanyId(list[0].id)
    })
  }, [])

  const cols: ColumnDef<LedgerRow>[] = [
    { accessorKey: "transaction_date", header: "Date", cell: ({ row }) => row.original.transaction_date?.slice(0, 10) || "-" },
    { accessorKey: "employee_id", header: "Employee" },
    { accessorKey: "transaction_type", header: "Type", cell: ({ row }) => <Badge variant="secondary">{row.original.transaction_type}</Badge> },
    { accessorKey: "opening_balance", header: "Opening" },
    { accessorKey: "accrued", header: "Accrued" },
    { accessorKey: "used", header: "Used" },
    { accessorKey: "encashed", header: "Encashed" },
    { accessorKey: "closing_balance", header: "Closing", cell: ({ row }) => <span className="font-semibold">{row.original.closing_balance}</span> },
  ]

  const fetch = async (p = page) => {
    if (!companyId) { toast.error("Select company"); return }
    setLoading(true)
    try {
      const params: Record<string, string> = { company_id: companyId, page: String(p), limit: String(limit) }
      if (employeeId) params.employee_id = employeeId
      if (type) params.transaction_type = type
      const res = await earnedLeaveApi.ledger(params)
      setData(res.data?.data || res.data || [])
      setTotal(res.data?.total || 0)
    } catch { toast.error("Failed to load ledger") }
    finally { setLoading(false) }
  }

  React.useEffect(() => { if (companyId) fetch(1) }, [companyId])

  return (
    <div className="flex flex-col gap-4 py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <FileTextIcon className="h-6 w-6 text-amber-600" />
        <div>
          <h1 className="text-2xl font-bold tracking-tight">EL Ledger</h1>
          <p className="text-sm text-muted-foreground">Append-only audit trail — every accrual, usage, adjustment, encashment</p>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader><CardTitle className="text-sm">Filters</CardTitle></CardHeader>
          <CardContent className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Company</label><select value={companyId} onChange={(e) => setCompanyId(e.target.value)} className="flex h-10 rounded-md border px-3 text-sm bg-background"><option value="">Select</option>{companies.map((c) => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}</select></div>
            <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Employee ID</label><input value={employeeId} onChange={(e) => setEmployeeId(e.target.value)} placeholder="EMP001" className="flex h-10 rounded-md border px-3 text-sm" /></div>
            <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Type</label><select value={type} onChange={(e) => setType(e.target.value)} className="flex h-10 rounded-md border px-3 text-sm bg-background"><option value="">All</option><option value="ACCRUAL">ACCRUAL</option><option value="LEAVE_USED">LEAVE_USED</option><option value="ADJUSTMENT_ADD">ADJUSTMENT_ADD</option><option value="ADJUSTMENT_DEDUCT">ADJUSTMENT_DEDUCT</option><option value="ENCASHMENT">ENCASHMENT</option><option value="REVERSAL">REVERSAL</option></select></div>
            <div className="flex items-end"><Button onClick={() => { setPage(1); fetch(1) }} disabled={loading}>{loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <SearchIcon className="mr-2 h-4 w-4" />}Search</Button></div>
          </CardContent>
        </Card>
      </div>

      <div className="px-4 lg:px-6">
        <DataTable data={data} columns={cols} loading={loading} serverSide page={page} pageSize={limit} total={total} pageCount={Math.ceil(total / limit)} onPageChange={(p) => { setPage(p); fetch(p) }} />
      </div>
    </div>
  )
}
