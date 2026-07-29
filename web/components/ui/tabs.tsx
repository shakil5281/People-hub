"use client"

import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { Tabs as TabsPrimitive } from "radix-ui"

import { cn } from "@/lib/utils"

type TabsContextValue = {
  value: string
  onValueChange: (value: string) => void
}

const TabsContext = React.createContext<TabsContextValue | null>(null)

function useTabsContext() {
  const ctx = React.useContext(TabsContext)
  if (!ctx) throw new Error("Tabs compound components must be used within <Tabs>")
  return ctx
}

function Tabs({
  className,
  orientation = "horizontal",
  value,
  onValueChange,
  defaultValue,
  ...props
}: React.ComponentProps<typeof TabsPrimitive.Root>) {
  const [internalValue, setInternalValue] = React.useState(defaultValue ?? "")

  const ctxValue = React.useMemo(() => ({
    value: value ?? internalValue,
    onValueChange: (v: string) => {
      onValueChange?.(v)
      setInternalValue(v)
    },
  }), [value, internalValue, onValueChange])

  return (
    <TabsContext.Provider value={ctxValue}>
      <TabsPrimitive.Root
        data-slot="tabs"
        data-orientation={orientation}
        value={ctxValue.value}
        onValueChange={ctxValue.onValueChange}
        className={cn(
          "group/tabs flex gap-2 data-horizontal:flex-col",
          className
        )}
        {...props}
      />
    </TabsContext.Provider>
  )
}

const tabsListVariants = cva(
  "group/tabs-list relative inline-flex w-fit items-center justify-center rounded-none px-2 py-1 text-muted-foreground group-data-horizontal/tabs:h-fit group-data-vertical/tabs:h-fit group-data-vertical/tabs:flex-col data-[variant=line]:rounded-none",
  {
    variants: {
      variant: {
        default: "bg-muted",
        line: "gap-1 bg-transparent",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

function TabsList({
  className,
  variant = "default",
  ...props
}: React.ComponentProps<typeof TabsPrimitive.List> &
  VariantProps<typeof tabsListVariants>) {
  const listRef = React.useRef<HTMLDivElement>(null)
  const indicatorRef = React.useRef<HTMLDivElement>(null)
  const { value } = useTabsContext()
  const [indicatorStyle, setIndicatorStyle] = React.useState<React.CSSProperties>({})

  React.useEffect(() => {
    const list = listRef.current
    if (!list) return

    const activeTrigger = list.querySelector<HTMLElement>('[data-state="active"]')
    if (!activeTrigger) return

    const listRect = list.getBoundingClientRect()
    const triggerRect = activeTrigger.getBoundingClientRect()

    setIndicatorStyle({
      width: triggerRect.width,
      height: triggerRect.height,
      transform: `translateX(${triggerRect.left - listRect.left}px)`,
    })
  }, [value])

  const handleMouseEnter = React.useCallback((e: React.MouseEvent) => {
    const target = (e.target as HTMLElement).closest<HTMLElement>("[data-slot='tabs-trigger']")
    if (!target || target.getAttribute("data-state") === "active") return

    const list = listRef.current
    if (!list) return

    const listRect = list.getBoundingClientRect()
    const triggerRect = target.getBoundingClientRect()
    indicatorRef.current?.style.setProperty("transition-duration", "200ms")
    setIndicatorStyle({
      width: triggerRect.width,
      height: triggerRect.height,
      transform: `translateX(${triggerRect.left - listRect.left}px)`,
    })
  }, [])

  const handleMouseLeave = React.useCallback(() => {
    const list = listRef.current
    if (!list) return

    const activeTrigger = list.querySelector<HTMLElement>('[data-state="active"]')
    if (!activeTrigger) return

    const listRect = list.getBoundingClientRect()
    const triggerRect = activeTrigger.getBoundingClientRect()
    indicatorRef.current?.style.setProperty("transition-duration", "300ms")
    setIndicatorStyle({
      width: triggerRect.width,
      height: triggerRect.height,
      transform: `translateX(${triggerRect.left - listRect.left}px)`,
    })
  }, [])

  return (
    <TabsPrimitive.List
      ref={listRef}
      data-slot="tabs-list"
      data-variant={variant}
      className={cn(tabsListVariants({ variant }), "relative", className)}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      {...props}
    >
      <div
        ref={indicatorRef}
        className="pointer-events-none absolute top-[3px] left-[3px] z-0 rounded-none bg-background shadow-sm transition-all duration-300 ease-out will-change-transform"
        style={indicatorStyle}
      />
      {props.children}
    </TabsPrimitive.List>
  )
}

function TabsTrigger({
  className,
  ...props
}: React.ComponentProps<typeof TabsPrimitive.Trigger>) {
  return (
    <TabsPrimitive.Trigger
      data-slot="tabs-trigger"
      className={cn(
        "relative z-10 inline-flex h-9 flex-1 items-center justify-center gap-1.5 rounded-md border border-transparent px-4 text-sm font-medium whitespace-nowrap text-foreground/60 transition-colors group-data-vertical/tabs:w-full group-data-vertical/tabs:justify-start hover:text-foreground focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:outline-1 focus-visible:outline-ring disabled:pointer-events-none disabled:opacity-50 has-data-[icon=inline-end]:pr-1 has-data-[icon=inline-start]:pl-1 dark:text-muted-foreground dark:hover:text-foreground group-data-[variant=default]/tabs-list:data-active:shadow-none group-data-[variant=line]/tabs-list:data-active:shadow-none [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        "group-data-[variant=line]/tabs-list:bg-transparent group-data-[variant=line]/tabs-list:data-active:bg-transparent dark:group-data-[variant=line]/tabs-list:data-active:border-transparent dark:group-data-[variant=line]/tabs-list:data-active:bg-transparent",
        "data-active:text-foreground dark:data-active:text-foreground",
        "after:absolute after:bg-foreground after:opacity-0 after:transition-opacity group-data-horizontal/tabs:after:inset-x-0 group-data-horizontal/tabs:after:bottom-[-5px] group-data-horizontal/tabs:after:h-0.5 group-data-vertical/tabs:after:inset-y-0 group-data-vertical/tabs:after:-right-1 group-data-vertical/tabs:after:w-0.5 group-data-[variant=line]/tabs-list:data-active:after:opacity-100",
        className
      )}
      {...props}
    />
  )
}

function TabsContent({
  className,
  ...props
}: React.ComponentProps<typeof TabsPrimitive.Content>) {
  return (
    <TabsPrimitive.Content
      data-slot="tabs-content"
      className={cn("flex-1 text-sm outline-none", className)}
      {...props}
    />
  )
}

export { Tabs, TabsList, TabsTrigger, TabsContent, tabsListVariants }
