<template>
  <div class="row justify-content-center">
    <div class="col-lg-7">
      <h2 class="fw-bold mb-1"><i class="bi bi-ui-checks"></i> Check the words</h2>
      <p class="text-muted small mb-3">
        One question per word: which sentence uses it correctly, its closest meaning, or the word that goes with it.
        A wrong answer brings the word back in Study soon.
      </p>

      <div v-if="loading" class="text-center py-5">
        <div class="spinner-border text-primary"></div>
        <div class="mt-3 text-muted">Preparing questions… the first time AI writes them (up to a minute), then they are kept.</div>
      </div>
      <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
      <div v-else-if="!questions.length" class="alert alert-light border">
        No questions for these words{{ warning ? ': ' + warning : '' }}.
      </div>

      <!-- finished -->
      <div v-else-if="idx >= questions.length" class="card border-0 shadow-sm">
        <div class="card-body text-center py-4">
          <div class="display-6 fw-bold" :class="right * 10 >= questions.length * 8 ? 'text-success' : 'text-primary'">
            {{ right }} / {{ questions.length }}
          </div>
          <p class="text-muted mb-3">{{ missed.length ? 'These come back in Study soon:' : 'All right. Well done!' }}</p>
          <div v-if="missed.length" class="mb-3">
            <span v-for="w in missed" :key="w" class="badge bg-warning-subtle text-dark border me-1">{{ w }}</span>
          </div>
          <router-link class="btn btn-primary" to="/review"><i class="bi bi-arrow-repeat"></i> Back to Study</router-link>
        </div>
      </div>

      <!-- question -->
      <div v-else class="card border-0 shadow-sm">
        <div class="card-body">
          <div class="d-flex justify-content-between small text-muted mb-2">
            <span>{{ KIND_LABEL[q.kind] ?? 'Question' }} · <strong>{{ q.word }}</strong></span>
            <span>{{ idx + 1 }} / {{ questions.length }}</span>
          </div>
          <p class="fs-5 mb-3">{{ q.prompt }}</p>
          <button v-for="(c, i) in q.choices" :key="i" type="button" class="btn w-100 text-start mb-2 py-2"
            :class="choiceClass(i)" :disabled="picked !== null" @click="pick(i)">
            <span class="badge bg-light text-dark me-2">{{ 'ABCD'[i] }}</span>{{ c }}
          </button>
          <button v-if="picked === null" type="button" class="btn btn-outline-secondary w-100 text-start mb-2 py-2 idk" @click="pick(-1)">
            <span class="badge bg-light text-dark me-2">0</span>? I don't know
          </button>
          <template v-if="picked !== null">
            <div class="alert py-2 mt-2" :class="picked === q.answer ? 'alert-success' : 'alert-warning'">
              <i class="bi" :class="picked === q.answer ? 'bi-check-circle-fill' : 'bi-arrow-counterclockwise'"></i>
              {{ picked === q.answer ? 'Correct!' : 'Not this time: the word comes back in Study soon.' }}
              <div v-if="q.explanation" class="small mt-1">{{ q.explanation }}</div>
            </div>
            <button class="btn btn-primary w-100" @click="next">
              {{ idx + 1 < questions.length ? 'Next' : 'Finish' }} <small class="opacity-75">(Enter)</small>
            </button>
          </template>
          <div v-else class="small text-muted text-center mt-1">Keys 1–4 or A–D, 0 = I don't know</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, errorMessage } from '../api'

interface WordQuestion { card_id: number; word: string; kind: string; prompt: string; choices: string[]; answer: number; explanation: string }

const KIND_LABEL: Record<string, string> = { usage: 'Usage', synonym: 'Closest meaning', collocation: 'Goes with' }

const route = useRoute()
const questions = ref<WordQuestion[]>([])
const idx = ref(0)
const picked = ref<number | null>(null)
const right = ref(0)
const missed = ref<string[]>([])
const loading = ref(true)
const error = ref('')
const warning = ref('')
const q = computed(() => questions.value[idx.value])

function choiceClass(i: number) {
  if (picked.value === null) return 'btn-outline-primary'
  if (i === q.value.answer) return 'btn-success'
  if (i === picked.value) return 'btn-danger'
  return 'btn-outline-secondary opacity-50'
}

function pick(i: number) {
  if (picked.value !== null) return
  picked.value = i
  const ok = i === q.value.answer
  if (ok) right.value++
  else missed.value.push(q.value.word)
  api.post('/practice/word-questions/answer', { card_id: q.value.card_id, correct: ok }).catch(() => {})
}

function next() {
  picked.value = null
  idx.value++
}

function onKey(e: KeyboardEvent) {
  if (!q.value || loading.value) return
  if (picked.value === null) {
    const n = '1234'.indexOf(e.key) >= 0 ? Number(e.key) - 1 : 'abcd'.indexOf(e.key.toLowerCase())
    if (n >= 0 && n < 4) pick(n)
    else if (e.key === '0') pick(-1)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    next()
  }
}

onMounted(async () => {
  window.addEventListener('keydown', onKey)
  const ids = String(route.query.ids ?? '').split(',').map(Number).filter((n) => n > 0)
  try {
    const { data } = await api.post('/practice/word-questions', { card_ids: ids }, { timeout: 90000 })
    questions.value = data.questions ?? []
    warning.value = data.warning ?? ''
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
})
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<style scoped>
.idk { border-style: dashed; }
</style>
