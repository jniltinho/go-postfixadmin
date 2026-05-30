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
            <Icon name="globe" :size="20" color="white" />
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
            <Icon name="mail" :size="20" color="white" />
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
            <Icon name="arrow-left-right" :size="20" color="white" />
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
      :initial-rows-per-page="10"
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
  { key: 'username',  label: 'ADMINISTRATOR' },
  { key: 'domain',    label: 'DOMAIN' },
  { key: 'action',    label: 'ACTION' },
  { key: 'data',      label: 'DESCRIPTION' },
]

async function loadDashboardData() {
  loading.value = true
  try {
    const res = await http.get(`${API_BASE}/dashboard`)
    const data = res.data?.data

    stats.value.domains   = data?.domains   ?? 0
    stats.value.mailboxes = data?.mailboxes ?? 0
    stats.value.aliases   = data?.aliases   ?? 0

    allRows.value = (data?.recent_logs ?? []).map((log: any, idx: number) => ({
      _key: String(idx),
      timestamp: log.timestamp ? new Date(log.timestamp).toLocaleString('pt-BR') : '-',
      username:  log.username  || '-',
      domain:    log.domain    || '-',
      action:    log.action    || '-',
      data:      log.data      || '',
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
.cta-card--disabled {
  opacity: 0.58;
  cursor: not-allowed;
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
.cta-btn--disabled,
.cta-btn--disabled:hover {
  background: #e5e7eb;
  color: #64748b;
  cursor: not-allowed;
  transform: none;
  box-shadow: 2px 2px 0px #1e293b;
}

/* ─── Table Controls & Custom Elements ─── */
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

.td-link { color: #3b82f6; font-weight: 700; }
.activity-action { font-weight: 900; text-transform: uppercase; }
.activity-data { word-break: break-all; }

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
