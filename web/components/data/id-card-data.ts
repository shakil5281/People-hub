"use client"

import { z } from "zod"
import type { Department, Designation } from "./organization-data"

export interface IdCard {
  id: string
  employee: string
  employee_id: string
  designation_id: string
  department_id: string
  card_no: string
  issued: string
  expiry: string
  status: "Active" | "Expired" | "Lost" | "Damaged"
  created_at: string
  updated_at: string
  department?: Department
  designation?: Designation
}

export const idCardSchema = z.object({
  employee: z.string().min(2, "Employee name is required"),
  employee_id: z.string().min(1, "Employee code is required"),
  designation_id: z.string().min(1, "Designation is required"),
  department_id: z.string().min(1, "Department is required"),
  card_no: z.string().min(1, "Card number is required"),
  issued: z.string().min(1, "Issue date is required"),
  expiry: z.string().min(1, "Expiry date is required"),
  status: z.enum(["Active", "Expired", "Lost", "Damaged"]),
})

export type IdCardFormData = z.infer<typeof idCardSchema>

export const idCardStatusOptions = [
  { value: "Active" as const, label: "Active" },
  { value: "Expired" as const, label: "Expired" },
  { value: "Lost" as const, label: "Lost" },
  { value: "Damaged" as const, label: "Damaged" },
]
