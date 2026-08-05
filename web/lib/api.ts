import api from "./axios-instance"

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  name: string
  email: string
  password: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: {
    id: string
    email: string
    name: string
    status: string
    force_password_change: boolean
  }
}

export const authApi = {
  login: (data: LoginRequest) => api.post<AuthResponse>("/auth/login", data),
  register: (data: RegisterRequest) => api.post<AuthResponse>("/auth/register", data),
  logout: () => api.post("/auth/logout"),
  me: () => api.get("/auth/me"),
  updateProfile: (data: Record<string, unknown>) => api.put("/auth/profile", data),
  changePassword: (data: { current_password: string; new_password: string }) => api.put("/auth/change-password", data),
  sessions: () => api.get("/auth/sessions"),
  logoutAll: () => api.post("/auth/logout-all"),
  forgotPassword: (data: { email: string }) => api.post("/auth/forgot-password", data),
  resetPassword: (data: { token: string; new_password: string }) => api.post("/auth/reset-password", data),
}

export const userApi = {
  list: (params?: Record<string, string>) => api.get<PaginatedResponse<UserListItem>>("/users", { params }),
  get: (id: string) => api.get<{ id: string; email: string; name: string; status: string; roles: Array<{ id: string; name: string }> }>(`/users/${id}`),
  create: (data: { email: string; name: string; role_ids?: string[] }) => api.post("/users", data),
  update: (id: string, data: { name?: string; status?: string }) => api.put(`/users/${id}`, data),
  delete: (id: string) => api.delete(`/users/${id}`),
  getRoles: (id: string) => api.get<{ data: Array<{ id: string; name: string; description: string }> }>(`/users/${id}/roles`),
  assignRoles: (id: string, data: { role_ids: string[] }) => api.put(`/users/${id}/roles`, data),
  resetPassword: (id: string) => api.post<{ generated_password: string; force_password_change: boolean; message: string }>(`/users/${id}/reset-password`),
}

export interface UserListItem {
  id: string
  email: string
  name: string
  status: string
  created_at: string
}

export const roleApi = {
  list: (companyId?: string) => api.get("/roles", { params: companyId ? { company_id: companyId } : undefined }),
  get: (id: string) => api.get(`/roles/${id}`),
  create: (data: { name: string; description?: string }) => api.post("/roles", data),
  update: (id: string, data: { name?: string; description?: string }) => api.put(`/roles/${id}`, data),
  delete: (id: string) => api.delete(`/roles/${id}`),
  assignPermissions: (id: string, data: { permission_ids: string[] }) => api.put(`/roles/${id}/permissions`, data),
}

export const permissionApi = {
  list: () => api.get("/permissions"),
}

export const settingsApi = {
  list: () => api.get("/settings"),
  update: (data: { settings: Record<string, string> }) => api.put("/settings", data),
}

export const companyApi = {
  list: (params?: Record<string, string>) => api.get("/companies", { params }),
  get: (id: string) => api.get(`/companies/${id}`),
  create: (data: Record<string, unknown>) => api.post("/companies", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/companies/${id}`, data),
  delete: (id: string) => api.delete(`/companies/${id}`),
}

export const employeeApi = {
  list: (params?: Record<string, string>) => api.get("/employees", { params }),
  get: (id: string) => api.get(`/employees/${id}`),
  getByCode: (code: string) => api.get(`/employees/by-code/${code}`),
  getProfile: (id: string, params?: { month?: number; year?: number }) =>
    api.get(`/employees/${id}/profile`, { params }),
  exportProfileExcel: (id: string, params?: { month?: number; year?: number }) =>
    api.get(`/employees/${id}/profile/export/excel`, { params, responseType: "blob" }),
  exportProfilePdf: (id: string, params?: { month?: number; year?: number }) =>
    api.get(`/employees/${id}/profile/export/pdf`, { params, responseType: "blob" }),
  create: (data: Record<string, unknown>) => api.post("/employees", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/employees/${id}`, data),
  delete: (id: string) => api.delete(`/employees/${id}`),
  importExcel: (file: File) => {
    const formData = new FormData()
    formData.append("file", file)
    return api.post("/employees/import", formData)
  },
  downloadTemplate: () => api.get("/employees/import/template", { responseType: "blob" }),
  exportExcel: (params?: Record<string, string>) => api.get("/employees/export/excel", { params, responseType: "blob" }),
  exportPdf: (params?: Record<string, string>) => api.get("/employees/export/pdf", { params, responseType: "blob" }),
}

