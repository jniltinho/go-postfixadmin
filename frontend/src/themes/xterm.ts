import type { ITheme } from '@xterm/xterm'

// The SSH terminal's colors: the classic xterm palette (the 16 ANSI colors; xterm.js derives the 256-color cube and
// gray ramp from the standard), the same in every panel theme, like a real terminal window (as in ad-manager).
export const xtermTheme: ITheme = {
  background: '#000000',
  foreground: '#e5e5e5',
  cursor: '#e5e5e5',
  cursorAccent: '#000000',
  selectionBackground: '#4d4d4d',
  scrollbarSliderBackground: 'rgba(255, 255, 255, 0.12)',
  scrollbarSliderHoverBackground: 'rgba(255, 255, 255, 0.25)',
  scrollbarSliderActiveBackground: 'rgba(255, 255, 255, 0.35)',
  black: '#000000',
  red: '#cd0000',
  green: '#00cd00',
  yellow: '#cdcd00',
  blue: '#0000ee',
  magenta: '#cd00cd',
  cyan: '#00cdcd',
  white: '#e5e5e5',
  brightBlack: '#7f7f7f',
  brightRed: '#ff0000',
  brightGreen: '#00ff00',
  brightYellow: '#ffff00',
  brightBlue: '#5c5cff',
  brightMagenta: '#ff00ff',
  brightCyan: '#00ffff',
  brightWhite: '#ffffff',
}
