"use client"

import * as React from "react"
import { ClipboardListIcon, Loader2, ChevronLeftIcon, ChevronRightIcon, FilterIcon, XIcon, FileText, FileSpreadsheet } from "lucide-react"
import { format } from "date-fns"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { attendanceApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"
import { formatCheck } from "@/lib/utils"

interface JobCardRecord {
  id: string
  employee_id: string
  date: string
  check_in: string | null
  check_out: string | null
  total_hours: string | null
  over_time: string | null
  status: string
  late_minutes: number
  shift_name: string
}

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

const today = new Date().toISOString().split("T")[0]

const statusMapEn: Record<string, string> = {
  present: "P", late: "L", absent: "A", half_day: "HD", leave: "V", on_leave: "V", weekend: "W", holiday: "H",
}

const statusMapBn: Record<string, string> = {
  present: "উপস্থিত", late: "বিলম্বে", absent: "অনুপস্থিত", half_day: "অর্ধদিবস", leave: "ছুটি", on_leave: "ছুটি", weekend: "সাপ্তাহিক ছুটি", holiday: "সরকারি ছুটি",
}

const dayMapBn: Record<string, string> = {
  Sun: "রবিবার", Mon: "সোমবার", Tue: "মঙ্গলবার", Wed: "বুধবার", Thu: "বৃহস্পতিবার", Fri: "শুক্রবার", Sat: "শনিবার",
}

export default function JobCardPage() {
  const [data, setData] = React.useState<JobCardRecord[]>([])
  const [employees, setEmployees] = React.useState<{employee_id: string; name_en: string; designation: string; department: string; phone: string; joining_date: string; company: string}[]>([])
  const [currentIndex, setCurrentIndex] = React.useState(0)
  const [loading, setLoading] = React.useState(false)
  const [fetchingList, setFetchingList] = React.useState(false)
  const [error, setError] = React.useState("")
  const [submitting, setSubmitting] = React.useState(false)
  const [lang, setLang] = React.useState<"en" | "bn">("en")
  const [filters, setFilters] = React.useState<Record<string, string>>({
    start_date: today,
    end_date: today,
  })
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)
 
  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      groupApi.list({ limit: "100" }),
      shiftApi.list({ limit: "100" }),
    ]).then(([cRes, dRes, gRes, sRes]) => {
      setCompanies(cRes.data?.data || [])
      setDepartments(dRes.data?.data || [])
      setGroups(gRes.data?.data || [])
      setShifts(sRes.data?.data || [])
    }).catch(() => {})
    handleApply()
  }, [])

  React.useEffect(() => {
    if (!filters.department_id) { setSections([]); setDesignations([]); setLines([]); return }
    sectionApi.list(filters.department_id, { limit: "100" }).then((r) => setSections(r.data?.data || [])).catch(() => setSections([]))
  }, [filters.department_id])

  React.useEffect(() => {
    if (!filters.section_id) { setDesignations([]); setLines([]); return }
    Promise.all([
      designationApi.list(filters.section_id, { limit: "100" }),
      lineApi.list(filters.section_id, { limit: "100" }),
    ]).then(([dr, lr]) => {
      setDesignations(dr.data?.data || [])
      setLines(lr.data?.data || [])
    }).catch(() => { setDesignations([]); setLines([]) })
  }, [filters.section_id])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date_range", label: "Date Range", type: "daterange-split", dateRangeKeys: { start: "start_date", end: "end_date" } },
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en })) },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    { key: "section_id", label: "Section", type: "select", options: sections.map((s) => ({ value: s.id, label: s.name })), disabled: !filters.department_id },
    { key: "designation_id", label: "Designation", type: "select", options: designations.map((d) => ({ value: d.id, label: d.name })), disabled: !filters.section_id },
    { key: "line_id", label: "Line", type: "select", options: lines.map((l) => ({ value: l.id, label: l.name })), disabled: !filters.section_id },
    { key: "group_id", label: "Group", type: "select", options: groups.map((g) => ({ value: g.id, label: g.name })) },
    { key: "shift_id", label: "Shift", type: "select", options: shifts.map((s) => ({ value: s.id, label: s.name })) },
    { key: "status", label: "Status", type: "select", options: [
      { value: "present", label: "Present" }, { value: "late", label: "Late" },
      { value: "absent", label: "Absent" }, { value: "half_day", label: "Half Day" },
    ] },
    { key: "employee_id", label: "Employee ID", type: "text", placeholder: "Enter employee code..." },
  ], [companies, departments, sections, designations, lines, groups, shifts, filters.department_id, filters.section_id])

  const buildParams = (f?: Record<string, string>) => {
    const params = f || filters
    const active: Record<string, string> = {
      start_date: params.start_date || today,
      end_date: params.end_date || today,
    }
    if (params.company_id) active.company_id = params.company_id
    if (params.department_id) active.department_id = params.department_id
    if (params.section_id) active.section_id = params.section_id
    if (params.designation_id) active.designation_id = params.designation_id
    if (params.line_id) active.line_id = params.line_id
    if (params.group_id) active.group_id = params.group_id
    if (params.shift_id) active.shift_id = params.shift_id
    if (params.status) active.status = params.status
    if (params.employee_id) active.employee_id = params.employee_id
    return active
  }

  const fetchEmployeeList = async (f?: Record<string, string>) => {
    setFetchingList(true)
    try {
      const params = buildParams(f || filters)
      params.list_mode = "true"
      const { data: res } = await attendanceApi.jobCard(params)
      setEmployees(res.data || [])
      setCurrentIndex(0)
      return res.data || []
    } catch {
      setError("Failed to load employee list")
      return []
    } finally {
      setFetchingList(false)
    }
  }

  const fetchEmployeeData = async (empId: string) => {
    setLoading(true)
    setError("")
    try {
      const params = buildParams()
      params.employee_id = empId
      delete params.list_mode
      const { data: res } = await attendanceApi.jobCard(params)
      setData(res.data || [])
    } catch {
      setError("Failed to load attendance data")
    } finally {
      setLoading(false)
    }
  }

  const handleApply = async () => {
    setSubmitting(true)
    const empList = await fetchEmployeeList()
    setSubmitting(false)
    if (empList.length > 0) {
      fetchEmployeeData(empList[0].employee_id)
    } else {
      setData([])
    }
    setMobileFilterOpen(false)
  }

  const handleReset = () => {
    setFilters({ start_date: today, end_date: today })
    setData([])
    setEmployees([])
    setCurrentIndex(0)
    setError("")
    setMobileFilterOpen(false)
  }

  const handlePrev = () => {
    if (currentIndex > 0) {
      const newIdx = currentIndex - 1
      setCurrentIndex(newIdx)
      fetchEmployeeData(employees[newIdx].employee_id)
    }
  }

  const handleNext = () => {
    if (currentIndex < employees.length - 1) {
      const newIdx = currentIndex + 1
      setCurrentIndex(newIdx)
      fetchEmployeeData(employees[newIdx].employee_id)
    }
  }

  const downloadExport = (res: { data: Blob }, filename: string) => {
    const blob = new Blob([res.data], { type: res.data.type || "application/octet-stream" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  }

  const handleExport = async (format: "pdf" | "excel", lang?: string) => {
    if (!currentEmpId) {
      toast.error("No employee selected to export")
      return
    }
    setSubmitting(true)
    try {
      const params = buildParams()
      params.employee_id = currentEmpId
      if (lang) params.lang = lang
      const res = format === "pdf"
        ? await attendanceApi.exportJobCardPdf(params)
        : await attendanceApi.exportJobCardExcel(params)
      const range = `${(filters.start_date || today).replace(/-/g, "")}_${(filters.end_date || today).replace(/-/g, "")}`
      const name = `job_card_${currentEmpId}_${range}${lang ? `_${lang}` : ""}.${format === "pdf" ? "pdf" : "xlsx"}`
      downloadExport(res, name)
      toast.success(format === "pdf" ? "PDF exported" : "Excel exported")
    } catch {
      toast.error("Failed to export")
    } finally {
      setSubmitting(false)
    }
  }

  const handleFilterChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const currentEmpId = employees[currentIndex]?.employee_id || ""
  const emp = employees[currentIndex] || null
  const totalByStatus = data.reduce<Record<string, number>>((acc, r) => {
    acc[r.status] = (acc[r.status] || 0) + 1
    return acc
  }, {})
  const totalLateMinutes = data.reduce((sum, r) => sum + (r.late_minutes || 0), 0)
  const totalOT = data.reduce((sum, r) => sum + (Number(r.over_time) || 0), 0)

  const activeStatusMap = lang === "bn" ? statusMapBn : statusMapEn

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-2">
          <div className="flex items-center gap-2">
            <ClipboardListIcon className="h-6 w-6 text-muted-foreground" />
            <div>
              <h1 className="text-lg md:text-3xl font-bold tracking-tight">
                {lang === "bn" ? "চাকুরী কার্ড" : "Job Card"}
              </h1>
              <p className="text-muted-foreground mt-1">
                {lang === "bn" ? "কর্মচারীদের উপস্থিতি জব কার্ড প্রতিবেদন" : "Employee attendance job card report"}
              </p>
            </div>
          </div>
          <div className="hidden md:flex items-center gap-2">
            <div className="flex border rounded-md overflow-hidden">
              <button
                className={`px-2 py-1 text-xs font-medium ${lang === "en" ? "bg-primary text-primary-foreground" : "bg-muted"}`}
                onClick={() => setLang("en")}
              >
                EN
              </button>
              <button
                className={`px-2 py-1 text-xs font-medium ${lang === "bn" ? "bg-primary text-primary-foreground" : "bg-muted"}`}
                onClick={() => setLang("bn")}
              >
                BN
              </button>
            </div>
            <Button variant="outline" size="sm" onClick={() => handleExport("excel", lang)} disabled={submitting || !currentEmpId}>
              <FileSpreadsheet className="h-4 w-4 mr-1.5 text-emerald-600" />
              {submitting ? "Exporting..." : "Excel"}
            </Button>
            <Button variant="outline" size="sm" onClick={() => handleExport("pdf", lang)} disabled={submitting || !currentEmpId}>
              <FileText className="h-4 w-4 mr-1.5 text-destructive" />
              {submitting ? "Exporting..." : "PDF"}
            </Button>
          </div>
        </div>
        <div className="md:hidden flex items-center justify-end gap-2 mt-3">
          <Sheet open={mobileFilterOpen} onOpenChange={setMobileFilterOpen}>
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
                <FilterBar
                  filters={filterDefs}
                  values={filters}
                  onChange={handleFilterChange}
                  onApply={handleApply}
                  onReset={handleReset}
                  submitting={submitting}
                  singleColumn
                  noBorder
                />
              </div>
            </SheetContent>
          </Sheet>
        </div>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleFilterChange}
          onApply={handleApply}
          onReset={handleReset}
          submitting={submitting}
        />
      </div>

      <div className="px-4 lg:px-6">
        <div className="rounded-lg border bg-card overflow-hidden">
          {employees.length > 0 && (
            <div className="border-b px-4 py-3 flex items-center justify-between">
              <Button variant="outline" size="sm" onClick={handlePrev} disabled={currentIndex === 0 || loading}>
                <ChevronLeftIcon className="h-4 w-4 mr-1" />Previous
              </Button>
              <span className="text-sm text-muted-foreground">Employee {currentIndex + 1} of {employees.length}</span>
              <Button variant="outline" size="sm" onClick={handleNext} disabled={currentIndex >= employees.length - 1 || loading}>
                Next<ChevronRightIcon className="h-4 w-4 ml-1" />
              </Button>
            </div>
          )}

          {emp && (
            <div className="border-b bg-muted/30 px-4 py-3">
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-sm">
                <div><span className="text-muted-foreground">{lang === "bn" ? "নাম: " : "Name: "}</span><span className="font-medium">{emp.name_en}</span></div>
                <div><span className="text-muted-foreground">{lang === "bn" ? "কর্মী আইডি: " : "Code: "}</span><span className="font-medium">{emp.employee_id}</span></div>
                <div><span className="text-muted-foreground">{lang === "bn" ? "পদবী: " : "Designation: "}</span><span className="font-medium">{emp.designation || "-"}</span></div>
                <div><span className="text-muted-foreground">{lang === "bn" ? "বিভাগ: " : "Department: "}</span><span className="font-medium">{emp.department || "-"}</span></div>
                <div><span className="text-muted-foreground">{lang === "bn" ? "মোবাইল: " : "Phone: "}</span><span className="font-medium">{emp.phone || "-"}</span></div>
                <div><span className="text-muted-foreground">{lang === "bn" ? "যোগদানের তারিখ: " : "Joining: "}</span><span className="font-medium">{emp.joining_date ? format(new Date(emp.joining_date), "dd-MM-yyyy") : "-"}</span></div>
                <div><span className="text-muted-foreground">{lang === "bn" ? "সময়কাল: " : "Period: "}</span><span className="font-medium">
                  {filters.start_date ? `${filters.start_date.split("-").reverse().join("-")} ${lang === "bn" ? "হতে" : "-"} ${(filters.end_date || filters.start_date).split("-").reverse().join("-")} ${lang === "bn" ? "পর্যন্ত" : ""}` : "-"}
                </span></div>
                <div><span className="text-muted-foreground">{lang === "bn" ? "মোট কার্যদিবস: " : "Days: "}</span><span className="font-medium">{data.length}</span></div>
              </div>
            </div>
          )}

          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="px-3 py-2.5 text-left font-semibold w-10">#</th>
                  <th className="px-3 py-2.5 text-left font-semibold">{lang === "bn" ? "তারিখ" : "Date"}</th>
                  <th className="px-3 py-2.5 text-left font-semibold">{lang === "bn" ? "শিফট" : "Shift"}</th>
                  <th className="px-3 py-2.5 text-left font-semibold">{lang === "bn" ? "বার" : "Day"}</th>
                  <th className="px-3 py-2.5 text-left font-semibold">{lang === "bn" ? "প্রবেশ" : "In Time"}</th>
                  <th className="px-3 py-2.5 text-left font-semibold">{lang === "bn" ? "প্রস্থান" : "Out Time"}</th>
                  <th className="px-3 py-2.5 text-left font-semibold">{lang === "bn" ? "কর্মঘণ্টা" : "Hours"}</th>
                  <th className="px-3 py-2.5 text-left font-semibold">{lang === "bn" ? "ওটি" : "OT"}</th>
                  <th className="px-3 py-2.5 text-left font-semibold">{lang === "bn" ? "বিলম্ব (মি.)" : "Late (min)"}</th>
                  <th className="px-3 py-2.5 text-left font-semibold">{lang === "bn" ? "অবস্থা" : "Status"}</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr><td colSpan={10} className="px-3 py-12 text-center"><Loader2 className="h-6 w-6 animate-spin mx-auto text-muted-foreground" /></td></tr>
                ) : error ? (
                  <tr><td colSpan={10} className="px-3 py-8 text-center text-destructive">{error}</td></tr>
                ) : data.length === 0 ? (
                  <tr><td colSpan={10} className="px-3 py-8 text-center text-muted-foreground">No records found</td></tr>
                ) : (
                  data.map((row, i) => {
                    const bg = row.status === "present" ? "bg-green-50/40"
                      : row.status === "late" ? "bg-orange-50/40"
                      : row.status === "absent" ? "bg-red-50/40"
                      : row.status === "half_day" ? "bg-yellow-50/40"
                      : row.status === "on_leave" ? "bg-indigo-50/40"
                      : row.status === "weekend" ? "bg-slate-50/40" : ""
                    const tc = row.status === "absent" ? "text-red-600"
                      : row.status === "present" ? "text-green-700"
                      : row.status === "late" ? "text-orange-700"
                      : row.status === "half_day" ? "text-yellow-700"
                      : row.status === "on_leave" ? "text-indigo-700"
                      : "text-muted-foreground"
                    const formattedDate = format(new Date(row.date), "dd-MM-yyyy")
                    const dayStr = format(new Date(row.date), "EEE")
                    const displayDay = lang === "bn" ? (dayMapBn[dayStr] || dayStr) : dayStr
                    return (
                      <tr key={row.id} className={`border-b last:border-0 ${bg} hover:bg-muted/50 transition-colors`}>
                        <td className="px-3 py-2">{i + 1}</td>
                        <td className="px-3 py-2">{formattedDate}</td>
                        <td className="px-3 py-2">{row.shift_name || "-"}</td>
                        <td className="px-3 py-2">{displayDay}</td>
                        <td className="px-3 py-2">{formatCheck(row.check_in)}</td>
                        <td className="px-3 py-2">{formatCheck(row.check_out)}</td>
                        <td className="px-3 py-2">{row.total_hours || "-"}</td>
                        <td className="px-3 py-2">{row.over_time || "-"}</td>
                        <td className="px-3 py-2">{row.late_minutes > 0 ? row.late_minutes : "-"}</td>
                        <td className={`px-3 py-2 font-semibold ${tc}`}>
                          {activeStatusMap[row.status] || row.status}
                        </td>
                      </tr>
                    )
                  })
                )}
              </tbody>
            </table>
          </div>

          {data.length > 0 && (
            <div className="border-t px-4 py-3">
              <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm">
                <span className="text-muted-foreground">{lang === "bn" ? "মোট কার্যদিবস" : "Total Days"}: <b>{data.length}</b></span>
                {Object.entries(totalByStatus).map(([s, c]) => (
                  <span key={s} className="text-muted-foreground">
                    {(lang === "bn" ? statusMapBn[s] : s === "on_leave" ? "Leave" : s.charAt(0).toUpperCase() + s.slice(1)) || s}: <b>{c}</b>
                  </span>
                ))}
                <span className="text-muted-foreground">{lang === "bn" ? "মোট বিলম্ব" : "Late Minutes"}: <b>{totalLateMinutes} {lang === "bn" ? "মিনিট" : "min"}</b></span>
                <span className="text-muted-foreground">{lang === "bn" ? "মোট ওভারটাইম" : "Total OT"}: <b>{totalOT ? totalOT.toFixed(2) : 0} {lang === "bn" ? "ঘণ্টা" : "hrs"}</b></span>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

