<template>
  <div class="login-page dot-pattern">
    <div class="w-full login-col">
      <!-- Brand Header -->
      <div class="brand-header">
        <div class="brand-icon-box">
          <Icon name="user" :size="32" class="text-on-action" />
        </div>
        <h1 class="brand-title-mono">Go-PostfixAdmin</h1>
        <p class="brand-subtitle-flat">{{ t.subtitle }}</p>
      </div>

      <!-- Square Card -->
      <div class="login-card-neo neo-shadow">
        <h2 class="card-heading-neo">{{ t.title }}</h2>

        <!-- Error Message -->
        <form @submit.prevent="handleLogin" class="space-y-5">
          <!-- E-mail -->
          <div class="form-group">
            <label class="form-label-neo">{{ t.email }}</label>
            <div class="relative-group">
              <div class="input-wrapper">
                <Icon name="mail" :size="18" class="input-icon-left text-fg-subtle" />
                <input
                  v-model="form.username"
                  type="email"
                  :placeholder="t.emailPlaceholder"
                  class="brutal-input"
                  :disabled="loading"
                  autocomplete="off"
                />
              </div>
            </div>
          </div>

          <!-- Password -->
          <div class="form-group">
            <label class="form-label-neo">{{ t.password }}</label>
            <div class="relative-group">
              <div class="input-wrapper">
                <Icon name="lock" :size="18" class="input-icon-left text-fg-subtle" />
                <input
                  v-model="form.password"
                  :type="showPassword ? 'text' : 'password'"
                  :placeholder="t.passwordPlaceholder"
                  class="brutal-input pr-10"
                  :disabled="loading"
                  autocomplete="off"
                />
                <button
                  type="button"
                  class="input-icon-right text-fg-subtle hover:text-fg"
                  @click="showPassword = !showPassword"
                  tabindex="-1"
                >
                  <Icon :name="showPassword ? 'eye-off' : 'eye'" :size="18" />
                </button>
              </div>
            </div>
          </div>

          <!-- Submit Button -->
          <div class="pt-2">
            <button type="submit" class="btn-authenticate" :disabled="loading">
              <span>{{ loading ? '...' : t.login }}</span>
              <Icon name="arrow-right" :size="18" class="btn-arrow" />
            </button>
          </div>
        </form>

        <div class="card-footer-neo">
          <p class="text-xs text-fg-muted">
            {{ t.adminLink1 }}
            <a href="/login" class="text-action font-bold hover:underline">
              {{ t.adminLink2 }}
            </a>
            {{ t.adminLink3 }}
          </p>
          <p class="secure-system-text">
            {{ t.secureSystem }}
          </p>
        </div>
      </div>

      <!-- Footer Info -->
      <p class="copyright-footer">&copy; {{ currentYear }} GO-POSTFIXADMIN. {{ appVersion }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLangStore } from '../../stores/lang'
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'
import http from '../../utils/http'
import { useToastStore } from '../../stores/toast'

const auth = useAuthStore()
const router = useRouter()
const toast = useToastStore()

const form = ref({ username: '', password: '' })
const loading = ref(false)
const showPassword = ref(false)
const langStore = useLangStore()
const activeLang = computed(() => langStore.lang) // saved choice (Settings), else the browser language
const appVersion = ref('1.0.0')
const currentYear = new Date().getFullYear()

const translations = {
  EN: {
    title: 'USER LOGIN',
    subtitle: 'USER PORTAL',
    email: 'E-MAIL',
    emailPlaceholder: 'user@domain.com',
    password: 'PASSWORD',
    passwordPlaceholder: '••••••••',
    login: 'LOGIN',
    adminLink1: 'Administrators should ',
    adminLink2: 'click here',
    adminLink3: ' to access the admin dashboard.',
    secureSystem: 'USER ACCESS PORTAL',
  },
  PT: {
    title: 'USER LOGIN',
    subtitle: 'PORTAL DO USUÁRIO',
    email: 'E-MAIL',
    emailPlaceholder: 'usuario@dominio.com',
    password: 'SENHA',
    passwordPlaceholder: '••••••••',
    login: 'LOGIN',
    adminLink1: 'Os administradores devem ',
    adminLink2: 'clicar aqui',
    adminLink3: ' para acessar o painel de controle.',
    secureSystem: 'PORTAL DE ACESSO DO USUÁRIO',
  },
  ES: {
    title: 'INICIAR SESIÓN',
    subtitle: 'PORTAL DE USUARIO',
    email: 'E-MAIL',
    emailPlaceholder: 'usuario@dominio.com',
    password: 'CONTRASEÑA',
    passwordPlaceholder: '••••••••',
    login: 'LOGIN',
    adminLink1: 'Los administradores deben ',
    adminLink2: 'hacer clic aquí',
    adminLink3: ' para acceder al panel de administración.',
    secureSystem: 'PORTAL DE ACCESO DE USUARIO',
  },
}

const t = computed(() => translations[activeLang.value as 'EN' | 'PT' | 'ES'])

async function handleLogin() {
  if (!form.value.username.trim()) {
    toast.error('Email is required')
    return
  }
  if (!form.value.password.trim()) {
    toast.error('Password is required')
    return
  }

  loading.value = true
  try {
    await auth.userLogin(form.value.username.trim(), form.value.password)
    router.push('/users/dashboard')
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || e.message || 'Login failed')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  // Autodetect browser language if supported

  // Fetch dynamic version from backend
  try {
    const { data } = await http.get(`${API_BASE}/version`)
    if (data && data.version) {
      appVersion.value = data.version
    }
  } catch (e) {
    // Fallback to static default if request fails
  }
})
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background-color: var(--color-surface-2);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 1.5rem;
  padding-top: 6vh;
  position: relative;
}

