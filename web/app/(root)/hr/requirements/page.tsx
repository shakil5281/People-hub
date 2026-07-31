"use client"

import * as React from "react"
import { ClipboardListIcon, PlusIcon, Loader2 } from "lucide-react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { sectionTabOptions } from "@/components/data/requirement-data"
import { requirementApi } from "@/lib/api"

interface SectionRow {
  section_name: string
  group_type: string
  designation: string
  required: number
  present: number
  net: number
}

export default function RequirementsPage() {
  const router = useRouter()
  const [allData, setAllData] = React.useState<SectionRow[]>([])
  const [loading, setLoading] = React.useState(true)
  const [activeTab, setActiveTab] = React.useState("Summary")

  const fetchData = React.useCallback(async () => {
    setLoading(true)
    try {
      const { data: res } = await requirementApi.sectionSummary()
      setAllData(Array.isArray(res.data) ? res.data : [])
    } catch {
      setAllData([])
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchData() }, [fetchData])

  const filteredData = React.useMemo(() => {
    if (activeTab === "Summary") return allData
    return allData.filter((r) => r.section_name === activeTab)
  }, [allData, activeTab])

  const groupedData = React.useMemo(() => {
    const groups: Record<string, SectionRow[]> = { Staff: [], Worker: [] }
    for (const row of filteredData) {
      const g = row.group_type === "Staff" ? "Staff" : "Worker"
      if (!groups[g]) groups[g] = []
      groups[g].push(row)
    }
    return groups
  }, [filteredData])

  const num = (v: unknown): number => {
    if (typeof v === "number") return v
    if (typeof v === "string") return parseFloat(v) || 0
    return 0
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <ClipboardListIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Requirements</h1>
            <p className="text-muted-foreground mt-1">Staffing requirements by section</p>
          </div>
        </div>
        <Button onClick={() => router.push("/hr/requirements/create")}>
          <PlusIcon className="mr-2 h-4 w-4" />
          Add Requirement
        </Button>
      </div>

      <div className="px-4 lg:px-6">
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="h-9 flex-nowrap">
            {sectionTabOptions.map((tab) => (
              <TabsTrigger key={tab} value={tab} className="h-8 px-4 text-sm">{tab}</TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </div>

      <div className="px-4 lg:px-6 space-y-6">
        {loading ? (
          <div className="flex items-center justify-center py-16 text-muted-foreground">
            <Loader2 className="h-6 w-6 animate-spin mr-2" />
            Loading...
          </div>
        ) : filteredData.length === 0 ? (
          <div className="flex items-center justify-center py-16 text-muted-foreground">
            No requirements found for this section
          </div>
        ) : (
          ["Staff", "Worker"].map((group) => {
            const rows = groupedData[group] || []
            if (rows.length === 0) return null
            const totalRequired = rows.reduce((s, r) => s + num(r.required), 0)
            const totalPresent = rows.reduce((s, r) => s + num(r.present), 0)
            const totalNet = rows.reduce((s, r) => s + num(r.net), 0)
            return (
              <div key={group} className="rounded-lg border bg-card overflow-hidden">
                <div className="px-4 py-3 border-b bg-muted/30 font-semibold text-base flex items-center gap-2">
                  {group}
                  <Badge variant="outline" className="text-xs">{rows.length}</Badge>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-left text-muted-foreground bg-muted/10">
                        <th className="py-3 px-4 font-semibold w-10">Sl</th>
                        <th className="py-3 px-4 font-semibold">Designation</th>
                        <th className="py-3 px-4 font-semibold text-right">Required</th>
                        <th className="py-3 px-4 font-semibold text-right">Present</th>
                        <th className="py-3 px-4 font-semibold text-right">Net</th>
                      </tr>
                    </thead>
                    <tbody>
                      {rows.map((row, i) => (
                        <tr key={i} className="border-b last:border-0 hover:bg-muted/20">
                          <td className="py-2.5 px-4 text-muted-foreground text-xs">{i + 1}</td>
                          <td className="py-2.5 px-4 font-medium">{row.designation}</td>
                          <td className="py-2.5 px-4 text-right">{num(row.required)}</td>
                          <td className="py-2.5 px-4 text-right">{num(row.present)}</td>
                          <td className="py-2.5 px-4 text-right">
                            <span className={
                              num(row.net) > 0 ? "text-destructive font-semibold" :
                              num(row.net) < 0 ? "text-green-600 font-semibold" : ""
                            }>
                              {num(row.net) > 0 ? `+${num(row.net)}` : num(row.net)}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                    <tfoot>
                      <tr className="border-t-2 border-muted bg-muted/40 font-semibold">
                        <td className="py-3 px-4" colSpan={1}></td>
                        <td className="py-3 px-4 text-base">Total</td>
                        <td className="py-3 px-4 text-right text-base">{num(totalRequired)}</td>
                        <td className="py-3 px-4 text-right text-base">{num(totalPresent)}</td>
                        <td className="py-3 px-4 text-right text-base">
                          <span className={
                            num(totalNet) > 0 ? "text-destructive" :
                            num(totalNet) < 0 ? "text-green-600" : ""
                          }>
                            {num(totalNet) > 0 ? `+${num(totalNet)}` : num(totalNet)}
                          </span>
                        </td>
                      </tr>
                    </tfoot>
                  </table>
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
