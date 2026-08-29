"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2 } from "lucide-react"
import { format } from "date-fns"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { DatePicker } from "@/components/ui/date-picker"
import { bulkRosterSchema, BulkRosterFormData } from "../data/roster-data"
import { shiftApi, departmentApi, sectionApi, lineApi, groupApi } from "@/lib/api"

interface BulkFormProps {
  onSuccess: (data: BulkRosterFormData & { company_id: string }) => void
  onCancel?: () => void
  companyId: string
}

export function RosterBulkForm({ onSuccess, onCancel, companyId }: BulkFormProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [shifts, setShifts] = React.useState<{ id: string; name: string }[]>([])
  const [departments, setDepartments] = React.useState<{ id: string; name: string }[]>([])
  const [sections, setSections] = React.useState<{ id: string; name: string }[]>([])
  const [lines, setLines] = React.useState<{ id: string; name: string }[]>([])
  const [groups, setGroups] = React.useState<{ id: string; name: string }[]>([])

  const { register, handleSubmit, watch, setValue, formState: { errors } } = useForm({
    resolver: zodResolver(bulkRosterSchema),
    defaultValues: {
      shift_id: "",
      from_date: "",
      to_date: "",
      reason: "",
      status: "active",
      employee_ids: [],
      department_id: "",
      section_id: "",
      line_id: "",
      group_id: "",
    },
  })

  const fromDateVal = watch("from_date")
  const toDateVal = watch("to_date")

  React.useEffect(() => {
    Promise.all([
      shiftApi.list(),
      departmentApi.list(),
      groupApi.list(),
    ]).then(([shiftRes, deptRes, groupRes]) => {
      const sd = shiftRes.data as unknown as { data?: {id:string;name:string}[] }
      setShifts(Array.isArray(sd?.data) ? sd.data : (Array.isArray(shiftRes.data) ? shiftRes.data as unknown as {id:string;name:string}[] : []))
      const dd = deptRes.data as unknown as { data?: {id:string;name:string}[] }
      setDepartments(Array.isArray(dd?.data) ? dd.data : (Array.isArray(deptRes.data) ? deptRes.data as unknown as {id:string;name:string}[] : []))
      const gd = groupRes.data as unknown as { data?: {id:string;name:string}[] }
      setGroups(Array.isArray(gd?.data) ? gd.data : (Array.isArray(groupRes.data) ? groupRes.data as unknown as {id:string;name:string}[] : []))
    })
  }, [])

  React.useEffect(() => {
    const deptId = watch("department_id")
    if (!deptId) { setSections([]); return }
    sectionApi.list(deptId).then((res) => {
      const sd = res.data as unknown as { data?: {id:string;name:string}[] }
      setSections(Array.isArray(sd?.data) ? sd.data : (Array.isArray(res.data) ? res.data as unknown as {id:string;name:string}[] : []))
    })
  }, [watch("department_id")])

  React.useEffect(() => {
    const secId = watch("section_id")
    if (!secId) { setLines([]); return }
    lineApi.list(secId).then((res) => {
      const ld = res.data as unknown as { data?: {id:string;name:string}[] }
      setLines(Array.isArray(ld?.data) ? ld.data : (Array.isArray(res.data) ? res.data as unknown as {id:string;name:string}[] : []))
    })
  }, [watch("section_id")])

  const onSubmit = async (data: BulkRosterFormData) => {
    setIsSubmitting(true)
    try {
      await onSuccess({ ...data, company_id: companyId })
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="space-y-2">
        <Label>Shift *</Label>
        <select
          className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
          {...register("shift_id")}
        >
          <option value="">Select Shift</option>
          {shifts.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
        </select>
        {errors.shift_id && <p className="text-sm text-destructive">{errors.shift_id.message}</p>}
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label>From Date *</Label>
          <DatePicker
            value={fromDateVal ? new Date(fromDateVal + "T00:00:00") : undefined}
            onChange={(date) => setValue("from_date", date ? format(date, "yyyy-MM-dd") : "", { shouldValidate: true })}
            placeholder="Select from date"
          />
          {errors.from_date && <p className="text-sm text-destructive">{errors.from_date.message}</p>}
        </div>
        <div className="space-y-2">
          <Label>To Date</Label>
          <DatePicker
            value={toDateVal ? new Date(toDateVal + "T00:00:00") : undefined}
            onChange={(date) => setValue("to_date", date ? format(date, "yyyy-MM-dd") : "")}
            placeholder="Select to date"
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Reason</Label>
        <Input placeholder="Bulk roster reason" {...register("reason")} />
      </div>

      <div className="rounded-md border p-3 space-y-3 bg-muted/30">
        <p className="text-sm font-medium">Assign to (leave empty to require manual IDs, or select org filter)</p>
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label>Department</Label>
            <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...register("department_id")}>
              <option value="">All Departments</option>
              {departments.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
            </select>
          </div>
          <div className="space-y-2">
            <Label>Section</Label>
            <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...register("section_id")}>
              <option value="">All Sections</option>
              {sections.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
            </select>
          </div>
          <div className="space-y-2">
            <Label>Line</Label>
            <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...register("line_id")}>
              <option value="">All Lines</option>
              {lines.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
            </select>
          </div>
          <div className="space-y-2">
            <Label>Group</Label>
            <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...register("group_id")}>
              <option value="">All Groups</option>
              {groups.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
            </select>
          </div>
        </div>
        <p className="text-xs text-muted-foreground">If no org filter selected, bulk will require explicit employee_ids (add via search in next step). Current implementation uses department/section/line/group filter to resolve employees automatically.</p>
      </div>

      <div className="flex justify-end gap-4 pt-4 border-t">
        {onCancel && <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>Cancel</Button>}
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</> : "Bulk Assign"}
        </Button>
      </div>
    </form>
  )
}
