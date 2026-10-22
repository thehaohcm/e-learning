<template>
  <div>
    <div class="d-flex flex-wrap gap-2 align-items-end mb-3">
      <div class="flex-grow-1">
        <h2 class="fw-bold mb-1"><i class="bi bi-collection"></i> {{ deck?.name ?? 'Browse cards' }}</h2>
        <small class="text-muted">{{ deck?.description }}</small>
      </div>
      <select v-model.number="deckId" class="form-select w-auto" @change="onDeckChange">
        <optgroup v-for="s in subjects" :key="s.id" :label="s.name">
          <option v-for="d in s.decks" :key="d.id" :value="d.id">{{ d.name }} ({{ d.card_count }})</option>
        </optgroup>
      </select>
    </div>

    <div class="row g-2 mb-3">
      <div class="col-md-6">
        <input v-model="search" @keyup.enter="load(1)" class="form-control" placeholder="Search this deck…" />
      </div>
      <div class="col-md-6 d-flex gap-2 flex-wrap justify-content-md-end">
        <button class="btn btn-primary" @click="load(1)"><i class="bi bi-search"></i></button>
        <button class="btn btn-success" @click="addAll" :disabled="!notAdded.length">
          <i class="bi bi-plus-circle"></i> Add {{ notAdded.length }} on this page
        </button>
        <router-link v-if="deckId" :to="`/review?deck_id=${deckId}`" class="btn btn-warning">
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
        <FlashCard :card="c" :can-delete="auth.isAdmin"
          @speak="(t: string) => speakEnglish(t)"
          @speak-sentence="(en: string, vi?: string) => vi ? speakBilingual(en, vi) : speakEnglish(en)"
          @add="enroll([c.id])" @remove="remove(c)" />
      </div>
    </div>
    <div v-else class="text-center py-5 text-muted">
      <i class="bi bi-inbox display-4"></i>
      <p class="mt-2">No cards found.</p>
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
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, errorMessage } from '../api'
import { auth } from '../auth'
import FlashCard from '../components/flashcard/FlashCard.vue'
import CardForm from '../components/CardForm.vue'
import { useSpeechSynthesis } from '../composables/useSpeechSynthesis'
import type { CardWithProgress, Subject } from '../types'

const route = useRoute()
const router = useRouter()
const { speakEnglish, speakBilingual } = useSpeechSynthesis()

const subjects = ref<Subject[]>([])
const deckId = ref<number>(Number(route.query.deck_id) || 0)
const cards = ref<CardWithProgress[]>([])
const search = ref('')
const page = ref(1)
const total = ref(0)
const limit = 30
const loading = ref(false)
const generating = ref(false)
const showForm = ref(false)
const message = ref('')
const messageOk = ref(true)

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
    const { data } = await api.get('/cards', { params: { deck_id: deckId.value, q: search.value || undefined, page: p, limit } })
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
  router.replace({ query: { deck_id: deckId.value } })
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

onMounted(async () => {
  try {
    const { data } = await api.get('/subjects')
    subjects.value = data
    if (!deckId.value) deckId.value = data[0]?.decks[0]?.id ?? 0
    await load(1)
  } catch (e) {
    flash(errorMessage(e), false)
  }
})
</script>
