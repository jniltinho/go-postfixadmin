<template>
  <div class="alias-page">

    <!-- ─── Header ─── -->
    <div class="dom-header">
      <div>
        <div class="dom-title">ALIASES</div>
        <div class="dom-subtitle">MANAGE EMAIL ALIASES</div>
      </div>
      <button class="btn-add-big" @click="openAdd">
        <Icon name="plus-circle" :size="20" style="margin-right:12px;vertical-align:middle" />
        ADD ALIAS
      </button>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <Icon name="alert-triangle" :size="16" class="mr-1" /> {{ error }}
    </div>

    <!-- Filter by Domain -->
    <div class="mb-8 flex items-center">
      <label class="mr-4 text-xs font-black uppercase tracking-widest text-brand-text">FILTER BY DOMAIN:</label>
      <select v-model="domainFilter" class="h-10 px-4 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors bg-white" @change="onFilterChange">
        <option value="">All Domains</option>
        <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
      </select>
      <a v-if="domainFilter" class="ml-4 text-xs font-bold text-red-500 hover:underline cursor-pointer" @click="domainFilter = ''; onFilterChange()">Clear Filter</a>
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
            <tr v-for="row in pagedRows" :key="row.address" class="table-row">
              <td class="table-td td-link">
                <div class="cell-with-icon">
                  <Icon name="arrow-left-right" :size="14" class="row-icon" />
                  {{ row.address }}
                </div>
              </td>
              <td class="table-td mono" style="word-break: break-all;">{{ row.goto }}</td>
              <td class="table-td mono">{{ row.domain }}</td>
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

    <!-- ══════════ ADD ALIAS MODAL (exact pattern from mailbox + form_add_alias.html) ══════════ -->
    <div v-if="showAdd" class="modal-overlay" @click.self="closeAdd">
      <div class="bg-white border-4 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Header -->
        <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="plus-circle" :size="20" class="mr-2" />
            ADD ALIAS
          </h3>
          <button @click="closeAdd" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <!-- Scrollable Body -->
        <div class="overflow-y-auto flex-1">
          <!-- Error -->
          <div v-if="addError" class="mx-6 mt-4 bg-red-50 border-2 border-red-600 p-3 flex items-start">
            <Icon name="alert-circle" :size="16" class="text-red-600 mr-2 mt-0.5 flex-shrink-0" />
            <p class="text-sm text-red-700 font-medium">{{ addError }}</p>
          </div>

          <form class="p-6 space-y-5" @submit.prevent="submitAdd">
            <!-- Alias Information Section -->
            <div class="border-2 border-brand-text p-4 space-y-4">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="info" :size="16" class="mr-2" />
                ALIAS INFORMATION
              </h4>

              <!-- Alias + Domain -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4 items-start">
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                    ALIAS <span class="text-red-500">*</span>
                  </label>
                  <input
                    v-model="addForm.localPart"
                    type="text"
                    required
                    placeholder="alias-name"
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm lowercase"
                    style="text-transform: lowercase;"
                    @input="onLocalPartInputAliases"
                  />
                  <p class="text-[10px] text-gray-500 mt-1">Local part of the alias address</p>
                </div>
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                    DOMAIN <span class="text-red-500">*</span>
                  </label>
                  <select
                    v-model="addForm.domain"
                    required
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors bg-white cursor-pointer text-sm"
                    @change="updateAliasPreview"
                  >
                    <option value="" disabled>Select a domain...</option>
                    <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
                  </select>
                </div>
              </div>

              <!-- Preview -->
              <div class="bg-gray-50 border border-gray-300 px-3 h-10 flex items-center gap-2">
                <span class="text-xs font-bold uppercase tracking-widest text-gray-400 whitespace-nowrap">FULL ADDRESS:</span>
                <p class="font-mono text-sm font-bold">
                  <span class="text-brand-primary">{{ addForm.localPart || 'alias' }}</span>@<span class="text-brand-primary">{{ addForm.domain || 'domain.com' }}</span>
                </p>
              </div>

              <!-- Goto / Recipients -->
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                  TO (RECIPIENTS) <span class="text-red-500">*</span>
                </label>
                <textarea
                  v-model="addForm.goto"
                  rows="4"
                  required
                  placeholder="recipient@example.com&#10;another@example.com"
                  class="w-full px-3 py-2 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors resize-y text-sm"
                ></textarea>
                <p class="text-[10px] text-gray-500 mt-1">One email address per line. Multiple recipients are supported.</p>
              </div>

              <!-- Active -->
              <div class="flex items-center pt-1">
                <input type="checkbox" id="alias-add-active" v-model="addForm.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                <label for="alias-add-active" class="ml-2 text-sm font-bold cursor-pointer">Active Alias</label>
              </div>
            </div>
          </form>
        </div>

        <!-- Footer -->
        <div class="bg-gray-50 px-6 py-4 flex items-center justify-end space-x-3 border-t-2 border-brand-text flex-shrink-0">
          <button type="button" @click="closeAdd"
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

    <!-- ══════════ EDIT ALIAS MODAL (exact pattern from form_edit_alias.html) ══════════ -->
    <div v-if="showEdit" class="modal-overlay" @click.self="closeEdit">
      <div class="bg-white border-4 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Header -->
        <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="edit" :size="20" class="mr-2" />
            EDIT ALIAS
            <span class="ml-2 text-gray-200 text-base font-mono">- {{ editForm.address }}</span>
          </h3>
          <button @click="closeEdit" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <!-- Body -->
        <div class="overflow-y-auto flex-1">
          <div v-if="editError" class="mx-6 mt-4 bg-red-50 border-2 border-red-600 p-3 flex items-start">
            <Icon name="alert-circle" :size="16" class="text-red-600 mr-2 mt-0.5 flex-shrink-0" />
            <p class="text-sm text-red-700 font-medium">{{ editError }}</p>
          </div>

          <form class="p-6 space-y-5" @submit.prevent="submitEdit">
            <div class="border-2 border-brand-text p-4 space-y-4">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="info" :size="16" class="mr-2" />
                ALIAS INFORMATION
              </h4>

              <!-- Read-only address -->
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">ALIAS ADDRESS</label>
                <input :value="editForm.address" readonly
                  class="w-full h-10 px-3 border-2 border-gray-300 bg-gray-50 font-medium font-mono text-sm cursor-not-allowed" />
                <p class="text-[10px] text-gray-500 mt-1">Alias address cannot be changed after creation</p>
              </div>

              <!-- Goto -->
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                  TO (RECIPIENTS) <span class="text-red-500">*</span>
                </label>
                <textarea
                  v-model="editForm.goto"
                  rows="4"
                  required
                  placeholder="recipient@example.com"
                  class="w-full px-3 py-2 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors resize-y text-sm"
                ></textarea>
                <p class="text-[10px] text-gray-500 mt-1">One email address per line. Multiple recipients are supported.</p>
              </div>

              <!-- Active -->
              <div class="flex items-center pt-1">
                <input type="checkbox" id="alias-edit-active" v-model="editForm.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                <label for="alias-edit-active" class="ml-2 text-sm font-bold cursor-pointer">Active Alias</label>
              </div>
            </div>
          </form>
        </div>

        <!-- Footer -->
        <div class="bg-gray-50 px-6 py-4 flex items-center justify-end space-x-3 border-t-2 border-brand-text flex-shrink-0">
          <button type="button" @click="closeEdit"
            class="bg-white hover:bg-gray-100 text-brand-text border-2 border-brand-text font-black px-6 py-2.5 shadow-[2px_2px_0px_#1E293B] hover:translate-y-px hover:shadow-[1px_1px_0px_#1E293B] active:translate-y-0.5 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center">
            <Icon name="x" :size="16" class="mr-2" />
            CANCEL
          </button>
          <button type="button" :disabled="savingEdit" @click="submitEdit"
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
        Are you sure you want to delete alias<br />
        <strong>{{ deleteTarget?.address }}</strong>?<br />
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
const error = ref('')

