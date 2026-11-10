import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { expect, it } from 'vite-plus/test'

const root = fileURLToPath(new URL('.', import.meta.url))
const files = (dir: string): string[] =>
  readdirSync(dir).flatMap((f) => {
    const p = join(dir, f)
    if (statSync(p).isDirectory()) return f === 'themes' ? [] : files(p)
    return /\.(vue|css)$/.test(f) ? [p] : []
  })
const palette =
  /\b(?:bg|text|border|ring|divide|outline|from|to|fill|stroke)-(?:white|black|gray|slate|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose|brand-\w+)\b/

it('components use theme roles, not literal colors (themes/ and the contract in style.css excepted)', () => {
  const bad: string[] = []
  for (const f of files(root)) {
    let src = readFileSync(f, 'utf8')
    if (f.endsWith('style.css')) src = src.replace(/@theme static \{[\s\S]*?\n\}/, '')
    src.split('\n').forEach((line, i) => {
      if (
        /#[0-9a-fA-F]{3,8}\b(?![\w-])|rgba?\(|(?:oklch|oklab|hsla?|lab|lch)\(|(?:^|[;{]\s*)(?:color|background(?:-color)?|fill|stroke)\s*:\s*(?:white|black)\b/.test(
          line.replace(/href="#[^"]*"|url\(#[^)]*\)/g, ''),
        ) ||
        palette.test(line)
      )
        bad.push(`${f.slice(root.length)}:${i + 1}: ${line.trim()}`)
    })
  }
  expect(bad).toEqual([])
})
