import { createRouter, createWebHistory } from 'vue-router'
import { auth } from './auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('./pages/Login.vue'), meta: { public: true } },
    { path: '/', component: () => import('./pages/Dashboard.vue') },
    { path: '/dashboard', redirect: '/' },
    { path: '/learn', component: () => import('./pages/Learn.vue') }, // ?deck_id=
    { path: '/review', component: () => import('./pages/Review.vue') }, // ?deck_id= / ?subject_id=
    { path: '/quiz', component: () => import('./pages/Quiz.vue') },
    { path: '/writing', component: () => import('./pages/Writing.vue') },
    { path: '/search', component: () => import('./pages/Search.vue') },
    { path: '/devices', component: () => import('./pages/Devices.vue') },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.beforeEach((to) => {
  if (!to.meta.public && !auth.token) {
    return { path: '/login', query: { next: to.fullPath } }
  }
  if (to.path === '/login' && auth.token) return '/'
})

export default router
