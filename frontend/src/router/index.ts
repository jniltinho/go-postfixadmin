import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../pages/LoginPage.vue')
  },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/dashboard'
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../pages/DashboardPage.vue'),
        meta: { title: 'Dashboard' }
      },
      {
        path: 'domains',
        name: 'Domains',
        component: () => import('../pages/DomainsPage.vue'),
        meta: { title: 'Domains' }
      },
      {
        path: 'mailboxes',
        name: 'Mailboxes',
        component: () => import('../pages/MailboxesPage.vue'),
        meta: { title: 'Mailboxes' }
      },
      {
        path: 'aliases',
        name: 'Aliases',
        component: () => import('../pages/AliasesPage.vue'),
        meta: { title: 'Aliases' }
      },
      {
        path: 'alias-domains',
        name: 'AliasDomains',
        component: () => import('../pages/AliasDomainsPage.vue'),
        meta: { title: 'Domain Aliases' }
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('../pages/LogsPage.vue'),
        meta: { title: 'Logs' }
      },
      {
        path: 'maillog',
        name: 'MailLog',
        component: () => import('../pages/MailLogPage.vue'),
        meta: { title: 'MailLog' }
      },
      {
        path: 'admins',
        name: 'Admins',
        component: () => import('../pages/AdminsPage.vue'),
        meta: { title: 'Administrators' }
      },
      {
        path: 'transports',
        name: 'Transports',
        component: () => import('../pages/TransportsPage.vue'),
        meta: { title: 'Transport List' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('../pages/SettingsPage.vue'),
        meta: { title: 'Settings' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guard
router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    next({ name: 'Login' })
  } else if (to.name === 'Login' && auth.isAuthenticated) {
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
