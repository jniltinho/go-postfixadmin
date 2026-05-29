import { defineStore } from 'pinia'
import axios from 'axios'
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
      const { data } = await axios.post('/api/v1/auth/login', {
        username,
        password
      })

      this.accessToken = data.data.access_token
      this.user = data.data.user
      this.isAuthenticated = true

      localStorage.setItem('access_token', data.data.access_token)
      localStorage.setItem('user_info', JSON.stringify(data.data.user))
      axios.defaults.headers.common['Authorization'] = `Bearer ${data.data.access_token}`
    },

    async refresh(): Promise<boolean> {
      try {
        const { data } = await axios.post('/api/v1/auth/refresh')
        const newToken = data.data?.access_token
        if (!newToken) return false

        this.accessToken = newToken
        this.isAuthenticated = true
        localStorage.setItem('access_token', newToken)
        axios.defaults.headers.common['Authorization'] = `Bearer ${newToken}`
        return true
      } catch {
        this.forceLogout()
        return false
      }
    },

    logout() {
      axios.post('/api/v1/auth/logout').catch(() => {})
      this.forceLogout()
    },

    forceLogout() {
      this.accessToken = null
      this.user = null
      this.isAuthenticated = false
      localStorage.removeItem('access_token')
      localStorage.removeItem('user_info')
      delete axios.defaults.headers.common['Authorization']
      router.push({ name: 'Login' })
    },

    initFromStorage() {
      const token = localStorage.getItem('access_token')
      if (token) {
        this.accessToken = token
        this.isAuthenticated = true
        axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
        const saved = localStorage.getItem('user_info')
        if (saved) {
          try { this.user = JSON.parse(saved) } catch { /* ignore */ }
        }
      }
    }
  }
})

// Axios interceptor: on 401 try token refresh once, then redirect to login
let isRefreshing = false
let failedQueue: Array<{ resolve: (v: unknown) => void; reject: (e: unknown) => void }> = []

function processQueue(error: unknown, token: string | null = null) {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error)
    } else {
      resolve(token)
    }
  })
  failedQueue = []
}

axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    // Skip refresh for auth endpoints themselves
    const isAuthEndpoint = originalRequest?.url?.includes('/api/v1/auth/')
    if (error.response?.status !== 401 || isAuthEndpoint || originalRequest._retry) {
      return Promise.reject(error)
    }

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        failedQueue.push({ resolve, reject })
      }).then((token) => {
        originalRequest.headers['Authorization'] = `Bearer ${token}`
        return axios(originalRequest)
      }).catch((err) => Promise.reject(err))
    }

    originalRequest._retry = true
    isRefreshing = true

    try {
      const { data } = await axios.post('/api/v1/auth/refresh')
      const newToken = data.data?.access_token
      if (!newToken) throw new Error('no token in refresh response')

      const store = useAuthStore()
      store.accessToken = newToken
      store.isAuthenticated = true
      localStorage.setItem('access_token', newToken)
      axios.defaults.headers.common['Authorization'] = `Bearer ${newToken}`
      originalRequest.headers['Authorization'] = `Bearer ${newToken}`

      processQueue(null, newToken)
      return axios(originalRequest)
    } catch (refreshError) {
      processQueue(refreshError, null)
      const store = useAuthStore()
      store.forceLogout()
      return Promise.reject(refreshError)
    } finally {
      isRefreshing = false
    }
  }
)
