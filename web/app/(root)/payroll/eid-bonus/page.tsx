"use client"

import * as React from "react"
import { GiftIcon, Loader2, DownloadIcon } from "lucide-react"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { eidBonusApi, companyApi } from "@/lib/api"

interface EidBonus {
  id: string
  employee_id: string
  year: number
  bonus_type: string
  gross_salary: number
  bonus_amount: number
  status: string
  employee?: {
    name_en: string
    department?: { name: string }
    designation_ref?: { name: string }
    line_ref?: { name: string }
    group_ref?: { name: string }
    account_number: string
    account_type: string
  }
}

interface SummaryRow {
  id: number
  group_key: string
  employees: number
  gross_salary: number
  bonus_amount: number
}

const selectClass = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"

export default function EidBonusPage() {
  const [tab, setTab] = React.useState("process")
  const [companies, setCompanies] = React.useState<{ id: string; company_name_en: string }[]>([])
  const [companyId, setCompanyId] = React.useState("")
  const [year, setYear] = React.useState(new Date().getFullYear().toString())
  const [bonusType, setBonusType] = React.useState("")

  // Process
  const [processing, setProcessing] = React.useState(false)

  // Sheet
  const [sheetData, setSheetData] = React.useState<EidBonus[]>([])
  const [sheetLoading, setSheetLoading] = React.useState(false)
  const [sheetTotals, setSheetTotals] = React.useState<{ gross_salary: number; bonus_amount: number } | null>(null)

  // Summary
  const [summaryData, setSummaryData] = React.useState<SummaryRow[]>([])
  const [summaryLoading, setSummaryLoading] = React.useState(false)
  const [groupBy, setGroupBy] = React.useState("department")

  // Bank Sheet
  const [bankData, setBankData] = React.useState<EidBonus[]>([])
  const [bankLoading, setBankLoading] = React.useState(false)
  const [bankTotals, setBankTotals] = React.useState<{ gross_salary: number; bonus_amount: number } | null>(null)

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then(({ data }) => {
      const list = Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : []
      setCompanies(list)
      if (list.length > 0 && !companyId) setCompanyId(list[0].id)
    })
  }, [])

  const params = React.useMemo(() => {
    const p: Record<string, string> = { company_id: companyId, year }
    if (bonusType) p.bonus_type = bonusType
    return p
  }, [companyId, year, bonusType])

  const fetchSheet = React.useCallback(async () => {
    if (!companyId || !year) return
    setSheetLoading(true)
    try {
      const { data } = await eidBonusApi.sheet(params)
      setSheetData(data.bonuses || [])
      setSheetTotals(data.totals || null)
    } catch { setSheetData([]); toast.error("Failed to load sheet") }
    finally { setSheetLoading(false) }
  }, [params])

  const fetchSummary = React.useCallback(async () => {
    if (!companyId || !year) return
    setSummaryLoading(true)
    try {
      const { data } = await eidBonusApi.summary({ ...params, group_by: groupBy })
      setSummaryData((data.summaries || []).map((s: SummaryRow, i: number) => ({ ...s, id: i + 1 })))
    } catch { setSummaryData([]); toast.error("Failed to load summary") }
    finally { setSummaryLoading(false) }
  }, [params, groupBy])

  const fetchBank = React.useCallback(async () => {
    if (!companyId || !year) return
    setBankLoading(true)
    try {
      const { data } = await eidBonusApi.bankSheet(params)
      setBankData(data.bonuses || [])
      setBankTotals(data.totals || null)
    } catch { setBankData([]); toast.error("Failed to load bank sheet") }
    finally { setBankLoading(false) }
  }, [params])

  React.useEffect(() => { if (tab === "sheet") fetchSheet() }, [tab, fetchSheet])
  React.useEffect(() => { if (tab === "summary") fetchSummary() }, [tab, fetchSummary])
  React.useEffect(() => { if (tab === "bank") fetchBank() }, [tab, fetchBank])

  const handleProcess = async () => {
    if (!companyId) { toast.error("Select a company"); return }
    setProcessing(true)
    try {
      const { data } = await eidBonusApi.process({ company_id: companyId, year: parseInt(year) })
      toast.success(data.message || "Eid bonus processed")
    } catch { toast.error("Failed to process eid bonus") }
    finally { setProcessing(false) }
  }

  const handleExport = async () => {
    try {
      const res = await eidBonusApi.exportExcel(params)
      const blob = new Blob([res.data as BlobPart], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `eid_bonus_${year}.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch { toast.error("Failed to export") }
  }

  const sheetColumns: ColumnDef<EidBonus>[] = [
    { accessorKey: "employee_id", header: "Employee ID" },
    { header: "Name", accessorFn: (r) => r.employee?.name_en || "-" },
    { header: "Dept", accessorFn: (r) => r.employee?.department?.name || "-" },
    { header: "Designation", accessorFn: (r) => r.employee?.designation_ref?.name || "-" },
    { accessorKey: "gross_salary", header: "Gross", cell: ({ row }) => `৳${row.original.gross_salary.toLocaleString()}` },
    { accessorKey: "bonus_amount", header: "Bonus", cell: ({ row }) => `৳${row.original.bonus_amount.toLocaleString()}` },
  ]

  const summaryColumns: ColumnDef<SummaryRow>[] = [
    { accessorKey: "group_key", header: groupBy === "line" ? "Line" : "Department" },
    { accessorKey: "employees", header: "Employees" },
    { header: "Gross Salary", accessorFn: (r) => `৳${r.gross_salary.toLocaleString()}` },
    { header: "Bonus Amount", accessorFn: (r) => `৳${r.bonus_amount.toLocaleString()}` },
  ]

  const bankColumns: ColumnDef<EidBonus>[] = [
    { accessorKey: "employee_id", header: "Employee ID" },
    { header: "Name", accessorFn: (r) => r.employee?.name_en || "-" },
    { header: "Account No", accessorFn: (r) => r.employee?.account_number || "-" },
    { header: "Account Type", accessorFn: (r) => r.employee?.account_type || "-" },
    { accessorKey: "bonus_amount", header: "Bonus Amount", cell: ({ row }) => `৳${row.original.bonus_amount.toLocaleString()}` },
  ]

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <GiftIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Eid Bonus</h1>
            <p className="text-muted-foreground mt-1">Manage festival eid bonus</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6 space-y-4">
        <div className="flex flex-wrap items-end gap-3">
          <div className="flex flex-col gap-1.5">
            <Label>Company</Label>
            <select value={companyId} onChange={(e) => setCompanyId(e.target.value)} className={selectClass + " w-48"}>
              <option value="">Select Company</option>
              {companies.map((c) => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
            </select>
          </div>
          <div className="flex flex-col gap-1.5">
            <Label>Year</Label>
            <Input type="number" value={year} onChange={(e) => setYear(e.target.value)} className="w-24" />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label>Bonus Type</Label>
            <select value={bonusType} onChange={(e) => setBonusType(e.target.value)} className={selectClass + " w-36"}>
              <option value="">All Types</option>
              <option value="eid">Eid</option>
            </select>
          </div>
          <Button variant="outline" onClick={handleExport} disabled={!companyId || !year}>
            <DownloadIcon className="mr-2 h-4 w-4" />
            Export Excel
          </Button>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Tabs value={tab} onValueChange={setTab}>
          <TabsList>
            <TabsTrigger value="process">Process</TabsTrigger>
            <TabsTrigger value="sheet">Sheet</TabsTrigger>
            <TabsTrigger value="summary">Summary</TabsTrigger>
            <TabsTrigger value="bank">Bank Sheet</TabsTrigger>
          </TabsList>

          <TabsContent value="process" className="mt-4">
            <Card>
              <CardHeader>
                <CardTitle>Process Eid Bonus</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground mb-4">
                  Process eid bonus for all active employees for the selected year. Bonus amount equals gross salary.
                </p>
                <Button onClick={handleProcess} disabled={processing || !companyId}>
                  {processing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  Run Process
                </Button>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="sheet" className="mt-4">
            <Card>
              <CardHeader>
                <CardTitle>Eid Bonus Sheet</CardTitle>
              </CardHeader>
              <CardContent>
                <DataTable
                  data={sheetData}
                  columns={sheetColumns}
                  enableDnd={false}
                  enableSelection={false}
                  loading={sheetLoading}
                />
                {sheetTotals && (
                  <div className="flex justify-end gap-6 mt-3 text-sm">
                    <span>Gross: <strong>৳{sheetTotals.gross_salary.toLocaleString()}</strong></span>
                    <span>Bonus: <strong>৳{sheetTotals.bonus_amount.toLocaleString()}</strong></span>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="summary" className="mt-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Eid Bonus Summary</CardTitle>
                <select value={groupBy} onChange={(e) => setGroupBy(e.target.value)} className={selectClass + " w-40"}>
                  <option value="department">Department</option>
                  <option value="line">Line</option>
                </select>
              </CardHeader>
              <CardContent>
                <DataTable
                  data={summaryData}
                  columns={summaryColumns}
                  enableDnd={false}
                  enableSelection={false}
                  loading={summaryLoading}
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="bank" className="mt-4">
            <Card>
              <CardHeader>
                <CardTitle>Bank Sheet</CardTitle>
              </CardHeader>
              <CardContent>
                <DataTable
                  data={bankData}
                  columns={bankColumns}
                  enableDnd={false}
                  enableSelection={false}
                  loading={bankLoading}
                />
                {bankTotals && (
                  <div className="flex justify-end gap-6 mt-3 text-sm">
                    <span>Total Bonus: <strong>৳{bankTotals.bonus_amount.toLocaleString()}</strong></span>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
