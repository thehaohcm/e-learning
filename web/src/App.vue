<template>
  <div id="app">
    <nav v-if="auth.token" class="navbar navbar-expand-lg navbar-dark shadow-lg modern-navbar">
      <div class="container">
        <router-link class="navbar-brand d-flex align-items-center gap-2 fw-bold" to="/">
          <div class="brand-icon">
            <i class="bi bi-mortarboard-fill"></i>
          </div>
          <div>
            <span class="brand-title">E-Learning</span>
            <small class="brand-subtitle">English · PMP</small>
          </div>
        </router-link>
        <button class="navbar-toggler border-0" type="button" @click="navOpen = !navOpen"
          :aria-expanded="navOpen" aria-label="Toggle navigation">
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" :class="{ show: navOpen }">
          <ul class="navbar-nav ms-auto align-items-lg-center gap-1">
            <li class="nav-item" v-for="l in links" :key="l.to">
              <router-link class="nav-link nav-pill" :to="l.to" active-class="active" :exact-active-class="l.exact ? 'active' : ''">
                <i :class="'bi ' + l.icon + ' me-1'"></i>{{ l.label }}
              </router-link>
            </li>
            <li class="nav-item ms-lg-2">
              <button class="btn btn-sm btn-outline-light rounded-pill" @click="signOut" :title="auth.user?.email">
                <i class="bi bi-box-arrow-right me-1"></i>Sign out
              </button>
            </li>
          </ul>
        </div>
      </div>
    </nav>

    <main class="container my-4">
      <router-view v-slot="{ Component }">
        <transition name="page-fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>

    <footer class="modern-footer">
      <div class="container text-center">
        <div class="d-flex justify-content-center align-items-center gap-2 mb-2">
          <div class="brand-icon-sm">
            <i class="bi bi-mortarboard-fill"></i>
          </div>
          <span class="fw-semibold">E-Learning</span>
        </div>
        <p class="text-muted small mb-0">Vue 3 · Go · PostgreSQL · FSRS spaced repetition</p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { auth, logout } from './auth'

const router = useRouter()
const route = useRoute()
const navOpen = ref(false)
watch(() => route.fullPath, () => { navOpen.value = false })

const links = [
  { to: '/', label: 'Home', icon: 'bi-speedometer2', exact: true },
  { to: '/review', label: 'Review', icon: 'bi-arrow-repeat' },
  { to: '/learn', label: 'Browse', icon: 'bi-collection' },
  { to: '/quiz', label: 'Quiz', icon: 'bi-question-circle' },
  { to: '/writing', label: 'Writing', icon: 'bi-pencil' },
  { to: '/search', label: 'Search', icon: 'bi-search' },
  { to: '/devices', label: 'Device & settings', icon: 'bi-cpu' },
]

function signOut() {
  logout()
  router.push('/login')
}
</script>

<style>
@import url("https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.1/font/bootstrap-icons.css");

#app {
  font-family: 'Inter', 'Segoe UI', system-ui, -apple-system, sans-serif;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f8fafc;
}

main {
  flex: 1;
}

/* ========== MODERN NAVBAR ========== */
.modern-navbar {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  padding: 0.5rem 0;
  border-bottom: 3px solid transparent;
  border-image: linear-gradient(90deg, #e94560, #f5a623, #7b2ff7, #00d2ff) 1;
}

.modern-navbar::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: rgba(255,255,255,0.1);
}

.brand-icon {
  width: 42px;
  height: 42px;
  background: linear-gradient(135deg, #f5a623, #e94560);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  color: white;
  box-shadow: 0 4px 15px rgba(233,69,96,0.3);
}

.brand-title {
  display: block;
  font-size: 1.3rem;
  line-height: 1.2;
  letter-spacing: -0.5px;
  background: linear-gradient(90deg, #fff, #e0e7ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.brand-subtitle {
  display: block;
  font-size: 0.65rem;
  color: #f5a623;
  letter-spacing: 2px;
  text-transform: uppercase;
  font-weight: 600;
}

/* Nav Pills */
.nav-pill {
  border-radius: 50px !important;
  padding: 0.5rem 1.1rem !important;
  margin: 0 2px;
  font-size: 0.9rem;
  color: rgba(255,255,255,0.7) !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.nav-pill::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50px;
  background: rgba(255,255,255,0.08);
  opacity: 0;
  transition: opacity 0.3s;
}

.nav-pill:hover::before {
  opacity: 1;
}

.nav-pill:hover {
  color: white !important;
  transform: translateY(-1px);
}

.nav-pill.active {
  background: linear-gradient(135deg, rgba(233,69,96,0.3), rgba(123,47,247,0.3)) !important;
  color: white !important;
  box-shadow: 0 2px 12px rgba(233,69,96,0.2);
  border: 1px solid rgba(255,255,255,0.15);
}

/* ========== PAGE TRANSITIONS ========== */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: all 0.25s ease;
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* ========== MODERN FOOTER ========== */
.modern-footer {
  background: linear-gradient(135deg, #1a1a2e, #16213e);
  padding: 1.5rem 0;
  margin-top: 3rem;
  border-top: 1px solid rgba(255,255,255,0.05);
}

.brand-icon-sm {
  width: 28px;
  height: 28px;
  background: linear-gradient(135deg, #f5a623, #e94560);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.9rem;
  color: white;
}

.modern-footer span {
  color: rgba(255,255,255,0.7);
}

/* ========== GLOBAL CARD STYLE ========== */
.card {
  border-radius: 16px !important;
  transition: all 0.3s ease;
}

.card.hoverable:hover {
  transform: translateY(-3px);
  box-shadow: 0 12px 40px rgba(0,0,0,0.1) !important;
}

/* ========== RESPONSIVE ========== */
@media (max-width: 768px) {
  .brand-title {
    font-size: 1.1rem;
  }
  .nav-pill {
    padding: 0.5rem 1rem !important;
    margin: 2px 0;
  }
}
</style>
