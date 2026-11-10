<template>
  <aside class="sidebar">
    <div class="sidebar-logo">
      <div class="logo-icon">
        <Icon name="mail" :size="20" class="text-on-action" />
      </div>
      <h1 class="logo-text">Go-PostfixAdmin</h1>
    </div>

    <nav class="sidebar-nav">
      <router-link
        v-for="item in mainNav"
        :key="item.to"
        :to="item.to"
        class="nav-item"
        active-class="nav-item--active"
        exact-active-class="nav-item--active"
      >
        <Icon :name="item.icon" :size="20" class="nav-icon" />
        <span class="nav-label">{{ item.label }}</span>
      </router-link>

      <div class="nav-section-title">SETTINGS</div>

      <template v-for="item in settingsNav" :key="item.to">
        <router-link
          v-if="!isSettingsDisabled(item)"
          :to="item.to"
          class="nav-item"
          active-class="nav-item--active"
          exact-active-class="nav-item--active"
        >
          <Icon :name="item.icon" :size="20" class="nav-icon" />
          <span class="nav-label">{{ 'i18n' in item ? settingsLabel : item.label }}</span>
        </router-link>
      </template>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useT } from '../pages/settings/i18n'

const auth = useAuthStore()
const t = useT()
// only the Settings entry is translated so far (the rest of the admin menu is English)
const settingsLabel = computed(() => t('tabsLabel'))

const mainNav = [
  { to: '/dashboard', icon: 'layout-dashboard', label: 'Dashboard' },
  { to: '/mailboxes', icon: 'mail', label: 'Mailboxes' },
  { to: '/aliases', icon: 'forward', label: 'Aliases' },
  { to: '/domains', icon: 'globe', label: 'My Domains' },
  { to: '/alias-domains', icon: 'arrow-left-right', label: 'Domain Aliases' },
  { to: '/logs', icon: 'history', label: 'Logs' },
  { to: '/maillog', icon: 'mail', label: 'MailLog' },
]

const settingsNav = [
  { to: '/admins', icon: 'users', label: 'Administrators', perm: 'admins:read' },
  { to: '/roles', icon: 'shield', label: 'Roles', perm: 'admins:write' },
  { to: '/transports', icon: 'arrow-up-down', label: 'Transport List', perm: 'transports:read' },
  { to: '/apikeys', icon: 'key', label: 'API Keys', perm: 'apikeys:read' },
  { to: '/settings', icon: 'settings', label: 'Settings', perm: null, i18n: true },
]

function isSettingsDisabled(item: { to: string; perm: string | null }) {
  if (!item.perm) return false
  return !auth.hasPermission(item.perm)
}
</script>

<style scoped>
.sidebar {
  grid-area: sidebar;
  background: var(--color-surface);
  border-right: 3px solid var(--color-line);
  display: flex;
  flex-direction: column;
}

.sidebar-logo {
  height: 3.5rem;
  padding: 0 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border-bottom: 3px solid var(--color-line);
  flex-shrink: 0;
}

.logo-icon {
  width: 2.5rem;
  height: 2.5rem;
  background: var(--color-action);
  border: 2px solid var(--color-line);
  box-shadow: var(--shadow-1);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logo-text {
  font-size: 1.0625rem;
  white-space: nowrap;
  font-weight: 700;
  font-family: monospace;
  color: var(--color-fg);
  letter-spacing: -0.01875rem;
}

.sidebar-nav {
  flex: 1;
  padding-top: 0.5rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--color-fg-2);
  text-decoration: none;
  transition:
    background 0.12s,
    border-color 0.12s,
    color 0.12s;
  border: 2px solid transparent;
  border-left-width: 8px;
  border-right-width: 0;
}

.nav-icon {
  color: var(--color-fg-subtle);
  flex-shrink: 0;
  transition: color 0.12s;
}

/* hover: tint and left bar only — a top/bottom border here drew a box closed by the sidebar edge (a stray line) */
.nav-item:hover {
  background: color-mix(in srgb, var(--color-action) 10%, transparent);
  border-left-color: var(--color-line-soft);
  color: var(--color-fg);
}

.nav-item--disabled,
.nav-item--disabled:hover {
  background: transparent;
  border-color: transparent;
  color: var(--color-fg-subtle);
  cursor: not-allowed;
  opacity: 0.48;
}
.nav-item:hover :deep(.nav-icon) {
  color: var(--color-action);
}
.nav-item--disabled:hover :deep(.nav-icon) {
  color: var(--color-fg-subtle);
}

/* selected: a tab that opens into the content (covers the sidebar edge) */
.nav-item--active,
.nav-item--active:hover {
  background: var(--color-page);
  border-color: var(--color-line);
  border-right-color: transparent;
  color: var(--color-fg);
  margin-right: -0.1875rem;
}

.nav-item--active :deep(.nav-icon) {
  color: var(--color-action);
}

.nav-section-title {
  padding: 1rem 1rem 0.375rem;
  font-size: 0.625rem;
  font-weight: 900;
  letter-spacing: 0.09375rem;
  text-transform: uppercase;
  color: var(--color-fg-subtle);
}
</style>
