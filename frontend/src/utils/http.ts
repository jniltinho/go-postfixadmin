import router from '../router'

export interface HttpResponse<T = any> {
  data: T
  status: number
  statusText: string
  headers: Headers
}

export class HttpError extends Error {
  response?: HttpResponse
  config?: any

  constructor(message: string, response?: HttpResponse, config?: any) {
    super(message)
    this.name = 'HttpError'
    this.response = response
    this.config = config
  }
}

let isRefreshing = false
let failedQueue: Array<{ resolve: (token: string) => void; reject: (err: any) => void }> = []

function processQueue(error: any, token: string | null = null) {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error)
    } else {
      resolve(token!)
    }
  })
  failedQueue = []
}

function buildUrl(url: string, params?: Record<string, any>): string {
  if (!params) return url
  // Using a dummy base for relative URLs to avoid errors in new URL()
  const base = url.startsWith('http') ? undefined : window.location.origin
  const urlObj = new URL(url, base)
  Object.entries(params).forEach(([key, val]) => {
    if (val !== undefined && val !== null) {
      urlObj.searchParams.set(key, String(val))
    }
  })
  if (url.startsWith('/')) {
    return urlObj.pathname + urlObj.search
  }
  return urlObj.toString()
}

async function request<T = any>(
  method: string,
  url: string,
  data?: any,
  config?: { headers?: Record<string, string>; params?: Record<string, any>; [key: string]: any }
): Promise<HttpResponse<T>> {
  const fullUrl = buildUrl(url, config?.params)
  const token = localStorage.getItem('access_token')

  const headers = new Headers(config?.headers)
  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const options: RequestInit = {
    method,
    headers,
    ...config
  }

  if (data !== undefined) {
    if (data instanceof FormData) {
      options.body = data
    } else {
      if (!headers.has('Content-Type')) {
        headers.set('Content-Type', 'application/json')
      }
      options.body = JSON.stringify(data)
    }
  }

  let res: Response
  try {
    res = await fetch(fullUrl, options)
  } catch (err: any) {
    throw new HttpError(err.message || 'Network Error', undefined, { method, url, data, config })
  }

  let responseData: any = null
  const contentType = res.headers.get('Content-Type') || ''
  try {
    if (contentType.includes('application/json')) {
      responseData = await res.json()
    } else {
      responseData = await res.text()
    }
  } catch {
    // If reading body fails, leave as null
  }

  const httpResponse: HttpResponse<T> = {
    data: responseData,
    status: res.status,
    statusText: res.statusText,
    headers: res.headers
  }

  if (res.status === 401 && !url.includes('/auth/')) {
    if (isRefreshing) {
      try {
        const newToken = await new Promise<string>((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
        // Retry original request with new token
        const newHeaders = new Headers(options.headers)
        newHeaders.set('Authorization', `Bearer ${newToken}`)
        options.headers = newHeaders
        return await request<T>(method, url, data, config)
      } catch (err) {
        throw new HttpError('Unauthorized after token refresh failed', httpResponse, { method, url, data, config })
      }
    }

    isRefreshing = true
    try {
      const refreshRes = await fetch('/api/v1/auth/refresh', { method: 'POST' })
      if (!refreshRes.ok) throw new Error('Refresh failed')
      const refreshData = await refreshRes.json()
      const newToken = refreshData.data?.access_token
      if (!newToken) throw new Error('No token returned')

      localStorage.setItem('access_token', newToken)
      processQueue(null, newToken)

      // Retry original request
      const newHeaders = new Headers(options.headers)
      newHeaders.set('Authorization', `Bearer ${newToken}`)
      options.headers = newHeaders
      return await request<T>(method, url, data, config)
    } catch (refreshError) {
      processQueue(refreshError, null)
      localStorage.removeItem('access_token')
      localStorage.removeItem('user_info')
      router.push({ name: 'Login' })
      throw new HttpError('Unauthorized', httpResponse, { method, url, data, config })
    } finally {
      isRefreshing = false
    }
  }

  if (!res.ok) {
    throw new HttpError('HTTP Error', httpResponse, { method, url, data, config })
  }

  return httpResponse
}

export const http = {
  get: <T = any>(url: string, config?: any) => request<T>('GET', url, undefined, config),
  post: <T = any>(url: string, data?: any, config?: any) => request<T>('POST', url, data, config),
  put: <T = any>(url: string, data?: any, config?: any) => request<T>('PUT', url, data, config),
  delete: <T = any>(url: string, config?: any) => request<T>('DELETE', url, undefined, config)
}

export default http
