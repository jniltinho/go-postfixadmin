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

const IDLE_KEY = 'idle_timeout'
const ACTIVITY_KEY = 'last_activity'
const DEFAULT_IDLE_SEC = 30 * 60

function saveIdleTimeout(sec: unknown) {
  if (typeof sec === 'number' && sec > 0) localStorage.setItem(IDLE_KEY, String(sec))
}

/** True when no user input was seen (in any tab) for longer than the server's session_idle_timeout. */
export function idleExpired(now = Date.now()): boolean {
  const last = Number(localStorage.getItem(ACTIVITY_KEY))
  if (!last) return false
  const sec = Number(localStorage.getItem(IDLE_KEY)) || DEFAULT_IDLE_SEC
  return now - last > sec * 1000
}

export function markActivity(now = Date.now()) {
  localStorage.setItem(ACTIVITY_KEY, String(now))
}

let idleWatch: ReturnType<typeof setInterval> | null = null
let nextRenew = 0

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
    hasPermission:
      (state) =>
      (perm: string): boolean => {
        if (!state.user) return false
        if (state.user.superadmin) return true
        const perms = state.user.permissions ?? []
        return perms.includes('*') || perms.includes(perm)
      },

    /** Returns true when the user holds at least one of the given roles. */
    hasRole:
      (state) =>
      (...roles: string[]): boolean => {
        if (!state.user) return false
        if (state.user.superadmin) return true
        const userRoles = state.user.roles ?? []
        return roles.some((r) => userRoles.includes(r))
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

      saveIdleTimeout(data.data.idle_timeout)
      markActivity()
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

      saveIdleTimeout(data.data.idle_timeout)
      markActivity()
      this.applySession(token, {
        ...data.data.user,
        permissions: decoded.permissions ?? [],
        roles: decoded.roles ?? [],
      })
    },

    async refresh(): Promise<boolean> {
      const before = localStorage.getItem('access_token')
      try {
        const { data } = await apiClient.post(`${API_BASE}/auth/refresh`)
        // a logout or a new login happened meanwhile (this tab or another): keep that state, not this answer
        const current = localStorage.getItem('access_token')
        if (current !== before) return current !== null
        const newToken = data.data?.access_token
        // a refresh that lands after an idle logout (here or in another tab) must not bring the session back
        if (!newToken || idleExpired()) return false
        saveIdleTimeout(data.data.idle_timeout)

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
        // a failure that lands after a new login must not log that new session out
        const current = localStorage.getItem('access_token')
        return current !== before && current !== null
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
        } catch {
          /* ignore corrupted storage */
        }
      }
    },

    /** Logs out after session_idle_timeout without keyboard, mouse or touch input; polling pages do not count. */
    startIdleWatch() {
      if (idleWatch) return
      let lastMark = 0
      const onActivity = () => {
        const now = Date.now()
        if (!this.isAuthenticated || now - lastMark < 10_000) return
        if (idleExpired(now)) return this.logout() // input after the deadline must not revive the session
        lastMark = now
        markActivity(now)
      }
      for (const ev of ['pointerdown', 'pointermove', 'keydown', 'wheel', 'touchstart'])
        window.addEventListener(ev, onActivity, { passive: true })
      // last_activity is shared, so an idle logout in one tab logs out the others on their next tick
      idleWatch = setInterval(() => {
        if (!this.isAuthenticated) return
        if (idleExpired()) return this.logout()
        // an active user may go a long time without API calls (typing a form): renew so the server sees the activity
        // the stored token is shared, so once one tab renewed it the others skip; a failure (e.g. 429) only logs
        // out once the token has really expired, and waits a minute before trying again
        const token = localStorage.getItem('access_token')
        if (!token || !isTokenExpired(token, 60) || Date.now() < nextRenew) return
        nextRenew = Date.now() + 60_000
        void this.refresh().then((ok) => {
          if (ok) nextRenew = 0
          else if (isTokenExpired(localStorage.getItem('access_token') ?? '', 0)) this.forceLogout()
        })
      }, 15_000)
    },

    async initAuth() {
      if (idleExpired()) {
        // before the router is installed: drop the session here, its guards send the user to the login page
        apiClient.post(`${API_BASE}/auth/logout`).catch(() => {})
        for (const k of ['access_token', 'user_info']) localStorage.removeItem(k)
        return
      }
      this.initFromStorage()
      if (this.accessToken) markActivity()

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
