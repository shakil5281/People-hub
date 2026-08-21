"use client"

import * as React from "react"
import { UsersIcon, PlusIcon, Loader2, FilterIcon, XIcon, MoonIcon } from "lucide-react"
import { useRouter } from "next/navigation"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { nightBillEmployeeListApi, companyApi } from "@/lib/api"
import { cn } from "@/lib/utils"

type NightBillEmployeeListRecord = {
  id: string
  company_id: string
  employee_id: string
  bill_type: "fixed" | "hourly"
  fixed_amount: number
  hourly_rate: number
  is_active: boolean
  employee?: { name_en?: string }
}

const billTypeBadge = (type: string) => (
  <Badge variant={type === "fixed" ? "default" : "secondary"} className="capitalize">{type}</Badge>
)

const selectClass = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"

export default function NightBillEmployeeListPage() {
  const router = useRouter()
  const [data, setData] = React.useState<NightBillEmployeeListRecord[]>([])
  const [loading, setLoading] = React.useState(true)
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editing, setEditing] = React.useState<NightBillEmployeeListRecord | null>(null)
  const [submitting, setSubmitting] = React.useState(false)
  const [deleting, setDeleting] = React.useState<NightBillEmployeeListRecord | null>(null)
  const [companies, setCompanies] = React.useState<Array<{ id: string; company_name_en: string }>>([])
  const [filters, setFilters] = React.useState<Record<string, string>>({})
  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  const [form, setForm] = React.useState<{
    company_id: string
    employee_id: string
    bill_type: "fixed" | "hourly"
    amount: string
  }>({ company_id: "", employee_id: "", bill_type: "fixed", amount: "100" })

  const filterDefs: FilterDef[] = React.useMemo(() => [
    {
      key: "company_id", label: "Company", type: "select",
      options: companies.map((c) => ({ value: c.id, label: c.company_name_en })),
    },
  ], [companies])

  const columns: ColumnDef<NightBillEmployeeListRecord>[] = React.useMemo(() => [
    { accessorKey: "employee_id", header: "Employee ID" },
    {
      header: "Employee",
      accessorFn: (r) => r.employee?.name_en || "-",
    },
    {
      accessorKey: "bill_type",
      header: "Bill Type",
      cell: ({ row }) => billTypeBadge(row.original.bill_type),
    },
    {
      header: "Amount (Tk)",
      accessorFn: (r) => r.bill_type === "fixed" ? r.fixed_amount : r.hourly_rate,
      cell: ({ row }) => row.original.bill_type === "fixed"
        ? `৳${row.original.fixed_amount}`
        : `৳${row.original.hourly_rate}/hr`,
    },
  ], [])

  const fetchData = React.useCallback(async (f: Record<string, string>, p?: number, l?: number) => {
    setLoading(true)
    try {
      const params: Record<string, string> = { page: String(p ?? page), limit: String(l ?? limit) }
      if (f.company_id) params.company_id = f.company_id
      const { data: res } = await nightBillEmployeeListApi.list(params)
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setData([])
      toast.error("Failed to load night bill employee list")
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
    setForm({ company_id: "", employee_id: "", bill_type: "fixed", amount: "100" })
    setEditing(null)
  }

  const handleCreate = async () => {
    if (!form.employee_id) { toast.error("Employee ID is required"); return }
    const amount = Number(form.amount) || 0
    if (amount <= 0) { toast.error("Amount must be greater than 0"); return }
    setSubmitting(true)
    try {
      await nightBillEmployeeListApi.create({
        company_id: form.company_id,
        employee_id: form.employee_id,
        bill_type: form.bill_type,
        fixed_amount: form.bill_type === "fixed" ? amount : 0,
        hourly_rate: form.bill_type === "hourly" ? amount : 0,
      })
      toast.success("Employee added to night bill list")
      setDialogOpen(false)
      resetForm()
      fetchData(filters, 1)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      toast.error(msg || "Failed to add employee")
    } finally { setSubmitting(false) }
  }

  const handleUpdate = async () => {
    if (!editing) return
    const amount = Number(form.amount) || 0
    setSubmitting(true)
    try {
      await nightBillEmployeeListApi.update(editing.id, {
        bill_type: form.bill_type,
        fixed_amount: form.bill_type === "fixed" ? amount : 0,
        hourly_rate: form.bill_type === "hourly" ? amount : 0,
      })
      toast.success("Updated successfully")
      setDialogOpen(false)
      resetForm()
      fetchData(filters, page)
    } catch {
      toast.error("Failed to update")
    } finally { setSubmitting(false) }
  }

  const handleDelete = async () => {
    if (!deleting) return
    try {
      await nightBillEmployeeListApi.delete(deleting.id)
      toast.success("Removed from night bill employee list")
      setDeleting(null)
      fetchData(filters, page)
    } catch {
      toast.error("Failed to delete")
    }
  }

  const openCreate = () => {
    resetForm()
    if (companies.length > 0) setForm((p) => ({ ...p, company_id: companies[0].id }))
    setDialogOpen(true)
  }

  const openEdit = (row: NightBillEmployeeListRecord) => {
    setEditing(row)
    setForm({
      company_id: row.company_id,
      employee_id: row.employee_id,
      bill_type: row.bill_type,
      amount: String(row.bill_type === "fixed" ? row.fixed_amount : row.hourly_rate),
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
    setFilters({})
    fetchData({}, 1, 20)
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <UsersIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Employee Night Bill List</h1>
            <p className="text-muted-foreground mt-1">Manage employees eligible for night bill</p>
          </div>
        </div>
        <div className="hidden md:flex items-center gap-2">
          <Button variant="outline" onClick={() => router.push("/attendance/night-bill")}>
            <MoonIcon className="mr-2 h-4 w-4" />
            Night Bill
          </Button>
          <Dialog
            open={dialogOpen}
            onOpenChange={(open) => { setDialogOpen(open); if (!open) resetForm() }}
          >
            <DialogTrigger asChild>
              <Button onClick={openCreate}>
                <PlusIcon className="mr-2 h-4 w-4" />
                Add Employee
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>{editing ? "Edit Employee" : "Add Employee"}</DialogTitle>
              </DialogHeader>
              <div className="space-y-3">
                <div className="flex flex-col gap-1.5">
                  <Label>Company</Label>
                  <select
                    className={selectClass}
                    value={form.company_id}
                    onChange={(e) => setForm((p) => ({ ...p, company_id: e.target.value }))}
                  >
                    <option value="">Select company</option>
                    {companies.map((c) => (
                      <option key={c.id} value={c.id}>{c.company_name_en}</option>
                    ))}
                  </select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label>Employee ID *</Label>
                  <Input
                    value={form.employee_id}
                    onChange={(e) => setForm((p) => ({ ...p, employee_id: e.target.value }))}
                    placeholder="e.g. 12"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label>Bill Type</Label>
                  <ButtonGroup className="w-full">
                    {(["fixed", "hourly"] as const).map((t) => (
                      <Button
                        key={t}
                        type="button"
                        onClick={() => setForm((p) => ({ ...p, bill_type: t }))}
                        className={cn("flex-1", form.bill_type === t && "ring-1 ring-ring")}
                        variant={form.bill_type === t ? "default" : "outline"}
                      >
                        {t === "fixed" ? "Fixed" : "Hourly"}
                      </Button>
                    ))}
                  </ButtonGroup>
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label>{form.bill_type === "fixed" ? "Fixed Amount (Tk)" : "Hourly Rate (Tk / hour)"}</Label>
                  <Input
                    type="number"
                    step="0.01"
                    min="0"
                    value={form.amount}
                    onChange={(e) => setForm((p) => ({ ...p, amount: e.target.value }))}
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-4">
                <Button variant="outline" onClick={() => { setDialogOpen(false); resetForm() }}>Cancel</Button>
                <Button onClick={editing ? handleUpdate : handleCreate} disabled={submitting}>
                  {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  {editing ? "Update" : "Add"}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <div className="md:hidden px-4 lg:px-6">
        <ButtonGroup className="w-full">
          <Button variant="outline" className="flex-1" onClick={() => router.push("/attendance/night-bill")}>
            <MoonIcon className="mr-2 h-4 w-4" />
            Night Bill
          </Button>
          <Sheet open={mobileFilterOpen} onOpenChange={setMobileFilterOpen}>
            <SheetTrigger asChild>
              <Button variant="outline" className="flex-1">
                <FilterIcon className="mr-2 h-4 w-4" />
                Filters
              </Button>
            </SheetTrigger>            <SheetContent side="right" className="w-full sm:max-w-md p-0 flex flex-col" showCloseButton={false}>
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
              <Button className="flex-1" onClick={openCreate}>
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
        onDelete={setDeleting}
        serverSide={true}
        page={page}
        pageSize={limit}
        pageCount={totalPages}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={(size) => { setLimit(size); setPage(1) }}
        loading={loading}
      />

      <AlertDialog open={!!deleting} onOpenChange={(open) => !open && setDeleting(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Employee</AlertDialogTitle>
            <AlertDialogDescription>
              This will remove {deleting?.employee?.name_en || deleting?.employee_id} from the night bill list.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}