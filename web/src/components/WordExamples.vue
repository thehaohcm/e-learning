<template>
  <div ref="root" class="word-examples" tabindex="-1" @keydown="onKeydown">
    <div v-if="loading" class="small text-muted py-2"><span class="spinner-border spinner-border-sm me-1"></span> Loading examples…</div>

    <template v-else>
      <div v-if="examples.length" class="d-flex flex-wrap justify-content-end gap-3 mb-1">
        <div class="form-check form-switch small mb-0" title="Type the missing word, then say the sentence">
          <input :id="`blank-${cardId}`" v-model="blanks" class="form-check-input" type="checkbox" />
          <label class="form-check-label" :for="`blank-${cardId}`">Fill in the blanks</label>
        </div>
        <div class="form-check form-switch small mb-0">
          <input :id="`vi-${cardId}`" v-model="showVi" class="form-check-input" type="checkbox" />
          <label class="form-check-label" :for="`vi-${cardId}`">Show Vietnamese</label>
        </div>
      </div>

      <div v-for="ex in examples" :key="ex.id" :data-id="ex.id" class="example py-2 border-bottom"
        :class="{ current: current === ex.id }">
        <div class="d-flex align-items-start gap-2">
          <span v-if="ex.context" class="badge bg-light text-secondary border flex-shrink-0 mt-1">{{ ex.context }}</span>
          <div class="flex-grow-1">
            <!-- blank: type the word (in the form used in the sentence) -->
            <div v-if="isBlank(ex)" class="sentence">
              <template v-for="(seg, k) in segments(ex.en)" :key="k">
                <template v-if="seg.hit">
                  <input v-if="seg.first" :ref="inputRef(ex.id)" v-model="st(ex).answer" class="blank form-control form-control-sm d-inline-block"
                    :class="{ 'is-invalid': st(ex).wrong }" :style="{ width: Math.max(5, seg.text.length + 2) + 'ch' }"
                    autocomplete="off" autocapitalize="off" spellcheck="false" :aria-label="`missing word in example`"
                    @focus="current = ex.id" @keydown.enter.prevent="checkBlank(ex, seg.text, $event)" />
                  <span v-else class="text-muted">_____</span>
                </template>
                <template v-else>{{ seg.text }}</template>
              </template>
              <div class="small mt-1">
                <button class="btn btn-link btn-sm p-0 me-2" :disabled="!st(ex).answer.trim()" @click="checkBlank(ex, firstHit(ex.en))">Check (Enter)</button>
                <span v-if="st(ex).wrong" class="text-danger me-2">Not this one: listen 🔊 and try again.</span>
                <button class="btn btn-link btn-sm p-0 text-muted" @click="reveal(ex)">Show</button>
              </div>
            </div>
            <!-- full sentence -->
            <div v-else role="button" title="Tap to show the Vietnamese" @click="toggle(ex.id)">
              <template v-for="(seg, k) in segments(ex.en)" :key="k">
                <strong v-if="seg.hit" class="text-primary">{{ seg.text }}</strong><template v-else>{{ seg.text }}</template>
              </template>
              <span v-if="st(ex).note" class="small ms-1" :class="st(ex).revealed ? 'text-muted' : 'text-success'">{{ st(ex).note }}</span>
              <div v-if="current === ex.id && blanks" class="small text-muted mt-1 keys">
                <kbd v-if="micOK">S</kbd><span v-if="micOK"> say it · </span><kbd>N</kbd> next · <kbd>P</kbd> previous
              </div>
            </div>
            <div v-if="showVi || opened.has(ex.id)" class="small text-muted fst-italic mt-1">{{ ex.vi }}</div>
          </div>
          <div class="d-flex gap-1 flex-shrink-0">
            <button class="btn btn-light btn-sm rounded-circle" title="Listen" @click="say(ex.en, RATE_REPLAY)">
              <i class="bi bi-volume-up-fill"></i>
            </button>
            <button class="btn btn-light btn-sm rounded-circle" title="Listen slowly" @click="say(ex.en, RATE_SLOW)">🐢</button>
            <button v-if="micOK && !isBlank(ex)" class="btn btn-sm rounded-circle" :class="practising === ex.id ? 'btn-danger' : 'btn-light'"
              title="Practise saying it" @click="practising = practising === ex.id ? 0 : ex.id">
              <i class="bi bi-mic-fill"></i>
            </button>
          </div>
        </div>
        <SpeakCheck v-if="practising === ex.id" :ref="setSpeak" :text="ex.en" :model-first="lastRead !== ex.id" label="Say it"
          class="mt-2 text-center" @scored="scored" />
      </div>

      <div v-if="!examples.length && !canGenerate" class="small text-muted py-2">No extra examples yet.</div>
      <div v-if="warning" class="small text-warning mt-2">{{ warning }}</div>
      <button v-if="canGenerate" class="btn btn-outline-primary btn-sm mt-2" :disabled="generating" @click="generate">
        <span v-if="generating" class="spinner-border spinner-border-sm me-1"></span>
        <i v-else class="bi bi-stars"></i>
        {{ generating ? 'AI is writing examples… (10–30 s)' : examples.length ? 'More examples with AI' : 'Get examples with AI' }}
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, reactive, ref, watch } from 'vue'
import { api, errorMessage } from '../api'
import SpeakCheck from './common/SpeakCheck.vue'
import { speechRecognitionSupported } from '../composables/useSpeechRecognition'
import { RATE_NORMAL, RATE_REPLAY, RATE_SLOW, useSpeechSynthesis } from '../composables/useSpeechSynthesis'
import { wordRegex } from '../utils/words'
import { useBlankNav } from '../composables/useBlankNav'

