<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
    <div class="bg-surface border-2 border-line w-full max-w-2xl max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="bg-action px-6 py-4 flex items-center justify-between flex-shrink-0">
        <h3 class="text-lg font-mono font-black uppercase tracking-tight text-on-action flex items-center">
          <Icon name="edit" :size="20" class="mr-2" />
          EDIT DOMAIN
          <span class="ml-2 text-on-action/80 text-base font-mono">- {{ form.domain }}</span>
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
          <!-- Basic Information -->
          <div class="border-2 border-line p-4 space-y-4">
            <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="info" :size="16" class="mr-2" />
              BASIC INFORMATION
            </h4>

            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">DOMAIN NAME</label>
              <input
                :value="form.domain"
                readonly
                class="w-full h-10 px-3 border-2 border-field-line bg-surface-3 text-fg-muted cursor-not-allowed font-medium transition-colors text-sm"
              />
              <p class="text-[0.625rem] text-fg-muted mt-1">Domain name cannot be changed after creation</p>
            </div>

            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">DESCRIPTION</label>
              <input
                v-model="form.description"
                type="text"
                placeholder="Optional description for this domain"
                class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium transition-colors text-sm"
              />
            </div>

            <div class="flex items-center pt-1">
              <input
                type="checkbox"
                id="dom-edit-active"
                v-model="form.active"
                class="w-5 h-5 border-2 border-line cursor-pointer"
              />
              <label for="dom-edit-active" class="ml-2 text-sm font-bold cursor-pointer">Active Domain</label>
            </div>
          </div>

          <!-- Advanced Settings: visible if domain_adset:read or domain_adset:write; editable only with domain_adset:write -->
          <details v-if="canViewAdvanced" class="border-2 border-line group" open>
            <summary
              class="p-3 cursor-pointer font-bold uppercase tracking-tight text-sm flex items-center bg-surface-2 hover:bg-surface-3 transition-colors"
            >
              <Icon name="settings" :size="16" class="mr-2" />
              ADVANCED SETTINGS
              <Icon name="chevron-down" :size="16" class="ml-auto group-open:rotate-180 transition-transform" />
            </summary>
            <div
              class="p-4 space-y-4 border-t-2 border-line"
              :class="{ 'opacity-60 pointer-events-none': !canEditAdvanced }"
            >
              <div
                v-if="!canEditAdvanced"
                class="bg-warn-soft border border-warn-line p-2 text-xs font-bold text-warn-strong flex items-center gap-2"
              >
                <Icon name="lock" :size="13" /> Read-only — you do not have permission to edit advanced settings.
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">ALIAS LIMIT</label>
                  <input
                    v-model.number="form.aliases"
                    type="number"
                    min="0"
                    :disabled="!canEditAdvanced"
                    class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium transition-colors text-sm disabled:bg-surface-3 disabled:cursor-not-allowed"
                  />
                  <p class="text-[0.625rem] text-fg-muted mt-1">Maximum number of aliases (0 = unlimited)</p>
                </div>
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">MAILBOX LIMIT</label>
                  <input
                    v-model.number="form.mailboxes"
                    type="number"
                    min="0"
                    :disabled="!canEditAdvanced"
                    class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium transition-colors text-sm disabled:bg-surface-3 disabled:cursor-not-allowed"
                  />
                  <p class="text-[0.625rem] text-fg-muted mt-1">Maximum number of mailboxes (0 = unlimited)</p>
                </div>
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1"
                    >QUOTA LIMIT (MB)</label
                  >
                  <input
                    v-model.number="form.quotaMB"
                    type="number"
                    min="0"
                    :disabled="!canEditAdvanced"
                    class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium transition-colors text-sm disabled:bg-surface-3 disabled:cursor-not-allowed"
                  />
                  <p class="text-[0.625rem] text-fg-muted mt-1">Maximum quota limit in MB (0 = unlimited)</p>
                </div>
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1"
                    >PASSWORD EXPIRY (DAYS)</label
                  >
                  <input
                    v-model.number="form.passwordExpiry"
                    type="number"
                    min="0"
                    placeholder="365"
                    :disabled="!canEditAdvanced"
                    class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium transition-colors text-sm disabled:bg-surface-3 disabled:cursor-not-allowed"
                  />
                  <p class="text-[0.625rem] text-fg-muted mt-1">Password expiry in days (empty = never)</p>
                </div>
              </div>

              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">TRANSPORT</label>
                <div class="relative">
                  <select
                    v-model="form.transport"
                    :disabled="!canEditAdvanced"
                    class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium transition-colors text-sm appearance-none bg-surface pr-10 disabled:bg-surface-3 disabled:cursor-not-allowed"
                    :class="canEditAdvanced ? 'cursor-pointer' : 'cursor-not-allowed'"
                  >
                    <option value="virtual">virtual</option>
                    <option v-for="t in transports" :key="t.id" :value="t.transport">
                      {{ t.domain }} → {{ t.transport }}
                    </option>
                  </select>
                  <div
                    class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-3 text-fg border-l-2 border-line"
                  >
                    <Icon name="chevron-down" :size="16" />
                  </div>
                </div>
                <p class="text-[0.625rem] text-fg-muted mt-1">Select the mail transport for this domain</p>
              </div>

              <div class="flex items-center pt-2">
                <input
                  type="checkbox"
                  id="dom-edit-backupmx"
                  v-model="form.backupmx"
                  :disabled="!canEditAdvanced"
                  class="w-5 h-5 border-2 border-line"
                  :class="canEditAdvanced ? 'cursor-pointer' : 'cursor-not-allowed'"
                />
                <label
                  for="dom-edit-backupmx"
                  class="ml-2 text-sm font-bold"
                  :class="canEditAdvanced ? 'cursor-pointer' : 'cursor-not-allowed'"
                  >Enable Backup MX</label
                >
                <span class="ml-auto text-[0.625rem] text-fg-muted hidden sm:block"
                  >Use this server as a backup mail exchanger</span
                >
              </div>
            </div>
          </details>
        </form>
      </div>

      <!-- Footer -->
      <div class="bg-surface-2 px-6 py-4 flex items-center justify-end space-x-3 border-t-2 border-line flex-shrink-0">
        <button
          type="button"
          @click="emit('update:modelValue', false)"
          class="bg-surface hover:bg-surface-3 text-fg border-2 border-line font-black px-6 py-2.5 shadow-(--shadow-2) hover:translate-y-px hover:shadow-(--shadow-1) active:translate-y-0.5 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center"
        >
          <Icon name="x" :size="16" class="mr-2" />
          CANCEL
        </button>
        <button
          type="button"
          :disabled="saving"
          @click="submit"
          class="bg-action hover:bg-surface hover:text-action text-on-action border-2 border-line font-black px-6 py-2.5 shadow-(--shadow-3) hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-(--shadow-4) active:translate-x-0 active:translate-y-0 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center disabled:opacity-60"
        >
          <Icon name="save" :size="16" class="mr-2" />
          {{ saving ? 'SAVING...' : 'UPDATE DOMAIN' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface Transport {
  id: number
  domain: string
  transport: string
  active: boolean
}

interface DomainEditForm {
  domain: string
  description: string
  aliases: number
  mailboxes: number
  quotaMB: number
  passwordExpiry: number | null
  transport: string
  active: boolean
  backupmx: boolean
}

const props = defineProps<{
  modelValue: boolean
  initialData: DomainEditForm | null
  transports: Transport[]
  saving: boolean
  canViewAdvanced: boolean
  canEditAdvanced: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  submit: [
    payload: {
      description: string
      aliases: number
      mailboxes: number
      quota: number
      transport: string
      backupmx: boolean
      active: boolean
      password_expiry: number | null
    },
  ]
}>()

const form = ref<DomainEditForm>({
  domain: '',
  description: '',
  aliases: 10,
  mailboxes: 10,
  quotaMB: 2048,
  passwordExpiry: null,
  transport: 'virtual',
  active: true,
  backupmx: false,
})

watch(
  () => props.initialData,
  (data) => {
    if (data) form.value = { ...data }
  },
  { immediate: true },
)

function submit() {
  emit('submit', {
    description: form.value.description,
    aliases: form.value.aliases,
    mailboxes: form.value.mailboxes,
    quota: form.value.quotaMB,
    transport: form.value.transport || 'virtual',
    backupmx: form.value.backupmx,
    active: form.value.active,
    password_expiry: form.value.passwordExpiry || null,
  })
}
</script>
