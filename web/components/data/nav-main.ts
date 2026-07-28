import type { LucideIcon } from "lucide-react"
import { LayoutDashboardIcon, ListIcon, ChartBarIcon } from "lucide-react"

export const navMain = [
  {
    title: "Dashboard",
    url: "/dashboard",
    icon: LayoutDashboardIcon,
  },
  {
    title: "Lifecycle",
    url: "/lifecycle",
    icon: ListIcon,
  },
  {
    title: "Analytics",
    url: "/analytics",
    icon: ChartBarIcon,
  },
] as const satisfies {
  title: string
  url: string
  icon: LucideIcon
}[]