<template>
  <div class="row justify-content-center">
    <div class="col-lg-9">
      <h2 class="fw-bold text-info mb-1"><i class="bi bi-pencil"></i> Writing Lab</h2>
      <p class="text-muted">AI scores your English writing, explains mistakes in Vietnamese and rewrites it at 3 levels.</p>

      <!-- task from a reading ("Write about this") -->
      <div v-if="readingId" class="card border-0 shadow-sm mb-3 border-start border-danger border-4">
        <div class="card-body">
          <div v-if="taskLoading" class="text-muted small"><span class="spinner-border spinner-border-sm me-2"></span>Preparing the task…</div>
          <div v-else-if="taskError" class="alert alert-warning py-2 mb-0">{{ taskError }}</div>
          <template v-else-if="task">
            <div class="d-flex flex-wrap justify-content-between gap-2 mb-2">
              <small class="text-muted">
                Write about: <router-link :to="`/reading?id=${readingId}`">{{ taskTitle }}</router-link>
              </small>
              <router-link to="/writing" class="small">Write on my own topic instead</router-link>
            </div>
            <p class="fw-semibold mb-1">{{ task.instruction_en }}</p>
            <p class="small text-muted mb-2">{{ task.instruction_vi }}</p>
            <div class="small mb-2">
              <span class="me-2"><i class="bi bi-rulers"></i> {{ task.min_words }}–{{ task.max_words }} words</span>
              <span class="text-muted">Use:</span>
              <span v-for="w in task.required_words" :key="w" class="badge ms-1"
                :class="used(w) ? 'bg-success' : 'bg-light text-dark border'">
                <i v-if="used(w)" class="bi bi-check-lg"></i> {{ w }}
              </span>
            </div>

            <!-- help when stuck: outline → phrases → sample -->
            <div class="d-flex flex-wrap gap-2 small">
              <button class="btn btn-sm" :class="help >= 1 ? 'btn-secondary' : 'btn-outline-secondary'" @click="help = Math.max(help, 1)">
                <i class="bi bi-list-ol"></i> Outline
              </button>
              <button class="btn btn-sm" :class="help >= 2 ? 'btn-secondary' : 'btn-outline-secondary'" @click="help = Math.max(help, 2)">
                <i class="bi bi-chat-quote"></i> Useful phrases
              </button>
              <button v-if="task.sample_answer" class="btn btn-sm" :class="help >= 3 ? 'btn-secondary' : 'btn-outline-secondary'" @click="showSample">
                <i class="bi bi-file-text"></i> Sample answer
              </button>
            </div>
            <ol v-if="help >= 1" class="small mt-2 mb-0">
              <li v-for="(o, i) in task.outline_vi" :key="i">{{ o }}</li>
            </ol>
            <div v-if="help >= 2" class="small mt-2">
              <span class="text-muted">Click to add to your text:</span>
              <ul class="list-unstyled mb-0 mt-1">
                <li v-for="(ph, i) in task.useful_phrases" :key="i">
                  <a href="#" @click.prevent="insert(ph.phrase)">{{ ph.phrase }}</a>
                  <span class="text-muted"> — {{ ph.meaning_vi }}</span>
                </li>
              </ul>
            </div>
            <div v-if="help >= 3" class="mt-2 p-2 rounded bg-light small">
              <div class="text-muted mb-1"><i class="bi bi-info-circle"></i> Read it, then close it and write your own version: don't copy.</div>
              <div style="white-space: pre-wrap">{{ task.sample_answer }}</div>
              <button class="btn btn-link btn-sm p-0 mt-1" @click="help = 2">Hide sample</button>
            </div>
          </template>
        </div>
      </div>

      <div class="card border-0 shadow-sm mb-4">
        <div class="card-body">
          <template v-if="!readingId">
            <label class="form-label fw-semibold">Task</label>
            <textarea v-model="prompt" class="form-control mb-3" rows="2"
              placeholder="E.g. Describe your daily work routine (IELTS Task 2 style questions work well too)"></textarea>
          </template>
          <div class="d-flex justify-content-between align-items-center mb-2">
            <label class="form-label fw-semibold mb-0">Your writing</label>
            <button type="button" class="btn btn-sm" :class="bankOpen ? 'btn-primary' : 'btn-outline-primary'" @click="toggleBank">
              <i class="bi bi-journal-bookmark"></i> Phrase bank
            </button>
          </div>

          <!-- phrase bank: what you want to say → the English frame (click to add it) -->
          <div v-if="bankOpen" class="bank border rounded p-2 mb-2">
            <div v-if="bankLoading" class="small text-muted"><span class="spinner-border spinner-border-sm me-1"></span>Loading…</div>
            <div v-else-if="bankError" class="small text-danger">{{ bankError }}</div>
            <template v-else>
              <div class="d-flex flex-wrap gap-1 mb-2">
                <input v-model="bankQuery" class="form-control form-control-sm bank-search" placeholder="Search: lý do, however, kết thư…" />
                <select v-model="bankGroup" class="form-select form-select-sm w-auto">
                  <option value="">All ({{ bankTotal }})</option>
                  <option v-for="g in bankGroups" :key="g.name" :value="g.name">{{ g.name }} ({{ g.items.length }})</option>
                </select>
              </div>
              <div class="bank-list">
                <template v-for="g in shownGroups" :key="g.name">
                  <div class="small fw-bold text-muted mt-1">{{ g.name }}</div>
                  <div v-for="ph in g.items" :key="ph.id" class="bank-item py-1">
                    <div class="d-flex align-items-start gap-2">
                      <div class="flex-grow-1 min-w-0 small">
                        <span class="text-muted">{{ ph.vi }}</span>
                        <i class="bi bi-arrow-right mx-1 text-muted"></i>
                        <a href="#" class="fw-semibold" title="Add to your text" @click.prevent="insert(frameText(ph.pattern))">{{ ph.pattern }}</a>
                        <button type="button" class="btn btn-link btn-sm p-0 ms-1 align-baseline" :title="openPhrase === ph.id ? 'Hide' : 'Example'"
                          @click="openPhrase = openPhrase === ph.id ? 0 : ph.id">
                          <i class="bi" :class="openPhrase === ph.id ? 'bi-chevron-up' : 'bi-info-circle'"></i>
                        </button>
                        <div v-if="openPhrase === ph.id" class="mt-1 ps-2 border-start">
                          <div v-if="ph.note_vi" class="text-muted">{{ ph.note_vi }}</div>
                          <div class="fst-italic" style="white-space: pre-line">“{{ ph.example }}”</div>
                          <div class="text-muted" style="white-space: pre-line">{{ ph.example_vi }}</div>
                        </div>
                      </div>
                      <i v-if="ph.in_study" class="bi bi-check2-circle text-success" title="In Study"></i>
                    </div>
                  </div>
                </template>
                <div v-if="!shownGroups.length" class="small text-muted">No frame matches.</div>
              </div>
              <div class="d-flex flex-wrap justify-content-between align-items-center gap-2 mt-2 pt-2 border-top small">
                <span class="text-muted">Click a frame to add it where the cursor is.</span>
                <span v-if="bankInStudy >= bankTotal" class="text-success"><i class="bi bi-check2-circle"></i> All {{ bankTotal }} in Study</span>
                <button v-else type="button" class="btn btn-sm btn-success" :disabled="bankBusy" @click="studyBank">
                  <i class="bi bi-plus-circle"></i> Learn these in Study ({{ bankInStudy }}/{{ bankTotal }})
                </button>
              </div>
              <div v-if="bankInfo" class="small text-success mt-1">{{ bankInfo }}</div>
            </template>
          </div>

          <textarea ref="textBox" v-model="text" class="form-control" rows="8" placeholder="Write here…"></textarea>
          <div class="d-flex flex-wrap justify-content-between align-items-center gap-2 mt-2">
            <small :class="countClass">
              {{ wordCount }} words<template v-if="task"> (target {{ task.min_words }}–{{ task.max_words }})</template>
              <template v-if="task && missingWords.length"> · not used yet: {{ missingWords.join(', ') }}</template>
            </small>
            <button class="btn btn-info text-white" :disabled="submitting || wordCount < 5 || !effectivePrompt.trim()" @click="submit">
              <i class="bi bi-send"></i> {{ submitting ? 'Sending…' : 'Get feedback' }}
            </button>
          </div>
          <div v-if="sampleSeen" class="small text-muted mt-1"><i class="bi bi-eye"></i> You looked at the sample answer.</div>
          <div v-if="error" class="alert alert-danger mt-3 mb-0 py-2">{{ error }}</div>
        </div>
      </div>

      <div v-if="current" class="card border-0 shadow-sm mb-4">
        <div class="card-body">
          <div v-if="current.status === 'processing' || current.status === 'pending'" class="text-center py-3">
            <div class="spinner-border spinner-border-sm me-2"></div> AI is reading your text… (usually 10-40 s)
          </div>
          <div v-else-if="current.status === 'failed'" class="alert alert-warning mb-0">{{ current.error_message || 'Evaluation failed.' }}</div>
          <template v-else>
            <div class="row g-2 mb-3 text-center">
              <div class="col-6 col-md-3" v-for="s in scores" :key="s.label">
                <div class="p-2 rounded bg-light">
                  <div class="fs-4 fw-bold" :class="`text-${s.color}`">
                    {{ Math.round(s.value ?? 0) }}<small class="fs-6 fw-normal text-muted">/100</small>
                  </div>
                  <small class="text-muted">{{ s.label }}</small>
                  <div v-if="s.label === 'Overall'" class="small fw-semibold" :class="scoreLevel(s.value).cls">{{ scoreLevel(s.value).text }}</div>
                </div>
              </div>
            </div>
            <p>{{ current.feedback?.summary_vi }}</p>

            <h6 v-if="current.feedback?.errors?.length" class="fw-bold mt-3">Mistakes</h6>
            <ul class="list-unstyled">
              <li v-for="(e, i) in current.feedback?.errors" :key="i" class="mb-2">
                <span class="text-danger text-decoration-line-through">{{ e.original }}</span>
                <i class="bi bi-arrow-right mx-1"></i>
                <span class="text-success fw-semibold">{{ e.correction }}</span>
                <div class="small text-muted">{{ e.explanation_vi }}</div>
              </li>
            </ul>

            <template v-if="current.feedback?.phrase_tips?.length">
              <h6 class="fw-bold mt-3"><i class="bi bi-journal-bookmark"></i> Use a frame</h6>
              <ul class="list-unstyled">
                <li v-for="(t, i) in current.feedback?.phrase_tips" :key="i" class="mb-2">
                  <span v-if="t.original" class="text-muted">{{ t.original }}</span>
                  <i v-if="t.original" class="bi bi-arrow-right mx-1"></i>
                  <span class="text-success fw-semibold">{{ t.rewrite }}</span>
                  <div class="small"><span class="badge bg-primary-subtle text-primary-emphasis">{{ t.pattern }}</span></div>
                </li>
              </ul>
            </template>

            <h6 class="fw-bold mt-3">Rewrites</h6>
            <ul class="nav nav-pills mb-2 small">
              <li class="nav-item" v-for="r in current.feedback?.rewrites" :key="r.level">
                <button class="nav-link py-1" :class="{ active: level === r.level }" @click="level = r.level">
                  {{ ['', 'Simple', 'Natural', 'Advanced'][r.level] || `Level ${r.level}` }}
                </button>
              </li>
            </ul>
            <div class="p-3 bg-light rounded" style="white-space: pre-wrap">{{ current.feedback?.rewrites?.find((r) => r.level === level)?.text }}</div>
            <AskTutor class="mt-3" :context="{
              task: current.prompt, learner: current.original_text,
              corrected: current.feedback?.rewrites?.find((r) => r.level === level)?.text,
              feedback: [current.feedback?.summary_vi, ...(current.feedback?.errors ?? []).map((e) => `${e.original} → ${e.correction}: ${e.explanation_vi}`)].filter(Boolean).join('\n') }" />
          </template>
        </div>
      </div>

      <div v-if="history.length">
        <h6 class="fw-bold">History</h6>
        <div class="list-group">
          <button v-for="h in history" :key="h.id" class="list-group-item list-group-item-action d-flex justify-content-between align-items-center"
            @click="open(h.id)">
            <span class="text-truncate me-2">{{ h.prompt }}</span>
            <span>
              <small class="text-muted me-2">{{ new Date(h.created_at).toLocaleDateString() }}</small>
              <span v-if="h.overall_score != null" class="badge bg-success">{{ Math.round(h.overall_score) }}/100</span>
              <span v-else class="badge bg-secondary">{{ h.status }}</span>
            </span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, errorMessage } from '../api'