export const groupApi = {
  list: (params?: Record<string, string>) => api.get("/groups", { params }),
  get: (id: string) => api.get(`/groups/${id}`),
  create: (data: Record<string, unknown>) => api.post("/groups", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/groups/${id}`, data),
  delete: (id: string) => api.delete(`/groups/${id}`),
}

export const floorApi = {
  list: (params?: Record<string, string>) => api.get("/floors", { params }),
  get: (id: string) => api.get(`/floors/${id}`),
  create: (data: Record<string, unknown>) => api.post("/floors", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/floors/${id}`, data),
  delete: (id: string) => api.delete(`/floors/${id}`),
}

export const departmentApi = {
  list: (params?: Record<string, string>) => api.get("/departments", { params }),
  get: (id: string) => api.get(`/departments/${id}`),
  create: (data: Record<string, unknown>) => api.post("/departments", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/departments/${id}`, data),
  delete: (id: string) => api.delete(`/departments/${id}`),
}

export const sectionApi = {
  list: (departmentId?: string, params?: Record<string, string>) => api.get("/sections", { params: departmentId ? { ...params, department_id: departmentId } : params }),
  get: (id: string) => api.get(`/sections/${id}`),
  create: (data: Record<string, unknown>) => api.post("/sections", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/sections/${id}`, data),
  delete: (id: string) => api.delete(`/sections/${id}`),
}

export const designationApi = {
  list: (sectionId?: string, params?: Record<string, string>) => api.get("/designations", { params: sectionId ? { ...params, section_id: sectionId } : params }),
  get: (id: string) => api.get(`/designations/${id}`),
  create: (data: Record<string, unknown>) => api.post("/designations", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/designations/${id}`, data),
  delete: (id: string) => api.delete(`/designations/${id}`),
}

export const lineApi = {
  list: (sectionId?: string, params?: Record<string, string>) => api.get("/lines", { params: sectionId ? { ...params, section_id: sectionId } : params }),
  get: (id: string) => api.get(`/lines/${id}`),
  create: (data: Record<string, unknown>) => api.post("/lines", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/lines/${id}`, data),
  delete: (id: string) => api.delete(`/lines/${id}`),
}

export const databaseApi = {
  backup: () => api.post("/database/backup"),
  listBackups: () => api.get("/database/backups"),
  export: (filename: string) => api.get("/database/export", { params: { filename }, responseType: "blob" }),
  deleteBackup: (filename: string) => api.delete("/database/backups", { params: { filename } }),
  importSql: (file: File) => {
    const formData = new FormData()
    formData.append("file", file)
    return api.post("/database/import", formData)
  },
  reset: () => api.post("/database/reset"),
}

export const dashboardApi = {
  stats: () => api.get("/dashboard/stats"),
}

export const organizationApi = {
  downloadTemplate: () => api.get("/organization/template", { responseType: "blob" }),
  importExcel: (file: File) => {
    const formData = new FormData()
    formData.append("file", file)
    return api.post("/organization/import", formData)
  },
}

