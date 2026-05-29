<template>
  <div class="admins-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">ADMINISTRATORS</div>
        <div class="page-subtitle">MANAGE SYSTEM ADMINISTRATORS</div>
      </div>
      <button v-if="isSuperAdmin" class="btn-primary" @click="openAdd">
        <Icon name="user-plus" :size="16" style="margin-right:6px;vertical-align:middle" />
        ADD ADMINISTRATOR
      </button>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <Icon name="alert-triangle" :size="16" class="mr-1" /> {{ error }}
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
        <button class="act-btn act-edit" @click="openEdit(row)">
          <Icon name="pencil" :size="12" style="margin-right:4px;vertical-align:middle" />EDIT
        </button>
        <button
          class="act-btn act-del"
          :class="{ 'act-disabled': !isSuperAdmin }"
          :disabled="!isSuperAdmin"
          @click="isSuperAdmin && confirmDelete(row)"
        >
          <Icon name="trash-2" :size="12" style="margin-right:4px;vertical-align:middle" />DELETE
        </button>
      </template>
    </AppTable>

    <!-- ══════════ ADD MODAL ══════════ -->
    <div v-if="showAdd" class="modal-overlay" @click.self="showAdd = false">
      <div class="modal-card">
        <div class="modal-head">
          <span class="modal-head-title">
            <Icon name="user-plus" :size="18" style="margin-right:8px;vertical-align:middle" />
            ADD NEW ADMINISTRATOR
          </span>
          <button class="modal-close" @click="showAdd = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="addError" class="modal-error">
            <Icon name="alert-triangle" :size="14" style="margin-right:6px;flex-shrink:0" />{{ addError }}
          </div>

          <!-- Admin Details -->
          <div class="info-card">
            <div class="info-card-title">
              <Icon name="user" :size="15" style="margin-right:6px" />
              ADMIN DETAILS
            </div>
            <div class="form-group">
              <label class="form-label">USERNAME (EMAIL) <span class="req">*</span></label>
              <input
                v-model="addForm.username"
                type="email"
                class="form-input"
                placeholder="admin@example.com"
                @input="addForm.username = addForm.username.toLowerCase()"
              />
            </div>
            <div class="check-row">
              <label class="check-label">
                <input type="checkbox" v-model="addForm.active" />
                Active
              </label>
              <label class="check-label" :title="'Grants full access to all domains and settings'">
                <input type="checkbox" v-model="addForm.superadmin" @change="onAddSuperAdminChange" />
                Super Administrator
              </label>
            </div>
          </div>

          <!-- Password -->
          <div class="info-card">
            <div class="info-card-title">
              <Icon name="key" :size="15" style="margin-right:6px" />
              PASSWORD
            </div>
            <div class="pw-row">
              <div class="pw-field">
                <label class="form-label">PASSWORD <span class="req">*</span></label>
                <div class="pw-wrap">
                  <input
                    v-model="addForm.password"
                    :type="showPw1 ? 'text' : 'password'"
                    class="form-input"
                    placeholder="Min. 8 characters"
                    @input="onAddPasswordInput"
                  />
                  <button type="button" class="pw-eye" @click="showPw1 = !showPw1">
                    <Icon :name="showPw1 ? 'eye-off' : 'eye'" :size="18" />
                  </button>
                </div>
              </div>
              <div class="pw-field">
                <label class="form-label">CONFIRM PASSWORD <span class="req">*</span></label>
                <div class="pw-wrap">
                  <input
                    v-model="addForm.passwordConfirm"
                    :type="showPw2 ? 'text' : 'password'"
                    class="form-input"
                    placeholder="Repeat password"
                    @input="onAddPasswordInput"
                  />
                  <button type="button" class="pw-eye" @click="showPw2 = !showPw2">
                    <Icon :name="showPw2 ? 'eye-off' : 'eye'" :size="18" />
                  </button>
                </div>
              </div>
              <div class="pw-gen-wrap">
                <label class="form-label">&nbsp;</label>
                <button type="button" class="btn-generate" @click="generatePassword(addForm)">
                  <Icon name="wand-2" :size="14" style="margin-right:4px;vertical-align:middle" />
                  GENERATE
                </button>
              </div>
            </div>
            <div v-if="addForm.password" class="pwd-strength">
              <div class="strength-bar-wrap">
                <div class="strength-bar" :style="{ width: addPwdStrength.pct + '%', background: addPwdStrength.color }"></div>
              </div>
              <span class="strength-label" :style="{ color: addPwdStrength.color }">{{ addPwdStrength.label }}</span>
            </div>
            <div v-if="addPwdMismatch" class="pwd-mismatch">Passwords do not match</div>
          </div>

          <!-- Domains -->
          <div class="info-card" :class="{ 'card-disabled': addForm.superadmin }">
            <div class="info-card-title">
              <Icon name="globe" :size="15" style="margin-right:6px" />
              ASSIGNED DOMAINS
              <span class="info-card-sub">Select domains this admin can manage</span>
            </div>
            <div v-if="addForm.superadmin" class="superadmin-note">
              Super admins have access to all domains automatically.
            </div>
            <div v-else-if="domains.length === 0" class="domains-empty">No domains available</div>
            <div v-else class="domains-grid">
              <label
                v-for="d in domains"
                :key="d.domain"
                class="domain-item"
                :class="{ 'domain-item-checked': addForm.domains.includes(d.domain) }"
              >
                <input
                  type="checkbox"
                  :value="d.domain"
                  :checked="addForm.domains.includes(d.domain)"
                  :disabled="addForm.superadmin"
                  @change="toggleAddDomain(d.domain)"
                />
                <span class="domain-name" :title="d.domain">{{ d.domain }}</span>
              </label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showAdd = false">
            <Icon name="x" :size="14" style="margin-right:4px;vertical-align:middle" />CANCEL
          </button>
          <button class="btn-primary" :disabled="savingAdd" @click="submitAdd">
            <Icon name="save" :size="14" style="margin-right:6px;vertical-align:middle" />
            {{ savingAdd ? 'SAVING...' : 'CREATE ADMIN' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ EDIT MODAL ══════════ -->
    <div v-if="showEdit" class="modal-overlay" @click.self="showEdit = false">
      <div class="modal-card">
        <div class="modal-head">
          <span class="modal-head-title">
            <Icon name="users" :size="18" style="margin-right:8px;vertical-align:middle" />
            EDIT ADMINISTRATOR
            <span class="modal-head-sub">— {{ editForm.username }}</span>
          </span>
          <button class="modal-close" @click="showEdit = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingEdit" class="loading-overlay">
            <div class="spinner mx-auto" style="width:32px;height:32px" />
            <span class="loading-text">LOADING...</span>
          </div>
          <template v-else>
            <div v-if="editError" class="modal-error">
              <Icon name="alert-triangle" :size="14" style="margin-right:6px;flex-shrink:0" />{{ editError }}
            </div>

            <!-- Admin Details -->
            <div class="info-card">
              <div class="info-card-title">
                <Icon name="user" :size="15" style="margin-right:6px" />
                ADMIN DETAILS
              </div>
              <div class="form-group">
                <label class="form-label">USERNAME</label>
                <input :value="editForm.username" class="form-input form-input-disabled" disabled />
                <span class="form-hint">Username cannot be changed after creation</span>
              </div>
              <div class="check-row">
                <label class="check-label" :class="{ 'check-disabled': !isSuperAdmin }">
                  <input type="checkbox" v-model="editForm.active" :disabled="!isSuperAdmin" />
                  Active
                </label>
                <label class="check-label" :class="{ 'check-disabled': !isSuperAdmin }">
                  <input type="checkbox" v-model="editForm.superadmin" :disabled="!isSuperAdmin" @change="onEditSuperAdminChange" />
                  Super Administrator
                </label>
              </div>
            </div>

            <!-- Change Password (collapsible) -->
            <details class="adv-details">
              <summary class="adv-summary">
                <span class="adv-summary-left">
                  <Icon name="key" :size="15" style="margin-right:6px" />
                  CHANGE PASSWORD
                </span>
                <Icon name="chevron-down" :size="18" class="adv-chevron" />
              </summary>
              <div class="adv-body">
                <div class="pw-row">
                  <div class="pw-field">
                    <label class="form-label">NEW PASSWORD</label>
                    <div class="pw-wrap">
                      <input
                        v-model="editForm.password"
                        :type="showPw3 ? 'text' : 'password'"
                        class="form-input"
                        placeholder="Min. 8 characters"
                        @input="onEditPasswordInput"
                      />
                      <button type="button" class="pw-eye" @click="showPw3 = !showPw3">
                        <Icon :name="showPw3 ? 'eye-off' : 'eye'" :size="18" />
                      </button>
                    </div>
                  </div>
                  <div class="pw-field">
                    <label class="form-label">CONFIRM NEW PASSWORD</label>
                    <div class="pw-wrap">
                      <input
                        v-model="editForm.passwordConfirm"
                        :type="showPw4 ? 'text' : 'password'"
                        class="form-input"
                        placeholder="Repeat new password"
                        @input="onEditPasswordInput"
                      />
                      <button type="button" class="pw-eye" @click="showPw4 = !showPw4">
                        <Icon :name="showPw4 ? 'eye-off' : 'eye'" :size="18" />
                      </button>
                    </div>
                  </div>
                  <div class="pw-gen-wrap">
                    <label class="form-label">&nbsp;</label>
                    <button type="button" class="btn-generate" @click="generatePassword(editForm)">
                      <Icon name="wand-2" :size="14" style="margin-right:4px;vertical-align:middle" />
                      GENERATE
                    </button>
                  </div>
                </div>
                <div v-if="editForm.password" class="pwd-strength">
                  <div class="strength-bar-wrap">
                    <div class="strength-bar" :style="{ width: editPwdStrength.pct + '%', background: editPwdStrength.color }"></div>
                  </div>
                  <span class="strength-label" :style="{ color: editPwdStrength.color }">{{ editPwdStrength.label }}</span>
                </div>
                <div v-if="editPwdMismatch" class="pwd-mismatch">Passwords do not match</div>
              </div>
            </details>

            <!-- Domains -->
            <div class="info-card" :class="{ 'card-disabled': editForm.superadmin || !isSuperAdmin }">
              <div class="info-card-title">
                <Icon name="globe" :size="15" style="margin-right:6px" />
                ASSIGNED DOMAINS
                <span class="info-card-sub">Select domains this admin can manage</span>
              </div>
              <div v-if="editForm.superadmin" class="superadmin-note">
                Super admins have access to all domains automatically.
              </div>
              <div v-else-if="editDomains.length === 0" class="domains-empty">No domains available</div>
              <div v-else class="domains-grid">
                <label
                  v-for="d in editDomains"
                  :key="d.domain"
                  class="domain-item"
                  :class="{ 'domain-item-checked': d.assigned }"
                >
                  <input
                    type="checkbox"
                    v-model="d.assigned"
                    :disabled="editForm.superadmin || !isSuperAdmin"
                  />
                  <span class="domain-name" :title="d.domain">{{ d.domain }}</span>
                </label>
              </div>
            </div>
          </template>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showEdit = false">
            <Icon name="x" :size="14" style="margin-right:4px;vertical-align:middle" />CANCEL
          </button>
          <button class="btn-primary" :disabled="savingEdit || loadingEdit" @click="submitEdit">
            <Icon name="save" :size="14" style="margin-right:6px;vertical-align:middle" />
            {{ savingEdit ? 'SAVING...' : 'SAVE CHANGES' }}
          </button>
        </div>
      </div>
    </div>

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
import axios from 'axios'
import { useToastStore } from '../stores/toast'
import { useAuthStore } from '../stores/auth'
import { calcStrength, generatePassword as fillPassword } from '../utils/password'

const toast = useToastStore()
const auth = useAuthStore()

const isSuperAdmin = computed(() => auth.user?.superadmin === true)

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
const error = ref('')

const columns = [
  { key: 'username',     label: 'USERNAME' },
  { key: 'domain_count', label: 'DOMAINS' },
  { key: 'active',       label: 'ACTIVE' },
  { key: 'modified',     label: 'MODIFIED' },
]

function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function load() {
  loading.value = true; error.value = ''
  try {
    const [admRes, domRes] = await Promise.all([
      axios.get(`${API_BASE}/admins`),
      axios.get(`${API_BASE}/domains`),
    ])
    allAdmins.value = admRes.data?.data ?? []
    domains.value = domRes.data?.data ?? []
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load administrators'
  } finally { loading.value = false }
}
onMounted(load)

function generatePassword(form: { password: string; passwordConfirm: string; [k: string]: any }) {
  fillPassword(form)
  if (form === addForm.value) {
    showPw1.value = true
    showPw2.value = true
    onAddPasswordInput()
  } else {
    showPw3.value = true
    showPw4.value = true
    onEditPasswordInput()
  }
}

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)
const addError = ref('')
const showPw1 = ref(false)
const showPw2 = ref(false)
const addPwdMismatch = ref(false)
const addForm = ref({
  username: '', password: '', passwordConfirm: '',
  active: true, superadmin: false, domains: [] as string[]
})

const addPwdStrength = computed(() => calcStrength(addForm.value.password))

function onAddPasswordInput() {
  addPwdMismatch.value = !!(addForm.value.passwordConfirm && addForm.value.password !== addForm.value.passwordConfirm)
}

function onAddSuperAdminChange() {
  if (addForm.value.superadmin) addForm.value.domains = []
}

function toggleAddDomain(domain: string) {
  const idx = addForm.value.domains.indexOf(domain)
  if (idx === -1) addForm.value.domains.push(domain)
  else addForm.value.domains.splice(idx, 1)
}

function openAdd() {
  addForm.value = { username: '', password: '', passwordConfirm: '', active: true, superadmin: false, domains: [] }
  addError.value = ''
  showPw1.value = false
  showPw2.value = false
  addPwdMismatch.value = false
  showAdd.value = true
}

async function submitAdd() {
  addError.value = ''
  const f = addForm.value
  if (!f.username) { addError.value = 'Username is required'; return }
  if (!f.password) { addError.value = 'Password is required'; return }
  if (f.password.length < 8) { addError.value = 'Password must be at least 8 characters'; return }
  if (f.password !== f.passwordConfirm) { addError.value = 'Passwords do not match'; return }
  savingAdd.value = true
  try {
    await axios.post(`${API_BASE}/admins`, {
      username: f.username.toLowerCase().trim(),
      password: f.password,
      active: f.active,
      superadmin: f.superadmin,
      domains: f.superadmin ? [] : f.domains,
    })
    showAdd.value = false
    toast.success(`Administrator ${f.username} created successfully`)
    await load()
  } catch (e: any) {
    addError.value = e?.response?.data?.error?.message || 'Failed to create administrator'
    toast.error(addError.value)
  } finally { savingAdd.value = false }
}

// ─── Edit modal ───
const showEdit = ref(false)
const loadingEdit = ref(false)
const savingEdit = ref(false)
const editError = ref('')
const showPw3 = ref(false)
const showPw4 = ref(false)
const editPwdMismatch = ref(false)
const editDomains = ref<DomainOption[]>([])
const editForm = ref({
  username: '', password: '', passwordConfirm: '',
  active: true, superadmin: false,
})

const editPwdStrength = computed(() => calcStrength(editForm.value.password))

function onEditPasswordInput() {
  editPwdMismatch.value = !!(editForm.value.passwordConfirm && editForm.value.password !== editForm.value.passwordConfirm)
}

function onEditSuperAdminChange() {
  if (editForm.value.superadmin) editDomains.value.forEach(d => d.assigned = true)
}

async function openEdit(row: Admin) {
  editError.value = ''
  editForm.value = { username: row.username, password: '', passwordConfirm: '', active: row.active, superadmin: row.superadmin }
  editDomains.value = []
  showPw3.value = false
  showPw4.value = false
  editPwdMismatch.value = false
  loadingEdit.value = true
  showEdit.value = true
  try {
    const res = await axios.get(`${API_BASE}/admins/${encodeURIComponent(row.username)}`)
    const d = res.data?.data
    editForm.value.active = d.admin.active
    editForm.value.superadmin = d.admin.superadmin
    editDomains.value = (d.domains ?? []).map((opt: DomainOption) => ({ ...opt }))
  } catch (e: any) {
    editError.value = e?.response?.data?.error?.message || 'Failed to load administrator'
  } finally { loadingEdit.value = false }
}

async function submitEdit() {
  editError.value = ''
  const f = editForm.value
  if (f.password && f.password.length < 8) { editError.value = 'Password must be at least 8 characters'; return }
  if (f.password && f.password !== f.passwordConfirm) { editError.value = 'Passwords do not match'; return }
  savingEdit.value = true
  try {
    const payload: any = {
      active: f.active,
      superadmin: f.superadmin,
      domains: f.superadmin ? [] : editDomains.value.filter(d => d.assigned).map(d => d.domain),
    }
    if (f.password) {
      payload.password = f.password
      payload.password_confirm = f.passwordConfirm
      payload.change_password = true
    }
    await axios.put(`${API_BASE}/admins/${encodeURIComponent(f.username)}`, payload)
    showEdit.value = false
    toast.success(`Administrator ${f.username} updated successfully`)
    await load()
  } catch (e: any) {
    editError.value = e?.response?.data?.error?.message || 'Failed to update administrator'
    toast.error(editError.value)
  } finally { savingEdit.value = false }
}

// ─── Delete ───
const showDeleteConfirm = ref(false)
const deletingRow = ref(false)
const deleteTarget = ref<Admin | null>(null)

function confirmDelete(row: Admin) { deleteTarget.value = row; showDeleteConfirm.value = true }

async function submitDelete() {
  if (!deleteTarget.value) return
  deletingRow.value = true
  try {
    const name = deleteTarget.value.username
    await axios.delete(`${API_BASE}/admins/${encodeURIComponent(name)}`)
    showDeleteConfirm.value = false; deleteTarget.value = null
    toast.success(`Administrator ${name} deleted successfully`)
    await load()
  } catch (e: any) {
    const msg = e?.response?.data?.error?.message || 'Failed to delete administrator'
    error.value = msg
    toast.error(msg)
    showDeleteConfirm.value = false
  } finally { deletingRow.value = false }
}
</script>

<style scoped>
.admins-page { background: #ebf2fe; padding: 24px 28px 40px; }

/* Duplicated styles removed — now centralized in global style.css (Tailwind v4 + Brutalist primitives) */

/* More duplicated table styles removed - centralized globally */
.badge-super { background: #ede9fe; color: #7c3aed; padding: 4px 8px; font-size: 12px; font-weight: 700; border: 2px solid #7c3aed; }
.act-disabled { opacity: .4; cursor: not-allowed; pointer-events: none; }
/* More duplicated pagination/table-footer styles removed - centralized globally */

/* Migration note: Delete modal converted to BrutalModal. Add/Edit modals still use legacy custom structure. Old wrapper CSS reduced. */

/* ─── Info card ─── */
.info-card { border: 2px solid #1e293b; padding: 16px; display: flex; flex-direction: column; gap: 14px; }
.info-card-title { font-size: 12px; font-weight: 900; color: #1e293b; letter-spacing: 0.6px; text-transform: uppercase; font-family: monospace; display: flex; align-items: center; gap: 4px; }
.info-card-sub { font-size: 10px; color: #94a3b8; font-weight: 500; letter-spacing: 0.5px; margin-left: 8px; text-transform: none; font-family: inherit; }
.card-disabled { opacity: .5; pointer-events: none; }

/* Heavy form/check styles deduplicated — now in global style.css */
.check-disabled input { cursor: not-allowed; }

/* ─── Password ─── */
.pw-row { display: flex; gap: 12px; align-items: flex-end; }
.pw-field { flex: 1; display: flex; flex-direction: column; gap: 3px; }
.pw-wrap { position: relative; }
.pw-wrap .form-input { padding-right: 40px; }
.pw-eye { position: absolute; right: 10px; top: 50%; transform: translateY(-50%); background: none; border: none; cursor: pointer; color: #64748b; padding: 0; display: flex; }
.pw-eye:hover { color: #1e293b; }
.pw-gen-wrap { display: flex; flex-direction: column; gap: 3px; flex-shrink: 0; }
.btn-generate { height: 40px; padding: 0 12px; background: #3b82f6; color: #fff; border: 2px solid #1e293b; cursor: pointer; display: flex; align-items: center; font-size: 11px; font-weight: 800; letter-spacing: 0.5px; transition: all .12s; box-shadow: 2px 2px 0 #1e293b; flex-shrink: 0; }
.btn-generate:hover { background: #fff; color: #3b82f6; }
.pwd-strength { display: flex; align-items: center; gap: 10px; margin-top: 4px; }
.strength-bar-wrap { flex: 1; height: 5px; background: #e2e8f0; border: 1px solid #d1d5db; overflow: hidden; }
.strength-bar { height: 100%; transition: width .3s, background .3s; }
.strength-label { font-size: 10px; font-weight: 800; letter-spacing: 0.5px; text-transform: uppercase; white-space: nowrap; }
.pwd-mismatch { font-size: 11px; color: #dc2626; font-weight: 600; }

/* Heavy deduplication — advanced/domain grid styles moved to global or simplified */</style>
