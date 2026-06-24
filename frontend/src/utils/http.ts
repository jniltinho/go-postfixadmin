import type { AxiosRequestConfig, AxiosResponse } from 'axios'
import { apiClient } from './client'

export type HttpConfig = AxiosRequestConfig
export type HttpResponse<T = any> = AxiosResponse<T>

export const http = {
  get: <T = any>(url: string, config?: HttpConfig) => apiClient.get<T>(url, config),
  post: <T = any>(url: string, data?: unknown, config?: HttpConfig) => apiClient.post<T>(url, data, config),
  put: <T = any>(url: string, data?: unknown, config?: HttpConfig) => apiClient.put<T>(url, data, config),
  delete: <T = any>(url: string, config?: HttpConfig) => apiClient.delete<T>(url, config),
}

export default http