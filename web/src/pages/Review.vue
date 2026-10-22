<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-center mb-3 gap-2">
      <div>
        <h2 class="fw-bold mb-0"><i class="bi bi-arrow-repeat"></i> Study</h2>
        <small class="text-muted">{{ scopeLabel }} · spaced repetition (FSRS)</small>
      </div>
      <div class="d-flex flex-wrap gap-3 align-items-center small">
        <div class="btn-group btn-group-sm" role="group" aria-label="Review mode">
          <button class="btn" :class="mode === 'choose' ? 'btn-primary' : 'btn-outline-primary'" @click="mode = 'choose'"
            title="Pick the right meaning: wrong answers are graded automatically">
            <i class="bi bi-ui-checks"></i> Choose
          </button>
          <button class="btn" :class="mode === 'self' ? 'btn-primary' : 'btn-outline-primary'" @click="mode = 'self'"
            title="Recall in your head, reveal, then grade yourself">
            <i class="bi bi-eye"></i> Self-check
          </button>
        </div>
        <div class="form-check form-switch mb-0">
          <input id="autospeak" v-model="autoSpeak" class="form-check-input" type="checkbox" />
          <label class="form-check-label" for="autospeak">Auto pronounce</label>
        </div>
        <span class="text-muted">Done: {{ done }}</span>
      </div>
    </div>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <div v-if="item" class="row justify-content-center">
      <div class="col-lg-8">
        <ForgottenWords :words="forgotten" compact class="mb-3" />
        <div class="progress mb-3" style="height: 6px">
          <div class="progress-bar bg-warning" :style="{ width: progressPct + '%' }"></div>
        </div>

        <div class="card border-0 shadow-lg">
          <div class="card-body p-4 p-md-5">
            <!-- badges -->
            <div class="text-center mb-3">
              <span v-if="item.is_new" class="badge bg-info">New · learn it</span>
              <span v-else class="badge bg-secondary">{{ stateLabel(item.progress) }}</span>
              <span class="badge bg-light text-dark ms-1">{{ kindLabel }}</span>
            </div>

            <!-- front: word + inline speaker -->
            <div class="text-center mb-3">
              <div class="d-inline-flex align-items-center gap-2">
                <h2 class="fw-bold mb-0" :class="item.card.kind === 'mcq' ? 'fs-5 text-dark text-start' : 'text-primary'">
                  {{ item.card.front }}
                </h2>
                <button v-if="isEnglish" class="btn btn-light rounded-circle speak-btn" type="button"
                  title="Listen to the word and an example (P)" aria-label="Listen" @click="speakCard(item, RATE_REPLAY)">
                  <i class="bi bi-volume-up-fill"></i>
                </button>
              </div>
              <div v-if="soundLocked && autoSpeak && isEnglish" class="small text-warning mt-1">
                <i class="bi bi-volume-mute"></i> Click anywhere or press a key to turn on sound
              </div>
              <div v-if="item.card.payload.ipa || item.card.payload.pos" class="text-muted small mt-1">
                {{ item.card.payload.ipa }}
                <span v-if="item.card.payload.pos" class="ms-1 fst-italic">{{ item.card.payload.pos }}</span>
              </div>
            </div>

            <!-- example sentence in English only (Vietnamese is shown after answering) -->
            <div v-if="phase === 'ask' && exampleParts" class="d-flex justify-content-center align-items-center gap-2 mb-3">
              <p class="fst-italic text-secondary mb-0 example">
                “{{ exampleParts[0] }}<strong class="text-primary">{{ exampleParts[1] }}</strong>{{ exampleParts[2] }}”
              </p>
              <button class="btn btn-light rounded-circle speak-btn speak-sm flex-shrink-0" type="button"
                title="Listen to the sentence (E)" aria-label="Listen to the sentence" @click="speakExample(RATE_REPLAY)">
                <i class="bi bi-volume-up-fill"></i>
              </button>
              <button class="btn btn-light rounded-circle speak-btn speak-sm flex-shrink-0" type="button"
                title="Listen slowly (S)" aria-label="Listen slowly" @click="speakExample(RATE_SLOW)">
                <span aria-hidden="true">🐢</span>
              </button>
            </div>
            <!-- shadowing: read the example aloud, speech-to-text checks each word -->
            <div v-if="phase === 'ask' && exampleText" class="text-center mb-3">
              <SpeakCheck ref="speakCheck" :text="exampleText" />
            </div>

            <!-- ASK: multiple choice -->
            <template v-if="phase === 'ask' && choices.length">
              <p class="text-center text-muted small mb-2">{{ item.card.kind === 'mcq' ? 'Choose the best answer' : 'What does it mean?' }}</p>
              <button v-for="(c, i) in choices" :key="c.id" type="button"
                class="btn btn-outline-primary w-100 text-start mb-2 py-2 choice"
                @click="pick(c.id)">
                <span class="badge bg-light text-dark me-2">{{ i + 1 }}</span>{{ c.text }}
              </button>
              <button type="button" class="btn btn-link text-muted w-100 mt-1" @click="dontKnow">
                I don't remember <small>(0)</small>
              </button>
            </template>

            <!-- ASK: self-check -->
            <div v-else-if="phase === 'ask'" class="text-center">
              <button class="btn btn-primary btn-lg px-5" @click="reveal">
                Show answer <small class="opacity-75 d-none d-md-inline">(Space)</small>
              </button>
            </div>

            <!-- REVEAL -->
            <template v-else>
              <!-- result of the multiple choice -->
              <div v-if="choices.length && outcome" class="mb-3">
                <button v-for="c in choices" :key="c.id" type="button" disabled
                  class="btn w-100 text-start mb-2 py-2 choice" :class="resultClass(c.id)">
                  {{ c.text }}
                </button>
              </div>
              <div v-if="outcome === 'correct'" class="alert alert-success py-2 text-center">
                <i class="bi bi-check-circle-fill"></i> Correct! How easy was it?
              </div>
              <div v-else-if="outcome === 'wrong'" class="alert alert-warning py-2 text-center">
                <i class="bi bi-arrow-counterclockwise"></i> No problem: this card comes back in {{ item.preview['1'] }}.
              </div>

              <CardDetails v-if="item.card.kind !== 'mcq'" :card="item.card" :speakable="isEnglish" class="text-start"
                @speak="(en: string, vi?: string) => vi ? speakBilingual(en, vi, RATE_REPLAY) : speakEnglish(en, RATE_REPLAY)" />
              <div v-else-if="item.card.payload.explanation" class="small text-muted">{{ item.card.payload.explanation }}</div>
              <div v-if="exampleText" class="text-center mt-3 pt-3 border-top">
                <div class="small text-muted mb-2">Practise saying: <em>“{{ exampleText }}”</em></div>
                <SpeakCheck ref="speakCheck" :text="exampleText" />
              </div>

              <!-- grading -->
              <div class="d-flex gap-2 mt-4">
                <template v-if="item.is_new">
                  <button class="btn btn-primary flex-fill py-2" :disabled="sending" @click="rate(3)">
                    <div class="fw-semibold">Got it</div>
                    <small class="opacity-75">practice again in {{ item.preview['3'] }} · Enter</small>
                  </button>
                  <button class="btn btn-outline-secondary flex-fill py-2" :disabled="sending" @click="rate(4)">
                    <div class="fw-semibold">I already know this</div>
                    <small class="opacity-75">next in {{ item.preview['4'] }}</small>
                  </button>
                </template>
                <button v-else-if="outcome === 'wrong'" class="btn btn-warning flex-fill py-2" :disabled="sending" @click="rate(1)">
                  <div class="fw-semibold">Continue</div>
                  <small class="opacity-75">Enter</small>
                </button>
                <template v-else>
                  <button v-for="r in visibleRatings" :key="r.value" class="btn flex-fill py-2"
                    :class="r.value === 3 ? 'btn-success' : r.cls" :disabled="sending" @click="rate(r.value)">
                    <div class="fw-semibold">{{ r.label }}</div>
                    <small class="d-block opacity-75">{{ item.preview[String(r.value)] }} · {{ r.value === 3 ? 'Enter' : r.value }}</small>
                  </button>
                </template>
              </div>
            </template>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="loaded" class="text-center py-5">
      <i class="bi bi-check-circle display-1 text-success"></i>
      <h3 class="mt-3">Nothing to study right now</h3>
      <ForgottenWords :words="forgotten" class="mx-auto my-3" style="max-width: 640px" />

      <!-- why it is empty -->
      <div v-if="empty" class="mx-auto mb-3" style="max-width: 640px">
        <div v-if="empty.new_left_today === 0" class="alert alert-info text-start">
          <div class="fw-semibold mb-1">
            <i class="bi bi-calendar-check"></i>
            You've started all {{ empty.new_limit_today }} new words for today ({{ empty.new_today }}/{{ empty.new_limit_today }}).
          </div>
          <div class="small mb-2">
            The daily limit keeps future reviews manageable. Want more today? It only applies to today;
            the daily amount is in <router-link to="/devices">Device &amp; settings</router-link>.
          </div>
          <button class="btn btn-primary btn-sm me-2" :disabled="addingMore" @click="learnMore(5)">
            <span v-if="addingMore" class="spinner-border spinner-border-sm me-1"></span>
            <i v-else class="bi bi-plus-lg"></i> Learn 5 more today
          </button>
          <button class="btn btn-outline-primary btn-sm" :disabled="addingMore" @click="learnMore(10)">+10</button>
        </div>
        <div v-else class="alert alert-light border text-start small">
          <i class="bi bi-info-circle"></i>
          No new cards left {{ route.query.deck_id ? 'in this deck' : route.query.subject_id ? 'in this subject' : 'in your active decks' }}:
          every card is already in your list. Add another deck or come back when cards are due.
        </div>
        <p v-if="empty.next_due_at" class="text-muted mb-0">
          <i class="bi bi-clock-history"></i> Next review {{ fromNow(empty.next_due_at) }}.
        </p>
      </div>
      <p v-else class="text-muted">Cards you just learned will come back in a few minutes.</p>
      <div class="d-flex flex-wrap gap-2 justify-content-center">
        <button class="btn btn-outline-primary" @click="load"><i class="bi bi-arrow-clockwise"></i> Check again</button>
        <router-link to="/learn" class="btn btn-primary"><i class="bi bi-plus-circle"></i> Add cards</router-link>
        <router-link to="/reading" class="btn btn-warning"><i class="bi bi-book-half"></i> Read a story with your words</router-link>
      </div>
    </div>

    <div v-else class="text-center py-5"><div class="spinner-border"></div></div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, errorMessage } from '../api'
