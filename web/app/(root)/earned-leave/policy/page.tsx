"use client"

import * as React from "react"
import { BookOpenIcon, PlusIcon, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { toast } from "sonner"
import { earnedLeaveApi, companyApi } from "@/lib/api"

interface Policy {
  id: string
  name: string
  accrual_rate: number
  max_balance: number
  company_id: string
  effective_from: string
  status: string
}

export default function ELPolicyPage() {
  const [companies, setCompanies] = React.useState<{ id: string; company_name_en: string }[]>([])
  const [companyId, setCompanyId] = React.useState("")
  const [data, setData] = React.useState<Policy[]>([])
  const [loading, setLoading] = React.useState(false)
  const [open, setOpen] = React.useState(false)
  const [saving, setSaving] = React.useState(false)
  const [form, setForm] = React.useState({ name: "", accrual_rate: "1", max_balance: "40", effective_from: new Date().toISOString().slice(0, 10) })

  const fetch = React.useCallback(async (cid: string) => {
    if (!cid) return
    setLoading(true)
    try {
      const res = await earnedLeaveApi.listPolicies({ company_id: cid })
      setData(res.data?.data || res.data || [])
    } catch { toast.error("Failed to load policies") }
    finally { setLoading(false) }
  }, [])

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((r) => {
      const list = r.data?.data || r.data || []
      setCompanies(Array.isArray(list) ? list : [])
      if (Array.isArray(list) && list.length > 0) {
        setCompanyId(list[0].id)
        fetch(list[0].id)
      }
    })
  }, [fetch])

  React.useEffect(() => { if (companyId) fetch(companyId) }, [companyId, fetch])

  const cols: ColumnDef<Policy>[] = [
    { accessorKey: "name", header: "Policy Name" },
    { accessorKey: "accrual_rate", header: "Rate / Period", cell: ({ row }) => `${row.original.accrual_rate}` },
    { accessorKey: "max_balance", header: "Max Balance" },
    { accessorKey: "effective_from", header: "Effective From", cell: ({ row }) => row.original.effective_from?.slice(0, 10) || "-" },
    { accessorKey: "status", header: "Status" },
  ]

  const handleCreate = async () => {
    if (!form.name || !form.effective_from) { toast.error("Name and effective_from required"); return }
    setSaving(true)
    try {
      await earnedLeaveApi.createPolicy({
        company_id: companyId,
        name: form.name,
        accrual_rate: Number(form.accrual_rate) || 1,
        max_balance: Number(form.max_balance) || 40,
        effective_from: form.effective_from,
        accrual_frequency: "MONTHLY",
        encashment_basis: "BASIC",
        encashment_divisor: 30,
        rounding_rule: "ROUND",
      })
      toast.success("Policy created")
      setOpen(false)
      fetch(companyId)
    } catch (e: any) { toast.error(e?.response?.data?.error || "Failed") }
    finally { setSaving(false) }
  }

  return (
    <div className="flex flex-col gap-4 py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <BookOpenIcon className="h-6 w-6 text-emerald-600" />
          <div>
            <h1 className="text-2xl font-bold tracking-tight">EL Policy</h1>
            <p className="text-sm text-muted-foreground">Configurable per company — accrual, max, carry, encashment basis</p>
          </div>
        </div>
        <Button onClick={() => setOpen(true)}><PlusIcon className="mr-2 h-4 w-4" />New Policy</Button>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader className="pb-3"><CardTitle className="text-sm">Company</CardTitle></CardHeader>
          <CardContent>
            <select value={companyId} onChange={(e) => setCompanyId(e.target.value)} className="flex h-10 w-full max-w-sm rounded-md border bg-background px-3 text-sm">
              <option value="">Select Company</option>
              {companies.map((c) => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
            </select>
          </CardContent>
        </Card>
      </div>

      <div className="px-4 lg:px-6">
        <DataTable data={data} columns={cols} loading={loading} enableSelection={false} />
      </div>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader><DialogTitle>New EL Policy</DialogTitle></DialogHeader>
          <div className="grid gap-3 py-2">
            <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Policy Name *</label><input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="Monthly 1 day" className="flex h-9 rounded-md border px-3 text-sm" /></div>
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Accrual Rate *</label><input type="number" step="0.5" value={form.accrual_rate} onChange={(e) => setForm({ ...form, accrual_rate: e.target.value })} className="flex h-9 rounded-md border px-3 text-sm" /></div>
              <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Max Balance</label><input type="number" value={form.max_balance} onChange={(e) => setForm({ ...form, max_balance: e.target.value })} className="flex h-9 rounded-md border px-3 text-sm" /></div>
            </div>
            <div className="flex flex-col gap-1.5"><label className="text-xs font-medium">Effective From *</label><input type="date" value={form.effective_from} onChange={(e) => setForm({ ...form, effective_from: e.target.value })} className="flex h-9 rounded-md border px-3 text-sm" /></div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
            <Button onClick={handleCreate} disabled={saving}>{saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
