import { readFileSync, readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vite-plus/test'
import { themes } from './index'

const dir = fileURLToPath(new URL('.', import.meta.url))
const css = (f: string) => readFileSync(dir + f, 'utf8')
const vars = (block: string) =>
  Object.fromEntries([...block.matchAll(/(--[\w-]+)\s*:\s*([^;]+);/g)].map((m) => [m[1], m[2].trim()]))

describe('theme catalog', () => {
  it('registry ids equal the [data-theme] blocks', () => {
    const ids = readdirSync(dir)
      .filter((f) => f.endsWith('.css') && !['terminal.css', 'index.css', 'shape.css'].includes(f))
      .flatMap((f) => [...css(f).matchAll(/^\[data-theme='([\w-]+)'\] \{/gm)].map((m) => m[1]))
    expect(new Set(ids).size).toBe(ids.length)
    expect([...ids].sort()).toEqual(themes.map((t) => t.id).sort())
  })

  it('includes the XP theme', () => {
    expect(themes.find((t) => t.id === 'xp')?.label).toBe('Windows XP')
  })

  it('default.css repeats the role defaults of style.css', () => {
    const defaults = vars(/@theme static \{([\s\S]*?)\n\}/.exec(css('../style.css'))![1])
    const block = vars(css('default.css'))
    for (const [k, v] of Object.entries(defaults)) if (k.startsWith('--color-')) expect(block[k], k).toBe(v)
  })

  it('every theme declares the color scheme of its registry entry', () => {
    const all = readdirSync(dir)
      .filter((f) => f.endsWith('.css'))
      .map(css)
      .join('\n')
    for (const t of themes)
      expect(all).toMatch(new RegExp(`\\[data-theme='${t.id}'\\] \\{\\s*color-scheme: ${t.scheme};`))
  })

  it('terminal.css only sets the console frame color', () => {
    expect(css('terminal.css')).toMatch(/--term-bg: #000000;/)
  })
})
