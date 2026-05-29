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
  grid-template-columns: 240px 1fr;
  grid-template-rows: 56px 1fr;
  grid-template-areas:
    "sidebar header"
    "sidebar content";
  height: 100vh;
  overflow: hidden;
  background: #ebf2fe;
}

.page-content {
  grid-area: content;
  background: #ebf2fe;
  overflow: auto;
}
</style>