import { usesWord } from '../utils/words'
import AskTutor from '../components/common/AskTutor.vue'

interface WritingTask {
  type: string
  instruction_en: string
  instruction_vi: string
  min_words: number
  max_words: number
  required_words: string[]
  outline_vi: string[]
  useful_phrases: { phrase: string; meaning_vi: string }[]
  sample_answer: string
}

interface Submission {
  id: number
  prompt: string
  status: string
  created_at: string
  error_message?: string
  original_text?: string
  overall_score?: number
  grammar_score?: number
  vocab_score?: number
  coherence_score?: number
  feedback?: {
    summary_vi?: string
    errors?: { original: string; correction: string; explanation_vi: string }[]
    rewrites?: { level: number; text: string }[]
    phrase_tips?: { original: string; pattern: string; rewrite: string }[]
  }
}

const prompt = ref('')
const text = ref('')
const submitting = ref(false)
const error = ref('')
const current = ref<Submission | null>(null)
const history = ref<Submission[]>([])
const level = ref(2)
let timer: number | undefined

const route = useRoute()
const readingId = computed(() => Number(route.query.reading_id) || 0)
const task = ref<WritingTask | null>(null)
const taskTitle = ref('')
const taskLoading = ref(false)
const taskError = ref('')
const help = ref(0) // 0 none, 1 outline, 2 + phrases, 3 + sample
const sampleSeen = ref(false)
const textBox = ref<HTMLTextAreaElement | null>(null)