import CardDetails from '../components/CardDetails.vue'
import SpeakCheck from '../components/common/SpeakCheck.vue'
import ForgottenWords from '../components/common/ForgottenWords.vue'
import { RATE_REPLAY, RATE_SLOW, useSpeechSynthesis } from '../composables/useSpeechSynthesis'
import { RATINGS, stateLabel } from '../types'
import type { QueueItem } from '../types'

type Mode = 'choose' | 'self'

const route = useRoute()
const { speakEnglish, speakBilingual } = useSpeechSynthesis()

const queue = ref<QueueItem[]>([])
const idx = ref(0)
const done = ref(0)
const phase = ref<'ask' | 'reveal'>('ask')
const outcome = ref<'correct' | 'wrong' | null>(null)
const picked = ref<number | null>(null)
const sending = ref(false)
const loaded = ref(false)
const error = ref('')
const autoSpeak = ref(storageGet('elearn.autospeak') !== 'false')
const mode = ref<Mode>(storageGet('elearn.reviewmode') === 'self' ? 'self' : 'choose')
let shownAt = Date.now()
// English cards graded "Again" in this session: offered for practice
const forgotten = ref<{ id: number; front: string }[]>([])

const item = computed(() => queue.value[idx.value] ?? null)
const isEnglish = computed(() => item.value?.card.kind === 'vocab' || item.value?.card.kind === 'phrasal')
const progressPct = computed(() => (queue.value.length ? Math.round((idx.value / queue.value.length) * 100) : 0))
const kindLabel = computed(() => ({ vocab: 'Word', phrasal: 'Phrasal verb', concept: 'Concept', mcq: 'Question' }[item.value?.card.kind ?? 'vocab']))
const scopeLabel = computed(() => (route.query.deck_id ? 'This deck' : route.query.subject_id ? 'This subject' : 'All subjects'))

