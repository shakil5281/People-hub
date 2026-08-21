"use client"

import * as React from "react"
import { format } from "date-fns"
import { DatabaseIcon, Loader2, DownloadIcon, Trash2Icon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { DatePicker } from "@/components/ui/date-picker"
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog"
import { dataLogApi } from "@/lib/api"
import { isSuperAdmin } from "@/lib/auth"

export default function LogCollectPage() {
  const [importing, setImporting] = React.useState(false)
  const [deleting, setDeleting] = React.useState(false)
  const [startDate, setStartDate] = React.useState<Date | undefined>(undefined)
  const [endDate, setEndDate] = React.useState<Date | undefined>(undefined)
  const [todayCount, setTodayCount] = React.useState(0)
  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false)
  const [mounted, setMounted] = React.useState(false)

  const fetchStats = async () => {
    try {
      const { data } = await dataLogApi.stats()
      setTodayCount(data.today_logs || 0)
    } catch {
      // ignore
    }
  }

  React.useEffect(() => {
    setMounted(true)
    setStartDate(new Date())
    setEndDate(new Date())
    fetchStats()
  }, [])

  const handleImport = async () => {
    setImporting(true)
    try {
      const body: Record<string, string> = {}
      if (startDate) body.start_date = format(startDate, "yyyy-MM-dd")
      if (endDate) body.end_date = format(endDate, "yyyy-MM-dd")
      const { data } = await dataLogApi.import(body)
      toast.success(`Imported ${data.imported || 0} records`)
      fetchStats()
    } catch (err: unknown) {
      const msg = err && typeof err === "object" && "response" in err
        ? (err as { response: { data?: { error?: string } } }).response?.data?.error
        : undefined
      toast.error(msg || "Failed to import data logs")
    } finally {
      setImporting(false)
    }
  }

  const handleDeleteAll = async () => {
    setDeleting(true)
    try {
      await dataLogApi.deleteAll()
      toast.success("All data logs deleted permanently")
      setTodayCount(0)
      setDeleteDialogOpen(false)
    } catch {
      toast.error("Failed to delete data logs")
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <DatabaseIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Log Collect</h1>
            <p className="text-muted-foreground mt-1">Import and view raw punch data from ZKTeco devices</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">Today&apos;s Collected Logs</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <p className="text-4xl font-bold">{todayCount}</p>
              {mounted && isSuperAdmin() && (
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
                          This action cannot be undone. This will permanently delete all data logs from the database.
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
            <CardTitle className="text-base">Import & Fetch Logs</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">Start Date</label>
                <DatePicker value={startDate} onChange={setStartDate} placeholder="Start date" />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">End Date</label>
                <DatePicker value={endDate} onChange={setEndDate} placeholder="End date" />
              </div>
              <div>
                <Button onClick={handleImport} disabled={importing}>
                  {importing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <DownloadIcon className="mr-2 h-4 w-4" />}
                  Import from MDB
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

    </div>
  )
}
