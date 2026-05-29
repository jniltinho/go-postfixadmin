import { defineStore } from 'pinia'
import http from '../utils/http'
import router from '../router'

interface User {
  username: string
  type: 'admin' | 'mailbox'
  superadmin: boolean
  domains: string[]
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    accessToken: localStorage.getItem('access_token') || null,
    isAuthenticated: false
  }),

  actions: {
    async login(username: string, password: string) {
      const { data } = await http.post(`${API_BASE}/auth/login`, {
        username,
        password
      })

      this.accessToken = data.data.access_token
      this.user = data.data.user
      this.isAuthenticated = true

      localStorage.setItem('access_token', data.data.access_token)
      localStorage.setItem('user_info', JSON.stringify(data.data.user))
    },

    async refresh(): Promise<boolean> {
      try {
        const { data } = await http.post(`${API_BASE}/auth/refresh`)
        const newToken = data.data?.access_token
        if (!newToken) return false

        this.accessToken = newToken
        this.isAuthenticated = true
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
      if (token) {
        this.accessToken = token
        this.isAuthenticated = true
        const saved = localStorage.getItem('user_info')
        if (saved) {
          try { this.user = JSON.parse(saved) } catch { /* ignore */ }
        }
      }
    }
  }
})

