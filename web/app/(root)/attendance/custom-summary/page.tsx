"use client"

import * as React from "react"
import { Loader2, DownloadIcon, FileTextIcon, ArrowLeftIcon, CalendarIcon, BuildingIcon } from "lucide-react"
import { toast } from "sonner"
import { attendanceApi, companyApi } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { DatePicker } from "@/components/ui/date-picker"
import Link from "next/link"

interface Company {
  id: string
  company_name_en: string
}

interface CustomSummaryRow {
  parent_section: string
  section: string
  present: number
  absent: number
  leave: number
  others: number
  total: number
  remarks: string
  is_subtotal: boolean
  is_grand_total: boolean
  style_type: string
}

interface CustomSummaryData {
  company_name: string
  company_address: string
  report_title: string
  date: string
  formatted_date: string
  rows: CustomSummaryRow[]
  grand_total: CustomSummaryRow
}

const today = new Date().toISOString().split("T")[0]

const toDate = (s: string): Date | undefined => {
  if (!s) return undefined
  const [y, m, d] = s.split("-").map(Number)
  if (y && m && d) return new Date(y, m - 1, d)
  return undefined
}

const fromDate = (d?: Date): string => {
  if (!d) return ""
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, "0")
  const day = String(d.getDate()).padStart(2, "0")
  return `${y}-${m}-${day}`
}

