import { createApp } from 'vue'
import { Quasar } from 'quasar'
import { createPinia } from 'pinia'
import router from './router'
import { useAuthStore } from './stores/auth'

import '@quasar/extras/material-icons/material-icons.css'
import './style.css'

// Using pre-compiled Quasar CSS to avoid Sass preprocessor issues during build
import 'quasar/dist/quasar.css'

import App from './App.vue'

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)

// Initialize auth from localStorage before router runs so guards see correct state
useAuthStore(pinia).initFromStorage()

app.use(Quasar, {
  plugins: {},
  config: {
    brand: {
      primary: '#3B82F6',
      secondary: '#6366ee',
      accent: '#F97316',
      dark: '#1E293B',
      positive: '#21BA45',
      negative: '#C10015',
      info: '#31CCEC',
      warning: '#F2C037'
    }
  }
})
app.use(router)

app.mount('#app')
