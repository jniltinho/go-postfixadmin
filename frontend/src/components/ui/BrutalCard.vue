<script setup lang="ts">
interface Props {
  padding?: 'none' | 'sm' | 'md' | 'lg'
  border?: boolean
}

withDefaults(defineProps<Props>(), {
  padding: 'md',
  border: true,
})
</script>

<template>
  <div
    class="brutal-card"
    :class="[
      `brutal-card--p-${padding}`,
      { 'brutal-card--border': border }
    ]"
  >
    <slot />
  </div>
</template>

<style scoped>
.brutal-card {
  background: #ffffff;
  box-shadow: 4px 4px 0 #1e293b;
}

.brutal-card--border {
  border: 3px solid #1e293b;
}

.brutal-card--p-none { padding: 0; }
.brutal-card--p-sm { padding: 12px; }
.brutal-card--p-md { padding: 18px; }
.brutal-card--p-lg { padding: 24px; }
</style>
