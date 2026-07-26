"use client"

import * as React from "react"
import { useRouter, useParams } from "next/navigation"
import {
  ArrowLeftIcon, Loader2, UserIcon, BanknoteIcon, CalendarCheckIcon,
  UserCircle, BriefcaseIcon, CalendarDays, Phone, Mail, MapPin,
  Heart, CreditCard, BadgeCheck, Clock
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { employeeApi } from "@/lib/api"
import { cn } from "@/lib/utils"

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
  present: "text-green-600 bg-green-50 border-green-200",
  absent: "text-red-600 bg-red-50 border-red-200",
  late: "text-yellow-600 bg-yellow-50 border-yellow-200",
  on_leave: "text-blue-600 bg-blue-50 border-blue-200",
  weekend: "text-purple-600 bg-purple-50 border-purple-200",
  half_day: "text-orange-600 bg-orange-50 border-orange-200",
}

const statusLabels: Record<string, string> = {
  present: "Present",
  absent: "Absent",
  late: "Late",
  on_leave: "On Leave",
  weekend: "Weekend",
  half_day: "Half Day",
}

function InfoRow({ label, value }: { label: string; value: string | number | null | undefined }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="text-sm font-medium">{value ?? "-"}</span>
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

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error || !profile) {
    return (
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          <Button variant="ghost" onClick={() => router.back()} className="mb-2">
            <ArrowLeftIcon className="mr-2 h-4 w-4" /> Back
          </Button>
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">
            {error || "Employee not found"}
          </div>
        </div>
      </div>
    )
  }

  const { employee, attendance, salary } = profile

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" onClick={() => router.back()}>
            <ArrowLeftIcon className="h-5 w-5" />
          </Button>
          <UserIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{employee.name_en}</h1>
            <p className="text-muted-foreground mt-1">
              {employee.employee_id} — {employee.designation_ref?.name || employee.designation_id || "N/A"}
              {employee.department && ` — ${employee.department.name}`}
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <Badge variant={employee.status === "active" ? "default" : "secondary"} className="capitalize">
            {employee.status}
          </Badge>
          <Button variant="outline" size="sm" onClick={() => router.push(`/hr/employees/${id}/edit`)}>
            Edit
          </Button>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="px-4 lg:px-6">
        <TabsList>
          <TabsTrigger value="info" className="flex items-center gap-2">
            <UserCircle className="h-4 w-4" /> Employee Info
          </TabsTrigger>
          <TabsTrigger value="salary" className="flex items-center gap-2">
            <BanknoteIcon className="h-4 w-4" /> Salary
          </TabsTrigger>
          <TabsTrigger value="attendance" className="flex items-center gap-2">
            <CalendarCheckIcon className="h-4 w-4" /> Attendance
          </TabsTrigger>
        </TabsList>

        <TabsContent value="info" className="mt-4 space-y-6">
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <UserCircle className="h-4 w-4 text-muted-foreground" /> Personal Info
                </CardTitle>
              </CardHeader>
              <CardContent className="grid gap-3">
                <InfoRow label="Name (EN)" value={employee.name_en} />
                <InfoRow label="Name (BN)" value={employee.name_bn} />
                <InfoRow label="Father's Name" value={employee.father_name} />
                <InfoRow label="Mother's Name" value={employee.mother_name} />
                <InfoRow label="Date of Birth" value={employee.date_of_birth?.split("T")[0]} />
                <InfoRow label="Gender" value={employee.gender} />
                <InfoRow label="Blood Group" value={employee.blood_group} />
                <InfoRow label="Marital Status" value={employee.marital_status} />
                <InfoRow label="Religion" value={employee.religion} />
                <InfoRow label="Nationality" value={employee.nationality} />
                <InfoRow label="NID" value={employee.nid} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <Heart className="h-4 w-4 text-muted-foreground" /> Family & Emergency
                </CardTitle>
              </CardHeader>
              <CardContent className="grid gap-3">
                <InfoRow label="Spouse Name" value={employee.spouse_name} />
                <InfoRow label="Emergency Contact" value={employee.emergency_contact} />
                <InfoRow label="Emergency Phone" value={employee.emergency_phone} />
                <InfoRow label="Dependents" value={employee.number_of_dependents} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <Phone className="h-4 w-4 text-muted-foreground" /> Contact
                </CardTitle>
              </CardHeader>
              <CardContent className="grid gap-3">
                <InfoRow label="Phone" value={employee.phone} />
                <InfoRow label="Email" value={employee.email} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <MapPin className="h-4 w-4 text-muted-foreground" /> Present Address
                </CardTitle>
              </CardHeader>
              <CardContent>
                <InfoRow label="Address" value={employee.present_address} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <MapPin className="h-4 w-4 text-muted-foreground" /> Permanent Address
                </CardTitle>
              </CardHeader>
              <CardContent>
                <InfoRow label="Address" value={employee.permanent_address} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <BriefcaseIcon className="h-4 w-4 text-muted-foreground" /> Office Info
                </CardTitle>
              </CardHeader>
              <CardContent className="grid gap-3">
                <InfoRow label="Employee ID" value={employee.employee_id} />
                <InfoRow label="Punch Number" value={employee.punch_number} />
                <InfoRow label="Employee Type" value={employee.employee_type} />
                <InfoRow label="Grade" value={employee.grade} />
                <InfoRow label="Joining Date" value={employee.joining_date?.split("T")[0]} />
                <InfoRow label="Department" value={employee.department?.name} />
                <InfoRow label="Section" value={employee.section_ref?.name} />
                <InfoRow label="Designation" value={employee.designation_ref?.name} />
                <InfoRow label="Line" value={employee.line_ref?.name} />
                <InfoRow label="Group" value={employee.group_ref?.name} />
                <InfoRow label="Floor" value={employee.floor_ref?.name} />
                <InfoRow label="Shift" value={employee.shift?.name} />
                <InfoRow label="Over Time Status" value={employee.over_time_status ? "Enabled" : "Disabled"} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <CreditCard className="h-4 w-4 text-muted-foreground" /> Bank Account
                </CardTitle>
              </CardHeader>
              <CardContent className="grid gap-3">
                <InfoRow label="Account Type" value={employee.account_type} />
                <InfoRow label="Account Number" value={employee.account_number} />
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="salary" className="mt-4 space-y-6">
          {salary ? (
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle className="text-base flex items-center gap-2">
                    <BanknoteIcon className="h-4 w-4 text-muted-foreground" /> Earnings
                  </CardTitle>
                </CardHeader>
                <CardContent className="grid gap-3">
                  <InfoRow label="Gross Salary" value={salary.gross_salary?.toLocaleString()} />
                  <InfoRow label="Basic Salary" value={salary.basic_salary?.toLocaleString()} />
                  <InfoRow label="House Rent" value={salary.house_rent?.toLocaleString()} />
                  <InfoRow label="Medical Allowance" value={salary.medical_allowance?.toLocaleString()} />
                  <InfoRow label="Transport Allowance" value={salary.transport_allowance?.toLocaleString()} />
                  <InfoRow label="Food Allowance" value={salary.food_allowance?.toLocaleString()} />
                  <InfoRow label="Other Allowance" value={salary.other_allowance?.toLocaleString()} />
                  <InfoRow label="OT Amount" value={salary.overtime_amount?.toLocaleString()} />
                  <InfoRow label="Attendance Bonus" value={salary.attendance_bonus?.toLocaleString()} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base flex items-center gap-2">
                    <BadgeCheck className="h-4 w-4 text-muted-foreground" /> Deductions & Net
                  </CardTitle>
                </CardHeader>
                <CardContent className="grid gap-3">
                  <InfoRow label="Absent Deduction" value={salary.absent_deduction?.toLocaleString()} />
                  <InfoRow label="Other Deduction" value={salary.other_deduction?.toLocaleString()} />
                  <InfoRow label="Total Deductions" value={salary.total_deductions?.toLocaleString()} />
                  <div className="border-t pt-2 mt-2">
                    <InfoRow label="Net Salary" value={salary.net_salary?.toLocaleString()} />
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" /> Attendance Summary
                  </CardTitle>
                </CardHeader>
                <CardContent className="grid gap-3">
                  <InfoRow label="Month" value={`${salary.month}/${salary.year}`} />
                  <InfoRow label="Total Days" value={salary.total_days} />
                  <InfoRow label="Present Days" value={salary.present_days} />
                  <InfoRow label="Absent Days" value={salary.absent_days} />
                  <InfoRow label="Late Days" value={salary.late_days} />
                  <InfoRow label="Leave Days" value={salary.leave_days} />
                  <InfoRow label="Weekend Days" value={salary.weekend_days} />
                </CardContent>
              </Card>
            </div>
          ) : (
            <Card>
              <CardContent className="py-8 text-center text-muted-foreground">
                No salary record found for the current month
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="attendance" className="mt-4 space-y-6">
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
            {attendance.length > 0 ? attendance.map((a) => (
              <Card key={a.status} className={cn("border-2", statusColors[a.status] || "")}>
                <CardContent className="flex flex-col items-center justify-center py-6 gap-1">
                  <span className="text-3xl font-bold">{a.count}</span>
                  <span className="text-sm font-medium capitalize">
                    {statusLabels[a.status] || a.status}
                  </span>
                </CardContent>
              </Card>
            )) : (
              <Card className="md:col-span-2 lg:col-span-3 xl:col-span-6">
                <CardContent className="py-8 text-center text-muted-foreground">
                  No attendance records for the current month
                </CardContent>
              </Card>
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
