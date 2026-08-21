"use client"

import * as React from "react"
import { TrendingUpIcon, ArrowLeftIcon, Loader2, SearchIcon, AlertCircle } from "lucide-react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { getApiBaseUrl } from "@/lib/utils"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { salaryIncrementApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, employeeApi } from "@/lib/api"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { DatePicker } from "@/components/ui/date-picker"
import { format } from "date-fns"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }

interface EmployeeRow {
  id: string
  employee_id: string
  name_en: string
  designation?: string
  department?: string
  gross_salary: number
  basic_salary: number
  image_url?: string
}

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

  const [employees, setEmployees] = React.useState<EmployeeRow[]>([])
  const [loading, setLoading] = React.useState(false)
  const [selectedRows, setSelectedRows] = React.useState<EmployeeRow[]>([])

  const [incrementType, setIncrementType] = React.useState("fixed")
  const [value, setValue] = React.useState("")
  const [incrementDate, setIncrementDate] = React.useState<Date | undefined>()
  const [effectiveDate, setEffectiveDate] = React.useState<Date | undefined>()
  const [applying, setApplying] = React.useState(false)
  const [dialogOpen, setDialogOpen] = React.useState(false)

  React.useEffect(() => {
    const init = async () => {
      const [cRes, dRes, gRes] = await Promise.all([
        companyApi.list({ limit: "100" }),
        departmentApi.list({ limit: "100" }),
        groupApi.list({ limit: "100" }),
      ])
      if (Array.isArray(cRes.data?.data)) setCompanies(cRes.data.data)
      else if (Array.isArray(cRes.data)) setCompanies(cRes.data)
      
      if (Array.isArray(dRes.data?.data)) setDepartments(dRes.data.data)
      else if (Array.isArray(dRes.data)) setDepartments(dRes.data)
      
      if (Array.isArray(gRes.data?.data)) setGroups(gRes.data.data)
      else if (Array.isArray(gRes.data)) setGroups(gRes.data)
    }
    init()
  }, [])

  const fetchSections = React.useCallback(async (deptId: string) => {
    try {
      const { data } = await sectionApi.list(deptId || undefined, { limit: "100" })
      setSections(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setSections([]) }
  }, [])

  const fetchDesignations = React.useCallback(async (secId: string) => {
    try {
      const { data } = await designationApi.list(secId || undefined, { limit: "100" })
      setDesignations(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setDesignations([]) }
  }, [])

  const fetchLines = React.useCallback(async (secId: string) => {
    try {
      const { data } = await lineApi.list(secId || undefined, { limit: "100" })
      setLines(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    } catch { setLines([]) }
  }, [])

  React.useEffect(() => {
    if (!departmentId) {
      setSections([])
      setDesignations([])
      setLines([])
      setSectionId("")
      setDesignationId("")
      setLineId("")
      return
    }
    fetchSections(departmentId)
    setSectionId("")
    setDesignationId("")
    setLineId("")
    setDesignations([])
    setLines([])
  }, [departmentId, fetchSections])

  React.useEffect(() => {
    if (!sectionId) {
      setDesignations([])
      setLines([])
      setDesignationId("")
      setLineId("")
      return
    }
    fetchDesignations(sectionId)
    fetchLines(sectionId)
    setDesignationId("")
    setLineId("")
  }, [sectionId, fetchDesignations, fetchLines])

  const handleSearch = async () => {
    if (!companyId) {
      toast.error("Company is required to search employees")
      return
    }
    setLoading(true)
    try {
      const params: Record<string, string> = { company_id: companyId, limit: "1000" }
      if (departmentId) params.department_id = departmentId
      if (sectionId) params.section_id = sectionId
      if (designationId) params.designation_id = designationId
      if (lineId) params.line_id = lineId
      if (groupId) params.group_id = groupId

      const res = await employeeApi.list(params)
      const data = res.data?.data || res.data || []
      setEmployees(Array.isArray(data) ? data : [])
      setSelectedRows([])
    } catch (err: unknown) {
      toast.error("Failed to fetch employees")
    } finally {
      setLoading(false)
    }
  }

  const handleApply = async () => {
    if (!value || !incrementDate || !effectiveDate) {
      toast.error("Value, increment date, and effective date are required")
      return
    }
    if (selectedRows.length === 0) {
      toast.error("Please select at least one employee")
      return
    }

    setApplying(true)
    try {
      const payload: Record<string, unknown> = {
        company_id: companyId,
        increment_type: incrementType,
        increment_date: format(incrementDate, "yyyy-MM-dd"),
        effective_date: format(effectiveDate, "yyyy-MM-dd"),
        value: Number(value),
        employee_ids: selectedRows.map(r => r.employee_id)
      }
      if (departmentId) payload.department_id = departmentId
      if (sectionId) payload.section_id = sectionId
      if (designationId) payload.designation_id = designationId
      if (lineId) payload.line_id = lineId
      if (groupId) payload.group_id = groupId

      const { data: res } = await salaryIncrementApi.bulkApply(payload)
      toast.success(`${res.message} - ${res.applied} employees`)
      setDialogOpen(false)
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

  const columns: ColumnDef<EmployeeRow>[] = [
    {
      accessorKey: "photo",
      header: "Photo",
      cell: ({ row }) => {
        const img = row.original.image_url
        const baseUrl = getApiBaseUrl().replace("/api/v1", "")
        return (
          <Avatar className="h-8 w-8">
            <AvatarImage src={img ? (img.startsWith("http") ? img : `${baseUrl}/${img.replace(/^\//, '')}`) : ""} />
            <AvatarFallback>{row.original.name_en?.charAt(0)}</AvatarFallback>
          </Avatar>
        )
      },
    },
    { accessorKey: "employee_id", header: "Emp ID" },
    { accessorKey: "name_en", header: "Name" },
    { accessorKey: "designation", header: "Designation", cell: ({ row }) => row.original.designation || "-" },
    { accessorKey: "department", header: "Department", cell: ({ row }) => row.original.department || "-" },
    { accessorKey: "gross_salary", header: "Gross Salary" },
  ]

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
            <p className="text-muted-foreground mt-1 ml-11">Filter and apply increment</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6 flex flex-col gap-6">
        {/* Filter Section */}
        <div className="rounded-lg border bg-card p-6 space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Advance Filter</h2>
            <Button onClick={handleSearch} disabled={loading || !companyId}>
              {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <SearchIcon className="mr-2 h-4 w-4" />}
              Search Employees
            </Button>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Company *</label>
              <select value={companyId} onChange={e => setCompanyId(e.target.value)} className={selectCls}>
                <option value="">Select Company</option>
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

        {/* Data Table Section */}
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Employee List</h2>
            {selectedRows.length > 0 && (
              <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                <DialogTrigger asChild>
                  <Button variant="default">
                    Apply Increment to {selectedRows.length} Employees
                  </Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Apply Increment</DialogTitle>
                    <DialogDescription>
                      Fill in the increment details for the selected {selectedRows.length} employees.
                    </DialogDescription>
                  </DialogHeader>
                  
                  <div className="grid gap-4 py-4">
                    <div className="grid grid-cols-2 gap-4">
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
                          className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
                        />
                      </div>
                      <div className="flex flex-col gap-1.5">
                        <label className={labelCls}>Increment Date</label>
                        <DatePicker
                          value={incrementDate}
                          onChange={setIncrementDate}
                          placeholder="Select date"
                        />
                      </div>
                      <div className="flex flex-col gap-1.5">
                        <label className={labelCls}>Effective Date</label>
                        <DatePicker
                          value={effectiveDate}
                          onChange={setEffectiveDate}
                          placeholder="Select date"
                        />
                      </div>
                    </div>
                  </div>

                  <DialogFooter>
                    <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
                    <Button onClick={handleApply} disabled={applying}>
                      {applying && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                      Submit Increment
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            )}
          </div>
          
          {employees.length === 0 && !loading ? (
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                No employees found. Please adjust filters and search.
              </AlertDescription>
            </Alert>
          ) : (
            <DataTable
              data={employees}
              columns={columns}
              loading={loading}
              enableSelection={true}
              onSelectionChange={setSelectedRows}
            />
          )}
        </div>
      </div>
    </div>
  )
}
