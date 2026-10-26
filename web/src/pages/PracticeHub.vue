<template>
  <div>
    <div class="mb-3">
      <h2 class="fw-bold mb-0"><i class="bi bi-lightning-charge"></i> Practice</h2>
      <small class="text-muted">{{ isEnglish ? 'Extra practice on top of your daily Study: use your words in context.' : `${topicName}: study and quiz your decks.` }}</small>
    </div>

    <div class="row g-3">
      <!-- forgotten words -->
      <div v-if="isEnglish" class="col-md-6">
        <div class="card border-0 shadow-sm h-100 hoverable" :class="{ 'border-warning border-2 border': forgotten.length }">
          <div class="card-body d-flex flex-column">
            <h5 class="fw-bold"><i class="bi bi-translate text-warning"></i> Forgotten words</h5>
            <p class="small text-muted mb-2">New example sentences for the words you missed today: translate them, then read them aloud.</p>
            <div v-if="loading" class="placeholder-glow"><span class="placeholder col-6"></span></div>
            <div v-else-if="forgotten.length" class="mb-2">
              <span class="fw-semibold">{{ forgotten.length }} word{{ forgotten.length > 1 ? 's' : '' }} waiting:</span>
              <span v-for="w in forgotten.slice(0, 8)" :key="w.id" class="badge bg-warning-subtle text-dark border ms-1">{{ w.front }}</span>
              <span v-if="forgotten.length > 8" class="small text-muted"> +{{ forgotten.length - 8 }}</span>
            </div>
            <div v-else class="small text-success mb-2"><i class="bi bi-check-circle"></i> Nothing waiting right now.</div>
            <router-link to="/practice/forgotten" class="btn mt-auto align-self-start"
              :class="forgotten.length ? 'btn-warning' : 'btn-outline-secondary disabled'">
              Practise {{ forgotten.length ? Math.min(forgotten.length, 10) : '' }} words
            </router-link>
          </div>
        </div>
      </div>

      <!-- quiz -->
      <div class="col-md-6">
        <div class="card border-0 shadow-sm h-100 hoverable">
          <div class="card-body d-flex flex-column">
            <h5 class="fw-bold"><i class="bi bi-question-circle text-success"></i> Quiz</h5>
            <p v-if="isEnglish" class="small text-muted mb-2">Test any deck: choose the meaning, listening mode, or <b>Write it</b>: write the English sentence for a Vietnamese one, using the word.</p>
            <p v-else class="small text-muted mb-2">10 random questions from your {{ topicName }} decks. Wrong answers come back in Study.</p>
            <div class="mt-auto d-flex gap-2">
              <router-link to="/quiz" class="btn btn-success">Start a quiz</router-link>
              <router-link v-if="isEnglish" to="/quiz?mode=write" class="btn btn-outline-success"><i class="bi bi-pencil"></i> Write it</router-link>
            </div>
          </div>
        </div>
      </div>

      <!-- PMP: the decks, to study or browse one -->
      <div v-if="!isEnglish" class="col-md-6">
        <div class="card border-0 shadow-sm h-100 hoverable">
          <div class="card-body d-flex flex-column">
            <h5 class="fw-bold"><i class="bi bi-collection text-primary"></i> {{ topicName }} decks</h5>
            <p class="small text-muted mb-2">Study one deck only, or browse its cards.</p>
            <div v-for="d in subject?.decks ?? []" :key="d.id" class="d-flex align-items-center gap-2 small mb-1">
              <span class="flex-grow-1">{{ d.name }} <span class="text-muted">· {{ d.learning_count }}/{{ d.card_count }}</span></span>
              <router-link :to="`/review?deck_id=${d.id}`" class="btn btn-sm btn-warning py-0">Study</router-link>
              <router-link :to="`/learn?deck_id=${d.id}`" class="btn btn-sm btn-outline-primary py-0">Browse</router-link>
            </div>
          </div>
        </div>
      </div>

      <!-- reading -->
      <div v-if="isEnglish" class="col-md-6">
        <div class="card border-0 shadow-sm h-100 hoverable">
          <div class="card-body d-flex flex-column">
            <h5 class="fw-bold"><i class="bi bi-book-half text-primary"></i> Reading &amp; listening</h5>
            <p class="small text-muted mb-2">A TOEIC/IELTS passage or a daily conversation written with your words, with questions, shadowing and role-play.</p>
            <div v-if="lastReading" class="small mb-2">
              Latest: <router-link :to="`/reading?id=${lastReading.id}`">{{ lastReading.title || 'Untitled' }}</router-link>
              <span v-if="lastReading.score != null" class="badge bg-success ms-1">{{ lastReading.score }}/{{ lastReading.total }}</span>
              <span v-else-if="lastReading.status === 'completed'" class="badge bg-light text-dark border ms-1">not answered</span>
            </div>
            <router-link to="/reading" class="btn btn-primary mt-auto align-self-start">Open Reading</router-link>
          </div>
        </div>
      </div>

      <!-- writing -->
      <div v-if="isEnglish" class="col-md-6">
        <div class="card border-0 shadow-sm h-100 hoverable">
          <div class="card-body d-flex flex-column">
            <h5 class="fw-bold"><i class="bi bi-pencil text-danger"></i> Writing</h5>
            <p class="small text-muted mb-2">Write an email, an opinion or a short text: AI scores it, lists your mistakes and shows better versions.</p>
            <div v-if="lastWriting" class="small mb-2">
              Latest: {{ shorten(lastWriting.prompt || lastWriting.original_text) }}
              <span v-if="lastWriting.overall_score != null" class="badge bg-primary ms-1">{{ Math.round(lastWriting.overall_score) }}/100</span>
            </div>
            <router-link to="/writing" class="btn btn-danger mt-auto align-self-start">Open Writing</router-link>
          </div>
        </div>
      </div>
    </div>
    <p class="small text-muted mt-3 mb-0"><i class="bi bi-info-circle"></i> Practice is web-only; the study device does flashcards, quizzes and listening.</p>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '../api'
import { isEnglish, loadSubjects, shortName, subject } from '../composables/useSubject'
const topicName = computed(() => (subject.value ? shortName(subject.value.slug, subject.value.name) : ''))

interface ReadingSummary { id: number; title: string; status: string; score: number | null; total: number | null }
interface WritingSummary { id: number; prompt: string; original_text: string; overall_score: number | null }

const loading = ref(true)
const forgotten = ref<{ id: number; front: string }[]>([])
const lastReading = ref<ReadingSummary | null>(null)
const lastWriting = ref<WritingSummary | null>(null)

const shorten = (s: string) => (s.length > 50 ? s.slice(0, 50) + '…' : s)

onMounted(async () => {
  loadSubjects()
  // each tile works on its own: one failing request does not hide the others
  const [f, r, w] = await Promise.allSettled([api.get('/practice/forgotten'), api.get('/readings'), api.get('/writing')])
  if (f.status === 'fulfilled') forgotten.value = f.value.data ?? []
  if (r.status === 'fulfilled') lastReading.value = r.value.data?.[0] ?? null
  if (w.status === 'fulfilled') lastWriting.value = w.value.data?.[0] ?? null
  loading.value = false
})
</script>
