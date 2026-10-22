<template>
  <div class="row g-4">
    <!-- left: create + history -->
    <div class="col-lg-4 order-lg-2">
      <div class="card border-0 shadow-sm mb-3">
        <div class="card-body">
          <h5 class="fw-bold mb-1"><i class="bi bi-stars"></i> New reading</h5>
          <p class="small text-muted">AI writes it with words you are learning plus a few you will meet next.</p>
          <div class="btn-group w-100 mb-2" role="group">
            <button class="btn btn-sm" :class="kind === 'reading' ? 'btn-primary' : 'btn-outline-primary'" @click="kind = 'reading'">
              <i class="bi bi-newspaper"></i> Reading (TOEIC/IELTS)
            </button>
            <button class="btn btn-sm" :class="kind === 'conversation' ? 'btn-primary' : 'btn-outline-primary'" @click="kind = 'conversation'">
              <i class="bi bi-chat-dots"></i> Daily conversation
            </button>
          </div>
          <select v-model="level" class="form-select form-select-sm mb-2">
            <option value="easy">Easy (A2–B1)</option>
            <option value="medium">Medium (B1–B2, TOEIC)</option>
            <option value="ielts">Hard (B2–C1, IELTS)</option>
          </select>
          <button class="btn btn-warning w-100" :disabled="creating || isProcessing" @click="create">
            <span v-if="creating || isProcessing" class="spinner-border spinner-border-sm me-1"></span>
            {{ creating || isProcessing ? 'AI is writing… (30–90 s)' : 'Create' }}
          </button>
          <div v-if="createError" class="alert alert-danger py-2 small mt-2 mb-0">{{ createError }}</div>
        </div>
      </div>

      <div class="card border-0 shadow-sm">
        <div class="card-body">
          <h6 class="fw-bold">History</h6>
          <div v-if="!history.length" class="small text-muted">No readings yet.</div>
          <div class="list-group list-group-flush">
            <button v-for="h in history" :key="h.id" type="button"
              class="list-group-item list-group-item-action px-1 d-flex justify-content-between align-items-center"
              :class="{ active: current?.id === h.id }" @click="open(h.id)">
              <span class="text-truncate me-2">
                <i :class="h.kind === 'conversation' ? 'bi bi-chat-dots' : 'bi bi-newspaper'"></i>
                {{ h.title || (h.status === 'processing' ? 'Writing…' : h.status === 'failed' ? 'Failed' : 'Untitled') }}
              </span>
              <span class="small flex-shrink-0">
                <span v-if="h.score != null" class="badge bg-success">{{ h.score }}/{{ h.total }}</span>
                <span v-else-if="h.status === 'processing'" class="spinner-border spinner-border-sm"></span>
                <span v-else class="text-muted">{{ levelLabel(h.level) }}</span>
              </span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- right: the reading -->
    <div class="col-lg-8 order-lg-1">
      <div v-if="!current" class="text-center text-muted py-5">
        <i class="bi bi-book-half display-3"></i>
        <p class="mt-2">Create a reading or open one from your history.</p>
      </div>

      <div v-else-if="current.status === 'processing'" class="text-center py-5">
        <div class="spinner-border text-warning"></div>
        <p class="mt-3 text-muted">AI is writing your {{ current.kind === 'conversation' ? 'conversation' : 'reading' }}…</p>
      </div>

      <div v-else-if="current.status === 'failed'" class="alert alert-warning">
        {{ current.error_message || 'This reading could not be created.' }}
        <button class="btn btn-sm btn-outline-secondary ms-2" @click="remove(current.id)">Delete</button>
      </div>

      <template v-else>
        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body p-4">
            <div class="d-flex justify-content-between align-items-start gap-2 mb-2">
              <div>
                <h3 class="fw-bold mb-1">{{ doc.title }}</h3>
                <span class="badge bg-light text-dark me-1">{{ current.kind === 'conversation' ? 'Conversation' : 'Reading' }}</span>
                <span class="badge bg-light text-dark">{{ levelLabel(current.level) }}</span>
              </div>
              <button class="btn btn-sm btn-outline-danger" title="Delete" @click="remove(current.id)"><i class="bi bi-trash"></i></button>
            </div>

            <!-- audio + translation controls -->
            <div class="d-flex flex-wrap gap-2 align-items-center my-3">
              <button class="btn btn-primary btn-sm" :disabled="!!practice" @click="playAll(1)"><i class="bi bi-play-fill"></i> Listen</button>
              <button class="btn btn-outline-primary btn-sm" :disabled="!!practice" @click="playAll(0.75)">🐢 Slow</button>
              <button class="btn btn-outline-secondary btn-sm" :disabled="playing < 0 && playingQ < 0" @click="stopAll"><i class="bi bi-stop-fill"></i> Stop</button>
              <span class="vr mx-1"></span>
              <button class="btn btn-outline-danger btn-sm" :disabled="!canSpeak || !!practice"
                :title="canSpeak ? 'Listen to each sentence and repeat it' : 'Needs Chrome/Edge and HTTPS'" @click="startPractice('shadow')">
                <i class="bi bi-mic"></i> Shadowing
              </button>
              <button v-if="current.kind === 'conversation'" class="btn btn-sm" :class="roleSetup ? 'btn-danger' : 'btn-outline-danger'"
                :disabled="!canSpeak || !!practice" :title="canSpeak ? 'Play one of the speakers' : 'Needs Chrome/Edge and HTTPS'"
                @click="roleSetup = !roleSetup">
                <i class="bi bi-people"></i> Role-play
              </button>
              <div class="ms-auto d-flex gap-3">
                <div class="form-check form-switch mb-0 small" title="Listening practice: hide the text until you check your answers">
                  <input id="hideText" v-model="hideText" class="form-check-input" type="checkbox" />
                  <label class="form-check-label" for="hideText">Hide {{ current.kind === 'conversation' ? 'conversation' : 'text' }}</label>
                </div>
                <div class="form-check form-switch mb-0 small">
                  <input id="vi" v-model="showVi" class="form-check-input" type="checkbox" :disabled="textHidden" />
                  <label class="form-check-label" for="vi">Vietnamese</label>
                </div>
              </div>
            </div>
            <!-- role-play setup -->
            <div v-if="roleSetup && !practice" class="d-flex flex-wrap align-items-center gap-2 p-2 mb-3 rounded bg-light small">
              <span>I play</span>
              <select v-model="rpRole" class="form-select form-select-sm w-auto">
                <option v-for="s in speakerOrder" :key="s" :value="s">{{ s }}</option>
              </select>
              <div class="btn-group btn-group-sm">
                <button class="btn" :class="rpLevel === 'easy' ? 'btn-secondary' : 'btn-outline-secondary'" @click="rpLevel = 'easy'"
                  title="Your lines are shown: read them aloud">Easy</button>
                <button class="btn" :class="rpLevel === 'hard' ? 'btn-secondary' : 'btn-outline-secondary'" @click="rpLevel = 'hard'"
                  title="Only the Vietnamese meaning is shown: say it in English">Hard</button>
              </div>
              <button class="btn btn-danger btn-sm" :disabled="!rpRole" @click="startPractice('roleplay')">Start</button>
              <span class="text-muted">{{ rpLevel === 'easy' ? 'Your lines are shown, read them aloud.' : 'You see only the Vietnamese meaning of your lines.' }}</span>
            </div>

            <SpeakingPractice v-if="practice" :key="practiceKey" :items="practiceItems" :mode="practice.mode"
              :role="practice.role" :level="practice.level" :say="sayAs"
              @highlight="(i: number) => (playing = i)" @close="practice = null" />

            <div v-if="textHidden" class="alert alert-info py-2 small">
              <i class="bi bi-headphones"></i> Listening practice: press <b>Listen</b>, then answer the questions.
              The text appears after you check your answers.
              <button class="btn btn-link btn-sm p-0 align-baseline" @click="hideText = false">Show it now</button>
            </div>

            <!-- vocabulary chips -->
            <div v-if="!textHidden" class="mb-3">
              <small class="text-muted d-block mb-1">Words in this text (tap to hear):</small>
              <button v-for="v in doc.vocab" :key="v.card_id" type="button"
                class="btn btn-sm me-1 mb-1" :class="v.is_new ? 'btn-outline-warning' : 'btn-outline-success'"
                @click="selectWord(v)">
                {{ v.word }} <span v-if="v.is_new" class="badge bg-warning text-dark ms-1">new</span>
              </button>
            </div>
            <div v-if="selected && !textHidden" class="alert alert-light border py-2 d-flex flex-wrap align-items-center gap-2">
              <strong>{{ selected.word }}</strong>
              <span>— {{ selected.meaning_vi || selected.meaning_en }}</span>
              <button class="btn btn-sm btn-light rounded-circle" @click="speakEnglish(selected.word).catch(() => {})"><i class="bi bi-volume-up-fill"></i></button>
              <button v-if="selected.is_new && !added.has(selected.card_id)" class="btn btn-sm btn-success ms-auto" @click="learn(selected)">
                <i class="bi bi-plus-circle"></i> Add to learning
              </button>
              <span v-else-if="added.has(selected.card_id)" class="small text-success ms-auto"><i class="bi bi-check2"></i> Added</span>
            </div>

            <!-- reading passage -->
            <div v-if="current.kind === 'reading'" class="passage" :class="{ blurred: textHidden }">
              <p v-for="(para, pi) in paragraphs" :key="pi">
                <span v-for="s in para" :key="s.index" class="sentence" :class="{ playing: playing === s.index }"
                  title="Click to listen to this sentence" @click="playOne(s.index)">
                  <template v-for="(seg, k) in s.segments" :key="k">
                    <mark v-if="seg.word" :class="seg.word.is_new ? 'new-word' : 'known-word'"
                      @click.stop="selectWord(seg.word)">{{ seg.text }}</mark><template v-else>{{ seg.text }}</template>
                  </template>{{ ' ' }}
                </span>
              </p>
              <div v-if="showVi && !textHidden && doc.translation_vi" class="translation">{{ doc.translation_vi }}</div>
            </div>

            <!-- conversation -->
            <div v-else :class="{ blurred: textHidden }">
              <p v-if="doc.situation" class="text-muted fst-italic">{{ doc.situation }}</p>
              <div v-for="(line, i) in doc.lines" :key="i" class="d-flex mb-2"
                :class="speakerIndex(line.speaker) % 2 === 0 ? 'justify-content-start' : 'justify-content-end'">
                <div class="bubble" :class="[speakerIndex(line.speaker) % 2 === 0 ? 'left' : 'right', { playing: playing === i }]">
                  <div class="small fw-semibold mb-1">
                    {{ line.speaker }}
                    <button class="btn btn-link btn-sm p-0 ms-1 align-baseline" title="Listen" @click="playOne(i)">
                      <i class="bi bi-volume-up"></i>
                    </button>
                  </div>
                  <div v-if="hiddenForRole(line.speaker)" class="text-muted fst-italic">… your line …</div>
                  <div v-else>
                    <template v-for="(seg, k) in segmentsOf(line.text)" :key="k">
                      <mark v-if="seg.word" :class="seg.word.is_new ? 'new-word' : 'known-word'"
                        @click="selectWord(seg.word)">{{ seg.text }}</mark><template v-else>{{ seg.text }}</template>
                    </template>
                  </div>
                  <div v-if="showVi && !textHidden && line.vi" class="small text-muted mt-1">{{ line.vi }}</div>
                </div>
              </div>
              <div v-if="doc.useful_phrases?.length && !textHidden" class="mt-3">
                <h6 class="fw-bold">Useful everyday phrases</h6>
                <ul class="mb-0">
                  <li v-for="(p, i) in doc.useful_phrases" :key="i">
                    <a href="#" @click.prevent="speakEnglish(p.phrase).catch(() => {})">{{ p.phrase }}</a>
                    <span class="text-muted"> — {{ p.meaning_vi }}</span>
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <!-- quiz -->
        <div class="card border-0 shadow-sm">
          <div class="card-body p-4">
            <div class="d-flex flex-wrap justify-content-between align-items-center gap-2 mb-3">
              <h5 class="fw-bold mb-0"><i class="bi bi-question-circle"></i> Questions</h5>
              <div class="d-flex flex-wrap align-items-center gap-2">
                <button v-if="playingQ >= 0" class="btn btn-outline-secondary btn-sm" @click="stopAll">
                  <i class="bi bi-stop-fill"></i> Stop
                </button>
                <button v-else class="btn btn-outline-primary btn-sm" title="Read all questions and their options" @click="playQuestions">
                  <i class="bi bi-volume-up-fill"></i> Listen to questions
                </button>
                <div class="form-check form-switch mb-0 small" title="Listening practice: hear the questions instead of reading them">
                  <input id="hideQ" v-model="hideQuestions" class="form-check-input" type="checkbox" />
                  <label class="form-check-label" for="hideQ">Hide questions</label>
                </div>
                <div class="form-check form-switch mb-0 small" title="Listening practice: read the question, hear the options A–D and pick a letter">
                  <input id="hideO" v-model="hideOptions" class="form-check-input" type="checkbox" />
                  <label class="form-check-label" for="hideO">Hide options</label>
                </div>
                <span v-if="results" class="badge fs-6" :class="scorePct >= 70 ? 'bg-success' : 'bg-warning text-dark'">
                  {{ current.score }}/{{ current.total }}
                </span>
              </div>
            </div>
            <div v-for="(q, qi) in doc.questions" :key="qi" class="mb-4 question" :class="{ playing: playingQ === qi }">
              <div class="d-flex align-items-start gap-2 mb-2">
                <p class="fw-semibold mb-0 flex-grow-1">
                  {{ qi + 1 }}.
                  <template v-if="questionsHidden"><span class="text-muted fw-normal">Listen to question {{ qi + 1 }}</span></template>
                  <template v-else>{{ q.question }}</template>
                </p>
                <button class="btn btn-light btn-sm rounded-circle flex-shrink-0" :title="heardQ.has(qi) ? 'Listen again (slower)' : 'Listen to the question and options'"
                  @click="playQuestion(qi)">
                  <i class="bi bi-volume-up-fill"></i>
                </button>
              </div>
              <div :class="{ 'd-flex gap-2': optionsHidden }">
                <button v-for="(c, ci) in q.choices" :key="ci" type="button" :disabled="!!results"
                  class="btn text-start mb-1 py-2 choice" :class="[choiceClass(qi, ci), optionsHidden ? 'flex-fill text-center' : 'w-100']"
                  @click="answers[qi] = ci">
                  <span class="badge bg-light text-dark" :class="{ 'me-2': !optionsHidden }">{{ 'ABCD'[ci] }}</span><template v-if="!optionsHidden">{{ c }}</template>
                </button>
              </div>
              <div v-if="results" class="small mt-2 p-2 rounded bg-light">
                <div :class="results[qi].correct ? 'text-success' : 'text-danger'" class="fw-semibold">
                  {{ results[qi].correct ? 'Correct' : results[qi].chosen < 0 ? 'Skipped' : 'Not quite' }}
                </div>
                <div>{{ results[qi].explanation }}</div>
                <div v-if="results[qi].evidence" class="fst-italic text-muted mt-1">“{{ results[qi].evidence }}”</div>
              </div>
            </div>
            <button v-if="!results" class="btn btn-success" :disabled="submitting" @click="submit">
              Check answers ({{ answeredCount }}/{{ doc.questions.length }})
            </button>
            <p v-if="results" class="small text-muted mb-0">
              Questions about your vocabulary also update its review schedule.
            </p>
          </div>
        </div>

        <!-- write about it -->
        <div class="card border-0 shadow-sm mt-3">
          <div class="card-body d-flex flex-wrap align-items-center gap-2">
            <div class="me-auto">
              <div class="fw-bold"><i class="bi bi-pencil text-danger"></i> Write about this</div>
              <small class="text-muted">
                {{ current.kind === 'conversation' ? 'A short message as one of the speakers' : 'A reply or your opinion' }},
                using words from this {{ current.kind === 'conversation' ? 'conversation' : 'text' }}. AI scores it.
              </small>
            </div>
            <router-link :to="`/writing?reading_id=${current.id}`" class="btn btn-outline-danger">Start writing</router-link>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, errorMessage } from '../api'