const search = ref('')
const rowsPerPage = ref(15)
const currentPage = ref(1)
const sortKey = ref('address')
const sortDir = ref<'asc' | 'desc'>('asc')
const domainFilter = ref('')

// Old custom columns/sort logic removed — now using BrutalDataTable (datatables.net-vue3)

const filteredRows = computed(() => {
  let rows = allAliases.value
  if (domainFilter.value) rows = rows.filter(r => r.domain === domainFilter.value)
  const q = search.value.toLowerCase()
  if (q) rows = rows.filter(r =>
    r.address.toLowerCase().includes(q) ||
    r.goto.toLowerCase().includes(q) ||
    r.domain.toLowerCase().includes(q)
  )
  return [...rows].sort((a, b) => {
    const av = String((a as any)[sortKey.value] ?? '')
    const bv = String((b as any)[sortKey.value] ?? '')
    return sortDir.value === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
  })
})

// Old manual pagination removed — DataTables handles length, search, paging, sorting now
watch([search, rowsPerPage, domainFilter], () => { currentPage.value = 1 })

// Old format helpers removed (DataTables render functions handle display now)

async function load() {
  loading.value = true; error.value = ''
  try {
    const [alRes, domRes] = await Promise.all([
      axios.get(`${API_BASE}/aliases`),
      axios.get(`${API_BASE}/domains`),
    ])
    allAliases.value = alRes.data?.data ?? []
    domains.value = domRes.data?.data ?? []
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load aliases'
  } finally { loading.value = false }
}
onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)
const addError = ref('')
const addForm = ref({ localPart: '', domain: '', goto: '', active: true })

