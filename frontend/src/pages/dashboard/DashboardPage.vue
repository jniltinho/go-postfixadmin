<template>
  <div class="dash-page">
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
            <Icon name="globe" :size="20" class="text-on-action" />
          </div>
          <span class="stat-badge badge-green">ACTIVE</span>
        </div>
        <p class="stat-label">TOTAL DOMAINS</p>
        <div class="stat-value-row">
          <div v-if="loading" class="skeleton w-16 h-10" />
          <span v-else class="stat-value">{{ stats.domains }}</span>
          <span class="stat-sub">AVAILABLE</span>
        </div>
      </div>

      <!-- Email Accounts -->
      <div class="stat-card">
        <div class="stat-card-top">
          <div class="stat-icon-wrap">
            <Icon name="mail" :size="20" class="text-on-action" />
          </div>
          <span class="stat-badge badge-green">IN USE</span>
        </div>
        <p class="stat-label">EMAIL ACCOUNTS</p>
        <div class="stat-value-row">
          <div v-if="loading" class="skeleton w-16 h-10" />
          <span v-else class="stat-value">{{ stats.mailboxes }}</span>
          <span class="stat-sub">ACCOUNTS</span>
        </div>
      </div>

      <!-- Total Aliases -->
      <div class="stat-card">
        <div class="stat-card-top">
          <div class="stat-icon-wrap">
            <Icon name="arrow-left-right" :size="20" class="text-on-action" />
          </div>
          <span class="stat-badge badge-green">ACTIVE</span>
        </div>
        <p class="stat-label">TOTAL ALIASES</p>
        <div class="stat-value-row">
          <div v-if="loading" class="skeleton w-16 h-10" />
          <span v-else class="stat-value">{{ stats.aliases }}</span>
          <span class="stat-sub">ALIASES</span>
        </div>
      </div>

      <!-- New Domain CTA -->
      <div class="cta-card cta-orange" :class="{ 'cta-card--disabled': !isSuperAdmin }">
        <div>
          <h3 class="cta-title">NEW DOMAIN?</h3>
          <p class="cta-sub">CREATE A NEW DOMAIN IN SECONDS.</p>
        </div>
        <router-link v-if="isSuperAdmin" to="/domains" class="cta-btn">
          ADD DOMAIN
          <Icon name="plus-circle" :size="16" />
        </router-link>
        <button v-else type="button" class="cta-btn cta-btn--disabled" disabled>
          ADD DOMAIN
          <Icon name="plus-circle" :size="16" />
        </button>
      </div>

      <!-- New Email CTA -->
      <div class="cta-card cta-blue">
        <div>
          <h3 class="cta-title">NEW EMAIL?</h3>
          <p class="cta-sub">CREATE A NEW ACCOUNT IN SECONDS.</p>
        </div>
        <router-link to="/mailboxes" class="cta-btn">
          ADD EMAIL
          <Icon name="plus-circle" :size="16" />
        </router-link>
      </div>
    </div>

    <!-- ─── Activity table ─── -->
    <AppTable
      title="RECENT ACTIVITY"
      :rows="allRows"
      :columns="columns"
      row-key="_key"
      :search-fields="['timestamp', 'username', 'domain', 'action', 'data']"
      default-sort-key="timestamp"
      default-sort-dir="desc"
      :initial-rows-per-page="12"
      :loading="loading"
      :show-actions="false"
    >
      <template #title-action>
        <router-link to="/logs" class="view-logs-link">
          VIEW LOGS
          <Icon name="arrow-right" :size="16" />
        </router-link>
      </template>
      <template #cell-username="{ value }">
        <span class="td-link">{{ value }}</span>
      </template>
      <template #cell-domain="{ value }">
        <span class="mono">{{ value }}</span>
      </template>
      <template #cell-action="{ value }">
        <span class="activity-action">{{ value }}</span>
      </template>
      <template #cell-data="{ value }">
        <span class="mono activity-data">{{ value || '—' }}</span>
      </template>
    </AppTable>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import http from '../../utils/http'
import { useToastStore } from '../../stores/toast'
import { useAuthStore } from '../../stores/auth'
import AppTable from '../../components/AppTable.vue'

const toast = useToastStore()
const auth = useAuthStore()
const isSuperAdmin = computed(() => auth.user?.superadmin === true)
const loading = ref(true)

const stats = ref({ domains: 0, mailboxes: 0, aliases: 0 })
const allRows = ref<any[]>([])

const columns = [
  { key: 'timestamp', label: 'DATE/TIME' },
  { key: 'username', label: 'ADMINISTRATOR' },
  { key: 'domain', label: 'DOMAIN' },
  { key: 'action', label: 'ACTION' },
  { key: 'data', label: 'DESCRIPTION' },
]