import { RATE_NORMAL, RATE_REPLAY, useSpeechSynthesis } from '../composables/useSpeechSynthesis'
import { speechRecognitionSupported } from '../composables/useSpeechRecognition'
import SpeakingPractice from '../components/reading/SpeakingPractice.vue'
import type { PracticeItem } from '../components/reading/practice'

interface Word { card_id: number; word: string; meaning_vi: string; meaning_en?: string; is_new: boolean }
interface Question { question: string; choices: string[]; answer: number }
interface Doc {
  title: string
  situation?: string
  passage?: string
  translation_vi?: string
  lines?: { speaker: string; text: string; vi?: string }[]
  useful_phrases?: { phrase: string; meaning_vi: string }[]
  questions: Question[]
  vocab: Word[]
}
interface Result { chosen: number; answer: number; correct: boolean; explanation: string; evidence: string }
interface Reading {
  id: number
  kind: 'reading' | 'conversation'
  level: string
  title: string
  status: string
  error_message?: string
  score: number | null
  total: number | null
  content: Doc
  results: Result[] | null
}
interface Summary { id: number; kind: string; level: string; title: string; status: string; score: number | null; total: number | null }
interface Segment { text: string; word?: Word }

const route = useRoute()
const router = useRouter()
const { speakEnglish, speakAs, stop } = useSpeechSynthesis()

