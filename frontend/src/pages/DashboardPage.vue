<template>
  <q-page class="dash-page">

    <!-- Page title -->
    <div class="dash-header">
      <h2 class="dash-title">OVERVIEW</h2>
      <p class="dash-subtitle">MONITORING YOUR EMAIL SERVER INFRASTRUCTURE.</p>
    </div>

    <!-- ─── Stats row ─── -->
    <div class="stats-row">

      <!-- Total Domains -->
      <div class="stat-card">
        <div class="stat-card-top">
          <div class="stat-icon-wrap">
            <q-icon name="public" size="20px" color="white" />
          </div>
          <span class="stat-badge badge-green">ACTIVE</span>
        </div>
        <p class="stat-label">TOTAL DOMAINS</p>
        <div class="stat-value-row">
          <q-skeleton v-if="loading" type="text" width="2rem" height="40px" />
          <span v-else class="stat-value">{{ stats.domains }}</span>
          <span class="stat-sub">AVAILABLE</span>
        </div>
      </div>

      <!-- Email Accounts -->
      <div class="stat-card">
        <div class="stat-card-top">
          <div class="stat-icon-wrap">
            <q-icon name="mail" size="20px" color="white" />
          </div>
          <span class="stat-badge badge-green">IN USE</span>
        </div>
        <p class="stat-label">EMAIL ACCOUNTS</p>
        <div class="stat-value-row">
          <q-skeleton v-if="loading" type="text" width="2rem" height="40px" />
          <span v-else class="stat-value">{{ stats.mailboxes }}</span>
          <span class="stat-sub">ACCOUNTS</span>
        </div>
      </div>

      <!-- Total Aliases -->
      <div class="stat-card">
        <div class="stat-card-top">
          <div class="stat-icon-wrap">
            <q-icon name="swap_horiz" size="20px" color="white" />
          </div>
          <span class="stat-badge badge-green">ACTIVE</span>
        </div>
        <p class="stat-label">TOTAL ALIASES</p>
        <div class="stat-value-row">
          <q-skeleton v-if="loading" type="text" width="2rem" height="40px" />
          <span v-else class="stat-value">{{ stats.aliases }}</span>
          <span class="stat-sub">ALIASES</span>
        </div>
      </div>

      <!-- New Domain CTA -->
      <div class="cta-card cta-orange">
        <div>
          <h3 class="cta-title">NEW DOMAIN?</h3>
          <p class="cta-sub">CREATE A NEW DOMAIN IN SECONDS.</p>
        </div>
        <router-link to="/domains" class="cta-btn">
          ADD DOMAIN
          <q-icon name="add_circle_outline" size="16px" />
        </router-link>
      </div>

      <!-- New Email CTA -->
      <div class="cta-card cta-blue">
        <div>
          <h3 class="cta-title">NEW EMAIL?</h3>
          <p class="cta-sub">CREATE A NEW ACCOUNT IN SECONDS.</p>
        </div>
        <router-link to="/mailboxes" class="cta-btn">
          ADD EMAIL
          <q-icon name="add_circle_outline" size="16px" />
        </router-link>
      </div>

    </div>

    <!-- Error -->
    <div v-if="error" class="error-banner">
      <q-icon name="warning" size="16px" /> {{ error }}
    </div>

    <!-- ─── Activity table ─── -->
    <div class="table-card">

      <!-- Table header bar -->
      <div class="table-topbar">
        <div class="table-topbar-title">RECENT ACTIVITY</div>
        <router-link to="/logs" class="view-logs-link">
          VIEW LOGS
          <q-icon name="arrow_forward" size="16px" />
        </router-link>
      </div>

      <!-- Controls: entries per page + search -->
      <div class="table-controls">
        <div class="entries-control">
          <select v-model="rowsPerPage" class="entries-select">
            <option :value="10">10</option>
            <option :value="25">25</option>
            <option :value="50">50</option>
          </select>
          <span class="entries-text">entries per page</span>
        </div>
        <div class="search-control">
          <span class="search-label">Search:</span>
          <input v-model="search" class="search-input" type="text" />
        </div>
      </div>

      <!-- Table -->
      <div class="table-wrap">
        <table class="activity-table">
          <thead>
            <tr class="table-head-row">
              <th v-for="col in columns" :key="col.key" @click="sortBy(col.key)" class="table-th">
                {{ col.label }}
                <span class="sort-arrows">
                  <span :class="sortKey === col.key && sortDir === 'asc' ? 'sort-active' : ''">▲</span>
                  <span :class="sortKey === col.key && sortDir === 'desc' ? 'sort-active' : ''">▼</span>
                </span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td :colspan="columns.length" class="table-loading">
                <q-spinner color="primary" size="24px" />
              </td>
            </tr>
            <tr v-else-if="pagedRows.length === 0">
              <td :colspan="columns.length" class="table-empty">
                <q-icon name="warning" size="16px" /> No data available
              </td>
            </tr>
            <tr v-for="(row, i) in pagedRows" :key="i" class="table-row">
              <td class="table-td">{{ row.timestamp }}</td>
              <td class="table-td td-link">{{ row.username }}</td>
              <td class="table-td">{{ row.domain }}</td>
              <td class="table-td td-bold">{{ row.action }}</td>
              <td class="table-td">{{ row.data }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Table footer: showing + pagination -->
      <div class="table-footer">
        <div class="showing-text">
          <template v-if="filteredRows.length === 0">
            Showing 0 entries
          </template>
          <template v-else>
            Showing {{ (currentPage - 1) * rowsPerPage + 1 }} to
            {{ Math.min(currentPage * rowsPerPage, filteredRows.length) }} of
            {{ filteredRows.length }} entries
          </template>
        </div>
        <div class="pagination">
          <button class="pg-btn" :disabled="currentPage === 1" @click="currentPage = 1">«</button>
          <button class="pg-btn" :disabled="currentPage === 1" @click="currentPage--">‹</button>
          <button
            v-for="p in pageButtons"
            :key="p"
            class="pg-btn"
            :class="{ 'pg-active': p === currentPage }"
            @click="currentPage = p"
          >{{ p }}</button>
          <button class="pg-btn" :disabled="currentPage === totalPages" @click="currentPage++">›</button>
          <button class="pg-btn" :disabled="currentPage === totalPages" @click="currentPage = totalPages">»</button>
        </div>
      </div>

    </div>
  </q-page>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import axios from 'axios'

const loading = ref(true)
const error = ref('')

const stats = ref({ domains: 0, mailboxes: 0, aliases: 0 })
const allRows = ref<any[]>([])

const search = ref('')
const rowsPerPage = ref(10)
const currentPage = ref(1)
const sortKey = ref('timestamp')
const sortDir = ref<'asc' | 'desc'>('desc')

const columns = [
  { key: 'timestamp', label: 'DATE/TIME' },
  { key: 'username',  label: 'ADMINISTRATOR' },
  { key: 'domain',    label: 'DOMAIN' },
  { key: 'action',    label: 'ACTION' },
  { key: 'data',      label: 'DESCRIPTION' },
]

function sortBy(key: string) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'asc'
  }
}

