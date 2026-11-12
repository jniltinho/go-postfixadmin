import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import { useAuthStore } from './stores/auth'
import { useThemeStore } from './stores/theme'

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
  void useThemeStore(pinia).init()

  const auth = useAuthStore(pinia)
  await auth.initAuth()
  auth.startIdleWatch()

  app.use(router)

  app.mount('#app')
}

bootstrap()
