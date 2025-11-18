<template>
  <div>
    <!-- Header Section -->
    <div class="text-center mb-5">
      <h1 class="display-5 fw-bold text-primary mb-3">
        <i class="bi bi-book"></i> Learn Vocabulary
      </h1>
      <div class="badge bg-secondary fs-6 px-3 py-2">
        <i class="bi bi-tag me-1"></i> Topic: {{ topic }}
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="vocabs.length === 0" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
      <p class="mt-3 text-muted">Loading vocabulary...</p>
    </div>

    <!-- Vocabulary Cards -->
    <div v-else class="row g-4">
      <div v-for="v in vocabs" :key="v.ID" class="col-md-6 col-lg-4">
        <div class="card border-0 shadow-sm h-100 vocabulary-card">
          <div class="card-body">
            <!-- Word Header -->
            <div class="d-flex justify-content-between align-items-start mb-3">
              <h3 class="card-title text-primary fw-bold mb-0">
                {{ v.Word }}
              </h3>
              <button
                @click="select(v.ID)"
                class="btn btn-outline-success btn-sm"
                :disabled="selectedWords.includes(v.ID)"
              >
                <i class="bi" :class="selectedWords.includes(v.ID) ? 'bi-check-circle-fill' : 'bi-plus-circle'"></i>
                {{ selectedWords.includes(v.ID) ? 'Selected' : 'Select' }}
              </button>
            </div>

            <!-- Meanings -->
            <div class="mb-3">
              <div class="mb-2">
                <span class="badge bg-primary me-2">EN</span>
                <span class="text-muted">{{ v.MeaningEn }}</span>
              </div>
              <div>
                <span class="badge bg-danger me-2">VI</span>
                <span class="text-muted">{{ v.MeaningVi }}</span>
              </div>
            </div>

            <!-- Examples -->
            <div class="bg-light p-3 rounded-3">
              <div class="mb-2">
                <small class="text-muted">
                  <i class="bi bi-quote"></i> {{ v.ExampleEn }}
                </small>
              </div>
              <div>
                <small class="text-muted">
                  <i class="bi bi-quote"></i> {{ v.ExampleVi }}
                </small>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Action Button -->
    <div class="text-center mt-5">
      <button @click="goQuiz" class="btn btn-primary btn-lg px-5 py-3 shadow-sm">
        <i class="bi bi-arrow-right-circle me-2"></i>
        <strong>Go to Quiz</strong>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

interface Vocab {
  ID: number
  Word: string
  MeaningEn: string
  MeaningVi: string
  ExampleEn: string
  ExampleVi: string
}

const router = useRouter()
const topic = ref('Environment')
const vocabs = ref<Vocab[]>([])
const selectedWords = ref<number[]>([])
const API = import.meta.env.VITE_API_BASE_URL || 'http://localhost:5173'

onMounted(async () => {
  try {
    const { data } = await axios.get(`${API}/api/vocabs?topic=${encodeURIComponent(topic.value)}`)
    vocabs.value = data
  } catch (error) {
    console.error('Error loading vocabulary:', error)
    vocabs.value = []
  }
})

async function select(id: number) {
  try {
    await axios.post(`${API}/api/vocabs/select`, {
      user_id: 1,
      vocab_ids: [id]
    })
    selectedWords.value.push(id)
    
    // Show success message
    const toast = document.createElement('div')
    toast.className = 'position-fixed top-0 end-0 p-3'
    toast.style.zIndex = '9999'
    toast.innerHTML = `
      <div class="toast show" role="alert">
        <div class="toast-header bg-success text-white">
          <i class="bi bi-check-circle me-2"></i>
          <strong class="me-auto">Success!</strong>
          <button type="button" class="btn-close btn-close-white" data-bs-dismiss="toast"></button>
        </div>
        <div class="toast-body">
          Word selected for study!
        </div>
      </div>
    `
    document.body.appendChild(toast)
    setTimeout(() => toast.remove(), 3000)
  } catch (error) {
    console.error('Error selecting word:', error)
    alert('Failed to select word. Please try again.')
  }
}

function goQuiz() {
  router.push('/quiz')
}
</script>

<style scoped>
.vocabulary-card {
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.vocabulary-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15) !important;
}

.card-title {
  font-size: 1.5rem;
}

.btn-lg {
  transition: all 0.3s ease;
}

.btn-lg:hover {
  transform: translateY(-2px);
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15) !important;
}
</style>
