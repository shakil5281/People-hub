"use client"

import * as React from "react"
import { CalendarDaysIcon, PlusIcon, Loader2, SunIcon, MoonIcon } from "lucide-react"
import { format } from "date-fns"
import { toast } from "sonner"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { DatePicker } from "@/components/ui/date-picker"
import { holidayApi, companyApi } from "@/lib/api"

interface Holiday {
  id: string
  company_id: string
  name: string
  date: string
  from_date: string | null
  to_date: string | null
  weekend_date: string | null
  type: string
  description: string
  status: string
}

interface Company {
  id: string
  company_name_en: string
}

const typeColors: Record<string, "default" | "secondary"> = {
  government: "default",
  weekend_change: "secondary",
}

const typeLabels: Record<string, string> = {
  government: "Government",
  weekend_change: "Weekend Change",
}

function formatDate(val: string | null | undefined) {
  if (!val) return "-"
  const d = val.includes("T") ? val.split("T")[0] : val
  const parts = d.split("-")
  if (parts.length !== 3) return val
  return `${parts[2]}-${parts[1]}-${parts[0]}`
}

const govColumns: ColumnDef<Holiday>[] = [
  { id: "sl", header: "#", cell: ({ row }) => row.index + 1 },
  { accessorKey: "name", header: "Holiday Name" },
  {
    id: "date_range",
    header: "Date",
    cell: ({ row }) => {
      const h = row.original
      if (h.from_date && h.to_date) {
        return `${formatDate(h.from_date)} to ${formatDate(h.to_date)}`
      }
      return formatDate(h.date)
    },
  },
  { accessorKey: "description", header: "Description", cell: ({ row }) => row.original.description || "-" },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => (
      <Badge variant={row.original.status === "active" ? "default" : "secondary"} className="capitalize">
        {row.original.status}
      </Badge>
    ),
  },
]

const wcColumns: ColumnDef<Holiday>[] = [
  { id: "sl", header: "#", cell: ({ row }) => row.index + 1 },
  { accessorKey: "name", header: "Change Name" },
  { accessorKey: "date", header: "General Duty Date", cell: ({ row }) => formatDate(row.original.date) },
  {
    accessorKey: "weekend_date",
    header: "Weekend Date",
    cell: ({ row }) => formatDate(row.original.weekend_date),
  },
  { accessorKey: "description", header: "Description", cell: ({ row }) => row.original.description || "-" },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => (
      <Badge variant={row.original.status === "active" ? "default" : "secondary"} className="capitalize">
        {row.original.status}
      </Badge>
    ),
  },
]