.dot-pattern {
  background-image: radial-gradient(circle, var(--color-field-line) 1.2px, transparent 1.2px);
  background-size: 1.5rem 1.5rem;
}

/* Floating Language Switcher */

.w-full {
  width: 100%;
}
.login-col {
  max-width: max(27.5rem, 23vw);
}

/* Brand Header */
.brand-header {
  text-align: center;
  margin-bottom: 1.75rem;
}

.brand-icon-box {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  background: var(--color-accent); /* Indigo color for user portal branding icon */
  border: 2px solid var(--color-line);
  box-shadow: var(--shadow-2);
  margin-bottom: 1rem;
  transition: transform 0.2s;
  cursor: pointer;
}

.brand-icon-box:hover {
  transform: translateY(-2px);
}

.brand-title-mono {
  font-size: 2.25rem;
  font-weight: 700;
  letter-spacing: -0.0625rem;
  color: var(--color-fg);
  margin: 0 0 0.375rem;
  line-height: 1;
  font-family: monospace;
}

.brand-subtitle-flat {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-fg-muted);
  letter-spacing: 0.03125rem;
  text-transform: uppercase;
}

/* Card */
.login-card-neo {
  background: var(--color-surface);
  border: 2px solid var(--color-line);
  padding: 2rem;
}

.neo-shadow {
  box-shadow: var(--shadow-4);
}

.card-heading-neo {
  font-size: 1.25rem;
  font-weight: 800;
  letter-spacing: 0.0625rem;
  color: var(--color-fg);
  margin: 0 0 1.5rem;
  text-transform: uppercase;
}

/* Form inputs */
.form-group {
  margin-bottom: 1.25rem;
  display: flex;
  flex-direction: column;
}

.form-label-neo {
  display: block;
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.03125rem;
  color: var(--color-fg);
  margin-bottom: 0.5rem;
  text-transform: uppercase;
}

.relative-group {
  position: relative;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.brutal-input {
  width: 100%;
  height: 2.875rem;
  padding-left: 2.625rem;
  padding-right: 0.75rem;
  font-size: 0.875rem;
  border: 2px solid var(--color-line);
  border-radius: 0;
  background: var(--color-surface);
  color: var(--color-fg);
  transition:
    border-color 0.1s,
    box-shadow 0.1s;
  outline: none;
}

.brutal-input:focus {
  border-color: var(--color-accent);
  box-shadow: 4px 4px 0px var(--color-accent);
}

.brutal-input:disabled {
  background: var(--color-surface-2);
  opacity: 0.7;
}

.input-icon-left {
  position: absolute;
  left: 0.875rem;
  pointer-events: none;
}

.input-icon-right {
  position: absolute;
  right: 0.75rem;
  background: none;
  border: none;
  padding: 0.25rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Authenticate Button */
.btn-authenticate {
  width: 100%;
  background: var(--color-accent); /* Indigo color for user portal login button */
  color: var(--color-on-action);
  border: 2px solid var(--color-line);
  font-weight: 700;
  padding: 0.875rem 1.25rem;
  font-size: 0.8125rem;
  letter-spacing: 0.0625rem;
  cursor: pointer;
  border-radius: 0;
  box-shadow: var(--shadow-3);
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
  text-transform: uppercase;
}

.btn-authenticate:hover:not(:disabled) {
  background: var(--color-surface);
  color: var(--color-accent);
}

.btn-authenticate:active:not(:disabled) {
  transform: translate(2px, 2px);
  box-shadow: none;
}

.btn-authenticate:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-arrow {
  margin-left: 0.5rem;
  transition: transform 0.2s;
}

.btn-authenticate:hover .btn-arrow {
  transform: translateX(3px);
}

.pt-2 {
  padding-top: 0.5rem;
}

/* Error Banner */
.error-banner-neo {
  background: var(--color-err-soft);
  border: 2px solid var(--color-err);
  color: var(--color-err);
  padding: 0.75rem 0.875rem;
  font-size: 0.8125rem;
  font-weight: 600;
  margin-bottom: 1.25rem;
  display: flex;
  align-items: center;
}

/* Card footer */
.card-footer-neo {
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 2px solid var(--color-surface-3);
  text-align: center;
}

.card-footer-neo p {
  margin: 0;
}

.text-xs {
  font-size: 0.75rem;
}
.text-fg-muted {
  color: var(--color-fg-muted);
}
.text-action {
  color: var(--color-accent);
}
.font-bold {
  font-weight: 700;
}

.card-footer-neo a:hover {
  text-decoration: underline;
}

.secure-system-text {
  font-size: 0.625rem;
  color: var(--color-fg-subtle);
  font-weight: 700;
  letter-spacing: 0.05rem;
  text-transform: uppercase;
  margin-top: 0.75rem !important;
}

/* Copyright Footer */
.copyright-footer {
  text-align: center;
  margin-top: 2rem;
  font-size: 0.6875rem;
  color: var(--color-fg-subtle);
  font-weight: 800;
  letter-spacing: 0.0625rem;
  text-transform: uppercase;
}
</style>
