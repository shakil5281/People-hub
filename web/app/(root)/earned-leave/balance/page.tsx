"use client"

import * as React from "react"
import { WalletIcon, SearchIcon, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { earnedLeaveApi, companyApi } from "@/lib/api"

interface BalanceRow {
  employee_id: string
  balance: number
  closing: number
  opening: number
  as_of: string
}

export default function ELBalancePage() {
  const [companies, setCompanies] = React.useState<{ id: string; company_name_en: string }[]>([])
  const [companyId, setCompanyId] = React.useState("")
  const [employeeId, setEmployeeId] = React.useState("")
  const [asOf, setAsOf] = React.useState(new Date().toISOString().slice(0, 10))
  const [data, setData] = React.useState<BalanceRow[]>([])
  const [loading, setLoading] = React.useState(false)

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((r) => {
      const list = r.data?.data || r.data || []
      setCompanies(Array.isArray(list) ? list : [])
      if (Array.isArray(list) && list.length > 0) setCompanyId(list[0].id)
    })
  }, [])

  const cols: ColumnDef<BalanceRow>[] = [
    { accessorKey: "employee_id", header: "Employee ID" },
    { accessorKey: "opening", header: "Opening" },
    { accessorKey: "balance", header: "Closing Balance", cell: ({ row }) => <span className="font-semibold">{row.original.closing ?? row.original.balance}</span> },
    { accessorKey: "as_of", header: "As Of" },
  ]

  const handleSearch = async () => {
    if (!companyId || !employeeId) { toast.error("Company and Employee ID required"); return }
    setLoading(true)
    try {
      const res = await earnedLeaveApi.balance({ company_id: companyId, employee_id: employeeId, as_of: asOf })
      const row = res.data
      setData(row ? [row] : [])
      if (!row || row.closing === 0) toast.info("No balance found (0)")
    } catch (e: any) { toast.error(e?.response?.data?.error || "Failed"); setData([]) }
    finally { setLoading(false) }
  }

  return (
    <div className="flex flex-col gap-4 py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <WalletIcon className="h-6 w-6 text-indigo-600" />
        <div>
          <h1 className="text-2xl font-bold tracking-tight">EL Balance</h1>
          <p className="text-sm text-muted-foreground">Ledger-driven closing balance (opening + accrued - used - encashed)</p>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader><CardTitle className="text-sm">Query</CardTitle></CardHeader>
          <CardContent className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Company *</label><select value={companyId} onChange={(e) => setCompanyId(e.target.value)} className="flex h-10 rounded-md border px-3 text-sm bg-background"><option value="">Select</option>{companies.map((c) => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}</select></div>
            <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Employee ID *</label><input value={employeeId} onChange={(e) => setEmployeeId(e.target.value)} placeholder="EMP001" className="flex h-10 rounded-md border px-3 text-sm" /></div>
            <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">As Of</label><input type="date" value={asOf} onChange={(e) => setAsOf(e.target.value)} className="flex h-10 rounded-md border px-3 text-sm" /></div>
            <div className="flex items-end"><Button onClick={handleSearch} disabled={loading} className="w-full sm:w-auto">{loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <SearchIcon className="mr-2 h-4 w-4" />}Search</Button></div>
          </CardContent>
        </Card>
      </div>

      <div className="px-4 lg:px-6">
        <DataTable data={data.map((d) => ({ ...d, id: d.employee_id }))} columns={cols as any} loading={loading} enableSelection={false} />
      </div>
    </div>
  )
}
