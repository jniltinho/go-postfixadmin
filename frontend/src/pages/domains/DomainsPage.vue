<template>
  <div class="dom-page">

    <!-- ─── Header ─── -->
    <div class="dom-header">
      <div>
        <div class="dom-title">DOMAINS</div>
        <div class="dom-subtitle">MANAGE AND MONITOR YOUR REGISTERED EMAIL DOMAINS.</div>
      </div>
      <button class="btn-add-big" :disabled="!isSuperAdmin" @click="openAdd">
        <Icon name="plus-circle" :size="20" style="margin-right:12px;vertical-align:middle" />
        ADD DOMAIN
      </button>
    </div>


    <!-- ─── Table ─── -->
    <AppTable
      :rows="allDomains"
      :columns="columns"
      row-key="domain"
      :search-fields="['domain', 'description', 'transport']"
      default-sort-key="domain"
      :loading="loading"
      :show-actions="true"
      @edit="openEdit"
      @delete="confirmDelete"
    >
      <template #cell-domain="{ row }">
        <div class="cell-with-icon">
          <Icon name="globe" :size="14" class="row-icon" />
          {{ row.domain }}
        </div>
      </template>
      <template #cell-description="{ value }">{{ value || '—' }}</template>
      <template #cell-aliases="{ row }">
        <div class="progress-bar-wrap">
          <div class="progress-bar-fill" :style="{ width: (row.aliases > 0 ? Math.min(100, Math.round(((row.alias_count ?? 0) / row.aliases) * 100)) : 0) + '%' }"></div>
          <div class="progress-bar-text">{{ row.alias_count ?? 0 }} / {{ row.aliases }}</div>
        </div>
      </template>
      <template #cell-mailboxes="{ row }">
        <div class="progress-bar-wrap">
          <div class="progress-bar-fill" :style="{ width: (row.mailboxes > 0 ? Math.min(100, Math.round(((row.mailbox_count ?? 0) / row.mailboxes) * 100)) : 0) + '%' }"></div>
          <div class="progress-bar-text">{{ row.mailbox_count ?? 0 }} / {{ row.mailboxes }}</div>
        </div>
      </template>
      <template #cell-transport="{ value }"><span class="mono">{{ value || 'virtual' }}</span></template>
      <template #cell-backupmx="{ value }">{{ value ? 'Yes' : 'No' }}</template>
      <template #cell-active="{ value }">
        <span :class="value ? 'badge-yes' : 'badge-no'">{{ value ? 'YES' : 'NO' }}</span>
      </template>
      <template #cell-modified="{ value }">{{ formatDate(value) }}</template>
      <template #cell-password_expiry="{ value }"><span class="mono">{{ value ?? 0 }}</span></template>
      <template #actions="{ row }">
        <button
          class="act-btn act-edit"
          :class="{ 'act-disabled': !isSuperAdmin }"
          :disabled="!isSuperAdmin"
          @click="openEdit(row)"
        >
          <Icon name="pencil" :size="12" style="margin-right:4px;vertical-align:middle" />EDIT
        </button>
        <button
          class="act-btn act-del"
          :class="{ 'act-disabled': !isSuperAdmin }"
          :disabled="!isSuperAdmin"
          @click="confirmDelete(row)"
        >
          <Icon name="trash-2" :size="12" style="margin-right:4px;vertical-align:middle" />DELETE
        </button>
      </template>
    </AppTable>

    <!-- ══════════ ADD MODAL ══════════ -->
    <DomainAddModal
      v-model="showAdd"
      :transports="transports"
      :saving="savingAdd"
      @submit="handleAdd"
    />

    <!-- ══════════ EDIT MODAL ══════════ -->
    <DomainEditModal
      v-model="showEdit"
      :initial-data="editInitialData"
      :transports="transports"
      :saving="savingEdit"
      @submit="handleEdit"
    />

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <ConfirmDialog
      v-model="showDeleteConfirm"
      title="CONFIRM DELETE"
      :loading="deletingRow"
      @confirm="submitDelete"
    >
      <p class="text-sm font-medium text-brand-text leading-relaxed">
        Are you sure you want to delete domain<br />
        <strong class="font-mono break-all">{{ deleteTarget?.domain }}</strong>?<br />
        <span class="text-xs text-red-600 font-bold mt-1 block">This will delete all mailboxes, aliases, and related data.</span>
      </p>
    </ConfirmDialog>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import http from '../../utils/http'
import { useToastStore } from '../../stores/toast'
import { useAuthStore } from '../../stores/auth'
import DomainAddModal from './DomainAddModal.vue'
import DomainEditModal from './DomainEditModal.vue'

const toast = useToastStore()
const auth = useAuthStore()
const isSuperAdmin = computed(() => auth.user?.superadmin === true)

