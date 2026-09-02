"use client"

import * as React from "react"
import { format } from "date-fns"
import { UsersIcon, PlusIcon, UploadIcon, DownloadIcon, Loader2, FilterIcon, XIcon, MoreHorizontalIcon } from "lucide-react"
import { DatePicker } from "@/components/ui/date-picker"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { useRouter, useSearchParams } from "next/navigation"
import Link from "next/link"
import { ButtonGroup } from "@/components/ui/button-group"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetFooter, SheetClose } from "@/components/ui/sheet"
import { statusOptionsEmployee, genderOptions, bloodGroupOptions } from "@/components/data/employee-data"
import {
  employeeApi,
  companyApi,
  departmentApi,
  sectionApi,
  designationApi,
  lineApi,
  shiftApi,
  groupApi,
  floorApi,
} from "@/lib/api"

interface Company {
  id: string
  company_name_en: string
}
interface Department {
  id: string
  name: string
}
interface Section {
  id: string
  name: string
}
interface Designation {
  id: string
  name: string
}
interface Line {
  id: string
  name: string
}
interface Shift {
  id: string
  name: string
}
interface Group {
  id: string
  name: string
}
interface Floor {
  id: string
  name: string
}

interface EmployeeRow {
  id: string
  employee_id: string
  punch_number: string
  name_en: string
  name_bn: string
  phone: string
  designation: string
  department: string
  section: string
  line: string
  group: string
  floor: string
  joining_date: string
  gross_salary: number
  status: string
  employee_type: string
  gender: string
  company_id: string
  shift_id: string | null
  department_id: string | null
  section_id: string | null
  designation_id: string | null
  line_id: string | null
  group_id: string | null
  floor_id: string | null
}

const columns: ColumnDef<EmployeeRow>[] = [
  { accessorKey: "employee_id", header: "Emp. ID" },
  {
    accessorKey: "name_en",
    header: "Name",
    cell: ({ row }) => (
      <Link
        href={`/hr/employees/${row.original.id}`}
        className="cursor-pointer text-inherit no-underline hover:text-green-600 hover:underline"
      >
        {row.original.name_en}
      </Link>
    ),
  },
  { accessorKey: "designation", header: "Designation", cell: ({ row }) => {
    const v = row.original.designation as unknown
    return (typeof v === "object" && v !== null ? (v as { name?: string }).name : v as string) || "-"
  }},
  { accessorKey: "department", header: "Department", cell: ({ row }) => {
    const v = row.original.department as unknown
    return (typeof v === "object" && v !== null ? (v as { name?: string }).name : v as string) || "-"
  }},
  { accessorKey: "section", header: "Section", cell: ({ row }) => {
    const v = row.original.section as unknown
    return (typeof v === "object" && v !== null ? (v as { name?: string }).name : v as string) || "-"
  }},
  { accessorKey: "punch_number", header: "Punch No" },
  { accessorKey: "phone", header: "Phone" },
  {
    accessorKey: "joining_date",
    header: "Joining Date",
    cell: ({ row }) => row.original.joining_date || "-",
  },
  {
    accessorKey: "gross_salary",
    header: "Salary",
    cell: ({ row }) => row.original.gross_salary?.toLocaleString() || "-",
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => (
      <Badge variant={row.original.status === "active" ? "default" : "secondary"} className="capitalize">
        {statusOptionsEmployee.find((s) => s.value === row.original.status)?.label || row.original.status}
      </Badge>
    ),
  },
]

const selectClass =
  "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"

