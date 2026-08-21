"use client"

import * as React from "react"
import { TrendingUpIcon, Loader2, SearchIcon, PlusIcon, TrashIcon, CheckIcon, XIcon, FileSpreadsheetIcon, FileTextIcon } from "lucide-react"
import Link from "next/link"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { salaryIncrementApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi } from "@/lib/api"
import { format } from "date-fns"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"

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
  increment_date: string
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
  const [month, setMonth] = React.useState(new Date().getMonth() + 1)
  const [year, setYear] = React.useState(new Date().getFullYear())
  const [status, setStatus] = React.useState("")

  const [data, setData] = React.useState<IncrementRecord[]>([])
  const [loading, setLoading] = React.useState(true)

  const [selectedRows, setSelectedRows] = React.useState<IncrementRecord[]>([])
  const [bulkActioning, setBulkActioning] = React.useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false)

  const [exportingExcel, setExportingExcel] = React.useState(false)
  const [exportingPdf, setExportingPdf] = React.useState(false)

  const buildExportParams = () => {
    const params: Record<string, string> = { company_id: companyId }
    if (departmentId) params.department_id = departmentId
    if (sectionId) params.section_id = sectionId
    if (designationId) params.designation_id = designationId
    if (lineId) params.line_id = lineId
    if (groupId) params.group_id = groupId
    if (month > 0) params.month = String(month)
    if (year > 0) params.year = String(year)
    if (status) params.status = status
    return params
  }

  const handleExportExcel = async () => {
    if (!companyId) return toast.error("Please select a company")
    setExportingExcel(true)
    try {
      const res = await salaryIncrementApi.exportExcel(buildExportParams())
      const url = window.URL.createObjectURL(new Blob([res.data]))
      const a = document.createElement("a")
      a.href = url
      a.download = `increments_${format(new Date(), "yyyyMMdd_HHmmss")}.xlsx`
      document.body.appendChild(a)
      a.click()
      a.remove()
      toast.success("Excel exported successfully")
    } catch {
      toast.error("Failed to export Excel")
    } finally {
      setExportingExcel(false)
    }
  }

  const handleExportPdf = async () => {
    if (!companyId) return toast.error("Please select a company")
    setExportingPdf(true)
    try {
      const res = await salaryIncrementApi.exportPdf(buildExportParams())
      const url = window.URL.createObjectURL(new Blob([res.data], { type: 'application/pdf' }))
      const a = document.createElement("a")
      a.href = url
      a.download = `increments_${format(new Date(), "yyyyMMdd_HHmmss")}.pdf`
      document.body.appendChild(a)
      a.click()
      a.remove()
      toast.success("PDF exported successfully")
    } catch {
      toast.error("Failed to export PDF")
    } finally {
      setExportingPdf(false)
    }
  }

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

  const fetchData = React.useCallback(async () => {
    if (!companyId) return
    setLoading(true)
    try {
      const params = buildExportParams()

      const { data: res } = await salaryIncrementApi.list(params)
      setData(Array.isArray(res.increments) ? res.increments : [])
      setSelectedRows([])
    } catch {
      toast.error("Failed to load increments")
    } finally {
      setLoading(false)
    }
  }, [companyId, departmentId, sectionId, designationId, lineId, groupId, month, year, status])

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

  const handleBulkApprove = async () => {
    if (selectedRows.length === 0) return
    const invalidRows = selectedRows.filter(r => r.status !== 'pending')
    if (invalidRows.length > 0) {
      toast.error("Only pending increments can be approved. Please unselect others.")
      return
    }
    setBulkActioning(true)
    try {
      await Promise.all(selectedRows.map(r => salaryIncrementApi.approve(r.id)))
      toast.success("Approved selected increments")
      fetchData()
    } catch {
      toast.error("Failed to approve some increments")
    } finally {
      setBulkActioning(false)
    }
  }

  const handleBulkReject = async () => {
    if (selectedRows.length === 0) return
    const invalidRows = selectedRows.filter(r => r.status !== 'pending')
    if (invalidRows.length > 0) {
      toast.error("Only pending increments can be rejected. Please unselect others.")
      return
    }
    setBulkActioning(true)
    try {
      await Promise.all(selectedRows.map(r => salaryIncrementApi.reject(r.id, { reason: "Bulk rejected" })))
      toast.success("Rejected selected increments")
      fetchData()
    } catch {
      toast.error("Failed to reject some increments")
    } finally {
      setBulkActioning(false)
    }
  }

  const handleBulkDelete = async () => {
    if (selectedRows.length === 0) return
    const invalidRows = selectedRows.filter(r => r.status !== 'pending')
    if (invalidRows.length > 0) {
      toast.error("Only pending increments can be deleted. Please unselect approved/rejected ones.")
      setDeleteDialogOpen(false)
      return
    }
    setBulkActioning(true)
    try {
      await Promise.all(selectedRows.map(r => salaryIncrementApi.delete(r.id)))
      toast.success("Deleted selected increments")
      fetchData()
    } catch {
      toast.error("Failed to delete some increments. Did you restart the server?")
    } finally {
      setBulkActioning(false)
      setDeleteDialogOpen(false)
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
    {
      accessorKey: "increment_date",
      header: "Increment Date",
      cell: ({ row }) => row.original.increment_date ? format(new Date(row.original.increment_date), "dd-MM-yyyy") : "-",
    },
    {
      accessorKey: "effective_date",
      header: "Effective Date",
      cell: ({ row }) => row.original.effective_date ? format(new Date(row.original.effective_date), "dd-MM-yyyy") : "-",
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const s = row.original.status
        return <Badge variant={statusVariant[s] || "secondary"} className="capitalize">{s}</Badge>
      },
    },
  ], [])

  const handleSearch = () => { fetchData() }

  const handleReset = () => {
    setDepartmentId("")
    setSectionId("")
    setDesignationId("")
    setLineId("")
    setGroupId("")
    setMonth(new Date().getMonth() + 1)
    setYear(new Date().getFullYear())
    setStatus("")
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
        <div className="flex items-center gap-2">
          <ButtonGroup>
            <Button onClick={handleExportExcel} disabled={exportingExcel} variant="outline" className="h-10">
              {exportingExcel ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileSpreadsheetIcon className="mr-2 h-4 w-4" />}
              Export Excel
            </Button>
            <Button onClick={handleExportPdf} disabled={exportingPdf} variant="outline" className="h-10">
              {exportingPdf ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileTextIcon className="mr-2 h-4 w-4" />}
              Export PDF
            </Button>
          </ButtonGroup>
          <Link href="/payroll/increment/create">
            <Button className="h-10">
              <PlusIcon className="mr-2 h-4 w-4" />
              Apply Increment
            </Button>
          </Link>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardContent className="pt-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
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
                <label className={labelCls}>Status</label>
                <select value={status} onChange={e => setStatus(e.target.value)} className={selectCls}>
                  <option value="">All</option>
                  <option value="pending">Pending</option>
                  <option value="approved">Approved</option>
                  <option value="rejected">Rejected</option>
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
            </div>
            <div className="mt-4 flex items-center justify-end gap-2">
              <Button variant="outline" onClick={handleReset} className="h-9">Reset</Button>
              <Button onClick={handleSearch} disabled={loading} className="h-9">
                {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <SearchIcon className="mr-2 h-4 w-4" />}
                Search
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="px-4 lg:px-6">
        <div className="mb-4 flex items-center gap-2">
          {selectedRows.length > 0 && (
            <>
              <Button onClick={handleBulkApprove} disabled={bulkActioning} variant="default" className="h-9 bg-green-600 hover:bg-green-700">
                {bulkActioning ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <CheckIcon className="mr-2 h-4 w-4" />}
                Approve ({selectedRows.length})
              </Button>
              <Button onClick={handleBulkReject} disabled={bulkActioning} variant="default" className="h-9 bg-orange-600 hover:bg-orange-700">
                {bulkActioning ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <XIcon className="mr-2 h-4 w-4" />}
                Reject ({selectedRows.length})
              </Button>
              <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
                <AlertDialogTrigger asChild>
                  <Button disabled={bulkActioning} variant="destructive" className="h-9">
                    {bulkActioning ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <TrashIcon className="mr-2 h-4 w-4" />}
                    Delete ({selectedRows.length})
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                    <AlertDialogDescription>
                      This action cannot be undone. This will permanently delete {selectedRows.length} selected increment(s). Only pending increments can be deleted.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction onClick={(e) => {
                      e.preventDefault()
                      handleBulkDelete()
                    }} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                      Delete
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </>
          )}
        </div>
        <DataTable
          columns={columns}
          data={data}
          loading={loading}
          enableSelection={true}
          onSelectionChange={setSelectedRows}
        />
      </div>
    </div>
  )
}
