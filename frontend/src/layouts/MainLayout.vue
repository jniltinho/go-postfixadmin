<template>
  <q-layout view="lHh Lpr lFf">

    <!-- ═══════════════════════════════ HEADER ═══════════════════════════════ -->
    <q-header class="header-bar">
      <q-toolbar class="header-toolbar">

        <!-- Breadcrumb -->
        <div class="breadcrumb">
          <span class="breadcrumb-system">SYSTEM</span>
          <q-icon name="chevron_right" size="16px" class="breadcrumb-icon" />
          <span class="breadcrumb-page">{{ pageTitle }}</span>
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
            <q-icon name="shield" size="20px" color="white" />
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
      :width="240"
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
  padding: 0 32px;
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
.breadcrumb-system {
  color: #94a3b8;
}
.breadcrumb-icon {
  color: #94a3b8;
  margin: 0 2px;
}
.breadcrumb-page {
  color: var(--color-brand-text);
}

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

/* ─── Sidebar ─── */
:deep(.q-drawer__content) {
  background: #ffffff;
  border-right: 3px solid #1e293b;
  display: flex;
  flex-direction: column;
  overflow: visible;
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
  font-size: 20px;
  font-weight: 700;
  color: var(--color-brand-text);
  letter-spacing: -0.5px;
  font-family: monospace;
}

:deep(.q-scrollarea__content) {
  width: 100%;
}
.sidebar-list {
  padding-top: 8px;
  flex: 1;
}

.nav-item {
  padding: 8px 16px;
  border-radius: 0;
  font-size: 16px;
  color: #374151;
  min-height: 40px;
  cursor: pointer;
  transition: background .12s, color .12s;
  border-top: 2px solid transparent;
  border-bottom: 2px solid transparent;
  border-left: 6px solid transparent;
  border-right: 0px solid transparent;
}
.nav-item :deep(.q-icon) {
  color: #9ca3af;
  transition: color .12s;
}
.nav-item:hover {
  background: rgba(59, 130, 246, .08);
  color: var(--color-brand-text);
  border-top-color: var(--color-brand-text);
  border-bottom-color: var(--color-brand-text);
  border-left-color: var(--color-brand-text);
}
.nav-item:hover :deep(.q-icon) {
  color: var(--color-brand-text);
}
.nav-item--active {
  position: relative;
  background: #ebf2fe !important;
  color: var(--color-brand-text) !important;
  border-top-color: var(--color-brand-text) !important;
  border-bottom-color: var(--color-brand-text) !important;
  border-left-color: var(--color-brand-text) !important;
  border-right-width: 0 !important;
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
.nav-item--active :deep(.q-icon) {
  color: var(--color-brand-text) !important;
}
.nav-icon-section {
  min-width: 36px;
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
</style>
