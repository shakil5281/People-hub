import type { LucideIcon } from "lucide-react"
import {
  Building2Icon,
  CalendarClockIcon,
  TimerIcon,
  MapPinIcon,
  UsersIcon,
  LayersIcon,
  BuildingIcon,
ClipboardCheckIcon,
  ClipboardEditIcon,
  FileTextIcon,
  IdCardIcon,
  ClockIcon,
  ChartColumnIcon,
  ClipboardListIcon,
  UserXIcon,
  MoonIcon,
  CalendarDaysIcon,
  TagIcon,
  CalendarCheckIcon,
  DollarSignIcon,
  ArrowUpDownIcon,
  FileSpreadsheetIcon,
  FileBarChartIcon,
  TrendingUpIcon,
  ReceiptIcon,
  CalendarRangeIcon,
  DatabaseIcon,
  RefreshCwIcon,
  BarChart3Icon,
  ShieldIcon,
  UserCogIcon,
  ShieldCheckIcon,
  Settings2Icon,
  Trash2Icon,
  AlertTriangleIcon,
  CalendarIcon,
  CoffeeIcon,
  GiftIcon,
} from "lucide-react"

export const monthlyReportNav = [
  { title: "Monthly Leave Report", url: "/leave/monthly-leave-report", icon: ChartColumnIcon },
  { title: "Monthly Attendance Report", url: "/attendance/monthly-attendance", icon: BarChart3Icon },
] as const satisfies { title: string; url: string; icon: LucideIcon }[]

export const informationNav = [
  { title: "Company", url: "/information/company", icon: Building2Icon },
  { title: "Shift", url: "/information/shift", icon: CalendarClockIcon },
  { title: "Temporary Shift", url: "/information/temporary-shift", icon: TimerIcon },
  { title: "Address", url: "/information/address", icon: MapPinIcon },
  { title: "Group", url: "/information/group", icon: UsersIcon },
  { title: "Floor", url: "/information/floor", icon: LayersIcon },
  { title: "Organization", url: "/information/organization", icon: BuildingIcon },
] as const satisfies { title: string; url: string; icon: LucideIcon }[]

export const hrNav = [
  { title: "Employees", url: "/hr/employees", icon: UsersIcon },
  { title: "Requirements", url: "/hr/requirements", icon: ClipboardListIcon },
  { title: "Seperation", url: "/hr/seperation", icon: UserXIcon },
  { title: "Id Card", url: "/hr/id-card", icon: IdCardIcon },
  { title: "Punishment", url: "/hr/punishment", icon: AlertTriangleIcon },
  { title: "Daily Schedule", url: "/hr/daily-schedule", icon: CalendarIcon },
  { title: "mCash", url: "/hr/mcash", icon: DollarSignIcon },
] as const satisfies { title: string; url: string; icon: LucideIcon }[]

export const attendanceNav = [
  { title: "Daily Attendance", url: "/attendance/daily-attendance", icon: ClipboardCheckIcon },
  { title: "Daily Summary", url: "/attendance/daily-summary", icon: FileTextIcon },
  { title: "Job Card", url: "/attendance/job-card", icon: IdCardIcon },
  { title: "Job Age", url: "/attendance/job-age", icon: ClockIcon },
  { title: "Over Time Sheet", url: "/attendance/over-time-sheet", icon: ClockIcon },
  { title: "Over Time Summary", url: "/attendance/over-time-summary", icon: ChartColumnIcon },
  { title: "OT Early Exit", url: "/attendance/ot-early-exit", icon: ClockIcon },
  { title: "Manual Attendance", url: "/attendance/manual-attendance", icon: ClipboardCheckIcon },
  { title: "Missing Attendance", url: "/attendance/missing-attendance", icon: UserXIcon },
  { title: "Custom Attendance", url: "/attendance/custom-attendance", icon: ClipboardEditIcon },
  { title: "Absent Status", url: "/attendance/absent-status", icon: UserXIcon },
  { title: "Night Bill", url: "/attendance/night-bill", icon: MoonIcon },
  { title: "Remove Attendance", url: "/attendance/remove-attendance", icon: Trash2Icon },
] as const satisfies { title: string; url: string; icon: LucideIcon }[]



