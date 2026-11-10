import { reactive, readonly } from 'vue'

export type ToastType = 'success' | 'error' | 'info' | 'warning'

export interface ToastOptions {
  /** milliseconds before it closes; 0 keeps it until dismissed */
  duration?: number
  /** reuse an id to replace a toast instead of stacking a duplicate */
  id?: string
}

export interface ToastItem {
  id: string
  type: ToastType
  message: string
  duration: number
}

export interface ToastConfig {
  duration: number
  max: number
  newestOnTop: boolean
}

const config: ToastConfig = { duration: 3000, max: 5, newestOnTop: true }
const state = reactive({ items: [] as ToastItem[] })
let seq = 0

/** Changes the defaults for toasts shown after the call. */
export function configureToast(options: Partial<ToastConfig>): void {
  Object.assign(config, options)
}

/** Closes one toast, or all of them without an id. */
export function dismiss(id?: string): void {
  if (id === undefined) {
    state.items.splice(0)
    return
  }
  const i = state.items.findIndex((t) => t.id === id)
  if (i >= 0) state.items.splice(i, 1)
}

function show(type: ToastType, message: string, options: ToastOptions = {}): string {
  const id = options.id ?? `toast-${++seq}`
  dismiss(id)
  const item: ToastItem = { id, type, message, duration: Math.max(0, options.duration ?? config.duration) }
  if (config.newestOnTop) state.items.unshift(item)
  else state.items.push(item)
  // the oldest go first when there are too many
  while (state.items.length > config.max) {
    if (config.newestOnTop) state.items.pop()
    else state.items.shift()
  }
  return id
}

export const toast = {
  success: (message: string, options?: ToastOptions) => show('success', message, options),
  error: (message: string, options?: ToastOptions) => show('error', message, options),
  info: (message: string, options?: ToastOptions) => show('info', message, options),
  warning: (message: string, options?: ToastOptions) => show('warning', message, options),
  dismiss,
}

/** The toasts on screen, read-only (ToastHost renders them). */
export const toasts = readonly(state.items) as readonly ToastItem[]
