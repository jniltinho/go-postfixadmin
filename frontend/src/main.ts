import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import { useAuthStore } from './stores/auth'
import Vue3Toastify from 'vue3-toastify'
import 'vue3-toastify/dist/index.css'

import './style.css'

import App from './App.vue'

import { Icon, ConfirmDialog } from './components/ui'
import AppTable from './components/AppTable.vue'

async function bootstrap() {
  const app = createApp(App)

  app.component('Icon', Icon)
  app.component('AppTable', AppTable)
  app.component('ConfirmDialog', ConfirmDialog)

  const pinia = createPinia()
  app.use(pinia)

  await useAuthStore(pinia).initAuth()

  app.use(Vue3Toastify)
  app.use(router)

  app.mount('#app')
}

bootstrap()