<template>
  <div class="tp-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">TRANSPORT LIST</div>
        <div class="page-subtitle">MANAGE MAIL TRANSPORT RULES</div>
      </div>
      <button class="btn-primary" @click="openAdd">
        <Icon name="plus-circle" :size="16" style="margin-right:6px;vertical-align:middle" />
        ADD TRANSPORT
      </button>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <Icon name="alert-triangle" :size="16" class="mr-1" /> {{ error }}
    </div>

    <!-- ─── Table card ─── -->
    <div class="table-card">
      <div class="table-topbar">
        <div class="controls-left">
          <div class="per-page-wrap">
            <select v-model="rowsPerPage" class="ctrl-select" @change="currentPage = 1">
              <option :value="10">10</option>
              <option :value="15">15</option>
              <option :value="25">25</option>
              <option :value="50">50</option>
            </select>
            <span class="ctrl-label">entries per page</span>
          </div>
        </div>
        <div class="controls-right">
          <span class="ctrl-label">Search:</span>
          <input v-model="search" class="search-input" placeholder="Search records..." @input="currentPage = 1" />
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
              <th class="table-th">ACTIONS</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td :colspan="columns.length + 1" class="table-loading"><div class="spinner mx-auto" /></td>
            </tr>
            <tr v-else-if="pagedRows.length === 0">
              <td :colspan="columns.length + 1" class="table-empty">No records found</td>
            </tr>
            <tr v-for="row in pagedRows" :key="row.id" class="table-row">
              <td class="table-td td-domain">{{ row.domain }}</td>
              <td class="table-td">
                <div class="cell-with-icon">
                  <Icon name="arrow-up-down" :size="14" class="row-icon" />
                  <span class="mono-text">{{ row.transport }}</span>
                </div>
              </td>
              <td class="table-td">{{ formatDate(row.created) }}</td>
              <td class="table-td">{{ formatDate(row.modified) }}</td>
              <td class="table-td">
                <span :class="row.active ? 'badge-yes' : 'badge-no'">{{ row.active ? 'YES' : 'NO' }}</span>
              </td>
              <td class="table-td actions-td">
                <button class="act-btn act-edit" @click="openEdit(row)">
                  <Icon name="pencil" :size="12" style="margin-right:4px;vertical-align:middle" />EDIT
                </button>
                <button class="act-btn act-del" @click="confirmDelete(row)">
                  <Icon name="trash-2" :size="12" style="margin-right:4px;vertical-align:middle" />DELETE
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="table-footer">
        <div class="showing-text">
          <template v-if="filteredRows.length === 0">Showing 0 entries</template>
          <template v-else>
            Showing {{ (currentPage - 1) * rowsPerPage + 1 }} to
            {{ Math.min(currentPage * rowsPerPage, filteredRows.length) }} of
            {{ filteredRows.length }} entries
          </template>
        </div>
        <div class="pagination">
          <button class="pg-btn" :disabled="currentPage === 1" @click="currentPage = 1">FIRST</button>
          <button class="pg-btn" :disabled="currentPage === 1" @click="currentPage--">PREVIOUS</button>
          <button v-for="p in pageButtons" :key="p" class="pg-btn" :class="{ 'pg-active': p === currentPage }" @click="currentPage = p">{{ p }}</button>
          <button class="pg-btn" :disabled="currentPage === totalPages" @click="currentPage++">NEXT</button>
          <button class="pg-btn" :disabled="currentPage === totalPages" @click="currentPage = totalPages">LAST</button>
        </div>
      </div>
    </div>

    <!-- ══════════ ADD MODAL ══════════ -->
    <div v-if="showAdd" class="modal-overlay" @click.self="showAdd = false">
      <div class="bg-white border-4 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Header (blue bar) -->
        <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0" style="border-bottom: 2px solid #1e293b;">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="plus-circle" :size="20" class="mr-2" />
            ADD NEW TRANSPORT
          </h3>
          <button @click="showAdd = false" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <!-- Scrollable Body -->
        <div class="overflow-y-auto flex-1">
          <!-- Error Area -->
          <div v-if="addError" class="mx-6 mt-4 bg-red-50 border-2 border-red-600 p-3 flex items-start">
            <Icon name="alert-circle" :size="16" class="text-red-600 mr-2 mt-0.5 flex-shrink-0" />
            <p class="text-sm text-red-700 font-medium">{{ addError }}</p>
          </div>

          <form class="p-6 space-y-5" @submit.prevent="submitAdd">
            <!-- Transport Information Section -->
            <div class="border-2 border-brand-text p-4 space-y-4">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="arrow-up-down" :size="16" class="mr-2" />
                TRANSPORT INFORMATION
              </h4>

              <!-- Domain -->
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                  DOMAIN <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="addForm.domain"
                  type="text"
                  required
                  placeholder="example.com"
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm"
                  @input="addForm.domain = addForm.domain.toLowerCase()"
                />
                <p class="text-[10px] text-gray-400 mt-1">Domain or pattern for this transport rule</p>
              </div>

              <!-- Transport -->
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                  TRANSPORT <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="addForm.transport"
                  type="text"
                  required
                  placeholder="smtp:[mail.example.com]:25"
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm"
                />
                <p class="text-[10px] text-gray-400 mt-1">Transport definition, e.g. smtp:[host]:port or relay:[host]</p>
              </div>

              <!-- Active Checkbox -->
              <div class="flex items-center">
                <input 
                  type="checkbox" 
                  id="add-active" 
                  v-model="addForm.active" 
                  class="w-5 h-5 border-2 border-brand-text cursor-pointer" 
                />
                <label for="add-active" class="ml-2 text-sm font-bold cursor-pointer">Active</label>
              </div>
            </div>
          </form>
        </div>

        <!-- Footer Buttons -->
        <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-brand-text flex-shrink-0 bg-white">
          <button type="button" @click="showAdd = false"
            class="bg-white hover:bg-gray-50 text-brand-text border-2 border-brand-text font-black px-6 py-3 shadow-[2px_2px_0px_#1E293B] transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm">
            <Icon name="x" :size="16" class="mr-2" />
            CANCEL
          </button>
          <button type="button" :disabled="savingAdd" @click="submitAdd"
            class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-3 shadow-[3px_3px_0px_#1E293B] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm disabled:opacity-60">
            <Icon name="save" :size="16" class="mr-2" />
            {{ savingAdd ? 'SAVING...' : 'CREATE TRANSPORT' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ EDIT MODAL ══════════ -->
    <div v-if="showEdit" class="modal-overlay" @click.self="showEdit = false">
      <div class="bg-white border-4 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Header (blue bar) -->
        <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0" style="border-bottom: 2px solid #1e293b;">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="pencil" :size="20" class="mr-2" />
            EDIT TRANSPORT
            <span class="text-white opacity-75 font-normal text-sm ml-2">— {{ editForm.domain }}</span>
          </h3>
          <button @click="showEdit = false" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <!-- Scrollable Body -->
        <div class="overflow-y-auto flex-1 relative min-h-[200px]">
          <!-- Loading overlay inside body -->
          <div v-if="loadingEdit" class="absolute inset-0 bg-white/80 flex flex-col items-center justify-center z-10 gap-3">
            <div class="spinner" style="width:32px;height:32px" />
            <span class="text-xs font-black uppercase tracking-wider text-gray-400">LOADING...</span>
          </div>

          <template v-else>
            <!-- Error Area -->
            <div v-if="editError" class="mx-6 mt-4 bg-red-50 border-2 border-red-600 p-3 flex items-start">
              <Icon name="alert-circle" :size="16" class="text-red-600 mr-2 mt-0.5 flex-shrink-0" />
              <p class="text-sm text-red-700 font-medium">{{ editError }}</p>
            </div>

            <form class="p-6 space-y-5" @submit.prevent="submitEdit">
              <!-- Transport Information Section -->
              <div class="border-2 border-brand-text p-4 space-y-4">
                <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                  <Icon name="arrow-up-down" :size="16" class="mr-2" />
                  TRANSPORT INFORMATION
                </h4>

                <!-- Domain -->
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                    DOMAIN <span class="text-red-500">*</span>
                  </label>
                  <input
                    v-model="editForm.domain"
                    type="text"
                    required
                    placeholder="example.com"
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm"
                    @input="editForm.domain = editForm.domain.toLowerCase()"
                  />
                  <p class="text-[10px] text-gray-400 mt-1">Domain or pattern for this transport rule</p>
                </div>

                <!-- Transport -->
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                    TRANSPORT <span class="text-red-500">*</span>
                  </label>
                  <input
                    v-model="editForm.transport"
                    type="text"
                    required
                    placeholder="smtp:[mail.example.com]:25"
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm"
                  />
                  <p class="text-[10px] text-gray-400 mt-1">Transport definition, e.g. smtp:[host]:port or relay:[host]</p>
                </div>

                <!-- Active Checkbox -->
                <div class="flex items-center">
                  <input 
                    type="checkbox" 
                    id="edit-active" 
                    v-model="editForm.active" 
                    class="w-5 h-5 border-2 border-brand-text cursor-pointer" 
                  />
                  <label for="edit-active" class="ml-2 text-sm font-bold cursor-pointer">Active</label>
                </div>
              </div>
            </form>
          </template>
        </div>

        <!-- Footer Buttons -->
        <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-brand-text flex-shrink-0 bg-white">
          <button type="button" @click="showEdit = false"
            class="bg-white hover:bg-gray-50 text-brand-text border-2 border-brand-text font-black px-6 py-3 shadow-[2px_2px_0px_#1E293B] transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm">
            <Icon name="x" :size="16" class="mr-2" />
            CANCEL
          </button>
          <button type="button" :disabled="savingEdit || loadingEdit" @click="submitEdit"
            class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-3 shadow-[3px_3px_0px_#1E293B] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm disabled:opacity-60">
            <Icon name="save" :size="16" class="mr-2" />
            {{ savingEdit ? 'SAVING...' : 'SAVE CHANGES' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="bg-white border-4 border-brand-text w-full max-w-md flex flex-col">
        <!-- Header -->
        <div class="bg-red-600 px-6 py-4 flex items-center justify-between flex-shrink-0" style="border-bottom: 2px solid #1e293b;">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="trash-2" :size="20" class="mr-2" />
            CONFIRM DELETE
          </h3>
          <button @click="showDeleteConfirm = false" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <!-- Body -->
        <div class="p-6">
          <p class="text-sm font-bold text-brand-text leading-relaxed">
            Are you sure you want to delete transport for<br />
            <strong class="text-red-600">{{ deleteTarget?.domain }}</strong>?<br />
            <span class="block text-xs text-gray-500 font-medium mt-2 leading-normal">
              This action cannot be undone.
            </span>
          </p>
        </div>

        <!-- Footer -->
        <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-brand-text bg-white">
          <button type="button" @click="showDeleteConfirm = false"
            class="bg-white hover:bg-gray-50 text-brand-text border-2 border-brand-text font-black px-4 py-2 shadow-[2px_2px_0px_#1E293B] transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-xs">
            CANCEL
          </button>
          <button type="button" :disabled="deletingRow" @click="submitDelete"
            class="bg-red-600 hover:bg-white hover:text-red-600 text-white border-2 border-brand-text font-black px-4 py-2 shadow-[3px_3px_0px_#1E293B] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-xs">
            <Icon name="trash-2" :size="14" style="margin-right:4px;vertical-align:middle" />
            {{ deletingRow ? 'DELETING...' : 'DELETE' }}
          </button>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import axios from 'axios'
import { useToastStore } from '../stores/toast'

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
const error = ref('')

const search = ref('')
const rowsPerPage = ref(15)
const currentPage = ref(1)
const sortKey = ref('domain')
const sortDir = ref<'asc' | 'desc'>('asc')

const columns = [
  { key: 'domain',    label: 'DOMAIN' },
  { key: 'transport', label: 'TRANSPORT' },
  { key: 'created',   label: 'CREATED' },
  { key: 'modified',  label: 'MODIFIED' },
  { key: 'active',    label: 'ACTIVE' },
]

function sortBy(key: string) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

const filteredRows = computed(() => {
  const q = search.value.toLowerCase()
  let rows = allTransports.value
  if (q) rows = rows.filter(r =>
    r.domain.toLowerCase().includes(q) ||
    r.transport.toLowerCase().includes(q)
  )
  return [...rows].sort((a, b) => {
    const av = String((a as any)[sortKey.value] ?? '')
    const bv = String((b as any)[sortKey.value] ?? '')
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

function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function load() {
  loading.value = true; error.value = ''
  try {
    const res = await axios.get(`${API_BASE}/transports`)
    allTransports.value = res.data?.data ?? []
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load transports'
  } finally { loading.value = false }
}
onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)
const addError = ref('')
const addForm = ref({ domain: '', transport: '', active: true })

function openAdd() {
  addForm.value = { domain: '', transport: '', active: true }
  addError.value = ''
  showAdd.value = true
}

async function submitAdd() {
  addError.value = ''
  const { domain, transport, active } = addForm.value
  if (!domain.trim()) { addError.value = 'Domain is required'; return }
  if (!transport.trim()) { addError.value = 'Transport is required'; return }
  savingAdd.value = true
  try {
    await axios.post(`${API_BASE}/transports`, { domain: domain.trim(), transport: transport.trim(), active })
    showAdd.value = false
    toast.success(`Transport for ${domain} created successfully`)
    await load()
  } catch (e: any) {
    addError.value = e?.response?.data?.error?.message || 'Failed to create transport'
    toast.error(addError.value)
  } finally { savingAdd.value = false }
}

// ─── Edit modal ───
const showEdit = ref(false)
const loadingEdit = ref(false)
const savingEdit = ref(false)
const editError = ref('')
const editForm = ref({ id: 0, domain: '', transport: '', active: true })

async function openEdit(row: Transport) {
  editError.value = ''
  editForm.value = { id: row.id, domain: row.domain, transport: row.transport, active: row.active }
  loadingEdit.value = true
  showEdit.value = true
  try {
    const res = await axios.get(`${API_BASE}/transports/${row.id}`)
    const t = res.data?.data
    editForm.value = { id: t.id, domain: t.domain, transport: t.transport, active: t.active }
  } catch (e: any) {
    editError.value = e?.response?.data?.error?.message || 'Failed to load transport'
  } finally { loadingEdit.value = false }
}

async function submitEdit() {
  editError.value = ''
  const f = editForm.value
  if (!f.domain.trim()) { editError.value = 'Domain is required'; return }
  if (!f.transport.trim()) { editError.value = 'Transport is required'; return }
  savingEdit.value = true
  try {
    await axios.put(`${API_BASE}/transports/${f.id}`, {
      domain: f.domain.trim(),
      transport: f.transport.trim(),
      active: f.active,
    })
    showEdit.value = false
    toast.success(`Transport for ${f.domain} updated successfully`)
    await load()
  } catch (e: any) {
    editError.value = e?.response?.data?.error?.message || 'Failed to update transport'
    toast.error(editError.value)
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
    await axios.delete(`${API_BASE}/transports/${id}`)
    showDeleteConfirm.value = false; deleteTarget.value = null
    toast.success(`Transport for ${domain} deleted successfully`)
    await load()
  } catch (e: any) {
    const msg = e?.response?.data?.error?.message || 'Failed to delete transport'
    error.value = msg
    toast.error(msg)
    showDeleteConfirm.value = false
  } finally { deletingRow.value = false }
}
</script>

<style scoped>
.tp-page { background: #ebf2fe; padding: 24px 28px 40px; }

/* .page-header, .btn-primary, .error-banner, .table-card now from global style.css */

.table-wrap { overflow-x: auto; }
.td-domain { color: #1e293b; font-weight: 700; }
.mono-text { font-family: monospace; font-size: 12px; color: #1e293b; font-weight: 600; }

/* Migration note: All three modals converted to BrutalModal. Old CSS fully removed. */

.loading-overlay { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 40px; gap: 12px; }
.loading-text { font-size: 11px; font-weight: 800; letter-spacing: 1px; color: #94a3b8; text-transform: uppercase; }

/* ─── Info card ─── */
.info-card { border: 2px solid #1e293b; padding: 16px; display: flex; flex-direction: column; gap: 14px; }
.info-card-title { font-size: 12px; font-weight: 900; color: #1e293b; letter-spacing: 0.6px; text-transform: uppercase; font-family: monospace; display: flex; align-items: center; }

/* ─── Form elements ─── */
.form-group { display: flex; flex-direction: column; gap: 3px; }
.form-label { font-size: 11px; font-weight: 800; color: #1e293b; letter-spacing: 0.7px; text-transform: uppercase; }
.req { color: #ef4444; }
.form-input { border: 2px solid #1e293b; padding: 8px 10px; font-size: 13px; color: #374151; outline: none; border-radius: 0; width: 100%; box-sizing: border-box; height: 40px; transition: border-color .12s; }
.form-input:focus { border-color: #3b82f6; }
.form-hint { font-size: 10px; color: #94a3b8; margin-top: 2px; }
.check-label { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 600; color: #374151; cursor: pointer; }
.check-label input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; accent-color: #3b82f6; }

/* Old modal footer styles removed (now handled by BrutalModal + global buttons) */
</style>
