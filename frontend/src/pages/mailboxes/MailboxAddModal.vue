<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
    <div class="bg-surface border-2 border-line w-full max-w-2xl max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="bg-action px-6 py-4 flex items-center justify-between flex-shrink-0">
        <h3 class="text-lg font-mono font-black uppercase tracking-tight text-on-action flex items-center">
          <Icon name="plus-circle" :size="20" class="mr-2" />
          ADD EMAIL ACCOUNT
        </h3>
        <button
          @click="emit('update:modelValue', false)"
          class="text-on-action hover:text-on-action/80 transition-colors"
        >
          <Icon name="x" :size="20" />
        </button>
      </div>

      <!-- Scrollable Body -->
      <div class="overflow-y-auto flex-1">
        <form class="p-6 space-y-5" @submit.prevent="submit">
          <!-- EMAIL ACCOUNT Section -->
          <div class="border-2 border-line p-4 space-y-4">
            <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="mail" :size="16" class="mr-2" />
              EMAIL ACCOUNT
            </h4>

            <!-- Username + Domain -->
            <div class="grid grid-cols-2 gap-3 items-start">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">
                  USERNAME <span class="text-err">*</span>
                </label>
                <input
                  v-model="form.localPart"
                  type="text"
                  required
                  minlength="4"
                  pattern="[a-z0-9._-]{4,}"
                  placeholder="username"
                  class="w-full h-10 px-3 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors text-sm lowercase"
                  style="text-transform: lowercase"
                  @input="onLocalPartInput"
                />
                <p class="text-[0.625rem] text-fg-subtle mt-1">
                  Lowercase letters, numbers, dots, hyphens (min. 4 chars)
                </p>
              </div>
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">
                  DOMAIN <span class="text-err">*</span>
                </label>
                <template v-if="selectedDomain">
                  <div
                    class="w-full h-10 px-3 border-2 border-line bg-surface-2 font-mono font-bold text-sm flex items-center"
                  >
                    {{ selectedDomain }}
                  </div>
                </template>
                <select
                  v-else
                  v-model="form.domain"
                  required
                  class="w-full h-10 px-3 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors cursor-pointer text-sm"
                >
                  <option value="" disabled>Select a domain...</option>
                  <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
                </select>
              </div>
            </div>

            <!-- Email Preview -->
            <div class="bg-surface-2 border border-field-line px-3 h-10 flex items-center gap-2">
              <span class="text-xs font-bold uppercase tracking-widest text-fg-subtle whitespace-nowrap"
                >FULL ADDRESS:</span
              >
              <p class="font-mono text-sm font-bold">
                <span class="text-icon-accent">{{ form.localPart || 'user' }}</span
                >@<span class="text-icon-accent">{{ form.domain || 'domain.com' }}</span>
              </p>
            </div>

            <!-- Display Name -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">DISPLAY NAME</label>
              <input
                v-model="form.name"
                type="text"
                placeholder="Full Name"
                class="w-full px-3 py-2 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors text-sm"
              />
            </div>

            <!-- Active + Welcome -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div class="flex items-center">
                <input
                  type="checkbox"
                  id="add-active"
                  v-model="form.active"
                  class="w-5 h-5 border-2 border-line cursor-pointer"
                />
                <label for="add-active" class="ml-2 text-sm font-bold cursor-pointer">Active Account</label>
              </div>
              <div class="flex items-center">
                <input
                  type="checkbox"
                  id="add-welcome"
                  v-model="form.sendWelcome"
                  class="w-5 h-5 border-2 border-line cursor-pointer"
                />
                <label for="add-welcome" class="ml-2 text-sm font-bold cursor-pointer">Send Welcome Email</label>
              </div>
            </div>
          </div>

          <!-- PASSWORD Section -->
          <div class="border-2 border-line p-4 space-y-3">
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
                  class="w-full h-10 px-3 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors text-sm pr-10"
                  @input="onPasswordInput"
                />
                <button
                  type="button"
                  class="absolute right-2 top-1/2 -translate-y-1/2 text-fg-muted hover:text-icon-accent transition-colors"
                  @click="showPw1 = !showPw1"
                >
                  <Icon :name="showPw1 ? 'eye-off' : 'eye'" :size="16" />
                </button>
              </div>
              <div class="relative flex-1">
                <input
                  :type="showPw2 ? 'text' : 'password'"
                  v-model="form.passwordConfirm"
                  required
                  minlength="8"
                  placeholder="Repeat password"
                  class="w-full h-10 px-3 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors text-sm pr-10"
                  @input="onPasswordInput"
                />
                <button
                  type="button"
                  class="absolute right-2 top-1/2 -translate-y-1/2 text-fg-muted hover:text-icon-accent transition-colors"
                  @click="showPw2 = !showPw2"
                >
                  <Icon :name="showPw2 ? 'eye-off' : 'eye'" :size="16" />
                </button>
              </div>
              <button
                type="button"
                @click="genPassword"
                class="bg-action hover:bg-surface hover:text-icon-accent text-on-action text-xs font-black px-3 h-10 border-2 border-line cursor-pointer uppercase tracking-widest flex items-center gap-1 whitespace-nowrap flex-shrink-0 transition-colors"
              >
                <Icon name="wand-2" :size="14" />
                GENERATE
              </button>
            </div>

            <p v-if="pwdMismatch" class="text-xs text-danger font-bold">Passwords do not match</p>

            <!-- Strength Meter -->
            <div v-if="form.password">
              <div class="flex items-center justify-between mb-1">
                <span class="text-xs font-black uppercase tracking-widest text-fg">STRENGTH</span>
                <span class="text-xs font-bold uppercase tracking-wider" :style="{ color: pwdStrength.color }">{{
                  pwdStrength.label
                }}</span>
              </div>
              <div class="h-2 bg-line-soft border border-line overflow-hidden">
                <div
                  class="h-full transition-all duration-300"
                  :style="{ width: pwdStrength.pct + '%', background: pwdStrength.color }"
                ></div>
              </div>
            </div>
          </div>

          <!-- ADVANCED SETTINGS -->
          <details class="border-2 border-line">
            <summary
              class="px-4 py-3 cursor-pointer font-bold uppercase tracking-tight text-xs flex items-center hover:bg-surface-2 transition-colors"
            >
              <Icon name="settings" :size="14" class="mr-2" />
              ADVANCED SETTINGS
              <Icon name="chevron-down" :size="14" class="ml-auto" />
            </summary>
            <div class="px-4 pb-4 pt-2 space-y-3 border-t-2 border-line">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">QUOTA (MB)</label>
                  <input
                    v-model.number="form.quotaMB"
                    type="number"
                    min="0"
                    class="w-full px-3 py-2 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors text-sm"
                  />
                  <p class="text-xs text-fg-muted mt-1">Storage limit in MB (0 = unlimited)</p>
                </div>
                <div class="flex items-center pt-5">
                  <input
                    type="checkbox"
                    id="add-smtp"
                    v-model="form.smtpActive"
                    class="w-5 h-5 border-2 border-line cursor-pointer"
                  />
                  <label for="add-smtp" class="ml-2 text-sm font-bold cursor-pointer"
                    >SMTP Active (can send email)</label
                  >
                </div>
              </div>
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">ALTERNATIVE EMAIL</label>
                <input
                  v-model="form.emailOther"
                  type="email"
                  placeholder="other@example.com"
                  class="w-full px-3 py-2 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors text-sm"
                />
              </div>
            </div>
          </details>
        </form>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-line flex-shrink-0 bg-surface">
        <button
          type="button"
          @click="emit('update:modelValue', false)"
          class="btn-cancel bg-surface hover:bg-surface-2 text-fg border-2 border-line font-black px-6 py-3 shadow-(--shadow-2) transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-(--shadow-3) active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm"
        >
          <Icon name="x" :size="16" class="mr-2" />
          CANCEL
        </button>
        <button
          type="button"
          :disabled="saving || pwdMismatch"
          @click="submit"
          class="bg-action hover:bg-surface hover:text-icon-accent text-on-action border-2 border-line font-black px-6 py-3 shadow-(--shadow-3) transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-(--shadow-4) active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm disabled:opacity-60"
        >
          <Icon name="save" :size="16" class="mr-2" />
          {{ saving ? 'SAVING...' : 'SAVE EMAIL ACCOUNT' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { calcStrength, generatePassword } from '../../utils/password'

interface Domain {
  domain: string
}

const props = defineProps<{
  modelValue: boolean
  domains: Domain[]
  saving: boolean
  selectedDomain?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  submit: [
    payload: {
      local_part: string
      domain: string
      name: string
      password: string
      quota: number
      active: boolean
      smtp_active: boolean
      send_welcome: boolean
      email_other: string
    },
  ]
}>()

const showPw1 = ref(false)
const showPw2 = ref(false)
const pwdMismatch = ref(false)

const form = ref({
  localPart: '',
  domain: '',
  name: '',
  quotaMB: 1024,
  active: true,
  smtpActive: true,
  sendWelcome: false,
  emailOther: '',
  password: '',
  passwordConfirm: '',
})

const selectedDomain = computed(() => props.selectedDomain?.trim() || '')
const pwdStrength = computed(() => calcStrength(form.value.password))

// Reset form when modal opens
watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      form.value = {
        localPart: '',
        domain: selectedDomain.value || props.domains[0]?.domain || '',
        name: '',
        quotaMB: 1024,
        active: true,
        smtpActive: true,
        sendWelcome: false,
        emailOther: '',
        password: '',
        passwordConfirm: '',
      }
      showPw1.value = false
      showPw2.value = false
      pwdMismatch.value = false
    }
  },
)

watch(selectedDomain, (domain) => {
  if (props.modelValue && domain) {
    form.value.domain = domain
  }
})

function onLocalPartInput(e: Event) {
  form.value.localPart = (e.target as HTMLInputElement).value.toLowerCase()
}

function onPasswordInput() {
  pwdMismatch.value = !!(form.value.passwordConfirm && form.value.password !== form.value.passwordConfirm)
}

function genPassword() {
  generatePassword(form.value)
  showPw1.value = true
  showPw2.value = true
  onPasswordInput()
}

function submit() {
  if (form.value.password !== form.value.passwordConfirm) {
    pwdMismatch.value = true
    return
  }

  emit('submit', {
    local_part: form.value.localPart.toLowerCase().trim(),
    domain: form.value.domain,
    name: form.value.name,
    password: form.value.password,
    quota: form.value.quotaMB > 0 ? form.value.quotaMB * 1048576 : 0,
    active: form.value.active,
    smtp_active: form.value.smtpActive,
    send_welcome: form.value.sendWelcome,
    email_other: form.value.emailOther,
  })
}
</script>
