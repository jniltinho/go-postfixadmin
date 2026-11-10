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
          <Icon name="refresh-cw" :size="16" style="margin-right:0.375rem;vertical-align:middle" />
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
.log-page { background: var(--color-page); padding: 1.5rem 1.75rem 2.5rem; }

.page-header {
  display: flex; justify-content: space-between; align-items: flex-end;
  margin-bottom: 1.25rem; gap: 1.25rem; flex-wrap: wrap;
}
.page-title {
  font-size: 1.75rem; font-weight: 900; color: var(--color-fg);
  letter-spacing: -0.03125rem; line-height: 1; font-family: monospace; text-transform: uppercase;
}
.page-subtitle {
  font-size: 0.625rem; color: var(--color-fg-subtle); letter-spacing: 0.05rem;
  margin-top: 0.375rem; text-transform: uppercase; font-weight: 700;
}

.refresh-controls { display: flex; align-items: center; gap: 0.875rem; }
.auto-refresh-chk { font-size: 0.6875rem; font-weight: 800; color: var(--color-fg); font-family: monospace; }

.btn-primary {
  background: var(--color-action); color: var(--color-on-action); border: 2px solid var(--color-line); padding: 0.625rem 1.125rem;
  font-size: 0.6875rem; font-weight: 800; letter-spacing: 0.0375rem; cursor: pointer;
  border-radius: 0; transition: all .15s; text-transform: uppercase;
  box-shadow: var(--shadow-2); display: flex; align-items: center;
}
.btn-primary:hover { background: var(--color-surface); color: var(--color-action); }
.btn-primary:active { transform: translate(0, 0); box-shadow: none; }

/* ─── Filter card ─── */
.filter-card {
  background: var(--color-surface); border: 2px solid var(--color-line); box-shadow: var(--shadow-2);
  padding: 0.875rem 1.125rem 1.125rem; margin-bottom: 1rem;
}
.filter-title {
  font-size: 0.75rem; font-weight: 900; color: var(--color-fg); letter-spacing: 0.0375rem;
  text-transform: uppercase; font-family: monospace;
  display: flex; align-items: center; gap: 0.375rem; margin-bottom: 0.875rem;
}
.filter-grid {
  display: grid; grid-template-columns: 1fr 1fr 1fr auto; gap: 0.875rem; align-items: end;
}
.form-group--btn { display: flex; align-items: flex-end; }

.btn-clear-filters {
  background: var(--color-surface-2); border: 1px solid var(--color-field-line); color: var(--color-fg-2);
  padding: 0 0.75rem; font-size: 0.625rem; font-weight: 800; letter-spacing: 0.025rem;
  cursor: pointer; border-radius: 0; transition: all .12s; text-transform: uppercase; height: 2.25rem;
}
.btn-clear-filters:hover { background: var(--color-surface-3); border-color: var(--color-fg-subtle); color: var(--color-fg); }

/* ─── Badges ─── */
.badge-action {
  padding: 0.1875rem 0.5rem; font-size: 0.625rem; font-weight: 800; border-radius: 0;
  display: inline-block; border: 1px solid var(--color-line); text-align: center;
  white-space: nowrap; box-shadow: var(--shadow-1);
}
.badge-action-create { background: var(--color-ok-soft); color: var(--color-ok-strong); }
.badge-action-edit   { background: var(--color-action-soft); color: var(--color-action-hover); }
.badge-action-delete { background: var(--color-err-soft); color: var(--color-danger); }
.badge-action-login  { background: var(--color-accent-soft); color: var(--color-accent-text); }
.badge-action-other  { background: var(--color-surface-3); color: var(--color-fg-2); }

.font-mono { font-family: 'Fira Code', var(--font-mono), monospace; }
.text-nowrap { white-space: nowrap; }
.desc-cell { word-break: break-all; }

/* ─── Form elements ─── */
.form-group { display: flex; flex-direction: column; gap: 0.1875rem; }
.form-label {
  font-size: 0.625rem; font-weight: 800; color: var(--color-fg);
  letter-spacing: 0.04375rem; text-transform: uppercase; margin-bottom: 0.125rem;
}
.form-input {
  border: 2px solid var(--color-line); padding: 0.375rem 0.625rem; font-size: 0.8125rem; color: var(--color-fg-2);
  outline: none; border-radius: 0; width: 100%; box-sizing: border-box; height: 2.25rem;
}
.form-select-plain {
  border: 2px solid var(--color-line); padding: 0 0.625rem; font-size: 0.75rem; color: var(--color-fg-2);
  background: var(--color-surface); border-radius: 0; width: 100%; height: 2.25rem;
  outline: none; cursor: pointer; font-weight: 700;
}

.check-label { display: flex; align-items: center; gap: 0.5rem; font-size: 0.8125rem; font-weight: 600; color: var(--color-fg-2); cursor: pointer; }
.check-label input[type="checkbox"] { width: 1.125rem; height: 1.125rem; cursor: pointer; accent-color: var(--color-action); }

@media (max-width: 900px) { .filter-grid { grid-template-columns: 1fr 1fr; } }
@media (max-width: 600px) { .filter-grid { grid-template-columns: 1fr; } }
</style>
