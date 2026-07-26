import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function getApiBaseUrl(): string {
  if (typeof window !== "undefined") {
    const port = window.location.port
    if (port === "3000" || port === "3050") {
      return process.env.NEXT_PUBLIC_API_URL || "http://localhost:5000/api/v1"
    }
    return window.location.protocol + "//" + window.location.host + "/api/v1"
  }
  return process.env.NEXT_PUBLIC_API_URL || "http://localhost:5000/api/v1"
}

export function getUploadBaseUrl(): string {
  if (typeof window !== "undefined") {
    const port = window.location.port
    if (port === "3000" || port === "3050") {
      return process.env.NEXT_PUBLIC_API_URL?.replace("/api/v1", "") || "http://localhost:5000"
    }
    return ""
  }
  return process.env.NEXT_PUBLIC_API_URL?.replace("/api/v1", "") || "http://localhost:5000"
}

export function formatCheck(val: string | null | undefined): string {
  if (!val) return "-"
  if (val.includes("T")) return val.slice(11, 16)
  if (val.length >= 19 && val[10] === " ") return val.slice(11, 16)
  if (val.length >= 5 && val[2] === ":") return val.slice(0, 5)
  return val
}
