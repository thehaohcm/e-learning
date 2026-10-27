<template>
  <div>
    <div class="d-flex flex-wrap gap-2 align-items-end mb-3">
      <div class="flex-grow-1">
        <h2 class="fw-bold mb-1"><i class="bi bi-collection"></i> {{ deck?.name ?? (isAll ? 'All decks' : 'Browse cards') }}</h2>
        <small class="text-muted">{{ isAll ? `Search every ${topicName} deck.` : deck?.description }}</small>
      </div>
      <select v-model.number="deckId" class="form-select w-auto" @change="onDeckChange">
        <option :value="ALL">🔍 All {{ topicName }} decks</option>
        <option v-for="d in subjects[0]?.decks ?? []" :key="d.id" :value="d.id">{{ d.name }} ({{ d.card_count }})</option>
      </select>
    </div>

    <div class="row g-2 mb-3">
      <div class="col-md-6">
        <input v-model="search" @keyup.enter="load(1)" class="form-control" :placeholder="isAll ? 'Search all decks: word, Vietnamese meaning, PMP term…' : 'Search this deck…'" />
      </div>
      <div class="col-md-6 d-flex gap-2 flex-wrap justify-content-md-end">
        <button class="btn btn-primary" @click="load(1)"><i class="bi bi-search"></i></button>
        <button v-if="notAdded.length" class="btn btn-success" @click="addAll">
          <i class="bi bi-plus-circle"></i> Add {{ notAdded.length }} to Study
        </button>
        <router-link v-if="deckId > 0" :to="`/review?deck_id=${deckId}`" class="btn btn-warning">
          <i class="bi bi-arrow-repeat"></i> Study deck
        </router-link>
      </div>
    </div>

    <!-- admin tools -->
    <div v-if="auth.isAdmin && deck" class="card border-0 bg-light mb-3">
      <div class="card-body py-2 d-flex flex-wrap gap-2 align-items-center">
        <small class="text-muted me-auto"><i class="bi bi-shield-lock"></i> Admin</small>
        <button class="btn btn-sm btn-warning" @click="generate" :disabled="generating">
          <i class="bi bi-magic"></i> {{ generating ? 'Generating… (up to 2 min)' : 'Generate 10 with AI' }}
        </button>
        <button class="btn btn-sm btn-outline-secondary" @click="showForm = !showForm">
          <i class="bi bi-plus-lg"></i> Add card manually
        </button>
      </div>
      <div v-if="showForm" class="card-body pt-0">
        <CardForm :kind="deck.default_kind" :deck-id="deck.id" @created="onCreated" />
      </div>
    </div>

    <div v-if="message" class="alert py-2" :class="messageOk ? 'alert-success' : 'alert-danger'">{{ message }}</div>

    <div v-if="loading" class="text-center py-5"><div class="spinner-border"></div></div>
    <div v-else-if="cards.length" class="row g-3">
      <div v-for="c in cards" :key="c.id" class="col-md-6 col-lg-4">
        <FlashCard :card="c" :can-delete="auth.isAdmin" :can-edit="auth.isAdmin"
          @speak="(t: string) => speakEnglish(t, RATE_REPLAY)"
          @speak-sentence="(en: string) => speakEnglish(en, RATE_REPLAY)"
          @add="enroll([c.id])" @remove="remove(c)" @edit="startEdit(c)" />
      </div>
    </div>
    <div v-else class="text-center py-5 text-muted">
      <i class="bi bi-inbox display-4"></i>
      <p class="mt-2">No cards found.</p>
    </div>

    <!-- edit card (admin) -->
    <div v-if="editCard" class="edit-overlay" @click.self="editCard = null">
      <div class="card border-0 shadow-lg edit-dialog">
        <div class="card-header bg-white d-flex align-items-center gap-2">
          <i class="bi bi-pencil-square text-primary"></i>
          <span class="fw-bold">Edit card</span>
          <small class="text-muted">· {{ editCard.kind }}</small>
          <button type="button" class="btn-close ms-auto" aria-label="Close" @click="editCard = null"></button>
        </div>
        <div class="card-body">
          <CardForm :key="editCard.id" :kind="editCard.kind" :deck-id="editCard.deck_id" :card="editCard"
            @saved="onSaved" @cancel="editCard = null" />
        </div>
      </div>
    </div>

    <nav v-if="pages > 1" class="d-flex justify-content-center mt-4">
      <ul class="pagination">
        <li class="page-item" :class="{ disabled: page <= 1 }"><a class="page-link" href="#" @click.prevent="load(page - 1)">‹</a></li>
        <li class="page-item disabled"><span class="page-link">{{ page }} / {{ pages }}</span></li>
        <li class="page-item" :class="{ disabled: page >= pages }"><a class="page-link" href="#" @click.prevent="load(page + 1)">›</a></li>
      </ul>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, errorMessage } from '../api'
