import { defineStore } from 'pinia'
import { apiClient } from '../utils/client'
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

function isTokenExpired(token: string, skewSec = 30): boolean {
  const decoded = decodeJWT(token)
  if (!decoded?.exp) return true
  return decoded.exp * 1000 <= Date.now() + skewSec * 1000
}

function userFromJWT(decoded: Record<string, any>): User {
  return {
    username: decoded.username ?? decoded.sub ?? '',
    type: (decoded.type as User['type']) ?? 'admin',
    superadmin: decoded.superadmin ?? false,
    domains: decoded.domains ?? [],
    permissions: decoded.permissions ?? [],
    roles: decoded.roles ?? [],
  }
}

function persistSession(token: string, user: User) {
  localStorage.setItem('access_token', token)
  localStorage.setItem('user_info', JSON.stringify(user))
}

function storedUserType(): User['type'] | undefined {
  try {
    const saved = localStorage.getItem('user_info')
    if (!saved) return undefined
    return JSON.parse(saved).type as User['type']
  } catch {
    return undefined
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
    applySession(token: string, user: User) {
      this.accessToken = token
      this.user = user
      this.isAuthenticated = true
      persistSession(token, user)
    },

    async userLogin(username: string, password: string) {
      const { data } = await apiClient.post(`${API_BASE}/auth/user-login`, { username, password })

      const token = data.data.access_token
      const decoded = decodeJWT(token)
      if (!decoded) throw new Error('invalid token')

      this.applySession(token, {
        ...data.data.user,
        permissions: decoded.permissions ?? [],
        roles: decoded.roles ?? [],
      })
    },

    async login(username: string, password: string) {
      const { data } = await apiClient.post(`${API_BASE}/auth/login`, { username, password })

      const token = data.data.access_token
      const decoded = decodeJWT(token)
      if (!decoded) throw new Error('invalid token')

      this.applySession(token, {
        ...data.data.user,
        permissions: decoded.permissions ?? [],
        roles: decoded.roles ?? [],
      })
    },

    async refresh(): Promise<boolean> {
      try {
        const { data } = await apiClient.post(`${API_BASE}/auth/refresh`)
        const newToken = data.data?.access_token
        if (!newToken) return false

        const decoded = decodeJWT(newToken)
        if (!decoded) return false

        const user = this.user
          ? {
              ...this.user,
              permissions: decoded.permissions ?? [],
              roles: decoded.roles ?? [],
            }
          : userFromJWT(decoded)

        this.applySession(newToken, user)
        return true
      } catch {
        return false
      }
    },

    logout() {
      apiClient.post(`${API_BASE}/auth/logout`).catch(() => {})
      this.forceLogout()
    },

    forceLogout() {
      const userType = this.user?.type ?? storedUserType()

      this.accessToken = null
      this.user = null
      this.isAuthenticated = false
      localStorage.removeItem('access_token')
      localStorage.removeItem('user_info')

      const routeName = userType === 'mailbox' ? 'UserLogin' : 'Login'
      router.push({ name: routeName })
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
          if (!parsed.permissions || !parsed.roles) {
            const decoded = decodeJWT(token)
            parsed.permissions = decoded?.permissions ?? []
            parsed.roles = decoded?.roles ?? []
          }
          this.user = parsed
        } catch { /* ignore corrupted storage */ }
      }
    },

    async initAuth() {
      this.initFromStorage()

      const token = this.accessToken
      if (!token) {
        const ok = await this.refresh()
        if (!ok) return
        return
      }

      if (!isTokenExpired(token)) return

      const ok = await this.refresh()
      if (!ok) this.forceLogout()
    },
  },
})