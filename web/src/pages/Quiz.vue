<template>
  <div>
    <div class="text-center mb-2">
      <h1 class="fw-bold text-success mb-2">
        <i class="bi bi-question-circle"></i> Vocabulary Quiz
      </h1>
    </div>

    <div v-if="isLoading" class="text-center py-5">
      <div class="spinner-border text-primary"></div>
      <p class="mt-3 text-muted">Loading quiz...</p>
    </div>

    <div v-else-if="questions.length === 0" class="text-center py-5">
      <i class="bi bi-inbox display-1 text-muted"></i>
      <p class="mt-3">No quiz questions available yet.</p>
      <router-link to="/learn" class="btn btn-primary">Learn vocabulary first</router-link>
    </div>

    <div v-else>
      <!-- Progress -->
      <div class="d-flex justify-content-between align-items-center mb-3">
        <span class="text-muted">Question {{ currentIdx + 1 }} / {{ questions.length }}</span>
        <span class="text-muted">Score: {{ correctCount }} correct</span>
      </div>
      <div class="progress mb-4" style="height: 8px">
        <div class="progress-bar bg-success" :style="{ width: ((currentIdx) / questions.length * 100) + '%' }"></div>
      </div>

      <!-- Current Question -->
      <div v-if="currentIdx < questions.length" class="row justify-content-center">
        <div class="col-md-8">
          <div class="card border-0 shadow-sm">
            <div class="card-body p-4">
              <h5 class="fw-bold mb-4">{{ currentQ?.prompt }}</h5>
              <div class="row g-2">
                <div v-for="(choice, idx) in currentQ?.choices || []" :key="idx" class="col-12">
                  <button
                    @click="answer(idx)"
                    class="btn w-100 py-3 text-start"
                    :class="answerClass(idx)"
                    :disabled="selectedAnswer !== null"
                  >
                    <span class="badge bg-light text-dark me-2">{{ ['A', 'B', 'C', 'D'][idx] }}</span>
                    {{ choice }}
                  </button>
                </div>
              </div>
              <div v-if="selectedAnswer !== null" class="mt-3 text-center">
                <span v-if="lastAnswer?.correct" class="text-success fw-bold">
                  <i class="bi bi-check-circle-fill"></i> Correct!
                </span>
                <span v-else class="text-danger fw-bold">
                  <i class="bi bi-x-circle-fill"></i> Incorrect. Correct: {{ ['A', 'B', 'C', 'D'][lastAnswer?.correct_answer_idx ?? 0] }}
                </span>
                <br>
                <button @click="nextQuestion" class="btn btn-primary mt-2">
                  {{ currentIdx < questions.length - 1 ? 'Next Question' : 'See Results' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Results -->
      <div v-else class="text-center">
        <div class="card border-0 shadow-lg">
          <div class="card-body py-5">
            <i class="bi bi-trophy-fill display-1 text-warning"></i>
            <h3 class="mt-3 fw-bold">Quiz Complete!</h3>
            <p class="lead">You scored {{ correctCount }} / {{ questions.length }} ({{ Math.round(correctCount / questions.length * 100) }}%)</p>
            <button @click="retake" class="btn btn-primary btn-lg me-2">
              <i class="bi bi-arrow-repeat me-2"></i>Retake
            </button>
            <router-link to="/dashboard" class="btn btn-outline-secondary btn-lg">
              Back to Dashboard
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, computed, onMounted } from 'vue'

const API = (import.meta as any).env.VITE_API_BASE_URL || 'http://localhost:5173'
const USER_ID = 1

interface Question {
  id: number
  prompt: string
  choices: string[]
}

const questions = ref<Question[]>([])
const currentIdx = ref(0)
const selectedAnswer = ref<number | null>(null)
const correctCount = ref(0)
const isLoading = ref(true)
const lastAnswer = ref<any>(null)

const currentQ = computed(() => questions.value[currentIdx.value] || null)

function answerClass(idx: number) {
  if (selectedAnswer.value === null) return 'btn-outline-primary'
  if (lastAnswer.value?.correct && idx === selectedAnswer.value) return 'btn-success text-white'
  if (!lastAnswer.value?.correct && idx === selectedAnswer.value) return 'btn-danger text-white'
  if (idx === lastAnswer.value?.correct_answer_idx) return 'btn-success text-white'
  return 'btn-outline-secondary'
}

async function answer(idx: number) {
  if (selectedAnswer.value !== null || !currentQ.value) return
  selectedAnswer.value = idx
  try {
    const { data } = await axios.post(`${API}/api/quiz/answer`, {
      question_id: currentQ.value.id,
      answer_idx: idx
    }, { headers: { 'X-User-ID': String(USER_ID) } })
    lastAnswer.value = data
    if (data.correct) correctCount.value++
  } catch (e) { console.error(e) }
}

function nextQuestion() {
  selectedAnswer.value = null
  lastAnswer.value = null
  currentIdx.value++
}

function retake() {
  currentIdx.value = 0
  selectedAnswer.value = null
  correctCount.value = 0
  lastAnswer.value = null
  loadQuiz()
}

async function loadQuiz() {
  isLoading.value = true
  try {
    const { data } = await axios.get(`${API}/api/quiz`)
    questions.value = data || []
  } catch (e) { console.error(e) }
  isLoading.value = false
}

onMounted(loadQuiz)
</script>