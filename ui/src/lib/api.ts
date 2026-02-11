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

export interface QuotaState {
  active_requests: number
  total_tokens_used: number
  in_cooldown: boolean
  cooldown_until?: string
  cooldown_remaining_seconds?: number
  learned_limit?: number
  learned_cooldown_seconds?: number
  last_exhausted_at?: string
  real_quota?: {
    remaining_fraction?: number
    remaining_tokens?: number
    window_reset_at?: string
    reset_in_seconds?: number
    fetched_at?: string
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
  subscription_type?: string
  status: 'active' | 'disabled' | 'error' | 'cooling' | 'unavailable'
  status_message?: string
  disabled: boolean
  last_refresh?: string
  expires_at?: string
  quota_state?: QuotaState
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
    cost_usd: number
  }
  by_provider: Record<string, { requests: number; tokens: { total: number; input: number; output: number }; cost_usd: number }>
  by_account: Record<string, { provider: string; auth_id: string; requests: number; tokens: { total: number; input: number; output: number } }>
  by_model: Record<string, { provider: string; requests: number; tokens: { total: number; input: number; output: number }; cost_usd: number }>
  by_ip?: Record<string, { requests: number; success: number; failure: number; tokens: { total: number; input: number; output: number; reasoning: number }; models: string[]; last_seen_at: string }>
  timeline?: {
    by_day: Array<{ day: string; requests: number; tokens: number }>
  }
}

export const managementApi = {
  getConfig: () => api.get<ApiResponse<Record<string, unknown>>>('/config'),
  getConfigYAML: () => api.get<string>('/config.yaml'),
  putConfigYAML: (yaml: string) => api.put('/config.yaml', yaml, { headers: { 'Content-Type': 'application/yaml' } }),

  getAuthFiles: () => api.get<ApiResponse<{ files: AuthFile[] }>>('/auth-files'),
  getAuthFileContent: (name: string) => api.get<Record<string, unknown>>('/auth-files/download', { params: { name } }),
  deleteAuthFile: (name: string) => api.delete('/auth-files', { params: { name } }),
  refreshAuthFile: (id: string) => api.post<ApiResponse<{ status: string; message: string }>>('/auth-files/refresh', null, { params: { id } }),
  importRawJSON: (jsonData: string) => api.post<ApiResponse<{ status: string; filename: string }>>('/auth-files/import', jsonData, { headers: { 'Content-Type': 'application/json' } }),

  getUsage: (params?: { days?: number; from?: string; to?: string }) =>
    api.get<ApiResponse<UsageStats>>('/usage', { params }),

  getLogs: (params?: { limit?: number; after?: number }) =>
    api.get<ApiResponse<{ lines: string[]; latest_timestamp: number }>>('/logs', { params }),
  deleteLogs: () => api.delete('/logs'),

  getDebug: () => api.get<ApiResponse<{ debug: boolean }>>('/debug'),
  setDebug: (value: boolean) => api.put('/debug', { value }),

  oauthStart: (provider: string, manual?: boolean) => api.post<{ auth_url: string; state: string }>('/oauth/start', { provider, manual }),
  oauthStatus: (state: string) => api.get<{ status: string }>(`/oauth/status/${state}`),
  oauthCancel: (state: string) => api.post(`/oauth/cancel/${state}`),
  oauthComplete: (state: string, code: string) => api.post<{ status: string; error?: string }>('/oauth/complete', { state, code }),
}

export default api
