<template>
  <div>
    <div class="text-center mb-5">
      <h1 class="display-4 fw-bold text-primary mb-3">
        <i class="bi bi-speedometer2"></i> Dashboard
      </h1>
      <p class="lead text-muted">Your AI-Powered Learning Journey</p>
    </div>

    <div class="row g-4 mb-5">
      <div class="col-md-3" v-for="stat in stats" :key="stat.label">
        <div class="card border-0 shadow-sm h-100 text-center" :class="'border-' + stat.color">
          <div class="card-body">
            <i :class="'bi ' + stat.icon + ' fs-1 text-' + stat.color"></i>
            <h3 class="mt-2 fw-bold">{{ stat.value }}</h3>
            <p class="text-muted mb-0">{{ stat.label }}</p>
          </div>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-md-4 mb-3">
        <router-link to="/learn" class="btn btn-primary btn-lg w-100 py-3 shadow-sm">
          <i class="bi bi-book me-2"></i><strong>Learn Vocabulary</strong>
          <br><small>Explore 3-level sentences</small>
        </router-link>
      </div>
      <div class="col-md-4 mb-3">
        <router-link to="/review" class="btn btn-warning btn-lg w-100 py-3 shadow-sm">
          <i class="bi bi-arrow-repeat me-2"></i><strong>Review Due Words</strong>
          <br><small>Spaced Repetition (SRS)</small>
        </router-link>
      </div>
      <div class="col-md-4 mb-3">
        <router-link to="/quiz" class="btn btn-success btn-lg w-100 py-3 shadow-sm">
          <i class="bi bi-question-circle me-2"></i><strong>Take Quiz</strong>
          <br><small>Test your knowledge</small>
        </router-link>
      </div>
    </div>
    <div class="row mt-3">
      <div class="col-md-4 mb-3">
        <router-link to="/writing" class="btn btn-info btn-lg w-100 py-3 shadow-sm text-white">
          <i class="bi bi-pencil me-2"></i><strong>Writing Lab</strong>
          <br><small>AI Feedback & 3-Level Rewrite</small>
        </router-link>
      </div>
      <div class="col-md-4 mb-3">
        <router-link to="/search" class="btn btn-secondary btn-lg w-100 py-3 shadow-sm">
          <i class="bi bi-search me-2"></i><strong>AI Search</strong>
          <br><small>Find vocabulary & examples</small>
        </router-link>
      </div>
      <div class="col-md-4 mb-3">
        <router-link to="/preferences" class="btn btn-outline-dark btn-lg w-100 py-3 shadow-sm">
          <i class="bi bi-gear me-2"></i><strong>Settings</strong>
          <br><small>Notification preferences</small>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, onMounted } from 'vue'

const API = (import.meta as any).env.VITE_API_BASE_URL || 'http://localhost:5173'
const USER_ID = 1

const stats = ref([
  { label: 'Active Words', value: 0, icon: 'bi-bookmark-check', color: 'primary' },
  { label: 'Mastered', value: 0, icon: 'bi-trophy', color: 'success' },
  { label: 'Due Reviews', value: 0, icon: 'bi-clock', color: 'warning' },
  { label: '7-Day Streak', value: 0, icon: 'bi-fire', color: 'danger' },
])

onMounted(async () => {
  try {
    const { data } = await axios.get(`${API}/api/dashboard`, { headers: { 'X-User-ID': String(USER_ID) } })
    stats.value[0].value = data.active_vocabs || 0
    stats.value[1].value = data.mastered_vocabs || 0
    stats.value[2].value = data.due_reviews || 0
    stats.value[3].value = data.streak_7d || 0
  } catch (e) {
    console.error('Dashboard load error:', e)
  }
})
</script>