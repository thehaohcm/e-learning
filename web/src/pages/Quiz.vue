<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-end mb-3 gap-2">
      <div>
        <h2 class="fw-bold text-success mb-0"><i class="bi bi-question-circle"></i> Quiz</h2>
        <small class="text-muted">Correct answers on due cards also count as a review.</small>
      </div>
      <div class="d-flex gap-2 align-items-center">
        <div class="form-check form-switch mb-0 small">
          <input id="listen" v-model="listening" class="form-check-input" type="checkbox" />
          <label class="form-check-label" for="listen">Listening mode</label>
        </div>
        <select v-model.number="deckId" class="form-select w-auto" @change="start">
          <optgroup v-for="s in subjects" :key="s.id" :label="s.name">
            <option v-for="d in s.decks" :key="d.id" :value="d.id">{{ d.name }}</option>
          </optgroup>
        </select>
      </div>
    </div>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-if="loading" class="text-center py-5"><div class="spinner-border text-success"></div></div>

    <div v-else-if="!questions.length" class="text-center py-5 text-muted">
      <i class="bi bi-inbox display-4"></i>
      <p class="mt-2">Not enough cards in this deck for a quiz yet (needs at least 4).</p>
    </div>

    <div v-else-if="idx < questions.length" class="row justify-content-center">
      <div class="col-lg-8">
        <div class="d-flex justify-content-between small text-muted mb-2">
          <span>Question {{ idx + 1 }} / {{ questions.length }}</span>
          <span>Score: {{ score }}</span>
        </div>
        <div class="card border-0 shadow-sm">
          <div class="card-body p-4">
            <h5 class="fw-bold mb-3">
              <template v-if="listening && q.speak && !answered">
                Listen and choose the meaning
              </template>
              <template v-else>{{ q.prompt }}</template>
            </h5>
            <button v-if="q.speak" class="btn btn-sm btn-outline-secondary mb-3" @click="speakEnglish(q.speak)">
              <i class="bi bi-volume-up"></i> Play
            </button>
            <button v-for="(c, i) in q.choices" :key="c.id" class="btn w-100 py-2 mb-2 text-start" :class="choiceClass(c.id)"
              :disabled="answered" @click="answer(c.id)">
              <span class="badge bg-light text-dark me-2">{{ 'ABCD'[i] }}</span>{{ c.text }}
            </button>

            <div v-if="result" class="mt-3">
              <div :class="result.correct ? 'text-success' : 'text-danger'" class="fw-bold mb-1">
                <i :class="result.correct ? 'bi bi-check-circle-fill' : 'bi bi-x-circle-fill'"></i>
                {{ result.correct ? 'Correct!' : 'Not quite.' }}
              </div>
              <div v-if="result.explanation" class="small text-muted mb-2">{{ result.explanation }}</div>
              <div v-else-if="result.card.example_en" class="small text-muted mb-2">
                <em>{{ result.card.example_en }}</em><br />{{ result.card.example_vi }}
              </div>
              <button class="btn btn-primary" @click="next">{{ idx < questions.length - 1 ? 'Next' : 'See result' }}</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="text-center">
      <div class="card border-0 shadow-sm">
        <div class="card-body py-5">
          <i class="bi bi-trophy-fill display-3 text-warning"></i>
          <h3 class="mt-3 fw-bold">{{ score }} / {{ questions.length }}</h3>
          <p class="text-muted">{{ Math.round((score / questions.length) * 100) }}% correct</p>
          <button class="btn btn-primary me-2" @click="start"><i class="bi bi-arrow-repeat"></i> New quiz</button>
          <router-link to="/" class="btn btn-outline-secondary">Home</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, errorMessage } from '../api'
import { useSpeechSynthesis } from '../composables/useSpeechSynthesis'
import type { Subject } from '../types'

interface Question { card_id: number; kind: string; mode: string; prompt: string; speak?: string; choices: { id: number; text: string }[] }
interface Result { correct: boolean; correct_choice_id: number; explanation?: string; card: { example_en?: string; example_vi?: string } }

const route = useRoute()
const router = useRouter()
const { speakEnglish } = useSpeechSynthesis()

const subjects = ref<Subject[]>([])
const deckId = ref<number>(Number(route.query.deck_id) || 0)
const questions = ref<Question[]>([])
const idx = ref(0)
const score = ref(0)
const chosen = ref<number | null>(null)
const result = ref<Result | null>(null)
const loading = ref(true)
const listening = ref(false)
const error = ref('')

const q = computed(() => questions.value[idx.value])
const answered = computed(() => chosen.value !== null)

watch(q, (cur) => { if (cur?.speak && listening.value) speakEnglish(cur.speak).catch(() => {}) })

function choiceClass(id: number) {
  if (!result.value) return 'btn-outline-primary'
  if (id === result.value.correct_choice_id) return 'btn-success'
  if (id === chosen.value) return 'btn-danger'
  return 'btn-outline-secondary'
}

async function answer(id: number) {
  if (answered.value) return
  chosen.value = id
  try {
    const { data } = await api.post('/quiz/answer', { card_id: q.value.card_id, choice_id: id })
    result.value = data
    if (data.correct) score.value++
  } catch (e) {
    error.value = errorMessage(e)
  }
}

function next() {
  chosen.value = null
  result.value = null
  idx.value++
}

async function start() {
  if (!deckId.value) return
  router.replace({ query: { deck_id: deckId.value } })
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.get('/quiz', { params: { deck_id: deckId.value, count: 10 } })
    questions.value = data
    idx.value = 0
    score.value = 0
    chosen.value = null
    result.value = null
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    const { data } = await api.get('/subjects')
    subjects.value = data
    if (!deckId.value) deckId.value = data[0]?.decks[0]?.id ?? 0
    await start()
  } catch (e) {
    error.value = errorMessage(e)
    loading.value = false
  }
})
</script>
