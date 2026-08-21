"use client"

export interface JobCardRecord {
  id: number
  employee: string
  employeeId: string
  company: string
  department: string
  designation: string
  line: string
  section: string
  group: string
  date: string
  shift: string
  inTime: string
  outTime: string
  late: string
  overTime: string
  status: "Present" | "Late" | "Absent" | "Half Day" | "Holiday" | "Leave"
}

export interface EmployeeInfo {
  employee: string
  employeeId: string
  company: string
  department: string
  designation: string
  line: string
  section: string
  group: string
}

const mockRecords: JobCardRecord[] = [
  { id: 1, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-01", shift: "Morning", inTime: "08:15 AM", outTime: "05:10 PM", late: "15m", overTime: "10m", status: "Present" },
  { id: 2, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-02", shift: "Morning", inTime: "08:00 AM", outTime: "05:00 PM", late: "-", overTime: "-", status: "Present" },
  { id: 3, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-03", shift: "Morning", inTime: "09:30 AM", outTime: "05:00 PM", late: "1h 30m", overTime: "-", status: "Late" },
  { id: 4, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-04", shift: "Morning", inTime: "07:55 AM", outTime: "04:45 PM", late: "-", overTime: "-", status: "Present" },
  { id: 5, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-05", shift: "Morning", inTime: "-", outTime: "-", late: "-", overTime: "-", status: "Absent" },
  { id: 6, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-06", shift: "Morning", inTime: "08:10 AM", outTime: "05:05 PM", late: "10m", overTime: "5m", status: "Present" },
  { id: 7, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-07", shift: "Morning", inTime: "08:00 AM", outTime: "04:00 PM", late: "-", overTime: "-", status: "Half Day" },
  { id: 8, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-08", shift: "Morning", inTime: "08:20 AM", outTime: "05:15 PM", late: "20m", overTime: "15m", status: "Present" },
  { id: 9, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-09", shift: "Evening", inTime: "02:00 PM", outTime: "11:00 PM", late: "-", overTime: "-", status: "Present" },
  { id: 10, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-10", shift: "Evening", inTime: "02:30 PM", outTime: "11:15 PM", late: "30m", overTime: "15m", status: "Late" },
  { id: 11, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-11", shift: "Morning", inTime: "08:00 AM", outTime: "05:00 PM", late: "-", overTime: "-", status: "Present" },
  { id: 12, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-12", shift: "Morning", inTime: "08:00 AM", outTime: "05:00 PM", late: "-", overTime: "-", status: "Holiday" },
  { id: 13, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-13", shift: "Morning", inTime: "08:15 AM", outTime: "05:10 PM", late: "15m", overTime: "10m", status: "Present" },
  { id: 14, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-14", shift: "Morning", inTime: "08:10 AM", outTime: "05:05 PM", late: "10m", overTime: "5m", status: "Present" },
  { id: 15, employee: "Rafiqul Islam", employeeId: "EMP001", company: "HR Hub Ltd", department: "Production", designation: "Senior Operator", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-15", shift: "Morning", inTime: "08:00 AM", outTime: "05:00 PM", late: "-", overTime: "-", status: "Present" },
  { id: 16, employee: "Shamima Akter", employeeId: "EMP002", company: "HR Hub Ltd", department: "Production", designation: "Supervisor", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-01", shift: "Morning", inTime: "08:00 AM", outTime: "04:55 PM", late: "-", overTime: "-", status: "Present" },
  { id: 17, employee: "Shamima Akter", employeeId: "EMP002", company: "HR Hub Ltd", department: "Production", designation: "Supervisor", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-02", shift: "Morning", inTime: "08:05 AM", outTime: "05:00 PM", late: "5m", overTime: "-", status: "Present" },
  { id: 18, employee: "Shamima Akter", employeeId: "EMP002", company: "HR Hub Ltd", department: "Production", designation: "Supervisor", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-03", shift: "Morning", inTime: "-", outTime: "-", late: "-", overTime: "-", status: "Leave" },
  { id: 19, employee: "Shamima Akter", employeeId: "EMP002", company: "HR Hub Ltd", department: "Production", designation: "Supervisor", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-04", shift: "Morning", inTime: "07:50 AM", outTime: "04:45 PM", late: "-", overTime: "-", status: "Present" },
  { id: 20, employee: "Shamima Akter", employeeId: "EMP002", company: "HR Hub Ltd", department: "Production", designation: "Supervisor", line: "Line-1", section: "Section-A", group: "Group-A", date: "2026-07-05", shift: "Morning", inTime: "08:00 AM", outTime: "04:30 PM", late: "-", overTime: "-", status: "Present" },
  { id: 21, employee: "Kamal Hossain", employeeId: "EMP003", company: "HR Hub Ltd", department: "Admin", designation: "Manager", line: "Line-2", section: "Section-B", group: "Group-B", date: "2026-07-01", shift: "Day", inTime: "09:00 AM", outTime: "05:00 PM", late: "1h", overTime: "-", status: "Late" },
  { id: 22, employee: "Kamal Hossain", employeeId: "EMP003", company: "HR Hub Ltd", department: "Admin", designation: "Manager", line: "Line-2", section: "Section-B", group: "Group-B", date: "2026-07-02", shift: "Day", inTime: "08:30 AM", outTime: "05:00 PM", late: "30m", overTime: "-", status: "Late" },
  { id: 23, employee: "Kamal Hossain", employeeId: "EMP003", company: "HR Hub Ltd", department: "Admin", designation: "Manager", line: "Line-2", section: "Section-B", group: "Group-B", date: "2026-07-03", shift: "Day", inTime: "09:30 AM", outTime: "05:00 PM", late: "1h 30m", overTime: "-", status: "Late" },
  { id: 24, employee: "Kamal Hossain", employeeId: "EMP003", company: "HR Hub Ltd", department: "Admin", designation: "Manager", line: "Line-2", section: "Section-B", group: "Group-B", date: "2026-07-04", shift: "Day", inTime: "10:00 AM", outTime: "06:00 PM", late: "2h", overTime: "1h", status: "Late" },
  { id: 25, employee: "Kamal Hossain", employeeId: "EMP003", company: "HR Hub Ltd", department: "Admin", designation: "Manager", line: "Line-2", section: "Section-B", group: "Group-B", date: "2026-07-05", shift: "Day", inTime: "-", outTime: "-", late: "-", overTime: "-", status: "Absent" },
  { id: 26, employee: "Nasrin Sultana", employeeId: "EMP004", company: "Tech Solutions BD", department: "IT", designation: "Software Developer", line: "Line-1", section: "Section-C", group: "Group-D", date: "2026-07-01", shift: "Flexible", inTime: "07:55 AM", outTime: "04:45 PM", late: "-", overTime: "-", status: "Present" },
  { id: 27, employee: "Nasrin Sultana", employeeId: "EMP004", company: "Tech Solutions BD", department: "IT", designation: "Software Developer", line: "Line-1", section: "Section-C", group: "Group-D", date: "2026-07-02", shift: "Flexible", inTime: "08:00 AM", outTime: "05:00 PM", late: "-", overTime: "-", status: "Present" },
  { id: 28, employee: "Nasrin Sultana", employeeId: "EMP004", company: "Tech Solutions BD", department: "IT", designation: "Software Developer", line: "Line-1", section: "Section-C", group: "Group-D", date: "2026-07-03", shift: "Flexible", inTime: "08:15 AM", outTime: "05:00 PM", late: "15m", overTime: "-", status: "Present" },
  { id: 29, employee: "Nasrin Sultana", employeeId: "EMP004", company: "Tech Solutions BD", department: "IT", designation: "Software Developer", line: "Line-1", section: "Section-C", group: "Group-D", date: "2026-07-04", shift: "Flexible", inTime: "-", outTime: "-", late: "-", overTime: "-", status: "Leave" },
  { id: 30, employee: "Nasrin Sultana", employeeId: "EMP004", company: "Tech Solutions BD", department: "IT", designation: "Software Developer", line: "Line-1", section: "Section-C", group: "Group-D", date: "2026-07-05", shift: "Flexible", inTime: "08:00 AM", outTime: "04:30 PM", late: "-", overTime: "-", status: "Present" },
  { id: 31, employee: "Jahangir Alam", employeeId: "EMP005", company: "Garments Corp", department: "QC", designation: "QC Inspector", line: "Line-5", section: "Section-D", group: "Group-C", date: "2026-07-01", shift: "Day", inTime: "08:00 AM", outTime: "05:00 PM", late: "-", overTime: "-", status: "Present" },
  { id: 32, employee: "Jahangir Alam", employeeId: "EMP005", company: "Garments Corp", department: "QC", designation: "QC Inspector", line: "Line-5", section: "Section-D", group: "Group-C", date: "2026-07-02", shift: "Day", inTime: "08:15 AM", outTime: "05:10 PM", late: "15m", overTime: "10m", status: "Present" },
  { id: 33, employee: "Jahangir Alam", employeeId: "EMP005", company: "Garments Corp", department: "QC", designation: "QC Inspector", line: "Line-5", section: "Section-D", group: "Group-C", date: "2026-07-03", shift: "Day", inTime: "08:00 AM", outTime: "05:00 PM", late: "-", overTime: "-", status: "Present" },
  { id: 34, employee: "Jahangir Alam", employeeId: "EMP005", company: "Garments Corp", department: "QC", designation: "QC Inspector", line: "Line-5", section: "Section-D", group: "Group-C", date: "2026-07-04", shift: "Day", inTime: "-", outTime: "-", late: "-", overTime: "-", status: "Absent" },
  { id: 35, employee: "Jahangir Alam", employeeId: "EMP005", company: "Garments Corp", department: "QC", designation: "QC Inspector", line: "Line-5", section: "Section-D", group: "Group-C", date: "2026-07-05", shift: "Day", inTime: "08:00 AM", outTime: "04:00 PM", late: "-", overTime: "-", status: "Half Day" },
  { id: 36, employee: "Maksuda Khatun", employeeId: "EMP008", company: "HR Hub Ltd", department: "Finance", designation: "Accountant", line: "Line-3", section: "Section-E", group: "Group-B", date: "2026-07-01", shift: "Morning", inTime: "08:20 AM", outTime: "05:15 PM", late: "20m", overTime: "15m", status: "Present" },
  { id: 37, employee: "Maksuda Khatun", employeeId: "EMP008", company: "HR Hub Ltd", department: "Finance", designation: "Accountant", line: "Line-3", section: "Section-E", group: "Group-B", date: "2026-07-02", shift: "Morning", inTime: "08:00 AM", outTime: "05:00 PM", late: "-", overTime: "-", status: "Present" },
  { id: 38, employee: "Maksuda Khatun", employeeId: "EMP008", company: "HR Hub Ltd", department: "Finance", designation: "Accountant", line: "Line-3", section: "Section-E", group: "Group-B", date: "2026-07-03", shift: "Morning", inTime: "08:10 AM", outTime: "05:05 PM", late: "10m", overTime: "5m", status: "Present" },
  { id: 39, employee: "Maksuda Khatun", employeeId: "EMP008", company: "HR Hub Ltd", department: "Finance", designation: "Accountant", line: "Line-3", section: "Section-E", group: "Group-B", date: "2026-07-04", shift: "Morning", inTime: "08:00 AM", outTime: "05:00 PM", late: "-", overTime: "-", status: "Present" },
  { id: 40, employee: "Maksuda Khatun", employeeId: "EMP008", company: "HR Hub Ltd", department: "Finance", designation: "Accountant", line: "Line-3", section: "Section-E", group: "Group-B", date: "2026-07-05", shift: "Morning", inTime: "-", outTime: "-", late: "-", overTime: "-", status: "Leave" },
]

export function getJobCardRecords(): JobCardRecord[] { return records }
let records = [...mockRecords]

export function getEmployees(): EmployeeInfo[] {
  const map = new Map<string, EmployeeInfo>()
  records.forEach((r) => {
    if (!map.has(r.employeeId))
      map.set(r.employeeId, { employee: r.employee, employeeId: r.employeeId, company: r.company, department: r.department, designation: r.designation, line: r.line, section: r.section, group: r.group })
  })
  return Array.from(map.values())
}

export const companyOptions = ["HR Hub Ltd", "Tech Solutions BD", "Garments Corp"].map((v) => ({ value: v, label: v }))
export const departmentOptions = ["Production", "Admin", "IT", "QC", "Finance"].map((v) => ({ value: v, label: v }))
export const designationOptions = ["Senior Operator", "Operator", "Supervisor", "Manager", "Software Developer", "QC Inspector", "Accountant"].map((v) => ({ value: v, label: v }))
export const lineOptions = ["Line-1", "Line-2", "Line-3", "Line-5"].map((v) => ({ value: v, label: v }))
export const sectionOptions = ["Section-A", "Section-B", "Section-C", "Section-D", "Section-E"].map((v) => ({ value: v, label: v }))
export const groupOptions = ["Group-A", "Group-B", "Group-C", "Group-D"].map((v) => ({ value: v, label: v }))
