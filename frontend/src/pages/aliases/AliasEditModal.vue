<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
    <div class="bg-surface border-2 border-line w-full max-w-2xl max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="bg-action px-6 py-4 flex items-center justify-between flex-shrink-0">
        <h3 class="text-lg font-mono font-black uppercase tracking-tight text-on-action flex items-center">
          <Icon name="edit" :size="20" class="mr-2" />
          EDIT ALIAS
          <span class="ml-2 text-on-action/80 text-base font-mono">- {{ form.address }}</span>
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
        <form class="p-6 space-y-5" @submit.prevent="submit">
          <div class="border-2 border-line p-4 space-y-4">
            <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="info" :size="16" class="mr-2" />
              ALIAS INFORMATION
            </h4>

            <!-- Read-only address -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">ALIAS ADDRESS</label>
              <input
                :value="form.address"
                readonly
                class="w-full h-10 px-3 border-2 border-field-line bg-surface-2 font-medium font-mono text-sm cursor-not-allowed"
              />
              <p class="text-[0.625rem] text-fg-muted mt-1">Alias address cannot be changed after creation</p>
            </div>

            <!-- Goto -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-fg mb-1">
                TO (RECIPIENTS) <span class="text-err">*</span>
              </label>
              <textarea
                v-model="form.goto"
                rows="4"
                required
                placeholder="recipient@example.com"
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
                id="alias-edit-active"
                v-model="form.active"
                class="w-5 h-5 border-2 border-line cursor-pointer"
              />
              <label for="alias-edit-active" class="ml-2 text-sm font-bold cursor-pointer">Active Alias</label>
            </div>
          </div>
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
          {{ saving ? 'SAVING...' : 'UPDATE ALIAS' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface AliasEditForm {
  address: string
  goto: string
  active: boolean
}

const props = defineProps<{
  modelValue: boolean
  initialData: AliasEditForm | null
  saving: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  submit: [payload: { goto: string; active: boolean }]
}>()

const form = ref<AliasEditForm>({ address: '', goto: '', active: true })

watch(
  () => props.initialData,
  (data) => {
    if (data) form.value = { ...data }
  },
  { immediate: true },
)

function submit() {
  emit('submit', {
    goto: form.value.goto.trim(),
    active: form.value.active,
  })
}
</script>
