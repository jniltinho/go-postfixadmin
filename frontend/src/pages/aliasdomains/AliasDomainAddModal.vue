<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
    <div class="bg-white border-2 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0">
        <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
          <Icon name="plus-circle" :size="20" class="mr-2" />
          ADD DOMAIN ALIAS
        </h3>
        <button @click="emit('update:modelValue', false)" class="text-white hover:text-gray-200 transition-colors">
          <Icon name="x" :size="20" />
        </button>
      </div>

      <!-- Body -->
      <div class="overflow-y-auto flex-1">
        <form class="p-6 space-y-5" @submit.prevent="submit">
          <div class="border-2 border-brand-text p-4 space-y-4">
            <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="info" :size="16" class="mr-2" />
              ALIAS DOMAIN INFORMATION
            </h4>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                  ALIAS DOMAIN <span class="text-red-500">*</span>
                </label>
                <select v-model="form.alias_domain" required
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors bg-white cursor-pointer text-sm">
                  <option value="" disabled>Select source domain...</option>
                  <option v-for="d in availableAlias" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
                </select>
                <p class="text-[10px] text-gray-500 mt-1">Domain that will be the alias (source)</p>
              </div>
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                  TARGET DOMAIN <span class="text-red-500">*</span>
                </label>
                <select v-model="form.target_domain" required
                  class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors bg-white cursor-pointer text-sm">
                  <option value="" disabled>Select target domain...</option>
                  <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
                </select>
                <p class="text-[10px] text-gray-500 mt-1">Domain that will receive the mail (target)</p>
              </div>
            </div>

            <div class="flex items-center pt-2">
              <input type="checkbox" id="ad-add-active" v-model="form.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
              <label for="ad-add-active" class="ml-2 text-sm font-bold cursor-pointer">Active Domain Alias</label>
            </div>
          </div>
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
          {{ saving ? 'SAVING...' : 'CREATE ALIAS' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface Domain { domain: string }

const props = defineProps<{
  modelValue: boolean
  domains: Domain[]
  availableAlias: Domain[]
  saving: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  submit: [payload: { alias_domain: string; target_domain: string; active: boolean }]
}>()

const form = ref({ alias_domain: '', target_domain: '', active: true })

watch(() => props.modelValue, (open) => {
  if (open) form.value = { alias_domain: '', target_domain: '', active: true }
})

function submit() {
  emit('submit', { ...form.value })
}
</script>
