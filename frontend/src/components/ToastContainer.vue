<script setup lang="ts">
import { ToastifyContainer } from 'vue3-toastify'
import 'vue3-toastify/dist/index.css'
</script>

<template>
  <ToastifyContainer
    position="top-right"
    theme="colored"
    transition="zoom"
    :auto-close="3000"
    :newest-on-top="true"
    :pause-on-hover="true"
    :pause-on-focus-loss="true"
    :close-on-click="true"
    role="alert"
  />
</template>

<style>
.Toastify__toast-container {
  width: 340px;
}

.Toastify__toast {
  min-height: 100px;
  border-radius: 0;
}
</style>
