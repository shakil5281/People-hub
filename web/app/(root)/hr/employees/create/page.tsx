"use client"

import { useRouter } from "next/navigation"
import { UsersIcon } from "lucide-react"
import { EmployeeForm } from "@/components/form/employee-form"

export default function CreateEmployeePage() {
  const router = useRouter()

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <UsersIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Create Employee</h1>
          <p className="text-muted-foreground mt-1">Add a new employee to the system</p>
        </div>
      </div>
      <div className="px-4 lg:px-6">
        <EmployeeForm onSuccess={() => router.push("/hr/employees")} onCancel={() => router.back()} />
      </div>
    </div>
  )
}
