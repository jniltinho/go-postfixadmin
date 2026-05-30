import { defineStore } from 'pinia'
import { ref } from 'vue'
import { toast as notify, type Id, type ToastOptions } from 'vue3-toastify'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface Toast {
  id: Id
  message: string
  type: ToastType
}

const defaultOptions: ToastOptions = {
  position: 'top-right',
  theme: 'colored',
  transition: 'zoom'
}

function options(duration = 3000): ToastOptions {
  return {
    ...defaultOptions,
    autoClose: duration > 0 ? duration : false
  }
}

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])

  function add(message: string, type: ToastType = 'info', duration = 3000): Id {
    return notify(message, {
      ...options(duration),
      type
    })
  }

  function remove(id?: Id): void {
    notify.remove(id)
  }

  function clear(): void {
    notify.clearAll()
  }

  const success = (message: string, duration?: number) => add(message, 'success', duration)
  const error = (message: string, duration?: number) => add(message, 'error', duration)
  const warning = (message: string, duration?: number) => add(message, 'warning', duration)
  const info = (message: string, duration?: number) => add(message, 'info', duration)

  return { toasts, add, remove, clear, success, error, warning, info }
})
