"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command"
import { navMain } from "@/components/data/nav-main"
import {
  informationNav,
  hrNav,
  attendanceNav,
  leaveNav,
  payrollNav,
  collectDataNav,
  monthlyReportNav,
  administratorNav,
} from "@/components/data/nav-group"
import { navSecondary } from "@/components/data/nav-secondary"
import { useSearchDialog } from "@/contexts/search-context"

export function SearchDialog() {
  const router = useRouter()
  const { open, setOpen } = useSearchDialog()

  const runCommand = React.useCallback(
    (url: string) => {
      setOpen(false)
      router.push(url)
    },
    [router, setOpen]
  )

  return (
    <CommandDialog open={open} onOpenChange={setOpen} showCloseButton>
      <CommandInput placeholder="Type a command or search..." />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        <CommandGroup heading="Main">
          {navMain.map((item) => (
            <CommandItem key={item.url} onSelect={() => runCommand(item.url)}>
              {React.createElement(item.icon, { className: "mr-2 h-4 w-4" })}
              <span>{item.title}</span>
            </CommandItem>
          ))}
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Information">
          {informationNav.map((item) => (
            <CommandItem key={item.url} onSelect={() => runCommand(item.url)}>
              {React.createElement(item.icon, { className: "mr-2 h-4 w-4" })}
              <span>{item.title}</span>
            </CommandItem>
          ))}
        </CommandGroup>
        <CommandGroup heading="Human Resource">
          {hrNav.map((item) => (
            <CommandItem key={item.url} onSelect={() => runCommand(item.url)}>
              {React.createElement(item.icon, { className: "mr-2 h-4 w-4" })}
              <span>{item.title}</span>
            </CommandItem>
          ))}
        </CommandGroup>
        <CommandGroup heading="Attendance">
          {attendanceNav.map((item) => (
            <CommandItem key={item.url} onSelect={() => runCommand(item.url)}>
              {React.createElement(item.icon, { className: "mr-2 h-4 w-4" })}
              <span>{item.title}</span>
            </CommandItem>
          ))}
        </CommandGroup>
        <CommandGroup heading="Leave">
          {leaveNav.map((item) => (
            <CommandItem key={item.url} onSelect={() => runCommand(item.url)}>
              {React.createElement(item.icon, { className: "mr-2 h-4 w-4" })}
              <span>{item.title}</span>
            </CommandItem>
          ))}
        </CommandGroup>
        <CommandGroup heading="Payroll">
          {payrollNav.map((item) => (
            <CommandItem key={item.url} onSelect={() => runCommand(item.url)}>
              {React.createElement(item.icon, { className: "mr-2 h-4 w-4" })}
              <span>{item.title}</span>
            </CommandItem>
          ))}
        </CommandGroup>
        <CommandGroup heading="Collect Data">
          {collectDataNav.map((item) => (
            <CommandItem key={item.url} onSelect={() => runCommand(item.url)}>
              {React.createElement(item.icon, { className: "mr-2 h-4 w-4" })}
              <span>{item.title}</span>
            </CommandItem>
          ))}
        </CommandGroup>
        <CommandGroup heading="Monthly Report">
          {monthlyReportNav.map((item) => (
            <CommandItem key={item.url} onSelect={() => runCommand(item.url)}>
              {React.createElement(item.icon, { className: "mr-2 h-4 w-4" })}
              <span>{item.title}</span>
            </CommandItem>
          ))}
        </CommandGroup>
        <CommandGroup heading="Administrator">
          {administratorNav.map((item) => (
            <CommandItem key={item.url} onSelect={() => runCommand(item.url)}>
              {React.createElement(item.icon, { className: "mr-2 h-4 w-4" })}
              <span>{item.title}</span>
            </CommandItem>
          ))}
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Other">
          {navSecondary.map((item) => (
            <CommandItem key={item.url} onSelect={() => runCommand(item.url)}>
              {React.createElement(item.icon, { className: "mr-2 h-4 w-4" })}
              <span>{item.title}</span>
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  )
}