const wordCount = computed(() => text.value.trim().split(/\s+/).filter(Boolean).length)
const used = (w: string) => usesWord(text.value, w)
const missingWords = computed(() => (task.value?.required_words ?? []).filter((w) => !used(w)))
const countClass = computed(() => {
  const t = task.value
  if (!t) return 'text-muted'
  if (wordCount.value < t.min_words) return 'text-muted'
  return wordCount.value > t.max_words ? 'text-warning' : 'text-success'
})
// what the evaluator sees as the task
const effectivePrompt = computed(() => {
  const t = task.value
  if (!readingId.value) return prompt.value
  if (!t) return ''
  return `${t.instruction_en} (${t.min_words}-${t.max_words} words; use these words: ${t.required_words.join(', ')})`
})

async function loadTask() {
  task.value = null
  help.value = 0
  sampleSeen.value = false
  if (!readingId.value) return
  taskLoading.value = true
  taskError.value = ''
  try {
    const { data } = await api.post(`/readings/${readingId.value}/writing-task`)
    task.value = data.task
    taskTitle.value = data.title
  } catch (e) {
    taskError.value = errorMessage(e)
  } finally {
    taskLoading.value = false
  }
}

function showSample() {
  help.value = 3
  sampleSeen.value = true
}

/** Add a useful phrase at the cursor. */
async function insert(phrase: string) {
  const el = textBox.value
  const p = phrase.replace(/…$/, ' ')
  if (!el) {
    text.value += (text.value && !text.value.endsWith(' ') ? ' ' : '') + p
    return
  }
  const start = el.selectionStart ?? text.value.length
  const before = text.value.slice(0, start)
  const sep = before && !/\s$/.test(before) ? ' ' : ''
  text.value = before + sep + p + text.value.slice(el.selectionEnd ?? start)
  await nextTick()
  const pos = (before + sep + p).length
  el.focus()
  el.setSelectionRange(pos, pos)
}

