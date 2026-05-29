<template>
  <div class="dom-page">

    <!-- ─── Header ─── -->
    <div class="dom-header">
      <div>
        <div class="dom-title">DOMAIN ALIASES</div>
        <div class="dom-subtitle">MANAGE ALIAS DOMAIN MAPPINGS</div>
      </div>
      <button class="btn-primary" @click="openAdd">
        <Icon name="plus-circle" :size="16" style="margin-right:6px;vertical-align:middle" />
        ADD DOMAIN ALIAS
      </button>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <Icon name="alert-triangle" :size="16" class="mr-1" /> {{ error }}
    </div>

    <!-- ─── Table card ─── -->
    <div class="table-card">

      <!-- Controls row -->
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

      <!-- Table -->
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
              <td :colspan="columns.length + 1" class="table-loading">
                <div class="spinner mx-auto" />
              </td>
            </tr>
            <tr v-else-if="pagedRows.length === 0">
              <td :colspan="columns.length + 1" class="table-empty">No records found</td>
            </tr>
            <tr v-for="row in pagedRows" :key="row.alias_domain" class="table-row">
              <td class="table-td td-link">
                <div class="cell-with-icon">
                  <Icon name="arrow-left-right" :size="14" class="row-icon" />
                  {{ row.alias_domain }}
                </div>
              </td>
              <td class="table-td mono">{{ row.target_domain }}</td>
              <td class="table-td">
                <span :class="row.active ? 'badge-yes' : 'badge-no'">{{ row.active ? 'YES' : 'NO' }}</span>
              </td>
              <td class="table-td">{{ formatDate(row.modified) }}</td>
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

      <!-- Footer -->
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
          <button
            v-for="p in pageButtons" :key="p"
            class="pg-btn" :class="{ 'pg-active': p === currentPage }"
            @click="currentPage = p"
          >{{ p }}</button>
          <button class="pg-btn" :disabled="currentPage === totalPages" @click="currentPage++">NEXT</button>
          <button class="pg-btn" :disabled="currentPage === totalPages" @click="currentPage = totalPages">LAST</button>
        </div>
      </div>
    </div>


    <!-- ══════════ ADD DOMAIN ALIAS MODAL (exact pattern from form_add_alias_domain.html) ══════════ -->
    <div v-if="showAdd" class="modal-overlay" @click.self="closeAddAliasDomain">
      <div class="bg-white border-4 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Header -->
        <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="plus-circle" :size="20" class="mr-2" />
            ADD DOMAIN ALIAS
          </h3>
          <button @click="closeAddAliasDomain" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <!-- Body -->
        <div class="overflow-y-auto flex-1">
          <div v-if="addError" class="mx-6 mt-4 bg-red-50 border-2 border-red-600 p-3 flex items-start">
            <Icon name="alert-circle" :size="16" class="text-red-600 mr-2 mt-0.5 flex-shrink-0" />
            <p class="text-sm text-red-700 font-medium">{{ addError }}</p>
          </div>

          <form class="p-6 space-y-5" @submit.prevent="submitAdd">
            <div class="border-2 border-brand-text p-4 space-y-4">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="info" :size="16" class="mr-2" />
                ALIAS DOMAIN INFORMATION
              </h4>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                    ALIAS DOMAIN <span class="text-red-500">*</span>
                  </label>
                  <select v-model="addForm.alias_domain" required
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors bg-white cursor-pointer text-sm">
                    <option value="" disabled>Select source domain...</option>
                    <option v-for="d in availableAliasDomains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
                  </select>
                  <p class="text-[10px] text-gray-500 mt-1">Domain that will be the alias (source)</p>
                </div>
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                    TARGET DOMAIN <span class="text-red-500">*</span>
                  </label>
                  <select v-model="addForm.target_domain" required
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors bg-white cursor-pointer text-sm">
                    <option value="" disabled>Select target domain...</option>
                    <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
                  </select>
                  <p class="text-[10px] text-gray-500 mt-1">Domain that will receive the mail (target)</p>
                </div>
              </div>

              <div class="flex items-center pt-2">
                <input type="checkbox" id="ad-add-active" v-model="addForm.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                <label for="ad-add-active" class="ml-2 text-sm font-bold cursor-pointer">Active Domain Alias</label>
              </div>
            </div>
          </form>
        </div>

        <!-- Footer -->
        <div class="bg-gray-50 px-6 py-4 flex items-center justify-end space-x-3 border-t-2 border-brand-text flex-shrink-0">
          <button type="button" @click="closeAddAliasDomain"
            class="bg-white hover:bg-gray-100 text-brand-text border-2 border-brand-text font-black px-6 py-2.5 shadow-[2px_2px_0px_#1E293B] hover:translate-y-px hover:shadow-[1px_1px_0px_#1E293B] active:translate-y-0.5 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center">
            <Icon name="x" :size="16" class="mr-2" />
            CANCEL
          </button>
          <button type="button" :disabled="savingAdd" @click="submitAdd"
            class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-2.5 shadow-[3px_3px_0px_#1E293B] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center disabled:opacity-60">
            <Icon name="plus-circle" :size="16" class="mr-2" />
            {{ savingAdd ? 'SAVING...' : 'CREATE ALIAS' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ EDIT DOMAIN ALIAS MODAL (exact pattern from form_edit_alias_domain.html) ══════════ -->
    <div v-if="showEdit" class="modal-overlay" @click.self="closeEditAliasDomain">
      <div class="bg-white border-4 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Header -->
        <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="edit" :size="20" class="mr-2" />
            EDIT DOMAIN ALIAS
            <span class="ml-2 text-gray-200 text-base font-mono">- {{ editForm.alias_domain }}</span>
          </h3>
          <button @click="closeEditAliasDomain" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <!-- Body -->
        <div class="overflow-y-auto flex-1">
          <div v-if="loadingEdit" class="flex items-center justify-center p-12">
            <div class="flex flex-col items-center">
              <Icon name="loader-2" :size="28" class="text-brand-primary animate-spin mb-2" />
              <span class="text-sm font-bold uppercase tracking-widest text-brand-text">LOADING...</span>
            </div>
          </div>

          <div v-else>
            <div v-if="editError" class="mx-6 mt-4 bg-red-50 border-2 border-red-600 p-3 flex items-start">
              <Icon name="alert-circle" :size="16" class="text-red-600 mr-2 mt-0.5 flex-shrink-0" />
              <p class="text-sm text-red-700 font-medium">{{ editError }}</p>
            </div>

            <form class="p-6 space-y-5" @submit.prevent="submitEdit">
              <div class="border-2 border-brand-text p-4 space-y-4">
                <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                  <Icon name="info" :size="16" class="mr-2" />
                  ALIAS DOMAIN INFORMATION
                </h4>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">ALIAS DOMAIN</label>
                    <input :value="editForm.alias_domain" readonly
                      class="w-full h-10 px-3 border-2 border-gray-300 bg-gray-50 font-medium font-mono text-sm cursor-not-allowed" />
                    <p class="text-[10px] text-gray-500 mt-1">Alias domain cannot be changed after creation</p>
                  </div>
                  <div>
                    <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                      TARGET DOMAIN <span class="text-red-500">*</span>
                    </label>
                    <select v-model="editForm.target_domain" required
                      class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors bg-white cursor-pointer text-sm">
                      <option value="" disabled>Select target domain...</option>
                      <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
                    </select>
                    <p class="text-[10px] text-gray-500 mt-1">Domain that receives the mail</p>
                  </div>
                </div>

                <div class="flex items-center pt-2">
                  <input type="checkbox" id="ad-edit-active" v-model="editForm.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                  <label for="ad-edit-active" class="ml-2 text-sm font-bold cursor-pointer">Active Domain Alias</label>
                </div>
              </div>
            </form>
          </div>
        </div>

        <!-- Footer -->
        <div class="bg-gray-50 px-6 py-4 flex items-center justify-end space-x-3 border-t-2 border-brand-text flex-shrink-0">
          <button type="button" @click="closeEditAliasDomain"
            class="bg-white hover:bg-gray-100 text-brand-text border-2 border-brand-text font-black px-6 py-2.5 shadow-[2px_2px_0px_#1E293B] hover:translate-y-px hover:shadow-[1px_1px_0px_#1E293B] active:translate-y-0.5 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center">
            <Icon name="x" :size="16" class="mr-2" />
            CANCEL
          </button>
          <button type="button" :disabled="savingEdit || loadingEdit" @click="submitEdit"
            class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-2.5 shadow-[3px_3px_0px_#1E293B] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center disabled:opacity-60">
            <Icon name="save" :size="16" class="mr-2" />
            {{ savingEdit ? 'SAVING...' : 'UPDATE ALIAS' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <BrutalModal v-model="showDeleteConfirm" title="CONFIRM DELETE" size="sm" danger>
      <p class="confirm-text">
        Are you sure you want to delete domain alias<br />
        <strong>{{ deleteTarget?.alias_domain }}</strong>?<br />
        <span class="confirm-sub">This action cannot be undone.</span>
      </p>

      <template #footer>
        <button class="btn-cancel" @click="showDeleteConfirm = false">CANCEL</button>
        <button class="btn-danger" :disabled="deletingRow" @click="submitDelete">
          <Icon name="trash-2" :size="14" style="margin-right:6px;vertical-align:middle" />
          {{ deletingRow ? 'DELETING...' : 'DELETE' }}
        </button>
      </template>
    </BrutalModal>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import axios from 'axios'
import { useToastStore } from '../stores/toast'

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
const error = ref('')

const search = ref('')
const rowsPerPage = ref(15)
const currentPage = ref(1)
const sortKey = ref('alias_domain')
const sortDir = ref<'asc' | 'desc'>('asc')

const columns = [
  { key: 'alias_domain',  label: 'ALIAS DOMAIN' },
  { key: 'target_domain', label: 'TARGET DOMAIN' },
  { key: 'active',        label: 'ACTIVE' },
  { key: 'modified',      label: 'MODIFIED' },
]

function sortBy(key: string) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

const filteredRows = computed(() => {
  const q = search.value.toLowerCase()
  let rows = allAliasDomains.value
  if (q) rows = rows.filter(r =>
    r.alias_domain.toLowerCase().includes(q) ||
    r.target_domain.toLowerCase().includes(q)
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
    const [adRes, domRes] = await Promise.all([
      axios.get(`${API_BASE}/alias-domains`),
      axios.get(`${API_BASE}/domains`),
    ])
    allAliasDomains.value = adRes.data?.data ?? []
    domains.value = domRes.data?.data ?? []
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load alias domains'
  } finally { loading.value = false }
}
onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)
const addError = ref('')
const addForm = ref({ alias_domain: '', target_domain: '', active: true })

function openAdd() {
  addForm.value = { alias_domain: '', target_domain: '', active: true }
  addError.value = ''
  showAdd.value = true
}

function closeAddAliasDomain() {
  showAdd.value = false
  addError.value = ''
}

function closeEditAliasDomain() {
  showEdit.value = false
  editError.value = ''
}

async function submitAdd() {
  addError.value = ''
  const { alias_domain, target_domain, active } = addForm.value
  if (!alias_domain) { addError.value = 'Please select a source alias domain'; return }
  if (!target_domain) { addError.value = 'Please select a target domain'; return }
  if (alias_domain === target_domain) { addError.value = 'Alias domain and target domain must be different'; return }
  savingAdd.value = true
  try {
    await axios.post(`${API_BASE}/alias-domains`, { alias_domain, target_domain, active })
    closeAddAliasDomain()
    toast.success(`Domain alias ${alias_domain} → ${target_domain} created successfully`)
    await load()
  } catch (e: any) {
    addError.value = e?.response?.data?.error?.message || 'Failed to create domain alias'
    toast.error(addError.value)
  } finally { savingAdd.value = false }
}

// ─── Edit modal ───
const showEdit = ref(false)
const loadingEdit = ref(false)
const savingEdit = ref(false)
const editError = ref('')
const editForm = ref({ alias_domain: '', target_domain: '', active: true })

async function openEdit(row: AliasDomain) {
  editError.value = ''
  loadingEdit.value = true
  showEdit.value = true
  try {
    const res = await axios.get(`${API_BASE}/alias-domains/${encodeURIComponent(row.alias_domain)}`)
    const ad = res.data?.data
    editForm.value = {
      alias_domain: ad.alias_domain,
      target_domain: ad.target_domain,
      active: ad.active,
    }
  } catch (e: any) {
    editError.value = e?.response?.data?.error?.message || 'Failed to load alias domain'
  } finally { loadingEdit.value = false }
}

async function submitEdit() {
  editError.value = ''
  const f = editForm.value
  if (!f.target_domain) { editError.value = 'Please select a target domain'; return }
  savingEdit.value = true
  try {
    await axios.put(`${API_BASE}/alias-domains/${encodeURIComponent(f.alias_domain)}`, {
      target_domain: f.target_domain,
      active: f.active,
    })
    closeEditAliasDomain()
    toast.success(`Domain alias ${f.alias_domain} updated successfully`)
    await load()
  } catch (e: any) {
    editError.value = e?.response?.data?.error?.message || 'Failed to update alias domain'
    toast.error(editError.value)
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
    await axios.delete(`${API_BASE}/alias-domains/${encodeURIComponent(name)}`)
    showDeleteConfirm.value = false; deleteTarget.value = null
    toast.success(`Domain alias ${name} deleted successfully`)
    await load()
  } catch (e: any) {
    const msg = e?.response?.data?.error?.message || 'Failed to delete alias domain'
    error.value = msg
    toast.error(msg)
    showDeleteConfirm.value = false
  } finally { deletingRow.value = false }
}

// =============================================
// DataTables.net-vue3 - matching the aligned pages
// =============================================
// domains that are not already used as an alias_domain (used in Add modal)
const availableAliasDomains = computed(() => {
  const used = new Set(allAliasDomains.value.map(a => a.alias_domain))
  return domains.value.filter(d => !used.has(d.domain))
})

</script>

<style scoped>
.dom-page { background: #ebf2fe; padding: 24px 28px 40px; }

/* .page-header, .btn-primary, .error-banner, .table-card, .table-topbar now centralized in global style.css */
.controls-left, .controls-right { display: flex; align-items: center; gap: 8px; }
.per-page-wrap { display: flex; align-items: center; gap: 6px; }
.ctrl-select { border: 1px solid #d1d5db; padding: 4px 6px; font-size: 13px; color: #374151; background: #fff; border-radius: 0; outline: none; }
.ctrl-label { font-size: 12px; color: #64748b; font-weight: 500; }
.search-input { border: 1px solid #d1d5db; padding: 4px 8px; font-size: 13px; color: #374151; width: 200px; outline: none; border-radius: 0; }
.search-input:focus { border-color: #3b82f6; }

.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 12.5px; }
.table-head-row { background: #3b82f6; }
.table-th { color: #fff; font-weight: 700; letter-spacing: 0.4px; padding: 10px; text-align: left; cursor: pointer; white-space: nowrap; user-select: none; font-size: 12px; }
.table-th:hover { background: #2563eb; }
.sort-arrows { margin-left: 4px; font-size: 9px; opacity: .5; }
.sort-active { opacity: 1 !important; }
.table-row:nth-child(even) { background: #f8fafc; }
.table-row:hover { background: #eff6ff; }
.table-td { padding: 6px 10px; color: #374151; border-bottom: 1px solid #f1f5f9; font-size: 12px; }
.td-link { color: #1e293b; font-weight: 600; }
.mono-text { font-family: monospace; font-size: 12px; color: #6b7280; }
.cell-with-icon { display: flex; align-items: center; gap: 6px; }
.row-icon { color: #3b82f6; flex-shrink: 0; }
.badge-yes { background: #dcfce7; color: #16a34a; padding: 2px 8px; font-size: 11px; font-weight: 700; }
.badge-no  { background: #fee2e2; color: #dc2626; padding: 2px 8px; font-size: 11px; font-weight: 700; }
.actions-td { display: flex; gap: 6px; align-items: center; justify-content: flex-end; }
.act-btn { padding: 4px 10px; font-size: 10px; font-weight: 800; cursor: pointer; border: 1px solid #1e293b; letter-spacing: 0.4px; border-radius: 0; display: inline-flex; align-items: center; transition: all .12s; box-shadow: 1px 1px 0 #1e293b; text-transform: uppercase; }
.act-btn:hover { transform: translate(-0.5px,-0.5px); }
.act-btn:active { transform: translate(0,0); box-shadow: none; }
.act-edit { background: #2563eb; color: #fff; }
.act-edit:hover { background: #fff; color: #2563eb; }
.act-del  { background: #dc2626; color: #fff; }
.act-del:hover { background: #fff; color: #dc2626; }
.table-loading, .table-empty { text-align: center; padding: 28px; color: #94a3b8; font-size: 13px; }
.table-footer { display: flex; justify-content: space-between; align-items: center; padding: 10px 14px; border-top: 1px solid #e2e8f0; }
.showing-text { font-size: 12.5px; color: #64748b; }
.pagination { display: flex; gap: 3px; }
.pg-btn { height: 28px; padding: 0 10px; font-size: 10px; font-weight: 700; color: #374151; background: #fff; border: 1px solid #d1d5db; border-radius: 0; cursor: pointer; letter-spacing: 0.4px; text-transform: uppercase; white-space: nowrap; }
.pg-btn:hover:not(:disabled) { border-color: #1e293b; color: #1e293b; background: #f8fafc; }
.pg-btn:disabled { opacity: .35; cursor: default; }
.pg-active { background: #3b82f6 !important; color: #fff !important; border-color: #3b82f6 !important; }

/* Old custom modal CSS fully removed — all modals use BrutalModal */

/* ─── Info card ─── */
.info-card { border: 2px solid #1e293b; padding: 16px; display: flex; flex-direction: column; gap: 14px; }
.info-card-title { font-size: 12px; font-weight: 900; color: #1e293b; letter-spacing: 0.6px; text-transform: uppercase; font-family: monospace; display: flex; align-items: center; }

/* ─── Form elements ─── */
.form-row2 { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-group { display: flex; flex-direction: column; gap: 3px; }
.form-label { font-size: 11px; font-weight: 800; color: #1e293b; letter-spacing: 0.7px; text-transform: uppercase; }
.req { color: #ef4444; }
.form-input { border: 2px solid #1e293b; padding: 8px 10px; font-size: 13px; color: #374151; outline: none; border-radius: 0; width: 100%; box-sizing: border-box; height: 40px; }
.form-input-disabled { background: #f1f5f9; color: #94a3b8; cursor: not-allowed; }
.form-select-plain { border: 2px solid #1e293b; padding: 0 10px; font-size: 13px; color: #374151; background: #fff; border-radius: 0; width: 100%; height: 40px; outline: none; cursor: pointer; }
.form-select-plain:focus { border-color: #3b82f6; }
.form-hint { font-size: 10px; color: #94a3b8; margin-top: 2px; }
.check-label { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 600; color: #374151; cursor: pointer; }
.check-label input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; accent-color: #3b82f6; }

/* Old modal footer styles fully removed — all modals now use BrutalModal + global button styles */
</style>