export default function CustomSummaryPage() {
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [companyId, setCompanyId] = React.useState("")
  const [date, setDate] = React.useState(today)
  const [data, setData] = React.useState<CustomSummaryData | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [exporting, setExporting] = React.useState<"excel" | "pdf" | null>(null)

  const fetchData = React.useCallback(async (cId: string, dStr: string) => {
    setLoading(true)
    try {
      const params: Record<string, string> = { date: dStr }
      if (cId) params.company_id = cId
      const res = await attendanceApi.customDailySummary(params)
      setData(res.data)
    } catch {
      toast.error("Failed to load custom summary report")
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    const init = async () => {
      try {
        const cRes = await companyApi.list({ limit: "100" })
        if (Array.isArray(cRes.data?.data) && cRes.data.data.length > 0) {
          setCompanies(cRes.data.data)
          setCompanyId(cRes.data.data[0].id)
          fetchData(cRes.data.data[0].id, today)
        } else {
          fetchData("", today)
        }
      } catch {
        fetchData("", today)
      }
    }
    init()
  }, [fetchData])

  const handleSearch = () => {
    fetchData(companyId, date)
  }

  const handleExport = async (kind: "excel" | "pdf") => {
    if (exporting) return
    setExporting(kind)
    try {
      const params: Record<string, string> = { date }
      if (companyId) params.company_id = companyId
      const res = kind === "pdf"
        ? await attendanceApi.exportCustomDailySummaryPdf(params)
        : await attendanceApi.exportCustomDailySummaryExcel(params)

      const url = window.URL.createObjectURL(new Blob([res.data]))
      const a = document.createElement("a")
      a.href = url
      a.download = `daily_summary_${date}.${kind === "pdf" ? "pdf" : "xlsx"}`
      a.click()
      window.URL.revokeObjectURL(url)
    } catch {
      toast.error(`Failed to export ${kind.toUpperCase()}`)
    } finally {
      setExporting(null)
    }
  }

  // Count sewing rows to merge parent cell
  const sewingRowsCount = data?.rows.filter((r) => r.parent_section === "Sewing").length || 0

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6 px-4 lg:px-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div className="flex items-center gap-3">
          <Link href="/attendance/daily-summary">
            <Button variant="outline" size="icon" className="h-9 w-9">
              <ArrowLeftIcon className="h-4 w-4" />
            </Button>
          </Link>
          <div>
            <h1 className="text-xl md:text-3xl font-bold tracking-tight">Custom Daily Summary</h1>
            <p className="text-muted-foreground text-xs md:text-sm mt-0.5">
              Daily attendance summary report by section, sewing lines, and staff
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => handleExport("excel")} disabled={loading || !!exporting}>
            {exporting === "excel" ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <DownloadIcon className="mr-1.5 h-4 w-4" />}
            Export Excel
          </Button>
          <Button variant="outline" size="sm" onClick={() => handleExport("pdf")} disabled={loading || !!exporting}>
            {exporting === "pdf" ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <FileTextIcon className="mr-1.5 h-4 w-4" />}
            Export PDF
          </Button>
        </div>
      </div>

      {/* Filter Bar */}
      <Card className="shadow-none border-slate-200">
        <CardContent className="p-4 flex flex-wrap items-center gap-4">
          {companies.length > 0 && (
            <div className="flex items-center gap-2 min-w-[200px]">
              <BuildingIcon className="h-4 w-4 text-muted-foreground" />
              <Select value={companyId} onValueChange={setCompanyId}>
                <SelectTrigger className="h-9">
                  <SelectValue placeholder="Select company" />
                </SelectTrigger>
                <SelectContent>
                  {companies.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.company_name_en}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          <div className="flex items-center gap-2">
            <DatePicker
              value={toDate(date)}
              onChange={(d) => setDate(fromDate(d) || today)}
            />
          </div>

          <Button size="sm" onClick={handleSearch} disabled={loading}>
            {loading ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : null}
            Generate Report
          </Button>
        </CardContent>
      </Card>

      {/* Report Document Preview */}
      {loading ? (
        <div className="flex items-center justify-center py-20 text-muted-foreground">
          <Loader2 className="h-8 w-8 animate-spin mr-2" />
          Loading report...
        </div>
      ) : data ? (
        <Card className="overflow-hidden border-slate-300 shadow-md max-w-4xl mx-auto w-full bg-white text-slate-900">
          <CardContent className="p-6 md:p-8 space-y-4">
            {/* Header Document Titles */}
            <div className="text-center space-y-1">
              <h2 className="text-2xl md:text-3xl font-extrabold tracking-wide text-purple-800 uppercase">
                {data.company_name}
              </h2>
              <p className="text-xs md:text-sm italic font-serif text-slate-700">{data.company_address}</p>
              <p className="text-xs md:text-sm italic font-serif text-slate-700">{data.report_title}</p>
              <div className="text-right text-xs md:text-sm font-semibold text-slate-800 pt-2">
                Date:- {data.formatted_date}
              </div>
            </div>

            {/* Custom Report Table */}
            <div className="overflow-x-auto border border-slate-300 rounded-sm">
              <table className="w-full text-xs md:text-sm border-collapse">
                <thead>
                  <tr className="border-b border-slate-300 bg-slate-50 text-blue-900">
                    <th colSpan={2} className="border-r border-slate-300 py-2 px-3 text-center font-bold">
                      Section
                    </th>
                    <th className="border-r border-slate-300 py-2 px-3 text-center font-bold w-20">Present</th>
                    <th className="border-r border-slate-300 py-2 px-3 text-center font-bold w-20">Abesnt</th>
                    <th className="border-r border-slate-300 py-2 px-3 text-center font-bold w-20">Leave</th>
                    <th className="border-r border-slate-300 py-2 px-3 text-center font-bold w-20">Others</th>
                    <th className="border-r border-slate-300 py-2 px-3 text-center font-bold w-20">Total</th>
                    <th className="py-2 px-3 text-center font-bold w-24">Remarks</th>
                  </tr>
                </thead>
                <tbody>
                  {data.rows.map((r, idx) => {
                    const isSewingFirst = r.parent_section === "Sewing" && (idx === 0 || data.rows[idx - 1]?.parent_section !== "Sewing")

                    const rowBgClass =
                      r.style_type === "subtotal" || r.style_type === "staff_total"
                        ? "bg-sky-100 font-bold"
                        : r.style_type === "worker_total"
                        ? "bg-sky-200 font-bold"
                        : "hover:bg-slate-50"

                    return (
                      <tr key={idx} className={`border-b border-slate-300 ${rowBgClass}`}>
                        {/* Parent Section column for Sewing */}
                        {r.parent_section === "Sewing" ? (
                          isSewingFirst ? (
                            <td
                              rowSpan={sewingRowsCount}
                              className="border-r border-slate-300 py-2 px-3 font-bold text-center align-middle bg-white w-24 text-slate-900"
                            >
                              Sewing
                            </td>
                          ) : null
                        ) : (
                          <td colSpan={2} className="border-r border-slate-300 py-1.5 px-3 font-medium text-slate-900">
                            {r.section}
                          </td>
                        )}

                        {r.parent_section === "Sewing" && (
                          <td className="border-r border-slate-300 py-1.5 px-3 font-medium text-slate-900">
                            {r.section}
                          </td>
                        )}

                        <td className="border-r border-slate-300 py-1.5 px-3 text-center text-slate-900">
                          {r.present}
                        </td>
                        <td className="border-r border-slate-300 py-1.5 px-3 text-center text-slate-900">
                          {r.absent}
                        </td>
                        <td className="border-r border-slate-300 py-1.5 px-3 text-center text-slate-900">
                          {r.leave}
                        </td>
                        <td className="border-r border-slate-300 py-1.5 px-3 text-center text-slate-900">
                          {r.others}
                        </td>
                        <td className="border-r border-slate-300 py-1.5 px-3 text-center text-slate-900">
                          {r.total}
                        </td>
                        <td className="py-1.5 px-3 text-center text-slate-500">{r.remarks}</td>
                      </tr>
                    )
                  })}
                </tbody>

                {/* Grand Total Row */}
                <tfoot>
                  <tr className="bg-[#00A0E9] text-white font-bold text-sm">
                    <td colSpan={2} className="border-r border-slate-300 py-2.5 px-3 text-center uppercase tracking-wide">
                      {data.grand_total.section}
                    </td>
                    <td className="border-r border-slate-300 py-2.5 px-3 text-center">{data.grand_total.present}</td>
                    <td className="border-r border-slate-300 py-2.5 px-3 text-center">{data.grand_total.absent}</td>
                    <td className="border-r border-slate-300 py-2.5 px-3 text-center">{data.grand_total.leave}</td>
                    <td className="border-r border-slate-300 py-2.5 px-3 text-center">{data.grand_total.others}</td>
                    <td className="border-r border-slate-300 py-2.5 px-3 text-center">{data.grand_total.total}</td>
                    <td className="py-2.5 px-3 text-center"></td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </CardContent>
        </Card>
      ) : null}
    </div>
  )
}