const filteredRows = computed(() => {
  const q = search.value.toLowerCase()
  const rows = q
    ? allRows.value.filter(r =>
        Object.values(r).some(v => String(v).toLowerCase().includes(q))
      )
    : allRows.value

  return [...rows].sort((a, b) => {
    const av = String(a[sortKey.value] ?? '')
    const bv = String(b[sortKey.value] ?? '')
    return sortDir.value === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredRows.value.length / rowsPerPage.value)))

const pageButtons = computed(() => {
  const total = totalPages.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const cur = currentPage.value
  const pages = new Set([1, total, cur, cur - 1, cur + 1].filter(p => p >= 1 && p <= total))
  return Array.from(pages).sort((a, b) => a - b)
})

const pagedRows = computed(() => {
  const start = (currentPage.value - 1) * rowsPerPage.value
  return filteredRows.value.slice(start, start + rowsPerPage.value)
})

watch([search, rowsPerPage], () => { currentPage.value = 1 })

async function loadDashboardData() {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get(`${API_BASE}/dashboard`)
    const data = res.data?.data

    stats.value.domains   = data?.domains   ?? 0
    stats.value.mailboxes = data?.mailboxes ?? 0
    stats.value.aliases   = data?.aliases   ?? 0

    allRows.value = (data?.recent_logs ?? []).map((log: any) => ({
      timestamp: log.timestamp ? new Date(log.timestamp).toLocaleString('pt-BR') : '-',
      username:  log.username  || '-',
      domain:    log.domain    || '-',
      action:    log.action    || '-',
      data:      log.data      || '',
    }))
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load dashboard data'
  } finally {
    loading.value = false
  }
}

