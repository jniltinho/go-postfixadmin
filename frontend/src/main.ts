import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import { useAuthStore } from './stores/auth'

import './style.css'

import App from './App.vue'

import { Icon, ConfirmDialog } from './components/ui'
import AppTable from './components/AppTable.vue'

const app = createApp(App)

app.component('Icon', Icon)
app.component('AppTable', AppTable)
app.component('ConfirmDialog', ConfirmDialog)

const pinia = createPinia()
app.use(pinia)

// Initialize auth from localStorage before router runs so guards see correct state
useAuthStore(pinia).initFromStorage()

app.use(router)

app.mount('#app')
