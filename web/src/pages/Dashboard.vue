<template>
  <div>
    <!-- Header Section -->
    <div class="text-center mb-5">
      <h1 class="display-4 fw-bold text-primary mb-3">
        <i class="bi bi-speedometer2"></i> Dashboard
      </h1>
      <p class="lead text-muted">Track your language learning progress</p>
    </div>

    <!-- Statistics Cards -->
    <div class="row g-4 mb-5">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body text-center">
            <div class="mb-3">
              <i class="bi bi-check-circle-fill text-success fs-1"></i>
            </div>
            <h3 class="card-title text-success fw-bold">{{ summary.learned }}</h3>
            <p class="card-text text-muted">Words Learned</p>
          </div>
        </div>
      </div>
      
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body text-center">
            <div class="mb-3">
              <i class="bi bi-exclamation-triangle-fill text-warning fs-1"></i>
            </div>
            <h3 class="card-title text-warning fw-bold">{{ summary.weak }}</h3>
            <p class="card-text text-muted">Words Need Practice</p>
          </div>
        </div>
      </div>
      
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body text-center">
            <div class="mb-3">
              <i class="bi bi-calendar-check-fill text-info fs-1"></i>
            </div>
            <h3 class="card-title text-info fw-bold">{{ summary.due }}</h3>
            <p class="card-text text-muted">Due Today</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="row">
      <div class="col-md-6 mb-3">
        <router-link to="/learn" class="btn btn-primary btn-lg w-100 py-3 shadow-sm">
          <i class="bi bi-book me-2"></i>
          <strong>Start Learning</strong>
          <br>
          <small class="opacity-75">Practice new vocabulary</small>
        </router-link>
      </div>
      <div class="col-md-6 mb-3">
        <router-link to="/quiz" class="btn btn-success btn-lg w-100 py-3 shadow-sm">
          <i class="bi bi-question-circle me-2"></i>
          <strong>Take Quiz</strong>
          <br>
          <small class="opacity-75">Test your knowledge</small>
        </router-link>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="!summary.learned && !summary.weak && !summary.due" class="text-center">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
      <p class="mt-3 text-muted">Loading your progress...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, onMounted } from 'vue'

interface Summary {
  learned: number
  weak: number
  due: number
}

const summary = ref<Summary>({ learned: 0, weak: 0, due: 0 })
const API = import.meta.env.VITE_API_BASE_URL || 'http://localhost:5173'

onMounted(async () => {
  try {
    const { data } = await axios.get(`${API}/api/dashboard`)
    summary.value = data
  } catch (error) {
    console.error('Error loading dashboard data:', error)
    // Set default values on error
    summary.value = { learned: 0, weak: 0, due: 0 }
  }
})
</script>

<style scoped>
.card {
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.card:hover {
  transform: translateY(-5px);
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15) !important;
}

.btn-lg {
  transition: all 0.3s ease;
}

.btn-lg:hover {
  transform: translateY(-2px);
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15) !important;
}

.display-4 {
  font-weight: 700;
}
</style>
