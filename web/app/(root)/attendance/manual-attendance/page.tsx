"use client"

import * as React from "react"
import { ClipboardCheckIcon, Loader2, UsersIcon, FilterIcon, XIcon } from "lucide-react"
import { format } from "date-fns"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { attendanceApi, employeeApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi } from "@/lib/api"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { DatePicker } from "@/components/ui/date-picker"
import { DateTimePicker } from "@/components/ui/date-time-picker"
import type { Company } from "@/components/data/company-data"
import type { Department, Section, Designation, Line } from "@/components/data/organization-data"
import type { Group } from "@/components/data/group-data"

interface EmployeeItem {
  id: string
  employee_id: string
  name_en: string
  name_bn: string
  punch_number: string
  company_id: string
  status: string
  employee_type: string
  department?: { id: string; name: string }
  designation_ref?: { id: string; name: string }
}

export default function ManualAttendancePage() {
  const [data, setData] = React.useState<EmployeeItem[]>([])
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [loading, setLoading] = React.useState(true)
  const [submitting, setSubmitting] = React.useState(false)
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)
  const [filters, setFilters] = React.useState<Record<string, string>>({
    status: "active",
  })
  const [selectedRows, setSelectedRows] = React.useState<EmployeeItem[]>([])

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const [entryDate, setEntryDate] = React.useState<Date>(new Date())
  const defaultIn = `${format(new Date(), "yyyy-MM-dd")}T07:55`
  const defaultOut = `${format(new Date(), "yyyy-MM-dd")}T17:00`
  const [checkIn, setCheckIn] = React.useState(defaultIn)
  const [checkOut, setCheckOut] = React.useState(defaultOut)
  const [entryStatus, setEntryStatus] = React.useState("present")

  const fetchData = React.useCallback(async (f?: Record<string, string>, p?: number, l?: number) => {
    setLoading(true)
    try {
      const params = { ...(f || {}), page: String(p ?? page), limit: String(l ?? limit) }
      const { data: res } = await employeeApi.list(params)
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      toast.error("Failed to load employees")
    } finally {
      setLoading(false)
    }
  }, [page, limit])

  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      sectionApi.list(),
      designationApi.list(),
      lineApi.list(),
      groupApi.list(),
    ]).then(([cRes, dRes, sRes, desRes, lRes, gRes]) => {
      setCompanies(Array.isArray(cRes.data?.data) ? cRes.data.data : [])
      setDepartments(Array.isArray(dRes.data?.data) ? dRes.data.data : [])
      setSections(Array.isArray(sRes.data?.data) ? sRes.data.data : [])
      setDesignations(Array.isArray(desRes.data?.data) ? desRes.data.data : [])
      setLines(Array.isArray(lRes.data?.data) ? lRes.data.data : [])
      setGroups(Array.isArray(gRes.data?.data) ? gRes.data.data : [])
    }).catch(() => {})
    fetchData(filters)
  }, [])

  React.useEffect(() => {
    fetchData(filters)
  }, [page, limit])

  const handleApply = async () => {
    setPage(1)
    const active: Record<string, string> = {}
    for (const [k, v] of Object.entries(filters)) {
      if (v) active[k] = v
    }
    await fetchData(active, 1)
  }

  const handleReset = () => {
    setPage(1)
    setLimit(20)
    const defaults: Record<string, string> = { status: "active" }
    setFilters(defaults)
    fetchData(defaults, 1, 20)
  }

  const handleChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleSubmit = async () => {
    if (!selectedRows.length) {
      toast.error("Please select employees from the table")
      return
    }

    setSubmitting(true)
    let successCount = 0
    let failCount = 0

    for (const emp of selectedRows) {
      try {
        const body: Record<string, unknown> = {
          employee_id: emp.employee_id,
          company_id: emp.company_id,
          date: format(entryDate, "yyyy-MM-dd"),
          status: entryStatus,
        }
        if (checkIn) body.check_in = checkIn
        if (checkOut) body.check_out = checkOut
        await attendanceApi.create(body)
        successCount++
      } catch {
        failCount++
      }
    }

    if (successCount > 0) {
      toast.success(`Attendance saved for ${successCount} employee(s)` + (failCount > 0 ? `, ${failCount} failed` : ""))
      setSelectedRows([])
      setEntryDate(new Date())
      setCheckIn("")
      setCheckOut("")
      setEntryStatus("present")
    } else {
      toast.error("Failed to save attendance")
    }
    setSubmitting(false)
  }

  const columns: ColumnDef<EmployeeItem>[] = React.useMemo(() => [
    { accessorKey: "employee_id", header: "Emp. ID" },
    { accessorKey: "name_en", header: "Name" },
    { accessorKey: "punch_number", header: "Punch No." },
    {
      accessorKey: "department.name",
      header: "Department",
      cell: ({ row }) => <span>{row.original.department?.name || "-"}</span>,
    },
    {
      accessorKey: "designation_ref.name",
      header: "Designation",
      cell: ({ row }) => <span>{row.original.designation_ref?.name || "-"}</span>,
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const s = row.original.status
        return <Badge variant={s === "active" ? "default" : "destructive"}>{s}</Badge>
      },
    },
  ], [])

  const filterDefs: FilterDef[] = [
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en })) },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    { key: "section_id", label: "Section", type: "select", options: sections.map((s) => ({ value: s.id, label: s.name })) },
    { key: "designation_id", label: "Designation", type: "select", options: designations.map((d) => ({ value: d.id, label: d.name })) },
    { key: "line_id", label: "Line", type: "select", options: lines.map((l) => ({ value: l.id, label: l.name })) },
    { key: "group_id", label: "Group", type: "select", options: groups.map((g) => ({ value: g.id, label: g.name })) },
    { key: "status", label: "Emp. Status", type: "select", options: [
      { value: "active", label: "Active" }, { value: "inactive", label: "Inactive" },
    ] },
    { key: "employee_id", label: "Employee Code", type: "text", placeholder: "Enter employee code..." },
  ]

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <ClipboardCheckIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Manual Attendance</h1>
            <p className="text-muted-foreground mt-1">Check employees from the table and submit bulk attendance</p>
          </div>
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
                  values={filters}
                  onChange={handleChange}
                  onApply={() => { handleApply(); setMobileFilterOpen(false) }}
                  onReset={() => { handleReset(); setMobileFilterOpen(false) }}
                  submitting={loading}
                  singleColumn
                  noBorder
                />
              </div>
            </SheetContent>
          </Sheet>
        </ButtonGroup>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleChange}
          onApply={handleApply}
          onReset={handleReset}
          submitting={loading}
        />
      </div>

      <div className="px-4 lg:px-6">
        <div className="rounded-lg border bg-card">
          <div className="flex items-center justify-between px-4 py-3 border-b bg-muted/30">
            <div className="flex items-center gap-2">
              <UsersIcon className="h-5 w-5 text-muted-foreground" />
              <span className="text-sm font-medium">Bulk Attendance Entry</span>
            </div>
            {selectedRows.length > 0 && (
              <Badge variant="default" className="text-xs">
                {selectedRows.length} employee(s) selected
              </Badge>
            )}
          </div>
          <div className="p-4">
            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5 max-w-[240px]">
                <label className="text-xs font-medium text-muted-foreground">Date</label>
                <DatePicker value={entryDate} onChange={(d) => d && setEntryDate(d)} />
              </div>
              <div className="flex flex-row flex-wrap gap-4">
                <div className="flex flex-col gap-1.5 min-w-[200px] flex-1">
                  <label className="text-xs font-medium text-muted-foreground">In Time</label>
                  <DateTimePicker value={checkIn} onChange={setCheckIn} />
                </div>
                <div className="flex flex-col gap-1.5 min-w-[200px] flex-1">
                  <label className="text-xs font-medium text-muted-foreground">Out Time</label>
                  <DateTimePicker value={checkOut} onChange={setCheckOut} />
                </div>
              </div>
              <div className="flex flex-row flex-wrap gap-4">
                <div className="flex flex-col gap-1.5 min-w-[140px]">
                  <label className="text-xs font-medium text-muted-foreground">Status</label>
                  <select
                    value={entryStatus}
                    onChange={(e) => setEntryStatus(e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  >
                    <option value="present">Present</option>
                    <option value="late">Late</option>
                    <option value="absent">Absent</option>
                    <option value="half_day">Half Day</option>
                    <option value="on_leave">On Leave</option>
                    <option value="weekend">Weekend</option>
                  </select>
                </div>
                <Button onClick={handleSubmit} disabled={submitting || !selectedRows.length} className="h-10 self-end">
                  {submitting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                  Submit {selectedRows.length > 0 ? `Selected (${selectedRows.length})` : ""}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <DataTable
          data={data}
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
          onSelectionChange={setSelectedRows}
        />
      </div>
    </div>
  )
}
