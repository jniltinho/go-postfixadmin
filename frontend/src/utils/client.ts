import axios from 'axios'

export const apiClient = axios.create({
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
})

apiClient.interceptors.request.use(config => {
  const token = localStorage.getItem('access_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

let refreshPromise: Promise<boolean> | null = null

async function refreshSession(): Promise<boolean> {
  const { useAuthStore } = await import('../stores/auth')
  return useAuthStore().refresh()
}

apiClient.interceptors.response.use(
  res => res,
  async (err) => {
    const original = err.config
    const url: string = original?.url ?? ''

    if (err.response?.status === 401 && !original._retry && !url.includes('/auth/')) {
      original._retry = true
      try {
        refreshPromise ??= refreshSession().finally(() => { refreshPromise = null })
        const ok = await refreshPromise
        if (!ok) {
          const { useAuthStore } = await import('../stores/auth')
          useAuthStore().forceLogout()
          return Promise.reject(err)
        }
        const newToken = localStorage.getItem('access_token')
        original.headers.Authorization = `Bearer ${newToken}`
        return apiClient(original)
      } catch {
        const { useAuthStore } = await import('../stores/auth')
        useAuthStore().forceLogout()
        return Promise.reject(err)
      }
    }
    return Promise.reject(err)
  }
)