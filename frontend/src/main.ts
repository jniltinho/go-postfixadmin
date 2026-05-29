import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import { useAuthStore } from './stores/auth'

import './style.css'

import App from './App.vue'

// Global UI components (neo-brutalist)
import { Icon, BrutalButton, BrutalCard, BrutalModal } from './components/ui'

// DataTables.net for all list tables (exact visual match to original 8081 server)
import DataTable from 'datatables.net-vue3'
import DataTablesCore from 'datatables.net'
import BrutalDataTable from './components/BrutalDataTable.vue'

// IMPORTANT: Must register the core library or you get "DataTables library not set" error
DataTable.use(DataTablesCore)

const app = createApp(App)

// Register brutalist components globally so we don't have to import them everywhere
app.component('Icon', Icon)
app.component('BrutalButton', BrutalButton)
app.component('BrutalCard', BrutalCard)
app.component('BrutalModal', BrutalModal)
app.component('DataTable', DataTable)
app.component('BrutalDataTable', BrutalDataTable)

const pinia = createPinia()
app.use(pinia)

// Initialize auth from localStorage before router runs so guards see correct state
useAuthStore(pinia).initFromStorage()

app.use(router)

app.mount('#app')
