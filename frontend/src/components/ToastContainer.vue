<script setup lang="ts">
import { ToastifyContainer } from 'vue3-toastify'
import 'vue3-toastify/dist/index.css'
</script>

<template>
  <ToastifyContainer
    position="top-right"
    theme="colored"
    transition="zoom"
    :auto-close="4000"
    :newest-on-top="true"
    :pause-on-hover="true"
    :pause-on-focus-loss="true"
    :close-on-click="true"
    role="alert"
  />
</template>

<style>
.Toastify__toast-container {
  z-index: 99999;
  width: 360px;
}

.Toastify__toast {
  width: 340px;
  min-height: 100px;
  padding: 14px 16px;
  font-family: var(--font-sans);
  font-size: 13px;
  font-weight: 700;
  border: 2px solid var(--color-brand-text);
  border-radius: 0;
  box-shadow: 4px 4px 0 rgba(30, 41, 59, 0.25);
}

.Toastify__toast-body {
  line-height: 1.35;
}

.Toastify__close-button {
  opacity: 0.9;
}
</style>
