<template>
  <div class="rm-page">
    <!-- ─── Header ─── -->
    <div class="page-header">
      <div>
        <div class="page-title">ROLE MANAGEMENT</div>
        <div class="page-subtitle">MANAGE RBAC ROLES AND THEIR PERMISSIONS</div>
      </div>
      <button class="btn-primary" @click="openCreate">
        <Icon name="plus-circle" :size="16" style="margin-right: 0.375rem; vertical-align: middle" />
        NEW ROLE
      </button>
    </div>

    <!-- ─── Roles Grid ─── -->
    <div v-if="loading" class="flex justify-center py-16">
      <div class="spinner" style="width: 2rem; height: 2rem" />
    </div>

    <div v-else class="roles-grid">
      <div v-for="role in roles" :key="role.id" class="role-card" :class="{ 'role-card--system': role.system }">
        <!-- Card header -->
        <div class="role-card-header">
          <div class="role-card-title">
            <Icon name="shield" :size="16" class="mr-2 text-action" />
            {{ role.name }}
            <span v-if="role.system" class="badge-system">SYSTEM</span>
          </div>
          <div class="flex gap-1">
            <button class="act-btn act-edit" title="Edit role" @click="openEdit(role)">
              <Icon name="pencil" :size="12" />
            </button>
            <button v-if="!role.system" class="act-btn act-del" title="Delete role" @click="confirmDelete(role)">
              <Icon name="trash-2" :size="12" />
            </button>
          </div>
        </div>

        <!-- Description -->
        <p class="role-card-desc">{{ role.description || '—' }}</p>

        <!-- Permissions -->
        <div class="role-perms">
          <span v-for="p in role.permissions" :key="p.id" class="perm-badge" :class="permClass(p.action)">{{
            p.name
          }}</span>
          <span v-if="!role.permissions?.length" class="text-xs text-fg-subtle">No permissions</span>
        </div>
      </div>
    </div>

    <!-- ══════════ CREATE ROLE MODAL ══════════ -->
    <div v-if="showCreate" class="modal-overlay" @click.self="showCreate = false">
      <div class="bg-surface border-2 border-line w-full max-w-lg max-h-[90vh] flex flex-col">
        <div class="bg-action px-6 py-4 flex items-center justify-between flex-shrink-0 border-b-2 border-line">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-on-action flex items-center">
            <Icon name="shield-plus" :size="20" class="mr-2" />
            CREATE CUSTOM ROLE
          </h3>
          <button @click="showCreate = false" class="text-on-action hover:text-on-action/80">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <div class="overflow-y-auto flex-1 p-6 space-y-5">
          <div class="border-2 border-line p-4 space-y-4">
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1"
                >ROLE NAME <span class="text-err">*</span></label
              >
              <input
                v-model="createForm.name"
                type="text"
                placeholder="e.g. billing_admin"
                class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium text-sm"
              />
            </div>
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">DESCRIPTION</label>
              <input
                v-model="createForm.description"
                type="text"
                placeholder="Short description of this role"
                class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium text-sm"
              />
            </div>
          </div>
          <div class="border-2 border-line p-4 space-y-3">
            <h4 class="text-xs font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="lock" :size="14" class="mr-2" /> PERMISSIONS
            </h4>
            <div class="grid grid-cols-2 gap-2 max-h-60 overflow-y-auto pr-1">
              <label
                v-for="p in allPermissions"
                :key="p.id"
                class="flex items-center p-2 border-2 border-line bg-surface hover:bg-surface-2 cursor-pointer select-none"
                :class="{ 'bg-action-soft/50 border-action': createForm.permissionIds.includes(p.id) }"
              >
                <input
                  type="checkbox"
                  :value="p.id"
                  v-model="createForm.permissionIds"
                  class="w-4 h-4 border-2 border-line cursor-pointer flex-shrink-0"
                />
                <span class="ml-2 text-xs font-bold truncate" :class="permClass(p.action)">{{ p.name }}</span>
              </label>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-line bg-surface">
          <button
            type="button"
            @click="showCreate = false"
            class="bg-surface hover:bg-surface-2 text-fg border-2 border-line font-black px-6 py-3 shadow-[2px_2px_0px_var(--color-shadow)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_var(--color-shadow)] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm transition-all"
          >
            <Icon name="x" :size="16" class="mr-2" /> CANCEL
          </button>
          <button
            type="button"
            :disabled="creating"
            @click="submitCreate"
            class="bg-action hover:bg-surface hover:text-action text-on-action border-2 border-line font-black px-6 py-3 shadow-[3px_3px_0px_var(--color-shadow)] hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_var(--color-shadow)] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm transition-all disabled:opacity-60"
          >
            <Icon name="plus-circle" :size="16" class="mr-2" />
            {{ creating ? 'CREATING...' : 'CREATE ROLE' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ EDIT ROLE MODAL ══════════ -->
    <div v-if="showEdit" class="modal-overlay" @click.self="showEdit = false">
      <div class="bg-surface border-2 border-line w-full max-w-lg max-h-[90vh] flex flex-col">
        <div class="bg-action px-6 py-4 flex items-center justify-between flex-shrink-0 border-b-2 border-line">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-on-action flex items-center">
            <Icon name="shield" :size="20" class="mr-2" />
            EDIT ROLE — {{ editTarget?.name }}
          </h3>
          <button @click="showEdit = false" class="text-on-action hover:text-on-action/80">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <div class="overflow-y-auto flex-1 p-6 space-y-5">
          <!-- System role notice -->
          <div
            v-if="editTarget?.system"
            class="bg-action-soft border-2 border-action-line-soft p-3 flex items-start gap-2"
          >
            <Icon name="info" :size="16" class="text-action mt-0.5 flex-shrink-0" />
            <p class="text-xs font-bold text-action-strong">
              This is a system role. You can update its description but permissions are managed by the system.
            </p>
          </div>

          <div class="border-2 border-line p-4 space-y-4">
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">DESCRIPTION</label>
              <input
                v-model="editForm.description"
                type="text"
                placeholder="Short description of this role"
                class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium text-sm"
              />
            </div>
          </div>

          <!-- Permissions — editable for custom roles, read-only for system roles -->
          <div class="border-2 border-line p-4 space-y-3">
            <h4 class="text-xs font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="lock" :size="14" class="mr-2" />
              PERMISSIONS
              <span v-if="editTarget?.system" class="ml-2 text-[0.625rem] font-normal text-fg-subtle normal-case"
                >(read-only for system roles)</span
              >
            </h4>
            <div class="grid grid-cols-2 gap-2 max-h-60 overflow-y-auto pr-1">
              <label
                v-for="p in allPermissions"
                :key="p.id"
                class="flex items-center p-2 border-2 select-none"
                :class="
                  editTarget?.system
                    ? editForm.permissionIds.includes(p.id)
                      ? 'border-action bg-action-soft/50 cursor-default'
                      : 'border-line-soft bg-surface-2 cursor-default opacity-50'
                    : editForm.permissionIds.includes(p.id)
                      ? 'border-action bg-action-soft/50 cursor-pointer hover:bg-action-soft'
                      : 'border-line bg-surface cursor-pointer hover:bg-surface-2'
                "
              >
                <input
                  type="checkbox"
                  :value="p.id"
                  v-model="editForm.permissionIds"
                  :disabled="editTarget?.system"
                  class="w-4 h-4 border-2 border-line flex-shrink-0"
                  :class="editTarget?.system ? 'cursor-default' : 'cursor-pointer'"
                />
                <span class="ml-2 text-xs font-bold truncate" :class="permClass(p.action)">{{ p.name }}</span>
              </label>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-line bg-surface">
          <button
            type="button"
            @click="showEdit = false"
            class="bg-surface hover:bg-surface-2 text-fg border-2 border-line font-black px-6 py-3 shadow-[2px_2px_0px_var(--color-shadow)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_var(--color-shadow)] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm transition-all"
          >
            <Icon name="x" :size="16" class="mr-2" /> CANCEL
          </button>
          <button
            type="button"
            :disabled="saving"
            @click="submitEdit"
            class="bg-action hover:bg-surface hover:text-action text-on-action border-2 border-line font-black px-6 py-3 shadow-[3px_3px_0px_var(--color-shadow)] hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_var(--color-shadow)] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm transition-all disabled:opacity-60"
          >
            <Icon name="save" :size="16" class="mr-2" />
            {{ saving ? 'SAVING...' : 'SAVE CHANGES' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <ConfirmDialog
      v-model="showDeleteConfirm"
      title="DELETE ROLE"
      confirm-label="DELETE"
      :loading="deleting"
      @confirm="submitDelete"
    >
      <p class="text-sm font-medium text-fg leading-relaxed">
        Delete role <strong class="font-mono">{{ deleteTarget?.name }}</strong
        >?<br />
        <span class="text-xs text-fg-muted mt-1 block"
          >All admin assignments for this role will be removed. This action is permanent.</span
        >
      </p>
    </ConfirmDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import http from '../../utils/http'
import { useToastStore } from '../../stores/toast'
import { ConfirmDialog } from '../../components/ui'

const toast = useToastStore()

interface Permission {
  id: number
  name: string
  resource: string
  action: string
}
interface Role {
  id: number
  name: string
  description: string
  system: boolean
  permissions: Permission[]
}

const roles = ref<Role[]>([])
const allPermissions = ref<Permission[]>([])
const loading = ref(true)

async function load() {
  loading.value = true
  try {
    const [rolesRes, permsRes] = await Promise.all([
      http.get(`${API_BASE}/rbac/roles`),
      http.get(`${API_BASE}/rbac/permissions`),
    ])
    roles.value = rolesRes.data?.data ?? []
    allPermissions.value = permsRes.data?.data ?? []
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to load roles')
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ─── Create ───
const showCreate = ref(false)
const creating = ref(false)
const createForm = ref({ name: '', description: '', permissionIds: [] as number[] })

function openCreate() {
  createForm.value = { name: '', description: '', permissionIds: [] }
  showCreate.value = true
}

async function submitCreate() {
  if (!createForm.value.name.trim()) {
    toast.error('Role name is required')
    return
  }
  creating.value = true
  try {
    await http.post(`${API_BASE}/rbac/roles`, {
      name: createForm.value.name.trim(),
      description: createForm.value.description.trim(),
      permission_ids: createForm.value.permissionIds,
    })
    toast.success(`Role "${createForm.value.name}" created`)
    showCreate.value = false
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to create role')
  } finally {
    creating.value = false
  }
}

// ─── Edit ───
const showEdit = ref(false)
const saving = ref(false)
const editTarget = ref<Role | null>(null)
const editForm = ref({ description: '', permissionIds: [] as number[] })

function openEdit(role: Role) {
  editTarget.value = role
  editForm.value = {
    description: role.description,
    permissionIds: role.permissions?.map((p) => p.id) ?? [],
  }
  showEdit.value = true
}

async function submitEdit() {
  if (!editTarget.value) return
  saving.value = true
  try {
    await http.put(`${API_BASE}/rbac/roles/${editTarget.value.id}`, {
      description: editForm.value.description.trim(),
      permission_ids: editTarget.value.system ? undefined : editForm.value.permissionIds,
    })
    toast.success(`Role "${editTarget.value.name}" updated`)
    showEdit.value = false
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to update role')
  } finally {
    saving.value = false
  }
}

// ─── Delete ───
const showDeleteConfirm = ref(false)
const deleting = ref(false)
const deleteTarget = ref<Role | null>(null)

function confirmDelete(role: Role) {
  deleteTarget.value = role
  showDeleteConfirm.value = true
}

async function submitDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await http.delete(`${API_BASE}/rbac/roles/${deleteTarget.value.id}`)
    toast.success(`Role "${deleteTarget.value.name}" deleted`)
    showDeleteConfirm.value = false
    deleteTarget.value = null
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to delete role')
    showDeleteConfirm.value = false
  } finally {
    deleting.value = false
  }
}

// ─── Helpers ───
function permClass(action: string): string {
  switch (action) {
    case 'delete':
      return 'perm-delete'
    case 'write':
      return 'perm-write'
    default:
      return 'perm-read'
  }
}
</script>

<style scoped>
.rm-page {
  background: var(--color-page);
  padding: 1.5rem 1.75rem 2.5rem;
}

.roles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(20rem, 1fr));
  gap: 1rem;
}

.role-card {
  background: var(--color-surface);
  border: 2px solid var(--color-line);
  box-shadow: 3px 3px 0 var(--color-shadow);
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}
.role-card--system {
  border-color: var(--color-action);
  box-shadow: 3px 3px 0 var(--color-action);
}

.role-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.role-card-title {
  font-size: 0.8125rem;
  font-weight: 900;
  font-family: monospace;
  color: var(--color-fg);
  letter-spacing: 0.01875rem;
  display: flex;
  align-items: center;
  text-transform: uppercase;
}
.badge-system {
  margin-left: 0.5rem;
  font-size: 0.5625rem;
  font-weight: 900;
  background: var(--color-action);
  color: var(--color-on-action);
  padding: 0.0625rem 0.375rem;
  letter-spacing: 0.03125rem;
}
.role-card-desc {
  font-size: 0.75rem;
  color: var(--color-fg-muted);
  margin: 0;
}

.role-perms {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}
.perm-badge {
  font-size: 0.625rem;
  font-weight: 700;
  padding: 0.125rem 0.375rem;
  border: 1px solid currentColor;
  font-family: monospace;
}
.perm-read {
  color: var(--color-ok-strong);
}
.perm-write {
  color: var(--color-warn-strong);
}
.perm-delete {
  color: var(--color-danger);
}

.act-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.025rem;
  cursor: pointer;
  border: 1px solid var(--color-line);
  text-transform: uppercase;
  transition: all 0.1s;
  background: var(--color-surface);
}
.act-edit {
  color: var(--color-fg);
}
.act-edit:hover {
  background: var(--color-strong-fill);
  color: var(--color-on-action);
}
.act-del {
  color: var(--color-danger);
  border-color: var(--color-danger);
}
.act-del:hover {
  background: var(--color-danger);
  color: var(--color-on-action);
}
</style>
