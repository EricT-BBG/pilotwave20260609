import { createRouter, createWebHistory } from 'vue-router';
import { resolveRouteAccess } from '../lib/route-guard';

const routes = [
  {
    path: '/',
    name: 'Landingpage',
    component: () => import('../views/Landingpage.vue')
  },
  {
    path: '/',
    name: 'Template',
    component: () => import('../components/Template.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: 'dashboard', name: 'Home', component: () => import('../views/Home.vue'), meta: { requiresAuth: true, hideNamespacePicker: true } },
      { path: 'users', name: 'Users', component: () => import('../views/user/User.vue'), meta: { requiresAuth: true, requiresAdmin: true } },
      { path: 'new/user', name: 'NewUser', component: () => import('../views/user/NewUser.vue'), meta: { requiresAuth: true, requiresAdmin: true } },
      { path: 'user/:id', name: 'UserDetail', component: () => import('../views/user/UserDetail.vue'), meta: { requiresAuth: true } },
      { path: 'gateways', name: 'Gateways', component: () => import('../views/gateway/Gateway.vue'), meta: { requiresAuth: true } },
      { path: 'tls-certificates', name: 'TLSCertificates', component: () => import('../views/gateway/TLSCertificates.vue'), meta: { requiresAuth: true } },
      { path: 'tls-certificates/:id', name: 'TLSCertificateDetail', component: () => import('../views/gateway/TLSCertificates.vue'), meta: { requiresAuth: true } },
      { path: 'new/gateway', name: 'NewGateway', component: () => import('../views/gateway/NewGateway.vue'), meta: { requiresAuth: true } },
      { path: 'gateway/:name', name: 'GatewayDetail', component: () => import('../views/gateway/GatewayDetail.vue'), meta: { requiresAuth: true } },
      { path: 'routers', name: 'Routers', component: () => import('../views/router/Router.vue'), meta: { requiresAuth: true } },
      { path: 'router/:name', name: 'RouterDetail', component: () => import('../views/router/RouterDetail.vue'), meta: { requiresAuth: true } },
      { path: 'new/router', name: 'NewRouter', component: () => import('../views/router/NewRouter.vue'), meta: { requiresAuth: true } },
      { path: 'requestauths', name: 'Requestauths', component: () => import('../views/auth/Requestauth.vue'), meta: { requiresAuth: true } },
      { path: 'requestauth/:name', name: 'RequestauthDetail', component: () => import('../views/auth/RequestauthDetail.vue'), meta: { requiresAuth: true } },
      { path: 'new/requestauth', name: 'NewRequestauth', component: () => import('../views/auth/NewRequestauth.vue'), meta: { requiresAuth: true } },
      { path: 'authpolicies', name: 'Authpolicy', component: () => import('../views/policy/Authpolicy.vue'), meta: { requiresAuth: true } },
      { path: 'authpolicy/:name', name: 'AuthpolicyDetail', component: () => import('../views/policy/AuthpolicyDetail.vue'), meta: { requiresAuth: true } },
      { path: 'new/authpolicy', name: 'NewAuthpolicy', component: () => import('../views/policy/NewAuthpolicy.vue'), meta: { requiresAuth: true } }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'PageNotfound',
    component: () => import('../views/Landingpage.vue')
  }
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
});

router.beforeEach((to) => {
  const access = resolveRouteAccess(to);
  return access === true ? true : access;
});

export default router;
