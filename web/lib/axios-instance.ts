import axios from "axios"
import { toast } from "sonner"
import { getApiBaseUrl } from "./utils"

let isRedirecting = false

function clearAuthAndRedirect() {
  if (isRedirecting) return
  isRedirecting = true
  localStorage.removeItem("access_token")
  localStorage.removeItem("refresh_token")
  document.cookie = "auth_token=; path=/; max-age=0"
  window.location.href = "/login"
}

const api = axios.create({
  baseURL: getApiBaseUrl(),
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
    if (error.response?.status !== 401) {
      if (error.response?.status === 403) {
        const message = error.response?.data?.error || "You don't have permission to perform this action"
        toast.error(message)
      }
      return Promise.reject(error)
    }

    const isRefreshRequest = error.config.url?.includes("/auth/refresh")
    if (isRefreshRequest) {
      clearAuthAndRedirect()
      return Promise.reject(error)
    }

    if (error.config._retry) {
      clearAuthAndRedirect()
      return Promise.reject(error)
    }

    error.config._retry = true

    const refreshToken = localStorage.getItem("refresh_token")
    if (!refreshToken) {
      clearAuthAndRedirect()
      return Promise.reject(error)
    }

    return axios
      .post(`${getApiBaseUrl()}/auth/refresh`, { refresh_token: refreshToken })
      .then((response) => {
        const { access_token, refresh_token } = response.data
        localStorage.setItem("access_token", access_token)
        localStorage.setItem("refresh_token", refresh_token)
        document.cookie = `auth_token=${access_token}; path=/; max-age=${7 * 24 * 60 * 60}; SameSite=Lax`

        error.config.headers.Authorization = `Bearer ${access_token}`
        return api(error.config)
      })
      .catch(() => {
        clearAuthAndRedirect()
        return Promise.reject(error)
      })
  }
)

export default api