function GovHolidayDialog({
  open,
  onOpenChange,
  editingItem,
  onSuccess,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
  editingItem: Holiday | null
  onSuccess: () => void
}) {
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [companyId, setCompanyId] = React.useState("")
  const [name, setName] = React.useState("")
  const [date, setDate] = React.useState<Date | undefined>()
  const [fromDate, setFromDate] = React.useState<Date | undefined>()
  const [toDate, setToDate] = React.useState<Date | undefined>()
  const [description, setDescription] = React.useState("")
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [error, setError] = React.useState("")
  const [useRange, setUseRange] = React.useState(false)

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((res) => {
      const list = Array.isArray(res.data?.data) ? res.data.data : Array.isArray(res.data) ? res.data : []
      setCompanies(list)
    }).catch(() => {})
  }, [])

  React.useEffect(() => {
    if (editingItem) {
      setName(editingItem.name)
      setDate(editingItem.date ? new Date(editingItem.date + "T00:00:00") : undefined)
      setFromDate(editingItem.from_date ? new Date(editingItem.from_date + "T00:00:00") : undefined)
      setToDate(editingItem.to_date ? new Date(editingItem.to_date + "T00:00:00") : undefined)
      setUseRange(!!(editingItem.from_date && editingItem.to_date))
      setDescription(editingItem.description || "")
      setCompanyId(editingItem.company_id)
    } else {
      setName("")
      setDate(undefined)
      setFromDate(undefined)
      setToDate(undefined)
      setUseRange(false)
      setDescription("")
      setCompanyId(companies.length === 1 ? companies[0].id : "")
    }
    setError("")
  }, [editingItem, open, companies])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name || !companyId) {
      setError("Name and company are required")
      return
    }
    if (!useRange && !date) {
      setError("Date is required")
      return
    }
    if (useRange && (!fromDate || !toDate)) {
      setError("From date and to date are required")
      return
    }
    setIsSubmitting(true)
    setError("")
    try {
      const payload: Record<string, unknown> = {
        name,
        description,
        type: "government",
      }
      if (useRange && fromDate && toDate) {
        payload.date = format(fromDate, "yyyy-MM-dd")
        payload.from_date = format(fromDate, "yyyy-MM-dd")
        payload.to_date = format(toDate, "yyyy-MM-dd")
      } else if (date) {
        payload.date = format(date, "yyyy-MM-dd")
      }
      if (editingItem) {
        await holidayApi.update(editingItem.id, payload)
        toast.success("Holiday updated successfully")
      } else {
        payload.company_id = companyId
        await holidayApi.create(payload)
        toast.success("Holiday created successfully")
      }
      onOpenChange(false)
      onSuccess()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to save"
      let detail = message
      if (typeof err === "object" && err !== null && "response" in err) {
        const axiosErr = err as { response?: { data?: { error?: string } } }
        detail = axiosErr.response?.data?.error || message
      }
      setError(detail)
      toast.error(detail)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <SunIcon className="h-5 w-5 text-muted-foreground" />
            {editingItem ? "Edit Government Holiday" : "Add Government Holiday"}
          </DialogTitle>
          <DialogDescription>
            {editingItem ? "Update the holiday details." : "Add a government-declared public holiday."}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
          )}

          {!editingItem && (
            <div className="space-y-2">
              <Label htmlFor="company_id">Company *</Label>
              <select
                id="company_id"
                value={companyId}
                onChange={(e) => setCompanyId(e.target.value)}
                className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              >
                <option value="">Select company</option>
                {companies.map((c) => (
                  <option key={c.id} value={c.id}>{c.company_name_en}</option>
                ))}
              </select>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="name">Holiday Name *</Label>
            <Input
              id="name"
              placeholder="e.g. Independence Day"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div className="flex items-center gap-2 mb-2">
            <input
              id="use-range"
              type="checkbox"
              checked={useRange}
              onChange={(e) => setUseRange(e.target.checked)}
              className="h-4 w-4 rounded border-gray-300"
            />
            <Label htmlFor="use-range" className="text-sm font-normal">Date range (multi-day holiday)</Label>
          </div>

          {useRange ? (
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>From Date *</Label>
                <DatePicker
                  value={fromDate}
                  onChange={(d) => setFromDate(d)}
                  placeholder="Start date"
                />
              </div>
              <div className="space-y-2">
                <Label>To Date *</Label>
                <DatePicker
                  value={toDate}
                  onChange={(d) => setToDate(d)}
                  placeholder="End date"
                />
              </div>
            </div>
          ) : (
            <div className="space-y-2">
              <Label>Date *</Label>
              <DatePicker
                value={date}
                onChange={(d) => setDate(d)}
                placeholder="Select holiday date"
              />
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Input
              id="description"
              placeholder="Optional description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <div className="flex justify-end gap-4 pt-4 border-t">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? (
                <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</>
              ) : (
                editingItem ? "Update" : "Add Holiday"
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function WcHolidayDialog({
  open,
  onOpenChange,
  editingItem,
  onSuccess,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
  editingItem: Holiday | null
  onSuccess: () => void
}) {
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [companyId, setCompanyId] = React.useState("")
  const [name, setName] = React.useState("")
  const [generalDutyDate, setGeneralDutyDate] = React.useState<Date | undefined>()
  const [weekendDate, setWeekendDate] = React.useState<Date | undefined>()
  const [description, setDescription] = React.useState("")
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [error, setError] = React.useState("")

  React.useEffect(() => {
    companyApi.list({ limit: "100" }).then((res) => {
      const list = Array.isArray(res.data?.data) ? res.data.data : Array.isArray(res.data) ? res.data : []
      setCompanies(list)
    }).catch(() => {})
  }, [])

  React.useEffect(() => {
    if (editingItem) {
      setName(editingItem.name)
      setGeneralDutyDate(editingItem.date ? new Date(editingItem.date + "T00:00:00") : undefined)
      setWeekendDate(editingItem.weekend_date ? new Date(editingItem.weekend_date + "T00:00:00") : undefined)
      setDescription(editingItem.description || "")
      setCompanyId(editingItem.company_id)
    } else {
      setName("")
      setGeneralDutyDate(undefined)
      setWeekendDate(undefined)
      setDescription("")
      setCompanyId(companies.length === 1 ? companies[0].id : "")
    }
    setError("")
  }, [editingItem, open, companies])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name || !generalDutyDate || !weekendDate || !companyId) {
      setError("Name, general duty date, weekend date, and company are required")
      return
    }
    setIsSubmitting(true)
    setError("")
    try {
      const dateStr = format(generalDutyDate, "yyyy-MM-dd")
      const weekendStr = format(weekendDate, "yyyy-MM-dd")
      if (editingItem) {
        await holidayApi.update(editingItem.id, {
          name,
          date: dateStr,
          weekend_date: weekendStr,
          description,
          type: "weekend_change",
        })
        toast.success("Weekend change updated successfully")
      } else {
        await holidayApi.create({
          name,
          date: dateStr,
          weekend_date: weekendStr,
          description,
          type: "weekend_change",
          company_id: companyId,
        })
        toast.success("Weekend change created successfully")
      }
      onOpenChange(false)
      onSuccess()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to save"
      let detail = message
      if (typeof err === "object" && err !== null && "response" in err) {
        const axiosErr = err as { response?: { data?: { error?: string } } }
        detail = axiosErr.response?.data?.error || message
      }
      setError(detail)
      toast.error(detail)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <MoonIcon className="h-5 w-5 text-muted-foreground" />
            {editingItem ? "Edit Weekend Change" : "Add Weekend Change"}
          </DialogTitle>
          <DialogDescription>
            {editingItem
              ? "Update the weekend change details."
              : "Change a weekend day to general duty. Select the general duty date and the weekend date being replaced."}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
          )}

          {!editingItem && (
            <div className="space-y-2">
              <Label htmlFor="company_id">Company *</Label>
              <select
                id="company_id"
                value={companyId}
                onChange={(e) => setCompanyId(e.target.value)}
                className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              >
                <option value="">Select company</option>
                {companies.map((c) => (
                  <option key={c.id} value={c.id}>{c.company_name_en}</option>
                ))}
              </select>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="name">Change Name *</Label>
            <Input
              id="name"
              placeholder="e.g. Friday General Duty"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>General Duty Date *</Label>
              <DatePicker
                value={generalDutyDate}
                onChange={(d) => setGeneralDutyDate(d)}
                placeholder="Select date"
              />
              <p className="text-xs text-muted-foreground">The date employees will work</p>
            </div>

            <div className="space-y-2">
              <Label>Weekend Date *</Label>
              <DatePicker
                value={weekendDate}
                onChange={(d) => setWeekendDate(d)}
                placeholder="Select weekend"
              />
              <p className="text-xs text-muted-foreground">The weekend day being replaced</p>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Input
              id="description"
              placeholder="e.g. Friday general duty for production target"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <div className="flex justify-end gap-4 pt-4 border-t">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? (
                <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</>
              ) : (
                editingItem ? "Update" : "Add Change"
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default function HolidayPage() {
  const [allData, setAllData] = React.useState<Holiday[]>([])
  const [loading, setLoading] = React.useState(true)

  const [govDialogOpen, setGovDialogOpen] = React.useState(false)
  const [wcDialogOpen, setWcDialogOpen] = React.useState(false)
  const [editingItem, setEditingItem] = React.useState<Holiday | null>(null)
  const [editingType, setEditingType] = React.useState<string>("")

  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)

  const fetchData = React.useCallback(async () => {
    setLoading(true)
    try {
      const { data: res } = await holidayApi.list({ page: String(page), limit: String(limit) })
      setAllData(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setAllData([])
    } finally {
      setLoading(false)
    }
  }, [page, limit])

  React.useEffect(() => { fetchData() }, [fetchData])

  const govData = React.useMemo(() => allData.filter((h) => h.type === "government"), [allData])
  const wcData = React.useMemo(() => allData.filter((h) => h.type === "weekend_change"), [allData])

  const openAdd = (type: string) => {
    setEditingItem(null)
    setEditingType(type)
    if (type === "government") setGovDialogOpen(true)
    else setWcDialogOpen(true)
  }

  const handleEdit = (item: Holiday) => {
    setEditingItem(item)
    setEditingType(item.type)
    if (item.type === "government") setGovDialogOpen(true)
    else setWcDialogOpen(true)
  }

  const handleDelete = async (item: Holiday) => {
    try {
      await holidayApi.delete(item.id)
      setAllData((prev) => prev.filter((d) => d.id !== item.id))
      toast.success("Holiday deleted successfully")
    } catch {
      toast.error("Failed to delete holiday")
    }
  }

  const sharedTableProps = {
    serverSide: true as const,
    page,
    pageSize: limit,
    pageCount: totalPages,
    total,
    onPageChange: setPage,
    onPageSizeChange: (size: number) => { setLimit(size); setPage(1) },
    loading,
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <CalendarDaysIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Holiday</h1>
          <p className="text-muted-foreground mt-1">Manage government holidays and weekend change / general duty</p>
        </div>
      </div>

      <Tabs defaultValue="government" className="px-4 lg:px-6">
        <TabsList>
          <TabsTrigger value="government" className="flex items-center gap-2">
            <SunIcon className="h-4 w-4" /> Government Holiday
          </TabsTrigger>
          <TabsTrigger value="weekend_change" className="flex items-center gap-2">
            <MoonIcon className="h-4 w-4" /> Change General Duty by Weekend
          </TabsTrigger>
        </TabsList>

        <TabsContent value="government" className="mt-4 space-y-4">
          <div className="flex justify-end">
            <Button onClick={() => openAdd("government")}>
              <PlusIcon className="mr-2 h-4 w-4" />
              Add Government Holiday
            </Button>
          </div>
          <DataTable
            data={govData}
            columns={govColumns}
            onEdit={handleEdit}
            onDelete={handleDelete}
            {...sharedTableProps}
          />
        </TabsContent>

        <TabsContent value="weekend_change" className="mt-4 space-y-4">
          <div className="flex justify-end">
            <Button onClick={() => openAdd("weekend_change")}>
              <PlusIcon className="mr-2 h-4 w-4" />
              Add Weekend Change
            </Button>
          </div>
          <DataTable
            data={wcData}
            columns={wcColumns}
            onEdit={handleEdit}
            onDelete={handleDelete}
            {...sharedTableProps}
          />
        </TabsContent>
      </Tabs>

      <GovHolidayDialog
        open={govDialogOpen}
        onOpenChange={setGovDialogOpen}
        editingItem={editingType === "government" ? editingItem : null}
        onSuccess={fetchData}
      />

      <WcHolidayDialog
        open={wcDialogOpen}
        onOpenChange={setWcDialogOpen}
        editingItem={editingType === "weekend_change" ? editingItem : null}
        onSuccess={fetchData}
      />
    </div>
  )
}
