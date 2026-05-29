<template>
  <q-page class="admins-page">

    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">ADMINISTRATORS</div>
        <div class="page-subtitle">MANAGE SYSTEM ADMINISTRATORS</div>
      </div>
      <button v-if="isSuperAdmin" class="btn-primary" @click="openAdd">
        <q-icon name="person_add" size="16px" style="margin-right:6px;vertical-align:middle" />
        ADD ADMINISTRATOR
      </button>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <q-icon name="warning" size="16px" /> {{ error }}
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
            <tr v-for="row in pagedRows" :key="row.username" class="table-row">
              <td class="table-td td-link">
                <div class="cell-with-icon">
                  <q-icon name="shield" size="14px" class="row-icon" />
                  {{ row.username }}
                </div>
              </td>
              <td class="table-td">
                <span v-if="row.domain_count === 'ALL'" class="badge-super">SUPER ADMIN</span>
                <span v-else class="mono-text">{{ row.domain_count }} domain(s)</span>
              </td>
              <td class="table-td">
                <span :class="row.active ? 'badge-yes' : 'badge-no'">{{ row.active ? 'YES' : 'NO' }}</span>
              </td>
              <td class="table-td">{{ formatDate(row.modified) }}</td>
              <td class="table-td actions-td">
                <button class="act-btn act-edit" @click="openEdit(row)">
                  <q-icon name="edit" size="12px" style="margin-right:4px;vertical-align:middle" />EDIT
                </button>
                <button
                  class="act-btn act-del"
                  :class="{ 'act-disabled': !isSuperAdmin }"
                  :disabled="!isSuperAdmin"
                  @click="isSuperAdmin && confirmDelete(row)"
                >
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
            <q-icon name="person_add" size="18px" style="margin-right:8px;vertical-align:middle" />
            ADD NEW ADMINISTRATOR
          </span>
          <button class="modal-close" @click="showAdd = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="addError" class="modal-error">
            <q-icon name="warning" size="14px" style="margin-right:6px;flex-shrink:0" />{{ addError }}
          </div>

          <!-- Admin Details -->
          <div class="info-card">
            <div class="info-card-title">
              <q-icon name="person" size="15px" style="margin-right:6px" />
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
              <q-icon name="key" size="15px" style="margin-right:6px" />
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
                    <q-icon :name="showPw1 ? 'visibility_off' : 'visibility'" size="18px" />
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
                    <q-icon :name="showPw2 ? 'visibility_off' : 'visibility'" size="18px" />
                  </button>
                </div>
              </div>
              <div class="pw-gen-wrap">
                <label class="form-label">&nbsp;</label>
                <button type="button" class="btn-generate" @click="generatePassword(addForm)">
                  <q-icon name="auto_fix_high" size="14px" style="margin-right:4px;vertical-align:middle" />
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
              <q-icon name="public" size="15px" style="margin-right:6px" />
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
            <q-icon name="close" size="14px" style="margin-right:4px;vertical-align:middle" />CANCEL
          </button>
          <button class="btn-primary" :disabled="savingAdd" @click="submitAdd">
            <q-icon name="save" size="14px" style="margin-right:6px;vertical-align:middle" />
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
            <q-icon name="manage_accounts" size="18px" style="margin-right:8px;vertical-align:middle" />
            EDIT ADMINISTRATOR
            <span class="modal-head-sub">— {{ editForm.username }}</span>
          </span>
          <button class="modal-close" @click="showEdit = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingEdit" class="loading-overlay">
            <q-spinner color="primary" size="32px" />
            <span class="loading-text">LOADING...</span>
          </div>
          <template v-else>
            <div v-if="editError" class="modal-error">
              <q-icon name="warning" size="14px" style="margin-right:6px;flex-shrink:0" />{{ editError }}
            </div>

            <!-- Admin Details -->
            <div class="info-card">
              <div class="info-card-title">
                <q-icon name="person" size="15px" style="margin-right:6px" />
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
                  <q-icon name="key" size="15px" style="margin-right:6px" />
                  CHANGE PASSWORD
                </span>
                <q-icon name="expand_more" size="18px" class="adv-chevron" />
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
                        <q-icon :name="showPw3 ? 'visibility_off' : 'visibility'" size="18px" />
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
                        <q-icon :name="showPw4 ? 'visibility_off' : 'visibility'" size="18px" />
                      </button>
                    </div>
                  </div>
                  <div class="pw-gen-wrap">
                    <label class="form-label">&nbsp;</label>
                    <button type="button" class="btn-generate" @click="generatePassword(editForm)">
                      <q-icon name="auto_fix_high" size="14px" style="margin-right:4px;vertical-align:middle" />
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
                <q-icon name="public" size="15px" style="margin-right:6px" />
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
            <q-icon name="close" size="14px" style="margin-right:4px;vertical-align:middle" />CANCEL
          </button>
          <button class="btn-primary" :disabled="savingEdit || loadingEdit" @click="submitEdit">
            <q-icon name="save" size="14px" style="margin-right:6px;vertical-align:middle" />
            {{ savingEdit ? 'SAVING...' : 'SAVE CHANGES' }}
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
            Are you sure you want to delete administrator<br />
            <strong>{{ deleteTarget?.username }}</strong>?<br />
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

