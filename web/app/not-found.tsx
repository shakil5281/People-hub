"use client"

import * as React from "react"
import Link from "next/link"
import { HomeIcon, RefreshCwIcon } from "lucide-react"
import { Button } from "@/components/ui/button"

export default function NotFound() {
  return (
    <div className="flex min-h-[70vh] flex-col items-center justify-center gap-4 px-4 text-center">
      <h1 className="text-9xl font-bold tracking-tight text-muted-foreground/20">404</h1>
      <div className="space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Page not found</h2>
        <p className="text-muted-foreground max-w-md">
          Could not find the page you are looking for. It might have been removed, renamed, or
          did not exist in the first place.
        </p>
      </div>
      <div className="flex gap-2">
        <Button asChild>
          <Link href="/dashboard">
            <HomeIcon className="mr-2 h-4 w-4" />
            Go home
          </Link>
        </Button>
        <Button variant="outline" asChild>
          <Link href="/dashboard">
            <RefreshCwIcon className="mr-2 h-4 w-4" />
            Refresh
          </Link>
        </Button>
      </div>
    </div>
  )
}