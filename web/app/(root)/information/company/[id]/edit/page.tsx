"use client"

import * as React from "react"
import { useRouter, useParams } from "next/navigation"
import { Loader2 } from "lucide-react"
import { CompanyForm } from "@/components/form/company-form"
import { companyApi } from "@/lib/api"
import type { Company } from "@/components/data/company-data"

export default function EditCompanyPage() {
  const router = useRouter()
  const params = useParams()
  const id = params.id as string
  const [company, setCompany] = React.useState<Company | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")

  React.useEffect(() => {
    async function fetchCompany() {
      try {
        const { data } = await companyApi.get(id)
        setCompany(data)
      } catch {
        setError("Failed to load company")
      } finally {
        setLoading(false)
      }
    }
    fetchCompany()
  }, [id])

  if (loading) {
    return (
      <div className="container mx-auto py-8 flex items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error || !company) {
    return (
      <div className="container mx-auto py-8">
        <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">
          {error || "Company not found"}
        </div>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight">Edit Company</h1>
        <p className="text-muted-foreground mt-1">Update company information</p>
      </div>
      <CompanyForm
        isEditing
        companyId={id}
        initialData={{
          company_name_en: company.company_name_en,
          company_name_bn: company.company_name_bn,
          address_bn: company.address_bn,
          address_en: company.address_en,
          phone: company.phone,
          email: company.email,
          signature: company.signature,
          status: company.status,
        }}
        onSuccess={() => router.push("/information/company")}
        onCancel={() => router.back()}
      />
    </div>
  )
}
