"use client"

import type { LucideIcon } from "lucide-react"
import { Loader2 } from "lucide-react"

interface CrudPageLayoutProps {
  icon: LucideIcon
  title: string
  description?: string
  loading?: boolean
  error?: string | null
  children: React.ReactNode
}

export function CrudPageLayout({ icon: Icon, title, description, loading, error, children }: CrudPageLayoutProps) {
  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6 flex items-center gap-2">
        <Icon className="h-6 w-6 text-muted-foreground" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{title}</h1>
          {description && <p className="text-muted-foreground mt-1">{description}</p>}
        </div>
      </div>
      {error && (
        <div className="px-4 lg:px-6">
          <div className="rounded-md bg-destructive/15 px-4 py-3 text-sm text-destructive">{error}</div>
        </div>
      )}
      <div className="px-4 lg:px-6">{children}</div>
    </div>
  )
}
