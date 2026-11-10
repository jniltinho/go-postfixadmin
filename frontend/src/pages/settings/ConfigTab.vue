<template>
  <div class="cf-wrap">
    <p v-if="loadError" class="st-msg st-msg--err" role="alert">
      <Icon name="circle-x" :size="18" />
      <span>{{ t('configLoadError') }}</span>
    </p>

    <div v-if="cfg" class="cf-cols">
      <form class="st-card cf-card" @submit.prevent="submit">
        <div class="st-card-head">
          <Icon name="sliders-horizontal" :size="18" />
          <span>{{ t('editable') }}</span>
          <span class="cf-file" :title="cfg.file"
            >{{ t('configFile') }}: <code>{{ cfg.file || '—' }}</code></span
          >
        </div>
        <!-- the card scrolls inside, so the page does not; head and save bar stay in view -->
        <div class="cf-scroll">
          <p v-if="!cfg.writable" class="st-msg" role="status">
            <Icon name="lock" :size="18" />
            <span>{{ t('notWritable') }}</span>
          </p>
          <div class="cf-grid">
            <div v-for="f in editable" :key="f.key" class="cf-field">
              <label class="cf-key" :for="'cf-' + f.key">
                <code>{{ f.key }}</code>
                <span v-if="f.restart" class="cf-tag">{{ t('restartTag') }}</span>
                <span v-if="f.env" class="cf-tag">{{ t('envOverride', { env: f.env }) }}</span>
              </label>
              <p class="cf-desc">{{ fieldDescs[lang.lang][f.key] }}</p>

              <label v-if="f.kind === 'bool'" class="cf-check">
                <input :id="'cf-' + f.key" v-model="form[f.key]" type="checkbox" :disabled="!f.editable" />
                <span>{{ form[f.key] ? t('yes') : t('no') }}</span>
              </label>
              <select
                v-else-if="f.key === 'ui.default_theme'"
                :id="'cf-' + f.key"
                v-model="form[f.key]"
                class="cf-input"
                :disabled="!f.editable"
              >
                <option v-for="th in themes" :key="th.id" :value="th.id">{{ themeLabel(th.label, lang.lang) }}</option>
              </select>
              <select
                v-else-if="f.kind === 'enum'"
                :id="'cf-' + f.key"
                v-model="form[f.key]"
                class="cf-input"
                :disabled="!f.editable"
              >
                <option v-for="o in f.options" :key="o" :value="o">{{ o }}</option>
              </select>
              <input
                v-else-if="f.kind === 'int'"
                :id="'cf-' + f.key"
                v-model.number="form[f.key]"
                class="cf-input"
                type="number"
                :min="f.min"
                :max="f.max"
                :disabled="!f.editable"
                required
              />
              <input
                v-else
                :id="'cf-' + f.key"
                v-model.trim="form[f.key]"
                class="cf-input"
                :placeholder="f.kind === 'duration' ? t('durationHint') : ''"
                :disabled="!f.editable"
                required
              />
            </div>
          </div>
        </div>
        <div class="cf-actions">
          <p class="cf-hint">{{ t('configHint') }}</p>
          <button type="submit" class="st-primary" :disabled="!dirty.length || saving || !cfg.writable">
            <Icon name="save" :size="16" />
            {{ saving ? t('saving') : t('save') }}
          </button>
        </div>
      </form>

      <section class="st-card cf-card">
        <div class="st-card-head">
          <Icon name="lock" :size="18" />
          <span>{{ t('readOnlyTitle') }}</span>
        </div>
        <div class="cf-scroll">
          <p class="st-hint cf-ro-hint">{{ t('readOnlyHint') }}</p>
          <div class="cf-sections">
            <div v-for="(fields, section) in readOnly" :key="section" class="cf-section">
              <h4 class="cf-section-name">[{{ section }}]</h4>
              <dl>
                <template v-for="f in fields" :key="f.key">
                  <dt>{{ f.key.split('.')[1] }}</dt>
                  <dd>
                    <template v-if="f.secret">
                      <Icon :name="f.set ? 'eye-off' : 'circle-minus'" :size="14" />
                      {{ f.set ? t('secretSet') : t('secretUnset') }}
                    </template>
                    <template v-else>{{ show(f) }}</template>
                  </dd>
                </template>
              </dl>
            </div>
          </div>
        </div>
      </section>
    </div>

    <ConfirmDialog
      v-if="cfg"
      v-model="askTerminal"
      :title="t('terminalOnTitle')"
      :confirm-label="t('terminalOnOk')"
      @confirm="confirmTerminal"
    >
      <p class="text-sm text-fg leading-relaxed">{{ t('terminalOnText', { target: terminalTarget }) }}</p>
    </ConfirmDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import http from '../../utils/http'
import { useToastStore } from '../../stores/toast'
import { themes } from '../../themes'
import { useLangStore } from '../../stores/lang'
import { useT, themeLabel, fieldDescs } from './i18n'

type Field = {
  key: string
  kind: 'bool' | 'int' | 'string' | 'enum' | 'duration' | 'list'
  value?: unknown
  editable: boolean
  configurable?: boolean
  secret?: boolean
  set?: boolean
  restart?: boolean
  env?: string
  options?: string[]
  min?: number
  max?: number
}
type Config = { file: string; writable: boolean; fields: Field[] }

const t = useT()
const toast = useToastStore()
const lang = useLangStore()
const cfg = ref<Config | null>(null)
const form = ref<Record<string, unknown>>({})
const loadError = ref(false)
const saving = ref(false)

