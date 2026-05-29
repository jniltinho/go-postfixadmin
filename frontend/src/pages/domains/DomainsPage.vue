<template>
  <div class="dom-page">

    <!-- ─── Header ─── -->
    <div class="dom-header">
      <div>
        <div class="dom-title">DOMAINS</div>
        <div class="dom-subtitle">MANAGE AND MONITOR YOUR REGISTERED EMAIL DOMAINS.</div>
      </div>
      <button class="btn-add-big" @click="openAdd">
        <Icon name="plus-circle" :size="20" style="margin-right:12px;vertical-align:middle" />
        ADD DOMAIN
      </button>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <Icon name="alert-triangle" :size="16" class="mr-1" /> {{ error }}
    </div>

    <!-- ─── Table ─── -->
    <AppTable
      :rows="allDomains"
      :columns="columns"
      row-key="domain"
      :search-fields="['domain', 'description', 'transport']"
      default-sort-key="domain"
      :loading="loading"
      @edit="openEdit"
      @delete="confirmDelete"
    >
      <template #cell-domain="{ row }">
        <div class="cell-with-icon">
          <Icon name="globe" :size="14" class="row-icon" />
          {{ row.domain }}
        </div>
      </template>
      <template #cell-description="{ value }">{{ value || '—' }}</template>
      <template #cell-aliases="{ row }">
        <div class="progress-bar-wrap">
          <div class="progress-bar-fill" :style="{ width: (row.aliases > 0 ? Math.min(100, Math.round(((row.alias_count ?? 0) / row.aliases) * 100)) : 0) + '%' }"></div>
          <div class="progress-bar-text">{{ row.alias_count ?? 0 }} / {{ row.aliases }}</div>
        </div>
      </template>
      <template #cell-mailboxes="{ row }">
        <div class="progress-bar-wrap">
          <div class="progress-bar-fill" :style="{ width: (row.mailboxes > 0 ? Math.min(100, Math.round(((row.mailbox_count ?? 0) / row.mailboxes) * 100)) : 0) + '%' }"></div>
          <div class="progress-bar-text">{{ row.mailbox_count ?? 0 }} / {{ row.mailboxes }}</div>
        </div>
      </template>
      <template #cell-transport="{ value }"><span class="mono">{{ value || 'virtual' }}</span></template>
      <template #cell-backupmx="{ value }">{{ value ? 'Yes' : 'No' }}</template>
      <template #cell-active="{ value }">
        <span :class="value ? 'badge-yes' : 'badge-no'">{{ value ? 'YES' : 'NO' }}</span>
      </template>
      <template #cell-modified="{ value }">{{ formatDate(value) }}</template>
      <template #cell-password_expiry="{ value }"><span class="mono">{{ value ?? 0 }}</span></template>
    </AppTable>

    <!-- ══════════ ADD DOMAIN MODAL (exact pattern from form_add_domain.html) ══════════ -->
    <div v-if="showAdd" class="modal-overlay" @click.self="closeAddDomain">
      <div class="bg-white border-4 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Header -->
        <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="plus-circle" :size="20" class="mr-2" />
            ADD DOMAIN
          </h3>
          <button @click="closeAddDomain" class="text-white hover:text-gray-200 transition-colors">
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
            <!-- Basic Information -->
            <div class="border-2 border-brand-text p-4 space-y-4">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="info" :size="16" class="mr-2" />
                BASIC INFORMATION
              </h4>

              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                  DOMAIN NAME <span class="text-red-500">*</span>
                </label>
                <input v-model="addForm.domain" type="text" required placeholder="example.com"
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                <p class="text-[10px] text-gray-500 mt-1">Enter a valid domain name (e.g., example.com)</p>
              </div>

              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">DESCRIPTION</label>
                <input v-model="addForm.description" type="text" placeholder="Optional description for this domain"
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
              </div>

              <div class="flex items-center pt-1">
                <input type="checkbox" id="dom-add-active" v-model="addForm.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                <label for="dom-add-active" class="ml-2 text-sm font-bold cursor-pointer">Active Domain</label>
              </div>
            </div>

            <!-- Advanced Settings (exact from reference) -->
            <details class="border-2 border-brand-text group" open>
              <summary class="p-3 cursor-pointer font-bold uppercase tracking-tight text-sm flex items-center bg-gray-50 hover:bg-gray-100 transition-colors">
                <Icon name="settings" :size="16" class="mr-2" />
                ADVANCED SETTINGS
                <Icon name="chevron-down" :size="16" class="ml-auto group-open:rotate-180 transition-transform" />
              </summary>
              <div class="p-4 space-y-4 border-t-2 border-brand-text">
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">ALIAS LIMIT</label>
                    <input v-model.number="addForm.aliases" type="number" min="0" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                    <p class="text-[10px] text-gray-500 mt-1">Maximum number of aliases (0 = unlimited)</p>
                  </div>
                  <div>
                    <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">MAILBOX LIMIT</label>
                    <input v-model.number="addForm.mailboxes" type="number" min="0" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                    <p class="text-[10px] text-gray-500 mt-1">Maximum number of mailboxes (0 = unlimited)</p>
                  </div>
                  <div>
                    <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">QUOTA LIMIT (MB)</label>
                    <input v-model.number="addForm.quotaMB" type="number" min="0" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                    <p class="text-[10px] text-gray-500 mt-1">Maximum quota limit in MB (0 = unlimited)</p>
                  </div>
                  <div>
                    <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">PASSWORD EXPIRY (DAYS)</label>
                    <input v-model.number="addForm.passwordExpiry" type="number" min="0" placeholder="365" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                    <p class="text-[10px] text-gray-500 mt-1">Password expiry in days (empty = never)</p>
                  </div>
                </div>

                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">TRANSPORT</label>
                  <div class="relative">
                    <select v-model="addForm.transport"
                      class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm appearance-none bg-white cursor-pointer pr-10">
                      <option value="virtual">virtual</option>
                      <option v-for="t in transports" :key="t.id" :value="t.transport">{{ t.domain }} → {{ t.transport }}</option>
                    </select>
                    <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-3 text-brand-text border-l-2 border-brand-text">
                      <Icon name="chevron-down" :size="16" />
                    </div>
                  </div>
                  <p class="text-[10px] text-gray-500 mt-1">Select the mail transport for this domain</p>
                </div>

                <div class="flex items-center pt-2">
                  <input type="checkbox" id="dom-add-backupmx" v-model="addForm.backupmx" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                  <label for="dom-add-backupmx" class="ml-2 text-sm font-bold cursor-pointer">Enable Backup MX</label>
                  <span class="ml-auto text-[10px] text-gray-500 hidden sm:block">Use this server as a backup mail exchanger</span>
                </div>
              </div>
            </details>
          </form>
        </div>

        <!-- Footer -->
        <div class="bg-gray-50 px-6 py-4 flex items-center justify-end space-x-3 border-t-2 border-brand-text flex-shrink-0">
          <button type="button" @click="closeAddDomain"
            class="bg-white hover:bg-gray-100 text-brand-text border-2 border-brand-text font-black px-6 py-2.5 shadow-[2px_2px_0px_#1E293B] hover:translate-y-px hover:shadow-[1px_1px_0px_#1E293B] active:translate-y-0.5 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center">
            <Icon name="x" :size="16" class="mr-2" />
            CANCEL
          </button>
          <button type="button" :disabled="savingAdd" @click="submitAdd"
            class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-2.5 shadow-[3px_3px_0px_#1E293B] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center disabled:opacity-60">
            <Icon name="plus-circle" :size="16" class="mr-2" />
            {{ savingAdd ? 'SAVING...' : 'SAVE DOMAIN' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ EDIT DOMAIN MODAL (exact pattern from form_edit_domain.html) ══════════ -->
    <div v-if="showEdit" class="modal-overlay" @click.self="closeEditDomain">
      <div class="bg-white border-4 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Header -->
        <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="edit" :size="20" class="mr-2" />
            EDIT DOMAIN
            <span class="ml-2 text-gray-200 text-base font-mono">- {{ editForm.domain }}</span>
          </h3>
          <button @click="closeEditDomain" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <!-- Scrollable Body -->
        <div class="overflow-y-auto flex-1">
          <!-- Error -->
          <div v-if="editError" class="mx-6 mt-4 bg-red-50 border-2 border-red-600 p-3 flex items-start">
            <Icon name="alert-circle" :size="16" class="text-red-600 mr-2 mt-0.5 flex-shrink-0" />
            <p class="text-sm text-red-700 font-medium">{{ editError }}</p>
          </div>

          <form class="p-6 space-y-5" @submit.prevent="submitEdit">
            <!-- Basic Information -->
            <div class="border-2 border-brand-text p-4 space-y-4">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="info" :size="16" class="mr-2" />
                BASIC INFORMATION
              </h4>

              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">DOMAIN NAME</label>
                <input :value="editForm.domain" readonly
                  class="w-full h-10 px-3 border-2 border-gray-300 bg-gray-100 text-gray-500 cursor-not-allowed font-medium transition-colors text-sm" />
                <p class="text-[10px] text-gray-500 mt-1">Domain name cannot be changed after creation</p>
              </div>

              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">DESCRIPTION</label>
                <input v-model="editForm.description" type="text" placeholder="Optional description for this domain"
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
              </div>

              <div class="flex items-center pt-1">
                <input type="checkbox" id="dom-edit-active" v-model="editForm.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                <label for="dom-edit-active" class="ml-2 text-sm font-bold cursor-pointer">Active Domain</label>
              </div>
            </div>

            <!-- Advanced Settings -->
            <details class="border-2 border-brand-text group" open>
              <summary class="p-3 cursor-pointer font-bold uppercase tracking-tight text-sm flex items-center bg-gray-50 hover:bg-gray-100 transition-colors">
                <Icon name="settings" :size="16" class="mr-2" />
                ADVANCED SETTINGS
                <Icon name="chevron-down" :size="16" class="ml-auto group-open:rotate-180 transition-transform" />
              </summary>
              <div class="p-4 space-y-4 border-t-2 border-brand-text">
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">ALIAS LIMIT</label>
                    <input v-model.number="editForm.aliases" type="number" min="0" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                    <p class="text-[10px] text-gray-500 mt-1">Maximum number of aliases (0 = unlimited)</p>
                  </div>
                  <div>
                    <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">MAILBOX LIMIT</label>
                    <input v-model.number="editForm.mailboxes" type="number" min="0" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                    <p class="text-[10px] text-gray-500 mt-1">Maximum number of mailboxes (0 = unlimited)</p>
                  </div>
                  <div>
                    <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">QUOTA LIMIT (MB)</label>
                    <input v-model.number="editForm.quotaMB" type="number" min="0" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                    <p class="text-[10px] text-gray-500 mt-1">Maximum quota limit in MB (0 = unlimited)</p>
                  </div>
                  <div>
                    <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">PASSWORD EXPIRY (DAYS)</label>
                    <input v-model.number="editForm.passwordExpiry" type="number" min="0" placeholder="365" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                    <p class="text-[10px] text-gray-500 mt-1">Password expiry in days (empty = never)</p>
                  </div>
                </div>

                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">TRANSPORT</label>
                  <div class="relative">
                    <select v-model="editForm.transport"
                      class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm appearance-none bg-white cursor-pointer pr-10">
                      <option value="virtual">virtual</option>
                      <option v-for="t in transports" :key="t.id" :value="t.transport">{{ t.domain }} → {{ t.transport }}</option>
                    </select>
                    <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-3 text-brand-text border-l-2 border-brand-text">
                      <Icon name="chevron-down" :size="16" />
                    </div>
                  </div>
                  <p class="text-[10px] text-gray-500 mt-1">Select the mail transport for this domain</p>
                </div>

                <div class="flex items-center pt-2">
                  <input type="checkbox" id="dom-edit-backupmx" v-model="editForm.backupmx" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                  <label for="dom-edit-backupmx" class="ml-2 text-sm font-bold cursor-pointer">Enable Backup MX</label>
                  <span class="ml-auto text-[10px] text-gray-500 hidden sm:block">Use this server as a backup mail exchanger</span>
                </div>
              </div>
            </details>
          </form>
        </div>

        <!-- Footer -->
        <div class="bg-gray-50 px-6 py-4 flex items-center justify-end space-x-3 border-t-2 border-brand-text flex-shrink-0">
          <button type="button" @click="closeEditDomain"
            class="bg-white hover:bg-gray-100 text-brand-text border-2 border-brand-text font-black px-6 py-2.5 shadow-[2px_2px_0px_#1E293B] hover:translate-y-px hover:shadow-[1px_1px_0px_#1E293B] active:translate-y-0.5 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center">
            <Icon name="x" :size="16" class="mr-2" />
            CANCEL
          </button>
          <button type="button" :disabled="savingEdit" @click="submitEdit"
            class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-2.5 shadow-[3px_3px_0px_#1E293B] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center disabled:opacity-60">
            <Icon name="save" :size="16" class="mr-2" />
            {{ savingEdit ? 'SAVING...' : 'UPDATE DOMAIN' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <ConfirmDialog
      v-model="showDeleteConfirm"
      title="CONFIRM DELETE"
      :loading="deletingRow"
      @confirm="submitDelete"
    >
      <p class="text-sm font-medium text-brand-text leading-relaxed">
        Are you sure you want to delete domain<br />
        <strong class="font-mono break-all">{{ deleteTarget?.domain }}</strong>?<br />
        <span class="text-xs text-red-600 font-bold mt-1 block">This will delete all mailboxes, aliases, and related data.</span>
      </p>
    </ConfirmDialog>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useToastStore } from '../../stores/toast'

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

const columns = [
  { key: 'domain',      label: 'DOMAIN' },
  { key: 'description', label: 'DESCRIPTION' },
  { key: 'aliases',     label: 'ALIASES' },
  { key: 'mailboxes',   label: 'EMAIL ACCOUNTS' },
  { key: 'transport',   label: 'TRANSPORT' },
  { key: 'backupmx',    label: 'BACKUP MX' },
  { key: 'active',      label: 'ACTIVE' },
  { key: 'modified',    label: 'MODIFIED' },
  { key: 'password_expiry', label: 'PASS EXPIRES' },
]

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

function closeAddDomain() {
  showAdd.value = false
  addError.value = ''
}

function closeEditDomain() {
  showEdit.value = false
  editError.value = ''
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
    closeAddDomain()
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
    closeEditDomain()
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

/* .btn-primary, .error-banner, .table-card now centralized in global style.css (Tailwind-friendly) */
.table-card {
  border: 2px solid #1e293b;
  box-shadow: 2px 2px 0 #1e293b;
  padding: 16px 8px 4px;
}

.table-topbar {
  display: flex; justify-content: space-between; align-items: center;
  padding: 0 0 12px; border-bottom: none; gap: 12px; flex-wrap: wrap;
  background: transparent;
}

.table-wrap { overflow-x: auto; }
.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12.5px;
  border: 2px solid #1e293b;
}
.table-head-row {
  background: #3b82f6;
  border-bottom: 2px solid #1e293b;
}
.table-th {
  color: #fff;
  font-weight: 700;
  letter-spacing: 0.4px;
  padding: 10px;
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
  user-select: none;
  font-size: 12px;
  font-family: var(--font-sans);
}
.table-th:hover { background: #2563eb; }
.sort-arrows { margin-left: 4px; font-size: 9px; opacity: .5; }
.sort-active { opacity: 1 !important; }

.table-row {
  border-bottom: 1px solid #e2e8f0;
}
.table-row:last-child {
  border-bottom: none;
}
.table-row:nth-child(even) { background: #f8fafc; }
.table-row:hover { background: #eff6ff; }
.table-td {
  padding: 8px 10px;
  color: #1e293b;
  font-size: 12px;
}
.td-link { color: #1e293b; font-weight: 600; }
.mono { font-family: monospace; font-size: 12px; color: #64748b; }

.progress-bar-wrap {
  width: 128px;
  background-color: #e2e8f0;
  height: 24px;
  position: relative;
  border: 1px solid #1e293b;
  overflow: hidden;
  box-shadow: 1px 1px 0px #1e293b;
}

.progress-bar-fill {
  background-color: #22c55e;
  height: 100%;
  position: absolute;
  top: 0;
  left: 0;
  transition: width 0.3s ease;
}

.progress-bar-text {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  color: #374151;
  z-index: 10;
}

.badge-yes {
  background: oklch(0.962 0.044 156.743);
  color: oklch(0.527 0.154 150.069);
  border: 2px solid oklch(0.527 0.154 150.069);
  padding: 4px 8px;
  font-size: 12px;
  font-weight: 700;
  display: inline-block;
}
.badge-no {
  background: #fee2e2;
  color: #dc2626;
  border: 2px solid #dc2626;
  padding: 4px 8px;
  font-size: 12px;
  font-weight: 700;
  display: inline-block;
}

.actions-td { display: flex; gap: 6px; align-items: center; justify-content: flex-end; }
.act-btn {
  padding: 4px 8px; font-size: 12px; font-weight: 800; cursor: pointer;
  border: 1px solid #1e293b; letter-spacing: 0.4px; border-radius: 0;
  display: inline-flex; align-items: center; transition: all .12s;
  box-shadow: 1px 1px 0 #1e293b; text-transform: uppercase;
}
.act-btn:hover { transform: translate(-0.5px, -0.5px); }
.act-btn:active { transform: translate(0,0); box-shadow: none; }
.act-edit { background: oklch(0.546 0.245 262.881); color: #fff; }
.act-edit:hover { background: #fff; color: oklch(0.546 0.245 262.881); }
.act-del  { background: oklch(0.577 0.245 27.325); color: #fff; }
.act-del:hover { background: #fff; color: oklch(0.577 0.245 27.325); }

.table-loading, .table-empty { text-align: center; padding: 28px; color: #94a3b8; font-size: 13px; }

.table-footer {
  display: flex; justify-content: space-between; align-items: center;
  padding: 12px 0 0; border-top: none;
  background: transparent;
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

/* Migration note: All modals (Add, Edit, Delete) converted to BrutalModal. Old CSS removed. */

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

/* ─── Modal footer styles now handled by BrutalModal + global .btn-cancel / .btn-danger */

.confirm-text { font-size: 14px; color: #374151; line-height: 1.8; text-align: center; margin: 8px 0; }
.confirm-sub { font-size: 12px; color: #dc2626; }
</style>