const kind = ref<'reading' | 'conversation'>('reading')
const level = ref('medium')
const creating = ref(false)
const createError = ref('')
const history = ref<Summary[]>([])
const current = ref<Reading | null>(null)
const showVi = ref(false)
const selected = ref<Word | null>(null)
const added = reactive(new Set<number>())
const answers = ref<number[]>([])
const submitting = ref(false)
const playing = ref(-1)
const playingQ = ref(-1) // question being read
const heardQ = reactive(new Set<number>()) // questions already heard once: replays are slower
const hideText = ref(storageGet('elearn.reading.hideText') === 'true')
const hideQuestions = ref(storageGet('elearn.reading.hideQ') === 'true')
const hideOptions = ref(storageGet('elearn.reading.hideO') === 'true')
watch(hideOptions, (v) => storageSet('elearn.reading.hideO', String(v)))
watch(hideText, (v) => storageSet('elearn.reading.hideText', String(v)))
watch(hideQuestions, (v) => storageSet('elearn.reading.hideQ', String(v)))
let playToken = 0
let pollTimer: number | undefined

// normalise: the API may send null for empty lists
const doc = computed<Doc>(() => {
  const c = current.value?.content
  return {
    ...(c ?? {}),
    title: c?.title ?? '',
    questions: c?.questions ?? [],
    vocab: c?.vocab ?? [],
    lines: c?.lines ?? [],
    useful_phrases: c?.useful_phrases ?? [],
  }
})
const results = computed(() => current.value?.results ?? null)
// listening practice: everything is shown again once the answers are checked
const textHidden = computed(() => hideText.value && !results.value)
const questionsHidden = computed(() => hideQuestions.value && !results.value)
const optionsHidden = computed(() => hideOptions.value && !results.value)

