<template>
  <div class="tm-wrap">
    <aside class="tm-side">
      <section class="st-card">
        <div class="st-card-head">
          <Icon name="server" :size="18" />
          <span>{{ t('server') }}</span>
          <button type="button" class="st-act tm-right" :disabled="loading" @click="load">
            <Icon name="refresh-cw" :size="14" />
            {{ t('refresh') }}
          </button>
        </div>
        <div v-if="info" class="tm-server">
          <strong>{{ t('thisServer') }}</strong>
          <span class="tm-mono">{{ info.target }}</span>
          <span v-if="info.host_key" class="tm-key">
            <Icon name="key-round" :size="14" />
            <span>{{ info.host_key }}</span>
          </span>
          <span v-else class="tm-key">
            <Icon name="key-round" :size="14" />
            <span>{{ t('noHostKey') }}</span>
          </span>
        </div>
      </section>

      <form v-if="info?.enabled && state !== 'open'" class="st-card tm-form" @submit.prevent="connect()">
        <div class="tm-field">
          <label class="tm-label" for="tm-user">{{ t('user') }}</label>
          <input id="tm-user" v-model="form.user" class="tm-input" autocomplete="username" required />
        </div>
        <div class="tm-field">
          <label class="tm-label" for="tm-pass">{{ t('password') }}</label>
          <input
            id="tm-pass"
            v-model="form.password"
            class="tm-input"
            type="password"
            autocomplete="current-password"
          />
        </div>
        <div class="tm-field">
          <span class="tm-label">{{ t('orKey') }}</span>
          <div class="tm-keyrow">
            <label class="st-act tm-file">
              <Icon name="file-key" :size="14" />
              {{ form.keyName || t('chooseFile') }}
              <input type="file" class="sr-only" @change="pickKey" />
            </label>
            <button v-if="form.key" type="button" class="st-act" @click="clearKey">{{ t('clear') }}</button>
          </div>
          <input
            v-if="form.key"
            v-model="form.passphrase"
            class="tm-input"
            type="password"
            :placeholder="t('passphrase')"
            :aria-label="t('passphrase')"
          />
        </div>
        <button type="submit" class="st-primary" :disabled="state === 'connecting'">
          <Icon name="plug" :size="16" />
          {{ state === 'connecting' ? t('connecting') : t('connect') }}
        </button>
      </form>

      <div class="tm-note">
        <p>{{ t('note1', { target: info?.target || 'terminal.host', term: 'xterm-256color' }) }}</p>
        <p>{{ t('note2', { n: info?.idle_minutes || 30 }) }}</p>
        <button v-if="info?.host_key && state === 'idle'" type="button" class="st-act" @click="resetKey">
          <Icon name="rotate-ccw" :size="14" />
          {{ t('forgetKey') }}
        </button>
      </div>
    </aside>

    <section class="st-card tm-main">
      <div class="st-card-head">
        <Icon name="square-terminal" :size="18" />
        <span>{{ t('terminal') }}</span>
        <span class="tm-status">
          {{ info?.target
          }}<template v-if="state === 'open'"> · {{ t('connectedAs', { user: connectedAs }) }}</template>
        </span>
        <button v-if="state === 'open'" type="button" class="st-act tm-right" @click="disconnect">
          <Icon name="unplug" :size="14" />
          {{ t('disconnect') }}
        </button>
      </div>
      <p v-if="info && !info.enabled" class="st-msg" role="status">
        <Icon name="triangle-alert" :size="18" />
        <span>{{ t('disabled', { enabled: '[terminal] enabled = true', hostport: 'host/port' }) }}</span>
      </p>
      <p v-if="error" class="st-msg st-msg--err" role="alert">
        <Icon name="circle-x" :size="18" />
        <span>{{ error && t(error) }}</span>
      </p>
      <div v-if="!started && info?.enabled" class="tm-empty">
        <Icon name="square-terminal" :size="40" />
        <p>{{ t('empty', { target: info.target }) }}</p>
      </div>
      <!-- frame (padding, border, black) outside the element xterm.js fits into: the fit addon measures its parent -->
      <div v-show="state !== 'idle' || started" class="tm-term">
        <div ref="termEl" class="tm-term-fit"></div>
      </div>
    </section>

    <ConfirmDialog v-model="askTrust" :title="t('trustTitle')" :confirm-label="t('trustOk')" @confirm="trustAndConnect">
      <p class="text-sm text-fg leading-relaxed">
        {{ t('trustText', { target: info?.target || '' }) }}
      </p>
      <p class="tm-fp">{{ pending?.key_type }} {{ pending?.fingerprint }}</p>
      <p class="text-xs text-fg-2">
        {{ t('trustCheck') }} <code>ssh-keygen -lf /etc/ssh/ssh_host_ed25519_key.pub</code>
      </p>
    </ConfirmDialog>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onUnmounted, ref, watch } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { xtermTheme } from '../../themes/xterm'
