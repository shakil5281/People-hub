"use client"

import * as React from "react"
import { TrendingUpIcon, Loader2, SearchIcon, PlusIcon } from "lucide-react"
import Link from "next/link"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { salaryIncrementApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi } from "@/lib/api"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }

interface IncrementRecord {
  id: string
  employee_id: string
  previous_gross: number
  increment_amount: number
  new_gross: number
  effective_date: string
  status: string
  remarks: string
  rejection_reason: string
  employee: {
    employee_id: string
    name_en: string
    designation_ref?: { name: string }
    department?: { name: string }
  }
}

const MONTHS = ["January","February","March","April","May","June","July","August","September","October","November","December"]
const currentYear = new Date().getFullYear()
const YEARS = Array.from({length:10},(_,i)=>currentYear-5+i)

const selectCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
const labelCls = "text-xs font-medium text-muted-foreground"

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  approved: "default",
  pending: "secondary",
  rejected: "destructive",
}

export default function IncrementPage() {
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
  const [month, setMonth] = React.useState(0)
  const [year, setYear] = React.useState(0)

  const [data, setData] = React.useState<IncrementRecord[]>([])
  const [loading, setLoading] = React.useState(true)

  const [actionId, setActionId] = React.useState<string | null>(null)

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

  const fetchData = React.useCallback(async () => {
    if (!companyId) return
    setLoading(true)
    try {
      const params: Record<string, string> = { company_id: companyId }
      if (departmentId) params.department_id = departmentId
      if (sectionId) params.section_id = sectionId
      if (designationId) params.designation_id = designationId
      if (lineId) params.line_id = lineId
      if (groupId) params.group_id = groupId
      if (month > 0) params.month = String(month)
      if (year > 0) params.year = String(year)

      const { data: res } = await salaryIncrementApi.list(params)
      setData(Array.isArray(res.increments) ? res.increments : [])
    } catch {
      toast.error("Failed to load increments")
    } finally {
      setLoading(false)
    }
  }, [companyId, departmentId, sectionId, designationId, lineId, groupId, month, year])

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

  React.useEffect(() => { if (companyId) fetchData() }, [companyId, fetchData])

  const handleApprove = async (id: string) => {
    setActionId(id)
    try {
      await salaryIncrementApi.approve(id)
      toast.success("Increment approved")
      fetchData()
    } catch (err: unknown) {
      const msg = typeof err === "object" && err !== null && "response" in err
        ? (err as any).response?.data?.error || "Failed to approve"
        : "Failed to approve"
      toast.error(msg)
    } finally {
      setActionId(null)
    }
  }

  const handleReject = async (id: string) => {
    setActionId(id)
    try {
      await salaryIncrementApi.reject(id)
      toast.success("Increment rejected")
      fetchData()
    } catch (err: unknown) {
      const msg = typeof err === "object" && err !== null && "response" in err
        ? (err as any).response?.data?.error || "Failed to reject"
        : "Failed to reject"
      toast.error(msg)
    } finally {
      setActionId(null)
    }
  }

  const columns: ColumnDef<IncrementRecord>[] = React.useMemo(() => [
    { accessorKey: "employee.name_en", header: "Employee" },
    { accessorKey: "employee_id", header: "Code" },
    { id: "designation", header: "Designation", accessorFn: (r) => r.employee?.designation_ref?.name || "-" },
    { id: "department", header: "Department", accessorFn: (r) => r.employee?.department?.name || "-" },
    {
      accessorKey: "previous_gross",
      header: "Current",
      cell: ({ row }) => row.original.previous_gross.toLocaleString(),
    },
    {
      accessorKey: "increment_amount",
      header: "Increment",
      cell: ({ row }) => <span className="text-green-600 font-semibold">+{row.original.increment_amount.toLocaleString()}</span>,
    },
    {
      accessorKey: "new_gross",
      header: "New Salary",
      cell: ({ row }) => row.original.new_gross.toLocaleString(),
    },
    { accessorKey: "effective_date", header: "Effective Date" },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const s = row.original.status
        return <Badge variant={statusVariant[s] || "secondary"} className="capitalize">{s}</Badge>
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => {
        const r = row.original
        if (r.status !== "pending") return null
        const busy = actionId === r.id
        if (busy) return <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
        return (
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="sm" onClick={() => handleApprove(r.id)} title="Approve" className="text-green-600 hover:text-green-700">
              Approve
            </Button>
            <Button variant="ghost" size="sm" onClick={() => handleReject(r.id)} title="Reject" className="text-red-600 hover:text-red-700">
              Reject
            </Button>
          </div>
        )
      },
    },
  ], [actionId])

  const handleSearch = () => { fetchData() }

  const handleReset = () => {
    setDepartmentId("")
    setSectionId("")
    setDesignationId("")
    setLineId("")
    setGroupId("")
    setMonth(0)
    setYear(0)
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <TrendingUpIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Increment</h1>
            <p className="text-muted-foreground mt-1">Manage salary increments</p>
          </div>
        </div>
        <Link href="/payroll/increment/create">
          <Button>
            <PlusIcon className="mr-2 h-4 w-4" />
            Apply Increment
          </Button>
        </Link>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardContent className="pt-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Company</label>
                <select value={companyId} onChange={e => setCompanyId(e.target.value)} className={selectCls}>
                  <option value="">Select</option>
                  {companies.map(c => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Month</label>
                <select value={month} onChange={e => setMonth(Number(e.target.value))} className={selectCls}>
                  <option value={0}>All</option>
                  {MONTHS.map((n, i) => <option key={n} value={i + 1}>{n}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Year</label>
                <select value={year} onChange={e => setYear(Number(e.target.value))} className={selectCls}>
                  <option value={0}>All</option>
                  {YEARS.map(y => <option key={y} value={y}>{y}</option>)}
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
            <div className="flex gap-2 mt-4">
              <Button onClick={handleSearch} disabled={loading || !companyId}>
                {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <SearchIcon className="mr-2 h-4 w-4" />}
                Search
              </Button>
              <Button variant="outline" onClick={handleReset}>Reset</Button>
            </div>
          </CardContent>
        </Card>
      </div>

      <DataTable data={data} columns={columns} loading={loading} />
    </div>
  )
}
