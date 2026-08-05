"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface TimePickerProps {
  value?: string
  onChange?: (time: string) => void
  className?: string
  disabled?: boolean
  autoFocus?: boolean
}

function parseTimeString(v: string): { hh: string; mm: string } {
  if (!v) return { hh: "", mm: "" }
  if (v.includes("T")) {
    const t = v.slice(11, 16)
    const p = t.split(":")
    return { hh: p[0] || "", mm: p[1] || "" }
  }
  if (v.includes(":") && v.length <= 5) {
    const p = v.split(":")
    return { hh: p[0] || "", mm: p[1] || "" }
  }
  return { hh: "", mm: "" }
}

export function TimePicker({ value, onChange, disabled, className, autoFocus }: TimePickerProps) {
  const init = parseTimeString(value || "")
  const [hh, setHhState] = React.useState(init.hh)
  const [mm, setMmState] = React.useState(init.mm)
  const hhRef = React.useRef(init.hh)
  const mmRef = React.useRef(init.mm)

  const setHh = React.useCallback((v: string) => {
    hhRef.current = v
    setHhState(v)
  }, [])

  const setMm = React.useCallback((v: string) => {
    mmRef.current = v
    setMmState(v)
  }, [])

  const hhInputRef = React.useRef<HTMLInputElement>(null)
  const mmInputRef = React.useRef<HTMLInputElement>(null)

  React.useEffect(() => {
    const p = parseTimeString(value || "")
    hhRef.current = p.hh
    mmRef.current = p.mm
    setHhState(p.hh)
    setMmState(p.mm)
    if (autoFocus) {
      hhInputRef.current?.focus()
      hhInputRef.current?.select()
    }
  }, [value, autoFocus])

  const emit = React.useCallback((h: string, m: string) => {
    const hNum = parseInt(h, 10)
    const mNum = parseInt(m, 10)
    if (!isNaN(hNum) && !isNaN(mNum) && hNum >= 0 && hNum <= 23 && mNum >= 0 && mNum <= 59) {
      onChange?.(`${h.padStart(2, "0")}:${m.padStart(2, "0")}`)
    }
  }, [onChange])

  const onHhText = (v: string) => {
    const c = v.replace(/\D/g, "").slice(0, 2)
    const num = parseInt(c, 10)
    if (c === "" || (!isNaN(num) && num <= 23)) {
      setHh(c)
    }
    if (c.length === 2) {
      mmInputRef.current?.focus()
      mmInputRef.current?.select()
    }
  }

  const onMmText = (v: string) => {
    const c = v.replace(/\D/g, "").slice(0, 2)
    const num = parseInt(c, 10)
    if (c === "" || (!isNaN(num) && num <= 59)) {
      setMm(c)
    }
    if (c.length === 2) {
      emit(hhRef.current, c)
    }
  }

  const onHhBlur = () => emit(hhRef.current, mmRef.current)
  const onMmBlur = () => emit(hhRef.current, mmRef.current)

  const onHhKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === ":") {
      e.preventDefault()
      mmInputRef.current?.focus()
      mmInputRef.current?.select()
    }
    if (e.key === "Enter") emit(hhRef.current, mmRef.current)
  }

  const onMmKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") emit(hhRef.current, mmRef.current)
  }

  return (
    <div className={cn("relative flex items-center w-[124px]", className)}>
      <div className="flex items-center h-10 w-full rounded-lg border border-input bg-transparent text-sm focus-within:border-ring focus-within:ring-3 focus-within:ring-ring/50">
        <input
          ref={hhInputRef}
          type="text"
          inputMode="numeric"
          placeholder="hh"
          value={hh}
          onChange={(e) => onHhText(e.target.value)}
          onKeyDown={onHhKeyDown}
          onBlur={onHhBlur}
          onFocus={(e) => e.target.select()}
          disabled={disabled}
          className="w-10 h-full text-center bg-transparent outline-none placeholder:text-muted-foreground text-sm"
        />
        <span className="text-muted-foreground select-none font-medium">:</span>
        <input
          ref={mmInputRef}
          type="text"
          inputMode="numeric"
          placeholder="mm"
          value={mm}
          onChange={(e) => onMmText(e.target.value)}
          onKeyDown={onMmKeyDown}
          onBlur={onMmBlur}
          onFocus={(e) => e.target.select()}
          disabled={disabled}
          className="flex-1 h-full text-center bg-transparent outline-none placeholder:text-muted-foreground text-sm min-w-0"
        />
      </div>
    </div>
  )
}
