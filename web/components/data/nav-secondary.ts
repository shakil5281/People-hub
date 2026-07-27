import type { LucideIcon } from "lucide-react"
import { Settings2Icon, CircleHelpIcon, SearchIcon, BellIcon } from "lucide-react"

export const navSecondary = [
  {
    title: "Notifications",
    url: "/notifications",
    icon: BellIcon,
  },
  {
    title: "Settings",
    url: "/admin/settings",
    icon: Settings2Icon,
  },
  {
    title: "Get Help",
    url: "/help",
    icon: CircleHelpIcon,
  },
  {
    title: "Search",
    url: "#search",
    icon: SearchIcon,
    isSearch: true,
  },
] as const satisfies {
  title: string
  url: string
  icon: LucideIcon
  isSearch?: boolean
}[]