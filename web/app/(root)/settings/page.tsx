"use client"

import * as React from "react"
import { useSearchParams, useRouter } from "next/navigation"
import {
  SettingsIcon,
  LockKeyholeIcon,
  BellIcon,
  MailIcon,
  SmartphoneIcon,
  GlobeIcon,
  MoonIcon,
  ShieldCheckIcon,
  Loader2,
  CheckCircleIcon,
} from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Separator } from "@/components/ui/separator"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { authApi } from "@/lib/api"
import { toast } from "sonner"

function SettingsContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const forceChange = searchParams.get("force_change") === "true"

  const [currentPassword, setCurrentPassword] = React.useState("")
  const [newPassword, setNewPassword] = React.useState("")
  const [confirmPassword, setConfirmPassword] = React.useState("")
  const [pwSubmitting, setPwSubmitting] = React.useState(false)
  const [pwError, setPwError] = React.useState("")
  const [pwSuccess, setPwSuccess] = React.useState(false)

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    setPwError("")

    if (newPassword.length < 12) {
      setPwError("Password must be at least 12 characters")
      return
    }
    if (newPassword !== confirmPassword) {
      setPwError("Passwords do not match")
      return
    }

    setPwSubmitting(true)
    try {
      await authApi.changePassword({ current_password: currentPassword, new_password: newPassword })
      setPwSuccess(true)
      toast.success("Password changed successfully")
      if (forceChange) {
        setTimeout(() => router.push("/dashboard"), 1500)
      }
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || "Failed to change password"
      setPwError(msg)
      toast.error(msg)
    } finally {
      setPwSubmitting(false)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      {forceChange && (
        <Alert className="border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/20">
          <LockKeyholeIcon className="h-4 w-4 text-amber-600" />
          <AlertDescription className="text-amber-700 dark:text-amber-400">
            You must change your password before continuing. This is required by your administrator.
          </AlertDescription>
        </Alert>
      )}

      <Card className={forceChange ? "border-amber-300 dark:border-amber-700" : ""}>
        <CardHeader>
          <div className="flex items-center gap-2">
            <LockKeyholeIcon className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-lg">Change Password</CardTitle>
          </div>
          <CardDescription>
            {forceChange
              ? "Set a new password to continue using the system"
              : "Update your account password"}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {pwSuccess ? (
            <div className="flex items-center gap-2 text-emerald-600">
              <CheckCircleIcon className="h-5 w-5" />
              <span className="font-medium">Password changed successfully{forceChange ? ", redirecting..." : ""}</span>
            </div>
          ) : (
            <form onSubmit={handleChangePassword} className="space-y-4 max-w-md">
              {pwError && (
                <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{pwError}</div>
              )}
              <div className="space-y-2">
                <Label htmlFor="current-password">Current Password</Label>
                <Input id="current-password" type="password" required
                  value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)}
                  placeholder="Enter current password" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="new-password">New Password</Label>
                <Input id="new-password" type="password" required minLength={12}
                  value={newPassword} onChange={(e) => setNewPassword(e.target.value)}
                  placeholder="Minimum 12 characters" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirm-password">Confirm New Password</Label>
                <Input id="confirm-password" type="password" required
                  value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="Re-enter new password" />
              </div>
              <Button type="submit" disabled={pwSubmitting}>
                {pwSubmitting ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Changing...</> : "Change Password"}
              </Button>
            </form>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <BellIcon className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-lg">Notification Preferences</CardTitle>
          </div>
          <CardDescription>Choose how you receive notifications</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <MailIcon className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-medium">Email Notifications</h3>
            </div>
            <div className="ml-6 space-y-3">
              {[
                { label: "Leave requests", desc: "When an employee requests leave" },
                { label: "Attendance alerts", desc: "Daily attendance summary and alerts" },
                { label: "Employee updates", desc: "New hires, terminations, and changes" },
              ].map((item) => (
                <div key={item.label} className="flex items-center justify-between">
                  <div>
                    <Label className="text-sm font-normal">{item.label}</Label>
                    <p className="text-xs text-muted-foreground">{item.desc}</p>
                  </div>
                  <Switch defaultChecked />
                </div>
              ))}
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <SmartphoneIcon className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-medium">Push Notifications</h3>
            </div>
            <div className="ml-6 space-y-3">
              {[
                { label: "Real-time alerts", desc: "Instant notifications in browser" },
                { label: "Overtime approvals", desc: "When overtime is approved or rejected" },
                { label: "System updates", desc: "Maintenance and feature updates" },
              ].map((item) => (
                <div key={item.label} className="flex items-center justify-between">
                  <div>
                    <Label className="text-sm font-normal">{item.label}</Label>
                    <p className="text-xs text-muted-foreground">{item.desc}</p>
                  </div>
                  <Switch defaultChecked={item.label !== "System updates"} />
                </div>
              ))}
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <GlobeIcon className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-medium">In-App Notifications</h3>
            </div>
            <div className="ml-6 space-y-3">
              {[
                { label: "Show in notification center", desc: "Display notifications in the app" },
                { label: "Sound alerts", desc: "Play sound for new notifications" },
                { label: "Desktop notifications", desc: "Show native OS notifications" },
              ].map((item) => (
                <div key={item.label} className="flex items-center justify-between">
                  <div>
                    <Label className="text-sm font-normal">{item.label}</Label>
                    <p className="text-xs text-muted-foreground">{item.desc}</p>
                  </div>
                  <Switch defaultChecked={item.label !== "Sound alerts"} />
                </div>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <MoonIcon className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-lg">Appearance</CardTitle>
          </div>
          <CardDescription>Customize the look and feel</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <Label>Dark Mode</Label>
              <p className="text-sm text-muted-foreground">Use dark theme across the app</p>
            </div>
            <Switch />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <ShieldCheckIcon className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-lg">Privacy & Security</CardTitle>
          </div>
          <CardDescription>Manage your security preferences</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <Label>Two-factor authentication</Label>
              <p className="text-sm text-muted-foreground">Add an extra layer of security</p>
            </div>
            <Switch />
          </div>
          <div className="flex items-center justify-between">
            <div>
              <Label>Login notifications</Label>
              <p className="text-sm text-muted-foreground">Alert on new device login</p>
            </div>
            <Switch defaultChecked />
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default function SettingsPage() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-2">
        <SettingsIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
          <p className="text-muted-foreground mt-1">Manage your preferences</p>
        </div>
      </div>
      <React.Suspense fallback={null}>
        <SettingsContent />
      </React.Suspense>
    </div>
  )
}