watch(readingId, loadTask)

// ---------------------------------------------------------------- phrase bank
interface Phrase { id: number; vi: string; pattern: string; note_vi?: string; example?: string; example_vi?: string; in_study: boolean }
interface PhraseGroup { name: string; items: Phrase[] }

const bankOpen = ref(storageGet('elearn.phrasebank') === 'true')
const bankGroups = ref<PhraseGroup[]>([])
const bankTotal = ref(0)
const bankInStudy = ref(0)
const bankLoading = ref(false)
const bankBusy = ref(false)
const bankError = ref('')
const bankInfo = ref('')
const bankQuery = ref('')
const bankGroup = ref('')
const openPhrase = ref(0)

function storageGet(k: string) {
  try { return localStorage.getItem(k) } catch { return null }
}
function storageSet(k: string, v: string) {
  try { localStorage.setItem(k, v) } catch { /* ignore */ }
}

/** Lower case, no Vietnamese accents: "ket thu" finds "Kết thư". */
const fold = (s: string) => s.toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '').replace(/đ/g, 'd')

const shownGroups = computed(() => {
  const q = fold(bankQuery.value.trim())
  return bankGroups.value
    .filter((g) => !bankGroup.value || g.name === bankGroup.value)
    .map((g) => ({
      name: g.name,
      items: !q ? g.items : g.items.filter((p) => fold(`${g.name} ${p.vi} ${p.pattern} ${p.example ?? ''}`).includes(q)),
    }))
    .filter((g) => g.items.length)
})