/** Choices for the current card: mcq cards always use their own; others only in "choose" mode. */
const choices = computed<{ id: number; text: string }[]>(() => {
  const it = item.value
  if (!it || it.is_new) return []
  if (it.card.kind === 'mcq') return (it.card.payload.choices ?? []).map((text, id) => ({ id, text }))
  return mode.value === 'choose' ? (it.choices ?? []) : []
})
const correctId = computed(() => (item.value?.card.kind === 'mcq' ? item.value.card.payload.answer ?? -1 : item.value?.card.id ?? -1))
// after a correct pick "Again" makes no sense, so only Hard / Good / Easy are offered
const visibleRatings = computed(() => (outcome.value === 'correct' ? RATINGS.filter((r) => r.value > 1) : RATINGS))

function storageGet(k: string) {
  try { return localStorage.getItem(k) } catch { return null }
}
function storageSet(k: string, v: string) {
  try { localStorage.setItem(k, v) } catch { /* ignore */ }
}
watch(autoSpeak, (v) => storageSet('elearn.autospeak', String(v)))
watch(mode, (v) => storageSet('elearn.reviewmode', v))

watch(item, (it) => {
  outcome.value = null
  picked.value = null
  // new cards are shown with the answer straight away: there is nothing to recall yet
  phase.value = it?.is_new ? 'reveal' : 'ask'
  shownAt = Date.now()
  if (it && autoSpeak.value) {
    if (soundLocked.value) return // the first click/keypress will read it
    speakCard(it)
  }
})

