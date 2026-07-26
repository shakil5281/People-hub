"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { format } from "date-fns"
import {
  Loader2, UserCircle, Languages, BriefcaseIcon, CalendarDays, UsersRound, Droplet, Heart, Globe, CreditCard,
  Phone, Mail, MapPin, AlertTriangle, PhoneCall, Users, Building2, Badge, Fingerprint,
  Award, CalendarClock, Clock, Layers, Star, Minus, Folder, LayoutGrid, UserCog, ToggleLeft,
  Banknote, Home, Stethoscope, Car, UtensilsCrossed, PlusCircle,
  Calculator, Smartphone, Hash, Camera, PenTool, MapPinned
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { toast } from "sonner"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ImageUpload } from "@/components/image-upload"
import { DatePicker } from "@/components/ui/date-picker"
import { employeeSchema, EmployeeFormData, genderOptions, bloodGroupOptions, maritalStatusOptions, statusOptionsEmployee, religionOptions, employeeTypeOptions } from "../data/employee-data"
import { employeeApi, companyApi, shiftApi, departmentApi, sectionApi, designationApi, lineApi, groupApi, floorApi, divisionApi, districtApi, upazilaApi, unionApi } from "@/lib/api"
import type { Company } from "@/components/data/company-data"
import type { Shift } from "@/components/data/shift-data"
import type { Department } from "@/components/data/organization-data"
import type { Section, Designation, Line } from "@/components/data/organization-data"
import type { Division, District, Upazila, Union } from "@/components/data/address-data"

interface NamedItem { id: string; name: string }

interface EmployeeFormProps {
  initialData?: Partial<EmployeeFormData>
  onSuccess: () => void
  onCancel?: () => void
  isEditing?: boolean
  employeeId?: string
}

