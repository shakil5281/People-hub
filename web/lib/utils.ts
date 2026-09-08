import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function getBasePath(): string {
  return process.env.NEXT_PUBLIC_BASE_PATH || ""
}

export function withBasePath(path: string): string {
  const base = getBasePath()
  if (!base || path.startsWith(base)) return path
  return `${base}${path}`
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
    return window.location.origin
  }
  return process.env.NEXT_PUBLIC_API_URL?.replace("/api/v1", "") || "http://localhost:5000"
}

export function resolveUploadUrl(val?: string): string {
  if (!val) return ""
  if (val.startsWith("http://") || val.startsWith("https://") || val.startsWith("blob:") || val.startsWith("data:")) {
    return val
  }
  const baseUrl = getUploadBaseUrl()
  const path = val.startsWith("/") ? val : `/${val}`
  // Respect NEXT_PUBLIC_BASE_PATH (e.g. /people-hub) when on same-origin
  const basePath = getBasePath()
  if (basePath && typeof window !== "undefined" && baseUrl === window.location.origin && !path.startsWith(basePath)) {
    return `${baseUrl}${withBasePath(path)}`
  }
  return `${baseUrl}${path}`
}

export function formatCheck(val: string | null | undefined): string {
  if (!val) return "-"
  if (val.includes("T")) return val.slice(11, 16)
  if (val.length >= 19 && val[10] === " ") return val.slice(11, 16)
  if (val.length >= 5 && val[2] === ":") return val.slice(0, 5)
  return val
}

export function downloadExport(res: { data: Blob | any }, filename: string) {
  const blob = res.data instanceof Blob ? res.data : new Blob([res.data], { type: res.data?.type || "application/octet-stream" })
  const url = URL.createObjectURL(blob)
  const a = document.createElement("a")
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

