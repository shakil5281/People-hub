"use client"

import { z } from "zod"

export interface Employee {
  id: string
  // Personal
  name_en: string
  name_bn: string
  father_name: string
  mother_name: string
  date_of_birth: string
  gender: string
  blood_group: string
  marital_status: string
  religion: string
  nationality: string
  nid: string
  phone: string
  email: string
  present_address: string
  permanent_address: string
  // Family
  spouse_name: string
  emergency_contact: string
  emergency_phone: string
  number_of_dependents: number
  // Office
  company_id: string
  employee_id: string
  punch_number: string
  employee_type: string
  grade: string
  joining_date: string
  shift_id: string | null
  department_id: string | null
  reports_to: string | null
  section_id: string | null
  designation_id: string | null
  line_id: string | null
  group_id: string | null
  floor_id: string | null
  // Relations (returned by API via Preload)
  department?: { id: string; name: string }
  section_ref?: { id: string; name: string }
  designation_ref?: { id: string; name: string }
  line_ref?: { id: string; name: string }
  group_ref?: { id: string; name: string }
  floor_ref?: { id: string; name: string }
  shift?: { id: string; name: string }
  // Address (present)
  present_division_id: string | null
  present_district_id: string | null
  present_upazila_id: string | null
  present_union_id: string | null
  present_post_office: string | null
  present_post_code: string | null
  // Address (permanent)
  permanent_division_id: string | null
  permanent_district_id: string | null
  permanent_upazila_id: string | null
  permanent_union_id: string | null
  permanent_post_office: string | null
  permanent_post_code: string | null
  // Salary
  gross_salary: number
  basic_salary: number
  house_rent: number
  transport_allowance: number
  food_allowance: number
  medical_allowance: number
  other_allowance: number
  // Account
  account_type: string
  account_number: string
  // Status & Files
  status: "active" | "inactive"
  over_time_status: boolean
  signature_url: string
  image_url: string
}

export const employeeSchema = z.object({
  // Personal
  name_en: z.string().min(2, "Name (English) is required"),
  name_bn: z.string().optional(),
  father_name: z.string().optional(),
  mother_name: z.string().optional(),
  date_of_birth: z.string().optional(),
  gender: z.string().optional(),
  blood_group: z.string().optional(),
  marital_status: z.string().optional(),
  religion: z.string().optional(),
  nationality: z.string().optional(),
  nid: z.string().optional(),
  phone: z.string().optional(),
  email: z.string().optional(),
  present_address: z.string().optional(),
  permanent_address: z.string().optional(),
  // Family
  spouse_name: z.string().optional(),
  emergency_contact: z.string().optional(),
  emergency_phone: z.string().optional(),
  number_of_dependents: z.number().optional(),
  // Office
  company_id: z.string().min(1, "Company is required"),
  employee_id: z.string().min(1, "Employee ID is required"),
  punch_number: z.string().min(1, "Punch number is required"),
  employee_type: z.string().optional(),
  grade: z.string().optional(),
  joining_date: z.string().min(1, "Joining date is required"),
  shift_id: z.string().optional().nullable(),
  department_id: z.string().optional().nullable(),
  reports_to: z.string().optional().nullable(),
  section_id: z.string().optional().nullable(),
  designation_id: z.string().optional().nullable(),
  line_id: z.string().optional().nullable(),
  group_id: z.string().optional().nullable(),
  floor_id: z.string().optional().nullable(),
  // Address (present)
  present_division_id: z.string().optional().nullable(),
  present_district_id: z.string().optional().nullable(),
  present_upazila_id: z.string().optional().nullable(),
  present_union_id: z.string().optional().nullable(),
  present_post_office: z.string().optional().nullable(),
  present_post_code: z.string().optional().nullable(),
  // Address (permanent)
  permanent_division_id: z.string().optional().nullable(),
  permanent_district_id: z.string().optional().nullable(),
  permanent_upazila_id: z.string().optional().nullable(),
  permanent_union_id: z.string().optional().nullable(),
  permanent_post_office: z.string().optional().nullable(),
  permanent_post_code: z.string().optional().nullable(),
  // Salary
  gross_salary: z.number().optional(),
  basic_salary: z.number().optional(),
  house_rent: z.number().optional(),
  transport_allowance: z.number().optional(),
  food_allowance: z.number().optional(),
  medical_allowance: z.number().optional(),
  other_allowance: z.number().optional(),
  // Account
  account_type: z.string().optional(),
  account_number: z.string().optional(),
  // Status & Files
  status: z.enum(["active", "inactive"]).optional(),
  over_time_status: z.boolean().optional(),
  image_url: z.string().optional(),
  signature_url: z.string().optional(),
})

export type EmployeeFormData = z.input<typeof employeeSchema>

export const genderOptions = [
  { value: "Male", label: "Male" },
  { value: "Female", label: "Female" },
  { value: "Other", label: "Other" },
]

export const bloodGroupOptions = [
  { value: "A+", label: "A+" }, { value: "A-", label: "A-" },
  { value: "B+", label: "B+" }, { value: "B-", label: "B-" },
  { value: "AB+", label: "AB+" }, { value: "AB-", label: "AB-" },
  { value: "O+", label: "O+" }, { value: "O-", label: "O-" },
]

export const maritalStatusOptions = [
  { value: "Single", label: "Single" },
  { value: "Married", label: "Married" },
  { value: "Divorced", label: "Divorced" },
  { value: "Widowed", label: "Widowed" },
]

export const statusOptionsEmployee = [
  { value: "active" as const, label: "Active" },
  { value: "inactive" as const, label: "Inactive" },
]

export const religionOptions = [
  { value: "Islam", label: "Islam" },
  { value: "Hinduism", label: "Hinduism" },
  { value: "Christianity", label: "Christianity" },
  { value: "Buddhism", label: "Buddhism" },
  { value: "Others", label: "Others" },
]

export const employeeTypeOptions = [
  { value: "Regular", label: "Regular" },
  { value: "Lefty", label: "Lefty" },
  { value: "Close", label: "Close" },
  { value: "Resign", label: "Resign" },
]
