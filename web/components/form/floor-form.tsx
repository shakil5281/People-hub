"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2 } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { floorSchema } from "../data/floor-data"
import { floorApi } from "@/lib/api"

interface FloorFormProps {
  initialData?: { name: string }
  onSuccess: () => void
  onCancel?: () => void
  isEditing?: boolean
  floorId?: string
}

export function FloorForm({ initialData, onSuccess, onCancel, isEditing = false, floorId }: FloorFormProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [error, setError] = React.useState("")

  const { register, handleSubmit, formState: { errors } } = useForm({
    resolver: zodResolver(floorSchema),
    defaultValues: {
      name: initialData?.name || "",
    },
  })

  const onSubmit = async (data: Record<string, unknown>) => {
    setIsSubmitting(true)
    setError("")
    try {
      if (isEditing && floorId) {
        await floorApi.update(floorId, data)
        toast.success("Floor updated successfully")
      } else {
        await floorApi.create(data)
        toast.success("Floor created successfully")
      }
      onSuccess()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to save floor"
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
        <Label htmlFor="name">Floor Name *</Label>
        <Input
          id="name"
          placeholder="Enter floor name"
          {...register("name")}
          aria-invalid={!!errors.name}
        />
        {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
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
            isEditing ? "Update Floor" : "Create Floor"
          )}
        </Button>
      </div>
    </form>
  )
}
