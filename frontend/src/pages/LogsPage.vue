<template>
  <div class="log-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">ADMIN ACTION LOGS</div>
        <div class="page-subtitle">AUDIT TRAIL OF ADMINISTRATIVE OPERATIONS</div>
      </div>
      
      <!-- Auto refresh controls -->
      <div class="refresh-controls">
        <label class="check-label auto-refresh-chk">
          <input type="checkbox" v-model="autoRefresh" />
          <span>AUTO REFRESH (10s)</span>
        </label>
        <button class="btn-primary btn-refresh-manual" @click="load(false)">
          <Icon name="refresh-cw" :size="16" style="margin-right:6px;vertical-align:middle" />
          REFRESH
        </button>
      </div>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <Icon name="alert-triangle" :size="16" class="mr-1" /> {{ error }}
    </div>

    <!-- ─── Filters & Search card ─── -->
    <div class="filter-card">
      <div class="filter-title">
        <Icon name="filter" :size="16" />
        SEARCH & ADVANCED FILTERS
      </div>
      <div class="filter-grid">
        <div class="form-group">
          <label class="form-label">Search Description</label>
          <input v-model="search" class="form-input" placeholder="Search logs..." />
        </div>
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
      </div>
    </div>

    <!-- ─── Table card ─── -->
    <div class="table-card">
      <div class="table-topbar">
        <div class="controls-left">
          <div class="per-page-wrap">
            <select v-model="rowsPerPage" class="ctrl-select">
              <option :value="10">10</option>
              <option :value="15">15</option>
              <option :value="25">25</option>
              <option :value="50">50</option>
              <option :value="100">100</option>
            </select>
            <span class="ctrl-label">entries per page</span>
          </div>
        </div>
        <div class="controls-right">
          <button class="btn-clear-filters" @click="clearFilters">
            CLEAR FILTERS
          </button>
        </div>
      </div>

      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr class="table-head-row">
              <th v-for="col in columns" :key="col.key" class="table-th" @click="sortBy(col.key)">
                {{ col.label }}
                <span class="sort-arrows">
                  <span :class="{ 'sort-active': sortKey === col.key && sortDir === 'asc' }">▲</span>
                  <span :class="{ 'sort-active': sortKey === col.key && sortDir === 'desc' }">▼</span>
                </span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading && !isQuietLoading">
              <td :colspan="columns.length" class="table-loading">
                <div class="spinner mx-auto" />
              </td>
            </tr>
            <tr v-else-if="rows.length === 0">
              <td :colspan="columns.length" class="table-empty">No admin action logs found</td>
            </tr>
            <tr v-for="(row, idx) in rows" :key="idx" class="table-row">
              <td class="table-td text-nowrap font-mono">{{ formatDate(row.timestamp) }}</td>
              <td class="table-td font-semibold text-break">{{ row.username || '—' }}</td>
              <td class="table-td text-break text-grey-7 font-mono">{{ row.domain || '—' }}</td>
              <td class="table-td">
                <span :class="getActionBadgeClass(row.action)">
                  {{ row.action.toUpperCase() }}
                </span>
              </td>
              <td class="table-td text-break desc-cell">{{ row.data || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="table-footer">
        <div class="showing-text">
          <template v-if="filteredCount === 0">Showing 0 entries</template>
          <template v-else>
            Showing {{ (currentPage - 1) * rowsPerPage + 1 }} to
            {{ Math.min(currentPage * rowsPerPage, filteredCount) }} of
            {{ filteredCount }} entries
            <span v-if="filteredCount !== totalRows" class="text-xs text-grey-6"> (filtered from {{ totalRows }} total)</span>
          </template>
        </div>
        <div class="pagination">
          <button class="pg-btn" :disabled="currentPage === 1" @click="changePage(1)">FIRST</button>
          <button class="pg-btn" :disabled="currentPage === 1" @click="changePage(currentPage - 1)">PREVIOUS</button>
          <button v-for="p in pageButtons" :key="p" class="pg-btn" :class="{ 'pg-active': p === currentPage }" @click="changePage(p)">{{ p }}</button>
          <button class="pg-btn" :disabled="currentPage === totalPages" @click="changePage(currentPage + 1)">NEXT</button>
          <button class="pg-btn" :disabled="currentPage === totalPages" @click="changePage(totalPages)">LAST</button>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import axios from 'axios'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()

interface LogEntry {
  timestamp: string
  username: string
  domain: string
  action: string
  data: string
}

// ─── State ───
const rows = ref<LogEntry[]>([])
const totalRows = ref(0)
const filteredCount = ref(0)
const loading = ref(true)
const isQuietLoading = ref(false)
const error = ref('')

// ─── Table & Filter controls ───
const search = ref('')
const filterAdmin = ref('')
const filterDomain = ref('')
const filterAction = ref('')
const rowsPerPage = ref(10)
const currentPage = ref(1)
const sortKey = ref('timestamp')
const sortDir = ref<'asc' | 'desc'>('desc')
const autoRefresh = ref(false)

const columns = [
  { key: 'timestamp', label: 'DATE/TIME' },
  { key: 'username',  label: 'ADMINISTRATOR' },
  { key: 'domain',    label: 'DOMAIN' },
  { key: 'action',    label: 'ACTION' },
  { key: 'data',      label: 'DESCRIPTION' },
]

// ─── Fetch data ───
async function load(quiet = false) {
  if (quiet) {
    isQuietLoading.value = true
  } else {
    loading.value = true
  }
  error.value = ''
  
  try {
    const res = await axios.get(`${API_BASE}/logs`, {
      params: {
        page: currentPage.value,
        per_page: rowsPerPage.value,
        search: search.value.trim(),
        admin: filterAdmin.value.trim(),
        domain: filterDomain.value.trim(),
        action: filterAction.value,
        sort: sortKey.value,
        order: sortDir.value,
      }
    })
    
    const paginated = res.data?.data
    rows.value = paginated?.data ?? []
    totalRows.value = paginated?.total ?? 0
    filteredCount.value = paginated?.filtered ?? 0
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load logs'
    toast.error(error.value)
  } finally {
    loading.value = false
    isQuietLoading.value = false
  }
}

// ─── Sorting ───
function sortBy(key: string) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'desc' // default desc for logs
  }
  currentPage.value = 1
  load()
}

