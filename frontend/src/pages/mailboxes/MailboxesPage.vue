<template>
  <div class="dom-page">

    <!-- ─── Header ─── -->
    <div class="dom-header">
      <div>
        <div class="dom-title">EMAIL ACCOUNTS</div>
        <div class="dom-subtitle">MANAGE YOUR EMAIL ACCOUNTS</div>
      </div>
      <button class="btn-add-big" @click="showAdd = true">
        <Icon name="plus-circle" :size="20" style="margin-right:12px;vertical-align:middle" />
        ADD EMAIL ACCOUNT
      </button>
    </div>


    <!-- ─── Table ─── -->
    <AppTable
      :rows="domainFilteredRows"
      :columns="columns"
      row-key="username"
      :search-fields="['username', 'name', 'domain']"
      default-sort-key="username"
      :loading="loading"
      @edit="openEdit"
      @delete="confirmDelete"
    >
      <template #toolbar>
        <div class="mb-4 flex items-center gap-3">
          <label class="text-xs font-black uppercase tracking-widest text-brand-text">FILTER BY DOMAIN:</label>
          <select v-model="domainFilter" class="h-9 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium bg-white">
            <option value="">All Domains</option>
            <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
          </select>
          <a v-if="domainFilter" class="ml-2 text-xs font-bold text-red-500 hover:underline cursor-pointer" @click="domainFilter = ''">Clear Filter</a>
        </div>
      </template>

      <template #cell-username="{ row }">
        <div class="cell-with-icon">
          <Icon name="mail" :size="14" class="row-icon" />
          {{ row.username }}
        </div>
      </template>
      <template #cell-name="{ value }">{{ value || '—' }}</template>
      <template #cell-domain="{ value }"><span class="mono">{{ value }}</span></template>
      <template #cell-quota="{ value }">{{ value === 0 ? 'Unlimited' : (value / 1048576) + ' MB' }}</template>
      <template #cell-active="{ value }">
        <span :class="value ? 'badge-yes' : 'badge-no'">{{ value ? 'YES' : 'NO' }}</span>
      </template>
      <template #cell-modified="{ value }">{{ formatDate(value) }}</template>
    </AppTable>

    <!-- ══════════ ADD MODAL ══════════ -->
    <MailboxAddModal
      v-model="showAdd"
      :domains="domains"
      :saving="savingAdd"
      @submit="handleAdd"
    />

    <!-- ══════════ EDIT MODAL ══════════ -->
    <MailboxEditModal
      v-model="showEdit"
      :initial-data="editInitialData"
      :saving="savingEdit"
      @submit="handleEdit"
    />

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <ConfirmDialog
      v-model="showDeleteConfirm"
      title="CONFIRM DELETE"
      :item-name="deleteTarget?.username"
      :loading="deletingRow"
      @confirm="submitDelete"
    />

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import http from '../../utils/http'
import { useToastStore } from '../../stores/toast'
import MailboxAddModal from './MailboxAddModal.vue'
import MailboxEditModal from './MailboxEditModal.vue'

const toast = useToastStore()

interface Mailbox {
  username: string
  domain: string
  name: string
  quota: number
  active: boolean
  smtp_active: boolean
  email_other: string
  modified: string
  created: string
}

interface Domain { domain: string }

const allMailboxes = ref<Mailbox[]>([])
const domains = ref<Domain[]>([])
const loading = ref(true)
const domainFilter = ref('')

const columns = [
  { key: 'username', label: 'EMAIL' },
  { key: 'name',     label: 'NAME' },
  { key: 'domain',   label: 'DOMAIN' },
  { key: 'quota',    label: 'QUOTA' },
  { key: 'active',   label: 'ACTIVE' },
  { key: 'modified', label: 'MODIFIED' },
]

const domainFilteredRows = computed(() => {
  if (!domainFilter.value) return allMailboxes.value
  return allMailboxes.value.filter(r => r.domain === domainFilter.value)
})

function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function load() {
  loading.value = true
  try {
    const [mbRes, domRes] = await Promise.all([
      http.get(`${API_BASE}/mailboxes`),
      http.get(`${API_BASE}/domains`),
    ])
    allMailboxes.value = mbRes.data?.data ?? []
    domains.value = domRes.data?.data ?? []
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load mailboxes')
  } finally { loading.value = false }
}
onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)

async function handleAdd(payload: any) {
  if (!payload.local_part || payload.local_part.length < 4) { toast.error('Username must be at least 4 characters'); return }
  if (!payload.domain) { toast.error('Please select a domain'); return }
  if (!payload.password || payload.password.length < 8) { toast.error('Password must be at least 8 characters'); return }
  savingAdd.value = true
  try {
    await http.post(`${API_BASE}/mailboxes`, payload)
    showAdd.value = false
    toast.success(`Mailbox ${payload.local_part}@${payload.domain} created successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to create mailbox')
  } finally { savingAdd.value = false }
}

// ─── Edit modal ───
const showEdit = ref(false)
const savingEdit = ref(false)
const editInitialData = ref<any>(null)
const editUsername = ref('')

async function openEdit(row: Mailbox) {
  try {
    const res = await http.get(`${API_BASE}/mailboxes/${encodeURIComponent(row.username)}`)
    const mb = res.data?.data
    editUsername.value = mb.username
    editInitialData.value = {
      username: mb.username,
      name: mb.name || '',
      quotaMB: mb.quota,
      active: mb.active,
      smtpActive: mb.smtp_active,
      emailOther: mb.email_other || '',
      password: '',
      passwordConfirm: '',
    }
    showEdit.value = true
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load mailbox')
  }
}

async function handleEdit(payload: any) {
  if (payload.change_password) {
    if (!payload.password || payload.password.length < 8) { toast.error('Password must be at least 8 characters'); return }
    if (payload.password !== payload.password_confirm) { toast.error('Passwords do not match'); return }
  }
  savingEdit.value = true
  try {
    await http.put(`${API_BASE}/mailboxes/${encodeURIComponent(editUsername.value)}`, payload)
    showEdit.value = false
    toast.success(`Mailbox ${editUsername.value} updated successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to update mailbox')
  } finally { savingEdit.value = false }
}

// ─── Delete ───
const showDeleteConfirm = ref(false)
const deletingRow = ref(false)
const deleteTarget = ref<Mailbox | null>(null)

function confirmDelete(row: Mailbox) { deleteTarget.value = row; showDeleteConfirm.value = true }

async function submitDelete() {
  if (!deleteTarget.value) return
  deletingRow.value = true
  try {
    const username = deleteTarget.value.username
    await http.delete(`${API_BASE}/mailboxes/${encodeURIComponent(username)}`)
    showDeleteConfirm.value = false; deleteTarget.value = null
    toast.success(`Mailbox ${username} deleted successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to delete mailbox')
    showDeleteConfirm.value = false
  } finally { deletingRow.value = false }
}
</script>

<style scoped>
.dom-page { background: #ebf2fe; padding: 24px 28px 40px; }
</style>
