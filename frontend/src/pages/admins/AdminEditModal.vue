<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
    <div class="bg-white border-2 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0 border-b-2 border-brand-text">
        <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
          <Icon name="users" :size="20" class="mr-2" />
          EDIT ADMINISTRATOR
          <span class="text-white/80 ml-2 font-normal">— {{ form.username }}</span>
        </h3>
        <button @click="emit('update:modelValue', false)" class="text-white hover:text-gray-200 transition-colors">
          <Icon name="x" :size="20" />
        </button>
      </div>

      <!-- Scrollable Body -->
      <div class="overflow-y-auto flex-1">
        <div v-if="loading" class="flex flex-col items-center justify-center p-12">
          <div class="spinner mb-2" style="width:32px;height:32px" />
          <span class="text-xs font-black uppercase tracking-widest text-brand-text">LOADING...</span>
        </div>
        <template v-else>
          <div class="p-6 space-y-5">
            <!-- Admin Details -->
            <div class="border-2 border-brand-text p-4 space-y-4">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="user" :size="16" class="mr-2" />
                ADMIN DETAILS
              </h4>

              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">USERNAME</label>
                <div class="h-10 px-3 flex items-center border-2 border-gray-300 bg-gray-50 font-mono text-sm font-medium text-gray-500">
                  {{ form.username }}
                </div>
                <p class="text-[10px] text-gray-400 mt-1">Username cannot be changed after creation</p>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div class="flex items-center">
                  <input type="checkbox" id="edit-active" v-model="form.active" :disabled="!isSuperAdmin" class="w-5 h-5 border-2 border-brand-text cursor-pointer disabled:opacity-50" />
                  <label for="edit-active" class="ml-2 text-sm font-bold cursor-pointer" :class="{ 'opacity-50': !isSuperAdmin }">Active</label>
                </div>
                <div class="flex items-center">
                  <input type="checkbox" id="edit-superadmin" v-model="form.superadmin" :disabled="!isSuperAdmin" @change="onSuperAdminChange" class="w-5 h-5 border-2 border-brand-text cursor-pointer disabled:opacity-50" />
                  <label for="edit-superadmin" class="ml-2 text-sm font-bold cursor-pointer" :class="{ 'opacity-50': !isSuperAdmin }" title="Grants full access to all domains and settings">Super Administrator</label>
                </div>
              </div>
            </div>

            <!-- Change Password -->
            <details class="border-2 border-brand-text">
              <summary class="px-4 py-3 cursor-pointer font-bold uppercase tracking-tight text-xs flex items-center hover:bg-gray-50 transition-colors">
                <Icon name="key" :size="14" class="mr-2" />
                CHANGE PASSWORD
                <Icon name="chevron-down" :size="14" class="ml-auto" />
              </summary>
              <div class="px-4 pb-4 pt-3 space-y-3 border-t-2 border-brand-text">
                <div class="flex gap-2 items-center">
                  <div class="relative flex-1">
                    <input :type="showPw3 ? 'text' : 'password'" v-model="form.password"
                      placeholder="New password (min. 8 chars)"
                      class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm pr-10"
                      @input="onPasswordInput" />
                    <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-brand-primary transition-colors" @click="showPw3 = !showPw3">
                      <Icon :name="showPw3 ? 'eye-off' : 'eye'" :size="16" />
                    </button>
                  </div>
                  <div class="relative flex-1">
                    <input :type="showPw4 ? 'text' : 'password'" v-model="form.passwordConfirm"
                      placeholder="Repeat new password"
                      class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm pr-10"
                      @input="onPasswordInput" />
                    <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-brand-primary transition-colors" @click="showPw4 = !showPw4">
                      <Icon :name="showPw4 ? 'eye-off' : 'eye'" :size="16" />
                    </button>
                  </div>
                  <button type="button" @click="genPassword"
                    class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white text-xs font-black px-3 h-10 border-2 border-brand-text cursor-pointer uppercase tracking-widest flex items-center gap-1 whitespace-nowrap flex-shrink-0 transition-colors">
                    <Icon name="wand-2" :size="14" />
                    GENERATE
                  </button>
                </div>

                <div v-if="form.password">
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-xs font-black uppercase tracking-widest text-brand-text">STRENGTH</span>
                    <span class="text-xs font-bold uppercase tracking-wider" :style="{ color: pwdStrength.color }">{{ pwdStrength.label }}</span>
                  </div>
                  <div class="h-2 bg-gray-200 border border-brand-text overflow-hidden">
                    <div class="h-full transition-all duration-300" :style="{ width: pwdStrength.pct + '%', background: pwdStrength.color }"></div>
                  </div>
                </div>
                <p v-if="pwdMismatch" class="text-xs text-red-600 font-bold">Passwords do not match</p>
              </div>
            </details>

            <!-- Assigned Domains -->
            <div class="border-2 border-brand-text p-4 space-y-4" :class="{ 'opacity-50 pointer-events-none': form.superadmin || !isSuperAdmin }">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="globe" :size="16" class="mr-2" />
                ASSIGNED DOMAINS
                <span class="text-[10px] text-gray-400 ml-2 font-normal uppercase tracking-wider">Select domains this admin can manage</span>
              </h4>

              <div v-if="form.superadmin" class="bg-blue-50 border-2 border-blue-300 p-3 text-sm text-blue-700 font-bold">
                Super admins have access to all domains automatically.
              </div>
              <div v-else-if="editDomains.length === 0" class="text-xs font-bold text-gray-400">No domains available.</div>
              <div v-else class="grid grid-cols-2 sm:grid-cols-4 gap-2">
                <label
                  v-for="d in editDomains" :key="d.domain"
                  class="flex items-center p-2 border-2 border-brand-text bg-white hover:bg-gray-50 transition-colors cursor-pointer select-none"
                  :class="{ 'bg-blue-50/50 border-brand-primary': d.assigned }"
                >
                  <input type="checkbox" v-model="d.assigned" :disabled="form.superadmin || !isSuperAdmin" class="w-4 h-4 border-2 border-brand-text cursor-pointer flex-shrink-0" />
                  <span class="ml-2 text-xs font-bold truncate text-brand-text" :title="d.domain">{{ d.domain }}</span>
                </label>
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-brand-text flex-shrink-0 bg-white">
        <button type="button" @click="emit('update:modelValue', false)"
          class="bg-white hover:bg-gray-50 text-brand-text border-2 border-brand-text font-black px-6 py-3 shadow-[2px_2px_0px_#1E293B] transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm">
          <Icon name="x" :size="16" class="mr-2" />
          CANCEL
        </button>
        <button type="button" :disabled="saving || loading" @click="submit"
          class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-3 shadow-[3px_3px_0px_#1E293B] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm disabled:opacity-60">
          <Icon name="save" :size="16" class="mr-2" />
          {{ saving ? 'SAVING...' : 'SAVE CHANGES' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { calcStrength, generatePassword as fillPassword } from '../../utils/password'

interface DomainOption { domain: string; assigned: boolean }

const props = defineProps<{
  modelValue: boolean
  isSuperAdmin: boolean
  saving: boolean
  loading: boolean
  editDomains: DomainOption[]
  initialData: { username: string; active: boolean; superadmin: boolean } | null
}>()

const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  submit: [payload: {
    active: boolean
    superadmin: boolean
    domains: string[]
    change_password?: boolean
    password?: string
    password_confirm?: string
  }]
}>()

const showPw3 = ref(false)
const showPw4 = ref(false)
const pwdMismatch = ref(false)

const form = ref({ username: '', password: '', passwordConfirm: '', active: true, superadmin: false })

const pwdStrength = computed(() => calcStrength(form.value.password))

watch(() => props.initialData, (data) => {
  if (data) {
    form.value = { ...data, password: '', passwordConfirm: '' }
    showPw3.value = false; showPw4.value = false; pwdMismatch.value = false
  }
}, { immediate: true })

function onPasswordInput() {
  pwdMismatch.value = !!(form.value.passwordConfirm && form.value.password !== form.value.passwordConfirm)
}

function onSuperAdminChange() {
  if (form.value.superadmin) props.editDomains.forEach(d => d.assigned = true)
}

function genPassword() {
  fillPassword(form.value)
  showPw3.value = true; showPw4.value = true
  onPasswordInput()
}

function submit() {
  const payload: any = {
    active: form.value.active,
    superadmin: form.value.superadmin,
    domains: form.value.superadmin ? [] : props.editDomains.filter(d => d.assigned).map(d => d.domain),
  }
  if (form.value.password) {
    payload.change_password = true
    payload.password = form.value.password
    payload.password_confirm = form.value.passwordConfirm
  }
  emit('submit', payload)
}
</script>
