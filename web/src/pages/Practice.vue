<template>
  <div class="row justify-content-center">
    <div class="col-lg-8">
      <div class="d-flex flex-wrap justify-content-between align-items-end gap-2 mb-3">
        <div>
          <h2 class="fw-bold mb-0"><i class="bi bi-translate"></i> Practise forgotten words</h2>
          <small class="text-muted">Translate each sentence into Vietnamese, then read it aloud.</small>
        </div>
        <div v-if="speechOK" class="form-check form-switch mb-0 small" title="Skip the reading-aloud step (e.g. in a quiet place)">
          <input id="nomic" v-model="noMic" class="form-check-input" type="checkbox" />
          <label class="form-check-label" for="nomic">Can't speak right now</label>
        </div>
      </div>

      <div v-if="warning" class="alert alert-warning py-2 small">{{ warning }}</div>

      <!-- loading: one AI request writes the sentences for all words -->
      <div v-if="phase === 'loading'" class="text-center py-5">
        <div class="spinner-border text-warning"></div>
        <p class="mt-3 text-muted">AI is writing example sentences for {{ idCount }} word{{ idCount > 1 ? 's' : '' }}… (up to ~40 s)</p>
      </div>

      <div v-else-if="phase === 'empty'" class="text-center py-5">
        <i class="bi bi-emoji-smile display-4 text-success"></i>
        <p class="mt-3 mb-1 fw-semibold">No forgotten words to practise right now.</p>
        <p class="text-muted small">Words you mark “Again” in Study or miss in a Quiz show up here.</p>
        <router-link to="/review" class="btn btn-primary me-2">Study</router-link>
        <router-link to="/quiz" class="btn btn-outline-primary">Quiz</router-link>
      </div>

      <div v-else-if="phase === 'error'" class="alert alert-danger">
        {{ error }}
        <button class="btn btn-sm btn-outline-danger ms-2" @click="load">Try again</button>
      </div>

      <!-- practising -->
      <template v-else-if="phase === 'practice' && word && ex">
        <div class="d-flex justify-content-between small text-muted mb-1">
          <span>Word {{ wi + 1 }} / {{ words.length }} · sentence {{ ei + 1 }} / {{ word.examples.length }}</span>
          <span>{{ ei === 0 ? 'required' : 'optional' }}</span>
        </div>
        <div class="progress mb-3" style="height: 6px">
          <div class="progress-bar bg-warning" :style="{ width: progressPct + '%' }"></div>
        </div>

        <div class="card border-0 shadow-sm">
          <div class="card-body p-4">
            <div class="d-flex align-items-center gap-2 mb-1">
              <h3 class="fw-bold text-primary mb-0">{{ word.word }}</h3>
              <button class="btn btn-light btn-sm rounded-circle" title="Listen" @click="say(word.word, RATE_REPLAY)">
                <i class="bi bi-volume-up-fill"></i>
              </button>
              <span v-if="word.ipa" class="text-muted small">{{ word.ipa }}</span>
            </div>
            <div class="text-muted mb-3">{{ word.meaning_vi }}</div>

            <!-- 1. the sentence -->
            <div class="d-flex align-items-start gap-2 mb-3">
              <p class="fs-5 mb-0 flex-grow-1 sentence">
                <template v-for="(seg, k) in segments(ex.en, word.word)" :key="k">
                  <strong v-if="seg.hit" class="text-primary">{{ seg.text }}</strong><template v-else>{{ seg.text }}</template>
                </template>
              </p>
              <button class="btn btn-light btn-sm rounded-circle flex-shrink-0" title="Listen again" @click="say(ex.en, RATE_REPLAY)">
                <i class="bi bi-volume-up-fill"></i>
              </button>
              <button class="btn btn-light btn-sm rounded-circle flex-shrink-0" title="Listen slowly" @click="say(ex.en, RATE_SLOW)">🐢</button>
            </div>

            <!-- 2. translate -->
            <label class="form-label small fw-semibold" :for="`tr-${ex.id}`">1. Translate into Vietnamese</label>
            <textarea :id="`tr-${ex.id}`" v-model="ex.translation" class="form-control mb-2" rows="2"
              :readonly="ex.checked" placeholder="Bản dịch của bạn…" @keydown.ctrl.enter.prevent="check"
              @keydown.meta.enter.prevent="check"></textarea>
            <div v-if="!ex.checked" class="d-flex align-items-center gap-2">
              <button class="btn btn-primary btn-sm" :disabled="ex.translation.trim().length < 2" @click="check">
                Check <small class="opacity-75 d-none d-md-inline">(Ctrl+Enter)</small>
              </button>
              <button class="btn btn-link btn-sm text-muted" @click="giveUp">I can't translate it</button>
            </div>
            <div v-else class="p-2 rounded bg-light small mb-3">
              <span class="text-muted">Sample translation:</span> {{ ex.vi }}
              <div class="text-muted mt-1">Compare with yours. AI checks all your translations at the end.</div>
            </div>

            <!-- 3. read it aloud -->
            <template v-if="ex.checked">
              <div class="small fw-semibold mb-2">
                2. Read it aloud
                <span v-if="!speakRequired" class="fw-normal text-muted">(skipped)</span>
                <span v-else-if="ex.best >= PASS" class="text-success fw-normal"><i class="bi bi-check-circle-fill"></i> {{ ex.best }}%</span>
                <span v-else class="fw-normal text-muted">
                  reach {{ PASS }}%<template v-if="ex.tries"> · best {{ ex.best }}% · try {{ ex.tries }}/{{ MAX_TRIES }}</template>
                </span>
              </div>
              <SpeakCheck v-if="speakRequired" :text="ex.en" class="mb-3" @scored="onScored" />
            </template>

            <!-- next -->
            <div v-if="exDone" class="d-flex flex-wrap gap-2 mt-3 pt-3 border-top">
              <button v-if="ei < word.examples.length - 1" class="btn btn-outline-primary" @click="nextSentence">
                Another sentence <small class="opacity-75">(optional)</small>
              </button>
              <button class="btn btn-success ms-auto" @click="nextWord">
                <template v-if="wi < words.length - 1">Next word <i class="bi bi-arrow-right"></i></template>
                <template v-else><i class="bi bi-check2-all"></i> Finish &amp; check my translations</template>
              </button>
            </div>
            <div v-else-if="ex.checked && speakRequired && ex.tries > 0 && ex.best < PASS" class="small text-muted mt-2">
              Not quite: press <b>Try again</b> to hear it slowly and repeat.
            </div>
          </div>
        </div>
      </template>

      <!-- grading: one AI request for all translations -->
      <div v-else-if="phase === 'grading'" class="text-center py-5">
        <div class="spinner-border text-success"></div>
        <p class="mt-3 text-muted">AI is checking your translations…</p>
      </div>

      <!-- summary -->
      <template v-else-if="phase === 'done'">
        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body text-center py-4">
            <i class="bi bi-trophy-fill display-5 text-warning"></i>
            <h4 class="fw-bold mt-2 mb-1">
              <template v-if="gradedCount">{{ correctCount }} / {{ gradedCount }} translations correct</template>
              <template v-else>Practice finished</template>
            </h4>
            <div class="text-muted small">{{ practisedCount }} sentences · {{ words.length }} words</div>
            <div v-if="gradeError" class="alert alert-warning py-2 small mt-3 mb-0">
              {{ gradeError }}
              <button class="btn btn-sm btn-outline-secondary ms-2" @click="grade">Check again</button>
            </div>
          </div>
        </div>

        <div v-for="w in words" :key="w.card_id" class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <h5 class="fw-bold mb-0">{{ w.word }} <small class="text-muted fw-normal">— {{ w.meaning_vi }}</small></h5>
            <div v-for="e in w.examples.filter((x) => x.checked)" :key="e.id" class="mt-3 small">
              <div class="d-flex align-items-start gap-2">
                <div class="flex-grow-1 fst-italic">{{ e.en }}</div>
                <button class="btn btn-light btn-sm rounded-circle flex-shrink-0" title="Listen" @click="say(e.en, RATE_REPLAY)">
                  <i class="bi bi-volume-up-fill"></i>
                </button>
              </div>
              <div class="mt-1">
                <span class="text-muted">You:</span> {{ e.translation.trim() || '—' }}
                <span v-if="e.result?.graded" class="badge ms-1" :class="e.result.correct ? 'bg-success' : 'bg-danger'">
                  {{ e.result.correct ? '✓' : '✗' }} {{ e.result.score }}
                </span>
                <span v-if="e.tries" class="badge ms-1" :class="e.best >= PASS ? 'bg-success-subtle text-success' : 'bg-light text-dark'">
                  <i class="bi bi-mic"></i> {{ e.best }}%
                </span>
              </div>
              <div v-if="e.result?.feedback_vi" :class="e.result.correct ? 'text-success' : 'text-danger'">{{ e.result.feedback_vi }}</div>
              <div class="text-muted"><span>Sample:</span> {{ e.vi }}</div>
              <AskTutor class="mt-1" :context="{
                task: `Translate into Vietnamese: ${e.en}`, word: w.word, learner: e.translation, sample: e.vi, feedback: e.result?.feedback_vi }" />
            </div>
          </div>
        </div>

        <div class="d-flex flex-wrap gap-2 justify-content-center mb-4">
          <button class="btn btn-outline-warning" @click="load"><i class="bi bi-arrow-repeat"></i> Practise these words again (new sentences)</button>
          <router-link to="/review" class="btn btn-primary">Back to Study</router-link>
          <router-link to="/quiz" class="btn btn-outline-primary">Quiz</router-link>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, errorMessage } from '../api'
