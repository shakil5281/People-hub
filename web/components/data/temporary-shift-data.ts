"use client"

import { z } from "zod"
import { temporaryShiftApi } from "@/lib/api"

export interface TempShift {
  id: string
  employee_id: string
  shift_id: string
  company_id: string
  date: string
  reason: string
  status: string
  employee?: { id: string; name_en: string; employee_id: string }
  shift?: { id: string; name: string }
}

export const tempShiftSchema = z.object({
  employee_id: z.string().min(1, "Employee is required"),
  shift_id: z.string().min(1, "Shift is required"),
  from_date: z.string().min(1, "From date is required"),
  to_date: z.string().optional(),
  reason: z.string().min(1, "Reason is required"),
  status: z.string().min(1),
  company_id: z.string().optional(),
})

export type TempShiftFormData = z.infer<typeof tempShiftSchema>

export async function getTempShifts(companyId?: string, params?: Record<string, string>): Promise<{ data: TempShift[]; total: number; total_pages: number }> {
  const reqParams: Record<string, string> = { ...params }
  if (companyId) reqParams.company_id = companyId
  const res = await temporaryShiftApi.list(reqParams)
  return {
    data: (res.data?.data || res.data || []) as TempShift[],
    total: res.data?.total || 0,
    total_pages: res.data?.total_pages || 0,
  }
}

export async function createTempShift(data: Record<string, unknown>): Promise<boolean> {
  const res = await temporaryShiftApi.create(data)
  return res.status === 200
}

export async function updateTempShift(id: string, data: Record<string, unknown>): Promise<boolean> {
  const res = await temporaryShiftApi.update(id, data)
  return res.status === 200
}

export async function deleteTempShift(id: string): Promise<boolean> {
  const res = await temporaryShiftApi.delete(id)
  return res.status === 200
}
