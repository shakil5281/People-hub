"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface CollapsibleContextValue {
  open: boolean
  onToggle: () => void
}

const CollapsibleContext = React.createContext<CollapsibleContextValue>({
  open: false,
  onToggle: () => {},
})

function Collapsible({
  open: controlledOpen,
  onOpenChange,
  children,
  asChild,
  className,
  ...props
}: React.ComponentProps<"div"> & {
  open?: boolean
  onOpenChange?: (open: boolean) => void
  asChild?: boolean
}) {
  const [internalOpen, setInternalOpen] = React.useState(false)
  const isControlled = controlledOpen !== undefined
  const open = isControlled ? controlledOpen : internalOpen

  const onToggle = React.useCallback(() => {
    if (isControlled) {
      onOpenChange?.(!controlledOpen)
    } else {
      setInternalOpen((prev) => !prev)
    }
  }, [isControlled, onOpenChange, controlledOpen])

  const contextValue = React.useMemo(() => ({ open, onToggle }), [open, onToggle])

  if (asChild && React.isValidElement(children)) {
    const child = React.Children.only(children) as React.ReactElement<any>
    return (
      <CollapsibleContext.Provider value={contextValue}>
        {React.cloneElement(child, {
          "data-slot": "collapsible",
          "data-state": open ? "open" : "closed",
          className: cn("group/collapsible", className, child.props.className),
          ...props,
        })}
      </CollapsibleContext.Provider>
    )
  }

  return (
    <div
      data-slot="collapsible"
      data-state={open ? "open" : "closed"}
      className={cn("group/collapsible", className)}
      {...props}
    >
      <CollapsibleContext.Provider value={contextValue}>
        {children}
      </CollapsibleContext.Provider>
    </div>
  )
}

function CollapsibleTrigger({
  children,
  asChild,
  className,
  ...props
}: React.ComponentProps<"button"> & {
  asChild?: boolean
}) {
  const { onToggle } = React.useContext(CollapsibleContext)

  if (asChild && React.isValidElement(children)) {
    const child = React.Children.only(children) as React.ReactElement<any>
    return React.cloneElement(child, {
      onClick: (e: React.MouseEvent) => {
        child.props.onClick?.(e)
        onToggle()
      },
      ...props,
    })
  }

  return (
    <button type="button" onClick={onToggle} className={cn(className)} {...props}>
      {children}
    </button>
  )
}

function CollapsibleContent({
  children,
  className,
  ...props
}: React.ComponentProps<"div">) {
  const { open } = React.useContext(CollapsibleContext)
  const contentRef = React.useRef<HTMLDivElement>(null)
  const [height, setHeight] = React.useState(0)

  React.useLayoutEffect(() => {
    if (contentRef.current) {
      setHeight(contentRef.current.scrollHeight)
    }
  }, [open])

  return (
    <div
      className={cn(
        "overflow-hidden transition-all duration-300 ease-in-out",
        className
      )}
      style={{ maxHeight: open ? height + 16 : 0 }}
      {...props}
    >
      <div ref={contentRef}>{children}</div>
    </div>
  )
}

export { Collapsible, CollapsibleContent, CollapsibleTrigger }
