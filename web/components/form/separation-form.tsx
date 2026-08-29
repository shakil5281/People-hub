"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { format, differenceInYears, differenceInMonths, differenceInDays, parseISO } from "date-fns"
import { Loader2, UserXIcon, SearchIcon, AlertTriangleIcon, CheckCircle2Icon, XCircleIcon, Building2Icon, BriefcaseIcon, CalendarIcon, PhoneIcon, LayersIcon } from "lucide-react"
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
  id?: string
  employee_id: string
  punch_number?: string
  name_en: string
  name_bn?: string
  department?: { name: string }
  department_id?: string
  section_ref?: { name: string }
  designation_ref?: { name: string }
  line_ref?: { name: string }
  group_ref?: { name: string }
  employee_type?: string
  status?: string
  joining_date?: string
  phone?: string
  gross_salary?: number
  image_url?: string
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
  const [searchedCode, setSearchedCode] = React.useState("")
  const [empNotEligible, setEmpNotEligible] = React.useState("")
  const [emp, setEmp] = React.useState<EmployeeInfo | null>(null)

  const { handleSubmit, formState: { errors }, setValue, watch, register } = useForm<SeparationFormData>({
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

  const handleEmployeeLookup = React.useCallback(async (overrideCode?: string) => {
    const code = (overrideCode ?? empId)?.trim()
    if (!code) {
      setEmp(null)
      setEmpNotFound(false)
      setEmpNotEligible("")
      setSearchedCode("")
      setValue("employee", "")
      setValue("department_id", "")
      return
    }

    setEmpLookupLoading(true)
    setEmpNotFound(false)
    setEmpNotEligible("")
    setError("")
    setSearchedCode(code)

    try {
      const { data } = await employeeApi.getByCode(code)
      const e = data as EmployeeInfo

      // Verify exact match on employee_id or punch_number
      const match = e && (
        String(e.employee_id).trim().toLowerCase() === code.toLowerCase() ||
        (e.punch_number && String(e.punch_number).trim().toLowerCase() === code.toLowerCase())
      )

      if (!match) {
        setEmp(null)
        setEmpNotFound(true)
        setValue("employee", "")
        setValue("department_id", "")
        return
      }

      setEmp(e)
      setEmpNotFound(false)
      setValue("employee_id", e.employee_id)
      setValue("employee", e.name_en || e.name_bn || "")
      setValue("department_id", e.department_id || "")

      if (e.status && e.status.toLowerCase() !== "active") {
        setEmpNotEligible(`Notice: Employee status is currently "${e.status}".`)
      }
    } catch {
      setEmp(null)
      setEmpNotFound(true)
      setValue("employee", "")
      setValue("department_id", "")
    } finally {
      setEmpLookupLoading(false)
    }
  }, [empId, setValue])

  // Debounced auto-search when user types employee ID
  React.useEffect(() => {
    if (isEditing) return
    const code = empId?.trim()
    if (!code) {
      setEmp(null)
      setEmpNotFound(false)
      setSearchedCode("")
      return
    }
    const timer = setTimeout(() => {
      handleEmployeeLookup(code)
    }, 400)
    return () => clearTimeout(timer)
  }, [empId, isEditing, handleEmployeeLookup])

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

  // On edit, auto-load employee info from initial employee_id
  React.useEffect(() => {
    if (isEditing && initialData?.employee_id && !emp) {
      employeeApi.getByCode(initialData.employee_id).then(({ data }) => {
        setEmp(data as EmployeeInfo)
      }).catch(() => {})
    }
  }, [isEditing, initialData?.employee_id, emp])

  const calcTenure = (joiningDate?: string) => {
    if (!joiningDate) return "-"
    try {
      const join = parseISO(joiningDate)
      const now = new Date()
      const y = differenceInYears(now, join)
      const m = differenceInMonths(now, join) % 12
      const d = differenceInDays(now, join) % 30
      const parts = []
      if (y > 0) parts.push(`${y}y`)
      if (m > 0) parts.push(`${m}m`)
      if (d > 0 && y === 0) parts.push(`${d}d`)
      return parts.join(" ") || "0 days"
    } catch {
      return "-"
    }
  }

  const showDetails = !!emp && !empNotFound

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {error && (
        <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
      )}
      {successMsg && (
        <div className="rounded-md bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-800">{successMsg}</div>
      )}

      {/* Employee Search Card */}
      <Card className="border border-border/80 shadow-xs">
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center justify-between text-base">
            <span className="flex items-center gap-2">
              <SearchIcon className="h-4 w-4 text-primary" />
              Search Employee
            </span>
            {showDetails && (
              <Badge variant="outline" className="text-xs bg-emerald-500/10 text-emerald-600 border-emerald-500/30">
                <CheckCircle2Icon className="mr-1 h-3 w-3" />
                Employee Found
              </Badge>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-end gap-3">
            <div className="flex-1 space-y-1.5">
              <Label htmlFor="emp_search" className="text-xs font-medium">Search by Employee ID / Code</Label>
              <div className="relative">
                <SearchIcon className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  id="emp_search"
                  value={empId || ""}
                  onChange={(e) => {
                    setValue("employee_id", e.target.value)
                    setEmpNotFound(false)
                    setEmpNotEligible("")
                    if (!e.target.value.trim()) {
                      setEmp(null)
                      setValue("employee", "")
                      setValue("department_id", "")
                    }
                  }}
                  onKeyDown={handleKeyDown}
                  placeholder="Enter Employee ID (e.g. 1024 or 1695)..."
                  className="pl-9 h-10 font-mono"
                  disabled={isEditing}
                  autoFocus={!isEditing}
                />
              </div>
            </div>
            <Button
              type="button"
              onClick={() => handleEmployeeLookup()}
              disabled={empLookupLoading || isEditing || !empId?.trim()}
              className="h-10 px-5"
            >
              {empLookupLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <SearchIcon className="h-4 w-4 mr-1" />}
              Search
            </Button>
          </div>

          {/* Not Found State Alert */}
          {empNotFound && searchedCode && (
            <div className="rounded-lg border border-destructive/30 bg-destructive/10 p-5 text-center mt-4 animate-in fade-in-50">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-destructive/20 text-destructive mb-2.5">
                <XCircleIcon className="h-6 w-6" />
              </div>
              <h3 className="text-sm font-semibold text-destructive uppercase tracking-wide">Do not data exist</h3>
              <p className="text-xs text-muted-foreground mt-1">
                No matching employee record found with ID: <strong className="text-foreground font-mono">&quot;{searchedCode}&quot;</strong>
              </p>
              <p className="text-[11px] text-muted-foreground mt-0.5">Please check the employee code and search again.</p>
            </div>
          )}

          {empNotEligible && (
            <div className="flex items-start gap-2 rounded-md bg-amber-500/10 border border-amber-500/30 p-3 text-xs text-amber-800 dark:text-amber-300 mt-3">
              <AlertTriangleIcon className="h-4 w-4 mt-0.5 shrink-0" />
              <span>{empNotEligible}</span>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Employee Profile Information Card */}
      {showDetails && (
        <>
          <Card className="border border-primary/20 bg-card/60 shadow-xs overflow-hidden">
            <div className="bg-primary/5 px-5 py-3 border-b border-primary/10 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/15 text-primary font-bold text-sm">
                  {emp.name_en ? emp.name_en.charAt(0).toUpperCase() : <UserXIcon className="h-5 w-5" />}
                </div>
                <div>
                  <h3 className="font-semibold text-sm leading-tight text-foreground">{emp.name_en || emp.name_bn || "-"}</h3>
                  <div className="flex items-center gap-2 mt-0.5">
                    <span className="text-xs font-mono text-muted-foreground">ID: {emp.employee_id}</span>
                    {emp.punch_number && <span className="text-xs font-mono text-muted-foreground">| Punch: {emp.punch_number}</span>}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-1.5">
                <Badge variant={emp.status?.toLowerCase() === "active" ? "default" : "destructive"} className="text-xs capitalize">
                  {emp.status || "Unknown"}
                </Badge>
                {emp.employee_type && (
                  <Badge variant="outline" className="text-xs">
                    {emp.employee_type}
                  </Badge>
                )}
              </div>
            </div>

            <CardContent className="p-5">
              <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
                <div className="space-y-1">
                  <div className="flex items-center gap-1 text-xs text-muted-foreground">
                    <Building2Icon className="h-3.5 w-3.5" />
                    <span>Department</span>
                  </div>
                  <p className="text-sm font-medium">{emp.department?.name || "-"}</p>
                </div>

                <div className="space-y-1">
                  <div className="flex items-center gap-1 text-xs text-muted-foreground">
                    <LayersIcon className="h-3.5 w-3.5" />
                    <span>Section</span>
                  </div>
                  <p className="text-sm font-medium">{emp.section_ref?.name || "-"}</p>
                </div>

                <div className="space-y-1">
                  <div className="flex items-center gap-1 text-xs text-muted-foreground">
                    <BriefcaseIcon className="h-3.5 w-3.5" />
                    <span>Designation</span>
                  </div>
                  <p className="text-sm font-medium">{emp.designation_ref?.name || "-"}</p>
                </div>

                <div className="space-y-1">
                  <div className="flex items-center gap-1 text-xs text-muted-foreground">
                    <LayersIcon className="h-3.5 w-3.5" />
                    <span>Line</span>
                  </div>
                  <p className="text-sm font-medium">{emp.line_ref?.name || "-"}</p>
                </div>

                <div className="space-y-1">
                  <div className="flex items-center gap-1 text-xs text-muted-foreground">
                    <CalendarIcon className="h-3.5 w-3.5" />
                    <span>Joining Date</span>
                  </div>
                  <p className="text-sm font-medium">
                    {emp.joining_date ? parseISO(emp.joining_date).toLocaleDateString("en-GB", { day: "2-digit", month: "short", year: "numeric" }) : "-"}
                  </p>
                </div>

                <div className="space-y-1">
                  <div className="flex items-center gap-1 text-xs text-muted-foreground">
                    <CalendarIcon className="h-3.5 w-3.5" />
                    <span>Job Age (Tenure)</span>
                  </div>
                  <p className="text-sm font-medium text-primary">{calcTenure(emp.joining_date)}</p>
                </div>

                {emp.phone && (
                  <div className="space-y-1">
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
                      <PhoneIcon className="h-3.5 w-3.5" />
                      <span>Phone</span>
                    </div>
                    <p className="text-sm font-medium">{emp.phone}</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Separation Details Form Card */}
          <Card className="border border-border/80 shadow-xs">
            <CardHeader className="pb-3">
              <CardTitle className="text-base flex items-center gap-2">
                <UserXIcon className="h-4 w-4 text-destructive" />
                Separation Details
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-3">
                <div className="space-y-1.5">
                  <Label htmlFor="type" className="text-xs font-medium">Separation Type *</Label>
                  <Select value={typeVal} onValueChange={(val) => setValue("type", val as SeparationFormData["type"])}>
                    <SelectTrigger id="type" className="h-9"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      {separationTypeOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="date" className="text-xs font-medium">Separation Date *</Label>
                  <DatePicker
                    value={dateVal ? new Date(dateVal) : undefined}
                    onChange={(date) => setValue("date", date ? format(date, "yyyy-MM-dd") : "")}
                    placeholder="Select separation date"
                  />
                  {errors.date && <p className="text-xs text-destructive">{errors.date.message}</p>}
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="status" className="text-xs font-medium">Status</Label>
                  <Input id="status" value="Pending" disabled className="h-9 text-muted-foreground bg-muted/40 text-xs" />
                  <p className="text-[11px] text-muted-foreground">Auto-set on create. Processed on separation date.</p>
                </div>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="reason" className="text-xs font-medium">Reason for Separation</Label>
                <Input id="reason" {...register("reason")} placeholder="Enter reason (optional)" className="h-9 text-sm" />
              </div>
            </CardContent>
          </Card>
        </>
      )}

      <div className="flex justify-end gap-3 pt-2">
        {onCancel && <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>Cancel</Button>}
        <Button type="submit" disabled={isSubmitting || !showDetails} className="min-w-[140px]">
          {isSubmitting ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</> : (isEditing ? "Update Separation" : "Create Separation")}
        </Button>
      </div>
    </form>
  )
}

