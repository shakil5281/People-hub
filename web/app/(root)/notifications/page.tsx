"use client"

import * as React from "react"
import { BellIcon, CheckCheckIcon, InfoIcon, AlertTriangleIcon, AlertCircleIcon, XIcon, Loader2 } from "lucide-react"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet"
import { cn } from "@/lib/utils"
import { notificationApi, type Notification } from "@/lib/api"
import { toast } from "sonner"

const typeConfig = {
  info: { icon: InfoIcon, color: "text-blue-500 bg-blue-50 dark:bg-blue-950/30" },
  warning: { icon: AlertTriangleIcon, color: "text-amber-500 bg-amber-50 dark:bg-amber-950/30" },
  error: { icon: AlertCircleIcon, color: "text-red-500 bg-red-50 dark:bg-red-950/30" },
  success: { icon: CheckCheckIcon, color: "text-emerald-500 bg-emerald-50 dark:bg-emerald-950/30" },
}

type FilterTab = "all" | "unread" | "read"

const typeVariant: Record<string, "destructive" | "secondary" | "default" | "outline"> = {
  error: "destructive",
  warning: "secondary",
  success: "default",
  info: "outline",
}

function relativeTime(dateStr: string) {
  const now = Date.now()
  const d = new Date(dateStr).getTime()
  const diff = now - d
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return "just now"
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  const days = Math.floor(hrs / 24)
  if (days < 7) return `${days}d ago`
  return new Date(dateStr).toLocaleDateString("en-GB")
}