// Browsers block speech until the user has interacted with the page (autoplay
// policy). Until then, wait for the first click/keypress and read the card.
const soundLocked = ref(!hasUserActivation())
function hasUserActivation(): boolean {
  const ua = (navigator as Navigator & { userActivation?: { hasBeenActive: boolean } }).userActivation
  return ua ? ua.hasBeenActive : false
}
function unlockSound() {
  if (!soundLocked.value) return
  soundLocked.value = false
  window.removeEventListener('pointerdown', unlockSound, true)
  window.removeEventListener('keydown', unlockSound, true)
  // read the card that is on screen, unless this click already moves to the next one
  setTimeout(() => { if (item.value && autoSpeak.value && phase.value === 'ask') speakCard(item.value) }, 50)
}

/**
 * Example sentence split around the studied word so it can be highlighted:
 * [before, word, after]. Matches the first word of the front plus common
 * endings ("negotiate" -> "negotiated"), case-insensitive.
 */
const exampleParts = computed<[string, string, string] | null>(() => {
  const it = item.value
  if (!it || (it.card.kind !== 'vocab' && it.card.kind !== 'phrasal')) return null
  const ex = exampleOf(it)
  if (!ex) return null
  const base = it.card.front.trim().split(/\s+/)[0].replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const stem = base.length > 3 && /[ey]$/i.test(base) ? base.slice(0, -1) : base // negotiate→negotiated, carry→carried
  const m = new RegExp(`\\b${stem}\\w*`, 'i').exec(ex)
  if (!m) return [ex, '', '']
  return [ex.slice(0, m.index), m[0], ex.slice(m.index + m[0].length)]
})

/** Example sentence of the current English card ('' if none): used by the 🎤 check. */
const exampleText = computed(() => (item.value && isEnglish.value ? exampleOf(item.value) : ''))
const speakCheck = ref<InstanceType<typeof SpeakCheck> | null>(null)

/** Read only the example sentence (user asked for it: slower) so the user can listen and repeat. */
function speakExample(rate: number) {
  const it = item.value
  if (!it) return
  const ex = exampleOf(it)
  if (ex) speakEnglish(ex, rate).catch(() => {})
}

/** Simplest English example sentence of a word/phrasal card. */
function exampleOf(it: QueueItem): string {
  const senses = [...(it.card.payload.senses ?? [])].sort((a, b) => a.level - b.level)
  return senses.find((s) => s.example_en)?.example_en ?? ''
}

/**
 * Read the word, then one example sentence that uses it (English only).
 * If the user moves to another card meanwhile, the old card stops speaking.
 */
async function speakCard(it: QueueItem, rate = 1) {
  if (it.card.kind !== 'vocab' && it.card.kind !== 'phrasal') return
  const id = it.card.id
  try {
    await speakEnglish(it.card.front, rate)
    const ex = exampleOf(it)
    if (!ex) return
    await new Promise((r) => setTimeout(r, 400))
    if (item.value?.card.id !== id) return
    await speakEnglish(ex, rate)
  } catch {
    /* speech was interrupted by the next card */
  }
}

// stats shown when the queue is empty
interface EmptyInfo { new_left_today: number; new_today: number; new_limit_today: number; next_due_at: string | null }
const empty = ref<EmptyInfo | null>(null)
const addingMore = ref(false)

/** Allow n more new cards today only, then reload the queue. */
async function learnMore(n: number) {
  addingMore.value = true
  try {
    await api.post('/study/extra-new', { count: n })
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    addingMore.value = false
  }
}

