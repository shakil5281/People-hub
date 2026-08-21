"use client"

import * as React from "react"
import { ShieldCheckIcon, PlusIcon, Loader2, Trash2Icon } from "lucide-react"
import { toast } from "sonner"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle,
} from "@/components/ui/dialog"
import { Checkbox } from "@/components/ui/checkbox"
import { roleApi, permissionApi } from "@/lib/api"

interface Role {
  id: string
  name: string
  description: string
  is_system: boolean
  created_at: string
}

interface Permission {
  id: string
  resource: string
  action: string
  description: string
}

export default function RolesPage() {
  const [roles, setRoles] = React.useState<Role[]>([])
  const [allPermissions, setAllPermissions] = React.useState<Permission[]>([])
  const [loading, setLoading] = React.useState(true)
  const [selectedRole, setSelectedRole] = React.useState<Role | null>(null)
  const [rolePermissions, setRolePermissions] = React.useState<string[]>([])
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [createOpen, setCreateOpen] = React.useState(false)
  const [newName, setNewName] = React.useState("")
  const [newDesc, setNewDesc] = React.useState("")
  const [saving, setSaving] = React.useState(false)

  const fetchData = React.useCallback(async () => {
    try {
      const [rRes, pRes] = await Promise.all([roleApi.list(), permissionApi.list()])
      setRoles(rRes.data?.data || [])
      setAllPermissions(pRes.data?.data || [])
    } catch {
      toast.error("Failed to load roles")
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchData() }, [fetchData])

  const openRole = async (role: Role) => {
    setSelectedRole(role)
    try {
      const { data: res } = await roleApi.get(role.id)
      setRolePermissions((res.permissions || []).map((p: Permission) => p.id))
    } catch {
      setRolePermissions([])
    }
    setDialogOpen(true)
  }

  const togglePermission = (permId: string) => {
    setRolePermissions((prev) =>
      prev.includes(permId) ? prev.filter((id) => id !== permId) : [...prev, permId]
    )
  }

  const savePermissions = async () => {
    if (!selectedRole) return
    setSaving(true)
    try {
      await roleApi.assignPermissions(selectedRole.id, { permission_ids: rolePermissions })
      toast.success("Permissions updated")
      setDialogOpen(false)
    } catch {
      toast.error("Failed to save permissions")
    } finally {
      setSaving(false)
    }
  }

  const handleCreate = async () => {
    if (!newName.trim()) return
    setSaving(true)
    try {
      await roleApi.create({ name: newName.trim(), description: newDesc.trim() })
      toast.success("Role created")
      setCreateOpen(false)
      setNewName("")
      setNewDesc("")
      fetchData()
    } catch {
      toast.error("Failed to create role")
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (role: Role) => {
    if (role.is_system) { toast.error("Cannot delete system role"); return }
    try {
      await roleApi.delete(role.id)
      toast.success("Role deleted")
      fetchData()
    } catch {
      toast.error("Failed to delete role")
    }
  }

  const groupedPerms = React.useMemo(() => {
    const map = new Map<string, Permission[]>()
    for (const p of allPermissions) {
      const list = map.get(p.resource) || []
      list.push(p)
      map.set(p.resource, list)
    }
    return Array.from(map.entries()).sort(([a], [b]) => a.localeCompare(b))
  }, [allPermissions])

  const columns: ColumnDef<Role>[] = [
    { id: "sl", header: "Sl", cell: ({ row }) => row.index + 1 },
    { accessorKey: "name", header: "Name", cell: ({ row }) => (
      <span className="font-medium capitalize">{row.original.name.replace("_", " ")}</span>
    )},
    { accessorKey: "description", header: "Description" },
    { accessorKey: "is_system", header: "Type", cell: ({ row }) => (
      row.original.is_system ? <Badge variant="secondary">System</Badge> : <Badge variant="outline">Custom</Badge>
    )},
    {
      id: "actions",
      header: () => null,
      cell: ({ row }) => (
        <div className="flex justify-end">
          {!row.original.is_system && (
            <Button variant="ghost" size="icon" className="size-8 text-muted-foreground hover:text-destructive"
              onClick={(e) => { e.stopPropagation(); handleDelete(row.original) }}>
              <Trash2Icon className="h-4 w-4" />
            </Button>
          )}
        </div>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <ShieldCheckIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Roles & Permissions</h1>
            <p className="text-muted-foreground mt-1">Manage roles and assign permissions</p>
          </div>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <PlusIcon className="mr-2 h-4 w-4" /> Add Role
        </Button>
      </div>

      <div className="px-4 lg:px-6">
        <DataTable
          data={roles}
          columns={columns}
          loading={loading}
          enableSelection={false}
          onEdit={undefined}
          onDelete={undefined}
        />
      </div>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="capitalize">{selectedRole?.name.replace("_", " ")}</DialogTitle>
            <DialogDescription>{selectedRole?.description || "Assign permissions to this role"}</DialogDescription>
          </DialogHeader>

          {groupedPerms.length > 0 ? (
            <div className="space-y-4">
              {groupedPerms.map(([resource, perms]) => (
                <div key={resource}>
                  <h4 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-2">{resource}</h4>
                  <div className="grid grid-cols-2 gap-2">
                    {perms.map((perm) => (
                      <label key={perm.id} className="flex items-start gap-2 rounded-md border p-2.5 cursor-pointer hover:bg-muted/50">
                        <Checkbox
                          checked={rolePermissions.includes(perm.id)}
                          onCheckedChange={() => togglePermission(perm.id)}
                          className="mt-0.5"
                        />
                        <div>
                          <p className="text-sm font-medium">{perm.action}</p>
                          {perm.description && <p className="text-xs text-muted-foreground">{perm.description}</p>}
                        </div>
                      </label>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground text-center py-4">No permissions defined</p>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button onClick={savePermissions} disabled={saving}>
              {saving ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</> : "Save Permissions"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Role</DialogTitle>
            <DialogDescription>Add a new custom role</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="role_name">Role Name</Label>
              <Input id="role_name" value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="e.g. hr_manager" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="role_desc">Description</Label>
              <Input id="role_desc" value={newDesc} onChange={(e) => setNewDesc(e.target.value)} placeholder="Optional description" />
            </div>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button onClick={handleCreate} disabled={saving || !newName.trim()}>
              {saving ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Creating...</> : "Create"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
