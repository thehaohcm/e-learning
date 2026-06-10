<template>
  <div>
    <div class="text-center mb-4">
      <h1 class="fw-bold text-warning mb-2">
        <i class="bi bi-arrow-repeat"></i> Spaced Repetition Review
      </h1>
      <p class="text-muted">Review words before you forget them (FSRS Algorithm)</p>
    </div>

    <!-- Due Reviews -->
    <div v-if="items.length > 0">
      <div class="progress mb-4" style="height: 8px">
        <div class="progress-bar bg-warning" :style="{ width: progressPct + '%' }"></div>
      </div>
      <p class="text-muted text-center">{{ reviewedCount }} / {{ items.length }} reviewed</p>

      <!-- Flashcard Carousel -->
      <div class="row justify-content-center">
        <div class="col-md-8">
          <div class="card border-0 shadow-lg review-card" v-if="currentItem">
            <div class="card-body text-center p-5">
              <h2 class="display-4 fw-bold text-primary mb-3">{{ currentItem.word }}</h2>
              <p class="text-muted mb-4">{{ currentItem.phonetics }}</p>

              <div class="bg-light p-4 rounded mb-4">
                <h5 class="mb-2">{{ currentItem.example_en }}</h5>
                <p class="text-muted mb-0">{{ currentItem.example_vi }}</p>
                <small class="d-block mt-2 text-secondary">
                  <strong>Meaning:</strong> {{ currentItem.meaning_en }} / {{ currentItem.meaning_vi }}
                </small>
              </div>

              <div class="mt-4">
                <button @click="speakCurrent" class="btn btn-outline-secondary btn-lg mb-3 w-100">
                  <i class="bi bi-volume-up me-2"></i> Listen (Anh-Việt)
                </button>
              </div>

              <!-- Rating Buttons -->
              <div class="d-flex gap-2 justify-content-center">
                <button @click="rate(1)" class="btn btn-outline-danger btn-lg flex-fill" title="Forgot">
                  <i class="bi bi-x-circle fs-4"></i><br><small>Forgot</small>
                </button>
                <button @click="rate(2)" class="btn btn-outline-warning btn-lg flex-fill" title="Hard">
                  <i class="bi bi-exclamation-circle fs-4"></i><br><small>Hard</small>
                </button>
                <button @click="rate(3)" class="btn btn-outline-success btn-lg flex-fill" title="Good">
                  <i class="bi bi-check-circle fs-4"></i><br><small>Good</small>
                </button>
                <button @click="rate(4)" class="btn btn-outline-primary btn-lg flex-fill" title="Easy">
                  <i class="bi bi-star-fill fs-4"></i><br><small>Easy</small>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="loaded" class="text-center py-5">
      <i class="bi bi-check-circle display-1 text-success"></i>
      <h3 class="mt-3">All caught up!</h3>
      <p class="text-muted">No words due for review right now.</p>
      <router-link to="/learn" class="btn btn-primary mt-2">
        <i class="bi bi-plus-circle me-2"></i> Learn New Words
      </router-link>
    </div>

    <div v-else class="text-center py-5 text-muted">
      <div class="spinner-border"></div>
      <p class="mt-2">Loading due reviews...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, computed, onMounted } from 'vue'
import { useSpeechSynthesis } from '../composables/useSpeechSynthesis'

const API = (import.meta as any).env.VITE_API_BASE_URL || 'http://localhost:5173'
const USER_ID = 1
const { speakBilingual } = useSpeechSynthesis()

interface DueItem {
  vocab_id: number
  word: string
  phonetics: string
  level: number
  example_en: string
  example_vi: string
  meaning_en: string
  meaning_vi: string
}

const items = ref<DueItem[]>([])
const currentIdx = ref(0)
const reviewedCount = ref(0)
const loaded = ref(false)

const currentItem = computed(() => items.value[currentIdx.value] || null)
const progressPct = computed(() => items.value.length > 0 ? Math.round(reviewedCount.value / items.value.length * 100) : 0)

async function loadDueReviews() {
  try {
    const { data } = await axios.get(`${API}/api/srs/due-reviews`, {
      params: { limit: 20 },
      headers: { 'X-User-ID': String(USER_ID) }
    })
    items.value = data || []
    currentIdx.value = 0
    reviewedCount.value = 0
    loaded.value = true
  } catch (e) {
    console.error(e)
    loaded.value = true
  }
}

async function rate(rating: number) {
  if (!currentItem.value) return
  try {
    await axios.post(`${API}/api/srs/review`, {
      vocab_id: currentItem.value.vocab_id,
      rating
    }, { headers: { 'X-User-ID': String(USER_ID) } })
    reviewedCount.value++
    if (currentIdx.value < items.value.length - 1) {
      currentIdx.value++
    } else {
      items.value = []
    }
  } catch (e) {
    console.error(e)
  }
}

async function speakCurrent() {
  if (!currentItem.value) return
  await speakBilingual(currentItem.value.example_en, currentItem.value.example_vi)
}

onMounted(loadDueReviews)
</script>