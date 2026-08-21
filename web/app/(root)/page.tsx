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

  return (
    <div className="flex flex-col gap-6 py-4 md:gap-8 md:py-6">

      {/* Hero */}
      <div className="px-4 lg:px-6">
        <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-sky-50 via-indigo-50/50 to-background border border-sky-100 p-6 md:p-8">
          <div className="absolute top-0 right-0 w-64 h-64 bg-sky-200/30 rounded-full -translate-y-1/2 translate-x-1/2 blur-3xl" />
          <div className="absolute bottom-0 left-0 w-48 h-48 bg-indigo-200/30 rounded-full translate-y-1/2 -translate-x-1/2 blur-3xl" />
          <div className="relative flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <SparklesIcon className="h-5 w-5 text-sky-500" />
                <span className="text-xs font-semibold uppercase tracking-widest text-sky-600">PeopleHub</span>
              </div>
              <h1 className="text-3xl md:text-4xl font-bold tracking-tight text-slate-800">
                Good to see you, <span className="text-sky-600">Admin</span>
              </h1>
              <p className="text-slate-500 mt-1 max-w-lg">
                Here&apos;s what&apos;s happening across your organization today.
              </p>
            </div>
            <div className="flex items-center gap-3">
              <div className="text-right">
                <p className="text-2xl font-bold tabular-nums text-sky-600">{attendanceRate}%</p>
                <p className="text-xs text-slate-500">Today Attendance</p>
              </div>
              <div className="h-12 w-px bg-sky-200" />
              <div className="text-right">
                <p className="text-2xl font-bold tabular-nums text-amber-500">{data?.pending_leaves ?? 0}</p>
                <p className="text-xs text-slate-500">Pending Leaves</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Stat Cards */}
      <div className="px-4 lg:px-6">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 md:gap-4">
          {[
            { title: "Total Employees", value: data?.total_employees ?? 0, icon: UsersIcon, change: "+12%", trend: "up" as const, color: "text-sky-500", bg: "bg-sky-50" },
            { title: "Active Employees", value: data?.active_employees ?? 0, icon: UsersIcon, change: `${attendanceRate}% today`, trend: attendanceRate > 70 ? "up" as const : "down" as const, color: "text-emerald-500", bg: "bg-emerald-50" },
            { title: "Today Attendance", value: data?.today_attendance ?? 0, icon: ClockIcon, change: `${data?.today_logs ?? 0} logs`, trend: "up" as const, color: "text-violet-500", bg: "bg-violet-50" },
            { title: "Pending Leaves", value: data?.pending_leaves ?? 0, icon: CalendarCheckIcon, change: `${data?.new_hires_month ?? 0} new hires`, trend: "up" as const, color: "text-amber-500", bg: "bg-amber-50" },
          ].map((stat) => {
            const Icon = stat.icon
            return (
              <Card key={stat.title} className="relative overflow-hidden group hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 border-slate-200/60">
                <div className="absolute top-0 right-0 w-24 h-24 bg-gradient-to-bl from-sky-100/40 to-transparent rounded-bl-full" />
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-xs font-medium text-slate-500">{stat.title}</CardTitle>
                  <div className={`p-1.5 rounded-md ${stat.bg} ${stat.color}`}>
                    <Icon className="h-3.5 w-3.5" />
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold tabular-nums text-slate-800">{stat.value.toLocaleString()}</div>
                  <div className="flex items-center gap-1 mt-1">
                    {stat.trend === "up" ? (
                      <TrendingUpIcon className="h-3 w-3 text-emerald-500" />
                    ) : (
                      <TrendingDownIcon className="h-3 w-3 text-red-400" />
                    )}
                    <span className={`text-xs ${stat.trend === "up" ? "text-emerald-600" : "text-red-500"}`}>
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
        <h2 className="text-sm font-semibold text-sky-700 uppercase tracking-wider mb-3 flex items-center gap-2">
          <SparklesIcon className="h-4 w-4 text-sky-500" />
          Quick Access
        </h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {QUICK_LINKS.map((link) => {
            const Icon = link.icon
            return (
              <a
                key={link.label}
                href={link.href}
                className="group relative overflow-hidden rounded-xl bg-gradient-to-br p-4 text-black shadow-md hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5"
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
        <Card className="lg:col-span-4 border-slate-200/60">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2 text-sky-800">
              <BuildingIcon className="h-4 w-4 text-sky-500" />
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
                      <span className="text-sm font-medium text-slate-700 w-36 md:w-48 truncate">{dept.name}</span>
                      <div className="flex-1 h-5 bg-sky-100 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-gradient-to-r from-sky-400 to-sky-500 rounded-full transition-all duration-500 group-hover:from-sky-500 group-hover:to-sky-600"
                          style={{ width: `${Math.min(100, (dept.count / maxCount) * 100)}%` }}
                        />
                      </div>
                      <span className="text-sm text-slate-600 w-8 text-right tabular-nums">{dept.count}</span>
                      <span className="text-xs text-slate-500 w-10 text-right tabular-nums">{pct}%</span>
                    </div>
                  )
                })}
              </div>
            ) : (
              <div className="py-8 text-center text-sm text-slate-400">No department data available</div>
            )}
          </CardContent>
        </Card>

        {/* Gender + Weekly Overview */}
        <div className="lg:col-span-3 flex flex-col gap-4 md:gap-6">
          {/* Gender Distribution */}
          <Card className="border-slate-200/60">
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2 text-violet-800">
                <UsersIcon className="h-4 w-4 text-violet-500" />
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
                        <span className="text-sm font-medium text-slate-700 w-16">{g.gender}</span>
                        <div className="flex-1 h-4 bg-violet-100 rounded-full overflow-hidden">
                          <div
                            className="h-full bg-gradient-to-r from-violet-400 to-violet-500 rounded-full transition-all duration-500"
                            style={{ width: `${Math.min(100, (g.count / maxCount) * 100)}%` }}
                          />
                        </div>
                        <span className="text-sm text-slate-600 w-8 text-right tabular-nums">{g.count}</span>
                      </div>
                    )
                  })}
                </div>
              ) : (
                <div className="py-6 text-center text-sm text-slate-400">No gender data</div>
              )}
            </CardContent>
          </Card>

          {/* Quick Stats */}
          <Card className="border-slate-200/60">
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2 text-emerald-800">
                <ActivityIcon className="h-4 w-4 text-emerald-500" />
                Weekly Overview
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-3 gap-3">
                <div className="rounded-lg bg-emerald-50 border border-emerald-200 p-3 text-center">
                  <p className="text-2xl font-bold text-emerald-500 tabular-nums">{totalPresent}</p>
                  <p className="text-xs text-emerald-600/80 mt-0.5 font-medium">Present</p>
                </div>
                <div className="rounded-lg bg-red-50 border border-red-200 p-3 text-center">
                  <p className="text-2xl font-bold text-red-400 tabular-nums">{totalAbsent}</p>
                  <p className="text-xs text-red-500/80 mt-0.5 font-medium">Absent</p>
                </div>
                <div className="rounded-lg bg-amber-50 border border-amber-200 p-3 text-center">
                  <p className="text-2xl font-bold text-amber-500 tabular-nums">{totalLate}</p>
                  <p className="text-xs text-amber-600/80 mt-0.5 font-medium">Late</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Bottom Row: Last 7 Days Attendance + Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-7 gap-4 md:gap-6 px-4 lg:px-6 pb-6">
        {/* Last 7 Days Attendance */}
        <Card className="lg:col-span-4 border-slate-200/60">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2 text-indigo-800">
              <CalendarDaysIcon className="h-4 w-4 text-indigo-500" />
              Last 7 Days Attendance
            </CardTitle>
          </CardHeader>
          <CardContent>
            {data?.last_7_days && data.last_7_days.length > 0 ? (
              <div className="space-y-1">
                <div className="grid grid-cols-4 gap-2 text-xs font-medium pb-2 border-b border-indigo-100">
                  <span className="text-indigo-600">Date</span>
                  <span className="text-right text-emerald-600">Present</span>
                  <span className="text-right text-red-500">Absent</span>
                  <span className="text-right text-amber-600">Late</span>
                </div>
                {data.last_7_days.map((d, i) => (
                  <div key={`${d.date}-${i}`} className="grid grid-cols-4 gap-2 text-sm py-2 border-b border-slate-100 last:border-0 hover:bg-indigo-50/50 transition-colors rounded-sm px-1">
                    <span className="font-medium text-slate-700">{d.date}</span>
                    <span className="text-right text-emerald-500 font-medium tabular-nums">{d.present}</span>
                    <span className="text-right text-red-400 font-medium tabular-nums">{d.absent}</span>
                    <span className="text-right text-amber-500 font-medium tabular-nums">{d.late}</span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="py-8 text-center text-sm text-slate-400">No attendance data</div>
            )}
          </CardContent>
        </Card>

        {/* Recent Activity */}
        <Card className="lg:col-span-3 border-slate-200/60">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2 text-amber-800">
              <ActivityIcon className="h-4 w-4 text-amber-500" />
              Recent Activity
            </CardTitle>
          </CardHeader>
          <CardContent>
            {data?.recent_activity && data.recent_activity.length > 0 ? (
              <div className="space-y-1">
                {data.recent_activity.map((activity, index) => (
                  <div key={index} className="flex items-center gap-3 p-2 rounded-lg hover:bg-slate-50 transition-all duration-200">
                    <div className={`shrink-0 p-1.5 rounded-full ${
                      activity.type === "leave_request" ? "bg-amber-100 text-amber-600" :
                      activity.type === "new_employee" ? "bg-emerald-100 text-emerald-600" :
                      "bg-slate-100 text-slate-500"
                    }`}>
                      {activity.type === "leave_request" ? <CalendarCheckIcon className="h-3.5 w-3.5" /> :
                       activity.type === "new_employee" ? <UserPlusIcon className="h-3.5 w-3.5" /> :
                       <ActivityIcon className="h-3.5 w-3.5" />}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-slate-700 truncate">{activity.description}</p>
                    </div>
                    <span className="text-xs text-slate-400 whitespace-nowrap">{activity.date}</span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="py-8 text-center text-sm text-slate-400">No recent activity</div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
