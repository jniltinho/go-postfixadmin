<template>
  <q-page class="dom-page">

    <!-- ─── Header ─── -->
    <div class="dom-header">
      <div>
        <div class="dom-title">MY DOMAINS</div>
        <div class="dom-subtitle">MANAGE YOUR MAIL DOMAINS</div>
      </div>
      <button class="btn-primary" @click="openAdd">
        <q-icon name="add_circle" size="16px" style="margin-right:6px;vertical-align:middle" />
        ADD DOMAIN
      </button>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <q-icon name="warning" size="16px" /> {{ error }}
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
                <q-spinner color="primary" size="24px" />
              </td>
            </tr>
            <tr v-else-if="pagedRows.length === 0">
              <td :colspan="columns.length + 1" class="table-empty">No records found</td>
            </tr>
            <tr v-for="row in pagedRows" :key="row.domain" class="table-row">
              <td class="table-td td-link">
                <div class="cell-with-icon">
                  <q-icon name="public" size="14px" class="row-icon" />
                  {{ row.domain }}
                </div>
              </td>
              <td class="table-td">{{ row.description || '—' }}</td>
              <td class="table-td">
                <span class="count-badge">{{ row.alias_count ?? 0 }}</span>
                <span class="count-sep">/</span>
                <span class="count-max">{{ row.aliases }}</span>
              </td>
              <td class="table-td">
                <span class="count-badge">{{ row.mailbox_count ?? 0 }}</span>
                <span class="count-sep">/</span>
                <span class="count-max">{{ row.mailboxes }}</span>
              </td>
              <td class="table-td mono">{{ row.transport || 'virtual' }}</td>
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

    <!-- ══════════ ADD MODAL ══════════ -->
    <div v-if="showAdd" class="modal-overlay" @click.self="showAdd = false">
      <div class="modal-card">

        <!-- Modal header -->
        <div class="modal-head">
          <span class="modal-head-title">
            <q-icon name="add_circle" size="18px" style="margin-right:8px;vertical-align:middle" />
            ADD NEW DOMAIN
          </span>
          <button class="modal-close" @click="showAdd = false">✕</button>
        </div>

        <!-- Modal body -->
        <div class="modal-body">

          <!-- Error -->
          <div v-if="addError" class="modal-error">
            <q-icon name="warning" size="14px" style="margin-right:6px;flex-shrink:0" />
            {{ addError }}
          </div>

          <!-- ── BASIC INFORMATION ── -->
          <div class="info-card">
            <div class="info-card-title">
              <q-icon name="info" size="15px" style="margin-right:6px" />
              BASIC INFORMATION
            </div>

            <div class="form-group">
              <label class="form-label">DOMAIN NAME <span class="req">*</span></label>
              <input v-model="addForm.domain" class="form-input" placeholder="example.com" />
              <span class="form-hint">Enter a valid domain name (e.g., example.com)</span>
            </div>

            <div class="form-group">
              <label class="form-label">DESCRIPTION</label>
              <input v-model="addForm.description" class="form-input" placeholder="Optional description for this domain" />
            </div>

            <label class="check-label">
              <input type="checkbox" v-model="addForm.active" />
              Active Domain
            </label>
          </div>

          <!-- ── ADVANCED SETTINGS (collapsible) ── -->
          <details class="adv-details" open>
            <summary class="adv-summary">
              <span class="adv-summary-left">
                <q-icon name="settings" size="15px" style="margin-right:6px" />
                ADVANCED SETTINGS
              </span>
              <q-icon name="expand_less" size="18px" class="adv-chevron" />
            </summary>

            <div class="adv-body">
              <div class="form-row2">
                <div class="form-group">
                  <label class="form-label">ALIAS LIMIT</label>
                  <input v-model.number="addForm.aliases" class="form-input" type="number" min="0" />
                  <span class="form-hint">Maximum number of aliases (0 = unlimited)</span>
                </div>
                <div class="form-group">
                  <label class="form-label">MAILBOX LIMIT</label>
                  <input v-model.number="addForm.mailboxes" class="form-input" type="number" min="0" />
                  <span class="form-hint">Maximum number of mailboxes (0 = unlimited)</span>
                </div>
                <div class="form-group">
                  <label class="form-label">QUOTA LIMIT (MB)</label>
                  <input v-model.number="addForm.quotaMB" class="form-input" type="number" min="0" />
                  <span class="form-hint">Maximum quota limit in MB (0 = unlimited)</span>
                </div>
                <div class="form-group">
                  <label class="form-label">PASSWORD EXPIRY (DAYS)</label>
                  <input v-model.number="addForm.passwordExpiry" class="form-input" type="number" min="0" placeholder="365" />
                  <span class="form-hint">Password expiry in days (empty = never)</span>
                </div>
              </div>

              <div class="form-group">
                <label class="form-label">TRANSPORT</label>
                <select v-model="addForm.transport" class="form-select-plain">
                  <option value="virtual">virtual</option>
                  <option v-for="t in transports" :key="t.id" :value="t.transport">
                    {{ t.domain }} -&gt; {{ t.transport }}
                  </option>
                </select>
                <span class="form-hint">Select the mail transport for this domain</span>
              </div>

              <label class="check-label backupmx-row">
                <input type="checkbox" v-model="addForm.backupmx" />
                Enable Backup MX
                <span class="backupmx-hint">Use this server as a backup mail exchanger</span>
              </label>
            </div>
          </details>

        </div>

        <!-- Modal footer -->
        <div class="modal-footer">
          <button class="btn-cancel" @click="showAdd = false">
            <q-icon name="close" size="14px" style="margin-right:4px;vertical-align:middle" />
            CANCEL
          </button>
          <button class="btn-primary" :disabled="savingAdd" @click="submitAdd">
            <q-icon name="add_circle" size="14px" style="margin-right:6px;vertical-align:middle" />
            {{ savingAdd ? 'SAVING...' : 'SAVE DOMAIN' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ EDIT MODAL ══════════ -->
    <div v-if="showEdit" class="modal-overlay" @click.self="showEdit = false">
      <div class="modal-card">

        <!-- Modal header -->
        <div class="modal-head">
          <span class="modal-head-title">
            <q-icon name="edit" size="18px" style="margin-right:8px;vertical-align:middle" />
            EDIT DOMAIN
            <span class="modal-head-sub">— {{ editForm.domain }}</span>
          </span>
          <button class="modal-close" @click="showEdit = false">✕</button>
        </div>

        <!-- Modal body -->
        <div class="modal-body">

          <!-- Error -->
          <div v-if="editError" class="modal-error">
            <q-icon name="warning" size="14px" style="margin-right:6px;flex-shrink:0" />
            {{ editError }}
          </div>

          <!-- ── BASIC INFORMATION ── -->
          <div class="info-card">
            <div class="info-card-title">
              <q-icon name="info" size="15px" style="margin-right:6px" />
              BASIC INFORMATION
            </div>

            <div class="form-group">
              <label class="form-label">DOMAIN NAME</label>
              <input :value="editForm.domain" class="form-input form-input-disabled" disabled />
              <span class="form-hint">Domain name cannot be changed after creation</span>
            </div>

            <div class="form-group">
              <label class="form-label">DESCRIPTION</label>
              <input v-model="editForm.description" class="form-input" placeholder="Optional description for this domain" />
            </div>

            <label class="check-label">
              <input type="checkbox" v-model="editForm.active" />
              Active Domain
            </label>
          </div>

          <!-- ── ADVANCED SETTINGS (collapsible) ── -->
          <details class="adv-details" open>
            <summary class="adv-summary">
              <span class="adv-summary-left">
                <q-icon name="settings" size="15px" style="margin-right:6px" />
                ADVANCED SETTINGS
              </span>
              <q-icon name="expand_less" size="18px" class="adv-chevron" />
            </summary>

            <div class="adv-body">
              <div class="form-row2">
                <div class="form-group">
                  <label class="form-label">ALIAS LIMIT</label>
                  <input v-model.number="editForm.aliases" class="form-input" type="number" min="0" />
                  <span class="form-hint">Maximum number of aliases (0 = unlimited)</span>
                </div>
                <div class="form-group">
                  <label class="form-label">MAILBOX LIMIT</label>
                  <input v-model.number="editForm.mailboxes" class="form-input" type="number" min="0" />
                  <span class="form-hint">Maximum number of mailboxes (0 = unlimited)</span>
                </div>
                <div class="form-group">
                  <label class="form-label">QUOTA LIMIT (MB)</label>
                  <input v-model.number="editForm.quotaMB" class="form-input" type="number" min="0" />
                  <span class="form-hint">Maximum quota limit in MB (0 = unlimited)</span>
                </div>
                <div class="form-group">
                  <label class="form-label">PASSWORD EXPIRY (DAYS)</label>
                  <input v-model.number="editForm.passwordExpiry" class="form-input" type="number" min="0" placeholder="365" />
                  <span class="form-hint">Password expiry in days (empty = never)</span>
                </div>
              </div>

              <div class="form-group">
                <label class="form-label">TRANSPORT</label>
                <select v-model="editForm.transport" class="form-select-plain">
                  <option value="virtual">virtual</option>
                  <option v-for="t in transports" :key="t.id" :value="t.transport">
                    {{ t.domain }} -&gt; {{ t.transport }}
                  </option>
                </select>
                <span class="form-hint">Select the mail transport for this domain</span>
              </div>

              <label class="check-label backupmx-row">
                <input type="checkbox" v-model="editForm.backupmx" />
                Enable Backup MX
                <span class="backupmx-hint">Use this server as a backup mail exchanger</span>
              </label>
            </div>
          </details>

        </div>

        <!-- Modal footer -->
        <div class="modal-footer">
          <button class="btn-cancel" @click="showEdit = false">
            <q-icon name="close" size="14px" style="margin-right:4px;vertical-align:middle" />
            CANCEL
          </button>
          <button class="btn-primary" :disabled="savingEdit" @click="submitEdit">
            <q-icon name="save" size="14px" style="margin-right:6px;vertical-align:middle" />
            {{ savingEdit ? 'SAVING...' : 'UPDATE DOMAIN' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="modal-card modal-card-sm">
        <div class="modal-head modal-head-danger">
          <span>CONFIRM DELETE</span>
          <button class="modal-close" @click="showDeleteConfirm = false">✕</button>
        </div>
        <div class="modal-body">
          <p class="confirm-text">
            Are you sure you want to delete domain<br />
            <strong>{{ deleteTarget?.domain }}</strong>?<br />
            <span class="confirm-sub">This will delete all mailboxes, aliases, and related data.</span>
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

// ─── Data ───
const allDomains = ref<Domain[]>([])
const transports = ref<Transport[]>([])
const loading = ref(true)
const error = ref('')

// ─── Table controls ───
const search = ref('')
const rowsPerPage = ref(15)
const currentPage = ref(1)
const sortKey = ref('domain')
const sortDir = ref<'asc' | 'desc'>('asc')

const columns = [
  { key: 'domain',      label: 'DOMAIN' },
  { key: 'description', label: 'DESCRIPTION' },
  { key: 'aliases',     label: 'ALIASES' },
  { key: 'mailboxes',   label: 'MAILBOXES' },
  { key: 'transport',   label: 'TRANSPORT' },
  { key: 'active',      label: 'ACTIVE' },
  { key: 'modified',    label: 'MODIFIED' },
]

function sortBy(key: string) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

const filteredRows = computed(() => {
  const q = search.value.toLowerCase()
  let rows = allDomains.value
  if (q) rows = rows.filter(r =>
    r.domain.toLowerCase().includes(q) ||
    (r.description || '').toLowerCase().includes(q) ||
    (r.transport || '').toLowerCase().includes(q)
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

// ─── Load data ───
async function load() {
  loading.value = true
  error.value = ''
  try {
    const [domRes, trRes] = await Promise.all([
      axios.get(`${API_BASE}/domains`),
      axios.get(`${API_BASE}/transports`),
    ])
    allDomains.value = domRes.data?.data ?? []
    transports.value = (trRes.data?.data ?? []).filter((t: Transport) => t.active)
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load data'
  } finally {
    loading.value = false
  }
}
onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)
const addError = ref('')

const addForm = ref({
  domain: '', description: '', aliases: 10, mailboxes: 10,
  quotaMB: 2048, passwordExpiry: 365 as number | null,
  transport: 'virtual', active: true, backupmx: false,
})

function openAdd() {
  addForm.value = {
    domain: '', description: '', aliases: 10, mailboxes: 10,
    quotaMB: 2048, passwordExpiry: 365,
    transport: 'virtual', active: true, backupmx: false,
  }
  addError.value = ''
  showAdd.value = true
}

async function submitAdd() {
  addError.value = ''
  const f = addForm.value
  if (!f.domain.trim()) { addError.value = 'Domain name is required'; return }

  savingAdd.value = true
  try {
    await axios.post(`${API_BASE}/domains`, {
      domain: f.domain.trim().toLowerCase(),
      description: f.description,
      aliases: f.aliases,
      mailboxes: f.mailboxes,
      quota: f.quotaMB,
      transport: f.transport || 'virtual',
      backupmx: f.backupmx,
      active: f.active,
      password_expiry: f.passwordExpiry || null,
    })
    showAdd.value = false
    toast.success(`Domain ${f.domain.trim().toLowerCase()} created successfully`)
    await load()
  } catch (e: any) {
    addError.value = e?.response?.data?.error?.message || 'Failed to create domain'
    toast.error(addError.value)
  } finally {
    savingAdd.value = false
  }
}

// ─── Edit modal ───
const showEdit = ref(false)
const savingEdit = ref(false)
const editError = ref('')

const editForm = ref({
  domain: '', description: '', aliases: 10, mailboxes: 10,
  quotaMB: 2048, passwordExpiry: null as number | null,
  transport: 'virtual', active: true, backupmx: false,
})

function openEdit(row: Domain) {
  editError.value = ''
  editForm.value = {
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

async function submitEdit() {
  editError.value = ''
  const f = editForm.value
  savingEdit.value = true
  try {
    await axios.put(`${API_BASE}/domains/${encodeURIComponent(f.domain)}`, {
      description: f.description,
      aliases: f.aliases,
      mailboxes: f.mailboxes,
      quota: f.quotaMB,
      transport: f.transport || 'virtual',
      backupmx: f.backupmx,
      active: f.active,
      password_expiry: f.passwordExpiry || null,
    })
    showEdit.value = false
    toast.success(`Domain ${f.domain} updated successfully`)
    await load()
  } catch (e: any) {
    editError.value = e?.response?.data?.error?.message || 'Failed to update domain'
    toast.error(editError.value)
  } finally {
    savingEdit.value = false
  }
}

// ─── Delete ───
const showDeleteConfirm = ref(false)
const deletingRow = ref(false)
const deleteTarget = ref<Domain | null>(null)

function confirmDelete(row: Domain) {
  deleteTarget.value = row
  showDeleteConfirm.value = true
}

async function submitDelete() {
  if (!deleteTarget.value) return
  deletingRow.value = true
  try {
    const domain = deleteTarget.value.domain
    await axios.delete(`${API_BASE}/domains/${encodeURIComponent(domain)}`)
    showDeleteConfirm.value = false
    deleteTarget.value = null
    toast.success(`Domain ${domain} deleted successfully`)
    await load()
  } catch (e: any) {
    const msg = e?.response?.data?.error?.message || 'Failed to delete domain'
    error.value = msg
    toast.error(msg)
    showDeleteConfirm.value = false
  } finally {
    deletingRow.value = false
  }
}
</script>

<style scoped>
.dom-page { background: #ebf2fe; padding: 24px 28px 40px; }

.dom-header {
  display: flex; justify-content: space-between; align-items: flex-start;
  margin-bottom: 20px;
}
.dom-title { font-size: 28px; font-weight: 900; color: #1e293b; letter-spacing: -0.5px; line-height: 1; font-family: monospace; text-transform: uppercase; }
.dom-subtitle { font-size: 10px; color: #94a3b8; letter-spacing: 0.8px; margin-top: 6px; text-transform: uppercase; font-weight: 700; }

.btn-primary {
  background: #3b82f6; color: #fff; border: 2px solid #1e293b; padding: 20px 32px;
  font-size: 16px; font-weight: 900; letter-spacing: 1.6px; cursor: pointer;
  border-radius: 0; transition: all .15s; text-transform: uppercase;
  box-shadow: 3px 3px 0 #1e293b; display: flex; align-items: center;
}
.btn-primary:hover:not(:disabled) {
  background: #fff; color: #3b82f6;
  transform: translate(-1px, -1px); box-shadow: 4px 4px 0 #1e293b;
}
.btn-primary:active:not(:disabled) { transform: translate(0,0); box-shadow: none; }
.btn-primary:disabled { opacity: .5; cursor: default; }

.error-banner {
  background: #fef2f2; border: 1px solid #fca5a5; color: #dc2626;
  padding: 10px 14px; font-size: 13px; margin-bottom: 18px;
  display: flex; align-items: center; gap: 6px;
}

/* ─── Table card ─── */
.table-card { background: #fff; border: 2px solid #1e293b; }

.table-topbar {
  display: flex; justify-content: space-between; align-items: center;
  padding: 10px 14px; border-bottom: 1px solid #e2e8f0; gap: 12px; flex-wrap: wrap;
}
.controls-left { display: flex; align-items: center; gap: 8px; }
.controls-right { display: flex; align-items: center; gap: 8px; }
.per-page-wrap { display: flex; align-items: center; gap: 6px; }
.ctrl-select {
  border: 1px solid #d1d5db; padding: 4px 6px; font-size: 13px; color: #374151;
  background: #fff; border-radius: 0; outline: none;
}
.ctrl-label { font-size: 12px; color: #64748b; font-weight: 500; }
.search-input {
  border: 1px solid #d1d5db; padding: 4px 8px; font-size: 13px; color: #374151;
  width: 200px; outline: none; border-radius: 0;
}
.search-input:focus { border-color: #3b82f6; }

.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 12.5px; }
.table-head-row { background: #3b82f6; }
.table-th {
  color: #fff; font-weight: 600; letter-spacing: 0.4px; padding: 10px 14px;
  text-align: left; cursor: pointer; white-space: nowrap; user-select: none;
}
.table-th:hover { background: #2563eb; }
.sort-arrows { margin-left: 4px; font-size: 9px; opacity: .5; }
.sort-active { opacity: 1 !important; }

.table-row:nth-child(even) { background: #f8fafc; }
.table-row:hover { background: #eff6ff; }
.table-td { padding: 6px 10px; color: #374151; border-bottom: 1px solid #f1f5f9; font-size: 12px; }
.td-link { color: #1e293b; font-weight: 600; }
.cell-with-icon { display: flex; align-items: center; gap: 6px; }
.row-icon { color: #3b82f6; flex-shrink: 0; }
.mono { font-family: monospace; font-size: 12px; color: #64748b; }

.count-badge { font-weight: 700; color: #1e293b; }
.count-sep { color: #94a3b8; margin: 0 2px; }
.count-max { color: #94a3b8; font-size: 11px; }

.badge-yes { background: #dcfce7; color: #16a34a; padding: 2px 8px; font-size: 11px; font-weight: 700; }
.badge-no  { background: #fee2e2; color: #dc2626; padding: 2px 8px; font-size: 11px; font-weight: 700; }

.actions-td { display: flex; gap: 6px; align-items: center; }
.act-btn {
  padding: 4px 10px; font-size: 10px; font-weight: 800; cursor: pointer;
  border: 1px solid #1e293b; letter-spacing: 0.4px; border-radius: 0;
  display: inline-flex; align-items: center; transition: all .12s;
  box-shadow: 1px 1px 0 #1e293b; text-transform: uppercase;
}
.act-btn:hover { transform: translate(-0.5px, -0.5px); }
.act-btn:active { transform: translate(0,0); box-shadow: none; }
.act-edit { background: #3b82f6; color: #fff; }
.act-edit:hover { background: #fff; color: #3b82f6; }
.act-del  { background: #ef4444; color: #fff; }
.act-del:hover { background: #fff; color: #ef4444; }

.table-loading, .table-empty { text-align: center; padding: 28px; color: #94a3b8; font-size: 13px; }

.table-footer {
  display: flex; justify-content: space-between; align-items: center;
  padding: 10px 14px; border-top: 1px solid #e2e8f0;
}
.showing-text { font-size: 12.5px; color: #64748b; }
.pagination { display: flex; gap: 3px; }
.pg-btn {
  height: 28px; padding: 0 10px; font-size: 10px; font-weight: 700;
  color: #374151; background: #fff; border: 1px solid #d1d5db; border-radius: 0;
  cursor: pointer; letter-spacing: 0.4px; text-transform: uppercase; white-space: nowrap;
}
.pg-btn:hover:not(:disabled) { border-color: #1e293b; color: #1e293b; background: #f8fafc; }
.pg-btn:disabled { opacity: .35; cursor: default; }
.pg-active { background: #3b82f6 !important; color: #fff !important; border-color: #3b82f6 !important; }

/* ─── Modals ─── */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,.5); z-index: 9000;
  display: flex; align-items: center; justify-content: center; padding: 20px;
}
.modal-card {
  background: #fff; border: 3px solid #1e293b; width: 100%; max-width: 620px;
  max-height: 90vh; display: flex; flex-direction: column; border-radius: 0;
  box-shadow: 4px 4px 0 #1e293b;
}
.modal-card-sm { max-width: 400px; }

.modal-head {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px 20px; background: #3b82f6; color: #fff; flex-shrink: 0;
}
.modal-head-title {
  font-size: 15px; font-weight: 900; letter-spacing: 0.3px; font-family: monospace;
  text-transform: uppercase; display: flex; align-items: center;
}
.modal-head-sub { font-size: 13px; color: rgba(255,255,255,.7); margin-left: 8px; font-weight: 400; }
.modal-head-danger { background: #dc2626; }
.modal-close {
  background: transparent; border: none; color: #fff; cursor: pointer;
  font-size: 18px; line-height: 1; padding: 2px 6px; font-weight: 300;
}
.modal-close:hover { opacity: .75; }

.modal-body { padding: 20px; overflow-y: auto; flex: 1; display: flex; flex-direction: column; gap: 16px; }

.modal-error {
  background: #fef2f2; border: 2px solid #dc2626; color: #dc2626;
  padding: 10px 14px; font-size: 13px; display: flex; align-items: flex-start;
}

/* ─── Basic info card ─── */
.info-card {
  border: 2px solid #1e293b; padding: 16px; display: flex; flex-direction: column; gap: 14px;
}
.info-card-title {
  font-size: 12px; font-weight: 900; color: #1e293b; letter-spacing: 0.6px;
  text-transform: uppercase; font-family: monospace; display: flex; align-items: center;
}

/* ─── Advanced settings collapsible ─── */
.adv-details { border: 2px solid #1e293b; }
.adv-summary {
  padding: 10px 14px; cursor: pointer; font-weight: 800; font-size: 12px;
  text-transform: uppercase; letter-spacing: 0.5px; display: flex; align-items: center;
  background: #f8fafc; list-style: none; user-select: none;
  border-bottom: 2px solid transparent; transition: background .12s;
}
.adv-details[open] .adv-summary { border-bottom-color: #1e293b; background: #f1f5f9; }
.adv-summary:hover { background: #e2e8f0; }
.adv-summary::-webkit-details-marker { display: none; }
.adv-summary-left { display: flex; align-items: center; flex: 1; color: #1e293b; }
.adv-chevron { color: #64748b; transition: transform .2s; }
.adv-details[open] .adv-chevron { transform: rotate(180deg); }
.adv-body { padding: 16px; display: flex; flex-direction: column; gap: 12px; }

/* ─── Form elements ─── */
.form-row2 { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-group { display: flex; flex-direction: column; gap: 3px; }
.form-label { font-size: 11px; font-weight: 800; color: #1e293b; letter-spacing: 0.7px; text-transform: uppercase; }
.req { color: #ef4444; }
.form-input {
  border: 2px solid #1e293b; padding: 8px 10px; font-size: 13px; color: #374151;
  outline: none; border-radius: 0; width: 100%; box-sizing: border-box;
  transition: border-color .12s;
}
.form-input:focus { border-color: #3b82f6; }
.form-input-disabled { background: #f1f5f9; color: #94a3b8; cursor: not-allowed; }
.form-hint { font-size: 10px; color: #94a3b8; margin-top: 2px; }

.form-select-plain {
  width: 100%; height: 40px; padding: 0 10px; font-size: 13px; color: #374151;
  border: 2px solid #1e293b; background: #fff; border-radius: 0; outline: none;
  cursor: pointer; box-sizing: border-box; transition: border-color .12s;
}
.form-select-plain:focus { border-color: #3b82f6; }

.check-label { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 600; color: #374151; cursor: pointer; }
.check-label input[type="checkbox"] { width: 18px; height: 18px; border: 2px solid #1e293b; cursor: pointer; accent-color: #3b82f6; }
.backupmx-row { justify-content: flex-start; }
.backupmx-hint { margin-left: auto; font-size: 10px; color: #94a3b8; }

/* ─── Modal footer ─── */
.modal-footer {
  display: flex; justify-content: flex-end; gap: 10px;
  padding: 14px 20px; border-top: 2px solid #e2e8f0; flex-shrink: 0;
  background: #f8fafc;
}
.btn-cancel {
  background: #fff; border: 2px solid #1e293b; color: #374151;
  padding: 8px 20px; font-size: 11px; font-weight: 800; cursor: pointer; border-radius: 0;
  text-transform: uppercase; letter-spacing: 0.5px;
  box-shadow: 2px 2px 0 #1e293b; transition: all .12s;
  display: flex; align-items: center;
}
.btn-cancel:hover { background: #f1f5f9; transform: translate(-1px,-1px); box-shadow: 3px 3px 0 #1e293b; }
.modal-footer .btn-primary { padding: 12px 24px !important; font-size: 14px !important; font-weight: 900 !important; letter-spacing: 1.4px !important; }
.modal-footer .btn-primary:hover:not(:disabled) { transform: translate(-1px,-1px); box-shadow: 4px 4px 0 #1e293b; }
.btn-cancel:active { transform: translate(0,0); box-shadow: none; }
.btn-danger {
  background: #ef4444; color: #fff; border: 2px solid #1e293b; padding: 8px 20px;
  font-size: 11px; font-weight: 800; cursor: pointer; border-radius: 0;
  text-transform: uppercase; letter-spacing: 0.5px; box-shadow: 2px 2px 0 #1e293b;
  transition: all .12s;
}
.btn-danger:hover:not(:disabled) { background: #dc2626; }
.btn-danger:active:not(:disabled) { transform: translate(0,0); box-shadow: none; }
.btn-danger:disabled { opacity: .5; cursor: default; }

.confirm-text { font-size: 14px; color: #374151; line-height: 1.8; text-align: center; margin: 8px 0; }
.confirm-sub { font-size: 12px; color: #dc2626; }
</style>
