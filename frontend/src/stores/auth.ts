import { defineStore } from 'pinia'
import http from '../utils/http'
import router from '../router'

interface User {
  username: string
  type: 'admin' | 'mailbox'
  superadmin: boolean
  domains: string[]
  permissions: string[]
  roles: string[]
}

/** Decode the payload of a JWT without verifying the signature. */
function decodeJWT(token: string): Record<string, any> | null {
  try {
    const payload = token.split('.')[1]
    return JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')))
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    accessToken: localStorage.getItem('access_token') || null,
    isAuthenticated: false,
  }),

  getters: {
    /**
     * Returns true when the current user holds the given permission.
     * The wildcard "*" grants all permissions (superadmin fast-path).
     * Always returns true when superadmin=true, regardless of the permission list.
     */
    hasPermission: (state) => (perm: string): boolean => {
      if (!state.user) return false
      if (state.user.superadmin) return true
      const perms = state.user.permissions ?? []
      return perms.includes('*') || perms.includes(perm)
    },

    /** Returns true when the user holds at least one of the given roles. */
    hasRole: (state) => (...roles: string[]): boolean => {
      if (!state.user) return false
      if (state.user.superadmin) return true
      const userRoles = state.user.roles ?? []
      return roles.some(r => userRoles.includes(r))
    },
  },

  actions: {
    async login(username: string, password: string) {
      const { data } = await http.post(`${API_BASE}/auth/login`, { username, password })

      const token = data.data.access_token
      const decoded = decodeJWT(token)

      this.accessToken = token
      this.user = {
        ...data.data.user,
        permissions: decoded?.permissions ?? [],
        roles:       decoded?.roles ?? [],
      }
      this.isAuthenticated = true

      localStorage.setItem('access_token', token)
      localStorage.setItem('user_info', JSON.stringify(this.user))
    },

    async refresh(): Promise<boolean> {
      try {
        const { data } = await http.post(`${API_BASE}/auth/refresh`)
        const newToken = data.data?.access_token
        if (!newToken) return false

        const decoded = decodeJWT(newToken)
        this.accessToken = newToken
        this.isAuthenticated = true

        if (this.user) {
          this.user.permissions = decoded?.permissions ?? []
          this.user.roles       = decoded?.roles ?? []
        }

        localStorage.setItem('access_token', newToken)
        return true
      } catch {
        this.forceLogout()
        return false
      }
    },

    logout() {
      http.post(`${API_BASE}/auth/logout`).catch(() => {})
      this.forceLogout()
    },

    forceLogout() {
      this.accessToken = null
      this.user = null
      this.isAuthenticated = false
      localStorage.removeItem('access_token')
      localStorage.removeItem('user_info')
      router.push({ name: 'Login' })
    },

    initFromStorage() {
      const token = localStorage.getItem('access_token')
      if (!token) return

      this.accessToken = token
      this.isAuthenticated = true

      const saved = localStorage.getItem('user_info')
      if (saved) {
        try {
          const parsed = JSON.parse(saved)
          // Backfill permissions/roles from the stored JWT if the saved user_info
          // predates the RBAC fields.
          if (!parsed.permissions || !parsed.roles) {
            const decoded = decodeJWT(token)
            parsed.permissions = decoded?.permissions ?? []
            parsed.roles       = decoded?.roles ?? []
          }
          this.user = parsed
        } catch { /* ignore corrupted storage */ }
      }
    },
  },
})
