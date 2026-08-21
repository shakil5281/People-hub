"use client"

import * as React from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Loader2, UserPlusIcon, UsersIcon, UserXIcon, ClockIcon, CheckCircleIcon } from "lucide-react"
import { dashboardApi, employeeApi, separationApi } from "@/lib/api"
import { toast } from "sonner"

interface Employee {
  id: string
  employee_id: string
  name_en: string
  status: string
  joining_date: string
}

interface Separation {
  id: string
  employee: string
  type: string
  date: string
  status: string
}

const cardBg = "bg-gradient-to-b from-white to-gray-100 dark:from-gray-900 dark:to-gray-800/80"

export default function LifecyclePage() {
  const [employees, setEmployees] = React.useState<Employee[]>([])
  const [separations, setSeparations] = React.useState<Separation[]>([])
  const [stats, setStats] = React.useState<{ new_hires_month: number; separations_month: number; active_employees: number } | null>(null)
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    Promise.all([
      dashboardApi.stats(),
      employeeApi.list({ status: "active", limit: "10" }),
      separationApi.list(),
    ])
      .then(([statsRes, empRes, sepRes]) => {
        setStats({ new_hires_month: statsRes.data.new_hires_month, separations_month: statsRes.data.separations_month, active_employees: statsRes.data.active_employees })
        setEmployees(Array.isArray(empRes.data?.data) ? empRes.data.data.slice(0, 10) : (Array.isArray(empRes.data) ? empRes.data.slice(0, 10) : []))
        setSeparations(Array.isArray(sepRes.data?.data) ? sepRes.data.data.slice(0, 10) : (Array.isArray(sepRes.data) ? sepRes.data.slice(0, 10) : []))
      })
      .catch(() => toast.error("Failed to load lifecycle data"))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const retentionRate = stats && (stats.active_employees + stats.separations_month) > 0
    ? Math.round((stats.active_employees / (stats.active_employees + stats.separations_month)) * 100)
    : 100

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <h1 className="text-3xl font-bold tracking-tight">Lifecycle Management</h1>
        <p className="text-muted-foreground mt-1">Employee lifecycle overview from hire to separation</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 px-4 lg:px-6">
        <Card className={`border-l-4 border-l-emerald-500 ${cardBg}`}>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <UserPlusIcon className="h-4 w-4 text-emerald-500" />
              New Hires (Month)
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-emerald-600">{stats?.new_hires_month ?? 0}</div>
          </CardContent>
        </Card>
        <Card className={`border-l-4 border-l-blue-500 ${cardBg}`}>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <UsersIcon className="h-4 w-4 text-blue-500" />
              Active Employees
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-blue-600">{stats?.active_employees ?? 0}</div>
          </CardContent>
        </Card>
        <Card className={`border-l-4 border-l-red-500 ${cardBg}`}>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <UserXIcon className="h-4 w-4 text-red-500" />
              Separations (Month)
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-red-600">{stats?.separations_month ?? 0}</div>
          </CardContent>
        </Card>
        <Card className={`border-l-4 border-l-purple-500 ${cardBg}`}>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <ClockIcon className="h-4 w-4 text-purple-500" />
              Retention Rate
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-purple-600">{retentionRate}%</div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 px-4 lg:px-6">
        <Card className={`lg:col-span-2 ${cardBg}`}>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <UserPlusIcon className="h-4 w-4 text-emerald-500" />
              Recent Hires
            </CardTitle>
          </CardHeader>
          <CardContent>
            {employees.length > 0 ? (
              <div className="space-y-2">
                {employees.map((emp) => (
                  <div key={emp.id} className="flex items-center justify-between p-2 rounded-lg hover:bg-muted/50 transition-colors">
                    <div>
                      <p className="text-sm font-medium">{emp.name_en}</p>
                      <p className="text-xs text-muted-foreground">ID: {emp.employee_id}</p>
                    </div>
                    <Badge variant="outline" className="text-xs">{emp.joining_date || "N/A"}</Badge>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-4">No employee data</p>
            )}
          </CardContent>
        </Card>

        <Card className={cardBg}>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <UserXIcon className="h-4 w-4 text-red-500" />
              Recent Separations
            </CardTitle>
          </CardHeader>
          <CardContent>
            {separations.length > 0 ? (
              <div className="space-y-2">
                {separations.map((sep) => (
                  <div key={sep.id} className="flex items-center justify-between p-2 rounded-lg hover:bg-muted/50 transition-colors">
                    <div>
                      <p className="text-sm font-medium truncate max-w-[140px]">{sep.employee}</p>
                      <p className="text-xs text-muted-foreground">{sep.type}</p>
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <span className="text-xs text-muted-foreground">{sep.date}</span>
                      {sep.status === "Approved" ? (
                        <CheckCircleIcon className="h-3.5 w-3.5 text-red-500" />
                      ) : (
                        <ClockIcon className="h-3.5 w-3.5 text-amber-500" />
                      )}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-4">No separation data</p>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 px-4 lg:px-6">
        <div className="lg:col-span-3">
          <Card className={cardBg}>
            <CardHeader>
              <CardTitle className="text-base">Employee Lifecycle Stages</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 sm:grid-cols-3 md:grid-cols-5 gap-3">
                <div className="text-center p-4 rounded-lg bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-800">
                  <div className="text-2xl font-bold text-blue-600">01</div>
                  <p className="text-sm font-medium mt-1">Requirement</p>
                  <p className="text-xs text-muted-foreground">Job requisition</p>
                </div>
                <div className="text-center p-4 rounded-lg bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800">
                  <div className="text-2xl font-bold text-green-600">02</div>
                  <p className="text-sm font-medium mt-1">Hiring</p>
                  <p className="text-xs text-muted-foreground">Onboarding</p>
                </div>
                <div className="text-center p-4 rounded-lg bg-teal-50 dark:bg-teal-950/20 border border-teal-200 dark:border-teal-800">
                  <div className="text-2xl font-bold text-teal-600">03</div>
                  <p className="text-sm font-medium mt-1">Active</p>
                  <p className="text-xs text-muted-foreground">Employment</p>
                </div>
                <div className="text-center p-4 rounded-lg bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-800">
                  <div className="text-2xl font-bold text-amber-600">04</div>
                  <p className="text-sm font-medium mt-1">Offboarding</p>
                  <p className="text-xs text-muted-foreground">Exit process</p>
                </div>
                <div className="text-center p-4 rounded-lg bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
                  <div className="text-2xl font-bold text-gray-600">05</div>
                  <p className="text-sm font-medium mt-1">Separated</p>
                  <p className="text-xs text-muted-foreground">Final settlement</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
