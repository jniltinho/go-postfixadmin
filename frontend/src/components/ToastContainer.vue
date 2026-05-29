<script setup lang="ts">
import { useToastStore } from '../stores/toast'

const toast = useToastStore()
</script>

<template>
  <Teleport to="body">
    <div class="toast-stack" aria-live="polite" aria-atomic="false">
      <TransitionGroup name="toast">
        <div
          v-for="t in toast.toasts"
          :key="t.id"
          class="toast"
          :class="`toast--${t.type}`"
          role="alert"
        >
          <svg class="toast-icon" width="16" height="16" viewBox="0 0 16 16" fill="none">
            <template v-if="t.type === 'success'">
              <circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/>
              <path d="M5 8l2 2 4-4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </template>
            <template v-else-if="t.type === 'error'">
              <circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/>
              <path d="M5.5 5.5l5 5M10.5 5.5l-5 5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </template>
            <template v-else-if="t.type === 'warning'">
              <path d="M8 2L14.5 13H1.5L8 2z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/>
              <path d="M8 6v3.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              <circle cx="8" cy="11" r=".6" fill="currentColor"/>
            </template>
            <template v-else>
              <circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/>
              <path d="M8 7v4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              <circle cx="8" cy="5" r=".6" fill="currentColor"/>
            </template>
          </svg>

          <span class="toast-msg">{{ t.message }}</span>

          <button class="toast-close" @click="toast.remove(t.id)" aria-label="Dismiss">
            <Icon name="x" :size="12" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-stack {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9000;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;
  width: 300px;
}

.toast {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-left-width: 3px;
  box-shadow: 4px 4px 0 rgba(30, 41, 59, .08);
  pointer-events: all;
  font-size: 13px;
  font-family: var(--font-sans);
  color: var(--color-brand-text);
}

.toast--success { border-left-color: var(--toast-success); }
.toast--error   { border-left-color: var(--toast-error); }
.toast--warning { border-left-color: var(--toast-warning); }
.toast--info    { border-left-color: var(--color-brand-primary); }

.toast-icon { flex-shrink: 0; }
.toast--success .toast-icon { color: var(--toast-success); }
.toast--error   .toast-icon { color: var(--toast-error); }
.toast--warning .toast-icon { color: var(--toast-warning); }
.toast--info    .toast-icon { color: var(--color-brand-primary); }

.toast-msg {
  flex: 1;
  line-height: 1.4;
  word-break: break-word;
}

.toast-close {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background: none;
  border: none;
  cursor: pointer;
  color: #94a3b8;
  padding: 0;
  transition: color .12s;
}
.toast-close:hover { color: var(--color-brand-text); }

/* Transition */
.toast-enter-active,
.toast-leave-active {
  transition: opacity 220ms ease, transform 220ms ease;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-12px);
}
</style>
