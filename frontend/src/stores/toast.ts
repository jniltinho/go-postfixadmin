import { defineStore } from 'pinia'
import { toast as notify, updateGlobalOptions, Zoom } from 'vue3-toastify'
import type { ToastOptions } from 'vue3-toastify'

export const useToastStore = defineStore('toast', () => {
  updateGlobalOptions({
    position: 'top-right',
    theme: 'colored',
    transition: Zoom,
    autoClose: 3000,
    newestOnTop: true,
    pauseOnHover: true,
    pauseOnFocusLoss: true,
    closeOnClick: true,
    closeButton: false,
  })

  return {
    success: (message: string, extra?: ToastOptions) => notify.success(message, extra),
    error: (message: string, extra?: ToastOptions) => notify.error(message, extra),
    info: (message: string, extra?: ToastOptions) => notify.info(message, extra),
    warning: (message: string, extra?: ToastOptions) => notify.warning(message, extra),
  }
})