// configurable keys stay in the editable card even when locked now (file permissions, environment override)
const editable = computed(() => cfg.value?.fields.filter((f) => f.configurable) ?? [])
const readOnly = computed(() => {
  const out: Record<string, Field[]> = {}
  for (const f of cfg.value?.fields ?? []) {
    if (f.configurable) continue
    const section = f.key.split('.')[0]
    ;(out[section] ??= []).push(f)
  }
  return out
})
const dirty = computed(() =>
  editable.value.filter((f) => f.editable && form.value[f.key] !== f.value).map((f) => f.key),
)

function show(f: Field) {
  if (f.kind === 'bool') return f.value ? t('yes') : t('no')
  if (Array.isArray(f.value)) return f.value.length ? f.value.join(', ') : '—'
  return f.value === '' || f.value == null ? '—' : String(f.value)
}

async function load() {
  loadError.value = false
  try {
    cfg.value = (await http.get(`${API_BASE}/system/config`)).data?.data ?? null
    form.value = Object.fromEntries((cfg.value?.fields ?? []).map((f) => [f.key, f.value]))
  } catch {
    loadError.value = true
  }
}

// enabling the web terminal is confirmed first; disabling (and everything else) saves directly
const askTerminal = ref(false)
const terminalTarget = computed(() => {
  const v = (k: string) => cfg.value?.fields.find((f) => f.key === k)?.value
  return `${v('terminal.host') ?? '127.0.0.1'}:${v('terminal.port') ?? 22}`
})
function submit() {
  if (dirty.value.includes('terminal.enabled') && form.value['terminal.enabled'] === true) askTerminal.value = true
  else save()
}
function confirmTerminal() {
  askTerminal.value = false
  save()
}

async function save() {
  saving.value = true
  const keys = dirty.value
  try {
    const res = (
      await http.put(`${API_BASE}/system/config`, { changes: Object.fromEntries(keys.map((k) => [k, form.value[k]])) })
    ).data?.data
    toast.success(t(res?.restart_required ? 'savedRestart' : 'saved', { keys: keys.join(', ') }))
    await load()
  } catch (e: any) {
    toast.error(t('saveError', { msg: e?.response?.data?.error?.message || e?.message || '' }))
  } finally {
    saving.value = false
  }
}

load()
</script>

<style scoped>
.cf-wrap {
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}
/* two cards side by side, as tall as the window below the tabs; each scrolls inside */
.cf-cols {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(0, 1fr);
  align-items: start;
  gap: 1.5rem;
}
.cf-card {
  display: flex;
  flex-direction: column;
  min-height: 0;
  /* as tall as its content, never taller than the window below the tabs (then it scrolls inside) */
  max-height: calc(100vh - 16rem);
}
.cf-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--color-line) transparent;
}
.cf-card .cf-actions {
  border-top: 1px solid var(--color-line-soft);
}
@media (max-width: 1100px) {
  .cf-cols {
    grid-template-columns: 1fr;
    height: auto;
  }
  .cf-scroll {
    overflow: visible;
  }
}
.cf-file {
  margin-left: auto;
  min-width: 0;
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0;
  color: var(--color-fg-2);
}
.cf-file code,
.cf-key code {
  font-family: var(--font-mono);
  color: var(--color-fg);
}
.cf-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(15rem, 1fr));
  gap: 1rem 1.5rem;
  padding: 1rem 1.25rem;
}
.cf-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.cf-key {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
  font-weight: 700;
}
.cf-tag {
  padding: 0 0.375rem;
  font-size: 0.6875rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.03125rem;
  color: var(--color-fg);
  background: var(--color-warn-soft);
  border: 1px solid var(--color-line);
  border-radius: var(--radius-sm);
}
.cf-desc {
  font-size: 0.75rem;
  line-height: 1.4;
  color: var(--color-fg-2);
}
.cf-input {
  min-height: 2.25rem;
  padding: 0.25rem 0.625rem;
  font-size: 0.875rem;
  color: var(--color-fg);
  background: var(--color-surface);
  border: 2px solid var(--color-line);
}
.cf-input:disabled,
.cf-check input:disabled + span {
  opacity: 0.7;
}
.cf-check {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  min-height: 2.25rem;
  font-size: 0.875rem;
  font-weight: 700;
  cursor: pointer;
}
.cf-check input {
  width: 1.25rem;
  height: 1.25rem;
  accent-color: var(--color-action);
}
.cf-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem 1.5rem;
  padding: 0.75rem 1.25rem;
}
.cf-hint {
  flex: 1 1 18rem;
  font-size: 0.75rem;
  line-height: 1.5;
  color: var(--color-fg-2);
}
.cf-ro-hint {
  padding: 0.75rem 1.25rem 0.5rem;
  font-size: 0.75rem;
}
.cf-sections {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(14rem, 1fr));
  gap: 0.75rem;
  padding: 0.25rem 1.25rem 1rem;
}
.cf-section {
  padding: 0.5rem 0.75rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-line-soft);
  border-radius: var(--radius-sm);
}
.cf-section-name {
  margin-bottom: 0.25rem;
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  font-weight: 800;
  color: var(--color-fg);
}
dl {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.125rem 0.75rem;
  font-size: 0.75rem;
}
dt {
  font-family: var(--font-mono);
  color: var(--color-fg-2);
}
dd {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-family: var(--font-mono);
  font-weight: 600;
  color: var(--color-fg);
  overflow-wrap: anywhere;
}
</style>
