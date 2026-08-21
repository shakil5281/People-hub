"use client"

import { z } from "zod"
import type { Department } from "./organization-data"

export interface Separation {
  id: string
  employee: string
  employee_id: string
  department_id: string
  type: "Resign" | "Lefty" | "Close"
  date: string
  status: "Approved" | "Pending" | "Rejected" | "Processed" | "Cancelled"
  reason: string
  created_at: string
  updated_at: string
  department?: Department
}

export const separationSchema = z.object({
  employee: z.string().min(2, "Employee name is required"),
  employee_id: z.string().min(1, "Employee code is required"),
  department_id: z.string().min(1, "Department is required"),
  type: z.enum(["Resign", "Lefty", "Close"]),
  date: z.string().min(1, "Date is required"),
  status: z.enum(["Approved", "Pending", "Rejected", "Processed", "Cancelled"]),
  reason: z.string().optional(),
})

export type SeparationFormData = z.infer<typeof separationSchema>

export const separationTypeOptions = [
  { value: "Resign" as const, label: "Resign" },
  { value: "Lefty" as const, label: "Lefty" },
  { value: "Close" as const, label: "Close" },
]

export const separationStatusOptions = [
  { value: "Approved" as const, label: "Approved" },
  { value: "Pending" as const, label: "Pending" },
  { value: "Rejected" as const, label: "Rejected" },
  { value: "Processed" as const, label: "Processed" },
  { value: "Cancelled" as const, label: "Cancelled" },
]