export const leaveNav = [
  { title: "Leave Type", url: "/leave/leave-type", icon: TagIcon },
  { title: "Leave", url: "/leave/leave", icon: CalendarCheckIcon },
  { title: "Leave Details", url: "/leave/leave-details", icon: FileTextIcon },
  { title: "Holiday", url: "/leave/holiday", icon: CalendarDaysIcon },
] as const satisfies { title: string; url: string; icon: LucideIcon }[]

export const payrollNav = [
  { title: "Salary Process", url: "/payroll/salary-process", icon: ArrowUpDownIcon },
  { title: "Salary Sheet", url: "/payroll/salary-sheet", icon: FileSpreadsheetIcon },
  { title: "Salary Summary", url: "/payroll/salary-summary", icon: FileBarChartIcon },
  { title: "Increment", url: "/payroll/increment", icon: TrendingUpIcon },
  { title: "PaySlip", url: "/payroll/payslip", icon: ReceiptIcon },
  { title: "Daily Salary Sheet", url: "/payroll/daily-salary-sheet", icon: CalendarRangeIcon },
  { title: "Daily Salary Summary", url: "/payroll/daily-salary-summary", icon: FileBarChartIcon },
  { title: "Night Bill", url: "/payroll/night-bill", icon: MoonIcon },
  { title: "Tiffin Bill", url: "/payroll/tiffin-bill", icon: CoffeeIcon },
  { title: "Eid Bonus", url: "/payroll/eid-bonus", icon: GiftIcon },
] as const satisfies { title: string; url: string; icon: LucideIcon }[]

export const collectDataNav = [
  { title: "Log Collect", url: "/collect-data/log-collect", icon: DatabaseIcon },
  { title: "Daily Process", url: "/collect-data/daily-process", icon: RefreshCwIcon },
  { title: "Monthly Process", url: "/collect-data/monthly-process", icon: CalendarRangeIcon },
] as const satisfies { title: string; url: string; icon: LucideIcon }[]

export const administratorNav = [
  { title: "User Management", url: "/admin/users", icon: UserCogIcon },
  { title: "Roles & Permissions", url: "/admin/roles", icon: ShieldCheckIcon },
  { title: "Database", url: "/admin/database", icon: DatabaseIcon },
  { title: "System Settings", url: "/admin/settings", icon: Settings2Icon },
] as const satisfies { title: string; url: string; icon: LucideIcon }[]

export const navGroup = [
  {
    title: "Information",
    icon: Building2Icon,
    url: "#",
    items: informationNav.map(({ title, url }) => ({ title, url })),
  },
  {
    title: "Human Resource",
    icon: UsersIcon,
    url: "#",
    items: hrNav.map(({ title, url }) => ({ title, url })),
  },
  {
    title: "Attendance",
    icon: ClipboardCheckIcon,
    url: "#",
    items: attendanceNav.map(({ title, url }) => ({ title, url })),
  },
  {
    title: "Leave",
    icon: CalendarDaysIcon,
    url: "#",
    items: leaveNav.map(({ title, url }) => ({ title, url })),
  },
  {
    title: "Payroll",
    icon: DollarSignIcon,
    url: "#",
    items: payrollNav.map(({ title, url }) => ({ title, url })),
  },
  {
    title: "Collect Data",
    icon: DatabaseIcon,
    url: "#",
    items: collectDataNav.map(({ title, url }) => ({ title, url })),
  },
  {
    title: "Monthly Report",
    icon: ChartColumnIcon,
    url: "#",
    items: monthlyReportNav.map(({ title, url }) => ({ title, url })),
  },
  {
    title: "Administrator",
    icon: ShieldIcon,
    url: "#",
    items: administratorNav.map(({ title, url }) => ({ title, url })),
  },
] as const satisfies {
  title: string
  url: string
  icon: LucideIcon
  items: { title: string; url: string }[]
}[]
