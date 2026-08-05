"use client"

import * as React from "react"
import { CalendarIcon, XIcon } from "lucide-react"
import { format, isValid } from "date-fns"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"

interface DatePickerProps {
  value?: Date | undefined
  onChange?: (date: Date | undefined) => void
  placeholder?: string
  className?: string
  disabled?: boolean
}

export function DatePicker({ value, onChange, placeholder = "Pick a date", className, disabled }: DatePickerProps) {
  const [open, setOpen] = React.useState(false)
  const [calendarMonth, setCalendarMonth] = React.useState<Date>(value || new Date())
  const [dd, setDd] = React.useState(value ? format(value, "dd") : "")
  const [mm, setMm] = React.useState(value ? format(value, "MM") : "")
  const [yyyy, setYyyy] = React.useState(value ? format(value, "yyyy") : "")

  const ddRef = React.useRef<HTMLInputElement>(null)
  const mmRef = React.useRef<HTMLInputElement>(null)
  const yyyyRef = React.useRef<HTMLInputElement>(null)

  React.useEffect(() => {
    setDd(value ? format(value, "dd") : "")
    setMm(value ? format(value, "MM") : "")
    setYyyy(value ? format(value, "yyyy") : "")
  }, [value])

  const tryComplete = (d: string, m: string, y: string) => {
    const day = parseInt(d, 10)
    const month = parseInt(m, 10)
    const year = parseInt(y, 10)
    if (day >= 1 && day <= 31 && month >= 1 && month <= 12 && year >= 1000 && year <= 9999) {
      const date = new Date(year, month - 1, day)
      if (isValid(date) && date.getDate() === day) {
        onChange?.(date)
      }
    }
  }

  const handleDdChange = (val: string) => {
    const cleaned = val.replace(/\D/g, "").slice(0, 2)
    setDd(cleaned)
    if (cleaned.length === 2) {
      mmRef.current?.focus()
      mmRef.current?.select()
    }
  }

  const handleMmChange = (val: string) => {
    const cleaned = val.replace(/\D/g, "").slice(0, 2)
    setMm(cleaned)
    if (cleaned.length === 2) {
      yyyyRef.current?.focus()
      yyyyRef.current?.select()
    }
  }

  const handleYyyyChange = (val: string) => {
    const cleaned = val.replace(/\D/g, "").slice(0, 4)
    setYyyy(cleaned)
    if (cleaned.length === 4) {
      tryComplete(dd, mm, cleaned)
    }
  }

  const handleDdBlur = () => tryComplete(dd, mm, yyyy)
  const handleMmBlur = () => tryComplete(dd, mm, yyyy)
  const handleYyyyBlur = () => tryComplete(dd, mm, yyyy)

  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation()
    onChange?.(undefined)
    setDd("")
    setMm("")
    setYyyy("")
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (nextOpen) {
      setCalendarMonth(value || new Date())
    }
    setOpen(nextOpen)
  }

  const displayValue = dd && mm && yyyy ? `${dd}/${mm}/${yyyy}` : ""

  return (
    <div className={cn("relative flex items-center", className)}>
      <div className="flex items-center h-10 w-48 rounded border border-input bg-transparent text-sm focus-within:border-ring focus-within:ring-3 focus-within:ring-ring/50">
        <input
          ref={ddRef}
          type="text"
          inputMode="numeric"
          placeholder="dd"
          value={dd}
          onChange={(e) => handleDdChange(e.target.value)}
          onBlur={handleDdBlur}
          onFocus={(e) => e.target.select()}
          disabled={disabled}
          className="w-10 h-full text-center bg-transparent outline-none placeholder:text-muted-foreground text-sm border-r border-input"
        />
        <input
          ref={mmRef}
          type="text"
          inputMode="numeric"
          placeholder="mm"
          value={mm}
          onChange={(e) => handleMmChange(e.target.value)}
          onBlur={handleMmBlur}
          onFocus={(e) => e.target.select()}
          disabled={disabled}
          className="w-10 h-full text-center bg-transparent outline-none placeholder:text-muted-foreground text-sm border-r border-input"
        />
        <input
          ref={yyyyRef}
          type="text"
          inputMode="numeric"
          placeholder="yyyy"
          value={yyyy}
          onChange={(e) => handleYyyyChange(e.target.value)}
          onBlur={handleYyyyBlur}
          onFocus={(e) => e.target.select()}
          disabled={disabled}
          className="w-14 h-full text-center bg-transparent outline-none placeholder:text-muted-foreground text-sm flex-1 min-w-0"
        />
        <div className="flex items-center pr-1 gap-0.5 shrink-0">
          {displayValue && (
            <button
              type="button"
              onClick={handleClear}
              className="p-1 text-muted-foreground hover:text-foreground"
            >
              <XIcon className="size-3.5" />
            </button>
          )}
          <Popover open={open} onOpenChange={handleOpenChange}>
            <PopoverTrigger asChild>
              <button
                type="button"
                disabled={disabled}
                className="p-1 text-muted-foreground hover:text-foreground disabled:pointer-events-none"
              >
                <CalendarIcon className="size-4" />
              </button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="end">
              <Calendar
                mode="single"
                selected={value}
                month={calendarMonth}
                onMonthChange={setCalendarMonth}
                onSelect={(date) => {
                  onChange?.(date)
                  setDd(date ? format(date, "dd") : "")
                  setMm(date ? format(date, "MM") : "")
                  setYyyy(date ? format(date, "yyyy") : "")
                  setOpen(false)
                }}
                autoFocus
              />
            </PopoverContent>
          </Popover>
        </div>
      </div>
    </div>
  )
}