export const shiftApi = {
  list: (params?: Record<string, string>) => api.get("/shifts", { params }),
  get: (id: string) => api.get(`/shifts/${id}`),
  create: (data: Record<string, unknown>) => api.post("/shifts", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/shifts/${id}`, data),
  delete: (id: string) => api.delete(`/shifts/${id}`),
}

export const attendanceApi = {
  list: (params?: Record<string, string>) => api.get("/attendance", { params }),
  get: (id: string) => api.get(`/attendance/${id}`),
  create: (data: Record<string, unknown>) => api.post("/attendance", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/attendance/${id}`, data),
  delete: (id: string) => api.delete(`/attendance/${id}`),
  clockIn: (data: Record<string, unknown>) => api.post("/attendance/clock-in", data),
  clockOut: (data: Record<string, unknown>) => api.post("/attendance/clock-out", data),
  jobCard: (params?: Record<string, string>) => api.get("/attendance/job-card", { params }),
  exportJobCardPdf: (params?: Record<string, string>) => api.get("/attendance/job-card/export", { params: { ...params, format: "pdf" }, responseType: "blob" }),
  exportJobCardExcel: (params?: Record<string, string>) => api.get("/attendance/job-card/export", { params: { ...params, format: "excel" }, responseType: "blob" }),
  stats: () => api.get("/attendance/stats"),
  deleteAll: () => api.delete("/attendance/delete-all"),
  missing: (params: Record<string, string>) => api.get("/attendance/missing", { params }),
  absent: (params: Record<string, string>) => api.get("/attendance/absent", { params }),
  exportAbsentExcel: (params?: Record<string, string>) => api.get("/attendance/absent/export/excel", { params, responseType: "blob" }),
  exportMissingExcel: (params?: Record<string, string>) => api.get("/attendance/missing/export/excel", { params, responseType: "blob" }),
  summary: (params?: Record<string, string>) => api.get("/attendance/summary", { params }),
  exportSummaryExcel: (params?: Record<string, string>) => api.get("/attendance/summary/export/excel", { params, responseType: "blob" }),
  overtime: (params?: Record<string, string>) => api.get("/attendance/overtime", { params }),
  exportOvertimeExcel: (params?: Record<string, string>) => api.get("/attendance/overtime/export/excel", { params, responseType: "blob" }),
  overtimeSummary: (params?: Record<string, string>) => api.get("/attendance/overtime-summary", { params }),
  exportOvertimeSummaryExcel: (params?: Record<string, string>) => api.get("/attendance/overtime-summary/export/excel", { params, responseType: "blob" }),
  monthlyReport: (params?: Record<string, string>) => api.get("/attendance/monthly-report", { params }),
  exportMonthlyReportExcel: (params?: Record<string, string>) => api.get("/attendance/monthly-report/export/excel", { params, responseType: "blob" }),
  exportMonthlyReportPdf: (params?: Record<string, string>) => api.get("/attendance/monthly-report/export/pdf", { params, responseType: "blob" }),
  customSummary: (data: { company_id: string; date: string; sections: Record<string, unknown>[] }) => api.post("/attendance/custom-summary?company_id=" + data.company_id + "&date=" + data.date, data.sections),
  exportExcel: (params?: Record<string, string>) => api.get("/attendance/export/excel", { params, responseType: "blob" }),
  bulkUpdateMissing: (data: { status?: string; inTime?: string; outTime?: string; attendanceIds: string[] }) => api.post("/attendance/bulk-update-missing", data),
}

export const otEarlyExitApi = {
  process: (data: { company_id: string; month: number; year: number }) => api.post("/attendance/ot-early-exit/process", data),
  list: (params: Record<string, string>) => api.get("/attendance/ot-early-exit", { params }),
  exportExcel: (params?: Record<string, string>) => api.get("/attendance/ot-early-exit/export/excel", { params, responseType: "blob" }),
}

export const missingAttendanceApi = {
  list: (params: Record<string, string>) => api.get("/missing-attendance", { params }),
  create: (data: Record<string, unknown>) => api.post("/missing-attendance", data),
  upsert: (data: Record<string, unknown>) => api.post("/missing-attendance/upsert", data),
  bulk: (data: { records: Record<string, unknown>[] }) => api.post("/missing-attendance/bulk", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/missing-attendance/${id}`, data),
  delete: (id: string) => api.delete(`/missing-attendance/${id}`),
}

export const dataLogApi = {
  import: (data: { file_path?: string; start_date?: string; end_date?: string }) => api.post("/data-logs/import", data),
  list: (params?: Record<string, string>) => api.get("/data-logs", { params }),
  process: (data: { date?: string; start_date?: string; end_date?: string; company_id: string }) => api.post("/data-logs/process", data),
  stats: () => api.get("/data-logs/stats"),
  deleteAll: () => api.delete("/data-logs/delete-all"),
}

export const divisionApi = {
  list: (params?: Record<string, string>) => api.get("/divisions", { params }),
  get: (id: string) => api.get(`/divisions/${id}`),
  create: (data: Record<string, unknown>) => api.post("/divisions", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/divisions/${id}`, data),
  delete: (id: string) => api.delete(`/divisions/${id}`),
}

export const districtApi = {
  list: (divisionId?: string, params?: Record<string, string>) => api.get("/districts", { params: divisionId ? { ...params, division_id: divisionId } : params }),
  get: (id: string) => api.get(`/districts/${id}`),
  create: (data: Record<string, unknown>) => api.post("/districts", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/districts/${id}`, data),
  delete: (id: string) => api.delete(`/districts/${id}`),
}

export const upazilaApi = {
  list: (districtId?: string, params?: Record<string, string>) => api.get("/upazilas", { params: districtId ? { ...params, district_id: districtId } : params }),
  get: (id: string) => api.get(`/upazilas/${id}`),
  create: (data: Record<string, unknown>) => api.post("/upazilas", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/upazilas/${id}`, data),
  delete: (id: string) => api.delete(`/upazilas/${id}`),
}

