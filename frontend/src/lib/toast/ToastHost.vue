<template>
  <Teleport to="body">
    <TransitionGroup
      tag="div"
      name="tst"
      class="tst-stack"
      :class="[`tst-${position}`, { 'tst-paused': paused }]"
      aria-label="Notifications"
    >
      <div
        v-for="t in toasts"
        :key="t.id"
        class="tst"
        :class="`tst--${t.type}`"
        :role="t.type === 'error' ? 'alert' : 'status'"
        :aria-live="t.type === 'error' ? 'assertive' : 'polite'"
        @click="closeOnClick && dismiss(t.id)"
      >
        <svg class="tst-icon" viewBox="0 0 24 24" aria-hidden="true">
          <circle cx="12" cy="12" r="10" />
          <path :d="icons[t.type]" />
        </svg>
        <p class="tst-msg">{{ t.message }}</p>
        <button type="button" class="tst-close" :aria-label="closeLabel" @click.stop="dismiss(t.id)">
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12M18 6L6 18" /></svg>
        </button>
        <span
          v-if="t.duration > 0"
          class="tst-bar"
          :style="{ animationDuration: `${t.duration}ms` }"
          @animationend="dismiss(t.id)"
        ></span>
      </div>
    </TransitionGroup>
  </Teleport>
</template>

<script setup lang="ts">
// Renders the toasts of ./toast.ts. Place it once (App.vue). The countdown is the progress bar's CSS animation:
// it pauses on hover/focus and, with pauseOnFocusLoss, while the window is in the background; its end closes the toast.
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { dismiss, toasts, type ToastType } from './toast'

withDefaults(
  defineProps<{
    position?: 'top-right' | 'top-left' | 'top-center' | 'bottom-right' | 'bottom-left' | 'bottom-center'
    closeOnClick?: boolean
    pauseOnFocusLoss?: boolean
    closeLabel?: string
  }>(),
  { position: 'top-right', closeOnClick: true, pauseOnFocusLoss: true, closeLabel: 'Close' },
)

const icons: Record<ToastType, string> = {
  success: 'M8 12.5l2.5 2.5L16 9.5',
  error: 'M9 9l6 6M15 9l-6 6',
  info: 'M12 11v5M12 8h.01',
  warning: 'M12 7v6M12 16h.01',
}

const paused = ref(false)
const blur = () => (paused.value = true)
const focus = () => (paused.value = false)
onMounted(() => {
  window.addEventListener('blur', blur)
  window.addEventListener('focus', focus)
})
onBeforeUnmount(() => {
  window.removeEventListener('blur', blur)
  window.removeEventListener('focus', focus)
})
</script>

<style>
/* Self-contained: every color and size is a --toast-* variable with a neutral default, so the module works as is in
   another project and a design system maps the variables to its own tokens (see README.md). */
.tst-stack {
  position: fixed;
  z-index: var(--toast-z, 9999);
  display: flex;
  flex-direction: column;
  gap: var(--toast-gap, 0.75rem);
  width: min(var(--toast-width, 22.5rem), calc(100vw - 2 * var(--toast-offset, 1rem)));
  pointer-events: none;
}
.tst-top-right {
  top: var(--toast-offset, 1rem);
  right: var(--toast-offset, 1rem);
}
.tst-top-left {
  top: var(--toast-offset, 1rem);
  left: var(--toast-offset, 1rem);
}
.tst-top-center {
  top: var(--toast-offset, 1rem);
  left: 50%;
  transform: translateX(-50%);
}
.tst-bottom-right {
  bottom: var(--toast-offset, 1rem);
  right: var(--toast-offset, 1rem);
}
.tst-bottom-left {
  bottom: var(--toast-offset, 1rem);
  left: var(--toast-offset, 1rem);
}
.tst-bottom-center {
  bottom: var(--toast-offset, 1rem);
  left: 50%;
  transform: translateX(-50%);
}
.tst {
  --tst-bg: var(--toast-info-bg, #1d4ed8);
  --tst-fg: var(--toast-info-fg, #ffffff);
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  min-height: var(--toast-min-height, 3.5rem);
  padding: var(--toast-padding, 0.875rem 1rem 1.125rem);
  overflow: hidden;
  font-family: var(--toast-font, inherit);
  font-size: var(--toast-font-size, 0.9375rem);
  font-weight: var(--toast-font-weight, 700);
  line-height: 1.4;
  color: var(--tst-fg);
  background: var(--tst-bg);
  border: var(--toast-border-width, 0) solid var(--toast-border-color, transparent);
  border-radius: var(--toast-radius, 0.375rem);
  box-shadow: var(--toast-shadow, 0 6px 20px rgba(0, 0, 0, 0.18));
  cursor: pointer;
  pointer-events: auto;
}
.tst--success {
  --tst-bg: var(--toast-success-bg, #15803d);
  --tst-fg: var(--toast-success-fg, #ffffff);
}
.tst--error {
  --tst-bg: var(--toast-error-bg, #b91c1c);
  --tst-fg: var(--toast-error-fg, #ffffff);
}
.tst--warning {
  --tst-bg: var(--toast-warning-bg, #b45309);
  --tst-fg: var(--toast-warning-fg, #ffffff);
}
.tst-icon {
  flex-shrink: 0;
  width: 1.375rem;
  height: 1.375rem;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}
.tst-msg {
  flex: 1;
  margin: 0;
  overflow-wrap: anywhere;
}
.tst-close {
  flex-shrink: 0;
  display: grid;
  place-items: center;
  width: 1.5rem;
  height: 1.5rem;
  margin: -0.125rem -0.25rem 0 0;
  padding: 0;
  color: inherit;
  background: none;
  border: 0;
  border-radius: 0.25rem;
  opacity: 0.8;
  cursor: pointer;
}
.tst-close:hover,
.tst-close:focus-visible {
  opacity: 1;
}
.tst-close:focus-visible {
  outline: 2px solid currentColor;
  outline-offset: 1px;
}
.tst-close svg {
  width: 1rem;
  height: 1rem;
  fill: none;
  stroke: currentColor;
  stroke-width: 2.5;
  stroke-linecap: round;
}
.tst-bar {
  position: absolute;
  left: 0;
  bottom: 0;
  width: 100%;
  height: var(--toast-bar-height, 0.3125rem);
  background: color-mix(in srgb, var(--tst-fg) 40%, transparent);
  transform-origin: left;
  animation: tst-countdown linear forwards;
}
.tst:hover .tst-bar,
.tst:focus-within .tst-bar,
.tst-paused .tst-bar {
  animation-play-state: paused;
}
@keyframes tst-countdown {
  from {
    transform: scaleX(1);
  }
  to {
    transform: scaleX(0);
  }
}
.tst-enter-active,
.tst-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}
.tst-enter-from,
.tst-leave-to {
  opacity: 0;
  transform: translateY(-0.5rem) scale(0.97);
}
.tst-move {
  transition: transform 0.2s ease;
}
@media (prefers-reduced-motion: reduce) {
  .tst-enter-active,
  .tst-leave-active,
  .tst-move {
    transition: none;
  }
}
</style>
