<template>
  <div class="ml-page">
    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">MAIL DELIVERY LOGS</div>
        <div class="page-subtitle">MONITOR INCOMING AND OUTGOING SMTP TRAFFIC</div>
      </div>
      <div class="refresh-controls">
        <label class="check-label auto-refresh-chk">
          <input type="checkbox" v-model="autoRefresh" />
          <span>AUTO REFRESH (10s)</span>
        </label>
        <button class="btn-primary" @click="load(false)">
          <Icon name="refresh-cw" :size="16" style="margin-right: 0.375rem; vertical-align: middle" />
          REFRESH
        </button>
      </div>
    </div>

    <!-- ─── AppTable ─── -->
    <AppTable
      title="MAIL DELIVERY LOGS"
      :rows="allRows"
      :columns="columns"
      row-key="_key"
      :search-fields="['logdate', 'm_from', 'm_to', 'domain_from', 'domain_to', 'host_ip', 'host_name', 'helo']"
      default-sort-key="logdate"
      default-sort-dir="desc"
      :initial-rows-per-page="12"
      :loading="loading"
      :show-actions="false"
    >
      <template #cell-logdate="{ value }">
        <span class="font-mono text-nowrap">{{ value }}</span>
      </template>
      <template #cell-m_from="{ value }">
        <span class="font-semibold text-break">{{ value || '—' }}</span>
      </template>
      <template #cell-m_to="{ value }">
        <span class="font-semibold text-break">{{ value || '—' }}</span>
      </template>
      <template #cell-domain_from="{ value }">
        <span class="text-grey text-break">{{ value || '—' }}</span>
      </template>
      <template #cell-domain_to="{ value }">
        <span class="text-grey text-break">{{ value || '—' }}</span>
      </template>
      <template #cell-host_ip="{ value, row }">
        <span class="ip-address font-mono" @click="copyText(row.host_ip)" title="Click to copy IP">
          {{ value || '—' }}
        </span>
      </template>
      <template #cell-host_name="{ value }">
        <span class="font-mono text-grey">{{ value || '—' }}</span>
      </template>
      <template #cell-helo="{ value }">
        <span class="font-mono text-grey">{{ value || '—' }}</span>
      </template>
      <template #cell-msgsize="{ value }">
        <span class="font-mono text-right text-nowrap">{{ formatSize(value) }}</span>
      </template>
    </AppTable>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import http from '../../utils/http'
import { useToastStore } from '../../stores/toast'
import AppTable from '../../components/AppTable.vue'

const toast = useToastStore()

interface MailLogEntry {
  _key: string
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

const allRows = ref<MailLogEntry[]>([])
const loading = ref(true)
const autoRefresh = ref(false)

const columns = [
  { key: 'logdate', label: 'DATE/TIME' },
  { key: 'm_from', label: 'SENDER (FROM)' },
  { key: 'm_to', label: 'RECIPIENT (TO)' },
  { key: 'domain_from', label: 'DOMAIN FROM' },
  { key: 'domain_to', label: 'DOMAIN TO' },
  { key: 'host_ip', label: 'CLIENT IP' },
  { key: 'host_name', label: 'HOST' },
  { key: 'helo', label: 'HELO' },
  { key: 'msgsize', label: 'SIZE' },
]

async function load(quiet = false) {
  if (!quiet) loading.value = true
  try {
    const res = await http.get(`${API_BASE}/maillog`, {
      params: { page: 1, per_page: 9999, sort: 'tstamp', order: 'desc' },
    })
    const data = res.data?.data?.data ?? []
    allRows.value = data.map((log: any, idx: number) => ({
      _key: String(idx),
      logdate: log.logdate ?? '',
      m_from: log.m_from || '—',
      m_to: log.m_to || '—',
      domain_from: log.domain_from || '—',
      domain_to: log.domain_to || '—',
      host_ip: log.host_ip || '',
      host_name: log.host_name || '—',
      helo: log.helo || '—',
      msgsize: log.msgsize ?? 0,
    }))
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load delivery logs')
  } finally {
    loading.value = false
  }
}

function formatSize(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
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
</script>

<style scoped>
.ml-page {
  background: var(--color-page);
  padding: 1.5rem 1.75rem 2.5rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 1.25rem;
  gap: 1.25rem;
  flex-wrap: wrap;
}
.page-title {
  font-size: 1.75rem;
  font-weight: 900;
  color: var(--color-fg);
  letter-spacing: -0.03125rem;
  line-height: 1;
  font-family: monospace;
  text-transform: uppercase;
}
.page-subtitle {
  font-size: 0.625rem;
  color: var(--color-fg-subtle);
  letter-spacing: 0.05rem;
  margin-top: 0.375rem;
  text-transform: uppercase;
  font-weight: 700;
}

.refresh-controls {
  display: flex;
  align-items: center;
  gap: 0.875rem;
}
.auto-refresh-chk {
  font-size: 0.6875rem;
  font-weight: 800;
  color: var(--color-fg);
  font-family: monospace;
}

.btn-primary {
  background: var(--color-action);
  color: var(--color-on-action);
  border: 2px solid var(--color-line);
  padding: 0.625rem 1.125rem;
  font-size: 0.6875rem;
  font-weight: 800;
  letter-spacing: 0.0375rem;
  cursor: pointer;
  border-radius: 0;
  transition: all 0.15s;
  text-transform: uppercase;
  box-shadow: 2px 2px 0 var(--color-shadow);
  display: flex;
  align-items: center;
}
.btn-primary:hover {
  background: var(--color-surface);
  color: var(--color-action);
}
.btn-primary:active {
  transform: translate(0, 0);
  box-shadow: none;
}

.check-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-fg-2);
  cursor: pointer;
}
.check-label input[type='checkbox'] {
  width: 1.125rem;
  height: 1.125rem;
  cursor: pointer;
  accent-color: var(--color-action);
}

.font-mono {
  font-family: 'Fira Code', var(--font-mono), monospace;
}
.font-semibold {
  font-weight: 600;
}
.text-nowrap {
  white-space: nowrap;
}
.text-right {
  text-align: right;
  display: block;
}
.text-break {
  word-break: break-all;
}
.text-grey {
  color: var(--color-fg-2);
}

.ip-address {
  cursor: pointer;
  color: var(--color-action);
  font-weight: 700;
  border-bottom: 1px dashed var(--color-action);
  padding: 0.0625rem 0;
}
.ip-address:hover {
  color: var(--color-action-strong);
  border-bottom-style: solid;
}
</style>