// ─── Pagination ───
const totalPages = computed(() => Math.max(1, Math.ceil(filteredCount.value / rowsPerPage.value)))
const pageButtons = computed(() => {
  const total = totalPages.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const cur = currentPage.value
  const pages = new Set([1, total, cur, cur - 1, cur + 1].filter(p => p >= 1 && p <= total))
  return Array.from(pages).sort((a, b) => a - b)
})

function changePage(p: number) {
  if (p < 1 || p > totalPages.value) return
  currentPage.value = p
  load()
}

// ─── Watchers for search & filters with debouncing ───
let searchTimeout: any = null
function triggerDebouncedLoad() {
  currentPage.value = 1
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    load()
  }, 400)
}

watch([search, filterAdmin, filterDomain], triggerDebouncedLoad)
watch(filterAction, () => {
  currentPage.value = 1
  load()
})

watch(rowsPerPage, () => {
  currentPage.value = 1
  load()
})

// ─── Clear filters ───
function clearFilters() {
  search.value = ''
  filterAdmin.value = ''
  filterDomain.value = ''
  filterAction.value = ''
  currentPage.value = 1
  load()
  toast.success('Filters cleared successfully')
}

// ─── Auto refresh ───
let refreshInterval: any = null
watch(autoRefresh, (newVal) => {
  if (newVal) {
    refreshInterval = setInterval(() => {
      load(true)
    }, 10000)
    toast.success('Auto-refresh (10s) enabled')
  } else {
    if (refreshInterval) clearInterval(refreshInterval)
  }
})

onMounted(() => {
  load()
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
  if (searchTimeout) clearTimeout(searchTimeout)
})

// ─── Helpers ───
function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

function getActionBadgeClass(action: string): string {
  const act = action.toLowerCase()
  if (act.includes('create') || act.includes('add')) {
    return 'badge-action badge-action-create'
  }
  if (act.includes('edit') || act.includes('update')) {
    return 'badge-action badge-action-edit'
  }
  if (act.includes('delete') || act.includes('remove')) {
    return 'badge-action badge-action-delete'
  }
  if (act.includes('login')) {
    return 'badge-action badge-action-login'
  }
  return 'badge-action badge-action-other'
}
</script>

