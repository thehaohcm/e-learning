<template>
  <div>
    <div class="text-center mb-4">
      <h1 class="fw-bold text-secondary mb-2">
        <i class="bi bi-search"></i> AI-Powered Search
      </h1>
      <p class="text-muted">Search vocabulary with full-text fuzzy matching</p>
    </div>

    <div class="row justify-content-center mb-4">
      <div class="col-md-8">
        <div class="input-group input-group-lg">
          <input v-model="query" @keyup.enter="search" class="form-control" placeholder="Search for words... e.g. 'agree', 'sustain', 'confid'">
          <button @click="search" class="btn btn-primary">
            <i class="bi bi-search"></i> Search
          </button>
        </div>
      </div>
    </div>

    <div v-if="results.length > 0" class="row g-3">
      <div v-for="v in results" :key="v.id" class="col-md-6 col-lg-4">
        <div class="card border-0 shadow-sm">
          <div class="card-body">
            <h5 class="text-primary fw-bold">{{ v.word }}</h5>
            <small class="text-muted">{{ v.phonetics }}</small>
            <div class="mt-2">
              <span v-for="s in v.sentences?.slice(0, 1)" :key="s.level" class="badge bg-light text-dark me-1">
                L{{ s.level }}: {{ s.meaning_en?.substring(0, 30) }}
              </span>
            </div>
            <div class="mt-2">
              <button @click="speakWord(v)" class="btn btn-sm btn-outline-secondary">
                <i class="bi bi-volume-up"></i>
              </button>
              <button @click="startLearn(v.id)" class="btn btn-sm btn-outline-success ms-1">
                <i class="bi bi-plus-circle"></i> Learn
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="searched" class="text-center py-5 text-muted">
      <i class="bi bi-search display-1"></i>
      <p class="mt-2">No results found for "{{ query }}"</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref } from 'vue'
import { useSpeechSynthesis } from '../composables/useSpeechSynthesis'

const API = (import.meta as any).env.VITE_API_BASE_URL || 'http://localhost:5173'
const USER_ID = 1
const { speakEnglish } = useSpeechSynthesis()

const query = ref('')
const results = ref<any[]>([])
const searched = ref(false)

async function search() {
  if (!query.value.trim()) return
  try {
    const { data } = await axios.get(`${API}/api/search`, { params: { q: query.value } })
    results.value = data.results || []
    searched.value = true
  } catch (e) { console.error(e) }
}

async function speakWord(v: any) {
  await speakEnglish(v.word)
}

async function startLearn(id: number) {
  try {
    await axios.post(`${API}/api/srs/start`, { vocab_ids: [id] }, { headers: { 'X-User-ID': String(USER_ID) } })
    alert('Word added!')
  } catch (e) { console.error(e) }
}
</script>