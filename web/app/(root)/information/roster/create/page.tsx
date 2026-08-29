"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { CalendarRangeIcon, ArrowLeftIcon, UsersIcon } from "lucide-react"
import { format } from "date-fns"
import { toast } from "sonner"
import type { ColumnDef } from "@tanstack/react-table"
import { DataTable } from "@/components/table/data-table"
import { FilterBar, FilterDef } from "@/components/filter-bar"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { DatePicker } from "@/components/ui/date-picker"
import { Badge } from "@/components/ui/badge"
import { employeeApi, companyApi, shiftApi, departmentApi, sectionApi, designationApi, lineApi, groupApi } from "@/lib/api"
import { bulkCreateRoster } from "@/components/data/roster-data"

interface EmployeeRow {
  id: string
  employee_id: string
  name_en: string
  departmentRef?: { name: string }
  sectionRef?: { name: string }
  designationRef?: { name: string }
  lineRef?: { name: string }
  groupRef?: { name: string }
  shift?: { name: string }
  status: string
}

export default function RosterCreatePage() {
  const router = useRouter()
  const [companyId, setCompanyId] = React.useState("")
  const [loading, setLoading] = React.useState(false)
  const [employees, setEmployees] = React.useState<EmployeeRow[]>([])
  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)
  const [selectedRows, setSelectedRows] = React.useState<EmployeeRow[]>([])
  const [dialogOpen, setDialogOpen] = React.useState(false)

  // filter values
  const [filterValues, setFilterValues] = React.useState<Record<string, string>>({})
  const [appliedFilters, setAppliedFilters] = React.useState<Record<string, string>>({})

  // options for selects
  const [departments, setDepartments] = React.useState<{ id: string; name: string }[]>([])
  const [sections, setSections] = React.useState<{ id: string; name: string }[]>([])
  const [designations, setDesignations] = React.useState<{ id: string; name: string }[]>([])
  const [lines, setLines] = React.useState<{ id: string; name: string }[]>([])
  const [groups, setGroups] = React.useState<{ id: string; name: string }[]>([])
  const [shifts, setShifts] = React.useState<{ id: string; name: string }[]>([])

  // dialog form state
  const [shiftId, setShiftId] = React.useState("")
  const [fromDate, setFromDate] = React.useState<Date | undefined>()
  const [toDate, setToDate] = React.useState<Date | undefined>()
  const [reason, setReason] = React.useState("")
  const [submitting, setSubmitting] = React.useState(false)

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((res) => {
      const list = (res.data as unknown as { data?: { id: string }[] })?.data || (Array.isArray(res.data) ? (res.data as unknown as { id: string }[]) : [])
      const cid = (list as { id: string }[])[0]?.id || ""
      setCompanyId(cid)
    })
    // load filter options — use same pattern as employee-form (destructure {data}, Array.isArray(data.data))
    departmentApi.list({ limit: "100" }).then(({ data }) => setDepartments(Array.isArray((data as unknown as { data?: unknown[] }).data) ? (data as unknown as { data: { id: string; name: string }[] }).data : [])).catch(() => {})
    groupApi.list({ limit: "100" }).then(({ data }) => setGroups(Array.isArray((data as unknown as { data?: unknown[] }).data) ? (data as unknown as { data: { id: string; name: string }[] }).data : [])).catch(() => {})
    shiftApi.list({ limit: "100" }).then(({ data }) => setShifts(Array.isArray((data as unknown as { data?: unknown[] }).data) ? (data as unknown as { data: { id: string; name: string }[] }).data : [])).catch(() => {})
  }, [])

  // cascading section / designation / line
  React.useEffect(() => {
    const deptId = filterValues.department_id || appliedFilters.department_id
    if (!deptId) { setSections([]); return }
    sectionApi.list(deptId, { limit: "100" }).then(({ data }) => setSections(Array.isArray((data as unknown as { data?: unknown[] }).data) ? (data as unknown as { data: { id: string; name: string }[] }).data : [])).catch(() => setSections([]))
  }, [filterValues.department_id, appliedFilters.department_id])

  React.useEffect(() => {
    const secId = filterValues.section_id || appliedFilters.section_id
    if (!secId) { setDesignations([]); setLines([]); return }
    designationApi.list(secId, { limit: "100" }).then(({ data }) => setDesignations(Array.isArray((data as unknown as { data?: unknown[] }).data) ? (data as unknown as { data: { id: string; name: string }[] }).data : [])).catch(() => setDesignations([]))
    lineApi.list(secId, { limit: "100" }).then(({ data }) => setLines(Array.isArray((data as unknown as { data?: unknown[] }).data) ? (data as unknown as { data: { id: string; name: string }[] }).data : [])).catch(() => setLines([]))
  }, [filterValues.section_id, appliedFilters.section_id])

  const fetchEmployees = React.useCallback(async (p?: number, l?: number, filters?: Record<string, string>) => {
    const f = filters ?? appliedFilters
    const params: Record<string, string> = {
      page: String(p ?? page),
      limit: String(l ?? limit),
      status: "active",
      ...f,
    }
    if (companyId) params.company_id = companyId
    Object.keys(params).forEach((k) => { if (!params[k]) delete params[k] })
    setLoading(true)
    try {
      const res = await employeeApi.list(params)
      const payload = res.data as unknown as { data?: EmployeeRow[]; total?: number; total_pages?: number }
      const list: EmployeeRow[] = Array.isArray(payload?.data) ? payload.data : Array.isArray(res.data) ? (res.data as unknown as EmployeeRow[]) : []
      setEmployees(list)
      setTotal(payload?.total ?? 0)
      setTotalPages(payload?.total_pages ?? 0)
    } catch {
      toast.error("Failed to load employees")
    } finally {
      setLoading(false)
    }
  }, [companyId, page, limit, appliedFilters])

  React.useEffect(() => {
    if (companyId) fetchEmployees()
  }, [companyId])

  React.useEffect(() => {
    if (companyId) fetchEmployees(page, limit, appliedFilters)
  }, [page, limit])

  const filterDefs: FilterDef[] = [
    { key: "search", label: "Search", type: "text", placeholder: "Name, code, phone..." },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    { key: "section_id", label: "Section", type: "select", options: sections.map((s) => ({ value: s.id, label: s.name })) },
    { key: "designation_id", label: "Designation", type: "select", options: designations.map((d) => ({ value: d.id, label: d.name })) },
    { key: "line_id", label: "Line", type: "select", options: lines.map((l) => ({ value: l.id, label: l.name })) },
    { key: "group_id", label: "Group", type: "select", options: groups.map((g) => ({ value: g.id, label: g.name })) },
  ]

  const handleApply = () => {
    setAppliedFilters({ ...filterValues })
    setPage(1)
    fetchEmployees(1, limit, filterValues)
  }
  const handleReset = () => {
    setFilterValues({})
    setAppliedFilters({})
    setPage(1)
    fetchEmployees(1, limit, {})
  }

  const handleOpenDialog = () => {
    if (selectedRows.length === 0) {
      toast.error("Select at least one employee")
      return
    }
    setDialogOpen(true)
  }

  const handleSubmit = async () => {
    if (!shiftId) { toast.error("Shift is required"); return }
    if (!fromDate) { toast.error("From date is required"); return }
    const employee_ids = selectedRows.map((r) => r.employee_id)
    const payload: Record<string, unknown> = {
      company_id: companyId,
      shift_id: shiftId,
      from_date: format(fromDate, "yyyy-MM-dd"),
      reason,
      employee_ids,
    }
    if (toDate) payload.to_date = format(toDate, "yyyy-MM-dd")
    setSubmitting(true)
    try {
      const ok = await bulkCreateRoster(payload)
      if (ok) {
        toast.success(`Roster assigned to ${employee_ids.length} employees`)
        setDialogOpen(false)
        setSelectedRows([])
        router.push("/information/roster")
      } else toast.error("Failed to assign roster")
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: string } } }
      toast.error(axiosErr.response?.data?.error || "Failed to assign roster")
    } finally {
      setSubmitting(false)
    }
  }

  const columns: ColumnDef<EmployeeRow>[] = [
    { accessorKey: "employee_id", header: "Code" },
    { accessorKey: "name_en", header: "Name", cell: ({ row }) => <span className="font-medium">{row.original.name_en}</span> },
    { accessorKey: "department", header: "Department", cell: ({ row }) => row.original.departmentRef?.name || "-" },
    { accessorKey: "section", header: "Section", cell: ({ row }) => row.original.sectionRef?.name || "-" },
    { accessorKey: "designation", header: "Designation", cell: ({ row }) => row.original.designationRef?.name || "-" },
    { accessorKey: "line", header: "Line", cell: ({ row }) => row.original.lineRef?.name || "-" },
    { accessorKey: "group", header: "Group", cell: ({ row }) => row.original.groupRef?.name || "-" },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => <Badge variant={row.original.status === "active" ? "default" : "secondary"}>{row.original.status}</Badge>,
    },
  ]

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" onClick={() => router.back()}>
            <ArrowLeftIcon className="h-5 w-5" />
          </Button>
          <CalendarRangeIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Add Roster</h1>
            <p className="text-muted-foreground mt-1">Advance filter → select active employees → assign shift via dialog</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="px-3 py-1">
            <UsersIcon className="mr-2 h-4 w-4" /> Selected: {selectedRows.length} / {total}
          </Badge>
          <Button
            variant="outline"
            disabled={loading || total === 0 || selectedRows.length === total}
            onClick={async () => {
              const params: Record<string, string> = {
                status: "active",
                limit: String(total || 5000),
                page: "1",
                ...appliedFilters,
              }
              if (companyId) params.company_id = companyId
              Object.keys(params).forEach((k) => { if (!params[k]) delete params[k] })
              setLoading(true)
              try {
                const res = await employeeApi.list(params)
                const payload = res.data as unknown as { data?: EmployeeRow[] }
                const all: EmployeeRow[] = Array.isArray(payload?.data) ? payload.data : Array.isArray(res.data) ? (res.data as unknown as EmployeeRow[]) : []
                setSelectedRows(all)
                toast.success(`All ${all.length} employees selected`)
              } catch {
                toast.error("Failed to select all")
              } finally {
                setLoading(false)
              }
            }}
          >
            Select All ({total})
          </Button>
          {selectedRows.length > 0 && (
            <Button onClick={handleOpenDialog}>
              <CalendarRangeIcon className="mr-2 h-4 w-4" /> Assign Roster ({selectedRows.length})
            </Button>
          )}
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
          applyLabel="Search"
        />
      </div>

      <div className="px-4 lg:px-6 space-y-3">
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">Showing active employees — select via checkbox to enable roster assignment. {selectedRows.length > 0 && <span className="font-medium text-foreground">{selectedRows.length} selected — click Assign Roster to open dialog</span>}</p>
        </div>
        <DataTable
          data={employees}
          columns={columns}
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

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Assign Roster</DialogTitle>
            <DialogDescription>
              Assign shift to {selectedRows.length} selected employee(s). Uses <b>DatePicker</b> for dates.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Shift *</Label>
              <select
                value={shiftId}
                onChange={(e) => setShiftId(e.target.value)}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="">Select Shift</option>
                {shifts.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label>From Date *</Label>
                <DatePicker value={fromDate} onChange={setFromDate} placeholder="Select from date" />
              </div>
              <div className="space-y-2">
                <Label>To Date</Label>
                <DatePicker value={toDate} onChange={setToDate} placeholder="Select to date" />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Reason</Label>
              <Input placeholder="Weekly roster, rotation..." value={reason} onChange={(e) => setReason(e.target.value)} />
            </div>
            <div className="flex justify-end gap-2 pt-4 border-t">
              <Button variant="outline" onClick={() => setDialogOpen(false)} disabled={submitting}>Cancel</Button>
              <Button onClick={handleSubmit} disabled={submitting}>{submitting ? "Saving..." : "Submit Roster"}</Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
