<script setup lang="ts">
/**
 * Neo-brutalist Modal
 *
 * Usage:
 *   <BrutalModal v-model="showAdd" title="ADD DOMAIN" size="md">
 *     ...
 *   </BrutalModal>
 */
import Icon from './Icon.vue'

interface Props {
  modelValue: boolean
  title?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  danger?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'close'): void
}>()

function close() {
  emit('update:modelValue', false)
  emit('close')
}

function onOverlayClick(e: MouseEvent) {
  if (e.target === e.currentTarget) close()
}

// Close on ESC
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.modelValue) close()
}

if (typeof window !== 'undefined') {
  window.addEventListener('keydown', onKeydown)
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="modelValue"
        class="brutal-modal-overlay"
        @click="onOverlayClick"
      >
        <div
          class="brutal-modal-card"
          :class="[
            `brutal-modal--${size}`,
            { 'brutal-modal--danger': danger }
          ]"
          @click.stop
        >
          <!-- Header -->
          <div class="brutal-modal-head" :class="{ 'brutal-modal-head--danger': danger }">
            <div class="flex items-center gap-2">
              <span class="brutal-modal-title">{{ title }}</span>
              <slot name="subtitle" />
            </div>
            <button class="brutal-modal-close" @click="close" aria-label="Close">
              <Icon name="x" :size="14" />
            </button>
          </div>

          <!-- Body -->
          <div class="brutal-modal-body">
            <slot />
          </div>

          <!-- Footer (actions) -->
          <div v-if="$slots.footer" class="brutal-modal-footer">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.brutal-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 20px;
}

.brutal-modal-card {
  background: #ffffff;
  border: 3px solid #1e293b;
  width: 100%;
  max-height: 90vh;
  overflow: auto;
  display: flex;
  flex-direction: column;
}

.brutal-modal--sm { max-width: 420px; }
.brutal-modal--md { max-width: 560px; }
.brutal-modal--lg { max-width: 720px; }
.brutal-modal--xl { max-width: 960px; }

.brutal-modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 3px solid #1e293b;
  background: #f8fafc;
  flex-shrink: 0;
}

.brutal-modal-head--danger {
  background: #fef2f2;
  border-bottom-color: #1e293b;
}

.brutal-modal-title {
  font-size: 13px;
  font-weight: 900;
  letter-spacing: 1px;
  text-transform: uppercase;
  color: #1e293b;
}

.brutal-modal-close {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid #1e293b;
  background: white;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  transition: all 0.1s;
}

.brutal-modal-close:hover {
  background: #fef2f2;
  color: #ef4444;
}

.brutal-modal-body {
  padding: 18px;
  flex: 1;
  overflow: auto;
}

.brutal-modal-footer {
  padding: 14px 18px;
  border-top: 2px solid #e2e8f0;
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  flex-shrink: 0;
  background: #fafafa;
}

/* Transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.15s ease;
}
.modal-enter-active .brutal-modal-card,
.modal-leave-active .brutal-modal-card {
  transition: transform 0.2s ease, opacity 0.15s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-from .brutal-modal-card,
.modal-leave-to .brutal-modal-card {
  transform: translateY(12px) scale(0.985);
  opacity: 0;
}
</style>
