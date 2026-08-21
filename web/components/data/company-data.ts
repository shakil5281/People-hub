"use client"

import { z } from "zod"

export interface Company {
  id: string
  company_name_en: string
  company_name_bn: string
  slug: string
  address_bn: string
  address_en: string
  phone: string
  email: string
  signature: string
  status: "active" | "inactive"
  created_at: string
  updated_at: string
}

export const companySchema = z.object({
  company_name_en: z.string().min(2, "Company Name (English) must be at least 2 characters"),
  company_name_bn: z.string().default(""),
  address_bn: z.string().default(""),
  address_en: z.string().default(""),
  phone: z.string().min(10, "Phone number must be at least 10 digits"),
  email: z.string().email("Invalid email address").or(z.string().length(0)).default(""),
  signature: z.string().default(""),
  status: z.enum(["active", "inactive"]).default("active"),
})

export type CompanyFormData = z.infer<typeof companySchema>

export const statusOptions: { value: "active" | "inactive"; label: string }[] = [
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
]
