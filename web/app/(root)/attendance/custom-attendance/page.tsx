"use client"

import * as React from "react"
import { ClipboardEditIcon, Loader2, PencilIcon, FilterIcon, XIcon, CheckIcon, SquareIcon, CheckSquareIcon } from "lucide-react"
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

interface CustomRecord {
  id: string
  employee_id: string
  employee_name: string
  designation: string
  shift_name: string
  check_in: string
  check_out: string
  status: string
  date: string
  company_id: string
}

const today = new Date().toISOString().split("T")[0]

const statusMap: Record<string, string> = {
  present: "P", late: "L", absent: "A", half_day: "H", on_leave: "Lv", weekend: "W",
}

export default function CustomAttendancePage() {
  const [data, setData] = React.useState<CustomRecord[]>([])
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
  const [selected, setSelected] = React.useState<CustomRecord | null>(null)
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
    {
      key: "status", label: "Status", type: "select",
      options: [
        { value: "present", label: "Present" },
        { value: "late", label: "Late" },
        { value: "absent", label: "Absent" },
        { value: "half_day", label: "Half Day" },
        { value: "on_leave", label: "On Leave" },
        { value: "weekend", label: "Weekend" },
      ],
    },
  ], [companies, departments, sections, designations, lines, groups, shifts])

  const buildParams = React.useCallback((f?: Record<string, string>, p?: number) => {
    const params = f || filters
    const active: Record<string, string> = {
      start_date: params.start_date || today,
      end_date: params.end_date || today,
      page: String(p || page),
      limit: String(limit),
    }
    if (params.employee_id) active.employee_id = params.employee_id
    if (params.company_id) active.company_id = params.company_id
    if (params.department_id) active.department_id = params.department_id
    if (params.section_id) active.section_id = params.section_id
    if (params.designation_id) active.designation_id = params.designation_id
    if (params.line_id) active.line_id = params.line_id
    if (params.group_id) active.group_id = params.group_id
    if (params.shift_id) active.shift_id = params.shift_id
    if (params.status) active.status = params.status
    return active
  }, [filters, page, limit])

  const fetchData = React.useCallback(async (f?: Record<string, string>, p?: number) => {
    setLoading(true)
    setError("")
    try {
      const params = buildParams(f, p)
      const { data: res } = await attendanceApi.custom(params)
      setData(res.data || [])
      setTotal(res.total || 0)
    } catch {
      setError("Failed to load attendance data")
    } finally {
      setLoading(false)
    }
  }, [buildParams])

  React.useEffect(() => {
    const init = async () => {
      const [cRes, dRes, secRes, desRes, lRes, gRes, sRes] = await Promise.all([
        companyApi.list({ limit: "100" }),
        departmentApi.list({ limit: "100" }),
        sectionApi.list(undefined, { limit: "100" }),
        designationApi.list(undefined, { limit: "100" }),
        lineApi.list(undefined, { limit: "100" }),
        groupApi.list({ limit: "100" }),
        shiftApi.list({ limit: "100" }),
      ])
      if (Array.isArray(cRes.data?.data)) setCompanies(cRes.data.data)
      if (Array.isArray(dRes.data?.data)) setDepartments(dRes.data.data)
      if (Array.isArray(secRes.data?.data)) setSections(secRes.data.data)
      if (Array.isArray(desRes.data?.data)) setDesignations(desRes.data.data)
      if (Array.isArray(lRes.data?.data)) setLines(lRes.data.data)
      if (Array.isArray(gRes.data?.data)) setGroups(gRes.data.data)
      if (Array.isArray(sRes.data?.data)) setShifts(sRes.data.data)
    }
    init()
    fetchData()
  }, [])

  const handleChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleApply = () => {
    setPage(1)
    fetchData(filters, 1)
  }

  const handleReset = () => {
    setFilters({ start_date: today, end_date: today })
    setPage(1)
    fetchData({ start_date: today, end_date: today }, 1)
  }

  const toggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const toggleSelectAll = () => {
    if (selectedIds.size === data.length) {
      setSelectedIds(new Set())
    } else {
      setSelectedIds(new Set(data.map((r) => r.id)))
    }
  }

  const openBulkDialog = () => {
    setBulkInTime("")
    setBulkOutTime("")
    setBulkStatus("present")
    setBulkDialogOpen(true)
  }

  const handleBulkSubmit = async () => {
    if (selectedIds.size === 0) {
      toast.error("Please select employees from the table first")
      return
    }
    setBulkSaving(true)
    try {
      const attendanceIds: string[] = []
      for (const row of data) {
        if (selectedIds.has(row.id)) {
          attendanceIds.push(row.id)
        }
      }
      if (attendanceIds.length === 0) {
        toast.error("No records to submit")
        return
      }
      const { data: res } = await attendanceApi.bulkUpdateMissing({
        status: bulkStatus || undefined,
        inTime: bulkInTime || undefined,
        outTime: bulkOutTime || undefined,
        attendanceIds,
      })
      toast.success(res?.message || `Bulk attendance saved for ${attendanceIds.length} record(s)`)
      setBulkDialogOpen(false)
      setSelectedIds(new Set())
      fetchData(filters, page)
    } catch {
      toast.error("Failed to save bulk attendance")
    } finally {
      setBulkSaving(false)
    }
  }

  function formatDateDDMMYYYY(v: string | null | undefined): string {
    if (!v) return "-"
    const d = v.slice(0, 10)
    if (!/^\d{4}-\d{2}-\d{2}$/.test(d)) return v
    return d.split("-").reverse().join("/")
  }

  const extractDate = (v: string): string => {
    const d = v ? v.slice(0, 10) : ""
    return /^\d{4}-\d{2}-\d{2}$/.test(d) ? d : ""
  }

  const openSheet = (row: CustomRecord) => {
    setSelected(row)
    const rowDate = row.date ? row.date.slice(0, 10) : ""

    // Source of truth is the database attendance record, not raw device punches.
    const dbIn = row.check_in || ""
    const dbOut = row.check_out || ""

    let inTimeVal = ""
    let outTimeVal = ""

    if (!dbIn && dbOut) {
      inTimeVal = extractDate(dbOut) || rowDate
      outTimeVal = dbOut
    } else if (dbIn && !dbOut) {
      inTimeVal = dbIn
      outTimeVal = extractDate(dbIn) || rowDate
    } else if (dbIn && dbOut) {
      inTimeVal = dbIn
      outTimeVal = dbOut
    } else {
      inTimeVal = rowDate
      outTimeVal = rowDate
    }

    setInTime(inTimeVal)
    setOutTime(outTimeVal)
    setEntryStatus(row.status || "present")
    setSheetKey((k) => k + 1)
    setSheetOpen(true)
  }

  const handleSave = async () => {
    if (!selected) return
    setSaving(true)
    try {
      const sendIn = inTime || null
      const sendOut = outTime || null

      const payload: Record<string, unknown> = {
        employee_id: selected.employee_id,
        company_id: selected.company_id,
        date: selected.date,
        check_in: sendIn,
        check_out: sendOut,
        status: entryStatus,
      }

      await attendanceApi.update(selected.id, payload)
      toast.success("Attendance updated successfully")
      setSheetOpen(false)
      fetchData(filters, page)
    } catch {
      toast.error("Failed to update attendance")
    } finally {
      setSaving(false)
    }
  }

  const totalPages = Math.ceil(total / limit)

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      {/* Header */}
      <div className="px-4 lg:px-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2">
            <ClipboardEditIcon className="h-6 w-6 text-muted-foreground" />
            <div>
              <h1 className="text-2xl font-bold tracking-tight md:text-3xl">Custom Attendance</h1>
              <p className="text-muted-foreground mt-0.5 text-sm md:mt-1">Update existing attendance records and the attendance report</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {selectedIds.size > 0 && (
              <Button onClick={openBulkDialog} className="flex-1 sm:flex-none">
                <CheckIcon className="mr-2 h-4 w-4" />
                Bulk Update ({selectedIds.size})
              </Button>
            )}
            <div className="md:hidden">
              <Sheet open={mobileFilterOpen} onOpenChange={setMobileFilterOpen}>
                <SheetTrigger asChild>
                  <Button variant="outline">
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
            </div>
          </div>
        </div>
      </div>

      {/* Desktop Filter */}
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

      {/* Error */}
      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      {/* Table */}
      <div className="px-4 lg:px-6">
        <div className="rounded-lg border bg-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="px-3 py-2.5 w-10">
                    <button type="button" onClick={toggleSelectAll} className="text-muted-foreground hover:text-foreground">
                      {selectedIds.size === data.length && data.length > 0 ? (
                        <CheckSquareIcon className="size-4" />
                      ) : (
                        <SquareIcon className="size-4" />
                      )}
                    </button>
                  </th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Employee ID</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Name</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground hidden md:table-cell">Designation</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground hidden lg:table-cell">Shift</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Date</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">In Time</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Out Time</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Status</th>
                  <th className="px-3 py-2.5 text-right font-medium text-muted-foreground w-16">Action</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={10} className="px-3 py-12 text-center">
                      <Loader2 className="h-6 w-6 animate-spin text-muted-foreground mx-auto" />
                    </td>
                  </tr>
                ) : error ? (
                  <tr>
                    <td colSpan={10} className="px-3 py-8 text-center text-destructive">
                      {error}
                    </td>
                  </tr>
                ) : data.length === 0 ? (
                  <tr>
                    <td colSpan={10} className="px-3 py-8 text-center text-muted-foreground">
                      No attendance records found.
                    </td>
                  </tr>
                ) : (
                  data.map((row) => (
                    <tr key={row.id} className="border-b last:border-0 hover:bg-muted/30">
                      <td className="px-3 py-2">
                        <button type="button" onClick={() => toggleSelect(row.id)} className="text-muted-foreground hover:text-foreground">
                          {selectedIds.has(row.id) ? (
                            <CheckSquareIcon className="size-4" />
                          ) : (
                            <SquareIcon className="size-4" />
                          )}
                        </button>
                      </td>
                      <td className="px-3 py-2 font-mono text-xs">{row.employee_id}</td>
                      <td className="px-3 py-2 font-medium">{row.employee_name}</td>
                      <td className="px-3 py-2 text-muted-foreground hidden md:table-cell">{row.designation || "-"}</td>
                      <td className="px-3 py-2 hidden lg:table-cell">{row.shift_name || "-"}</td>
                      <td className="px-3 py-2">{formatDateDDMMYYYY(row.date)}</td>
                      <td className="px-3 py-2">
                        {row.check_in ? (
                          <span className="text-green-600 font-medium">{formatCheck(row.check_in)}</span>
                        ) : (
                          <span className="text-destructive font-medium">--:--</span>
                        )}
                      </td>
                      <td className="px-3 py-2">
                        {row.check_out ? (
                          <span className="text-green-600 font-medium">{formatCheck(row.check_out)}</span>
                        ) : (
                          <span className="text-destructive font-medium">--:--</span>
                        )}
                      </td>
                      <td className="px-3 py-2">
                        <Badge variant={row.status === "absent" ? "destructive" : "secondary"}>
                          {statusMap[row.status] || row.status}
                        </Badge>
                      </td>
                      <td className="px-3 py-2 text-right">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-8"
                          onClick={() => openSheet(row)}
                        >
                          <PencilIcon className="h-4 w-4" />
                        </Button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
          {total > 0 && (
            <div className="border-t bg-muted/30 px-4 py-3 flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Total Records: <b>{total}</b></span>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => { setPage(page - 1); fetchData(filters, page - 1) }}
                >
                  Previous
                </Button>
                <span className="text-muted-foreground">
                  Page {page} of {totalPages || 1}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= totalPages}
                  onClick={() => { setPage(page + 1); fetchData(filters, page + 1) }}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Edit Sheet */}
      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent side="right" className="w-full sm:max-w-md max-h-[100vh] overflow-y-auto">
          <SheetHeader>
            <SheetTitle>Edit Attendance</SheetTitle>
            <SheetDescription>Update check-in, check-out time and status</SheetDescription>
          </SheetHeader>

          {selected && (
            <div className="flex flex-col gap-4 p-4">
              <div className="rounded-lg bg-muted/30 p-3 space-y-1 text-sm">
                <div><span className="text-muted-foreground">Employee: </span><span className="font-medium">{selected.employee_name}</span></div>
                <div><span className="text-muted-foreground">Code: </span><span className="font-medium">{selected.employee_id}</span></div>
                <div><span className="text-muted-foreground">Designation: </span><span className="font-medium">{selected.designation || "-"}</span></div>
                <div><span className="text-muted-foreground">Shift: </span><span className="font-medium">{selected.shift_name || "-"}</span></div>
                <div><span className="text-muted-foreground">Date: </span><span className="font-medium">{formatDateDDMMYYYY(selected.date)}</span></div>
              </div>

              <div className="flex flex-col gap-3">
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">In Time</label>
                  <DateTimePicker
                    key={`edit-in-${sheetKey}`}
                    value={inTime}
                    onChange={setInTime}
                    autoFocusTime={!selected?.check_in}
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Out Time</label>
                  <DateTimePicker
                    key={`edit-out-${sheetKey}`}
                    value={outTime}
                    onChange={setOutTime}
                    autoFocusTime={!!selected?.check_in && !selected?.check_out}
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Status</label>
                  <select
                    value={entryStatus}
                    onChange={(e) => setEntryStatus(e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  >
                    <option value="present">Present</option>
                    <option value="absent">Absent</option>
                    <option value="late">Late</option>
                    <option value="half_day">Half Day</option>
                    <option value="on_leave">On Leave</option>
                    <option value="weekend">Weekend</option>
                  </select>
                </div>
                <div className="flex items-center gap-2 pt-1 text-xs text-muted-foreground">
                  Status will be: <Badge variant={entryStatus === "absent" ? "destructive" : "secondary"}>{statusMap[entryStatus] || entryStatus}</Badge>
                </div>
              </div>

              <Button onClick={handleSave} disabled={saving} className="w-full mt-2">
                {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Update Attendance Report
              </Button>
            </div>
          )}
        </SheetContent>
      </Sheet>

      {/* Bulk Update Dialog */}
      <Dialog open={bulkDialogOpen} onOpenChange={setBulkDialogOpen}>
        <DialogContent className="min-w-87.5 sm:min-w-[700px]">
          <DialogHeader>
            <DialogTitle className="text-lg">Bulk Attendance Update</DialogTitle>
            <p className="text-sm text-muted-foreground">{selectedIds.size} attendance record(s) selected</p>
          </DialogHeader>

          <div className="flex flex-col gap-4 py-2">
            <div className="rounded-md bg-muted/40 px-3 py-2.5 text-sm text-muted-foreground">
              The original attendance date for each record will be used automatically.
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">In Time</label>
                <TimePicker value={bulkInTime} onChange={setBulkInTime} />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Out Time</label>
                <TimePicker value={bulkOutTime} onChange={setBulkOutTime} />
              </div>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium">Status</label>
              <select
                value={bulkStatus}
                onChange={(e) => setBulkStatus(e.target.value)}
                className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              >
                <option value="present">Present</option>
                <option value="absent">Absent</option>
                <option value="late">Late</option>
                <option value="half_day">Half Day</option>
                <option value="on_leave">On Leave</option>
                <option value="weekend">Weekend</option>
              </select>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setBulkDialogOpen(false)}>Cancel</Button>
            <Button onClick={handleBulkSubmit} disabled={bulkSaving || selectedRows.length === 0}>
              {bulkSaving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Submit
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}