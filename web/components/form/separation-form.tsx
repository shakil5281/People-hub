"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { format } from "date-fns"
import { Loader2, UserXIcon, SearchIcon, AlertTriangleIcon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { DatePicker } from "@/components/ui/date-picker"
import {
  separationSchema, SeparationFormData,
  separationTypeOptions,
} from "../data/separation-data"
import { separationApi, employeeApi } from "@/lib/api"

interface EmployeeInfo {
  employee_id: string
  name_en: string
  name_bn?: string
  department?: { name: string }
  department_id?: string
  section_ref?: { name: string }
  designation_ref?: { name: string }
  line_ref?: { name: string }
  employee_type?: string
  status?: string
  joining_date?: string
  phone?: string
}

interface SeparationFormProps {
  initialData?: Partial<SeparationFormData>
  onSuccess: () => void
  onCancel?: () => void
  isEditing?: boolean
  separationId?: string
}

export function SeparationForm({ initialData, onSuccess, onCancel, isEditing = false, separationId }: SeparationFormProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [error, setError] = React.useState("")
  const [successMsg, setSuccessMsg] = React.useState("")
  const [empLookupLoading, setEmpLookupLoading] = React.useState(false)
  const [empNotFound, setEmpNotFound] = React.useState(false)
  const [empNotEligible, setEmpNotEligible] = React.useState("")
  const [emp, setEmp] = React.useState<EmployeeInfo | null>(null)

  const { handleSubmit, formState: { errors }, setValue, watch, register, reset } = useForm<SeparationFormData>({
    resolver: zodResolver(separationSchema),
    defaultValues: {
      employee: "",
      employee_id: "",
      department_id: "",
      type: "Resign",
      date: new Date().toISOString().split("T")[0],
      status: "Pending",
      reason: "",
      ...initialData,
    },
  })

  const empId = watch("employee_id")
  const empName = watch("employee")
  const typeVal = watch("type")
  const dateVal = watch("date")

  const handleEmployeeLookup = React.useCallback(async () => {
    const code = empId?.trim()
    if (!code) return
    setEmpLookupLoading(true)
    setEmpNotFound(false)
    setEmpNotEligible("")
    setError("")
    setEmp(null)
    setValue("employee", "")
    setValue("department_id", "")
    try {
      const { data } = await employeeApi.getByCode(code)
      const e = data as EmployeeInfo
      if (e.status !== "active" || (e.employee_type && e.employee_type.toLowerCase() !== "regular")) {
        setEmpNotEligible(`Employee is ${e.status || "unknown"} / ${e.employee_type || "unknown"}. Only active Regular employees can be separated.`)
        setEmp(e)
        return
      }
      setEmp(e)
      setValue("employee_id", e.employee_id)
      setValue("employee", e.name_en || e.name_bn || "")
      setValue("department_id", e.department_id || "")
    } catch {
      setEmpNotFound(true)
      setValue("employee", "")
      setValue("department_id", "")
      toast.error("Employee not found")
    } finally {
      setEmpLookupLoading(false)
    }
  }, [empId, setValue])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault()
      handleEmployeeLookup()
    }
  }

  const onSubmit = async (data: SeparationFormData) => {
    setIsSubmitting(true)
    setError("")
    setSuccessMsg("")
    try {
      if (isEditing && separationId) {
        await separationApi.update(separationId, data)
        toast.success("Separation updated")
      } else {
        const { data: res } = await separationApi.create(data)
        if (res.auto_processed) {
          setSuccessMsg(`Employee processed: type → ${res.new_employee_type}, status → ${res.employee_status}`)
          toast.success("Separation created and processed immediately")
        } else {
          toast.success("Separation created (pending — will process on separation date)")
        }
      }
      onSuccess()
    } catch (err: unknown) {
      let detail = "Failed to save separation"
      if (typeof err === "object" && err !== null && "response" in err) {
        const axiosErr = err as { response?: { data?: { error?: string } } }
        detail = axiosErr.response?.data?.error || detail
      }
      setError(detail)
      toast.error(detail)
    } finally {
      setIsSubmitting(false)
    }
  }

  const isLocked = !!empName && !isEditing
  const showDetails = !!empName && (isEditing || (emp && !empNotEligible))

  // On edit, auto-load employee info from initial employee_id
  React.useEffect(() => {
    if (isEditing && initialData?.employee_id && !emp) {
      employeeApi.getByCode(initialData.employee_id).then(({ data }) => {
        setEmp(data as EmployeeInfo)
      }).catch(() => {})
    }
  }, [isEditing, initialData?.employee_id])

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {error && (
        <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
      )}
      {successMsg && (
        <div className="rounded-md bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-800">{successMsg}</div>
      )}

      <Card>
        <CardHeader><CardTitle className="flex items-center gap-2"><SearchIcon className="h-5 w-5" /> Employee Lookup</CardTitle></CardHeader>
        <CardContent>
          <div className="flex items-end gap-3">
            <div className="flex-1 space-y-2">
              <Label htmlFor="emp_search">Employee ID / Punch Number</Label>
              <Input
                id="emp_search"
                value={empId || ""}
                onChange={(e) => { setValue("employee_id", e.target.value); setEmpNotFound(false); setEmpNotEligible(""); setEmp(null) }}
                onKeyDown={handleKeyDown}
                placeholder="Type employee code and press Enter"
                disabled={isEditing}
              />
              {empNotFound && <p className="text-sm text-destructive">Employee not found</p>}
              {empNotEligible && (
                <div className="flex items-start gap-2 rounded-md bg-amber-50 border border-amber-200 p-2 text-sm text-amber-800">
                  <AlertTriangleIcon className="h-4 w-4 mt-0.5 shrink-0" />
                  <span>{empNotEligible}</span>
                </div>
              )}
            </div>
            <Button type="button" onClick={handleEmployeeLookup} disabled={empLookupLoading || isEditing || !empId?.trim()}>
              {empLookupLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <SearchIcon className="h-4 w-4" />}
              Search
            </Button>
          </div>
        </CardContent>
      </Card>

      {showDetails && (
        <>
          <Card>
            <CardHeader><CardTitle className="flex items-center gap-2 text-base"><UserXIcon className="h-4 w-4" /> Employee Information</CardTitle></CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
                <div>
                  <Label className="text-xs text-muted-foreground">Name</Label>
                  <p className="font-medium text-sm mt-1">{empName}</p>
                </div>
                <div>
                  <Label className="text-xs text-muted-foreground">Department</Label>
                  <p className="font-medium text-sm mt-1">{emp?.department?.name || "-"}</p>
                </div>
                <div>
                  <Label className="text-xs text-muted-foreground">Designation</Label>
                  <p className="font-medium text-sm mt-1">{emp?.designation_ref?.name || "-"}</p>
                </div>
                <div>
                  <Label className="text-xs text-muted-foreground">Type</Label>
                  <Badge variant="secondary" className="mt-1">{emp?.employee_type || "-"}</Badge>
                </div>
                <div>
                  <Label className="text-xs text-muted-foreground">Status</Label>
                  <Badge variant={emp?.status === "active" ? "default" : "destructive"} className="mt-1">{emp?.status || "-"}</Badge>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle className="text-base">Separation Details</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-3">
                <div className="space-y-2">
                  <Label htmlFor="type">Separation Type *</Label>
                  <Select value={typeVal} onValueChange={(val) => setValue("type", val as SeparationFormData["type"])}>
                    <SelectTrigger id="type"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      {separationTypeOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="date">Separation Date</Label>
                  <DatePicker
                    value={dateVal ? new Date(dateVal) : undefined}
                    onChange={(date) => setValue("date", date ? format(date, "yyyy-MM-dd") : "")}
                    placeholder="Select separation date"
                  />
                  {errors.date && <p className="text-sm text-destructive">{errors.date.message}</p>}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="status">Status</Label>
                  <Input id="status" value="Pending" disabled className="text-muted-foreground" />
                  <p className="text-xs text-muted-foreground">Auto-set on create. Process via actions after save.</p>
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="reason">Reason</Label>
                <Input id="reason" {...register("reason")} placeholder="Reason for separation" />
              </div>
            </CardContent>
          </Card>
        </>
      )}

      <div className="flex justify-end gap-3">
        {onCancel && <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>Cancel</Button>}
        <Button type="submit" disabled={isSubmitting || !empName || (!isEditing && !!empNotEligible)}>
          {isSubmitting ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</> : (isEditing ? "Update" : "Create Separation")}
        </Button>
      </div>
    </form>
  )
}