/** "in 7 min", "in 3 h", "tomorrow at 08:15" */
function fromNow(iso: string): string {
  const t = new Date(iso)
  const min = Math.max(1, Math.round((t.getTime() - Date.now()) / 60000))
  const hhmm = t.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  if (min < 60) return `in ${min} min (${hhmm})`
  if (min < 12 * 60) return `in ${Math.round(min / 60)} h (${hhmm})`
  const days = Math.round((startOfDay(t) - startOfDay(new Date())) / 86400000)
  return days === 1 ? `tomorrow at ${hhmm}` : `on ${t.toLocaleDateString()} at ${hhmm}`
}
function startOfDay(d: Date): number {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime()
}

async function load() {
  loaded.value = false
  error.value = ''
  try {
    const params = { deck_id: route.query.deck_id, subject_id: route.query.subject_id, limit: 20 }
    const { data } = await api.get('/study/queue', { params })
    queue.value = data
    idx.value = 0
    empty.value = null
    if (!data.length) {
      // explain why there is nothing (daily limit reached? next card due when?)
      try {
        empty.value = (await api.get('/stats')).data
      } catch { /* the generic message is shown */ }
    }
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loaded.value = true
  }
}

function reveal() {
  phase.value = 'reveal'
}

function pick(id: number) {
  picked.value = id
  outcome.value = id === correctId.value ? 'correct' : 'wrong'
  phase.value = 'reveal'
}

function dontKnow() {
  picked.value = null
  outcome.value = 'wrong'
  phase.value = 'reveal'
}

function resultClass(id: number) {
  if (id === correctId.value) return 'btn-success'
  if (id === picked.value) return 'btn-danger'
  return 'btn-outline-secondary opacity-50'
}

async function rate(rating: number) {
  if (!item.value || sending.value) return
  sending.value = true
  try {
    const it = item.value
    await api.post('/study/review', { card_id: it.card.id, rating, duration_ms: Date.now() - shownAt })
    if (rating === 1 && isEnglish.value && !forgotten.value.some((w) => w.id === it.card.id)) {
      forgotten.value.push({ id: it.card.id, front: it.card.front })
    }
    done.value++
    if (idx.value < queue.value.length - 1) idx.value++
    else await load() // learning steps come back within minutes
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    sending.value = false
  }
}

function onKey(e: KeyboardEvent) {
  const it = item.value
  if (!it || sending.value || (e.target as HTMLElement)?.tagName === 'INPUT') return
  if (e.key === 'p' || e.key === 'P') {
    speakCard(it, RATE_REPLAY)
    return
  }
  if (e.key === 'e' || e.key === 'E') {
    speakExample(RATE_REPLAY)
    return
  }
  if (e.key === 's' || e.key === 'S') {
    speakExample(RATE_SLOW)
    return
  }
  if ((e.key === 'r' || e.key === 'R') && speakCheck.value) {
    speakCheck.value.toggle()
    return
  }
  if (phase.value === 'ask') {
    if (choices.value.length) {
      const n = Number(e.key)
      if (n >= 1 && n <= choices.value.length) pick(choices.value[n - 1].id)
      else if (e.key === '0') dontKnow()
    } else if (e.code === 'Space') {
      e.preventDefault()
      reveal()
    }
    return
  }
  // reveal phase
  if (e.key === 'Enter') {
    e.preventDefault()
    rate(outcome.value === 'wrong' ? 1 : 3)
  } else if (!it.is_new && outcome.value !== 'wrong' && ['1', '2', '3', '4'].includes(e.key)) {
    const r = Number(e.key)
    if (outcome.value === 'correct' && r === 1) return
    rate(r)
  }
}

watch(() => route.fullPath, load)
onMounted(() => {
  load()
  window.addEventListener('keydown', onKey)
  if (soundLocked.value) {
    window.addEventListener('pointerdown', unlockSound, true)
    window.addEventListener('keydown', unlockSound, true)
  }
})
onUnmounted(() => {
  window.removeEventListener('keydown', onKey)
  window.removeEventListener('pointerdown', unlockSound, true)
  window.removeEventListener('keydown', unlockSound, true)
})
</script>

<style scoped>
.speak-btn {
  width: 40px;
  height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #0d6efd;
  border: 1px solid #dee2e6;
}
.speak-btn:hover {
  background: #e7f1ff;
}
.choice {
  white-space: normal;
}
.example {
  font-size: 1.05rem;
}
.speak-sm {
  width: 32px;
  height: 32px;
  font-size: 0.85rem;
}
</style>
