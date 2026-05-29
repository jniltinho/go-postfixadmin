<template>
  <div class="dom-page">

    <!-- ─── Header ─── -->
    <div class="dom-header">
      <div>
        <div class="dom-title">DOMAIN ALIASES</div>
        <div class="dom-subtitle">MANAGE ALIAS DOMAIN MAPPINGS</div>
      </div>
      <button class="btn-add-big" @click="showAdd = true">
        <Icon name="plus-circle" :size="20" style="margin-right:12px;vertical-align:middle" />
        ADD DOMAIN ALIAS
      </button>
    </div>

    <!-- ─── Table ─── -->
    <AppTable
      :rows="allAliasDomains"
      :columns="columns"
      row-key="alias_domain"
      :search-fields="['alias_domain', 'target_domain']"
      default-sort-key="alias_domain"
      :loading="loading"
      @edit="openEdit"
      @delete="confirmDelete"
    >
      <template #cell-alias_domain="{ row }">
        <div class="cell-with-icon">
          <Icon name="arrow-left-right" :size="14" class="row-icon" />
          {{ row.alias_domain }}
        </div>
      </template>
      <template #cell-target_domain="{ value }"><span class="mono">{{ value }}</span></template>
      <template #cell-active="{ value }">
        <span :class="value ? 'badge-yes' : 'badge-no'">{{ value ? 'YES' : 'NO' }}</span>
      </template>
      <template #cell-modified="{ value }">{{ formatDate(value) }}</template>
    </AppTable>

    <!-- ══════════ ADD MODAL ══════════ -->
    <AliasDomainAddModal
      v-model="showAdd"
      :domains="domains"
      :available-alias="availableAliasDomains"
      :saving="savingAdd"
      @submit="handleAdd"
    />

    <!-- ══════════ EDIT MODAL ══════════ -->
    <AliasDomainEditModal
      v-model="showEdit"
      :domains="domains"
      :initial-data="editInitialData"
      :saving="savingEdit"
      :loading="loadingEdit"
      @submit="handleEdit"
    />

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <ConfirmDialog
      v-model="showDeleteConfirm"
      title="CONFIRM DELETE"
      :item-name="deleteTarget?.alias_domain"
      :loading="deletingRow"
      @confirm="submitDelete"
    />

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import http from '../../utils/http'
import { useToastStore } from '../../stores/toast'
import AliasDomainAddModal from './AliasDomainAddModal.vue'
import AliasDomainEditModal from './AliasDomainEditModal.vue'

const toast = useToastStore()

interface AliasDomain {
  alias_domain: string
  target_domain: string
  active: boolean
  created: string
  modified: string
}

interface Domain { domain: string }

const allAliasDomains = ref<AliasDomain[]>([])
const domains = ref<Domain[]>([])
const loading = ref(true)

const availableAliasDomains = computed(() => {
  const used = new Set(allAliasDomains.value.map(a => a.alias_domain))
  return domains.value.filter(d => !used.has(d.domain))
})

const columns = [
  { key: 'alias_domain',  label: 'ALIAS DOMAIN' },
  { key: 'target_domain', label: 'TARGET DOMAIN' },
  { key: 'active',        label: 'ACTIVE' },
  { key: 'modified',      label: 'MODIFIED' },
]

function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function load() {
  loading.value = true
  try {
    const [adRes, domRes] = await Promise.all([
      http.get(`${API_BASE}/alias-domains`),
      http.get(`${API_BASE}/domains`),
    ])
    allAliasDomains.value = adRes.data?.data ?? []
    domains.value = domRes.data?.data ?? []
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load alias domains')
  } finally { loading.value = false }
}
onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)

async function handleAdd(payload: { alias_domain: string; target_domain: string; active: boolean }) {
  if (!payload.alias_domain) { toast.error('Please select a source alias domain'); return }
  if (!payload.target_domain) { toast.error('Please select a target domain'); return }
  if (payload.alias_domain === payload.target_domain) { toast.error('Alias domain and target domain must be different'); return }
  savingAdd.value = true
  try {
    await http.post(`${API_BASE}/alias-domains`, payload)
    showAdd.value = false
    toast.success(`Domain alias ${payload.alias_domain} → ${payload.target_domain} created successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to create domain alias')
  } finally { savingAdd.value = false }
}

// ─── Edit modal ───
const showEdit = ref(false)
const loadingEdit = ref(false)
const savingEdit = ref(false)
const editInitialData = ref<{ alias_domain: string; target_domain: string; active: boolean } | null>(null)
const editAliasDomain = ref('')

async function openEdit(row: AliasDomain) {
  editAliasDomain.value = row.alias_domain
  editInitialData.value = { alias_domain: row.alias_domain, target_domain: row.target_domain, active: row.active }
  loadingEdit.value = true
  showEdit.value = true
  try {
    const res = await http.get(`${API_BASE}/alias-domains/${encodeURIComponent(row.alias_domain)}`)
    const ad = res.data?.data
    editInitialData.value = { alias_domain: ad.alias_domain, target_domain: ad.target_domain, active: ad.active }
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load alias domain')
  } finally { loadingEdit.value = false }
}

async function handleEdit(payload: { target_domain: string; active: boolean }) {
  if (!payload.target_domain) { toast.error('Please select a target domain'); return }
  savingEdit.value = true
  try {
    await http.put(`${API_BASE}/alias-domains/${encodeURIComponent(editAliasDomain.value)}`, payload)
    showEdit.value = false
    toast.success(`Domain alias ${editAliasDomain.value} updated successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to update alias domain')
  } finally { savingEdit.value = false }
}

// ─── Delete ───
const showDeleteConfirm = ref(false)
const deletingRow = ref(false)
const deleteTarget = ref<AliasDomain | null>(null)

function confirmDelete(row: AliasDomain) { deleteTarget.value = row; showDeleteConfirm.value = true }

async function submitDelete() {
  if (!deleteTarget.value) return
  deletingRow.value = true
  try {
    const name = deleteTarget.value.alias_domain
    await http.delete(`${API_BASE}/alias-domains/${encodeURIComponent(name)}`)
    showDeleteConfirm.value = false; deleteTarget.value = null
    toast.success(`Domain alias ${name} deleted successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to delete alias domain')
    showDeleteConfirm.value = false
  } finally { deletingRow.value = false }
}
</script>

<style scoped>
.dom-page { background: #ebf2fe; padding: 24px 28px 40px; }
</style>
