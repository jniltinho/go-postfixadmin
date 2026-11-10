<template>
  <div class="app-shell">
    <HeaderBar />
    <SidebarNav />
    <main class="page-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import HeaderBar from '../components/HeaderBar.vue'
import SidebarNav from '../components/SidebarNav.vue'
</script>

<style scoped>
.app-shell {
  display: grid;
  grid-template-columns: 15rem 1fr;
  grid-template-rows: 3.5rem 1fr;
  grid-template-areas:
    'sidebar header'
    'sidebar content';
  height: 100vh;
  overflow: hidden;
  background: var(--color-page);
}

.page-content {
  grid-area: content;
  background: var(--color-page);
  overflow: auto;
}
</style>
