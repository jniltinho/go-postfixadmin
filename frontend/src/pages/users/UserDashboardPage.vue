<template>
  <div class="user-portal-layout dot-pattern min-h-screen flex flex-col bg-login-page text-fg">
    <!-- Top Nav Bar -->
    <header class="h-14 bg-surface border-b-2 border-line flex items-center justify-between px-8 z-10">
      <div class="flex items-center">
        <a
          href="/users/dashboard"
          class="w-10 h-10 bg-accent border-2 border-line neo-shadow-subtle flex items-center justify-center mr-3 hover:-translate-y-0.5 hover:shadow-(--shadow-3) transition-all cursor-pointer"
        >
          <Icon name="user" class="text-on-action" :size="20" />
        </a>
        <a
          href="/users/dashboard"
          class="text-xl font-mono font-bold tracking-tight hover:text-icon-accent transition-colors"
          >{{ t.userPortal }}</a
        >
      </div>

      <div class="flex items-center space-x-4">
        <label class="flex items-center gap-1 text-[0.625rem] font-black uppercase tracking-widest text-fg-muted">
          <Icon name="palette" :size="14" />
          <span class="sr-only">{{ t.theme }}</span>
          <select
            :value="theme.current"
            :aria-label="t.theme"
            class="bg-surface text-fg border border-line-soft hover:border-line px-1 py-0.5 text-[0.6875rem] font-bold normal-case tracking-normal"
            @change="theme.select(($event.target as HTMLSelectElement).value)"
          >
            <option v-for="th in themes" :key="th.id" :value="th.id">
              {{ themeLabel(th.label, activeLang as Lang) }}
            </option>
          </select>
        </label>

        <!-- Language (the portal has no Settings page) -->
        <label class="flex items-center gap-1 text-[0.625rem] font-black uppercase tracking-widest text-fg-muted">
          <Icon name="languages" :size="14" />
          <select
            :value="activeLang"
            :aria-label="t.language"
            class="bg-surface text-fg border border-line-soft hover:border-line px-1 py-0.5 text-[0.6875rem] font-bold normal-case tracking-normal"
            @change="selectLanguage(($event.target as HTMLSelectElement).value as 'PT' | 'EN' | 'ES')"
          >
            <option value="EN">English</option>
            <option value="PT">Português</option>
            <option value="ES">Español</option>
          </select>
        </label>

        <div class="flex items-center space-x-3">
          <div class="text-right">
            <p class="text-xs font-black uppercase text-fg tracking-tight">
              {{ auth.user?.username?.toUpperCase() }}
            </p>
            <p class="text-[0.625rem] font-bold text-fg-subtle uppercase tracking-widest leading-none mt-0.5 font-mono">
              {{ t.userRole }}
            </p>
          </div>
          <div
            class="w-10 h-10 border-2 border-line bg-accent flex items-center justify-center text-on-action font-black shadow-subtle shrink-0"
          >
            <Icon name="user" :size="20" />
          </div>
        </div>
        <div class="h-8 w-1 bg-strong-fill mx-2"></div>
        <button
          @click="handleLogout"
          class="flex items-center py-2 px-4 border-2 border-transparent font-bold text-err hover:border-err hover:bg-err-soft transition-all cursor-pointer"
        >
          <Icon name="log-out" :size="20" class="mr-2" />
          <span>{{ t.logout }}</span>
        </button>
      </div>
    </header>

    <!-- Main Content Area -->
    <main class="flex-1 px-12 pt-6 pb-12 overflow-y-auto w-full">
      <div class="max-w-[max(72rem,60vw)] mx-auto">
        <!-- Title Block -->
        <div class="mb-10 text-left">
          <h2 class="text-4xl font-mono font-black uppercase tracking-tight text-fg mb-2">
            {{ t.dashboardTitle }}
          </h2>
          <p class="text-xs font-bold uppercase tracking-widest text-fg-subtle">
            {{ t.dashboardDesc }}
          </p>
        </div>

        <!-- User Info Card -->
        <div class="bg-surface border-2 border-line neo-shadow px-8 py-5 mb-8">
          <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div class="flex items-center">
              <div
                class="w-16 h-16 bg-accent border-2 border-line flex items-center justify-center text-on-action mr-6 shrink-0"
              >
                <Icon name="user" :size="32" />
              </div>
              <div>
                <h3 class="text-2xl font-mono font-black tracking-tight text-fg">
                  {{ userProfile.name || getFallbackName(auth.user?.username) }}
                </h3>
                <p class="text-sm font-bold text-fg-muted font-mono">
                  {{ userProfile.username || auth.user?.username }}
                </p>
              </div>
            </div>
            <div>
              <button @click="showVacationModal = true" class="btn-vacation">
                <Icon name="plane-takeoff" :size="18" class="mr-2" />
                {{ t.configureVacation }}
              </button>
            </div>
          </div>
        </div>

        <!-- Settings Cards Grid -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <!-- Configure Forwarding Card -->
          <div class="bg-surface border-2 border-line neo-shadow p-8 h-full flex flex-col">
            <h3 class="text-xl font-mono font-black uppercase tracking-tight text-fg mb-6 flex items-center">
              <Icon name="corner-up-right" :size="20" class="mr-2 text-fg" />
              {{ t.forwardingTitle }}
            </h3>
            <p class="text-xs text-fg-muted mb-6 leading-relaxed">
              {{ t.forwardingDesc }}
            </p>

            <form @submit.prevent="saveForwarding" class="space-y-6 flex-1 flex flex-col justify-between">
              <div class="form-group flex-1">
                <label class="block text-xs font-black uppercase tracking-widest text-fg mb-2">
                  {{ t.forwardTo }}
                </label>
                <textarea
                  v-model="forwardingForm.forwarding"
                  rows="4"
                  class="w-full px-4 py-3 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors resize-y text-sm bg-surface"
                  :placeholder="t.forwardToPlaceholder"
                  :disabled="savingForwarding"
                ></textarea>
                <p class="text-xs text-fg-muted leading-normal mt-2">
                  {{ t.forwardHelp1 }}
                  <br />{{ t.forwardHelp2 }}
                </p>
              </div>

              <button type="submit" class="btn-submit-blue" :disabled="savingForwarding">
                <Icon name="save" :size="20" class="mr-2" />
                <span>{{ savingForwarding ? '...' : t.saveForwarding }}</span>
              </button>
            </form>
          </div>

          <!-- Change Password Card -->
          <div class="bg-surface border-2 border-line neo-shadow p-8 h-full flex flex-col">
            <h3 class="text-xl font-mono font-black uppercase tracking-tight text-fg mb-6 flex items-center">
              <Icon name="key" :size="20" class="mr-2 text-fg" />
              {{ t.passwordTitle }}
            </h3>
            <p class="text-xs text-fg-muted mb-6 leading-relaxed">
              {{ t.passwordDesc }}
            </p>

            <form @submit.prevent="changePassword" class="space-y-4 flex-1 flex flex-col justify-between">
              <div class="space-y-4">
                <!-- Current Password -->
                <div class="form-group">
                  <label class="block text-xs font-black uppercase tracking-widest text-fg mb-2">
                    {{ t.currentPassword }} <span class="text-err">*</span>
                  </label>
                  <div class="relative group">
                    <div
                      class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-fg-subtle group-focus-within:text-icon-accent transition-colors"
                    >
                      <Icon name="lock" :size="20" />
                    </div>
                    <input
                      v-model="passwordForm.current_password"
                      type="password"
                      required
                      :placeholder="t.placeholderCurrentPassword"
                      class="w-full pl-11 pr-4 py-2 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors h-10 text-sm bg-surface"
                      :disabled="savingPassword"
                    />
                  </div>
                </div>

                <!-- New Password & Confirm Password -->
                <div class="mb-4">
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-2">
                    <div>
                      <label class="block text-xs font-black uppercase tracking-widest text-fg">
                        {{ t.newPassword }} <span class="text-err">*</span>
                      </label>
                    </div>
                    <div>
                      <label class="block text-xs font-black uppercase tracking-widest text-fg">
                        {{ t.confirmPassword }} <span class="text-err">*</span>
                      </label>
                    </div>
                  </div>

                  <div class="flex flex-col md:flex-row gap-2 items-stretch md:items-center">
                    <div class="relative flex-1 flex items-center">
                      <input
                        v-model="passwordForm.new_password"
                        :type="showNewPassword ? 'text' : 'password'"
                        required
                        :placeholder="t.placeholderNewPassword"
                        class="w-full h-10 px-3 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors text-sm pr-10 bg-surface"
                        :disabled="savingPassword"
                      />
                      <button
                        type="button"
                        @click="showNewPassword = !showNewPassword"
                        class="absolute right-2 top-1/2 -translate-y-1/2 text-fg-muted hover:text-icon-accent transition-colors cursor-pointer bg-none border-none p-1 flex items-center"
                      >
                        <Icon :name="showNewPassword ? 'eye-off' : 'eye'" :size="16" />
                      </button>
                    </div>

                    <div class="relative flex-1 flex items-center">
                      <input
                        v-model="passwordForm.confirm_password"
                        :type="showConfirmPassword ? 'text' : 'password'"
                        required
                        :placeholder="t.placeholderConfirmPassword"
                        class="w-full h-10 px-3 border-2 border-line focus:border-icon-accent focus:outline-none font-medium transition-colors text-sm pr-10 bg-surface"
                        :disabled="savingPassword"
                      />
                      <button
                        type="button"
                        @click="showConfirmPassword = !showConfirmPassword"
                        class="absolute right-2 top-1/2 -translate-y-1/2 text-fg-muted hover:text-icon-accent transition-colors cursor-pointer bg-none border-none p-1 flex items-center"
                      >
                        <Icon :name="showConfirmPassword ? 'eye-off' : 'eye'" :size="16" />
                      </button>
                    </div>

                    <button
                      type="button"
                      @click="generateRandomPassword"
                      class="bg-accent hover:bg-surface hover:text-accent text-on-action text-xs font-black px-3 h-10 border-2 border-line cursor-pointer uppercase tracking-widest flex items-center justify-center gap-1 shrink-0 transition-colors shadow-(--shadow-1)"
                      title="Generate password"
                    >
                      <Icon name="wand-2" :size="14" />
                    </button>
                  </div>
                </div>

                <!-- Password Strength Meter -->
                <div v-if="passwordForm.new_password" class="mb-4">
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-xs font-black uppercase tracking-widest text-fg">{{ t.strength }}</span>
                    <span
                      class="text-xs font-bold uppercase tracking-wider font-mono"
                      :style="{ color: strengthInfo.color }"
                      >{{ strengthInfo.text }}</span
                    >
                  </div>
                  <div class="h-2 bg-line-soft border border-line overflow-hidden">
                    <div
                      class="h-full transition-all duration-300"
                      :style="{ width: strengthInfo.percent, background: strengthInfo.color }"
                    ></div>
                  </div>
                </div>
              </div>

              <button type="submit" class="btn-submit-orange" :disabled="savingPassword">
                <Icon name="shield" :size="20" class="mr-2" />
                <span>{{ savingPassword ? '...' : t.changePasswordBtn }}</span>
              </button>
            </form>
          </div>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="bg-surface border-t-2 border-line mt-auto">
      <div class="py-4 px-8 text-center">
        <p class="text-[0.625rem] text-fg-subtle font-bold uppercase tracking-widest">
          &copy; {{ currentYear }} GO-POSTFIXADMIN. {{ appVersion }}
        </p>
      </div>
    </footer>

    <!-- Auto-Reply Vacation Modal -->
    <div v-if="showVacationModal" class="vacation-modal-backdrop flex items-start justify-center px-4 pt-[5vh]">
      <div class="vacation-modal-box bg-surface border-2 border-line w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Modal Header -->
        <header class="bg-accent px-6 py-4 flex items-center justify-between text-on-action border-b-2 border-line">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight flex items-center">
            <Icon name="plane-takeoff" :size="20" class="mr-2" />
            {{ t.vacationTitle }}
          </h3>
          <button @click="showVacationModal = false" class="text-on-action hover:text-on-action/80 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </header>

        <!-- Scrollable Body -->
        <div class="overflow-y-auto flex-1 p-6 space-y-5">
          <!-- Dates Card -->
          <div class="border-2 border-line p-4 space-y-4">
            <h4 class="text-xs font-mono font-black uppercase tracking-widest text-fg flex items-center">
              <Icon name="calendar" :size="14" class="mr-1.5" />
              {{ t.startDate }} / {{ t.endDate }}
            </h4>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label class="block text-[0.625rem] font-black uppercase tracking-widest text-fg-muted mb-1">
                  {{ t.startDate }}
                </label>
                <input
                  v-model="vacationForm.activefrom"
                  type="datetime-local"
                  required
                  class="w-full h-10 px-3 border-2 border-line focus:border-accent focus:outline-none font-medium transition-colors text-sm bg-surface"
                />
              </div>
              <div>
                <label class="block text-[0.625rem] font-black uppercase tracking-widest text-fg-muted mb-1">
                  {{ t.endDate }}
                </label>
                <input
                  v-model="vacationForm.activeuntil"
                  type="datetime-local"
                  required
                  class="w-full h-10 px-3 border-2 border-line focus:border-accent focus:outline-none font-medium transition-colors text-sm bg-surface"
                />
              </div>
            </div>
          </div>

          <!-- Message Config Card -->
          <div class="border-2 border-line p-4 space-y-4">
            <h4 class="text-xs font-mono font-black uppercase tracking-widest text-fg flex items-center">
              <Icon name="message-square-dashed" :size="14" class="mr-1.5" />
              {{ t.messageBody }}
            </h4>

            <!-- Reply Interval -->
            <div>
              <label class="block text-[0.625rem] font-black uppercase tracking-widest text-fg-muted mb-1">
                {{ t.replyOption }}
              </label>
              <select
                v-model="vacationForm.interval_time"
                class="w-full h-10 px-3 border-2 border-line focus:border-accent focus:outline-none font-medium transition-colors appearance-none bg-surface text-sm cursor-pointer"
              >
                <option :value="0">{{ t.replyOnce }}</option>
                <option :value="86400">{{ t.replyEveryDay }}</option>
                <option :value="604800">{{ t.replyEvery7Days }}</option>
              </select>
            </div>

            <!-- Subject -->
            <div>
              <label class="block text-[0.625rem] font-black uppercase tracking-widest text-fg-muted mb-1">
                {{ t.subject }}
              </label>
              <input
                v-model="vacationForm.subject"
                type="text"
                required
                class="w-full h-10 px-3 border-2 border-line focus:border-accent focus:outline-none font-medium transition-colors text-sm bg-surface"
              />
            </div>

            <!-- Message Body -->
            <div>
              <label class="block text-[0.625rem] font-black uppercase tracking-widest text-fg-muted mb-1">
                {{ t.messageBody }}
              </label>
              <textarea
                v-model="vacationForm.body"
                rows="4"
                required
                class="w-full px-3 py-2 border-2 border-line focus:border-accent focus:outline-none font-medium transition-colors resize-y text-sm bg-surface"
              ></textarea>
            </div>

            <!-- Checkbox: Active -->
            <div class="flex items-center gap-2 pt-1">
              <input
                v-model="vacationForm.active"
                type="checkbox"
                id="active-chk"
                class="w-4 h-4 cursor-pointer accent-brand-secondary"
              />
              <label for="active-chk" class="text-xs font-black text-fg cursor-pointer">
                {{ t.active }}
              </label>
            </div>
          </div>
        </div>

        <!-- Footer actions -->
        <footer class="border-t-2 border-line px-6 py-4 flex items-center justify-end gap-3 shrink-0">
          <button @click="showVacationModal = false" class="btn-cancel btn-modal-cancel">
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
import { useAuthStore } from '../../stores/auth'
import { useThemeStore } from '../../stores/theme'
import { themes } from '../../themes'
import { themeLabel } from '../settings/i18n'
import { useLangStore, type Lang } from '../../stores/lang'
import { useRouter } from 'vue-router'
import { useToastStore } from '../../stores/toast'
import { calcStrength, generatePassword as fillPassword } from '../../utils/password'
import http from '../../utils/http'

