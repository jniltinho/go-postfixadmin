<template>
  <div class="tp-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">TRANSPORT LIST</div>
        <div class="page-subtitle">MANAGE MAIL TRANSPORT RULES</div>
      </div>
      <button class="btn-primary" @click="showAdd = true">
        <Icon name="plus-circle" :size="16" style="margin-right:6px;vertical-align:middle" />
        ADD TRANSPORT
      </button>
    </div>

    <!-- ─── Table ─── -->
    <AppTable
      :rows="allTransports"
      :columns="columns"
      row-key="id"
      :search-fields="['domain', 'transport']"
      default-sort-key="domain"
      :loading="loading"
      @edit="openEdit"
      @delete="confirmDelete"
    >
      <template #cell-transport="{ row }">
        <div class="cell-with-icon">
          <Icon name="arrow-up-down" :size="14" class="row-icon" />
          <span class="mono-text">{{ row.transport }}</span>
        </div>
      </template>
      <template #cell-created="{ value }">{{ formatDate(value) }}</template>
      <template #cell-modified="{ value }">{{ formatDate(value) }}</template>
      <template #cell-active="{ value }">
        <span :class="value ? 'badge-yes' : 'badge-no'">{{ value ? 'YES' : 'NO' }}</span>
      </template>
    </AppTable>

    <!-- ══════════ ADD MODAL ══════════ -->
    <TransportAddModal
      v-model="showAdd"
      :saving="savingAdd"
      @submit="handleAdd"
    />

    <!-- ══════════ EDIT MODAL ══════════ -->
    <TransportEditModal
      v-model="showEdit"
      :initial-data="editInitialData"
      :saving="savingEdit"
      :loading="loadingEdit"
      @submit="handleEdit"
    />

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <ConfirmDialog
      v-model="showDeleteConfirm"
      title="CONFIRM DELETE"
      :item-name="deleteTarget?.domain"
      :loading="deletingRow"
      @confirm="submitDelete"
    />

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import http from '../../utils/http'
import { useToastStore } from '../../stores/toast'
import TransportAddModal from './TransportAddModal.vue'
import TransportEditModal from './TransportEditModal.vue'

const toast = useToastStore()

interface Transport {
  id: number
  domain: string
  transport: string
  active: boolean
  created: string
  modified: string
}

const allTransports = ref<Transport[]>([])
const loading = ref(true)

const columns = [
  { key: 'domain',    label: 'DOMAIN' },
  { key: 'transport', label: 'TRANSPORT' },
  { key: 'created',   label: 'CREATED' },
  { key: 'modified',  label: 'MODIFIED' },
  { key: 'active',    label: 'ACTIVE' },
]

function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function load() {
  loading.value = true
  try {
    const res = await http.get(`${API_BASE}/transports`)
    allTransports.value = res.data?.data ?? []
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load transports')
  } finally { loading.value = false }
}
onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)

async function handleAdd(payload: { domain: string; transport: string; active: boolean }) {
  if (!payload.domain.trim()) { toast.error('Domain is required'); return }
  if (!payload.transport.trim()) { toast.error('Transport is required'); return }
  savingAdd.value = true
  try {
    await http.post(`${API_BASE}/transports`, payload)
    showAdd.value = false
    toast.success(`Transport for ${payload.domain} created successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to create transport')
  } finally { savingAdd.value = false }
}

// ─── Edit modal ───
const showEdit = ref(false)
const loadingEdit = ref(false)
const savingEdit = ref(false)
const editInitialData = ref<{ id: number; domain: string; transport: string; active: boolean } | null>(null)

async function openEdit(row: Transport) {
  editInitialData.value = { id: row.id, domain: row.domain, transport: row.transport, active: row.active }
  loadingEdit.value = true
  showEdit.value = true
  try {
    const res = await http.get(`${API_BASE}/transports/${row.id}`)
    const t = res.data?.data
    editInitialData.value = { id: t.id, domain: t.domain, transport: t.transport, active: t.active }
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load transport')
  } finally { loadingEdit.value = false }
}

async function handleEdit(payload: { domain: string; transport: string; active: boolean }) {
  if (!payload.domain.trim()) { toast.error('Domain is required'); return }
  if (!payload.transport.trim()) { toast.error('Transport is required'); return }
  const id = editInitialData.value?.id
  savingEdit.value = true
  try {
    await http.put(`${API_BASE}/transports/${id}`, payload)
    showEdit.value = false
    toast.success(`Transport for ${payload.domain} updated successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to update transport')
  } finally { savingEdit.value = false }
}

// ─── Delete ───
const showDeleteConfirm = ref(false)
const deletingRow = ref(false)
const deleteTarget = ref<Transport | null>(null)

function confirmDelete(row: Transport) { deleteTarget.value = row; showDeleteConfirm.value = true }

async function submitDelete() {
  if (!deleteTarget.value) return
  deletingRow.value = true
  try {
    const { id, domain } = deleteTarget.value
    await http.delete(`${API_BASE}/transports/${id}`)
    showDeleteConfirm.value = false; deleteTarget.value = null
    toast.success(`Transport for ${domain} deleted successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to delete transport')
    showDeleteConfirm.value = false
  } finally { deletingRow.value = false }
}
</script>

<style scoped>
.tp-page { background: #ebf2fe; padding: 24px 28px 40px; }
.mono-text { font-family: monospace; font-size: 12px; color: #1e293b; font-weight: 600; }
</style>
