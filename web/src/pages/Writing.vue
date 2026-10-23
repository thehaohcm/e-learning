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
          <label class="form-label fw-semibold">Your writing</label>
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
                  <div class="fs-4 fw-bold" :class="`text-${s.color}`">{{ Math.round(s.value ?? 0) }}</div>
                  <small class="text-muted">{{ s.label }}</small>
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
              <span v-if="h.overall_score != null" class="badge bg-success">{{ Math.round(h.overall_score) }}</span>
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
})
onUnmounted(() => clearTimeout(timer))
</script>