const auth = useAuthStore()
const theme = useThemeStore()
const langStore = useLangStore()
const router = useRouter()
const toast = useToastStore()

const activeLang = ref('EN')
const showVacationModal = ref(false)
const appVersion = ref('1.0.0')
const currentYear = new Date().getFullYear()

const userProfile = ref({ username: '', name: '' })
const forwardingForm = ref({ forwarding: '' })
const passwordForm = ref({ current_password: '', new_password: '', confirm_password: '' })
const vacationForm = ref({
  active: false,
  subject: 'Out of Office',
  body: 'I will be out of the office from <date> to <date>. In case of urgency, please contact <contact person>.',
  activefrom: '',
  activeuntil: '',
  interval_time: 0,
})

const showNewPassword = ref(false)
const showConfirmPassword = ref(false)

const savingForwarding = ref(false)
const savingPassword = ref(false)
const savingVacation = ref(false)

function getFallbackName(email?: string) {
  if (!email) return 'User'
  const parts = email.split('@')
  return parts[0].charAt(0).toUpperCase() + parts[0].slice(1)
}

const translations = {
  EN: {
    userPortal: 'User Portal',
    theme: 'Theme',
    language: 'Language',
    userRole: 'USER',
    dashboardTitle: 'MY ACCOUNT',
    dashboardDesc: 'MANAGE YOUR PASSWORD AND FORWARDING SETTINGS.',
    profileName: 'Name',
    profileEmail: 'Email',
    configureVacation: 'CONFIGURE VACATION',
    forwardingTitle: 'CONFIGURE FORWARDING',
    forwardingDesc: 'Configure where your emails should be delivered. Enter one address per line.',
    forwardTo: 'FORWARD TO',
    forwardToPlaceholder: 'user@domain.com',
    forwardHelp1: 'To receive locally and forward, include your own address in the list.',
    forwardHelp2: 'To disable forwarding, just leave your own address.',
    saveForwarding: 'SAVE FORWARDING',
    passwordTitle: 'CHANGE PASSWORD',
    passwordDesc: 'Update your mailbox password.',
    currentPassword: 'CURRENT PASSWORD',
    placeholderCurrentPassword: 'Your current password',
    newPassword: 'NEW PASSWORD',
    placeholderNewPassword: 'Min. 8 characters',
    confirmPassword: 'CONFIRM PASSWORD',
    placeholderConfirmPassword: 'Repeat new password',
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
    logout: 'Logout',
  },
  PT: {
    userPortal: 'Portal do Usuário',
    theme: 'Tema',
    language: 'Idioma',
    userRole: 'USUÁRIO',
    dashboardTitle: 'MINHA CONTA',
    dashboardDesc: 'GERENCIE SUAS CONFIGURAÇÕES DE SENHA E ENCAMINHAMENTO.',
    profileName: 'Nome',
    profileEmail: 'E-mail',
    configureVacation: 'CONFIGURAR FÉRIAS',
    forwardingTitle: 'CONFIGURAR ENCAMINHAMENTO',
    forwardingDesc: 'Configure para onde seus e-mails devem ser entregues. Insira um endereço por linha.',
    forwardTo: 'ENCAMINHAR PARA',
    forwardToPlaceholder: 'usuario@dominio.com',
    forwardHelp1: 'Para receber localmente e encaminhar, inclua seu próprio endereço na lista.',
    forwardHelp2: 'Para desativar o encaminhamento, basta deixar apenas seu próprio endereço.',
    saveForwarding: 'SALVAR ENCAMINHAMENTO',
    passwordTitle: 'ALTERAR SENHA',
    passwordDesc: 'Atualize a senha da sua caixa de correio.',
    currentPassword: 'SENHA ATUAL',
    placeholderCurrentPassword: 'Sua senha atual',
    newPassword: 'NOVA SENHA',
    placeholderNewPassword: 'Mín. 8 caracteres',
    confirmPassword: 'CONFIRMAR SENHA',
    placeholderConfirmPassword: 'Repita a nova senha',
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
    logout: 'Sair',
  },
  ES: {
    userPortal: 'Portal de Usuario',
    theme: 'Tema',
    language: 'Idioma',
    userRole: 'USUARIO',
    dashboardTitle: 'MI CUENTA',
    dashboardDesc: 'ADMINISTRE SUS CONFIGURACIONES DE CONTRASEÑA Y REENVÍO.',
    profileName: 'Nombre',
    profileEmail: 'Correo electrónico',
    configureVacation: 'CONFIGURAR VACACIONES',
    forwardingTitle: 'CONFIGURAR REENVÍO',
    forwardingDesc: 'Configure a dónde se devem entregar sus correos. Ingrese una dirección por línea.',
    forwardTo: 'REENVIAR A',
    forwardToPlaceholder: 'usuario@dominio.com',
    forwardHelp1: 'Para recibir localmente y reenviar, incluya su propia dirección en la lista.',
    forwardHelp2: 'Para deshabilitar el reenvío, simplemente deje su propia dirección.',
    saveForwarding: 'GUARDAR REENVÍO',
    passwordTitle: 'CAMBIAR CONTRASEÑA',
    passwordDesc: 'Actualice la contraseña de su buzón de correo.',
    currentPassword: 'CONTRASEÑA ACTUAL',
    placeholderCurrentPassword: 'Su contraseña actual',
    newPassword: 'NUEVA CONTRASEÑA',
    placeholderNewPassword: 'Mín. 8 caracteres',
    confirmPassword: 'CONFIRMAR CONTRASEÑA',
    placeholderConfirmPassword: 'Repita la nueva contraseña',
    strength: 'FUERZA',
    changePasswordBtn: 'CAMBIAR CONTRASEÑA',
    vacationTitle: 'CONFIGURACIÓN DE AUTO-RESPUESTA',
    startDate: 'DE FECHA',
    endDate: 'A FECHA',
    replyOption: 'OPCIÓN DE RESPUESTA',
    replyOnce: 'Responder una vez',
    replyEveryDay: 'Responder todos os días',
    replyEvery7Days: 'Responder cada 7 días',
    subject: 'ASUNTO',
    messageBody: 'CUERPO DEL MENSAJE',
    active: 'Activar Auto-Respuesta',
    cancel: 'CANCELAR',
    save: 'GUARDAR MENSAJE',
    successForwarding: 'Reenvío actualizado con éxito',
    successPassword: 'Contraseña actualizada con éxito',
    successVacation: 'Auto-respuesta actualizada con éxito',
    logout: 'Cerrar Sesión',
  },
}

