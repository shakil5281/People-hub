"use client"

import { useRouter } from "next/navigation"
import { UserXIcon } from "lucide-react"
import { SeparationForm } from "@/components/form/separation-form"

export default function CreateSeparationPage() {
  const router = useRouter()

  const handleSuccess = () => {
    router.push("/hr/seperation")
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <UserXIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Create Separation</h1>
          <p className="text-muted-foreground mt-1">Record an employee separation</p>
        </div>
      </div>
      <div className="px-4 lg:px-6">
        <SeparationForm onSuccess={handleSuccess} onCancel={() => router.back()} />
      </div>
    </div>
  )
}
