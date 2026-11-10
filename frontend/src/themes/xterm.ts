import type { ITheme } from '@xterm/xterm'

const roles: Partial<Record<keyof ITheme, string>> = {
  background: 'bg',
  foreground: 'fg',
  cursor: 'cursor',
  cursorAccent: 'bg',
  selectionBackground: 'selection',
  black: 'black',
  red: 'red',
  green: 'green',
  yellow: 'yellow',
  blue: 'blue',
  magenta: 'magenta',
  cyan: 'cyan',
  white: 'white',
  brightBlack: 'bright-black',
  brightRed: 'bright-red',
  brightGreen: 'bright-green',
  brightYellow: 'bright-yellow',
  brightBlue: 'bright-blue',
  brightMagenta: 'bright-magenta',
  brightCyan: 'bright-cyan',
  brightWhite: 'bright-white',
}

// the xterm.js palette of the active theme, from the --term-* roles in terminal.css
export function xtermTheme(el: Element = document.documentElement): ITheme {
  const css = getComputedStyle(el)
  const theme: Record<string, string> = {}
  for (const [key, role] of Object.entries(roles)) {
    const v = css.getPropertyValue(`--term-${role}`).trim()
    if (v) theme[key] = v
  }
  return theme as ITheme
}
