import { defineStore } from 'pinia'
import http from '../utils/http'

export type Lang = 'EN' | 'PT' | 'ES'
const LANGS: Lang[] = ['EN', 'PT', 'ES']
const KEY = 'lang'

// saved choice > browser language > English (the login and portal pages detect the browser the same way)
export function initialLang(saved: string | null, browser: string): Lang {
  const s = (saved || '').toUpperCase() as Lang
  if (LANGS.includes(s)) return s
  const b = browser.slice(0, 2).toUpperCase() as Lang
  return LANGS.includes(b) ? b : 'EN'
}

function saved(): string | null {
  try {
    return localStorage.getItem(KEY)
  } catch {
    return null
  }
}

const browser = () => (typeof navigator === 'undefined' ? '' : navigator.language)

export const useLangStore = defineStore('lang', {
  // auto: no saved choice, the browser language is used
  state: () => ({
    lang: initialLang(saved(), browser()),
    auto: !LANGS.includes((saved() || '').toUpperCase() as Lang),
  }),
  actions: {
    // 'AUTO' forgets the saved choice and follows the browser again
    async set(choice: Lang | 'AUTO') {
      const lang = choice === 'AUTO' ? initialLang(null, browser()) : choice
      this.lang = lang
      this.auto = choice === 'AUTO'
      try {
        if (choice === 'AUTO') localStorage.removeItem(KEY)
        else localStorage.setItem(KEY, choice)
      } catch {
        // private mode: this page only
      }
      try {
        await http.get(`/lang/${lang.toLowerCase()}`) // the backend's lang cookie (server-rendered messages)
      } catch {
        // the UI language does not depend on it
      }
    },
  },
})
