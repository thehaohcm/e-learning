<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-end mb-3 gap-2">
      <div>
        <h2 class="fw-bold text-success mb-0"><i class="bi bi-question-circle"></i> Quiz</h2>
        <small class="text-muted">Correct answers on due cards also count as a review.</small>
      </div>
      <div class="d-flex flex-wrap gap-2 align-items-center">
        <div class="btn-group btn-group-sm" role="group" aria-label="Quiz mode">
          <button class="btn" :class="qmode === 'choose' ? 'btn-success' : 'btn-outline-success'" @click="setMode('choose')"
            title="Pick the right meaning"><i class="bi bi-ui-checks"></i> Choose</button>
          <button class="btn" :class="qmode === 'write' ? 'btn-success' : 'btn-outline-success'" @click="setMode('write')"
            title="Write the English sentence for a Vietnamese one"><i class="bi bi-pencil"></i> Write it</button>
        </div>
        <div v-if="qmode === 'choose'" class="form-check form-switch mb-0 small">
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
    <WriteIt v-if="qmode === 'write'" :key="deckId" :deck-id="deckId" />
    <div v-else-if="loading" class="text-center py-5"><div class="spinner-border text-success"></div></div>

    <div v-else-if="!questions.length" class="text-center py-5 text-muted">
      <i class="bi bi-inbox display-4"></i>
      <p class="mt-2">Not enough cards in this deck for a quiz yet (needs at least 4).</p>
    </div>

    <div v-else-if="idx < questions.length" class="row justify-content-center">
      <div class="col-lg-8">
        <div class="d-flex justify-content-between small text-muted mb-2">
          <span>
            <button type="button" class="btn btn-sm btn-outline-secondary py-0 me-2" :disabled="idx === 0" title="Previous question (←)" @click="back">
              <i class="bi bi-chevron-left"></i> Back
            </button>
            Question {{ idx + 1 }} / {{ questions.length }}
          </span>
          <span>Score: {{ score }}</span>
        </div>
        <div class="card border-0 shadow-sm">
          <div class="card-body p-4">
            <h5 class="fw-bold mb-3">
              <template v-if="listening && q.speak && !solved">
                Listen and choose the meaning
              </template>
              <template v-else>{{ q.prompt }}</template>
            </h5>
            <button v-if="q.speak" class="btn btn-sm btn-outline-secondary mb-3" @click="speakEnglish(q.speak, RATE_REPLAY).catch(() => {})">
              <i class="bi bi-volume-up"></i> Play
            </button>
            <button v-for="(c, i) in q.choices" :key="c.id" class="btn w-100 py-2 mb-2 text-start" :class="choiceClass(c.id)"
              :disabled="solved || wrong.has(c.id) || checking" @click="answer(c.id)">
              <span class="badge bg-light text-dark me-2">{{ 'ABCD'[i] }}</span>{{ c.text }}
            </button>

            <!-- wrong, still trying: the example sentence is the hint -->
            <div v-if="result && !solved" class="mt-3">
              <div class="d-flex flex-wrap align-items-center gap-2">
                <span class="text-danger fw-bold me-2"><i class="bi bi-x-circle-fill"></i> Not quite, listen to the example and try again</span>
                <template v-if="hintText">
                  <button class="btn btn-light btn-sm rounded-circle" title="Listen again" @click="speakHint(RATE_REPLAY)">
                    <i class="bi bi-volume-up-fill"></i>
                  </button>
                  <button class="btn btn-light btn-sm rounded-circle" title="Listen slowly" @click="speakHint(RATE_SLOW)">🐢</button>
                  <button class="btn btn-outline-secondary btn-sm" @click="showHint = !showHint">
                    <i :class="showHint ? 'bi bi-eye-slash' : 'bi bi-card-text'"></i> {{ showHint ? 'Hide text' : 'Show sentence' }}
                  </button>
                </template>
                <button class="btn btn-link btn-sm p-0 ms-auto" @click="reveal">Show answer</button>
              </div>
              <div v-if="showHint && hintText" class="mt-2 p-2 rounded bg-light fst-italic">“{{ hintText }}”</div>
            </div>

            <div v-if="result && solved" class="mt-3">
              <div :class="result.correct ? 'text-success' : 'text-warning'" class="fw-bold mb-1">
                <i :class="result.correct ? 'bi bi-check-circle-fill' : 'bi bi-info-circle-fill'"></i>
                {{ result.correct ? 'Correct!' : revealed ? 'Here is the answer' : `Found it on try ${wrong.size + 1} (counted as wrong)` }}
              </div>
              <div v-if="result.explanation" class="small text-muted mb-2">{{ result.explanation }}</div>
              <div v-else-if="result.card.example_en" class="small mb-2 d-flex align-items-start gap-2">
                <div>
                  <em>{{ result.card.example_en }}</em><br /><span class="text-muted">{{ result.card.example_vi }}</span>
                </div>
                <button class="btn btn-light btn-sm rounded-circle" title="Listen again" @click="speakResolved(RATE_REPLAY)">
                  <i class="bi bi-volume-up-fill"></i>
                </button>
              </div>
              <SpeakCheck v-if="!result.explanation && result.card.example_en" :text="result.card.example_en" class="mb-3" />
              <button v-if="idx > 0" class="btn btn-outline-secondary me-2" @click="back"><i class="bi bi-chevron-left"></i> Back</button>
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
          <ForgottenWords :words="forgotten" class="mx-auto mb-4" style="max-width: 640px" />
          <button class="btn btn-outline-secondary me-2" @click="back"><i class="bi bi-chevron-left"></i> Back to questions</button>
          <button class="btn btn-primary me-2" @click="start"><i class="bi bi-arrow-repeat"></i> New quiz</button>
          <router-link to="/" class="btn btn-outline-secondary">Home</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, errorMessage } from '../api'
