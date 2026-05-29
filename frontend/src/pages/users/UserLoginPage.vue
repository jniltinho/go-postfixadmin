<template>
  <div class="login-page dot-pattern">

    <!-- Floating Language Switcher -->
    <div class="lang-bar">
      <button
        v-for="lang in ['PT', 'EN', 'ES']"
        :key="lang"
        class="lang-btn"
        :class="{ 'lang-active': activeLang === lang }"
        @click="selectLanguage(lang as 'PT' | 'EN' | 'ES')"
      >
        {{ lang }}
      </button>
    </div>

    <div class="w-full max-w-md">
      <!-- Brand Header -->
      <div class="brand-header">
        <div class="brand-icon-box">
          <Icon name="user" :size="32" color="white" />
        </div>
        <h1 class="brand-title-mono">Go-PostfixAdmin</h1>
        <p class="brand-subtitle-flat">{{ t.subtitle }}</p>
      </div>

      <!-- Square Card -->
      <div class="login-card-neo neo-shadow">
        <h2 class="card-heading-neo">{{ t.title }}</h2>

        <!-- Error Message -->
        <div v-if="error" class="error-banner-neo">
          <Icon name="alert-triangle" :size="16" class="mr-1.5 flex-shrink-0" />
          {{ error }}
        </div>

        <form @submit.prevent="handleLogin" class="space-y-5">
          <!-- E-mail -->
          <div class="form-group">
            <label class="form-label-neo">{{ t.email }}</label>
            <div class="relative-group">
              <div class="input-wrapper">
                <Icon name="mail" :size="18" class="input-icon-left text-[#94a3b8]" />
                <input
                  v-model="form.username"
                  type="text"
                  :placeholder="t.emailPlaceholder"
                  class="brutal-input"
                  :disabled="loading"
                  autocomplete="username"
                />
              </div>
            </div>
          </div>

          <!-- Password -->
          <div class="form-group">
            <label class="form-label-neo">{{ t.password }}</label>
            <div class="relative-group">
              <div class="input-wrapper">
                <Icon name="lock" :size="18" class="input-icon-left text-[#94a3b8]" />
                <input
                  v-model="form.password"
                  :type="showPassword ? 'text' : 'password'"
                  :placeholder="t.passwordPlaceholder"
                  class="brutal-input pr-10"
                  :disabled="loading"
                  autocomplete="current-password"
                />
                <button
                  type="button"
                  class="input-icon-right text-[#94a3b8] hover:text-[#1e293b]"
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
            <button
              type="submit"
              class="btn-authenticate"
              :disabled="loading"
            >
              <span>{{ loading ? '...' : t.login }}</span>
              <Icon name="arrow-right" :size="18" class="btn-arrow" />
            </button>
          </div>
        </form>

        <div class="card-footer-neo">
          <p class="text-xs text-gray-500">
            {{ t.adminLink1 }}
            <a href="/login" class="text-brand-primary font-bold hover:underline">
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
      <p class="copyright-footer">
        &copy; {{ currentYear }} GO-POSTFIXADMIN. {{ appVersion }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import axios from 'axios'

const auth = useAuthStore()
const router = useRouter()

const form = ref({ username: '', password: '' })
const loading = ref(false)
const error = ref('')
const showPassword = ref(false)
const activeLang = ref('EN')
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
    secureSystem: 'USER ACCESS PORTAL'
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
    secureSystem: 'PORTAL DE ACESSO DO USUÁRIO'
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
    secureSystem: 'PORTAL DE ACCESO DE USUARIO'
  }
}

const t = computed(() => translations[activeLang.value as 'EN' | 'PT' | 'ES'])

async function selectLanguage(lang: 'PT' | 'EN' | 'ES') {
  activeLang.value = lang
  try {
    await axios.get(`/lang/${lang.toLowerCase()}`)
  } catch (e) {
    // Ignore language set backend error
  }
}

