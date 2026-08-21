"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2 } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { TimePicker } from "@/components/ui/time-picker"
import { NativeSelect, NativeSelectOption } from "@/components/ui/native-select"
import { shiftSchema, ShiftFormData, statusOptions, dayOptions } from "../data/shift-data"
import { shiftApi, companyApi } from "@/lib/api"
import type { Company } from "@/components/data/company-data"

interface ShiftFormProps {
  initialData?: Partial<ShiftFormData>
  onSuccess: () => void
  onCancel?: () => void
  isEditing?: boolean
  shiftId?: string
}

export function ShiftForm({ initialData, onSuccess, onCancel, isEditing = false, shiftId }: ShiftFormProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [error, setError] = React.useState("")
  const [companies, setCompanies] = React.useState<Company[]>([])

  const { handleSubmit, formState: { errors }, setValue, watch, register } = useForm({
    resolver: zodResolver(shiftSchema),
    defaultValues: {
      company_id: initialData?.company_id || "",
      name: initialData?.name || "",
      shift_type: (initialData?.shift_type as "day" | "night" | "general") || "day",
      start_time: initialData?.start_time || "",
      end_time: initialData?.end_time || "",
      late_grace_minutes: initialData?.late_grace_minutes ?? 0,
      weekend_days: initialData?.weekend_days || "",
      status: (initialData?.status as "active" | "inactive") || "active",
    },
  })

  const startTime = watch("start_time")
  const endTime = watch("end_time")
  const shiftType = watch("shift_type")
  const statusVal = watch("status")
  const companyId = watch("company_id")
  const weekendDay = watch("weekend_days")

  React.useEffect(() => {
    companyApi.list().then((res) => {
      const list = Array.isArray(res.data?.data) ? res.data.data : Array.isArray(res.data) ? res.data : []
      setCompanies(list)
    }).catch(() => {})
  }, [])

  const onSubmit = async (data: ShiftFormData) => {
    setIsSubmitting(true)
    setError("")
    try {
      if (isEditing && shiftId) {
        await shiftApi.update(shiftId, data)
        toast.success("Shift updated successfully")
      } else {
        await shiftApi.create(data)
        toast.success("Shift created successfully")
      }
      onSuccess()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to save shift"
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
        <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="space-y-2">
        <Label>Company *</Label>
        <NativeSelect
          value={companyId}
          onChange={(e) => setValue("company_id", e.target.value)}
        >
          <NativeSelectOption value="">Select company</NativeSelectOption>
          {companies.map((c) => (
            <NativeSelectOption key={c.id} value={c.id}>{c.company_name_en}</NativeSelectOption>
          ))}
        </NativeSelect>
        {errors.company_id && <p className="text-sm text-destructive">{errors.company_id.message}</p>}
      </div>

      <div className="space-y-2">
        <Label>Shift Name *</Label>
        <Input
          placeholder="Morning Shift"
          {...register("name")}
          aria-invalid={!!errors.name}
        />
        {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
      </div>

      <div className="space-y-2">
        <Label>Shift Type *</Label>
        <NativeSelect
          value={shiftType}
          onChange={(e) => setValue("shift_type", e.target.value as "day" | "night" | "general")}
        >
          <NativeSelectOption value="day">Day</NativeSelectOption>
          <NativeSelectOption value="night">Night</NativeSelectOption>
          <NativeSelectOption value="general">General</NativeSelectOption>
        </NativeSelect>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label>Start Time *</Label>
          <TimePicker value={startTime} onChange={(val) => setValue("start_time", val)} />
          {errors.start_time && <p className="text-sm text-destructive">{errors.start_time.message}</p>}
        </div>
        <div className="space-y-2">
          <Label>End Time *</Label>
          <TimePicker value={endTime} onChange={(val) => setValue("end_time", val)} />
          {errors.end_time && <p className="text-sm text-destructive">{errors.end_time.message}</p>}
        </div>
      </div>

      <div className="space-y-2">
        <Label>Late Grace (minutes)</Label>
        <Input
          type="number"
          min="0"
          {...register("late_grace_minutes", { valueAsNumber: true })}
        />
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label>Weekend Day</Label>
          <NativeSelect
            value={weekendDay}
            onChange={(e) => setValue("weekend_days", e.target.value)}
          >
            <NativeSelectOption value="">None</NativeSelectOption>
            {dayOptions.map((opt) => (
              <NativeSelectOption key={opt.value} value={opt.value}>{opt.label}</NativeSelectOption>
            ))}
          </NativeSelect>
        </div>
        <div className="space-y-2">
          <Label>Status</Label>
          <NativeSelect
            value={statusVal}
            onChange={(e) => setValue("status", e.target.value as "active" | "inactive")}
          >
            {statusOptions.map((opt) => (
              <NativeSelectOption key={opt.value} value={opt.value}>{opt.label}</NativeSelectOption>
            ))}
          </NativeSelect>
        </div>
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
            isEditing ? "Update Shift" : "Create Shift"
          )}
        </Button>
      </div>
    </form>
  )
}
