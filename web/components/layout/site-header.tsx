"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { SidebarTrigger } from "@/components/ui/sidebar"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { authApi, notificationApi, companyApi } from "@/lib/api"
import {
  SearchIcon,
  SettingsIcon,
  LogOutIcon,
  Loader2,
  UserIcon,
  BellDotIcon,
} from "lucide-react"
import { toast } from "sonner"
import { useSearchDialog } from "@/contexts/search-context"

interface UserData {
  id: string
  name: string
  email: string
  roles?: { name: string }[]
}

export function SiteHeader() {
  const router = useRouter()
  const [user, setUser] = React.useState<UserData | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [loggingOut, setLoggingOut] = React.useState(false)
  const [notifCount, setNotifCount] = React.useState(0)
  const [companyName, setCompanyName] = React.useState("")

  React.useEffect(() => {
    authApi.me()
      .then((res) => setUser(res.data))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  React.useEffect(() => {
    const fetchCount = async () => {
      try {
        const { data } = await notificationApi.unreadCount()
        setNotifCount(data.count)
      } catch {}
    }
    fetchCount()
    const interval = setInterval(fetchCount, 30000)
    return () => clearInterval(interval)
  }, [])

  React.useEffect(() => {
    companyApi.list({ limit: "1" })
      .then((res) => {
        const companies = res.data?.data
        if (companies?.length > 0) setCompanyName(companies[0].company_name_en || "")
      })
      .catch(() => {})
  }, [])

  const handleLogout = async () => {
    setLoggingOut(true)
    try {
      await authApi.logout()
    } catch {
    } finally {
      localStorage.removeItem("access_token")
      localStorage.removeItem("refresh_token")
      document.cookie = "auth_token=; path=/; max-age=0"
      toast.success("Logged out successfully")
      router.push("/login")
    }
  }

  const { setOpen } = useSearchDialog()

  const initials = user?.name
    ? user.name.split(" ").map((n) => n[0]).join("").toUpperCase().slice(0, 2)
    : "U"

  return (
    <header className="sticky top-0 z-10 flex h-(--header-height) shrink-0 items-center gap-2 border-b border-border/50 bg-background/80 backdrop-blur-sm transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-(--header-height)">
      <div className="flex w-full items-center gap-2 px-4 lg:gap-3 lg:px-6">
        <SidebarTrigger className="-ml-1" />

        {/* <Separator orientation="vertical" className="mx-1 data-[orientation=vertical]:h-4" /> */}

        <div className="relative hidden sm:flex flex-1 max-w-sm">
          <SearchIcon className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search... (Ctrl+K)"
            className="pl-9 h-9 text-sm cursor-pointer rounded-2xl"
            onFocus={() => setOpen(true)}
            readOnly
          />
        </div>

        <div className="flex items-center gap-1 ml-auto">
          {companyName && (
            <div className="hidden md:flex items-center mr-1 pr-2 border-r border-border">
              <span className="text-sm font-semibold text-foreground truncate max-w-[180px]">{companyName}</span>
            </div>
          )}
          <Button variant="ghost" size="icon" className="relative h-9 w-9 group" onClick={() => router.push("/notifications")}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4 transition-all duration-300 group-hover:scale-110 group-hover:text-amber-500 group-hover:rotate-12">
              <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9" />
              <path d="M10.3 21a1.94 1.94 0 0 0 3.4 0" />
            </svg>
            {notifCount > 0 && (
              <span className="absolute -top-0.5 -right-0.5 h-4 min-w-4 px-1 text-[10px] font-bold flex items-center justify-center rounded-full bg-red-500 text-white animate-pulse">
                {notifCount > 99 ? "99+" : notifCount}
              </span>
            )}
          </Button>

          <Button variant="ghost" size="icon" className="h-9 w-9 group" onClick={() => router.push("/admin/settings")}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4 transition-all duration-500 group-hover:rotate-90 group-hover:text-primary">
              <circle cx="12" cy="12" r="3" />
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
            </svg>
          </Button>

          <Separator orientation="vertical" className="mx-1 h-6" />

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-9 w-9 rounded-full p-0">
                {loading ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Avatar className="h-8 w-8">
                    <AvatarFallback className="text-xs font-medium bg-primary/10 text-primary">
                      {initials}
                    </AvatarFallback>
                  </Avatar>
                )}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56" align="end" sideOffset={8}>
              <DropdownMenuLabel className="p-0 font-normal">
                <div className="flex items-center gap-2 px-1 py-1.5">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback className="text-xs font-medium bg-primary/10 text-primary">
                      {initials}
                    </AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">{user?.name || "User"}</span>
                    <span className="truncate text-xs text-muted-foreground">{user?.email || ""}</span>
                  </div>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuGroup>
                <DropdownMenuItem onClick={() => router.push("/profile")}>
                  <UserIcon className="mr-2 h-4 w-4" />
                  Profile
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => router.push("/admin/settings")}>
                  <SettingsIcon className="mr-2 h-4 w-4" />
                  Settings
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={handleLogout}
                disabled={loggingOut}
                className="text-destructive focus:text-destructive"
              >
                <LogOutIcon className="mr-2 h-4 w-4" />
                {loggingOut ? "Logging out..." : "Log out"}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  )
}
