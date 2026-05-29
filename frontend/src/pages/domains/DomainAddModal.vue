<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
    <div class="bg-white border-2 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0">
        <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
          <Icon name="plus-circle" :size="20" class="mr-2" />
          ADD DOMAIN
        </h3>
        <button @click="emit('update:modelValue', false)" class="text-white hover:text-gray-200 transition-colors">
          <Icon name="x" :size="20" />
        </button>
      </div>

      <!-- Scrollable Body -->
      <div class="overflow-y-auto flex-1">
        <form class="p-6 space-y-5" @submit.prevent="submit">
          <!-- Basic Information -->
          <div class="border-2 border-brand-text p-4 space-y-4">
            <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="info" :size="16" class="mr-2" />
              BASIC INFORMATION
            </h4>

            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                DOMAIN NAME <span class="text-red-500">*</span>
              </label>
              <input v-model="form.domain" type="text" required placeholder="example.com"
                class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
              <p class="text-[10px] text-gray-500 mt-1">Enter a valid domain name (e.g., example.com)</p>
            </div>

            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">DESCRIPTION</label>
              <input v-model="form.description" type="text" placeholder="Optional description for this domain"
                class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
            </div>

            <div class="flex items-center pt-1">
              <input type="checkbox" id="dom-add-active" v-model="form.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
              <label for="dom-add-active" class="ml-2 text-sm font-bold cursor-pointer">Active Domain</label>
            </div>
          </div>

          <!-- Advanced Settings -->
          <details class="border-2 border-brand-text group" open>
            <summary class="p-3 cursor-pointer font-bold uppercase tracking-tight text-sm flex items-center bg-gray-50 hover:bg-gray-100 transition-colors">
              <Icon name="settings" :size="16" class="mr-2" />
              ADVANCED SETTINGS
              <Icon name="chevron-down" :size="16" class="ml-auto group-open:rotate-180 transition-transform" />
            </summary>
            <div class="p-4 space-y-4 border-t-2 border-brand-text">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">ALIAS LIMIT</label>
                  <input v-model.number="form.aliases" type="number" min="0"
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                  <p class="text-[10px] text-gray-500 mt-1">Maximum number of aliases (0 = unlimited)</p>
                </div>
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">MAILBOX LIMIT</label>
                  <input v-model.number="form.mailboxes" type="number" min="0"
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                  <p class="text-[10px] text-gray-500 mt-1">Maximum number of mailboxes (0 = unlimited)</p>
                </div>
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">QUOTA LIMIT (MB)</label>
                  <input v-model.number="form.quotaMB" type="number" min="0"
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                  <p class="text-[10px] text-gray-500 mt-1">Maximum quota limit in MB (0 = unlimited)</p>
                </div>
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">PASSWORD EXPIRY (DAYS)</label>
                  <input v-model.number="form.passwordExpiry" type="number" min="0" placeholder="365"
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm" />
                  <p class="text-[10px] text-gray-500 mt-1">Password expiry in days (empty = never)</p>
                </div>
              </div>

              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">TRANSPORT</label>
                <div class="relative">
                  <select v-model="form.transport"
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm appearance-none bg-white cursor-pointer pr-10">
                    <option value="virtual">virtual</option>
                    <option v-for="t in transports" :key="t.id" :value="t.transport">{{ t.domain }} → {{ t.transport }}</option>
                  </select>
                  <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-3 text-brand-text border-l-2 border-brand-text">
                    <Icon name="chevron-down" :size="16" />
                  </div>
                </div>
                <p class="text-[10px] text-gray-500 mt-1">Select the mail transport for this domain</p>
              </div>

              <div class="flex items-center pt-2">
                <input type="checkbox" id="dom-add-backupmx" v-model="form.backupmx" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                <label for="dom-add-backupmx" class="ml-2 text-sm font-bold cursor-pointer">Enable Backup MX</label>
                <span class="ml-auto text-[10px] text-gray-500 hidden sm:block">Use this server as a backup mail exchanger</span>
              </div>
            </div>
          </details>
        </form>
      </div>

      <!-- Footer -->
      <div class="bg-gray-50 px-6 py-4 flex items-center justify-end space-x-3 border-t-2 border-brand-text flex-shrink-0">
        <button type="button" @click="emit('update:modelValue', false)"
          class="bg-white hover:bg-gray-100 text-brand-text border-2 border-brand-text font-black px-6 py-2.5 shadow-[2px_2px_0px_#1E293B] hover:translate-y-px hover:shadow-[1px_1px_0px_#1E293B] active:translate-y-0.5 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center">
          <Icon name="x" :size="16" class="mr-2" />
          CANCEL
        </button>
        <button type="button" :disabled="saving" @click="submit"
          class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-2.5 shadow-[3px_3px_0px_#1E293B] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center disabled:opacity-60">
          <Icon name="plus-circle" :size="16" class="mr-2" />
          {{ saving ? 'SAVING...' : 'SAVE DOMAIN' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface Transport { id: number; domain: string; transport: string; active: boolean }

const props = defineProps<{
  modelValue: boolean
  transports: Transport[]
  saving: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  submit: [payload: {
    domain: string
    description: string
    aliases: number
    mailboxes: number
    quota: number
    transport: string
    backupmx: boolean
    active: boolean
    password_expiry: number | null
  }]
}>()

const form = ref({
  domain: '', description: '', aliases: 10, mailboxes: 10,
  quotaMB: 2048, passwordExpiry: 365 as number | null,
  transport: 'virtual', active: true, backupmx: false,
})

watch(() => props.modelValue, (open) => {
  if (open) {
    form.value = {
      domain: '', description: '', aliases: 10, mailboxes: 10,
      quotaMB: 2048, passwordExpiry: 365,
      transport: 'virtual', active: true, backupmx: false,
    }
  }
})

function submit() {
  emit('submit', {
    domain: form.value.domain.trim().toLowerCase(),
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
