import { createRouter, createWebHistory } from 'vue-router'
import Learn from './pages/Learn.vue'
import Quiz from './pages/Quiz.vue'
import Dashboard from './pages/Dashboard.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: Dashboard },
    { path: '/learn', component: Learn },
    { path: '/quiz', component: Quiz }
  ]
})

