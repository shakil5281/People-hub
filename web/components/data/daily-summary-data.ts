"use client"

export interface DailySummary {
  id: number
  date: string
  company: string
  department: string
  line: string
  present: number
  late: number
  absent: number
  halfDay: number
  leave: number
  holiday: number
  total: number
}

const mockSummaries: DailySummary[] = [
  { id: 1, date: "2026-07-13", company: "HR Hub Ltd", department: "Production", line: "Line-1", present: 18, late: 2, absent: 1, halfDay: 1, leave: 1, holiday: 0, total: 23 },
  { id: 2, date: "2026-07-13", company: "HR Hub Ltd", department: "Production", line: "Line-2", present: 20, late: 1, absent: 0, halfDay: 0, leave: 1, holiday: 0, total: 22 },
  { id: 3, date: "2026-07-13", company: "HR Hub Ltd", department: "Admin", line: "Line-2", present: 10, late: 1, absent: 0, halfDay: 0, leave: 0, holiday: 0, total: 11 },
  { id: 4, date: "2026-07-13", company: "HR Hub Ltd", department: "Security", line: "Line-1", present: 14, late: 0, absent: 0, halfDay: 1, leave: 1, holiday: 0, total: 16 },
  { id: 5, date: "2026-07-13", company: "HR Hub Ltd", department: "IT", line: "Line-3", present: 5, late: 1, absent: 0, halfDay: 0, leave: 0, holiday: 0, total: 6 },
  { id: 6, date: "2026-07-13", company: "HR Hub Ltd", department: "QC", line: "Line-2", present: 8, late: 1, absent: 1, halfDay: 0, leave: 0, holiday: 0, total: 10 },
  { id: 7, date: "2026-07-13", company: "HR Hub Ltd", department: "Finance", line: "Line-2", present: 6, late: 1, absent: 0, halfDay: 0, leave: 0, holiday: 0, total: 7 },
  { id: 8, date: "2026-07-13", company: "HR Hub Ltd", department: "Cleaning", line: "Line-3", present: 8, late: 1, absent: 1, halfDay: 0, leave: 0, holiday: 0, total: 10 },
  { id: 9, date: "2026-07-13", company: "HR Hub Ltd", department: "Logistics", line: "Line-3", present: 7, late: 1, absent: 0, halfDay: 0, leave: 0, holiday: 0, total: 8 },
  { id: 10, date: "2026-07-13", company: "HR Hub Ltd", department: "HR", line: "Line-1", present: 5, late: 0, absent: 0, halfDay: 0, leave: 1, holiday: 0, total: 6 },
  { id: 11, date: "2026-07-13", company: "Tech Solutions BD", department: "IT", line: "Line-1", present: 12, late: 1, absent: 0, halfDay: 0, leave: 1, holiday: 0, total: 14 },
  { id: 12, date: "2026-07-13", company: "Tech Solutions BD", department: "HR", line: "Line-2", present: 4, late: 1, absent: 1, halfDay: 0, leave: 0, holiday: 0, total: 6 },
  { id: 13, date: "2026-07-13", company: "Garments Corp", department: "Production", line: "Line-4", present: 22, late: 0, absent: 0, halfDay: 0, leave: 1, holiday: 0, total: 23 },
  { id: 14, date: "2026-07-13", company: "Garments Corp", department: "Production", line: "Line-5", present: 20, late: 1, absent: 1, halfDay: 0, leave: 0, holiday: 0, total: 22 },
  { id: 15, date: "2026-07-13", company: "Garments Corp", department: "QC", line: "Line-5", present: 6, late: 0, absent: 0, halfDay: 0, leave: 0, holiday: 0, total: 6 },
  { id: 16, date: "2026-07-14", company: "HR Hub Ltd", department: "Production", line: "Line-1", present: 17, late: 0, absent: 1, halfDay: 0, leave: 1, holiday: 0, total: 19 },
  { id: 17, date: "2026-07-14", company: "HR Hub Ltd", department: "Production", line: "Line-2", present: 19, late: 1, absent: 0, halfDay: 0, leave: 1, holiday: 0, total: 21 },
  { id: 18, date: "2026-07-14", company: "HR Hub Ltd", department: "Admin", line: "Line-2", present: 9, late: 0, absent: 0, halfDay: 1, leave: 1, holiday: 0, total: 11 },
  { id: 19, date: "2026-07-14", company: "HR Hub Ltd", department: "Security", line: "Line-1", present: 14, late: 0, absent: 0, halfDay: 0, leave: 0, holiday: 0, total: 14 },
  { id: 20, date: "2026-07-14", company: "HR Hub Ltd", department: "IT", line: "Line-3", present: 4, late: 1, absent: 0, halfDay: 0, leave: 1, holiday: 0, total: 6 },
  { id: 21, date: "2026-07-14", company: "HR Hub Ltd", department: "QC", line: "Line-2", present: 7, late: 1, absent: 0, halfDay: 0, leave: 0, holiday: 0, total: 8 },
  { id: 22, date: "2026-07-14", company: "HR Hub Ltd", department: "Finance", line: "Line-2", present: 5, late: 0, absent: 0, halfDay: 0, leave: 0, holiday: 0, total: 5 },
  { id: 23, date: "2026-07-14", company: "HR Hub Ltd", department: "Cleaning", line: "Line-3", present: 7, late: 0, absent: 1, halfDay: 0, leave: 0, holiday: 0, total: 8 },
  { id: 24, date: "2026-07-14", company: "HR Hub Ltd", department: "Logistics", line: "Line-3", present: 6, late: 0, absent: 0, halfDay: 0, leave: 1, holiday: 0, total: 7 },
  { id: 25, date: "2026-07-14", company: "HR Hub Ltd", department: "HR", line: "Line-1", present: 4, late: 0, absent: 1, halfDay: 0, leave: 0, holiday: 0, total: 5 },
  { id: 26, date: "2026-07-14", company: "Tech Solutions BD", department: "IT", line: "Line-1", present: 10, late: 1, absent: 1, halfDay: 0, leave: 1, holiday: 0, total: 13 },
  { id: 27, date: "2026-07-14", company: "Tech Solutions BD", department: "HR", line: "Line-2", present: 3, late: 0, absent: 1, halfDay: 0, leave: 0, holiday: 0, total: 4 },
  { id: 28, date: "2026-07-14", company: "Garments Corp", department: "Production", line: "Line-4", present: 21, late: 0, absent: 0, halfDay: 1, leave: 0, holiday: 0, total: 22 },
  { id: 29, date: "2026-07-14", company: "Garments Corp", department: "Production", line: "Line-5", present: 19, late: 1, absent: 0, halfDay: 0, leave: 1, holiday: 0, total: 21 },
  { id: 30, date: "2026-07-14", company: "Garments Corp", department: "QC", line: "Line-5", present: 5, late: 1, absent: 0, halfDay: 0, leave: 0, holiday: 0, total: 6 },
  { id: 31, date: "2026-07-15", company: "HR Hub Ltd", department: "Production", line: "Line-1", present: 16, late: 0, absent: 0, halfDay: 1, leave: 1, holiday: 1, total: 19 },
  { id: 32, date: "2026-07-15", company: "HR Hub Ltd", department: "Production", line: "Line-2", present: 18, late: 0, absent: 0, halfDay: 0, leave: 0, holiday: 0, total: 18 },
  { id: 33, date: "2026-07-15", company: "HR Hub Ltd", department: "Admin", line: "Line-2", present: 8, late: 0, absent: 0, halfDay: 0, leave: 0, holiday: 1, total: 9 },
  { id: 34, date: "2026-07-15", company: "HR Hub Ltd", department: "Security", line: "Line-1", present: 12, late: 0, absent: 1, halfDay: 0, leave: 0, holiday: 0, total: 13 },
  { id: 35, date: "2026-07-15", company: "Tech Solutions BD", department: "IT", line: "Line-1", present: 11, late: 1, absent: 0, halfDay: 0, leave: 0, holiday: 0, total: 12 },
]

export function getDailySummaries(): DailySummary[] { return summaries }
let summaries = [...mockSummaries]

export const companyOptions = ["HR Hub Ltd", "Tech Solutions BD", "Garments Corp"].map((v) => ({ value: v, label: v }))
export const departmentOptions = ["Production", "Admin", "IT", "QC", "Security", "Cleaning", "Finance", "Logistics", "HR"].map((v) => ({ value: v, label: v }))
export const lineOptions = ["Line-1", "Line-2", "Line-3", "Line-4", "Line-5"].map((v) => ({ value: v, label: v }))
