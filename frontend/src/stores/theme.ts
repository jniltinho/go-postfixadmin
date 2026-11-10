import { defineStore } from 'pinia'
import { themes, type Theme } from '../themes'

const KEY = 'theme'
const SCHEME_KEY = 'theme-scheme' // read by the inline script in index.html before the bundle loads

const find = (id: string | null | undefined) => themes.find((t) => t.id === id)

// stored choice > server default (ui.default_theme) > Default
export function resolveTheme(stored: string | null, serverDefault?: string | null): Theme {
  return find(stored) || find(serverDefault) || themes[0]
}

function read(): string | null {
  try {
    return localStorage.getItem(KEY)
  } catch {
    return null
  }
}

function apply(t: Theme) {
  const el = document.documentElement
  el.dataset.theme = t.id
  el.style.colorScheme = t.scheme
  document.querySelector('meta[name="color-scheme"]')?.setAttribute('content', t.scheme)
}

export const useThemeStore = defineStore('theme', {
  state: () => ({ current: resolveTheme(read()).id, serverDefault: null as string | null, picked: false }),
  actions: {
    async init() {
      try {
        const res = await fetch(`${API_BASE}/version`)
        if (res.ok) this.serverDefault = (await res.json()).default_theme ?? null
      } catch {
        // server default unavailable: keep the stored choice or Default
      }
      if (this.picked) return // chosen while /version was loading: keep it even when storage is unavailable
      const t = resolveTheme(read(), this.serverDefault)
      this.current = t.id
      apply(t)
    },
    select(id: string) {
      const t = resolveTheme(id, this.serverDefault)
      this.picked = true
      try {
        localStorage.setItem(KEY, t.id)
        localStorage.setItem(SCHEME_KEY, t.scheme)
      } catch {
        // private mode: the choice lasts for this page only
      }
      this.current = t.id
      apply(t)
    },
  },
})
