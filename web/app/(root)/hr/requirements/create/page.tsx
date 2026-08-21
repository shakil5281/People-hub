"use client"

import { useRouter } from "next/navigation"
import { ClipboardListIcon } from "lucide-react"
import { RequirementForm } from "@/components/form/requirement-form"

export default function CreateRequirementPage() {
  const router = useRouter()

  const handleSuccess = () => {
    router.push("/hr/requirements")
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <ClipboardListIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Create Requirement</h1>
          <p className="text-muted-foreground mt-1">Add a new recruitment requirement</p>
        </div>
      </div>
      <div className="px-4 lg:px-6">
        <RequirementForm onSuccess={handleSuccess} onCancel={() => router.back()} />
      </div>
    </div>
  )
}
