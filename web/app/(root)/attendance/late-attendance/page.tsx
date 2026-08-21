"use client"

import * as React from "react"
import { ClockIcon, Loader2, PencilIcon, FilterIcon, XIcon, CheckIcon, SquareIcon, CheckSquareIcon } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { attendanceApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"
import { formatCheck } from "@/lib/utils"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { DateTimePicker } from "@/components/ui/date-time-picker"
import { TimePicker } from "@/components/ui/time-picker"
import { toast } from "sonner"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

interface LateRecord {
  id: string
  employee_id: string
  employee_name: string
  designation: string
  shift_name: string
  check_in: string
  check_out: string
  status: string
  late_minutes: number
  date: string
  company_id: string
}

const today = new Date().toISOString().split("T")[0]

const statusMap: Record<string, string> = {
  present: "P", late: "L", absent: "A", half_day: "H", on_leave: "Lv", weekend: "W",
}

export default function LateAttendancePage() {
  const [data, setData] = React.useState<LateRecord[]>([])
  const [total, setTotal] = React.useState(0)
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState("")
  const [page, setPage] = React.useState(1)
  const [limit] = React.useState(20)
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])
  const [filters, setFilters] = React.useState<Record<string, string>>({
    start_date: today,
    end_date: today,
  })

  const [sheetOpen, setSheetOpen] = React.useState(false)
  const [selected, setSelected] = React.useState<LateRecord | null>(null)
  const [inTime, setInTime] = React.useState("")
  const [outTime, setOutTime] = React.useState("")
  const [entryStatus, setEntryStatus] = React.useState("present")
  const [saving, setSaving] = React.useState(false)
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)
  const [sheetKey, setSheetKey] = React.useState(0)

  // Bulk selection
  const [selectedIds, setSelectedIds] = React.useState<Set<string>>(new Set())
  const [bulkDialogOpen, setBulkDialogOpen] = React.useState(false)
  const [bulkInTime, setBulkInTime] = React.useState("")
  const [bulkOutTime, setBulkOutTime] = React.useState("")
  const [bulkStatus, setBulkStatus] = React.useState("present")
  const [bulkSaving, setBulkSaving] = React.useState(false)

  const selectedRows = React.useMemo(
    () => data.filter((r) => selectedIds.has(r.id)),
    [data, selectedIds],
  )

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date_range", label: "Date Range", type: "daterange-split", dateRangeKeys: { start: "start_date", end: "end_date" } },
    { key: "employee_id", label: "Employee ID", type: "text" },
    {
      key: "company_id", label: "Company", type: "select",
      options: companies.map((c) => ({ value: c.id, label: c.company_name_en })),
    },
    {
      key: "department_id", label: "Department", type: "select",
      options: departments.map((d) => ({ value: d.id, label: d.name })),
    },
    {
      key: "section_id", label: "Section", type: "select",
      options: sections.map((s) => ({ value: s.id, label: s.name })),
    },
    {
      key: "designation_id", label: "Designation", type: "select",
      options: designations.map((d) => ({ value: d.id, label: d.name })),
    },
    {
      key: "line_id", label: "Line", type: "select",
      options: lines.map((l) => ({ value: l.id, label: l.name })),
    },
    {
      key: "group_id", label: "Group", type: "select",
      options: groups.map((g) => ({ value: g.id, label: g.name })),
    },
    {
      key: "shift_id", label: "Shift", type: "select",
      options: shifts.map((s) => ({ value: s.id, label: s.name })),
    },
  ], [companies, departments, sections, designations, lines, groups, shifts])

  const loadData = React.useCallback(
    async (p: number, f: Record<string, string>) => {
      setLoading(true)
      setError("")
      try {
        const query: Record<string, string> = {
          page: String(p),
          limit: String(limit),
          start_date: f.start_date || today,
          end_date: f.end_date || today,
        }
        if (f.company_id) query.company_id = f.company_id
        if (f.department_id) query.department_id = f.department_id
        if (f.section_id) query.section_id = f.section_id
        if (f.designation_id) query.designation_id = f.designation_id
        if (f.line_id) query.line_id = f.line_id
        if (f.group_id) query.group_id = f.group_id
        if (f.shift_id) query.shift_id = f.shift_id
        if (f.employee_id) query.employee_id = f.employee_id

        const res = await attendanceApi.late(query)
        const items: LateRecord[] = res.data?.data || []
        const totalCount: number = res.data?.meta?.total || items.length
        setData(items)
        setTotal(totalCount)
      } catch {
        setError("Failed to load late attendance records")
      } finally {
        setLoading(false)
      }
    },
    [limit],
  )

  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      groupApi.list({ limit: "100" }),
      shiftApi.list({ limit: "100" }),
    ])
      .then(([cRes, dRes, gRes, sRes]) => {
        setCompanies(cRes.data?.data || [])
        setDepartments(dRes.data?.data || [])
        setGroups(gRes.data?.data || [])
        setShifts(sRes.data?.data || [])
      })
      .catch(() => {})
    loadData(1, filters)
  }, [])

  React.useEffect(() => {
    if (!filters.department_id) { setSections([]); setDesignations([]); setLines([]); return }
    sectionApi.list(filters.department_id, { limit: "100" })
      .then((r) => setSections(r.data?.data || []))
      .catch(() => setSections([]))
  }, [filters.department_id])

  React.useEffect(() => {
    if (!filters.section_id) { setDesignations([]); setLines([]); return }
    Promise.all([
      designationApi.list(filters.section_id, { limit: "100" }),
      lineApi.list(filters.section_id, { limit: "100" }),
    ])
      .then(([dr, lr]) => {
        setDesignations(dr.data?.data || [])
        setLines(lr.data?.data || [])
      })
      .catch(() => { setDesignations([]); setLines([]) })
  }, [filters.section_id])

  const handleFilterChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleApply = () => {
    setPage(1)
    loadData(1, filters)
    setMobileFilterOpen(false)
  }

  const handleReset = () => {
    const empty: Record<string, string> = { start_date: today, end_date: today }
    setFilters(empty)
    setPage(1)
    loadData(1, empty)
    setMobileFilterOpen(false)
  }

  const handleOpenEdit = (rec: LateRecord) => {
    setSelected(rec)
    setInTime(rec.check_in || "")
    setOutTime(rec.check_out || "")
    setEntryStatus("present")
    setSheetKey((k) => k + 1)
    setSheetOpen(true)
  }

  const handleSaveSingle = async () => {
    if (!selected) return
    setSaving(true)
    try {
      await attendanceApi.fixSingleLate({
        attendance_id: selected.id,
        employee_id: selected.employee_id,
        company_id: selected.company_id,
        date: selected.date,
        check_in: inTime,
        check_out: outTime,
        status: entryStatus,
      })
      toast.success("Late attendance fixed and saved to missing attendance!")
      setSheetOpen(false)
      loadData(page, filters)
    } catch {
      toast.error("Failed to fix late attendance")
    } finally {
      setSaving(false)
    }
  }

  const toggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const toggleAll = () => {
    if (selectedIds.size === data.length) {
      setSelectedIds(new Set())
    } else {
      setSelectedIds(new Set(data.map((d) => d.id)))
    }
  }

  const handleOpenBulk = () => {
    setBulkInTime("")
    setBulkOutTime("")
    setBulkStatus("present")
    setBulkDialogOpen(true)
  }

  const handleSaveBulk = async () => {
    if (selectedIds.size === 0) return
    setBulkSaving(true)
    try {
      const items = selectedRows.map((r) => ({
        attendance_id: r.id,
        employee_id: r.employee_id,
        company_id: r.company_id,
        date: r.date,
        check_in: bulkInTime || r.check_in,
        check_out: bulkOutTime || r.check_out,
        status: bulkStatus,
      }))
      await attendanceApi.fixBulkLate({ items })
      toast.success(`Fixed ${items.length} late attendance records!`)
      setBulkDialogOpen(false)
      setSelectedIds(new Set())
      loadData(page, filters)
    } catch {
      toast.error("Failed to fix bulk late attendance")
    } finally {
      setBulkSaving(false)
    }
  }

  const totalPages = Math.ceil(total / limit)

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-2">
          <div className="flex items-center gap-2">
            <ClockIcon className="h-6 w-6 text-orange-500" />
            <div>
              <h1 className="text-lg md:text-3xl font-bold tracking-tight">Late Attendance</h1>
              <p className="text-muted-foreground mt-1">Manage and fix employee late attendance records into preset attendance</p>
            </div>
          </div>

          <div className="md:hidden flex items-center justify-between gap-2 mt-3">
            <div className="flex items-center gap-2">
              <Sheet open={mobileFilterOpen} onOpenChange={setMobileFilterOpen}>
                <SheetTrigger asChild>
                  <Button variant="outline" className="w-full justify-start">
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
                    <FilterBar
                      filters={filterDefs}
                      values={filters}
                      onChange={handleFilterChange}
                      onApply={handleApply}
                      onReset={handleReset}
                      submitting={loading}
                      singleColumn
                      noBorder
                    />
                  </div>
                </SheetContent>
              </Sheet>

              {selectedIds.size > 0 && (
                <Button size="sm" onClick={handleOpenBulk} className="bg-emerald-600 hover:bg-emerald-700 text-white">
                  <CheckIcon className="h-4 w-4 mr-1" />
                  Fix Late ({selectedIds.size})
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleFilterChange}
          onApply={handleApply}
          onReset={handleReset}
          submitting={loading}
        />
      </div>

      <div className="px-4 lg:px-6">
        <div className="rounded-lg border bg-card overflow-hidden">
          {selectedIds.size > 0 && (
            <div className="bg-muted/60 border-b px-4 py-2.5 flex items-center justify-between">
              <span className="text-sm font-medium">
                {selectedIds.size} of {data.length} row(s) selected
              </span>
              <Button size="sm" onClick={handleOpenBulk} className="bg-emerald-600 hover:bg-emerald-700 text-white">
                <CheckIcon className="h-4 w-4 mr-1" />
                Fix Late ({selectedIds.size})
              </Button>
            </div>
          )}

          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/30">
                  <th className="px-3 py-2.5 w-10 text-center">
                    <button type="button" onClick={toggleAll} className="p-0.5 rounded hover:bg-muted text-muted-foreground">
                      {data.length > 0 && selectedIds.size === data.length ? (
                        <CheckSquareIcon className="h-4 w-4 text-primary" />
                      ) : (
                        <SquareIcon className="h-4 w-4" />
                      )}
                    </button>
                  </th>
                  <th className="px-3 py-2.5 text-left font-semibold w-10">#</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Employee</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Shift</th>
                  <th className="px-3 py-2.5 text-left font-semibold">In Time</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Out Time</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Late (min)</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Status</th>
                  <th className="px-3 py-2.5 text-left font-semibold">Date</th>
                  <th className="px-3 py-2.5 text-right font-semibold">Action</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={10} className="px-3 py-12 text-center">
                      <Loader2 className="h-6 w-6 animate-spin mx-auto text-muted-foreground" />
                    </td>
                  </tr>
                ) : error ? (
                  <tr>
                    <td colSpan={10} className="px-3 py-8 text-center text-destructive">{error}</td>
                  </tr>
                ) : data.length === 0 ? (
                  <tr>
                    <td colSpan={10} className="px-3 py-8 text-center text-muted-foreground">No late attendance records found</td>
                  </tr>
                ) : (
                  data.map((row, i) => {
                    const isChecked = selectedIds.has(row.id)
                    return (
                      <tr key={row.id} className={`border-b last:border-0 hover:bg-muted/50 transition-colors ${isChecked ? "bg-muted/40" : ""}`}>
                        <td className="px-3 py-2 text-center">
                          <button type="button" onClick={() => toggleSelect(row.id)} className="p-0.5 rounded hover:bg-muted text-muted-foreground">
                            {isChecked ? (
                              <CheckSquareIcon className="h-4 w-4 text-primary" />
                            ) : (
                              <SquareIcon className="h-4 w-4" />
                            )}
                          </button>
                        </td>
                        <td className="px-3 py-2">{(page - 1) * limit + i + 1}</td>
                        <td className="px-3 py-2 font-medium">
                          <div>{row.employee_name}</div>
                          <div className="text-xs text-muted-foreground">{row.employee_id} {row.designation ? `• ${row.designation}` : ""}</div>
                        </td>
                        <td className="px-3 py-2">{row.shift_name || "-"}</td>
                        <td className="px-3 py-2">{formatCheck(row.check_in)}</td>
                        <td className="px-3 py-2">{formatCheck(row.check_out)}</td>
                        <td className="px-3 py-2 text-orange-600 font-medium">{row.late_minutes} min</td>
                        <td className="px-3 py-2">
                          <Badge variant="outline" className="border-orange-200 bg-orange-50 text-orange-700">
                            {statusMap[row.status] || row.status}
                          </Badge>
                        </td>
                        <td className="px-3 py-2">{row.date}</td>
                        <td className="px-3 py-2 text-right">
                          <Button variant="ghost" size="icon-sm" onClick={() => handleOpenEdit(row)}>
                            <PencilIcon className="h-3.5 w-3.5" />
                          </Button>
                        </td>
                      </tr>
                    )
                  })
                )}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="border-t px-4 py-3 flex items-center justify-between">
              <span className="text-xs text-muted-foreground">
                Showing {(page - 1) * limit + 1}–{Math.min(page * limit, total)} of {total}
              </span>
              <div className="flex gap-1">
                <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => { setPage(p => p - 1); loadData(page - 1, filters) }}>
                  Previous
                </Button>
                <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => { setPage(p => p + 1); loadData(page + 1, filters) }}>
                  Next
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Single Edit Sheet */}
      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent side="right" key={sheetKey} className="sm:max-w-md p-0 flex flex-col">
          <SheetHeader className="px-6 py-4 border-b">
            <SheetTitle>Fix Late Attendance</SheetTitle>
            <SheetDescription>
              Adjust Check-In/Check-Out to grant preset attendance. This change will be saved to missing attendance.
            </SheetDescription>
          </SheetHeader>

          {selected && (
            <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
              <div className="rounded-md border p-3 bg-muted/30 text-sm space-y-1">
                <div className="font-semibold">{selected.employee_name} ({selected.employee_id})</div>
                <div className="text-muted-foreground">Date: {selected.date} | Shift: {selected.shift_name || "General"}</div>
                <div className="text-orange-600 font-medium">Late Minutes: {selected.late_minutes} min</div>
              </div>

              <div className="space-y-1">
                <label className="text-sm font-medium">Check-In</label>
                <DateTimePicker
                  value={inTime}
                  onChange={setInTime}
                />
              </div>

              <div className="space-y-1">
                <label className="text-sm font-medium">Check-Out</label>
                <DateTimePicker
                  value={outTime}
                  onChange={setOutTime}
                />
              </div>

              <div className="space-y-1">
                <label className="text-sm font-medium">Status</label>
                <select
                  value={entryStatus}
                  onChange={(e) => setEntryStatus(e.target.value)}
                  className="w-full border rounded-md p-2 text-sm bg-background"
                >
                  <option value="present">Present (P)</option>
                  <option value="late">Late (L)</option>
                  <option value="absent">Absent (A)</option>
                </select>
              </div>
            </div>
          )}

          <div className="border-t p-4 flex justify-end gap-2 bg-muted/20">
            <SheetClose asChild>
              <Button variant="outline">Cancel</Button>
            </SheetClose>
            <Button onClick={handleSaveSingle} disabled={saving} className="bg-emerald-600 hover:bg-emerald-700 text-white">
              {saving ? "Saving..." : "Save & Fix"}
            </Button>
          </div>
        </SheetContent>
      </Sheet>

      {/* Bulk Edit Dialog */}
      <Dialog open={bulkDialogOpen} onOpenChange={setBulkDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Bulk Fix Late Attendance ({selectedIds.size} employees)</DialogTitle>
          </DialogHeader>

          <div className="space-y-4 py-2 text-sm">
            <p className="text-muted-foreground">
              Fix late attendance for {selectedIds.size} selected employee(s). The records will be updated and persisted in missing attendance.
            </p>

            <div className="space-y-1">
              <label className="text-sm font-medium">Override Check-In Time (Optional)</label>
              <TimePicker value={bulkInTime} onChange={setBulkInTime} />
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium">Override Check-Out Time (Optional)</label>
              <TimePicker value={bulkOutTime} onChange={setBulkOutTime} />
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium">New Status</label>
              <select
                value={bulkStatus}
                onChange={(e) => setBulkStatus(e.target.value)}
                className="w-full border rounded-md p-2 text-sm bg-background"
              >
                <option value="present">Present (P)</option>
                <option value="late">Late (L)</option>
                <option value="absent">Absent (A)</option>
              </select>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setBulkDialogOpen(false)}>Cancel</Button>
            <Button onClick={handleSaveBulk} disabled={bulkSaving} className="bg-emerald-600 hover:bg-emerald-700 text-white">
              {bulkSaving ? "Applying..." : "Apply Bulk Fix"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
