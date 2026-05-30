<template>
  <div class="admins-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">ADMINISTRATORS</div>
        <div class="page-subtitle">MANAGE SYSTEM ADMINISTRATORS</div>
      </div>
      <button class="btn-primary" :disabled="!isSuperAdmin" @click="openAdd">
        <Icon name="user-plus" :size="16" style="margin-right:6px;vertical-align:middle" />
        ADD ADMINISTRATOR
      </button>
    </div>

    <!-- ─── Table ─── -->
    <AppTable
      :rows="allAdmins"
      :columns="columns"
      row-key="username"
      :search-fields="['username']"
      default-sort-key="username"
      :loading="loading"
      @edit="openEdit"
      @delete="confirmDelete"
    >
      <template #cell-username="{ row }">
        <div class="cell-with-icon">
          <Icon name="shield" :size="14" class="row-icon" />
          {{ row.username }}
        </div>
      </template>
      <template #cell-domain_count="{ value }">
        <span v-if="value === 'ALL'" class="badge-super">SUPER ADMIN</span>
        <span v-else class="mono-text">{{ value }} domain(s)</span>
      </template>
      <template #cell-active="{ value }">
        <span :class="value ? 'badge-yes' : 'badge-no'">{{ value ? 'YES' : 'NO' }}</span>
      </template>
      <template #cell-modified="{ value }">{{ formatDate(value) }}</template>
      <template #actions="{ row }">
        <button
          class="act-btn act-edit"
          :class="{ 'act-disabled': !canEditAdmin(row) }"
          :disabled="!canEditAdmin(row)"
          @click="openEdit(row)"
        >
          <Icon name="pencil" :size="12" style="margin-right:4px;vertical-align:middle" />EDIT
        </button>
        <button
          class="act-btn act-del"
          :class="{ 'act-disabled': !canDeleteAdmin(row) }"
          :disabled="!canDeleteAdmin(row)"
          @click="confirmDelete(row)"
        >
          <Icon name="trash-2" :size="12" style="margin-right:4px;vertical-align:middle" />DELETE
        </button>
      </template>
    </AppTable>

    <!-- ══════════ ADD MODAL ══════════ -->
    <AdminAddModal
      v-model="showAdd"
      :domains="domains"
      :saving="savingAdd"
      @submit="handleAdd"
    />

    <!-- ══════════ EDIT MODAL ══════════ -->
    <AdminEditModal
      v-model="showEdit"
      :is-super-admin="isSuperAdmin"
      :initial-data="editInitialData"
      :edit-domains="editDomains"
      :saving="savingEdit"
      :loading="loadingEdit"
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
import { useAuthStore } from '../../stores/auth'
import AdminAddModal from './AdminAddModal.vue'
import AdminEditModal from './AdminEditModal.vue'

const toast = useToastStore()
const auth = useAuthStore()
const isSuperAdmin = computed(() => auth.user?.superadmin === true)
const currentUsername = computed(() => auth.user?.username || '')

interface Admin {
  username: string
  active: boolean
  superadmin: boolean
  domain_count: string
  created: string
  modified: string
}

interface Domain { domain: string }
interface DomainOption { domain: string; assigned: boolean }

const allAdmins = ref<Admin[]>([])
const domains = ref<Domain[]>([])
const loading = ref(true)

const columns = [
  { key: 'username',     label: 'USERNAME' },
  { key: 'domain_count', label: 'DOMAINS' },
  { key: 'active',       label: 'ACTIVE' },
  { key: 'modified',     label: 'MODIFIED' },
]

function canEditAdmin(row: Admin): boolean {
  return isSuperAdmin.value || row.username === currentUsername.value
}

function canDeleteAdmin(row: Admin): boolean {
  return isSuperAdmin.value && row.username !== currentUsername.value
}

function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function load() {
  loading.value = true
  try {
    const [admRes, domRes] = await Promise.all([
      http.get(`${API_BASE}/admins`),
      http.get(`${API_BASE}/domains`),
    ])
    allAdmins.value = admRes.data?.data ?? []
    domains.value = domRes.data?.data ?? []
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load administrators')
  } finally { loading.value = false }
}
onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)

