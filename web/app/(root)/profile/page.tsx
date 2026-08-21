"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { UserIcon, Loader2, SaveIcon, KeyIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { authApi } from "@/lib/api"
import { toast } from "sonner"

export default function ProfilePage() {
  const router = useRouter()
  const [user, setUser] = React.useState<{ id: string; name: string; email: string; status: string; created_at: string } | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [name, setName] = React.useState("")

  const [pwSaving, setPwSaving] = React.useState(false)
  const [currentPassword, setCurrentPassword] = React.useState("")
  const [newPassword, setNewPassword] = React.useState("")
  const [confirmPassword, setConfirmPassword] = React.useState("")
  const [pwError, setPwError] = React.useState("")

  React.useEffect(() => {
    authApi.me()
      .then((res) => {
        setUser(res.data)
        setName(res.data.name || "")
      })
      .catch(() => router.push("/login"))
      .finally(() => setLoading(false))
  }, [router])

  const handleUpdateProfile = async () => {
    if (!name.trim()) return
    setSaving(true)
    try {
      const { data } = await authApi.updateProfile({ name })
      setUser(data)
      toast.success("Profile updated")
    } catch {
      toast.error("Failed to update profile")
    } finally {
      setSaving(false)
    }
  }

  const handleChangePassword = async () => {
    setPwError("")
    if (newPassword.length < 12) {
      setPwError("New password must be at least 12 characters")
      return
    }
    if (newPassword !== confirmPassword) {
      setPwError("Passwords do not match")
      return
    }
    setPwSaving(true)
    try {
      await authApi.changePassword({ current_password: currentPassword, new_password: newPassword })
      toast.success("Password changed successfully. Please login again.")
      setCurrentPassword("")
      setNewPassword("")
      setConfirmPassword("")
    } catch (err: unknown) {
      const msg = err && typeof err === "object" && "response" in err
        ? (err as { response: { data: { error: string } } }).response?.data?.error || "Failed to change password"
        : "Failed to change password"
      setPwError(msg)
    } finally {
      setPwSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-2">
        <UserIcon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Profile</h1>
          <p className="text-muted-foreground mt-1">Manage your account settings</p>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Personal Information</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>Email</Label>
              <Input value={user?.email || ""} disabled />
            </div>
            <div className="space-y-2">
              <Label>Name</Label>
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Status</Label>
              <Input value={user?.status || ""} disabled />
            </div>
            <div className="space-y-2">
              <Label>Member Since</Label>
              <Input value={user?.created_at ? new Date(user.created_at).toLocaleDateString() : ""} disabled />
            </div>
            <Button onClick={handleUpdateProfile} disabled={saving || !name.trim()}>
              {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              <SaveIcon className="mr-2 h-4 w-4" />
              Save Changes
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Change Password</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {pwError && (
              <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{pwError}</div>
            )}
            <div className="space-y-2">
              <Label>Current Password</Label>
              <Input type="password" value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>New Password</Label>
              <Input type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Confirm New Password</Label>
              <Input type="password" value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)} />
            </div>
            <Button onClick={handleChangePassword} disabled={pwSaving || !currentPassword || !newPassword || !confirmPassword}>
              {pwSaving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              <KeyIcon className="mr-2 h-4 w-4" />
              Update Password
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