import { RATE_REPLAY, RATE_SLOW, useSpeechSynthesis } from '../composables/useSpeechSynthesis'
import SpeakCheck from '../components/common/SpeakCheck.vue'
import ForgottenWords from '../components/common/ForgottenWords.vue'
import WriteIt from '../components/quiz/WriteIt.vue'
import type { Subject } from '../types'

interface Question { card_id: number; kind: string; mode: string; prompt: string; speak?: string; choices: { id: number; text: string }[] }
interface Result {
  correct: boolean
  correct_choice_id: number
  explanation?: string
  card: { front: string; kind: string; example_en?: string; example_vi?: string }
}

const route = useRoute()
const router = useRouter()
const { speakEnglish } = useSpeechSynthesis()

const subjects = ref<Subject[]>([])
const deckId = ref<number>(Number(route.query.deck_id) || 0)
const questions = ref<Question[]>([])
const idx = ref(0)
const score = ref(0)
const result = ref<Result | null>(null) // result of the FIRST attempt (this is what counts)
const wrong = reactive(new Set<number>()) // options already tried and wrong
const solved = ref(false) // correct option found or answer revealed
const revealed = ref(false)
const showHint = ref(false)
const checking = ref(false)
const loading = ref(true)
const listening = ref(false)
const error = ref('')
let speakToken = 0
// English words answered wrong on the first try: offered for practice at the end
const forgotten = ref<{ id: number; front: string }[]>([])
function remember(r: Result | null) {
  if (!r || (r.card.kind !== 'vocab' && r.card.kind !== 'phrasal')) return
  const id = q.value.card_id
  if (!forgotten.value.some((w) => w.id === id)) forgotten.value.push({ id, front: r.card.front })
}

const q = computed(() => questions.value[idx.value])

// don't re-read questions that were already answered (going back)
watch(q, (cur) => { if (cur?.speak && listening.value && !solved.value) speakEnglish(cur.speak).catch(() => {}) })

function choiceClass(id: number) {
  if (!result.value) return 'btn-outline-primary'
  if (solved.value && id === result.value.correct_choice_id) return 'btn-success'
  if (wrong.has(id)) return 'btn-outline-danger text-decoration-line-through'
  return solved.value ? 'btn-outline-secondary opacity-50' : 'btn-outline-primary'
}

/**
 * Read the word, then its example sentence. Used once the question is solved,
 * and after a wrong first try on "What does X mean?" questions (reading X does
 * not give the answer away there; on "Which one means" it would).
 */
async function speakResolved(rate = 1) {
  const r = result.value
  if (!r || r.card.kind === 'mcq' || r.card.kind === 'concept') return
  const token = ++speakToken
  try {
    await speakEnglish(r.card.front, rate)
    if (!r.card.example_en) return
    await new Promise((res) => setTimeout(res, 400))
    if (token === speakToken) await speakEnglish(r.card.example_en, rate)
  } catch {
    /* interrupted */
  }
}