import SpeakCheck from '../components/common/SpeakCheck.vue'
import AskTutor from '../components/common/AskTutor.vue'
import { speechRecognitionSupported } from '../composables/useSpeechRecognition'
import { RATE_NORMAL, RATE_REPLAY, RATE_SLOW, useSpeechSynthesis } from '../composables/useSpeechSynthesis'

interface GradeResult { example_id: number; correct: boolean; score: number; feedback_vi: string; graded: boolean }
interface Ex {
  id: number
  en: string
  vi: string
  translation: string
  checked: boolean // the sample translation was shown
  best: number // best reading-aloud score
  tries: number
  result?: GradeResult
}
interface Word { card_id: number; word: string; ipa?: string; meaning_vi: string; examples: Ex[] }

const PASS = 80
const MAX_TRIES = 3 // after 3 tries the learner may go on anyway

const route = useRoute()
const { speakEnglish } = useSpeechSynthesis()
const speechOK = speechRecognitionSupported

const phase = ref<'loading' | 'empty' | 'error' | 'practice' | 'grading' | 'done'>('loading')
const words = ref<Word[]>([])
const wi = ref(0)
const ei = ref(0)
const error = ref('')
const warning = ref('')
const gradeError = ref('')
const noMic = ref(storageGet('elearn.practice.nomic') === 'true')
watch(noMic, (v) => storageSet('elearn.practice.nomic', String(v)))

