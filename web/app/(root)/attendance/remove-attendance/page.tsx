"use client"

import * as React from "react"
import { Trash2Icon, Loader2, AlertTriangleIcon, FilterIcon, XIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { format } from "date-fns"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { attendanceApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"
import { formatCheck } from "@/lib/utils"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { toast } from "sonner"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

interface AttendanceRecord {
  id: string
  employee_id: string
  company_id: string
  date: string
  check_in: string | null
  check_out: string | null
  total_hours: string | null
  status: string
  employee_name?: string
  designation?: string
}

const today = new Date().toISOString().split("T")[0]

export default function RemoveAttendancePage() {
  const [data, setData] = React.useState<AttendanceRecord[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])
  const [selectedRows, setSelectedRows] = React.useState<AttendanceRecord[]>([])
  const [deleting, setDeleting] = React.useState(false)
  const [confirmOpen, setConfirmOpen] = React.useState(false)
  const [deleteType, setDeleteType] = React.useState<"selected" | "all">("selected")
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  const [filters, setFilters] = React.useState<Record<string, string>>({
    start_date: today,
    end_date: today,
  })

  const fetchData = React.useCallback(async (f?: Record<string, string>, p?: number, l?: number) => {
    setLoading(true)
    setError("")
    try {
      const params: Record<string, string> = {
        start_date: f?.start_date || filters.start_date || today,
        end_date: f?.end_date || filters.end_date || today,
        page: String(p ?? page),
        limit: String(l ?? limit),
      }
      if (f?.company_id || filters.company_id) params.company_id = f?.company_id || filters.company_id
      if (f?.department_id || filters.department_id) params.department_id = f?.department_id || filters.department_id
      if (f?.section_id || filters.section_id) params.section_id = f?.section_id || filters.section_id
      if (f?.designation_id || filters.designation_id) params.designation_id = f?.designation_id || filters.designation_id
      if (f?.line_id || filters.line_id) params.line_id = f?.line_id || filters.line_id
      if (f?.group_id || filters.group_id) params.group_id = f?.group_id || filters.group_id
      if (f?.shift_id || filters.shift_id) params.shift_id = f?.shift_id || filters.shift_id
      if (f?.status || filters.status) params.status = f?.status || filters.status
      const { data: res } = await attendanceApi.list(params)
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setError("Failed to load attendance records")
    } finally {
      setLoading(false)
    }
  }, [page, limit, filters])

  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      sectionApi.list(undefined, { limit: "100" }),
      designationApi.list(undefined, { limit: "100" }),
      lineApi.list(undefined, { limit: "100" }),
      groupApi.list({ limit: "100" }),
      shiftApi.list({ limit: "100" }),
    ]).then(([cRes, dRes, secRes, desRes, lRes, gRes, sRes]) => {
      if (Array.isArray(cRes.data?.data)) setCompanies(cRes.data.data)
      if (Array.isArray(dRes.data?.data)) setDepartments(dRes.data.data)
      if (Array.isArray(secRes.data?.data)) setSections(secRes.data.data)
      if (Array.isArray(desRes.data?.data)) setDesignations(desRes.data.data)
      if (Array.isArray(lRes.data?.data)) setLines(lRes.data.data)
      if (Array.isArray(gRes.data?.data)) setGroups(gRes.data.data)
      if (Array.isArray(sRes.data?.data)) setShifts(sRes.data.data)
    }).catch(() => {})
    fetchData()
  }, [])

  React.useEffect(() => {
    fetchData()
  }, [page, limit])

  const handleChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleApply = () => {
    setPage(1)
    fetchData(filters, 1)
  }

  const handleReset = () => {
    const defaults: Record<string, string> = { start_date: today, end_date: today }
    setFilters(defaults)
    setPage(1)
    fetchData(defaults, 1)
  }

  const handleDelete = async () => {
    if (deleteType === "all") {
      setDeleting(true)
      try {
        await attendanceApi.deleteAll()
        toast.success("All attendance records deleted")
        setSelectedRows([])
        fetchData()
      } catch {
        toast.error("Failed to delete all records")
      } finally {
        setDeleting(false)
        setConfirmOpen(false)
      }
      return
    }

    if (!selectedRows.length) return
    setDeleting(true)
    try {
      const ids = selectedRows.map((r) => r.id)
      await attendanceApi.deleteBulk(ids)
      toast.success(`Deleted ${ids.length} attendance record(s)`)
      setSelectedRows([])
      fetchData()
    } catch {
      toast.error("Failed to delete selected records")
    } finally {
      setDeleting(false)
      setConfirmOpen(false)
    }
  }

  const columns: ColumnDef<AttendanceRecord>[] = React.useMemo(() => [
    { id: "sl", header: "Sl", cell: ({ row }) => (page - 1) * limit + row.index + 1 },
    {
      accessorKey: "employee_id",
      header: "Employee ID",
      cell: ({ row }) => row.original.employee_id,
    },
    {
      accessorKey: "employee_name",
      header: "Name",
      cell: ({ row }) => row.original.employee_name || "-",
    },
    {
      accessorKey: "designation",
      header: "Designation",
      cell: ({ row }) => row.original.designation || "-",
    },
    { accessorKey: "date", header: "Date", cell: ({ row }) => row.original.date?.split("-").reverse().join("-") },
    { accessorKey: "check_in", header: "In", cell: ({ row }) => formatCheck(row.original.check_in) },
    { accessorKey: "check_out", header: "Out", cell: ({ row }) => formatCheck(row.original.check_out) },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const s = row.original.status
        const v = s === "present" ? "default" : s === "absent" ? "destructive" : "secondary"
        return <Badge variant={v} className="capitalize">{s}</Badge>
      },
    },
  ], [page, limit])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date_range", label: "Date Range", type: "daterange-split", dateRangeKeys: { start: "start_date", end: "end_date" } },
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en })) },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    { key: "section_id", label: "Section", type: "select", options: sections.map((s) => ({ value: s.id, label: s.name })) },
    { key: "designation_id", label: "Designation", type: "select", options: designations.map((d) => ({ value: d.id, label: d.name })) },
    { key: "line_id", label: "Line", type: "select", options: lines.map((l) => ({ value: l.id, label: l.name })) },
    { key: "group_id", label: "Group", type: "select", options: groups.map((g) => ({ value: g.id, label: g.name })) },
    { key: "shift_id", label: "Shift", type: "select", options: shifts.map((s) => ({ value: s.id, label: s.name })) },
    {
      key: "status", label: "Status", type: "select", options: [
        { value: "present", label: "Present" },
        { value: "late", label: "Late" },
        { value: "absent", label: "Absent" },
        { value: "half_day", label: "Half Day" },
        { value: "on_leave", label: "On Leave" },
        { value: "weekend", label: "Weekend" },
      ],
    },
  ], [companies, departments, sections, designations, lines, groups, shifts])

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Trash2Icon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Remove Attendance</h1>
            <p className="text-muted-foreground mt-1">Select and delete attendance records</p>
          </div>
        </div>
        <div className="hidden md:flex items-center gap-2">
          {selectedRows.length > 0 && (
            <Button
              variant="destructive"
              size="sm"
              onClick={() => { setDeleteType("selected"); setConfirmOpen(true) }}
              disabled={deleting}
            >
              {deleting ? <Loader2 className="mr-1 h-4 w-4 animate-spin" /> : <Trash2Icon className="mr-1 h-4 w-4" />}
              Delete Selected ({selectedRows.length})
            </Button>
          )}
          <Button
            variant="outline"
            size="sm"
            className="text-destructive hover:bg-destructive/10 border-destructive/30"
            onClick={() => { setDeleteType("all"); setConfirmOpen(true) }}
            disabled={deleting || total === 0}
          >
            <Trash2Icon className="mr-1 h-4 w-4" />
            Delete All
          </Button>
        </div>
      </div>

      <div className="md:hidden px-4 lg:px-6 mt-3">
        {selectedRows.length > 0 && (
          <ButtonGroup className="w-full mb-3">
            <Button
              variant="destructive"
              onClick={() => { setDeleteType("selected"); setConfirmOpen(true) }}
              disabled={deleting}
              className="flex-1"
            >
              {deleting ? <Loader2 className="mr-1 h-4 w-4 animate-spin" /> : <Trash2Icon className="mr-1 h-4 w-4" />}
              Delete Selected ({selectedRows.length})
            </Button>
          </ButtonGroup>
        )}
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

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

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

      <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <AlertTriangleIcon className="h-5 w-5 text-destructive" />
              Confirm Delete
            </AlertDialogTitle>
            <AlertDialogDescription>
              {deleteType === "all"
                ? "This will permanently delete ALL attendance records. This action cannot be undone."
                : `Are you sure you want to delete ${selectedRows.length} selected attendance record(s)? This action cannot be undone.`
              }
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
              {deleting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
