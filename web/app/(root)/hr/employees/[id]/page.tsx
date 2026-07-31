"use client"

import * as React from "react"
import { useRouter, useParams } from "next/navigation"
import {
  ArrowLeftIcon, Loader2, UserCircle, BanknoteIcon, CalendarCheckIcon,
  BriefcaseIcon, Phone, Mail, MapPin, Heart, CreditCard, Clock,
  Cake, Droplets, Users, Shield, CalendarDays, Building2, Fingerprint,
  ChevronRight, BadgePercent, TrendingUp, TrendingDown, UserCheck,
  Globe, Hash, BookUser, Palette, Mars, Venus, CheckCheck, X,
  Baby, Ban, RefreshCw, FileSpreadsheet, FileText, Download
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { employeeApi } from "@/lib/api"
import { cn } from "@/lib/utils"
import { toast } from "sonner"

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
  permanent_address: string
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

interface ProfileResponse {
  employee: Employee
  attendance: AttendanceCount[]
  salary?: Salary
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

function StatusDot({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    active: "bg-green-500",
    inactive: "bg-gray-400",
    terminated: "bg-red-500",
    suspended: "bg-yellow-500",
  }
  return <span className={cn("inline-block h-2.5 w-2.5 rounded-full", colorMap[status] || "bg-gray-400")} />
}

function InfoRow({ label, value, icon: Icon }: { label: string; value: string | number | null | undefined; icon?: React.ElementType }) {
  return (
    <div className="flex items-start gap-2.5 group">
      {Icon && <Icon className="h-4 w-4 text-muted-foreground/50 mt-0.5 shrink-0 group-hover:text-muted-foreground/80 transition-colors" />}
      <div className="flex flex-col min-w-0">
        <span className="text-[11px] uppercase tracking-wider text-muted-foreground/60 font-medium">{label}</span>
        <span className="text-sm font-medium truncate">{value ?? "—"}</span>
      </div>
    </div>
  )
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
      downloadBlob(res.data, `profile_${employee?.employee_id || id}.${ext}`, mime)
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
  const imageInitial = employee.name_en?.charAt(0)?.toUpperCase() || "?"
  const totalAttendance = attendance.reduce((sum, a) => sum + a.count, 0)

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
              {employee.image_url ? (
                <div className="relative h-20 w-20 overflow-hidden rounded-2xl border-2 border-border/50 shadow-sm lg:h-24 lg:w-24">
                  <img src={employee.image_url} alt={employee.name_en} className="h-full w-full object-cover" />
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
          </TabsList>

          {/* === Employee Info Tab === */}
          <TabsContent value="info" className="mt-6 space-y-6">
            <div className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">

              {/* Personal Information */}
              <Card className="overflow-hidden border-t-2 border-t-primary/20">
                <CardHeader className="bg-muted/30 pb-3">
                  <CardTitle className="text-sm flex items-center gap-2">
                    <UserCircle className="h-4 w-4 text-primary" /> Personal Information
                  </CardTitle>
                </CardHeader>
                <CardContent className="grid gap-4 pt-4">
                  <InfoRow label="Full Name (EN)" value={employee.name_en} icon={UserCircle} />
                  <InfoRow label="Full Name (BN)" value={employee.name_bn} icon={Globe} />
                  <InfoRow label="Father's Name" value={employee.father_name} icon={Users} />
                  <InfoRow label="Mother's Name" value={employee.mother_name} icon={Users} />
                  <InfoRow label="Date of Birth" value={employee.date_of_birth?.split("T")[0]} icon={Cake} />
                  <InfoRow label="Gender" value={employee.gender} icon={employee.gender?.toLowerCase() === "male" ? Mars : Venus} />
                  <InfoRow label="Blood Group" value={employee.blood_group} icon={Droplets} />
                  <InfoRow label="Marital Status" value={employee.marital_status} icon={Heart} />
                  <InfoRow label="Religion" value={employee.religion} icon={Palette} />
                  <InfoRow label="Nationality" value={employee.nationality} icon={Globe} />
                  <InfoRow label="NID Number" value={employee.nid} icon={Fingerprint} />
                </CardContent>
              </Card>

              {/* Contact */}
              <Card className="overflow-hidden border-t-2 border-t-blue-500/20">
                <CardHeader className="bg-muted/30 pb-3">
                  <CardTitle className="text-sm flex items-center gap-2">
                    <Phone className="h-4 w-4 text-blue-500" /> Contact Information
                  </CardTitle>
                </CardHeader>
                <CardContent className="grid gap-4 pt-4">
                  <InfoRow label="Phone Number" value={employee.phone} icon={Phone} />
                  <InfoRow label="Email Address" value={employee.email} icon={Mail} />
                  <div className="border-t pt-3">
                    <p className="text-[11px] uppercase tracking-wider text-muted-foreground/60 font-medium mb-3 flex items-center gap-1.5">
                      <MapPin className="h-3 w-3" /> Present Address
                    </p>
                    <p className="text-sm">{employee.present_address || "—"}</p>
                  </div>
                  <div className="border-t pt-3">
                    <p className="text-[11px] uppercase tracking-wider text-muted-foreground/60 font-medium mb-3 flex items-center gap-1.5">
                      <MapPin className="h-3 w-3" /> Permanent Address
                    </p>
                    <p className="text-sm">{employee.permanent_address || "—"}</p>
                  </div>
                </CardContent>
              </Card>

              {/* Family & Emergency */}
              <Card className="overflow-hidden border-t-2 border-t-rose-500/20">
                <CardHeader className="bg-muted/30 pb-3">
                  <CardTitle className="text-sm flex items-center gap-2">
                    <Heart className="h-4 w-4 text-rose-500" /> Family & Emergency
                  </CardTitle>
                </CardHeader>
                <CardContent className="grid gap-4 pt-4">
                  <InfoRow label="Spouse Name" value={employee.spouse_name} icon={Heart} />
                  <InfoRow label="Emergency Contact" value={employee.emergency_contact} icon={Users} />
                  <InfoRow label="Emergency Phone" value={employee.emergency_phone} icon={Phone} />
                  <InfoRow label="Number of Dependents" value={employee.number_of_dependents} icon={Baby} />
                </CardContent>
              </Card>

              {/* Office Information */}
              <Card className="overflow-hidden border-t-2 border-t-amber-500/20 md:col-span-2">
                <CardHeader className="bg-muted/30 pb-3">
                  <CardTitle className="text-sm flex items-center gap-2">
                    <BriefcaseIcon className="h-4 w-4 text-amber-500" /> Office Information
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-4">
                  <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    <InfoRow label="Employee ID" value={employee.employee_id} icon={Hash} />
                    <InfoRow label="Punch Number" value={employee.punch_number} icon={Fingerprint} />
                    <InfoRow label="Employee Type" value={employee.employee_type} icon={BriefcaseIcon} />
                    <InfoRow label="Grade" value={employee.grade} icon={BadgePercent} />
                    <InfoRow label="Joining Date" value={employee.joining_date?.split("T")[0]} icon={CalendarDays} />
                    <InfoRow label="Department" value={employee.department?.name} icon={Building2} />
                    <InfoRow label="Section" value={employee.section_ref?.name} icon={Building2} />
                    <InfoRow label="Designation" value={employee.designation_ref?.name} icon={Shield} />
                    <InfoRow label="Line" value={employee.line_ref?.name} icon={ChevronRight} />
                    <InfoRow label="Group" value={employee.group_ref?.name} icon={Users} />
                    <InfoRow label="Floor" value={employee.floor_ref?.name} icon={Building2} />
                    <InfoRow label="Shift" value={employee.shift?.name} icon={Clock} />
                    <InfoRow label="Over Time" value={employee.over_time_status ? "Enabled" : "Disabled"} icon={Clock} />
                    <InfoRow label="Company" value={employee.company?.company_name_en} icon={Building2} />
                  </div>
                </CardContent>
              </Card>

              {/* Bank Account */}
              <Card className="overflow-hidden border-t-2 border-t-emerald-500/20">
                <CardHeader className="bg-muted/30 pb-3">
                  <CardTitle className="text-sm flex items-center gap-2">
                    <CreditCard className="h-4 w-4 text-emerald-500" /> Bank Account
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
            {salary ? (
              <div className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">

                {/* Net Salary Highlight */}
                <Card className="md:col-span-2 xl:col-span-3 overflow-hidden border-t-2 border-t-primary/30 bg-gradient-to-r from-primary/[0.03] to-transparent">
                  <CardContent className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-6">
                    <div>
                      <p className="text-xs uppercase tracking-wider text-muted-foreground/60 font-medium">Net Salary</p>
                      <p className="text-3xl font-bold text-primary">৳{(salary.net_salary || 0).toLocaleString()}</p>
                      <p className="text-xs text-muted-foreground mt-1">
                        {salary.month}/{salary.year} • {salary.total_days} days • Status: <Badge variant="secondary" className="text-[10px] capitalize ml-1">{salary.status}</Badge>
                      </p>
                    </div>
                    <div className="flex gap-4 text-center">
                      <div>
                        <p className="text-2xl font-bold text-green-600">{salary.present_days}</p>
                        <p className="text-[11px] text-muted-foreground">Present</p>
                      </div>
                      <div className="w-px bg-border" />
                      <div>
                        <p className="text-2xl font-bold text-red-600">{salary.absent_days}</p>
                        <p className="text-[11px] text-muted-foreground">Absent</p>
                      </div>
                      <div className="w-px bg-border" />
                      <div>
                        <p className="text-2xl font-bold text-yellow-600">{salary.late_days}</p>
                        <p className="text-[11px] text-muted-foreground">Late</p>
                      </div>
                      <div className="w-px bg-border" />
                      <div>
                        <p className="text-2xl font-bold text-blue-600">{salary.leave_days}</p>
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
                    <InfoRow label="Gross Salary" value={`৳${(salary.gross_salary || 0).toLocaleString()}`} icon={BanknoteIcon} />
                    <div className="border-t my-1" />
                    <InfoRow label="Basic Salary (50%)" value={`৳${(salary.basic_salary || 0).toLocaleString()}`} icon={BanknoteIcon} />
                    <InfoRow label="House Rent (25%)" value={`৳${(salary.house_rent || 0).toLocaleString()}`} icon={BanknoteIcon} />
                    <InfoRow label="Medical Allowance" value={`৳${(salary.medical_allowance || 0).toLocaleString()}`} icon={BanknoteIcon} />
                    <InfoRow label="Transport Allowance" value={`৳${(salary.transport_allowance || 0).toLocaleString()}`} icon={BanknoteIcon} />
                    <InfoRow label="Food Allowance" value={`৳${(salary.food_allowance || 0).toLocaleString()}`} icon={BanknoteIcon} />
                    <InfoRow label="Other Allowance" value={`৳${(salary.other_allowance || 0).toLocaleString()}`} icon={BanknoteIcon} />
                    <div className="border-t my-1" />
                    <InfoRow label="Overtime Amount" value={`৳${(salary.overtime_amount || 0).toLocaleString()}`} icon={Clock} />
                    <InfoRow label="Attendance Bonus" value={`৳${(salary.attendance_bonus || 0).toLocaleString()}`} icon={CheckCheck} />
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
                    <InfoRow label="Absent Deduction" value={`৳${(salary.absent_deduction || 0).toLocaleString()}`} icon={TrendingDown} />
                    <InfoRow label="Other Deduction" value={`৳${(salary.other_deduction || 0).toLocaleString()}`} icon={TrendingDown} />
                    <div className="border-t my-1" />
                    <div className="flex items-center justify-between rounded-lg bg-red-50 dark:bg-red-950/20 px-3 py-2.5 -mx-1">
                      <span className="text-xs font-medium text-red-700 dark:text-red-400">Total Deductions</span>
                      <span className="text-sm font-bold text-red-600 dark:text-red-400">৳{(salary.total_deductions || 0).toLocaleString()}</span>
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
                    <InfoRow label="Month" value={`${salary.month}/${salary.year}`} icon={CalendarDays} />
                    <div className="border-t my-1" />
                    <InfoRow label="Total Days" value={salary.total_days} icon={CalendarDays} />
                    <InfoRow label="Present Days" value={salary.present_days} icon={UserCheck} />
                    <InfoRow label="Absent Days" value={salary.absent_days} icon={X} />
                    <InfoRow label="Late Days" value={salary.late_days} icon={Clock} />
                    <InfoRow label="Leave Days" value={salary.leave_days} icon={CalendarDays} />
                    <InfoRow label="Weekend / Holiday" value={salary.weekend_days} icon={RefreshCw} />
                  </CardContent>
                </Card>

              </div>
            ) : (
              <Card className="border-dashed">
                <CardContent className="flex flex-col items-center gap-3 py-12">
                  <BanknoteIcon className="h-10 w-10 text-muted-foreground/30" />
                  <p className="text-sm text-muted-foreground">No salary record found for the current month</p>
                </CardContent>
              </Card>
            )}
          </TabsContent>

          {/* === Attendance Tab === */}
          <TabsContent value="attendance" className="mt-6 space-y-6">
            <div>
              <div className="flex items-center justify-between mb-4">
                <p className="text-sm text-muted-foreground">
                  Attendance breakdown for <span className="font-medium text-foreground">{new Date().toLocaleString("en-GB", { month: "long", year: "numeric" })}</span>
                  {totalAttendance > 0 && <span className="ml-2">— {totalAttendance} total records</span>}
                </p>
              </div>
              {attendance.length > 0 ? (
                <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
                  {attendance.map((a) => {
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
                <Card className="border-dashed">
                  <CardContent className="flex flex-col items-center gap-3 py-12">
                    <CalendarCheckIcon className="h-10 w-10 text-muted-foreground/30" />
                    <p className="text-sm text-muted-foreground">No attendance records for the current month</p>
                  </CardContent>
                </Card>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