import { useAuthStore } from '../../stores/auth'
import { useToastStore } from '../../stores/toast'
import http from '../../utils/http'
import { useT, type Key } from './i18n'

type Info = { enabled: boolean; target: string; host_key?: string; idle_minutes: number }

// active: this tab is shown. The component stays mounted (an open session survives tab switches), so it reloads the
// terminal settings each time it is shown again: enabling it in Configuration must show up here without F5.
const props = defineProps<{ active?: boolean }>()
const auth = useAuthStore()
const t = useT()
const toast = useToastStore()
const info = ref<Info | null>(null)
const loading = ref(false)
const form = ref({ user: 'root', password: '', key: '', keyName: '', passphrase: '' })
const state = ref<'idle' | 'connecting' | 'open'>('idle')
const started = ref(false)
const error = ref<Key | ''>('')
const connectedAs = ref('')
const askTrust = ref(false)
const pending = ref<{ fingerprint: string; key_type: string } | null>(null)
const termEl = ref<HTMLElement | null>(null)
let ws: WebSocket | null = null
let term: Terminal | null = null
let fit: FitAddon | null = null
let resizeObs: ResizeObserver | null = null
let attempts = 0 // the latest connect: an older socket's answers are ignored

// the server's error codes, which are also dictionary keys
const codes: Key[] = [
  'auth_failed',
  'hostkey_changed',
  'unreachable',
  'timeout',
  'closed',
  'idle',
  'bad_request',
  'forbidden',
  'busy',
  'internal',
  'off',
]
const codeKey = (c: string): Key => (codes.includes(c as Key) ? (c as Key) : 'internal')

async function load() {
  loading.value = true
  try {
    // also refreshes the access token when it expired (the WebSocket sends it in its first message)
    info.value = (await http.get(`${API_BASE}/system/terminal`)).data?.data ?? null
  } catch (e: any) {
    error.value = e?.response?.status === 403 ? 'forbidden' : 'loadError'
  } finally {
    loading.value = false
  }
}

async function pickKey(ev: Event) {
  const input = ev.target as HTMLInputElement
  const f = input.files?.[0]
  input.value = ''
  if (!f) return
  if (f.size > 32 << 10) return toast.error(t('keyTooLarge'))
  form.value.key = await f.text()
  form.value.keyName = f.name
}
function clearKey() {
  form.value.key = form.value.keyName = form.value.passphrase = ''
}
function forget() {
  form.value.password = ''
  clearKey()
}

function ensureTerm() {
  if (term || !termEl.value) return
  term = new Terminal({
    cursorBlink: true,
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace',
    // follows the root font size (16px up to FullHD, larger on QHD/4K)
    fontSize: Math.round(parseFloat(getComputedStyle(document.documentElement).fontSize) * 0.8125),
    theme: xtermTheme,
  })
  fit = new FitAddon()
  term.loadAddon(fit)
  term.open(termEl.value)
  fit.fit()
  term.onData(
    (d) => ws?.readyState === WebSocket.OPEN && state.value === 'open' && ws.send(new TextEncoder().encode(d)),
  )
  term.onResize(({ cols, rows }) => state.value === 'open' && ws?.send(JSON.stringify({ type: 'resize', cols, rows })))
  resizeObs = new ResizeObserver(() => fit?.fit())
  resizeObs.observe(termEl.value)
}

async function connect(trust = '') {
  const n = ++attempts // a newer connect or leaving the page (onUnmounted) makes this one stale
  error.value = ''
  state.value = 'connecting'
  started.value = true
  await load() // fresh token and target
  if (n !== attempts) return
  if (!info.value?.enabled) {
    state.value = 'idle'
    return
  }
  await nextTick()
  if (n !== attempts) return
  ensureTerm()
  if (!term) return
  term.reset()
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  ws?.close()
  const sock = new WebSocket(`${proto}://${location.host}${API_BASE}/system/terminal/ws`)
  ws = sock
  sock.binaryType = 'arraybuffer'
  let answered = false
  sock.onopen = () => {
    const { user, password, key, passphrase } = form.value
    sock.send(
      JSON.stringify({
        token: auth.accessToken,
        user,
        password,
        key,
        passphrase,
        cols: term!.cols,
        rows: term!.rows,
        trust,
      }),
    )
  }
  sock.onmessage = (ev) => {
    if (typeof ev.data !== 'string') return term?.write(new Uint8Array(ev.data))
    answered = true
    const m = JSON.parse(ev.data)
    if (m.type === 'ready') {
      state.value = 'open'
      connectedAs.value = form.value.user
      forget() // the credential is not kept once the shell is open
      term?.focus()
      load()
    } else if (m.type === 'hostkey') {
      state.value = 'idle'
      if (n !== attempts) return
      pending.value = { fingerprint: m.fingerprint, key_type: m.key_type }
      askTrust.value = true
    } else if (m.type === 'error') {
      error.value = codeKey(m.code)
      if (state.value === 'open') term?.write('\r\n\x1b[2m[' + t(error.value) + ']\x1b[0m\r\n')
    }
  }
  sock.onclose = () => {
    if (ws !== sock) return
    if (!answered && state.value === 'connecting') error.value = 'openFailed'
    if (state.value === 'open' && !error.value) term?.write('\r\n\x1b[2m[' + t('closedNote') + ']\x1b[0m\r\n')
    state.value = 'idle'
    ws = null
  }
}