interface Example { id: number; en: string; vi: string; context?: string }

// excludeId: the example already shown as the card's main example (not listed twice)
const props = defineProps<{ cardId: number; word: string; excludeId?: number }>()
const { speakEnglish } = useSpeechSynthesis()
const micOK = speechRecognitionSupported

const examples = ref<Example[]>([])
const canGenerate = ref(false)
const loading = ref(true)
const generating = ref(false)
const warning = ref('')
const practising = ref(0) // example id with the 🎤 panel open
const opened = reactive(new Set<number>()) // examples whose Vietnamese was tapped open
const showVi = ref(storageGet('elearn.examples.vi') === 'true')
watch(showVi, (v) => storageSet('elearn.examples.vi', String(v)))
const blanks = ref(storageGet('elearn.examples.blanks') !== 'false') // on by default
watch(blanks, (v) => storageSet('elearn.examples.blanks', String(v)))

// keyboard: Enter checks a blank, S says the sentence, N / P next / previous
// sentence just filled (and read aloud by the app): S listens straight away;
// for other sentences the 🎤 panel reads the sentence first
const lastRead = ref(0)
let speakRef: { toggle: () => void } | null = null
const setSpeak = (el: unknown) => { speakRef = (el as { toggle: () => void } | null) ?? null }
const { current, root, inputRef, solved, wrong, scored, onKeydown } = useBlankNav({
  ids: () => examples.value.map((e) => e.id),
  isOpen: (id) => {
    const ex = examples.value.find((e) => e.id === id)
    return !!ex && isBlank(ex)
  },
  speak: async (id) => {
    if (!micOK) return
    practising.value = id
    await nextTick()
    speakRef?.toggle()
  },
})

// fill-in-the-blank state per example
interface BlankState { answer: string; solved: boolean; revealed: boolean; wrong: boolean; note: string }
const states = reactive(new Map<number, BlankState>())
function st(ex: Example): BlankState {
  let s = states.get(ex.id)
  if (!s) {
    s = { answer: '', solved: false, revealed: false, wrong: false, note: '' }
    states.set(ex.id, s)
  }
  return s
}
const isBlank = (ex: Example) => blanks.value && !st(ex).solved && segments(ex.en).some((s) => s.hit)
const firstHit = (text: string) => segments(text).find((s) => s.hit)?.text ?? ''