onMounted(loadDashboardData)
</script>

<style scoped>
.dash-page {
  background: #ebf2fe;
  padding: 24px 28px 40px;
}

/* ─── Page heading ─── */
.dash-header { margin-bottom: 18px; }
.dash-title {
  font-size: 26px;
  font-weight: 900;
  color: #1e293b;
  font-family: monospace;
  letter-spacing: -0.5px;
  line-height: 1;
  text-transform: uppercase;
  margin: 0 0 4px;
}
.dash-subtitle {
  font-size: 11px;
  font-weight: 700;
  color: #94a3b8;
  letter-spacing: 0.8px;
  margin: 0;
  text-transform: uppercase;
}

/* ─── Stats row ─── */
.stats-row {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 16px;
  margin-bottom: 18px;
}

/* Stat card */
.stat-card {
  background: #ffffff;
  border: 2px solid #1e293b;
  padding: 18px 18px 14px;
  box-shadow: 2px 2px 0px #1e293b;
  cursor: pointer;
  transition: transform .2s, box-shadow .2s;
}
.stat-card:hover {
  transform: translate(-1px, -1px);
  box-shadow: 3px 3px 0px #1e293b;
}

.stat-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.stat-icon-wrap {
  width: 40px;
  height: 40px;
  background: #3b82f6;
  border: 2px solid #1e293b;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-badge {
  font-size: 9px;
  font-weight: 900;
  letter-spacing: 0.6px;
  padding: 2px 7px;
  color: #fff;
  border: 2px solid #1e293b;
  text-transform: uppercase;
}
.badge-green { background: #22c55e; }

.stat-label {
  font-size: 10px;
  font-weight: 900;
  color: #94a3b8;
  letter-spacing: 0.6px;
  text-transform: uppercase;
  margin: 0 0 2px;
}

.stat-value-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
}
.stat-value {
  font-size: 36px;
  font-weight: 900;
  color: #1e293b;
  font-family: monospace;
  line-height: 1.1;
}
.stat-sub {
  font-size: 10px;
  font-weight: 700;
  color: #94a3b8;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

/* CTA cards */
.cta-card {
  border: 2px solid #1e293b;
  padding: 18px 18px 14px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  box-shadow: 3px 3px 0px #1e293b;
}
.cta-orange { background: #f97316; }
.cta-blue   { background: #3b82f6; }

.cta-title {
  font-size: 14px;
  font-weight: 900;
  color: #fff;
  letter-spacing: 0.3px;
  text-transform: uppercase;
  margin: 0 0 4px;
}
.cta-sub {
  font-size: 11px;
  font-weight: 700;
  color: rgba(255, 255, 255, .85);
  margin: 0 0 14px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}
.cta-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  background: #fff;
  color: #1e293b;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  padding: 7px 12px;
  text-decoration: none;
  border: 2px solid #1e293b;
  box-shadow: 2px 2px 0px #1e293b;
  transition: transform .15s, box-shadow .15s;
}
.cta-btn:hover {
  transform: translate(-1px, -1px);
  box-shadow: 3px 3px 0px #1e293b;
}

/* Error banner */
.error-banner {
  background: #fef2f2;
  border: 2px solid #fca5a5;
  color: #dc2626;
  padding: 10px 14px;
  font-size: 13px;
  margin-bottom: 18px;
  display: flex;
  align-items: center;
  gap: 6px;
}

/* ─── Table card ─── */
.table-card {
  background: #ffffff;
  border: 2px solid #1e293b;
  box-shadow: 2px 2px 0px #1e293b;
}

.table-topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px 10px;
  border-bottom: 2px solid #1e293b;
}
.table-topbar-title {
  font-size: 13px;
  font-weight: 900;
  color: #1e293b;
  letter-spacing: 0.6px;
  text-transform: uppercase;
}
.view-logs-link {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 900;
  color: #3b82f6;
  text-decoration: none;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}
.view-logs-link:hover { text-decoration: underline; }

/* Controls */
.table-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  border-bottom: 1px solid #f1f5f9;
}
.entries-control {
  display: flex;
  align-items: center;
  gap: 6px;
}
.entries-select {
  border: 1px solid #d1d5db;
  padding: 2px 4px;
  font-size: 13px;
  color: #374151;
  background: #fff;
}
.entries-text { font-size: 13px; color: #64748b; }

.search-control {
  display: flex;
  align-items: center;
  gap: 8px;
}
.search-label { font-size: 13px; color: #374151; font-weight: 500; }
.search-input {
  border: 1px solid #d1d5db;
  padding: 4px 8px;
  font-size: 13px;
  color: #374151;
  width: 180px;
  outline: none;
}
.search-input:focus { border-color: #3b82f6; }

/* Table */
.table-wrap { overflow-x: auto; }
.activity-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12.5px;
  border: 2px solid #1e293b;
}
.table-head-row {
  background: #3b82f6;
  border-bottom: 2px solid #1e293b;
}
.table-th {
  color: #fff;
  font-weight: 900;
  letter-spacing: 0.6px;
  text-transform: uppercase;
  padding: 10px 14px;
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
  user-select: none;
  font-size: 11px;
}
.table-th:hover { background: #2563eb; }
.sort-arrows {
  margin-left: 4px;
  font-size: 9px;
  opacity: .6;
}
.sort-active { opacity: 1 !important; }

.table-row:nth-child(even) { background: #f8fafc; }
.table-row:hover { background: #eff6ff; }
.table-td {
  padding: 9px 14px;
  color: #374151;
  border-bottom: 1px solid #f1f5f9;
  white-space: nowrap;
}
.td-link { color: #3b82f6; }
.td-bold { font-weight: 700; color: #1e293b; }

.table-loading,
.table-empty {
  text-align: center;
  padding: 24px;
  color: #94a3b8;
  font-size: 13px;
}

/* Footer */
.table-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  border-top: 1px solid #e2e8f0;
}
.showing-text { font-size: 12.5px; color: #64748b; }

.pagination { display: flex; gap: 3px; }
.pg-btn {
  min-width: 28px;
  height: 28px;
  padding: 0 6px;
  font-size: 12px;
  font-weight: 600;
  color: #374151;
  background: #fff;
  border: 1px solid #d1d5db;
  cursor: pointer;
  transition: all .12s;
}
.pg-btn:hover:not(:disabled) { border-color: #3b82f6; color: #3b82f6; }
.pg-btn:disabled { opacity: .4; cursor: default; }
.pg-active {
  background: #3b82f6 !important;
  color: #fff !important;
  border-color: #1e293b !important;
}

/* Responsive */
@media (max-width: 1100px) {
  .stats-row {
    grid-template-columns: repeat(3, 1fr);
  }
}
@media (max-width: 700px) {
  .stats-row {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