const t = computed(() => translations[activeLang.value as 'EN' | 'PT' | 'ES'])

// Password strength calculator
const strengthInfo = computed(() => {
  const pass = passwordForm.value.new_password
  const res = calcStrength(pass)
  return {
    percent: `${res.pct}%`,
    text: res.label,
    color: res.color,
  }
})

// Generate random strong password
function generateRandomPassword() {
  const tempForm = { password: '', passwordConfirm: '' }
  fillPassword(tempForm)
  passwordForm.value.new_password = tempForm.password
  passwordForm.value.confirm_password = tempForm.passwordConfirm
  showNewPassword.value = true
  showConfirmPassword.value = true
}

// Select portal language
async function selectLanguage(lang: 'PT' | 'EN' | 'ES') {
  activeLang.value = lang
  await langStore.set(lang) // shared with the admin UI; also sets the backend lang cookie
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
    const profileRes = await http.get(`${API_BASE}/user/me`)
    userProfile.value = profileRes.data

    // Fetch forwarding alias
    const forwardingRes = await http.get(`${API_BASE}/user/forwarding`)
    // Replace commas with newlines for convenient textual layout
    forwardingForm.value.forwarding = forwardingRes.data.goto.split(',').join('\n')

    // Fetch vacation auto-reply settings
    const vacationRes = await http.get(`${API_BASE}/user/vacation`)
    vacationForm.value = vacationRes.data
  } catch (e) {
    toast.error('Failed to load user portal settings.')
  }
}

