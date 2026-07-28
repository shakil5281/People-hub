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

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { setOpen } = useSearchDialog()

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:p-1.5!"
            >
              <a href="/">
                <svg viewBox="0 0 24 24" className="size-5!" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <circle cx="9" cy="7" r="3" />
                  <circle cx="17" cy="7" r="3" />
                  <path d="M3 21v-2a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v2" />
                  <path d="M13 21v-2a4 4 0 0 1 4-4h2a4 4 0 0 1 2 2v2" />
                  <circle cx="12" cy="17" r="1" fill="currentColor" />
                </svg>
                <span className="text-base font-semibold">People Hub</span>
              </a>
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