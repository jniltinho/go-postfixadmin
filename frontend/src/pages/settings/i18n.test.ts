import { describe, expect, it } from 'vite-plus/test'
import { dicts, themeLabel } from './i18n'
import { initialLang } from '../../stores/lang'

const vars = (s: string) => [...s.matchAll(/\{(\w+)\}/g)].map((m) => m[1]).sort()

describe('Settings translations', () => {
  it('PT and ES have every EN key, non-empty, with the same placeholders', () => {
    for (const lang of ['PT', 'ES'] as const) {
      for (const [k, v] of Object.entries(dicts.EN)) {
        const tr = (dicts[lang] as Record<string, string>)[k]
        expect(tr, `${lang}.${k}`).toBeTruthy()
        expect(vars(tr), `${lang}.${k}`).toEqual(vars(v))
      }
      expect(Object.keys(dicts[lang]).sort()).toEqual(Object.keys(dicts.EN).sort())
    }
  })

  it('translates only the descriptive words of theme names', () => {
    expect(themeLabel('Breeze (dark)', 'PT')).toBe('Breeze (escuro)')
    expect(themeLabel('High contrast (dark)', 'ES')).toBe('Alto contraste (oscuro)')
    expect(themeLabel('Default', 'PT')).toBe('Padrão')
    expect(themeLabel('Windows XP', 'ES')).toBe('Windows XP')
  })

  it('language: saved choice, else browser, else English', () => {
    expect(initialLang('es', 'pt-BR')).toBe('ES')
    expect(initialLang(null, 'pt-BR')).toBe('PT')
    expect(initialLang('xx', 'de-DE')).toBe('EN')
  })
})