// Save Forwarding settings
async function saveForwarding() {
  savingForwarding.value = true
  try {
    await http.post(`${API_BASE}/user/forwarding`, {
      forwarding: forwardingForm.value.forwarding,
    })
    toast.success(t.value.successForwarding)
    await fetchUserData()
  } catch (e: any) {
    toast.error(e?.response?.data?.error?.message || e.message || 'Failed to update forwarding.')
  } finally {
    savingForwarding.value = false
  }
}

// Save New password
async function changePassword() {
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    toast.error('Passwords do not match.')
    return
  }

  savingPassword.value = true
  try {
    await http.post(`${API_BASE}/user/password`, {
      current_password: passwordForm.value.current_password,
      new_password: passwordForm.value.new_password,
      confirm_password: passwordForm.value.confirm_password,
    })
    toast.success(t.value.successPassword)
    passwordForm.value = { current_password: '', new_password: '', confirm_password: '' }
  } catch (e: any) {
    toast.error(e?.response?.data?.error || e.message || 'Failed to update password.')
  } finally {
    savingPassword.value = false
  }
}

// Save Vacation setting from modal
async function saveVacation() {
  savingVacation.value = true
  try {
    await http.post(`${API_BASE}/user/vacation`, vacationForm.value)
    toast.success(t.value.successVacation)
    showVacationModal.value = false
    await fetchUserData()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || e.message || 'Failed to update auto-reply settings.')
  } finally {
    savingVacation.value = false
  }
}

