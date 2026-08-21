import api from "./axios-instance"

interface CrudApiOptions {
  basePath: string
  listParams?: boolean
}

export function createCrudApi(basePath: string, options?: CrudApiOptions) {
  const withParams = options?.listParams ?? true
  return {
    list: (params?: Record<string, string>) =>
      api.get(basePath, params && Object.keys(params).length > 0 ? { params } : {}),
    get: (id: string) => api.get(`${basePath}/${id}`),
    create: (data: Record<string, unknown>) => api.post(basePath, data),
    update: (id: string, data: Record<string, unknown>) => api.put(`${basePath}/${id}`, data),
    delete: (id: string) => api.delete(`${basePath}/${id}`),
  }
}
