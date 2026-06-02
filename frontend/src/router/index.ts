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
      { path: '', redirect: '/dashboard' },
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
        meta: { title: 'Administrators', requirePermission: 'admins:read' }
      },
      {
        path: 'transports',
        name: 'Transports',
        component: () => import('../pages/transports/TransportsPage.vue'),
        meta: { title: 'Transport List', requirePermission: 'transports:read' }
      },
      {
        path: 'apikeys',
        name: 'APIKeys',
        component: () => import('../pages/APIKeysPage.vue'),
        meta: { title: 'API Keys', requirePermission: 'apikeys:read' }
      },
      {
        path: 'roles',
        name: 'Roles',
        component: () => import('../pages/rbac/RoleManagementPage.vue'),
        meta: { title: 'Role Management', requirePermission: 'admins:write' }
      },
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

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()

  if (to.meta.requiresUserAuth && (!auth.isAuthenticated || auth.user?.type !== 'mailbox')) {
    return next({ name: 'UserLogin' })
  }

  if (to.meta.requiresAuth && (!auth.isAuthenticated || auth.user?.type !== 'admin')) {
    return next({ name: 'Login' })
  }

  // Permission-based guard: redirect to dashboard with a toast if access is denied.
  const requiredPerm = to.meta.requirePermission as string | undefined
  if (requiredPerm && !auth.hasPermission(requiredPerm)) {
    return next({ name: 'Dashboard' })
  }

  if (to.name === 'Login' && auth.isAuthenticated && auth.user?.type === 'admin') {
    return next({ name: 'Dashboard' })
  }

  if (to.name === 'UserLogin' && auth.isAuthenticated && auth.user?.type === 'mailbox') {
    return next({ name: 'UserDashboard' })
  }

  next()
})

router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} — Go-PostfixAdmin` : 'Go-PostfixAdmin'
})

export default router
