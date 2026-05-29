<template>
  <div class="app-shell">
    <!-- ═══════════════════════════════ HEADER ═══════════════════════════════ -->
    <header class="header-bar">
      <div class="header-toolbar">
        <!-- Breadcrumb -->
        <div class="breadcrumb">
          <span class="breadcrumb-system">SYSTEM</span>
          <Icon name="chevron-right" :size="16" class="breadcrumb-icon" />
          <span class="breadcrumb-page">{{ pageTitle }}</span>
        </div>

        <div class="flex-1" />

        <!-- Language switcher -->
        <div class="lang-switcher">
          <button
            v-for="lang in ['PT','EN','ES']"
            :key="lang"
            class="lang-btn"
            :class="{ 'lang-active': activeLang === lang }"
            @click="activeLang = lang"
          >
            {{ lang }}
          </button>
        </div>

        <!-- User info -->
        <div v-if="auth.user" class="user-block">
          <div class="user-info">
            <span class="user-name">{{ auth.user.username }}</span>
            <span class="user-role">{{ auth.user.superadmin ? 'SUPERADMIN' : 'ADMIN' }}</span>
          </div>
          <div class="user-avatar">
            <Icon name="shield" :size="20" color="white" />
          </div>
          <div class="header-divider" />
          <button class="logout-btn" @click="logout">
            <Icon name="log-out" :size="22" />
            <span>Logout</span>
          </button>
        </div>
      </div>
    </header>

    <!-- ═══════════════════════════════ SIDEBAR (always visible on desktop) ═══════════════════════════════ -->
    <aside class="sidebar">
      <!-- Logo / branding -->
      <div class="sidebar-logo">
        <div class="logo-icon">
          <Icon name="mail" :size="18" color="white" />
        </div>
        <span class="logo-text">PostfixAdmin</span>
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
          <Icon :name="item.icon" :size="18" class="nav-icon" />
          <span class="nav-label">{{ item.label }}</span>
        </router-link>

        <div class="settings-label">SETTINGS</div>

        <router-link
          v-for="item in settingsNav"
          :key="item.to"
          :to="item.to"
          class="nav-item"
          active-class="nav-item--active"
          exact-active-class="nav-item--active"
        >
          <Icon :name="item.icon" :size="18" class="nav-icon" />
          <span class="nav-label">{{ item.label }}</span>
        </router-link>
      </nav>
    </aside>

    <!-- ═══════════════════════════════ CONTENT ═══════════════════════════════ -->
    <main class="page-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const activeLang = ref('EN')

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
  Settings: 'SETTINGS',
}

const pageTitle = computed(() => routeTitles[route.name as string] ?? 'DASHBOARD')

// Note: icon names are now Lucide (kebab-case)
const mainNav = [
  { to: '/dashboard',     icon: 'layout-dashboard', label: 'Dashboard' },
  { to: '/mailboxes',     icon: 'mail',             label: 'Mailboxes' },
  { to: '/aliases',       icon: 'forward',          label: 'Aliases' },
  { to: '/domains',       icon: 'globe',            label: 'My Domains' },
  { to: '/alias-domains', icon: 'arrow-left-right', label: 'Domain Aliases' },
  { to: '/logs',          icon: 'history',          label: 'Logs' },
  { to: '/maillog',       icon: 'mail',             label: 'MailLog' },
]

const settingsNav = [
  { to: '/admins',     icon: 'users',    label: 'Administrators' },
  { to: '/transports', icon: 'arrow-up-down', label: 'Transport List' },
  { to: '/settings',   icon: 'settings', label: 'Settings' },
]

function logout() {
  auth.logout()
}
</script>

<style scoped>
/* ═══════════════════════════════════════════════════════════════════════════
   MainLayout — Neo-Brutalist (no Quasar)
   ═══════════════════════════════════════════════════════════════════════════ */

.app-shell {
  display: grid;
  grid-template-columns: 240px 1fr;
  grid-template-rows: 56px 1fr;
  grid-template-areas:
    "sidebar header"
    "sidebar content";
  min-height: 100vh;
  background: #ebf2fe;
}

/* ─── Header ─── */
.header-bar {
  grid-area: header;
  height: 56px;
  background: #ffffff;
  border-bottom: 3px solid #1e293b;
  z-index: 100;
}

.header-toolbar {
  height: 100%;
  padding: 0 32px;
  display: flex;
  align-items: center;
}

