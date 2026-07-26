"use client"

import * as React from "react"
import { useRouter, useParams } from "next/navigation"
import { UsersIcon, Loader2 } from "lucide-react"
import { EmployeeForm } from "@/components/form/employee-form"
import { employeeApi } from "@/lib/api"
import type { Employee } from "@/components/data/employee-data"

export default function EditEmployeePage() {
  const router = useRouter()
  const params = useParams()
  const id = params.id as string
  const [employee, setEmployee] = React.useState<Employee | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")

  React.useEffect(() => {
    async function fetchEmployee() {
      try {
        const { data } = await employeeApi.get(id)
        setEmployee(data)
      } catch {
        setError("Failed to load employee")
      } finally {
        setLoading(false)
      }
    }
    fetchEmployee()
  }, [id])

  if (loading) {
    return (
      <div className="container mx-auto py-8 flex items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error || !employee) {
    return (
      <div className="container mx-auto py-8">
        <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">
          {error || "Employee not found"}
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <UsersIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Edit Employee</h1>
          <p className="text-muted-foreground mt-1">Update employee information</p>
        </div>
      </div>
      <div className="px-4 lg:px-6">
        <EmployeeForm
          isEditing
          employeeId={id}
          initialData={{
            name_en: employee.name_en,
            name_bn: employee.name_bn,
            father_name: employee.father_name,
            mother_name: employee.mother_name,
            date_of_birth: employee.date_of_birth?.split("T")[0] || "",
            gender: employee.gender,
            blood_group: employee.blood_group,
            marital_status: employee.marital_status,
            religion: employee.religion,
            nationality: employee.nationality,
            nid: employee.nid,
            phone: employee.phone,
            email: employee.email,
            present_address: employee.present_address,
            permanent_address: employee.permanent_address,
            spouse_name: employee.spouse_name,
            emergency_contact: employee.emergency_contact,
            emergency_phone: employee.emergency_phone,
            number_of_dependents: employee.number_of_dependents || 0,
            company_id: employee.company_id,
            employee_id: employee.employee_id,
            punch_number: employee.punch_number,
            employee_type: employee.employee_type,
            grade: employee.grade,
            joining_date: employee.joining_date?.split("T")[0] || "",
            shift_id: employee.shift_id,
            department_id: employee.department_id,
            section_id: employee.section_id,
            designation_id: employee.designation_id,
            line_id: employee.line_id,
            group_id: employee.group_id,
            floor_id: employee.floor_id,
            reports_to: employee.reports_to,
            present_division_id: employee.present_division_id,
            present_district_id: employee.present_district_id,
            present_upazila_id: employee.present_upazila_id,
            present_union_id: employee.present_union_id,
            permanent_division_id: employee.permanent_division_id,
            permanent_district_id: employee.permanent_district_id,
            permanent_upazila_id: employee.permanent_upazila_id,
            permanent_union_id: employee.permanent_union_id,
            transport_allowance: employee.transport_allowance || 0,
            food_allowance: employee.food_allowance || 0,
            other_allowance: employee.other_allowance || 0,
            gross_salary: employee.gross_salary || 0,
            basic_salary: employee.basic_salary || 0,
            house_rent: employee.house_rent || 0,
            medical_allowance: employee.medical_allowance || 0,
            account_type: employee.account_type,
            account_number: employee.account_number,
            status: employee.status,
            over_time_status: employee.over_time_status,
            image_url: employee.image_url,
            signature_url: employee.signature_url,
          }}
          onSuccess={() => router.push("/hr/employees")}
          onCancel={() => router.back()}
        />
      </div>
    </div>
  )
}
