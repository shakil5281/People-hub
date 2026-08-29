"use client"

import { z } from "zod"
import { rosterApi } from "@/lib/api"

export interface Roster {
  id: string
  company_id: string
  employee_id: string
  shift_id: string
  date: string
  reason: string
  status: string
  employee?: { id: string; name_en: string; employee_id: string }
  shift?: { id: string; name: string }
}

export const rosterSchema = z.object({
  employee_id: z.string().min(1, "Employee is required"),
  shift_id: z.string().min(1, "Shift is required"),
  from_date: z.string().min(1, "From date is required"),
  to_date: z.string().optional(),
  reason: z.string().optional(),
  status: z.string().default("active"),
  company_id: z.string().optional(),
})

export const bulkRosterSchema = z.object({
  shift_id: z.string().min(1, "Shift is required"),
  from_date: z.string().min(1, "From date is required"),
  to_date: z.string().optional(),
  reason: z.string().optional(),
  status: z.string().optional(),
  employee_ids: z.array(z.string()).optional(),
  department_id: z.string().optional(),
  section_id: z.string().optional(),
  line_id: z.string().optional(),
  group_id: z.string().optional(),
})

export type RosterFormData = z.infer<typeof rosterSchema>
export type BulkRosterFormData = z.infer<typeof bulkRosterSchema>

export async function getRosters(companyId?: string, params?: Record<string, string>): Promise<{ data: Roster[]; total: number; total_pages: number }> {
  const reqParams: Record<string, string> = { ...params }
  if (companyId) reqParams.company_id = companyId
  const res = await rosterApi.list(reqParams)
  return {
    data: (res.data?.data || res.data || []) as Roster[],
    total: res.data?.total || 0,
    total_pages: res.data?.total_pages || 0,
  }
}

export async function createRoster(data: Record<string, unknown>): Promise<boolean> {
  const res = await rosterApi.create(data)
  return res.status === 201 || res.status === 200
}

export async function bulkCreateRoster(data: Record<string, unknown>): Promise<boolean> {
  const res = await rosterApi.bulkCreate(data)
  return res.status === 201 || res.status === 200
}

export async function updateRoster(id: string, data: Record<string, unknown>): Promise<boolean> {
  const res = await rosterApi.update(id, data)
  return res.status === 200
}

export async function deleteRoster(id: string): Promise<boolean> {
  const res = await rosterApi.delete(id)
  return res.status === 200
}
