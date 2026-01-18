import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const api = axios.create({
  baseURL: '/v1/management',
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use((config) => {
  const managementKey = useAuthStore.getState().managementKey
  if (managementKey) {
    config.headers['X-Management-Key'] = managementKey
  }
  return config
})

export interface ApiResponse<T> {
  data: T
  meta: {
    timestamp: string
    version: string
  }
}

export interface AuthFile {
  id: string
  name: string
  type: string
  provider: string
  label: string
  email?: string
  account_type?: string
  status: 'active' | 'disabled' | 'error' | 'cooling' | 'unavailable'
  status_message?: string
  disabled: boolean
  last_refresh?: string
}

export interface UsageStats {
  summary: {
    total_requests: number
    success_count: number
    failure_count: number
    tokens: {
      total: number
      input: number
      output: number
    }
  }
  by_provider: Record<string, { requests: number; tokens: { total: number } }>
  by_account: Record<string, { requests: number; tokens: { total: number } }>
  timeline?: {
    by_day: Array<{ day: string; requests: number; tokens: number }>
  }
}

export const managementApi = {
  getConfig: () => api.get<ApiResponse<Record<string, unknown>>>('/config'),
  getConfigYAML: () => api.get<string>('/config.yaml'),
  putConfigYAML: (yaml: string) => api.put('/config.yaml', yaml, { headers: { 'Content-Type': 'application/yaml' } }),

  getAuthFiles: () => api.get<ApiResponse<{ files: AuthFile[] }>>('/auth-files'),
  deleteAuthFile: (name: string) => api.delete('/auth-files', { params: { name } }),

  getUsage: (params?: { days?: number; from?: string; to?: string }) =>
    api.get<ApiResponse<UsageStats>>('/usage', { params }),

  getLogs: (params?: { limit?: number; after?: number }) =>
    api.get<ApiResponse<{ lines: string[]; latest_timestamp: number }>>('/logs', { params }),
  deleteLogs: () => api.delete('/logs'),

  getDebug: () => api.get<ApiResponse<{ debug: boolean }>>('/debug'),
  setDebug: (value: boolean) => api.put('/debug', { value }),

  oauthStart: (provider: string) => api.post<{ auth_url: string; state: string }>('/oauth/start', { provider }),
  oauthStatus: (state: string) => api.get<{ status: string }>(`/oauth/status/${state}`),
  oauthCancel: (state: string) => api.post(`/oauth/cancel/${state}`),
}

export default api
