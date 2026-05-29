<template>
  <q-layout view="lHh Lpr lFf">

    <!-- ═══════════════════════════════ HEADER ═══════════════════════════════ -->
    <q-header class="header-bar">
      <q-toolbar class="header-toolbar q-px-md">

        <!-- Breadcrumb -->
        <div class="breadcrumb">
          SYSTEM <span class="breadcrumb-sep">›</span> {{ pageTitle }}
        </div>

        <q-space />

        <!-- Language switcher -->
        <div class="lang-switcher">
          <button
            v-for="lang in ['PT','EN','ES']"
            :key="lang"
            class="lang-btn"
            :class="{ 'lang-active': activeLang === lang }"
            @click="activeLang = lang"
          >{{ lang }}</button>
        </div>

        <!-- User info -->
        <div v-if="auth.user" class="user-block">
          <div class="user-info">
            <span class="user-name">{{ auth.user.username }}</span>
            <span class="user-role">{{ auth.user.superadmin ? 'SUPERADMIN' : 'ADMIN' }}</span>
          </div>
          <div class="user-avatar">
            <q-icon name="shield" size="18px" color="white" />
          </div>
          <div class="header-divider" />
          <button class="logout-btn" @click="logout">
            <q-icon name="exit_to_app" size="26px" />
            Logout
          </button>
        </div>

      </q-toolbar>
    </q-header>

    <!-- ═══════════════════════════════ SIDEBAR ═══════════════════════════════ -->
    <q-drawer
      v-model="leftDrawerOpen"
      show-if-above
      :width="210"
      class="sidebar"
    >
      <!-- Logo / branding at top of sidebar -->
      <div class="sidebar-logo">
        <div class="logo-icon">
          <q-icon name="mail" size="18px" color="white" />
        </div>
        <span class="logo-text">PostfixAdmin</span>
      </div>

      <q-list class="sidebar-list">

        <q-item
          v-for="item in mainNav"
          :key="item.to"
          :to="item.to"
          exact
          class="nav-item"
          active-class="nav-item--active"
        >
          <q-item-section avatar class="nav-icon-section">
            <q-icon :name="item.icon" size="18px" />
          </q-item-section>
          <q-item-section class="nav-label">{{ item.label }}</q-item-section>
        </q-item>

        <div class="settings-label">SETTINGS</div>

        <q-item
          v-for="item in settingsNav"
          :key="item.to"
          :to="item.to"
          exact
          class="nav-item"
          active-class="nav-item--active"
        >
          <q-item-section avatar class="nav-icon-section">
            <q-icon :name="item.icon" size="18px" />
          </q-item-section>
          <q-item-section class="nav-label">{{ item.label }}</q-item-section>
        </q-item>

      </q-list>
    </q-drawer>

    <!-- ═══════════════════════════════ CONTENT ═══════════════════════════════ -->
    <q-page-container class="page-content">
      <router-view />
    </q-page-container>

  </q-layout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const leftDrawerOpen = ref(false)
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

const mainNav = [
  { to: '/dashboard',     icon: 'dashboard',   label: 'Dashboard' },
  { to: '/mailboxes',     icon: 'mail',        label: 'Mailboxes' },
  { to: '/aliases',       icon: 'forward',     label: 'Aliases' },
  { to: '/domains',       icon: 'public',      label: 'My Domains' },
  { to: '/alias-domains', icon: 'swap_horiz',  label: 'Domain Aliases' },
  { to: '/logs',          icon: 'history',     label: 'Logs' },
  { to: '/maillog',       icon: 'mail_outline',label: 'MailLog' },
]

const settingsNav = [
  { to: '/admins',     icon: 'people',    label: 'Administrators' },
  { to: '/transports', icon: 'swap_vert', label: 'Transport List' },
  { to: '/settings',   icon: 'settings',  label: 'Settings' },
]

function logout() {
  auth.logout()
}
</script>

<style scoped>
/* ─── Page content background ─── */
.page-content {
  background: #ebf2fe;
}

/* ─── Header ─── */
.header-bar {
  background: #ffffff;
  border-bottom: 3px solid #1e293b;
}

.header-toolbar {
  min-height: 56px;
  height: 56px;
}

/* Breadcrumb */
.breadcrumb {
  font-size: 11px;
  font-weight: 800;
  color: #94a3b8;
  letter-spacing: 0.8px;
  text-transform: uppercase;
}
.breadcrumb-sep {
  margin: 0 6px;
  color: #94a3b8;
}

/* Language switcher */
.lang-switcher {
  display: flex;
  gap: 4px;
  margin-right: 16px;
}
.lang-btn {
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
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
  font-weight: 700;
  color: var(--color-brand-text);
  text-transform: uppercase;
  letter-spacing: 0.2px;
}
.user-role {
  font-size: 10px;
  font-weight: 600;
  color: #64748b;
  letter-spacing: 0.5px;
}
.user-avatar {
  width: 32px;
  height: 32px;
  border: 2px solid #1e293b;
  border-radius: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-brand-primary);
  box-shadow: 1px 1px 0 #1e293b;
}
.header-divider {
  width: 3px;
  height: 28px;
  background: #1e293b;
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

/* ─── Sidebar ─── */
.sidebar :deep(.q-drawer__content) {
  background: #ffffff;
  border-right: 3px solid #1e293b;
  display: flex;
  flex-direction: column;
}

/* Sidebar logo/branding section */
.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 59px;
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
  font-size: 17px;
  font-weight: 700;
  color: var(--color-brand-text);
  letter-spacing: -0.3px;
  font-family: monospace;
}

.sidebar :deep(.q-scrollarea__content) {
  width: 100%;
}
.sidebar-list {
  padding-top: 8px;
  flex: 1;
}

.nav-item {
  padding: 8px 14px;
  border-radius: 0;
  font-size: 13px;
  color: #374151;
  min-height: 40px;
  cursor: pointer;
  transition: background .12s, color .12s;
  border: 2px solid transparent;
}
.nav-item :deep(.q-icon) {
  color: #6b7280;
  transition: color .12s;
}
.nav-item:hover {
  background: rgba(59, 130, 246, .08);
  color: var(--color-brand-text);
  border-color: var(--color-brand-text);
}
.nav-item:hover :deep(.q-icon) {
  color: var(--color-brand-text);
}
.nav-item--active {
  background: rgba(59, 130, 246, .08) !important;
  color: var(--color-brand-primary) !important;
  border-color: var(--color-brand-text) !important;
  border-left-width: 6px !important;
  border-right-color: transparent !important;
}
.nav-item--active :deep(.q-icon) {
  color: var(--color-brand-primary) !important;
}
.nav-icon-section {
  min-width: 36px;
}
.nav-label {
  font-weight: 600;
}

.settings-label {
  margin: 16px 14px 4px;
  font-size: 10px;
  font-weight: 800;
  color: #94a3b8;
  letter-spacing: 1.4px;
  text-transform: uppercase;
}
</style>
