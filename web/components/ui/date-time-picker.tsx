"use client"

import * as React from "react"
import { format, isValid } from "date-fns"
import { cn } from "@/lib/utils"
import { DatePicker } from "@/components/ui/date-picker"
import { TimePicker } from "@/components/ui/time-picker"

function toIso(v: string): string {
  if (!v) return ""
  if (v.includes("T")) {
    return v.length === 16 ? v + ":00" : v
  }
  if (v.includes(" ")) {
    return v.replace(" ", "T")
  }
  return v + "T00:00:00"
}

function parseTime(v: string): string {
  if (!v) return ""
  if (v.includes("T")) return v.slice(11, 16)
  const spaceIdx = v.indexOf(" ")
  if (spaceIdx !== -1) {
    const timePart = v.slice(spaceIdx + 1)
    if (timePart.includes(":")) {
      const parts = timePart.split(":")
      if (parts[0].length <= 2) {
        return `${parts[0].padStart(2, "0")}:${parts[1]?.slice(0, 2).padStart(2, "0") || "00"}`
      }
    }
  }
  if (v.includes(":") && v.length <= 5) return v.slice(0, 5)
  return ""
}

interface DateTimePickerProps {
  value?: string
  onChange?: (value: string) => void
  className?: string
  disabled?: boolean
  autoFocusTime?: boolean
}

export function DateTimePicker({ value, onChange, className, disabled, autoFocusTime }: DateTimePickerProps) {
  const iso = value ? toIso(value) : ""
  const parsed = iso ? new Date(iso) : undefined
  const dateObj = parsed && isValid(parsed) ? parsed : undefined
  const dateStr = dateObj ? format(dateObj, "yyyy-MM-dd") : ""
  const timeStr = value ? parseTime(value) : ""

  const [dateVal, setDateVal] = React.useState(dateObj)
  const [timeVal, setTimeVal] = React.useState(timeStr)

  React.useEffect(() => {
    const i = value ? toIso(value) : ""
    const p = i ? new Date(i) : undefined
    setDateVal(p && isValid(p) ? p : undefined)
    setTimeVal(value ? parseTime(value) : "")
  }, [value])

  const emit = React.useCallback((d: Date | undefined, t: string) => {
    if (d && isValid(d)) {
      const ds = format(d, "yyyy-MM-dd")
      if (t) {
        onChange?.(`${ds}T${t}`)
      } else {
        onChange?.(ds)
      }
    } else if (t) {
      onChange?.(t)
    } else {
      onChange?.("")
    }
  }, [onChange])

  return (
    <div className={cn("flex items-center gap-1", className)}>
      <DatePicker
        value={dateVal}
        onChange={(d) => {
          setDateVal(d)
          emit(d, timeVal)
        }}
        disabled={disabled}
      />
      <TimePicker
        value={timeVal}
        onChange={(t) => {
          setTimeVal(t)
          emit(dateVal, t)
        }}
        disabled={disabled}
        autoFocus={autoFocusTime}
      />
    </div>
  )
}
