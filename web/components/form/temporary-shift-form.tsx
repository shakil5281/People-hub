"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { tempShiftSchema, TempShiftFormData, TempShift } from "../data/temporary-shift-data"
import { employeeApi, shiftApi } from "@/lib/api"

interface TempShiftFormProps {
  initialData?: Partial<TempShift>
  onSuccess: (data: TempShiftFormData) => void
  onCancel?: () => void
  isEditing?: boolean
  tempShiftId?: string
}

interface EmployeeOption {
  id: string
  name_en: string
  employee_id: string
}

interface ShiftOption {
  id: string
  name: string
}

export function TempShiftForm({ initialData, onSuccess, onCancel, isEditing = false, tempShiftId }: TempShiftFormProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [isLoading, setIsLoading] = React.useState(false)
  const [employees, setEmployees] = React.useState<EmployeeOption[]>([])
  const [shifts, setShifts] = React.useState<ShiftOption[]>([])

  const { register, handleSubmit, reset, formState: { errors } } = useForm({
    resolver: zodResolver(tempShiftSchema),
    defaultValues: {
      employee_id: initialData?.employee_id || "",
      shift_id: initialData?.shift_id || "",
      from_date: "",
      to_date: "",
      reason: "",
      status: "active",
    },
  })

  React.useEffect(() => {
    setIsLoading(true)
    Promise.all([
      employeeApi.list({ status: "active" }),
      shiftApi.list(),
    ]).then(([empRes, shiftRes]) => {
      const empList = Array.isArray(empRes.data) ? empRes.data : []
      const shiftList = Array.isArray(shiftRes.data) ? shiftRes.data : []
      setEmployees(empList)
      setShifts(shiftList)
    }).finally(() => setIsLoading(false))
  }, [])

  React.useEffect(() => {
    if (initialData && isEditing) {
      reset({
        employee_id: initialData.employee_id || "",
        shift_id: initialData.shift_id || "",
        from_date: initialData.date || "",
        to_date: initialData.date || "",
        reason: initialData.reason || "",
        status: initialData.status || "active",
      })
    }
  }, [initialData, isEditing, reset])

  const onSubmit = async (data: TempShiftFormData) => {
    setIsSubmitting(true)
    try {
      await onSuccess(data)
    } finally {
      setIsSubmitting(false)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="employee_id">Employee *</Label>
          <select
            id="employee_id"
            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            {...register("employee_id")}
          >
            <option value="">Select Employee</option>
            {employees.map((emp) => (
              <option key={emp.id} value={emp.id}>
                {emp.name_en} ({emp.employee_id})
              </option>
            ))}
          </select>
          {errors.employee_id && <p className="text-sm text-destructive">{errors.employee_id.message}</p>}
        </div>
        <div className="space-y-2">
          <Label htmlFor="shift_id">Assigned Shift *</Label>
          <select
            id="shift_id"
            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            {...register("shift_id")}
          >
            <option value="">Select Shift</option>
            {shifts.map((s) => (
              <option key={s.id} value={s.id}>{s.name}</option>
            ))}
          </select>
          {errors.shift_id && <p className="text-sm text-destructive">{errors.shift_id.message}</p>}
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="from_date">From Date *</Label>
          <Input id="from_date" type="date" {...register("from_date")} aria-invalid={!!errors.from_date} />
          {errors.from_date && <p className="text-sm text-destructive">{errors.from_date.message}</p>}
        </div>
        <div className="space-y-2">
          <Label htmlFor="to_date">To Date</Label>
          <Input id="to_date" type="date" {...register("to_date")} />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="reason">Reason *</Label>
        <Input id="reason" placeholder="Replacement, Overtime, Emergency..." {...register("reason")} aria-invalid={!!errors.reason} />
        {errors.reason && <p className="text-sm text-destructive">{errors.reason.message}</p>}
      </div>

      <div className="flex justify-end gap-4 pt-4 border-t">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>Cancel</Button>
        )}
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? (
            <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</>
          ) : (
            isEditing ? "Update" : "Create"
          )}
        </Button>
      </div>
    </form>
  )
}
