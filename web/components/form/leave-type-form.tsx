"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2, PlusIcon, XIcon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { leaveTypeSchema, type LeaveTypeFormData, type LeaveType } from "../data/leave-type-data"
import { leaveTypeApi, companyApi } from "@/lib/api"

interface Company { id: string; company_name_en: string }

interface LeaveTypeFormProps {
  initialData?: LeaveType | null
  onSuccess: () => void
  onCancel?: () => void
  isEditing?: boolean
  leaveTypeId?: string
}

export function LeaveTypeForm({ initialData, onSuccess, onCancel, isEditing = false, leaveTypeId }: LeaveTypeFormProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [error, setError] = React.useState("")
  const [companies, setCompanies] = React.useState<Company[]>([])

  const { register, handleSubmit, formState: { errors }, setValue } = useForm({
    resolver: zodResolver(leaveTypeSchema),
    defaultValues: {
      company_id: initialData?.company_id || "",
      name: initialData?.name || "",
      total_days: initialData?.total_days || 1,
    },
  })

  React.useEffect(() => {
    companyApi.list().then((res) => {
      const list: Company[] = res.data?.data ?? []
      setCompanies(list)
      if (!initialData?.company_id && list.length === 1) {
        setValue("company_id", list[0].id)
      }
    }).catch(() => {})
  }, [])

  const onSubmit = async (raw: any) => {
    setIsSubmitting(true)
    setError("")
    try {
      if (isEditing && leaveTypeId) {
        const data = {
          ...raw,
          total_days: Number(raw.total_days),
        }
        await leaveTypeApi.update(leaveTypeId, data as unknown as Record<string, unknown>)
        toast.success("Leave type updated successfully")
      } else {
        const code = raw.name.split(" ").map((w: string) => w[0]).join("").toUpperCase()
        const data = {
          ...raw,
          total_days: Number(raw.total_days),
          code,
        }
        await leaveTypeApi.create(data as unknown as Record<string, unknown>)
        toast.success("Leave type created successfully")
      }
      onSuccess()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to save leave type"
      let detail = message
      if (typeof err === "object" && err !== null && "response" in err) {
        const axiosErr = err as { response?: { data?: { error?: string } } }
        detail = axiosErr.response?.data?.error || message
      }
      setError(detail)
      toast.error(detail)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {error && (
        <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
      )}

      <div className="space-y-2">
        <Label htmlFor="company_id">Company *</Label>
        <select
          id="company_id"
          {...register("company_id")}
          className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        >
          <option value="">Select company</option>
          {companies.map((c) => (
            <option key={c.id} value={c.id}>{c.company_name_en}</option>
          ))}
        </select>
        {errors.company_id && <p className="text-sm text-destructive">{errors.company_id.message}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="name">Leave Type *</Label>
        <Input id="name" placeholder="e.g. Annual Leave" {...register("name")} aria-invalid={!!errors.name} />
        {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="total_days">Balance *</Label>
        <Input id="total_days" type="number" min="1" {...register("total_days")} />
        {errors.total_days && <p className="text-sm text-destructive">{errors.total_days.message}</p>}
      </div>

      <div className="flex justify-end gap-4 pt-4 border-t">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>
            Cancel
          </Button>
        )}
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? (
            <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</>
          ) : (
            isEditing ? "Update Leave Type" : "Create Leave Type"
          )}
        </Button>
      </div>
    </form>
  )
}
