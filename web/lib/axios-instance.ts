import axios from "axios"
import { toast } from "sonner"
import { getApiBaseUrl, withBasePath } from "./utils"

let isRedirecting = false
let isRefreshing = false
type QueueItem = { resolve: (v: string) => void; reject: (e: unknown) => void }
let failedQueue: QueueItem[] = []

function processQueue(error: unknown, token: string | null = null) {
  failedQueue.forEach((prom) => {
    if (token) prom.resolve(token)
    else prom.reject(error)
  })
  failedQueue = []
}

function clearAuthAndRedirect() {
  if (isRedirecting) return
  isRedirecting = true
  try {
    localStorage.removeItem("access_token")
    localStorage.removeItem("refresh_token")
  } catch {}
  document.cookie = "auth_token=; path=/; max-age=0; SameSite=Lax"
  // ensure cookie cleared for HttpOnly case via extra path
  document.cookie = "auth_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax"
  window.location.href = withBasePath("/login")
}

const api = axios.create({
  baseURL: getApiBaseUrl(),
  withCredentials: true,
})

api.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("access_token")
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const url = error.config?.url || ""
    const status = error.response?.status

    // 429 is rate-limit — surface once and don't trigger refresh loop
    if (status === 429) {
      if (!error.config._retry429) {
        error.config._retry429 = true
        toast.error("Too many requests — please wait a moment")
      }
      return Promise.reject(error)
    }

    if (status !== 401) {
      if (status === 403) {
        const message = error.response?.data?.error || "You don't have permission to perform this action"
        toast.error(message)
      }
      return Promise.reject(error)
    }

    // Do not intercept login or register — let the calling code handle errors
    if (url.includes("/auth/login") || url.includes("/auth/register")) {
      return Promise.reject(error)
    }

    const isRefreshRequest = url.includes("/auth/refresh")
    if (isRefreshRequest) {
      clearAuthAndRedirect()
      return Promise.reject(error)
    }

    if (error.config._retry) {
      clearAuthAndRedirect()
      return Promise.reject(error)
    }

    const refreshToken = localStorage.getItem("refresh_token")
    if (!refreshToken) {
      clearAuthAndRedirect()
      return Promise.reject(error)
    }

    if (isRefreshing) {
      return new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject })
      })
        .then((newToken) => {
          error.config.headers.Authorization = `Bearer ${newToken}`
          error.config._retry = true
          return api(error.config)
        })
        .catch((err) => Promise.reject(err))
    }

    error.config._retry = true
    isRefreshing = true

    return axios
      .post(`${getApiBaseUrl()}/auth/refresh`, { refresh_token: refreshToken })
      .then((response) => {
        const { access_token, refresh_token: newRefresh } = response.data
        localStorage.setItem("access_token", access_token)
        localStorage.setItem("refresh_token", newRefresh)
        document.cookie = `auth_token=${access_token}; path=/; max-age=${7 * 24 * 60 * 60}; SameSite=Lax`
        api.defaults.headers.common.Authorization = `Bearer ${access_token}`
        processQueue(null, access_token)
        error.config.headers.Authorization = `Bearer ${access_token}`
        return api(error.config)
      })
      .catch((refreshError) => {
        processQueue(refreshError, null)
        clearAuthAndRedirect()
        return Promise.reject(refreshError)
      })
      .finally(() => {
        isRefreshing = false
      })
  }
)

export default api