function storageGet(k: string) {
  try { return localStorage.getItem(k) } catch { return null }
}
function storageSet(k: string, v: string) {
  try { localStorage.setItem(k, v) } catch { /* ignore */ }
}
const isProcessing = computed(() => history.value.some((h) => h.status === 'processing'))
const answeredCount = computed(() => answers.value.filter((a) => a >= 0).length)
const scorePct = computed(() => (current.value?.total ? ((current.value.score ?? 0) / current.value.total) * 100 : 0))

const levelLabel = (l: string) => ({ easy: 'Easy', medium: 'Medium', ielts: 'IELTS' }[l] ?? l)

// ---------------------------------------------------------------- highlighting

/** Regexp for one vocab item, matching common forms (negotiate→negotiated, carry out→carried out). */
function wordSource(word: string): string {
  const parts = word.toLowerCase().trim().split(/\s+/).map((p, i) => {
    let q = p.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    if (i === 0 && q.length > 3 && /[ey]$/.test(q)) q = q.slice(0, -1)
    return q + '\\w*'
  })
  return '\\b' + parts.join('(?:\\W+\\w+){0,3}?\\W+') + '\\b'
}

const vocabRegex = computed(() => {
  const vocab = doc.value.vocab ?? []
  if (!vocab.length) return null
  // longer items first so "carry out" wins over "carry"
  const sorted = [...vocab].sort((a, b) => b.word.length - a.word.length)
  return { re: new RegExp(sorted.map((v) => `(${wordSource(v.word)})`).join('|'), 'gi'), sorted }
})

