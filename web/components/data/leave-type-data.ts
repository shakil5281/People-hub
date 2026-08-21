"use client"

import { z } from "zod"

export interface LeaveType {
  id: string
  company_id: string
  name: string
  total_days: number
  status: string
}

export const leaveTypeSchema = z.object({
  company_id: z.string().min(1, "Company is required"),
  name: z.string().min(1, "Name is required"),
  total_days: z.coerce.number().min(1, "Must be at least 1"),
})

export type LeaveTypeFormData = z.infer<typeof leaveTypeSchema>
