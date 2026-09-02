"use client"

import * as React from "react"
import {
  BanknoteIcon,
  DownloadIcon,
  UploadIcon,
  PencilIcon,
  Loader2,
  FilterIcon,
  XIcon,
  CheckCircle2Icon,
  AlertCircleIcon,
  CreditCardIcon,
  FileSpreadsheetIcon,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog"
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetClose,
} from "@/components/ui/sheet"
import {
  companyApi,
  departmentApi,
  sectionApi,
  designationApi,
  lineApi,
  groupApi,
  shiftApi,
  salaryAccountApi,
} from "@/lib/api"
import { toast } from "sonner"

interface Company { id: string; company_name_en: string }
interface Department { id: string; name: string }
interface Section { id: string; name: string }
interface Designation { id: string; name: string }
interface Line { id: string; name: string }
interface Group { id: string; name: string }
interface Shift { id: string; name: string }

interface EmployeeAccountRecord {
  id: string
  employee_id: string
  punch_number: string
  name_en: string
  name_bn: string
  phone: string
  designation: string
  department: string
  section: string
  line: string
  group: string
  gross_salary: number
  account_type: string
  account_number: string
  status: string
  company_id: string
}

interface ImportResult {
  success: boolean
  message: string
  total_rows: number
  updated: number
  skipped: number
  errors: string[]
}