onMounted(async () => {
  // Autodetect browser language if supported
  activeLang.value = langStore.lang // saved choice, else the browser language

  // Fetch dynamic version from backend
  try {
    const { data } = await http.get(`${API_BASE}/version`)
    if (data && data.version) {
      appVersion.value = data.version
    }
  } catch (e) {
    // Fallback to static default if request fails
  }

  fetchUserData()
})
</script>

<style scoped>
.user-portal-layout {
  min-height: 100vh;
  background-color: var(--color-login-page);
  color: var(--color-fg);
  font-family: var(--default-font-family);
  display: flex;
  flex-direction: column;
}

.dot-pattern {
  background-image: radial-gradient(circle, var(--color-field-line) 1px, transparent 1px);
  background-size: 2rem 2rem;
}

.neo-shadow {
  box-shadow: var(--shadow-4);
}

.neo-shadow-subtle {
  box-shadow: var(--shadow-1);
}

/* Configure Vacation Button */
.btn-vacation {
  background: var(--color-accent);
  color: var(--color-on-action);
  border: 2px solid var(--color-line);
  font-weight: 900;
  font-size: 0.75rem;
  padding: 0.75rem 1.5rem;
  letter-spacing: 0.0625rem;
  cursor: pointer;
  box-shadow: var(--shadow-3);
  transition: all 0.15s;
  display: flex;
  align-items: center;
  text-transform: uppercase;
}

