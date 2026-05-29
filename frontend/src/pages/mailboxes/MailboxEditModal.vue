<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
    <div class="modal-card" style="border: 2px solid #1e293b; max-width: 720px;">
      <!-- Header -->
      <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0" style="border-bottom: 2px solid #1e293b;">
        <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
          <Icon name="edit" :size="20" class="mr-2" />
          EDIT EMAIL ACCOUNT
        </h3>
        <button @click="emit('update:modelValue', false)" class="text-white hover:text-gray-200 transition-colors">
          <Icon name="x" :size="20" />
        </button>
      </div>

      <div class="modal-body">
        <!-- EMAIL ACCOUNT Section -->
        <div class="border-2 border-brand-text p-4 space-y-4">
          <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
            <Icon name="mail" :size="16" class="mr-2" />
            EMAIL ACCOUNT
          </h4>

          <div>
            <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">EMAIL ADDRESS</label>
            <div class="h-10 px-3 flex items-center border-2 border-gray-300 bg-gray-50 font-mono text-sm font-medium text-gray-700">
              {{ form.username }}
            </div>
          </div>

          <div>
            <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">DISPLAY NAME</label>
            <input v-model="form.name" type="text"
              class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm"
              placeholder="Full name" />
          </div>

          <div class="flex items-center">
            <input type="checkbox" id="edit-active" v-model="form.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
            <label for="edit-active" class="ml-2 text-sm font-bold cursor-pointer">Active Account</label>
          </div>
        </div>

        <!-- CHANGE PASSWORD (collapsible) -->
        <details class="border-2 border-brand-text">
          <summary class="px-4 py-3 cursor-pointer font-bold uppercase tracking-tight text-xs flex items-center hover:bg-gray-50 transition-colors">
            <Icon name="key" :size="14" class="mr-2" />
            CHANGE PASSWORD
            <Icon name="chevron-down" :size="14" class="ml-auto" />
          </summary>
          <div class="px-4 pb-4 pt-3 space-y-3 border-t-2 border-brand-text">
            <div class="flex gap-2 items-center">
              <div class="relative flex-1">
                <input v-model="form.password" :type="showPw3 ? 'text' : 'password'"
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm pr-10"
                  placeholder="New password (min 8 chars)"
                  @input="onPasswordInput" />
                <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-brand-primary" @click="showPw3 = !showPw3">
                  <Icon :name="showPw3 ? 'eye-off' : 'eye'" :size="16" />
                </button>
              </div>
              <div class="relative flex-1">
                <input v-model="form.passwordConfirm" :type="showPw4 ? 'text' : 'password'"
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm pr-10"
                  placeholder="Confirm new password"
                  @input="onPasswordInput" />
                <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-brand-primary" @click="showPw4 = !showPw4">
                  <Icon :name="showPw4 ? 'eye-off' : 'eye'" :size="16" />
                </button>
              </div>
              <button type="button"
                class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white text-xs font-black px-3 h-10 border-2 border-brand-text cursor-pointer uppercase tracking-widest flex items-center gap-1 whitespace-nowrap flex-shrink-0 transition-colors"
                @click="genPassword">
                <Icon name="wand-2" :size="14" class="mr-1" />
                GENERATE
              </button>
            </div>

            <div v-if="form.password" class="pwd-strength">
              <div class="flex items-center justify-between mb-1">
                <span class="text-xs font-black uppercase tracking-widest text-brand-text">STRENGTH</span>
                <span class="text-xs font-bold uppercase tracking-wider" :style="{ color: pwdStrength.color }">{{ pwdStrength.label }}</span>
              </div>
              <div class="h-2 bg-gray-200 border border-brand-text overflow-hidden">
                <div class="h-full transition-all duration-300" :style="{ width: pwdStrength.pct + '%', background: pwdStrength.color }"></div>
              </div>
            </div>
            <div v-if="pwdMismatch" class="text-xs text-red-600 font-bold">Passwords do not match</div>
          </div>
        </details>

        <!-- ADVANCED SETTINGS (open by default) -->
        <details class="border-2 border-brand-text" open>
          <summary class="px-4 py-3 cursor-pointer font-bold uppercase tracking-tight text-xs flex items-center hover:bg-gray-50 transition-colors">
            <Icon name="settings" :size="14" class="mr-2" />
            ADVANCED SETTINGS
            <Icon name="chevron-down" :size="14" class="ml-auto" />
          </summary>
          <div class="px-4 pb-4 pt-2 space-y-3 border-t-2 border-brand-text">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">QUOTA (MB)</label>
                <input v-model.number="form.quotaMB" type="number" min="0"
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm" />
                <p class="text-[10px] text-gray-400 mt-1">Storage limit in MB (0 = unlimited)</p>
              </div>
              <div class="flex items-center pt-5">
                <input type="checkbox" id="edit-smtp" v-model="form.smtpActive" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                <label for="edit-smtp" class="ml-2 text-sm font-bold cursor-pointer">SMTP Active (can send email)</label>
              </div>
            </div>
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">ALTERNATIVE EMAIL</label>
              <input v-model="form.emailOther" type="email"
                class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm"
                placeholder="other@example.com" />
            </div>
          </div>
        </details>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-brand-text flex-shrink-0 bg-white">
        <button type="button" @click="emit('update:modelValue', false)"
          class="bg-white hover:bg-gray-50 text-brand-text border-2 border-brand-text font-black px-6 py-3 shadow-[2px_2px_0px_#1E293B] transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm">
          <Icon name="x" :size="16" class="mr-2" />
          CANCEL
        </button>
        <button type="button" @click="submit" :disabled="saving"
          class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-3 shadow-[3px_3px_0px_#1E293B] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm">
          <Icon name="save" :size="16" class="mr-2" />
          {{ saving ? 'SAVING...' : 'UPDATE EMAIL ACCOUNT' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { calcStrength, generatePassword } from '../../utils/password'

interface MailboxEditForm {
  username: string
  name: string
  quotaMB: number
  active: boolean
  smtpActive: boolean
  emailOther: string
  password: string
  passwordConfirm: string
}

const props = defineProps<{
  modelValue: boolean
  initialData: MailboxEditForm | null
  saving: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  submit: [payload: {
    name: string
    quota: number
    active: boolean
    smtp_active: boolean
    email_other: string
    change_password?: boolean
    password?: string
    password_confirm?: string
  }]
}>()

const showPw3 = ref(false)
const showPw4 = ref(false)
const pwdMismatch = ref(false)

const form = ref<MailboxEditForm>({
  username: '',
  name: '',
  quotaMB: 0,
  active: true,
  smtpActive: true,
  emailOther: '',
  password: '',
  passwordConfirm: '',
})

const pwdStrength = computed(() => calcStrength(form.value.password))

// Populate form when initialData changes (modal opens)
watch(() => props.initialData, (data) => {
  if (data) {
    form.value = { ...data }
    showPw3.value = false
    showPw4.value = false
    pwdMismatch.value = false
  }
}, { immediate: true })

function onPasswordInput() {
  pwdMismatch.value = !!(form.value.passwordConfirm && form.value.password !== form.value.passwordConfirm)
}

function genPassword() {
  generatePassword(form.value)
  showPw3.value = true
  showPw4.value = true
  onPasswordInput()
}

function submit() {
  const payload: any = {
    name: form.value.name,
    quota: form.value.quotaMB,
    active: form.value.active,
    smtp_active: form.value.smtpActive,
    email_other: form.value.emailOther,
  }
  if (form.value.password) {
    payload.change_password = true
    payload.password = form.value.password
    payload.password_confirm = form.value.passwordConfirm
  }
  emit('submit', payload)
}
</script>
