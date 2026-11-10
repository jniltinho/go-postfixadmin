<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
    <div class="bg-surface border-2 border-line w-full max-w-2xl max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="bg-action px-6 py-4 flex items-center justify-between flex-shrink-0">
        <h3 class="text-lg font-mono font-black uppercase tracking-tight text-on-action flex items-center">
          <Icon name="plus-circle" :size="20" class="mr-2" />
          ADD ALIAS
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
          <div class="border-2 border-line p-4 space-y-4">
            <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="info" :size="16" class="mr-2" />
              ALIAS INFORMATION
            </h4>

            <!-- Alias + Domain -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 items-start">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">
                  ALIAS <span class="text-err">*</span>
                </label>
                <input
                  v-model="form.localPart"
                  type="text"
                  required
                  placeholder="alias-name"
                  class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium transition-colors text-sm lowercase"
                  style="text-transform: lowercase"
                  @input="onLocalPartInput"
                />
                <p class="text-[0.625rem] text-fg-muted mt-1">Local part of the alias address</p>
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
                  class="w-full h-10 px-3 border-2 border-line focus:border-action focus:outline-none font-medium transition-colors bg-surface cursor-pointer text-sm"
                >
                  <option value="" disabled>Select a domain...</option>
                  <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
                </select>
              </div>
            </div>

            <!-- Preview -->
            <div class="bg-surface-2 border border-field-line px-3 h-10 flex items-center gap-2">
              <span class="text-xs font-bold uppercase tracking-widest text-fg-subtle whitespace-nowrap"
                >FULL ADDRESS:</span
              >
              <p class="font-mono text-sm font-bold">
                <span class="text-action">{{ form.localPart || 'alias' }}</span
                >@<span class="text-action">{{ form.domain || 'domain.com' }}</span>
              </p>
            </div>

            <!-- Goto / Recipients -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">
                TO (RECIPIENTS) <span class="text-err">*</span>
              </label>
              <textarea
                v-model="form.goto"
                rows="4"
                required
                placeholder="recipient@example.com&#10;another@example.com"
                class="w-full px-3 py-2 border-2 border-line focus:border-action focus:outline-none font-medium transition-colors resize-y text-sm"
              ></textarea>
              <p class="text-[0.625rem] text-fg-muted mt-1">
                One email address per line. Multiple recipients are supported.
              </p>
            </div>

            <!-- Active -->
            <div class="flex items-center pt-1">
              <input
                type="checkbox"
                id="alias-add-active"
                v-model="form.active"
                class="w-5 h-5 border-2 border-line cursor-pointer"
              />
              <label for="alias-add-active" class="ml-2 text-sm font-bold cursor-pointer">Active Alias</label>
            </div>
          </div>
        </form>
      </div>

      <!-- Footer -->
      <div class="bg-surface-2 px-6 py-4 flex items-center justify-end space-x-3 border-t-2 border-line flex-shrink-0">
        <button
          type="button"
          @click="emit('update:modelValue', false)"
          class="bg-surface hover:bg-surface-3 text-fg border-2 border-line font-black px-6 py-2.5 shadow-[2px_2px_0px_var(--color-shadow)] hover:translate-y-px hover:shadow-[1px_1px_0px_var(--color-shadow)] active:translate-y-0.5 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center"
        >
          <Icon name="x" :size="16" class="mr-2" />
          CANCEL
        </button>
        <button
          type="button"
          :disabled="saving"
          @click="submit"
          class="bg-action hover:bg-surface hover:text-action text-on-action border-2 border-line font-black px-6 py-2.5 shadow-[3px_3px_0px_var(--color-shadow)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_var(--color-shadow)] active:translate-x-0 active:translate-y-0 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center disabled:opacity-60"
        >
          <Icon name="plus-circle" :size="16" class="mr-2" />
          {{ saving ? 'SAVING...' : 'CREATE ALIAS' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

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
      goto: string
      active: boolean
    },
  ]
}>()

const selectedDomain = computed(() => props.selectedDomain?.trim() || '')
const form = ref({ localPart: '', domain: '', goto: '', active: true })

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      form.value = {
        localPart: '',
        domain: selectedDomain.value || props.domains[0]?.domain || '',
        goto: '',
        active: true,
      }
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

function submit() {
  emit('submit', {
    local_part: form.value.localPart.toLowerCase().trim(),
    domain: form.value.domain,
    goto: form.value.goto.trim(),
    active: form.value.active,
  })
}
</script>
