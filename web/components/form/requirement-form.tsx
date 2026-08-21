"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2, ClipboardListIcon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { requirementSchema, RequirementFormData, statusOptions, priorityOptions, groupTypeOptions } from "../data/requirement-data"
import { requirementApi, departmentApi, sectionApi, designationApi } from "@/lib/api"
import type { Department, Section, Designation } from "@/components/data/organization-data"

interface RequirementFormProps {
  initialData?: Partial<RequirementFormData>
  onSuccess: () => void
  onCancel?: () => void
  isEditing?: boolean
  requirementId?: string
}

export function RequirementForm({ initialData, onSuccess, onCancel, isEditing = false, requirementId }: RequirementFormProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [error, setError] = React.useState("")
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])

  const { handleSubmit, formState: { errors }, setValue, watch, register } = useForm<RequirementFormData>({
    resolver: zodResolver(requirementSchema),
    defaultValues: {
      department_id: "",
      section_id: "",
      designation_id: "",
      group_type: "Worker",
      vacancies: 1,
      applicants: 0,
      status: "Open",
      priority: "Medium",
      description: "",
      ...initialData,
    },
  })

  const departmentId = watch("department_id")
  const sectionId = watch("section_id")
  const groupType = watch("group_type")
  const statusVal = watch("status")
  const priorityVal = watch("priority")

  React.useEffect(() => {
    departmentApi.list({ limit: "200" }).then(({ data }) => {
      setDepartments(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
    }).catch(() => {})
  }, [])

  React.useEffect(() => {
    if (departmentId) {
      sectionApi.list(departmentId, { limit: "200" }).then(({ data }) => {
        setSections(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
      }).catch(() => setSections([]))
      if (!isEditing) { setValue("section_id", ""); setValue("designation_id", "") }
      setDesignations([])
    } else {
      setSections([])
      setDesignations([])
    }
  }, [departmentId, setValue, isEditing])

  React.useEffect(() => {
    if (sectionId) {
      designationApi.list(sectionId).then(({ data }) => {
        setDesignations(Array.isArray(data?.data) ? data.data : Array.isArray(data) ? data : [])
      }).catch(() => setDesignations([]))
      if (!isEditing) setValue("designation_id", "")
    } else {
      setDesignations([])
    }
  }, [sectionId, setValue, isEditing])

  const onSubmit = async (data: RequirementFormData) => {
    setIsSubmitting(true)
    setError("")
    try {
      if (isEditing && requirementId) {
        await requirementApi.update(requirementId, data)
        toast.success("Requirement updated successfully")
      } else {
        await requirementApi.create(data)
        toast.success("Requirement created successfully")
      }
      onSuccess()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to save requirement"
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
        <CardHeader><CardTitle className="flex items-center gap-2"><ClipboardListIcon className="h-5 w-5" /> Requirement Details</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="department_id">Department *</Label>
              <Select value={departmentId} onValueChange={(val) => setValue("department_id", val)}>
                <SelectTrigger id="department_id" className="w-full"><SelectValue placeholder="Select department" /></SelectTrigger>
                <SelectContent>
                  {departments.map((d) => <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>)}
                </SelectContent>
              </Select>
              {errors.department_id && <p className="text-sm text-destructive">{errors.department_id.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="section_id">Section *</Label>
              <Select value={sectionId} onValueChange={(val) => setValue("section_id", val)} disabled={!departmentId}>
                <SelectTrigger id="section_id" className="w-full"><SelectValue placeholder={departmentId ? "Select section" : "Select department first"} /></SelectTrigger>
                <SelectContent>
                  {sections.map((s) => <SelectItem key={s.id} value={s.id}>{s.name}</SelectItem>)}
                </SelectContent>
              </Select>
              {errors.section_id && <p className="text-sm text-destructive">{errors.section_id.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="designation_id">Designation *</Label>
              <Select value={watch("designation_id")} onValueChange={(val) => setValue("designation_id", val)} disabled={!sectionId}>
                <SelectTrigger id="designation_id" className="w-full"><SelectValue placeholder={sectionId ? "Select designation" : "Select section first"} /></SelectTrigger>
                <SelectContent>
                  {designations.map((d) => <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>)}
                </SelectContent>
              </Select>
              {errors.designation_id && <p className="text-sm text-destructive">{errors.designation_id.message}</p>}
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="group_type">Group *</Label>
              <Select value={groupType} onValueChange={(val) => setValue("group_type", val as "Staff" | "Worker")}>
                <SelectTrigger id="group_type" className="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  {groupTypeOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="status">Status *</Label>
              <Select value={statusVal} onValueChange={(val) => setValue("status", val as "Open" | "Closed")}>
                <SelectTrigger id="status" className="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  {statusOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="vacancies">Required *</Label>
              <Input id="vacancies" type="number" min={1} {...register("vacancies", { valueAsNumber: true })} aria-invalid={!!errors.vacancies} />
              {errors.vacancies && <p className="text-sm text-destructive">{errors.vacancies.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="applicants">Applicants</Label>
              <Input id="applicants" type="number" min={0} {...register("applicants", { valueAsNumber: true })} aria-invalid={!!errors.applicants} />
              {errors.applicants && <p className="text-sm text-destructive">{errors.applicants.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="priority">Priority *</Label>
              <Select value={priorityVal} onValueChange={(val) => setValue("priority", val as "High" | "Medium" | "Low")}>
                <SelectTrigger id="priority" className="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  {priorityOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-1">
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Input id="description" {...register("description")} placeholder="Additional notes..." />
            </div>
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