function openAdd() {
  addForm.value = { localPart: '', domain: domains.value[0]?.domain || '', goto: '', active: true }
  addError.value = ''
  showAdd.value = true
}

function closeAdd() {
  showAdd.value = false
  addForm.value = { localPart: '', domain: '', goto: '', active: true }
  addError.value = ''
}

function onLocalPartInputAliases(e: Event) {
  const val = (e.target as HTMLInputElement).value.toLowerCase()
  addForm.value.localPart = val
}

function updateAliasPreview() {
  // reactive binding handles preview
}

async function submitAdd() {
  addError.value = ''
  const { localPart, domain, goto: gotoVal, active } = addForm.value
  if (!localPart || !domain) { addError.value = 'Alias and domain are required'; return }
  if (!gotoVal.trim()) { addError.value = 'At least one recipient is required'; return }
  savingAdd.value = true
  try {
    await axios.post(`${API_BASE}/aliases`, {
      local_part: localPart.toLowerCase().trim(),
      domain,
      goto: gotoVal.trim(),
      active,
    })
    closeAdd()
    toast.success(`Alias ${localPart}@${domain} created successfully`)
    await load()
  } catch (e: any) {
    addError.value = e?.response?.data?.error?.message || 'Failed to create alias'
    toast.error(addError.value)
  } finally { savingAdd.value = false }
}

// ─── Edit modal ───
const showEdit = ref(false)
const savingEdit = ref(false)
const editError = ref('')
const editForm = ref({ address: '', goto: '', active: true })

async function openEdit(row: Alias) {
  editError.value = ''
  try {
    const res = await axios.get(`${API_BASE}/aliases/${encodeURIComponent(row.address)}`)
    const al = res.data?.data
    editForm.value = {
      address: al.address,
      goto: (al.goto || '').replace(/,/g, '\n'),
      active: al.active,
    }
    showEdit.value = true
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load alias'
  }
}

function closeEdit() {
  showEdit.value = false
  editError.value = ''
}

async function submitEdit() {
  editError.value = ''
  const f = editForm.value
  if (!f.goto.trim()) { editError.value = 'At least one recipient is required'; return }
  savingEdit.value = true
  try {
    await axios.put(`${API_BASE}/aliases/${encodeURIComponent(f.address)}`, {
      goto: f.goto.trim(),
      active: f.active,
    })
    closeEdit()
    toast.success(`Alias ${f.address} updated successfully`)
    await load()
  } catch (e: any) {
    editError.value = e?.response?.data?.error?.message || 'Failed to update alias'
    toast.error(editError.value)
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
    await axios.delete(`${API_BASE}/aliases/${encodeURIComponent(address)}`)
    showDeleteConfirm.value = false; deleteTarget.value = null
    toast.success(`Alias ${address} deleted successfully`)
    await load()
  } catch (e: any) {
    const msg = e?.response?.data?.error?.message || 'Failed to delete alias'
    error.value = msg
    toast.error(msg)
    showDeleteConfirm.value = false
  } finally { deletingRow.value = false }
}

const columns = [
  { key: 'address',  label: 'ALIAS' },
  { key: 'goto',     label: 'TO' },
  { key: 'domain',   label: 'DOMAIN' },
  { key: 'active',   label: 'ACTIVE' },
  { key: 'modified', label: 'MODIFIED' },
]

function sortBy(key: string) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

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

function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function onFilterChange() {
  // filter auto-applied in computed filteredRows
}
</script>

<style scoped>
.alias-page { background: #ebf2fe; padding: 24px 28px 40px; }

/* Heavy CSS deduplication complete on AliasesPage */</style>
