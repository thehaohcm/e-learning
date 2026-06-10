<template>
  <div>
    <div class="text-center mb-4">
      <h1 class="fw-bold text-primary mb-2">
        <i class="bi bi-book"></i> Vocabulary Learning
      </h1>
      <p class="text-muted">AI-generated 3-Level Sentences for each word</p>
      <button @click="generateRandom" class="btn btn-warning btn-lg mt-2" :disabled="generating">
        <i :class="generating ? 'bi bi-hourglass-split' : 'bi bi-magic'"></i>
        {{ generating ? 'Generating...' : '✨ Generate Random by AI' }}
      </button>
      <p v-if="generatedCount > 0" class="text-success mt-1">
        <i class="bi bi-check-circle"></i> Generated {{ generatedCount }} words!
      </p>
    </div>

    <!-- Search & Topic Filter -->
    <div class="row mb-4">
      <div class="col-md-6">
        <input v-model="searchQuery" @keyup.enter="loadVocabs" class="form-control" placeholder="Search vocabulary...">
      </div>
      <div class="col-md-3">
        <select v-model="selectedTopic" @change="loadVocabs" class="form-select">
          <option value="">All Topics</option>
          <option v-for="t in topics" :key="t.id" :value="t.slug">{{ t.name_en }}</option>
        </select>
      </div>
      <div class="col-md-3">
        <button @click="loadVocabs" class="btn btn-primary w-100">
          <i class="bi bi-search"></i> Search
        </button>
      </div>
    </div>

    <!-- Vocabulary Cards with 3-Level Toggle -->
    <div v-if="vocabs.length > 0" class="row g-4">
      <div v-for="v in vocabs" :key="v.id" class="col-md-6 col-lg-4">
        <div class="card border-0 shadow-sm flashcard">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-start mb-2">
              <h4 class="text-primary fw-bold mb-0">{{ v.word }}</h4>
              <div class="d-flex gap-1">
                <button @click="speakWord(v)" class="btn btn-outline-secondary btn-sm" title="Listen">
                  <i :class="speakingWordId === v.id ? 'bi-volume-up-fill' : 'bi-volume-up'"></i>
                </button>
                <button @click="startLearning(v.id)" class="btn btn-outline-success btn-sm" title="Start learning">
                  <i class="bi bi-plus-circle"></i>
                </button>
              </div>
            </div>
            <small class="text-muted">{{ v.phonetics }} <span class="badge bg-light text-dark ms-1">{{ v.pos }}</span></small>

            <!-- Level Tabs -->
            <div class="mt-3">
              <ul class="nav nav-pills nav-fill mb-2">
                <li class="nav-item" v-for="lvl in levels" :key="lvl.level">
                  <button class="nav-link btn-sm"
                    :class="{ active: activeLevel[v.id] === lvl.level }"
                    @click="toggleLevel(v.id, lvl.level)"
                    :style="activeLevel[v.id] === lvl.level ? 'background:' + lvl.color : ''">
                    {{ lvl.label }}
                  </button>
                </li>
              </ul>

              <!-- Active Level Content -->
              <div v-for="s in getSentences(v, activeLevel[v.id])" :key="s.level" class="p-2 rounded" style="background: #f8f9fa">
                <div class="mb-2">
                  <span class="badge bg-primary me-1">EN</span> {{ s.meaning_en }}
                </div>
                <div class="mb-2">
                  <span class="badge bg-danger me-1">VI</span> {{ s.meaning_vi }}
                </div>
                <div class="mb-1 border-start border-3 border-primary ps-2">
                  <strong>Example:</strong><br>
                  <em>{{ s.example_en }}</em>
                </div>
                <div class="border-start border-3 border-warning ps-2 mb-2">
                  <em>{{ s.example_vi }}</em>
                </div>
                <div class="d-flex gap-1">
                  <button @click="speakSentence(s.example_en, s.example_vi)" class="btn btn-sm btn-outline-primary">
                    <i class="bi bi-play-circle"></i> Listen (Anh-Việt)
                  </button>
                </div>
                <small v-if="s.usage_note" class="text-muted d-block mt-1"><i class="bi bi-info-circle"></i> {{ s.usage_note }}</small>
              </div>
              <div v-if="getSentences(v, activeLevel[v.id]).length === 0" class="text-muted text-center py-2">
                <small>No sentence available for this level</small>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="text-center py-5 text-muted">
      <i class="bi bi-book display-1"></i>
      <p class="mt-2">Loading vocabulary...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref, onMounted, reactive } from 'vue'
import { useSpeechSynthesis } from '../composables/useSpeechSynthesis'

const API = (import.meta as any).env.VITE_API_BASE_URL || 'http://localhost:5173'
const USER_ID = 1
const { speakBilingual, speakEnglish, stop, isSpeaking } = useSpeechSynthesis()

interface Sentence {
  level: number
  meaning_en: string
  meaning_vi: string
  example_en: string
  example_vi: string
  usage_note: string
}

interface Vocab {
  id: number
  word: string
  phonetics: string
  pos: string
  sentences: Sentence[]
}

const vocabs = ref<Vocab[]>([])
const topics = ref<any[]>([])
const searchQuery = ref('')
const selectedTopic = ref('')
const activeLevel = reactive<Record<number, number>>({})
const speakingWordId = ref<number | null>(null)
const generating = ref(false)
const generatedCount = ref(0)

const levels = [
  { level: 1, label: 'Level 1 Simple', color: '#28a745' },
  { level: 2, label: 'Level 2 Normal', color: '#fd7e14' },
  { level: 3, label: 'Level 3 Complex', color: '#6f42c1' },
]

function getSentences(v: Vocab, level: number): Sentence[] {
  return v.sentences?.filter(s => s.level === level) || []
}

function toggleLevel(vocabId: number, level: number) {
  activeLevel[vocabId] = level
}

async function loadVocabs() {
  try {
    const params: any = { limit: 50 }
    if (selectedTopic.value) params.topic = selectedTopic.value
    if (searchQuery.value) params.search = searchQuery.value
    const { data } = await axios.get(`${API}/api/vocabs`, { params })
    vocabs.value = data.data || data || []
    vocabs.value.forEach(v => {
      if (!activeLevel[v.id]) activeLevel[v.id] = 1
    })
  } catch (e) {
    console.error('Load vocabs error:', e)
  }
}

async function loadTopics() {
  try {
    const { data } = await axios.get(`${API}/api/topics`)
    topics.value = data || []
  } catch (e) { console.error(e) }
}

async function speakWord(v: Vocab) {
  speakingWordId.value = v.id
  await speakEnglish(v.word)
  speakingWordId.value = null
}

async function speakSentence(en: string, vi: string) {
  await speakBilingual(en, vi)
}

async function startLearning(vocabId: number) {
  try {
    await axios.post(`${API}/api/srs/start`, { vocab_ids: [vocabId] }, { headers: { 'X-User-ID': String(USER_ID) } })
    alert('Word added to learning list!')
  } catch (e) {
    console.error(e)
  }
}

async function generateRandom() {
  generating.value = true
  try {
    const { data } = await axios.post(`${API}/api/vocabs/generate-random`, { count: 10 })
    generatedCount.value = data.generated || 0
    await loadVocabs()
  } catch (e) {
    console.error('Generate random error:', e)
    alert('Failed to generate vocabulary')
  }
  generating.value = false
}

onMounted(() => {
  loadTopics()
  loadVocabs()
})
</script>