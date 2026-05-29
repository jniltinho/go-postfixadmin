<script setup lang="ts">
/**
 * Neo-brutalist button
 *
 * Variants:
 * - primary (blue)
 * - danger (red)
 * - ghost (transparent with border)
 */
interface Props {
  variant?: 'primary' | 'danger' | 'ghost' | 'default'
  disabled?: boolean
  type?: 'button' | 'submit' | 'reset'
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  type: 'button',
  size: 'md',
})

const emit = defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()

function onClick(e: MouseEvent) {
  if (!props.disabled) emit('click', e)
}
</script>

<template>
  <button
    :type="type"
    :disabled="disabled"
    class="brutal-btn"
    :class="[
      `brutal-btn--${variant}`,
      `brutal-btn--${size}`,
      { 'brutal-btn--disabled': disabled }
    ]"
    @click="onClick"
  >
    <slot />
  </button>
</template>

<style scoped>
.brutal-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-weight: 700;
  letter-spacing: 0.5px;
  border: 2px solid var(--color-brand-text);
  background: white;
  color: var(--color-brand-text);
  cursor: pointer;
  transition: all 0.1s ease;
  text-transform: uppercase;
  font-size: 12px;
  line-height: 1;
  box-shadow: 2px 2px 0 var(--color-brand-text);
}

.brutal-btn--sm {
  padding: 6px 12px;
  font-size: 11px;
}

.brutal-btn--md {
  padding: 9px 18px;
}

.brutal-btn--lg {
  padding: 12px 24px;
  font-size: 13px;
}

.brutal-btn--primary {
  background: var(--color-brand-primary);
  color: white;
  border-color: var(--color-brand-text);
}

.brutal-btn--primary:hover:not(:disabled) {
  background: #2563eb;
  transform: translate(-1px, -1px);
  box-shadow: 3px 3px 0 var(--color-brand-text);
}

.brutal-btn--danger {
  background: #ef4444;
  color: white;
  border-color: #1e293b;
}

.brutal-btn--danger:hover:not(:disabled) {
  background: #dc2626;
  transform: translate(-1px, -1px);
  box-shadow: 3px 3px 0 #1e293b;
}

.brutal-btn--ghost {
  background: transparent;
  border-color: #64748b;
  color: #475569;
  box-shadow: none;
}

.brutal-btn--ghost:hover:not(:disabled) {
  border-color: var(--color-brand-text);
  color: var(--color-brand-text);
  background: #f8fafc;
}

.brutal-btn--disabled,
.brutal-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
  box-shadow: 1px 1px 0 #94a3b8;
}
</style>
