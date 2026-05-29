<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="modelValue" class="modal-overlay" @click.self="close">
        <div class="bg-white border-2 border-brand-text w-full max-w-sm flex flex-col">

          <!-- Red danger header -->
          <div class="bg-red-600 px-5 py-4 flex items-center justify-between flex-shrink-0 border-b-2 border-brand-text">
            <h3 class="text-sm font-mono font-black uppercase tracking-tight text-white flex items-center gap-2">
              <Icon name="alert-triangle" :size="16" />
              {{ title }}
            </h3>
            <button @click="close" class="text-white hover:text-red-200 transition-colors cursor-pointer">
              <Icon name="x" :size="18" />
            </button>
          </div>

          <!-- Body -->
          <div class="px-6 py-5">
            <slot>
              <p class="text-sm font-medium text-brand-text leading-relaxed">
                Are you sure you want to delete<br />
                <strong class="font-mono break-all">{{ itemName }}</strong>?<br />
                <span class="text-xs text-gray-500 mt-1 block">This action cannot be undone.</span>
              </p>
            </slot>
          </div>

          <!-- Footer -->
          <div class="flex items-center justify-end gap-3 px-6 py-4 border-t-2 border-brand-text flex-shrink-0 bg-white">
            <button
              type="button"
              @click="close"
              class="bg-white hover:bg-gray-50 text-brand-text border-2 border-brand-text font-black px-5 py-2.5 shadow-[2px_2px_0px_#1E293B] transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest text-xs flex items-center"
            >
              <Icon name="x" :size="13" class="mr-1.5" />
              CANCEL
            </button>
            <button
              type="button"
              :disabled="loading"
              @click="emit('confirm')"
              class="bg-red-600 hover:bg-white hover:text-red-600 text-white border-2 border-brand-text font-black px-5 py-2.5 shadow-[3px_3px_0px_#1E293B] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest text-xs flex items-center disabled:opacity-60 disabled:cursor-not-allowed"
            >
              <Icon name="trash-2" :size="13" class="mr-1.5" />
              {{ loading ? 'DELETING...' : confirmLabel }}
            </button>
          </div>

        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import Icon from './Icon.vue'

const props = withDefaults(defineProps<{
  modelValue: boolean
  title?: string
  itemName?: string
  loading?: boolean
  confirmLabel?: string
}>(), {
  title: 'CONFIRM DELETE',
  itemName: '',
  loading: false,
  confirmLabel: 'DELETE',
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  confirm: []
}>()

function close() {
  emit('update:modelValue', false)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.modelValue) close()
}

if (typeof window !== 'undefined') {
  window.addEventListener('keydown', onKeydown)
}
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.15s ease;
}
.modal-enter-active > div,
.modal-leave-active > div {
  transition: transform 0.18s ease, opacity 0.15s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-from > div,
.modal-leave-to > div {
  transform: translateY(10px) scale(0.98);
  opacity: 0;
}
</style>
