"use client"

import * as React from "react"
import { useRouter, useParams } from "next/navigation"
import { IdCardIcon, Loader2 } from "lucide-react"
import { toast } from "sonner"
import { IdCardForm } from "@/components/form/id-card-form"
import { idCardApi } from "@/lib/api"
import type { IdCardFormData } from "@/components/data/id-card-data"

export default function EditIdCardPage() {
  const router = useRouter()
  const params = useParams()
  const id = params.id as string
  const [card, setCard] = React.useState<IdCardFormData | null>(null)
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    idCardApi.get(id).then(({ data }) => {
      setCard({
        employee: data.employee,
        employee_id: data.employee_id,
        designation_id: data.designation_id,
        department_id: data.department_id,
        card_no: data.card_no,
        issued: data.issued,
        expiry: data.expiry,
        status: data.status,
      })
    }).catch(() => {
      toast.error("ID card not found")
      router.push("/hr/id-card")
    }).finally(() => setLoading(false))
  }, [id, router])

  const handleSuccess = () => {
    router.push("/hr/id-card")
  }

  if (loading) {
    return <div className="flex items-center justify-center py-24"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>
  }

  if (!card) {
    return (
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          <h1 className="text-3xl font-bold tracking-tight">ID Card Not Found</h1>
          <p className="text-muted-foreground mt-1">The requested ID card does not exist.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <IdCardIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Edit ID Card</h1>
          <p className="text-muted-foreground mt-1">Update employee ID card details</p>
        </div>
      </div>
      <div className="px-4 lg:px-6">
        <IdCardForm
          initialData={card}
          onSuccess={handleSuccess}
          onCancel={() => router.push("/hr/id-card")}
          isEditing
          cardId={id}
        />
      </div>
    </div>
  )
}
