"use client"

export interface User {
  id: string
  email: string
  name: string
  status: string
  created_at: string
}

export const statusOptions = [
  { value: "active", label: "Active" },
  { value: "deactivated", label: "Deactivated" },
  { value: "locked", label: "Locked" },
  { value: "pending", label: "Pending" },
]

export const statusBadgeVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  active: "default",
  deactivated: "secondary",
  locked: "destructive",
  pending: "outline",
  deleted: "destructive",
}