export default function EmployeesPage() {
  const router = useRouter()
  const [data, setData] = React.useState<EmployeeRow[]>([])
  const [loading, setLoading] = React.useState(true)
  const [submitting, setSubmitting] = React.useState(false)
  const [exporting, setExporting] = React.useState(false)
  const [error, setError] = React.useState("")

  const [filters, setFilters] = React.useState<Record<string, string>>({ employee_type: "Regular", status: "active" })
  const [appliedFilters, setAppliedFilters] = React.useState<Record<string, string>>({ employee_type: "Regular", status: "active" })
  const [joiningFromDate, setJoiningFromDate] = React.useState<Date | undefined>(undefined)
  const [joiningToDate, setJoiningToDate] = React.useState<Date | undefined>(undefined)
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [floors, setFloors] = React.useState<Floor[]>([])

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)
  const isFirstMountRef = React.useRef(true)

  const fetchEmployees = async (f?: Record<string, string>, p?: number, l?: number) => {
    setError("")
    setLoading(true)
    try {
      const params: Record<string, string> = { ...(f || {}), page: String(p ?? page), limit: String(l ?? limit) }
      console.log("[Employees] fetch with params", params)
      const { data: res } = await employeeApi.list(params)
      const rows = Array.isArray(res.data) ? res.data : []
      setData(rows)
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
      // Fallback: if employee_id was searched and list is empty, try direct lookup (exact code / punch)
      if (rows.length === 0 && params["employee_id"]) {
        try {
          const { data: emp } = await employeeApi.getByCode(params["employee_id"])
          if (emp && (emp as unknown as { employee_id?: string }).employee_id) {
            const single = emp as unknown as EmployeeRow
            // map single employee to row shape if needed
            setData([single as unknown as EmployeeRow])
            setTotal(1)
            setTotalPages(1)
            setError("")
          }
        } catch {
          // keep empty — will show empty state with hint
        }
      }
    } catch {
      setError("Failed to load employees")
    } finally {
      setLoading(false)
    }
  }

  // Restore from sessionStorage + URL on mount
  React.useEffect(() => {
    try {
      const raw = sessionStorage.getItem("employees_advanced_filters_v2")
      if (raw) {
        const saved = JSON.parse(raw)
        if (saved.filters && typeof saved.filters === "object") setFilters(saved.filters)
        if (saved.appliedFilters && typeof saved.appliedFilters === "object") setAppliedFilters(saved.appliedFilters)
        if (saved.joiningFromDate) {
          const d = new Date(saved.joiningFromDate)
          if (!isNaN(d.getTime())) setJoiningFromDate(d)
        }
        if (saved.joiningToDate) {
          const d = new Date(saved.joiningToDate)
          if (!isNaN(d.getTime())) setJoiningToDate(d)
        }
        if (typeof saved.page === "number" && saved.page >= 1) setPage(saved.page)
        if (typeof saved.limit === "number" && saved.limit >= 1) setLimit(saved.limit)
        if (saved.filters?.department_id) {
          sectionApi.list(saved.filters.department_id, { limit: "100" }).then(({ data: secData }) => {
            setSections(Array.isArray(secData.data) ? secData.data : [])
          }).catch(() => {})
        }
        if (saved.filters?.section_id) {
          Promise.all([
            designationApi.list(saved.filters.section_id, { limit: "100" }),
            lineApi.list(saved.filters.section_id, { limit: "100" }),
          ]).then(([desigRes, lineRes]) => {
            setDesignations(Array.isArray(desigRes.data?.data) ? desigRes.data.data : [])
            setLines(Array.isArray(lineRes.data?.data) ? lineRes.data.data : [])
          }).catch(() => {})
        }
      }
    } catch {
      // ignore corrupt storage
    }
  }, [])

  // Persist to sessionStorage on change (after first mount)
  React.useEffect(() => {
    if (isFirstMountRef.current) {
      isFirstMountRef.current = false
      return
    }
    try {
      const payload = {
        filters,
        appliedFilters,
        joiningFromDate: joiningFromDate ? joiningFromDate.toISOString() : null,
        joiningToDate: joiningToDate ? joiningToDate.toISOString() : null,
        page,
        limit,
      }
      sessionStorage.setItem("employees_advanced_filters_v2", JSON.stringify(payload))
    } catch {}
  }, [filters, appliedFilters, joiningFromDate, joiningToDate, page, limit])

  React.useEffect(() => {
    const init = async () => {
      try {
        const [compRes, deptRes, shiftRes, groupRes, floorRes] = await Promise.all([
          companyApi.list({ limit: "100" }),
          departmentApi.list({ limit: "100" }),
          shiftApi.list({ limit: "100" }),
          groupApi.list({ limit: "100" }),
          floorApi.list({ limit: "100" }),
        ])
        setCompanies(Array.isArray(compRes.data?.data) ? compRes.data.data : [])
        setDepartments(Array.isArray(deptRes.data?.data) ? deptRes.data.data : [])
        setShifts(Array.isArray(shiftRes.data?.data) ? shiftRes.data.data : [])
        setGroups(Array.isArray(groupRes.data?.data) ? groupRes.data.data : [])
        setFloors(Array.isArray(floorRes.data?.data) ? floorRes.data.data : [])
      } catch {
        // dropdowns will be empty
      }
    }
    init()
  }, [])

  // Always fetch with appliedFilters + pagination (ensures filter persists)
  React.useEffect(() => {
    fetchEmployees(appliedFilters, page, limit)
  }, [appliedFilters, page, limit])

  const handleDepartmentChange = async (value: string) => {
    setFilters((prev) => ({
      ...prev,
      department_id: value,
      section_id: "",
      designation_id: "",
      line_id: "",
    }))
    if (value) {
      try {
        const { data: secData } = await sectionApi.list(value, { limit: "100" })
        setSections(Array.isArray(secData.data) ? secData.data : [])
      } catch {
        setSections([])
      }
    } else {
      setSections([])
    }
    setDesignations([])
    setLines([])
  }

  const handleSectionChange = async (value: string) => {
    setFilters((prev) => ({
      ...prev,
      section_id: value,
      designation_id: "",
      line_id: "",
    }))
    if (value) {
      try {
        const [desigRes, lineRes] = await Promise.all([
          designationApi.list(value, { limit: "100" }),
          lineApi.list(value, { limit: "100" }),
        ])
        setDesignations(Array.isArray(desigRes.data?.data) ? desigRes.data.data : [])
        setLines(Array.isArray(lineRes.data?.data) ? lineRes.data.data : [])
      } catch {
        setDesignations([])
        setLines([])
      }
    } else {
      setDesignations([])
      setLines([])
    }
  }

  const handleApply = async () => {
    setSubmitting(true)
    setError("")
    const active = { ...filters }
    if (joiningFromDate) active.joining_from = format(joiningFromDate, "yyyy-MM-dd")
    else delete active.joining_from
    if (joiningToDate) active.joining_to = format(joiningToDate, "yyyy-MM-dd")
    else delete active.joining_to
    // If searching by exact employee_id/punch, ignore narrow HR filters that may hide the result
    // (keep status/employee_type as they are usually intended, but clear org hierarchy if it blocks global search)
    // We keep company_id/department etc. but user can Reset if needed — log for debug
    const cleaned = Object.fromEntries(Object.entries(active).filter(([, v]) => v !== "")) as Record<string, string>
    if (cleaned.employee_id) {
      console.log("[Employees] employee_id search — active filters", cleaned)
    }
    setAppliedFilters(cleaned)
    if (page !== 1) {
      setPage(1)
      setSubmitting(false)
    } else {
      await fetchEmployees(cleaned, 1, limit)
      setSubmitting(false)
    }
  }

  const handleReset = async () => {
    const resetFilters = { status: "active", employee_type: "Regular" } as Record<string, string>
    setFilters(resetFilters)
    setAppliedFilters(resetFilters)
    setJoiningFromDate(undefined)
    setJoiningToDate(undefined)
    setSections([])
    setDesignations([])
    setLines([])
    setError("")
    setSubmitting(true)
    if (page !== 1 || limit !== 20) {
      setPage(1)
      setLimit(20)
      setSubmitting(false)
    } else {
      await fetchEmployees(resetFilters, 1, 20)
      setSubmitting(false)
    }
  }

  const handleEdit = (emp: EmployeeRow) => router.push(`/hr/employees/${emp.id}/edit`)
  const handleDelete = async (emp: EmployeeRow) => {
    try {
      await employeeApi.delete(emp.id)
      setData((prev) => prev.filter((e) => e.id !== emp.id))
      setTotal((prev) => Math.max(0, prev - 1))
    } catch {
      setError("Failed to delete employee")
    }
  }

  const handleExport = async () => {
    setExporting(true)
    try {
      const cleaned = Object.fromEntries(Object.entries(appliedFilters).filter(([, v]) => v !== ""))
      const res = await employeeApi.exportExcel(cleaned)
      const blob = new Blob([res.data], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `ManPower list.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setError("Failed to export employees")
    } finally {
      setExporting(false)
    }
  }

  const setFilter = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const renderFilterFields = () => (
    <>
      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Company</label>
        <select
          value={filters.company_id || ""}
          onChange={(e) => setFilter("company_id", e.target.value)}
          className={selectClass}
        >
          <option value="">— All —</option>
          {companies.map((c) => (
            <option key={c.id} value={c.id}>
              {c.company_name_en}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Department</label>
        <select
          value={filters.department_id || ""}
          onChange={(e) => handleDepartmentChange(e.target.value)}
          className={selectClass}
        >
          <option value="">— All —</option>
          {departments.map((d) => (
            <option key={d.id} value={d.id}>
              {d.name}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Section</label>
        <select
          value={filters.section_id || ""}
          onChange={(e) => handleSectionChange(e.target.value)}
          className={selectClass}
          disabled={!filters.department_id}
        >
          <option value="">— All —</option>
          {sections.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Designation</label>
        <select
          value={filters.designation_id || ""}
          onChange={(e) => setFilter("designation_id", e.target.value)}
          className={selectClass}
          disabled={!filters.section_id}
        >
          <option value="">— All —</option>
          {designations.map((d) => (
            <option key={d.id} value={d.id}>
              {d.name}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Line</label>
        <select
          value={filters.line_id || ""}
          onChange={(e) => setFilter("line_id", e.target.value)}
          className={selectClass}
          disabled={!filters.section_id}
        >
          <option value="">— All —</option>
          {lines.map((l) => (
            <option key={l.id} value={l.id}>
              {l.name}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Shift</label>
        <select
          value={filters.shift_id || ""}
          onChange={(e) => setFilter("shift_id", e.target.value)}
          className={selectClass}
        >
          <option value="">— All —</option>
          {shifts.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Group</label>
        <select
          value={filters.group_id || ""}
          onChange={(e) => setFilter("group_id", e.target.value)}
          className={selectClass}
        >
          <option value="">— All —</option>
          {groups.map((g) => (
            <option key={g.id} value={g.id}>
              {g.name}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Floor</label>
        <select
          value={filters.floor_id || ""}
          onChange={(e) => setFilter("floor_id", e.target.value)}
          className={selectClass}
        >
          <option value="">— All —</option>
          {floors.map((f) => (
            <option key={f.id} value={f.id}>
              {f.name}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Status</label>
        <select
          value={filters.status || ""}
          onChange={(e) => setFilter("status", e.target.value)}
          className={selectClass}
        >
          <option value="">— All —</option>
          {statusOptionsEmployee.map((s) => (
            <option key={s.value} value={s.value}>
              {s.label}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Employee ID</label>
        <input
          type="text"
          value={filters.employee_id || ""}
          onChange={(e) => setFilter("employee_id", e.target.value)}
          placeholder="Search by code..."
          className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Gender</label>
        <select
          value={filters.gender || ""}
          onChange={(e) => setFilter("gender", e.target.value)}
          className={selectClass}
        >
          <option value="">— All —</option>
          {genderOptions.map((g) => (
            <option key={g.value} value={g.value}>
              {g.label}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Employee Type</label>
        <select
          value={filters.employee_type || ""}
          onChange={(e) => setFilter("employee_type", e.target.value)}
          className={selectClass}
        >
          <option value="">— All —</option>
          <option value="Regular">Regular</option>
          <option value="Lefty">Lefty</option>
          <option value="Resign">Resign</option>
          <option value="Close">Close</option>
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Blood Group</label>
        <select
          value={filters.blood_group || ""}
          onChange={(e) => setFilter("blood_group", e.target.value)}
          className={selectClass}
        >
          <option value="">— All —</option>
          {bloodGroupOptions.map((b) => (
            <option key={b.value} value={b.value}>
              {b.label}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Min Salary</label>
        <input
          type="number"
          value={filters.min_salary || ""}
          onChange={(e) => setFilter("min_salary", e.target.value)}
          placeholder="0"
          min="0"
          className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Max Salary</label>
        <input
          type="number"
          value={filters.max_salary || ""}
          onChange={(e) => setFilter("max_salary", e.target.value)}
          placeholder="999999"
          min="0"
          className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Joining Date From</label>
        <DatePicker value={joiningFromDate} onChange={setJoiningFromDate} placeholder="From" />
      </div>

      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted-foreground">Joining Date To</label>
        <DatePicker value={joiningToDate} onChange={setJoiningToDate} placeholder="To" />
      </div>
    </>
  )

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-2">
          <div className="flex items-center gap-2">
            <UsersIcon className="h-6 w-6 text-muted-foreground" />
            <div>
              <h1 className="text-lg md:text-3xl font-bold tracking-tight">Employees</h1>
              <p className="text-muted-foreground mt-1">Manage employee records</p>
            </div>
          </div>
          <ButtonGroup className="hidden md:flex">
            <Button variant="outline" onClick={handleExport} disabled={exporting}>
              {exporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <DownloadIcon className="mr-2 h-4 w-4" />}
              Export
            </Button>
            <Button variant="outline" onClick={() => router.push("/hr/employees/import")}>
              <UploadIcon className="mr-2 h-4 w-4" />
              Import
            </Button>
            <Button onClick={() => router.push("/hr/employees/create")}>
              <PlusIcon className="mr-2 h-4 w-4" />
              Add Employee
            </Button>
          </ButtonGroup>
        </div>
        <div className="md:hidden flex items-center justify-end gap-2 mt-3">
          <Sheet>
            <SheetTrigger asChild>
              <Button variant="outline">
                <FilterIcon className="h-4 w-4" />
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
                <div className="flex flex-col gap-4">{renderFilterFields()}</div>
              </div>
              <SheetFooter className="px-4 py-3 border-t">
                <div className="flex items-center gap-2 w-full">
                  <SheetClose asChild>
                    <Button onClick={handleApply} disabled={submitting} className="flex-1">
                      {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                      Apply
                    </Button>
                  </SheetClose>
                  <Button variant="outline" onClick={handleReset} disabled={submitting} className="flex-1">
                    Reset
                  </Button>
                </div>
              </SheetFooter>
            </SheetContent>
          </Sheet>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline">
                <MoreHorizontalIcon className="h-4 w-4" />
                More
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-44">
              <DropdownMenuItem onClick={handleExport} disabled={exporting}>
                <DownloadIcon className="mr-2 h-4 w-4" />
                {exporting ? "Exporting..." : "Export"}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => router.push("/hr/employees/import")}>
                <UploadIcon className="mr-2 h-4 w-4" />
                Import
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => router.push("/hr/employees/create")}>
                <PlusIcon className="mr-2 h-4 w-4" />
                Add Employee
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      {!loading && data.length === 0 && appliedFilters.employee_id && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-900 px-4 py-3 text-sm">
            <p className="font-medium text-amber-800 dark:text-amber-300">No results for Employee ID “{appliedFilters.employee_id}” with current filters.</p>
            <p className="text-amber-700 dark:text-amber-400 mt-1">Active filters: {Object.entries(appliedFilters).map(([k,v])=>`${k}=${v}`).join(", ") || "none"} — try <button onClick={handleReset} className="underline font-medium">Reset</button> to clear company/department filters or check the ID. Direct API <code className="bg-white/60 dark:bg-black/20 px-1 rounded">/api/v1/employees?employee_id={appliedFilters.employee_id}</code> does return data, so the web filter is over-narrow.</p>
          </div>
        </div>
      )}



      {/* Desktop: inline filter card */}
      <div className="px-4 lg:px-6 hidden md:block">
        <div className="rounded-lg border bg-card p-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">{renderFilterFields()}</div>
          <div className="flex items-center gap-2 mt-4">
            <Button onClick={handleApply} disabled={submitting}>
              {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Apply
            </Button>
            <Button variant="outline" onClick={handleReset} disabled={submitting}>
              Reset
            </Button>
          </div>
        </div>
      </div>

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
      />
    </div>
  )
}
