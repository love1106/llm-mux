import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  managementKey: string | null
  setManagementKey: (key: string | null) => void
  isAuthenticated: boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      managementKey: null,
      isAuthenticated: false,
      setManagementKey: (key) => {
        set({ managementKey: key, isAuthenticated: !!key })
      },
    }),
    {
      name: 'llm-mux-auth',
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.isAuthenticated = !!state.managementKey
        }
      },
    }
  )
)
