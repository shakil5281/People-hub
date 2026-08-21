"use client"

import * as React from "react"
import { ChartColumnIcon, FileSpreadsheetIcon, Loader2, FilterIcon, XIcon } from "lucide-react"
import { FilterBar } from "@/components/filter-bar"
import type { FilterDef } from "@/components/filter-bar"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from "@/components/ui/sheet"
import { attendanceApi, companyApi } from "@/lib/api"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"

interface OvertimeSummaryRecord {
  id: string
  name: string
  employee_count: number
  total_hours: number
}

interface Company { id: string; company_name_en: string }

const groupTabs = [
  { value: "department", label: "Department" },
  { value: "section", label: "Section" },
  { value: "designation", label: "Designation" },
  { value: "line", label: "Line" },
]

const today = new Date().toISOString().split("T")[0]

export default function OverTimeSummaryPage() {
  const [activeTab, setActiveTab] = React.useState("department")
  const [data, setData] = React.useState<OvertimeSummaryRecord[]>([])
  const [loading, setLoading] = React.useState(false)
  const [submitting, setSubmitting] = React.useState(false)
  const [filters, setFilters] = React.useState<Record<string, string>>({ date: today })
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [mobileFilterOpen, setMobileFilterOpen] = React.useState(false)
  const [exporting, setExporting] = React.useState(false)
  const [error, setError] = React.useState("")

  const groupLabel = groupTabs.find((t) => t.value === activeTab)?.label || "Department"

  const columns: ColumnDef<OvertimeSummaryRecord>[] = [
    { id: "sl", header: "Sl", cell: ({ row }) => row.index + 1 },
    { accessorKey: "name", header: groupLabel },
    { accessorKey: "employee_count", header: "Employees" },
    {
      accessorKey: "total_hours", header: "Total Hours",
      cell: ({ row }) => (row.original.total_hours ?? 0).toFixed(2),
    },
  ]

  const buildExportParams = () => {
    const date = filters.date || today
    const params: Record<string, string> = { start_date: date, end_date: date, group_by: activeTab }
    if (filters.company_id) params.company_id = filters.company_id
    return params
  }

  const handleExport = async () => {
    setExporting(true)
    try {
      const res = await attendanceApi.exportOvertimeSummaryExcel(buildExportParams())
      const blob = new Blob([res.data], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `overtime_summary_${activeTab}_${filters.date || today}.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setError("Failed to export overtime summary")
    } finally {
      setExporting(false)
    }
  }

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((res) => {
      setCompanies(res.data?.data || [])
    }).catch(() => {})
    fetchData()
  }, [])

  const filterDefs: FilterDef[] = React.useMemo(() => [
    { key: "date", label: "Date", type: "datepicker" },
    {
      key: "company_id", label: "Company", type: "select",
      options: companies.map((c) => ({ value: c.id, label: c.company_name_en })),
    },
  ], [companies])

  const fetchData = async (f?: Record<string, string>) => {
    setLoading(true)
    try {
      const params = f || filters
      const active: Record<string, string> = { start_date: params.date || today, end_date: params.date || today, group_by: activeTab }
      if (params.company_id) active.company_id = params.company_id
      const { data: res } = await attendanceApi.overtimeSummary(active)
      setData(res.summaries || [])
    } catch {
      setData([])
    } finally {
      setLoading(false)
    }
  }

  const handleApply = () => {
    setSubmitting(true)
    fetchData(filters).finally(() => setSubmitting(false))
  }

  const handleReset = () => {
    setFilters({ date: today })
    setData([])
  }

  const handleFilterChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const handleTabChange = (value: string) => {
    setActiveTab(value)
    const date = filters.date || today
    const active: Record<string, string> = { start_date: date, end_date: date, group_by: value }
    if (filters.company_id) active.company_id = filters.company_id
    setLoading(true)
    attendanceApi.overtimeSummary(active).then(({ data: res }) => {
      setData(res.summaries || [])
    }).catch(() => setData([])).finally(() => setLoading(false))
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <ChartColumnIcon className="h-6 w-6 text-muted-foreground" />
            <div>
              <h1 className="text-3xl font-bold tracking-tight">Over Time Summary</h1>
              <p className="text-muted-foreground mt-1">Overtime summary grouped by department, section, designation, or line</p>
            </div>
          </div>
          <div className="hidden md:flex gap-2">
            <Button onClick={handleExport} disabled={exporting} className="bg-primary text-primary-foreground hover:bg-primary/90">
              {exporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileSpreadsheetIcon className="mr-2 h-4 w-4" />}
              {exporting ? "Exporting..." : "Export Excel"}
            </Button>
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
                    onChange={handleFilterChange}
                    onApply={() => { handleApply(); setMobileFilterOpen(false) }}
                    onReset={() => { handleReset(); setMobileFilterOpen(false) }}
                    submitting={submitting}
                    singleColumn
                    noBorder
                  />
                </div>
              </SheetContent>
            </Sheet>
            <Button onClick={handleExport} disabled={exporting} className="flex-1 bg-primary text-primary-foreground hover:bg-primary/90">
              {exporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <FileSpreadsheetIcon className="mr-2 h-4 w-4" />}
              {exporting ? "Exporting..." : "Export"}
            </Button>
          </ButtonGroup>
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

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      <div className="px-4 lg:px-6">
        <Tabs value={activeTab} onValueChange={handleTabChange}>
          <TabsList>
            {groupTabs.map((t) => (
              <TabsTrigger key={t.value} value={t.value}>{t.label}</TabsTrigger>
            ))}
          </TabsList>
          {groupTabs.map((t) => (
            <TabsContent key={t.value} value={t.value} className="mt-4">
              <DataTable data={t.value === activeTab ? data : []} columns={columns} loading={loading} />
            </TabsContent>
          ))}
        </Tabs>
      </div>
    </div>
  )
}