function trustAndConnect() {
  askTrust.value = false
  if (pending.value) connect(pending.value.fingerprint)
}

function disconnect() {
  ws?.close()
}

async function resetKey() {
  if (!confirm(t('forgetConfirm'))) return
  try {
    await http.delete(`${API_BASE}/system/terminal/hostkey`)
    toast.success(t('forgotten'))
    load()
  } catch {
    toast.error(t('forgetFailed'))
  }
}

// logout or session loss closes the terminal
watch(
  () => auth.accessToken,
  (t) => !t && ws?.close(),
)

watch(
  () => props.active,
  (shown, was) => shown && was === false && state.value === 'idle' && load(),
)

onUnmounted(() => {
  attempts++ // a trust confirmed after leaving the tab connects nothing
  ws?.close()
  ws = null
  resizeObs?.disconnect()
  term?.dispose()
  term = null
})

load()
</script>

<style scoped>
.tm-wrap {
  display: grid;
  /* minmax(0, 1fr): the terminal column may shrink below the xterm canvas, which then refits */
  grid-template-columns: minmax(17rem, 22rem) minmax(0, 1fr);
  gap: 1.75rem;
  align-items: start;
}
.tm-side {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
.tm-right {
  margin-left: auto;
}
.tm-server {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  margin: 1.25rem;
  padding: 0.875rem 1rem;
  font-size: 0.875rem;
  color: var(--color-fg);
  background: var(--color-action-soft);
  border: 2px solid var(--color-line);
  border-left: 6px solid var(--color-action);
}
.tm-mono {
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  color: var(--color-fg);
}
.tm-key {
  display: flex;
  gap: 0.375rem;
  align-items: flex-start;
  font-family: var(--font-mono);
  font-size: 0.75rem;
  line-height: 1.5;
  color: var(--color-fg-2);
  overflow-wrap: anywhere;
}
.tm-key :first-child {
  flex-shrink: 0;
  margin-top: 0.125rem;
}
.tm-form {
  display: flex;
  flex-direction: column;
  gap: 1.125rem;
  padding: 1.25rem;
}
.tm-field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.tm-label {
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.0625rem;
  color: var(--color-fg);
}
.tm-input {
  min-height: 2.75rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.9375rem;
  color: var(--color-fg);
  background: var(--color-surface);
  border: 2px solid var(--color-line);
}
.tm-input::placeholder {
  color: var(--color-fg-2);
}
.tm-keyrow {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.tm-file {
  cursor: pointer;
}
.tm-file:focus-within {
  outline: 3px solid var(--color-fg);
  outline-offset: 2px;
}
.tm-note {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  align-items: flex-start;
  font-size: 0.8125rem;
  line-height: 1.6;
  color: var(--color-fg-2);
}
.tm-note code {
  font-family: var(--font-mono);
  font-weight: 700;
  color: var(--color-fg);
}
.tm-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  min-height: 22rem;
}
.tm-status {
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  font-weight: 600;
  letter-spacing: 0;
  color: var(--color-fg-2);
}
.tm-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  min-height: 18rem;
  margin: 1.25rem;
  font-size: 0.875rem;
  color: var(--color-fg-2);
  border: 2px dashed var(--color-line-soft);
}
.tm-empty code {
  font-family: var(--font-mono);
  color: var(--color-fg);
}
.tm-term {
  overflow: hidden;
  min-height: 16rem;
  /* the window below the page header, tabs and card head */
  height: calc(100vh - 24rem);
  margin: 1.25rem;
  padding: 0.5rem 0.75rem;
  background: var(--term-bg);
  border: 2px solid var(--color-line);
}
.tm-term-fit {
  width: 100%;
  height: 100%;
  overflow: hidden;
}
.tm-fp {
  margin: 0.75rem 0;
  padding: 0.625rem 0.75rem;
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--color-fg);
  overflow-wrap: anywhere;
  background: var(--color-surface-2);
  border: 2px solid var(--color-line);
}
@media (max-width: 1100px) {
  .tm-wrap {
    grid-template-columns: 1fr;
  }
}
</style>
