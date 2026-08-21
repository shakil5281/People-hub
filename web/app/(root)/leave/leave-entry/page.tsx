"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { Loader2, ArrowLeft, SearchIcon, UserIcon, CalendarDays, BriefcaseIcon, CalendarCheckIcon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { DatePicker } from "@/components/ui/date-picker"
import { format, differenceInDays, isBefore } from "date-fns"
import { leaveApi, employeeApi, leaveTypeApi, leaveBalanceApi } from "@/lib/api"

interface EmployeeInfo {
  id: string
  employee_id: string
  name_en: string
  name_bn?: string
  designation_ref?: { name: string }
  department?: { name: string }
  company_id: string
  joining_date: string
  status: string
}

interface LeaveType { id: string; name: string; total_days: number; code?: string }

interface BalanceRow {
  leave_type: string
  total_days: number
  used_days: number
  pending_days: number
  remaining_days: number
}

interface PreviousLeave {
  id: string
  leave_type?: { name: string }
  from_date: string
  to_date: string
  total_days: number
  status: string
  reason?: string
}

export default function LeaveEntryPage() {
  const router = useRouter()
  const [searchCode, setSearchCode] = React.useState("")
  const [empLookupLoading, setEmpLookupLoading] = React.useState(false)
  const [empNotFound, setEmpNotFound] = React.useState(false)
  const [emp, setEmp] = React.useState<EmployeeInfo | null>(null)

  const [leaveTypes, setLeaveTypes] = React.useState<LeaveType[]>([])
  const [balance, setBalance] = React.useState<BalanceRow[]>([])
  const [balanceLoading, setBalanceLoading] = React.useState(false)
  const [previousLeaves, setPreviousLeaves] = React.useState<PreviousLeave[]>([])
  const [prevLeavesLoading, setPrevLeavesLoading] = React.useState(false)

  const [leaveTypeId, setLeaveTypeId] = React.useState("")
  const [fromDate, setFromDate] = React.useState<Date | undefined>(undefined)
  const [toDate, setToDate] = React.useState<Date | undefined>(undefined)
  const [reason, setReason] = React.useState("")
  const [submitting, setSubmitting] = React.useState(false)

  const thisYear = new Date().getFullYear()

  const handleEmployeeLookup = React.useCallback(async () => {
    const code = searchCode.trim()
    if (!code) return
    setEmpLookupLoading(true)
    setEmpNotFound(false)
    setEmp(null)
    setBalance([])
    setPreviousLeaves([])
    setLeaveTypes([])
    setLeaveTypeId("")
    setFromDate(undefined)
    setToDate(undefined)
    setReason("")
    try {
      const { data } = await employeeApi.getByCode(code)
      const e = data as EmployeeInfo
      setEmp(e)
      setBalanceLoading(true)
      setPrevLeavesLoading(true)
      const [balRes, leavesRes, ltRes] = await Promise.all([
        leaveBalanceApi.list({ employee_id: e.employee_id, year: String(thisYear) }),
        leaveApi.list({ employee_id: e.employee_id, limit: "10", page: "1" }),
        leaveTypeApi.list(e.company_id, { limit: "50" }),
      ])
      setLeaveTypes(Array.isArray(ltRes.data?.data) ? ltRes.data.data : [])
      const rawBal = Array.isArray(balRes.data?.data) ? balRes.data.data : []
      setBalance(rawBal.map((r: any) => ({
        leave_type: r.leave_type || r.leave_type_name || "-",
        total_days: r.total_days ?? r.total ?? 0,
        used_days: r.used_days ?? r.used ?? 0,
        pending_days: r.pending_days ?? r.pending ?? 0,
        remaining_days: r.remaining_days ?? r.remaining ?? 0,
      })))
      const rawLeaves = Array.isArray(leavesRes.data?.data) ? leavesRes.data.data : []
      setPreviousLeaves(rawLeaves.slice(0, 10))
    } catch {
      setEmpNotFound(true)
      toast.error("Employee not found")
    } finally {
      setEmpLookupLoading(false)
      setBalanceLoading(false)
      setPrevLeavesLoading(false)
    }
  }, [searchCode, thisYear])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault()
      handleEmployeeLookup()
    }
  }

  const totalDays = fromDate && toDate
    ? differenceInDays(toDate, fromDate) + 1
    : 0

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!emp || !leaveTypeId || !fromDate || !toDate) {
      toast.error("Please fill all required fields")
      return
    }
    if (isBefore(toDate, fromDate)) {
      toast.error("To date must be after or equal to from date")
      return
    }
    setSubmitting(true)
    try {
      await leaveApi.apply({
        company_id: emp.company_id,
        employee_id: emp.employee_id,
        leave_type_id: leaveTypeId,
        from_date: format(fromDate, "yyyy-MM-dd"),
        to_date: format(toDate, "yyyy-MM-dd"),
        reason: reason || undefined,
      })
      toast.success("Leave applied successfully")
      router.push("/leave/leave")
    } catch {
      toast.error("Failed to apply leave")
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => router.back()}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Leave Entry</h1>
          <p className="text-muted-foreground mt-1">Apply for employee leave</p>
        </div>
      </div>

      <div className="px-4 lg:px-6 max-w-4xl space-y-6">
        {/* Employee Search */}
        <Card>
          <CardHeader><CardTitle className="flex items-center gap-2 text-lg"><SearchIcon className="h-5 w-5" /> Employee Lookup</CardTitle></CardHeader>
          <CardContent>
            <div className="flex items-end gap-3">
              <div className="flex-1 space-y-2">
                <Label htmlFor="emp_search">Employee ID / Punch Number</Label>
                <Input
                  id="emp_search"
                  value={searchCode}
                  onChange={(e) => { setSearchCode(e.target.value); setEmpNotFound(false) }}
                  onKeyDown={handleKeyDown}
                  placeholder="Type employee ID / punch number and press Enter"
                />
                {empNotFound && <p className="text-sm text-destructive">Employee not found with this ID</p>}
              </div>
              <Button type="button" onClick={handleEmployeeLookup} disabled={empLookupLoading || !searchCode.trim()}>
                {empLookupLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <SearchIcon className="h-4 w-4" />}
                Search
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Employee Information */}
        {emp && (
          <Card>
            <CardHeader><CardTitle className="flex items-center gap-2 text-lg"><UserIcon className="h-5 w-5" /> Employee Information</CardTitle></CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <div>
                  <Label className="text-xs text-muted-foreground">Employee Name</Label>
                  <p className="font-medium text-sm mt-1">{emp.name_en || emp.name_bn || "-"}</p>
                </div>
                <div>
                  <Label className="text-xs text-muted-foreground">Designation</Label>
                  <div className="flex items-center gap-1.5 mt-1">
                    <BriefcaseIcon className="h-3.5 w-3.5 text-muted-foreground" />
                    <p className="font-medium text-sm">{emp.designation_ref?.name || "-"}</p>
                  </div>
                </div>
                <div>
                  <Label className="text-xs text-muted-foreground">Joining Date</Label>
                  <div className="flex items-center gap-1.5 mt-1">
                    <CalendarDays className="h-3.5 w-3.5 text-muted-foreground" />
                    <p className="font-medium text-sm">{emp.joining_date ? format(new Date(emp.joining_date), "dd-MM-yyyy") : "-"}</p>
                  </div>
                </div>
                <div>
                  <Label className="text-xs text-muted-foreground">Employee ID</Label>
                  <p className="font-medium text-sm mt-1">{emp.employee_id}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Leave Balance */}
        {emp && (
          <Card>
            <CardHeader><CardTitle className="flex items-center gap-2 text-lg"><CalendarCheckIcon className="h-5 w-5" /> Leave Balance ({thisYear})</CardTitle></CardHeader>
            <CardContent>
              {balanceLoading ? (
                <div className="flex items-center gap-2 text-sm text-muted-foreground"><Loader2 className="h-4 w-4 animate-spin" /> Loading balance...</div>
              ) : balance.length === 0 ? (
                <p className="text-sm text-muted-foreground">No leave allocation found for this year</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-left text-muted-foreground">
                        <th className="pb-2 font-medium">Leave Type</th>
                        <th className="pb-2 font-medium text-right">Total</th>
                        <th className="pb-2 font-medium text-right">Used</th>
                        <th className="pb-2 font-medium text-right">Pending</th>
                        <th className="pb-2 font-medium text-right">Remaining</th>
                      </tr>
                    </thead>
                    <tbody>
                      {balance.map((row, i) => (
                        <tr key={i} className="border-b last:border-0">
                          <td className="py-2">{row.leave_type}</td>
                          <td className="py-2 text-right">{row.total_days}</td>
                          <td className="py-2 text-right">{row.used_days}</td>
                          <td className="py-2 text-right">{row.pending_days}</td>
                          <td className="py-2 text-right">
                            <Badge variant={row.remaining_days <= 0 ? "destructive" : row.remaining_days <= 2 ? "secondary" : "default"}>
                              {row.remaining_days}
                            </Badge>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Previous Leaves */}
        {emp && (
          <Card>
            <CardHeader><CardTitle className="flex items-center gap-2 text-lg"><CalendarDays className="h-5 w-5" /> Previous Leave Applications</CardTitle></CardHeader>
            <CardContent>
              {prevLeavesLoading ? (
                <div className="flex items-center gap-2 text-sm text-muted-foreground"><Loader2 className="h-4 w-4 animate-spin" /> Loading previous leaves...</div>
              ) : previousLeaves.length === 0 ? (
                <p className="text-sm text-muted-foreground">No previous leave applications found</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-left text-muted-foreground">
                        <th className="pb-2 font-medium">#</th>
                        <th className="pb-2 font-medium">Leave Type</th>
                        <th className="pb-2 font-medium">From</th>
                        <th className="pb-2 font-medium">To</th>
                        <th className="pb-2 font-medium text-right">Days</th>
                        <th className="pb-2 font-medium">Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      {previousLeaves.map((lv, i) => (
                        <tr key={lv.id || i} className="border-b last:border-0">
                          <td className="py-2">{i + 1}</td>
                          <td className="py-2">{lv.leave_type?.name || "-"}</td>
                          <td className="py-2">{lv.from_date ? format(new Date(lv.from_date), "dd-MM-yyyy") : "-"}</td>
                          <td className="py-2">{lv.to_date ? format(new Date(lv.to_date), "dd-MM-yyyy") : "-"}</td>
                          <td className="py-2 text-right">{lv.total_days}</td>
                          <td className="py-2">
                            <Badge variant={
                              lv.status === "approved" ? "default" :
                              lv.status === "pending" ? "secondary" :
                              lv.status === "rejected" ? "destructive" : "outline"
                            }>
                              {lv.status}
                            </Badge>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Leave Application Form */}
        {emp && (
          <form onSubmit={handleSubmit}>
            <Card>
              <CardHeader><CardTitle className="flex items-center gap-2 text-lg"><CalendarDays className="h-5 w-5" /> Leave Application</CardTitle></CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Company</Label>
                  <Input value={emp.company_id} disabled className="text-muted-foreground" />
                  <p className="text-xs text-muted-foreground">Auto-filled from employee record</p>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="leave_type_id">Leave Type *</Label>
                  <select
                    id="leave_type_id"
                    value={leaveTypeId}
                    onChange={(e) => setLeaveTypeId(e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  >
                    <option value="">Select leave type</option>
                    {leaveTypes.map((l) => (
                      <option key={l.id} value={l.id}>{l.name} ({l.total_days} days)</option>
                    ))}
                  </select>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>From Date *</Label>
                    <DatePicker value={fromDate} onChange={setFromDate} />
                  </div>
                  <div className="space-y-2">
                    <Label>To Date *</Label>
                    <DatePicker value={toDate} onChange={setToDate} />
                  </div>
                </div>

                {totalDays > 0 && (
                  <p className="text-sm text-muted-foreground">
                    Total days: <span className="font-medium text-foreground">{totalDays}</span>
                  </p>
                )}

                <div className="space-y-2">
                  <Label htmlFor="reason">Reason</Label>
                  <Textarea id="reason" value={reason} onChange={(e) => setReason(e.target.value)} rows={3} placeholder="Optional reason..." />
                </div>

                <div className="flex gap-4 pt-4 border-t">
                  <Button type="button" variant="outline" onClick={() => router.back()}>Cancel</Button>
                  <Button type="submit" disabled={submitting || !leaveTypeId || !fromDate || !toDate}>
                    {submitting ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Submitting...</> : "Apply Leave"}
                  </Button>
                </div>
              </CardContent>
            </Card>
          </form>
        )}
      </div>
    </div>
  )
}
