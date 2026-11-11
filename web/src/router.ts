import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import { auth } from './auth'
import { subjectSlug } from './composables/useSubject'

// The current topic shows as the first URL segment (e.g. /pentest/labs,
// /eng/review) so each topic has its own shareable/bookmarkable URLs.
export const SUBJECT_PREFIXES: Record<string, string> = {
  english: 'eng',
  pmp: 'pmp',
  pentest: 'pentest',
  aws: 'aws',
}
export const PREFIX_TO_SUBJECT: Record<string, string> = Object.fromEntries(
  Object.entries(SUBJECT_PREFIXES).map(([slug, prefix]) => [prefix, slug]),
)
const PREFIX_PATTERN = Object.values(SUBJECT_PREFIXES).join('|')

/** The URL prefix for whatever topic is currently selected (falls back to 'eng'). */
export function currentPrefix(): string {
  return SUBJECT_PREFIXES[subjectSlug.value] ?? 'eng'
}

function redirectToCurrentSubject(to: RouteLocationNormalized) {
  if (to.path === '/search') {
    return { path: `/${currentPrefix()}/learn`, query: { deck_id: 'all', q: to.query.q } }
  }
  const rest = to.path === '/' || to.path === '/dashboard' ? '' : to.path
  return { path: `/${currentPrefix()}${rest}`, query: to.query }
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('./pages/Login.vue'), meta: { public: true } },
    { path: '/share', component: () => import('./pages/Share.vue') }, // PWA share target (Android): ?text=&title=&url= — fixed URL, no topic prefix
    {
      // every real page lives under /<topic-prefix>/..., so the address bar
      // always shows which topic (English/PMP/Pentest/AWS) is open
      path: `/:subjectPrefix(${PREFIX_PATTERN})`,
      children: [
        { path: '', component: () => import('./pages/Dashboard.vue') },
        { path: 'learn', component: () => import('./pages/Learn.vue') }, // ?deck_id=
        { path: 'review', component: () => import('./pages/Review.vue') }, // ?deck_id= / ?subject_id=
        { path: 'quiz', component: () => import('./pages/Quiz.vue') },
        { path: 'reading', component: () => import('./pages/Reading.vue') }, // ?id=
        { path: 'practice', component: () => import('./pages/PracticeHub.vue') },
        { path: 'labs', component: () => import('./pages/Labs.vue') }, // ?id= (pentest step-by-step command practice)
        { path: 'practice/forgotten', component: () => import('./pages/Practice.vue') }, // ?ids=1,2,3, default: today's forgotten words
        { path: 'practice/check', component: () => import('./pages/WordQuestions.vue') }, // ?ids=1,2,3: AI multiple choice (shared with the device)
        { path: 'writing', component: () => import('./pages/Writing.vue') },
        { path: 'cheatsheet', component: () => import('./pages/CheatSheet.vue') }, // per-course notes + AI cheat sheet
        { path: 'roleplay', component: () => import('./pages/Roleplay.vue') }, // PMP: play the PM with an AI client (English speaking)
        { path: 'daily', component: () => import('./pages/DailyQuiz.vue') }, // PMP: Daily 20 exam-style questions
        { path: 'courses', component: () => import('./pages/Courses.vue') }, // admin: manage subjects/decks/cards
        { path: 'devices', component: () => import('./pages/Devices.vue') },
      ],
    },
    // every old/unprefixed link (router-link to="/review?...", a bookmark, a
    // shared notification link, /, /dashboard, /search) lands here and gets
    // sent to the same page under the current topic's prefix — so nothing
    // elsewhere in the app had to change to start using prefixed URLs.
    { path: '/:pathMatch(.*)*', redirect: redirectToCurrentSubject },
  ],
})

router.beforeEach((to) => {
  if (!to.meta.public && !auth.token) {
    return { path: '/login', query: { next: to.fullPath } }
  }
  if (to.path === '/login' && auth.token) return `/${currentPrefix()}`
})

export default router
