<template>
  <q-page class="alias-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">ALIASES</div>
        <div class="page-subtitle">MANAGE EMAIL ALIASES</div>
      </div>
      <button class="btn-primary" @click="openAdd">
        <q-icon name="add_circle" size="16px" style="margin-right:6px;vertical-align:middle" />
        ADD ALIAS
      </button>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <q-icon name="warning" size="16px" /> {{ error }}
    </div>

    <!-- ─── Domain Filter ─── -->
    <div class="filter-section">
      <label class="filter-label">FILTER BY DOMAIN:</label>
      <select v-model="domainFilter" class="filter-select" @change="currentPage = 1">
        <option value="">All Domains</option>
        <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
      </select>
      <a v-if="domainFilter" class="clear-filter" @click="domainFilter = ''; currentPage = 1">Clear Filter</a>
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
              <td :colspan="columns.length + 1" class="table-loading"><q-spinner color="primary" size="24px" /></td>
            </tr>
            <tr v-else-if="pagedRows.length === 0">
              <td :colspan="columns.length + 1" class="table-empty">No records found</td>
            </tr>
            <tr v-for="row in pagedRows" :key="row.address" class="table-row">
              <td class="table-td td-link">
                <div class="cell-with-icon">
                  <q-icon name="forward" size="14px" class="row-icon" />
                  {{ row.address }}
                </div>
              </td>
              <td class="table-td goto-cell">{{ formatGoto(row.goto) }}</td>
              <td class="table-td">{{ row.domain }}</td>
              <td class="table-td">
                <span :class="row.active ? 'badge-yes' : 'badge-no'">{{ row.active ? 'YES' : 'NO' }}</span>
              </td>
              <td class="table-td">{{ formatDate(row.modified) }}</td>
              <td class="table-td actions-td">
                <button class="act-btn act-edit" @click="openEdit(row)">
                  <q-icon name="edit" size="12px" style="margin-right:4px;vertical-align:middle" />EDIT
                </button>
                <button class="act-btn act-del" @click="confirmDelete(row)">
                  <q-icon name="delete" size="12px" style="margin-right:4px;vertical-align:middle" />DELETE
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
      <div class="modal-card">
        <div class="modal-head">
          <span class="modal-head-title">
            <q-icon name="add_circle" size="18px" style="margin-right:8px;vertical-align:middle" />
            ADD NEW ALIAS
          </span>
          <button class="modal-close" @click="showAdd = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="addError" class="modal-error">
            <q-icon name="warning" size="14px" style="margin-right:6px;flex-shrink:0" />{{ addError }}
          </div>

          <!-- ALIAS INFORMATION card -->
          <div class="info-card">
            <div class="info-card-title">
              <q-icon name="forward" size="15px" style="margin-right:6px" />
              ALIAS INFORMATION
            </div>

            <div class="form-row2">
              <div class="form-group">
                <label class="form-label">ALIAS <span class="req">*</span></label>
                <input v-model="addForm.localPart" class="form-input" placeholder="alias-name" @input="addForm.localPart = addForm.localPart.toLowerCase()" />
                <span class="form-hint">Local part of the alias address</span>
              </div>
              <div class="form-group">
                <label class="form-label">DOMAIN <span class="req">*</span></label>
                <select v-model="addForm.domain" class="form-select-plain">
                  <option value="" disabled>Select a domain...</option>
                  <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
                </select>
              </div>
            </div>

            <!-- Preview -->
            <div class="email-preview">
              <span class="email-preview-label">FULL ADDRESS:</span>
              <span class="email-preview-value">
                <span class="email-preview-local">{{ addForm.localPart || 'alias' }}</span>@<span class="email-preview-domain">{{ addForm.domain || 'domain.com' }}</span>
              </span>
            </div>

            <div class="form-group">
              <label class="form-label">TO (RECIPIENTS) <span class="req">*</span></label>
              <textarea v-model="addForm.goto" class="form-textarea" rows="4" placeholder="recipient@example.com&#10;another@example.com" />
              <span class="form-hint">One email address per line. Multiple recipients are supported.</span>
            </div>

            <label class="check-label">
              <input type="checkbox" v-model="addForm.active" />
              Active Alias
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showAdd = false">
            <q-icon name="close" size="14px" style="margin-right:4px;vertical-align:middle" />CANCEL
          </button>
          <button class="btn-primary" :disabled="savingAdd" @click="submitAdd">
            <q-icon name="add_circle" size="14px" style="margin-right:6px;vertical-align:middle" />
            {{ savingAdd ? 'SAVING...' : 'CREATE ALIAS' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ EDIT MODAL ══════════ -->
    <div v-if="showEdit" class="modal-overlay" @click.self="showEdit = false">
      <div class="modal-card">
        <div class="modal-head">
          <span class="modal-head-title">
            <q-icon name="edit" size="18px" style="margin-right:8px;vertical-align:middle" />
            EDIT ALIAS
            <span class="modal-head-sub">— {{ editForm.address }}</span>
          </span>
          <button class="modal-close" @click="showEdit = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="editError" class="modal-error">
            <q-icon name="warning" size="14px" style="margin-right:6px;flex-shrink:0" />{{ editError }}
          </div>

          <!-- ALIAS INFORMATION card -->
          <div class="info-card">
            <div class="info-card-title">
              <q-icon name="forward" size="15px" style="margin-right:6px" />
              ALIAS INFORMATION
            </div>

            <div class="form-group">
              <label class="form-label">ALIAS ADDRESS</label>
              <input :value="editForm.address" class="form-input form-input-disabled" disabled />
              <span class="form-hint">Alias address cannot be changed after creation</span>
            </div>

            <div class="form-group">
              <label class="form-label">TO (RECIPIENTS) <span class="req">*</span></label>
              <textarea v-model="editForm.goto" class="form-textarea" rows="4" placeholder="recipient@example.com" />
              <span class="form-hint">One email address per line. Multiple recipients are supported.</span>
            </div>

            <label class="check-label">
              <input type="checkbox" v-model="editForm.active" />
              Active Alias
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showEdit = false">
            <q-icon name="close" size="14px" style="margin-right:4px;vertical-align:middle" />CANCEL
          </button>
          <button class="btn-primary" :disabled="savingEdit" @click="submitEdit">
            <q-icon name="save" size="14px" style="margin-right:6px;vertical-align:middle" />
            {{ savingEdit ? 'SAVING...' : 'UPDATE ALIAS' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="modal-card modal-card-sm">
        <div class="modal-head modal-head-danger">
          <span class="modal-head-title">
            <q-icon name="delete" size="16px" style="margin-right:6px;vertical-align:middle" />CONFIRM DELETE
          </span>
          <button class="modal-close" @click="showDeleteConfirm = false">✕</button>
        </div>
        <div class="modal-body">
          <p class="confirm-text">
            Are you sure you want to delete alias<br />
            <strong>{{ deleteTarget?.address }}</strong>?<br />
            <span class="confirm-sub">This action cannot be undone.</span>
          </p>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showDeleteConfirm = false">CANCEL</button>
          <button class="btn-danger" :disabled="deletingRow" @click="submitDelete">
            {{ deletingRow ? 'DELETING...' : 'DELETE' }}
          </button>
        </div>
      </div>
    </div>

  </q-page>
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
watch([search, rowsPerPage, domainFilter], () => { currentPage.value = 1 })

function formatGoto(goto: string): string {
  if (!goto) return '—'
  const parts = goto.split(',').map(s => s.trim()).filter(Boolean)
  return parts.length === 1 ? parts[0] : `${parts[0]} (+${parts.length - 1} more)`
}
function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

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
    showAdd.value = false
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
    showEdit.value = false
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
</script>

<style scoped>
.alias-page { background: #ebf2fe; padding: 24px 28px 40px; }

.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 20px; }
.page-title { font-size: 28px; font-weight: 900; color: #1e293b; letter-spacing: -0.5px; line-height: 1; font-family: monospace; text-transform: uppercase; }
.page-subtitle { font-size: 10px; color: #94a3b8; letter-spacing: 0.8px; margin-top: 6px; text-transform: uppercase; font-weight: 700; }

.btn-primary {
  background: #3b82f6; color: #fff; border: 2px solid #1e293b; padding: 20px 32px;
  font-size: 16px; font-weight: 900; letter-spacing: 1.6px; cursor: pointer;
  border-radius: 0; transition: all .15s; text-transform: uppercase;
  box-shadow: 3px 3px 0 #1e293b; display: flex; align-items: center;
}
.btn-primary:hover:not(:disabled) { background: #fff; color: #3b82f6; transform: translate(-1px,-1px); box-shadow: 4px 4px 0 #1e293b; }
.btn-primary:active:not(:disabled) { transform: translate(0,0); box-shadow: none; }
.btn-primary:disabled { opacity: .5; cursor: default; }

.filter-section { display: flex; align-items: center; gap: 12px; margin-bottom: 20px; }
.filter-label { font-size: 10px; font-weight: 800; color: #1e293b; letter-spacing: 1.2px; text-transform: uppercase; white-space: nowrap; }
.filter-select { height: 38px; padding: 0 12px; border: 2px solid #1e293b; background: #fff; font-size: 13px; font-weight: 500; color: #374151; border-radius: 0; outline: none; cursor: pointer; }
.filter-select:focus { border-color: #3b82f6; }
.clear-filter { font-size: 12px; color: #ef4444; cursor: pointer; text-decoration: underline; font-weight: 600; }

.error-banner { background: #fef2f2; border: 1px solid #fca5a5; color: #dc2626; padding: 10px 14px; font-size: 13px; margin-bottom: 18px; display: flex; align-items: center; gap: 6px; }

.table-card { background: #fff; border: 2px solid #1e293b; box-shadow: 2px 2px 0 #1e293b; }
.table-topbar { display: flex; justify-content: space-between; align-items: center; padding: 10px 14px; border-bottom: 1px solid #e2e8f0; gap: 12px; flex-wrap: wrap; }
.controls-left, .controls-right { display: flex; align-items: center; gap: 8px; }
.per-page-wrap { display: flex; align-items: center; gap: 6px; }
.ctrl-select { border: 1px solid #d1d5db; padding: 4px 6px; font-size: 13px; color: #374151; background: #fff; border-radius: 0; outline: none; }
.ctrl-label { font-size: 12px; color: #64748b; font-weight: 500; }
.search-input { border: 1px solid #d1d5db; padding: 4px 8px; font-size: 13px; color: #374151; width: 200px; outline: none; border-radius: 0; }
.search-input:focus { border-color: #3b82f6; }

.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 12.5px; }
.table-head-row { background: #3b82f6; }
.table-th { color: #fff; font-weight: 600; letter-spacing: 0.4px; padding: 10px 14px; text-align: left; cursor: pointer; white-space: nowrap; user-select: none; }
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
.badge-yes { background: #dcfce7; color: #16a34a; padding: 2px 8px; font-size: 11px; font-weight: 700; }
.badge-no  { background: #fee2e2; color: #dc2626; padding: 2px 8px; font-size: 11px; font-weight: 700; }
.actions-td { display: flex; gap: 6px; align-items: center; }
.act-btn { padding: 4px 10px; font-size: 10px; font-weight: 800; cursor: pointer; border: 1px solid #1e293b; letter-spacing: 0.4px; border-radius: 0; display: inline-flex; align-items: center; transition: all .12s; box-shadow: 1px 1px 0 #1e293b; text-transform: uppercase; }
.act-btn:hover { transform: translate(-0.5px,-0.5px); box-shadow: 2px 2px 0 #1e293b; }
.act-btn:active { transform: translate(0,0); box-shadow: none; }
.act-edit { background: #3b82f6; color: #fff; }
.act-edit:hover { background: #fff; color: #3b82f6; }
.act-del  { background: #ef4444; color: #fff; }
.act-del:hover { background: #fff; color: #ef4444; }
.table-loading, .table-empty { text-align: center; padding: 28px; color: #94a3b8; font-size: 13px; }
.table-footer { display: flex; justify-content: space-between; align-items: center; padding: 10px 14px; border-top: 1px solid #e2e8f0; }
.showing-text { font-size: 12.5px; color: #64748b; }
.pagination { display: flex; gap: 3px; }
.pg-btn { height: 28px; padding: 0 10px; font-size: 10px; font-weight: 700; color: #374151; background: #fff; border: 1px solid #d1d5db; border-radius: 0; cursor: pointer; letter-spacing: 0.4px; text-transform: uppercase; white-space: nowrap; }
.pg-btn:hover:not(:disabled) { border-color: #1e293b; color: #1e293b; background: #f8fafc; }
.pg-btn:disabled { opacity: .35; cursor: default; }
.pg-active { background: #3b82f6 !important; color: #fff !important; border-color: #3b82f6 !important; }

/* ─── Modals ─── */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,.5); z-index: 9000; display: flex; align-items: center; justify-content: center; padding: 20px; }
.modal-card { background: #fff; border: 3px solid #1e293b; width: 100%; max-width: 600px; max-height: 90vh; display: flex; flex-direction: column; border-radius: 0; box-shadow: 4px 4px 0 #1e293b; }
.modal-card-sm { max-width: 420px; }
.modal-head { display: flex; justify-content: space-between; align-items: center; padding: 14px 20px; background: #3b82f6; color: #fff; flex-shrink: 0; }
.modal-head-title { font-size: 15px; font-weight: 900; letter-spacing: 0.3px; font-family: monospace; text-transform: uppercase; display: flex; align-items: center; }
.modal-head-sub { font-size: 13px; color: rgba(255,255,255,.7); margin-left: 8px; font-weight: 400; }
.modal-head-danger { background: #dc2626; }
.modal-close { background: transparent; border: none; color: #fff; cursor: pointer; font-size: 18px; line-height: 1; padding: 2px 6px; }
.modal-close:hover { opacity: .75; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; display: flex; flex-direction: column; gap: 16px; }
.modal-error { background: #fef2f2; border: 2px solid #dc2626; color: #dc2626; padding: 10px 14px; font-size: 13px; display: flex; align-items: flex-start; }

/* ─── Info card ─── */
.info-card { border: 2px solid #1e293b; padding: 16px; display: flex; flex-direction: column; gap: 14px; }
.info-card-title { font-size: 12px; font-weight: 900; color: #1e293b; letter-spacing: 0.6px; text-transform: uppercase; font-family: monospace; display: flex; align-items: center; }

/* ─── Email preview ─── */
.email-preview { background: #f8fafc; border: 1px solid #d1d5db; padding: 0 12px; height: 40px; display: flex; align-items: center; gap: 8px; }
.email-preview-label { font-size: 10px; font-weight: 800; letter-spacing: 1px; color: #94a3b8; text-transform: uppercase; white-space: nowrap; }
.email-preview-value { font-family: monospace; font-size: 13px; font-weight: 700; }
.email-preview-local { color: #3b82f6; }
.email-preview-domain { color: #3b82f6; }

/* ─── Form elements ─── */
.form-row2 { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-group { display: flex; flex-direction: column; gap: 3px; }
.form-label { font-size: 11px; font-weight: 800; color: #1e293b; letter-spacing: 0.7px; text-transform: uppercase; }
.req { color: #ef4444; }
.form-input { border: 2px solid #1e293b; padding: 8px 10px; font-size: 13px; color: #374151; outline: none; border-radius: 0; width: 100%; box-sizing: border-box; transition: border-color .12s; height: 40px; }
.form-input:focus { border-color: #3b82f6; }
.form-input-disabled { background: #f1f5f9; color: #94a3b8; cursor: not-allowed; }
.form-select-plain { border: 2px solid #1e293b; padding: 0 10px; font-size: 13px; color: #374151; background: #fff; border-radius: 0; width: 100%; height: 40px; outline: none; cursor: pointer; }
.form-select-plain:focus { border-color: #3b82f6; }
.form-textarea { border: 2px solid #1e293b; padding: 8px 10px; font-size: 13px; color: #374151; outline: none; border-radius: 0; width: 100%; box-sizing: border-box; resize: vertical; font-family: inherit; transition: border-color .12s; }
.form-textarea:focus { border-color: #3b82f6; }
.form-hint { font-size: 10px; color: #94a3b8; margin-top: 2px; }
.check-label { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 600; color: #374151; cursor: pointer; }
.check-label input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; accent-color: #3b82f6; }

/* ─── Modal footer ─── */
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 14px 20px; border-top: 2px solid #e2e8f0; flex-shrink: 0; background: #f8fafc; }
.btn-cancel { background: #fff; border: 2px solid #1e293b; color: #374151; padding: 8px 20px; font-size: 11px; font-weight: 800; cursor: pointer; border-radius: 0; text-transform: uppercase; letter-spacing: 0.5px; box-shadow: 2px 2px 0 #1e293b; transition: all .12s; display: flex; align-items: center; }
.btn-cancel:hover { background: #f1f5f9; transform: translate(-1px,-1px); box-shadow: 3px 3px 0 #1e293b; }
.btn-cancel:active { transform: translate(0,0); box-shadow: none; }
.btn-danger { background: #ef4444; color: #fff; border: 2px solid #1e293b; padding: 8px 20px; font-size: 11px; font-weight: 800; cursor: pointer; border-radius: 0; text-transform: uppercase; letter-spacing: 0.5px; box-shadow: 2px 2px 0 #1e293b; transition: all .12s; }
.btn-danger:hover:not(:disabled) { background: #dc2626; transform: translate(-1px,-1px); box-shadow: 3px 3px 0 #1e293b; }
.btn-danger:active:not(:disabled) { transform: translate(0,0); box-shadow: none; }
.btn-danger:disabled { opacity: .5; cursor: default; }
.confirm-text { font-size: 14px; color: #374151; line-height: 1.8; text-align: center; margin: 8px 0; }
.confirm-sub { font-size: 12px; color: #dc2626; }
</style>
