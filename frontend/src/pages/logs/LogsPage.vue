<template>
  <div class="log-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">ADMIN ACTION LOGS</div>
        <div class="page-subtitle">AUDIT TRAIL OF ADMINISTRATIVE OPERATIONS</div>
      </div>
      <div class="refresh-controls">
        <label class="check-label auto-refresh-chk">
          <input type="checkbox" v-model="autoRefresh" />
          <span>AUTO REFRESH (10s)</span>
        </label>
        <button class="btn-primary" @click="load(false)">
          <Icon name="refresh-cw" :size="16" style="margin-right:6px;vertical-align:middle" />
          REFRESH
        </button>
      </div>
    </div>

    <!-- ─── AppTable ─── -->
    <AppTable
      title="AUDIT TRAIL"
      :rows="filteredRows"
      :columns="columns"
      row-key="_key"
      :search-fields="['username', 'domain', 'action', 'data']"
      default-sort-key="timestamp"
      default-sort-dir="desc"
      :initial-rows-per-page="12"
      :loading="loading"
      :show-actions="false"
    >
      <template #toolbar>
        <div class="filter-card">
          <div class="filter-title">
            <Icon name="filter" :size="16" />
            ADVANCED FILTERS
          </div>
          <div class="filter-grid">
            <div class="form-group">
              <label class="form-label">Filter Admin</label>
              <input v-model="filterAdmin" class="form-input" placeholder="e.g. admin@domain.com" />
            </div>
            <div class="form-group">
              <label class="form-label">Filter Domain</label>
              <input v-model="filterDomain" class="form-input" placeholder="e.g. example.com" />
            </div>
            <div class="form-group">
              <label class="form-label">Filter Action</label>
              <select v-model="filterAction" class="form-select-plain">
                <option value="">ALL ACTIONS</option>
                <option value="create">CREATE</option>
                <option value="edit">EDIT</option>
                <option value="delete">DELETE</option>
                <option value="login">LOGIN</option>
                <option value="logout">LOGOUT</option>
              </select>
            </div>
            <div class="form-group form-group--btn">
              <button class="btn-clear-filters" @click="clearFilters">CLEAR FILTERS</button>
            </div>
          </div>
        </div>
      </template>

      <template #cell-timestamp="{ value }">
        <span class="font-mono text-nowrap">{{ formatDate(value) }}</span>
      </template>
      <template #cell-domain="{ value }">
        <span class="font-mono">{{ value }}</span>
      </template>
      <template #cell-action="{ value }">
        <span :class="getActionBadgeClass(value)">{{ value?.toUpperCase() }}</span>
      </template>
      <template #cell-data="{ value }">
        <span class="desc-cell">{{ value || '—' }}</span>
      </template>
    </AppTable>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import http from '../../utils/http'
import { useToastStore } from '../../stores/toast'
import AppTable from '../../components/AppTable.vue'

const toast = useToastStore()

interface LogEntry {
  _key: string
  timestamp: string
  username: string
  domain: string
  action: string
  data: string
}

const allRows = ref<LogEntry[]>([])
const loading = ref(true)
const filterAdmin = ref('')
const filterDomain = ref('')
const filterAction = ref('')
const autoRefresh = ref(false)

const columns = [
  { key: 'timestamp', label: 'DATE/TIME' },
  { key: 'username',  label: 'ADMINISTRATOR' },
  { key: 'domain',    label: 'DOMAIN' },
  { key: 'action',    label: 'ACTION' },
  { key: 'data',      label: 'DESCRIPTION' },
]

const filteredRows = computed(() => {
  return allRows.value.filter(row => {
    if (filterAdmin.value && !row.username.toLowerCase().includes(filterAdmin.value.toLowerCase())) return false
    if (filterDomain.value && !row.domain.toLowerCase().includes(filterDomain.value.toLowerCase())) return false
    if (filterAction.value && !row.action.toLowerCase().includes(filterAction.value.toLowerCase())) return false
    return true
  })
})

async function load(quiet = false) {
  if (!quiet) loading.value = true
  try {
    const res = await http.get(`${API_BASE}/logs`, {
      params: { page: 1, per_page: 9999, sort: 'timestamp', order: 'desc' }
    })
    const data = res.data?.data?.data ?? []
    allRows.value = data.map((log: any, idx: number) => ({
      _key: String(idx),
      timestamp: log.timestamp ?? '',
      username:  log.username  || '—',
      domain:    log.domain    || '—',
      action:    log.action    || '',
      data:      log.data      || '',
    }))
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load logs')
  } finally {
    loading.value = false
  }
}

function clearFilters() {
  filterAdmin.value = ''
  filterDomain.value = ''
  filterAction.value = ''
  toast.success('Filters cleared')
}

let refreshInterval: any = null
watch(autoRefresh, (val) => {
  if (val) {
    refreshInterval = setInterval(() => load(true), 10000)
    toast.success('Auto-refresh (10s) enabled')
  } else {
    if (refreshInterval) clearInterval(refreshInterval)
  }
})

