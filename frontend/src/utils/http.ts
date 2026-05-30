import router from '../router'

type QueryParams = Record<string, string | number | boolean | null | undefined>

export interface HttpConfig extends Omit<RequestInit, 'body' | 'method'> {
  params?: QueryParams
}

export interface HttpResponse<T = any> {
  data: T
  status: number
  statusText: string
  headers: Headers
}

interface RequestContext {
  method: string
  url: string
  data?: unknown
  config?: HttpConfig
}

export class HttpError<T = any> extends Error {
  response?: HttpResponse<T>
  config?: RequestContext

  constructor(message: string, response?: HttpResponse<T>, config?: RequestContext) {
    super(message)
    this.name = 'HttpError'
    this.response = response
    this.config = config
  }
}

let refreshPromise: Promise<string> | null = null

function buildUrl(url: string, params?: QueryParams) {
  if (!params) return url

  const base = url.startsWith('http') ? undefined : window.location.origin
  const parsedUrl = new URL(url, base)

  Object.entries(params).forEach(([key, value]) => {
    if (value !== null && value !== undefined) {
      parsedUrl.searchParams.set(key, String(value))
    }
  })

  return url.startsWith('/') ? `${parsedUrl.pathname}${parsedUrl.search}` : parsedUrl.toString()
}

function buildHeaders(headers?: HeadersInit, token = localStorage.getItem('access_token')) {
  const requestHeaders = new Headers(headers)

  if (token && !requestHeaders.has('Authorization')) {
    requestHeaders.set('Authorization', `Bearer ${token}`)
  }

  return requestHeaders
}

function buildRequestInit(method: string, data: unknown, config: HttpConfig = {}, token?: string): RequestInit {
  const { params: _params, headers: configHeaders, ...fetchConfig } = config
  const headers = buildHeaders(configHeaders, token)
  const init: RequestInit = {
    ...fetchConfig,
    method,
    headers
  }

  if (data === undefined) return init

  if (data instanceof FormData) {
    init.body = data
    return init
  }

  if (!headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  init.body = JSON.stringify(data)

  return init
}

async function parseResponseBody(response: Response) {
  if (response.status === 204) return null

  const contentType = response.headers.get('Content-Type') || ''
  const body = contentType.includes('application/json')
    ? await response.json()
    : await response.text()

  return body === '' ? null : body
}

async function toHttpResponse<T>(response: Response): Promise<HttpResponse<T>> {
  let data: T | null = null

  try {
    data = await parseResponseBody(response)
  } catch {
    data = null
  }

  return {
    data: data as T,
    status: response.status,
    statusText: response.statusText,
    headers: response.headers
  }
}

function getErrorMessage(response: HttpResponse) {
  if (response.data && typeof response.data === 'object' && 'message' in response.data) {
    return String(response.data.message)
  }

  return response.statusText || 'HTTP Error'
}

function isAuthUrl(url: string) {
  return url.includes('/auth/')
}

async function refreshAccessToken() {
  const response = await fetch(`${API_BASE}/auth/refresh`, { method: 'POST' })
  if (!response.ok) throw new Error('Refresh failed')

  const body = await response.json()
  const token = body.data?.access_token
  if (!token) throw new Error('No token returned')

  localStorage.setItem('access_token', token)
  return token
}

function logout() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('user_info')
  router.push({ name: 'Login' })
}

async function request<T = any>(
  method: string,
  url: string,
  data?: unknown,
  config?: HttpConfig,
  retryOnUnauthorized = true,
  token?: string
): Promise<HttpResponse<T>> {
  const context = { method, url, data, config }
  const fullUrl = buildUrl(url, config?.params)
  const init = buildRequestInit(method, data, config, token)

  let response: Response
  try {
    response = await fetch(fullUrl, init)
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Network Error'
    throw new HttpError(message, undefined, context)
  }

  const httpResponse = await toHttpResponse<T>(response)

  if (response.status === 401 && retryOnUnauthorized && !isAuthUrl(url)) {
    try {
      refreshPromise ??= refreshAccessToken().finally(() => {
        refreshPromise = null
      })

      const newToken = await refreshPromise
      return request<T>(method, url, data, config, false, newToken)
    } catch {
      logout()
      throw new HttpError('Unauthorized', httpResponse, context)
    }
  }

  if (!response.ok) {
    throw new HttpError(getErrorMessage(httpResponse), httpResponse, context)
  }

  return httpResponse
}

export const http = {
  get: <T = any>(url: string, config?: HttpConfig) => request<T>('GET', url, undefined, config),
  post: <T = any>(url: string, data?: unknown, config?: HttpConfig) => request<T>('POST', url, data, config),
  put: <T = any>(url: string, data?: unknown, config?: HttpConfig) => request<T>('PUT', url, data, config),
  delete: <T = any>(url: string, config?: HttpConfig) => request<T>('DELETE', url, undefined, config)
}

export default http
