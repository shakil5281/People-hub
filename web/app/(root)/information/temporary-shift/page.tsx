"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { TimerIcon, PlusIcon, Trash2Icon } from "lucide-react"
import { format } from "date-fns"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { toast } from "sonner"
import { FilterBar, FilterDef } from "@/components/filter-bar"
import { TempShift, getTempShifts, updateTempShift, deleteTempShift, bulkDeleteTempShifts, TempShiftFormData } from "@/components/data/temporary-shift-data"
import { TempShiftForm } from "@/components/form/temporary-shift-form"
import { companyApi, shiftApi, departmentApi, sectionApi, designationApi, lineApi, groupApi } from "@/lib/api"

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  active: "default",
  inactive: "secondary",
}

const columns: ColumnDef<TempShift>[] = [
  {
    accessorKey: "employee",
    header: "Employee",
    cell: ({ row }) => row.original.employee?.name_en || "-",
  },
  {
    accessorKey: "employee_id",
    header: "Code",
    cell: ({ row }) => row.original.employee?.employee_id || "-",
  },
  {
    accessorKey: "shift",
    header: "Assigned Shift",
    cell: ({ row }) => row.original.shift?.name || "-",
  },
  { accessorKey: "date", header: "Date", cell: ({ row }) => row.original.date ? format(new Date(row.original.date), "dd-MM-yyyy") : "-" },
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

export default function TemporaryShiftPage() {
  const router = useRouter()
  const [data, setData] = React.useState<TempShift[]>([])
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editing, setEditing] = React.useState<TempShift | null>(null)
  const [companyId, setCompanyId] = React.useState("")

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)
  const [loading, setLoading] = React.useState(false)

  const [filterValues, setFilterValues] = React.useState<Record<string, string>>({})
  const [appliedFilters, setAppliedFilters] = React.useState<Record<string, string>>({})
  const [selectedRows, setSelectedRows] = React.useState<TempShift[]>([])
  const [shifts, setShifts] = React.useState<{ id: string; name: string }[]>([])
  const [departments, setDepartments] = React.useState<{ id: string; name: string }[]>([])
  const [sections, setSections] = React.useState<{ id: string; name: string }[]>([])
  const [designations, setDesignations] = React.useState<{ id: string; name: string }[]>([])
  const [lines, setLines] = React.useState<{ id: string; name: string }[]>([])
  const [groups, setGroups] = React.useState<{ id: string; name: string }[]>([])

  const filterDefs: FilterDef[] = [
    { key: "employee_id", label: "Employee Code", type: "text", placeholder: "EMP001" },
    { key: "search", label: "Search", type: "text", placeholder: "Code or reason..." },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    { key: "section_id", label: "Section", type: "select", options: sections.map((s) => ({ value: s.id, label: s.name })) },
    { key: "designation_id", label: "Designation", type: "select", options: designations.map((d) => ({ value: d.id, label: d.name })) },
    { key: "line_id", label: "Line", type: "select", options: lines.map((l) => ({ value: l.id, label: l.name })) },
    { key: "group_id", label: "Group", type: "select", options: groups.map((g) => ({ value: g.id, label: g.name })) },
    { key: "shift_id", label: "Shift", type: "select", options: shifts.map((s) => ({ value: s.id, label: s.name })) },
    { key: "status", label: "Status", type: "select", options: [{ value: "active", label: "Active" }, { value: "inactive", label: "Inactive" }] },
    { key: "dateRange", label: "Date Range", type: "daterange", dateRangeKeys: { start: "from_date", end: "to_date" } },
  ]

  const refreshData = React.useCallback(async (p?: number, l?: number, filters?: Record<string, string>) => {
    const f = filters ?? appliedFilters
    setLoading(true)
    try {
      const params: Record<string, string> = { page: String(p ?? page), limit: String(l ?? limit), ...f }
      Object.keys(params).forEach((k) => { if (!params[k]) delete params[k] })
      const result = await getTempShifts(companyId || undefined, params)
      setData(result.data)
      setTotal(result.total)
      setTotalPages(result.total_pages)
    } catch {
      toast.error("Failed to load temporary shifts")
    } finally {
      setLoading(false)
    }
  }, [companyId, page, limit, appliedFilters])

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((res) => {
      const list = Array.isArray(res.data?.data) ? res.data.data : []
      const cid = list[0]?.id || ""
      setCompanyId(cid)
    })
    shiftApi.list({ limit: "100" }).then(({ data }) => {
      const payload = data as unknown as { data?: { id: string; name: string }[] }
      setShifts(Array.isArray(payload?.data) ? payload.data : [])
    }).catch(() => setShifts([]))
    departmentApi.list({ limit: "100" }).then(({ data }) => {
      const payload = data as unknown as { data?: { id: string; name: string }[] }
      setDepartments(Array.isArray(payload?.data) ? payload.data : [])
    }).catch(() => setDepartments([]))
    groupApi.list({ limit: "100" }).then(({ data }) => {
      const payload = data as unknown as { data?: { id: string; name: string }[] }
      setGroups(Array.isArray(payload?.data) ? payload.data : [])
    }).catch(() => setGroups([]))
  }, [])

  React.useEffect(() => {
    const deptId = filterValues.department_id || appliedFilters.department_id
    if (!deptId) { setSections([]); return }
    sectionApi.list(deptId, { limit: "100" }).then(({ data }) => {
      const payload = data as unknown as { data?: { id: string; name: string }[] }
      setSections(Array.isArray(payload?.data) ? payload.data : [])
    }).catch(() => setSections([]))
  }, [filterValues.department_id, appliedFilters.department_id])

  React.useEffect(() => {
    const secId = filterValues.section_id || appliedFilters.section_id
    if (!secId) { setDesignations([]); setLines([]); return }
    designationApi.list(secId, { limit: "100" }).then(({ data }) => {
      const payload = data as unknown as { data?: { id: string; name: string }[] }
      setDesignations(Array.isArray(payload?.data) ? payload.data : [])
    }).catch(() => setDesignations([]))
    lineApi.list(secId, { limit: "100" }).then(({ data }) => {
      const payload = data as unknown as { data?: { id: string; name: string }[] }
      setLines(Array.isArray(payload?.data) ? payload.data : [])
    }).catch(() => setLines([]))
  }, [filterValues.section_id, appliedFilters.section_id])

  React.useEffect(() => {
    if (companyId) {
      refreshData()
    }
  }, [companyId])

  React.useEffect(() => {
    if (companyId) {
      refreshData(page, limit, appliedFilters)
    }
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

  const handleAdd = () => { router.push("/information/temporary-shift/create") }
  const handleEdit = (item: TempShift) => { setEditing(item); setDialogOpen(true) }
  const handleDelete = async (item: TempShift) => {
    const ok = await deleteTempShift(item.id)
    if (ok) { toast.success("Temporary shift deleted — employee reverts to roster, else default shift"); setSelectedRows([]); refreshData() } else toast.error("Delete failed")
  }

  const handleBulkDelete = async () => {
    if (selectedRows.length === 0) { toast.error("No rows selected"); return }
    const ids = selectedRows.map((r) => r.id)
    try {
      const result = await bulkDeleteTempShifts(ids)
      if (result.ok) {
        toast.success(`${result.count} temporary shift(s) deleted — ${result.fallback || "employees revert to roster, else default shift"}`)
        setSelectedRows([])
        refreshData()
      } else toast.error("Bulk delete failed")
    } catch {
      toast.error("Bulk delete failed")
    }
  }
  const handleFormSuccess = async (formData: TempShiftFormData) => {
    // Only Edit uses dialog now; Add goes to /create page (single-day bulk)
    const ok = await updateTempShift(editing!.id, formData)
    if (ok) { toast.success("Temporary shift updated"); refreshData(); setDialogOpen(false); setEditing(null) }
    else toast.error("Update failed")
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <TimerIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Temporary Shift</h1>
            <p className="text-muted-foreground mt-1">Manage temporary shift assignments — deleting reverts the employee to roster shift, otherwise default shift</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {selectedRows.length > 0 && (
            <Button variant="destructive" size="sm" onClick={handleBulkDelete}>
              <Trash2Icon className="mr-2 h-4 w-4" /> Delete ({selectedRows.length})
            </Button>
          )}
          <Button onClick={handleAdd}>
            <PlusIcon className="mr-2 h-4 w-4" />
            Add
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
        onPageSizeChange={(size) => { setLimit(size); setPage(1); }}
        loading={loading}
        enableSelection={true}
        enableDnd={false}
        onSelectionChange={setSelectedRows}
      />
      </div>
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editing ? "Edit Temporary Shift" : "Add Temporary Shift"}</DialogTitle>
            <DialogDescription>
              {editing ? "Update the temporary shift details." : "Fill in the temporary shift details."}
            </DialogDescription>
          </DialogHeader>
          <TempShiftForm
            initialData={editing || undefined}
            onSuccess={handleFormSuccess}
            onCancel={() => { setDialogOpen(false); setEditing(null) }}
            isEditing={!!editing}
            tempShiftId={editing?.id}
          />
        </DialogContent>
      </Dialog>
    </div>
  )
}
