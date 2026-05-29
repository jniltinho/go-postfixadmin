<template>
  <div class="user-portal-layout dot-pattern min-h-screen">
    <!-- Top Nav Bar -->
    <header class="top-nav neo-shadow-sm border-b-2 border-[#1E293B]">
      <div class="flex items-center justify-between px-6 h-16 max-w-6xl mx-auto w-full">
        <!-- Logo -->
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 bg-brand-primary border-2 border-[#1E293B] shadow-[1px_1px_0px_#1E293B] flex items-center justify-center text-white">
            <Icon name="user" :size="16" />
          </div>
          <span class="font-mono font-black text-lg tracking-tight text-[#1E293B]">User Portal</span>
        </div>

        <!-- Right Side Nav Actions -->
        <div class="flex items-center gap-4">
          <!-- Language Selector -->
          <div class="flex items-center space-x-1.5">
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

          <!-- User Info Badge -->
          <div class="user-badge-box">
            <div class="text-[10px] font-black text-[#1E293B] uppercase font-mono truncate max-w-[150px]">
              {{ auth.user?.username }}
            </div>
            <div class="text-[8px] font-bold text-gray-500 uppercase tracking-wider -mt-0.5">
              USER
            </div>
          </div>

          <!-- Logout Button -->
          <button @click="handleLogout" class="logout-btn flex items-center gap-1.5">
            <Icon name="log-out" :size="16" />
            <span class="font-black text-xs uppercase tracking-wider">{{ t.logout }}</span>
          </button>
        </div>
      </div>
    </header>

    <!-- Main Content Area -->
    <main class="max-w-4xl mx-auto px-6 py-10 w-full">
      <!-- Title Block -->
      <div class="mb-8">
        <h2 class="text-4xl font-mono font-black uppercase tracking-tight text-[#1E293B] mb-2">
          MY ACCOUNT
        </h2>
        <p class="text-xs font-bold uppercase tracking-widest text-gray-400">
          MANAGE YOUR PASSWORD AND FORWARDING SETTINGS.
        </p>
      </div>

      <!-- User Info Banner -->
      <div class="bg-white border-4 border-[#1E293B] shadow-[2px_2px_0px_#1E293B] px-8 py-5 mb-8 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div class="flex items-center">
          <div class="w-14 h-14 bg-brand-primary border-2 border-[#1E293B] flex items-center justify-center text-white mr-5 shrink-0 shadow-[1px_1px_0px_#1E293B]">
            <Icon name="user" :size="24" />
          </div>
          <div>
            <h3 class="text-2xl font-mono font-black tracking-tight text-[#1E293B]">
              {{ userProfile.name || 'User' }}
            </h3>
            <p class="text-sm font-bold text-gray-500 font-mono">
              {{ auth.user?.username }}
            </p>
          </div>
        </div>
        <button @click="showVacationModal = true" class="btn-vacation">
          <Icon name="plane" :size="18" class="mr-2" />
          {{ t.configureVacation }}
        </button>
      </div>

      <!-- Notification Banners -->
      <div v-if="successMsg" class="banner-success mb-6">
        <Icon name="check-circle" :size="16" class="mr-2" />
        {{ successMsg }}
      </div>
      <div v-if="errorMsg" class="banner-error mb-6">
        <Icon name="alert-triangle" :size="16" class="mr-2" />
        {{ errorMsg }}
      </div>

      <!-- Settings Cards Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
        <!-- Configure Forwarding Card -->
        <div class="bg-white border-4 border-[#1E293B] shadow-[2px_2px_0px_#1E293B] p-8 flex flex-col h-full">
          <h3 class="text-xl font-mono font-black uppercase tracking-tight text-[#1E293B] mb-4 flex items-center">
            <Icon name="corner-up-right" :size="20" class="mr-2 text-[#1E293B]" />
            {{ t.forwardingTitle }}
          </h3>
          <p class="text-xs text-gray-500 mb-6 leading-relaxed">
            {{ t.forwardingDesc }}
          </p>

          <form @submit.prevent="saveForwarding" class="space-y-6 flex-1 flex flex-col justify-between">
            <div class="form-group flex-1">
              <label class="block text-xs font-black uppercase tracking-widest text-[#1E293B] mb-2">
                {{ t.forwardTo }}
              </label>
              <textarea
                v-model="forwardingForm.forwarding"
                rows="5"
                class="brutal-textarea"
                placeholder="user@domain.com"
                :disabled="savingForwarding"
              ></textarea>
              <p class="text-[10px] text-gray-400 leading-normal mt-2">
                {{ t.forwardHelp1 }}
                <br>{{ t.forwardHelp2 }}
              </p>
            </div>

            <button type="submit" class="btn-submit-blue" :disabled="savingForwarding">
              <Icon name="save" :size="18" class="mr-2" />
              <span>{{ savingForwarding ? '...' : t.saveForwarding }}</span>
            </button>
          </form>
        </div>

        <!-- Change Password Card -->
        <div class="bg-white border-4 border-[#1E293B] shadow-[2px_2px_0px_#1E293B] p-8 flex flex-col h-full">
          <h3 class="text-xl font-mono font-black uppercase tracking-tight text-[#1E293B] mb-4 flex items-center">
            <Icon name="key" :size="20" class="mr-2 text-[#1E293B]" />
            {{ t.passwordTitle }}
          </h3>
          <p class="text-xs text-gray-500 mb-6 leading-relaxed">
            {{ t.passwordDesc }}
          </p>

          <form @submit.prevent="changePassword" class="space-y-5 flex-1 flex flex-col justify-between">
            <div class="space-y-4">
              <!-- Current Password -->
              <div class="form-group">
                <label class="block text-xs font-black uppercase tracking-widest text-[#1E293B] mb-1.5">
                  {{ t.currentPassword }} <span class="text-red-500">*</span>
                </label>
                <div class="relative flex items-center">
                  <Icon name="lock" :size="16" class="absolute left-3.5 text-gray-400" />
                  <input
                    v-model="passwordForm.current_password"
                    type="password"
                    required
                    placeholder="Your current password"
                    class="brutal-input-sm pl-10"
                    :disabled="savingPassword"
                  />
                </div>
              </div>

              <!-- New Password & Confirm Password -->
              <div class="grid grid-cols-2 gap-3 mb-1">
                <div>
                  <label class="block text-[10px] font-black uppercase tracking-widest text-[#1E293B] mb-1.5">
                    {{ t.newPassword }} <span class="text-red-500">*</span>
                  </label>
                </div>
                <div>
                  <label class="block text-[10px] font-black uppercase tracking-widest text-[#1E293B] mb-1.5">
                    {{ t.confirmPassword }} <span class="text-red-500">*</span>
                  </label>
                </div>
              </div>

              <div class="flex items-center gap-2">
                <div class="relative flex-1 flex items-center">
                  <input
                    v-model="passwordForm.new_password"
                    :type="showNewPassword ? 'text' : 'password'"
                    required
                    placeholder="Min. 8 characters"
                    class="brutal-input-sm pr-9 text-xs"
                    :disabled="savingPassword"
                  />
                  <button
                    type="button"
                    class="absolute right-2.5 text-gray-400 hover:text-[#1E293B]"
                    @click="showNewPassword = !showNewPassword"
                    tabindex="-1"
                  >
                    <Icon :name="showNewPassword ? 'eye-off' : 'eye'" :size="14" />
                  </button>
                </div>

                <div class="relative flex-1 flex items-center">
                  <input
                    v-model="passwordForm.confirm_password"
                    :type="showConfirmPassword ? 'text' : 'password'"
                    required
                    placeholder="Repeat password"
                    class="brutal-input-sm pr-9 text-xs"
                    :disabled="savingPassword"
                  />
                  <button
                    type="button"
                    class="absolute right-2.5 text-gray-400 hover:text-[#1E293B]"
                    @click="showConfirmPassword = !showConfirmPassword"
                    tabindex="-1"
                  >
                    <Icon :name="showConfirmPassword ? 'eye-off' : 'eye'" :size="14" />
                  </button>
                </div>

                <!-- Generate password button -->
                <button
                  type="button"
                  @click="generateRandomPassword"
                  class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-[#1E293B] h-9 w-9 flex items-center justify-center shadow-[1px_1px_0px_#1E293B] active:translate-x-0.5 active:translate-y-0.5 active:shadow-none transition-all shrink-0"
                  title="Generate dynamic password"
                >
                  <Icon name="wand" :size="14" />
                </button>
              </div>

              <!-- Password Strength Meter -->
              <div v-if="passwordForm.new_password" class="strength-meter pt-1">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-[10px] font-black uppercase tracking-widest text-[#1E293B]">{{ t.strength }}</span>
                  <span class="text-[10px] font-bold uppercase tracking-wider font-mono text-[#1E293B]">{{ strengthInfo.text }}</span>
                </div>
                <div class="h-1.5 bg-gray-200 border border-[#1E293B] overflow-hidden">
                  <div class="h-full transition-all duration-300" :class="strengthInfo.color" :style="{ width: strengthInfo.percent }"></div>
                </div>
              </div>
            </div>

            <button type="submit" class="btn-submit-orange mt-6" :disabled="savingPassword">
              <Icon name="shield" :size="18" class="mr-2" />
              <span>{{ savingPassword ? '...' : t.changePasswordBtn }}</span>
            </button>
          </form>
        </div>
      </div>
    </main>

    <!-- Auto-Reply Vacation Modal -->
    <div v-if="showVacationModal" class="vacation-modal-backdrop flex items-center justify-center px-4">
      <div class="vacation-modal-box bg-white border-4 border-[#1E293B] w-full max-w-2xl max-h-[90vh] flex flex-col shadow-[4px_4px_0px_#1E293B]">
        <!-- Modal Header -->
        <header class="bg-brand-primary px-6 py-4 flex items-center justify-between text-white border-b-2 border-[#1E293B]">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight flex items-center">
            <Icon name="plane" :size="20" class="mr-2" />
            {{ t.vacationTitle }}
          </h3>
          <button @click="showVacationModal = false" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </header>

        <!-- Scrollable Body -->
        <div class="overflow-y-auto flex-1 p-6 space-y-6">
          <!-- Dates Card -->
          <div class="border-2 border-[#1E293B] p-4 space-y-4">
            <h4 class="text-xs font-mono font-black uppercase tracking-widest text-[#1E293B] flex items-center">
              <Icon name="calendar" :size="14" class="mr-1.5" />
              {{ t.startDate }} / {{ t.endDate }}
            </h4>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label class="block text-[10px] font-black uppercase tracking-widest text-gray-500 mb-1">
                  {{ t.startDate }}
                </label>
                <input
                  v-model="vacationForm.activefrom"
                  type="datetime-local"
                  required
                  class="brutal-input-sm px-3 text-xs"
                />
              </div>
              <div>
                <label class="block text-[10px] font-black uppercase tracking-widest text-gray-500 mb-1">
                  {{ t.endDate }}
                </label>
                <input
                  v-model="vacationForm.activeuntil"
                  type="datetime-local"
                  required
                  class="brutal-input-sm px-3 text-xs"
                />
              </div>
            </div>
          </div>

          <!-- Message Config Card -->
          <div class="border-2 border-[#1E293B] p-4 space-y-4">
            <h4 class="text-xs font-mono font-black uppercase tracking-widest text-[#1E293B] flex items-center">
              <Icon name="message-square" :size="14" class="mr-1.5" />
              {{ t.messageBody }}
            </h4>

            <!-- Reply Interval -->
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-gray-500 mb-1">
                {{ t.replyOption }}
              </label>
              <select v-model="vacationForm.interval_time" class="brutal-select text-xs">
                <option :value="0">{{ t.replyOnce }}</option>
                <option :value="86400">{{ t.replyEveryDay }}</option>
                <option :value="604800">{{ t.replyEvery7Days }}</option>
              </select>
            </div>

            <!-- Subject -->
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-gray-500 mb-1">
                {{ t.subject }}
              </label>
              <input
                v-model="vacationForm.subject"
                type="text"
                required
                class="brutal-input-sm text-xs"
              />
            </div>

            <!-- Message Body -->
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-gray-500 mb-1">
                {{ t.messageBody }}
              </label>
              <textarea
                v-model="vacationForm.body"
                rows="4"
                required
                class="brutal-textarea text-xs"
              ></textarea>
            </div>

            <!-- Checkbox: Active -->
            <div class="flex items-center gap-2 pt-1">
              <input
                v-model="vacationForm.active"
                type="checkbox"
                id="active-chk"
                class="w-4 h-4 cursor-pointer accent-brand-primary"
              />
              <label for="active-chk" class="text-xs font-black text-[#1E293B] cursor-pointer">
                {{ t.active }}
              </label>
            </div>
          </div>
        </div>

        <!-- Footer actions -->
        <footer class="border-t-2 border-[#1E293B] px-6 py-4 flex items-center justify-end gap-3 shrink-0">
          <button @click="showVacationModal = false" class="btn-modal-cancel">
            <Icon name="x" :size="14" class="mr-1.5" />
            <span>{{ t.cancel }}</span>
          </button>
          <button @click="saveVacation" class="btn-modal-save" :disabled="savingVacation">
            <Icon name="save" :size="14" class="mr-1.5" />
            <span>{{ savingVacation ? '...' : t.save }}</span>
          </button>
        </footer>
      </div>
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

const activeLang = ref('EN')
const showVacationModal = ref(false)

const userProfile = ref({ username: '', name: '' })
const forwardingForm = ref({ forwarding: '' })
const passwordForm = ref({ current_password: '', new_password: '', confirm_password: '' })
const vacationForm = ref({
  active: false,
  subject: 'Out of Office',
  body: 'I will be out of the office from <date> to <date>. In case of urgency, please contact <contact person>.',
  activefrom: '',
  activeuntil: '',
  interval_time: 0
})

const showNewPassword = ref(false)
const showConfirmPassword = ref(false)

const savingForwarding = ref(false)
const savingPassword = ref(false)
const savingVacation = ref(false)

const successMsg = ref('')
const errorMsg = ref('')

const translations = {
  EN: {
    dashboardTitle: 'My Account',
    dashboardDesc: 'Manage your password and forwarding settings.',
    profileName: 'Name',
    profileEmail: 'Email',
    configureVacation: 'CONFIGURE VACATION',
    forwardingTitle: 'CONFIGURE FORWARDING',
    forwardingDesc: 'Configure where your emails should be delivered. Enter one address per line.',
    forwardTo: 'FORWARD TO',
    forwardHelp1: 'To receive locally and forward, include your own address in the list.',
    forwardHelp2: 'To disable forwarding, just leave your own address.',
    saveForwarding: 'SAVE FORWARDING',
    passwordTitle: 'CHANGE PASSWORD',
    passwordDesc: 'Update your mailbox password.',
    currentPassword: 'CURRENT PASSWORD',
    newPassword: 'NEW PASSWORD',
    confirmPassword: 'CONFIRM PASSWORD',
    strength: 'STRENGTH',
    changePasswordBtn: 'CHANGE PASSWORD',
    vacationTitle: 'AUTO-REPLY CONFIGURATION',
    startDate: 'START DATE',
    endDate: 'END DATE',
    replyOption: 'REPLY OPTION',
    replyOnce: 'Reply once',
    replyEveryDay: 'Reply every day',
    replyEvery7Days: 'Reply every 7 days',
    subject: 'SUBJECT',
    messageBody: 'MESSAGE BODY',
    active: 'Activate Auto-Reply',
    cancel: 'CANCEL',
    save: 'EDIT / SET MESSAGE',
    successForwarding: 'Forwarding updated successfully',
    successPassword: 'Password updated successfully',
    successVacation: 'Auto-reply updated successfully',
    logout: 'Logout'
  },
  PT: {
    dashboardTitle: 'Minha Conta',
    dashboardDesc: 'Gerencie suas configurações de senha e encaminhamento.',
    profileName: 'Nome',
    profileEmail: 'E-mail',
    configureVacation: 'CONFIGURAR FÉRIAS',
    forwardingTitle: 'CONFIGURAR ENCAMINHAMENTO',
    forwardingDesc: 'Configure para onde seus e-mails devem ser entregues. Insira um endereço por linha.',
    forwardTo: 'ENCAMINHAR PARA',
    forwardHelp1: 'Para receber localmente e encaminhar, inclua seu próprio endereço na lista.',
    forwardHelp2: 'Para desativar o encaminhamento, basta deixar apenas seu próprio endereço.',
    saveForwarding: 'SALVAR ENCAMINHAMENTO',
    passwordTitle: 'ALTERAR SENHA',
    passwordDesc: 'Atualize a senha da sua caixa de correio.',
    currentPassword: 'SENHA ATUAL',
    newPassword: 'NOVA SENHA',
    confirmPassword: 'CONFIRMAR SENHA',
    strength: 'FORÇA',
    changePasswordBtn: 'ALTERAR SENHA',
    vacationTitle: 'CONFIGURAÇÃO DE AUTO-RESPOSTA',
    startDate: 'DATA DE INÍCIO',
    endDate: 'DATA DE TÉRMINO',
    replyOption: 'OPÇÃO DE RESPOSTA',
    replyOnce: 'Responder uma vez',
    replyEveryDay: 'Responder todos os dias',
    replyEvery7Days: 'Responder a cada 7 dias',
    subject: 'ASSUNTO',
    messageBody: 'CORPO DA MENSAGEM',
    active: 'Ativar Auto-Resposta',
    cancel: 'CANCELAR',
    save: 'SALVAR MENSAGEM',
    successForwarding: 'Encaminhamento atualizado com sucesso',
    successPassword: 'Senha atualizada com sucesso',
    successVacation: 'Auto-resposta atualizada com sucesso',
    logout: 'Sair'
  },
  ES: {
    dashboardTitle: 'Mi Cuenta',
    dashboardDesc: 'Administre sus configuraciones de contraseña y reenvío.',
    profileName: 'Nombre',
    profileEmail: 'Correo electrónico',
    configureVacation: 'CONFIGURAR VACACIONES',
    forwardingTitle: 'CONFIGURAR REENVÍO',
    forwardingDesc: 'Configure a dónde se devem entregar sus correos. Ingrese una dirección por línea.',
    forwardTo: 'REENVIAR A',
    forwardHelp1: 'Para recibir localmente y reenviar, incluya su propia dirección en la lista.',
    forwardHelp2: 'Para deshabilitar el reenvío, simplemente deje su propia dirección.',
    saveForwarding: 'GUARDAR REENVÍO',
    passwordTitle: 'CAMBIAR CONTRASEÑA',
    passwordDesc: 'Actualice la contraseña de su buzón de correo.',
    currentPassword: 'CONTRASEÑA ACTUAL',
    newPassword: 'NUEVA CONTRASEÑA',
    confirmPassword: 'CONFIRMAR CONTRASEÑA',
    strength: 'FUERZA',
    changePasswordBtn: 'CAMBIAR CONTRASEÑA',
    vacationTitle: 'CONFIGURACIÓN DE AUTO-RESPUESTA',
    startDate: 'FECHA DE INICIO',
    endDate: 'FECHA DE TÉRMINO',
    replyOption: 'OPCIÓN DE RESPUESTA',
    replyOnce: 'Responder una vez',
    replyEveryDay: 'Responder todos los días',
    replyEvery7Days: 'Responder cada 7 dias',
    subject: 'ASUNTO',
    messageBody: 'CUERPO DEL MENSAJE',
    active: 'Activar Auto-Respuesta',
    cancel: 'CANCELAR',
    save: 'GUARDAR MENSAJE',
    successForwarding: 'Reenvío actualizado con éxito',
    successPassword: 'Contraseña actualizada con éxito',
    successVacation: 'Auto-respuesta actualizada con éxito',
    logout: 'Cerrar Sesión'
  }
}

const t = computed(() => translations[activeLang.value as 'EN' | 'PT' | 'ES'])

// Password strength calculator
const strengthInfo = computed(() => {
  const pass = passwordForm.value.new_password
  if (!pass) return { score: 0, text: '', color: 'bg-gray-200', percent: '0%' }
  let score = 0
  if (pass.length >= 8) score++
  if (/[A-Z]/.test(pass)) score++
  if (/[a-z]/.test(pass)) score++
  if (/[0-9]/.test(pass)) score++
  if (/[!@#$%^&*]/.test(pass)) score++

  if (score <= 2) return { score, text: 'Weak', color: 'bg-red-500', percent: '33%' }
  if (score <= 4) return { score, text: 'Medium', color: 'bg-orange-500', percent: '66%' }
  return { score, text: 'Strong', color: 'bg-green-500', percent: '100%' }
})

// Generate random strong password
function generateRandomPassword() {
  const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*'
  let pass = ''
  for (let i = 0; i < 12; i++) {
    pass += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  passwordForm.value.new_password = pass
  passwordForm.value.confirm_password = pass
}

// Select portal language
async function selectLanguage(lang: 'PT' | 'EN' | 'ES') {
  activeLang.value = lang
  try {
    await axios.get(`/lang/${lang.toLowerCase()}`)
  } catch (e) {
    // Ignore language set backend error
  }
}

// Logout virtual mailbox user
function handleLogout() {
  auth.logout()
  router.push('/users/login')
}

// Fetch user dashboard profile, forwarding and vacation details
async function fetchUserData() {
  try {
    // Fetch profile
    const profileRes = await axios.get(`${API_BASE}/user/me`)
    userProfile.value = profileRes.data

    // Fetch forwarding alias
    const forwardingRes = await axios.get(`${API_BASE}/user/forwarding`)
    // Replace commas with newlines for convenient textual layout
    forwardingForm.value.forwarding = forwardingRes.data.goto.split(',').join('\n')

    // Fetch vacation auto-reply settings
    const vacationRes = await axios.get(`${API_BASE}/user/vacation`)
    vacationForm.value = vacationRes.data
  } catch (e) {
    showError('Failed to load user portal settings.')
  }
}

// Save Forwarding settings
async function saveForwarding() {
  savingForwarding.value = true
  clearNotifications()
  try {
    await axios.post(`${API_BASE}/user/forwarding`, {
      forwarding: forwardingForm.value.forwarding
    })
    showSuccess(t.value.successForwarding)
    await fetchUserData()
  } catch (e: any) {
    showError(e?.response?.data?.error?.message || e.message || 'Failed to update forwarding.')
  } finally {
    savingForwarding.value = false
  }
}

// Save New password
async function changePassword() {
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    showError('Passwords do not match.')
    return
  }

  savingPassword.value = true
  clearNotifications()
  try {
    await axios.post(`${API_BASE}/user/password`, {
      current_password: passwordForm.value.current_password,
      new_password: passwordForm.value.new_password,
      confirm_password: passwordForm.value.confirm_password
    })
    showSuccess(t.value.successPassword)
    passwordForm.value = { current_password: '', new_password: '', confirm_password: '' }
  } catch (e: any) {
    showError(e?.response?.data?.error || e.message || 'Failed to update password.')
  } finally {
    savingPassword.value = false
  }
}

// Save Vacation setting from modal
async function saveVacation() {
  savingVacation.value = true
  clearNotifications()
  try {
    await axios.post(`${API_BASE}/user/vacation`, vacationForm.value)
    showSuccess(t.value.successVacation)
    showVacationModal.value = false
    await fetchUserData()
  } catch (e: any) {
    showError(e?.response?.data?.error || e.message || 'Failed to update auto-reply settings.')
  } finally {
    savingVacation.value = false
  }
}

function showSuccess(msg: string) {
  successMsg.value = msg
  setTimeout(() => { successMsg.value = '' }, 5000)
}

function showError(msg: string) {
  errorMsg.value = msg
  setTimeout(() => { errorMsg.value = '' }, 6000)
}

function clearNotifications() {
  successMsg.value = ''
  errorMsg.value = ''
}

onMounted(() => {
  // Autodetect browser language if supported
  const navLang = navigator.language.slice(0, 2).toUpperCase()
  if (['PT', 'ES'].includes(navLang)) {
    activeLang.value = navLang
  }

  fetchUserData()
})
</script>

<style scoped>
.user-portal-layout {
  min-height: 100vh;
  background-color: #F8FAFC;
  color: #1E293B;
  font-family: sans-serif;
  display: flex;
  flex-direction: column;
}

.dot-pattern {
  background-image: radial-gradient(circle, #cbd5e1 1.2px, transparent 1.2px);
  background-size: 24px 24px;
}

/* Header style matching admin layouts */
.top-nav {
  background: #ffffff;
}

.neo-shadow-sm {
  box-shadow: 2px 2px 0px #1E293B;
}

/* Language Switches */
.lang-btn {
  padding: 3px 8px;
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
  background: #6366F1 !important; /* Blue/Indigo style for active user portal flags selection */
  color: #ffffff !important;
  border-color: #1E293B !important;
  box-shadow: 1px 1px 0px #1E293B;
}

/* User identity container box */
.user-badge-box {
  background: #ffffff;
  border: 2px solid #1E293B;
  padding: 4px 12px;
  box-shadow: 2px 2px 0px #1E293B;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
}

/* Logout */
.logout-btn {
  background: none;
  border: none;
  color: #EF4444;
  cursor: pointer;
  transition: opacity 0.15s;
}

.logout-btn:hover {
  opacity: 0.8;
}

/* Configure Vacation Button */
.btn-vacation {
  background: #6366F1; /* Purple/Indigo color matching screenshot button */
  color: #ffffff;
  border: 2px solid #1E293B;
  font-weight: 900;
  font-size: 12px;
  padding: 10px 18px;
  letter-spacing: 1px;
  cursor: pointer;
  box-shadow: 3px 3px 0px #1E293B;
  transition: all .15s;
  display: flex;
  align-items: center;
  text-transform: uppercase;
}

.btn-vacation:hover {
  background: #ffffff;
  color: #6366F1;
}

.btn-vacation:active {
  transform: translate(2px, 2px);
  box-shadow: none;
}

/* Configure Forwarding Save Button */
.btn-submit-blue {
  background: #6366F1; /* Blue/Indigo color matching screenshot button */
  color: #ffffff;
  border: 2px solid #1E293B;
  font-weight: 900;
  font-size: 13px;
  padding: 12px 20px;
  letter-spacing: 1px;
  cursor: pointer;
  box-shadow: 3px 3px 0px #1E293B;
  transition: all .15s;
  display: flex;
  align-items: center;
  justify-content: center;
  text-transform: uppercase;
  width: 100%;
}

.btn-submit-blue:hover:not(:disabled) {
  background: #ffffff;
  color: #6366F1;
}

.btn-submit-blue:active:not(:disabled) {
  transform: translate(2px, 2px);
  box-shadow: none;
}

/* Change Password Save Button */
.btn-submit-orange {
  background: #F97316; /* Orange color matching screenshot button */
  color: #ffffff;
  border: 2px solid #1E293B;
  font-weight: 900;
  font-size: 13px;
  padding: 12px 20px;
  letter-spacing: 1px;
  cursor: pointer;
  box-shadow: 3px 3px 0px #1E293B;
  transition: all .15s;
  display: flex;
  align-items: center;
  justify-content: center;
  text-transform: uppercase;
  width: 100%;
}

.btn-submit-orange:hover:not(:disabled) {
  background: #ffffff;
  color: #F97316;
}

.btn-submit-orange:active:not(:disabled) {
  transform: translate(2px, 2px);
  box-shadow: none;
}

/* Form inputs & areas */
.brutal-textarea {
  width: 100%;
  padding: 12px;
  font-size: 13px;
  border: 2px solid #1E293B;
  border-radius: 0;
  background: #ffffff;
  color: #1E293B;
  transition: border-color 0.1s, box-shadow 0.1s;
  outline: none;
  font-family: monospace;
}

.brutal-textarea:focus {
  border-color: #6366F1;
  box-shadow: 3px 3px 0px #6366F1;
}

.brutal-input-sm {
  width: 100%;
  height: 38px;
  padding-left: 10px;
  padding-right: 10px;
  font-size: 13px;
  border: 2px solid #1E293B;
  border-radius: 0;
  background: #ffffff;
  color: #1E293B;
  transition: border-color 0.1s, box-shadow 0.1s;
  outline: none;
}

.brutal-input-sm:focus {
  border-color: #6366F1;
  box-shadow: 3px 3px 0px #6366F1;
}

/* Vacation Modal backdrop and box styles */
.vacation-modal-backdrop {
  position: fixed;
  inset: 0;
  background-color: rgba(30, 41, 59, 0.6);
  z-index: 1000;
  backdrop-filter: blur(1.5px);
}

.vacation-modal-box {
  animation: modalEntrance 0.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes modalEntrance {
  from {
    transform: scale(0.95);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

.brutal-select {
  width: 100%;
  height: 38px;
  padding-left: 10px;
  padding-right: 32px;
  font-size: 13px;
  border: 2px solid #1E293B;
  border-radius: 0;
  background: #ffffff;
  color: #1E293B;
  outline: none;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%231e293b' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e");
  background-repeat: no-repeat;
  background-position: right 10px center;
  background-size: 16px;
}

.brutal-select:focus {
  border-color: #6366F1;
  box-shadow: 3px 3px 0px #6366F1;
}

.btn-modal-cancel {
  background: #ffffff;
  color: #1E293B;
  border: 2px solid #1E293B;
  font-weight: 900;
  font-size: 12px;
  padding: 8px 16px;
  letter-spacing: 0.5px;
  cursor: pointer;
  box-shadow: 2px 2px 0px #1E293B;
  transition: all .15s;
  display: flex;
  align-items: center;
  text-transform: uppercase;
}

.btn-modal-cancel:hover {
  background: #F8FAFC;
}

.btn-modal-cancel:active {
  transform: translate(1px, 1px);
  box-shadow: none;
}

.btn-modal-save {
  background: #6366F1;
  color: #ffffff;
  border: 2px solid #1E293B;
  font-weight: 900;
  font-size: 12px;
  padding: 8px 16px;
  letter-spacing: 0.5px;
  cursor: pointer;
  box-shadow: 3px 3px 0px #1E293B;
  transition: all .15s;
  display: flex;
  align-items: center;
  text-transform: uppercase;
}

.btn-modal-save:hover:not(:disabled) {
  background: #ffffff;
  color: #6366F1;
}

.btn-modal-save:active:not(:disabled) {
  transform: translate(2px, 2px);
  box-shadow: none;
}

/* Success and error banner styling */
.banner-success {
  background: #F0FDF4;
  border: 2px solid #22C55E;
  color: #15803D;
  padding: 12px 16px;
  font-weight: 700;
  font-size: 13px;
  display: flex;
  align-items: center;
}

.banner-error {
  background: #FEF2F2;
  border: 2px solid #EF4444;
  color: #B91C1C;
  padding: 12px 16px;
  font-weight: 700;
  font-size: 13px;
  display: flex;
  align-items: center;
}
</style>