async function handleLogin() {
  if (!form.value.username.trim()) {
    error.value = 'Email is required'
    return
  }
  if (!form.value.password.trim()) {
    error.value = 'Password is required'
    return
  }

  loading.value = true
  error.value = ''
  try {
    await auth.login(form.value.username.trim(), form.value.password)
    if (auth.user?.type === 'mailbox') {
      router.push('/users/dashboard')
    } else {
      // If an admin logs in here, redirect to admin dashboard
      router.push('/dashboard')
    }
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || e.message || 'Login failed'
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  // Autodetect browser language if supported
  const navLang = navigator.language.slice(0, 2).toUpperCase()
  if (['PT', 'ES'].includes(navLang)) {
    activeLang.value = navLang
  }

  // Fetch dynamic version from backend
  try {
    const { data } = await axios.get(`${API_BASE}/version`)
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
  background-color: #F8FAFC;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 24px;
  padding-top: 6vh;
  position: relative;
}

.dot-pattern {
  background-image: radial-gradient(circle, #cbd5e1 1.2px, transparent 1.2px);
  background-size: 24px 24px;
}

/* Floating Language Switcher */
.lang-bar {
  position: fixed;
  top: 24px;
  right: 24px;
  display: flex;
  gap: 8px;
  z-index: 100;
}

.lang-btn {
  padding: 4px 10px;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.8px;
  border: 1px solid #cbd5e1;
  background: #ffffff;
  color: #94a3b8;
  cursor: pointer;
  border-radius: 0;
  transition: all .15s;
  text-transform: uppercase;
}

.lang-btn:hover {
  border-color: #1E293B;
  color: #1E293B;
}

.lang-active {
  background: #6366F1 !important; /* Blue/Indigo color for user portal login button selection */
  color: #ffffff !important;
  border-color: #1E293B !important;
  box-shadow: 1px 1px 0px #1E293B;
}

.w-full { width: 100%; }
.max-w-md { max-width: 440px; }

/* Brand Header */
.brand-header {
  text-align: center;
  margin-bottom: 28px;
}

.brand-icon-box {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  background: #6366F1; /* Indigo color for user portal branding icon */
  border: 2px solid #1E293B;
  box-shadow: 2px 2px 0px #1E293B;
  margin-bottom: 16px;
  transition: transform 0.2s;
  cursor: pointer;
}

.brand-icon-box:hover {
  transform: translateY(-2px);
}

.brand-title-mono {
  font-size: 36px;
  font-weight: 700;
  letter-spacing: -1px;
  color: #1E293B;
  margin: 0 0 6px;
  line-height: 1;
  font-family: monospace;
}

.brand-subtitle-flat {
  font-size: 13px;
  font-weight: 600;
  color: #64748B;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

/* Card */
.login-card-neo {
  background: #ffffff;
  border: 2px solid #1E293B;
  padding: 32px;
}

.neo-shadow {
  box-shadow: 4px 4px 0px #1E293B;
}

.card-heading-neo {
  font-size: 20px;
  font-weight: 800;
  letter-spacing: 1px;
  color: #1E293B;
  margin: 0 0 24px;
  text-transform: uppercase;
}

/* Form inputs */
.form-group {
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
}

.form-label-neo {
  display: block;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.5px;
  color: #1E293B;
  margin-bottom: 8px;
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
  height: 46px;
  padding-left: 42px;
  padding-right: 12px;
  font-size: 14px;
  border: 2px solid #1E293B;
  border-radius: 0;
  background: #ffffff;
  color: #1E293B;
  transition: border-color 0.1s, box-shadow 0.1s;
  outline: none;
}

.brutal-input:focus {
  border-color: #6366F1;
  box-shadow: 4px 4px 0px #6366F1;
}

.brutal-input:disabled {
  background: #f8fafc;
  opacity: 0.7;
}

.input-icon-left {
  position: absolute;
  left: 14px;
  pointer-events: none;
}

.input-icon-right {
  position: absolute;
  right: 12px;
  background: none;
  border: none;
  padding: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Authenticate Button */
.btn-authenticate {
  width: 100%;
  background: #6366F1; /* Indigo color for user portal login button */
  color: #ffffff;
  border: 2px solid #1E293B;
  font-weight: 700;
  padding: 14px 20px;
  font-size: 13px;
  letter-spacing: 1px;
  cursor: pointer;
  border-radius: 0;
  box-shadow: 3px 3px 0px #1E293B;
  transition: all .15s;
  display: flex;
  align-items: center;
  justify-content: center;
  text-transform: uppercase;
}

.btn-authenticate:hover:not(:disabled) {
  background: #ffffff;
  color: #6366F1;
}

.btn-authenticate:active:not(:disabled) {
  transform: translate(2px, 2px);
  box-shadow: none;
}

.btn-authenticate:disabled {
  opacity: .6;
  cursor: not-allowed;
}

.btn-arrow {
  margin-left: 8px;
  transition: transform 0.2s;
}

.btn-authenticate:hover .btn-arrow {
  transform: translateX(3px);
}

.pt-2 {
  padding-top: 8px;
}

/* Error Banner */
.error-banner-neo {
  background: #FEF2F2;
  border: 2px solid #EF4444;
  color: #EF4444;
  padding: 12px 14px;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
}

/* Card footer */
.card-footer-neo {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 2px solid #F1F5F9;
  text-align: center;
}

.card-footer-neo p {
  margin: 0;
}

.text-xs { font-size: 12px; }
.text-gray-500 { color: #64748B; }
.text-brand-primary { color: #6366F1; }
.font-bold { font-weight: 700; }

.card-footer-neo a:hover {
  text-decoration: underline;
}

.secure-system-text {
  font-size: 10px;
  color: #94A3B8;
  font-weight: 700;
  letter-spacing: 0.8px;
  text-transform: uppercase;
  margin-top: 12px !important;
}

/* Copyright Footer */
.copyright-footer {
  text-align: center;
  margin-top: 32px;
  font-size: 11px;
  color: #94A3B8;
  font-weight: 800;
  letter-spacing: 1px;
  text-transform: uppercase;
}
</style>
