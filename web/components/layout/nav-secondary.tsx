"use client"

import * as React from "react"

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar"

export function NavSecondary({
  items,
  onSearchClick,
  ...props
}: {
  items: {
    title: string
    url: string
    icon: React.ReactNode
    isSearch?: boolean
  }[]
  onSearchClick?: () => void
} & React.ComponentPropsWithoutRef<typeof SidebarGroup>) {
  const { isMobile, setOpenMobile } = useSidebar()

  const closeOnMobile = () => {
    if (isMobile) setOpenMobile(false)
  }

  return (
    <SidebarGroup {...props}>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton asChild>
                {item.isSearch ? (
                  <button type="button" onClick={onSearchClick}>
                    {item.icon}
                    <span>{item.title}</span>
                  </button>
                ) : (
                  <a href={item.url} onClick={closeOnMobile}>
                    {item.icon}
                    <span>{item.title}</span>
                  </a>
                )}
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}
