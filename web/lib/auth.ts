// Utility to check if current user has a specific role
export function hasRole(role: string): boolean {
  if (typeof window === "undefined") return false
  try {
    const token = localStorage.getItem("access_token")
    if (!token) return false
    const payload = JSON.parse(atob(token.split(".")[1]))
    const roles: string[] = payload.roles || []
    return roles.includes(role) || roles.includes("super_admin")
  } catch {
    return false
  }
}

export function isSuperAdmin(): boolean {
  return hasRole("superadmin")
}
