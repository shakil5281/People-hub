"use client"

import * as React from "react"
import { TrendingUpIcon, ArrowLeftIcon, Loader2 } from "lucide-react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { salaryIncrementApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi } from "@/lib/api"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }

const selectCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
const labelCls = "text-xs font-medium text-muted-foreground"

export default function CreateIncrementPage() {
  const router = useRouter()

  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])

  const [companyId, setCompanyId] = React.useState("")
  const [departmentId, setDepartmentId] = React.useState("")
  const [sectionId, setSectionId] = React.useState("")
  const [designationId, setDesignationId] = React.useState("")
  const [lineId, setLineId] = React.useState("")
  const [groupId, setGroupId] = React.useState("")

  const [incrementType, setIncrementType] = React.useState("fixed")
  const [value, setValue] = React.useState("")
  const [incrementDate, setIncrementDate] = React.useState("")
  const [effectiveDate, setEffectiveDate] = React.useState("")

  const [applying, setApplying] = React.useState(false)

  React.useEffect(() => {
    const init = async () => {
      const [cRes, dRes, gRes] = await Promise.all([
        companyApi.list({ limit: "100" }),
        departmentApi.list({ limit: "100" }),
        groupApi.list({ limit: "100" }),
      ])
      const clist = Array.isArray(cRes.data?.data) ? cRes.data.data : (Array.isArray(cRes.data) ? cRes.data : [])
      if (clist.length > 0) { setCompanies(clist); setCompanyId(clist[0].id) }
      if (Array.isArray(dRes.data?.data)) setDepartments(dRes.data.data)
      if (Array.isArray(gRes.data?.data)) setGroups(gRes.data.data)
      else if (Array.isArray(gRes.data)) setGroups(gRes.data)
    }
    init()
  }, [])

  const fetchSections = React.useCallback(async (deptId: string) => {
    try {
      const { data } = await sectionApi.list(deptId || undefined)
      setSections(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setSections([]) }
  }, [])

  const fetchDesignations = React.useCallback(async (secId: string) => {
    try {
      const { data } = await designationApi.list(secId || undefined)
      setDesignations(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setDesignations([]) }
  }, [])

  const fetchLines = React.useCallback(async (secId: string) => {
    try {
      const { data } = await lineApi.list(secId || undefined)
      setLines(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setLines([]) }
  }, [])

  React.useEffect(() => {
    fetchSections(departmentId)
    setSectionId("")
    setDesignationId("")
    setLineId("")
    setDesignations([])
    setLines([])
  }, [departmentId, fetchSections])

  React.useEffect(() => {
    fetchDesignations(sectionId)
    fetchLines(sectionId)
    setDesignationId("")
    setLineId("")
  }, [sectionId, fetchDesignations, fetchLines])

  const handleApply = async () => {
    if (!value || !incrementDate || !effectiveDate) {
      toast.error("Value, increment date, and effective date are required")
      return
    }
    setApplying(true)
    try {
      const payload: Record<string, unknown> = {
        company_id: companyId,
        increment_type: incrementType,
        increment_date: incrementDate,
        effective_date: effectiveDate,
        value: Number(value),
      }
      if (departmentId) payload.department_id = departmentId
      if (sectionId) payload.section_id = sectionId
      if (designationId) payload.designation_id = designationId
      if (lineId) payload.line_id = lineId
      if (groupId) payload.group_id = groupId

      const { data: res } = await salaryIncrementApi.bulkApply(payload)
      toast.success(`${res.message} — ${res.applied} of ${res.total} eligible employees`)
      router.push("/payroll/increment")
    } catch (err: unknown) {
      const msg = typeof err === "object" && err !== null && "response" in err
        ? (err as any).response?.data?.error || "Failed to apply increment"
        : "Failed to apply increment"
      toast.error(msg)
    } finally {
      setApplying(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <TrendingUpIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <div className="flex items-center gap-3">
              <Button variant="ghost" size="icon" onClick={() => router.back()} className="h-8 w-8">
                <ArrowLeftIcon className="h-4 w-4" />
              </Button>
              <h1 className="text-3xl font-bold tracking-tight">Apply Bulk Increment</h1>
            </div>
            <p className="text-muted-foreground mt-1 ml-11">Apply salary increment to eligible employees</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6 max-w-3xl">
        <div className="rounded-lg border bg-card">
          <div className="p-6 space-y-6">
            <h2 className="text-lg font-semibold">Filter Employees</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Company</label>
                <select value={companyId} onChange={e => setCompanyId(e.target.value)} className={selectCls}>
                  <option value="">Select</option>
                  {companies.map(c => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Department</label>
                <select value={departmentId} onChange={e => setDepartmentId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Section</label>
                <select value={sectionId} onChange={e => setSectionId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {sections.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Designation</label>
                <select value={designationId} onChange={e => setDesignationId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {designations.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Line</label>
                <select value={lineId} onChange={e => setLineId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {lines.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Group</label>
                <select value={groupId} onChange={e => setGroupId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {groups.map(g => <option key={g.id} value={g.id}>{g.name}</option>)}
                </select>
              </div>
            </div>
          </div>

          <div className="border-t p-6 space-y-6">
            <h2 className="text-lg font-semibold">Increment Details</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Increment Type</label>
                <select
                  value={incrementType}
                  onChange={(e) => setIncrementType(e.target.value)}
                  className={selectCls}
                >
                  <option value="fixed">Fixed Amount</option>
                  <option value="percentage">Percentage (%)</option>
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>{incrementType === "percentage" ? "Percentage (%)" : "Amount (BDT)"}</label>
                <input
                  type="number"
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  placeholder={incrementType === "percentage" ? "e.g. 10" : "e.g. 3000"}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Increment Date</label>
                <input
                  type="date"
                  value={incrementDate}
                  onChange={(e) => setIncrementDate(e.target.value)}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Effective Date</label>
                <input
                  type="date"
                  value={effectiveDate}
                  onChange={(e) => setEffectiveDate(e.target.value)}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                />
              </div>
            </div>
          </div>

          <div className="border-t p-6 flex items-center justify-end gap-2">
            <Button variant="outline" onClick={() => router.back()}>Cancel</Button>
            <Button onClick={handleApply} disabled={applying}>
              {applying ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              {applying ? "Applying..." : "Apply Increment"}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
