<template>
  <div class="st-page">
    <div class="page-header">
      <div>
        <div class="page-title">{{ t('settings') }}</div>
        <div class="page-subtitle">{{ t('subtitle') }}</div>
      </div>
    </div>

    <nav class="st-tabs" role="tablist" :aria-label="t('tabsLabel')">
      <!-- custom: only the query changes, so a plain router-link would mark every tab as the current page -->
      <router-link v-for="tb in tabs" :key="tb.id" v-slot="{ href, navigate }" :to="{ query: { tab: tb.id } }" custom>
        <a
          :href="href"
          role="tab"
          class="st-tab"
          :class="{ 'st-tab--on': tab === tb.id }"
          :aria-selected="tab === tb.id"
          @click="navigate"
        >
          <Icon :name="tb.icon" :size="14" />
          {{ tb.label }}
        </a>
      </router-link>
    </nav>

    <section v-if="tab === 'appearance'" class="st-panel st-card" role="tabpanel">
      <div class="st-card-head">
        <Icon name="languages" :size="18" />
        <span>{{ t('language') }}</span>
      </div>
      <div class="st-lang">
        <select
          class="st-select"
          :aria-label="t('languageAria')"
          :value="lang.auto ? 'AUTO' : lang.lang"
          @change="lang.set(($event.target as HTMLSelectElement).value as Lang | 'AUTO')"
        >
          <option value="AUTO">{{ t('languageAuto', { lang: browserLang }) }}</option>
          <option value="EN">English</option>
          <option value="PT">Português</option>
          <option value="ES">Español</option>
        </select>
        <p class="st-lang-hint">{{ t('languageHint') }}</p>
      </div>
    </section>

    <section v-if="tab === 'appearance'" class="st-panel st-card" role="tabpanel">
      <div class="st-card-head">
        <Icon name="palette" :size="18" />
        <span>{{ t('theme') }}</span>
        <span class="st-count">{{ t('themesCount', { n: themes.length }) }}</span>
      </div>
      <div class="st-grid" role="radiogroup" :aria-label="t('themeAria')">
        <button
          v-for="th in themes"
          :key="th.id"
          type="button"
          role="radio"
          :aria-checked="theme.current === th.id"
          class="st-theme"
          :class="{ 'st-theme--on': theme.current === th.id }"
          @click="theme.select(th.id)"
        >
          <span class="st-preview" :data-theme="th.id" aria-hidden="true">
            <span class="st-bar"></span>
            <span class="st-body">
              <span class="st-line"></span>
              <span class="st-line st-line--short"></span>
              <span class="st-btn"></span>
            </span>
          </span>
          <span class="st-label">
            <Icon v-if="theme.current === th.id" name="circle-check" :size="16" />
            {{ themeLabel(th.label, lang.lang) }}
          </span>
        </button>
      </div>
      <p class="st-hint">
        {{ t('themeHint').split('{code}')[0] }}<code>{{ theme.serverDefault || 'default' }}</code
        >{{ t('themeHint').split('{code}')[1] }}
      </p>
    </section>

    <section v-else-if="tab === 'services'" class="st-panel st-card" role="tabpanel">
      <div class="st-card-head">
        <Icon name="server" :size="18" />
        <span>{{ t('mailServices') }}</span>
        <button type="button" class="st-act st-right" :disabled="loading" @click="loadServices">
          <Icon name="refresh-cw" :size="14" />
          {{ t('refresh') }}
        </button>
      </div>
      <p v-if="error" class="st-msg st-msg--err" role="alert">
        <Icon name="circle-x" :size="18" />
        <span>{{ t(error) }}</span>
      </p>
      <p v-else-if="services && !services.available" class="st-msg" role="status">
        <Icon name="triangle-alert" :size="18" />
        <span :title="services.message">{{ t('servicesUnavailable') }}</span>
      </p>
      <div v-if="services?.services.length" class="st-card-body">
        <table class="st-table">
          <thead>
            <tr>
              <th scope="col">{{ t('unit') }}</th>
              <th scope="col">{{ t('status') }}</th>
              <th scope="col">{{ t('boot') }}</th>
              <th scope="col">{{ t('since') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in services.services" :key="s.name">
              <td class="st-mono">{{ s.name }}</td>
              <td>
                <span v-if="!s.found" class="st-badge">
                  <Icon name="circle-minus" :size="14" />
                  {{ t('notInstalled') }}
                </span>
                <span v-else class="st-badge" :class="s.active === 'active' ? 'st-badge--ok' : 'st-badge--err'">
                  <Icon :name="s.active === 'active' ? 'circle-check' : 'circle-x'" :size="14" />
                  {{ s.active || t('unknown') }}<span class="st-sub">/{{ s.sub }}</span>
                </span>
              </td>
              <td>{{ s.found ? s.enabled : '—' }}</td>
              <td class="st-mono">{{ s.since || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="st-hint">
        {{ t('servicesHint').split('{code}')[0] }}<code>systemctl show</code>{{ t('servicesHint').split('{code}')[1] }}
      </p>
    </section>

    <!-- mounted on first visit and then only hidden, so an open SSH session survives switching tabs -->
    <TerminalTab v-if="terminalSeen" v-show="tab === 'terminal'" class="st-panel" role="tabpanel" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { themes } from '../../themes'
import { useThemeStore } from '../../stores/theme'
import { useAuthStore } from '../../stores/auth'
import TerminalTab from './TerminalTab.vue'
import http from '../../utils/http'
import { useT, themeLabel, type Key } from './i18n'
import { useLangStore, initialLang, type Lang } from '../../stores/lang'

type Service = { name: string; active: string; sub: string; enabled: string; since: string; found: boolean }
type Services = { available: boolean; message?: string; services: Service[] }

const auth = useAuthStore()
const t = useT()
const lang = useLangStore()
const browserLang = { EN: 'English', PT: 'Português', ES: 'Español' }[initialLang(null, navigator.language)]
// the terminal is a shell on the host: superadmins only (the server checks it again)
const tabs = computed(() => [
  { id: 'appearance', label: t('tabAppearance'), icon: 'palette' },
  { id: 'services', label: t('tabServices'), icon: 'server' },
  ...(auth.user?.superadmin ? [{ id: 'terminal', label: t('tabTerminal'), icon: 'square-terminal' }] : []),
])
const route = useRoute()
const tab = computed(() => (tabs.value.some((x) => x.id === route.query.tab) ? route.query.tab : 'appearance'))
const theme = useThemeStore()

const services = ref<Services | null>(null)
const loading = ref(false)
const error = ref<Key | ''>('')

async function loadServices() {
  loading.value = true
  error.value = ''
  try {
    services.value = (await http.get(`${API_BASE}/system/services`)).data?.data ?? null
  } catch (e: any) {
    error.value = e?.response?.status === 403 ? 'servicesForbidden' : 'servicesError'
  } finally {
    loading.value = false
  }
}

const terminalSeen = ref(false)
watch(
  tab,
  (t) => {
    if (t === 'services') loadServices()
    if (t === 'terminal') terminalSeen.value = true
  },
  { immediate: true },
)
</script>

<style scoped>
.st-lang {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 1rem 1.5rem;
  padding: 1.25rem;
}
.st-select {
  min-width: 18rem;
  min-height: 2.75rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-fg);
  background: var(--color-surface);
  border: 2px solid var(--color-line);
}
.st-lang-hint {
  font-size: 0.8125rem;
  color: var(--color-fg-2);
}
/* tabs sit on the line; the panel below keeps its distance (.st-panel) */
.st-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
  margin-top: 0.5rem;
  border-bottom: 3px solid var(--color-line);
}
.st-tab {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.25rem;
  padding: 0.625rem 1.25rem;
  font-size: 0.8125rem;
  font-weight: 900;
  letter-spacing: 0.0625rem;
  color: var(--color-fg-2);
  background: var(--color-surface-3);
  border: 2px solid var(--color-line);
  border-bottom: 0;
  text-decoration: none;
}
.st-tab:hover {
  color: var(--color-fg);
  background: var(--color-surface);
}
.st-tab--on {
  margin-top: 0;
  margin-bottom: -3px;
  padding-bottom: calc(0.625rem + 3px);
  color: var(--color-fg);
  background: var(--color-surface);
  border-width: 3px 3px 0;
  box-shadow: inset 0 4px 0 var(--color-action);
}
.st-right {
  margin-left: auto;
}
.st-count {
  margin-left: auto;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0;
  color: var(--color-fg-2);
}
.st-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
  color: var(--color-fg);
  border: 2px solid var(--color-line);
}
.st-table th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.0625rem;
  color: var(--color-surface);
  background: var(--color-fg);
}
.st-table td {
  padding: 0.875rem 1rem;
  border-bottom: 1px solid var(--color-line-soft);
}
.st-mono {
  font-family: var(--font-mono);
}
.st-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.625rem;
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--color-fg);
  background: var(--color-surface-3);
  border: 2px solid var(--color-line);
  border-left-width: 6px;
}
.st-badge--ok {
  background: var(--color-ok-soft);
  border-left-color: var(--color-ok-strong);
}
.st-badge--err {
  background: var(--color-err-soft);
  border-left-color: var(--color-danger);
}
.st-sub {
  color: var(--color-fg-2);
}
.st-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
  gap: 1.25rem;
  padding: 1.25rem;
}
.st-theme {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
  padding: 0.5rem;
  background: var(--color-surface);
  border: 2px solid var(--color-line-soft);
  cursor: pointer;
  text-align: left;
  color: var(--color-fg);
}
.st-theme:hover {
  border-color: var(--color-line);
}
.st-theme--on {
  border: 3px solid var(--color-line);
  padding: calc(0.5rem - 1px);
  box-shadow: 4px 4px 0 var(--color-shadow);
}
/* the preview carries its own data-theme, so these roles resolve to that theme's values */
.st-preview {
  display: flex;
  height: 6rem;
  background: var(--color-page);
  border: 1px solid var(--color-line);
  font-family: var(--font-sans);
}
.st-bar {
  width: 28%;
  background: var(--color-sidebar);
  border-right: 2px solid var(--color-line);
}
.st-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  padding: 0.625rem;
}
.st-line {
  height: 0.5rem;
  background: var(--color-fg);
}
.st-line--short {
  width: 60%;
  background: var(--color-fg-muted);
}
.st-btn {
  margin-top: auto;
  width: 50%;
  height: 0.875rem;
  background: var(--color-action);
  border: 1px solid var(--color-line);
}
.st-label {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0 0.125rem 0.125rem;
  font-size: 0.875rem;
  font-weight: 800;
}
</style>