async function loadDashboardData() {
  loading.value = true
  try {
    const res = await http.get(`${API_BASE}/dashboard`)
    const data = res.data?.data

    stats.value.domains = data?.domains ?? 0
    stats.value.mailboxes = data?.mailboxes ?? 0
    stats.value.aliases = data?.aliases ?? 0

    allRows.value = (data?.recent_logs ?? []).map((log: any, idx: number) => ({
      _key: String(idx),
      timestamp: log.timestamp ? new Date(log.timestamp).toLocaleString('pt-BR') : '-',
      username: log.username || '-',
      domain: log.domain || '-',
      action: log.action || '-',
      data: log.data || '',
    }))
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load dashboard data')
  } finally {
    loading.value = false
  }
}

onMounted(loadDashboardData)
</script>

<style scoped>
.dash-page {
  background: var(--color-page);
  padding: 1.5rem 1.75rem 2.5rem;
}

/* ─── Page heading ─── */
.dash-header {
  margin-bottom: 1.125rem;
}
.dash-title {
  font-size: 1.625rem;
  font-weight: 900;
  color: var(--color-fg);
  font-family: monospace;
  letter-spacing: -0.03125rem;
  line-height: 1;
  text-transform: uppercase;
  margin: 0 0 0.25rem;
}
.dash-subtitle {
  font-size: 0.6875rem;
  font-weight: 700;
  color: var(--color-fg-subtle);
  letter-spacing: 0.05rem;
  margin: 0;
  text-transform: uppercase;
}

/* ─── Stats row ─── */
.stats-row {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 1rem;
  margin-bottom: 1.125rem;
}

/* Stat card */
.stat-card {
  background: var(--color-surface);
  border: 2px solid var(--color-line);
  padding: 1.125rem 1.125rem 0.875rem;
  box-shadow: var(--shadow-2);
  cursor: pointer;
  transition:
    transform 0.2s,
    box-shadow 0.2s;
}
.stat-card:hover {
  transform: translate(-1px, -1px);
  box-shadow: var(--shadow-3);
}

.stat-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.625rem;
}

.stat-icon-wrap {
  width: 2.5rem;
  height: 2.5rem;
  background: var(--color-action);
  border: 2px solid var(--color-line);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-badge {
  font-size: 0.5625rem;
  font-weight: 900;
  letter-spacing: 0.0375rem;
  padding: 0.125rem 0.4375rem;
  color: var(--color-on-action);
  border: 2px solid var(--color-line);
  text-transform: uppercase;
}
.badge-green {
  background: var(--color-ok);
}

.stat-label {
  font-size: 0.625rem;
  font-weight: 900;
  color: var(--color-fg-subtle);
  letter-spacing: 0.0375rem;
  text-transform: uppercase;
  margin: 0 0 0.125rem;
}

.stat-value-row {
  display: flex;
  align-items: baseline;
  gap: 0.375rem;
}
.stat-value {
  font-size: 2.25rem;
  font-weight: 900;
  color: var(--color-fg);
  font-family: monospace;
  line-height: 1.1;
}
.stat-sub {
  font-size: 0.625rem;
  font-weight: 700;
  color: var(--color-fg-subtle);
  letter-spacing: 0.03125rem;
  text-transform: uppercase;
}

/* CTA cards */
.cta-card {
  border: 2px solid var(--color-line);
  padding: 1.125rem 1.125rem 0.875rem;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  box-shadow: var(--shadow-3);
}
.cta-card--disabled {
  opacity: 0.58;
  cursor: not-allowed;
}
.cta-orange {
  background: var(--color-warn);
}
.cta-blue {
  background: var(--color-action);
}

.cta-title {
  font-size: 0.875rem;
  font-weight: 900;
  color: var(--color-on-action);
  letter-spacing: 0.01875rem;
  text-transform: uppercase;
  margin: 0 0 0.25rem;
}
.cta-sub {
  font-size: 0.6875rem;
  font-weight: 700;
  color: color-mix(in srgb, var(--color-on-action) 85%, transparent);
  margin: 0 0 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.01875rem;
}
.cta-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.3125rem;
  background: var(--color-surface);
  color: var(--color-fg);
  font-size: 0.6875rem;
  font-weight: 900;
  letter-spacing: 0.03125rem;
  text-transform: uppercase;
  padding: 0.4375rem 0.75rem;
  text-decoration: none;
  border: 2px solid var(--color-line);
  box-shadow: var(--shadow-2);
  transition:
    transform 0.15s,
    box-shadow 0.15s;
}
.cta-btn:hover {
  transform: translate(-1px, -1px);
  box-shadow: var(--shadow-3);
}
.cta-btn--disabled,
.cta-btn--disabled:hover {
  background: var(--color-line-soft);
  color: var(--color-fg-muted);
  cursor: not-allowed;
  transform: none;
  box-shadow: var(--shadow-2);
}

/* ─── Table Controls & Custom Elements ─── */
.view-logs-link {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.6875rem;
  font-weight: 900;
  color: var(--color-action);
  text-decoration: none;
  letter-spacing: 0.03125rem;
  text-transform: uppercase;
}
.view-logs-link:hover {
  text-decoration: underline;
}

.td-link {
  color: var(--color-action);
  font-weight: 700;
}
.activity-action {
  font-weight: 900;
  text-transform: uppercase;
}
.activity-data {
  word-break: break-all;
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
