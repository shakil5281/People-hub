"use client"

import * as React from "react"
import { UserXIcon, PlusIcon, RotateCcwIcon, Loader2, CheckCircleIcon, XCircleIcon, FilterIcon, XIcon, FileSpreadsheetIcon, FileTextIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { useRouter } from "next/navigation"
import { Separation, separationTypeOptions, separationStatusOptions } from "@/components/data/separation-data"
import { separationApi, companyApi, departmentApi, sectionApi, designationApi, lineApi, groupApi } from "@/lib/api"
import type { Department, Section, Designation, Line } from "@/components/data/organization-data"
import type { Company } from "@/components/data/company-data"
import type { Group } from "@/components/data/group-data"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { ButtonGroup } from "@/components/ui/button-group"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"

const today = new Date().toISOString().split("T")[0]
const firstOfMonth = new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString().split("T")[0]

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  Approved: "default", Pending: "secondary", Rejected: "destructive", Processed: "outline", Cancelled: "outline",
}

export default function SeperationPage() {
  const router = useRouter()
  const [data, setData] = React.useState<Separation[]>([])
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [loading, setLoading] = React.useState(true)
  const [processing, setProcessing] = React.useState(false)
  const [processingId, setProcessingId] = React.useState<string | null>(null)
  const [exporting, setExporting] = React.useState<"excel" | "pdf" | null>(null)
  const [filters, setFilters] = React.useState<Record<string, string>>({
    date_from: firstOfMonth,
    date_to: today,
  })
  const [submitting, setSubmitting] = React.useState(false)
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const fetchData = React.useCallback(async (f?: Record<string, string>, p?: number, l?: number) => {
    setLoading(true)
    try {
      const params = { ...(f || {}), page: String(p ?? page), limit: String(l ?? limit) }
      const { data: res } = await separationApi.list(params)
      setData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      toast.error("Failed to load separations")
    } finally {
      setLoading(false)
    }
  }, [page, limit])

  const loadSections = React.useCallback(async (departmentId: string) => {
    try {
      const res = await sectionApi.list(departmentId, { limit: "100" })
      setSections(Array.isArray(res.data?.data) ? res.data.data : [])
    } catch {
      setSections([])
    }
  }, [])

  const loadDesignations = React.useCallback(async (sectionId: string) => {
    try {
      const res = await designationApi.list(sectionId, { limit: "100" })
      setDesignations(Array.isArray(res.data?.data) ? res.data.data : [])
    } catch {
      setDesignations([])
    }
  }, [])

  const loadLines = React.useCallback(async (sectionId: string) => {
    try {
      const res = await lineApi.list(sectionId, { limit: "100" })
      setLines(Array.isArray(res.data?.data) ? res.data.data : [])
    } catch {
      setLines([])
    }
  }, [])

  React.useEffect(() => {
    Promise.all([
      companyApi.list({ limit: "100" }),
      departmentApi.list({ limit: "100" }),
      groupApi.list({ limit: "100" }),
    ]).then(([cRes, dRes, gRes]) => {
      setCompanies(Array.isArray(cRes.data?.data) ? cRes.data.data : [])
      setDepartments(Array.isArray(dRes.data?.data) ? dRes.data.data : [])
      setGroups(Array.isArray(gRes.data?.data) ? gRes.data.data : [])
    }).catch(() => {})
    fetchData(filters)
  }, [])

  React.useEffect(() => {
    if (!filters.department_id) {
      setSections([])
      setDesignations([])
      setLines([])
      setFilters((prev) => {
        const next = { ...prev }
        delete next.section_id
        delete next.designation_id
        delete next.line_id
        return next
      })
      return
    }
    loadSections(filters.department_id)
  }, [filters.department_id, loadSections])

  React.useEffect(() => {
    if (!filters.section_id) {
      setDesignations([])
      setLines([])
      setFilters((prev) => {
        const next = { ...prev }
        delete next.designation_id
        delete next.line_id
        return next
      })
      return
    }
    loadDesignations(filters.section_id)
    loadLines(filters.section_id)
  }, [filters.section_id, loadDesignations, loadLines])

  React.useEffect(() => {
    fetchData(filters)
  }, [page, limit])

  const handleApply = async () => {
    setPage(1)
    const active: Record<string, string> = {}
    for (const [k, v] of Object.entries(filters)) {
      if (v) active[k] = v
    }
    setSubmitting(true)
    await fetchData(active, 1)
    setSubmitting(false)
  }

  const handleReset = () => {
    setPage(1)
    setLimit(20)
    setSections([])
    setDesignations([])
    setLines([])
    setFilters({ date_from: firstOfMonth, date_to: today })
    fetchData({ date_from: firstOfMonth, date_to: today }, 1, 20)
  }

  const handleChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleEdit = (s: Separation) => router.push(`/hr/seperation/${s.id}/edit`)

  const handleDelete = async (s: Separation) => {
    try {
      await separationApi.delete(s.id)
      toast.success("Separation deleted")
      fetchData(filters)
    } catch (err: unknown) {
      const msg = extractError(err)
      toast.error(msg || "Failed to delete separation")
    }
  }

  const handleProcessBatch = async () => {
    setProcessing(true)
    try {
      const { data: res } = await separationApi.process()
      toast.success(res.processed === 0 ? "No pending separations due" : (res.message || `Processed ${res.processed} separation(s)`))
      fetchData(filters)
    } catch {
      toast.error("Failed to process separations")
    } finally {
      setProcessing(false)
    }
  }

  const handleProcessOne = async (s: Separation) => {
    setProcessingId(s.id)
    try {
      const { data: res } = await separationApi.processOne(s.id)
      toast.success(res.message || `Processed ${s.employee}`)
      fetchData(filters)
    } catch (err: unknown) {
      toast.error(extractError(err) || "Failed to process")
    } finally {
      setProcessingId(null)
    }
  }

  const handleCancel = async (s: Separation) => {
    setProcessingId(s.id)
    try {
      await separationApi.cancel(s.id)
      toast.success(`Cancelled separation for ${s.employee}`)
      fetchData(filters)
    } catch (err: unknown) {
      toast.error(extractError(err) || "Failed to cancel")
    } finally {
      setProcessingId(null)
    }
  }

  const handleExport = async (kind: "excel" | "pdf") => {
    setExporting(kind)
    try {
      const active: Record<string, string> = {}
      for (const [k, v] of Object.entries(filters)) {
        if (v) active[k] = v
      }
      const res = kind === "excel" ? await separationApi.exportExcel(active) : await separationApi.exportPdf(active)
      const blob = new Blob([res.data], {
        type: kind === "excel" ? "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" : "application/pdf",
      })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement("a")
      link.href = url
      const ext = kind === "excel" ? "xlsx" : "pdf"
      link.download = `separation_report_${new Date().toISOString().slice(0, 10)}.${ext}`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
      toast.success(`Separation report (${kind.toUpperCase()}) exported successfully`)
    } catch (err: unknown) {
      toast.error(`Failed to export ${kind.toUpperCase()}`)
    } finally {
      setExporting(null)
    }
  }

  const handleFormExportPdf = async (s: Separation) => {
    setProcessingId(s.id)
    try {
      const res = await separationApi.exportFormPdf(s.id, "en")
      const blob = new Blob([res.data], { type: "application/pdf" })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement("a")
      link.href = url
      link.download = `separation_form_${s.employee_id || s.id}.pdf`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
      toast.success(`Separation form PDF generated for ${s.employee}`)
    } catch {
      toast.error("Failed to export separation form PDF")
    } finally {
      setProcessingId(null)
    }
  }

  const columns: ColumnDef<Separation>[] = React.useMemo(() => [
    { accessorKey: "employee", header: "Employee" },
    { accessorKey: "employee_id", header: "Emp. ID" },
    {
      accessorKey: "department",
      header: "Department",
      cell: ({ row }) => <span>{row.original.department?.name || row.original.department_id}</span>,
    },
    { accessorKey: "type", header: "Separation Type" },
    { accessorKey: "date", header: "Separation Date" },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const s = row.original.status
        return <Badge variant={statusVariant[s] || "secondary"}>{s}</Badge>
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => {
        const s = row.original
        const busy = processingId === s.id
        if (busy) return <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
        return (
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleFormExportPdf(s) }} title="Export Resignation Form PDF">
              <FileTextIcon className="h-4 w-4 text-blue-600" />
            </Button>
            {(s.status === "Pending" || s.status === "Approved") && (
              <>
                <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleProcessOne(s) }} title="Process now">
                  <CheckCircleIcon className="h-4 w-4 text-green-600" />
                </Button>
                <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleCancel(s) }} title="Cancel">
                  <XCircleIcon className="h-4 w-4 text-amber-600" />
                </Button>
              </>
            )}
          </div>
        )
      },
    },
  ], [processingId])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date_from", label: "Separation Start Date", type: "datepicker" },
    { key: "date_to", label: "Separation End Date", type: "datepicker" },
    { key: "company_id", label: "Company", type: "select", options: companies.map((c) => ({ value: c.id, label: c.company_name_en })) },
    { key: "department_id", label: "Department", type: "select", options: departments.map((d) => ({ value: d.id, label: d.name })) },
    {
      key: "section_id",
      label: "Section",
      type: "select",
      options: sections.map((s) => ({ value: s.id, label: s.name })),
      disabled: !filters.department_id,
    },
    {
      key: "designation_id",
      label: "Designation",
      type: "select",
      options: designations.map((d) => ({ value: d.id, label: d.name })),
      disabled: !filters.section_id,
    },
    {
      key: "line_id",
      label: "Line",
      type: "select",
      options: lines.map((l) => ({ value: l.id, label: l.name })),
      disabled: !filters.section_id,
    },
    { key: "group_id", label: "Group", type: "select", options: groups.map((g) => ({ value: g.id, label: g.name })) },
    { key: "employee", label: "Employee Name", type: "text", placeholder: "Filter by name..." },
    { key: "employee_id", label: "Emp. ID", type: "text", placeholder: "Filter by Emp. ID (e.g. 1)..." },
    { key: "type", label: "Separation Type", type: "select", options: separationTypeOptions.map((o) => ({ value: o.value, label: o.label })) },
    { key: "status", label: "Status", type: "select", options: separationStatusOptions.map((o) => ({ value: o.value, label: o.label })) },
  ], [companies, departments, sections, designations, lines, groups, filters.department_id, filters.section_id])

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <UserXIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Separation</h1>
            <p className="text-muted-foreground mt-1">Manage employee separations</p>
          </div>
        </div>
        <ButtonGroup className="hidden md:flex">
          <Button variant="outline" size="sm" onClick={() => handleExport("excel")} disabled={loading || !!exporting}>
            {exporting === "excel" ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileSpreadsheetIcon className="mr-2 h-4 w-4 text-green-600" />}
            Export Excel
          </Button>
          <Button variant="outline" size="sm" onClick={() => handleExport("pdf")} disabled={loading || !!exporting}>
            {exporting === "pdf" ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileTextIcon className="mr-2 h-4 w-4 text-red-600" />}
            Export PDF
          </Button>
          <Button variant="outline" size="sm" onClick={handleProcessBatch} disabled={processing}>
            {processing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RotateCcwIcon className="mr-2 h-4 w-4" />}
            {processing ? "Processing..." : "Process Due"}
          </Button>
          <Button size="sm" onClick={() => router.push("/hr/seperation/create")}>
            <PlusIcon className="mr-2 h-4 w-4" />
            Add Separation
          </Button>
        </ButtonGroup>
      </div>

      <div className="md:hidden px-4 lg:px-6">
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
                  submitting={submitting}
                  singleColumn
                  noBorder
                />
              </div>
            </SheetContent>
          </Sheet>
        </ButtonGroup>
        <ButtonGroup className="w-full mt-2">
          <Button variant="outline" size="sm" onClick={() => handleExport("excel")} disabled={loading || !!exporting} className="flex-1">
            {exporting === "excel" ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <FileSpreadsheetIcon className="mr-1.5 h-4 w-4 text-green-600" />}
            Excel
          </Button>
          <Button variant="outline" size="sm" onClick={() => handleExport("pdf")} disabled={loading || !!exporting} className="flex-1">
            {exporting === "pdf" ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <FileTextIcon className="mr-1.5 h-4 w-4 text-red-600" />}
            PDF
          </Button>
          <Button onClick={handleProcessBatch} disabled={processing} variant="outline" size="sm" className="flex-1">
            {processing ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <RotateCcwIcon className="mr-1.5 h-4 w-4" />}
            Process
          </Button>
          <Button onClick={() => router.push("/hr/seperation/create")} size="sm" className="flex-1">
            <PlusIcon className="mr-1.5 h-4 w-4" />
            Add
          </Button>
        </ButtonGroup>
      </div>

      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleChange}
          onApply={handleApply}
          onReset={handleReset}
          submitting={submitting}
        />
      </div>

      <DataTable
        data={data}
        columns={columns}
        onEdit={handleEdit}
        onDelete={handleDelete}
        serverSide={true}
        page={page}
        pageSize={limit}
        pageCount={totalPages}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={(size) => { setLimit(size); setPage(1); }}
        loading={loading}
      />
    </div>
  )
}

function extractError(err: unknown): string {
  if (typeof err === "object" && err !== null && "response" in err) {
    const ae = err as { response?: { data?: { error?: string } } }
    return ae.response?.data?.error || ""
  }
  return ""
}