const norm = (s: string) => s.toLowerCase().replace(/[^a-z' ]/g, ' ').replace(/\s+/g, ' ').trim()

/**
 * Accept the exact form used in the sentence; the base word counts too, with a
 * note showing the right form (negotiate → here "negotiated"). Spelling must be right.
 */
function checkBlank(ex: Example, expected: string, e?: Event) {
  const s = st(ex)
  const typed = norm(s.answer)
  if (!typed) return
  if (typed === norm(expected)) {
    solve(ex, '')
  } else if (typed === norm(props.word)) {
    solve(ex, `✓ (here: “${expected}”)`)
  } else {
    s.wrong = true
    if (e) wrong(e) // select the wrong answer: typing replaces it
  }
}

function solve(ex: Example, note: string) {
  const s = st(ex)
  s.solved = true
  s.wrong = false
  s.note = note
  say(ex.en, RATE_NORMAL) // read the English sentence…
  opened.add(ex.id) // …and show its Vietnamese meaning (not read)
  lastRead.value = ex.id
  practising.value = 0
  solved(ex.id) // keyboard stays here: S to say it, N for the next blank
}

function reveal(ex: Example) {
  const s = st(ex)
  s.solved = true
  s.revealed = true
  s.wrong = false
  s.note = '(shown)'
  say(ex.en, RATE_REPLAY)
  opened.add(ex.id)
  lastRead.value = ex.id
  solved(ex.id)
}

function storageGet(k: string) {
  try { return localStorage.getItem(k) } catch { return null }
}
function storageSet(k: string, v: string) {
  try { localStorage.setItem(k, v) } catch { /* ignore */ }
}

function say(text: string, rate: number) {
  speakEnglish(text, rate).catch(() => {})
}

function toggle(id: number) {
  if (opened.has(id)) opened.delete(id)
  else opened.add(id)
}

/** Sentence split around the word (any form) so it can be highlighted / blanked. */
function segments(text: string): { text: string; hit: boolean; first?: boolean }[] {
  const out: { text: string; hit: boolean; first?: boolean }[] = []
  let last = 0
  for (const m of text.matchAll(wordRegex(props.word, 'gi'))) {
    const i = m.index ?? 0
    if (i > last) out.push({ text: text.slice(last, i), hit: false })
    out.push({ text: m[0], hit: true, first: !out.some((s) => s.hit) })
    last = i + m[0].length
  }
  if (last < text.length) out.push({ text: text.slice(last), hit: false })
  return out
}

function apply(data: { examples?: Example[]; can_generate?: boolean; warning?: string }) {
  examples.value = (data.examples ?? []).filter((e) => e.id !== props.excludeId)
  canGenerate.value = !!data.can_generate
  warning.value = data.warning ?? ''
}

async function load() {
  loading.value = true
  practising.value = 0
  opened.clear()
  states.clear()
  try {
    apply((await api.get(`/cards/${props.cardId}/examples`)).data)
  } catch (e) {
    warning.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

/** One AI request fills the word up to 6 examples (kept for everyone). */
async function generate() {
  generating.value = true
  warning.value = ''
  try {
    apply((await api.post(`/cards/${props.cardId}/examples`)).data)
  } catch (e) {
    warning.value = errorMessage(e)
  } finally {
    generating.value = false
  }
}

watch(() => props.cardId, load)
onMounted(load)
</script>

<style scoped>
.example:last-of-type { border-bottom: 0 !important; }
.sentence { line-height: 2; }
.blank { padding: 0 0.3rem; height: 1.7rem; vertical-align: baseline; }
.word-examples:focus { outline: none; }
.example.current { box-shadow: inset 3px 0 0 var(--bs-primary); padding-left: 0.5rem; }
.keys kbd { font-size: 0.7rem; padding: 0.05rem 0.3rem; }
</style>
