"use client"

import * as React from "react"
import {
  RefreshCwIcon,
  DatabaseIcon,
  CheckCircle2Icon,
  XCircleIcon,
  AlertCircleIcon,
  Loader2Icon,
  SearchIcon,
  HardDriveIcon,
  FilterIcon,
} from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Checkbox } from "@/components/ui/checkbox"
import { zktecoSyncApi } from "@/lib/api"

interface EmployeeSyncItem {
  id: string
  employee_id: string
  punch_number: string
  name: string
  department: string
  designation: string
  is_synched: boolean
}

export default function ZKTecoSyncPage() {
  const [mdbPath, setMdbPath] = React.useState(`C:\\Program Files (x86)\\ZKTeco\\att2000.mdb`)
  const [employees, setEmployees] = React.useState<EmployeeSyncItem[]>([])
  const [loading, setLoading] = React.useState(false)
  const [syncing, setSyncing] = React.useState(false)
  const [testingConn, setTestingConn] = React.useState(false)
  const [connStatus, setConnStatus] = React.useState<{ connected: boolean; message: string } | null>(null)
  const [selectedIds, setSelectedIds] = React.useState<string[]>([])
  const [search, setSearch] = React.useState("")
  const [filterStatus, setFilterStatus] = React.useState<"all" | "synced" | "unsynced">("all")
  const [syncedCount, setSyncedCount] = React.useState(0)
  const [mdbExists, setMdbExists] = React.useState(false)

  const fetchStatus = React.useCallback(async () => {
    setLoading(true)
    try {
      const { data: res } = await zktecoSyncApi.getStatus({ mdb_path: mdbPath })
      setEmployees(res.data || [])
      setSyncedCount(res.synced_count || 0)
      setMdbExists(res.mdb_exists || false)
    } catch {
      toast.error("Failed to fetch employee sync status")
    } finally {
      setLoading(false)
    }
  }, [mdbPath])

  React.useEffect(() => {
    fetchStatus()
  }, [fetchStatus])

  const handleTestConnection = async () => {
    setTestingConn(true)
    setConnStatus(null)
    try {
      const { data: res } = await zktecoSyncApi.testConnection({ mdb_path: mdbPath })
      setConnStatus({ connected: res.connected, message: res.message })
      if (res.connected) {
        toast.success(res.message)
      } else {
        toast.error(res.message)
      }
    } catch {
      setConnStatus({ connected: false, message: "Connection request failed" })
      toast.error("Connection test failed")
    } finally {
      setTestingConn(false)
    }
  }

  const handleSyncSelected = async () => {
    if (selectedIds.length === 0) {
      toast.error("Please select at least one employee to sync")
      return
    }
    setSyncing(true)
    try {
      const { data: res } = await zktecoSyncApi.sync({
        mdb_path: mdbPath,
        employee_ids: selectedIds,
        sync_all: false,
      })
      if (res.success) {
        toast.success(res.message || `Successfully synced ${res.synced_count} employees`)
      } else {
        toast.warning(res.message || `Synced ${res.synced_count} employees, ${res.failed_count} failed`)
      }
      setSelectedIds([])
      // Use real-time status data from sync response if available
      if ((res as any).status_data) {
        setEmployees((res as any).status_data)
        setSyncedCount((res as any).status_synced_count || 0)
        setMdbExists((res as any).mdb_exists ?? mdbExists)
      } else {
        fetchStatus()
      }
    } catch {
      toast.error("Failed to execute sync operation")
    } finally {
      setSyncing(false)
    }
  }

  const handleSyncAll = async () => {
    setSyncing(true)
    try {
      const { data: res } = await zktecoSyncApi.sync({
        mdb_path: mdbPath,
        sync_all: true,
      })
      if (res.success) {
        toast.success(res.message || `Successfully synced ${res.synced_count} employees`)
      } else {
        toast.warning(res.message || `Synced ${res.synced_count} employees, ${res.failed_count} failed`)
      }
      setSelectedIds([])
      // Use real-time status data from sync response if available
      if ((res as any).status_data) {
        setEmployees((res as any).status_data)
        setSyncedCount((res as any).status_synced_count || 0)
        setMdbExists((res as any).mdb_exists ?? mdbExists)
      } else {
        fetchStatus()
      }
    } catch {
      toast.error("Failed to execute bulk sync")
    } finally {
      setSyncing(false)
    }
  }

  const handleSyncSingle = async (emp: EmployeeSyncItem) => {
    setSyncing(true)
    try {
      const { data: res } = await zktecoSyncApi.sync({
        mdb_path: mdbPath,
        employee_ids: [emp.id],
        sync_all: false,
      })
      if (res.synced_count > 0) {
        toast.success(`Successfully synced ${emp.name} (${emp.punch_number}) to MDB`)
      } else {
        toast.error(`Failed to sync ${emp.name}`)
      }
      // Use real-time status data from sync response if available
      if ((res as any).status_data) {
        setEmployees((res as any).status_data)
        setSyncedCount((res as any).status_synced_count || 0)
        setMdbExists((res as any).mdb_exists ?? mdbExists)
      } else {
        fetchStatus()
      }
    } catch {
      toast.error("Failed to sync employee")
    } finally {
      setSyncing(false)
    }
  }

  const filteredEmployees = React.useMemo(() => {
    return employees.filter((emp) => {
      const matchesSearch =
        search === "" ||
        emp.name.toLowerCase().includes(search.toLowerCase()) ||
        emp.employee_id.toLowerCase().includes(search.toLowerCase()) ||
        emp.punch_number.toLowerCase().includes(search.toLowerCase()) ||
        (emp.department && emp.department.toLowerCase().includes(search.toLowerCase())) ||
        (emp.designation && emp.designation.toLowerCase().includes(search.toLowerCase()))

      const matchesStatus =
        filterStatus === "all" ||
        (filterStatus === "synced" && emp.is_synched) ||
        (filterStatus === "unsynced" && !emp.is_synched)

      return matchesSearch && matchesStatus
    })
  }, [employees, search, filterStatus])

  const toggleSelectAll = () => {
    if (selectedIds.length === filteredEmployees.length) {
      setSelectedIds([])
    } else {
      setSelectedIds(filteredEmployees.map((e) => e.id))
    }
  }

  const toggleSelectOne = (id: string) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id]
    )
  }

  return (
    <div className="flex flex-col gap-6 py-6 px-4 lg:px-8 max-w-7xl mx-auto w-full">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b pb-5">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-lg bg-primary/10 text-primary">
              <RefreshCwIcon className="h-6 w-6" />
            </div>
            <div>
              <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">ZKTeco Database Sync</h1>
              <p className="text-sm text-muted-foreground mt-0.5">
                Sync submitted employee names and punch numbers into ZKTeco <code className="font-mono text-xs bg-muted px-1.5 py-0.5 rounded">att2000.mdb</code> database (<span className="font-medium">USERINFO</span> table).
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={fetchStatus} disabled={loading}>
            <RefreshCwIcon className={`h-4 w-4 mr-2 ${loading ? "animate-spin" : ""}`} />
            Refresh Status
          </Button>
        </div>
      </div>

      {/* MDB Configuration Card */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base flex items-center gap-2">
            <HardDriveIcon className="h-4 w-4 text-primary" />
            ZKTeco Access Database File Path
          </CardTitle>
          <CardDescription>
            Specify the location of the <code className="font-mono text-xs bg-muted px-1.5 py-0.5 rounded">att2000.mdb</code> database file on your server or local computer.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1">
              <DatabaseIcon className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                value={mdbPath}
                onChange={(e) => setMdbPath(e.target.value)}
                placeholder="C:\att2000.mdb or C:\Program Files (x86)\ZKTeco\att2000.mdb"
                className="pl-9 font-mono text-sm"
              />
            </div>
            <Button
              variant="secondary"
              onClick={handleTestConnection}
              disabled={testingConn || !mdbPath.trim()}
              className="sm:w-auto"
            >
              {testingConn ? <Loader2Icon className="h-4 w-4 mr-2 animate-spin" /> : <CheckCircle2Icon className="h-4 w-4 mr-2" />}
              Test Connection
            </Button>
          </div>

          {connStatus && (
            <div className={`p-3 rounded-lg border text-sm flex items-start gap-2.5 ${
              connStatus.connected ? "bg-emerald-50/50 border-emerald-200 text-emerald-900 dark:bg-emerald-950/20 dark:border-emerald-900 dark:text-emerald-300" : "bg-destructive/10 border-destructive/20 text-destructive"
            }`}>
              {connStatus.connected ? (
                <CheckCircle2Icon className="h-5 w-5 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5" />
              ) : (
                <XCircleIcon className="h-5 w-5 text-destructive shrink-0 mt-0.5" />
              )}
              <div className="flex-1">
                <span className="font-semibold">{connStatus.connected ? "Connection Successful" : "Connection Failed"}</span>
                <p className="text-xs mt-0.5 opacity-90">{connStatus.message}</p>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Important Lock Banner Callout */}
      <div className="p-4 rounded-xl border bg-amber-500/10 border-amber-500/20 text-amber-900 dark:text-amber-200 flex items-start gap-3 text-sm">
        <AlertCircleIcon className="h-5 w-5 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
        <div className="space-y-1">
          <p className="font-semibold text-amber-900 dark:text-amber-200">
            Important Notice Before Syncing to ZKTeco Database:
          </p>
          <p className="text-xs text-amber-800 dark:text-amber-300 leading-relaxed">
            If <strong>ZKTeco Software (Att.exe)</strong> or <strong>Microsoft Access</strong> is open on your desktop, Windows locks <code className="font-mono bg-amber-200/50 dark:bg-amber-900/50 px-1 py-0.5 rounded">att2000.mdb</code> from receiving updates. 
            <br />
            <strong>To sync successfully:</strong> Please close ZKTeco software (Att.exe) and MS Access on your computer, then click <strong>Sync All Employees</strong> below.
          </p>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card className="bg-card">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-xs font-medium text-muted-foreground">Total Active Employees</p>
              <h3 className="text-2xl font-bold mt-1">{employees.length}</h3>
            </div>
            <div className="p-2.5 rounded-full bg-primary/10 text-primary">
              <DatabaseIcon className="h-5 w-5" />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-card">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-xs font-medium text-muted-foreground">Synced in MDB (att2000)</p>
              <h3 className="text-2xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">{syncedCount}</h3>
            </div>
            <div className="p-2.5 rounded-full bg-emerald-100 dark:bg-emerald-950/40 text-emerald-600 dark:text-emerald-400">
              <CheckCircle2Icon className="h-5 w-5" />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-card">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-xs font-medium text-muted-foreground">Pending Sync</p>
              <h3 className="text-2xl font-bold text-amber-600 dark:text-amber-400 mt-1">{employees.length - syncedCount}</h3>
            </div>
            <div className="p-2.5 rounded-full bg-amber-100 dark:bg-amber-950/40 text-amber-600 dark:text-amber-400">
              <AlertCircleIcon className="h-5 w-5" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Actions & Filters */}
      <Card>
        <CardContent className="p-4 sm:p-6 space-y-4">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 flex-1">
              <div className="relative flex-1 max-w-md">
                <SearchIcon className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search by name, ID, or punch number..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="pl-9 text-sm"
                />
              </div>

              <div className="flex items-center gap-2">
                <FilterIcon className="h-4 w-4 text-muted-foreground" />
                <select
                  value={filterStatus}
                  aria-label="Filter Status"
                  onChange={(e) => setFilterStatus(e.target.value as "all" | "synced" | "unsynced")}
                  className="h-10 rounded-md border border-input bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  <option value="all">All Employees</option>
                  <option value="unsynced">Pending Sync Only</option>
                  <option value="synced">Synced Only</option>
                </select>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Button
                variant="default"
                onClick={handleSyncSelected}
                disabled={syncing || selectedIds.length === 0}
              >
                {syncing ? <Loader2Icon className="h-4 w-4 mr-2 animate-spin" /> : <RefreshCwIcon className="h-4 w-4 mr-2" />}
                Sync Selected ({selectedIds.length})
              </Button>

              <Button
                variant="outline"
                onClick={handleSyncAll}
                disabled={syncing || employees.length === 0}
              >
                {syncing ? <Loader2Icon className="h-4 w-4 mr-2 animate-spin" /> : <RefreshCwIcon className="h-4 w-4 mr-2" />}
                Sync All Employees
              </Button>
            </div>
          </div>

          {/* Employee Table */}
          <div className="border rounded-lg overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/40 font-medium text-muted-foreground">
                  <th className="py-3 px-4 text-left w-12">
                    <Checkbox
                      checked={filteredEmployees.length > 0 && selectedIds.length === filteredEmployees.length}
                      onCheckedChange={toggleSelectAll}
                    />
                  </th>
                  <th className="py-3 px-4 text-left font-semibold">Employee ID</th>
                  <th className="py-3 px-4 text-left font-semibold">Punch No. (Badge)</th>
                  <th className="py-3 px-4 text-left font-semibold">Employee Name</th>
                  <th className="py-3 px-4 text-left font-semibold">Department</th>
                  <th className="py-3 px-4 text-left font-semibold">Designation</th>
                  <th className="py-3 px-4 text-center font-semibold">Sync Status</th>
                  <th className="py-3 px-4 text-right font-semibold">Action</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={8} className="py-12 text-center text-muted-foreground">
                      <Loader2Icon className="h-6 w-6 animate-spin mx-auto mb-2" />
                      Loading employee sync status...
                    </td>
                  </tr>
                ) : filteredEmployees.length === 0 ? (
                  <tr>
                    <td colSpan={8} className="py-12 text-center text-muted-foreground">
                      No employees found matching filter criteria.
                    </td>
                  </tr>
                ) : (
                  filteredEmployees.map((emp) => {
                    const isSelected = selectedIds.includes(emp.id)
                    return (
                      <tr
                        key={emp.id}
                        className={`border-b last:border-0 hover:bg-muted/50 transition-colors ${
                          isSelected ? "bg-primary/5" : ""
                        }`}
                      >
                        <td className="py-3 px-4">
                          <Checkbox
                            checked={isSelected}
                            onCheckedChange={() => toggleSelectOne(emp.id)}
                          />
                        </td>
                        <td className="py-3 px-4 font-mono font-medium">{emp.employee_id}</td>
                        <td className="py-3 px-4 font-mono font-bold text-primary">{emp.punch_number || emp.employee_id}</td>
                        <td className="py-3 px-4 font-medium">{emp.name}</td>
                        <td className="py-3 px-4 text-muted-foreground">{emp.department || "-"}</td>
                        <td className="py-3 px-4 text-muted-foreground">{emp.designation || "-"}</td>
                        <td className="py-3 px-4 text-center">
                          {emp.is_synched ? (
                            <Badge variant="outline" className="bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800">
                              <CheckCircle2Icon className="h-3.5 w-3.5 mr-1" />
                              Synced in MDB
                            </Badge>
                          ) : (
                            <Badge variant="outline" className="bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800">
                              <AlertCircleIcon className="h-3.5 w-3.5 mr-1" />
                              Pending Sync
                            </Badge>
                          )}
                        </td>
                        <td className="py-3 px-4 text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleSyncSingle(emp)}
                            disabled={syncing}
                            className="h-8 px-2.5 text-xs"
                          >
                            <RefreshCwIcon className="h-3.5 w-3.5 mr-1 text-primary" />
                            Sync to MDB
                          </Button>
                        </td>
                      </tr>
                    )
                  })
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
