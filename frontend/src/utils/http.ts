import axios, { type AxiosRequestConfig, type AxiosResponse } from 'axios'


const client = axios.create({
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
})

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
  window.location.href = '/login'
}

// Auto-retry on 401 with token refresh; logout if refresh also fails
client.interceptors.response.use(
  res => res,
  async (err) => {
    const original = err.config
    const url: string = original?.url ?? ''

    if (err.response?.status === 401 && !original._retry && !url.includes('/auth/')) {
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
    return Promise.reject(err)
  }
)

export type HttpConfig = AxiosRequestConfig
export type HttpResponse<T = any> = AxiosResponse<T>

export const http = {
  get: <T = any>(url: string, config?: HttpConfig) => client.get<T>(url, config),
  post: <T = any>(url: string, data?: unknown, config?: HttpConfig) => client.post<T>(url, data, config),
  put: <T = any>(url: string, data?: unknown, config?: HttpConfig) => client.put<T>(url, data, config),
  delete: <T = any>(url: string, config?: HttpConfig) => client.delete<T>(url, config),
}

export default http
