import { toast } from "sonner"

export function handleApiError(err: unknown, fallback = "Operation failed"): string {
  const message = err instanceof Error ? err.message : fallback
  if (typeof err === "object" && err !== null && "response" in err) {
    const axiosErr = err as { response?: { data?: { error?: string } } }
    return axiosErr.response?.data?.error || message
  }
  return message
}

export function handleApiErrorWithToast(err: unknown, fallback = "Operation failed"): string {
  const detail = handleApiError(err, fallback)
  toast.error(detail)
  return detail
}
