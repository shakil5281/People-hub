"use client"

import { z } from "zod"

export interface Shift {
  id: string
  company_id: string
  name: string
  shift_type: string
  start_time: string
  end_time: string
  late_grace_minutes: number
  weekend_days: string
  status: "active" | "inactive"
  created_at: string
  updated_at: string
}

export const shiftSchema = z.object({
  company_id: z.string().min(1, "Company is required"),
  name: z.string().min(2, "Shift name must be at least 2 characters"),
  shift_type: z.enum(["day", "night", "general"] as const).default("day"),
  start_time: z.string().min(1, "Start time is required"),
  end_time: z.string().min(1, "End time is required"),
  late_grace_minutes: z.number().min(0, "Must be 0 or more").default(0),
  weekend_days: z.string().default(""),
  status: z.enum(["active", "inactive"]).default("active"),
})

export type ShiftFormData = z.infer<typeof shiftSchema>

export const statusOptions = [
  { value: "active" as const, label: "Active" },
  { value: "inactive" as const, label: "Inactive" },
]

export const dayOptions = [
  { value: "Fri", label: "Friday" },
  { value: "Sat", label: "Saturday" },
  { value: "Sun", label: "Sunday" },
  { value: "Mon", label: "Monday" },
  { value: "Tue", label: "Tuesday" },
  { value: "Wed", label: "Wednesday" },
  { value: "Thu", label: "Thursday" },
]
