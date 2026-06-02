<template>
  <div class="border-2 border-brand-text p-4 space-y-4">
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
    <div v-if="loadingRoles" class="flex items-center gap-2 text-xs text-gray-400 font-bold">
      <div class="spinner" style="width:14px;height:14px" /> LOADING...
    </div>
    <div v-else-if="assignments.length === 0" class="text-xs font-bold text-gray-400">
      No roles assigned. Using default permissions from domain_admins.
    </div>
    <div v-else class="space-y-2">
      <div
        v-for="a in assignments"
        :key="a.id"
        class="assignment-row"
      >
        <div class="assignment-info">
          <Icon name="shield" :size="13" class="text-brand-primary flex-shrink-0" />
          <span class="font-mono font-bold text-xs text-brand-text uppercase">{{ a.role.name }}</span>
          <span v-if="a.domain" class="text-xs text-gray-500">@ {{ a.domain }}</span>
          <span v-else class="text-xs text-gray-400">(global)</span>
        </div>
        <button class="act-btn act-del" @click="removeRole(a)" :disabled="removing === a.id">
          <Icon name="x" :size="11" />
        </button>
      </div>
    </div>

    <!-- Assign modal -->
    <div v-if="showAssign" class="modal-overlay" @click.self="showAssign = false">
      <div class="bg-white border-2 border-brand-text w-full max-w-md max-h-[80vh] flex flex-col">
        <div class="bg-brand-primary px-5 py-3 flex items-center justify-between border-b-2 border-brand-text">
          <h3 class="text-base font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="shield-plus" :size="16" class="mr-2" />
            ASSIGN ROLE — {{ username }}
          </h3>
          <button @click="showAssign = false" class="text-white hover:text-gray-200">
            <Icon name="x" :size="18" />
          </button>
        </div>

        <div class="overflow-y-auto flex-1 p-5 space-y-4">
          <div>
            <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">ROLE <span class="text-red-500">*</span></label>
            <select v-model.number="assignForm.roleId"
              class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm bg-white cursor-pointer">
              <option :value="0" disabled>Select a role…</option>
              <option v-for="r in availableRoles" :key="r.id" :value="r.id">{{ r.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">DOMAIN SCOPE</label>
            <input v-model="assignForm.domain" type="text" placeholder="Leave empty for global scope (all domains)"
              class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm" />
            <p class="text-[10px] text-gray-400 mt-1">Optional. Restricts the role to a specific domain.</p>
          </div>
        </div>

        <div class="flex items-center justify-end space-x-3 px-5 py-3 border-t-2 border-brand-text bg-white">
          <button type="button" @click="showAssign = false"
            class="bg-white text-brand-text border-2 border-brand-text font-black px-5 py-2 cursor-pointer uppercase tracking-widest flex items-center text-xs hover:bg-gray-50 transition-colors">
            <Icon name="x" :size="14" class="mr-1" /> CANCEL
          </button>
          <button type="button" :disabled="assigning || !assignForm.roleId" @click="submitAssign"
            class="bg-brand-primary text-white border-2 border-brand-text font-black px-5 py-2 cursor-pointer uppercase tracking-widest flex items-center text-xs hover:bg-white hover:text-brand-primary transition-colors disabled:opacity-60">
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

interface Role { id: number; name: string; description: string; system: boolean }
interface Assignment { id: number; username: string; role_id: number; domain: string; role: Role }

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
  padding: 6px 8px;
  border: 2px solid #e2e8f0;
  background: #f8fafc;
}
.assignment-info {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}
.act-btn {
  display: inline-flex; align-items: center; gap: 3px;
  padding: 3px 7px; font-size: 10px; font-weight: 800;
  letter-spacing: 0.4px; cursor: pointer;
  border: 1px solid #1e293b; text-transform: uppercase;
  transition: all .1s; background: #fff;
}
.act-add { color: #1e293b; }
.act-add:hover:not(:disabled) { background: #1e293b; color: #fff; }
.act-del { color: #dc2626; border-color: #dc2626; }
.act-del:hover:not(:disabled) { background: #dc2626; color: #fff; }
.act-btn:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