function openAdd() {
  if (!isSuperAdmin.value) {
    toast.error('Only superadmins can create administrators')
    return
  }
  showAdd.value = true
}

async function handleAdd(payload: any) {
  if (!isSuperAdmin.value) {
    toast.error('Only superadmins can create administrators')
    showAdd.value = false
    return
  }
  if (!payload.username) { toast.error('Username is required'); return }
  if (!payload.password || payload.password.length < 8) { toast.error('Password must be at least 8 characters'); return }
  savingAdd.value = true
  try {
    await http.post(`${API_BASE}/admins`, payload)
    showAdd.value = false
    toast.success(`Administrator ${payload.username} created successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to create administrator')
  } finally { savingAdd.value = false }
}

// ─── Edit modal ───
const showEdit = ref(false)
const loadingEdit = ref(false)
const savingEdit = ref(false)
const editInitialData = ref<{ username: string; active: boolean; superadmin: boolean } | null>(null)
const editDomains = ref<DomainOption[]>([])
const editUsername = ref('')

async function openEdit(row: Admin) {
  if (!canEditAdmin(row)) {
    toast.error('You cannot edit this administrator')
    return
  }
  editUsername.value = row.username
  editInitialData.value = { username: row.username, active: row.active, superadmin: row.superadmin }
  editDomains.value = []
  loadingEdit.value = true
  showEdit.value = true
  try {
    const res = await http.get(`${API_BASE}/admins/${encodeURIComponent(row.username)}`)
    const d = res.data?.data
    editInitialData.value = { username: row.username, active: d.admin.active, superadmin: d.admin.superadmin }
    editDomains.value = (d.domains ?? [])
      .filter((opt: DomainOption) => isSuperAdmin.value || opt.assigned)
      .map((opt: DomainOption) => ({ ...opt }))
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load administrator')
  } finally { loadingEdit.value = false }
}

async function handleEdit(payload: any) {
  if (payload.change_password) {
    if (!payload.password || payload.password.length < 8) { toast.error('Password must be at least 8 characters'); return }
    if (payload.password !== payload.password_confirm) { toast.error('Passwords do not match'); return }
  }
  savingEdit.value = true
  try {
    await http.put(`${API_BASE}/admins/${encodeURIComponent(editUsername.value)}`, payload)
    showEdit.value = false
    toast.success(`Administrator ${editUsername.value} updated successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to update administrator')
  } finally { savingEdit.value = false }
}

// ─── Delete ───
const showDeleteConfirm = ref(false)
const deletingRow = ref(false)
const deleteTarget = ref<Admin | null>(null)

function confirmDelete(row: Admin) {
  if (!canDeleteAdmin(row)) {
    toast.error(row.username === currentUsername.value
      ? 'You cannot delete your own administrator account'
      : 'You cannot delete this administrator')
    return
  }
  deleteTarget.value = row
  showDeleteConfirm.value = true
}

async function submitDelete() {
  if (!deleteTarget.value || !canDeleteAdmin(deleteTarget.value)) return
  deletingRow.value = true
  try {
    const name = deleteTarget.value.username
    await http.delete(`${API_BASE}/admins/${encodeURIComponent(name)}`)
    showDeleteConfirm.value = false; deleteTarget.value = null
    toast.success(`Administrator ${name} deleted successfully`)
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to delete administrator')
    showDeleteConfirm.value = false
  } finally { deletingRow.value = false }
}
</script>

<style scoped>
.admins-page { background: #ebf2fe; padding: 24px 28px 40px; }
.badge-super { background: #ede9fe; color: #7c3aed; padding: 4px 8px; font-size: 12px; font-weight: 700; border: 2px solid #7c3aed; }
.act-disabled { opacity: .4; cursor: not-allowed; pointer-events: none; }
.btn-primary:disabled { opacity: .55; cursor: not-allowed; transform: none; }
</style>