import { auth } from '../auth'
import FlashCard from '../components/flashcard/FlashCard.vue'
import CardForm from '../components/CardForm.vue'
import { RATE_REPLAY, useSpeechSynthesis } from '../composables/useSpeechSynthesis'
import type { CardWithProgress, Subject } from '../types'
import { loadSubjects, shortName, subject, subjectId } from '../composables/useSubject'

const route = useRoute()
const router = useRouter()
const { speakEnglish } = useSpeechSynthesis()

// only the topic chosen in the menu
const subjects = computed<Subject[]>(() => (subject.value ? [subject.value] : []))
const topicName = computed(() => (subject.value ? shortName(subject.value.slug, subject.value.name) : ''))
const ALL = -1 // "All decks": search everything (the old Search page)
const deckId = ref<number>(route.query.deck_id === 'all' ? ALL : Number(route.query.deck_id) || 0)
const isAll = computed(() => deckId.value === ALL)
const cards = ref<CardWithProgress[]>([])
const search = ref(String(route.query.q ?? ''))
const page = ref(1)
const total = ref(0)
const limit = 30
const loading = ref(false)
const generating = ref(false)
const showForm = ref(false)
const message = ref('')
const messageOk = ref(true)
const editCard = ref<CardWithProgress | null>(null)

const deck = computed(() => subjects.value.flatMap((s) => s.decks).find((d) => d.id === deckId.value))
const pages = computed(() => Math.max(1, Math.ceil(total.value / limit)))
const notAdded = computed(() => cards.value.filter((c) => !c.progress).map((c) => c.id))

function flash(msg: string, ok = true) {
  message.value = msg
  messageOk.value = ok
  setTimeout(() => { if (message.value === msg) message.value = '' }, 4000)
}

async function load(p = 1) {
  if (!deckId.value) return
  loading.value = true
  try {
    const params = { deck_id: isAll.value ? undefined : deckId.value, subject_id: isAll.value ? subjectId.value || undefined : undefined, q: search.value || undefined, page: p, limit }
    const { data } = await api.get('/cards', { params })
    cards.value = data.data
    total.value = data.total
    page.value = p
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    loading.value = false
  }
}

function onDeckChange() {
  router.replace({ query: { deck_id: isAll.value ? 'all' : deckId.value } })
  search.value = ''
  load(1)
}

async function enroll(ids: number[]) {
  try {
    const { data } = await api.post('/study/enroll', { card_ids: ids })
    flash(`Added ${data.added} card(s) to your learning list.`)
    await load(page.value)
  } catch (e) {
    flash(errorMessage(e), false)
  }
}

const addAll = () => enroll(notAdded.value)

async function remove(c: CardWithProgress) {
  if (!confirm(`Delete “${c.front}” for everyone? This also removes its learning history.`)) return
  try {
    await api.delete(`/cards/${c.id}`)
    cards.value = cards.value.filter((x) => x.id !== c.id)
    total.value--
  } catch (e) {
    flash(errorMessage(e), false)
  }
}

async function generate() {
  generating.value = true
  try {
    const { data } = await api.post(`/decks/${deckId.value}/generate`, { count: 10 })
    flash(`AI added ${data.created} new card(s)${data.skipped ? `, skipped ${data.skipped} duplicate/invalid` : ''}.`)
    await load(pages.value)
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    generating.value = false
  }
}

function onCreated() {
  flash('Card added.')
  load(page.value)
}

function startEdit(c: CardWithProgress) {
  editCard.value = c
}

async function onSaved() {
  editCard.value = null
  flash('Card updated.')
  await load(page.value)
}

/** The deck asked for when it belongs to the topic, else the topic's first deck ("all" stays). */
function deckOfTopic(want: number): number {
  const decks = subject.value?.decks ?? []
  return want === ALL || decks.some((d) => d.id === want) ? want : decks[0]?.id ?? 0
}

watch(subjectId, (id, old) => {
  if (!old || id === old) return
  deckId.value = deckOfTopic(deckId.value === ALL ? ALL : 0)
  load(1)
})

onMounted(async () => {
  try {
    await loadSubjects()
    deckId.value = deckOfTopic(deckId.value)
    await load(1)
  } catch (e) {
    flash(errorMessage(e), false)
  }
})
</script>

<style scoped>
.edit-overlay {
  position: fixed;
  inset: 0;
  z-index: 1060;
  background: rgba(15, 23, 42, 0.5);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 1.5rem 1rem;
  overflow-y: auto;
}
.edit-dialog {
  width: 100%;
  max-width: 640px;
  border-radius: 14px;
}
</style>
