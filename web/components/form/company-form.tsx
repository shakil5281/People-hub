"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2, Building2, MapPin, Phone, Mail } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { companySchema, CompanyFormData, statusOptions } from "../data/company-data"
import { companyApi } from "@/lib/api"
import { ImageUpload } from "@/components/image-upload"

interface CompanyFormProps {
  initialData?: Partial<CompanyFormData>
  onSuccess: () => void
  onCancel?: () => void
  isEditing?: boolean
  companyId?: string
}

export function CompanyForm({ initialData, onSuccess, onCancel, isEditing = false, companyId }: CompanyFormProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [error, setError] = React.useState("")

  const { register, handleSubmit, watch, setValue, formState: { errors } } = useForm({
    resolver: zodResolver(companySchema),
    defaultValues: {
      company_name_en: initialData?.company_name_en || "",
      company_name_bn: initialData?.company_name_bn || "",
      address_bn: initialData?.address_bn || "",
      address_en: initialData?.address_en || "",
      phone: initialData?.phone || "",
      email: initialData?.email || "",
      signature: initialData?.signature || "",
      status: (initialData?.status as "active" | "inactive") || "active",
    },
  })

  const signatureValue = watch("signature")

  const onSubmit = async (data: Record<string, unknown>) => {
    setIsSubmitting(true)
    setError("")
    try {
      if (isEditing && companyId) {
        await companyApi.update(companyId, data)
        toast.success("Company updated successfully")
      } else {
        await companyApi.create(data)
        toast.success("Company created successfully")
      }
      onSuccess()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to save company"
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
        <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Company Information
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="company_name_en">Company Name (English) *</Label>
              <Input
                id="company_name_en"
                placeholder="Acme Corporation"
                {...register("company_name_en")}
                aria-invalid={!!errors.company_name_en}
              />
              {errors.company_name_en && <p className="text-sm text-destructive">{errors.company_name_en.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="company_name_bn">Company Name (Bengali)</Label>
              <Input
                id="company_name_bn"
                placeholder="‡Kv¤úvbxi bvg"
                className="bangla-input"
                {...register("company_name_bn")}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MapPin className="h-5 w-5" />
            Address
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4">
            <div className="space-y-2">
              <Label htmlFor="address_en">Address (English)</Label>
              <textarea
                id="address_en"
                className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                placeholder="Address in English"
                {...register("address_en")}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="address_bn">Address (Bengali)</Label>
              <textarea
                id="address_bn"
                className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 bangla-input"
                placeholder="wVKvbv"
                {...register("address_bn")}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Phone className="h-5 w-5" />
            Contact Information
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="phone">Phone *</Label>
              <div className="relative">
                <Phone className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="phone"
                  placeholder="+1 (555) 123-4567"
                  className="pl-10"
                  {...register("phone")}
                  aria-invalid={!!errors.phone}
                />
              </div>
              {errors.phone && <p className="text-sm text-destructive">{errors.phone.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="email"
                  type="email"
                  placeholder="company@example.com"
                  className="pl-10"
                  {...register("email")}
                  aria-invalid={!!errors.email}
                />
              </div>
              {errors.email && <p className="text-sm text-destructive">{errors.email.message}</p>}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Signature
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <Label>Company Signature</Label>
            <ImageUpload
              value={signatureValue}
              onChange={(url) => setValue("signature", url)}
              label="Upload Signature"
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Status
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <select
              id="status"
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              {...register("status")}
            >
              {statusOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end gap-4 pt-4 border-t">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>
            Cancel
          </Button>
        )}
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Saving...
            </>
          ) : (
            isEditing ? "Update Company" : "Create Company"
          )}
        </Button>
      </div>
    </form>
  )
}
