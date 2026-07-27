"use client"

import * as React from "react"
import {
  SettingsIcon, Building2Icon, ShieldCheckIcon, BellIcon, ClockIcon,
  DollarSignIcon, CalendarIcon, Loader2, SaveIcon
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Badge } from "@/components/ui/badge"
import { toast } from "sonner"
import { settingsApi } from "@/lib/api"

const defaultSettings: Record<string, string> = {
  company_name: "",
  company_name_bn: "",
  company_email: "",
  company_phone: "",
  company_address: "",
  ot_rate: "1.5",
  attendance_bonus: "500",
  late_grace_minutes: "15",
  daily_work_hours: "8",
  weekend_days: "Fri",
  overtime_enabled: "true",
  basic_salary_pct: "50",
  house_rent_pct: "25",
  medical_pct: "10",
  pf_pct: "0",
  tax_pct: "0",
  min_password_length: "12",
  token_expiry_hours: "24",
  session_timeout_minutes: "60",
  mfa_enabled: "false",
  leave_approval_method: "manual",
  max_carry_forward: "10",
  email_alerts: "true",
  sms_alerts: "false",
  attendance_alerts: "true",
  leave_alerts: "true",
}

const inputClass = "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"

export default function AdminSettingsPage() {
  const [settings, setSettings] = React.useState<Record<string, string>>(defaultSettings)
  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [activeTab, setActiveTab] = React.useState("company")

  React.useEffect(() => {
    settingsApi.list()
      .then(({ data }) => {
        if (data?.data) setSettings((prev) => ({ ...prev, ...data.data }))
      })
      .catch(() => toast.error("Failed to load settings"))
      .finally(() => setLoading(false))
  }, [])

  const update = (key: string, value: string) => {
    setSettings((prev) => ({ ...prev, [key]: value }))
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await settingsApi.update({ settings })
      toast.success("Settings saved successfully")
    } catch {
      toast.error("Failed to save settings")
    } finally {
      setSaving(false)
    }
  }

  const tf = (key: string, label: string, type = "text", placeholder = "", hint = "") => (
    <div className="space-y-1.5">
      <Label htmlFor={key}>{label}</Label>
      <Input id={key} type={type} value={settings[key] || ""} onChange={(e) => update(key, e.target.value)} placeholder={placeholder} className={inputClass} />
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  )

  const sf = (key: string, label: string, options: { value: string; label: string }[]) => (
    <div className="space-y-1.5">
      <Label htmlFor={key}>{label}</Label>
      <Select value={settings[key] || ""} onValueChange={(v) => update(key, v)}>
        <SelectTrigger id={key} className={inputClass}>
          <SelectValue placeholder="Select..." />
        </SelectTrigger>
        <SelectContent>
          {options.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
        </SelectContent>
      </Select>
    </div>
  )

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const SectionHeader = ({ icon: Icon, title, desc }: { icon: React.ComponentType<{ className?: string }>; title: string; desc: string }) => (
    <div className="flex items-center gap-3 pb-4 border-b">
      <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
        <Icon className="h-5 w-5 text-primary" />
      </div>
      <div>
        <h3 className="font-semibold">{title}</h3>
        <p className="text-sm text-muted-foreground">{desc}</p>
      </div>
    </div>
  )

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <SettingsIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">System Settings</h1>
            <p className="text-muted-foreground mt-1">Configure system-wide parameters</p>
          </div>
        </div>
        <Button onClick={handleSave} disabled={saving} size="lg">
          {saving ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</> : <><SaveIcon className="mr-2 h-4 w-4" /> Save All Settings</>}
        </Button>
      </div>

      <div className="px-4 lg:px-6">
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="flex-wrap w-full max-w-3xl">
            <TabsTrigger value="company"><Building2Icon className="h-4 w-4 mr-1.5" />Company</TabsTrigger>
            <TabsTrigger value="attendance"><ClockIcon className="h-4 w-4 mr-1.5" />Attendance</TabsTrigger>
            <TabsTrigger value="payroll"><DollarSignIcon className="h-4 w-4 mr-1.5" />Payroll</TabsTrigger>
            <TabsTrigger value="leave"><CalendarIcon className="h-4 w-4 mr-1.5" />Leave</TabsTrigger>
            <TabsTrigger value="security"><ShieldCheckIcon className="h-4 w-4 mr-1.5" />Security</TabsTrigger>
          </TabsList>

          <div className="mt-6 space-y-6">
            <TabsContent value="company">
              <Card>
                <CardContent className="pt-6 space-y-6">
                  <SectionHeader icon={Building2Icon} title="Company Information" desc="Basic company details displayed on reports and ID cards" />
                  <div className="grid gap-4 sm:grid-cols-2">
                    {tf("company_name", "Company Name (English)", "text", "Ekushe Fashions Ltd.")}
                    {tf("company_name_bn", "Company Name (Bangla)", "text", "কোম্পানির নাম")}
                    {tf("company_email", "Email Address", "email", "info@company.com")}
                    {tf("company_phone", "Phone Number", "text", "+880 1700-000000")}
                  </div>
                  {tf("company_address", "Full Address", "text", "House, Road, Sector, City")}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="attendance">
              <Card>
                <CardContent className="pt-6 space-y-6">
                  <SectionHeader icon={ClockIcon} title="Attendance Rules" desc="Configure how attendance is calculated and displayed" />
                  <div className="grid gap-4 sm:grid-cols-3">
                    {tf("daily_work_hours", "Daily Work Hours", "number", "", "Standard working hours per day")}
                    {tf("late_grace_minutes", "Late Grace (Minutes)", "number", "", "Allowed late minutes before marked late")}
                    {tf("ot_rate", "Overtime Rate Multiplier", "number", "", "e.g. 1.5 = time and a half")}
                  </div>
                  <div className="grid gap-4 sm:grid-cols-3">
                    {tf("weekend_days", "Weekend Days", "text", "Fri", 'Comma-separated: "Fri,Sat"')}
                    {sf("overtime_enabled", "Overtime Calculation", [
                      { value: "true", label: "Enabled" },
                      { value: "false", label: "Disabled" },
                    ])}
                    {tf("attendance_bonus", "Attendance Bonus (BDT)", "number", "", "Bonus for perfect attendance")}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="payroll">
              <Card>
                <CardContent className="pt-6 space-y-6">
                  <SectionHeader icon={DollarSignIcon} title="Payroll Rules" desc="Salary breakdown percentages and deduction rules" />
                  <div className="grid gap-4 sm:grid-cols-3">
                    {tf("basic_salary_pct", "Basic Salary (%)", "number", "", "% of gross salary")}
                    {tf("house_rent_pct", "House Rent (%)", "number", "", "% of gross salary")}
                    {tf("medical_pct", "Medical (%)", "number", "", "% of gross salary")}
                  </div>
                  <div className="grid gap-4 sm:grid-cols-3">
                    {tf("pf_pct", "Provident Fund (%)", "number", "", "% of basic salary")}
                    {tf("tax_pct", "Tax (%)", "number", "", "% of gross salary")}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="leave">
              <Card>
                <CardContent className="pt-6 space-y-6">
                  <SectionHeader icon={CalendarIcon} title="Leave Rules" desc="Configure leave approval and carry forward policies" />
                  <div className="grid gap-4 sm:grid-cols-3">
                    <div className="space-y-1.5">
                      <Label>Leave Approval Method</Label>
                      <Select value={settings.leave_approval_method || "manual"} onValueChange={(v) => update("leave_approval_method", v)}>
                        <SelectTrigger className={inputClass}>
                          <SelectValue placeholder="Select..." />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="manual">Manual (Manager)</SelectItem>
                          <SelectItem value="automatic">Automatic</SelectItem>
                          <SelectItem value="hod">HOD Approval</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    {tf("max_carry_forward", "Max Carry Forward (Days)", "number", "", "Leave days carried to next year")}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="security">
              <div className="space-y-6">
                <Card>
                  <CardContent className="pt-6 space-y-6">
                    <SectionHeader icon={ShieldCheckIcon} title="Authentication" desc="Password policies and token management" />
                    <div className="grid gap-4 sm:grid-cols-3">
                      {tf("min_password_length", "Min Password Length", "number")}
                      {tf("token_expiry_hours", "Token Expiry (Hours)", "number", "", "Access token lifetime")}
                      {tf("session_timeout_minutes", "Session Timeout (Minutes)", "number", "", "Auto logout after inactivity")}
                    </div>
                    {sf("mfa_enabled", "Multi-Factor Authentication", [
                      { value: "true", label: "Enabled" },
                      { value: "false", label: "Disabled" },
                    ])}
                  </CardContent>
                </Card>

                <Card>
                  <CardContent className="pt-6 space-y-6">
                    <SectionHeader icon={BellIcon} title="Notifications" desc="Email and SMS alert preferences" />
                    <div className="grid gap-4 sm:grid-cols-2">
                      {sf("email_alerts", "Email Alerts", [{ value: "true", label: "Enabled" }, { value: "false", label: "Disabled" }])}
                      {sf("sms_alerts", "SMS Alerts", [{ value: "true", label: "Enabled" }, { value: "false", label: "Disabled" }])}
                      {sf("attendance_alerts", "Attendance Alerts", [{ value: "true", label: "Enabled" }, { value: "false", label: "Disabled" }])}
                      {sf("leave_alerts", "Leave Alerts", [{ value: "true", label: "Enabled" }, { value: "false", label: "Disabled" }])}
                    </div>
                  </CardContent>
                </Card>
              </div>
            </TabsContent>
          </div>
        </Tabs>
      </div>
    </div>
  )
}
