<template>
  <div class="st-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">SYSTEM SETTINGS</div>
        <div class="page-subtitle">MANAGE YOUR ACCOUNT, INTEGRATIONS AND API KEYS</div>
      </div>
      <button class="btn-primary" @click="openAdd">
        <Icon name="plus-circle" :size="16" style="margin-right:6px;vertical-align:middle" />
        GENERATE API KEY
      </button>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <Icon name="alert-triangle" :size="16" class="mr-1" /> {{ error }}
    </div>

    <!-- ─── Description Callout ─── -->
    <div class="info-callout">
      <Icon name="key" :size="24" class="callout-icon" />
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
            <Icon name="refresh-cw" :size="14" /> REFRESH
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
                <div class="spinner mx-auto" />
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
                  <Icon name="trash-2" :size="12" style="margin-right:4px;vertical-align:middle" />REVOKE
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ══════════ GENERATE MODAL ══════════ -->
    <div v-if="showAdd" class="modal-overlay" @click.self="showAdd = false">
      <div class="bg-white border-4 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Header (blue bar) -->
        <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0" style="border-bottom: 2px solid #1e293b;">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="key" :size="20" class="mr-2" />
            GENERATE NEW API KEY
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
            <!-- Token Details Section -->
            <div class="border-2 border-brand-text p-4 space-y-4">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="info" :size="16" class="mr-2" />
                TOKEN DETAILS
              </h4>

              <!-- Key Name -->
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                  KEY NAME <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="addForm.name"
                  type="text"
                  required
                  placeholder="e.g. My Swagger Client, Python Script"
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm"
                />
                <p class="text-[10px] text-gray-400 mt-1">Enter a descriptive name to identify this token later</p>
              </div>

              <!-- Expiration -->
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                  EXPIRATION
                </label>
                <select
                  v-model.number="addForm.daysValid"
                  required
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors cursor-pointer text-sm bg-white"
                >
                  <option :value="0">Never expires</option>
                  <option :value="30">30 Days</option>
                  <option :value="60">60 Days</option>
                  <option :value="90">90 Days</option>
                  <option :value="365">1 Year (365 Days)</option>
                </select>
                <p class="text-[10px] text-gray-400 mt-1">Choose when this token should expire for safety</p>
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
            <Icon name="plus-circle" :size="16" class="mr-2" />
            {{ savingAdd ? 'GENERATING...' : 'GENERATE TOKEN' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ KEY CREATED SUCCESS MODAL ══════════ -->
    <div v-if="showSuccessModal" class="modal-overlay" @click.self="showSuccessModal = false">
      <div class="bg-white border-4 border-brand-text w-full max-w-xl max-h-[90vh] flex flex-col">
        <!-- Header (emerald green success bar) -->
        <div class="bg-emerald-600 px-6 py-4 flex items-center justify-between flex-shrink-0" style="border-bottom: 2px solid #1e293b;">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="check-circle" :size="20" class="mr-2" />
            API KEY GENERATED SUCCESSFULLY!
          </h3>
          <button @click="showSuccessModal = false" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <div class="overflow-y-auto flex-1 p-6 space-y-5">
          <!-- Warning Banner -->
          <div class="bg-amber-50 border-2 border-amber-600 p-3 flex items-start">
            <Icon name="alert-triangle" :size="20" class="text-amber-600 mr-2 mt-0.5 flex-shrink-0" />
            <p class="text-sm text-amber-700 font-bold">
              Make sure to copy your API key now. You will not be able to see it again!
            </p>
          </div>

          <!-- Token Details Block -->
          <div class="border-2 border-brand-text p-4 space-y-3">
            <h4 class="text-xs font-mono font-black uppercase tracking-tight flex items-center text-brand-text">
              YOUR PERSONAL ACCESS TOKEN
            </h4>
            <div 
              @click="copyText(createdKeyToken)"
              class="w-full bg-slate-50 hover:bg-blue-50 border-2 border-dashed border-blue-500 hover:border-solid p-4 text-center font-mono text-base font-bold text-blue-600 cursor-pointer break-all transition-all"
              title="Click to copy token"
            >
              {{ createdKeyToken }}
            </div>
            <p class="text-[10px] text-gray-400 text-center">Click the token block above to copy</p>
          </div>
        </div>

        <!-- Footer -->
        <div class="flex items-center justify-end px-6 py-4 border-t-2 border-brand-text bg-white">
          <button type="button" @click="copyAndClose"
            class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-3 shadow-[3px_3px_0px_#1E293B] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm">
            <Icon name="copy" :size="16" class="mr-2" />
            COPY & CLOSE
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ REVOKE CONFIRM ══════════ -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="bg-white border-4 border-brand-text w-full max-w-md flex flex-col">
        <!-- Header -->
        <div class="bg-red-600 px-6 py-4 flex items-center justify-between flex-shrink-0" style="border-bottom: 2px solid #1e293b;">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="trash-2" :size="20" class="mr-2" />
            REVOKE TOKEN
          </h3>
          <button @click="showDeleteConfirm = false" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <!-- Body -->
        <div class="p-6">
          <p class="text-sm font-bold text-brand-text leading-relaxed">
            Are you sure you want to revoke API key<br />
            <strong class="text-red-600">{{ deleteTarget?.name }}</strong>?<br />
            <span class="block text-xs text-gray-500 font-medium mt-2 leading-normal">
              Any external client or script using this key will immediately be blocked. This action is permanent.
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
            {{ deletingRow ? 'REVOKING...' : 'REVOKE' }}
          </button>
        </div>
      </div>
    </div>

  </div>
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

/* .page-header, .btn-primary, .error-banner, .table-card now from global style.css */

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
.table-card { background: #fff; border: 3px solid #1e293b; box-shadow: 3px 3px 0 #1e293b; }
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
  color: #fff; font-weight: 700; letter-spacing: 0.4px; padding: 10px;
  text-align: left; white-space: nowrap; user-select: none; font-size: 12px;
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

.actions-td { display: flex; gap: 6px; align-items: center; justify-content: flex-end; }
.act-btn {
  padding: 4px 10px; font-size: 10px; font-weight: 800; cursor: pointer;
  border: 1px solid #1e293b; letter-spacing: 0.4px; border-radius: 0;
  display: inline-flex; align-items: center; transition: all .12s;
  box-shadow: 1px 1px 0 #1e293b; text-transform: uppercase;
}
.act-btn:hover { transform: translate(-0.5px, -0.5px); }
.act-btn:active { transform: translate(0,0); box-shadow: none; }
.act-del  { background: #dc2626; color: #fff; }
.act-del:hover { background: #fff; color: #dc2626; }

.table-loading, .table-empty { text-align: center; padding: 28px; color: #94a3b8; font-size: 13px; }

/* Migration note: Main Add and Revoke modals converted to BrutalModal. Success modal still legacy. Old CSS cleaned. */

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

/* Old modal footer styles removed (handled by BrutalModal + global buttons) */
</style>