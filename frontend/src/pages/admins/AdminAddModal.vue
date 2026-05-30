<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
    <div class="bg-white border-2 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0 border-b-2 border-brand-text">
        <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
          <Icon name="user-plus" :size="20" class="mr-2" />
          ADD NEW ADMINISTRATOR
        </h3>
        <button @click="emit('update:modelValue', false)" class="text-white hover:text-gray-200 transition-colors">
          <Icon name="x" :size="20" />
        </button>
      </div>

      <!-- Scrollable Body -->
      <div class="overflow-y-auto flex-1">
        <form class="p-6 space-y-5" @submit.prevent="submit">
          <!-- Admin Details -->
          <div class="border-2 border-brand-text p-4 space-y-4">
            <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="user" :size="16" class="mr-2" />
              ADMIN DETAILS
            </h4>

            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                USERNAME (EMAIL) <span class="text-red-500">*</span>
              </label>
              <input
                v-model="form.username"
                type="email"
                required
                placeholder="admin@example.com"
                class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm lowercase"
                style="text-transform: lowercase;"
                @input="form.username = form.username.toLowerCase()"
              />
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div class="flex items-center">
                <input type="checkbox" id="add-active" v-model="form.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                <label for="add-active" class="ml-2 text-sm font-bold cursor-pointer">Active</label>
              </div>
              <div class="flex items-center">
                <input type="checkbox" id="add-superadmin" v-model="form.superadmin" @change="onSuperAdminChange" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                <label for="add-superadmin" class="ml-2 text-sm font-bold cursor-pointer" title="Grants full access to all domains and settings">Super Administrator</label>
              </div>
            </div>
          </div>

          <!-- Password -->
          <div class="border-2 border-brand-text p-4 space-y-3">
            <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="key" :size="16" class="mr-2" />
              PASSWORD
            </h4>

            <div class="flex gap-2 items-center">
              <div class="relative flex-1">
                <input
                  :type="showPw1 ? 'text' : 'password'"
                  v-model="form.password"
                  required
                  minlength="8"
                  placeholder="New password"
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm pr-10"
                  @input="onPasswordInput"
                />
                <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-brand-primary transition-colors" @click="showPw1 = !showPw1">
                  <Icon :name="showPw1 ? 'eye-off' : 'eye'" :size="16" />
                </button>
              </div>
              <div class="relative flex-1">
                <input
                  :type="showPw2 ? 'text' : 'password'"
                  v-model="form.passwordConfirm"
                  required
                  placeholder="Repeat password"
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm pr-10"
                  @input="onPasswordInput"
                />
                <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-brand-primary transition-colors" @click="showPw2 = !showPw2">
                  <Icon :name="showPw2 ? 'eye-off' : 'eye'" :size="16" />
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

          <!-- Assigned Domains -->
          <div class="border-2 border-brand-text p-4 space-y-4" :class="{ 'opacity-50 pointer-events-none': form.superadmin }">
            <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="globe" :size="16" class="mr-2" />
              ASSIGNED DOMAINS
              <span class="text-[10px] text-gray-400 ml-2 font-normal uppercase tracking-wider">Select domains this admin can manage</span>
            </h4>

            <div v-if="form.superadmin" class="bg-blue-50 border-2 border-blue-300 p-3 text-sm text-blue-700 font-bold">
              Super admins have access to all domains automatically.
            </div>
            <div v-else-if="domains.length === 0" class="text-xs font-bold text-gray-400">No domains available.</div>
            <div v-else class="grid grid-cols-2 sm:grid-cols-4 gap-2">
              <label
                v-for="d in domains" :key="d.domain"
                class="flex items-center p-2 border-2 border-brand-text bg-white hover:bg-gray-50 transition-colors cursor-pointer select-none"
                :class="{ 'bg-blue-50/50 border-brand-primary': form.domains.includes(d.domain) }"
              >
                <input
                  type="checkbox"
                  :value="d.domain"
                  :checked="form.domains.includes(d.domain)"
                  :disabled="form.superadmin"
                  class="w-4 h-4 border-2 border-brand-text cursor-pointer flex-shrink-0"
                  @change="toggleDomain(d.domain)"
                />
                <span class="ml-2 text-xs font-bold truncate text-brand-text" :title="d.domain">{{ d.domain }}</span>
              </label>
            </div>
          </div>
        </form>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-brand-text flex-shrink-0 bg-white">
        <button type="button" @click="emit('update:modelValue', false)"
          class="bg-white hover:bg-gray-50 text-brand-text border-2 border-brand-text font-black px-6 py-3 shadow-[2px_2px_0px_#1E293B] transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm">
          <Icon name="x" :size="16" class="mr-2" />
          CANCEL
        </button>
        <button type="button" :disabled="saving || pwdMismatch" @click="submit"
          class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-3 shadow-[3px_3px_0px_#1E293B] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm disabled:opacity-60">
          <Icon name="save" :size="16" class="mr-2" />
          {{ saving ? 'SAVING...' : 'CREATE ADMIN' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { calcStrength, generatePassword as fillPassword } from '../../utils/password'

interface Domain { domain: string }

const props = defineProps<{
  modelValue: boolean
  domains: Domain[]
  saving: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  submit: [payload: {
    username: string
    password: string
    active: boolean
    superadmin: boolean
    domains: string[]
  }]
}>()

const showPw1 = ref(false)
const showPw2 = ref(false)
const pwdMismatch = ref(false)

const form = ref({
  username: '', password: '', passwordConfirm: '',
  active: true, superadmin: false, domains: [] as string[],
})

const pwdStrength = computed(() => calcStrength(form.value.password))

watch(() => props.modelValue, (open) => {
  if (open) {
    form.value = { username: '', password: '', passwordConfirm: '', active: true, superadmin: false, domains: [] }
    showPw1.value = false; showPw2.value = false; pwdMismatch.value = false
  }
})

function onPasswordInput() {
  pwdMismatch.value = !!(form.value.passwordConfirm && form.value.password !== form.value.passwordConfirm)
}

function onSuperAdminChange() {
  if (form.value.superadmin) form.value.domains = []
}

function toggleDomain(domain: string) {
  const idx = form.value.domains.indexOf(domain)
  if (idx === -1) form.value.domains.push(domain)
  else form.value.domains.splice(idx, 1)
}

function genPassword() {
  fillPassword(form.value)
  showPw1.value = true; showPw2.value = true
  onPasswordInput()
}

function submit() {
  if (form.value.password !== form.value.passwordConfirm) {
    pwdMismatch.value = true
    return
  }

  emit('submit', {
    username: form.value.username.toLowerCase().trim(),
    password: form.value.password,
    active: form.value.active,
    superadmin: form.value.superadmin,
    domains: form.value.superadmin ? [] : form.value.domains,
  })
}
</script>
