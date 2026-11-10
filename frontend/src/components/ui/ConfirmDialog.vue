<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="modelValue" class="modal-overlay" @click.self="close">
        <div class="bg-surface border-2 border-line w-full max-w-sm flex flex-col">
          <!-- Red danger header -->
          <div class="bg-danger px-5 py-4 flex items-center justify-between flex-shrink-0 border-b-2 border-line">
            <h3 class="text-sm font-mono font-black uppercase tracking-tight text-on-action flex items-center gap-2">
              <Icon name="alert-triangle" :size="16" />
              {{ title }}
            </h3>
            <button @click="close" class="text-on-action hover:text-err-line-soft transition-colors cursor-pointer">
              <Icon name="x" :size="18" />
            </button>
          </div>

          <!-- Body -->
          <div class="px-6 py-5">
            <slot>
              <p class="text-sm font-medium text-fg leading-relaxed">
                Are you sure you want to delete<br />
                <strong class="font-mono break-all">{{ itemName }}</strong
                >?<br />
                <span class="text-xs text-fg-muted mt-1 block">This action cannot be undone.</span>
              </p>
            </slot>
          </div>

          <!-- Footer -->
          <div class="flex items-center justify-end gap-3 px-6 py-4 border-t-2 border-line flex-shrink-0 bg-surface">
            <button
              type="button"
              @click="close"
              class="bg-surface hover:bg-surface-2 text-fg border-2 border-line font-black px-5 py-2.5 shadow-(--shadow-2) transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-(--shadow-3) active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest text-xs flex items-center"
            >
              <Icon name="x" :size="13" class="mr-1.5" />
              CANCEL
            </button>
            <button
              type="button"
              :disabled="loading"
              @click="emit('confirm')"
              class="bg-danger hover:bg-surface hover:text-danger text-on-action border-2 border-line font-black px-5 py-2.5 shadow-(--shadow-3) transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-(--shadow-4) active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest text-xs flex items-center disabled:opacity-60 disabled:cursor-not-allowed"
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

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    itemName?: string
    loading?: boolean
    confirmLabel?: string
  }>(),
  {
    title: 'CONFIRM DELETE',
    itemName: '',
    loading: false,
    confirmLabel: 'DELETE',
  },
)

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
  transition:
    transform 0.18s ease,
    opacity 0.15s ease;
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