export const unionApi = {
  list: (upazilaId?: string, params?: Record<string, string>) => api.get("/unions", { params: upazilaId ? { ...params, upazila_id: upazilaId } : params }),
  get: (id: string) => api.get(`/unions/${id}`),
  create: (data: Record<string, unknown>) => api.post("/unions", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/unions/${id}`, data),
  delete: (id: string) => api.delete(`/unions/${id}`),
}

export const requirementApi = {
  list: (params?: Record<string, string>) => api.get("/requirements", { params }),
  get: (id: string) => api.get(`/requirements/${id}`),
  create: (data: Record<string, unknown>) => api.post("/requirements", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/requirements/${id}`, data),
  delete: (id: string) => api.delete(`/requirements/${id}`),
  sectionSummary: () => api.get("/requirements/section-summary"),
}

export const separationApi = {
  list: (params?: Record<string, string>) => api.get("/separations", { params }),
  get: (id: string) => api.get(`/separations/${id}`),
  create: (data: Record<string, unknown>) => api.post("/separations", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/separations/${id}`, data),
  delete: (id: string) => api.delete(`/separations/${id}`),
  process: (date?: string) => api.post("/separations/process" + (date ? `?date=${date}` : "")),
  processOne: (id: string) => api.post(`/separations/${id}/process`),
  cancel: (id: string) => api.post(`/separations/${id}/cancel`),
}

export const idCardApi = {
  list: (params?: Record<string, string>) => api.get("/id-cards", { params }),
  get: (id: string) => api.get(`/id-cards/${id}`),
  create: (data: Record<string, unknown>) => api.post("/id-cards", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/id-cards/${id}`, data),
  delete: (id: string) => api.delete(`/id-cards/${id}`),
  generate: (employeeIds: string[]) => api.post("/id-cards/generate", { employee_ids: employeeIds }),
}

export const leaveTypeApi = {
  list: (companyId?: string, params?: Record<string, string>) => api.get("/leave-types", { params: companyId ? { ...params, company_id: companyId } : params }),
  get: (id: string) => api.get(`/leave-types/${id}`),
  create: (data: Record<string, unknown>) => api.post("/leave-types", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/leave-types/${id}`, data),
  delete: (id: string) => api.delete(`/leave-types/${id}`),
}

export const leaveApi = {
  list: (params?: Record<string, string>) => api.get("/leaves", { params }),
  get: (id: string) => api.get(`/leaves/${id}`),
  apply: (data: Record<string, unknown>) => api.post("/leaves", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/leaves/${id}`, data),
  cancel: (id: string) => api.delete(`/leaves/${id}`),
  approve: (id: string) => api.put(`/leaves/${id}/approve`),
  reject: (id: string, reason: string) => api.put(`/leaves/${id}/reject`, { rejection_reason: reason }),
  exportFormPdf: (id: string, lang?: string) => api.get(`/leaves/${id}/export/pdf`, { params: { lang }, responseType: "blob" }),
}

export const leaveBalanceApi = {
  list: (params?: Record<string, string>) => api.get("/leave-balance", { params }),
}

export const leaveReportApi = {
  monthly: (params?: Record<string, string>) => api.get("/leave-reports/monthly", { params }),
}

