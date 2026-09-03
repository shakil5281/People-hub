"use client"

import * as React from "react"
import { useParams } from "next/navigation"
import { Loader2, CheckIcon, ShieldIcon, RotateCcwIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { earnedLeaveApi } from "@/lib/api"

interface Item {
  id: string
  employee_id: string
  opening: number
  accrued: number
  closing: number
  encashable_days: number
  salary_basis: number
  rate: number
  amount: number
  employee?: { name_en: string; department?: { name: string } }
}

export default function ELSalarySheetDetailPage() {
  const params = useParams<{ id: string }>()
  const id = params.id as string
  const [sheet, setSheet] = React.useState<any>(null)
  const [items, setItems] = React.useState<Item[]>([])
  const [loading, setLoading] = React.useState(true)

  const fetch = async () => {
    setLoading(true)
    try {
      const res = await earnedLeaveApi.getSalarySheet(id)
      setSheet(res.data?.sheet || res.data)
      setItems(res.data?.items || [])
    } catch { toast.error("Failed to load sheet") }
    finally { setLoading(false) }
  }

  React.useEffect(() => { fetch() }, [id])

  const cols: ColumnDef<Item>[] = [
    { accessorKey: "employee_id", header: "Employee ID" },
    { accessorKey: "employee", header: "Name", cell: ({ row }) => row.original.employee?.name_en || "-" },
    { accessorKey: "encashable_days", header: "EL Days" },
    { accessorKey: "salary_basis", header: "Basis", cell: ({ row }) => `৳${Number(row.original.salary_basis || 0).toLocaleString()}` },
    { accessorKey: "rate", header: "Rate" },
    { accessorKey: "amount", header: "Amount", cell: ({ row }) => <span className="font-semibold">৳{Number(row.original.amount || 0).toLocaleString()}</span> },
  ]

  const handle = async (action: "approve" | "finalize" | "reverse") => {
    try {
      if (action === "approve") await earnedLeaveApi.approveSheet(id)
      if (action === "finalize") await earnedLeaveApi.finalizeSheet(id)
      if (action === "reverse") await earnedLeaveApi.reverseSheet(id)
      toast.success(`${action} successful`)
      fetch()
    } catch (e: any) { toast.error(e?.response?.data?.error || "Failed") }
  }

  if (loading) return <div className="p-6 flex justify-center"><Loader2 className="h-6 w-6 animate-spin" /></div>
  if (!sheet) return <div className="p-6 text-center text-muted-foreground">Sheet not found</div>

  return (
    <div className="flex flex-col gap-4 py-6">
      <div className="px-4 lg:px-6">
        <h1 className="text-2xl font-bold tracking-tight">EL Salary Sheet — {sheet.period}</h1>
        <p className="text-sm text-muted-foreground">Status: <b>{sheet.status}</b> • Employees: {sheet.total_employees} • Total: ৳{Number(sheet.total_amount || 0).toLocaleString()}</p>
      </div>

      <div className="px-4 lg:px-6 flex gap-2">
        <Button variant="outline" onClick={() => handle("approve")} disabled={sheet.status !== "CALCULATED"}><CheckIcon className="mr-2 h-4 w-4" />Approve</Button>
        <Button onClick={() => handle("finalize")} disabled={sheet.status !== "APPROVED"}><ShieldIcon className="mr-2 h-4 w-4" />Finalize</Button>
        <Button variant="destructive" onClick={() => handle("reverse")} disabled={sheet.status === "REVERSED"}><RotateCcwIcon className="mr-2 h-4 w-4" />Reverse</Button>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader><CardTitle className="text-sm">Items ({items.length})</CardTitle></CardHeader>
          <CardContent className="p-0">
            <DataTable data={items} columns={cols} enableSelection={false} />
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