onMounted(() => load())
onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
})

function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}

function getActionBadgeClass(action: string): string {
  const act = (action || '').toLowerCase()
  if (act.includes('create') || act.includes('add'))    return 'badge-action badge-action-create'
  if (act.includes('edit')   || act.includes('update')) return 'badge-action badge-action-edit'
  if (act.includes('delete') || act.includes('remove')) return 'badge-action badge-action-delete'
  if (act.includes('login'))                            return 'badge-action badge-action-login'
  return 'badge-action badge-action-other'
}
</script>

<style scoped>
.log-page { background: #ebf2fe; padding: 24px 28px 40px; }

.page-header {
  display: flex; justify-content: space-between; align-items: flex-end;
  margin-bottom: 20px; gap: 20px; flex-wrap: wrap;
}
.page-title {
  font-size: 28px; font-weight: 900; color: #1e293b;
  letter-spacing: -0.5px; line-height: 1; font-family: monospace; text-transform: uppercase;
}
.page-subtitle {
  font-size: 10px; color: #94a3b8; letter-spacing: 0.8px;
  margin-top: 6px; text-transform: uppercase; font-weight: 700;
}

.refresh-controls { display: flex; align-items: center; gap: 14px; }
.auto-refresh-chk { font-size: 11px; font-weight: 800; color: #1e293b; font-family: monospace; }

.btn-primary {
  background: #3b82f6; color: #fff; border: 2px solid #1e293b; padding: 10px 18px;
  font-size: 11px; font-weight: 800; letter-spacing: 0.6px; cursor: pointer;
  border-radius: 0; transition: all .15s; text-transform: uppercase;
  box-shadow: 2px 2px 0 #1e293b; display: flex; align-items: center;
}
.btn-primary:hover { background: #fff; color: #3b82f6; }
.btn-primary:active { transform: translate(0, 0); box-shadow: none; }

/* ─── Filter card ─── */
.filter-card {
  background: #fff; border: 2px solid #1e293b; box-shadow: 2px 2px 0 #1e293b;
  padding: 14px 18px 18px; margin-bottom: 16px;
}
.filter-title {
  font-size: 12px; font-weight: 900; color: #1e293b; letter-spacing: 0.6px;
  text-transform: uppercase; font-family: monospace;
  display: flex; align-items: center; gap: 6px; margin-bottom: 14px;
}
.filter-grid {
  display: grid; grid-template-columns: 1fr 1fr 1fr auto; gap: 14px; align-items: end;
}
.form-group--btn { display: flex; align-items: flex-end; }

.btn-clear-filters {
  background: #f8fafc; border: 1px solid #cbd5e1; color: #475569;
  padding: 0 12px; font-size: 10px; font-weight: 800; letter-spacing: 0.4px;
  cursor: pointer; border-radius: 0; transition: all .12s; text-transform: uppercase; height: 36px;
}
.btn-clear-filters:hover { background: #f1f5f9; border-color: #94a3b8; color: #1e293b; }

/* ─── Badges ─── */
.badge-action {
  padding: 3px 8px; font-size: 10px; font-weight: 800; border-radius: 0;
  display: inline-block; border: 1px solid #1e293b; text-align: center;
  white-space: nowrap; box-shadow: 1px 1px 0 #1e293b;
}
.badge-action-create { background: #dcfce7; color: #16a34a; }
.badge-action-edit   { background: #eff6ff; color: #2563eb; }
.badge-action-delete { background: #fee2e2; color: #dc2626; }
.badge-action-login  { background: #faf5ff; color: #9333ea; }
.badge-action-other  { background: #f1f5f9; color: #475569; }

.font-mono { font-family: 'Fira Code', var(--font-mono), monospace; }
.text-nowrap { white-space: nowrap; }
.desc-cell { word-break: break-all; }

/* ─── Form elements ─── */
.form-group { display: flex; flex-direction: column; gap: 3px; }
.form-label {
  font-size: 10px; font-weight: 800; color: #1e293b;
  letter-spacing: 0.7px; text-transform: uppercase; margin-bottom: 2px;
}
.form-input {
  border: 2px solid #1e293b; padding: 6px 10px; font-size: 13px; color: #374151;
  outline: none; border-radius: 0; width: 100%; box-sizing: border-box; height: 36px;
}
.form-select-plain {
  border: 2px solid #1e293b; padding: 0 10px; font-size: 12px; color: #374151;
  background: #fff; border-radius: 0; width: 100%; height: 36px;
  outline: none; cursor: pointer; font-weight: 700;
}

.check-label { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 600; color: #374151; cursor: pointer; }
.check-label input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; accent-color: #3b82f6; }

@media (max-width: 900px) { .filter-grid { grid-template-columns: 1fr 1fr; } }
@media (max-width: 600px) { .filter-grid { grid-template-columns: 1fr; } }
</style>
