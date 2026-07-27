"use client"

import * as React from "react"
import { TrendingUpIcon, Loader2, PlusIcon, CheckCircleIcon, XCircleIcon, FilterIcon, XIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { salaryIncrementApi, companyApi } from "@/lib/api"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"

interface Company { id: string; company_name_en: string }

interface IncrementRecord {
  id: string
  employee_id: string
  previous_gross: number
  increment_amount: number
  new_gross: number
  effective_date: string
  status: string
  remarks: string
  rejection_reason: string
  employee: {
    employee_id: string
    name_en: string
  }
}

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  approved: "default",
  pending: "secondary",
  rejected: "destructive",
}

export default function IncrementPage() {
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [companyId, setCompanyId] = React.useState("")
  const [data, setData] = React.useState<IncrementRecord[]>([])
  const [loading, setLoading] = React.useState(true)
  const [actionId, setActionId] = React.useState<string | null>(null)
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [formData, setFormData] = React.useState({ employee_id: "", increment_amount: "", effective_date: "", remarks: "" })
  const [creating, setCreating] = React.useState(false)
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  const fetchData = React.useCallback(async () => {
    if (!companyId) return
    setLoading(true)
    try {
      const { data: res } = await salaryIncrementApi.list({ company_id: companyId })
      setData(Array.isArray(res.increments) ? res.increments : [])
    } catch {
      toast.error("Failed to load increments")
    } finally {
      setLoading(false)
    }
  }, [companyId])

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((c) => {
      const clist = Array.isArray(c.data?.data) ? c.data.data : (Array.isArray(c.data) ? c.data : [])
      if (clist.length > 0) { setCompanies(clist); setCompanyId(clist[0].id) }
    }).catch(() => {})
  }, [])

  React.useEffect(() => { fetchData() }, [fetchData])

  const handleApprove = async (id: string) => {
    setActionId(id)
    try {
      await salaryIncrementApi.approve(id)
      toast.success("Increment approved")
      fetchData()
    } catch (err: unknown) {
      toast.error(extractError(err) || "Failed to approve")
    } finally {
      setActionId(null)
    }
  }

  const handleReject = async (id: string) => {
    setActionId(id)
    try {
      await salaryIncrementApi.reject(id)
      toast.success("Increment rejected")
      fetchData()
    } catch (err: unknown) {
      toast.error(extractError(err) || "Failed to reject")
    } finally {
      setActionId(null)
    }
  }

  const handleCreate = async () => {
    if (!companyId || !formData.employee_id || !formData.increment_amount || !formData.effective_date) {
      toast.error("Employee ID, increment amount, and effective date are required")
      return
    }
    setCreating(true)
    try {
      await salaryIncrementApi.create({
        company_id: companyId,
        employee_id: formData.employee_id,
        increment_amount: Number(formData.increment_amount),
        effective_date: formData.effective_date,
        remarks: formData.remarks,
      })
      toast.success("Increment applied")
      setDialogOpen(false)
      setFormData({ employee_id: "", increment_amount: "", effective_date: "", remarks: "" })
      fetchData()
    } catch (err: unknown) {
      toast.error(extractError(err) || "Failed to create increment")
    } finally {
      setCreating(false)
    }
  }

  const columns: ColumnDef<IncrementRecord>[] = React.useMemo(() => [
    { accessorKey: "employee.name_en", header: "Employee" },
    { accessorKey: "employee_id", header: "Code" },
    {
      accessorKey: "previous_gross",
      header: "Current Salary",
      cell: ({ row }) => row.original.previous_gross.toLocaleString(),
    },
    {
      accessorKey: "increment_amount",
      header: "Increment",
      cell: ({ row }) => <span className="text-green-600 font-semibold">+{row.original.increment_amount.toLocaleString()}</span>,
    },
    {
      accessorKey: "new_gross",
      header: "New Salary",
      cell: ({ row }) => row.original.new_gross.toLocaleString(),
    },
    { accessorKey: "effective_date", header: "Effective Date" },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const s = row.original.status
        return <Badge variant={statusVariant[s] || "secondary"} className="capitalize">{s}</Badge>
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => {
        const r = row.original
        if (r.status !== "pending") return null
        const busy = actionId === r.id
        if (busy) return <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
        return (
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleApprove(r.id) }} title="Approve">
              <CheckCircleIcon className="h-4 w-4 text-green-600" />
            </Button>
            <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleReject(r.id) }} title="Reject">
              <XCircleIcon className="h-4 w-4 text-red-600" />
            </Button>
          </div>
        )
      },
    },
  ], [actionId])

  const filterDefs: FilterDef[] = [
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en })) },
  ]

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <TrendingUpIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Increment</h1>
            <p className="text-muted-foreground mt-1">Manage salary increments</p>
          </div>
        </div>
        <div className="hidden md:block">
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button>
                <PlusIcon className="mr-2 h-4 w-4" />
                Apply Increment
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Apply Salary Increment</DialogTitle>
              </DialogHeader>
              <div className="flex flex-col gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Employee ID</label>
                  <input value={formData.employee_id} onChange={(e) => setFormData((p) => ({ ...p, employee_id: e.target.value }))} placeholder="e.g. EMP001" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Increment Amount</label>
                  <input type="number" value={formData.increment_amount} onChange={(e) => setFormData((p) => ({ ...p, increment_amount: e.target.value }))} placeholder="e.g. 3000" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Effective Date</label>
                  <input type="date" value={formData.effective_date} onChange={(e) => setFormData((p) => ({ ...p, effective_date: e.target.value }))} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Remarks (optional)</label>
                  <textarea value={formData.remarks} onChange={(e) => setFormData((p) => ({ ...p, remarks: e.target.value }))} placeholder="Reason for increment..." className="flex h-20 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" />
                </div>
                <Button onClick={handleCreate} disabled={creating}>
                  {creating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                  Apply
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <div className="md:hidden px-4 lg:px-6">
        <ButtonGroup className="w-full">
          <Sheet open={mobileFilterOpen} onOpenChange={setMobileFilterOpen}>
            <SheetTrigger asChild>
              <Button variant="outline" className="flex-1">
                <FilterIcon className="mr-2 h-4 w-4" />
                Filters
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-full sm:max-w-md p-0 flex flex-col" showCloseButton={false}>
              <SheetHeader className="px-4 py-3 border-b flex flex-row items-center justify-between">
                <SheetTitle className="text-base">Filters</SheetTitle>
                <SheetClose asChild>
                  <Button variant="ghost" size="icon-sm">
                    <XIcon className="h-4 w-4" />
                  </Button>
                </SheetClose>
              </SheetHeader>
              <div className="flex-1 overflow-y-auto px-4 py-4">
                <FilterBar
                  filters={filterDefs}
                  values={{ company_id: companyId }}
                  onChange={(key, value) => { if (key === "company_id") setCompanyId(value) }}
                  onApply={() => { fetchData(); setMobileFilterOpen(false) }}
                  onReset={() => {}}
                  submitting={false}
                  singleColumn
                  noBorder
                />
              </div>
            </SheetContent>
          </Sheet>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button className="flex-1">
                <PlusIcon className="mr-2 h-4 w-4" />
                Apply
              </Button>
            </DialogTrigger>
          </Dialog>
        </ButtonGroup>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar
          filters={filterDefs}
          values={{ company_id: companyId }}
          onChange={(key, value) => { if (key === "company_id") setCompanyId(value) }}
          onApply={() => fetchData()}
          onReset={() => {}}
          submitting={false}
        />
      </div>

      <DataTable data={data} columns={columns} loading={loading} />
    </div>
  )
}

function extractError(err: unknown): string {
  if (typeof err === "object" && err !== null && "response" in err) {
    const ae = err as { response?: { data?: { error?: string } } }
    return ae.response?.data?.error || ""
  }
  return ""
}