.btn-vacation:hover {
  background: var(--color-surface);
  color: var(--color-accent);
  transform: translate(-1px, -1px);
  box-shadow: var(--shadow-4);
}

.btn-vacation:active {
  transform: translate(0px, 0px);
  box-shadow: none;
}

/* Configure Forwarding Save Button */
.btn-submit-blue {
  background: var(--color-accent);
  color: var(--color-on-action);
  border: 2px solid var(--color-line);
  font-weight: 900;
  font-size: 0.8125rem;
  padding: 0.875rem 1.25rem;
  letter-spacing: 0.0625rem;
  cursor: pointer;
  box-shadow: var(--shadow-3);
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
  text-transform: uppercase;
  width: 100%;
}

.btn-submit-blue:hover:not(:disabled) {
  background: var(--color-surface);
  color: var(--color-accent);
  transform: translate(-1px, -1px);
  box-shadow: var(--shadow-4);
}

.btn-submit-blue:active:not(:disabled) {
  transform: translate(0px, 0px);
  box-shadow: none;
}

/* Change Password Save Button */
.btn-submit-orange {
  background: var(--color-warn);
  color: var(--color-on-action);
  border: 2px solid var(--color-line);
  font-weight: 900;
  font-size: 0.8125rem;
  padding: 0.875rem 1.25rem;
  letter-spacing: 0.0625rem;
  cursor: pointer;
  box-shadow: var(--shadow-3);
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
  text-transform: uppercase;
  width: 100%;
}

