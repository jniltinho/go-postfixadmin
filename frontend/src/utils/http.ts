import axios, { type AxiosRequestConfig, type AxiosResponse } from 'axios'
import router from '../router'

const client = axios.create({
  headers: { 'Content-Type': 'application/json' },
})

// Attach JWT token to every request
client.interceptors.request.use(config => {
  const token = localStorage.getItem('access_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

let refreshPromise: Promise<string> | null = null

async function refreshAccessToken(): Promise<string> {
  const res = await axios.post(`${API_BASE}/auth/refresh`)
  const token = res.data?.data?.access_token
  if (!token) throw new Error('No token returned')
  localStorage.setItem('access_token', token)
  return token
}

function logout() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('user_info')
  router.push({ name: 'Login' })
}

// Auto-retry on 401 with token refresh
client.interceptors.response.use(
  res => res,
  async error => {
    const original = error.config
    const isAuthUrl = (original?.url ?? '').includes('/auth/')

    if (error.response?.status === 401 && !original._retry && !isAuthUrl) {
      original._retry = true
      try {
        refreshPromise ??= refreshAccessToken().finally(() => { refreshPromise = null })
        const newToken = await refreshPromise
        original.headers.Authorization = `Bearer ${newToken}`
        return client(original)
      } catch {
        logout()
      }
    }
    return Promise.reject(error)
  }
)

export type HttpConfig = AxiosRequestConfig
export type HttpResponse<T = any> = AxiosResponse<T>

export const http = {
  get:    <T = any>(url: string, config?: HttpConfig) => client.get<T>(url, config),
  post:   <T = any>(url: string, data?: unknown, config?: HttpConfig) => client.post<T>(url, data, config),
  put:    <T = any>(url: string, data?: unknown, config?: HttpConfig) => client.put<T>(url, data, config),
  delete: <T = any>(url: string, config?: HttpConfig) => client.delete<T>(url, config),
}

export default http
