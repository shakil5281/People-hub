"use client"

import { CompanyForm } from "@/components/form/company-form"
import { useRouter } from "next/navigation"

export default function CreateCompanyPage() {
  const router = useRouter()

  return (
    <div className="container mx-auto py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight">Create Company</h1>
        <p className="text-muted-foreground mt-1">Add a new company to the system</p>
      </div>
      <CompanyForm
        onSuccess={() => router.push("/information/company")}
        onCancel={() => router.back()}
      />
    </div>
  )
}