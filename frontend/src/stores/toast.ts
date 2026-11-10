import { defineStore } from 'pinia'
import { toast, type ToastOptions } from '../lib/toast'

// app-wide toasts (lib/toast, no external dependency); the store keeps the call sites as they were
export const useToastStore = defineStore('toast', () => ({
  success: (message: string, options?: ToastOptions) => toast.success(message, options),
  error: (message: string, options?: ToastOptions) => toast.error(message, options),
  info: (message: string, options?: ToastOptions) => toast.info(message, options),
  warning: (message: string, options?: ToastOptions) => toast.warning(message, options),
}))
