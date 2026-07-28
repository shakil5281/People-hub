"use client"

import * as React from "react"
import { format } from "date-fns"
import { RefreshCwIcon, Loader2, Trash2Icon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { DatePicker } from "@/components/ui/date-picker"
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog"
import { dataLogApi, attendanceApi, companyApi } from "@/lib/api"
import type { Company } from "@/components/data/company-data"
import { isSuperAdmin } from "@/lib/auth"

export default function DailyProcessPage() {
  const [processing, setProcessing] = React.useState(false)
  const [deleting, setDeleting] = React.useState(false)
  const [startDate, setStartDate] = React.useState<Date | undefined>(new Date())
  const [endDate, setEndDate] = React.useState<Date | undefined>(new Date())
  const [companyId, setCompanyId] = React.useState("")
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [todayCount, setTodayCount] = React.useState(0)
  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false)

  const fetchStats = async () => {
    try {
      const { data } = await attendanceApi.stats()
      setTodayCount(data.today_count || 0)
    } catch {
      // ignore
    }
  }

  const fetchCompanies = async () => {
    try {
      const { data } = await companyApi.list({ limit: "100" })
      setCompanies(Array.isArray(data?.data) ? data.data : (Array.isArray(data) ? data : []))
    } catch {
      // ignore
    }
  }

  React.useEffect(() => {
    fetchStats()
    fetchCompanies()
  }, [])

  const handleProcess = async () => {
    if (!companyId) {
      toast.error("Please select a company")
      return
    }
    setProcessing(true)
    try {
      const body: { company_id: string; start_date?: string; end_date?: string } = { company_id: companyId }
      if (startDate) body.start_date = format(startDate, "yyyy-MM-dd")
      if (endDate) body.end_date = format(endDate, "yyyy-MM-dd")
      const { data } = await dataLogApi.process(body)
      toast.success(data.message || "Data processed successfully")
      fetchStats()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to process data"
      if (typeof err === "object" && err !== null && "response" in err) {
        const axiosErr = err as { response?: { data?: { error?: string } } }
        toast.error(axiosErr.response?.data?.error || message)
      } else {
        toast.error(message)
      }
    } finally {
      setProcessing(false)
    }
  }

  const handleDeleteAll = async () => {
    setDeleting(true)
    try {
      await attendanceApi.deleteAll()
      toast.success("All attendance records deleted permanently")
      setTodayCount(0)
      setDeleteDialogOpen(false)
    } catch {
      toast.error("Failed to delete attendance records")
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <RefreshCwIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Daily Process</h1>
            <p className="text-muted-foreground mt-1">Process raw punch data into attendance records</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">Today&apos;s Attendance</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <p className="text-4xl font-bold">{todayCount}</p>
              {isSuperAdmin() && (
                <>
                  <Button variant="destructive" size="sm" onClick={() => setDeleteDialogOpen(true)}>
                    <Trash2Icon className="mr-2 h-4 w-4" />
                    Delete All
                  </Button>
                  <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                        <AlertDialogDescription>
                          This action cannot be undone. This will permanently delete all attendance records from the database.
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>Cancel</AlertDialogCancel>
                        <AlertDialogAction onClick={handleDeleteAll} disabled={deleting}>
                          {deleting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                          Delete
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">Process Parameters</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">Company</label>
                <select
                  value={companyId}
                  onChange={(e) => setCompanyId(e.target.value)}
                  className="flex h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
                >
                  <option value="">Select company</option>
                  {companies.map((c) => (
                    <option key={c.id} value={c.id}>{c.company_name_en}</option>
                  ))}
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">Start Date</label>
                <DatePicker value={startDate} onChange={setStartDate} placeholder="Start date" />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">End Date</label>
                <DatePicker value={endDate} onChange={setEndDate} placeholder="End date" />
              </div>
              <div>
                <Button onClick={handleProcess} disabled={processing}>
                  {processing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCwIcon className="mr-2 h-4 w-4" />}
                  Process Data
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

    </div>
  )
}