/* Breadcrumb */
.breadcrumb {
  display: flex;
  align-items: center;
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 1.2px;
  text-transform: uppercase;
}
.breadcrumb-system { color: #94a3b8; }
.breadcrumb-icon { color: #94a3b8; margin: 0 2px; }
.breadcrumb-page { color: var(--color-brand-text); }

/* Language switcher */
.lang-switcher {
  display: flex;
  gap: 4px;
  margin-right: 16px;
}
.lang-btn {
  padding: 2px 8px;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 1px;
  border: 1px solid #d1d5db;
  border-radius: 0;
  cursor: pointer;
  color: #64748b;
  background: #fff;
  transition: all .15s;
}
.lang-btn:hover { color: var(--color-brand-primary); border-color: #1e293b; }
.lang-active {
  background: var(--color-brand-primary);
  color: #fff !important;
  border-color: var(--color-brand-primary);
  box-shadow: 1px 1px 0 #1e293b;
}

/* User block */
.user-block {
  display: flex;
  align-items: center;
  gap: 10px;
}
.user-info {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  line-height: 1.2;
}
.user-name {
  font-size: 12px;
  font-weight: 900;
  color: var(--color-brand-text);
  text-transform: uppercase;
  letter-spacing: -0.3px;
}
.user-role {
  font-size: 10px;
  font-weight: 700;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 1.6px;
}
.user-avatar {
  width: 40px;
  height: 40px;
  border: 2px solid #1e293b;
  border-radius: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-brand-primary);
  box-shadow: 1px 1px 0 #1e293b;
}
.header-divider {
  width: 4px;
  height: 32px;
  background: #1e293b;
  margin: 0 8px;
}
.logout-btn {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 8px 16px;
  font-size: 14px;
  font-weight: 700;
  color: #ef4444;
  background: transparent;
  border: 2px solid transparent;
  cursor: pointer;
  border-radius: 0;
  transition: all .15s;
}
.logout-btn:hover {
  border-color: #ef4444;
  background: #fef2f2;
}

/* ─── Sidebar (always visible) ─── */
.sidebar {
  grid-area: sidebar;
  background: #ffffff;
  border-right: 3px solid #1e293b;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 56px;
  padding: 0 16px;
  border-bottom: 3px solid #1e293b;
  flex-shrink: 0;
}

.logo-icon {
  width: 36px;
  height: 36px;
  background: var(--color-brand-primary);
  border: 2px solid #1e293b;
  box-shadow: 1px 1px 0 #1e293b;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logo-text {
  font-size: 20px;
  font-weight: 700;
  color: var(--color-brand-text);
  letter-spacing: -0.5px;
  font-family: monospace;
}

.sidebar-nav {
  padding-top: 8px;
  flex: 1;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  font-size: 16px;
  font-weight: 700;
  color: #374151;
  text-decoration: none;
  transition: all .12s;
  border-top: 2px solid transparent;
  border-bottom: 2px solid transparent;
  border-left: 8px solid transparent;
}

.nav-icon {
  color: #9ca3af;
  transition: color .12s;
  flex-shrink: 0;
}

.nav-item:hover {
  background: rgba(59, 130, 246, 0.08);
  color: var(--color-brand-text);
  border-top-color: var(--color-brand-text);
  border-bottom-color: var(--color-brand-text);
  border-left-color: var(--color-brand-text);
}

.nav-item:hover .nav-icon {
  color: var(--color-brand-text);
}

.nav-item--active {
  background: #ebf2fe;
  color: var(--color-brand-text);
  border-top-color: var(--color-brand-text);
  border-bottom-color: var(--color-brand-text);
  border-left-color: var(--color-brand-text);
  position: relative;
}

.nav-item--active::after {
  content: '';
  position: absolute;
  top: -2px;
  bottom: -2px;
  right: -3px;
  width: 3px;
  background: #ebf2fe;
  z-index: 10;
}

.nav-item--active .nav-icon {
  color: var(--color-brand-text);
}

.nav-label {
  font-weight: 700;
}

.settings-label {
  margin: 16px 16px 4px;
  font-size: 10px;
  font-weight: 900;
  color: #94a3b8;
  letter-spacing: 1px;
  text-transform: uppercase;
}

/* ─── Main content area ─── */
.page-content {
  grid-area: content;
  background: #ebf2fe;
  overflow: auto;
}
</style>