interface Domain {
  domain: string
  description: string
  aliases: number
  mailboxes: number
  maxquota: number
  quota: number
  transport: string
  backupmx: boolean
  active: boolean
  password_expiry: number | null
  created: string
  modified: string
  alias_count: number
  mailbox_count: number
}

interface Transport {
  id: number
  domain: string
  transport: string
  active: boolean
}

const allDomains = ref<Domain[]>([])
const transports = ref<Transport[]>([])
const loading = ref(true)

const columns = [
  { key: 'domain',          label: 'DOMAIN' },
  { key: 'description',     label: 'DESCRIPTION' },
  { key: 'aliases',         label: 'ALIASES' },
  { key: 'mailboxes',       label: 'EMAIL ACCOUNTS' },
  { key: 'transport',       label: 'TRANSPORT' },
  { key: 'backupmx',        label: 'BACKUP MX' },
  { key: 'active',          label: 'ACTIVE' },
  { key: 'modified',        label: 'MODIFIED' },
  { key: 'password_expiry', label: 'PASS EXPIRES' },
]

function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function load() {
  loading.value = true
  try {
    const domRes = await http.get(`${API_BASE}/domains`)
    allDomains.value = domRes.data?.data ?? []

    if (isSuperAdmin.value) {
      const trRes = await http.get(`${API_BASE}/transports`)
      transports.value = (trRes.data?.data ?? []).filter((t: Transport) => t.active)
    } else {
      transports.value = []
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load data')
  } finally { loading.value = false }
}
onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)

function openAdd() {
  if (!isSuperAdmin.value) {
    toast.error('Only superadmins can create domains')
    return
  }
  showAdd.value = true
}

async function handleAdd(payload: any) {
  if (!isSuperAdmin.value) {
    toast.error('Only superadmins can create domains')
    showAdd.value = false
    return
  }
  if (!payload.domain.trim()) { toast.error('Domain name is required'); return }
  savingAdd.value = true
  try {
    await http.post(`${API_BASE}/domains`, payload)
    showAdd.value = false
    toast.success(`Domain ${payload.domain} created successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to create domain')
  } finally { savingAdd.value = false }
}

// ─── Edit modal ───
const showEdit = ref(false)
const savingEdit = ref(false)
const editInitialData = ref<any>(null)
const editDomainName = ref('')

function openEdit(row: Domain) {
  if (!isSuperAdmin.value) {
    toast.error('Only superadmins can edit domains')
    return
  }
  editDomainName.value = row.domain
  editInitialData.value = {
    domain: row.domain,
    description: row.description || '',
    aliases: row.aliases,
    mailboxes: row.mailboxes,
    quotaMB: row.quota,
    passwordExpiry: row.password_expiry ?? null,
    transport: row.transport || 'virtual',
    active: row.active,
    backupmx: row.backupmx,
  }
  showEdit.value = true
}

async function handleEdit(payload: any) {
  if (!isSuperAdmin.value) {
    toast.error('Only superadmins can edit domains')
    showEdit.value = false
    return
  }
  savingEdit.value = true
  try {
    await http.put(`${API_BASE}/domains/${encodeURIComponent(editDomainName.value)}`, payload)
    showEdit.value = false
    toast.success(`Domain ${editDomainName.value} updated successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to update domain')
  } finally { savingEdit.value = false }
}

// ─── Delete ───
const showDeleteConfirm = ref(false)
const deletingRow = ref(false)
const deleteTarget = ref<Domain | null>(null)

function confirmDelete(row: Domain) {
  if (!isSuperAdmin.value) {
    toast.error('Only superadmins can delete domains')
    return
  }
  deleteTarget.value = row
  showDeleteConfirm.value = true
}

async function submitDelete() {
  if (!deleteTarget.value || !isSuperAdmin.value) return
  deletingRow.value = true
  try {
    const domain = deleteTarget.value.domain
    await http.delete(`${API_BASE}/domains/${encodeURIComponent(domain)}`)
    showDeleteConfirm.value = false; deleteTarget.value = null
    toast.success(`Domain ${domain} deleted successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to delete domain')
    showDeleteConfirm.value = false
  } finally { deletingRow.value = false }
}
</script>

<style scoped>
.dom-page { background: #ebf2fe; padding: 24px 28px 40px; }

.progress-bar-wrap {
  width: 128px; background-color: #e2e8f0; height: 24px;
  position: relative; border: 1px solid #1e293b;
  overflow: hidden; box-shadow: 1px 1px 0px #1e293b;
}
.progress-bar-fill {
  background-color: #22c55e; height: 100%;
  position: absolute; top: 0; left: 0; transition: width 0.3s ease;
}
.progress-bar-text {
  position: absolute; inset: 0; display: flex; align-items: center;
  justify-content: center; font-size: 11px; font-weight: 700; color: #374151; z-index: 10;
}
.act-disabled { opacity: .4; cursor: not-allowed; pointer-events: none; }
.btn-add-big:disabled { opacity: .55; cursor: not-allowed; transform: none; }
</style>
