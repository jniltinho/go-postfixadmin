import { describe, expect, it } from 'vite-plus/test'
import { resolveTheme } from './theme'

describe('resolveTheme', () => {
  it('stored choice wins over the server default', () => expect(resolveTheme('matrix', 'xp').id).toBe('matrix'))
  it('unknown stored id falls back to the server default', () => expect(resolveTheme('nope', 'xp').id).toBe('xp'))
  it('server default when nothing is stored', () => expect(resolveTheme(null, 'xp').id).toBe('xp'))
  it('Default when the server default is missing or unknown', () => {
    expect(resolveTheme(null, null).id).toBe('default')
    expect(resolveTheme(null, 'bogus').id).toBe('default')
  })
})

describe('theme store init', () => {
  it('keeps a theme picked while /version is loading, even without storage', async () => {
    const { createPinia, setActivePinia } = await import('pinia')
    const { vi } = await import('vite-plus/test')
    const { useThemeStore } = await import('./theme')
    let release!: (v: unknown) => void
    vi.stubGlobal('API_BASE', '/api/v1')
    vi.stubGlobal('localStorage', {
      getItem: () => null,
      setItem: () => {
        throw new Error('storage blocked')
      },
    })
    vi.stubGlobal('document', { documentElement: { dataset: {}, style: {} }, querySelector: () => null })
    vi.stubGlobal('fetch', () => new Promise((r) => (release = r)))
    setActivePinia(createPinia())
    const store = useThemeStore()
    const done = store.init()
    store.select('matrix')
    release({ ok: true, json: async () => ({ default_theme: 'xp' }) })
    await done
    expect(store.current).toBe('matrix')
    expect(store.serverDefault).toBe('xp')
    vi.unstubAllGlobals()
  })
})