// words to practise: ?ids=1,2,3 from a quiz/study session, otherwise the
// words graded "Again" today and not practised yet (asked from the server)
const ids = ref<number[]>(
  String(route.query.ids ?? '')
    .split(',')
    .map(Number)
    .filter((n) => n > 0),
)
const idCount = computed(() => Math.min(ids.value.length, 10))
const word = computed(() => words.value[wi.value])
const ex = computed(() => word.value?.examples[ei.value])
const speakRequired = computed(() => speechOK && !noMic.value)
const exDone = computed(() => {
  const e = ex.value
  if (!e?.checked) return false
  return !speakRequired.value || e.best >= PASS || e.tries >= MAX_TRIES
})
const progressPct = computed(() => (words.value.length ? Math.round((wi.value / words.value.length) * 100) : 0))
const practised = computed(() => words.value.flatMap((w) => w.examples.filter((e) => e.checked)))
const practisedCount = computed(() => practised.value.length)
const gradedCount = computed(() => practised.value.filter((e) => e.result?.graded).length)
const correctCount = computed(() => practised.value.filter((e) => e.result?.graded && e.result.correct).length)

function storageGet(k: string) {
  try { return localStorage.getItem(k) } catch { return null }
}
function storageSet(k: string, v: string) {
  try { localStorage.setItem(k, v) } catch { /* ignore */ }
}

