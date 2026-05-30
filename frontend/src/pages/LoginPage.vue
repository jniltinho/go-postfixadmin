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
          <Icon name="mail" :size="32" color="white" />
        </div>
        <h1 class="brand-title-mono">Go-PostfixAdmin</h1>
        <p class="brand-subtitle-flat">{{ t.adminSubtitle }}</p>
      </div>

      <!-- Square Card -->
      <div class="login-card-neo neo-shadow">
        <h2 class="card-heading-neo">{{ t.title }}</h2>

        <!-- Error Message -->
        <form @submit.prevent="handleLogin" class="space-y-5">
          <!-- Username -->
          <div class="form-group">
            <label class="form-label-neo">{{ t.username }}</label>
            <div class="relative-group">
              <div class="input-wrapper">
                <Icon name="mail" :size="18" class="input-icon-left text-[#94a3b8]" />
                <input
                  v-model="form.username"
                  type="email"
                  :placeholder="t.usernamePlaceholder"
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
                <Icon name="lock" :size="18" class="input-icon-left text-[#94a3b8]" />
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
              <span>{{ loading ? '...' : t.authenticate }}</span>
              <Icon name="arrow-right" :size="18" class="btn-arrow" />
            </button>
          </div>
        </form>

        <div class="card-footer-neo">
          <p class="text-xs text-gray-500">
            {{ t.userLink1 }}
            <a href="/users/login" class="text-brand-primary font-bold hover:underline">
              {{ t.userLink2 }}
            </a>
            {{ t.userLink3 }}
          </p>
          <p class="secure-system-text">
            {{ t.secureSystem }}
          </p>
        </div>
      </div>

      <!-- Footer Info -->
      <p class="copyright-footer">
        &copy; {{ currentYear }} Go-Postfixadmin. {{ appVersion }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import http from '../utils/http'
import { useToastStore } from '../stores/toast'

const auth = useAuthStore()
const router = useRouter()
const toast = useToastStore()

const form = ref({ username: '', password: '' })
const loading = ref(false)
const showPassword = ref(false)
const activeLang = ref('EN')
const appVersion = ref('1.0.0')
const currentYear = new Date().getFullYear()

const translations = {
  EN: {
    title: 'LOGIN',
    adminSubtitle: 'ADMINISTRATIVE CONTROL PANEL',
    username: 'USERNAME',
    usernamePlaceholder: 'admin@domain.com',
    password: 'PASSWORD',
    passwordPlaceholder: '••••••••',
    authenticate: 'AUTHENTICATE',
    userLink1: 'Users should ',
    userLink2: 'click here',
    userLink3: ' to access the user section.',
    secureSystem: 'SECURE ADMINISTRATIVE ACCESS SYSTEM'
  },
  PT: {
    title: 'ENTRAR',
    adminSubtitle: 'PAINEL DE CONTROLE ADMINISTRATIVO',
    username: 'USUÁRIO',
    usernamePlaceholder: 'admin@dominio.com',
    password: 'SENHA',
    passwordPlaceholder: '••••••••',
    authenticate: 'AUTENTICAR',
    userLink1: 'Os usuários devem ',
    userLink2: 'clicar aqui',
    userLink3: ' para acessar a seção do usuário.',
    secureSystem: 'SISTEMA DE ACESSO ADMINISTRATIVO SEGURO'
  },
  ES: {
    title: 'INICIAR SESIÓN',
    adminSubtitle: 'PANEL DE CONTROL ADMINISTRATIVO',
    username: 'USUARIO',
    usernamePlaceholder: 'admin@dominio.com',
    password: 'CONTRASEÑA',
    passwordPlaceholder: '••••••••',
    authenticate: 'AUTENTICAR',
    userLink1: 'Los usuarios deben ',
    userLink2: 'hacer clic aquí',
    userLink3: ' para acceder a la sección de usuario.',
    secureSystem: 'SISTEMA DE ACCESO ADMINISTRATIVO SEGURO'
  }
}

const t = computed(() => translations[activeLang.value as 'EN' | 'PT' | 'ES'])

async function selectLanguage(lang: 'PT' | 'EN' | 'ES') {
  activeLang.value = lang
  try {
    await http.get(`/lang/${lang.toLowerCase()}`)
  } catch (e) {
    // Ignore language set backend error
  }
}

async function handleLogin() {
  if (!form.value.username.trim()) {
    toast.error('Username is required')
    return
  }
  if (!form.value.password.trim()) {
    toast.error('Password is required')
    return
  }

  loading.value = true
  try {
    await auth.login(form.value.username.trim(), form.value.password)
    router.push('/dashboard')
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || e.message || 'Login failed')
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
  background: #3B82F6 !important;
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
  background: #3B82F6;
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

/* Migration note: Form inputs converted from Quasar q-input to native brutal inputs. */
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
  border-color: #3B82F6;
  box-shadow: 4px 4px 0px #3B82F6;
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
  background: #3B82F6;
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
  color: #3B82F6;
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
.text-brand-primary { color: #3B82F6; }
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
