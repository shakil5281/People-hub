"use client"

import * as React from "react"
import { useRouter, useParams } from "next/navigation"
import { UserXIcon, Loader2 } from "lucide-react"
import { toast } from "sonner"
import { SeparationForm } from "@/components/form/separation-form"
import { separationApi } from "@/lib/api"
import type { SeparationFormData } from "@/components/data/separation-data"

export default function EditSeparationPage() {
  const router = useRouter()
  const params = useParams()
  const id = params.id as string
  const [separation, setSeparation] = React.useState<SeparationFormData | null>(null)
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    separationApi.get(id).then(({ data }) => {
      setSeparation({
        employee: data.employee,
        employee_id: data.employee_id,
        department_id: data.department_id,
        type: data.type,
        date: data.date,
        status: data.status,
        reason: data.reason || "",
      })
    }).catch(() => {
      toast.error("Separation not found")
      router.push("/hr/seperation")
    }).finally(() => setLoading(false))
  }, [id, router])

  const handleSuccess = () => {
    router.push("/hr/seperation")
  }

  if (loading) {
    return <div className="flex items-center justify-center py-24"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>
  }

  if (!separation) {
    return (
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          <h1 className="text-3xl font-bold tracking-tight">Separation Not Found</h1>
          <p className="text-muted-foreground mt-1">The requested separation does not exist.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <UserXIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Edit Separation</h1>
          <p className="text-muted-foreground mt-1">Update employee separation record</p>
        </div>
      </div>
      <div className="px-4 lg:px-6">
        <SeparationForm
          initialData={separation}
          onSuccess={handleSuccess}
          onCancel={() => router.push("/hr/seperation")}
          isEditing
          separationId={id}
        />
      </div>
    </div>
  )
}
