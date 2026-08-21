"use client"

import * as React from "react"
import { UtensilsCrossedIcon, PlusIcon, Loader2, FilterIcon, XIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { tiffinBillApi, companyApi } from "@/lib/api"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"

interface TiffinBill {
  id: string
  employee_id: string
  date: string
  amount: number
  month: number
  year: number
  status: string
  employee?: { name_en: string }
}

const monthNames = ["January","February","March","April","May","June","July","August","September","October","November","December"]
const currentYear = new Date().getFullYear()

const statusBadge = (status: string) => {
  const variant = status === "active" ? "default" : status === "inactive" ? "secondary" : "outline"
  return <Badge variant={variant} className="capitalize">{status}</Badge>
}

const selectClass = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"

export default function TiffinBillPage() {
  const [data, setData] = React.useState<TiffinBill[]>([])
  const [loading, setLoading] = React.useState(true)
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editing, setEditing] = React.useState<TiffinBill | null>(null)
  const [submitting, setSubmitting] = React.useState(false)
  const [companies, setCompanies] = React.useState<Array<{ id: string; company_name_en: string }>>([])
  const [filters, setFilters] = React.useState<Record<string, string>>({
    month: String(new Date().getMonth() + 1),
    year: String(currentYear),
  })
  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const [form, setForm] = React.useState({ employee_id: "", date: "", amount: "" })
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  const filterDefs: FilterDef[] = React.useMemo(() => [
    {
      key: "company_id", label: "Company", type: "select",
      options: companies.map((c) => ({ value: c.id, label: c.company_name_en })),
    },
    {
      key: "month", label: "Month", type: "select",
      options: monthNames.map((name, i) => ({ value: String(i + 1), label: name })),
    },
    {
      key: "year", label: "Year", type: "select",
      options: Array.from({ length: 5 }, (_, i) => {
        const y = currentYear - 2 + i
        return { value: String(y), label: String(y) }
      }),
    },
  ], [companies])

  const columns: ColumnDef<TiffinBill>[] = React.useMemo(() => [
    { accessorKey: "employee_id", header: "Employee ID" },
    {
      header: "Employee",
      accessorFn: (r) => r.employee?.name_en || "-",
    },
    { accessorKey: "date", header: "Date" },
    { accessorKey: "amount", header: "Amount", cell: ({ row }) => `৳${row.original.amount}` },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => statusBadge(row.original.status),
    },
  ], [])

  const fetchData = React.useCallback(async (f: Record<string, string>, p?: number, l?: number) => {
    setLoading(true)
    try {
      const params: Record<string, string> = { page: String(p ?? page), limit: String(l ?? limit) }
      if (f.company_id) params.company_id = f.company_id
      if (f.month) params.month = f.month
      if (f.year) params.year = f.year
      const { data: res } = await tiffinBillApi.list(params)
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setData([])
      toast.error("Failed to load tiffin bills")
    } finally {
      setLoading(false)
    }
  }, [page, limit])

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((res) => {
      setCompanies(Array.isArray(res.data?.data) ? res.data.data : [])
    }).catch(() => {})
    fetchData(filters)
  }, [])

  React.useEffect(() => { fetchData(filters) }, [page, limit])

  const resetForm = () => {
    setForm({ employee_id: "", date: "", amount: "" })
    setEditing(null)
  }

  const handleCreate = async () => {
    if (!form.employee_id || !form.date) { toast.error("Employee ID and date are required"); return }
    setSubmitting(true)
    try {
      await tiffinBillApi.create({
        employee_id: form.employee_id,
        date: form.date,
        amount: form.amount ? Number(form.amount) : 0,
      })
      toast.success("Tiffin bill created")
      setDialogOpen(false)
      resetForm()
      fetchData(filters, 1)
    } catch { toast.error("Failed to create tiffin bill") }
    finally { setSubmitting(false) }
  }

  const handleUpdate = async () => {
    if (!editing) return
    setSubmitting(true)
    try {
      await tiffinBillApi.update(editing.id, {
        employee_id: form.employee_id,
        date: form.date,
        amount: form.amount ? Number(form.amount) : 0,
      })
      toast.success("Tiffin bill updated")
      setDialogOpen(false)
      resetForm()
      fetchData(filters, page)
    } catch { toast.error("Failed to update tiffin bill") }
    finally { setSubmitting(false) }
  }

  const handleDelete = async (row: TiffinBill) => {
    try {
      await tiffinBillApi.delete(row.id)
      toast.success("Tiffin bill deleted")
      fetchData(filters, page)
    } catch { toast.error("Failed to delete tiffin bill") }
  }

  const openEdit = (row: TiffinBill) => {
    setEditing(row)
    setForm({
      employee_id: row.employee_id,
      date: row.date,
      amount: String(row.amount ?? ""),
    })
    setDialogOpen(true)
  }

  const handleChange = (key: string, value: string) => setFilters((prev) => ({ ...prev, [key]: value }))

  const handleApply = () => {
    setPage(1)
    const active: Record<string, string> = {}
    for (const [k, v] of Object.entries(filters)) if (v) active[k] = v
    fetchData(active, 1)
  }

  const handleReset = () => {
    setPage(1)
    setLimit(20)
    const defaults = { month: String(new Date().getMonth() + 1), year: String(currentYear) }
    setFilters(defaults)
    fetchData(defaults, 1, 20)
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <UtensilsCrossedIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Tiffin Bill</h1>
            <p className="text-muted-foreground mt-1">Manage employee tiffin bills</p>
          </div>
        </div>
        <div className="hidden md:block">
          <Dialog open={dialogOpen} onOpenChange={(open) => { setDialogOpen(open); if (!open) resetForm() }}>
            <DialogTrigger asChild>
              <Button>
                <PlusIcon className="mr-2 h-4 w-4" />
                Add Tiffin Bill
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>{editing ? "Edit Tiffin Bill" : "Add Tiffin Bill"}</DialogTitle>
              </DialogHeader>
              <div className="space-y-3">
                <div className="flex flex-col gap-1.5">
                  <Label>Employee ID *</Label>
                  <Input value={form.employee_id} onChange={(e) => setForm((p) => ({ ...p, employee_id: e.target.value }))} placeholder="EMP-001" />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label>Date *</Label>
                  <Input type="date" value={form.date} onChange={(e) => setForm((p) => ({ ...p, date: e.target.value }))} />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label>Amount</Label>
                  <Input type="number" step="0.01" value={form.amount} onChange={(e) => setForm((p) => ({ ...p, amount: e.target.value }))} placeholder="0" />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-4">
                <Button variant="outline" onClick={() => { setDialogOpen(false); resetForm() }}>Cancel</Button>
                <Button onClick={editing ? handleUpdate : handleCreate} disabled={submitting}>
                  {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  {editing ? "Update" : "Create"}
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
                <FilterBar filters={filterDefs} values={filters} onChange={handleChange} onApply={() => { handleApply(); setMobileFilterOpen(false) }} onReset={() => { handleReset(); setMobileFilterOpen(false) }} submitting={loading} singleColumn noBorder />
              </div>
            </SheetContent>
          </Sheet>
          <Dialog open={dialogOpen} onOpenChange={(open) => { setDialogOpen(open); if (!open) resetForm() }}>
            <DialogTrigger asChild>
              <Button className="flex-1">
                <PlusIcon className="mr-2 h-4 w-4" />
                Add
              </Button>
            </DialogTrigger>
          </Dialog>
        </ButtonGroup>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar filters={filterDefs} values={filters} onChange={handleChange} onApply={handleApply} onReset={handleReset} submitting={loading} />
      </div>

      <DataTable
        data={data}
        columns={columns}
        onEdit={openEdit}
        onDelete={handleDelete}
        serverSide={true}
        page={page}
        pageSize={limit}
        pageCount={totalPages}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={(size) => { setLimit(size); setPage(1) }}
        loading={loading}
      />
    </div>
  )
}
