<template>
  <div>
    <div class="text-center mb-4">
      <h1 class="fw-bold text-info mb-2">
        <i class="bi bi-pencil"></i> Writing Lab
      </h1>
      <p class="text-muted">Submit your writing - AI will evaluate and provide 3-level rewrites</p>
    </div>

    <div class="row justify-content-center">
      <div class="col-md-8">
        <div class="card border-0 shadow-sm mb-4">
          <div class="card-body">
            <div class="mb-3">
              <label class="form-label fw-bold">Writing Prompt</label>
              <textarea v-model="prompt" class="form-control" rows="2" placeholder="E.g., 'Describe your daily work routine using at least 5 vocabulary words'"></textarea>
            </div>
            <div class="mb-3">
              <label class="form-label fw-bold">Your Writing</label>
              <textarea v-model="text" class="form-control" rows="8" placeholder="Write your essay here..."></textarea>
              <small class="text-muted">{{ text.length }} characters</small>
            </div>
            <button @click="submit" class="btn btn-info btn-lg w-100 text-white" :disabled="submitting">
              <i class="bi bi-send me-2"></i>
              {{ submitting ? 'Submitting...' : 'Submit for AI Feedback' }}
            </button>
          </div>
        </div>

        <!-- Feedback Result -->
        <div v-if="feedback && feedback.status === 'completed'" class="card border-0 shadow-lg">
          <div class="card-body">
            <h4 class="fw-bold mb-3">AI Feedback</h4>
            <div class="row g-3 mb-3">
              <div class="col-6 col-md-3" v-for="s in scores" :key="s.label">
                <div class="text-center p-2 rounded bg-light">
                  <div class="fw-bold" :class="'text-' + s.color">{{ s.value }}%</div>
                  <small class="text-muted">{{ s.label }}</small>
                </div>
              </div>
            </div>
            <div class="text-center">
              <p class="text-muted">AI is evaluating your writing... Detailed 3-level rewrites will appear here.</p>
            </div>
          </div>
        </div>

        <div v-if="feedback && feedback.status === 'pending'" class="alert alert-info text-center">
          <div class="spinner-border spinner-border-sm me-2"></div>
          AI is evaluating your writing... <button @click="checkStatus(feedback.id)" class="btn btn-sm btn-outline-info ms-2">Check Status</button>
        </div>
      </div>
    </div>

    <!-- Writing History -->
    <div v-if="history.length > 0" class="row justify-content-center mt-5">
      <div class="col-md-8">
        <h5 class="fw-bold mb-3">Your Writing History</h5>
        <div v-for="h in history" :key="h.id" class="card border-0 shadow-sm mb-2">
          <div class="card-body d-flex justify-content-between align-items-center">
            <div>
              <strong>{{ h.prompt?.substring(0, 60) }}...</strong>
              <br><small class="text-muted">{{ h.status }} | {{ new Date(h.created_at).toLocaleDateString() }}</small>
            </div>
            <span v-if="h.overall_score" class="badge bg-success fs-6">{{ Math.round(h.overall_score) }}%</span>
            <span v-else class="badge bg-secondary">Pending</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, onMounted } from 'vue'

const API = (import.meta as any).env.VITE_API_BASE_URL || 'http://localhost:5173'
const USER_ID = 1

const prompt = ref('')
const text = ref('')
const submitting = ref(false)
const feedback = ref<any>(null)
const history = ref<any[]>([])

const scores = [
  { label: 'Grammar', value: 72, color: 'primary' },
  { label: 'Vocabulary', value: 78, color: 'success' },
  { label: 'Coherence', value: 75, color: 'warning' },
  { label: 'Overall', value: 75, color: 'info' },
]

async function submit() {
  if (!prompt.value || !text.value) return alert('Please fill in both prompt and text')
  submitting.value = true
  try {
    const { data } = await axios.post(`${API}/api/writing/submit`, {
      prompt: prompt.value,
      text: text.value,
      required_vocab_ids: []
    }, { headers: { 'X-User-ID': String(USER_ID) } })
    feedback.value = data
    loadHistory()
  } catch (e) {
    console.error(e)
  }
  submitting.value = false
}

async function checkStatus(id: number) {
  try {
    const { data } = await axios.get(`${API}/api/writing/${id}`)
    feedback.value = data
  } catch (e) { console.error(e) }
}

async function loadHistory() {
  try {
    const { data } = await axios.get(`${API}/api/writing/history`, {
      headers: { 'X-User-ID': String(USER_ID) }
    })
    history.value = data || []
  } catch (e) { console.error(e) }
}

onMounted(loadHistory)
</script>