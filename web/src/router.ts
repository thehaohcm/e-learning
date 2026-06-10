import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from './pages/Dashboard.vue'
import Learn from './pages/Learn.vue'
import Review from './pages/Review.vue'
import Quiz from './pages/Quiz.vue'
import Writing from './pages/Writing.vue'
import Search from './pages/Search.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: Dashboard },
    { path: '/dashboard', component: Dashboard },
    { path: '/learn', component: Learn },
    { path: '/review', component: Review },
    { path: '/quiz', component: Quiz },
    { path: '/writing', component: Writing },
    { path: '/search', component: Search },
  ]
})