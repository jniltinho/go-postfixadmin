<template>
  <q-page class="st-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">SYSTEM SETTINGS</div>
        <div class="page-subtitle">MANAGE YOUR ACCOUNT, INTEGRATIONS AND API KEYS</div>
      </div>
      <button class="btn-primary" @click="openAdd">
        <q-icon name="add_circle" size="16px" style="margin-right:6px;vertical-align:middle" />
        GENERATE API KEY
      </button>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <q-icon name="warning" size="16px" /> {{ error }}
    </div>

    <!-- ─── Description Callout ─── -->
    <div class="info-callout">
      <q-icon name="vpn_key" size="24px" class="callout-icon" />
      <div>
        <div class="callout-heading">BearerAuth API Keys (Personal Access Tokens)</div>
        <div class="callout-text">
          Generate API keys to securely authenticate external requests (such as Swagger UI or automated scripts) without exposing your password.
          To authenticate, pass the token as an HTTP header: <code>Authorization: Bearer pfa_...</code>.
        </div>
      </div>
    </div>

    <!-- ─── Table card ─── -->
    <div class="table-card">
      <div class="table-topbar">
        <div class="controls-left">
          <span class="ctrl-label-bold">ACTIVE PERSONAL ACCESS TOKENS</span>
        </div>
        <div class="controls-right">
          <button class="btn-refresh" @click="load">
            <q-icon name="refresh" size="14px" /> REFRESH
          </button>
        </div>
      </div>

      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr class="table-head-row">
              <th class="table-th">KEY NAME</th>
              <th class="table-th">TOKEN PREVIEW</th>
              <th class="table-th">CREATED</th>
              <th class="table-th">EXPIRES AT</th>
              <th class="table-th">ACTIVE</th>
              <th class="table-th text-center">ACTIONS</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="6" class="table-loading">
                <q-spinner color="primary" size="24px" />
              </td>
            </tr>
            <tr v-else-if="keys.length === 0">
              <td colspan="6" class="table-empty">No API keys generated yet. Click "Generate API Key" to create one.</td>
            </tr>
            <tr v-for="row in keys" :key="row.id" class="table-row">
              <td class="table-td font-semibold">{{ row.name }}</td>
              <td class="table-td font-mono text-grey-7">{{ row.preview }}</td>
              <td class="table-td">{{ formatDate(row.created) }}</td>
              <td class="table-td font-semibold" :class="{ 'text-red-500': isExpired(row.expires_at) }">
                {{ row.expires_at ? formatDate(row.expires_at) : 'NEVER' }}
              </td>
              <td class="table-td">
                <button
                  class="badge-toggle"
                  :class="row.active ? 'badge-yes' : 'badge-no'"
                  @click="toggleActive(row)"
                  :disabled="updatingRowId === row.id"
                  title="Click to toggle status"
                >
                  {{ row.active ? 'YES' : 'NO' }}
                </button>
              </td>
              <td class="table-td actions-td text-center justify-center">
                <button class="act-btn act-del" @click="confirmDelete(row)">
                  <q-icon name="delete" size="12px" style="margin-right:4px;vertical-align:middle" />REVOKE
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ══════════ GENERATE MODAL ══════════ -->
    <div v-if="showAdd" class="modal-overlay" @click.self="showAdd = false">
      <div class="modal-card">
        <div class="modal-head">
          <span class="modal-head-title">
            <q-icon name="vpn_key" size="18px" style="margin-right:8px;vertical-align:middle" />
            GENERATE NEW API KEY
          </span>
          <button class="modal-close" @click="showAdd = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="addError" class="modal-error">
            <q-icon name="warning" size="14px" style="margin-right:6px;flex-shrink:0" />
            {{ addError }}
          </div>

          <div class="info-card">
            <div class="info-card-title">
              <q-icon name="info" size="15px" style="margin-right:6px" />
              TOKEN DETAILS
            </div>

            <div class="form-group">
              <label class="form-label">KEY NAME <span class="req">*</span></label>
              <input v-model="addForm.name" class="form-input" placeholder="e.g. My Swagger Client, Python Script" />
              <span class="form-hint">Enter a descriptive name to identify this token later</span>
            </div>

            <div class="form-group">
              <label class="form-label">EXPIRATION</label>
              <select v-model.number="addForm.daysValid" class="form-select-plain">
                <option :value="0">Never expires</option>
                <option :value="30">30 Days</option>
                <option :value="60">60 Days</option>
                <option :value="90">90 Days</option>
                <option :value="365">1 Year (365 Days)</option>
              </select>
              <span class="form-hint">Choose when this token should expire for safety</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showAdd = false">
            <q-icon name="close" size="14px" style="margin-right:4px;vertical-align:middle" />CANCEL
          </button>
          <button class="btn-primary" :disabled="savingAdd" @click="submitAdd">
            <q-icon name="add_circle" size="14px" style="margin-right:6px;vertical-align:middle" />
            {{ savingAdd ? 'GENERATING...' : 'GENERATE TOKEN' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ KEY CREATED SUCCESS MODAL ══════════ -->
    <div v-if="showSuccessModal" class="modal-overlay">
      <div class="modal-card">
        <div class="modal-head modal-head-success">
          <span class="modal-head-title">
            <q-icon name="check_circle" size="18px" style="margin-right:8px;vertical-align:middle" />
            API KEY GENERATED SUCCESSFULLY!
          </span>
        </div>
        <div class="modal-body">
          <div class="success-banner">
            <q-icon name="warning" size="20px" class="q-mr-sm" />
            <span>Make sure to copy your API key now. You will not be able to see it again!</span>
          </div>

          <div class="info-card">
            <div class="info-card-title">YOUR PERSONAL ACCESS TOKEN</div>
            <div class="raw-token-box font-mono" @click="copyText(createdKeyToken)">
              {{ createdKeyToken }}
            </div>
            <span class="form-hint text-center">Click the token block or the button below to copy</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-primary" @click="copyAndClose">
            <q-icon name="content_copy" size="14px" style="margin-right:6px;vertical-align:middle" />
            COPY & CLOSE
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ REVOKE CONFIRM ══════════ -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="modal-card modal-card-sm">
        <div class="modal-head modal-head-danger">
          <span class="modal-head-title">
            <q-icon name="delete" size="16px" style="margin-right:6px;vertical-align:middle" />
            REVOKE TOKEN
          </span>
          <button class="modal-close" @click="showDeleteConfirm = false">✕</button>
        </div>
        <div class="modal-body">
          <p class="confirm-text">
            Are you sure you want to revoke API key<br />
            <strong>{{ deleteTarget?.name }}</strong>?<br />
            <span class="confirm-sub">Any external client or script using this key will immediately be blocked. This action is permanent.</span>
          </p>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showDeleteConfirm = false">CANCEL</button>
          <button class="btn-danger" :disabled="deletingRow" @click="submitDelete">
            {{ deletingRow ? 'REVOKING...' : 'REVOKE' }}
          </button>
        </div>
      </div>
    </div>

  </q-page>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()

interface ApiKey {
  id: number
  name: string
  preview: string
  created: string
  expires_at: string | null
  active: boolean
}

// ─── State ───
const keys = ref<ApiKey[]>([])
const loading = ref(true)
const error = ref('')
const updatingRowId = ref<number | null>(null)

// ─── Load data ───
async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get(`${API_BASE}/settings/apikeys`)
    keys.value = res.data?.data ?? []
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load settings data'
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)
const addError = ref('')
const addForm = ref({ name: '', daysValid: 0 })

const showSuccessModal = ref(false)
const createdKeyToken = ref('')

function openAdd() {
  addForm.value = { name: '', daysValid: 0 }
  addError.value = ''
  showAdd.value = true
}

async function submitAdd() {
  addError.value = ''
  const f = addForm.value
  if (!f.name.trim()) {
    addError.value = 'Key name is required'
    return
  }

  savingAdd.value = true
  try {
    const res = await axios.post(`${API_BASE}/settings/apikeys`, {
      name: f.name.trim(),
      days_valid: f.daysValid
    })
    
    createdKeyToken.value = res.data?.data?.token ?? ''
    showAdd.value = false
    showSuccessModal.value = true
    await load()
  } catch (e: any) {
    addError.value = e?.response?.data?.error?.message || 'Failed to generate API Key'
    toast.error(addError.value)
  } finally {
    savingAdd.value = false
  }
}

function copyText(txt: string) {
  if (!txt) return
  navigator.clipboard.writeText(txt)
  toast.success('API Key copied to clipboard!')
}

function copyAndClose() {
  copyText(createdKeyToken.value)
  showSuccessModal.value = false
  createdKeyToken.value = ''
}

// ─── Toggle active status ───
async function toggleActive(row: ApiKey) {
  updatingRowId.value = row.id
  const targetState = !row.active
  try {
    await axios.put(`${API_BASE}/settings/apikeys/${row.id}`, { active: targetState })
    row.active = targetState
    toast.success(`Key "${row.name}" status updated successfully`)
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to update key status')
  } finally {
    updatingRowId.value = null
  }
}

// ─── Revoke / Delete ───
const showDeleteConfirm = ref(false)
const deletingRow = ref(false)
const deleteTarget = ref<ApiKey | null>(null)

function confirmDelete(row: ApiKey) {
  deleteTarget.value = row
  showDeleteConfirm.value = true
}

async function submitDelete() {
  if (!deleteTarget.value) return
  deletingRow.value = true
  try {
    const id = deleteTarget.value.id
    const name = deleteTarget.value.name
    await axios.delete(`${API_BASE}/settings/apikeys/${id}`)
    showDeleteConfirm.value = false
    deleteTarget.value = null
    toast.success(`API Key "${name}" revoked successfully`)
    await load()
  } catch (e: any) {
    const msg = e?.response?.data?.error?.message || 'Failed to revoke API key'
    toast.error(msg)
    showDeleteConfirm.value = false
  } finally {
    deletingRow.value = false
  }
}

// ─── Helpers ───
function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function isExpired(ts: string | null): boolean {
  if (!ts) return false
  return new Date(ts) < new Date()
}
</script>

<style scoped>
.st-page { background: #ebf2fe; padding: 24px 28px 40px; }

.page-header {
  display: flex; justify-content: space-between; align-items: flex-start;
  margin-bottom: 20px;
}
.page-title { font-size: 28px; font-weight: 900; color: #1e293b; letter-spacing: -0.5px; line-height: 1; font-family: monospace; text-transform: uppercase; }
.page-subtitle { font-size: 10px; color: #94a3b8; letter-spacing: 0.8px; margin-top: 6px; text-transform: uppercase; font-weight: 700; }

.btn-primary {
  background: #3b82f6; color: #fff; border: 2px solid #1e293b; padding: 14px 24px;
  font-size: 11px; font-weight: 800; letter-spacing: 0.6px; cursor: pointer;
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

/* Callout */
.info-callout {
  background: #fff; border: 2px solid #1e293b; box-shadow: 2px 2px 0 #1e293b;
  padding: 16px 20px; display: flex; align-items: flex-start; gap: 14px; margin-bottom: 20px;
}
.callout-icon { color: #3b82f6; flex-shrink: 0; margin-top: 2px; }
.callout-heading { font-size: 14px; font-weight: 800; color: #1e293b; text-transform: uppercase; margin-bottom: 4px; }
.callout-text { font-size: 12px; color: #475569; line-height: 1.6; }
.callout-text code { background: #f1f5f9; padding: 2px 6px; border: 1px solid #cbd5e1; font-family: monospace; font-size: 11.5px; font-weight: bold; }

/* ─── Table card ─── */
.table-card { background: #fff; border: 2px solid #1e293b; box-shadow: 2px 2px 0 #1e293b; }
.table-topbar {
  display: flex; justify-content: space-between; align-items: center;
  padding: 12px 14px; border-bottom: 1px solid #e2e8f0; gap: 12px;
}
.ctrl-label-bold { font-size: 13px; color: #1e293b; font-weight: 800; letter-spacing: 0.5px; }

.btn-refresh {
  background: #f8fafc; border: 1px solid #cbd5e1; color: #475569; padding: 5px 12px;
  font-size: 10px; font-weight: 800; letter-spacing: 0.4px; cursor: pointer;
  border-radius: 0; transition: all .12s; text-transform: uppercase;
}
.btn-refresh:hover { background: #f1f5f9; border-color: #94a3b8; color: #1e293b; }

.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 12.5px; }
.table-head-row { background: #3b82f6; }
.table-th {
  color: #fff; font-weight: 600; letter-spacing: 0.4px; padding: 10px 14px;
  text-align: left; white-space: nowrap; user-select: none;
}
.table-row:nth-child(even) { background: #f8fafc; }
.table-row:hover { background: #eff6ff; }
.table-td { padding: 8px 14px; color: #374151; border-bottom: 1px solid #f1f5f9; font-size: 12px; }
.mono { font-family: monospace; font-size: 12px; color: #64748b; }
.font-semibold { font-weight: 700; color: #1e293b; }
.font-mono { font-family: monospace; }
.text-grey-7 { color: #475569; }

.badge-toggle {
  border: 1px solid #1e293b; padding: 3px 10px; font-size: 10px; font-weight: 850;
  cursor: pointer; border-radius: 0; box-shadow: 1px 1px 0 #1e293b; transition: all .1s;
}
.badge-toggle:hover:not(:disabled) { transform: translate(-0.5px, -0.5px); box-shadow: 1.5px 1.5px 0 #1e293b; }
.badge-toggle:active:not(:disabled) { transform: translate(0,0); box-shadow: none; }
.badge-toggle:disabled { opacity: 0.6; cursor: default; }

.badge-yes { background: #dcfce7; color: #16a34a; }
.badge-no  { background: #fee2e2; color: #dc2626; }

.actions-td { display: flex; gap: 6px; align-items: center; }
.act-btn {
  padding: 4px 10px; font-size: 10px; font-weight: 800; cursor: pointer;
  border: 1px solid #1e293b; letter-spacing: 0.4px; border-radius: 0;
  display: inline-flex; align-items: center; transition: all .12s;
  box-shadow: 1px 1px 0 #1e293b; text-transform: uppercase;
}
.act-btn:hover { transform: translate(-0.5px, -0.5px); box-shadow: 2px 2px 0 #1e293b; }
.act-btn:active { transform: translate(0,0); box-shadow: none; }
.act-del  { background: #ef4444; color: #fff; }
.act-del:hover { background: #fff; color: #ef4444; }

.table-loading, .table-empty { text-align: center; padding: 28px; color: #94a3b8; font-size: 13px; }

/* ─── Modals ─── */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,.5); z-index: 9000;
  display: flex; align-items: center; justify-content: center; padding: 20px;
}
.modal-card {
  background: #fff; border: 3px solid #1e293b; width: 100%; max-width: 500px;
  max-height: 90vh; display: flex; flex-direction: column; border-radius: 0;
  box-shadow: 4px 4px 0 #1e293b;
}
.modal-card-sm { max-width: 400px; }

.modal-head {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px 20px; background: #3b82f6; color: #fff; flex-shrink: 0;
}
.modal-head-title {
  font-size: 14px; font-weight: 900; letter-spacing: 0.5px; font-family: monospace;
  text-transform: uppercase; display: flex; align-items: center;
}
.modal-head-success { background: #16a34a; }
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

.success-banner {
  background: #fef3c7; border: 2px solid #d97706; color: #d97706;
  padding: 12px; font-size: 12.5px; font-weight: 700; display: flex; align-items: center;
}

/* ─── Basic info card ─── */
.info-card {
  border: 2px solid #1e293b; padding: 16px; display: flex; flex-direction: column; gap: 14px;
}
.info-card-title {
  font-size: 12px; font-weight: 900; color: #1e293b; letter-spacing: 0.6px;
  text-transform: uppercase; font-family: monospace; display: flex; align-items: center;
}

.raw-token-box {
  background: #f8fafc; border: 2px dashed #3b82f6; padding: 14px; font-size: 15px;
  font-weight: 700; color: #2563eb; text-align: center; cursor: pointer;
  transition: all .15s; word-break: break-all;
}
.raw-token-box:hover {
  background: #eff6ff; border-style: solid;
}

/* ─── Form elements ─── */
.form-group { display: flex; flex-direction: column; gap: 3px; }
.form-label { font-size: 11px; font-weight: 800; color: #1e293b; letter-spacing: 0.7px; text-transform: uppercase; }
.req { color: #ef4444; }
.form-input { border: 2px solid #1e293b; padding: 8px 10px; font-size: 13px; color: #374151; outline: none; border-radius: 0; width: 100%; box-sizing: border-box; height: 40px; }
.form-select-plain { border: 2px solid #1e293b; padding: 0 10px; font-size: 13px; color: #374151; background: #fff; border-radius: 0; width: 100%; height: 40px; outline: none; cursor: pointer; }
.form-select-plain:focus { border-color: #3b82f6; }
.form-hint { font-size: 10px; color: #94a3b8; margin-top: 2px; }

/* ─── Modal footer ─── */
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 14px 20px; border-top: 2px solid #e2e8f0; flex-shrink: 0; background: #f8fafc; }
.btn-cancel { background: #fff; border: 2px solid #1e293b; color: #374151; padding: 8px 20px; font-size: 11px; font-weight: 800; cursor: pointer; border-radius: 0; text-transform: uppercase; letter-spacing: 0.5px; box-shadow: 2px 2px 0 #1e293b; transition: all .12s; display: flex; align-items: center; }
.btn-cancel:hover { background: #f1f5f9; transform: translate(-1px,-1px); box-shadow: 3px 3px 0 #1e293b; }
.btn-cancel:active { transform: translate(0,0); box-shadow: none; }
.btn-danger { background: #ef4444; color: #fff; border: 2px solid #1e293b; padding: 8px 20px; font-size: 11px; font-weight: 800; cursor: pointer; border-radius: 0; text-transform: uppercase; letter-spacing: 0.5px; box-shadow: 2px 2px 0 #1e293b; transition: all .12s; }
.btn-danger:hover:not(:disabled) { background: #dc2626; transform: translate(-1px,-1px); box-shadow: 3px 3px 0 #1e293b; }
.btn-danger:active:not(:disabled) { transform: translate(0,0); box-shadow: none; }
.btn-danger:disabled { opacity: .5; cursor: default; }
.confirm-text { font-size: 13px; color: #374151; line-height: 1.8; text-align: center; margin: 8px 0; }
.confirm-sub { font-size: 11px; color: #dc2626; display: block; margin-top: 6px; }
</style>