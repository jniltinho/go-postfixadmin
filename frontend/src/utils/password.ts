const CHARS = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*'

const STRENGTH_LEVELS = [
  { pct: 20,  label: 'VERY WEAK',   color: '#ef4444' },
  { pct: 40,  label: 'WEAK',        color: '#f97316' },
  { pct: 60,  label: 'FAIR',        color: '#eab308' },
  { pct: 80,  label: 'STRONG',      color: '#22c55e' },
  { pct: 100, label: 'VERY STRONG', color: '#16a34a' },
]

export function calcStrength(pwd: string) {
  if (!pwd) return { pct: 0, label: '', color: '#e2e8f0' }
  let score = 0
  if (pwd.length >= 8)          score++
  if (pwd.length >= 12)         score++
  if (/[A-Z]/.test(pwd))        score++
  if (/[0-9]/.test(pwd))        score++
  if (/[^A-Za-z0-9]/.test(pwd)) score++
  return STRENGTH_LEVELS[Math.min(score - 1, 4)] ?? STRENGTH_LEVELS[0]
}

export function generatePassword(form: { password: string; passwordConfirm: string; [k: string]: any }) {
  const pwd = Array.from({ length: 16 }, () => CHARS[Math.floor(Math.random() * CHARS.length)]).join('')
  form.password = pwd
  form.passwordConfirm = pwd
}