const search = ref('')
const rowsPerPage = ref(15)
const currentPage = ref(1)
const sortKey = ref('username')
const sortDir = ref<'asc' | 'desc'>('asc')

const columns = [
  { key: 'username',     label: 'USERNAME' },
  { key: 'domain_count', label: 'DOMAINS' },
  { key: 'active',       label: 'ACTIVE' },
  { key: 'modified',     label: 'MODIFIED' },
]

function sortBy(key: string) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

const filteredRows = computed(() => {
  const q = search.value.toLowerCase()
  let rows = allAdmins.value
  if (q) rows = rows.filter(r =>
    r.username.toLowerCase().includes(q)
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

.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 20px; }
.page-title { font-size: 28px; font-weight: 900; color: #1e293b; letter-spacing: -0.5px; line-height: 1; font-family: monospace; text-transform: uppercase; }
.page-subtitle { font-size: 10px; color: #94a3b8; letter-spacing: 0.8px; margin-top: 6px; text-transform: uppercase; font-weight: 700; }

.btn-primary {
  background: #3b82f6; color: #fff; border: 2px solid #1e293b; padding: 14px 24px;
  font-size: 11px; font-weight: 800; letter-spacing: 0.6px; cursor: pointer;
  border-radius: 0; transition: all .15s; text-transform: uppercase;
  box-shadow: 3px 3px 0 #1e293b; display: flex; align-items: center;
}
.btn-primary:hover:not(:disabled) { background: #fff; color: #3b82f6; transform: translate(-1px,-1px); box-shadow: 4px 4px 0 #1e293b; }
.btn-primary:active:not(:disabled) { transform: translate(0,0); box-shadow: none; }
.btn-primary:disabled { opacity: .5; cursor: default; }

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
.mono-text { font-family: monospace; font-size: 12px; color: #6b7280; }
.cell-with-icon { display: flex; align-items: center; gap: 6px; }
.row-icon { color: #3b82f6; flex-shrink: 0; }
.badge-yes   { background: #dcfce7; color: #16a34a; padding: 2px 8px; font-size: 11px; font-weight: 700; }
.badge-no    { background: #fee2e2; color: #dc2626; padding: 2px 8px; font-size: 11px; font-weight: 700; }
.badge-super { background: #ede9fe; color: #7c3aed; padding: 2px 8px; font-size: 11px; font-weight: 700; border: 1px solid #c4b5fd; }
.actions-td { display: flex; gap: 6px; align-items: center; }
.act-btn { padding: 4px 10px; font-size: 10px; font-weight: 800; cursor: pointer; border: 1px solid #1e293b; letter-spacing: 0.4px; border-radius: 0; display: inline-flex; align-items: center; transition: all .12s; box-shadow: 1px 1px 0 #1e293b; text-transform: uppercase; }
.act-btn:hover { transform: translate(-0.5px,-0.5px); box-shadow: 2px 2px 0 #1e293b; }
.act-btn:active { transform: translate(0,0); box-shadow: none; }
.act-edit { background: #3b82f6; color: #fff; }
.act-edit:hover { background: #fff; color: #3b82f6; }
.act-del  { background: #ef4444; color: #fff; }
.act-del:hover { background: #fff; color: #ef4444; }
.act-disabled { opacity: .4; cursor: not-allowed; pointer-events: none; }
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

.loading-overlay { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 40px; gap: 12px; }
.loading-text { font-size: 11px; font-weight: 800; letter-spacing: 1px; color: #94a3b8; text-transform: uppercase; }

/* ─── Info card ─── */
.info-card { border: 2px solid #1e293b; padding: 16px; display: flex; flex-direction: column; gap: 14px; }
.info-card-title { font-size: 12px; font-weight: 900; color: #1e293b; letter-spacing: 0.6px; text-transform: uppercase; font-family: monospace; display: flex; align-items: center; gap: 4px; }
.info-card-sub { font-size: 10px; color: #94a3b8; font-weight: 500; letter-spacing: 0.5px; margin-left: 8px; text-transform: none; font-family: inherit; }
.card-disabled { opacity: .5; pointer-events: none; }

/* ─── Form elements ─── */
.form-group { display: flex; flex-direction: column; gap: 3px; }
.form-label { font-size: 11px; font-weight: 800; color: #1e293b; letter-spacing: 0.7px; text-transform: uppercase; }
.req { color: #ef4444; }
.form-input { border: 2px solid #1e293b; padding: 8px 10px; font-size: 13px; color: #374151; outline: none; border-radius: 0; width: 100%; box-sizing: border-box; height: 40px; transition: border-color .12s; }
.form-input:focus { border-color: #3b82f6; }
.form-input-disabled { background: #f1f5f9; color: #94a3b8; cursor: not-allowed; }
.form-hint { font-size: 10px; color: #94a3b8; margin-top: 2px; }

.check-row { display: flex; gap: 24px; flex-wrap: wrap; }
.check-label { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 600; color: #374151; cursor: pointer; }
.check-label input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; accent-color: #3b82f6; }
.check-disabled { opacity: .5; cursor: not-allowed; }
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
.btn-generate:hover { background: #fff; color: #3b82f6; transform: translate(-1px,-1px); box-shadow: 3px 3px 0 #1e293b; }
.pwd-strength { display: flex; align-items: center; gap: 10px; margin-top: 4px; }
.strength-bar-wrap { flex: 1; height: 5px; background: #e2e8f0; border: 1px solid #d1d5db; overflow: hidden; }
.strength-bar { height: 100%; transition: width .3s, background .3s; }
.strength-label { font-size: 10px; font-weight: 800; letter-spacing: 0.5px; text-transform: uppercase; white-space: nowrap; }
.pwd-mismatch { font-size: 11px; color: #dc2626; font-weight: 600; }

/* ─── Collapsible sections ─── */
.adv-details { border: 2px solid #1e293b; }
.adv-summary { list-style: none; padding: 12px 16px; cursor: pointer; font-size: 12px; font-weight: 900; color: #1e293b; letter-spacing: 0.6px; text-transform: uppercase; font-family: monospace; display: flex; align-items: center; justify-content: space-between; background: #f8fafc; transition: background .12s; user-select: none; }
.adv-summary:hover { background: #f1f5f9; }
.adv-summary::-webkit-details-marker { display: none; }
.adv-summary-left { display: flex; align-items: center; }
.adv-chevron { transition: transform .2s; }
.adv-details[open] .adv-chevron { transform: rotate(180deg); }
.adv-body { padding: 16px; display: flex; flex-direction: column; gap: 12px; border-top: 2px solid #1e293b; }

/* ─── Domains grid ─── */
.domains-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 8px; max-height: 180px; overflow-y: auto; padding: 10px; background: #f8fafc; border: 2px solid #e2e8f0; }
.domain-item { display: flex; align-items: center; gap: 8px; background: #fff; padding: 6px 10px; border: 1px solid #e2e8f0; cursor: pointer; font-size: 12px; font-weight: 500; color: #374151; transition: border-color .12s; }
.domain-item:hover { border-color: #3b82f6; }
.domain-item-checked { border-color: #3b82f6; background: #eff6ff; }
.domain-item input[type="checkbox"] { width: 14px; height: 14px; cursor: pointer; accent-color: #3b82f6; flex-shrink: 0; }
.domain-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.domains-empty { font-size: 12px; color: #94a3b8; text-align: center; padding: 12px; }
.superadmin-note { font-size: 12px; color: #7c3aed; background: #ede9fe; padding: 8px 12px; border: 1px solid #c4b5fd; font-weight: 600; }

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
