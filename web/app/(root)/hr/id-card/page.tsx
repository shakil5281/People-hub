"use client"

import * as React from "react"
import { IdCardIcon, DownloadIcon, Loader2, FilterIcon, XIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
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
  idCardApi,
} from "@/lib/api"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Shift { id: string; name: string }
interface Group { id: string; name: string }
interface Floor { id: string; name: string }

interface EmployeeRow {
  id: string
  employee_id: string
  punch_number: string
  name_en: string
  phone: string
  blood_group: string
  designation: string
  department: string
  status: string
}

const columns: ColumnDef<EmployeeRow>[] = [
  { accessorKey: "employee_id", header: "Emp. ID" },
  { accessorKey: "name_en", header: "Name" },
  { accessorKey: "designation", header: "Designation", cell: ({ row }) => row.original.designation || "-" },
  { accessorKey: "department", header: "Department", cell: ({ row }) => row.original.department || "-" },
  { accessorKey: "punch_number", header: "Punch No" },
  { accessorKey: "phone", header: "Phone" },
  { accessorKey: "blood_group", header: "Blood Group" },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const s = row.original.status
      return <span className={s === "active" ? "text-green-600 font-medium" : "text-muted-foreground"}>{s}</span>
    },
  },
]

const selectClass = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"

export default function IdCardPage() {
  const [data, setData] = React.useState<EmployeeRow[]>([])
  const [loading, setLoading] = React.useState(true)
  const [generating, setGenerating] = React.useState(false)
  const [selectedRows, setSelectedRows] = React.useState<EmployeeRow[]>([])
  const [error, setError] = React.useState("")
  const [filters, setFilters] = React.useState<Record<string, string>>({ employee_type: "Regular", status: "active" })
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

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

  const fetchEmployees = async (f?: Record<string, string>, p?: number, l?: number) => {
    setError("")
    setLoading(true)
    try {
      const params = { ...(f || {}), page: String(p ?? page), limit: String(l ?? limit) }
      const { data: res } = await employeeApi.list(params)
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setError("Failed to load employees")
    } finally {
      setLoading(false)
    }
  }

  React.useEffect(() => {
    Promise.allSettled([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      shiftApi.list({ limit: "100" }),
      groupApi.list({ limit: "100" }),
      floorApi.list({ limit: "100" }),
    ]).then((results) => {
      const getData = (r: PromiseSettledResult<any>, path: string = "data") => {
        if (r.status === "fulfilled") {
          const d = path ? r.value.data?.data || r.value.data : r.value.data
          return Array.isArray(d) ? d : []
        }
        return []
      }
      setCompanies(getData(results[0]))
      setDepartments(getData(results[1]))
      setShifts(getData(results[2]))
      setGroups(getData(results[3]))
      setFloors(getData(results[4]))
    })
    fetchEmployees(filters)
  }, [])

  React.useEffect(() => {
    fetchEmployees(filters)
  }, [page, limit])

  const handleDepartmentChange = async (value: string) => {
    setFilter("department_id", value)
    setFilter("section_id", "")
    setFilter("designation_id", "")
    setFilter("line_id", "")
    if (value) {
      try {
        const [secRes] = await Promise.all([
          sectionApi.list(value, { limit: "100" }),
        ])
        setSections(Array.isArray(secRes.data?.data) ? secRes.data.data : [])
      } catch { setSections([]) }
    } else { setSections([]); setDesignations([]); setLines([]) }
  }

  const handleSectionChange = async (value: string) => {
    setFilter("section_id", value)
    setFilter("designation_id", "")
    setFilter("line_id", "")
    if (value) {
      try {
        const [desigRes, lineRes] = await Promise.all([
          designationApi.list(value, { limit: "100" }),
          lineApi.list(value, { limit: "100" }),
        ])
        setDesignations(Array.isArray(desigRes.data?.data) ? desigRes.data.data : [])
        setLines(Array.isArray(lineRes.data?.data) ? lineRes.data.data : [])
      } catch { setDesignations([]); setLines([]) }
    } else { setDesignations([]); setLines([]) }
  }

  const handleApply = async () => {
    setPage(1)
    const active = Object.fromEntries(Object.entries(filters).filter(([, v]) => v !== ""))
    await fetchEmployees(active, 1)
  }

  const handleReset = async () => {
    setPage(1)
    setLimit(20)
    setFilters({ employee_type: "Regular", status: "active" })
    setSections([])
    setDesignations([])
    setLines([])
    await fetchEmployees({ status: "active", employee_type: "Regular" }, 1, 20)
  }

  const setFilter = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleGenerate = async () => {
    if (selectedRows.length === 0) return
    setGenerating(true)
    try {
      const employeeIds = selectedRows.map((r) => r.employee_id)
      const res = await idCardApi.generate(employeeIds)
      const { data: base64Data, filename } = res.data
      const binaryStr = atob(base64Data)
      const bytes = new Uint8Array(binaryStr.length)
      for (let i = 0; i < binaryStr.length; i++) {
        bytes[i] = binaryStr.charCodeAt(i)
      }
      const blob = new Blob([bytes], { type: "application/pdf" })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = filename || `id_cards_${new Date().toISOString().slice(0, 10)}.pdf`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setError("Failed to generate ID cards")
    } finally {
      setGenerating(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <IdCardIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">ID Card</h1>
            <p className="text-muted-foreground mt-1">Generate employee ID cards</p>
          </div>
        </div>
        <div className="hidden md:block">
          <Button onClick={handleGenerate} disabled={generating || selectedRows.length === 0}>
            {generating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <DownloadIcon className="mr-2 h-4 w-4" />}
            {generating ? "Generating..." : `Download Selected (${selectedRows.length})`}
          </Button>
        </div>
      </div>

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

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
                <div className="grid grid-cols-1 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Company</label>
                    <select value={filters.company_id || ""} onChange={(e) => setFilter("company_id", e.target.value)} className={selectClass}>
                      <option value="">— All —</option>
                      {companies.map((c) => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Department</label>
                    <select value={filters.department_id || ""} onChange={(e) => handleDepartmentChange(e.target.value)} className={selectClass}>
                      <option value="">— All —</option>
                      {departments.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Section</label>
                    <select value={filters.section_id || ""} onChange={(e) => handleSectionChange(e.target.value)} className={selectClass} disabled={!filters.department_id}>
                      <option value="">— All —</option>
                      {sections.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Designation</label>
                    <select value={filters.designation_id || ""} onChange={(e) => setFilter("designation_id", e.target.value)} className={selectClass} disabled={!filters.section_id}>
                      <option value="">— All —</option>
                      {designations.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Line</label>
                    <select value={filters.line_id || ""} onChange={(e) => setFilter("line_id", e.target.value)} className={selectClass} disabled={!filters.section_id}>
                      <option value="">— All —</option>
                      {lines.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Shift</label>
                    <select value={filters.shift_id || ""} onChange={(e) => setFilter("shift_id", e.target.value)} className={selectClass}>
                      <option value="">— All —</option>
                      {shifts.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Group</label>
                    <select value={filters.group_id || ""} onChange={(e) => setFilter("group_id", e.target.value)} className={selectClass}>
                      <option value="">— All —</option>
                      {groups.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Floor</label>
                    <select value={filters.floor_id || ""} onChange={(e) => setFilter("floor_id", e.target.value)} className={selectClass}>
                      <option value="">— All —</option>
                      {floors.map((f) => <option key={f.id} value={f.id}>{f.name}</option>)}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Status</label>
                    <select value={filters.status || ""} onChange={(e) => setFilter("status", e.target.value)} className={selectClass}>
                      <option value="">— All —</option>
                      {statusOptionsEmployee.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Employee Type</label>
                    <select value={filters.employee_type || ""} onChange={(e) => setFilter("employee_type", e.target.value)} className={selectClass}>
                      <option value="">— All —</option>
                      <option value="Regular">Regular</option>
                      <option value="Lefty">Lefty</option>
                      <option value="Resign">Resign</option>
                      <option value="Close">Close</option>
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Employee ID</label>
                    <input type="text" value={filters.employee_id || ""} onChange={(e) => setFilter("employee_id", e.target.value)} placeholder="Search by code..." className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Gender</label>
                    <select value={filters.gender || ""} onChange={(e) => setFilter("gender", e.target.value)} className={selectClass}>
                      <option value="">— All —</option>
                      {genderOptions.map((g) => <option key={g.value} value={g.value}>{g.label}</option>)}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Blood Group</label>
                    <select value={filters.blood_group || ""} onChange={(e) => setFilter("blood_group", e.target.value)} className={selectClass}>
                      <option value="">— All —</option>
                      {bloodGroupOptions.map((b) => <option key={b.value} value={b.value}>{b.label}</option>)}
                    </select>
                  </div>
                </div>
                <div className="flex items-center gap-2 mt-4">
                  <Button onClick={() => { handleApply(); setMobileFilterOpen(false) }}>Apply</Button>
                  <Button variant="outline" onClick={() => { handleReset(); setMobileFilterOpen(false) }}>Reset</Button>
                </div>
              </div>
            </SheetContent>
          </Sheet>
          <Button onClick={handleGenerate} disabled={generating || selectedRows.length === 0} className="flex-1">
            {generating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <DownloadIcon className="mr-2 h-4 w-4" />}
            {generating ? "Gen..." : `Download (${selectedRows.length})`}
          </Button>
        </ButtonGroup>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <div className="rounded-lg border bg-card p-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Company</label>
              <select value={filters.company_id || ""} onChange={(e) => setFilter("company_id", e.target.value)} className={selectClass}>
                <option value="">— All —</option>
                {companies.map((c) => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Department</label>
              <select value={filters.department_id || ""} onChange={(e) => handleDepartmentChange(e.target.value)} className={selectClass}>
                <option value="">— All —</option>
                {departments.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Section</label>
              <select value={filters.section_id || ""} onChange={(e) => handleSectionChange(e.target.value)} className={selectClass} disabled={!filters.department_id}>
                <option value="">— All —</option>
                {sections.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Designation</label>
              <select value={filters.designation_id || ""} onChange={(e) => setFilter("designation_id", e.target.value)} className={selectClass} disabled={!filters.section_id}>
                <option value="">— All —</option>
                {designations.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Line</label>
              <select value={filters.line_id || ""} onChange={(e) => setFilter("line_id", e.target.value)} className={selectClass} disabled={!filters.section_id}>
                <option value="">— All —</option>
                {lines.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Shift</label>
              <select value={filters.shift_id || ""} onChange={(e) => setFilter("shift_id", e.target.value)} className={selectClass}>
                <option value="">— All —</option>
                {shifts.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Group</label>
              <select value={filters.group_id || ""} onChange={(e) => setFilter("group_id", e.target.value)} className={selectClass}>
                <option value="">— All —</option>
                {groups.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Floor</label>
              <select value={filters.floor_id || ""} onChange={(e) => setFilter("floor_id", e.target.value)} className={selectClass}>
                <option value="">— All —</option>
                {floors.map((f) => <option key={f.id} value={f.id}>{f.name}</option>)}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Status</label>
              <select value={filters.status || ""} onChange={(e) => setFilter("status", e.target.value)} className={selectClass}>
                <option value="">— All —</option>
                {statusOptionsEmployee.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Employee Type</label>
              <select value={filters.employee_type || ""} onChange={(e) => setFilter("employee_type", e.target.value)} className={selectClass}>
                <option value="">— All —</option>
                <option value="Regular">Regular</option>
                <option value="Lefty">Lefty</option>
                <option value="Resign">Resign</option>
                <option value="Close">Close</option>
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Employee ID</label>
              <input type="text" value={filters.employee_id || ""} onChange={(e) => setFilter("employee_id", e.target.value)} placeholder="Search by code..." className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Gender</label>
              <select value={filters.gender || ""} onChange={(e) => setFilter("gender", e.target.value)} className={selectClass}>
                <option value="">— All —</option>
                {genderOptions.map((g) => <option key={g.value} value={g.value}>{g.label}</option>)}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Blood Group</label>
              <select value={filters.blood_group || ""} onChange={(e) => setFilter("blood_group", e.target.value)} className={selectClass}>
                <option value="">— All —</option>
                {bloodGroupOptions.map((b) => <option key={b.value} value={b.value}>{b.label}</option>)}
              </select>
            </div>
          </div>

          <div className="flex items-center gap-2 mt-4">
            <Button onClick={handleApply}>Apply</Button>
            <Button variant="outline" onClick={handleReset}>Reset</Button>
          </div>
        </div>
      </div>

      <DataTable
        data={data}
        columns={columns}
        enableSelection={true}
        onSelectionChange={setSelectedRows}
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