async function loadBank() {
  if (bankGroups.value.length || bankLoading.value) return
  bankLoading.value = true
  bankError.value = ''
  try {
    const { data } = await api.get('/writing/phrases')
    bankGroups.value = data.groups ?? []
    bankTotal.value = data.total ?? 0
    bankInStudy.value = data.in_study ?? 0
  } catch (e) {
    bankError.value = errorMessage(e)
  } finally {
    bankLoading.value = false
  }
}

function toggleBank() {
  bankOpen.value = !bankOpen.value
  storageSet('elearn.phrasebank', String(bankOpen.value))
  if (bankOpen.value) loadBank()
}

/** The frame as text to start from: "I would be grateful if you could + V" → "I would be grateful if you could ". */
function frameText(pattern: string): string {
  let t = pattern.replace(/^…\s*/, '')
  const plus = t.indexOf(' + ')
  if (plus > 0) t = t.slice(0, plus)
  t = t.replace(/\s*…\s*$/, '').trimEnd()
  return t + ' '
}

async function studyBank() {
  bankBusy.value = true
  bankInfo.value = ''
  try {
    const { data } = await api.post('/writing/phrases/study')
    bankGroups.value = bankGroups.value.map((g) => ({ ...g, items: g.items.map((p) => ({ ...p, in_study: true })) }))
    bankInStudy.value = bankTotal.value
    bankInfo.value = data.added
      ? `${data.added} frames added to Study: they come with your new cards each day (deck “Writing phrases”).`
      : 'Already in Study.'
  } catch (e) {
    bankError.value = errorMessage(e)
  } finally {
    bankBusy.value = false
  }
}
/** What a score (0-100) means; the AI is given the same bands. */
function scoreLevel(v?: number | null): { text: string; cls: string } {
  const n = Math.round(v ?? 0)
  if (n >= 80) return { text: 'Good', cls: 'text-success' }
  if (n >= 60) return { text: 'Fair', cls: 'text-primary' }
  if (n >= 40) return { text: 'Needs work', cls: 'text-warning' }
  return { text: 'Redo', cls: 'text-danger' }
}
const scores = computed(() => [
  { label: 'Overall', value: current.value?.overall_score, color: 'info' },
  { label: 'Grammar', value: current.value?.grammar_score, color: 'primary' },
  { label: 'Vocabulary', value: current.value?.vocab_score, color: 'success' },
  { label: 'Coherence', value: current.value?.coherence_score, color: 'warning' },
])

async function submit() {
  submitting.value = true
  error.value = ''
  try {
    const { data } = await api.post('/writing', { prompt: effectivePrompt.value, text: text.value })
    current.value = data
    poll(data.id)
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    submitting.value = false
  }
}

function poll(id: number) {
  clearTimeout(timer)
  timer = window.setTimeout(async () => {
    await open(id)
    if (current.value?.status === 'processing' || current.value?.status === 'pending') poll(id)
    else loadHistory()
  }, 3000)
}

async function open(id: number) {
  try {
    const { data } = await api.get(`/writing/${id}`)
    current.value = data
  } catch (e) {
    error.value = errorMessage(e)
  }
}

async function loadHistory() {
  try {
    const { data } = await api.get('/writing')
    history.value = data
  } catch { /* ignore */ }
}

onMounted(() => {
  loadHistory()
  loadTask()
  if (bankOpen.value) loadBank()
})
onUnmounted(() => clearTimeout(timer))
</script>

<style scoped>
.bank { background: #f8fafc; }
.bank-list { max-height: 320px; overflow-y: auto; }
.bank-item + .bank-item { border-top: 1px dashed #e2e8f0; }
.bank-search { flex: 1 1 180px; }
.min-w-0 { min-width: 0; }
</style>
