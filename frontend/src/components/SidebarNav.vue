<template>
  <aside class="sidebar">
    <div class="sidebar-logo">
      <div class="logo-icon">
        <Icon name="mail" :size="20" color="white" />
      </div>
      <h1 class="logo-text">PostfixAdmin</h1>
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

      <router-link
        v-for="item in settingsNav"
        :key="item.to"
        :to="item.to"
        class="nav-item"
        active-class="nav-item--active"
        exact-active-class="nav-item--active"
      >
        <Icon :name="item.icon" :size="20" class="nav-icon" />
        <span class="nav-label">{{ item.label }}</span>
      </router-link>
    </nav>
  </aside>
</template>

<script setup lang="ts">
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
  { to: '/admins',     icon: 'users',        label: 'Administrators' },
  { to: '/transports', icon: 'arrow-up-down', label: 'Transport List' },
  { to: '/settings',   icon: 'settings',     label: 'Settings' },
]
</script>

<style scoped>
.sidebar {
  grid-area: sidebar;
  background: #ffffff;
  border-right: 3px solid #1e293b;
  display: flex;
  flex-direction: column;
}

.sidebar-logo {
  height: 56px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 3px solid #1e293b;
  flex-shrink: 0;
}

.logo-icon {
  width: 40px;
  height: 40px;
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
  font-family: monospace;
  color: var(--color-brand-text);
  letter-spacing: -0.3px;
}

.sidebar-nav {
  flex: 1;
  padding-top: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  font-size: 14px;
  font-weight: 700;
  color: #374151;
  text-decoration: none;
  transition: background .12s, border-color .12s, color .12s;
  border: 2px solid transparent;
  border-left-width: 8px;
  border-right-width: 0;
}

.nav-icon {
  color: #9ca3af;
  flex-shrink: 0;
  transition: color .12s;
}

.nav-item:hover {
  background: rgba(59, 130, 246, 0.1);
  border-color: #1e293b;
  border-right-color: transparent !important;
  color: #1e293b;
}
.nav-item:hover :deep(.nav-icon) { color: var(--color-brand-primary); }

.nav-item--active {
  background: #ebf2fe;
  border-color: #1e293b;
  border-right-color: transparent !important;
  color: #1e293b;
  margin-right: -3px;
}


.nav-item--active :deep(.nav-icon) { color: var(--color-brand-primary); }

.nav-section-title {
  padding: 16px 16px 6px;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 1.5px;
  text-transform: uppercase;
  color: #94a3b8;
}

</style>
