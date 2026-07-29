"use client"

import * as React from "react"

import { NavDocuments } from "@/components/layout/nav-documents"
import { NavMain } from "@/components/layout/nav-main"
import { NavSecondary } from "@/components/layout/nav-secondary"
import { NavUser } from "@/components/layout/nav-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import { NavGroup } from "./nav-group"
import { navMain, navGroup, navSecondary, documents } from "../data"
import { useSearchDialog } from "@/contexts/search-context"
import Link from "next/link"

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { setOpen } = useSearchDialog()

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:p-1.5! hover:bg-transparent!"
            >
              <Link href="/">
                <svg viewBox="0 0 32 32" className="size-6!" fill="none">
                  <defs>
                    <linearGradient id="ph-logo" x1="0" y1="0" x2="32" y2="32" gradientUnits="userSpaceOnUse">
                      <stop offset="0%" stopColor="#10b981" />
                      <stop offset="100%" stopColor="#9333ea" />
                    </linearGradient>
                  </defs>
                  <circle cx="16" cy="16" r="14" className="stroke-[1.5]" stroke="url(#ph-logo)" fill="url(#ph-logo)" fillOpacity="0.08" />
                  <circle cx="11" cy="11" r="3" fill="url(#ph-logo)" />
                  <circle cx="21" cy="11" r="3" fill="url(#ph-logo)" />
                  <circle cx="16" cy="22" r="3" fill="url(#ph-logo)" />
                  <path d="M11 14v3a2 2 0 0 0 2 2h1" className="stroke-[1.5]" stroke="url(#ph-logo)" strokeLinecap="round" />
                  <path d="M21 14v3a2 2 0 0 1-2 2h-1" className="stroke-[1.5]" stroke="url(#ph-logo)" strokeLinecap="round" />
                  <path d="M14 19l2 3 2-3" className="stroke-[1.5]" stroke="url(#ph-logo)" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
                <span className="text-base font-semibold bg-gradient-to-r from-emerald-500 to-purple-600 bg-clip-text text-transparent">People Hub</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain
          items={navMain.map((item) => ({
            ...item,
            icon: React.createElement(item.icon),
          }))}
        />
        <NavGroup items={navGroup} />
        <NavSecondary
          items={navSecondary.map((item) => ({
            ...item,
            icon: React.createElement(item.icon),
          }))}
          className="mt-auto"
          onSearchClick={() => setOpen(true)}
        />
      </SidebarContent>
    </Sidebar>
  )
}