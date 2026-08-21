import type { LucideIcon } from "lucide-react"
import { DatabaseIcon, FileChartColumnIcon, FileIcon } from "lucide-react"

export const documents = [
  {
    name: "Data Library",
    url: "/documents/library",
    icon: DatabaseIcon,
  },
  {
    name: "Reports",
    url: "/documents/reports",
    icon: FileChartColumnIcon,
  },
  {
    name: "Word Assistant",
    url: "/documents/word-assistant",
    icon: FileIcon,
  },
] as const satisfies {
  name: string
  url: string
  icon: LucideIcon
}[]