"use client"

import * as React from "react"
import { ShieldCheckIcon, PlusIcon, Loader2, Trash2Icon, PencilIcon, SearchIcon, UsersIcon, KeyRoundIcon, Settings2Icon } from "lucide-react"
import { toast } from "sonner"
import { DataTable } from "@/components/table/data-table"
import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Card, CardContent } from "@/components/ui/card"
import {
  Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle,
} from "@/components/ui/dialog"
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
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

  // filters
  const [q, setQ] = React.useState("")
  const [typeFilter, setTypeFilter] = React.useState<"all"|"system"|"custom">("all")

  // manage perms
  const [selectedRole, setSelectedRole] = React.useState<Role | null>(null)
  const [rolePermissions, setRolePermissions] = React.useState<string[]>([])
  const [initialPerms, setInitialPerms] = React.useState<string[]>([])
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [permSearch, setPermSearch] = React.useState("")

  // create / edit
  const [createOpen, setCreateOpen] = React.useState(false)
  const [editingRole, setEditingRole] = React.useState<Role | null>(null)
  const [newName, setNewName] = React.useState("")
  const [newDesc, setNewDesc] = React.useState("")
  const [saving, setSaving] = React.useState(false)

  // delete confirm
  const [deleteTarget, setDeleteTarget] = React.useState<Role | null>(null)

  // role -> permission count cache
  const [permCountByRole, setPermCountByRole] = React.useState<Record<string, number>>({})

  const fetchData = React.useCallback(async () => {
    setLoading(true)
    try {
      const [rRes, pRes] = await Promise.allSettled([roleApi.list(), permissionApi.list()])
      if (rRes.status === "fulfilled") {
        const r = (rRes.value.data?.data || rRes.value.data || []) as Role[]
        setRoles(Array.isArray(r) ? r : [])
        // fetch perm counts per role in parallel (best-effort, non-blocking)
        if (Array.isArray(r) && r.length > 0) {
          Promise.all(r.map(async (role) => {
            try {
              const { data } = await roleApi.get(role.id) as unknown as { data: { permissions: Permission[] }; permissions?: Permission[] }
              const perms = (data as { permissions?: Permission[] }).permissions || (data as { data?: Permission[] }).data || []
              return { id: role.id, count: Array.isArray(perms) ? perms.length : 0 }
            } catch { return { id: role.id, count: 0 } }
          })).then((results) => {
            const counts: Record<string, number> = {}
            results.forEach(({ id, count }) => { counts[id] = count })
            setPermCountByRole(counts)
          })
        }
      } else {
        toast.error("Failed to load roles")
      }
      if (pRes.status === "fulfilled") {
        const p = (pRes.value.data?.data || pRes.value.data || []) as Permission[]
        setAllPermissions(Array.isArray(p) ? p : [])
      } else {
        // permissions load failure is non-critical
        console.warn("Failed to load permissions", pRes.reason)
      }
    } catch {
      toast.error("Failed to load roles")
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchData() }, [fetchData])

  const filteredRoles = React.useMemo(() => {
    return roles.filter((r) => {
      if (typeFilter === "system" && !r.is_system) return false
      if (typeFilter === "custom" && r.is_system) return false
      if (q.trim()) {
        const qq = q.toLowerCase()
        if (!r.name.toLowerCase().includes(qq) && !(r.description||"").toLowerCase().includes(qq)) return false
      }
      return true
    })
  }, [roles, q, typeFilter])

  const stats = React.useMemo(() => ({
    total: roles.length,
    system: roles.filter((r)=>r.is_system).length,
    custom: roles.filter((r)=>!r.is_system).length,
    perms: allPermissions.length,
  }), [roles, allPermissions])

  const openRole = async (role: Role) => {
    setSelectedRole(role)
    setPermSearch("")
    try {
      const { data: res } = await roleApi.get(role.id) as unknown as { data: { permissions: Permission[] } }
      const ids = (res.permissions || []).map((p) => p.id)
      setRolePermissions(ids)
      setInitialPerms(ids)
    } catch {
      setRolePermissions([])
      setInitialPerms([])
    }
    setDialogOpen(true)
  }

  const openEdit = (role: Role) => {
    if (role.is_system) { toast.error("System role cannot be edited"); return }
    setEditingRole(role)
    setNewName(role.name)
    setNewDesc(role.description || "")
    setCreateOpen(true)
  }

  const togglePermission = (permId: string) => {
    setRolePermissions((prev) => prev.includes(permId) ? prev.filter((id) => id !== permId) : [...prev, permId])
  }

  const toggleResourceAll = (perms: Permission[], checked: boolean) => {
    const ids = perms.map((p)=>p.id)
    setRolePermissions((prev) => {
      if (checked) {
        const set = new Set(prev)
        ids.forEach((id)=> set.add(id))
        return Array.from(set)
      } else {
        return prev.filter((id)=> !ids.includes(id))
      }
    })
  }

  const hasDiff = React.useMemo(() => {
    if (rolePermissions.length !== initialPerms.length) return true
    const a = new Set(rolePermissions); const b = new Set(initialPerms)
    for (const id of a) if (!b.has(id)) return true
    return false
  }, [rolePermissions, initialPerms])

  const savePermissions = async () => {
    if (!selectedRole) return
    if (!hasDiff) { setDialogOpen(false); return }
    setSaving(true)
    try {
      await roleApi.assignPermissions(selectedRole.id, { permission_ids: rolePermissions })
      toast.success("Permissions updated — users will get new permissions on next login")
      setDialogOpen(false)
      fetchData()
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error || "Failed to save permissions"
      toast.error(msg)
    } finally {
      setSaving(false)
    }
  }

  const handleCreateOrUpdate = async () => {
    const name = newName.trim()
    if (!name) { toast.error("Role name is required"); return }
    if (!/^[a-z0-9_]{2,50}$/.test(name)) {
      toast.error("Name must be lowercase letters, numbers, underscores (2-50)")
      return
    }
    setSaving(true)
    try {
      if (editingRole) {
        await roleApi.update(editingRole.id, { name, description: newDesc.trim() })
        toast.success("Role updated")
      } else {
        await roleApi.create({ name, description: newDesc.trim() })
        toast.success("Role created")
      }
      setCreateOpen(false)
      setEditingRole(null)
      setNewName(""); setNewDesc("")
      fetchData()
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error || "Failed to save role"
      toast.error(msg)
    } finally {
      setSaving(false)
    }
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    try {
      await roleApi.delete(deleteTarget.id)
      toast.success("Role deleted")
      setDeleteTarget(null)
      fetchData()
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error || "Failed to delete role"
      toast.error(msg)
    }
  }

  const groupedPerms = React.useMemo(() => {
    const map = new Map<string, Permission[]>()
    for (const p of allPermissions) {
      if (permSearch.trim()) {
        const qq = permSearch.toLowerCase()
        if (!p.resource.toLowerCase().includes(qq) && !p.action.toLowerCase().includes(qq) && !(p.description||"").toLowerCase().includes(qq)) continue
      }
      const list = map.get(p.resource) || []
      list.push(p)
      map.set(p.resource, list)
    }
    return Array.from(map.entries()).sort(([a], [b]) => a.localeCompare(b))
  }, [allPermissions, permSearch])

  const columns: ColumnDef<Role>[] = [
    { id: "sl", header: "SL", cell: ({ row }) => <span className="text-muted-foreground">{row.index + 1}</span>, size: 50 },
    { accessorKey: "name", header: "Role", cell: ({ row }) => (
      <div className="flex flex-col">
        <span className="font-medium capitalize">{row.original.name.replace(/_/g, " ")}</span>
        <span className="text-xs text-muted-foreground font-mono">{row.original.name}</span>
      </div>
    )},
    { accessorKey: "description", header: "Description", cell: ({ row }) => <span className="text-sm truncate max-w-[260px] inline-block">{row.original.description || "—"}</span> },
    { id: "perms", header: "Permissions", cell: ({ row }) => (
      <Badge variant="secondary" className="font-mono">{permCountByRole[row.original.id] ?? "—"}</Badge>
    )},
    { accessorKey: "is_system", header: "Type", cell: ({ row }) => (
      row.original.is_system ? <Badge variant="secondary"><Settings2Icon className="mr-1 h-3 w-3"/>System</Badge> : <Badge variant="outline">Custom</Badge>
    )},
    {
      id: "actions",
      header: () => <span className="sr-only">Actions</span>,
      cell: ({ row }) => (
        <div className="flex justify-end gap-1">
          <Button variant="ghost" size="sm" onClick={(e)=>{e.stopPropagation(); openRole(row.original)}} title="Manage permissions">
            <KeyRoundIcon className="h-4 w-4" /> <span className="hidden xl:inline ml-1">Manage</span>
          </Button>
          {!row.original.is_system && (
            <>
              <Button variant="ghost" size="icon" className="size-8" onClick={(e)=>{e.stopPropagation(); openEdit(row.original)}} title="Edit">
                <PencilIcon className="h-4 w-4" />
              </Button>
              <Button variant="ghost" size="icon" className="size-8 text-muted-foreground hover:text-destructive"
                onClick={(e) => { e.stopPropagation(); setDeleteTarget(row.original) }} title="Delete">
                <Trash2Icon className="h-4 w-4" />
              </Button>
            </>
          )}
        </div>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      {/* Header */}
      <div className="px-4 lg:px-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <div className="flex size-10 items-center justify-center rounded-lg bg-primary/10">
            <ShieldCheckIcon className="h-6 w-6 text-primary" />
          </div>
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Roles & Permissions</h1>
            <p className="text-muted-foreground mt-1 text-sm">Company-scoped RBAC — permissions are <code className="font-mono text-xs bg-muted px-1 py-0.5 rounded">resource.action</code> checked in JWT</p>
          </div>
        </div>
        <Button onClick={() => { setEditingRole(null); setNewName(""); setNewDesc(""); setCreateOpen(true) }}>
          <PlusIcon className="mr-2 h-4 w-4" /> Add Role
        </Button>
      </div>

      {/* Stats */}
      <div className="px-4 lg:px-6 grid gap-3 grid-cols-2 lg:grid-cols-4">
        <Card><CardContent className="p-4 flex items-center gap-3"><div className="rounded-md bg-blue-500/10 p-2"><ShieldCheckIcon className="h-5 w-5 text-blue-600"/></div><div><p className="text-2xl font-bold">{stats.total}</p><p className="text-xs text-muted-foreground">Total Roles</p></div></CardContent></Card>
        <Card><CardContent className="p-4 flex items-center gap-3"><div className="rounded-md bg-zinc-500/10 p-2"><Settings2Icon className="h-5 w-5 text-zinc-600"/></div><div><p className="text-2xl font-bold">{stats.system}</p><p className="text-xs text-muted-foreground">System Roles</p></div></CardContent></Card>
        <Card><CardContent className="p-4 flex items-center gap-3"><div className="rounded-md bg-emerald-500/10 p-2"><UsersIcon className="h-5 w-5 text-emerald-600"/></div><div><p className="text-2xl font-bold">{stats.custom}</p><p className="text-xs text-muted-foreground">Custom Roles</p></div></CardContent></Card>
        <Card><CardContent className="p-4 flex items-center gap-3"><div className="rounded-md bg-amber-500/10 p-2"><KeyRoundIcon className="h-5 w-5 text-amber-600"/></div><div><p className="text-2xl font-bold">{stats.perms}</p><p className="text-xs text-muted-foreground">Permissions</p></div></CardContent></Card>
      </div>

      {/* Filters */}
      <div className="px-4 lg:px-6 flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="relative flex-1 max-w-sm">
          <SearchIcon className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input placeholder="Search role name or description..." className="pl-8" value={q} onChange={(e)=>setQ(e.target.value)} />
        </div>
        <div className="flex items-center gap-1 rounded-lg border p-1">
          {(["all","system","custom"] as const).map((t)=> (
            <Button key={t} variant={typeFilter===t? "secondary":"ghost"} size="sm" className="capitalize h-7" onClick={()=>setTypeFilter(t)}>{t}</Button>
          ))}
        </div>
        <span className="text-xs text-muted-foreground ml-auto hidden sm:inline">{filteredRoles.length} of {roles.length} roles</span>
      </div>

      {/* Table */}
      <div className="px-4 lg:px-6">
        <DataTable
          data={filteredRoles}
          columns={columns}
          loading={loading}
          enableSelection={false}
          enableDnd={false}
        />
      </div>

      {/* Manage Permissions Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-3xl max-h-[85vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle className="capitalize flex items-center gap-2">
              <KeyRoundIcon className="h-5 w-5 text-muted-foreground" />
              {selectedRole?.name.replace(/_/g, " ")}
              {selectedRole?.is_system && <Badge variant="secondary" className="ml-2">System</Badge>}
            </DialogTitle>
            <DialogDescription>{selectedRole?.description || "Assign permissions — checked permissions are encoded as resource.action in JWT at login"}</DialogDescription>
          </DialogHeader>

          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <SearchIcon className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input placeholder="Filter permissions (resource, action)..." className="pl-8 h-9" value={permSearch} onChange={(e)=>setPermSearch(e.target.value)} />
            </div>
            <Badge variant="outline" className="shrink-0">{rolePermissions.length} selected</Badge>
          </div>

          <div className="flex-1 overflow-y-auto pr-1 space-y-4">
            {groupedPerms.length > 0 ? groupedPerms.map(([resource, perms]) => {
              const allChecked = perms.every((p)=> rolePermissions.includes(p.id))
              const someChecked = perms.some((p)=> rolePermissions.includes(p.id))
              return (
                <div key={resource} className="rounded-lg border">
                  <div className="flex items-center justify-between px-3 py-2 bg-muted/40 border-b">
                    <div className="flex items-center gap-2">
                      <Checkbox checked={allChecked} onCheckedChange={(v)=> toggleResourceAll(perms, Boolean(v))} aria-label={`Select all ${resource}`} />
                      <h4 className="text-sm font-semibold uppercase tracking-wider">{resource}</h4>
                      <Badge variant="outline" className="text-xs">{perms.filter((p)=> rolePermissions.includes(p.id)).length}/{perms.length}</Badge>
                    </div>
                    {someChecked && !allChecked && <span className="text-xs text-muted-foreground">partial</span>}
                  </div>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 p-3">
                    {perms.map((perm) => (
                      <label key={perm.id} className={`flex items-start gap-2 rounded-md border p-2.5 cursor-pointer transition-colors ${rolePermissions.includes(perm.id) ? "bg-primary/5 border-primary/20" : "hover:bg-muted/50"}`}>
                        <Checkbox
                          checked={rolePermissions.includes(perm.id)}
                          onCheckedChange={() => togglePermission(perm.id)}
                          className="mt-0.5"
                        />
                        <div className="min-w-0">
                          <p className="text-sm font-medium font-mono">{perm.action}</p>
                          {perm.description && <p className="text-xs text-muted-foreground line-clamp-2">{perm.description}</p>}
                        </div>
                      </label>
                    ))}
                  </div>
                </div>
              )
            }) : (
              <p className="text-sm text-muted-foreground text-center py-8">No permissions match filter</p>
            )}
          </div>

          <div className="flex justify-between items-center pt-2 border-t">
            <p className="text-xs text-muted-foreground hidden sm:block">Changes affect JWT on next login/refresh</p>
            <div className="flex gap-2 ml-auto">
              <Button variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
              <Button onClick={savePermissions} disabled={saving || !hasDiff}>
                {saving ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</> : hasDiff ? "Save Changes" : "No changes"}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Create / Edit Role Dialog */}
      <Dialog open={createOpen} onOpenChange={(o)=>{ setCreateOpen(o); if(!o) setEditingRole(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingRole ? "Edit Role" : "Create Role"}</DialogTitle>
            <DialogDescription>{editingRole ? "Update custom role — system roles are locked" : "Create a company-scoped custom role. Name becomes JWT claim."}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="role_name">Role Name * <span className="text-xs font-normal text-muted-foreground">(lowercase, _ )</span></Label>
              <Input id="role_name" value={newName} onChange={(e)=> setNewName(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ""))} placeholder="e.g. hr_manager" />
              {newName && <p className="text-xs text-muted-foreground font-mono">claim: <Badge variant="outline">{newName || "—"}</Badge></p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="role_desc">Description</Label>
              <Textarea id="role_desc" value={newDesc} onChange={(e)=> setNewDesc(e.target.value)} placeholder="Optional — what this role can do" rows={2} />
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => { setCreateOpen(false); setEditingRole(null) }}>Cancel</Button>
            <Button onClick={handleCreateOrUpdate} disabled={saving || !newName.trim()}>
              {saving ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Saving...</> : editingRole ? "Save" : "Create"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <AlertDialog open={!!deleteTarget} onOpenChange={(o)=> !o && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete role “{deleteTarget?.name.replace(/_/g, " ")}”?</AlertDialogTitle>
            <AlertDialogDescription>Custom roles are company-scoped. Deleting removes it from all users. System roles cannot be deleted. This cannot be undone.</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={confirmDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">Delete</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
