"use client"

import * as React from "react"
import {
  BuildingIcon,
  UsersIcon,
  ClockIcon,
  CalendarCheckIcon,
  UserPlusIcon,
  UserXIcon,
  Loader2,
  TrendingUpIcon,
  TrendingDownIcon,
  ActivityIcon,
  BarChart3Icon,
  SparklesIcon,
} from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
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
  monthly_attendance: { month: string; present: number; absent: number; late: number }[]
  recent_activity: { type: string; description: string; date: string }[]
}

const QUICK_LINKS = [
  { label: "Attendance", href: "/attendance/daily-attendance", icon: ClockIcon, color: "from-violet-500 to-purple-600" },
  { label: "Employees", href: "/hr/employees", icon: UsersIcon, color: "from-blue-500 to-indigo-600" },
  { label: "Salary Sheet", href: "/payroll/salary-sheet", icon: BarChart3Icon, color: "from-emerald-500 to-teal-600" },
  { label: "Leave", href: "/leave/leave", icon: CalendarCheckIcon, color: "from-orange-500 to-amber-600" },
]

export default function HomePage() {
  const [data, setData] = React.useState<DashboardStats | null>(null)
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    dashboardApi.stats()
      .then((res) => setData(res.data))
      .catch(() => toast.error("Failed to load dashboard"))
      .finally(() => setLoading(false))
  }, [])

  const totalPresent = data?.monthly_attendance?.reduce((s, m) => s + m.present, 0) || 0
  const totalAbsent = data?.monthly_attendance?.reduce((s, m) => s + m.absent, 0) || 0
  const totalLate = data?.monthly_attendance?.reduce((s, m) => s + m.late, 0) || 0
  const attendanceRate = data?.active_employees
    ? Math.round((data.today_attendance / data.active_employees) * 100)
    : 0

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[80vh]">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-10 w-10 animate-spin text-primary" />
          <p className="text-muted-foreground animate-pulse">Loading dashboard...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6 py-4 md:gap-8 md:py-6">

      {/* Hero */}
      <div className="px-4 lg:px-6">
        <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-primary/10 via-primary/5 to-background border p-6 md:p-8">
          <div className="absolute top-0 right-0 w-64 h-64 bg-primary/5 rounded-full -translate-y-1/2 translate-x-1/2 blur-3xl" />
          <div className="absolute bottom-0 left-0 w-48 h-48 bg-primary/5 rounded-full translate-y-1/2 -translate-x-1/2 blur-3xl" />
          <div className="relative flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <SparklesIcon className="h-5 w-5 text-primary" />
                <span className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">PeopleHub</span>
              </div>
              <h1 className="text-3xl md:text-4xl font-bold tracking-tight">
                Good to see you, <span className="text-primary">Admin</span>
              </h1>
              <p className="text-muted-foreground mt-1 max-w-lg">
                Here&apos;s what&apos;s happening across your organization today.
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
          {[
            { title: "Total Employees", value: data?.total_employees ?? 0, icon: UsersIcon, change: "+12%", trend: "up" as const, color: "text-blue-600" },
            { title: "Active Employees", value: data?.active_employees ?? 0, icon: UsersIcon, change: `${attendanceRate}% today`, trend: attendanceRate > 70 ? "up" as const : "down" as const, color: "text-emerald-600" },
            { title: "Today Attendance", value: data?.today_attendance ?? 0, icon: ClockIcon, change: `${data?.today_logs ?? 0} logs`, trend: "up" as const, color: "text-violet-600" },
            { title: "Pending Leaves", value: data?.pending_leaves ?? 0, icon: CalendarCheckIcon, change: `${data?.new_hires_month ?? 0} new hires`, trend: "up" as const, color: "text-amber-600" },
          ].map((stat) => {
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

      {/* Quick Links */}
      <div className="px-4 lg:px-6">
        <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3">Quick Access</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {QUICK_LINKS.map((link) => {
            const Icon = link.icon
            return (
              <a
                key={link.label}
                href={link.href}
                className="group relative overflow-hidden rounded-xl bg-gradient-to-br p-4 text-white shadow-md hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5"
                style={{ backgroundImage: `linear-gradient(135deg, ${link.color.replace("from-", "").split(" ")[0]}, ${link.color.replace("to-", "").split(" ")[0]})` }}
              >
                <div className="absolute top-2 right-2 opacity-10">
                  <Icon className="h-16 w-16" />
                </div>
                <Icon className="h-6 w-6 mb-2 opacity-90" />
                <p className="font-semibold text-sm">{link.label}</p>
              </a>
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

        {/* Gender + Monthly Attendance */}
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
                <ActivityIcon className="h-4 w-4 text-muted-foreground" />
                Monthly Overview
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

      {/* Bottom Row: Monthly Attendance Table + Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-7 gap-4 md:gap-6 px-4 lg:px-6 pb-6">
        {/* Monthly Attendance */}
        <Card className="lg:col-span-4">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <BarChart3Icon className="h-4 w-4 text-muted-foreground" />
              Monthly Attendance Trend
            </CardTitle>
          </CardHeader>
          <CardContent>
            {data?.monthly_attendance && data.monthly_attendance.length > 0 ? (
              <div className="space-y-1">
                <div className="grid grid-cols-4 gap-2 text-xs text-muted-foreground font-medium pb-2 border-b">
                  <span>Month</span>
                  <span className="text-right">Present</span>
                  <span className="text-right">Absent</span>
                  <span className="text-right">Late</span>
                </div>
                {data.monthly_attendance.map((m) => (
                  <div key={m.month} className="grid grid-cols-4 gap-2 text-sm py-2 border-b last:border-0 hover:bg-muted/30 transition-colors rounded-sm px-1">
                    <span className="font-medium">{m.month}</span>
                    <span className="text-right text-emerald-600 font-medium tabular-nums">{m.present}</span>
                    <span className="text-right text-red-600 font-medium tabular-nums">{m.absent}</span>
                    <span className="text-right text-amber-600 font-medium tabular-nums">{m.late}</span>
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
