<template>
  <div>
    <h2 class="fw-bold text-secondary mb-3"><i class="bi bi-search"></i> Search all decks</h2>
    <div class="input-group input-group-lg mb-4">
      <input v-model="query" @keyup.enter="search" class="form-control" placeholder="Word, Vietnamese meaning, PMP term…" />
      <button @click="search" class="btn btn-primary"><i class="bi bi-search"></i></button>
    </div>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-if="results.length" class="row g-3">
      <div v-for="c in results" :key="c.id" class="col-md-6 col-lg-4">
        <FlashCard :card="c" @speak="(t: string) => speakEnglish(t)"
          @speak-sentence="(en: string, vi?: string) => vi ? speakBilingual(en, vi) : speakEnglish(en)"
          @add="add(c)" />
      </div>
    </div>
    <p v-else-if="searched" class="text-center text-muted py-5">No results for “{{ lastQuery }}”.</p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { api, errorMessage } from '../api'
import FlashCard from '../components/flashcard/FlashCard.vue'
import { useSpeechSynthesis } from '../composables/useSpeechSynthesis'
import type { CardWithProgress } from '../types'

const { speakEnglish, speakBilingual } = useSpeechSynthesis()
const query = ref('')
const lastQuery = ref('')
const results = ref<CardWithProgress[]>([])
const searched = ref(false)
const error = ref('')

async function search() {
  const q = query.value.trim()
  if (!q) return
  error.value = ''
  try {
    const { data } = await api.get('/cards', { params: { q, limit: 30 } })
    results.value = data.data
    lastQuery.value = q
    searched.value = true
  } catch (e) {
    error.value = errorMessage(e)
  }
}

async function add(c: CardWithProgress) {
  try {
    await api.post('/study/enroll', { card_ids: [c.id] })
    const { data } = await api.get(`/cards/${c.id}`)
    results.value = results.value.map((x) => (x.id === c.id ? data : x))
  } catch (e) {
    error.value = errorMessage(e)
  }
}
</script>