function segmentsOf(text: string): Segment[] {
  const vr = vocabRegex.value
  if (!vr) return [{ text }]
  const out: Segment[] = []
  let last = 0
  vr.re.lastIndex = 0
  for (const m of text.matchAll(vr.re)) {
    const idx = m.index ?? 0
    const group = m.slice(1).findIndex((g) => g !== undefined)
    if (idx > last) out.push({ text: text.slice(last, idx) })
    out.push({ text: m[0], word: vr.sorted[group] })
    last = idx + m[0].length
  }
  if (last < text.length) out.push({ text: text.slice(last) })
  return out
}

interface Sentence { index: number; text: string; segments: Segment[] }

/** Passage → paragraphs → sentences (each sentence can be played on its own). */
const paragraphs = computed<Sentence[][]>(() => {
  let index = 0
  return (doc.value.passage ?? '')
    .split(/\n\s*\n/)
    .map((p) => p.trim())
    .filter(Boolean)
    .map((p) =>
      p.split(/(?<=[.!?])\s+(?=["'A-Z0-9])/).map((t) => ({ index: index++, text: t, segments: segmentsOf(t) })),
    )
})
const sentences = computed(() => paragraphs.value.flat())

const speakerOrder = computed(() => {
  const order: string[] = []
  for (const l of doc.value.lines ?? []) if (!order.includes(l.speaker)) order.push(l.speaker)
  return order
})
const speakerIndex = (name: string) => Math.max(0, speakerOrder.value.indexOf(name))

// ---------------------------------------------------------------- audio

/** Number of playable units: sentences for a reading, lines for a conversation. */
const unitCount = computed(() => (current.value?.kind === 'conversation' ? doc.value.lines?.length ?? 0 : sentences.value.length))

async function speakUnit(i: number, rate: number) {
  if (current.value?.kind === 'conversation') {
    const line = doc.value.lines![i]
    await speakAs(line.text, speakerIndex(line.speaker), rate)
  } else {
    await speakEnglish(sentences.value[i].text, rate)
  }
}

/** Read everything sentence by sentence (long texts stop early in some browsers if read in one go). */
async function playAll(rate: number) {
  const token = ++playToken
  playingQ.value = -1
  try {
    for (let i = 0; i < unitCount.value; i++) {
      if (token !== playToken) return
      playing.value = i
      await speakUnit(i, rate)
      await new Promise((r) => setTimeout(r, current.value?.kind === 'conversation' ? 350 : 150))
    }
  } catch {
    /* stopped */
  } finally {
    if (token === playToken) playing.value = -1
  }
}

async function playOne(i: number) {
  const token = ++playToken
  playing.value = i
  try {
    await speakUnit(i, 1)
  } catch {
    /* stopped */
  } finally {
    if (token === playToken) playing.value = -1
  }
}

function stopAll() {
  playToken++
  playing.value = -1
  playingQ.value = -1
  stop()
}

const pause = (ms: number) => new Promise((r) => setTimeout(r, ms))

/** "Question 1. … A. … B. …" — first time at normal speed, replays slower. */
async function readQuestion(qi: number, token: number) {
  const q = doc.value.questions[qi]
  const rate = heardQ.has(qi) ? RATE_REPLAY : RATE_NORMAL
  playingQ.value = qi
  await speakEnglish(`Question ${qi + 1}. ${q.question}`, rate)
  for (let ci = 0; ci < q.choices.length; ci++) {
    if (token !== playToken) return
    await pause(450)
    if (token !== playToken) return
    await speakEnglish(`${'ABCD'[ci]}. ${q.choices[ci]}`, rate)
  }
  heardQ.add(qi)
}

async function playQuestion(qi: number) {
  const token = ++playToken
  playing.value = -1
  try {
    await readQuestion(qi, token)
  } catch {
    /* stopped */
  } finally {
    if (token === playToken) playingQ.value = -1
  }
}

async function playQuestions() {
  const token = ++playToken
  playing.value = -1
  try {
    for (let qi = 0; qi < doc.value.questions.length; qi++) {
      if (token !== playToken) return
      await readQuestion(qi, token)
      await pause(900)
    }
  } catch {
    /* stopped */
  } finally {
    if (token === playToken) playingQ.value = -1
  }
}

// ---------------------------------------------------------------- speaking practice

const canSpeak = speechRecognitionSupported
const practice = ref<{ mode: 'shadow' | 'roleplay'; role?: string; level?: 'easy' | 'hard' } | null>(null)
const practiceKey = ref(0)
const roleSetup = ref(false)
const rpRole = ref('')
const rpLevel = ref<'easy' | 'hard'>('easy')
watch(speakerOrder, (s) => { if (!s.includes(rpRole.value)) rpRole.value = s[0] ?? '' }, { immediate: true })

/** Sentences of a reading, or lines of a conversation, for the practice panel. */
const practiceItems = computed<PracticeItem[]>(() =>
  current.value?.kind === 'conversation'
    ? (doc.value.lines ?? []).map((l, i) => ({ index: i, text: l.text, vi: l.vi, speaker: l.speaker }))
    : sentences.value.map((s) => ({ index: s.index, text: s.text })),
)

function startPractice(mode: 'shadow' | 'roleplay') {
  stopAll()
  roleSetup.value = false
  practice.value = mode === 'shadow' ? { mode } : { mode, role: rpRole.value, level: rpLevel.value }
  practiceKey.value++
}

/** Hard role-play: the learner's own lines are hidden in the conversation. */
function hiddenForRole(speaker: string) {
  const p = practice.value
  return !!p && p.mode === 'roleplay' && p.level === 'hard' && p.role === speaker
}

function sayAs(text: string, speaker: string | undefined, rate: number): Promise<void> {
  return speaker ? speakAs(text, speakerIndex(speaker), rate) : speakEnglish(text, rate)
}

function selectWord(w: Word) {
  selected.value = w
  speakEnglish(w.word).catch(() => {})
}

// ---------------------------------------------------------------- data

async function loadHistory() {
  try {
    const { data } = await api.get('/readings')
    history.value = Array.isArray(data) ? data : []
  } catch { /* ignore */ }
}

async function open(id: number) {
  stopAll()
  practice.value = null
  roleSetup.value = false
  heardQ.clear()
  selected.value = null
  try {
    const { data } = await api.get(`/readings/${id}`)
    if (data.content) {
      data.content.questions = data.content.questions ?? []
      data.content.vocab = data.content.vocab ?? []
    }
    current.value = data
    answers.value = (data.content?.questions ?? []).map((_: unknown, i: number) => data.results?.[i]?.chosen ?? -1)
    if (route.query.id !== String(id)) router.replace({ query: { id } })
    if (data.status === 'processing') schedulePoll(id)
  } catch (e) {
    createError.value = errorMessage(e)
  }
}

function schedulePoll(id: number) {
  clearTimeout(pollTimer)
  pollTimer = window.setTimeout(async () => {
    await loadHistory()
    if (current.value?.id === id) await open(id)
    else if (history.value.some((h) => h.status === 'processing')) schedulePoll(id)
  }, 4000)
}

async function create() {
  creating.value = true
  createError.value = ''
  try {
    const { data } = await api.post('/readings', { kind: kind.value, level: level.value })
    await loadHistory()
    await open(data.id)
  } catch (e) {
    createError.value = errorMessage(e)
  } finally {
    creating.value = false
  }
}

async function submit() {
  if (!current.value) return
  submitting.value = true
  try {
    const { data } = await api.post(`/readings/${current.value.id}/answers`, { answers: answers.value })
    current.value.results = data.results
    current.value.score = data.score
    current.value.total = data.total
    // reveal the correct answers in the question list
    data.results.forEach((r: Result, i: number) => { current.value!.content.questions[i].answer = r.answer })
    loadHistory()
  } catch (e) {
    createError.value = errorMessage(e)
  } finally {
    submitting.value = false
  }
}

function choiceClass(qi: number, ci: number) {
  const r = results.value?.[qi]
  if (!r) return answers.value[qi] === ci ? 'btn-primary' : 'btn-outline-primary'
  if (ci === r.answer) return 'btn-success'
  if (ci === r.chosen) return 'btn-danger'
  return 'btn-outline-secondary opacity-50'
}

async function learn(w: Word) {
  try {
    await api.post('/study/enroll', { card_ids: [w.card_id] })
    added.add(w.card_id)
  } catch (e) {
    createError.value = errorMessage(e)
  }
}

async function remove(id: number) {
  if (!confirm('Delete this reading?')) return
  try {
    await api.delete(`/readings/${id}`)
    if (current.value?.id === id) {
      current.value = null
      router.replace({ query: {} })
    }
    loadHistory()
  } catch (e) {
    createError.value = errorMessage(e)
  }
}

onMounted(async () => {
  await loadHistory()
  const id = Number(route.query.id) || history.value[0]?.id
  if (id) open(id)
})
onUnmounted(() => {
  clearTimeout(pollTimer)
  stopAll()
})
</script>

<style scoped>
.passage {
  font-size: 1.08rem;
  line-height: 1.8;
}
.sentence {
  cursor: pointer;
  border-radius: 4px;
  transition: background 0.2s;
}
.sentence:hover {
  background: #f1f5ff;
}
.sentence.playing,
.bubble.playing {
  background: #fff3cd;
}
mark {
  padding: 0 2px;
  border-radius: 3px;
  cursor: pointer;
}
mark.known-word {
  background: #d1e7dd;
}
mark.new-word {
  background: #ffe69c;
  font-weight: 600;
}
.translation {
  border-left: 3px solid #ffc107;
  padding-left: 0.75rem;
  color: #6c757d;
  white-space: pre-wrap;
}
.bubble {
  max-width: 80%;
  padding: 0.6rem 0.9rem;
  border-radius: 14px;
  transition: background 0.2s;
}
.bubble.left {
  background: #f1f3f5;
  border-top-left-radius: 4px;
}
.bubble.right {
  background: #e7f1ff;
  border-top-right-radius: 4px;
}
.choice {
  white-space: normal;
}
/* listening practice: the text stays in place (and shows progress) but can't be read */
.blurred {
  filter: blur(6px);
  user-select: none;
}
.blurred mark,
.blurred a {
  pointer-events: none;
}
.question {
  border-radius: 0.5rem;
  transition: background-color 0.2s;
}
.question.playing {
  background: rgba(13, 110, 253, 0.06);
  box-shadow: 0 0 0 6px rgba(13, 110, 253, 0.06);
}
</style>
