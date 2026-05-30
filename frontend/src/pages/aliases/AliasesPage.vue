<template>
  <div class="alias-page">

    <!-- ─── Header ─── -->
    <div class="dom-header">
      <div>
        <div class="dom-title">ALIASES</div>
        <div class="dom-subtitle">MANAGE EMAIL ALIASES</div>
      </div>
      <button class="btn-add-big" @click="showAdd = true">
        <Icon name="plus-circle" :size="20" style="margin-right:12px;vertical-align:middle" />
        ADD ALIAS
      </button>
    </div>


    <!-- ─── Table ─── -->
    <AppTable
      :rows="domainFilteredRows"
      :columns="columns"
      row-key="address"
      :search-fields="['address', 'goto', 'domain']"
      default-sort-key="address"
      :loading="loading"
      @edit="openEdit"
      @delete="confirmDelete"
    >
      <template #toolbar>
        <div class="mb-4 flex items-center gap-3">
          <label class="text-xs font-black uppercase tracking-widest text-brand-text">FILTER BY DOMAIN:</label>
          <select v-model="domainFilter" class="h-9 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium bg-white">
            <option value="">All Domains</option>
            <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
          </select>
          <a v-if="domainFilter" class="ml-2 text-xs font-bold text-red-500 hover:underline cursor-pointer" @click="domainFilter = ''">Clear Filter</a>
        </div>
      </template>

      <template #cell-address="{ row }">
        <div class="cell-with-icon">
          <Icon name="arrow-left-right" :size="14" class="row-icon" />
          {{ row.address }}
        </div>
      </template>
      <template #cell-goto="{ value }"><span class="mono" style="word-break:break-all">{{ value }}</span></template>
      <template #cell-domain="{ value }"><span class="mono">{{ value }}</span></template>
      <template #cell-active="{ value }">
        <span :class="value ? 'badge-yes' : 'badge-no'">{{ value ? 'YES' : 'NO' }}</span>
      </template>
      <template #cell-modified="{ value }">{{ formatDate(value) }}</template>
    </AppTable>

    <!-- ══════════ ADD MODAL ══════════ -->
    <AliasAddModal
      v-model="showAdd"
      :domains="domains"
      :saving="savingAdd"
      :selected-domain="domainFilter || undefined"
      @submit="handleAdd"
    />

    <!-- ══════════ EDIT MODAL ══════════ -->
    <AliasEditModal
      v-model="showEdit"
      :initial-data="editInitialData"
      :saving="savingEdit"
      @submit="handleEdit"
    />

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <ConfirmDialog
      v-model="showDeleteConfirm"
      title="CONFIRM DELETE"
      :item-name="deleteTarget?.address"
      :loading="deletingRow"
      @confirm="submitDelete"
    />

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import http from '../../utils/http'
import { useToastStore } from '../../stores/toast'
import AliasAddModal from './AliasAddModal.vue'
import AliasEditModal from './AliasEditModal.vue'

const toast = useToastStore()

interface Alias {
  address: string
  goto: string
  domain: string
  active: boolean
  modified: string
  created: string
}

interface Domain { domain: string }

const allAliases = ref<Alias[]>([])
const domains = ref<Domain[]>([])
const loading = ref(true)
const domainFilter = ref('')

const domainFilteredRows = computed(() => {
  if (!domainFilter.value) return allAliases.value
  return allAliases.value.filter(r => r.domain === domainFilter.value)
})

const columns = [
  { key: 'address',  label: 'ALIAS' },
  { key: 'goto',     label: 'TO' },
  { key: 'domain',   label: 'DOMAIN' },
  { key: 'active',   label: 'ACTIVE' },
  { key: 'modified', label: 'MODIFIED' },
]

function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function load() {
  loading.value = true
  try {
    const [alRes, domRes] = await Promise.all([
      http.get(`${API_BASE}/aliases`),
      http.get(`${API_BASE}/domains`),
    ])
    allAliases.value = alRes.data?.data ?? []
    domains.value = domRes.data?.data ?? []
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load aliases')
  } finally { loading.value = false }
}
onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)

async function handleAdd(payload: any) {
  if (!payload.local_part || !payload.domain) { toast.error('Alias and domain are required'); return }
  if (!payload.goto) { toast.error('At least one recipient is required'); return }
  savingAdd.value = true
  try {
    await http.post(`${API_BASE}/aliases`, payload)
    showAdd.value = false
    toast.success(`Alias ${payload.local_part}@${payload.domain} created successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to create alias')
  } finally { savingAdd.value = false }
}

// ─── Edit modal ───
const showEdit = ref(false)
const savingEdit = ref(false)
const editInitialData = ref<any>(null)
const editAddress = ref('')

async function openEdit(row: Alias) {
  try {
    const res = await http.get(`${API_BASE}/aliases/${encodeURIComponent(row.address)}`)
    const al = res.data?.data
    editAddress.value = al.address
    editInitialData.value = {
      address: al.address,
      goto: (al.goto || '').replace(/,/g, '\n'),
      active: al.active,
    }
    showEdit.value = true
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load alias')
  }
}

async function handleEdit(payload: any) {
  if (!payload.goto.trim()) { toast.error('At least one recipient is required'); return }
  savingEdit.value = true
  try {
    await http.put(`${API_BASE}/aliases/${encodeURIComponent(editAddress.value)}`, payload)
    showEdit.value = false
    toast.success(`Alias ${editAddress.value} updated successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to update alias')
  } finally { savingEdit.value = false }
}

// ─── Delete ───
const showDeleteConfirm = ref(false)
const deletingRow = ref(false)
const deleteTarget = ref<Alias | null>(null)

function confirmDelete(row: Alias) { deleteTarget.value = row; showDeleteConfirm.value = true }

async function submitDelete() {
  if (!deleteTarget.value) return
  deletingRow.value = true
  try {
    const address = deleteTarget.value.address
    await http.delete(`${API_BASE}/aliases/${encodeURIComponent(address)}`)
    showDeleteConfirm.value = false; deleteTarget.value = null
    toast.success(`Alias ${address} deleted successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to delete alias')
    showDeleteConfirm.value = false
  } finally { deletingRow.value = false }
}
</script>

<style scoped>
.alias-page { background: #ebf2fe; padding: 24px 28px 40px; }
</style>
