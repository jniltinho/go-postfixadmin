<template>
  <div class="st-page">
    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">API KEYS</div>
        <div class="page-subtitle">MANAGE YOUR PERSONAL ACCESS TOKENS</div>
      </div>
      <button class="btn-primary" @click="openAdd">
        <Icon name="plus-circle" :size="16" style="margin-right: 0.375rem; vertical-align: middle" />
        GENERATE API KEY
      </button>
    </div>

    <!-- ─── Description Callout ─── -->
    <div class="info-callout">
      <Icon name="key" :size="24" class="callout-icon" />
      <div>
        <div class="callout-heading">BearerAuth API Keys (Personal Access Tokens)</div>
        <div class="callout-text">
          Generate API keys to securely authenticate external requests (such as Swagger UI or automated scripts) without
          exposing your password. To authenticate, pass the token as an HTTP header:
          <code>Authorization: Bearer pfa_...</code>.
        </div>
      </div>
    </div>

    <!-- ─── Table ─── -->
    <AppTable
      :rows="keys"
      :columns="columns"
      row-key="id"
      :search-fields="['name', 'preview']"
      default-sort-key="name"
      :loading="loading"
    >
      <template #topbar-extra>
        <button class="btn-refresh" @click="load"><Icon name="refresh-cw" :size="14" /> REFRESH</button>
      </template>

      <template #cell-preview="{ value }">
        <span class="font-mono text-grey-7">{{ value }}</span>
      </template>
      <template #cell-created="{ value }">{{ formatDate(value) }}</template>
      <template #cell-expires_at="{ value, row }">
        <span :class="{ 'text-err font-bold': isExpired(row.expires_at) }">
          {{ value ? formatDate(value) : 'NEVER' }}
        </span>
      </template>
      <template #cell-active="{ row }">
        <button
          class="badge-toggle"
          :class="row.active ? 'badge-yes' : 'badge-no'"
          :disabled="updatingRowId === row.id"
          title="Click to toggle status"
          @click="toggleActive(row)"
        >
          {{ row.active ? 'YES' : 'NO' }}
        </button>
      </template>
      <template #actions="{ row }">
        <button class="act-btn act-del" @click="confirmDelete(row)">
          <Icon name="trash-2" :size="12" style="margin-right: 0.25rem; vertical-align: middle" />REVOKE
        </button>
      </template>
    </AppTable>

    <!-- ══════════ GENERATE MODAL ══════════ -->
    <div v-if="showAdd" class="modal-overlay" @click.self="showAdd = false">
      <div class="bg-surface border-2 border-line w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Header (blue bar) -->
        <div
          class="bg-action px-6 py-4 flex items-center justify-between flex-shrink-0"
          style="border-bottom: 2px solid var(--color-line)"
        >
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-on-action flex items-center">
            <Icon name="key" :size="20" class="mr-2" />
            GENERATE NEW API KEY
          </h3>
          <button @click="showAdd = false" class="text-on-action hover:text-on-action/80 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <!-- Scrollable Body -->
        <div class="overflow-y-auto flex-1">
          <form class="p-6 space-y-5" @submit.prevent="submitAdd">
            <!-- Token Details Section -->
            <div class="border-2 border-line p-4 space-y-4">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="info" :size="16" class="mr-2" />
                TOKEN DETAILS
              </h4>

              <!-- Key Name -->
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">
                  KEY NAME <span class="text-err">*</span>
                </label>
                <input
                  v-model="addForm.name"
                  type="text"
                  required
                  placeholder="e.g. My Swagger Client, Python Script"
                  class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium transition-colors text-sm"
                />
                <p class="text-[0.625rem] text-fg-subtle mt-1">Enter a descriptive name to identify this token later</p>
              </div>

              <!-- Expiration -->
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1"> EXPIRATION </label>
                <select
                  v-model.number="addForm.daysValid"
                  required
                  class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium transition-colors cursor-pointer text-sm bg-surface"
                >
                  <option :value="0">Never expires</option>
                  <option :value="30">30 Days</option>
                  <option :value="60">60 Days</option>
                  <option :value="90">90 Days</option>
                  <option :value="365">1 Year (365 Days)</option>
                </select>
                <p class="text-[0.625rem] text-fg-subtle mt-1">Choose when this token should expire for safety</p>
              </div>
            </div>
          </form>
        </div>

        <!-- Footer Buttons -->
        <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-line flex-shrink-0 bg-surface">
          <button
            type="button"
            @click="showAdd = false"
            class="bg-surface hover:bg-surface-2 text-fg border-2 border-line font-black px-6 py-3 shadow-[2px_2px_0px_var(--color-shadow)] transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_var(--color-shadow)] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm"
          >
            <Icon name="x" :size="16" class="mr-2" />
            CANCEL
          </button>
          <button
            type="button"
            :disabled="savingAdd"
            @click="submitAdd"
            class="bg-action hover:bg-surface hover:text-action text-on-action border-2 border-line font-black px-6 py-3 shadow-[3px_3px_0px_var(--color-shadow)] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_var(--color-shadow)] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm disabled:opacity-60"
          >
            <Icon name="plus-circle" :size="16" class="mr-2" />
            {{ savingAdd ? 'GENERATING...' : 'GENERATE TOKEN' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ KEY CREATED SUCCESS MODAL ══════════ -->
    <div v-if="showSuccessModal" class="modal-overlay" @click.self="showSuccessModal = false">
      <div class="bg-surface border-2 border-line w-full max-w-xl max-h-[90vh] flex flex-col">
        <!-- Header (emerald green success bar) -->
        <div
          class="bg-ok-strong px-6 py-4 flex items-center justify-between flex-shrink-0"
          style="border-bottom: 2px solid var(--color-line)"
        >
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-on-action flex items-center">
            <Icon name="check-circle" :size="20" class="mr-2" />
            API KEY GENERATED SUCCESSFULLY!
          </h3>
          <button @click="showSuccessModal = false" class="text-on-action hover:text-on-action/80 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <div class="overflow-y-auto flex-1 p-6 space-y-5">
          <!-- Warning Banner -->
          <div class="bg-warn-soft border-2 border-warn-strong p-3 flex items-start">
            <Icon name="alert-triangle" :size="20" class="text-warn-strong mr-2 mt-0.5 flex-shrink-0" />
            <p class="text-sm text-warn-strong font-bold">
              Make sure to copy your API key now. You will not be able to see it again!
            </p>
          </div>

          <!-- Token Details Block -->
          <div class="border-2 border-line p-4 space-y-3">
            <h4 class="text-xs font-mono font-black uppercase tracking-tight flex items-center text-fg">
              YOUR PERSONAL ACCESS TOKEN
            </h4>
            <div
              @click="copyText(createdKeyToken)"
              class="w-full bg-surface-2 hover:bg-action-soft border-2 border-dashed border-action hover:border-solid p-4 text-center font-mono text-base font-bold text-action-hover cursor-pointer break-all transition-all"
              title="Click to copy token"
            >
              {{ createdKeyToken }}
            </div>
            <p class="text-[0.625rem] text-fg-subtle text-center">Click the token block above to copy</p>
          </div>
        </div>

        <!-- Footer -->
        <div class="flex items-center justify-end px-6 py-4 border-t-2 border-line bg-surface">
          <button
            type="button"
            @click="copyAndClose"
            class="bg-action hover:bg-surface hover:text-action text-on-action border-2 border-line font-black px-6 py-3 shadow-[3px_3px_0px_var(--color-shadow)] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_var(--color-shadow)] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm"
          >
            <Icon name="copy" :size="16" class="mr-2" />
            COPY & CLOSE
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ REVOKE CONFIRM ══════════ -->
    <ConfirmDialog
      v-model="showDeleteConfirm"
      title="REVOKE TOKEN"
      confirm-label="REVOKE"
      :loading="deletingRow"
      @confirm="submitDelete"
    >
      <p class="text-sm font-medium text-fg leading-relaxed">
        Are you sure you want to revoke API key<br />
        <strong class="font-mono break-all">{{ deleteTarget?.name }}</strong
        >?<br />
        <span class="text-xs text-fg-muted mt-1 block"
          >Any external client or script using this key will immediately be blocked. This action is permanent.</span
        >
      </p>
    </ConfirmDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import http from '../utils/http'
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

const columns = [
  { key: 'name', label: 'KEY NAME' },
  { key: 'preview', label: 'TOKEN PREVIEW' },
  { key: 'created', label: 'CREATED' },
  { key: 'expires_at', label: 'EXPIRES AT' },
  { key: 'active', label: 'ACTIVE' },
]

// ─── State ───
const keys = ref<ApiKey[]>([])
const loading = ref(true)
const updatingRowId = ref<number | null>(null)

// ─── Load data ───
async function load() {
  loading.value = true
  try {
    const res = await http.get(`${API_BASE}/settings/apikeys`)
    keys.value = res.data?.data ?? []
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load settings data')
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)
const addForm = ref({ name: '', daysValid: 0 })

const showSuccessModal = ref(false)
const createdKeyToken = ref('')

function openAdd() {
  addForm.value = { name: '', daysValid: 0 }
  showAdd.value = true
}

async function submitAdd() {
  const f = addForm.value
  if (!f.name.trim()) {
    toast.error('Key name is required')
    return
  }

  savingAdd.value = true
  try {
    const res = await http.post(`${API_BASE}/settings/apikeys`, {
      name: f.name.trim(),
      days_valid: f.daysValid,
    })

    createdKeyToken.value = res.data?.data?.token ?? ''
    showAdd.value = false
    showSuccessModal.value = true
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to generate API Key')
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
    await http.put(`${API_BASE}/settings/apikeys/${row.id}`, { active: targetState })
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
    await http.delete(`${API_BASE}/settings/apikeys/${id}`)
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
  return new Date(ts).toLocaleString('pt-BR', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function isExpired(ts: string | null): boolean {
  if (!ts) return false
  return new Date(ts) < new Date()
}
</script>

<style scoped>
.st-page {
  background: var(--color-page);
  padding: 1.5rem 1.75rem 2.5rem;
}

/* .page-header, .btn-primary, .table-card now from global style.css */

/* Callout */
.info-callout {
  background: var(--color-surface);
  border: 2px solid var(--color-line);
  box-shadow: 2px 2px 0 var(--color-shadow);
  padding: 1rem 1.25rem;
  display: flex;
  align-items: flex-start;
  gap: 0.875rem;
  margin-bottom: 1.25rem;
}
.callout-icon {
  color: var(--color-action);
  flex-shrink: 0;
  margin-top: 0.125rem;
}
.callout-heading {
  font-size: 0.875rem;
  font-weight: 800;
  color: var(--color-fg);
  text-transform: uppercase;
  margin-bottom: 0.25rem;
}
.callout-text {
  font-size: 0.75rem;
  color: var(--color-fg-2);
  line-height: 1.6;
}
.callout-text code {
  background: var(--color-surface-3);
  padding: 0.125rem 0.375rem;
  border: 1px solid var(--color-field-line);
  font-family: monospace;
  font-size: 0.71875rem;
  font-weight: bold;
}

.ctrl-label-bold {
  font-size: 0.8125rem;
  color: var(--color-fg);
  font-weight: 800;
  letter-spacing: 0.03125rem;
}

.btn-refresh {
  display: inline-flex;
  align-items: center;
  gap: 0.3125rem;
  white-space: nowrap;
  background: var(--color-surface-2);
  border: 2px solid var(--color-field-line);
  color: var(--color-fg-2);
  padding: 0.25rem 0.75rem;
  height: 1.9375rem;
  box-sizing: border-box;
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.025rem;
  cursor: pointer;
  border-radius: 0;
  transition: all 0.12s;
  text-transform: uppercase;
}
.btn-refresh:hover {
  background: var(--color-surface-3);
  border-color: var(--color-fg-subtle);
  color: var(--color-fg);
}

.table-wrap {
  overflow-x: auto;
}
.mono {
  font-family: monospace;
  font-size: 0.75rem;
  color: var(--color-fg-muted);
}
.font-semibold {
  font-weight: 700;
  color: var(--color-fg);
}
.font-mono {
  font-family: monospace;
}
.text-grey-7 {
  color: var(--color-fg-2);
}

.badge-toggle {
  border: 1px solid var(--color-line);
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
  border-radius: 0;
  box-shadow: 1px 1px 0 var(--color-shadow);
  transition: all 0.1s;
}
.badge-toggle:hover:not(:disabled) {
  transform: translate(-0.5px, -0.5px);
  box-shadow: 1.5px 1.5px 0 var(--color-shadow);
}
.badge-toggle:active:not(:disabled) {
  transform: translate(0, 0);
  box-shadow: none;
}
.badge-toggle:disabled {
  opacity: 0.6;
  cursor: default;
}

.table-loading,
.table-empty {
  text-align: center;
  padding: 1.75rem;
  color: var(--color-fg-subtle);
  font-size: 0.8125rem;
}

/* Migration note: Main Add and Revoke modals converted to BrutalModal. Success modal still legacy. Old CSS cleaned. */

.success-banner {
  background: var(--color-warn-soft);
  border: 2px solid var(--color-warn-strong);
  color: var(--color-warn-strong);
  padding: 0.75rem;
  font-size: 0.78125rem;
  font-weight: 700;
  display: flex;
  align-items: center;
}

/* ─── Basic info card ─── */
.info-card {
  border: 2px solid var(--color-line);
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
}
.info-card-title {
  font-size: 0.75rem;
  font-weight: 900;
  color: var(--color-fg);
  letter-spacing: 0.0375rem;
  text-transform: uppercase;
  font-family: monospace;
  display: flex;
  align-items: center;
}

.raw-token-box {
  background: var(--color-surface-2);
  border: 2px dashed var(--color-action);
  padding: 0.875rem;
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--color-action-hover);
  text-align: center;
  cursor: pointer;
  transition: all 0.15s;
  word-break: break-all;
}
.raw-token-box:hover {
  background: var(--color-action-soft);
  border-style: solid;
}

/* ─── Form elements ─── */
.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.1875rem;
}
.form-label {
  font-size: 0.6875rem;
  font-weight: 800;
  color: var(--color-fg);
  letter-spacing: 0.04375rem;
  text-transform: uppercase;
}
.req {
  color: var(--color-err);
}
.form-input {
  border: 2px solid var(--color-line);
  padding: 0.5rem 0.625rem;
  font-size: 0.8125rem;
  color: var(--color-fg-2);
  outline: none;
  border-radius: 0;
  width: 100%;
  box-sizing: border-box;
  height: 2.5rem;
}
.form-select-plain {
  border: 2px solid var(--color-line);
  padding: 0 0.625rem;
  font-size: 0.8125rem;
  color: var(--color-fg-2);
  background: var(--color-surface);
  border-radius: 0;
  width: 100%;
  height: 2.5rem;
  outline: none;
  cursor: pointer;
}
.form-select-plain:focus {
  border-color: var(--color-action);
}
.form-hint {
  font-size: 0.625rem;
  color: var(--color-fg-subtle);
  margin-top: 0.125rem;
}

/* Old modal footer styles removed (handled by BrutalModal + global buttons) */
</style>
