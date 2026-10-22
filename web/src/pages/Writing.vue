<template>
  <div class="row justify-content-center">
    <div class="col-lg-9">
      <h2 class="fw-bold text-info mb-1"><i class="bi bi-pencil"></i> Writing Lab</h2>
      <p class="text-muted">AI scores your English writing, explains mistakes in Vietnamese and rewrites it at 3 levels.</p>

      <div class="card border-0 shadow-sm mb-4">
        <div class="card-body">
          <label class="form-label fw-semibold">Task</label>
          <textarea v-model="prompt" class="form-control mb-3" rows="2"
            placeholder="E.g. Describe your daily work routine (IELTS Task 2 style questions work well too)"></textarea>
          <label class="form-label fw-semibold">Your writing</label>
          <textarea v-model="text" class="form-control" rows="8" placeholder="Write here…"></textarea>
          <div class="d-flex justify-content-between align-items-center mt-2">
            <small class="text-muted">{{ wordCount }} words</small>
            <button class="btn btn-info text-white" :disabled="submitting || wordCount < 5 || !prompt.trim()" @click="submit">
              <i class="bi bi-send"></i> {{ submitting ? 'Sending…' : 'Get feedback' }}
            </button>
          </div>
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
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { api, errorMessage } from '../api'

interface Submission {
  id: number
  prompt: string
  status: string
  created_at: string
  error_message?: string
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

const wordCount = computed(() => text.value.trim().split(/\s+/).filter(Boolean).length)
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
    const { data } = await api.post('/writing', { prompt: prompt.value, text: text.value })
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

onMounted(loadHistory)
onUnmounted(() => clearTimeout(timer))
</script>
