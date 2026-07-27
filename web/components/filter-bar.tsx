"use client"

import * as React from "react"
import { format } from "date-fns"
import { Loader2, RotateCcwIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { DatePicker } from "@/components/ui/date-picker"
import { DateRangePicker } from "@/components/ui/date-range-picker"

export interface FilterDef {
  key: string
  label: string
  type: "select" | "text" | "number" | "date" | "datepicker" | "daterange" | "daterange-split"
  options?: { value: string; label: string }[]
  placeholder?: string
  disabled?: boolean
  dateRangeKeys?: { start: string; end: string }
}

interface FilterBarProps {
  filters: FilterDef[]
  values: Record<string, string>
  onChange: (key: string, value: string) => void
  onApply: () => void
  onReset: () => void
  submitting: boolean
  singleColumn?: boolean
  noBorder?: boolean
}

const selectClass =
  "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:opacity-50 disabled:cursor-not-allowed"

export function FilterBar({ filters, values, onChange, onApply, onReset, submitting, singleColumn, noBorder }: FilterBarProps) {
  const hasFilters = Object.values(values).some((v) => v !== "")

  return (
    <div className={noBorder ? "p-4" : "rounded-lg border bg-card p-4"}>
      <div className={singleColumn ? "grid grid-cols-1 gap-3" : "grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3"}>
        {filters.map((f) => (
          <div key={f.key} className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-muted-foreground">{f.label}</label>
            {f.type === "select" ? (
              <select
                value={values[f.key] || ""}
                onChange={(e) => onChange(f.key, e.target.value)}
                className={selectClass}
                disabled={f.disabled}
              >
                <option value="">— All —</option>
                {f.options?.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            ) : f.type === "datepicker" ? (
              <DatePicker
                value={values[f.key] ? new Date(values[f.key] + "T00:00:00") : undefined}
                onChange={(date) => onChange(f.key, date ? format(date, "yyyy-MM-dd") : "")}
                placeholder={f.placeholder || "Pick a date"}
              />
            ) : f.type === "daterange" ? (
              <DateRangePicker
                value={{
                  from: values[f.dateRangeKeys?.start || "start_date"] ? new Date(values[f.dateRangeKeys?.start || "start_date"] + "T00:00:00") : undefined,
                  to: values[f.dateRangeKeys?.end || "end_date"] ? new Date(values[f.dateRangeKeys?.end || "end_date"] + "T00:00:00") : undefined,
                }}
                onChange={(range) => {
                  onChange(f.dateRangeKeys?.start || "start_date", range?.from ? format(range.from, "yyyy-MM-dd") : "")
                  onChange(f.dateRangeKeys?.end || "end_date", range?.to ? format(range.to, "yyyy-MM-dd") : "")
                }}
                placeholder={f.placeholder || "Pick a date range"}
              />
            ) : f.type === "daterange-split" ? (
              <div className="flex gap-2 items-center">
                <DatePicker
                  value={values[f.dateRangeKeys?.start || "start_date"] ? new Date(values[f.dateRangeKeys?.start || "start_date"] + "T00:00:00") : undefined}
                  onChange={(date) => onChange(f.dateRangeKeys?.start || "start_date", date ? format(date, "yyyy-MM-dd") : "")}
                  placeholder="Start"
                />
                <span className="text-muted-foreground text-sm">-</span>
                <DatePicker
                  value={values[f.dateRangeKeys?.end || "end_date"] ? new Date(values[f.dateRangeKeys?.end || "end_date"] + "T00:00:00") : undefined}
                  onChange={(date) => onChange(f.dateRangeKeys?.end || "end_date", date ? format(date, "yyyy-MM-dd") : "")}
                  placeholder="End"
                />
              </div>
            ) : (
              <input
                type={f.type}
                value={values[f.key] || ""}
                onChange={(e) => onChange(f.key, e.target.value)}
                placeholder={f.placeholder || `Filter by ${f.label.toLowerCase()}...`}
                min={f.type === "number" ? 0 : undefined}
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
            )}
          </div>
        ))}
      </div>
      <div className="flex items-center gap-2 mt-4">
        <Button onClick={onApply} disabled={submitting}>
          {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Apply
        </Button>
        {hasFilters && (
          <Button variant="outline" onClick={onReset} disabled={submitting}>
            <RotateCcwIcon className="mr-1 size-3.5" />
            Reset
          </Button>
        )}
      </div>
    </div>
  )
}
