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
        component: () => import('../pages/dashboard/DashboardPage.vue'),
        meta: { title: 'Dashboard' }
      },
      {
        path: 'domains',
        name: 'Domains',
        component: () => import('../pages/domains/DomainsPage.vue'),
        meta: { title: 'Domains' }
      },
      {
        path: 'mailboxes',
        name: 'Mailboxes',
        component: () => import('../pages/mailboxes/MailboxesPage.vue'),
        meta: { title: 'Mailboxes' }
      },
      {
        path: 'aliases',
        name: 'Aliases',
        component: () => import('../pages/aliases/AliasesPage.vue'),
        meta: { title: 'Aliases' }
      },
      {
        path: 'alias-domains',
        name: 'AliasDomains',
        component: () => import('../pages/aliasdomains/AliasDomainsPage.vue'),
        meta: { title: 'Domain Aliases' }
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('../pages/logs/LogsPage.vue'),
        meta: { title: 'Logs' }
      },
      {
        path: 'maillog',
        name: 'MailLog',
        component: () => import('../pages/maillog/MailLogPage.vue'),
        meta: { title: 'MailLog' }
      },
      {
        path: 'admins',
        name: 'Admins',
        component: () => import('../pages/admins/AdminsPage.vue'),
        meta: { title: 'Administrators' }
      },
      {
        path: 'transports',
        name: 'Transports',
        component: () => import('../pages/transports/TransportsPage.vue'),
        meta: { title: 'Transport List' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('../pages/SettingsPage.vue'),
        meta: { title: 'Settings' }
      }
    ]
  },
  {
    path: '/users/login',
    name: 'UserLogin',
    component: () => import('../pages/users/UserLoginPage.vue')
  },
  {
    path: '/users/dashboard',
    name: 'UserDashboard',
    component: () => import('../pages/users/UserDashboardPage.vue'),
    meta: { requiresUserAuth: true, title: 'My Account' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guard
router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()

  if (to.meta.requiresUserAuth && (!auth.isAuthenticated || auth.user?.type !== 'mailbox')) {
    next({ name: 'UserLogin' })
  } else if (to.meta.requiresAuth && (!auth.isAuthenticated || auth.user?.type !== 'admin')) {
    next({ name: 'Login' })
  } else if (to.name === 'Login' && auth.isAuthenticated && auth.user?.type === 'admin') {
    next({ name: 'Dashboard' })
  } else if (to.name === 'UserLogin' && auth.isAuthenticated && auth.user?.type === 'mailbox') {
    next({ name: 'UserDashboard' })
  } else {
    next()
  }
})

router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} — Go-PostfixAdmin` : 'Go-PostfixAdmin'
})

export default router
