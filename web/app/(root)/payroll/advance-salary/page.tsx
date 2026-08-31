"use client"

import * as React from "react"
import { BanknoteIcon, Loader2, SearchIcon, PlusIcon, TrashIcon, CheckIcon, XIcon, FileSpreadsheetIcon, FileTextIcon } from "lucide-react"
import Link from "next/link"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { advanceSalaryApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi } from "@/lib/api"
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

interface AdvanceRecord {
  id: string
  employee_id: string
  amount: number
  advance_date: string
  deduction_month: number
  deduction_year: number
  status: string
  reason: string
  employee: {
    employee_id: string
    name_en: string
    designation_ref?: { name: string }
    department?: { name: string }
  }
}

export default function AdvanceSalaryPage() {
  const [data, setData] = React.useState<AdvanceRecord[]>([])
  const [loading, setLoading] = React.useState(true)
  const [isExportingExcel, setExportingExcel] = React.useState(false)
  const [isExportingPdf, setExportingPdf] = React.useState(false)
  const [bulkActionLoading, setBulkActionLoading] = React.useState(false)

  const [selectedRows, setSelectedRows] = React.useState<AdvanceRecord[]>([])
  const [deleteId, setDeleteId] = React.useState<string | null>(null)

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
  const [status, setStatus] = React.useState("all")
  
  const currentDate = new Date()
  const [month, setMonth] = React.useState(currentDate.getMonth() + 1)
  const [year, setYear] = React.useState(currentDate.getFullYear())

  React.useEffect(() => {
    const init = async () => {
      try {
        const [comp, dept, grp] = await Promise.all([
          companyApi.list(),
          departmentApi.list({ limit: "100" }),
          groupApi.list({ limit: "100" })
        ])
        setCompanies(Array.isArray(comp.data?.data) ? comp.data.data : Array.isArray(comp.data) ? comp.data : [])
        setDepartments(Array.isArray(dept.data?.data) ? dept.data.data : Array.isArray(dept.data) ? dept.data : [])
        setGroups(Array.isArray(grp.data?.data) ? grp.data.data : Array.isArray(grp.data) ? grp.data : [])
      } catch (err) {
        console.error("Failed to load initial dropdowns", err)
      }
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

  const loadData = React.useCallback(async () => {
    if (!companyId) {
      setData([])
      setLoading(false)
      return
    }
    setLoading(true)
    try {
      const params: Record<string, string> = { company_id: companyId, month: String(month), year: String(year), limit: "5000" }
      if (departmentId) params.department_id = departmentId
      if (sectionId) params.section_id = sectionId
      if (designationId) params.designation_id = designationId
      if (lineId) params.line_id = lineId
      if (groupId) params.group_id = groupId
      if (status !== "all") params.status = status

      const res = await advanceSalaryApi.list(params)
      setData(res.data?.advances || [])
    } catch (err: any) {
      toast.error(err.response?.data?.error || "Failed to load advances")
    } finally {
      setLoading(false)
    }
  }, [companyId, departmentId, sectionId, designationId, lineId, groupId, month, year, status])

  React.useEffect(() => {
    loadData()
  }, [loadData])

  const exportExcel = async () => {
    if (!companyId) return toast.error("Select a company first")
    setExportingExcel(true)
    try {
      const params: Record<string, string> = { company_id: companyId, month: String(month), year: String(year), limit: "5000" }
      if (departmentId) params.department_id = departmentId
      if (sectionId) params.section_id = sectionId
      if (designationId) params.designation_id = designationId
      if (lineId) params.line_id = lineId
      if (groupId) params.group_id = groupId
      if (status !== "all") params.status = status

      const res = await advanceSalaryApi.exportExcel(params)
      const url = window.URL.createObjectURL(new Blob([res.data]))
      const link = document.createElement("a")
      link.href = url
      link.setAttribute("download", `advance_report_${format(new Date(), "yyyyMMdd_HHmmss")}.xlsx`)
      document.body.appendChild(link)
      link.click()
      link.remove()
    } catch {
      toast.error("Failed to export Excel")
    } finally {
      setExportingExcel(false)
    }
  }

  const exportPdf = async () => {
    if (!companyId) return toast.error("Select a company first")
    setExportingPdf(true)
    try {
      const params: Record<string, string> = { company_id: companyId, month: String(month), year: String(year), limit: "5000" }
      if (departmentId) params.department_id = departmentId
      if (sectionId) params.section_id = sectionId
      if (designationId) params.designation_id = designationId
      if (lineId) params.line_id = lineId
      if (groupId) params.group_id = groupId
      if (status !== "all") params.status = status

      const res = await advanceSalaryApi.exportPdf(params)
      const url = window.URL.createObjectURL(new Blob([res.data], { type: "application/pdf" }))
      const link = document.createElement("a")
      link.href = url
      link.setAttribute("download", `advance_report_${format(new Date(), "yyyyMMdd_HHmmss")}.pdf`)
      document.body.appendChild(link)
      link.click()
      link.remove()
    } catch {
      toast.error("Failed to export PDF")
    } finally {
      setExportingPdf(false)
    }
  }

  const handleDelete = async () => {
    if (!deleteId) return
    try {
      await advanceSalaryApi.delete(deleteId)
      toast.success("Advance deleted successfully")
      setDeleteId(null)
      loadData()
    } catch (err: any) {
      toast.error(err.response?.data?.error || "Failed to delete")
    }
  }

  const handleBulkApprove = async () => {
    if (!selectedRows.length) return toast.error("No rows selected")
    const pending = selectedRows.filter(r => r.status === "pending")
    if (!pending.length) return toast.error("No pending advances selected")
    
    setBulkActionLoading(true)
    let count = 0
    for (const row of pending) {
      try {
        await advanceSalaryApi.approve(row.id)
        count++
      } catch (err) {
        console.error("Failed to approve", row.id)
      }
    }
    toast.success(`Approved ${count} advances`)
    setSelectedRows([])
    setBulkActionLoading(false)
    loadData()
  }

  const handleBulkReject = async () => {
    if (!selectedRows.length) return toast.error("No rows selected")
    const pending = selectedRows.filter(r => r.status === "pending")
    if (!pending.length) return toast.error("No pending advances selected")
    
    setBulkActionLoading(true)
    let count = 0
    for (const row of pending) {
      try {
        await advanceSalaryApi.reject(row.id)
        count++
      } catch (err) {
        console.error("Failed to reject", row.id)
      }
    }
    toast.success(`Rejected ${count} advances`)
    setSelectedRows([])
    setBulkActionLoading(false)
    loadData()
  }

  const handleBulkDelete = async () => {
    if (!selectedRows.length) return toast.error("No rows selected")
    const pending = selectedRows.filter(r => r.status === "pending")
    if (!pending.length) return toast.error("No pending advances selected")
    
    if (!confirm(`Are you sure you want to delete ${pending.length} pending advances?`)) return

    setBulkActionLoading(true)
    let count = 0
    for (const row of pending) {
      try {
        await advanceSalaryApi.delete(row.id)
        count++
      } catch (err) {
        console.error("Failed to delete", row.id)
      }
    }
    toast.success(`Deleted ${count} advances`)
    setSelectedRows([])
    setBulkActionLoading(false)
    loadData()
  }

  const columns: ColumnDef<AdvanceRecord>[] = [
    { accessorKey: "employee.employee_id", header: "Emp ID" },
    { accessorKey: "employee.name_en", header: "Name" },
    { accessorKey: "employee.designation_ref.name", header: "Designation" },
    { accessorKey: "employee.department.name", header: "Department" },
    {
      accessorKey: "amount",
      header: "Amount",
      cell: ({ row }) => <span className="font-medium text-emerald-600 dark:text-emerald-400">{row.original.amount.toFixed(2)}</span>
    },
    {
      accessorKey: "advance_date",
      header: "Advance Date",
      cell: ({ row }) => format(new Date(row.original.advance_date), "dd-MM-yyyy")
    },
    {
      accessorKey: "deduction_month",
      header: "Target Month",
      cell: ({ row }) => <span>{row.original.deduction_month}/{row.original.deduction_year}</span>
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const s = row.original.status
        if (s === "approved" || s === "deducted") return <Badge className="bg-emerald-500 hover:bg-emerald-600">{s.toUpperCase()}</Badge>
        if (s === "rejected") return <Badge variant="destructive">REJECTED</Badge>
        return <Badge variant="secondary" className="bg-amber-500 text-white hover:bg-amber-600">PENDING</Badge>
      }
    },
    {
      id: "actions",
      header: "Actions",
      cell: ({ row }) => {
        const record = row.original
        if (record.status !== "pending") return null
        return (
          <Button variant="ghost" size="icon" className="text-red-500 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950/50" onClick={() => setDeleteId(record.id)}>
            <TrashIcon className="h-4 w-4" />
          </Button>
        )
      },
    },
  ]

  const selectCls = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
  const labelCls = "text-xs font-semibold text-muted-foreground uppercase tracking-wider"

  return (
    <div className="flex flex-col gap-6 p-4 lg:p-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-2 text-primary">
          <div className="rounded-lg bg-primary/10 p-2">
            <BanknoteIcon className="h-6 w-6" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight">Advance Salary</h1>
            <p className="text-sm text-muted-foreground">Manage and review employee salary advances</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <ButtonGroup>
            <Button variant="outline" size="sm" onClick={exportExcel} disabled={isExportingExcel || !companyId}>
              {isExportingExcel ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileSpreadsheetIcon className="mr-2 h-4 w-4" />}
              Excel
            </Button>
            <Button variant="outline" size="sm" onClick={exportPdf} disabled={isExportingPdf || !companyId}>
              {isExportingPdf ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileTextIcon className="mr-2 h-4 w-4" />}
              PDF
            </Button>
          </ButtonGroup>
          <Button asChild>
            <Link href="/payroll/advance-salary/create">
              <PlusIcon className="mr-2 h-4 w-4" />
              Apply Advance
            </Link>
          </Button>
        </div>
      </div>

      <Card className="border-emerald-100 dark:border-emerald-900/50 shadow-sm">
        <CardContent className="p-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Company</label>
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
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Month</label>
              <select value={month} onChange={e => setMonth(Number(e.target.value))} className={selectCls}>
                {Array.from({ length: 12 }, (_, i) => i + 1).map(m => (
                  <option key={m} value={m}>{new Date(2000, m - 1).toLocaleString('default', { month: 'short' })}</option>
                ))}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className={labelCls}>Year</label>
              <select value={year} onChange={e => setYear(Number(e.target.value))} className={selectCls}>
                {Array.from({ length: 5 }, (_, i) => currentDate.getFullYear() - 2 + i).map(y => (
                  <option key={y} value={y}>{y}</option>
                ))}
              </select>
            </div>
            <div className="flex flex-col gap-1.5 sm:col-span-2 lg:col-span-1">
              <label className={labelCls}>Status</label>
              <select value={status} onChange={e => setStatus(e.target.value)} className={selectCls}>
                <option value="all">All</option>
                <option value="pending">Pending</option>
                <option value="approved">Approved</option>
                <option value="rejected">Rejected</option>
                <option value="deducted">Deducted</option>
              </select>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="space-y-4">
        {selectedRows.length > 0 && (
          <div className="flex items-center gap-2 p-2 bg-muted/50 rounded-lg border">
            <span className="text-sm font-medium mr-2">{selectedRows.length} selected</span>
            <Button size="sm" variant="outline" className="text-emerald-600 hover:text-emerald-700" onClick={handleBulkApprove} disabled={bulkActionLoading}>
              <CheckIcon className="h-4 w-4 mr-1" /> Approve
            </Button>
            <Button size="sm" variant="outline" className="text-amber-600 hover:text-amber-700" onClick={handleBulkReject} disabled={bulkActionLoading}>
              <XIcon className="h-4 w-4 mr-1" /> Reject
            </Button>
            <Button size="sm" variant="outline" className="text-red-600 hover:text-red-700" onClick={handleBulkDelete} disabled={bulkActionLoading}>
              <TrashIcon className="h-4 w-4 mr-1" /> Delete
            </Button>
          </div>
        )}
        <div className="rounded-xl border bg-card text-card-foreground shadow-sm">
          {loading ? (
            <div className="flex h-64 items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-primary/50" />
            </div>
          ) : !companyId ? (
            <div className="flex h-64 flex-col items-center justify-center text-muted-foreground">
              <SearchIcon className="mb-2 h-10 w-10 opacity-20" />
              <p>Select a company to view advances</p>
            </div>
          ) : (
            <DataTable 
              columns={columns} 
              data={data} 
              enableSelection={true}
              onSelectionChange={setSelectedRows}
            />
          )}
        </div>
      </div>

      <AlertDialog open={!!deleteId} onOpenChange={() => setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete this pending advance salary request. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-red-500 hover:bg-red-600 text-white">Delete</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
