"use client"

import * as React from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Loader2,
  UsersIcon,
  BuildingIcon,
  ClockIcon,
  UserPlusIcon,
  TrendingUpIcon,
  CalendarCheckIcon,
  DollarSignIcon,
  BriefcaseIcon,
} from "lucide-react"
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  ChartLegend,
  ChartLegendContent,
} from "@/components/ui/chart"
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  PieChart,
  Pie,
  Cell,
  AreaChart,
  Area,
  LineChart,
  Line,
} from "recharts"
import { dashboardApi, salaryApi, companyApi } from "@/lib/api"
import { toast } from "sonner"

const COLORS = ["#8884D8", "#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF", "#FF9F40"]
const cardBg = "bg-gradient-to-b from-white to-gray-100 dark:from-gray-900 dark:to-gray-800/80"

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
}

interface SalarySummary {
  total_gross: number
  total_net: number
  total_employees: number
}

export default function AnalyticsPage() {
  const [data, setData] = React.useState<DashboardStats | null>(null)
  const [salaryData, setSalaryData] = React.useState<SalarySummary | null>(null)
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    const fetchAll = async () => {
      try {
        const [statsRes, companiesRes] = await Promise.all([
          dashboardApi.stats(),
          companyApi.list({ limit: "1" }),
        ])
        setData(statsRes.data)
        const companies = Array.isArray(companiesRes.data?.data)
          ? companiesRes.data.data
          : (Array.isArray(companiesRes.data) ? companiesRes.data : [])
        if (companies.length > 0) {
          const now = new Date()
          const m = now.getMonth() + 1
          const y = now.getFullYear()
          salaryApi.summary({ company_id: companies[0].id, month: String(m), year: String(y) })
            .then((sRes) => {
              const d = sRes.data?.data || sRes.data
              if (d && d.total_gross !== undefined) setSalaryData(d)
              else setSalaryData(null)
            })
            .catch(() => setSalaryData(null))
        }
      } catch {
        toast.error("Failed to load analytics data")
      } finally {
        setLoading(false)
      }
    }
    fetchAll()
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const totalEmployees = data?.total_employees ?? 0
  const activeEmployees = data?.active_employees ?? 0

  const genderPieData = (data?.gender_distribution ?? []).map(g => ({
    name: g.gender,
    value: g.count,
  }))

  const deptBarData = (data?.department_counts ?? []).map(d => ({
    name: d.name.length > 12 ? d.name.slice(0, 11) + "\u2026" : d.name,
    employees: d.count,
  }))

  const attendanceTrendData = (data?.monthly_attendance ?? []).map(m => ({
    month: m.month.slice(-2),
    present: m.present,
    absent: m.absent,
    late: m.late,
  }))

  const totalPresent = attendanceTrendData.reduce((s, m) => s + m.present, 0)
  const totalAbsent = attendanceTrendData.reduce((s, m) => s + m.absent, 0)
  const totalLate = attendanceTrendData.reduce((s, m) => s + m.late, 0)

  const kpiCards = [
    { title: "Total Workforce", value: totalEmployees, icon: UsersIcon, color: "text-blue-600 bg-blue-100 dark:bg-blue-950/40" },
    { title: "Active Employees", value: activeEmployees, icon: BriefcaseIcon, color: "text-green-600 bg-green-100 dark:bg-green-950/40" },
    { title: "Departments", value: data?.total_departments ?? 0, icon: BuildingIcon, color: "text-purple-600 bg-purple-100 dark:bg-purple-950/40" },
    { title: "Today Attendance", value: data?.today_attendance ?? 0, icon: ClockIcon, color: "text-cyan-600 bg-cyan-100 dark:bg-cyan-950/40" },
    { title: "Pending Leaves", value: data?.pending_leaves ?? 0, icon: CalendarCheckIcon, color: "text-orange-600 bg-orange-100 dark:bg-orange-950/40" },
    { title: "New Hires", value: data?.new_hires_month ?? 0, icon: UserPlusIcon, color: "text-emerald-600 bg-emerald-100 dark:bg-emerald-950/40" },
  ]

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <h1 className="text-3xl font-bold tracking-tight">Analytics</h1>
        <p className="text-muted-foreground mt-1">HR analytics and workforce metrics</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 px-4 lg:px-6">
        {kpiCards.map((kpi) => {
          const Icon = kpi.icon
          return (
            <Card key={kpi.title} className={cardBg}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">{kpi.title}</CardTitle>
                <div className={`p-2 rounded-md ${kpi.color}`}>
                  <Icon className="h-4 w-4" />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold tabular-nums">{kpi.value.toLocaleString()}</div>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* Attendance Summary */}
      <div className="px-4 lg:px-6">
        <div className="flex items-center gap-2 mb-3">
          <ClockIcon className="h-5 w-5 text-muted-foreground" />
          <h2 className="text-lg font-semibold">Attendance Summary</h2>
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-7 px-4 lg:px-6">
        <Card className={`xl:col-span-4 ${cardBg}`}>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Monthly Attendance Trend</CardTitle>
          </CardHeader>
          <CardContent>
            {attendanceTrendData.length > 0 ? (
              <ChartContainer config={{
                present: { label: "Present", color: "#22c55e" },
                absent: { label: "Absent", color: "#ef4444" },
                late: { label: "Late", color: "#f59e0b" },
              }}>
                <ResponsiveContainer width="100%" height={280}>
                  <AreaChart data={attendanceTrendData} margin={{ left: 0, right: 0, top: 5, bottom: 0 }}>
                    <defs>
                      <linearGradient id="attPresent" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#22c55e" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#22c55e" stopOpacity={0}/>
                      </linearGradient>
                      <linearGradient id="attAbsent" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#ef4444" stopOpacity={0.25}/>
                        <stop offset="95%" stopColor="#ef4444" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--border))" />
                    <XAxis dataKey="month" tickLine={false} axisLine={false} tick={{ fontSize: 12 }} />
                    <YAxis tickLine={false} axisLine={false} tick={{ fontSize: 12 }} />
                    <Tooltip content={<ChartTooltipContent />} />
                    <Area type="monotone" dataKey="present" stroke="#22c55e" strokeWidth={2} fill="url(#attPresent)" />
                    <Area type="monotone" dataKey="absent" stroke="#ef4444" strokeWidth={2} fill="url(#attAbsent)" />
                    <Area type="monotone" dataKey="late" stroke="#f59e0b" strokeWidth={2} fillOpacity={0.1} fill="#f59e0b" />
                  </AreaChart>
                </ResponsiveContainer>
                <ChartLegend>
                  <ChartLegendContent />
                </ChartLegend>
              </ChartContainer>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-12">No attendance data</p>
            )}
          </CardContent>
        </Card>

        <Card className={`xl:col-span-3 ${cardBg}`}>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Attendance Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            {totalPresent + totalAbsent + totalLate > 0 ? (
              <div className="space-y-5">
                {[
                  { label: "Present", value: totalPresent, color: "bg-green-500", textColor: "text-green-600" },
                  { label: "Absent", value: totalAbsent, color: "bg-red-500", textColor: "text-red-600" },
                  { label: "Late", value: totalLate, color: "bg-amber-500", textColor: "text-amber-600" },
                ].map((item) => {
                  const pct = Math.round((item.value / (totalPresent + totalAbsent + totalLate)) * 100)
                  return (
                    <div key={item.label}>
                      <div className="flex items-center justify-between text-sm mb-1">
                        <span className="font-medium">{item.label}</span>
                        <span className={`tabular-nums ${item.textColor}`}>{item.value} ({pct}%)</span>
                      </div>
                      <div className="h-2.5 bg-muted rounded-full overflow-hidden">
                        <div className={`h-full ${item.color} rounded-full transition-all duration-500`} style={{ width: `${pct}%` }} />
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-8">No data</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Salary Summary */}
      <div className="px-4 lg:px-6 mt-2">
        <div className="flex items-center gap-2 mb-3">
          <DollarSignIcon className="h-5 w-5 text-muted-foreground" />
          <h2 className="text-lg font-semibold">Salary Summary</h2>
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-7 px-4 lg:px-6">
        <Card className={`xl:col-span-3 ${cardBg}`}>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Salary Overview</CardTitle>
          </CardHeader>
          <CardContent>
            {salaryData ? (
              <div className="space-y-5">
                <div className="grid grid-cols-2 gap-4">
                  <div className="rounded-lg bg-green-50 dark:bg-green-950/30 p-4 text-center">
                    <p className="text-xs text-muted-foreground mb-1">Total Gross</p>
                    <p className="text-xl font-bold text-green-600 tabular-nums">৳{salaryData.total_gross.toLocaleString()}</p>
                  </div>
                  <div className="rounded-lg bg-blue-50 dark:bg-blue-950/30 p-4 text-center">
                    <p className="text-xs text-muted-foreground mb-1">Total Net</p>
                    <p className="text-xl font-bold text-blue-600 tabular-nums">৳{salaryData.total_net.toLocaleString()}</p>
                  </div>
                </div>
                <div className="rounded-lg bg-muted/50 p-4 text-center">
                  <p className="text-xs text-muted-foreground mb-1">Employees Processed</p>
                  <p className="text-xl font-bold tabular-nums">{salaryData.total_employees}</p>
                </div>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-8">No salary summary available. Process a salary month first.</p>
            )}
          </CardContent>
        </Card>

        <Card className={`xl:col-span-4 ${cardBg}`}>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Employees by Department</CardTitle>
          </CardHeader>
          <CardContent>
            {deptBarData.length > 0 ? (
              <ChartContainer config={{ employees: { label: "Employees", color: "#8884D8" } }}>
                <ResponsiveContainer width="100%" height={280}>
                  <BarChart data={deptBarData} margin={{ left: 0, right: 0, top: 5, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--border))" />
                    <XAxis dataKey="name" tickLine={false} axisLine={false} tick={{ fontSize: 11 }} />
                    <YAxis tickLine={false} axisLine={false} tick={{ fontSize: 12 }} />
                    <Tooltip content={<ChartTooltipContent />} />
                    <Bar dataKey="employees" radius={[6, 6, 0, 0]}>
                      {deptBarData.map((_, i) => (
                        <Cell key={i} fill={COLORS[i % COLORS.length]} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </ChartContainer>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-12">No department data</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Male & Female Chart */}
      <div className="px-4 lg:px-6 mt-2">
        <div className="flex items-center gap-2 mb-3">
          <UsersIcon className="h-5 w-5 text-muted-foreground" />
          <h2 className="text-lg font-semibold">Gender Distribution</h2>
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-7 px-4 lg:px-6 pb-6">
        <Card className={`xl:col-span-3 ${cardBg}`}>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Male / Female Ratio</CardTitle>
          </CardHeader>
          <CardContent>
            {genderPieData.length > 0 ? (
              <ChartContainer config={{}}>
                <ResponsiveContainer width="100%" height={300}>
                  <PieChart>
                    <Pie
                      data={genderPieData}
                      cx="50%"
                      cy="50%"
                      innerRadius={70}
                      outerRadius={110}
                      paddingAngle={3}
                      dataKey="value"
                      nameKey="name"
                      label={({ name, percent }) => `${name} ${((percent ?? 0) * 100).toFixed(0)}%`}
                      labelLine={false}
                    >
                      {genderPieData.map((_, index) => (
                        <Cell key={index} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(value) => [value, "employees"]} />
                  </PieChart>
                </ResponsiveContainer>
                <ChartLegend>
                  <ChartLegendContent />
                </ChartLegend>
              </ChartContainer>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-12">No gender data</p>
            )}
          </CardContent>
        </Card>

        <Card className={`xl:col-span-4 ${cardBg}`}>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Gender Trend</CardTitle>
          </CardHeader>
          <CardContent>
            {genderPieData.length > 0 ? (
              <ChartContainer config={{
                male: { label: "Male", color: "#8884D8" },
                female: { label: "Female", color: "#FF6384" },
              }}>
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart
                    data={[{ name: "Employees", male: genderPieData.find(g => g.name === "Male")?.value || 0, female: genderPieData.find(g => g.name === "Female")?.value || 0 }]}
                    margin={{ left: 0, right: 0, top: 5, bottom: 0 }}
                    layout="vertical"
                  >
                    <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="hsl(var(--border))" />
                    <XAxis type="number" tickLine={false} axisLine={false} tick={{ fontSize: 12 }} />
                    <YAxis type="category" dataKey="name" tickLine={false} axisLine={false} tick={{ fontSize: 12 }} />
                    <Tooltip content={<ChartTooltipContent />} />
                    <Bar dataKey="male" fill="#8884D8" radius={[0, 6, 6, 0]} />
                    <Bar dataKey="female" fill="#FF6384" radius={[0, 6, 6, 0]} />
                  </BarChart>
                </ResponsiveContainer>
                <ChartLegend>
                  <ChartLegendContent />
                </ChartLegend>
              </ChartContainer>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-12">No gender data</p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
