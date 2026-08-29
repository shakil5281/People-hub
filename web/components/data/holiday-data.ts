"use client"

import { z } from "zod"

export interface Holiday {
  id: string
  company_id: string
  name: string
  date: string
  from_date?: string | null
  to_date?: string | null
  weekend_date?: string | null
  type: string
  description: string
  status: string
}

export const bulkAdvanceHolidaySchema = z.object({
  name: z.string().min(1, "Name is required"),
  from_date: z.string().min(1, "From date is required"),
  to_date: z.string().optional(),
  description: z.string().optional(),
  advance_only: z.boolean().optional(),
})

export type BulkAdvanceHolidayFormData = z.infer<typeof bulkAdvanceHolidaySchema>
