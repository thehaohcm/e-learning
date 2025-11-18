<template>
  <div>
    <!-- Header Section -->
    <div class="text-center mb-5">
      <h1 class="display-5 fw-bold text-primary mb-3">
        <i class="bi bi-question-circle"></i> Quiz Time!
      </h1>
      <p class="lead text-muted">Test your vocabulary knowledge</p>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
      <p class="mt-3 text-muted">Preparing your quiz...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="questions.length === 0" class="text-center py-5">
      <div class="alert alert-warning" role="alert">
        <i class="bi bi-exclamation-triangle me-2"></i>
        No quiz questions available. Please try again later.
      </div>
    </div>

    <!-- Quiz Questions -->
    <div v-else>
      <!-- Progress Bar -->
      <div class="mb-4">
        <div class="d-flex justify-content-between align-items-center mb-2">
          <span class="text-muted">Progress</span>
          <span class="text-muted">{{ answeredQuestions }}/{{ questions.length }} questions</span>
        </div>
        <div class="progress" style="height: 8px;">
          <div
            class="progress-bar bg-primary"
            role="progressbar"
            :style="{ width: progressPercentage + '%' }"
            :aria-valuenow="answeredQuestions"
            aria-valuemin="0"
            :aria-valuemax="questions.length"
          ></div>
        </div>
      </div>

      <!-- Questions List -->
      <div class="row g-4">
        <div v-for="q in questions" :key="q.ID" class="col-12">
          <div class="card border-0 shadow-sm quiz-card">
            <div class="card-body">
              <!-- Question Header -->
              <div class="d-flex justify-content-between align-items-start mb-4">
                <h5 class="card-title text-dark fw-bold">
                  <i class="bi bi-question-circle-fill text-primary me-2"></i>
                  {{ q.Prompt }}
                </h5>
                <span
                  v-if="questionResults[q.ID] !== undefined"
                  class="badge"
                  :class="questionResults[q.ID] ? 'bg-success' : 'bg-danger'"
                >
                  <i class="bi" :class="questionResults[q.ID] ? 'bi-check-circle' : 'bi-x-circle'"></i>
                  {{ questionResults[q.ID] ? 'Correct' : 'Incorrect' }}
                </span>
              </div>

              <!-- Answer Choices -->
              <div class="row g-3">
                <div
                  v-for="(choice, idx) in JSON.parse(q.Choices)"
                  :key="idx"
                  class="col-md-6"
                >
                  <button
                    @click="answer(q.ID, idx)"
                    class="btn btn-outline-primary w-100 py-3 choice-btn"
                    :disabled="questionResults[q.ID] !== undefined"
                    :class="{
                      'btn-success': questionResults[q.ID] === true && selectedAnswers[q.ID] === idx,
                      'btn-danger': questionResults[q.ID] === false && selectedAnswers[q.ID] === idx,
                      'btn-primary': questionResults[q.ID] === true && correctAnswers[q.ID] === idx
                    }"
                  >
                    <span class="fw-semibold">{{ choice }}</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Quiz Results -->
      <div v-if="answeredQuestions === questions.length" class="text-center mt-5">
        <div class="card border-0 shadow-lg">
          <div class="card-body py-5">
            <div class="mb-4">
              <i
                class="bi fs-1"
                :class="scorePercentage >= 80 ? 'bi-trophy-fill text-warning' : scorePercentage >= 60 ? 'bi-award-fill text-info' : 'bi-stars text-secondary'"
              ></i>
            </div>
            <h3 class="fw-bold mb-3">
              Quiz Complete!
            </h3>
            <p class="lead text-muted mb-4">
              You scored <span class="fw-bold text-primary">{{ correctAnswersCount }}</span> out of
              <span class="fw-bold">{{ questions.length }}</span> ({{ scorePercentage }}%)
            </p>
            <button @click="retakeQuiz" class="btn btn-primary btn-lg px-4">
              <i class="bi bi-arrow-repeat me-2"></i>Retake Quiz
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, onMounted, computed } from 'vue'

interface Question {
  ID: number
  Prompt: string
  Choices: string
  AnswerIdx: number
}

const questions = ref<Question[]>([])
const questionResults = ref<Record<number, boolean>>({})
const selectedAnswers = ref<Record<number, number>>({})
const correctAnswers = ref<Record<number, number>>({})
const isLoading = ref(true)
const API = import.meta.env.VITE_API_BASE_URL || 'http://localhost:5173'

onMounted(async () => {
  try {
    console.log('Fetching quiz from:', `${API}/api/quiz`)
    const { data } = await axios.get(`${API}/api/quiz`)
    console.log('Quiz data received:', data)
    questions.value = data
    console.log('Questions set to:', questions.value)
  } catch (error) {
    console.error('Error loading quiz:', error)
    questions.value = []
  } finally {
    isLoading.value = false
  }
})

async function answer(questionId: number, answerIdx: number) {
  try {
    selectedAnswers.value[questionId] = answerIdx
    
    const { data } = await axios.post(`${API}/api/quiz/attempt`, {
      user_id: 1,
      question_id: questionId,
      answer_idx: answerIdx
    })
    
    questionResults.value[questionId] = data.correct
    correctAnswers.value[questionId] = data.correct_answer_idx
    
    // Show feedback
    if (data.correct) {
      showToast('Correct! Well done!', 'success')
    } else {
      showToast('Incorrect. Keep trying!', 'danger')
    }
  } catch (error) {
    console.error('Error submitting answer:', error)
    alert('Failed to submit answer. Please try again.')
  }
}

function showToast(message: string, type: string) {
  const toast = document.createElement('div')
  toast.className = 'position-fixed top-0 end-0 p-3'
  toast.style.zIndex = '9999'
  toast.innerHTML = `
    <div class="toast show" role="alert">
      <div class="toast-header bg-${type} text-white">
        <i class="bi bi-info-circle me-2"></i>
        <strong class="me-auto">Notification</strong>
        <button type="button" class="btn-close btn-close-white" data-bs-dismiss="toast"></button>
      </div>
      <div class="toast-body">
        ${message}
      </div>
    </div>
  `
  document.body.appendChild(toast)
  setTimeout(() => toast.remove(), 3000)
}

function retakeQuiz() {
  // Reset all quiz data
  questionResults.value = {}
  selectedAnswers.value = {}
  correctAnswers.value = {}
}

const answeredQuestions = computed(() => {
  return Object.keys(questionResults.value).length
})

const progressPercentage = computed(() => {
  return questions.value.length > 0
    ? Math.round((answeredQuestions.value / questions.value.length) * 100)
    : 0
})

const correctAnswersCount = computed(() => {
  return Object.values(questionResults.value).filter(result => result === true).length
})

const scorePercentage = computed(() => {
  return questions.value.length > 0
    ? Math.round((correctAnswersCount.value / questions.value.length) * 100)
    : 0
})
</script>

<style scoped>
.quiz-card {
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.quiz-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15) !important;
}

.choice-btn {
  transition: all 0.3s ease;
  font-size: 1rem;
}

.choice-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 0.25rem 0.5rem rgba(0, 0, 0, 0.1);
}

.choice-btn:disabled {
  cursor: not-allowed;
  opacity: 0.8;
}

.progress {
  border-radius: 10px;
}

.progress-bar {
  transition: width 0.6s ease;
}

.display-5 {
  font-weight: 700;
}
</style>
