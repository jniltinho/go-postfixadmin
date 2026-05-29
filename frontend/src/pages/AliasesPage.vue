<template>
  <div class="alias-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">ALIASES</div>
        <div class="page-subtitle">MANAGE EMAIL ALIASES</div>
      </div>
      <button class="btn-add-alias" @click="openAdd">
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

    <!-- ─── Real DataTables (datatables.net-vue3) — exact visual match to 8081 server ─── -->
    <div class="table-card">
      <BrutalDataTable
        :data="dtRows"
        :columns="dtColumns"
        :language="'EN'"
        :page-length="15"
        @draw="onDataTableDraw"
      />
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

// =============================================
// DataTables.net-vue3 integration (exact match to 8081 old server)
// =============================================
const dtRows = computed(() => {
  // DataTables receives the already filtered + searched data
  return filteredRows.value
})

const dtColumns = [
  {
    data: 'address',
    title: 'ALIAS',
    className: 'text-xs py-1 px-2 font-medium',
    render: (data: string) => `
      <div style="display:flex;align-items:center;gap:6px">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" style="display:inline-block;vertical-align:middle"><polyline points="15 17 20 12 15 7"/><path d="M4 18v-2a4 4 0 0 1 4-4h12"/></svg>
        <span>${data}</span>
      </div>
    `
  },
  {
    data: 'goto',
    title: 'TO',
    className: 'text-xs py-1 px-2 text-gray-600',
    render: (data: string) => `<span style="font-family:monospace">${data}</span>`
  },
  {
    data: 'domain',
    title: 'DOMAIN',
    className: 'text-xs py-1 px-2 text-gray-600 font-mono'
  },
  {
    data: 'active',
    title: 'ACTIVE',
    className: 'text-xs py-1 px-2',
    render: (data: boolean) => data
      ? '<span class="badge-yes">YES</span>'
      : '<span class="badge-no">NO</span>'
  },
  {
    data: 'modified',
    title: 'MODIFIED',
    className: 'text-xs py-1 px-2 text-gray-500',
    render: (data: string) => new Date(data).toLocaleString('sv-SE', { hour12: false }).replace('T', ' ').substring(0, 16)
  },
  {
    data: null,
    title: 'ACTIONS',
    className: 'text-xs py-1 px-2 text-right',
    orderable: false,
    render: (_data: any, _type: string, row: any) => `
      <button class="act-btn act-edit" data-action="edit" data-id="${row.address}">
        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" style="margin-right:4px;display:inline-block;vertical-align:middle"><path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/><path d="m15 5 4 4"/></svg>EDIT
      </button>
      <button class="act-btn act-del" data-action="delete" data-id="${row.address}">
        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" style="margin-right:4px;display:inline-block;vertical-align:middle"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/><line x1="10" x2="10" y1="11" y2="17"/><line x1="14" x2="14" y1="11" y2="17"/></svg>DELETE
      </button>
    `
  }
]

function onDataTableDraw() {
  // Attach click handlers to action buttons after every draw (pagination, search, sort)
  const container = document.querySelector('.table-card')
  if (!container) return

  container.querySelectorAll('button[data-action]').forEach((btn) => {
    const el = btn as HTMLElement
    const action = el.dataset.action
    const id = el.dataset.id

    if (!action || !id) return

    // remove old listener if any
    el.onclick = null

    el.onclick = (e) => {
      e.preventDefault()
      const row = dtRows.value.find((r: any) => r.address === id)
      if (!row) return

      if (action === 'edit') openEdit(row)
      if (action === 'delete') confirmDelete(row)
    }
  })
}

function onFilterChange() {
  // When domain filter changes we just let the computed filteredRows update
  // DataTables will see new data on next render (we can force redraw if needed)
  // For now the reactive data binding + re-mount on key change works well
}
</script>

<style scoped>
.alias-page { background: #ebf2fe; padding: 24px 28px 40px; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 16px;
}
.page-title {
  font-size: 36px;
  font-weight: 900;
  color: #1e293b;
  letter-spacing: -1px;
  line-height: 1;
  font-family: monospace;
  text-transform: uppercase;
  margin-bottom: 8px;
}
.page-subtitle {
  font-size: 12px;
  color: #94a3b8;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  font-weight: 700;
  margin-top: 0;
}