export default function SalaryAccountPage() {
  const [data, setData] = React.useState<EmployeeAccountRecord[]>([])
  const [total, setTotal] = React.useState(0)
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState("")
  const [page, setPage] = React.useState(1)
  const [limit] = React.useState(20)

  // Lookups
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<Group[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])

  // Filter values
  const [filters, setFilters] = React.useState<Record<string, string>>({
    status: "active",
  })
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)

  // Single edit modal state
  const [editOpen, setEditOpen] = React.useState(false)
  const [selectedEmp, setSelectedEmp] = React.useState<EmployeeAccountRecord | null>(null)
  const [editAccountType, setEditAccountType] = React.useState("")
  const [editAccountNumber, setEditAccountNumber] = React.useState("")
  const [savingEdit, setSavingEdit] = React.useState(false)

  // Import modal state
  const [importOpen, setImportOpen] = React.useState(false)
  const [importFile, setImportFile] = React.useState<File | null>(null)
  const [importing, setImporting] = React.useState(false)
  const [importResult, setImportResult] = React.useState<ImportResult | null>(null)
  const [downloadingTemplate, setDownloadingTemplate] = React.useState(false)
  const [exportingExcel, setExportingExcel] = React.useState(false)

  const filterDefs: FilterDef[] = React.useMemo(() => [
    {
      key: "company_id",
      label: "Company",
      type: "select",
      options: companies.map((c) => ({ value: c.id, label: c.company_name_en })),
    },
    {
      key: "department_id",
      label: "Department",
      type: "select",
      options: departments.map((d) => ({ value: d.id, label: d.name })),
    },
    {
      key: "section_id",
      label: "Section",
      type: "select",
      options: sections.map((s) => ({ value: s.id, label: s.name })),
    },
    {
      key: "designation_id",
      label: "Designation",
      type: "select",
      options: designations.map((d) => ({ value: d.id, label: d.name })),
    },
    {
      key: "line_id",
      label: "Line",
      type: "select",
      options: lines.map((l) => ({ value: l.id, label: l.name })),
    },
    {
      key: "group_id",
      label: "Group",
      type: "select",
      options: groups.map((g) => ({ value: g.id, label: g.name })),
    },
    {
      key: "shift_id",
      label: "Shift",
      type: "select",
      options: shifts.map((s) => ({ value: s.id, label: s.name })),
    },
    {
      key: "account_type",
      label: "Account Type",
      type: "select",
      options: [
        { value: "mCash", label: "mCash" },
        { value: "Card", label: "Card" },
        { value: "none", label: "None / Unassigned" },
      ],
    },
    {
      key: "status",
      label: "Status",
      type: "select",
      options: [
        { value: "active", label: "Active" },
        { value: "inactive", label: "Inactive" },
        { value: "", label: "All Statuses" },
      ],
    },
    {
      key: "employee_id",
      label: "Employee ID",
      type: "text",
      placeholder: "Search by ID...",
    },
  ], [companies, departments, sections, designations, lines, groups, shifts])

  const buildParams = React.useCallback((f?: Record<string, string>, p?: number) => {
    const activeFilters = f || filters
    const params: Record<string, string> = {
      page: String(p || page),
      limit: String(limit),
    }
    if (activeFilters.company_id) params.company_id = activeFilters.company_id
    if (activeFilters.department_id) params.department_id = activeFilters.department_id
    if (activeFilters.section_id) params.section_id = activeFilters.section_id
    if (activeFilters.designation_id) params.designation_id = activeFilters.designation_id
    if (activeFilters.line_id) params.line_id = activeFilters.line_id
    if (activeFilters.group_id) params.group_id = activeFilters.group_id
    if (activeFilters.shift_id) params.shift_id = activeFilters.shift_id
    if (activeFilters.account_type) params.account_type = activeFilters.account_type
    if (activeFilters.status) params.status = activeFilters.status
    if (activeFilters.employee_id) params.employee_id = activeFilters.employee_id
    return params
  }, [filters, page, limit])

  const fetchData = React.useCallback(async (f?: Record<string, string>, p?: number) => {
    setLoading(true)
    setError("")
    try {
      const params = buildParams(f, p)
      const { data: res } = await salaryAccountApi.list(params)
      setData(res.data || [])
      setTotal(res.total || 0)
    } catch {
      setError("Failed to load employee salary account records")
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

  const handleFilterChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleApplyFilters = () => {
    setPage(1)
    fetchData(filters, 1)
  }

  const handleResetFilters = () => {
    const resetValues = { status: "active" }
    setFilters(resetValues)
    setPage(1)
    fetchData(resetValues, 1)
  }

  const openEditModal = (emp: EmployeeAccountRecord) => {
    setSelectedEmp(emp)
    setEditAccountType(emp.account_type || "")
    setEditAccountNumber(emp.account_number || "")
    setEditOpen(true)
  }

  const handleSaveEdit = async () => {
    if (!selectedEmp) return
    if (editAccountType && !editAccountNumber) {
      toast.error("Account number is required when account type is selected")
      return
    }
    if (!editAccountType && editAccountNumber) {
      toast.error("Account type is required when account number is provided")
      return
    }
    setSavingEdit(true)
    try {
      await salaryAccountApi.update(selectedEmp.id, {
        account_type: editAccountType,
        account_number: editAccountNumber,
      })
      toast.success("Salary account updated successfully")
      setEditOpen(false)
      fetchData(filters, page)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to update salary account"
      toast.error(msg)
    } finally {
      setSavingEdit(false)
    }
  }

  const handleDownloadTemplate = async () => {
    setDownloadingTemplate(true)
    try {
      const response = await salaryAccountApi.downloadTemplate()
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement("a")
      link.href = url
      link.setAttribute("download", "salary_account_template.xlsx")
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
      toast.success("Demo Excel format downloaded")
    } catch {
      toast.error("Failed to download template")
    } finally {
      setDownloadingTemplate(false)
    }
  }

  const handleExportExcel = async () => {
    setExportingExcel(true)
    try {
      const params = buildParams(filters)
      delete params.page
      delete params.limit
      const response = await salaryAccountApi.exportExcel(params)
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement("a")
      link.href = url
      link.setAttribute("download", `salary_accounts_${new Date().toISOString().slice(0, 10)}.xlsx`)
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
      toast.success("Salary accounts Excel exported successfully")
    } catch {
      toast.error("Failed to export Excel file")
    } finally {
      setExportingExcel(false)
    }
  }

  const handleImportSubmit = async () => {
    if (!importFile) {
      toast.error("Please select an Excel file to upload")
      return
    }
    setImporting(true)
    setImportResult(null)
    try {
      const { data: res } = await salaryAccountApi.importExcel(importFile)
      setImportResult(res)
      if (res.success) {
        toast.success(res.message || "Import completed successfully")
        fetchData(filters, page)
      } else {
        toast.warning(res.message || "Import finished with warnings")
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to import Excel file"
      toast.error(msg)
    } finally {
      setImporting(false)
    }
  }

  const totalPages = Math.ceil(total / limit)

  const renderAccountBadge = (type: string) => {
    const norm = (type || "").toLowerCase()
    if (norm === "mcash") {
      return <Badge className="bg-emerald-600 hover:bg-emerald-700 text-white">mCash</Badge>
    }
    if (norm === "card") {
      return <Badge className="bg-blue-600 hover:bg-blue-700 text-white">Card</Badge>
    }
    return <Badge variant="outline" className="text-muted-foreground">Unassigned</Badge>
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      {/* Header */}
      <div className="px-4 lg:px-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2.5">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <BanknoteIcon className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight md:text-3xl">Salary Account</h1>
              <p className="text-muted-foreground text-sm">
                Manage employee salary account numbers and bulk import account details
              </p>
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleExportExcel}
              disabled={exportingExcel || loading}
            >
              {exportingExcel ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <FileSpreadsheetIcon className="mr-2 h-4 w-4 text-emerald-600" />
              )}
              Export Excel
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleDownloadTemplate}
              disabled={downloadingTemplate}
            >
              {downloadingTemplate ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <DownloadIcon className="mr-2 h-4 w-4" />
              )}
              Demo Format
            </Button>
            <Button
              size="sm"
              onClick={() => {
                setImportFile(null)
                setImportResult(null)
                setImportOpen(true)
              }}
            >
              <UploadIcon className="mr-2 h-4 w-4" />
              Import Excel
            </Button>
            <div className="md:hidden">
              <Sheet open={mobileFilterOpen} onOpenChange={setMobileFilterOpen}>
                <SheetContent side="right" className="w-full sm:max-w-md p-0 flex flex-col" showCloseButton={false}>
                  <SheetHeader className="px-4 py-3 border-b flex flex-row items-center justify-between">
                    <SheetTitle className="text-base">Filters</SheetTitle>
                    <SheetClose asChild>
                      <Button variant="ghost" size="icon">
                        <XIcon className="h-4 w-4" />
                      </Button>
                    </SheetClose>
                  </SheetHeader>
                  <div className="flex-1 overflow-y-auto px-4 py-4">
                    <FilterBar
                      filters={filterDefs}
                      values={filters}
                      onChange={handleFilterChange}
                      onApply={() => { handleApplyFilters(); setMobileFilterOpen(false) }}
                      onReset={() => { handleResetFilters(); setMobileFilterOpen(false) }}
                      submitting={loading}
                      singleColumn
                      noBorder
                    />
                  </div>
                </SheetContent>
              </Sheet>
              <Button variant="outline" size="sm" onClick={() => setMobileFilterOpen(true)}>
                <FilterIcon className="mr-2 h-4 w-4" />
                Filters
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Desktop FilterBar */}
      <div className="px-4 lg:px-6 hidden md:block">
        <FilterBar
          filters={filterDefs}
          values={filters}
          onChange={handleFilterChange}
          onApply={handleApplyFilters}
          onReset={handleResetFilters}
          submitting={loading}
        />
      </div>

      {/* Error Message */}
      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      {/* Data Table */}
      <div className="px-4 lg:px-6">
        <div className="rounded-lg border bg-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50 text-left">
                  <th className="px-4 py-3 font-medium text-muted-foreground">Emp. ID</th>
                  <th className="px-4 py-3 font-medium text-muted-foreground hidden sm:table-cell">Punch No</th>
                  <th className="px-4 py-3 font-medium text-muted-foreground">Employee Name</th>
                  <th className="px-4 py-3 font-medium text-muted-foreground hidden md:table-cell">Designation</th>
                  <th className="px-4 py-3 font-medium text-muted-foreground hidden lg:table-cell">Department / Section</th>
                  <th className="px-4 py-3 font-medium text-muted-foreground text-right hidden sm:table-cell">Gross Salary</th>
                  <th className="px-4 py-3 font-medium text-muted-foreground text-center">Account Type</th>
                  <th className="px-4 py-3 font-medium text-muted-foreground">Account Number</th>
                  <th className="px-4 py-3 font-medium text-muted-foreground text-center hidden md:table-cell">Status</th>
                  <th className="px-4 py-3 font-medium text-muted-foreground text-right w-16">Action</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={10} className="px-4 py-12 text-center">
                      <Loader2 className="h-6 w-6 animate-spin text-muted-foreground mx-auto" />
                    </td>
                  </tr>
                ) : data.length === 0 ? (
                  <tr>
                    <td colSpan={10} className="px-4 py-10 text-center text-muted-foreground">
                      No employee salary account records found.
                    </td>
                  </tr>
                ) : (
                  data.map((row) => (
                    <tr key={row.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 font-mono font-medium text-xs sm:text-sm text-foreground">
                        {row.employee_id}
                      </td>
                      <td className="px-4 py-3 font-mono text-xs text-muted-foreground hidden sm:table-cell">
                        {row.punch_number || "-"}
                      </td>
                      <td className="px-4 py-3">
                        <div className="font-medium text-foreground">{row.name_en}</div>
                        {row.name_bn && <div className="text-xs text-muted-foreground">{row.name_bn}</div>}
                      </td>
                      <td className="px-4 py-3 text-muted-foreground hidden md:table-cell">
                        {(() => { const v = row.designation as unknown; return (typeof v === "object" && v !== null ? (v as { name?: string }).name : v as string) || "-" })()}
                      </td>
                      <td className="px-4 py-3 text-xs text-muted-foreground hidden lg:table-cell">
                        <div>{(() => { const v = row.department as unknown; return (typeof v === "object" && v !== null ? (v as { name?: string }).name : v as string) || "-" })()}</div>
                        {(() => { const v = row.section as unknown; const d = typeof v === "object" && v !== null ? (v as { name?: string }).name : v as string; return d ? <div className="text-muted-foreground/75">{d}</div> : null })()}
                      </td>
                      <td className="px-4 py-3 text-right font-medium hidden sm:table-cell">
                        ৳{row.gross_salary?.toLocaleString() || "0"}
                      </td>
                      <td className="px-4 py-3 text-center">
                        {renderAccountBadge(row.account_type)}
                      </td>
                      <td className="px-4 py-3 font-mono text-xs sm:text-sm">
                        {row.account_number ? (
                          <span className="font-semibold text-foreground tracking-wide">
                            {row.account_number}
                          </span>
                        ) : (
                          <span className="text-muted-foreground/50">--</span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-center hidden md:table-cell">
                        <Badge variant={row.status === "active" ? "secondary" : "outline"} className="capitalize text-xs">
                          {row.status}
                        </Badge>
                      </td>
                      <td className="px-4 py-3 text-right">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8"
                          onClick={() => openEditModal(row)}
                          title="Edit Salary Account"
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

          {/* Pagination Footer */}
          {total > 0 && (
            <div className="border-t bg-muted/30 px-4 py-3 flex items-center justify-between text-sm">
              <span className="text-muted-foreground">
                Total Records: <b>{total}</b>
              </span>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => {
                    setPage(page - 1)
                    fetchData(filters, page - 1)
                  }}
                >
                  Previous
                </Button>
                <span className="text-xs text-muted-foreground">
                  Page {page} of {totalPages || 1}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= totalPages}
                  onClick={() => {
                    setPage(page + 1)
                    fetchData(filters, page + 1)
                  }}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Edit Salary Account Dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <CreditCardIcon className="h-5 w-5 text-primary" />
              Edit Salary Account
            </DialogTitle>
            <DialogDescription>
              Update account type and account number for the employee.
            </DialogDescription>
          </DialogHeader>

          {selectedEmp && (
            <div className="flex flex-col gap-4 py-2">
              <div className="rounded-lg bg-muted/40 p-3 space-y-1 text-sm border">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Employee:</span>
                  <span className="font-semibold">{selectedEmp.name_en}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Employee ID:</span>
                  <span className="font-mono">{selectedEmp.employee_id}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Designation:</span>
                  <span>{selectedEmp.designation || "-"}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Gross Salary:</span>
                  <span className="font-medium">৳{selectedEmp.gross_salary?.toLocaleString() || "0"}</span>
                </div>
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-foreground">Account Type</label>
                <select
                  value={editAccountType}
                  onChange={(e) => setEditAccountType(e.target.value)}
                  className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                >
                  <option value="">-- Select Account Type --</option>
                  <option value="mCash">mCash</option>
                  <option value="Card">Card</option>
                </select>
                {editAccountType === "mCash" && (
                  <p className="text-xs text-muted-foreground">Standard mCash format: 12 digits</p>
                )}
                {editAccountType === "Card" && (
                  <p className="text-xs text-muted-foreground">Standard Card format: 17 digits</p>
                )}
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-foreground">Account Number</label>
                <input
                  type="text"
                  value={editAccountNumber}
                  onChange={(e) => setEditAccountNumber(e.target.value)}
                  placeholder="Enter account number..."
                  className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring font-mono"
                />
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSaveEdit} disabled={savingEdit}>
              {savingEdit && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Save Account
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Import Excel Dialog */}
      <Dialog open={importOpen} onOpenChange={setImportOpen}>
        <DialogContent className="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <FileSpreadsheetIcon className="h-5 w-5 text-primary" />
              Import Salary Accounts from Excel
            </DialogTitle>
            <DialogDescription>
              Upload an Excel (.xlsx / .xls) file to update employee salary account numbers in bulk.
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-4 py-2">
            {/* Download template notice */}
            <div className="flex items-center justify-between rounded-lg border bg-muted/40 p-3">
              <div className="text-xs text-muted-foreground">
                Required columns: <b className="text-foreground">EmployeeId</b>, <b className="text-foreground">AccountType</b> (mCash or Card), <b className="text-foreground">AccountNumber</b>
              </div>
              <Button
                variant="outline"
                size="sm"
                className="h-8 text-xs"
                onClick={handleDownloadTemplate}
                disabled={downloadingTemplate}
              >
                <DownloadIcon className="mr-1.5 h-3.5 w-3.5" />
                Demo Template
              </Button>
            </div>

            {/* File dropzone / upload input */}
            <div
              className="flex flex-col items-center justify-center rounded-lg border-2 border-dashed border-muted-foreground/25 p-6 text-center hover:bg-muted/20 transition-colors cursor-pointer"
              onClick={() => document.getElementById("salary-account-excel-input")?.click()}
            >
              <FileSpreadsheetIcon className="h-10 w-10 text-muted-foreground/60 mb-2" />
              <input
                id="salary-account-excel-input"
                type="file"
                accept=".xlsx,.xls"
                className="hidden"
                onChange={(e) => {
                  if (e.target.files && e.target.files[0]) {
                    setImportFile(e.target.files[0])
                    setImportResult(null)
                  }
                }}
              />
              <p className="text-sm font-medium">
                {importFile ? importFile.name : "Click to browse or drag and drop Excel file"}
              </p>
              <p className="text-xs text-muted-foreground mt-1">
                {importFile ? `${(importFile.size / 1024).toFixed(1)} KB` : "Supports .xlsx and .xls formats"}
              </p>
            </div>

            {/* Import results summary */}
            {importResult && (
              <div className="rounded-lg border bg-muted/30 p-3 space-y-2 text-sm">
                <div className="flex items-center gap-2 font-medium">
                  {importResult.updated > 0 ? (
                    <CheckCircle2Icon className="h-4 w-4 text-green-600" />
                  ) : (
                    <AlertCircleIcon className="h-4 w-4 text-amber-600" />
                  )}
                  <span>{importResult.message}</span>
                </div>
                <div className="grid grid-cols-3 gap-2 text-xs pt-1">
                  <div className="rounded bg-background p-2 border text-center">
                    <div className="text-muted-foreground">Total Rows</div>
                    <div className="text-base font-bold">{importResult.total_rows}</div>
                  </div>
                  <div className="rounded bg-green-500/10 p-2 border border-green-500/20 text-center text-green-700 dark:text-green-400">
                    <div>Updated</div>
                    <div className="text-base font-bold">{importResult.updated}</div>
                  </div>
                  <div className="rounded bg-amber-500/10 p-2 border border-amber-500/20 text-center text-amber-700 dark:text-amber-400">
                    <div>Skipped</div>
                    <div className="text-base font-bold">{importResult.skipped}</div>
                  </div>
                </div>

                {importResult.errors && importResult.errors.length > 0 && (
                  <div className="mt-2 max-h-32 overflow-y-auto rounded bg-destructive/10 p-2 text-xs text-destructive space-y-1">
                    <div className="font-semibold">Import Errors / Skipped Items:</div>
                    {importResult.errors.map((err, idx) => (
                      <div key={idx}>• {err}</div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setImportOpen(false)}>
              Close
            </Button>
            <Button onClick={handleImportSubmit} disabled={!importFile || importing}>
              {importing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Import & Update Accounts
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