export function EmployeeForm({ initialData, onSuccess, onCancel, isEditing = false, employeeId }: EmployeeFormProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [error, setError] = React.useState("")
  const [companies, setCompanies] = React.useState<Company[]>([])
  const [shifts, setShifts] = React.useState<Shift[]>([])
  const [departments, setDepartments] = React.useState<Department[]>([])
  const [sections, setSections] = React.useState<Section[]>([])
  const [designations, setDesignations] = React.useState<Designation[]>([])
  const [lines, setLines] = React.useState<Line[]>([])
  const [groups, setGroups] = React.useState<NamedItem[]>([])
  const [floors, setFloors] = React.useState<NamedItem[]>([])
  // Present address
  const [presentDivisions, setPresentDivisions] = React.useState<Division[]>([])
  const [presentDistricts, setPresentDistricts] = React.useState<District[]>([])
  const [presentUpazilas, setUpPresentUpazilas] = React.useState<Upazila[]>([])
  const [presentUnions, setPresentUnions] = React.useState<Union[]>([])
  // Permanent address
  const [permDivisions, setPermDivisions] = React.useState<Division[]>([])
  const [permDistricts, setPermDistricts] = React.useState<District[]>([])
  const [permUpazilas, setPermUpazilas] = React.useState<Upazila[]>([])
  const [permUnions, setPermUnions] = React.useState<Union[]>([])

  const { register, handleSubmit, watch, setValue, formState: { errors } } = useForm<EmployeeFormData>({
    resolver: zodResolver(employeeSchema),
    defaultValues: {
      company_id: "",
      employee_id: "",
      punch_number: "",
      employee_type: "Regular",
      joining_date: "",
      shift_id: null,
      status: "active",
      name_en: "", name_bn: "", father_name: "", mother_name: "", date_of_birth: "",
      gender: "", blood_group: "", marital_status: "", religion: "", nationality: "Bangladeshi",
      nid: "",
      phone: "", email: "", present_address: "", permanent_address: "",
      spouse_name: "", emergency_contact: "", emergency_phone: "", number_of_dependents: 0,
      grade: "",
      department_id: null, section_id: null, designation_id: null, line_id: null,
      group_id: null, floor_id: null,
      reports_to: null,
      present_division_id: null, present_district_id: null, present_upazila_id: null, present_union_id: null,
      permanent_division_id: null, permanent_district_id: null, permanent_upazila_id: null, permanent_union_id: null,
      gross_salary: 0, basic_salary: 0, house_rent: 0,
      transport_allowance: 450, food_allowance: 1250,
      medical_allowance: 750, other_allowance: 0,
      account_type: "", account_number: "",
      image_url: "", signature_url: "",
      ...initialData,
    },
  })

  const watchDept = watch("department_id")
  const watchSection = watch("section_id")
  const watchPresentDiv = watch("present_division_id")
  const watchPresentDist = watch("present_district_id")
  const watchPresentUpa = watch("present_upazila_id")
  const watchPermDiv = watch("permanent_division_id")
  const watchPermDist = watch("permanent_district_id")
  const watchPermUpa = watch("permanent_upazila_id")
  const watchDesig = watch("designation_id")
  const watchLine = watch("line_id")
  const watchGroup = watch("group_id")
  const watchFloor = watch("floor_id")
  const watchGrossSalary = watch("gross_salary")

  // Fetch static lists
  React.useEffect(() => {
    companyApi.list().then(({ data }) => setCompanies(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
    shiftApi.list().then(({ data }) => setShifts(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
    departmentApi.list().then(({ data }) => setDepartments(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
    groupApi.list().then(({ data }) => setGroups(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
    floorApi.list().then(({ data }) => setFloors(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
    divisionApi.list().then(({ data }) => { const d = Array.isArray(data.data) ? data.data : []; setPresentDivisions(d); setPermDivisions(d) }).catch((err) => console.error("Failed to fetch reference data", err))
  }, [])

  // Track previous values to avoid clearing on initial mount
  const prevDept = React.useRef(watchDept)
  const prevSection = React.useRef(watchSection)
  const prevPresentDiv = React.useRef(watchPresentDiv)
  const prevPresentDist = React.useRef(watchPresentDist)
  const prevPresentUpa = React.useRef(watchPresentUpa)
  const prevPermDiv = React.useRef(watchPermDiv)
  const prevPermDist = React.useRef(watchPermDist)
  const prevPermUpa = React.useRef(watchPermUpa)

  // Cascading: department → section → designation + line
  React.useEffect(() => {
    const changed = watchDept !== prevDept.current
    prevDept.current = watchDept
    if (watchDept) {
      sectionApi.list(watchDept).then(({ data }) => setSections(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
      if (changed) {
        setValue("section_id", null)
        setValue("designation_id", null)
        setValue("line_id", null)
      }
    } else {
      setSections([]); setDesignations([]); setLines([])
    }
  }, [watchDept])

  React.useEffect(() => {
    const changed = watchSection !== prevSection.current
    prevSection.current = watchSection
    if (watchSection) {
      designationApi.list(watchSection).then(({ data }) => setDesignations(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
      lineApi.list(watchSection).then(({ data }) => setLines(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
      if (changed) {
        setValue("designation_id", null)
        setValue("line_id", null)
      }
    } else {
      setDesignations([]); setLines([])
    }
  }, [watchSection])

  // Cascading: present address
  React.useEffect(() => {
    const changed = watchPresentDiv !== prevPresentDiv.current
    prevPresentDiv.current = watchPresentDiv
    if (watchPresentDiv) {
      districtApi.list(watchPresentDiv).then(({ data }) => setPresentDistricts(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
      if (changed) {
        setValue("present_district_id", null); setValue("present_upazila_id", null); setValue("present_union_id", null)
      }
    } else { setPresentDistricts([]); setUpPresentUpazilas([]); setPresentUnions([]) }
  }, [watchPresentDiv])
  React.useEffect(() => {
    const changed = watchPresentDist !== prevPresentDist.current
    prevPresentDist.current = watchPresentDist
    if (watchPresentDist) {
      upazilaApi.list(watchPresentDist).then(({ data }) => setUpPresentUpazilas(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
      if (changed) {
        setValue("present_upazila_id", null); setValue("present_union_id", null)
      }
    } else { setUpPresentUpazilas([]); setPresentUnions([]) }
  }, [watchPresentDist])
  React.useEffect(() => {
    const changed = watchPresentUpa !== prevPresentUpa.current
    prevPresentUpa.current = watchPresentUpa
    if (watchPresentUpa) {
      unionApi.list(watchPresentUpa).then(({ data }) => setPresentUnions(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
      if (changed) {
        setValue("present_union_id", null)
      }
    } else { setPresentUnions([]) }
  }, [watchPresentUpa])

  // Cascading: permanent address
  React.useEffect(() => {
    const changed = watchPermDiv !== prevPermDiv.current
    prevPermDiv.current = watchPermDiv
    if (watchPermDiv) {
      districtApi.list(watchPermDiv).then(({ data }) => setPermDistricts(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
      if (changed) {
        setValue("permanent_district_id", null); setValue("permanent_upazila_id", null); setValue("permanent_union_id", null)
      }
    } else { setPermDistricts([]); setPermUpazilas([]); setPermUnions([]) }
  }, [watchPermDiv])
  React.useEffect(() => {
    const changed = watchPermDist !== prevPermDist.current
    prevPermDist.current = watchPermDist
    if (watchPermDist) {
      upazilaApi.list(watchPermDist).then(({ data }) => setPermUpazilas(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
      if (changed) {
        setValue("permanent_upazila_id", null); setValue("permanent_union_id", null)
      }
    } else { setPermUpazilas([]); setPermUnions([]) }
  }, [watchPermDist])
  React.useEffect(() => {
    const changed = watchPermUpa !== prevPermUpa.current
    prevPermUpa.current = watchPermUpa
    if (watchPermUpa) {
      unionApi.list(watchPermUpa).then(({ data }) => setPermUnions(Array.isArray(data.data) ? data.data : [])).catch((err) => console.error("Failed to fetch reference data", err))
      if (changed) {
        setValue("permanent_union_id", null)
      }
    } else { setPermUnions([]) }
  }, [watchPermUpa])

  // Auto-calculate salary from gross
  // Backend formula:
  //   core       = Gross - 450(Transport) - 1250(Food) - 750(Medical)
  //   Basic      = core / 1.5
  //   House Rent = core - Basic
  const prevGross = React.useRef(watchGrossSalary)
  React.useEffect(() => {
    const changed = watchGrossSalary !== prevGross.current
    prevGross.current = watchGrossSalary
    if ((changed || (watchGrossSalary ?? 0) > 0) && (watchGrossSalary ?? 0) > 0) {
      const gross = watchGrossSalary ?? 0
      const transport = 450
      const food = 1250
      const medical = 750
      const core = gross - transport - food - medical
      const basic = Math.round(core / 1.5)
      const houseRent = Math.round(core - basic)
      setValue("transport_allowance", transport)
      setValue("food_allowance", food)
      setValue("medical_allowance", medical)
      setValue("basic_salary", basic)
      setValue("house_rent", houseRent)
    }
  }, [watchGrossSalary])

  const onSubmit = async (data: EmployeeFormData) => {
    setIsSubmitting(true)
    setError("")
    try {
      const payload = {
        ...data,
        shift_id: data.shift_id || null,
        department_id: data.department_id || null,
        section_id: data.section_id || null,
        designation_id: data.designation_id || null,
        line_id: data.line_id || null,
        group_id: data.group_id || null,
        floor_id: data.floor_id || null,
        reports_to: data.reports_to || null,
        present_division_id: data.present_division_id || null,
        present_district_id: data.present_district_id || null,
        present_upazila_id: data.present_upazila_id || null,
        present_union_id: data.present_union_id || null,
        permanent_division_id: data.permanent_division_id || null,
        permanent_district_id: data.permanent_district_id || null,
        permanent_upazila_id: data.permanent_upazila_id || null,
        permanent_union_id: data.permanent_union_id || null,
      }
      if (isEditing && employeeId) {
        await employeeApi.update(employeeId, payload)
      } else {
        await employeeApi.create(payload)
      }
      onSuccess()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to save employee"
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
      {error && <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>}

      {/* Personal Information */}
      <Card>
        <CardHeader><CardTitle className="flex items-center text-neutral-600 gap-2 text-lg">Personal Information</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="name_en" className="flex items-center text-red-500 gap-1.5 px-2"><UserCircle className="h-3.5 w-3.5" /> Name (English) *</Label>
              <Input size='md' id="name_en" placeholder="Full name" {...register("name_en")} aria-invalid={!!errors.name_en} />
              {errors.name_en && <p className="text-sm text-destructive">{errors.name_en.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="name_bn" className="flex items-center text-neutral-600 gap-1.5 px-2"><Languages className="h-3.5 w-3.5" /> Name (Bangla)</Label>
              <Input size='md' id="name_bn" className="bangla-input" placeholder="cy‡iv bvg" {...register("name_bn")} />
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="father_name" className="flex items-center text-neutral-600 gap-1.5 px-2"><UserCircle className="h-3.5 w-3.5" /> Father&apos;s Name</Label>
              <Input size='md' id="father_name" placeholder="Father's name" {...register("father_name")} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="mother_name" className="flex items-center text-neutral-600 gap-1.5 px-2"><UserCircle className="h-3.5 w-3.5" /> Mother&apos;s Name</Label>
              <Input size='md' id="mother_name" placeholder="Mother's name" {...register("mother_name")} />
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="date_of_birth" className="flex items-center text-red-500 gap-1.5 px-2"><CalendarDays className="h-3.5 w-3.5" /> Date of Birth</Label>
              <DatePicker
                value={watch("date_of_birth") ? new Date(watch("date_of_birth")!) : undefined}
                onChange={(date) => setValue("date_of_birth", date ? format(date, "yyyy-MM-dd") : "")}
                placeholder="Select date of birth"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="gender" className="flex items-center text-neutral-600 gap-1.5 px-2"><UsersRound className="h-3.5 w-3.5" /> Gender</Label>
              <select id="gender" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...register("gender")}>
                <option value="">Select</option>
                {genderOptions.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="blood_group" className="flex items-center text-neutral-600 gap-1.5 px-2"><Droplet className="h-3.5 w-3.5" /> Blood Group</Label>
              <select id="blood_group" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...register("blood_group")}>
                <option value="">Select</option>
                {bloodGroupOptions.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-4">
            <div className="space-y-2">
              <Label htmlFor="marital_status" className="flex items-center text-neutral-600 gap-1.5 px-2"><Heart className="h-3.5 w-3.5" /> Marital Status</Label>
              <select id="marital_status" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...register("marital_status")}>
                <option value="">Select</option>
                {maritalStatusOptions.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="religion" className="flex items-center text-neutral-600 gap-1.5 px-2"><Globe className="h-3.5 w-3.5" /> Religion</Label>
              <select id="religion" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...register("religion")}>
                <option value="">Select</option>
                {religionOptions.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="nationality" className="flex items-center text-neutral-600 gap-1.5 px-2"><Globe className="h-3.5 w-3.5" /> Nationality</Label>
              <Input size='md' id="nationality" placeholder="Bangladeshi" {...register("nationality")} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="nid" className="flex items-center text-neutral-600 gap-1.5 px-2"><CreditCard className="h-3.5 w-3.5" /> NID Number</Label>
              <Input size='md' id="nid" placeholder="National ID" {...register("nid")} />
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="phone" className="flex items-center text-neutral-600 gap-1.5 px-2"><Phone className="h-3.5 w-3.5" /> Phone</Label>
              <Input size='md' id="phone" placeholder="01XXXXXXXXX" {...register("phone")} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email" className="flex items-center text-neutral-600 gap-1.5 px-2"><Mail className="h-3.5 w-3.5" /> Email</Label>
              <Input size='md' id="email" type="email" placeholder="email@example.com" {...register("email")} />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Present Address */}
      <Card>
        <CardHeader><CardTitle className="flex items-center text-neutral-600 gap-2 text-lg"><MapPinned className="h-4 w-4" /> Present Address</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-4">
            <div className="space-y-2">
              <Label>Division</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watchPresentDiv || ""} onChange={(e) => setValue("present_division_id", e.target.value || null)}>
                <option value="">Select</option>
                {presentDivisions.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label>District</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watchPresentDist || ""} onChange={(e) => setValue("present_district_id", e.target.value || null)}>
                <option value="">Select</option>
                {presentDistricts.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label>Upazila</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watchPresentUpa || ""} onChange={(e) => setValue("present_upazila_id", e.target.value || null)}>
                <option value="">Select</option>
                {presentUpazilas.map((u) => <option key={u.id} value={u.id}>{u.name}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label>Union</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watch("present_union_id") || ""} onChange={(e) => setValue("present_union_id", e.target.value || null)}>
                <option value="">Select</option>
                {presentUnions.map((u) => <option key={u.id} value={u.id}>{u.name}</option>)}
              </select>
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="present_address" className="flex items-center text-neutral-600 gap-1.5 px-2"><MapPin className="h-3.5 w-3.5" /> Address Details</Label>
            <Input size='md' id="present_address" placeholder="House, road, village etc." {...register("present_address")} />
          </div>
        </CardContent>
      </Card>

      {/* Permanent Address */}
      <Card>
        <CardHeader><CardTitle className="flex items-center text-neutral-600 gap-2 text-lg"><MapPinned className="h-4 w-4" /> Permanent Address</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-4">
            <div className="space-y-2">
              <Label>Division</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watchPermDiv || ""} onChange={(e) => setValue("permanent_division_id", e.target.value || null)}>
                <option value="">Select</option>
                {permDivisions.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label>District</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watchPermDist || ""} onChange={(e) => setValue("permanent_district_id", e.target.value || null)}>
                <option value="">Select</option>
                {permDistricts.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label>Upazila</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watchPermUpa || ""} onChange={(e) => setValue("permanent_upazila_id", e.target.value || null)}>
                <option value="">Select</option>
                {permUpazilas.map((u) => <option key={u.id} value={u.id}>{u.name}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label>Union</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watch("permanent_union_id") || ""} onChange={(e) => setValue("permanent_union_id", e.target.value || null)}>
                <option value="">Select</option>
                {permUnions.map((u) => <option key={u.id} value={u.id}>{u.name}</option>)}
              </select>
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="permanent_address" className="flex items-center text-neutral-600 gap-1.5 px-2"><MapPin className="h-3.5 w-3.5" /> Address Details</Label>
            <Input size='md' id="permanent_address" placeholder="House, road, village etc." {...register("permanent_address")} />
          </div>
        </CardContent>
      </Card>

      {/* Family & Emergency */}
      <Card>
        <CardHeader><CardTitle className="flex items-center text-neutral-600 gap-2 text-lg">Family & Emergency</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="spouse_name" className="flex items-center text-neutral-600 gap-1.5 px-2"><Heart className="h-3.5 w-3.5" /> Spouse Name</Label>
              <Input size='md' id="spouse_name" placeholder="Spouse name" {...register("spouse_name")} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="emergency_contact" className="flex items-center text-neutral-600 gap-1.5 px-2"><AlertTriangle className="h-3.5 w-3.5" /> Emergency Contact Name</Label>
              <Input size='md' id="emergency_contact" placeholder="Contact name" {...register("emergency_contact")} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="emergency_phone" className="flex items-center text-neutral-600 gap-1.5 px-2"><PhoneCall className="h-3.5 w-3.5" /> Emergency Phone</Label>
              <Input size='md' id="emergency_phone" placeholder="01XXXXXXXXX" {...register("emergency_phone")} />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="number_of_dependents" className="flex items-center text-neutral-600 gap-1.5 px-2"><Users className="h-3.5 w-3.5" /> Number of Dependents</Label>
            <Input size='md' id="number_of_dependents" type="number" min="0" {...register("number_of_dependents", { valueAsNumber: true })} />
          </div>
        </CardContent>
      </Card>

      {/* Office Details */}
      <Card>
        <CardHeader><CardTitle className="flex items-center text-neutral-600 gap-2 text-lg">Office Details</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="company_id" className="flex items-center text-red-500 gap-1.5 px-2"><Building2 className="h-3.5 w-3.5" /> Company *</Label>
              <select id="company_id" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watch("company_id") || ""} onChange={(e) => setValue("company_id", e.target.value)}>
                <option value="">Select company</option>
                {companies.map((c) => <option key={c.id} value={c.id}>{c.company_name_en}</option>)}
              </select>
              {errors.company_id && <p className="text-sm text-destructive">{errors.company_id.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="employee_id" className="flex items-center text-red-500 gap-1.5 px-2"><Badge className="h-3.5 w-3.5" /> Emp. ID *</Label>
              <Input size='md' id="employee_id" placeholder="EMP001" {...register("employee_id")} aria-invalid={!!errors.employee_id} />
              {errors.employee_id && <p className="text-sm text-destructive">{errors.employee_id.message}</p>}
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="punch_number" className="flex items-center text-red-500 gap-1.5 px-2"><Fingerprint className="h-3.5 w-3.5" /> Punch Number *</Label>
              <Input size='md' id="punch_number" placeholder="Badge number" {...register("punch_number")} aria-invalid={!!errors.punch_number} />
              {errors.punch_number && <p className="text-sm text-destructive">{errors.punch_number.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="employee_type" className="flex items-center text-neutral-600 gap-1.5 px-2"><Badge className="h-3.5 w-3.5" /> Employee Type</Label>
              <select id="employee_type" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...register("employee_type")}>
                <option value="">Select</option>
                {employeeTypeOptions.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="joining_date" className="flex items-center text-red-500 gap-1.5 px-2"><CalendarClock className="h-3.5 w-3.5" /> Joining Date *</Label>
              <DatePicker
                value={watch("joining_date") ? new Date(watch("joining_date")!) : undefined}
                onChange={(date) => setValue("joining_date", date ? format(date, "yyyy-MM-dd") : "")}
                placeholder="Select joining date"
              />
              {errors.joining_date && <p className="text-sm text-destructive">{errors.joining_date.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="shift_id" className="flex items-center text-neutral-600 gap-1.5 px-2"><Clock className="h-3.5 w-3.5" /> Shift</Label>
              <select id="shift_id" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watch("shift_id") || ""} onChange={(e) => setValue("shift_id", e.target.value || null)}>
                <option value="">No shift</option>
                {shifts.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
            </div>
          </div>
          {/* Department → Section → Designation + Line */}
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label className="flex items-center text-neutral-600 gap-1.5 px-2"><LayoutGrid className="h-3.5 w-3.5" /> Department</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watchDept || ""} onChange={(e) => setValue("department_id", e.target.value || null)}>
                <option value="">Select department</option>
                {departments.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label className="flex items-center text-neutral-600 gap-1.5 px-2"><Layers className="h-3.5 w-3.5" /> Section</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watchSection || ""} onChange={(e) => setValue("section_id", e.target.value || null)}>
                <option value="">Select section</option>
                {sections.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label className="flex items-center text-neutral-600 gap-1.5 px-2"><Award className="h-3.5 w-3.5" /> Designation</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watch("designation_id") || ""} onChange={(e) => setValue("designation_id", e.target.value || null)}>
                <option value="">Select designation</option>
                {designations.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label className="flex items-center text-neutral-600 gap-1.5 px-2"><Minus className="h-3.5 w-3.5" /> Line</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watch("line_id") || ""} onChange={(e) => setValue("line_id", e.target.value || null)}>
                <option value="">Select line</option>
                {lines.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
              </select>
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="space-y-2">
              <Label className="flex items-center text-neutral-600 gap-1.5 px-2"><Folder className="h-3.5 w-3.5" /> Group</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watch("group_id") || ""} onChange={(e) => setValue("group_id", e.target.value || null)}>
                <option value="">Select group</option>
                {groups.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label className="flex items-center text-neutral-600 gap-1.5 px-2"><Layers className="h-3.5 w-3.5" /> Floor</Label>
              <select className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watch("floor_id") || ""} onChange={(e) => setValue("floor_id", e.target.value || null)}>
                <option value="">Select floor</option>
                {floors.map((f) => <option key={f.id} value={f.id}>{f.name}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="grade" className="flex items-center text-neutral-600 gap-1.5 px-2"><Star className="h-3.5 w-3.5" /> Grade</Label>
              <Input size='md' id="grade" placeholder="Grade" {...register("grade")} />
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="reports_to" className="flex items-center text-neutral-600 gap-1.5 px-2"><UserCog className="h-3.5 w-3.5" /> Reports To</Label>
              <Input size='md' id="reports_to" placeholder="Manager ID" {...register("reports_to")} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="status" className="flex items-center text-neutral-600 gap-1.5 px-2"><ToggleLeft className="h-3.5 w-3.5" /> Status</Label>
              <select id="status" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...register("status")}>
                {statusOptionsEmployee.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="over_time_status" className="flex items-center text-neutral-600 gap-1.5 px-2"><Clock className="h-3.5 w-3.5" /> OverTime Status</Label>
              <select id="over_time_status" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={watch("over_time_status") ? "true" : "false"} onChange={(e) => setValue("over_time_status", e.target.value === "true")}>
                <option value="false">Disable</option>
                <option value="true">Enable</option>
              </select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Salary */}
      <Card>
        <CardHeader><CardTitle className="flex items-center text-neutral-600 gap-2 text-lg">Salary & Benefits</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="gross_salary" className="flex items-center text-red-500 gap-1.5 px-2"><Calculator className="h-3.5 w-3.5" /> Gross Salary *</Label>
              <Input size='md' id="gross_salary" type="number" min="0" {...register("gross_salary", { valueAsNumber: true })} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="basic_salary" className="flex items-center text-neutral-600 gap-1.5 px-2"><Banknote className="h-3.5 w-3.5" /> Basic Salary</Label>
              <Input size='md' id="basic_salary" type="number" {...register("basic_salary", { valueAsNumber: true })} readOnly className="bg-muted" />
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="house_rent" className="flex items-center text-neutral-600 gap-1.5 px-2"><Home className="h-3.5 w-3.5" /> House Rent</Label>
              <Input size='md' id="house_rent" type="number" {...register("house_rent", { valueAsNumber: true })} readOnly className="bg-muted" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="transport_allowance" className="flex items-center text-neutral-600 gap-1.5 px-2"><Car className="h-3.5 w-3.5" /> Transport Allowance</Label>
              <Input size='md' id="transport_allowance" type="number" {...register("transport_allowance", { valueAsNumber: true })} readOnly className="bg-muted" />
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="food_allowance" className="flex items-center text-neutral-600 gap-1.5 px-2"><UtensilsCrossed className="h-3.5 w-3.5" /> Food Allowance</Label>
              <Input size='md' id="food_allowance" type="number" {...register("food_allowance", { valueAsNumber: true })} readOnly className="bg-muted" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="medical_allowance" className="flex items-center text-neutral-600 gap-1.5 px-2"><Stethoscope className="h-3.5 w-3.5" /> Medical Allowance</Label>
              <Input size='md' id="medical_allowance" type="number" {...register("medical_allowance", { valueAsNumber: true })} readOnly className="bg-muted" />
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="other_allowance" className="flex items-center text-neutral-600 gap-1.5 px-2"><PlusCircle className="h-3.5 w-3.5" /> Other Allowance</Label>
              <Input size='md' id="other_allowance" type="number" min="0" {...register("other_allowance", { valueAsNumber: true })} />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Account Information */}
      <Card>
        <CardHeader><CardTitle className="flex items-center text-neutral-600 gap-2 text-lg">Account Information</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="account_type" className="flex items-center text-neutral-600 gap-1.5 px-2"><Smartphone className="h-3.5 w-3.5" /> Account Type</Label>
              <select id="account_type" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" {...register("account_type")}>
                <option value="">Select type</option>
                <option value="mCash">mCash</option>
                <option value="Card">Card</option>
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="account_number" className="flex items-center text-neutral-600 gap-1.5 px-2"><Hash className="h-3.5 w-3.5" /> Account Number</Label>
              <Input size='md' id="account_number" placeholder="Account number" maxLength={17} {...register("account_number")} />
              {watch("account_type") === "mCash" && <p className="text-xs text-muted-foreground">Must be 12 digits for mCash</p>}
              {watch("account_type") === "Card" && <p className="text-xs text-muted-foreground">Must be 17 digits for Card</p>}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Image & Signature */}
      <Card>
        <CardHeader><CardTitle className="flex items-center text-neutral-600 gap-2 text-lg">Image & Signature</CardTitle></CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label className="flex items-center text-neutral-600 gap-1.5 px-2"><Camera className="h-3.5 w-3.5" /> Photo</Label>
              <ImageUpload value={watch("image_url")} onChange={(url) => setValue("image_url", url)} label="Upload Photo" />
            </div>
            <div className="space-y-2">
              <Label className="flex items-center text-neutral-600 gap-1.5 px-2"><PenTool className="h-3.5 w-3.5" /> Signature</Label>
              <ImageUpload value={watch("signature_url")} onChange={(url) => setValue("signature_url", url)} label="Upload Signature" />
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end gap-4 pt-4 border-t">
        {onCancel && <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>Cancel</Button>}
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</> : isEditing ? "Update Employee" : "Create Employee"}
        </Button>
      </div>
    </form>
  )
}
