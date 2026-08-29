"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { CalendarRangeIcon, PlusIcon, UsersIcon, Trash2Icon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { toast } from "sonner"
import { Roster, getRosters, createRoster, updateRoster, deleteRoster, RosterFormData, bulkCreateRoster } from "@/components/data/roster-data"
import { RosterForm } from "@/components/form/roster-form"
import { RosterBulkForm } from "@/components/form/roster-bulk-form"
import { BulkRosterFormData } from "@/components/data/roster-data"
import { FilterBar, FilterDef } from "@/components/filter-bar"
import { rosterApi, companyApi } from "@/lib/api"

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  active: "default",
  inactive: "secondary",
}

const columns: ColumnDef<Roster>[] = [
  {
    accessorKey: "employee",
    header: "Employee",
    cell: ({ row }) => row.original.employee?.name_en || "-",
  },
  {
    accessorKey: "employee_id",
    header: "Code",
    cell: ({ row }) => row.original.employee?.employee_id || row.original.employee_id || "-",
  },
  {
    accessorKey: "shift",
    header: "Shift",
    cell: ({ row }) => row.original.shift?.name || "-",
  },
  { accessorKey: "date", header: "Date" },
  { accessorKey: "reason", header: "Reason", cell: ({ row }) => row.original.reason || "-" },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const s = row.original.status
      return <Badge variant={statusVariant[s] || "secondary"}>{s}</Badge>
    },
  },
]

