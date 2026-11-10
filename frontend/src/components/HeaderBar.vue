<template>
  <header class="header-bar">
    <div class="header-toolbar">
      <div class="breadcrumb">
        <span class="breadcrumb-system">SYSTEM</span>
        <Icon name="chevron-right" :size="16" class="breadcrumb-icon" />
        <span class="breadcrumb-page">{{ pageTitle }}</span>
      </div>

      <div class="flex-1" />

      <div v-if="auth.user" class="user-block">
        <div class="user-info">
          <span class="user-name">{{ auth.user.username }}</span>
          <span class="user-role">{{ auth.user.superadmin ? 'SUPERADMIN' : 'ADMIN' }}</span>
        </div>
        <div class="user-avatar">
          <Icon name="shield" :size="20" class="text-on-action" />
        </div>
        <div class="header-divider" />
        <button class="logout-btn" @click="auth.logout()">
          <Icon name="log-out" :size="22" />
          <span>Logout</span>
        </button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useT } from '../pages/settings/i18n'

const auth = useAuthStore()
const route = useRoute()
const t = useT()

const routeTitles: Record<string, string> = {
  Dashboard: 'DASHBOARD',
  Mailboxes: 'EMAIL ACCOUNTS',
  Aliases: 'ALIASES',
  Domains: 'MY DOMAINS',
  AliasDomains: 'DOMAIN ALIASES',
  Logs: 'LOGS',
  MailLog: 'MAILLOG',
  Admins: 'ADMINISTRATORS',
  Transports: 'TRANSPORT LIST',
  APIKeys: 'API KEYS',
  Roles: 'ROLE MANAGEMENT',
  Settings: 'SETTINGS',
}

const pageTitle = computed(() =>
  route.name === 'Settings' ? t('settings') : (routeTitles[route.name as string] ?? 'DASHBOARD'),
)
</script>

<style scoped>
.header-bar {
  grid-area: header;
  height: 3.5rem;
  background: var(--color-surface);
  border-bottom: 3px solid var(--color-line);
  z-index: 100;
}

.header-toolbar {
  height: 100%;
  padding: 0 2rem;
  display: flex;
  align-items: center;
}

.breadcrumb {
  display: flex;
  align-items: center;
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.075rem;
  text-transform: uppercase;
}
.breadcrumb-system {
  color: var(--color-fg-subtle);
}
.breadcrumb-icon {
  color: var(--color-fg-subtle);
  margin: 0 0.125rem;
}
.breadcrumb-page {
  color: var(--color-fg);
}

.user-block {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}
.user-info {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  line-height: 1.2;
}
.user-name {
  font-size: 0.75rem;
  font-weight: 900;
  color: var(--color-fg);
  text-transform: uppercase;
  letter-spacing: -0.01875rem;
}
.user-role {
  font-size: 0.625rem;
  font-weight: 700;
  color: var(--color-fg-subtle);
  text-transform: uppercase;
  letter-spacing: 0.1rem;
}
.user-avatar {
  width: 2.5rem;
  height: 2.5rem;
  border: 2px solid var(--color-line);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-action);
  box-shadow: 1px 1px 0 var(--color-shadow);
}
.header-divider {
  width: 0.25rem;
  height: 2rem;
  background: var(--color-strong-fill);
  margin: 0 0.5rem;
}
.logout-btn {
  display: flex;
  align-items: center;
  gap: 0.4375rem;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--color-err);
  background: transparent;
  border: 2px solid transparent;
  cursor: pointer;
  border-radius: 0;
  transition: all 0.15s;
}
.logout-btn:hover {
  border-color: var(--color-err);
  background: var(--color-err-soft);
}
</style>
