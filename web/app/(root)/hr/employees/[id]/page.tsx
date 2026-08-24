"use client"

import * as React from "react"
import { useRouter, useParams } from "next/navigation"
import {
  ArrowLeftIcon, Loader2, UserCircle, BanknoteIcon, CalendarCheckIcon,
  BriefcaseIcon, Phone, Mail, MapPin, Heart, CreditCard, Clock,
  Cake, Droplets, Users, Shield, CalendarDays, Building2, Fingerprint,
  ChevronRight, BadgePercent, TrendingUp, TrendingDown, UserCheck,
  Globe, Hash, BookUser, Palette, Mars, Venus, CheckCheck, X,
  Baby, Ban, RefreshCw, FileSpreadsheet, FileText, Download, Filter,
  BarChart3, Activity, PieChart as PieChartIcon, LineChart as LineChartIcon, Layers, Sparkles
} from "lucide-react"
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  Legend as RechartsLegend,
  PieChart,
  Pie,
  Cell,
  Line,
  ComposedChart
} from "recharts"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { employeeApi } from "@/lib/api"
import { cn, getUploadBaseUrl } from "@/lib/utils"
import { toast } from "sonner"

interface AddressItem {
  id: string
  name: string
  name_bn?: string
}

interface Employee {
  id: string
  employee_id: string
  punch_number: string
  name_en: string
  name_bn: string
  father_name: string
  mother_name: string
  date_of_birth: string
  gender: string
  blood_group: string
  marital_status: string
  religion: string
  nationality: string
  nid: string
  phone: string
  email: string
  present_address: string
  present_address_bn: string
  present_post_office: string | null
  present_post_code: string | null
  present_division_id: string | null
  present_district_id: string | null
  present_upazila_id: string | null
  present_union_id: string | null
  present_division?: AddressItem
  present_district?: AddressItem
  present_upazila?: AddressItem
  present_union?: AddressItem

  permanent_address: string
  permanent_address_bn: string
  permanent_post_office: string | null
  permanent_post_code: string | null
  permanent_division_id: string | null
  permanent_district_id: string | null
  permanent_upazila_id: string | null
  permanent_union_id: string | null
  permanent_division?: AddressItem
  permanent_district?: AddressItem
  permanent_upazila?: AddressItem
  permanent_union?: AddressItem

  spouse_name: string
  emergency_contact: string
  emergency_phone: string
  number_of_dependents: number
  company_id: string
  employee_type: string
  grade: string
  joining_date: string
  shift_id: string | null
  department_id: string | null
  section_id: string | null
  designation_id: string | null
  line_id: string | null
  group_id: string | null
  floor_id: string | null
  reports_to: string | null
  gross_salary: number
  basic_salary: number
  house_rent: number
  transport_allowance: number
  food_allowance: number
  medical_allowance: number
  other_allowance: number
  account_type: string
  account_number: string
  status: string
  over_time_status: boolean
  signature_url: string
  image_url: string
  created_at: string
  department?: { id: string; name: string }
  section_ref?: { id: string; name: string }
  designation_ref?: { id: string; name: string }
  line_ref?: { id: string; name: string }
  group_ref?: { id: string; name: string }
  floor_ref?: { id: string; name: string }
  shift?: { id: string; name: string }
  company?: { id: string; company_name_en: string }
}

interface AttendanceCount {
  status: string
  count: number
}

interface Salary {
  id: string
  basic_salary: number
  house_rent: number
  medical_allowance: number
  transport_allowance: number
  food_allowance: number
  other_allowance: number
  gross_salary: number
  absent_deduction: number
  other_deduction: number
  total_deductions: number
  overtime_hours: number
  overtime_rate: number
  overtime_amount: number
  attendance_bonus: number
  net_salary: number
  present_days: number
  absent_days: number
  late_days: number
  leave_days: number
  weekend_days: number
  total_days: number
  month: number
  year: number
  status: string
}

interface MonthSalaryRecord {
  month: number
  year: number
  month_name: string
  salary: Salary | null
}

interface MonthAttendanceRecord {
  month: number
  year: number
  month_name: string
  total_days: number
  present_days: number
  absent_days: number
  late_days: number
  leave_days: number
  weekend_days: number
  half_days: number
  breakdown: AttendanceCount[]
}

interface ProfileResponse {
  employee: Employee
  attendance: AttendanceCount[]
  salary?: Salary
  salary_history?: MonthSalaryRecord[]
  attendance_history?: MonthAttendanceRecord[]
}

const statusColors: Record<string, string> = {
  present: "text-green-600 bg-green-50 border-green-200 dark:bg-green-950/20 dark:border-green-800 dark:text-green-400",
  absent: "text-red-600 bg-red-50 border-red-200 dark:bg-red-950/20 dark:border-red-800 dark:text-red-400",
  late: "text-yellow-600 bg-yellow-50 border-yellow-200 dark:bg-yellow-950/20 dark:border-yellow-800 dark:text-yellow-400",
  on_leave: "text-blue-600 bg-blue-50 border-blue-200 dark:bg-blue-950/20 dark:border-blue-800 dark:text-blue-400",
  weekend: "text-purple-600 bg-purple-50 border-purple-200 dark:bg-purple-950/20 dark:border-purple-800 dark:text-purple-400",
  half_day: "text-orange-600 bg-orange-50 border-orange-200 dark:bg-orange-950/20 dark:border-orange-800 dark:text-orange-400",
}

const statusLabels: Record<string, string> = {
  present: "Present",
  absent: "Absent",
  late: "Late",
  on_leave: "On Leave",
  weekend: "Weekend",
  half_day: "Half Day",
}

const attendanceIcons: Record<string, React.ElementType> = {
  present: UserCheck,
  absent: X,
  late: Clock,
  on_leave: CalendarDays,
  weekend: RefreshCw,
  half_day: Ban,
}

const monthsList = [
  { value: "all", label: "All Months" },
  { value: "1", label: "January" },
  { value: "2", label: "February" },
  { value: "3", label: "March" },
  { value: "4", label: "April" },
  { value: "5", label: "May" },
  { value: "6", label: "June" },
  { value: "7", label: "July" },
  { value: "8", label: "August" },
  { value: "9", label: "September" },
  { value: "10", label: "October" },
  { value: "11", label: "November" },
  { value: "12", label: "December" },
]

function StatusDot({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    active: "bg-green-500",
    inactive: "bg-gray-400",
    terminated: "bg-red-500",
    suspended: "bg-yellow-500",
  }
  return <span className={cn("inline-block h-2.5 w-2.5 rounded-full", colorMap[status] || "bg-gray-400")} />
}

function InfoRow({
  label,
  value,
  icon: Icon,
  valueClassName
}: {
  label: string
  value: string | number | null | undefined
  icon?: React.ElementType
  valueClassName?: string
}) {
  return (
    <div className="flex items-start gap-2.5 group">
      {Icon && <Icon className="h-4 w-4 text-muted-foreground/50 mt-0.5 shrink-0 group-hover:text-muted-foreground/80 transition-colors" />}
      <div className="flex flex-col min-w-0">
        <span className="text-[11px] uppercase tracking-wider text-muted-foreground/60 font-medium">{label}</span>
        <span className={cn("text-sm font-medium truncate", valueClassName)}>{value ?? "—"}</span>
      </div>
    </div>
  )
}

function formatAddressEn(
  details?: string | null,
  union?: string | null,
  postOffice?: string | null,
  postCode?: string | null,
  upazila?: string | null,
  district?: string | null,
  division?: string | null
) {
  const parts: string[] = []
  if (details) parts.push(details)
  if (union) parts.push(`Union: ${union}`)
  if (postOffice && postCode) {
    parts.push(`Post Office: ${postOffice} - ${postCode}`)
  } else if (postOffice) {
    parts.push(`Post Office: ${postOffice}`)
  } else if (postCode) {
    parts.push(`Post Code: ${postCode}`)
  }
  if (upazila) parts.push(`Upazila: ${upazila}`)
  if (district) parts.push(`District: ${district}`)
  if (division) parts.push(`Division: ${division}`)
  return parts.length > 0 ? parts.join(", ") : "—"
}

function formatAddressBn(
  details?: string | null,
  union?: string | null,
  postOffice?: string | null,
  postCode?: string | null,
  upazila?: string | null,
  district?: string | null,
  division?: string | null
) {
  const parts: string[] = []
  if (details) parts.push(details)
  if (union) parts.push(`ইউনিয়ন: ${union}`)
  if (postOffice && postCode) {
    parts.push(`ডাকঘর: ${postOffice} - ${postCode}`)
  } else if (postOffice) {
    parts.push(`ডাকঘর: ${postOffice}`)
  } else if (postCode) {
    parts.push(`পোস্ট কোড: ${postCode}`)
  }
  if (upazila) parts.push(`উপজেলা: ${upazila}`)
  if (district) parts.push(`জেলা: ${district}`)
  if (division) parts.push(`বিভাগ: ${division}`)
  return parts.length > 0 ? parts.join(", ") : "—"
}

function StatCard({ label, value, icon: Icon, color }: { label: string; value: string | number; icon: React.ElementType; color: string }) {
  return (
    <div className={cn("flex items-center gap-3 rounded-xl border px-4 py-3 transition-all hover:shadow-sm", color)}>
      <div className={cn("flex h-9 w-9 items-center justify-center rounded-lg", color.split(" ")[0]?.replace("text-", "bg-")?.replace("green", "green") || "bg-muted")}>
        <Icon className="h-4 w-4" />
      </div>
      <div>
        <p className="text-xs text-muted-foreground">{label}</p>
        <p className="text-lg font-bold">{value}</p>
      </div>
    </div>
  )
}

