"use client"

import { z } from "zod"

export interface DailyAttendance {
  id: number
  employee: string
  employeeId: string
  company: string
  department: string
  designation: string
  line: string
  group: string
  date: string
  checkIn: string
  checkOut: string
  status: "Present" | "Late" | "Absent" | "Half Day" | "Holiday" | "Leave"
  late: string
  overTime: string
  note: string
}

export const dailyAttendanceSchema = z.object({
  employee: z.string().min(2, "Employee name is required"),
  employeeId: z.string().min(1, "Employee code is required"),
  company: z.string().min(1, "Company is required"),
  department: z.string().min(1, "Department is required"),
  designation: z.string().min(1, "Designation is required"),
  line: z.string().min(1, "Line is required"),
  group: z.string().min(1, "Group is required"),
  date: z.string().min(1, "Date is required"),
  checkIn: z.string().min(1, "Check-in time is required"),
  checkOut: z.string().min(1, "Check-out time is required"),
  status: z.enum(["Present", "Late", "Absent", "Half Day", "Holiday", "Leave"]),
  note: z.string().optional(),
})

export type DailyAttendanceFormData = z.infer<typeof dailyAttendanceSchema>

const mockAttendance: DailyAttendance[] = [
  { id: 1, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", group: "Group-A", date: "2026-07-13", checkIn: "08:15 AM", checkOut: "05:10 PM", status: "Present", late: "15m", overTime: "10m", note: "" },
  { id: 2, employee: "Shamima Akter", employeeId: "EMP002", company: "HR Hub Ltd", department: "Production", designation: "Supervisor", line: "Line-1", group: "Group-A", date: "2026-07-13", checkIn: "08:00 AM", checkOut: "04:55 PM", status: "Present", late: "-", overTime: "-", note: "" },
  { id: 3, employee: "Kamal Hossain", employeeId: "EMP003", company: "HR Hub Ltd", department: "Admin", designation: "Manager", line: "Line-2", group: "Group-B", date: "2026-07-13", checkIn: "09:30 AM", checkOut: "05:00 PM", status: "Late", late: "1h 30m", overTime: "-", note: "Traffic jam" },
  { id: 4, employee: "Nasrin Sultana", employeeId: "EMP004", company: "HR Hub Ltd", department: "IT", designation: "Developer", line: "Line-3", group: "Group-C", date: "2026-07-13", checkIn: "07:55 AM", checkOut: "04:45 PM", status: "Present", late: "-", overTime: "-", note: "" },
  { id: 5, employee: "Jahangir Alam", employeeId: "EMP005", company: "HR Hub Ltd", department: "QC", designation: "QC Inspector", line: "Line-2", group: "Group-B", date: "2026-07-13", checkIn: "-", checkOut: "-", status: "Absent", late: "-", overTime: "-", note: "Sick leave not applied" },
  { id: 6, employee: "Farida Begum", employeeId: "EMP006", company: "HR Hub Ltd", department: "Cleaning", designation: "Cleaner", line: "Line-3", group: "Group-C", date: "2026-07-13", checkIn: "08:10 AM", checkOut: "05:05 PM", status: "Present", late: "10m", overTime: "5m", note: "" },
  { id: 7, employee: "Abdur Rahman", employeeId: "EMP007", company: "HR Hub Ltd", department: "Security", designation: "Security Guard", line: "Line-1", group: "Group-A", date: "2026-07-13", checkIn: "08:00 AM", checkOut: "04:00 PM", status: "Half Day", late: "-", overTime: "-", note: "Half day leave approved" },
  { id: 8, employee: "Maksuda Khatun", employeeId: "EMP008", company: "HR Hub Ltd", department: "Finance", designation: "Accountant", line: "Line-2", group: "Group-B", date: "2026-07-13", checkIn: "08:20 AM", checkOut: "05:15 PM", status: "Present", late: "20m", overTime: "15m", note: "" },
  { id: 9, employee: "Shahidul Islam", employeeId: "EMP009", company: "HR Hub Ltd", department: "Logistics", designation: "Driver", line: "Line-3", group: "Group-C", date: "2026-07-13", checkIn: "09:00 AM", checkOut: "05:30 PM", status: "Late", late: "1h", overTime: "30m", note: "" },
  { id: 10, employee: "Rokeya Begum", employeeId: "EMP010", company: "HR Hub Ltd", department: "HR", designation: "HR Executive", line: "Line-1", group: "Group-A", date: "2026-07-13", checkIn: "07:50 AM", checkOut: "04:50 PM", status: "Present", late: "-", overTime: "-", note: "" },
  { id: 11, employee: "Delwar Hossain", employeeId: "EMP011", company: "Tech Solutions BD", department: "IT", designation: "Software Developer", line: "Line-1", group: "Group-D", date: "2026-07-13", checkIn: "08:30 AM", checkOut: "06:00 PM", status: "Late", late: "30m", overTime: "1h", note: "" },
  { id: 12, employee: "Nargis Akter", employeeId: "EMP012", company: "Tech Solutions BD", department: "HR", designation: "HR Executive", line: "Line-2", group: "Group-D", date: "2026-07-13", checkIn: "08:05 AM", checkOut: "05:00 PM", status: "Present", late: "5m", overTime: "-", note: "" },
  { id: 13, employee: "Shafiqur Rahman", employeeId: "EMP013", company: "Garments Corp", department: "Production", designation: "Operator", line: "Line-4", group: "Group-C", date: "2026-07-13", checkIn: "07:50 AM", checkOut: "05:20 PM", status: "Present", late: "-", overTime: "20m", note: "" },
  { id: 14, employee: "Tahmina Begum", employeeId: "EMP014", company: "Garments Corp", department: "QC", designation: "QC Inspector", line: "Line-5", group: "Group-B", date: "2026-07-13", checkIn: "08:00 AM", checkOut: "05:00 PM", status: "Present", late: "-", overTime: "-", note: "" },
  { id: 15, employee: "Mizanur Rahman", employeeId: "EMP015", company: "HR Hub Ltd", department: "Production", designation: "Operator", line: "Line-1", group: "Group-A", date: "2026-07-13", checkIn: "08:45 AM", checkOut: "05:15 PM", status: "Late", late: "45m", overTime: "15m", note: "" },
  { id: 16, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", group: "Group-A", date: "2026-07-14", checkIn: "08:10 AM", checkOut: "05:05 PM", status: "Present", late: "10m", overTime: "5m", note: "" },
  { id: 17, employee: "Shamima Akter", employeeId: "EMP002", company: "HR Hub Ltd", department: "Production", designation: "Supervisor", line: "Line-1", group: "Group-A", date: "2026-07-14", checkIn: "08:00 AM", checkOut: "05:00 PM", status: "Present", late: "-", overTime: "-", note: "" },
  { id: 18, employee: "Kamal Hossain", employeeId: "EMP003", company: "HR Hub Ltd", department: "Admin", designation: "Manager", line: "Line-2", group: "Group-B", date: "2026-07-14", checkIn: "-", checkOut: "-", status: "Leave", late: "-", overTime: "-", note: "Annual leave" },
  { id: 19, employee: "Nasrin Sultana", employeeId: "EMP004", company: "HR Hub Ltd", department: "IT", designation: "Developer", line: "Line-3", group: "Group-C", date: "2026-07-14", checkIn: "07:50 AM", checkOut: "04:30 PM", status: "Present", late: "-", overTime: "-", note: "" },
  { id: 20, employee: "Jahangir Alam", employeeId: "EMP005", company: "HR Hub Ltd", department: "QC", designation: "QC Inspector", line: "Line-2", group: "Group-B", date: "2026-07-14", checkIn: "08:00 AM", checkOut: "05:00 PM", status: "Present", late: "-", overTime: "-", note: "" },
  { id: 21, employee: "Abdur Rahman", employeeId: "EMP007", company: "HR Hub Ltd", department: "Security", designation: "Security Guard", line: "Line-1", group: "Group-A", date: "2026-07-14", checkIn: "07:45 PM", checkOut: "06:00 AM", status: "Present", late: "-", overTime: "1h", note: "Night shift" },
  { id: 22, employee: "Delwar Hossain", employeeId: "EMP011", company: "Tech Solutions BD", department: "IT", designation: "Software Developer", line: "Line-1", group: "Group-D", date: "2026-07-14", checkIn: "09:00 AM", checkOut: "06:30 PM", status: "Late", late: "1h", overTime: "1h 30m", note: "" },
  { id: 23, employee: "Nargis Akter", employeeId: "EMP012", company: "Tech Solutions BD", department: "HR", designation: "HR Executive", line: "Line-2", group: "Group-D", date: "2026-07-14", checkIn: "-", checkOut: "-", status: "Absent", late: "-", overTime: "-", note: "" },
  { id: 24, employee: "Shafiqur Rahman", employeeId: "EMP013", company: "Garments Corp", department: "Production", designation: "Operator", line: "Line-4", group: "Group-C", date: "2026-07-14", checkIn: "07:55 AM", checkOut: "05:10 PM", status: "Present", late: "-", overTime: "10m", note: "" },
  { id: 25, employee: "Tahmina Begum", employeeId: "EMP014", company: "Garments Corp", department: "QC", designation: "QC Inspector", line: "Line-5", group: "Group-B", date: "2026-07-14", checkIn: "08:30 AM", checkOut: "05:00 PM", status: "Late", late: "30m", overTime: "-", note: "" },
  { id: 26, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", group: "Group-A", date: "2026-07-15", checkIn: "08:00 AM", checkOut: "05:00 PM", status: "Holiday", late: "-", overTime: "-", note: "National holiday" },
  { id: 27, employee: "Shamima Akter", employeeId: "EMP002", company: "HR Hub Ltd", department: "Production", designation: "Supervisor", line: "Line-1", group: "Group-A", date: "2026-07-15", checkIn: "08:00 AM", checkOut: "04:00 PM", status: "Half Day", late: "-", overTime: "-", note: "Personal work" },
  { id: 28, employee: "Mizanur Rahman", employeeId: "EMP015", company: "HR Hub Ltd", department: "Production", designation: "Operator", line: "Line-2", group: "Group-A", date: "2026-07-15", checkIn: "08:00 AM", checkOut: "06:00 PM", status: "Present", late: "-", overTime: "1h", note: "Overtime approved" },
  { id: 29, employee: "Shahidul Islam", employeeId: "EMP009", company: "HR Hub Ltd", department: "Logistics", designation: "Driver", line: "Line-3", group: "Group-C", date: "2026-07-15", checkIn: "07:30 AM", checkOut: "07:00 PM", status: "Present", late: "-", overTime: "2h", note: "Extra delivery duty" },
  { id: 30, employee: "Delwar Hossain", employeeId: "EMP011", company: "Tech Solutions BD", department: "IT", designation: "Software Developer", line: "Line-1", group: "Group-D", date: "2026-07-15", checkIn: "10:00 AM", checkOut: "07:00 PM", status: "Late", late: "2h", overTime: "2h", note: "Came late, stayed late" },
]

let records = [...mockAttendance]
let nextId = 31

export function getDailyAttendance(): DailyAttendance[] { return records }
export function getDailyAttendanceRecord(id: number): DailyAttendance | undefined { return records.find((r) => r.id === id) }

export function createDailyAttendance(data: DailyAttendanceFormData): DailyAttendance {
  const r: DailyAttendance = { ...data, note: data.note || "", late: computeLate(data.checkIn), overTime: computeOverTime(data.checkOut), id: nextId++ }
  records.push(r)
  return r
}

export function updateDailyAttendance(id: number, data: Partial<DailyAttendanceFormData>): DailyAttendance | null {
  const index = records.findIndex((r) => r.id === id)
  if (index === -1) return null
  const existing = records[index]
  const checkIn = data.checkIn ?? existing.checkIn
  const checkOut = data.checkOut ?? existing.checkOut
  records[index] = { ...existing, ...data, note: data.note ?? existing.note, late: computeLate(checkIn), overTime: computeOverTime(checkOut) }
  return records[index]
}

export function deleteDailyAttendance(id: number): boolean {
  const index = records.findIndex((r) => r.id === id)
  if (index === -1) return false
  records.splice(index, 1)
  return true
}

function parseTime(time: string): number | null {
  const match = time.match(/(\d+):(\d+)\s*(AM|PM)/i)
  if (!match) return null
  let hours = Number(match[1])
  const minutes = Number(match[2])
  const ampm = match[3].toUpperCase()
  if (ampm === "PM" && hours !== 12) hours += 12
  if (ampm === "AM" && hours === 12) hours = 0
  return hours * 60 + minutes
}

function computeLate(checkIn: string): string {
  if (!checkIn || checkIn === "-") return "-"
  const totalMinutes = parseTime(checkIn)
  if (totalMinutes === null) return "-"
  const standardMinutes = 8 * 60
  if (totalMinutes <= standardMinutes) return "-"
  const diff = totalMinutes - standardMinutes
  const h = Math.floor(diff / 60)
  const m = diff % 60
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}

function computeOverTime(checkOut: string): string {
  if (!checkOut || checkOut === "-") return "-"
  const totalMinutes = parseTime(checkOut)
  if (totalMinutes === null) return "-"
  const standardMinutes = 17 * 60 // 5:00 PM
  if (totalMinutes <= standardMinutes) return "-"
  const diff = totalMinutes - standardMinutes
  const h = Math.floor(diff / 60)
  const m = diff % 60
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}

export const attendanceStatusOptions = [
  { value: "Present" as const, label: "Present" },
  { value: "Late" as const, label: "Late" },
  { value: "Absent" as const, label: "Absent" },
  { value: "Half Day" as const, label: "Half Day" },
  { value: "Holiday" as const, label: "Holiday" },
  { value: "Leave" as const, label: "Leave" },
]

export const companyOptions = ["HR Hub Ltd", "Tech Solutions BD", "Garments Corp"].map((v) => ({ value: v, label: v }))
export const departmentOptions = [
  "Production", "Admin", "IT", "QC", "Security", "Cleaning", "Finance", "Logistics", "HR",
].map((v) => ({ value: v, label: v }))
export const designationOptions = [
  "Senior Operator", "Operator", "Supervisor", "Manager", "Developer", "QC Inspector",
  "Security Guard", "Cleaner", "Accountant", "Driver", "HR Executive",
].map((v) => ({ value: v, label: v }))
export const lineOptions = ["Line-1", "Line-2", "Line-3", "Line-4", "Line-5"].map((v) => ({ value: v, label: v }))
export const groupOptions = ["Group-A", "Group-B", "Group-C", "Group-D"].map((v) => ({ value: v, label: v }))
