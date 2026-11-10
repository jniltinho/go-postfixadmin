<script setup lang="ts">
/**
 * Icon wrapper for @lucide/vue
 *
 * Usage:
 *   <Icon name="mail" :size="18" />
 *   <Icon name="trash-2" :size="16" class="text-danger" />
 *
 * All lucide icon names are supported (kebab-case or PascalCase).
 * See: https://lucide.dev/icons
 */
import { computed, type Component } from 'vue'
import * as LucideIcons from '@lucide/vue'

interface Props {
  name: string
  size?: number | string
  color?: string
}

const props = withDefaults(defineProps<Props>(), {
  size: 16,
})

const iconComponent = computed<Component | undefined>(() => {
  // Accept both "mail" and "Mail", "trash-2" and "Trash2"
  const pascal = props.name
    .split(/[-_]/)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
    .join('')

  return (LucideIcons as Record<string, any>)[pascal]
})

const iconSize = computed(() => {
  const s = props.size
  return typeof s === 'number' ? `${s / 16}rem` : s // rem: icons scale with the root font size
})
</script>

<template>
  <component
    :is="iconComponent"
    v-if="iconComponent"
    :size="iconSize"
    :color="color"
    :class="$attrs.class"
    v-bind="$attrs"
  />
  <span v-else class="text-err text-[0.625rem]">?</span>
</template>
