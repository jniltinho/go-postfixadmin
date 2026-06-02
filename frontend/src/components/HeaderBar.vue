<template>
  <header class="header-bar">
    <div class="header-toolbar">

      <div class="breadcrumb">
        <span class="breadcrumb-system">SYSTEM</span>
        <Icon name="chevron-right" :size="16" class="breadcrumb-icon" />
        <span class="breadcrumb-page">{{ pageTitle }}</span>
      </div>

      <div class="flex-1" />

      <div class="lang-switcher">
        <button
          v-for="lang in ['PT','EN','ES']"
          :key="lang"
          class="lang-btn"
          :class="{ 'lang-active': activeLang === lang }"
          @click="activeLang = lang"
        >{{ lang }}</button>
      </div>

      <div v-if="auth.user" class="user-block">
        <div class="user-info">
          <span class="user-name">{{ auth.user.username }}</span>
          <span class="user-role">{{ auth.user.superadmin ? 'SUPERADMIN' : 'ADMIN' }}</span>
        </div>
        <div class="user-avatar">
          <Icon name="shield" :size="20" color="white" />
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
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const activeLang = ref('EN')

const routeTitles: Record<string, string> = {
  Dashboard:   'DASHBOARD',
  Mailboxes:   'EMAIL ACCOUNTS',
  Aliases:     'ALIASES',
  Domains:     'MY DOMAINS',
  AliasDomains:'DOMAIN ALIASES',
  Logs:        'LOGS',
  MailLog:     'MAILLOG',
  Admins:      'ADMINISTRATORS',
  Transports:  'TRANSPORT LIST',
  APIKeys:     'API KEYS',
  Roles:       'ROLE MANAGEMENT',
}

const pageTitle = computed(() => routeTitles[route.name as string] ?? 'DASHBOARD')
</script>

<style scoped>
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
</style>
