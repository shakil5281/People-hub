"use client"

import * as React from "react"
import { UserCogIcon, PlusIcon, KeyRoundIcon, ShieldCheckIcon } from "lucide-react"
import { toast } from "sonner"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import {
  Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle,
} from "@/components/ui/dialog"
import { userApi } from "@/lib/api"
import { User, statusOptions, statusBadgeVariant } from "@/components/data/user-data"
import { isSuperAdmin } from "@/lib/auth"

const columns: ColumnDef<User>[] = [
  {
    id: "sl",
    header: "SL",
    cell: ({ row }) => row.index + 1,
  },
  { accessorKey: "name", header: "Name" },
  { accessorKey: "email", header: "Email" },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => (
      <Badge variant={statusBadgeVariant[row.original.status] || "outline"} className="capitalize">
        {row.original.status}
      </Badge>
    ),
  },
  {
    accessorKey: "created_at",
    header: "Created",
    cell: ({ row }) => new Date(row.original.created_at).toLocaleDateString("en-GB"),
  },
]

export default function UserManagementPage() {
  const [users, setUsers] = React.useState<User[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [page, setPage] = React.useState(1)
  const [limit, setLimit] = React.useState(20)
  const [total, setTotal] = React.useState(0)
  const [totalPages, setTotalPages] = React.useState(0)
  const [search, setSearch] = React.useState("")

  // Create/Edit dialog
  const [formOpen, setFormOpen] = React.useState(false)
  const [editingUser, setEditingUser] = React.useState<User | null>(null)
  const [formName, setFormName] = React.useState("")
  const [formEmail, setFormEmail] = React.useState("")
  const [formStatus, setFormStatus] = React.useState("active")
  const [formSubmitting, setFormSubmitting] = React.useState(false)
  const [formError, setFormError] = React.useState("")

  // Password reset dialog
  const [resetOpen, setResetOpen] = React.useState(false)
  const [resetUser, setResetUser] = React.useState<User | null>(null)
  const [resetResult, setResetResult] = React.useState<string | null>(null)
  const [resetSubmitting, setResetSubmitting] = React.useState(false)

  // Roles dialog
  const [rolesOpen, setRolesOpen] = React.useState(false)
  const [rolesUser, setRolesUser] = React.useState<User | null>(null)
  const [allRoles, setAllRoles] = React.useState<Array<{ id: string; name: string }>>([])
  const [selectedRoleIds, setSelectedRoleIds] = React.useState<string[]>([])
  const [rolesSubmitting, setRolesSubmitting] = React.useState(false)
  const [rolesLoading, setRolesLoading] = React.useState(false)

  const fetchUsers = React.useCallback(async (p?: number, s?: string) => {
    setLoading(true)
    setError("")
    try {
      const params: Record<string, string> = { page: String(p ?? page), limit: String(limit) }
      if (s ?? search) params.search = s ?? search
      const { data: res } = await userApi.list(params)
      setUsers(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setError("Failed to load users")
    } finally {
      setLoading(false)
    }
  }, [page, limit, search])

  React.useEffect(() => { fetchUsers() }, [fetchUsers])

  const openCreate = () => {
    setEditingUser(null)
    setFormName("")
    setFormEmail("")
    setFormStatus("active")
    setFormError("")
    setFormOpen(true)
  }

  const openEdit = (user: User) => {
    setEditingUser(user)
    setFormName(user.name)
    setFormEmail(user.email)
    setFormStatus(user.status)
    setFormError("")
    setFormOpen(true)
  }

  const handleFormSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormSubmitting(true)
    setFormError("")
    try {
      if (editingUser) {
        await userApi.update(editingUser.id, { name: formName, status: formStatus })
        toast.success("User updated")
      } else {
        const { data } = await userApi.create({ email: formEmail, name: formName })
        const pw = (data as { generated_password?: string }).generated_password
        if (pw) toast.success(`User created. Password: ${pw}`, { duration: 10000 })
        else toast.success("User created")
      }
      setFormOpen(false)
      fetchUsers()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to save user"
      const detail = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || msg
      setFormError(detail)
      toast.error(detail)
    } finally {
      setFormSubmitting(false)
    }
  }

  const handleDelete = async (user: User) => {
    if (!confirm(`Delete user "${user.name}"? This soft-deletes the account.`)) return
    try {
      await userApi.delete(user.id)
      toast.success("User deleted")
      fetchUsers()
    } catch {
      toast.error("Failed to delete user")
    }
  }

  const openReset = (user: User) => {
    setResetUser(user)
    setResetResult(null)
    setResetOpen(true)
  }

  const handleReset = async () => {
    if (!resetUser) return
    setResetSubmitting(true)
    try {
      const { data } = await userApi.resetPassword(resetUser.id)
      setResetResult(data.generated_password)
      toast.success("Password reset. Share the new password securely.")
      fetchUsers()
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || "Failed to reset password"
      toast.error(msg)
    } finally {
      setResetSubmitting(false)
    }
  }

  const openRoles = async (user: User) => {
    setRolesUser(user)
    setRolesLoading(true)
    setRolesOpen(true)
    try {
      const { data: userData } = await userApi.get(user.id)
      const currentRoles = (userData as { roles?: Array<{ id: string; name: string }> }).roles || []
      setAllRoles(currentRoles)
      setSelectedRoleIds(currentRoles.map((r) => r.id))
    } catch {
      toast.error("Failed to load user roles")
    } finally {
      setRolesLoading(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <UserCogIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">User Management</h1>
            <p className="text-muted-foreground mt-1">Manage system users, roles, and passwords</p>
          </div>
        </div>
        {isSuperAdmin() && (
          <Button onClick={openCreate}>
            <PlusIcon className="mr-2 h-4 w-4" />
            Add User
          </Button>
        )}
      </div>

      <div className="px-4 lg:px-6 flex items-center gap-3">
        <Input
          placeholder="Search by name or email..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="max-w-xs"
          onKeyDown={(e) => { if (e.key === "Enter") { setPage(1); fetchUsers(1, search) } }}
        />
        <Button variant="secondary" size="sm" onClick={() => { setPage(1); fetchUsers(1, search) }}>
          Search
        </Button>
        {search && (
          <Button variant="ghost" size="sm" onClick={() => { setSearch(""); setPage(1); fetchUsers(1, "") }}>
            Clear
          </Button>
        )}
      </div>

      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}

      <DataTable
        data={users}
        columns={[
          ...columns,
          {
            id: "actions",
            header: "Actions",
            cell: ({ row }) => (
              <div className="flex items-center gap-1">
                {isSuperAdmin() && (
                  <>
                    <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => openRoles(row.original)} title="Manage Roles">
                      <ShieldCheckIcon className="h-4 w-4" />
                    </Button>
                    <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => openReset(row.original)} title="Reset Password">
                      <KeyRoundIcon className="h-4 w-4" />
                    </Button>
                  </>
                )}
              </div>
            ),
            enableSorting: false,
            enableHiding: false,
          },
        ]}
        onEdit={isSuperAdmin() ? openEdit : undefined}
        onDelete={isSuperAdmin() ? handleDelete : undefined}
        serverSide={true}
        page={page}
        pageSize={limit}
        pageCount={totalPages}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={(size) => { setLimit(size); setPage(1) }}
        loading={loading}
      />

      {/* Create/Edit Dialog */}
      <Dialog open={formOpen} onOpenChange={setFormOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingUser ? "Edit User" : "Add New User"}</DialogTitle>
            <DialogDescription>
              {editingUser
                ? "Update user name and status."
                : "Create a new system user. A secure password will be auto-generated."}
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleFormSubmit} className="space-y-4">
            {formError && (
              <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{formError}</div>
            )}

            <div className="space-y-2">
              <Label htmlFor="form-name">Name *</Label>
              <Input id="form-name" required placeholder="Full name" value={formName}
                onChange={(e) => setFormName(e.target.value)} />
            </div>

            <div className="space-y-2">
              <Label htmlFor="form-email">Email *</Label>
              <Input id="form-email" type="email" required placeholder="user@example.com" value={formEmail}
                onChange={(e) => setFormEmail(e.target.value)} disabled={!!editingUser} />
            </div>

            {editingUser && (
              <div className="space-y-2">
                <Label htmlFor="form-status">Status</Label>
                <Select value={formStatus} onValueChange={setFormStatus}>
                  <SelectTrigger id="form-status"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {statusOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            <div className="flex justify-end gap-4 pt-4 border-t">
              <Button type="button" variant="outline" onClick={() => setFormOpen(false)} disabled={formSubmitting}>
                Cancel
              </Button>
              <Button type="submit" disabled={formSubmitting}>
                {formSubmitting ? "Saving..." : editingUser ? "Update User" : "Create User"}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Reset Password Dialog */}
      {isSuperAdmin() && (
        <Dialog open={resetOpen} onOpenChange={setResetOpen}>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>Reset Password</DialogTitle>
              <DialogDescription>
                Reset password for <strong>{resetUser?.name}</strong> ({resetUser?.email}).
                The user will be logged out of all sessions and forced to change password on next login.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              {resetResult && (
                <div className="rounded-md bg-emerald-50 dark:bg-emerald-950/20 p-4">
                  <p className="text-sm font-medium text-emerald-700 dark:text-emerald-400">New Password</p>
                  <p className="mt-1 text-lg font-mono font-bold tracking-wider select-all">{resetResult}</p>
                  <p className="mt-2 text-xs text-muted-foreground">
                    Copy this password now. It will not be shown again.
                  </p>
                </div>
              )}
              <div className="flex justify-end gap-3 pt-4 border-t">
                <Button variant="outline" onClick={() => { setResetOpen(false); setResetResult(null) }}>
                  Close
                </Button>
                {!resetResult && (
                  <Button onClick={handleReset} disabled={resetSubmitting} variant="destructive">
                    {resetSubmitting ? "Resetting..." : "Reset Password"}
                  </Button>
                )}
              </div>
            </div>
          </DialogContent>
        </Dialog>
      )}

      {/* Roles Dialog */}
      {isSuperAdmin() && (
        <Dialog open={rolesOpen} onOpenChange={setRolesOpen}>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>Manage Roles</DialogTitle>
              <DialogDescription>
                Current roles for <strong>{rolesUser?.name}</strong> ({rolesUser?.email}).
                Enter comma-separated role names below to update.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              {rolesLoading ? (
                <p className="text-sm text-muted-foreground">Loading roles...</p>
              ) : (
                <>
                  <div className="flex flex-wrap gap-2">
                    {allRoles.length === 0 && (
                      <p className="text-sm text-muted-foreground">No roles assigned</p>
                    )}
                    {allRoles.map((role) => (
                      <Badge key={role.id} variant="secondary">{role.name}</Badge>
                    ))}
                  </div>
                </>
              )}
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}
