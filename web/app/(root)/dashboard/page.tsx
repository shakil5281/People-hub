"use client"

import * as React from "react"
import {
  UsersIcon,
  BuildingIcon,
  ClockIcon,
  CalendarCheckIcon,
  UserPlusIcon,
  UserXIcon,
  ActivityIcon,
  Loader2,
  TrendingUpIcon,
  TrendingDownIcon,
  BarChart3Icon,
  SparklesIcon,
  CalendarDaysIcon,
} from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { dashboardApi } from "@/lib/api"
import { toast } from "sonner"

interface DashboardStats {
  total_employees: number
  active_employees: number
  total_departments: number
  total_sections: number
  today_attendance: number
  today_logs: number
  pending_leaves: number
  new_hires_month: number
  separations_month: number
  gender_distribution: { gender: string; count: number }[]
  department_counts: { name: string; count: number }[]
  last_7_days: { date: string; present: number; absent: number; late: number }[]
  recent_activity: { type: string; description: string; date: string }[]
}

export default function DashboardPage() {
  const [data, setData] = React.useState<DashboardStats | null>(null)
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    dashboardApi.stats()
      .then((res) => setData(res.data))
      .catch(() => toast.error("Failed to load dashboard"))
      .finally(() => setLoading(false))
  }, [])

  const totalPresent = data?.last_7_days?.reduce((s, m) => s + m.present, 0) || 0
  const totalAbsent = data?.last_7_days?.reduce((s, m) => s + m.absent, 0) || 0
  const totalLate = data?.last_7_days?.reduce((s, m) => s + m.late, 0) || 0
  const attendanceRate = data?.active_employees
    ? Math.round((data.today_attendance / data.active_employees) * 100)
    : 0

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[80vh]">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-10 w-10 animate-spin text-sky-500" />
          <p className="text-sky-600 animate-pulse font-medium">Loading dashboard...</p>
        </div>
      </div>
    )
  }

  const statCards = [
    { title: "Total Employees", value: data?.total_employees ?? 0, icon: UsersIcon, color: "text-blue-600", change: `${data?.active_employees ?? 0} active`, trend: "up" as const },
    { title: "Active Employees", value: data?.active_employees ?? 0, icon: UsersIcon, color: "text-emerald-600", change: `${attendanceRate}% today`, trend: attendanceRate > 70 ? "up" as const : "down" as const },
    { title: "Today Attendance", value: data?.today_attendance ?? 0, icon: ClockIcon, color: "text-violet-600", change: `${data?.today_logs ?? 0} logs`, trend: "up" as const },
    { title: "Pending Leaves", value: data?.pending_leaves ?? 0, icon: CalendarCheckIcon, color: "text-amber-600", change: `${data?.new_hires_month ?? 0} new hires`, trend: "up" as const },
    { title: "Departments", value: data?.total_departments ?? 0, icon: BuildingIcon, color: "text-indigo-600", change: "Active", trend: "up" as const },
    { title: "New Hires", value: data?.new_hires_month ?? 0, icon: UserPlusIcon, color: "text-emerald-600", change: "This month", trend: "up" as const },
    { title: "Separations", value: data?.separations_month ?? 0, icon: UserXIcon, color: "text-red-600", change: "This month", trend: "down" as const },
    { title: "Sections", value: data?.total_sections ?? 0, icon: BuildingIcon, color: "text-teal-600", change: "Active", trend: "up" as const },
  ]

  return (
    <div className="flex flex-col gap-6 py-4 md:gap-8 md:py-6">

      {/* Hero */}
      <div className="px-4 lg:px-6">
        <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-sky-50 via-indigo-50/50 to-background border border-sky-100 p-6 md:p-8">
          <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-bl from-sky-200/30 to-transparent rounded-full -translate-y-1/2 translate-x-1/2 blur-3xl" />
          <div className="absolute bottom-0 left-0 w-48 h-48 bg-gradient-to-tr from-indigo-200/20 to-transparent rounded-full translate-y-1/2 -translate-x-1/2 blur-3xl" />
          <div className="relative flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <SparklesIcon className="h-5 w-5 text-sky-500" />
                <span className="text-xs font-semibold uppercase tracking-widest text-sky-600">Dashboard</span>
              </div>
              <h1 className="text-3xl md:text-4xl font-bold tracking-tight">
                HR Overview
              </h1>
              <p className="text-muted-foreground mt-1 max-w-lg">
                Key metrics and performance indicators across your organization.
              </p>
            </div>
            <div className="flex items-center gap-3">
              <div className="text-right">
                <p className="text-2xl font-bold tabular-nums">{attendanceRate}%</p>
                <p className="text-xs text-muted-foreground">Today Attendance</p>
              </div>
              <div className="h-12 w-px bg-border" />
              <div className="text-right">
                <p className="text-2xl font-bold tabular-nums">{data?.pending_leaves ?? 0}</p>
                <p className="text-xs text-muted-foreground">Pending Leaves</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Stat Cards */}
      <div className="px-4 lg:px-6">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 md:gap-4">
          {statCards.map((stat) => {
            const Icon = stat.icon
            return (
              <Card key={stat.title} className="relative overflow-hidden group hover:shadow-md transition-all duration-300">
                <div className="absolute top-0 right-0 w-24 h-24 bg-gradient-to-bl from-primary/[0.03] to-transparent rounded-bl-full" />
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-xs font-medium text-muted-foreground">{stat.title}</CardTitle>
                  <div className={`p-1.5 rounded-md bg-background border ${stat.color}`}>
                    <Icon className="h-3.5 w-3.5" />
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold tabular-nums">{stat.value.toLocaleString()}</div>
                  <div className="flex items-center gap-1 mt-1">
                    {stat.trend === "up" ? (
                      <TrendingUpIcon className="h-3 w-3 text-emerald-500" />
                    ) : (
                      <TrendingDownIcon className="h-3 w-3 text-red-500" />
                    )}
                    <span className={`text-xs ${stat.trend === "up" ? "text-emerald-600" : "text-red-600"}`}>
                      {stat.change}
                    </span>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      </div>

      {/* Charts & Data */}
      <div className="grid grid-cols-1 lg:grid-cols-7 gap-4 md:gap-6 px-4 lg:px-6">
        {/* Department Distribution */}
        <Card className="lg:col-span-4">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <BuildingIcon className="h-4 w-4 text-muted-foreground" />
              Department Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            {data?.department_counts && data.department_counts.length > 0 ? (
              <div className="space-y-3">
                {data.department_counts.slice(0, 8).map((dept) => {
                  const maxCount = Math.max(...data.department_counts.map(d => d.count))
                  const pct = Math.round((dept.count / (data.active_employees || 1)) * 100)
                  return (
                    <div key={dept.name} className="flex items-center gap-3 group">
                      <span className="text-sm font-medium w-36 md:w-48 truncate">{dept.name}</span>
                      <div className="flex-1 h-5 bg-muted rounded-full overflow-hidden">
                        <div
                          className="h-full bg-gradient-to-r from-primary to-primary/60 rounded-full transition-all duration-500 group-hover:from-primary/90 group-hover:to-primary/50"
                          style={{ width: `${Math.min(100, (dept.count / maxCount) * 100)}%` }}
                        />
                      </div>
                      <span className="text-sm text-muted-foreground w-8 text-right tabular-nums">{dept.count}</span>
                      <span className="text-xs text-muted-foreground w-10 text-right tabular-nums">{pct}%</span>
                    </div>
                  )
                })}
              </div>
            ) : (
              <div className="py-8 text-center text-sm text-muted-foreground">No department data available</div>
            )}
          </CardContent>
        </Card>

        {/* Gender + Quick Stats */}
        <div className="lg:col-span-3 flex flex-col gap-4 md:gap-6">
          {/* Gender Distribution */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <UsersIcon className="h-4 w-4 text-muted-foreground" />
                Gender Distribution
              </CardTitle>
            </CardHeader>
            <CardContent>
              {data?.gender_distribution && data.gender_distribution.length > 0 ? (
                <div className="space-y-3">
                  {data.gender_distribution.map((g) => {
                    const maxCount = Math.max(...data.gender_distribution.map(d => d.count))
                    return (
                      <div key={g.gender} className="flex items-center gap-3">
                        <span className="text-sm font-medium w-16">{g.gender}</span>
                        <div className="flex-1 h-4 bg-muted rounded-full overflow-hidden">
                          <div
                            className="h-full bg-gradient-to-r from-primary to-primary/60 rounded-full transition-all duration-500"
                            style={{ width: `${Math.min(100, (g.count / maxCount) * 100)}%` }}
                          />
                        </div>
                        <span className="text-sm text-muted-foreground w-8 text-right tabular-nums">{g.count}</span>
                      </div>
                    )
                  })}
                </div>
              ) : (
                <div className="py-6 text-center text-sm text-muted-foreground">No gender data</div>
              )}
            </CardContent>
          </Card>

          {/* Quick Stats */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <BarChart3Icon className="h-4 w-4 text-muted-foreground" />
                Last 7 Days Overview
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-3 gap-3">
                <div className="rounded-lg bg-emerald-50 dark:bg-emerald-950/30 p-3 text-center">
                  <p className="text-2xl font-bold text-emerald-600 tabular-nums">{totalPresent}</p>
                  <p className="text-xs text-emerald-600/70 mt-0.5">Present</p>
                </div>
                <div className="rounded-lg bg-red-50 dark:bg-red-950/30 p-3 text-center">
                  <p className="text-2xl font-bold text-red-600 tabular-nums">{totalAbsent}</p>
                  <p className="text-xs text-red-600/70 mt-0.5">Absent</p>
                </div>
                <div className="rounded-lg bg-amber-50 dark:bg-amber-950/30 p-3 text-center">
                  <p className="text-2xl font-bold text-amber-600 tabular-nums">{totalLate}</p>
                  <p className="text-xs text-amber-600/70 mt-0.5">Late</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Bottom Row: 7-Day Attendance Table + Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-7 gap-4 md:gap-6 px-4 lg:px-6 pb-6">
        {/* Last 7 Days Attendance */}
        <Card className="lg:col-span-4">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <CalendarDaysIcon className="h-4 w-4 text-muted-foreground" />
              Daily Attendance Breakdown
            </CardTitle>
          </CardHeader>
          <CardContent>
            {data?.last_7_days && data.last_7_days.length > 0 ? (
              <div className="space-y-1">
                <div className="grid grid-cols-4 gap-2 text-xs text-muted-foreground font-medium pb-2 border-b">
                  <span>Date</span>
                  <span className="text-right">Present</span>
                  <span className="text-right">Absent</span>
                  <span className="text-right">Late</span>
                </div>
                {data.last_7_days.map((d, i) => (
                  <div key={`${d.date}-${i}`} className="grid grid-cols-4 gap-2 text-sm py-2 border-b last:border-0 hover:bg-muted/30 transition-colors rounded-sm px-1">
                    <span className="font-medium">{d.date}</span>
                    <span className="text-right text-emerald-600 font-medium tabular-nums">{d.present}</span>
                    <span className="text-right text-red-600 font-medium tabular-nums">{d.absent}</span>
                    <span className="text-right text-amber-600 font-medium tabular-nums">{d.late}</span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="py-8 text-center text-sm text-muted-foreground">No attendance data</div>
            )}
          </CardContent>
        </Card>

        {/* Recent Activity */}
        <Card className="lg:col-span-3">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <ActivityIcon className="h-4 w-4 text-muted-foreground" />
              Recent Activity
            </CardTitle>
          </CardHeader>
          <CardContent>
            {data?.recent_activity && data.recent_activity.length > 0 ? (
              <div className="space-y-1">
                {data.recent_activity.map((activity, index) => (
                  <div key={index} className="flex items-center gap-3 p-2 rounded-lg hover:bg-muted/50 transition-all duration-200">
                    <div className={`shrink-0 p-1.5 rounded-full ${
                      activity.type === "leave_request" ? "bg-orange-100 text-orange-600 dark:bg-orange-950/40" :
                      activity.type === "new_employee" ? "bg-green-100 text-green-600 dark:bg-green-950/40" :
                      "bg-gray-100 text-gray-600 dark:bg-gray-800"
                    }`}>
                      {activity.type === "leave_request" ? <CalendarCheckIcon className="h-3.5 w-3.5" /> :
                       activity.type === "new_employee" ? <UserPlusIcon className="h-3.5 w-3.5" /> :
                       <ActivityIcon className="h-3.5 w-3.5" />}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm truncate">{activity.description}</p>
                    </div>
                    <span className="text-xs text-muted-foreground whitespace-nowrap">{activity.date}</span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="py-8 text-center text-sm text-muted-foreground">No recent activity</div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
