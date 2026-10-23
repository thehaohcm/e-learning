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
    { path: '/reading', component: () => import('./pages/Reading.vue') }, // ?id=
    { path: '/practice', component: () => import('./pages/PracticeHub.vue') },
    { path: '/practice/forgotten', component: () => import('./pages/Practice.vue') }, // ?ids=1,2,3, default: today's forgotten words
    { path: '/writing', component: () => import('./pages/Writing.vue') },
    // search is part of Browse now ("All decks")
    { path: '/search', redirect: (to) => ({ path: '/learn', query: { deck_id: 'all', q: to.query.q } }) },
    { path: '/devices', component: () => import('./pages/Devices.vue') },
    { path: '/share', component: () => import('./pages/Share.vue') }, // PWA share target (Android): ?text=&title=&url=
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