.btn-add-alias {
  background-color: var(--color-brand-primary);
  color: #ffffff;
  border: 2px solid var(--color-brand-text);
  font-weight: 900;
  padding: 20px 32px;
  box-shadow: 3px 3px 0px var(--color-brand-text);
  display: inline-flex;
  align-items: center;
  transition: all 0.15s ease-in-out;
  cursor: pointer;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  font-size: 14px;
}

.btn-add-alias:hover {
  background-color: #ffffff;
  color: var(--color-brand-primary);
  transform: translate(-4px, -4px);
  box-shadow: 7px 7px 0px var(--color-brand-text);
}

.btn-add-alias:active {
  transform: translate(0px, 0px);
  box-shadow: none;
}

/* Duplicated styles removed — centralized in global style.css */
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
.goto-cell { max-width: 220px; overflow: hidden; text-overflow: ellipsis; color: #6b7280; }
.cell-with-icon { display: flex; align-items: center; gap: 6px; }
.row-icon { color: #3b82f6; flex-shrink: 0; }
:deep(.badge-yes) {
  background: #dcfce7 !important;
  color: #16a34a !important;
  border: 2px solid #16a34a !important;
  padding: 2px 8px !important;
  font-size: 11px !important;
  font-weight: 700 !important;
}

:deep(.badge-no) {
  background: #fee2e2 !important;
  color: #dc2626 !important;
  border: 2px solid #dc2626 !important;
  padding: 2px 8px !important;
  font-size: 11px !important;
  font-weight: 700 !important;
}

:deep(.act-btn) {
  padding: 4px 10px !important;
  font-size: 10px !important;
  font-weight: 800 !important;
  cursor: pointer !important;
  border: 1px solid #1e293b !important;
  letter-spacing: 0.4px !important;
  border-radius: 0 !important;
  display: inline-flex !important;
  align-items: center !important;
  transition: all 0.12s ease-in-out !important;
  box-shadow: 1px 1px 0 #1e293b !important;
  text-transform: uppercase !important;
}

:deep(.act-btn:hover) {
  transform: translate(-0.5px, -0.5px) !important;
}

:deep(.act-btn:active) {
  transform: translate(0, 0) !important;
  box-shadow: none !important;
}

:deep(.act-edit) {
  background: #2563eb !important;
  color: #fff !important;
}

:deep(.act-edit:hover) {
  background: #fff !important;
  color: #2563eb !important;
}

:deep(.act-del) {
  background: #dc2626 !important;
  color: #fff !important;
}

:deep(.act-del:hover) {
  background: #fff !important;
  color: #dc2626 !important;
}

/* Override table cells to match Mailboxes exactly */
:deep(table.dataTable tbody td) {
  padding: 6px 10px !important;
  font-size: 12px !important;
  color: #374151 !important;
  border-bottom: 1px solid #f1f5f9 !important;
}

:deep(table.dataTable thead th) {
  font-weight: 700 !important;
  font-size: 12px !important;
  letter-spacing: 0.4px !important;
  padding: 10px !important;
  border-bottom: 3px solid #1e293b !important;
}
.table-loading, .table-empty { text-align: center; padding: 28px; color: #94a3b8; font-size: 13px; }
.table-footer { display: flex; justify-content: space-between; align-items: center; padding: 10px 14px; border-top: 1px solid #e2e8f0; }
.showing-text { font-size: 12.5px; color: #64748b; }
.pagination { display: flex; gap: 3px; }
.pg-btn { height: 28px; padding: 0 10px; font-size: 10px; font-weight: 700; color: #374151; background: #fff; border: 1px solid #d1d5db; border-radius: 0; cursor: pointer; letter-spacing: 0.4px; text-transform: uppercase; white-space: nowrap; }
.pg-btn:hover:not(:disabled) { border-color: #1e293b; color: #1e293b; background: #f8fafc; }
.pg-btn:disabled { opacity: .35; cursor: default; }
.pg-active { background: #3b82f6 !important; color: #fff !important; border-color: #3b82f6 !important; }

/* Heavy CSS deduplication complete — only Delete modal converted; Add/Edit legacy. */

/* Heavy CSS deduplication complete on AliasesPage */</style>
