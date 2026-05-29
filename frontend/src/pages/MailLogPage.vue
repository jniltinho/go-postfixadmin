<template>
  <q-page class="ml-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">MAIL DELIVERY LOGS</div>
        <div class="page-subtitle">MONITOR INCOMING AND OUTGOING SMTP TRAFFIC</div>
      </div>
      
      <!-- Auto refresh controls -->
      <div class="refresh-controls">
        <label class="check-label auto-refresh-chk">
          <input type="checkbox" v-model="autoRefresh" />
          <span>AUTO REFRESH (10s)</span>
        </label>
        <button class="btn-primary btn-refresh-manual" @click="load(false)">
          <q-icon name="refresh" size="16px" style="margin-right:6px;vertical-align:middle" />
          REFRESH
        </button>
      </div>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <q-icon name="warning" size="16px" /> {{ error }}
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
          <span class="ctrl-label">Search:</span>
          <input v-model="search" class="search-input" placeholder="Search sender, recipient, IP..." />
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
                <q-spinner color="primary" size="24px" />
              </td>
            </tr>
            <tr v-else-if="rows.length === 0">
              <td :colspan="columns.length" class="table-empty">No delivery logs found</td>
            </tr>
            <tr v-for="(row, idx) in rows" :key="idx" class="table-row">
              <td class="table-td text-nowrap font-mono">{{ row.logdate }}</td>
              <td class="table-td font-semibold text-break">{{ row.m_from || '—' }}</td>
              <td class="table-td font-semibold text-break">{{ row.m_to || '—' }}</td>
              <td class="table-td text-break text-grey-7">{{ row.domain_from || '—' }}</td>
              <td class="table-td text-break text-grey-7">{{ row.domain_to || '—' }}</td>
              <td class="table-td font-mono">
                <span class="ip-address" @click="copyText(row.host_ip)" title="Click to copy IP">
                  {{ row.host_ip || '—' }}
                </span>
              </td>
              <td class="table-td text-break text-grey-7 font-mono">{{ row.host_name || '—' }}</td>
              <td class="table-td text-break text-grey-7 font-mono">{{ row.helo || '—' }}</td>
              <td class="table-td font-mono text-right text-nowrap">{{ formatSize(row.msgsize) }}</td>
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

  </q-page>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import axios from 'axios'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()

interface MailLogEntry {
  logdate: string
  m_from: string
  m_to: string
  domain_from: string
  domain_to: string
  host_ip: string
  host_name: string
  helo: string
  msgsize: number
}

// ─── State ───
const rows = ref<MailLogEntry[]>([])
const totalRows = ref(0)
const filteredCount = ref(0)
const loading = ref(true)
const isQuietLoading = ref(false)
const error = ref('')

// ─── Table controls ───
const search = ref('')
const rowsPerPage = ref(15)
const currentPage = ref(1)
const sortKey = ref('logdate') // Maps to tstamp field
const sortDir = ref<'asc' | 'desc'>('desc')
const autoRefresh = ref(false)

const columns = [
  { key: 'logdate',     label: 'DATE/TIME' },
  { key: 'm_from',      label: 'SENDER (FROM)' },
  { key: 'm_to',        label: 'RECIPIENT (TO)' },
  { key: 'domain_from', label: 'DOMAIN FROM' },
  { key: 'domain_to',   label: 'DOMAIN TO' },
  { key: 'host_ip',     label: 'CLIENT IP' },
  { key: 'host_name',   label: 'HOST' },
  { key: 'helo',        label: 'HELO' },
  { key: 'msgsize',     label: 'SIZE' },
]

// Map the sorting key to backend fields
function getBackendSortKey(key: string): string {
  switch (key) {
    case 'logdate':
      return 'tstamp'
    default:
      return key
  }
}

// ─── Fetch data ───
async function load(quiet = false) {
  if (quiet) {
    isQuietLoading.value = true
  } else {
    loading.value = true
  }
  error.value = ''
  
  try {
    const res = await axios.get(`${API_BASE}/maillog`, {
      params: {
        page: currentPage.value,
        per_page: rowsPerPage.value,
        search: search.value.trim(),
        sort: getBackendSortKey(sortKey.value),
        order: sortDir.value,
      }
    })
    
    const paginated = res.data?.data
    rows.value = paginated?.data ?? []
    totalRows.value = paginated?.total ?? 0
    filteredCount.value = paginated?.filtered ?? 0
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load delivery logs'
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
    sortDir.value = 'desc' // default desc for date or size or others
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

// ─── Watchers for search & pagination ───
let searchTimeout: any = null
watch(search, () => {
  currentPage.value = 1
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    load()
  }, 400)
})

