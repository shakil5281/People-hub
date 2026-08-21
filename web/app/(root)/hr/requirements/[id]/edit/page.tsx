"use client"

import * as React from "react"
import { useRouter, useParams } from "next/navigation"
import { ClipboardListIcon, Loader2 } from "lucide-react"
import { toast } from "sonner"
import { RequirementForm } from "@/components/form/requirement-form"
import { requirementApi } from "@/lib/api"
import type { RequirementFormData } from "@/components/data/requirement-data"

export default function EditRequirementPage() {
  const router = useRouter()
  const params = useParams()
  const id = params.id as string
  const [requirement, setRequirement] = React.useState<RequirementFormData | null>(null)
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    requirementApi.get(id).then(({ data }) => {
      setRequirement({
        position: data.position,
        department_id: data.department_id || "",
        section_id: data.section_id || "",
        designation_id: data.designation_id || "",
        group_type: data.group_type || "Worker",
        vacancies: data.vacancies,
        applicants: data.applicants,
        status: data.status,
        priority: data.priority,
        description: data.description || "",
      })
    }).catch(() => {
      toast.error("Requirement not found")
      router.push("/hr/requirements")
    }).finally(() => setLoading(false))
  }, [id, router])

  const handleSuccess = () => {
    router.push("/hr/requirements")
  }

  if (loading) {
    return <div className="flex items-center justify-center py-24"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>
  }

  if (!requirement) {
    return (
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          <h1 className="text-3xl font-bold tracking-tight">Requirement Not Found</h1>
          <p className="text-muted-foreground mt-1">The requested requirement does not exist.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <ClipboardListIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Edit Requirement</h1>
          <p className="text-muted-foreground mt-1">Update recruitment requirement</p>
        </div>
      </div>
      <div className="px-4 lg:px-6">
        <RequirementForm
          initialData={requirement}
          onSuccess={handleSuccess}
          onCancel={() => router.push("/hr/requirements")}
          isEditing
          requirementId={id}
        />
      </div>
    </div>
  )
}
