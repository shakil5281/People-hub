"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2, IdCardIcon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import {
  idCardSchema, IdCardFormData, idCardStatusOptions,
} from "../data/id-card-data"
import { idCardApi, departmentApi, designationApi } from "@/lib/api"
import type { Department, Designation } from "@/components/data/organization-data"

interface IdCardFormProps {
  initialData?: Partial<IdCardFormData>
  onSuccess: () => void
  onCancel?: () => void
  isEditing?: boolean
  cardId?: string
}

export function IdCardForm({ initialData, onSuccess, onCancel, isEditing = false, cardId }: IdCardFormProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [error, setError] = React.useState("")
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])

  const { handleSubmit, formState: { errors }, setValue, watch, register } = useForm<IdCardFormData>({
    resolver: zodResolver(idCardSchema),
    defaultValues: {
      employee: "", employee_id: "", designation_id: "", department_id: "",
      card_no: "", issued: "", expiry: "", status: "Active",
      ...initialData,
    },
  })

  const departmentId = watch("department_id")
  const designationId = watch("designation_id")
  const statusVal = watch("status")

  React.useEffect(() => {
    departmentApi.list().then(({ data }) => {
      setDepartments(Array.isArray(data) ? data : [])
    }).catch(() => {})
    designationApi.list().then(({ data }) => {
      setDesignations(Array.isArray(data) ? data : [])
    }).catch(() => {})
  }, [])

  const onSubmit = async (data: IdCardFormData) => {
    setIsSubmitting(true)
    setError("")
    try {
      if (isEditing && cardId) {
        await idCardApi.update(cardId, data)
        toast.success("ID card updated successfully")
      } else {
        await idCardApi.create(data)
        toast.success("ID card created successfully")
      }
      onSuccess()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to save ID card"
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
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {error && (
        <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
      )}
      <Card>
        <CardHeader><CardTitle className="flex items-center gap-2"><IdCardIcon className="h-5 w-5" /> ID Card Details</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="employee">Employee Name *</Label>
              <Input id="employee" {...register("employee")} aria-invalid={!!errors.employee} />
              {errors.employee && <p className="text-sm text-destructive">{errors.employee.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="employee_id">Emp. ID *</Label>
              <Input id="employee_id" {...register("employee_id")} aria-invalid={!!errors.employee_id} />
              {errors.employee_id && <p className="text-sm text-destructive">{errors.employee_id.message}</p>}
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="designation_id">Designation *</Label>
              <Select value={designationId} onValueChange={(val) => setValue("designation_id", val)}>
                <SelectTrigger id="designation_id"><SelectValue placeholder="Select designation" /></SelectTrigger>
                <SelectContent>
                  {designations.map((d) => <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>)}
                </SelectContent>
              </Select>
              {errors.designation_id && <p className="text-sm text-destructive">{errors.designation_id.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="department_id">Department *</Label>
              <Select value={departmentId} onValueChange={(val) => setValue("department_id", val)}>
                <SelectTrigger id="department_id"><SelectValue placeholder="Select department" /></SelectTrigger>
                <SelectContent>
                  {departments.map((d) => <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>)}
                </SelectContent>
              </Select>
              {errors.department_id && <p className="text-sm text-destructive">{errors.department_id.message}</p>}
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="card_no">Card Number *</Label>
              <Input id="card_no" {...register("card_no")} aria-invalid={!!errors.card_no} />
              {errors.card_no && <p className="text-sm text-destructive">{errors.card_no.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="issued">Issue Date *</Label>
              <Input id="issued" type="date" {...register("issued")} aria-invalid={!!errors.issued} />
              {errors.issued && <p className="text-sm text-destructive">{errors.issued.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="expiry">Expiry Date *</Label>
              <Input id="expiry" type="date" {...register("expiry")} aria-invalid={!!errors.expiry} />
              {errors.expiry && <p className="text-sm text-destructive">{errors.expiry.message}</p>}
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="status">Status *</Label>
            <Select value={statusVal} onValueChange={(val) => setValue("status", val as IdCardFormData["status"])}>
              <SelectTrigger id="status"><SelectValue /></SelectTrigger>
              <SelectContent>
                {idCardStatusOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>
      <div className="flex justify-end gap-3">
        {onCancel && <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>Cancel</Button>}
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</> : (isEditing ? "Update" : "Create")}
        </Button>
      </div>
    </form>
  )
}