export default function RosterPage() {
  const router = useRouter()
  const [data, setData] = React.useState<Roster[]>([])
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [bulkDialogOpen, setBulkDialogOpen] = React.useState(false)
  const [editing, setEditing] = React.useState<Roster | null>(null)
  const [companyId, setCompanyId] = React.useState("")
  const [loading, setLoading] = React.useState(false)

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const [filterValues, setFilterValues] = React.useState<Record<string, string>>({})
  const [appliedFilters, setAppliedFilters] = React.useState<Record<string, string>>({})
  const [selectedRows, setSelectedRows] = React.useState<Roster[]>([])

  const filterDefs: FilterDef[] = [
    { key: "employee_id", label: "Employee ID", type: "text", placeholder: "EMP001" },
    { key: "search", label: "Search", type: "text", placeholder: "Reason or code" },
    { key: "status", label: "Status", type: "select", options: [{ value: "active", label: "Active" }, { value: "inactive", label: "Inactive" }] },
    { key: "dateRange", label: "Date Range", type: "daterange", dateRangeKeys: { start: "from_date", end: "to_date" } },
  ]

  const refreshData = React.useCallback(async (p?: number, l?: number, filters?: Record<string, string>) => {
    const f = filters ?? appliedFilters
    setLoading(true)
    try {
      const params: Record<string, string> = { page: String(p ?? page), limit: String(l ?? limit), ...f }
      // remove empty
      Object.keys(params).forEach((k) => { if (!params[k]) delete params[k] })
      const result = await getRosters(companyId || undefined, params)
      setData(result.data)
      setTotal(result.total)
      setTotalPages(result.total_pages)
    } catch {
      toast.error("Failed to load rosters")
    } finally {
      setLoading(false)
    }
  }, [companyId, page, limit, appliedFilters])

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((res) => {
      const list = (res.data as unknown as { data?: { id: string }[] })?.data || (Array.isArray(res.data) ? res.data as unknown as { id: string }[] : [])
      const cid = (list as { id: string }[])[0]?.id || ""
      setCompanyId(cid)
    })
  }, [])

  React.useEffect(() => {
    if (companyId) refreshData()
  }, [companyId])

  React.useEffect(() => {
    if (companyId) refreshData(page, limit, appliedFilters)
  }, [page, limit])

  const handleApply = () => {
    setAppliedFilters({ ...filterValues })
    setPage(1)
    refreshData(1, limit, filterValues)
  }

  const handleReset = () => {
    setFilterValues({})
    setAppliedFilters({})
    setPage(1)
    refreshData(1, limit, {})
  }

  const handleAdd = () => { router.push("/information/roster/create") }
  const handleEdit = (item: Roster) => { setEditing(item); setDialogOpen(true) }
  const handleDelete = async (item: Roster) => {
    const ok = await deleteRoster(item.id)
    if (ok) { toast.success("Roster deleted"); refreshData() } else toast.error("Delete failed")
  }

  const handleBulkDelete = async () => {
    if (selectedRows.length === 0) { toast.error("No rows selected"); return }
    const ids = selectedRows.map((r) => r.id)
    try {
      await rosterApi.bulkDelete(ids)
      toast.success(`${ids.length} rosters deleted`)
      setSelectedRows([])
      refreshData()
    } catch {
      toast.error("Bulk delete failed")
    }
  }

  const handleFormSuccess = async (formData: RosterFormData) => {
    const payload: Record<string, unknown> = { ...formData, company_id: companyId }
    let ok = false
    if (editing) {
      ok = await updateRoster(editing.id, { shift_id: formData.shift_id, date: formData.from_date, reason: formData.reason, status: formData.status })
    } else {
      ok = await createRoster(payload)
    }
    if (ok) { toast.success(editing ? "Roster updated" : "Roster created"); refreshData(); setDialogOpen(false); setEditing(null) }
    else toast.error("Save failed")
  }

  const handleBulkSuccess = async (formData: BulkRosterFormData & { company_id: string }) => {
    const payload: Record<string, unknown> = { ...formData }
    const ok = await bulkCreateRoster(payload)
    if (ok) { toast.success("Bulk roster assigned"); refreshData(); setBulkDialogOpen(false) }
    else toast.error("Bulk assign failed")
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <CalendarRangeIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Roster</h1>
            <p className="text-muted-foreground mt-1">Plan weekly/monthly shift assignments — roster overrides employee default (temporary shift still wins for emergencies)</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {selectedRows.length > 0 && (
            <Button variant="destructive" size="sm" onClick={handleBulkDelete}>
              <Trash2Icon className="mr-2 h-4 w-4" /> Delete ({selectedRows.length})
            </Button>
          )}
          <Button variant="outline" onClick={() => setBulkDialogOpen(true)}>
            <UsersIcon className="mr-2 h-4 w-4" /> Bulk Assign
          </Button>
          <Button onClick={handleAdd}>
            <PlusIcon className="mr-2 h-4 w-4" /> Add Roster
          </Button>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <FilterBar
          filters={filterDefs}
          values={filterValues}
          onChange={(k, v) => setFilterValues((prev) => ({ ...prev, [k]: v }))}
          onApply={handleApply}
          onReset={handleReset}
          submitting={loading}
        />
      </div>

      <div className="px-4 lg:px-6">
        <DataTable
          data={data}
          columns={columns}
          onEdit={handleEdit}
          onDelete={handleDelete}
          serverSide={true}
          page={page}
          pageSize={limit}
          pageCount={totalPages}
          total={total}
          onPageChange={setPage}
          onPageSizeChange={(size) => { setLimit(size); setPage(1) }}
          loading={loading}
          enableSelection={true}
          enableDnd={false}
          onSelectionChange={setSelectedRows}
        />
      </div>

      <Dialog open={dialogOpen} onOpenChange={(o) => { setDialogOpen(o); if (!o) setEditing(null) }}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editing ? "Edit Roster" : "Add Roster"}</DialogTitle>
            <DialogDescription>{editing ? "Update roster shift assignment." : "Assign a shift to an employee for a date range. Uses DatePicker."}</DialogDescription>
          </DialogHeader>
          <RosterForm
            initialData={editing || undefined}
            onSuccess={handleFormSuccess}
            onCancel={() => { setDialogOpen(false); setEditing(null) }}
            isEditing={!!editing}
          />
        </DialogContent>
      </Dialog>

      <Dialog open={bulkDialogOpen} onOpenChange={setBulkDialogOpen}>
        <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Bulk Roster Assign</DialogTitle>
            <DialogDescription>Assign a shift to multiple employees via org filter over a date range.</DialogDescription>
          </DialogHeader>
          <RosterBulkForm companyId={companyId} onSuccess={handleBulkSuccess} onCancel={() => setBulkDialogOpen(false)} />
        </DialogContent>
      </Dialog>
    </div>
  )
}
