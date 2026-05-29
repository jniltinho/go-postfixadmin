<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
    <div class="bg-white border-2 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0 border-b-2 border-brand-text">
        <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
          <Icon name="plus-circle" :size="20" class="mr-2" />
          ADD NEW TRANSPORT
        </h3>
        <button @click="emit('update:modelValue', false)" class="text-white hover:text-gray-200 transition-colors">
          <Icon name="x" :size="20" />
        </button>
      </div>

      <!-- Scrollable Body -->
      <div class="overflow-y-auto flex-1">
        <form class="p-6 space-y-5" @submit.prevent="submit">
          <div class="border-2 border-brand-text p-4 space-y-4">
            <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="arrow-up-down" :size="16" class="mr-2" />
              TRANSPORT INFORMATION
            </h4>

            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                DOMAIN <span class="text-red-500">*</span>
              </label>
              <input
                v-model="form.domain"
                type="text"
                required
                placeholder="example.com"
                class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm"
                @input="form.domain = form.domain.toLowerCase()"
              />
              <p class="text-[10px] text-gray-400 mt-1">Domain or pattern for this transport rule</p>
            </div>

            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                TRANSPORT <span class="text-red-500">*</span>
              </label>
              <input
                v-model="form.transport"
                type="text"
                required
                placeholder="smtp:[mail.example.com]:25"
                class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm"
              />
              <p class="text-[10px] text-gray-400 mt-1">Transport definition, e.g. smtp:[host]:port or relay:[host]</p>
            </div>

            <div class="flex items-center">
              <input type="checkbox" id="tp-add-active" v-model="form.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
              <label for="tp-add-active" class="ml-2 text-sm font-bold cursor-pointer">Active</label>
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
        <button type="button" :disabled="saving" @click="submit"
          class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-3 shadow-[3px_3px_0px_#1E293B] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm disabled:opacity-60">
          <Icon name="save" :size="16" class="mr-2" />
          {{ saving ? 'SAVING...' : 'CREATE TRANSPORT' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  modelValue: boolean
  saving: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  submit: [payload: { domain: string; transport: string; active: boolean }]
}>()

const form = ref({ domain: '', transport: '', active: true })

watch(() => props.modelValue, (open) => {
  if (open) form.value = { domain: '', transport: '', active: true }
})

function submit() {
  emit('submit', {
    domain: form.value.domain.trim(),
    transport: form.value.transport.trim(),
    active: form.value.active,
  })
}
</script>
