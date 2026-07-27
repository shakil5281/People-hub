"use client"

import * as React from "react"
import { UserXIcon, Loader2, PencilIcon, FilterIcon, XIcon } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { attendanceApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"
import { formatCheck } from "@/lib/utils"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { DateTimePicker } from "@/components/ui/date-time-picker"
import { toast } from "sonner"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

interface Punch {
  time: string
  type: string
}

interface MissingRecord {
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
  punches?: Punch[]
}

const today = new Date().toISOString().split("T")[0]

const statusMap: Record<string, string> = {
  present: "P", late: "L", absent: "A", half_day: "H", on_leave: "Lv", weekend: "W",
}

export default function MissingAttendancePage() {
  const [data, setData] = React.useState<MissingRecord[]>([])
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
  const [selected, setSelected] = React.useState<MissingRecord | null>(null)
  const [inTime, setInTime] = React.useState("")
  const [outTime, setOutTime] = React.useState("")
  const [saving, setSaving] = React.useState(false)
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  const computedStatus = React.useMemo(() => {
    const hasIn = inTime.length > 0 && inTime.includes(":") && !inTime.endsWith("T")
    const hasOut = outTime.length > 0 && outTime.includes(":") && !outTime.endsWith("T")
    if (hasIn && hasOut) return "present"
    if (hasIn || hasOut) return "late"
    return selected?.status || "absent"
  }, [inTime, outTime, selected])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date_range", label: "Date Range", type: "daterange-split", dateRangeKeys: { start: "start_date", end: "end_date" } },
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
      const { data: res } = await attendanceApi.missing(params)
      setData(res.data || [])
      setTotal(res.total || 0)
    } catch {
      setError("Failed to load missing attendance data")
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

  function toDT(v: string): string {
    if (!v) return v
    return v.replace(" ", "T")
  }
  function fromDT(v: string): string {
    if (!v) return v
    return v.replace("T", " ")
  }

  const openSheet = (row: MissingRecord) => {
    setSelected(row)
    const punches = row.punches || []
    const punchIn = punches.find((x) => x.type === "I" || x.type === "i" || x.type === "0")
    const punchOut = punches.find((x) => x.type === "O" || x.type === "o" || x.type === "1")
    const bestIn = punchIn?.time || row.check_in
    const bestOut = punchOut?.time || row.check_out
    setInTime(bestIn ? toDT(bestIn) : row.date)
    setOutTime(bestOut ? toDT(bestOut) : row.date)
    setSheetOpen(true)
  }

  const handleSave = async () => {
    if (!selected) return
    setSaving(true)
    try {
      const sendIn = inTime && inTime.includes("T") ? fromDT(inTime) : ""
      const sendOut = outTime && outTime.includes("T") ? fromDT(outTime) : ""
      const dateUsed = sendIn ? sendIn.slice(0, 10) : sendOut ? sendOut.slice(0, 10) : selected.date
      await attendanceApi.create({
        employee_id: selected.employee_id,
        company_id: selected.company_id,
        date: dateUsed,
        ...(sendIn ? { check_in: sendIn } : {}),
        ...(sendOut ? { check_out: sendOut } : {}),
        status: computedStatus,
      })
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
      <div className="px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <UserXIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Missing Attendance</h1>
            <p className="text-muted-foreground mt-1">Attendance records missing check-in or check-out time</p>
          </div>
        </div>
        <div className="md:hidden mt-3">
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
        <div className="rounded-lg border bg-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground w-10">Sl</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Employee ID</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Name</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Designation</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Shift</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Date</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">In Time</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Out Time</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground">Status</th>
                  <th className="px-3 py-2.5 text-left font-medium text-muted-foreground w-16">Action</th>
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
                      No missing attendance found.
                    </td>
                  </tr>
                ) : (
                  data.map((row, i) => (
                    <tr key={row.id} className="border-b last:border-0 hover:bg-muted/30">
                      <td className="px-3 py-2">{(page - 1) * limit + i + 1}</td>
                      <td className="px-3 py-2 font-mono text-xs">{row.employee_id}</td>
                      <td className="px-3 py-2 font-medium">{row.employee_name}</td>
                      <td className="px-3 py-2 text-muted-foreground">{row.designation || "-"}</td>
                      <td className="px-3 py-2">{row.shift_name || "-"}</td>
                      <td className="px-3 py-2">{row.date?.split("-").reverse().join("-")}</td>
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
                      <td className="px-3 py-2">
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
              <span className="text-muted-foreground">Total Missing: <b>{total}</b></span>
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

      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent side="right">
          <SheetHeader>
            <SheetTitle>Edit Attendance</SheetTitle>
            <SheetDescription>Update check-in and check-out time</SheetDescription>
          </SheetHeader>

          {selected && (
            <div className="flex flex-col gap-4 p-4">
              <div className="rounded-lg bg-muted/30 p-3 space-y-1 text-sm">
                <div><span className="text-muted-foreground">Employee: </span><span className="font-medium">{selected.employee_name}</span></div>
                <div><span className="text-muted-foreground">Code: </span><span className="font-medium">{selected.employee_id}</span></div>
                <div><span className="text-muted-foreground">Designation: </span><span className="font-medium">{selected.designation || "-"}</span></div>
                <div><span className="text-muted-foreground">Shift: </span><span className="font-medium">{selected.shift_name || "-"}</span></div>
                <div><span className="text-muted-foreground">Date: </span><span className="font-medium">{selected.date}</span></div>
                {selected.punches && selected.punches.length > 0 && (
                  <div className="pt-1.5 border-t border-border mt-1.5">
                    <span className="text-muted-foreground text-xs">Punch Logs: </span>
                    <div className="flex flex-wrap gap-1.5 mt-1">
                      {selected.punches.map((p, i) => (
                        <Badge key={i} variant={p.type === "I" ? "default" : "secondary"} className="text-xs">
                          {p.type === "I" ? "IN" : "OUT"} {formatCheck(p.time)}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              <div className="flex flex-col gap-3">
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">In Time</label>
                  <DateTimePicker value={inTime} onChange={setInTime} />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Out Time</label>
                  <DateTimePicker value={outTime} onChange={setOutTime} />
                </div>
                <div className="flex items-center gap-2 pt-1 text-xs text-muted-foreground">
                  Status will be: <Badge variant={computedStatus === "absent" ? "destructive" : "secondary"}>{statusMap[computedStatus] || computedStatus}</Badge>
                </div>
              </div>

              <Button onClick={handleSave} disabled={saving} className="w-full mt-2">
                {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Save
              </Button>
            </div>
          )}
        </SheetContent>
      </Sheet>
    </div>
  )
}
