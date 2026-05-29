import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface Toast {
  id: number
  message: string
  type: ToastType
}

let _nextId = 0

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])

  function add(message: string, type: ToastType = 'info', duration = 4000): number {
    const id = ++_nextId
    toasts.value.push({ id, message, type })
    if (duration > 0) setTimeout(() => remove(id), duration)
    return id
  }

  function remove(id: number): void {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  const success = (msg: string, duration?: number) => add(msg, 'success', duration)
  const error   = (msg: string, duration?: number) => add(msg, 'error',   duration)
  const warning = (msg: string, duration?: number) => add(msg, 'warning', duration)
  const info    = (msg: string, duration?: number) => add(msg, 'info',    duration)

  return { toasts, add, remove, success, error, warning, info }
})