<style scoped>
.log-page { background: #ebf2fe; padding: 24px 28px 40px; }

.page-header { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 20px; gap: 20px; flex-wrap: wrap; }
.page-title { font-size: 28px; font-weight: 900; color: #1e293b; letter-spacing: -0.5px; line-height: 1; font-family: monospace; text-transform: uppercase; }
.page-subtitle { font-size: 10px; color: #94a3b8; letter-spacing: 0.8px; margin-top: 6px; text-transform: uppercase; font-weight: 700; }

.refresh-controls { display: flex; align-items: center; gap: 14px; }
.auto-refresh-chk { font-size: 11px; font-weight: 800; color: #1e293b; font-family: monospace; }

.btn-primary {
  background: #3b82f6; color: #fff; border: 2px solid #1e293b; padding: 10px 18px;
  font-size: 11px; font-weight: 800; letter-spacing: 0.6px; cursor: pointer;
  border-radius: 0; transition: all .15s; text-transform: uppercase;
  box-shadow: 2px 2px 0 #1e293b; display: flex; align-items: center;
}
.btn-primary:hover:not(:disabled) { background: #fff; color: #3b82f6; }
.btn-primary:active:not(:disabled) { transform: translate(0,0); box-shadow: none; }

.error-banner { background: #fef2f2; border: 1px solid #fca5a5; color: #dc2626; padding: 10px 14px; font-size: 13px; margin-bottom: 18px; display: flex; align-items: center; gap: 6px; }

/* ─── Filter card ─── */
.filter-card { background: #fff; border: 2px solid #1e293b; box-shadow: 2px 2px 0 #1e293b; padding: 14px 18px 18px; margin-bottom: 16px; }
.filter-title { font-size: 12px; font-weight: 900; color: #1e293b; letter-spacing: 0.6px; text-transform: uppercase; font-family: monospace; display: flex; align-items: center; gap: 6px; margin-bottom: 14px; }
.filter-grid { display: grid; grid-template-columns: 2fr 1fr 1fr 1fr; gap: 14px; }

@media (max-width: 900px) {
  .filter-grid { grid-template-columns: 1fr 1fr; }
}
@media (max-width: 600px) {
  .filter-grid { grid-template-columns: 1fr; }
}

/* ─── Table Controls & Custom Elements ─── */

.btn-clear-filters {
  background: #f8fafc; border: 1px solid #cbd5e1; color: #475569; padding: 5px 12px;
  font-size: 10px; font-weight: 800; letter-spacing: 0.4px; cursor: pointer;
  border-radius: 0; transition: all .12s; text-transform: uppercase;
}
.btn-clear-filters:hover { background: #f1f5f9; border-color: #94a3b8; color: #1e293b; }

.table-wrap { overflow-x: auto; }
.text-nowrap { white-space: nowrap; }
.font-mono { font-family: 'Fira Code', var(--font-mono), monospace; }
.font-semibold { font-weight: 600; }
.text-break { word-break: break-all; }
.text-grey-7 { color: #475569; }
.desc-cell { max-width: 400px; }

/* ─── Badges ─── */
.badge-action { padding: 3px 8px; font-size: 10px; font-weight: 800; border-radius: 0; display: inline-block; border: 1px solid #1e293b; text-align: center; white-space: nowrap; box-shadow: 1px 1px 0 #1e293b; }
.badge-action-create { background: #dcfce7; color: #16a34a; }
.badge-action-edit   { background: #eff6ff; color: #2563eb; }
.badge-action-delete { background: #fee2e2; color: #dc2626; }
.badge-action-login  { background: #faf5ff; color: #9333ea; }
.badge-action-other  { background: #f1f5f9; color: #475569; }

.table-loading, .table-empty { text-align: center; padding: 40px; color: #94a3b8; font-size: 13px; }
.showing-text { font-size: 12.5px; color: #64748b; }
.pagination { display: flex; gap: 3px; }
.pg-btn { height: 28px; padding: 0 10px; font-size: 10px; font-weight: 700; color: #374151; background: #fff; border: 1px solid #d1d5db; border-radius: 0; cursor: pointer; letter-spacing: 0.4px; text-transform: uppercase; white-space: nowrap; }
.pg-btn:hover:not(:disabled) { border-color: #1e293b; color: #1e293b; background: #f8fafc; }
.pg-btn:disabled { opacity: .35; cursor: default; }
.pg-active { background: #3b82f6 !important; color: #fff !important; border-color: #3b82f6 !important; }

/* ─── Form elements ─── */
.form-group { display: flex; flex-direction: column; gap: 3px; }
.form-label { font-size: 10px; font-weight: 800; color: #1e293b; letter-spacing: 0.7px; text-transform: uppercase; margin-bottom: 2px; }
.form-input { border: 2px solid #1e293b; padding: 6px 10px; font-size: 13px; color: #374151; outline: none; border-radius: 0; width: 100%; box-sizing: border-box; height: 36px; }
.form-select-plain { border: 2px solid #1e293b; padding: 0 10px; font-size: 12px; color: #374151; background: #fff; border-radius: 0; width: 100%; height: 36px; outline: none; cursor: pointer; font-weight: 700; }
.form-select-plain:focus { border-color: #3b82f6; }

.check-label { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 600; color: #374151; cursor: pointer; }
.check-label input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; accent-color: #3b82f6; }
</style>