export default function NotificationsPage() {
  const [notifications, setNotifications] = React.useState<Notification[]>([])
  const [loading, setLoading] = React.useState(true)
  const [filter, setFilter] = React.useState<FilterTab>("all")
  const [page, setPage] = React.useState(1)
  const [totalPages, setTotalPages] = React.useState(0)
  const [total, setTotal] = React.useState(0)

  const fetchData = React.useCallback(async (p?: number) => {
    setLoading(true)
    try {
      const params: Record<string, string> = { page: String(p ?? page), limit: "20" }
      if (filter === "unread") params.is_read = "false"
      else if (filter === "read") params.is_read = "true"
      const { data: res } = await notificationApi.list(params)
      setNotifications(Array.isArray(res.data) ? res.data : [])
      setTotal(res.total ?? 0)
      setTotalPages(res.total_pages ?? 0)
    } catch {
      setNotifications([])
    } finally {
      setLoading(false)
    }
  }, [page, filter])

  React.useEffect(() => {
    fetchData()
  }, [fetchData])

  React.useEffect(() => {
    setPage(1)
  }, [filter])

  const unreadCount = React.useMemo(() => notifications.filter((n) => !n.is_read).length, [notifications])

  const handleMarkAsRead = async (id: string) => {
    try {
      await notificationApi.markAsRead(id)
      setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, is_read: true, read_at: new Date().toISOString() } : n)))
    } catch {
      toast.error("Failed to mark as read")
    }
  }

  const handleMarkAllAsRead = async () => {
    try {
      await notificationApi.markAllAsRead()
      setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true, read_at: new Date().toISOString() })))
      toast.success("All marked as read")
    } catch {
      toast.error("Failed to mark all as read")
    }
  }

  const handleDismiss = async (id: string) => {
    try {
      await notificationApi.delete(id)
      setNotifications((prev) => prev.filter((n) => n.id !== id))
    } catch {
      toast.error("Failed to dismiss")
    }
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <BellIcon className="h-6 w-6 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Notifications</h1>
            <p className="text-muted-foreground mt-1">Stay updated with the latest activities</p>
          </div>
        </div>
      </div>

      <div className="px-4 lg:px-6">
        <Tabs value={filter} onValueChange={(v) => setFilter(v as FilterTab)}>
          <div className="flex items-center justify-between">
            <TabsList>
              <TabsTrigger value="all">
                All
                <Badge variant="secondary" className="ml-1.5 text-[10px] px-1.5">{total}</Badge>
              </TabsTrigger>
              <TabsTrigger value="unread">
                Unread
                {unreadCount > 0 && (
                  <Badge variant="default" className="ml-1.5 text-[10px] px-1.5">{unreadCount}</Badge>
                )}
              </TabsTrigger>
              <TabsTrigger value="read">Read</TabsTrigger>
            </TabsList>
            <div className="hidden md:block">
              {unreadCount > 0 && (
                <Button variant="outline" size="sm" onClick={handleMarkAllAsRead}>
                  <CheckCheckIcon className="mr-1.5 h-4 w-4" />
                  Mark all as read
                </Button>
              )}
            </div>
            <div className="md:hidden">
              {unreadCount > 0 && (
                <Sheet>
                  <SheetTrigger asChild>
                    <Button variant="outline" size="sm">
                      <CheckCheckIcon className="h-4 w-4" />
                    </Button>
                  </SheetTrigger>
                  <SheetContent side="right" showCloseButton>
                    <SheetHeader>
                      <SheetTitle>Actions</SheetTitle>
                    </SheetHeader>
                    <div className="flex flex-col gap-2 mt-4">
                      <Button variant="outline" onClick={handleMarkAllAsRead} className="w-full">
                        <CheckCheckIcon className="mr-2 h-4 w-4" />
                        Mark all as read
                      </Button>
                    </div>
                  </SheetContent>
                </Sheet>
              )}
            </div>
          </div>

          <TabsContent value={filter} className="mt-4">
            {loading ? (
              <div className="flex items-center justify-center py-16 text-muted-foreground">
                <Loader2 className="h-6 w-6 animate-spin mr-2" />
                Loading...
              </div>
            ) : notifications.length === 0 ? (
              <Card>
                <CardContent className="flex flex-col items-center gap-3 py-12">
                  <BellIcon className="h-10 w-10 text-muted-foreground/40" />
                  <p className="text-sm text-muted-foreground">
                    {filter === "all"
                      ? "No notifications yet"
                      : filter === "unread"
                      ? "All caught up! No unread notifications"
                      : "No read notifications"}
                  </p>
                </CardContent>
              </Card>
            ) : (
              <div className="space-y-2">
                {notifications.map((n) => {
                  const config = typeConfig[n.type as keyof typeof typeConfig] || typeConfig.info
                  const Icon = config.icon
                  return (
                    <Card key={n.id} className={cn("transition-colors", !n.is_read && "border-l-2 border-l-primary bg-muted/30")}>
                      <CardContent className="flex items-start gap-3 p-4">
                        <div className={cn("flex h-9 w-9 shrink-0 items-center justify-center rounded-full", config.color)}>
                          <Icon className="h-4 w-4" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-2">
                            <p className="text-sm font-medium truncate">{n.title}</p>
                            <div className="flex items-center gap-2 shrink-0">
                              {!n.is_read && <span className="h-2 w-2 shrink-0 rounded-full bg-primary" />}
                              <Badge variant={typeVariant[n.type] || "outline"} className="text-[10px] capitalize">{n.type}</Badge>
                              <span className="text-xs text-muted-foreground/60">{relativeTime(n.created_at)}</span>
                            </div>
                          </div>
                          <p className="text-sm text-muted-foreground mt-0.5">{n.message}</p>
                        </div>
                        <div className="flex shrink-0 gap-1">
                          {!n.is_read && (
                            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleMarkAsRead(n.id)} title="Mark as read">
                              <CheckCheckIcon className="h-3.5 w-3.5 text-muted-foreground" />
                            </Button>
                          )}
                          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleDismiss(n.id)} title="Dismiss">
                            <XIcon className="h-3.5 w-3.5 text-muted-foreground" />
                          </Button>
                        </div>
                      </CardContent>
                    </Card>
                  )
                })}
                {totalPages > 1 && (
                  <div className="flex justify-center gap-2 pt-2">
                    <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => Math.max(1, p - 1))}>
                      Previous
                    </Button>
                    <span className="flex items-center text-sm text-muted-foreground px-2">
                      Page {page} of {totalPages}
                    </span>
                    <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
                      Next
                    </Button>
                  </div>
                )}
              </div>
            )}
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
