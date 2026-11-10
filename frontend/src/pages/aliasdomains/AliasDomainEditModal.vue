<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
    <div class="bg-surface border-2 border-line w-full max-w-2xl max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="bg-action px-6 py-4 flex items-center justify-between flex-shrink-0">
        <h3 class="text-lg font-mono font-black uppercase tracking-tight text-on-action flex items-center">
          <Icon name="edit" :size="20" class="mr-2" />
          EDIT DOMAIN ALIAS
          <span class="ml-2 text-on-action/80 text-base font-mono">- {{ form.alias_domain }}</span>
        </h3>
        <button
          @click="emit('update:modelValue', false)"
          class="text-on-action hover:text-on-action/80 transition-colors"
        >
          <Icon name="x" :size="20" />
        </button>
      </div>

      <!-- Body -->
      <div class="overflow-y-auto flex-1">
        <div v-if="loading" class="flex items-center justify-center p-12">
          <div class="flex flex-col items-center">
            <Icon name="loader-2" :size="28" class="text-icon-accent animate-spin mb-2" />
            <span class="text-sm font-bold uppercase tracking-widest text-fg">LOADING...</span>
          </div>
        </div>

        <div v-else>
          <form class="p-6 space-y-5" @submit.prevent="submit">
            <div class="border-2 border-line p-4 space-y-4">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="info" :size="16" class="mr-2" />
                ALIAS DOMAIN INFORMATION
              </h4>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">ALIAS DOMAIN</label>
                  <input
                    :value="form.alias_domain"
                    readonly
                    class="w-full h-10 px-3 border-2 border-field-line bg-surface-2 font-medium font-mono text-sm cursor-not-allowed"
                  />
                  <p class="text-[0.625rem] text-fg-muted mt-1">Alias domain cannot be changed after creation</p>
                </div>
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">
                    TARGET DOMAIN <span class="text-err">*</span>
                  </label>
                  <select
                    v-model="form.target_domain"
                    required
                    class="w-full h-10 px-3 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors bg-surface cursor-pointer text-sm"
                  >
                    <option value="" disabled>Select target domain...</option>
                    <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
                  </select>
                  <p class="text-[0.625rem] text-fg-muted mt-1">Domain that receives the mail</p>
                </div>
              </div>

              <div class="flex items-center pt-2">
                <input
                  type="checkbox"
                  id="ad-edit-active"
                  v-model="form.active"
                  class="w-5 h-5 border-2 border-line cursor-pointer"
                />
                <label for="ad-edit-active" class="ml-2 text-sm font-bold cursor-pointer">Active Domain Alias</label>
              </div>
            </div>
          </form>
        </div>
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
          :disabled="saving || loading"
          @click="submit"
          class="bg-action hover:bg-surface hover:text-icon-accent text-on-action border-2 border-line font-black px-6 py-2.5 shadow-(--shadow-3) hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-(--shadow-4) active:translate-x-0 active:translate-y-0 active:shadow-none transition-all uppercase tracking-widest text-sm flex items-center disabled:opacity-60"
        >
          <Icon name="save" :size="16" class="mr-2" />
          {{ saving ? 'SAVING...' : 'UPDATE ALIAS' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface Domain {
  domain: string
}
interface AliasDomainEditForm {
  alias_domain: string
  target_domain: string
  active: boolean
}

const props = defineProps<{
  modelValue: boolean
  domains: Domain[]
  initialData: AliasDomainEditForm | null
  saving: boolean
  loading: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  submit: [payload: { target_domain: string; active: boolean }]
}>()

const form = ref<AliasDomainEditForm>({ alias_domain: '', target_domain: '', active: true })

watch(
  () => props.initialData,
  (data) => {
    if (data) form.value = { ...data }
  },
  { immediate: true },
)

function submit() {
  emit('submit', { target_domain: form.value.target_domain, active: form.value.active })
}
</script>