.btn-submit-orange:hover:not(:disabled) {
  background: var(--color-surface);
  color: var(--color-warn);
  transform: translate(-1px, -1px);
  box-shadow: var(--shadow-4);
}

.btn-submit-orange:active:not(:disabled) {
  transform: translate(0px, 0px);
  box-shadow: none;
}

/* Vacation Modal backdrop and box styles */
.vacation-modal-backdrop {
  position: fixed;
  inset: 0;
  background-color: color-mix(in srgb, var(--color-shadow) 50%, transparent);
  z-index: 1000;
  backdrop-filter: blur(1px);
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

.btn-modal-cancel {
  background: var(--color-surface);
  color: var(--color-fg);
  border: 2px solid var(--color-line);
  font-weight: 900;
  font-size: 0.75rem;
  padding: 0.75rem 1.5rem;
  letter-spacing: 0.03125rem;
  cursor: pointer;
  box-shadow: var(--shadow-2);
  transition: all 0.15s;
  display: flex;
  align-items: center;
  text-transform: uppercase;
}

.btn-modal-cancel:hover {
  background: var(--color-surface-2);
  transform: translate(-2px, -2px);
  box-shadow: var(--shadow-3);
}

.btn-modal-cancel:active {
  transform: translate(0px, 0px);
  box-shadow: none;
}

.btn-modal-save {
  background: var(--color-accent);
  color: var(--color-on-action);
  border: 2px solid var(--color-line);
  font-weight: 900;
  font-size: 0.75rem;
  padding: 0.75rem 1.5rem;
  letter-spacing: 0.03125rem;
  cursor: pointer;
  box-shadow: var(--shadow-3);
  transition: all 0.15s;
  display: flex;
  align-items: center;
  text-transform: uppercase;
}

.btn-modal-save:hover:not(:disabled) {
  background: var(--color-surface);
  color: var(--color-accent);
  transform: translate(-1px, -1px);
  box-shadow: var(--shadow-4);
}

.btn-modal-save:active:not(:disabled) {
  transform: translate(0px, 0px);
  box-shadow: none;
}
</style>
