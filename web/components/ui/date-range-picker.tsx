"use client"

import * as React from "react"
import { CalendarIcon } from "lucide-react"
import { format } from "date-fns"
import type { DateRange } from "react-day-picker"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"

interface DateRangePickerProps {
  value?: DateRange | undefined
  onChange?: (range: DateRange | undefined) => void
  placeholder?: string
  className?: string
  disabled?: boolean
}

export function DateRangePicker({ value, onChange, placeholder = "Pick a date range", className, disabled }: DateRangePickerProps) {
  const [open, setOpen] = React.useState(false)

  const displayValue =
    value?.from
      ? value.to
        ? `${format(value.from, "dd/MM/yyyy")} - ${format(value.to, "dd/MM/yyyy")}`
        : `${format(value.from, "dd/MM/yyyy")} - ...`
      : ""

  return (
    <div className={cn("relative", className)}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            disabled={disabled}
            className={cn(
              "h-10 w-64 justify-between text-left font-normal",
              !value?.from && "text-muted-foreground"
            )}
          >
            {displayValue || placeholder}
            <CalendarIcon className="ml-2 size-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="range"
            selected={value}
            onSelect={onChange}
            autoFocus
          />
        </PopoverContent>
      </Popover>
    </div>
  )
}
