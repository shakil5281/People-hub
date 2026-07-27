"use client"

import * as React from "react"
import { UserXIcon, DownloadIcon, Loader2, FilterIcon, XIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { attendanceApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, shiftApi, groupApi } from "@/lib/api"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Shift { id: string; name: string }
interface Group { id: string; name: string }

interface AbsentRecord {
  id: string
  employee_id: string
  date: string
  status: string
  check_in: string | null
  employee?: { employee_id: string; name_en: string; designation_ref?: { name: string }; department?: { name: string }; section_ref?: { name: string } }
}

const columns: ColumnDef<AbsentRecord>[] = [
  {
    id: "sl", header: "Sl",
    cell: ({ row }) => row.index + 1,
  },
  {
    accessorKey: "date",
    header: "Date",
    cell: ({ row }) => row.original.date,
  },
  {
    accessorKey: "employee_id",
    header: "Employee ID",
    cell: ({ row }) => row.original.employee?.employee_id || row.original.employee_id,
  },
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => row.original.employee?.name_en || "-",
  },
  {
    accessorKey: "designation",
    header: "Designation",
    cell: ({ row }) => row.original.employee?.designation_ref?.name || "-",
  },
  {
    id: "department",
    header: "Department",
    cell: ({ row }) => row.original.employee?.department?.name || "-",
  },
  {
    id: "section",
    header: "Section",
    cell: ({ row }) => row.original.employee?.section_ref?.name || "-",
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: () => <span className="text-red-600 font-bold">A</span>,
  },
]

const selectClass = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"

const today = new Date().toISOString().split("T")[0]

export default function AbsentStatusPage() {
  const [data, setData] = React.useState<AbsentRecord[]>([])
  const [loading, setLoading] = React.useState(true)
  const [exporting, setExporting] = React.useState(false)
  const [error, setError] = React.useState("")

  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const [filters, setFilters] = React.useState<Record<string, string>>({
    start_date: today,
    end_date: today,
  })
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  const fetchData = async (f?: Record<string, string>, p?: number, l?: number) => {
    setError("")
    setLoading(true)
    try {
      const params = { ...(f || filters), page: String(p ?? page), limit: String(l ?? limit) }
      const { data: res } = await attendanceApi.absent(params)
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setError("Failed to load absent records")
    } finally {
      setLoading(false)
    }
  }

  React.useEffect(() => {
    const init = async () => {
      const [cRes, dRes, sRes, gRes] = await Promise.all([
        companyApi.list({ limit: "100" }),
        departmentApi.list({ limit: "100" }),
        shiftApi.list({ limit: "100" }),
        groupApi.list({ limit: "100" }),
      ])
      setCompanies(Array.isArray(cRes.data?.data) ? cRes.data.data : [])
      setDepartments(Array.isArray(dRes.data?.data) ? dRes.data.data : [])
      setShifts(Array.isArray(sRes.data?.data) ? sRes.data.data : [])
      setGroups(Array.isArray(gRes.data?.data) ? gRes.data.data : [])
    }
    init()
    fetchData()
  }, [])

  React.useEffect(() => {
    fetchData(filters)
  }, [page, limit])

  const handleDepartmentChange = async (value: string) => {
    setFilter("department_id", value)
    setFilter("section_id", "")
    setFilter("designation_id", "")
    setFilter("line_id", "")
    if (value) {
      try {
        const [secRes] = await Promise.all([sectionApi.list(value, { limit: "100" })])
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
    await fetchData(active, 1)
  }

  const handleReset = () => {
    setPage(1)
    setLimit(20)
    setFilters({ start_date: today, end_date: today })
    setSections([])
    setDesignations([])
    setLines([])
    fetchData({ start_date: today, end_date: today }, 1, 20)
  }

  const handleExport = async () => {
    setExporting(true)
    try {
      const active = Object.fromEntries(Object.entries(filters).filter(([, v]) => v !== ""))
      const res = await attendanceApi.exportAbsentExcel(active)
      const blob = new Blob([res.data], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `absent_report_${filters.start_date || today}.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setError("Failed to export absent report")
    } finally {
      setExporting(false)
    }
  }

  const setFilter = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <UserXIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Absent Report</h1>
            <p className="text-muted-foreground mt-1">View absent employee records</p>
          </div>
        </div>
        <div className="hidden md:block">
          <Button onClick={handleExport} disabled={exporting} variant="outline">
            {exporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <DownloadIcon className="mr-2 h-4 w-4" />}
            {exporting ? "Exporting..." : "Export Excel"}
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
                    <label className="text-xs font-medium text-muted-foreground">Start Date</label>
                    <input type="date" value={filters.start_date || ""} onChange={(e) => setFilter("start_date", e.target.value)} className={selectClass} />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">End Date</label>
                    <input type="date" value={filters.end_date || ""} onChange={(e) => setFilter("end_date", e.target.value)} className={selectClass} />
                  </div>
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
                </div>
                <div className="flex items-center gap-2 mt-4">
                  <Button onClick={() => { handleApply(); setMobileFilterOpen(false) }}>Apply</Button>
                  <Button variant="outline" onClick={() => { handleReset(); setMobileFilterOpen(false) }}>Reset</Button>
                </div>
              </div>
            </SheetContent>
          </Sheet>
          <Button onClick={handleExport} disabled={exporting} variant="outline" className="flex-1">
            {exporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <DownloadIcon className="mr-2 h-4 w-4" />}
            {exporting ? "Exporting..." : "Export"}
          </Button>
        </ButtonGroup>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <div className="rounded-lg border bg-card p-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Start Date</label>
              <input type="date" value={filters.start_date || ""} onChange={(e) => setFilter("start_date", e.target.value)} className={selectClass} />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">End Date</label>
              <input type="date" value={filters.end_date || ""} onChange={(e) => setFilter("end_date", e.target.value)} className={selectClass} />
            </div>

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