/** Regexp matching the studied word in common forms (negotiate→negotiated, carry out→carried out). */
function wordRegex(word: string): RegExp {
  const parts = word.toLowerCase().trim().split(/\s+/).map((p, i) => {
    let x = p.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    if (i === 0 && x.length > 3 && /[ey]$/.test(x)) x = x.slice(0, -1)
    return x + '\\w*'
  })
  return new RegExp('\\b' + parts.join('(?:\\W+\\w+){0,3}?\\W+') + '\\b', 'gi')
}

/**
 * Example sentence used as a hint after a wrong answer. On "Which one means…"
 * questions the options are words, so the word itself is hidden.
 */
const hintText = computed(() => {
  const r = result.value
  const ex = r?.card.example_en
  if (!r || !ex) return ''
  return q.value?.mode === 'term' ? ex.replace(wordRegex(r.card.front), '_____') : ex
})

async function speakHint(rate: number) {
  const text = hintText.value.replace(/_____/g, 'blank')
  if (!text) return
  ++speakToken
  try {
    await speakEnglish(text, rate)
  } catch {
    /* interrupted */
  }
}

async function answer(id: number) {
  if (solved.value || wrong.has(id) || checking.value) return
  // later attempts are checked locally: only the first one is scored and sent to the server
  if (result.value) {
    if (id === result.value.correct_choice_id) {
      solved.value = true
      speakResolved()
    } else {
      wrong.add(id)
      speakHint(1)
    }
    return
  }
  checking.value = true
  try {
    const { data } = await api.post('/quiz/answer', { card_id: q.value.card_id, choice_id: id })
    result.value = data
    if (data.correct) {
      score.value++
      solved.value = true
      speakResolved()
    } else {
      wrong.add(id)
      remember(data)
      speakHint(1)
    }
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    checking.value = false
  }
}

function reveal() {
  revealed.value = true
  solved.value = true
  speakResolved()
}

function resetQuestion() {
  speakToken++
  showHint.value = false
  result.value = null
  wrong.clear()
  solved.value = false
  revealed.value = false
}

/** State of questions already visited, so "Back" shows them as they were (score is not changed). */
interface Saved { result: Result | null; wrong: number[]; solved: boolean; revealed: boolean }
let saved: Record<number, Saved> = {}

function saveState() {
  saved[idx.value] = { result: result.value, wrong: [...wrong], solved: solved.value, revealed: revealed.value }
}

function go(i: number) {
  if (i < 0 || i > questions.value.length) return
  if (idx.value < questions.value.length) saveState()
  resetQuestion()
  const st = saved[i]
  if (st) {
    result.value = st.result
    st.wrong.forEach((w) => wrong.add(w))
    solved.value = st.solved
    revealed.value = st.revealed
  }
  idx.value = i
}

function next() {
  go(idx.value + 1)
}

function back() {
  go(idx.value - 1)
}

function onKey(e: KeyboardEvent) {
  if (qmode.value === 'write') return
  const tag = (e.target as HTMLElement)?.tagName
  if (tag === 'INPUT' || tag === 'SELECT' || tag === 'TEXTAREA') return
  if (e.key === 'ArrowLeft' && idx.value > 0) back()
  else if (e.key === 'ArrowRight' && solved.value) next()
}
onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))

// "choose" = multiple choice (+ listening), "write" = Write it (Vietnamese -> English sentence)
const qmode = ref<'choose' | 'write'>(route.query.mode === 'write' ? 'write' : 'choose')

function setMode(m: 'choose' | 'write') {
  if (qmode.value === m) return
  qmode.value = m
  start()
}

async function start() {
  if (!deckId.value) return
  router.replace({ query: { deck_id: deckId.value, ...(qmode.value === 'write' ? { mode: 'write' } : {}) } })
  if (qmode.value === 'write') {
    loading.value = false // WriteIt loads its own sentences
    return
  }
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.get('/quiz', { params: { deck_id: deckId.value, count: 10 } })
    questions.value = data
    saved = {}
    forgotten.value = []
    idx.value = 0
    score.value = 0
    resetQuestion()
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
