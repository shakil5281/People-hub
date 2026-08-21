"use client"

import * as React from "react"
import { AlertTriangleIcon, RefreshCwIcon, HomeIcon } from "lucide-react"
import Link from "next/link"
import { Button } from "@/components/ui/button"

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <div className="flex min-h-[70vh] flex-col items-center justify-center gap-4 px-4 text-center">
      <AlertTriangleIcon className="h-16 w-16 text-destructive" />
      <div className="space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Something went wrong!</h2>
        <p className="text-muted-foreground max-w-md">
          {error.message || "An unexpected error occurred. Please try again."}
        </p>
        {error.digest && (
          <p className="text-sm font-mono text-muted-foreground">Error ID: {error.digest}</p>
        )}
      </div>
      <div className="flex gap-2">
        <Button onClick={reset}>
          <RefreshCwIcon className="mr-2 h-4 w-4" />
          Try again
        </Button>
        <Button variant="outline" asChild>
          <Link href="/dashboard">
            <HomeIcon className="mr-2 h-4 w-4" />
            Go home
          </Link>
        </Button>
      </div>
    </div>
  )
}