export default function EmployeeProfilePage() {
  const router = useRouter()
  const params = useParams()
  const id = params.id as string

  const [profile, setProfile] = React.useState<ProfileResponse | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [activeTab, setActiveTab] = React.useState("info")
  const [exporting, setExporting] = React.useState<"excel" | "pdf" | null>(null)
  const [isImageOpen, setIsImageOpen] = React.useState(false)

  const currentYearStr = new Date().getFullYear().toString()
  const [salaryYearFilter, setSalaryYearFilter] = React.useState<string>(currentYearStr)
  const [salaryMonthFilter, setSalaryMonthFilter] = React.useState<string>("all")
  const [selectedSalaryKey, setSelectedSalaryKey] = React.useState<string>("")

  const [attendanceYearFilter, setAttendanceYearFilter] = React.useState<string>(currentYearStr)
  const [attendanceMonthFilter, setAttendanceMonthFilter] = React.useState<string>("all")
  const [selectedAttendanceKey, setSelectedAttendanceKey] = React.useState<string>("")

  React.useEffect(() => {
    async function fetchProfile() {
      try {
        const { data } = await employeeApi.getProfile(id)
        setProfile(data)
      } catch {
        setError("Failed to load employee profile")
      } finally {
        setLoading(false)
      }
    }
    fetchProfile()
  }, [id])

  const salary_history = profile?.salary_history || []
  const attendance_history = profile?.attendance_history || []

  // Salary Available Years
  const salaryAvailableYears = React.useMemo(() => {
    const set = new Set(salary_history.map((s) => s.year.toString()))
    set.add(currentYearStr)
    return Array.from(set).sort((a, b) => Number(b) - Number(a))
  }, [salary_history, currentYearStr])

  // Filtered Salary History
  const filteredSalaryHistory = React.useMemo(() => {
    return salary_history.filter((s) => {
      const matchYear = salaryYearFilter === "all" || s.year.toString() === salaryYearFilter
      const matchMonth = salaryMonthFilter === "all" || s.month.toString() === salaryMonthFilter
      return matchYear && matchMonth
    })
  }, [salary_history, salaryYearFilter, salaryMonthFilter])

  // Attendance Available Years
  const attendanceAvailableYears = React.useMemo(() => {
    const set = new Set(attendance_history.map((a) => a.year.toString()))
    set.add(currentYearStr)
    return Array.from(set).sort((a, b) => Number(b) - Number(a))
  }, [attendance_history, currentYearStr])

  // Filtered Attendance History
  const filteredAttendanceHistory = React.useMemo(() => {
    return attendance_history.filter((a) => {
      const matchYear = attendanceYearFilter === "all" || a.year.toString() === attendanceYearFilter
      const matchMonth = attendanceMonthFilter === "all" || a.month.toString() === attendanceMonthFilter
      return matchYear && matchMonth
    })
  }, [attendance_history, attendanceYearFilter, attendanceMonthFilter])

  // Chronological salary chart data
  const salaryChartData = React.useMemo(() => {
    return [...filteredSalaryHistory].reverse().map((item) => {
      const s = item.salary
      return {
        name: `${item.month_name?.slice(0, 3)} '${String(item.year).slice(-2)}`,
        fullMonth: `${item.month_name} ${item.year}`,
        net: s ? s.net_salary || 0 : 0,
        gross: s ? s.gross_salary || 0 : 0,
        basic: s ? s.basic_salary || 0 : 0,
        deductions: s ? s.total_deductions || 0 : 0,
        ot: s ? s.overtime_amount || 0 : 0,
        allowances: s
          ? (s.house_rent || 0) + (s.medical_allowance || 0) + (s.transport_allowance || 0) + (s.food_allowance || 0) + (s.other_allowance || 0)
          : 0,
      }
    })
  }, [filteredSalaryHistory])

  // Chronological attendance chart data
  const attendanceChartData = React.useMemo(() => {
    return [...filteredAttendanceHistory].reverse().map((item) => {
      const workingDays = item.present_days + item.absent_days + item.late_days
      const rate = workingDays > 0 ? Math.round(((item.present_days + item.late_days) / workingDays) * 100) : 0
      return {
        name: `${item.month_name?.slice(0, 3)} '${String(item.year).slice(-2)}`,
        fullMonth: `${item.month_name} ${item.year}`,
        present: item.present_days,
        absent: item.absent_days,
        late: item.late_days,
        leave: item.leave_days,
        weekend: item.weekend_days,
        total: item.total_days,
        rate: rate,
      }
    })
  }, [filteredAttendanceHistory])

  // Analytical KPIs
  const salaryKPIs = React.useMemo(() => {
    const generated = salary_history.filter((s) => s.salary !== null)
    const totalNet = generated.reduce((acc, s) => acc + (s.salary?.net_salary || 0), 0)
    const avgNet = generated.length > 0 ? Math.round(totalNet / generated.length) : 0
    let highestNet = 0
    let highestMonth = "—"
    generated.forEach((s) => {
      if (s.salary && s.salary.net_salary > highestNet) {
        highestNet = s.salary.net_salary
        highestMonth = `${s.month_name} ${s.year}`
      }
    })
    return {
      totalNet,
      avgNet,
      highestNet,
      highestMonth,
      generatedCount: generated.length,
    }
  }, [salary_history])

  const attendanceKPIs = React.useMemo(() => {
    const totalPresent = attendance_history.reduce((sum, a) => sum + a.present_days, 0)
    const totalAbsent = attendance_history.reduce((sum, a) => sum + a.absent_days, 0)
    const totalLate = attendance_history.reduce((sum, a) => sum + a.late_days, 0)
    const totalLeave = attendance_history.reduce((sum, a) => sum + a.leave_days, 0)
    const totalDays = attendance_history.reduce((sum, a) => sum + a.total_days, 0)
    const workingDays = totalPresent + totalAbsent + totalLate
    const avgRate = workingDays > 0 ? Math.round(((totalPresent + totalLate) / workingDays) * 100) : 0
    return {
      totalPresent,
      totalAbsent,
      totalLate,
      totalLeave,
      totalDays,
      avgRate,
    }
  }, [attendance_history])

  const downloadBlob = (data: Blob, filename: string, mime: string) => {
    const blob = new Blob([data], { type: mime })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  }

  const handleExport = async (format: "excel" | "pdf") => {
    setExporting(format)
    try {
      const res =
        format === "excel"
          ? await employeeApi.exportProfileExcel(id)
          : await employeeApi.exportProfilePdf(id)
      const ext = format === "excel" ? "xlsx" : "pdf"
      const mime =
        format === "excel"
          ? "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
          : "application/pdf"
      downloadBlob(res.data, `profile_${profile?.employee?.employee_id || id}.${ext}`, mime)
    } catch {
      toast.error(`Failed to export ${format.toUpperCase()}`)
    } finally {
      setExporting(null)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-10 w-10 animate-spin text-primary/60" />
          <p className="text-sm text-muted-foreground animate-pulse">Loading profile...</p>
        </div>
      </div>
    )
  }

  if (error || !profile) {
    return (
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6 px-4 lg:px-6">
        <Button variant="ghost" onClick={() => router.back()} className="w-fit -ml-2">
          <ArrowLeftIcon className="mr-2 h-4 w-4" /> Back
        </Button>
        <Card className="border-destructive/30 bg-destructive/5">
          <CardContent className="flex items-center gap-3 py-6">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-destructive/10">
              <X className="h-5 w-5 text-destructive" />
            </div>
            <p className="text-sm font-medium text-destructive">{error || "Employee not found"}</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  const { employee, attendance, salary } = profile

  const selectedSalaryRecord = filteredSalaryHistory.find(
    (s) => `${s.year}-${s.month}` === selectedSalaryKey
  ) || filteredSalaryHistory[0] || null

  const currentSalary = selectedSalaryRecord?.salary || null

  const earningsPieData: { name: string; value: number; color: string }[] = currentSalary
    ? [
        { name: "Basic Salary (50%)", value: currentSalary.basic_salary || 0, color: "#2563eb" },
        { name: "House Rent", value: currentSalary.house_rent || 0, color: "#3b82f6" },
        { name: "Medical", value: currentSalary.medical_allowance || 0, color: "#06b6d4" },
        { name: "Transport", value: currentSalary.transport_allowance || 0, color: "#10b981" },
        { name: "Food", value: currentSalary.food_allowance || 0, color: "#f59e0b" },
        { name: "Other Allow.", value: currentSalary.other_allowance || 0, color: "#8b5cf6" },
        { name: "Overtime", value: currentSalary.overtime_amount || 0, color: "#ec4899" },
        { name: "Att. Bonus", value: currentSalary.attendance_bonus || 0, color: "#14b8a6" },
      ].filter((item) => item.value > 0)
    : []

  const selectedAttendanceRecord = filteredAttendanceHistory.find(
    (a) => `${a.year}-${a.month}` === selectedAttendanceKey
  ) || filteredAttendanceHistory[0] || null

  const currentAttendanceCounts = selectedAttendanceRecord?.breakdown?.length
    ? selectedAttendanceRecord.breakdown
    : []

  const imageInitial = employee.name_en?.charAt(0)?.toUpperCase() || "?"
  const totalAttendance = attendance.reduce((sum, a) => sum + a.count, 0)
  const baseUrl = getUploadBaseUrl()
  const fullImageUrl = employee.image_url
    ? employee.image_url.startsWith("http")
      ? employee.image_url
      : employee.image_url.startsWith("/")
      ? `${baseUrl}${employee.image_url}`
      : `${baseUrl}/${employee.image_url}`
    : ""

  return (
    <div className="flex flex-col gap-4 pb-8 md:gap-6 md:pb-10">

      {/* Header */}
      <div className="relative overflow-hidden bg-gradient-to-b from-primary/5 via-primary/[0.02] to-background border-b">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,var(--color-primary)/0.08,transparent_50%)]" />
        <div className="relative px-4 pt-4 pb-6 lg:px-6 lg:pt-6 lg:pb-8">
          <Button variant="ghost" size="sm" onClick={() => router.back()} className="mb-3 -ml-2 text-muted-foreground hover:text-foreground">
            <ArrowLeftIcon className="mr-1.5 h-4 w-4" /> Back
          </Button>

          <div className="flex flex-col sm:flex-row sm:items-center gap-4 sm:gap-6">
            {/* Avatar */}
            <div className="shrink-0 self-start sm:self-center">
              {fullImageUrl ? (
                <div 
                  className="relative h-20 w-20 overflow-hidden rounded-2xl border-2 border-primary/20 shadow-sm lg:h-24 lg:w-24 bg-gradient-to-br from-primary/20 to-primary/10 flex items-center justify-center cursor-pointer hover:opacity-90 transition-opacity" 
                  onClick={() => setIsImageOpen(true)}
                >
                  <img
                    src={fullImageUrl}
                    alt={employee.name_en}
                    className="h-full w-full object-cover"
                    onError={(e) => {
                      const target = e.currentTarget
                      target.style.display = "none"
                      const parent = target.parentElement
                      if (parent && !parent.querySelector(".fallback-initial")) {
                        const span = document.createElement("span")
                        span.className = "fallback-initial text-2xl font-bold text-primary/70 lg:text-3xl"
                        span.innerText = imageInitial
                        parent.appendChild(span)
                      }
                    }}
                  />
                </div>
              ) : (
                <div className="flex h-20 w-20 items-center justify-center rounded-2xl bg-gradient-to-br from-primary/20 to-primary/10 border-2 border-primary/20 shadow-sm lg:h-24 lg:w-24">
                  <span className="text-2xl font-bold text-primary/70 lg:text-3xl">{imageInitial}</span>
                </div>
              )}
            </div>

            <div className="flex-1 min-w-0">
              <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3">
                <h1 className="text-2xl font-bold tracking-tight lg:text-3xl truncate">{employee.name_en}</h1>
                <div className="flex items-center gap-2 shrink-0">
                  <Badge
                    variant={employee.status === "active" ? "default" : "secondary"}
                    className="gap-1.5 capitalize px-3 py-1 text-[11px]"
                  >
                    <StatusDot status={employee.status} />
                    {employee.status}
                  </Badge>
                  {employee.over_time_status && (
                    <Badge variant="outline" className="gap-1 text-[11px] border-amber-300 text-amber-700 dark:border-amber-700 dark:text-amber-400">
                      <Clock className="h-3 w-3" /> OT
                    </Badge>
                  )}
                </div>
              </div>
              <p className="text-sm text-muted-foreground mt-1.5">
                <span className="font-medium text-foreground/80">{employee.employee_id}</span>
                {employee.designation_ref?.name && (
                  <>
                    <span className="mx-2 text-muted-foreground/40">•</span>
                    {employee.designation_ref.name}
                  </>
                )}
                {employee.department?.name && (
                  <>
                    <span className="mx-2 text-muted-foreground/40">•</span>
                    {employee.department.name}
                  </>
                )}
              </p>
              <div className="flex flex-wrap items-center gap-x-4 gap-y-1 mt-2 text-xs text-muted-foreground/70">
                {employee.company?.company_name_en && (
                  <span className="flex items-center gap-1">
                    <Building2 className="h-3 w-3" /> {employee.company.company_name_en}
                  </span>
                )}
                {employee.joining_date && (
                  <span className="flex items-center gap-1">
                    <CalendarDays className="h-3 w-3" /> Joined {new Date(employee.joining_date).toLocaleDateString("en-GB", { day: "numeric", month: "short", year: "numeric" })}
                  </span>
                )}
                {employee.grade && (
                  <span className="flex items-center gap-1">
                    <BadgePercent className="h-3 w-3" /> Grade {employee.grade}
                  </span>
                )}
              </div>
            </div>

            <div className="shrink-0 self-start sm:self-center flex flex-wrap items-center gap-2">
              <div className="flex items-center gap-1.5">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleExport("excel")}
                  disabled={exporting !== null}
                  className="gap-1.5"
                >
                  {exporting === "excel" ? <Loader2 className="h-4 w-4 animate-spin" /> : <FileSpreadsheet className="h-4 w-4 text-green-600 dark:text-green-400" />}
                  <span className="hidden sm:inline">Excel</span>
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleExport("pdf")}
                  disabled={exporting !== null}
                  className="gap-1.5"
                >
                  {exporting === "pdf" ? <Loader2 className="h-4 w-4 animate-spin" /> : <FileText className="h-4 w-4 text-red-600 dark:text-red-400" />}
                  <span className="hidden sm:inline">PDF</span>
                </Button>
              </div>
              <Button size="sm" onClick={() => router.push(`/hr/employees/${id}/edit`)} className="gap-2">
                <UserCircle className="h-4 w-4" /> Edit
              </Button>
            </div>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        {/* Quick Stats */}
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 -mt-3 mb-6">
          <StatCard
            label="Gross Salary"
            value={`৳${(employee.gross_salary || 0).toLocaleString()}`}
            icon={BanknoteIcon}
            color="border-green-200 dark:border-green-800"
          />
          <StatCard
            label="Present"
            value={attendance.find(a => a.status === "present")?.count || 0}
            icon={UserCheck}
            color="border-green-200 dark:border-green-800"
          />
          <StatCard
            label="Absent"
            value={attendance.find(a => a.status === "absent")?.count || 0}
            icon={X}
            color="border-red-200 dark:border-red-800"
          />
          <StatCard
            label="Late"
            value={attendance.find(a => a.status === "late")?.count || 0}
            icon={Clock}
            color="border-yellow-200 dark:border-yellow-800"
          />
          <StatCard
            label="On Leave"
            value={attendance.find(a => a.status === "on_leave")?.count || 0}
            icon={CalendarDays}
            color="border-blue-200 dark:border-blue-800"
          />
          <StatCard
            label="Weekend"
            value={attendance.find(a => a.status === "weekend")?.count || 0}
            icon={RefreshCw}
            color="border-purple-200 dark:border-purple-800"
          />
        </div>

        {/* Tabs */}
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="w-full lg:w-fit">
            <TabsTrigger value="info" className="gap-2">
              <BookUser className="h-4 w-4" /> <span className="hidden sm:inline">Employee</span> Info
            </TabsTrigger>
            <TabsTrigger value="salary" className="gap-2">
              <BanknoteIcon className="h-4 w-4" /> Salary
            </TabsTrigger>
            <TabsTrigger value="attendance" className="gap-2">
              <CalendarCheckIcon className="h-4 w-4" /> Attendance
            </TabsTrigger>
            <TabsTrigger value="analytics" className="gap-2">
              <Activity className="h-4 w-4 text-purple-500" /> Advance Charts
            </TabsTrigger>
          </TabsList>

          {/* === Employee Info Tab === */}
          <TabsContent value="info" className="mt-6 space-y-6">
            
            {/* 1. All-in-One Personal Information Card (Personal, Contact, Family & Emergency) */}
            <Card className="overflow-hidden border-t-4 border-t-primary shadow-sm">
              <CardHeader className="bg-muted/30 pb-3 border-b">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-1">
                  <CardTitle className="text-base flex items-center gap-2 font-semibold">
                    <UserCircle className="h-5 w-5 text-primary" /> Personal Information
                  </CardTitle>
                  <span className="text-xs text-muted-foreground font-normal">
                    ব্যক্তিগত, যোগাযোগ ও পারিবারিক তথ্য
                  </span>
                </div>
              </CardHeader>
              <CardContent className="p-5 space-y-6">
                
                {/* Section A: Identity & Demographics */}
                <div>
                  <h4 className="text-xs font-semibold uppercase tracking-wider text-primary/80 mb-3 flex items-center gap-1.5">
                    <UserCheck className="h-3.5 w-3.5" /> Identity & Demographics (ব্যক্তিগত পরিচিতি)
                  </h4>
                  <div className="grid gap-3.5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 bg-muted/20 p-4 rounded-xl border border-muted-foreground/10">
                    <InfoRow label="Full Name (EN)" value={employee.name_en} icon={UserCircle} />
                    <InfoRow label="Full Name (BN)" value={employee.name_bn} icon={Globe} valueClassName="sutonnymj !text-base font-normal tracking-wide" />
                    <InfoRow label="Father's Name" value={employee.father_name} icon={Users} />
                    <InfoRow label="Mother's Name" value={employee.mother_name} icon={Users} />
                    <InfoRow label="Date of Birth" value={employee.date_of_birth?.split("T")[0]} icon={Cake} />
                    <InfoRow label="Gender" value={employee.gender} icon={employee.gender?.toLowerCase() === "male" ? Mars : Venus} />
                    <InfoRow label="Blood Group" value={employee.blood_group} icon={Droplets} />
                    <InfoRow label="Marital Status" value={employee.marital_status} icon={Heart} />
                    <InfoRow label="Religion" value={employee.religion} icon={Palette} />
                    <InfoRow label="Nationality" value={employee.nationality} icon={Globe} />
                    <InfoRow label="NID Number" value={employee.nid} icon={Fingerprint} />
                  </div>
                </div>

                {/* Section B: Contact & Family / Emergency Details */}
                <div className="grid gap-5 md:grid-cols-2">
                  {/* Contact Information Sub-block */}
                  <div className="rounded-xl border border-blue-200/60 dark:border-blue-900/50 bg-blue-50/25 dark:bg-blue-950/15 p-4 space-y-3">
                    <h4 className="text-xs font-semibold uppercase tracking-wider text-blue-600 dark:text-blue-400 flex items-center gap-1.5">
                      <Phone className="h-3.5 w-3.5" /> Contact Details (যোগাযোগ)
                    </h4>
                    <div className="grid gap-3 sm:grid-cols-2 pt-1">
                      <InfoRow label="Phone Number" value={employee.phone} icon={Phone} />
                      <InfoRow label="Email Address" value={employee.email} icon={Mail} />
                    </div>
                  </div>

                  {/* Family & Emergency Sub-block */}
                  <div className="rounded-xl border border-rose-200/60 dark:border-rose-900/50 bg-rose-50/25 dark:bg-rose-950/15 p-4 space-y-3">
                    <h4 className="text-xs font-semibold uppercase tracking-wider text-rose-600 dark:text-rose-400 flex items-center gap-1.5">
                      <Heart className="h-3.5 w-3.5" /> Family & Emergency (পারিবারিক ও জরুরি)
                    </h4>
                    <div className="grid gap-3 sm:grid-cols-2 pt-1">
                      <InfoRow label="Spouse Name" value={employee.spouse_name} icon={Heart} />
                      <InfoRow label="Emergency Contact" value={employee.emergency_contact} icon={Users} />
                      <InfoRow label="Emergency Phone" value={employee.emergency_phone} icon={Phone} />
                      <InfoRow label="Number of Dependents" value={employee.number_of_dependents} icon={Baby} />
                    </div>
                  </div>
                </div>

              </CardContent>
            </Card>

            {/* 2. Address Information (Present & Permanent side-by-side) */}
            <div className="grid gap-6 lg:grid-cols-2">
              
              {/* Present Address */}
              <Card className="overflow-hidden border-t-4 border-t-indigo-500 shadow-sm">
                <CardHeader className="bg-muted/30 pb-3 border-b">
                  <CardTitle className="text-sm flex items-center justify-between">
                    <span className="flex items-center gap-2 font-semibold">
                      <MapPin className="h-4 w-4 text-indigo-500" /> Present Address
                    </span>
                    <span className="text-xs text-muted-foreground font-normal">বর্তমান ঠিকানা</span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4 pt-4">
                  {/* Full Address Banner */}
                  <div className="rounded-xl bg-indigo-50/60 dark:bg-indigo-950/25 p-3.5 border border-indigo-100 dark:border-indigo-900/50 space-y-2">
                    <div>
                      <p className="text-[10px] uppercase font-bold tracking-wider text-indigo-700 dark:text-indigo-400">Full Address (English)</p>
                      <p className="text-xs font-medium text-foreground/90 mt-0.5 leading-relaxed">
                        {formatAddressEn(
                          employee.present_address,
                          employee.present_union?.name,
                          employee.present_post_office,
                          employee.present_post_code,
                          employee.present_upazila?.name,
                          employee.present_district?.name,
                          employee.present_division?.name
                        )}
                      </p>
                    </div>
                    <div className="border-t border-indigo-200/50 dark:border-indigo-800/50 pt-2">
                      <p className="text-[10px] uppercase font-bold tracking-wider text-indigo-700 dark:text-indigo-400">পূর্ণ ঠিকানা (বাংলা)</p>
                      <p className="text-xs font-medium text-foreground/90 mt-0.5 leading-relaxed">
                        {employee.present_address_bn && (
                          <span className="sutonnymj text-sm mr-1.5">{employee.present_address_bn},</span>
                        )}
                        {formatAddressBn(
                          null,
                          employee.present_union?.name_bn || employee.present_union?.name,
                          employee.present_post_office,
                          employee.present_post_code,
                          employee.present_upazila?.name_bn || employee.present_upazila?.name,
                          employee.present_district?.name_bn || employee.present_district?.name,
                          employee.present_division?.name_bn || employee.present_division?.name
                        )}
                      </p>
                    </div>
                  </div>

                  {/* Hierarchical Fields */}
                  <div className="grid grid-cols-2 gap-3 pt-1">
                    <InfoRow
                      label="Division (বিভাগ)"
                      value={employee.present_division ? `${employee.present_division.name}${employee.present_division.name_bn ? ` (${employee.present_division.name_bn})` : ""}` : null}
                    />
                    <InfoRow
                      label="District (জেলা)"
                      value={employee.present_district ? `${employee.present_district.name}${employee.present_district.name_bn ? ` (${employee.present_district.name_bn})` : ""}` : null}
                    />
                    <InfoRow
                      label="Upazila (উপজেলা)"
                      value={employee.present_upazila ? `${employee.present_upazila.name}${employee.present_upazila.name_bn ? ` (${employee.present_upazila.name_bn})` : ""}` : null}
                    />
                    <InfoRow
                      label="Union (ইউনিয়ন)"
                      value={employee.present_union ? `${employee.present_union.name}${employee.present_union.name_bn ? ` (${employee.present_union.name_bn})` : ""}` : null}
                    />
                    <InfoRow label="Post Office (ডাকঘর)" value={employee.present_post_office} />
                    <InfoRow label="Post Code (পোস্ট কোড)" value={employee.present_post_code} />
                  </div>

                  {/* Details En & Bn */}
                  <div className="border-t pt-3 space-y-2">
                    <div>
                      <span className="text-[11px] uppercase tracking-wider text-muted-foreground/60 font-medium">Address Details (EN)</span>
                      <p className="text-xs font-medium text-foreground/80 mt-0.5">{employee.present_address || "—"}</p>
                    </div>
                    <div>
                      <span className="text-[11px] uppercase tracking-wider text-muted-foreground/60 font-medium">ঠিকানার বিবরণ (বাংলা)</span>
                      <p className="text-base font-normal text-foreground/90 mt-0.5 sutonnymj">{employee.present_address_bn || "—"}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Permanent Address */}
              <Card className="overflow-hidden border-t-4 border-t-teal-500 shadow-sm">
                <CardHeader className="bg-muted/30 pb-3 border-b">
                  <CardTitle className="text-sm flex items-center justify-between">
                    <span className="flex items-center gap-2 font-semibold">
                      <MapPin className="h-4 w-4 text-teal-500" /> Permanent Address
                    </span>
                    <span className="text-xs text-muted-foreground font-normal">স্থায়ী ঠিকানা</span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4 pt-4">
                  {/* Full Address Banner */}
                  <div className="rounded-xl bg-teal-50/60 dark:bg-teal-950/25 p-3.5 border border-teal-100 dark:border-teal-900/50 space-y-2">
                    <div>
                      <p className="text-[10px] uppercase font-bold tracking-wider text-teal-700 dark:text-teal-400">Full Address (English)</p>
                      <p className="text-xs font-medium text-foreground/90 mt-0.5 leading-relaxed">
                        {formatAddressEn(
                          employee.permanent_address,
                          employee.permanent_union?.name,
                          employee.permanent_post_office,
                          employee.permanent_post_code,
                          employee.permanent_upazila?.name,
                          employee.permanent_district?.name,
                          employee.permanent_division?.name
                        )}
                      </p>
                    </div>
                    <div className="border-t border-teal-200/50 dark:border-teal-800/50 pt-2">
                      <p className="text-[10px] uppercase font-bold tracking-wider text-teal-700 dark:text-teal-400">পূর্ণ ঠিকানা (বাংলা)</p>
                      <p className="text-xs font-medium text-foreground/90 mt-0.5 leading-relaxed">
                        {employee.permanent_address_bn && (
                          <span className="sutonnymj text-sm mr-1.5">{employee.permanent_address_bn},</span>
                        )}
                        {formatAddressBn(
                          null,
                          employee.permanent_union?.name_bn || employee.permanent_union?.name,
                          employee.permanent_post_office,
                          employee.permanent_post_code,
                          employee.permanent_upazila?.name_bn || employee.permanent_upazila?.name,
                          employee.permanent_district?.name_bn || employee.permanent_district?.name,
                          employee.permanent_division?.name_bn || employee.permanent_division?.name
                        )}
                      </p>
                    </div>
                  </div>

                  {/* Hierarchical Fields */}
                  <div className="grid grid-cols-2 gap-3 pt-1">
                    <InfoRow
                      label="Division (বিভাগ)"
                      value={employee.permanent_division ? `${employee.permanent_division.name}${employee.permanent_division.name_bn ? ` (${employee.permanent_division.name_bn})` : ""}` : null}
                    />
                    <InfoRow
                      label="District (জেলা)"
                      value={employee.permanent_district ? `${employee.permanent_district.name}${employee.permanent_district.name_bn ? ` (${employee.permanent_district.name_bn})` : ""}` : null}
                    />
                    <InfoRow
                      label="Upazila (উপজেলা)"
                      value={employee.permanent_upazila ? `${employee.permanent_upazila.name}${employee.permanent_upazila.name_bn ? ` (${employee.permanent_upazila.name_bn})` : ""}` : null}
                    />
                    <InfoRow
                      label="Union (ইউনিয়ন)"
                      value={employee.permanent_union ? `${employee.permanent_union.name}${employee.permanent_union.name_bn ? ` (${employee.permanent_union.name_bn})` : ""}` : null}
                    />
                    <InfoRow label="Post Office (ডাকঘর)" value={employee.permanent_post_office} />
                    <InfoRow label="Post Code (পোস্ট কোড)" value={employee.permanent_post_code} />
                  </div>

                  {/* Details En & Bn */}
                  <div className="border-t pt-3 space-y-2">
                    <div>
                      <span className="text-[11px] uppercase tracking-wider text-muted-foreground/60 font-medium">Address Details (EN)</span>
                      <p className="text-xs font-medium text-foreground/80 mt-0.5">{employee.permanent_address || "—"}</p>
                    </div>
                    <div>
                      <span className="text-[11px] uppercase tracking-wider text-muted-foreground/60 font-medium">ঠিকানার বিবরণ (বাংলা)</span>
                      <p className="text-base font-normal text-foreground/90 mt-0.5 sutonnymj">{employee.permanent_address_bn || "—"}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>

            </div>

            {/* 3. Office & Financial Information */}
            <div className="grid gap-6 lg:grid-cols-3">
              
              {/* Office Information (2 cols) */}
              <Card className="overflow-hidden border-t-4 border-t-amber-500 shadow-sm lg:col-span-2">
                <CardHeader className="bg-muted/30 pb-3 border-b">
                  <CardTitle className="text-sm flex items-center justify-between">
                    <span className="flex items-center gap-2 font-semibold">
                      <BriefcaseIcon className="h-4 w-4 text-amber-500" /> Office Information
                    </span>
                    <span className="text-xs text-muted-foreground font-normal">দাপ্তরিক তথ্য</span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-4">
                  <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    <InfoRow label="Employee ID" value={employee.employee_id} icon={Hash} />
                    <InfoRow label="Punch Number" value={employee.punch_number} icon={Fingerprint} />
                    <InfoRow label="Employee Type" value={employee.employee_type} icon={BriefcaseIcon} />
                    <InfoRow label="Grade" value={employee.grade} icon={BadgePercent} />
                    <InfoRow label="Joining Date" value={employee.joining_date?.split("T")[0]} icon={CalendarDays} />
                    <InfoRow label="Company" value={employee.company?.company_name_en} icon={Building2} />
                    <InfoRow label="Department" value={employee.department?.name} icon={Building2} />
                    <InfoRow label="Section" value={employee.section_ref?.name} icon={Building2} />
                    <InfoRow label="Designation" value={employee.designation_ref?.name} icon={Shield} />
                    <InfoRow label="Line" value={employee.line_ref?.name} icon={ChevronRight} />
                    <InfoRow label="Group" value={employee.group_ref?.name} icon={Users} />
                    <InfoRow label="Floor" value={employee.floor_ref?.name} icon={Building2} />
                    <InfoRow label="Shift" value={employee.shift?.name} icon={Clock} />
                    <InfoRow label="Over Time" value={employee.over_time_status ? "Enabled" : "Disabled"} icon={Clock} />
                  </div>
                </CardContent>
              </Card>

              {/* Bank Account (1 col) */}
              <Card className="overflow-hidden border-t-4 border-t-emerald-500 shadow-sm lg:col-span-1">
                <CardHeader className="bg-muted/30 pb-3 border-b">
                  <CardTitle className="text-sm flex items-center justify-between">
                    <span className="flex items-center gap-2 font-semibold">
                      <CreditCard className="h-4 w-4 text-emerald-500" /> Bank Account
                    </span>
                    <span className="text-xs text-muted-foreground font-normal">ব্যাংক হিসাব</span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="grid gap-4 pt-4">
                  <InfoRow label="Account Type" value={employee.account_type} icon={CreditCard} />
                  <InfoRow label="Account Number" value={employee.account_number} icon={Hash} />
                </CardContent>
              </Card>

            </div>

          </TabsContent>

          {/* === Salary Tab === */}
          <TabsContent value="salary" className="mt-6 space-y-6">
            
            {/* Filter Controls Bar */}
            <div className="flex flex-wrap items-center justify-between gap-3 bg-muted/30 p-3.5 rounded-xl border shadow-sm">
              <div className="flex items-center gap-2">
                <Filter className="h-4 w-4 text-primary" />
                <span className="text-xs font-semibold text-foreground">Filter Salary Timeline:</span>
              </div>
              <div className="flex flex-wrap items-center gap-2.5">
                <div className="flex items-center gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Year:</label>
                  <select
                    value={salaryYearFilter}
                    onChange={(e) => setSalaryYearFilter(e.target.value)}
                    className="h-8 rounded-lg border border-input bg-background px-2.5 text-xs font-medium focus:outline-none focus:ring-1 focus:ring-primary shadow-sm"
                  >
                    <option value="all">All Years</option>
                    {salaryAvailableYears.map((y) => (
                      <option key={y} value={y}>{y}</option>
                    ))}
                  </select>
                </div>

                <div className="flex items-center gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Month:</label>
                  <select
                    value={salaryMonthFilter}
                    onChange={(e) => setSalaryMonthFilter(e.target.value)}
                    className="h-8 rounded-lg border border-input bg-background px-2.5 text-xs font-medium focus:outline-none focus:ring-1 focus:ring-primary shadow-sm"
                  >
                    {monthsList.map((m) => (
                      <option key={m.value} value={m.value}>{m.label}</option>
                    ))}
                  </select>
                </div>

                {(salaryYearFilter !== currentYearStr || salaryMonthFilter !== "all") && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 px-2.5 text-xs text-muted-foreground hover:text-foreground"
                    onClick={() => {
                      setSalaryYearFilter(currentYearStr)
                      setSalaryMonthFilter("all")
                    }}
                  >
                    Reset Filter
                  </Button>
                )}
              </div>
            </div>

            {/* Top Stats Overview */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
              <Card className="p-4 flex flex-col justify-between border-l-4 border-l-primary shadow-sm">
                <span className="text-xs text-muted-foreground font-medium">Joining Date</span>
                <span className="text-base font-bold text-foreground mt-1 flex items-center gap-1.5">
                  <CalendarDays className="h-4 w-4 text-primary" />
                  {employee.joining_date?.split("T")[0] || "—"}
                </span>
                <span className="text-[10px] text-muted-foreground mt-0.5">Start of tenure</span>
              </Card>

              <Card className="p-4 flex flex-col justify-between border-l-4 border-l-blue-500 shadow-sm">
                <span className="text-xs text-muted-foreground font-medium">Displayed Months</span>
                <span className="text-2xl font-bold text-blue-600 dark:text-blue-400 mt-1">
                  {filteredSalaryHistory.length}
                </span>
                <span className="text-[10px] text-muted-foreground mt-0.5">Matching filter</span>
              </Card>

              <Card className="p-4 flex flex-col justify-between border-l-4 border-l-emerald-500 shadow-sm">
                <span className="text-xs text-muted-foreground font-medium">Generated Salaries</span>
                <span className="text-2xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">
                  {filteredSalaryHistory.filter((s) => s.salary !== null).length}
                </span>
                <span className="text-[10px] text-muted-foreground mt-0.5">Processed & available</span>
              </Card>

              <Card className="p-4 flex flex-col justify-between border-l-4 border-l-amber-500 shadow-sm">
                <span className="text-xs text-muted-foreground font-medium">Missing / Unprocessed</span>
                <span className="text-2xl font-bold text-amber-600 dark:text-amber-400 mt-1">
                  {filteredSalaryHistory.filter((s) => s.salary === null).length}
                </span>
                <span className="text-[10px] text-muted-foreground mt-0.5">null records</span>
              </Card>
            </div>

            {/* Visual Salary Trajectory Chart */}
            {salaryChartData.length > 0 && (
              <Card className="shadow-sm overflow-hidden border-t-2 border-t-emerald-500">
                <CardHeader className="bg-muted/30 pb-3 border-b flex flex-row items-center justify-between">
                  <div>
                    <CardTitle className="text-sm font-semibold flex items-center gap-2">
                      <LineChartIcon className="h-4 w-4 text-emerald-500" /> Salary Trajectory & Pay Trends
                    </CardTitle>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      Net salary vs gross compensation over chronological timeline.
                    </p>
                  </div>
                </CardHeader>
                <CardContent className="p-4 pt-6">
                  <div className="h-[250px] w-full">
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart data={salaryChartData} margin={{ top: 10, right: 20, left: 0, bottom: 0 }}>
                        <defs>
                          <linearGradient id="salaryNetGradTab" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#10b981" stopOpacity={0.4} />
                            <stop offset="95%" stopColor="#10b981" stopOpacity={0.0} />
                          </linearGradient>
                          <linearGradient id="salaryGrossGradTab" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.2} />
                            <stop offset="95%" stopColor="#3b82f6" stopOpacity={0.0} />
                          </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" className="stroke-muted/40" />
                        <XAxis dataKey="name" tick={{ fontSize: 11 }} tickLine={false} />
                        <YAxis tick={{ fontSize: 11 }} tickLine={false} tickFormatter={(val) => `৳${val >= 1000 ? `${(val/1000).toFixed(0)}k` : val}`} />
                        <RechartsTooltip
                          formatter={(value: any, name: any) => [`৳${Number(value || 0).toLocaleString()}`, name === "net" ? "Net Salary" : name === "gross" ? "Gross Salary" : name === "basic" ? "Basic Salary" : "Deductions"]}
                          labelFormatter={(label, payload) => payload?.[0]?.payload?.fullMonth || label}
                          contentStyle={{ backgroundColor: "rgba(15, 23, 42, 0.95)", borderColor: "rgba(255,255,255,0.1)", borderRadius: "8px", color: "#fff", fontSize: "12px" }}
                        />
                        <RechartsLegend verticalAlign="top" height={32} iconType="circle" />
                        <Area type="monotone" dataKey="gross" stroke="#3b82f6" strokeWidth={2} fillOpacity={1} fill="url(#salaryGrossGradTab)" name="Gross Salary" />
                        <Area type="monotone" dataKey="net" stroke="#10b981" strokeWidth={2.5} fillOpacity={1} fill="url(#salaryNetGradTab)" name="Net Salary" />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Monthly History Table */}
            <Card className="shadow-sm overflow-hidden">
              <CardHeader className="bg-muted/30 pb-3 border-b flex flex-row items-center justify-between">
                <div>
                  <CardTitle className="text-sm font-semibold flex items-center gap-2">
                    <BanknoteIcon className="h-4 w-4 text-primary" /> Monthly Salary History ({salaryYearFilter === "all" ? "All Years" : salaryYearFilter})
                  </CardTitle>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Full month-by-month salary timeline. Unprocessed months display as <span className="font-mono text-amber-600 dark:text-amber-400 font-semibold">null</span>.
                  </p>
                </div>
              </CardHeader>
              <div className="overflow-x-auto">
                <table className="w-full text-xs text-left">
                  <thead className="bg-muted/50 text-muted-foreground font-medium border-b uppercase text-[10px] tracking-wider">
                    <tr>
                      <th className="py-2.5 px-4">Month / Year</th>
                      <th className="py-2.5 px-3">Gross Salary</th>
                      <th className="py-2.5 px-3">Basic (50%)</th>
                      <th className="py-2.5 px-3">Allowances</th>
                      <th className="py-2.5 px-3">OT Amount</th>
                      <th className="py-2.5 px-3">Deductions</th>
                      <th className="py-2.5 px-3">Net Salary</th>
                      <th className="py-2.5 px-3">Status</th>
                      <th className="py-2.5 px-3 text-right">Action</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {filteredSalaryHistory.length > 0 ? (
                      filteredSalaryHistory.map((item) => {
                        const itemKey = `${item.year}-${item.month}`
                        const isSelected = selectedSalaryRecord && `${selectedSalaryRecord.year}-${selectedSalaryRecord.month}` === itemKey
                        const s = item.salary
                        const totalAllowances = s 
                          ? (s.house_rent || 0) + (s.medical_allowance || 0) + (s.transport_allowance || 0) + (s.food_allowance || 0) + (s.other_allowance || 0)
                          : 0

                        return (
                          <tr 
                            key={itemKey} 
                            className={cn(
                              "transition-colors hover:bg-muted/40 cursor-pointer",
                              isSelected ? "bg-primary/5 font-medium" : ""
                            )}
                            onClick={() => setSelectedSalaryKey(itemKey)}
                          >
                            <td className="py-3 px-4 font-semibold text-foreground flex items-center gap-2">
                              <span className={cn("h-2 w-2 rounded-full shrink-0", s ? "bg-emerald-500" : "bg-amber-400")} />
                              {item.month_name} {item.year}
                            </td>
                            <td className="py-3 px-3">
                              {s ? `৳${(s.gross_salary || 0).toLocaleString()}` : <span className="text-muted-foreground/60 italic font-mono text-[11px]">null</span>}
                            </td>
                            <td className="py-3 px-3">
                              {s ? `৳${(s.basic_salary || 0).toLocaleString()}` : <span className="text-muted-foreground/60 italic font-mono text-[11px]">null</span>}
                            </td>
                            <td className="py-3 px-3">
                              {s ? `৳${totalAllowances.toLocaleString()}` : <span className="text-muted-foreground/60 italic font-mono text-[11px]">null</span>}
                            </td>
                            <td className="py-3 px-3">
                              {s ? `৳${(s.overtime_amount || 0).toLocaleString()}` : <span className="text-muted-foreground/60 italic font-mono text-[11px]">null</span>}
                            </td>
                            <td className="py-3 px-3">
                              {s ? (
                                <span className="text-red-500 font-medium">-৳{(s.total_deductions || 0).toLocaleString()}</span>
                              ) : (
                                <span className="text-muted-foreground/60 italic font-mono text-[11px]">null</span>
                              )}
                            </td>
                            <td className="py-3 px-3">
                              {s ? (
                                <span className="text-emerald-600 dark:text-emerald-400 font-bold text-sm">৳{(s.net_salary || 0).toLocaleString()}</span>
                              ) : (
                                <span className="text-amber-600/80 dark:text-amber-400/80 italic font-mono text-xs">null</span>
                              )}
                            </td>
                            <td className="py-3 px-3">
                              {s ? (
                                <Badge className="bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30 text-[10px] capitalize">
                                  {s.status || "Generated"}
                                </Badge>
                              ) : (
                                <Badge variant="outline" className="bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30 text-[10px]">
                                  null (Not Generated)
                                </Badge>
                              )}
                            </td>
                            <td className="py-3 px-3 text-right">
                              <Button 
                                size="sm" 
                                variant={isSelected ? "default" : "outline"} 
                                className="h-7 text-xs px-2.5"
                                onClick={(e) => {
                                  e.stopPropagation()
                                  setSelectedSalaryKey(itemKey)
                                }}
                              >
                                {isSelected ? "Selected" : "View"}
                              </Button>
                            </td>
                          </tr>
                        )
                      })
                    ) : (
                      <tr>
                        <td colSpan={9} className="py-8 text-center text-muted-foreground">
                          No salary records match the selected Year / Month filter
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </Card>

            {/* Selected Month Detailed Breakdown */}
            {currentSalary ? (
              <div className="space-y-4 pt-2">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-semibold flex items-center gap-2 text-primary">
                    <CheckCheck className="h-4 w-4" /> Detailed Breakdown: {selectedSalaryRecord?.month_name} {selectedSalaryRecord?.year}
                  </h3>
                  <Badge variant="secondary" className="text-xs">
                    Status: <span className="capitalize font-semibold ml-1">{currentSalary.status || "Paid"}</span>
                  </Badge>
                </div>

                <div className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
                  {/* Net Salary Highlight */}
                  <Card className="md:col-span-2 xl:col-span-3 overflow-hidden border-t-2 border-t-primary/30 bg-gradient-to-r from-primary/[0.03] to-transparent">
                    <CardContent className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-6">
                      <div>
                        <p className="text-xs uppercase tracking-wider text-muted-foreground/60 font-medium">Net Salary ({selectedSalaryRecord?.month_name} {selectedSalaryRecord?.year})</p>
                        <p className="text-3xl font-bold text-primary">৳{(currentSalary.net_salary || 0).toLocaleString()}</p>
                        <p className="text-xs text-muted-foreground mt-1">
                          {currentSalary.month}/{currentSalary.year} • {currentSalary.total_days} total days in month
                        </p>
                      </div>
                      <div className="flex gap-4 text-center">
                        <div>
                          <p className="text-2xl font-bold text-green-600">{currentSalary.present_days}</p>
                          <p className="text-[11px] text-muted-foreground">Present</p>
                        </div>
                        <div className="w-px bg-border" />
                        <div>
                          <p className="text-2xl font-bold text-red-600">{currentSalary.absent_days}</p>
                          <p className="text-[11px] text-muted-foreground">Absent</p>
                        </div>
                        <div className="w-px bg-border" />
                        <div>
                          <p className="text-2xl font-bold text-yellow-600">{currentSalary.late_days}</p>
                          <p className="text-[11px] text-muted-foreground">Late</p>
                        </div>
                        <div className="w-px bg-border" />
                        <div>
                          <p className="text-2xl font-bold text-blue-600">{currentSalary.leave_days}</p>
                          <p className="text-[11px] text-muted-foreground">Leave</p>
                        </div>
                      </div>
                    </CardContent>
                  </Card>

                  {/* Earnings */}
                  <Card className="overflow-hidden border-t-2 border-t-green-500/20">
                    <CardHeader className="bg-muted/30 pb-3">
                      <CardTitle className="text-sm flex items-center gap-2">
                        <TrendingUp className="h-4 w-4 text-green-500" /> Earnings Breakdown
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="grid gap-4 pt-4">
                      <InfoRow label="Gross Salary" value={`৳${(currentSalary.gross_salary || 0).toLocaleString()}`} icon={BanknoteIcon} />
                      <div className="border-t my-1" />
                      <InfoRow label="Basic Salary (50%)" value={`৳${(currentSalary.basic_salary || 0).toLocaleString()}`} icon={BanknoteIcon} />
                      <InfoRow label="House Rent (25%)" value={`৳${(currentSalary.house_rent || 0).toLocaleString()}`} icon={BanknoteIcon} />
                      <InfoRow label="Medical Allowance" value={`৳${(currentSalary.medical_allowance || 0).toLocaleString()}`} icon={BanknoteIcon} />
                      <InfoRow label="Transport Allowance" value={`৳${(currentSalary.transport_allowance || 0).toLocaleString()}`} icon={BanknoteIcon} />
                      <InfoRow label="Food Allowance" value={`৳${(currentSalary.food_allowance || 0).toLocaleString()}`} icon={BanknoteIcon} />
                      <InfoRow label="Other Allowance" value={`৳${(currentSalary.other_allowance || 0).toLocaleString()}`} icon={BanknoteIcon} />
                      <div className="border-t my-1" />
                      <InfoRow label="Overtime Amount" value={`৳${(currentSalary.overtime_amount || 0).toLocaleString()}`} icon={Clock} />
                      <InfoRow label="Attendance Bonus" value={`৳${(currentSalary.attendance_bonus || 0).toLocaleString()}`} icon={CheckCheck} />
                    </CardContent>
                  </Card>

                  {/* Deductions */}
                  <Card className="overflow-hidden border-t-2 border-t-red-500/20">
                    <CardHeader className="bg-muted/30 pb-3">
                      <CardTitle className="text-sm flex items-center gap-2">
                        <TrendingDown className="h-4 w-4 text-red-500" /> Deductions
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="grid gap-4 pt-4">
                      <InfoRow label="Absent Deduction" value={`৳${(currentSalary.absent_deduction || 0).toLocaleString()}`} icon={TrendingDown} />
                      <InfoRow label="Other Deduction" value={`৳${(currentSalary.other_deduction || 0).toLocaleString()}`} icon={TrendingDown} />
                      <div className="border-t my-1" />
                      <div className="flex items-center justify-between rounded-lg bg-red-50 dark:bg-red-950/20 px-3 py-2.5 -mx-1">
                        <span className="text-xs font-medium text-red-700 dark:text-red-400">Total Deductions</span>
                        <span className="text-sm font-bold text-red-600 dark:text-red-400">৳{(currentSalary.total_deductions || 0).toLocaleString()}</span>
                      </div>
                    </CardContent>
                  </Card>

                  {/* Attendance Summary */}
                  <Card className="overflow-hidden border-t-2 border-t-purple-500/20">
                    <CardHeader className="bg-muted/30 pb-3">
                      <CardTitle className="text-sm flex items-center gap-2">
                        <CalendarCheckIcon className="h-4 w-4 text-purple-500" /> Attendance Summary
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="grid gap-4 pt-4">
                      <InfoRow label="Month" value={`${currentSalary.month}/${currentSalary.year}`} icon={CalendarDays} />
                      <div className="border-t my-1" />
                      <InfoRow label="Total Days" value={currentSalary.total_days} icon={CalendarDays} />
                      <InfoRow label="Present Days" value={currentSalary.present_days} icon={UserCheck} />
                      <InfoRow label="Absent Days" value={currentSalary.absent_days} icon={X} />
                      <InfoRow label="Late Days" value={currentSalary.late_days} icon={Clock} />
                      <InfoRow label="Leave Days" value={currentSalary.leave_days} icon={CalendarDays} />
                      <InfoRow label="Weekend / Holiday" value={currentSalary.weekend_days} icon={RefreshCw} />
                    </CardContent>
                  </Card>
                </div>
              </div>
            ) : selectedSalaryRecord ? (
              <Card className="border-dashed bg-muted/10">
                <CardContent className="flex flex-col items-center gap-3 py-10">
                  <BanknoteIcon className="h-9 w-9 text-muted-foreground/40" />
                  <p className="text-sm font-medium text-foreground">
                    No salary record for {selectedSalaryRecord.month_name} {selectedSalaryRecord.year}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Salary has not been generated for this month yet. Values are reported as <span className="font-mono text-amber-600 dark:text-amber-400 font-semibold">null</span>.
                  </p>
                </CardContent>
              </Card>
            ) : null}
          </TabsContent>

          {/* === Attendance Tab === */}
          <TabsContent value="attendance" className="mt-6 space-y-6">

            {/* Filter Controls Bar */}
            <div className="flex flex-wrap items-center justify-between gap-3 bg-muted/30 p-3.5 rounded-xl border shadow-sm">
              <div className="flex items-center gap-2">
                <Filter className="h-4 w-4 text-primary" />
                <span className="text-xs font-semibold text-foreground">Filter Attendance:</span>
              </div>
              <div className="flex flex-wrap items-center gap-2.5">
                <div className="flex items-center gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Year:</label>
                  <select
                    value={attendanceYearFilter}
                    onChange={(e) => setAttendanceYearFilter(e.target.value)}
                    className="h-8 rounded-lg border border-input bg-background px-2.5 text-xs font-medium focus:outline-none focus:ring-1 focus:ring-primary shadow-sm"
                  >
                    <option value="all">All Years</option>
                    {attendanceAvailableYears.map((y) => (
                      <option key={y} value={y}>{y}</option>
                    ))}
                  </select>
                </div>

                <div className="flex items-center gap-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Month:</label>
                  <select
                    value={attendanceMonthFilter}
                    onChange={(e) => setAttendanceMonthFilter(e.target.value)}
                    className="h-8 rounded-lg border border-input bg-background px-2.5 text-xs font-medium focus:outline-none focus:ring-1 focus:ring-primary shadow-sm"
                  >
                    {monthsList.map((m) => (
                      <option key={m.value} value={m.value}>{m.label}</option>
                    ))}
                  </select>
                </div>

                {(attendanceYearFilter !== currentYearStr || attendanceMonthFilter !== "all") && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 px-2.5 text-xs text-muted-foreground hover:text-foreground"
                    onClick={() => {
                      setAttendanceYearFilter(currentYearStr)
                      setAttendanceMonthFilter("all")
                    }}
                  >
                    Reset Filter
                  </Button>
                )}
              </div>
            </div>

            {/* Top Attendance Stats */}
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
              <Card className="p-3.5 border-l-4 border-l-blue-500 shadow-sm">
                <span className="text-[11px] text-muted-foreground font-medium uppercase">Tracked Months</span>
                <p className="text-xl font-bold text-foreground mt-1">{filteredAttendanceHistory.length}</p>
                <span className="text-[10px] text-muted-foreground">Matching filter</span>
              </Card>

              <Card className="p-3.5 border-l-4 border-l-green-500 shadow-sm">
                <span className="text-[11px] text-muted-foreground font-medium uppercase">Total Present</span>
                <p className="text-xl font-bold text-green-600 mt-1">
                  {filteredAttendanceHistory.reduce((sum, a) => sum + a.present_days, 0)}
                </p>
                <span className="text-[10px] text-muted-foreground">Matching filter</span>
              </Card>

              <Card className="p-3.5 border-l-4 border-l-red-500 shadow-sm">
                <span className="text-[11px] text-muted-foreground font-medium uppercase">Total Absent</span>
                <p className="text-xl font-bold text-red-600 mt-1">
                  {filteredAttendanceHistory.reduce((sum, a) => sum + a.absent_days, 0)}
                </p>
                <span className="text-[10px] text-muted-foreground">Matching filter</span>
              </Card>

              <Card className="p-3.5 border-l-4 border-l-yellow-500 shadow-sm">
                <span className="text-[11px] text-muted-foreground font-medium uppercase">Total Late</span>
                <p className="text-xl font-bold text-yellow-600 mt-1">
                  {filteredAttendanceHistory.reduce((sum, a) => sum + a.late_days, 0)}
                </p>
                <span className="text-[10px] text-muted-foreground">Matching filter</span>
              </Card>

              <Card className="p-3.5 border-l-4 border-l-blue-400 shadow-sm">
                <span className="text-[11px] text-muted-foreground font-medium uppercase">Total Leave</span>
                <p className="text-xl font-bold text-blue-500 mt-1">
                  {filteredAttendanceHistory.reduce((sum, a) => sum + a.leave_days, 0)}
                </p>
                <span className="text-[10px] text-muted-foreground">Matching filter</span>
              </Card>

              <Card className="p-3.5 border-l-4 border-l-purple-500 shadow-sm">
                <span className="text-[11px] text-muted-foreground font-medium uppercase">Weekend / Hol</span>
                <p className="text-xl font-bold text-purple-600 mt-1">
                  {filteredAttendanceHistory.reduce((sum, a) => sum + a.weekend_days, 0)}
                </p>
                <span className="text-[10px] text-muted-foreground">Matching filter</span>
              </Card>
            </div>

            {/* Visual Attendance Trend Chart */}
            {attendanceChartData.length > 0 && (
              <Card className="shadow-sm overflow-hidden border-t-2 border-t-blue-500">
                <CardHeader className="bg-muted/30 pb-3 border-b flex flex-row items-center justify-between">
                  <div>
                    <CardTitle className="text-sm font-semibold flex items-center gap-2">
                      <BarChart3 className="h-4 w-4 text-blue-500" /> Attendance Dynamics & Distribution
                    </CardTitle>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      Monthly counts for Present, Absent, Late, and Leave days across tenure.
                    </p>
                  </div>
                </CardHeader>
                <CardContent className="p-4 pt-6">
                  <div className="h-[250px] w-full">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart data={attendanceChartData} margin={{ top: 10, right: 20, left: 0, bottom: 0 }}>
                        <CartesianGrid strokeDasharray="3 3" className="stroke-muted/40" />
                        <XAxis dataKey="name" tick={{ fontSize: 11 }} tickLine={false} />
                        <YAxis tick={{ fontSize: 11 }} tickLine={false} />
                        <RechartsTooltip
                          formatter={(value: any, name: any) => [value, String(name).toUpperCase()]}
                          labelFormatter={(label, payload) => payload?.[0]?.payload?.fullMonth || label}
                          contentStyle={{ backgroundColor: "rgba(15, 23, 42, 0.95)", borderColor: "rgba(255,255,255,0.1)", borderRadius: "8px", color: "#fff", fontSize: "12px" }}
                        />
                        <RechartsLegend verticalAlign="top" height={32} iconType="circle" />
                        <Bar dataKey="present" fill="#16a34a" stackId="a" name="Present" radius={[0, 0, 0, 0]} />
                        <Bar dataKey="late" fill="#eab308" stackId="a" name="Late" />
                        <Bar dataKey="leave" fill="#3b82f6" stackId="a" name="Leave" />
                        <Bar dataKey="absent" fill="#ef4444" stackId="a" name="Absent" radius={[4, 4, 0, 0]} />
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Monthly Attendance Timeline Table */}
            <Card className="shadow-sm overflow-hidden">
              <CardHeader className="bg-muted/30 pb-3 border-b flex flex-row items-center justify-between">
                <div>
                  <CardTitle className="text-sm font-semibold flex items-center gap-2">
                    <CalendarCheckIcon className="h-4 w-4 text-primary" /> Monthly Attendance Summary ({attendanceYearFilter === "all" ? "All Years" : attendanceYearFilter})
                  </CardTitle>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Aggregated total attendance records across all months of employment.
                  </p>
                </div>
              </CardHeader>
              <div className="overflow-x-auto">
                <table className="w-full text-xs text-left">
                  <thead className="bg-muted/50 text-muted-foreground font-medium border-b uppercase text-[10px] tracking-wider">
                    <tr>
                      <th className="py-2.5 px-4">Month / Year</th>
                      <th className="py-2.5 px-3">Total Days</th>
                      <th className="py-2.5 px-3">Present</th>
                      <th className="py-2.5 px-3">Absent</th>
                      <th className="py-2.5 px-3">Late</th>
                      <th className="py-2.5 px-3">Leave</th>
                      <th className="py-2.5 px-3">Weekend</th>
                      <th className="py-2.5 px-3">Presence Rate</th>
                      <th className="py-2.5 px-3 text-right">Action</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {filteredAttendanceHistory.length > 0 ? (
                      filteredAttendanceHistory.map((item) => {
                        const itemKey = `${item.year}-${item.month}`
                        const isSelected = selectedAttendanceRecord && `${selectedAttendanceRecord.year}-${selectedAttendanceRecord.month}` === itemKey
                        const workingDays = item.present_days + item.absent_days + item.late_days
                        const rate = workingDays > 0 ? Math.round(((item.present_days + item.late_days) / workingDays) * 100) : 0

                        return (
                          <tr 
                            key={itemKey} 
                            className={cn(
                              "transition-colors hover:bg-muted/40 cursor-pointer",
                              isSelected ? "bg-primary/5 font-medium" : ""
                            )}
                            onClick={() => setSelectedAttendanceKey(itemKey)}
                          >
                            <td className="py-3 px-4 font-semibold text-foreground flex items-center gap-2">
                              <span className={cn("h-2 w-2 rounded-full shrink-0", item.total_days > 0 ? "bg-green-500" : "bg-muted-foreground/30")} />
                              {item.month_name} {item.year}
                            </td>
                            <td className="py-3 px-3 font-semibold">
                              {item.total_days > 0 ? item.total_days : <span className="text-muted-foreground/60 italic font-mono text-[11px]">0</span>}
                            </td>
                            <td className="py-3 px-3 font-semibold text-green-600 dark:text-green-400">
                              {item.present_days}
                            </td>
                            <td className="py-3 px-3 font-semibold text-red-600 dark:text-red-400">
                              {item.absent_days}
                            </td>
                            <td className="py-3 px-3 font-semibold text-yellow-600 dark:text-yellow-400">
                              {item.late_days}
                            </td>
                            <td className="py-3 px-3 font-semibold text-blue-600 dark:text-blue-400">
                              {item.leave_days}
                            </td>
                            <td className="py-3 px-3 font-semibold text-purple-600 dark:text-purple-400">
                              {item.weekend_days}
                            </td>
                            <td className="py-3 px-3">
                              {item.total_days > 0 ? (
                                <Badge variant="outline" className={cn(
                                  "text-[10px] font-semibold",
                                  rate >= 90 ? "bg-green-50 text-green-700 border-green-200" : rate >= 75 ? "bg-yellow-50 text-yellow-700 border-yellow-200" : "bg-red-50 text-red-700 border-red-200"
                                )}>
                                  {rate}%
                                </Badge>
                              ) : (
                                <Badge variant="outline" className="text-[10px] text-muted-foreground bg-muted/20">
                                  No Records
                                </Badge>
                              )}
                            </td>
                            <td className="py-3 px-3 text-right">
                              <Button 
                                size="sm" 
                                variant={isSelected ? "default" : "outline"} 
                                className="h-7 text-xs px-2.5"
                                onClick={(e) => {
                                  e.stopPropagation()
                                  setSelectedAttendanceKey(itemKey)
                                }}
                              >
                                {isSelected ? "Selected" : "View"}
                              </Button>
                            </td>
                          </tr>
                        )
                      })
                    ) : (
                      <tr>
                        <td colSpan={9} className="py-8 text-center text-muted-foreground">
                          No attendance records match the selected Year / Month filter
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </Card>

            {/* Selected Month Cards */}
            {selectedAttendanceRecord && (
              <div className="space-y-3 pt-2">
                <div className="flex items-center justify-between">
                  <p className="text-sm font-semibold text-foreground flex items-center gap-2">
                    <CalendarDays className="h-4 w-4 text-primary" /> Breakdown for {selectedAttendanceRecord.month_name} {selectedAttendanceRecord.year}
                    <span className="text-xs text-muted-foreground font-normal">
                      ({selectedAttendanceRecord.total_days} total records)
                    </span>
                  </p>
                </div>

                {currentAttendanceCounts.length > 0 ? (
                  <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
                    {currentAttendanceCounts.map((a) => {
                      const Icon = attendanceIcons[a.status] || CalendarCheckIcon
                      return (
                        <Card key={a.status} className={cn("border-2 transition-all hover:shadow-md hover:-translate-y-0.5", statusColors[a.status] || "")}>
                          <CardContent className="flex flex-col items-center justify-center py-6 gap-2">
                            <Icon className="h-6 w-6 opacity-60" />
                            <span className="text-3xl font-bold tabular-nums">{a.count}</span>
                            <span className="text-xs font-medium uppercase tracking-wider opacity-70">
                              {statusLabels[a.status] || a.status}
                            </span>
                          </CardContent>
                        </Card>
                      )
                    })}
                  </div>
                ) : (
                  <Card className="border-dashed bg-muted/10">
                    <CardContent className="flex flex-col items-center gap-3 py-10">
                      <CalendarCheckIcon className="h-9 w-9 text-muted-foreground/40" />
                      <p className="text-sm text-muted-foreground">
                        No attendance records found for {selectedAttendanceRecord.month_name} {selectedAttendanceRecord.year}
                      </p>
                    </CardContent>
                  </Card>
                )}
              </div>
            )}

          </TabsContent>

          {/* === Advance Charts & Analytics Tab === */}
          <TabsContent value="analytics" className="mt-6 space-y-6">
            {/* Analytical KPI Cards */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
              <Card className="p-4 flex flex-col justify-between border-l-4 border-l-emerald-500 shadow-sm bg-gradient-to-br from-emerald-500/[0.04] to-transparent">
                <span className="text-xs text-muted-foreground font-medium flex items-center gap-1.5">
                  <BanknoteIcon className="h-3.5 w-3.5 text-emerald-600" /> Avg Monthly Net Pay
                </span>
                <span className="text-2xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">
                  ৳{salaryKPIs.avgNet.toLocaleString()}
                </span>
                <span className="text-[10px] text-muted-foreground mt-0.5">Over {salaryKPIs.generatedCount} generated months</span>
              </Card>

              <Card className="p-4 flex flex-col justify-between border-l-4 border-l-blue-500 shadow-sm bg-gradient-to-br from-blue-500/[0.04] to-transparent">
                <span className="text-xs text-muted-foreground font-medium flex items-center gap-1.5">
                  <TrendingUp className="h-3.5 w-3.5 text-blue-600" /> Highest Net Salary
                </span>
                <span className="text-2xl font-bold text-blue-600 dark:text-blue-400 mt-1">
                  ৳{salaryKPIs.highestNet.toLocaleString()}
                </span>
                <span className="text-[10px] text-muted-foreground mt-0.5">{salaryKPIs.highestMonth}</span>
              </Card>

              <Card className="p-4 flex flex-col justify-between border-l-4 border-l-purple-500 shadow-sm bg-gradient-to-br from-purple-500/[0.04] to-transparent">
                <span className="text-xs text-muted-foreground font-medium flex items-center gap-1.5">
                  <Activity className="h-3.5 w-3.5 text-purple-600" /> Avg Presence Rate
                </span>
                <span className="text-2xl font-bold text-purple-600 dark:text-purple-400 mt-1">
                  {attendanceKPIs.avgRate}%
                </span>
                <span className="text-[10px] text-muted-foreground mt-0.5">{attendanceKPIs.totalPresent} present days recorded</span>
              </Card>

              <Card className="p-4 flex flex-col justify-between border-l-4 border-l-amber-500 shadow-sm bg-gradient-to-br from-amber-500/[0.04] to-transparent">
                <span className="text-xs text-muted-foreground font-medium flex items-center gap-1.5">
                  <Sparkles className="h-3.5 w-3.5 text-amber-600" /> Career Total Pay
                </span>
                <span className="text-2xl font-bold text-amber-600 dark:text-amber-400 mt-1">
                  ৳{salaryKPIs.totalNet.toLocaleString()}
                </span>
                <span className="text-[10px] text-muted-foreground mt-0.5">Disbursed net compensation</span>
              </Card>
            </div>

            {/* Main Charts Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Chart 1: Salary Dynamics */}
              <Card className="shadow-sm overflow-hidden border-t-2 border-t-emerald-500">
                <CardHeader className="bg-muted/30 pb-3 border-b flex flex-row items-center justify-between">
                  <div>
                    <CardTitle className="text-sm font-semibold flex items-center gap-2">
                      <TrendingUp className="h-4 w-4 text-emerald-500" /> Compensation & Net Salary Trajectory
                    </CardTitle>
                    <p className="text-xs text-muted-foreground mt-0.5">Chronological net vs gross pay progression</p>
                  </div>
                </CardHeader>
                <CardContent className="p-4 pt-6">
                  <div className="h-[300px] w-full">
                    {salaryChartData.length > 0 ? (
                      <ResponsiveContainer width="100%" height="100%">
                        <AreaChart data={salaryChartData} margin={{ top: 10, right: 20, left: 0, bottom: 0 }}>
                          <defs>
                            <linearGradient id="advNetGrad" x1="0" y1="0" x2="0" y2="1">
                              <stop offset="5%" stopColor="#10b981" stopOpacity={0.4} />
                              <stop offset="95%" stopColor="#10b981" stopOpacity={0.0} />
                            </linearGradient>
                          </defs>
                          <CartesianGrid strokeDasharray="3 3" className="stroke-muted/40" />
                          <XAxis dataKey="name" tick={{ fontSize: 11 }} tickLine={false} />
                          <YAxis tick={{ fontSize: 11 }} tickLine={false} tickFormatter={(v) => `৳${v >= 1000 ? `${(v/1000).toFixed(0)}k` : v}`} />
                          <RechartsTooltip
                            formatter={(value: any, name: any) => [`৳${Number(value || 0).toLocaleString()}`, name === "net" ? "Net Salary" : name === "gross" ? "Gross Salary" : "Basic Salary"]}
                            labelFormatter={(label, payload) => payload?.[0]?.payload?.fullMonth || label}
                            contentStyle={{ backgroundColor: "rgba(15, 23, 42, 0.95)", borderColor: "rgba(255,255,255,0.1)", borderRadius: "8px", color: "#fff", fontSize: "12px" }}
                          />
                          <RechartsLegend verticalAlign="top" height={36} iconType="circle" />
                          <Area type="monotone" dataKey="gross" stroke="#3b82f6" strokeWidth={2} fill="#3b82f6" fillOpacity={0.1} name="Gross Salary" />
                          <Area type="monotone" dataKey="net" stroke="#10b981" strokeWidth={2.5} fillOpacity={1} fill="url(#advNetGrad)" name="Net Salary" />
                          <Line type="monotone" dataKey="basic" stroke="#8b5cf6" strokeWidth={1.5} strokeDasharray="3 3" name="Basic (50%)" dot={false} />
                        </AreaChart>
                      </ResponsiveContainer>
                    ) : (
                      <div className="flex h-full items-center justify-center text-xs text-muted-foreground">No salary data recorded yet</div>
                    )}
                  </div>
                </CardContent>
              </Card>

              {/* Chart 2: Attendance & Presence Rate */}
              <Card className="shadow-sm overflow-hidden border-t-2 border-t-purple-500">
                <CardHeader className="bg-muted/30 pb-3 border-b flex flex-row items-center justify-between">
                  <div>
                    <CardTitle className="text-sm font-semibold flex items-center gap-2">
                      <CalendarCheckIcon className="h-4 w-4 text-purple-500" /> Attendance Breakdown & Presence Rate
                    </CardTitle>
                    <p className="text-xs text-muted-foreground mt-0.5">Monthly presence vs absence and punctuality rate</p>
                  </div>
                </CardHeader>
                <CardContent className="p-4 pt-6">
                  <div className="h-[300px] w-full">
                    {attendanceChartData.length > 0 ? (
                      <ResponsiveContainer width="100%" height="100%">
                        <ComposedChart data={attendanceChartData} margin={{ top: 10, right: 20, left: 0, bottom: 0 }}>
                          <CartesianGrid strokeDasharray="3 3" className="stroke-muted/40" />
                          <XAxis dataKey="name" tick={{ fontSize: 11 }} tickLine={false} />
                          <YAxis yAxisId="left" tick={{ fontSize: 11 }} tickLine={false} />
                          <YAxis yAxisId="right" orientation="right" tick={{ fontSize: 11 }} tickLine={false} tickFormatter={(v) => `${v}%`} domain={[0, 100]} />
                          <RechartsTooltip
                            formatter={(value: any, name: any) => [name === "rate" ? `${value}%` : value, name === "rate" ? "Presence Rate" : String(name).toUpperCase()]}
                            labelFormatter={(label, payload) => payload?.[0]?.payload?.fullMonth || label}
                            contentStyle={{ backgroundColor: "rgba(15, 23, 42, 0.95)", borderColor: "rgba(255,255,255,0.1)", borderRadius: "8px", color: "#fff", fontSize: "12px" }}
                          />
                          <RechartsLegend verticalAlign="top" height={36} iconType="circle" />
                          <Bar yAxisId="left" dataKey="present" fill="#16a34a" stackId="a" name="Present" />
                          <Bar yAxisId="left" dataKey="late" fill="#eab308" stackId="a" name="Late" />
                          <Bar yAxisId="left" dataKey="leave" fill="#3b82f6" stackId="a" name="Leave" />
                          <Bar yAxisId="left" dataKey="absent" fill="#ef4444" stackId="a" name="Absent" radius={[4, 4, 0, 0]} />
                          <Line yAxisId="right" type="monotone" dataKey="rate" stroke="#8b5cf6" strokeWidth={2.5} name="Presence Rate %" dot={{ r: 3 }} />
                        </ComposedChart>
                      </ResponsiveContainer>
                    ) : (
                      <div className="flex h-full items-center justify-center text-xs text-muted-foreground">No attendance records yet</div>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Earnings Composition Donut Section */}
            {earningsPieData.length > 0 && (
              <Card className="shadow-sm overflow-hidden border-t-2 border-t-blue-500">
                <CardHeader className="bg-muted/30 pb-3 border-b flex flex-row items-center justify-between">
                  <div>
                    <CardTitle className="text-sm font-semibold flex items-center gap-2">
                      <PieChartIcon className="h-4 w-4 text-blue-500" /> Salary Allowances & Earnings Composition ({selectedSalaryRecord?.month_name} {selectedSalaryRecord?.year})
                    </CardTitle>
                    <p className="text-xs text-muted-foreground mt-0.5">Component share of gross salary earnings</p>
                  </div>
                </CardHeader>
                <CardContent className="p-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-center">
                    <div className="h-[240px] w-full flex items-center justify-center">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={earningsPieData}
                            dataKey="value"
                            nameKey="name"
                            cx="50%"
                            cy="50%"
                            innerRadius={55}
                            outerRadius={85}
                            paddingAngle={3}
                          >
                            {earningsPieData.map((entry, index) => (
                              <Cell key={`cell-${index}`} fill={entry.color} />
                            ))}
                          </Pie>
                          <RechartsTooltip
                            formatter={(value: any, name: any) => [`৳${Number(value).toLocaleString()}`, name]}
                            contentStyle={{ backgroundColor: "rgba(15, 23, 42, 0.95)", borderColor: "rgba(255,255,255,0.1)", borderRadius: "8px", color: "#fff", fontSize: "12px" }}
                          />
                        </PieChart>
                      </ResponsiveContainer>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      {earningsPieData.map((item) => (
                        <div key={item.name} className="flex items-center gap-2 p-2.5 rounded-lg border bg-muted/20">
                          <span className="h-3 w-3 rounded-full shrink-0" style={{ backgroundColor: item.color }} />
                          <div className="min-w-0">
                            <p className="text-[11px] text-muted-foreground truncate">{item.name}</p>
                            <p className="text-xs font-bold text-foreground">৳{item.value.toLocaleString()}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
          </TabsContent>
        </Tabs>
      </div>

      {/* Lightbox Overlay */}
      <div 
        className={cn(
          "fixed inset-0 z-[100] flex items-center justify-center bg-black/80 p-4 transition-all duration-300 ease-out",
          isImageOpen ? "opacity-100 pointer-events-auto" : "opacity-0 pointer-events-none"
        )}
        onClick={() => setIsImageOpen(false)}
      >
        <div 
          className={cn(
            "relative max-w-3xl w-full max-h-[90vh] flex items-center justify-center transition-transform duration-300 ease-out",
            isImageOpen ? "scale-100 translate-y-0" : "scale-95 translate-y-4"
          )}
          onClick={(e) => e.stopPropagation()}
        >
          {fullImageUrl && (
             <img 
               src={fullImageUrl} 
               alt={employee?.name_en || "Profile"} 
               className="max-w-full max-h-[90vh] object-contain rounded-xl shadow-2xl"
             />
          )}
          <button 
            className="absolute -top-12 right-0 p-2 text-white/70 hover:text-white hover:bg-white/10 rounded-full transition-colors"
            onClick={() => setIsImageOpen(false)}
          >
            <X className="w-6 h-6" />
          </button>
        </div>
      </div>
    </div>
  )
}