export const holidayApi = {
  list: (params?: Record<string, string>) => api.get("/holidays", { params }),
  get: (id: string) => api.get(`/holidays/${id}`),
  create: (data: Record<string, unknown>) => api.post("/holidays", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/holidays/${id}`, data),
  delete: (id: string) => api.delete(`/holidays/${id}`),
}

export const salaryApi = {
  process: (data: { company_id: string; month: number; year: number }) => api.post("/salary/process", data),
  sheet: (params?: Record<string, string>) => api.get("/salary/sheet", { params }),
  sheetExport: (params: Record<string, string>) => api.get("/salary/sheet/export", { params, responseType: "blob" }),
  sheetExportPdf: (params: Record<string, string>) => api.get("/salary/sheet/export/pdf", { params, responseType: "blob" }),
  payslip: (params?: Record<string, string>) => api.get("/salary/payslip", { params }),
  payslipExport: (params: Record<string, string>) => api.get("/salary/payslip/export", { params, responseType: "blob" }),
  payslipExportPdf: (params: Record<string, string>) => api.get("/salary/payslip/export/pdf", { params, responseType: "blob" }),
  list: (params?: Record<string, string>) => api.get("/salary/list", { params }),
  summary: (params?: Record<string, string>) => api.get("/salary/summary", { params }),
  summaryExport: (params: Record<string, string>) => api.get("/salary/summary/export", { params, responseType: "blob" }),
  summaryExportPdf: (params: Record<string, string>) => api.get("/salary/summary/export/pdf", { params, responseType: "blob" }),
  dailySheet: (params?: Record<string, string>) => api.get("/salary/daily-sheet", { params }),
  bankSheet: (params?: Record<string, string>) => api.get("/salary/bank-sheet", { params }),
  bankSheetExport: (params: Record<string, string>) => api.get("/salary/bank-sheet/export", { params, responseType: "blob" }),
}

export const salaryIncrementApi = {
  list: (params?: Record<string, string>) => api.get("/salary/increments", { params }),
  create: (data: Record<string, unknown>) => api.post("/salary/increments", data),
  bulkApply: (data: Record<string, unknown>) => api.post("/salary/increments/bulk-apply", data),
  approve: (id: string, data?: Record<string, unknown>) => api.put(`/salary/increments/${id}/approve`, data || {}),
  reject: (id: string, data?: { reason: string }) => api.put(`/salary/increments/${id}/reject`, data || {}),
}

export const punishmentApi = {
  list: (params?: Record<string, string>) => api.get("/punishments", { params }),
  create: (data: Record<string, unknown>) => api.post("/punishments", data),
  calculate: (data: Record<string, unknown>) => api.post<{ amount: number }>("/punishments/calculate", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/punishments/${id}`, data),
  delete: (id: string) => api.delete(`/punishments/${id}`),
}

export const dailyScheduleApi = {
  list: (params?: Record<string, string>) => api.get("/daily-schedules", { params }),
  create: (data: Record<string, unknown>) => api.post("/daily-schedules", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/daily-schedules/${id}`, data),
  delete: (id: string) => api.delete(`/daily-schedules/${id}`),
}

export const nightBillApi = {
  list: (params?: Record<string, string>) => api.get("/night-bills", { params }),
  bulkCreate: (data: Record<string, unknown>) => api.post("/night-bills/bulk", data),
  listEligible: (params?: Record<string, string>) => api.get("/night-bills/eligible", { params }),
  delete: (id: string) => api.delete(`/night-bills/${id}`),
}

export const tiffinBillApi = {
  list: (params?: Record<string, string>) => api.get("/tiffin-bills", { params }),
  create: (data: Record<string, unknown>) => api.post("/tiffin-bills", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/tiffin-bills/${id}`, data),
  delete: (id: string) => api.delete(`/tiffin-bills/${id}`),
}

export const temporaryShiftApi = {
  list: (params?: Record<string, string>) => api.get("/temporary-shifts", { params }),
  get: (id: string) => api.get(`/temporary-shifts/${id}`),
  create: (data: Record<string, unknown>) => api.post("/temporary-shifts", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/temporary-shifts/${id}`, data),
  delete: (id: string) => api.delete(`/temporary-shifts/${id}`),
}

export const uploadApi = {
  file: (file: File) => {
    const formData = new FormData()
    formData.append("file", file)
    return api.post<{ url: string; filename: string }>("/upload", formData)
  },
}

export const notificationApi = {
  list: (params?: Record<string, string>) => api.get<PaginatedResponse<Notification>>("/notifications", { params }),
  unreadCount: () => api.get<{ count: number }>("/notifications/unread-count"),
  markAsRead: (id: string) => api.put(`/notifications/${id}/read`),
  markAllAsRead: () => api.put("/notifications/read-all"),
  delete: (id: string) => api.delete(`/notifications/${id}`),
}

export interface Notification {
  id: string
  user_id: string
  title: string
  message: string
  type: "info" | "warning" | "error" | "success"
  is_read: boolean
  read_at: string | null
  metadata: string | null
  created_at: string
  user?: { id: string; name: string; email: string }
}

export const eidBonusApi = {
  process: (data: { company_id: string; year: number }) => api.post("/eid-bonus/process", data),
  sheet: (params?: Record<string, string>) => api.get("/eid-bonus/sheet", { params }),
  summary: (params?: Record<string, string>) => api.get("/eid-bonus/summary", { params }),
  bankSheet: (params?: Record<string, string>) => api.get("/eid-bonus/bank-sheet", { params }),
  exportExcel: (params: Record<string, string>) => api.get("/eid-bonus/export/excel", { params, responseType: "blob" }),
}
