"use client"

import * as React from "react"
import { format } from "date-fns"
import { DollarSignIcon, SearchIcon, Loader2 } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { DatePicker } from "@/components/ui/date-picker"
import { toast } from "sonner"
import { employeeApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, shiftApi } from "@/lib/api"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

interface EmployeeRecord {
  id: string
  employee_id: string
  name_en: string
  phone: string
  nid: string
  date_of_birth: string
  present_address: string
  permanent_address: string
  present_post_office: string | null
  present_post_code: string | null
  permanent_post_office: string | null
  permanent_post_code: string | null
  present_division?: { name: string }
  present_district?: { name: string }
  permanent_division?: { name: string }
  permanent_district?: { name: string }
}

const selectCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
const labelCls = "text-xs font-medium text-muted-foreground"

export default function MCashPage() {
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])

  const [companyId, setCompanyId] = React.useState("")
  const [departmentId, setDepartmentId] = React.useState("")
  const [sectionId, setSectionId] = React.useState("")
  const [designationId, setDesignationId] = React.useState("")
  const [lineId, setLineId] = React.useState("")
  const [groupId, setGroupId] = React.useState("")
  const [shiftId, setShiftId] = React.useState("")

  const [joiningFrom, setJoiningFrom] = React.useState<Date | undefined>()
  const [joiningTo, setJoiningTo] = React.useState<Date | undefined>()

  const [data, setData] = React.useState<EmployeeRecord[]>([])
  const [loading, setLoading] = React.useState(false)

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

  React.useEffect(() => {
    const init = async () => {
      const [cRes, dRes, gRes, sRes] = await Promise.all([
        companyApi.list({ limit: "100" }),
        departmentApi.list({ limit: "100" }),
        groupApi.list({ limit: "100" }),
        shiftApi.list({ limit: "100" }),
      ])
      const clist = Array.isArray(cRes.data?.data) ? cRes.data.data : (Array.isArray(cRes.data) ? cRes.data : [])
      if (clist.length > 0) { setCompanies(clist); setCompanyId(clist[0].id) }
      if (Array.isArray(dRes.data?.data)) setDepartments(dRes.data.data)
      if (Array.isArray(gRes.data?.data)) setGroups(gRes.data.data)
      else if (Array.isArray(gRes.data)) setGroups(gRes.data)
      if (Array.isArray(sRes.data?.data)) setShifts(sRes.data.data)
      else if (Array.isArray(sRes.data)) setShifts(sRes.data)
    }
    init()
  }, [])

  const fetchData = React.useCallback(async () => {
    if (!companyId) return
    setLoading(true)
    try {
      const params: Record<string, string> = { company_id: companyId, limit: "1000" }
      if (departmentId) params.department_id = departmentId
      if (sectionId) params.section_id = sectionId
      if (designationId) params.designation_id = designationId
      if (lineId) params.line_id = lineId
      if (groupId) params.group_id = groupId
      if (shiftId) params.shift_id = shiftId
      if (joiningFrom) params.joining_from = format(joiningFrom, "yyyy-MM-dd")
      if (joiningTo) params.joining_to = format(joiningTo, "yyyy-MM-dd")

      const { data: res } = await employeeApi.list(params)
      const list = Array.isArray(res.data) ? res.data : (Array.isArray(res) ? res : [])
      setData(list)
    } catch {
      toast.error("Failed to load employees")
    } finally {
      setLoading(false)
    }
  }, [companyId, departmentId, sectionId, designationId, lineId, groupId, shiftId, joiningFrom, joiningTo])

  const handleSearch = () => { fetchData() }

  const handleReset = () => {
    setDepartmentId("")
    setSectionId("")
    setDesignationId("")
    setLineId("")
    setGroupId("")
    setShiftId("")
    setJoiningFrom(undefined)
    setJoiningTo(undefined)
    setData([])
  }

  const columns: ColumnDef<EmployeeRecord>[] = React.useMemo(() => [
    {
      id: "sl",
      header: "SL",
      cell: ({ row }) => row.index + 1,
    },
    { accessorKey: "name_en", header: "Employee Name" },
    { accessorKey: "phone", header: "Mobile Number" },
    { accessorKey: "nid", header: "NID No" },
    {
      accessorKey: "date_of_birth",
      header: "Date of Birth",
      cell: ({ row }) => row.original.date_of_birth || "-",
    },
    {
      id: "present_division",
      header: "Present Division",
      accessorFn: (r) => r.present_division?.name || "-",
    },
    {
      id: "present_district",
      header: "Present District",
      accessorFn: (r) => r.present_district?.name || "-",
    },
    {
      accessorKey: "present_post_office",
      header: "Present Post Office",
      cell: ({ row }) => row.original.present_post_office || "-",
    },
    {
      accessorKey: "present_post_code",
      header: "Present Post Code",
      cell: ({ row }) => row.original.present_post_code || "-",
    },
    {
      accessorKey: "present_address",
      header: "Present Address",
      cell: ({ row }) => row.original.present_address || "-",
    },
    {
      id: "permanent_division",
      header: "Permanent Division",
      accessorFn: (r) => r.permanent_division?.name || "-",
    },
    {
      id: "permanent_district",
      header: "Permanent District",
      accessorFn: (r) => r.permanent_district?.name || "-",
    },
    {
      accessorKey: "permanent_post_office",
      header: "Permanent Post Office",
      cell: ({ row }) => row.original.permanent_post_office || "-",
    },
    {
      accessorKey: "permanent_post_code",
      header: "Permanent Post Code",
      cell: ({ row }) => row.original.permanent_post_code || "-",
    },
    {
      accessorKey: "permanent_address",
      header: "Permanent Address",
      cell: ({ row }) => row.original.permanent_address || "-",
    },
  ], [])

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <DollarSignIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">mCash</h1>
            <p className="text-muted-foreground mt-1">Employee information for mCash</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardContent className="pt-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Joining Date From</label>
                <DatePicker value={joiningFrom} onChange={setJoiningFrom} placeholder="From" />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Joining Date To</label>
                <DatePicker value={joiningTo} onChange={setJoiningTo} placeholder="To" />
              </div>
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
              <div className="flex flex-col gap-1.5">
                <label className={labelCls}>Shift</label>
                <select value={shiftId} onChange={e => setShiftId(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  {shifts.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
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
