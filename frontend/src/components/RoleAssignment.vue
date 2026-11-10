<template>
  <div class="border-2 border-line p-4 space-y-4">
    <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center justify-between">
      <span class="flex items-center">
        <Icon name="shield" :size="16" class="mr-2" />
        RBAC ROLES
      </span>
      <button class="act-btn act-add" @click="openAssign" :disabled="loadingRoles">
        <Icon name="plus" :size="12" /> ASSIGN
      </button>
    </h4>

    <!-- Current assignments -->
    <div v-if="loadingRoles" class="flex items-center gap-2 text-xs text-fg-subtle font-bold">
      <div class="spinner" style="width: 0.875rem; height: 0.875rem" />
      LOADING...
    </div>
    <div v-else-if="assignments.length === 0" class="text-xs font-bold text-fg-subtle">
      No roles assigned. Using default permissions from domain_admins.
    </div>
    <div v-else class="space-y-2">
      <div v-for="a in assignments" :key="a.id" class="assignment-row">
        <div class="assignment-info">
          <Icon name="shield" :size="13" class="text-icon-accent flex-shrink-0" />
          <span class="font-mono font-bold text-xs text-fg uppercase">{{ a.role.name }}</span>
          <span v-if="a.domain" class="text-xs text-fg-muted">@ {{ a.domain }}</span>
          <span v-else class="text-xs text-fg-subtle">(global)</span>
        </div>
        <button class="act-btn act-del" @click="removeRole(a)" :disabled="removing === a.id">
          <Icon name="x" :size="11" />
        </button>
      </div>
    </div>

    <!-- Assign modal -->
    <div v-if="showAssign" class="modal-overlay" @click.self="showAssign = false">
      <div class="bg-surface border-2 border-line w-full max-w-md max-h-[80vh] flex flex-col">
        <div class="bg-action px-5 py-3 flex items-center justify-between border-b-2 border-line">
          <h3 class="text-base font-mono font-black uppercase tracking-tight text-on-action flex items-center">
            <Icon name="shield-plus" :size="16" class="mr-2" />
            ASSIGN ROLE — {{ username }}
          </h3>
          <button @click="showAssign = false" class="text-on-action hover:text-on-action/80">
            <Icon name="x" :size="18" />
          </button>
        </div>

        <div class="overflow-y-auto flex-1 p-5 space-y-4">
          <div>
            <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1"
              >ROLE <span class="text-err">*</span></label
            >
            <select
              v-model.number="assignForm.roleId"
              class="w-full h-10 px-3 border-2 border-line focus:border-icon-accent focus:outline-none font-medium text-sm bg-surface cursor-pointer"
            >
              <option :value="0" disabled>Select a role…</option>
              <option v-for="r in availableRoles" :key="r.id" :value="r.id">{{ r.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">DOMAIN SCOPE</label>
            <input
              v-model="assignForm.domain"
              type="text"
              placeholder="Leave empty for global scope (all domains)"
              class="w-full h-10 px-3 border-2 border-line focus:border-icon-accent focus:outline-none font-medium text-sm"
            />
            <p class="text-[0.625rem] text-fg-subtle mt-1">Optional. Restricts the role to a specific domain.</p>
          </div>
        </div>

        <div class="flex items-center justify-end space-x-3 px-5 py-3 border-t-2 border-line bg-surface">
          <button
            type="button"
            @click="showAssign = false"
            class="bg-surface text-fg border-2 border-line font-black px-5 py-2 cursor-pointer uppercase tracking-widest flex items-center text-xs hover:bg-surface-2 transition-colors"
          >
            <Icon name="x" :size="14" class="mr-1" /> CANCEL
          </button>
          <button
            type="button"
            :disabled="assigning || !assignForm.roleId"
            @click="submitAssign"
            class="bg-action text-on-action border-2 border-line font-black px-5 py-2 cursor-pointer uppercase tracking-widest flex items-center text-xs hover:bg-surface hover:text-icon-accent transition-colors disabled:opacity-60"
          >
            <Icon name="check" :size="14" class="mr-1" />
            {{ assigning ? 'ASSIGNING...' : 'ASSIGN' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import http from '../utils/http'
import { useToastStore } from '../stores/toast'

const props = defineProps<{ username: string }>()
const toast = useToastStore()

interface Role {
  id: number
  name: string
  description: string
  system: boolean
}
interface Assignment {
  id: number
  username: string
  role_id: number
  domain: string
  role: Role
}

const assignments = ref<Assignment[]>([])
const availableRoles = ref<Role[]>([])
const loadingRoles = ref(true)

async function load() {
  loadingRoles.value = true
  try {
    const [assignRes, rolesRes] = await Promise.all([
      http.get(`${API_BASE}/rbac/admins/${encodeURIComponent(props.username)}/roles`),
      http.get(`${API_BASE}/rbac/roles`),
    ])
    assignments.value = assignRes.data?.data ?? []
    availableRoles.value = rolesRes.data?.data ?? []
  } catch {
    // Non-fatal: RBAC may not be enabled yet.
  } finally {
    loadingRoles.value = false
  }
}

onMounted(load)

// ─── Assign ───
const showAssign = ref(false)
const assigning = ref(false)
const assignForm = ref({ roleId: 0, domain: '' })

function openAssign() {
  assignForm.value = { roleId: 0, domain: '' }
  showAssign.value = true
}

async function submitAssign() {
  if (!assignForm.value.roleId) return
  assigning.value = true
  try {
    await http.post(`${API_BASE}/rbac/admins/${encodeURIComponent(props.username)}/roles`, {
      role_id: assignForm.value.roleId,
      domain: assignForm.value.domain.trim(),
    })
    toast.success('Role assigned successfully')
    showAssign.value = false
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to assign role')
  } finally {
    assigning.value = false
  }
}

// ─── Remove ───
const removing = ref<number | null>(null)

async function removeRole(a: Assignment) {
  if (!confirm(`Remove role "${a.role.name}" from ${props.username}?`)) return
  removing.value = a.id
  try {
    await http.delete(`${API_BASE}/rbac/admins/${encodeURIComponent(props.username)}/roles/${a.id}`)
    toast.success('Role removed')
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || 'Failed to remove role')
  } finally {
    removing.value = null
  }
}
</script>

<style scoped>
.assignment-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.375rem 0.5rem;
  border: 2px solid var(--color-line-soft);
  background: var(--color-surface-2);
}
.assignment-info {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  min-width: 0;
}
.act-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.1875rem;
  padding: 0.1875rem 0.4375rem;
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.025rem;
  cursor: pointer;
  border: 1px solid var(--color-line);
  text-transform: uppercase;
  transition: all 0.1s;
  background: var(--color-surface);
}
.act-add {
  color: var(--color-fg);
}
.act-add:hover:not(:disabled) {
  background: var(--color-strong-fill);
  color: var(--color-on-action);
}
.act-del {
  color: var(--color-danger);
  border-color: var(--color-danger);
}
.act-del:hover:not(:disabled) {
  background: var(--color-danger);
  color: var(--color-on-action);
}
.act-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