watch(rowsPerPage, () => {
  currentPage.value = 1
  load()
})

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
function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function copyText(txt: string) {
  if (!txt) return
  navigator.clipboard.writeText(txt)
  toast.success('IP address copied to clipboard!')
}
</script>

<style scoped>
.ml-page { background: #ebf2fe; padding: 24px 28px 40px; }

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
.btn-primary:hover:not(:disabled) { background: #fff; color: #3b82f6; transform: translate(-1px,-1px); box-shadow: 3px 3px 0 #1e293b; }
.btn-primary:active:not(:disabled) { transform: translate(0,0); box-shadow: none; }

.error-banner { background: #fef2f2; border: 1px solid #fca5a5; color: #dc2626; padding: 10px 14px; font-size: 13px; margin-bottom: 18px; display: flex; align-items: center; gap: 6px; }

/* ─── Table card ─── */
.table-card { background: #fff; border: 2px solid #1e293b; box-shadow: 2px 2px 0 #1e293b; }
.table-topbar { display: flex; justify-content: space-between; align-items: center; padding: 10px 14px; border-bottom: 1px solid #e2e8f0; gap: 12px; flex-wrap: wrap; }
.controls-left, .controls-right { display: flex; align-items: center; gap: 8px; }
.per-page-wrap { display: flex; align-items: center; gap: 6px; }
.ctrl-select { border: 1px solid #d1d5db; padding: 4px 6px; font-size: 13px; color: #374151; background: #fff; border-radius: 0; outline: none; }
.ctrl-label { font-size: 12px; color: #64748b; font-weight: 500; }
.search-input { border: 1px solid #d1d5db; padding: 4px 8px; font-size: 13px; color: #374151; width: 280px; outline: none; border-radius: 0; }
.search-input:focus { border-color: #3b82f6; }

.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.table-head-row { background: #3b82f6; }
.table-th { color: #fff; font-weight: 600; letter-spacing: 0.4px; padding: 10px 14px; text-align: left; cursor: pointer; white-space: nowrap; user-select: none; }
.table-th:hover { background: #2563eb; }
.sort-arrows { margin-left: 4px; font-size: 9px; opacity: .5; }
.sort-active { opacity: 1 !important; }
.table-row:nth-child(even) { background: #f8fafc; }
.table-row:hover { background: #eff6ff; }
.table-td { padding: 8px 14px; color: #374151; border-bottom: 1px solid #f1f5f9; font-size: 12px; }
.text-nowrap { white-space: nowrap; }
.text-right { text-align: right; }
.font-mono { font-family: 'Fira Code', var(--font-mono), monospace; }
.font-semibold { font-weight: 600; }
.text-break { word-break: break-all; }
.text-grey-7 { color: #475569; }

.ip-address { cursor: pointer; color: #3b82f6; font-weight: 700; border-bottom: 1px dashed #3b82f6; padding: 1px 0; }
.ip-address:hover { color: #1d4ed8; border-bottom-style: solid; }

.table-loading, .table-empty { text-align: center; padding: 40px; color: #94a3b8; font-size: 13px; }
.table-footer { display: flex; justify-content: space-between; align-items: center; padding: 10px 14px; border-top: 1px solid #e2e8f0; }
.showing-text { font-size: 12.5px; color: #64748b; }
.pagination { display: flex; gap: 3px; }
.pg-btn { height: 28px; padding: 0 10px; font-size: 10px; font-weight: 700; color: #374151; background: #fff; border: 1px solid #d1d5db; border-radius: 0; cursor: pointer; letter-spacing: 0.4px; text-transform: uppercase; white-space: nowrap; }
.pg-btn:hover:not(:disabled) { border-color: #1e293b; color: #1e293b; background: #f8fafc; }
.pg-btn:disabled { opacity: .35; cursor: default; }
.pg-active { background: #3b82f6 !important; color: #fff !important; border-color: #3b82f6 !important; }

.check-label { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 600; color: #374151; cursor: pointer; }
.check-label input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; accent-color: #3b82f6; }
</style>