function say(text: string, rate: number) {
  speakEnglish(text, rate).catch(() => {})
}

// read each new sentence once at normal speed
watch(ex, (e, old) => {
  if (e && e !== old && phase.value === 'practice') say(e.en, RATE_NORMAL)
})

/** Sentence split around the practised word (any form: negotiate → negotiated, carry out → carried out). */
function segments(text: string, w: string): { text: string; hit: boolean }[] {
  const parts = w.toLowerCase().trim().split(/\s+/).map((p, i) => {
    let x = p.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    if (i === 0 && x.length > 3 && /[ey]$/.test(x)) x = x.slice(0, -1)
    return x + '\\w*'
  })
  const re = new RegExp('\\b' + parts.join('(?:\\W+\\w+){0,3}?\\W+') + '\\b', 'gi')
  const out: { text: string; hit: boolean }[] = []
  let last = 0
  for (const m of text.matchAll(re)) {
    const i = m.index ?? 0
    if (i > last) out.push({ text: text.slice(last, i), hit: false })
    out.push({ text: m[0], hit: true })
    last = i + m[0].length
  }
  if (last < text.length) out.push({ text: text.slice(last), hit: false })
  return out
}

function check() {
  const e = ex.value
  if (!e || e.checked || e.translation.trim().length < 2) return
  e.checked = true
}

function giveUp() {
  const e = ex.value
  if (!e) return
  e.translation = ''
  e.checked = true
}

function onScored(pct: number) {
  const e = ex.value
  if (!e) return
  e.tries++
  e.best = Math.max(e.best, pct)
}

function nextSentence() {
  if (word.value && ei.value < word.value.examples.length - 1) ei.value++
}

function nextWord() {
  if (wi.value < words.value.length - 1) {
    wi.value++
    ei.value = 0
  } else {
    grade()
  }
}

async function load() {
  phase.value = 'loading'
  error.value = ''
  warning.value = ''
  gradeError.value = ''
  try {
    if (!ids.value.length) {
      const { data } = await api.get('/practice/forgotten')
      ids.value = (data ?? []).map((w: { id: number }) => w.id).slice(0, 10)
    }
    if (!ids.value.length) {
      phase.value = 'empty'
      return
    }
    const { data } = await api.post('/practice/examples', { card_ids: ids.value.slice(0, 10) })
    words.value = (data.words ?? []).map((w: Omit<Word, 'examples'> & { examples: { id: number; en: string; vi: string }[] }) => ({
      ...w,
      examples: (w.examples ?? []).map((e) => ({ ...e, translation: '', checked: false, best: 0, tries: 0 })),
    }))
    warning.value = data.warning ?? ''
    if (!words.value.length) {
      phase.value = 'error'
      error.value = 'No example sentences could be prepared for these words.'
      return
    }
    wi.value = 0
    ei.value = 0
    phase.value = 'practice' // the watch on `ex` reads the first sentence
  } catch (e) {
    phase.value = 'error'
    error.value = errorMessage(e)
  }
}

/** One AI request checks every translation of the session. */
async function grade() {
  const items = practised.value
    .filter((e) => e.translation.trim())
    .map((e) => ({ example_id: e.id, translation: e.translation.trim() }))
  gradeError.value = ''
  if (!items.length) {
    phase.value = 'done'
    return
  }
  phase.value = 'grading'
  try {
    const { data } = await api.post('/practice/grade', { items })
    const byId = new Map<number, GradeResult>((data.results ?? []).map((r: GradeResult) => [r.example_id, r]))
    for (const e of practised.value) e.result = byId.get(e.id)
  } catch (e) {
    gradeError.value = `Could not check the translations: ${errorMessage(e)}`
  } finally {
    phase.value = 'done'
  }
}

onMounted(load)
</script>

<style scoped>
.sentence { line-height: 1.6; }
</style>
