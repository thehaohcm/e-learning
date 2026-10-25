<template>
  <div ref="root" class="card border-0 shadow-sm text-start cloze" tabindex="-1" @keydown="onKey">
    <div class="card-body p-4">
      <div class="d-flex flex-wrap align-items-center gap-2 mb-2">
        <h5 class="fw-bold mb-0"><i class="bi bi-input-cursor-text text-primary"></i> Round check: fill in the blanks</h5>
        <div class="btn-group btn-group-sm ms-auto" role="group" aria-label="Answer by">
          <button v-for="m in MODES" :key="m.value" type="button" class="btn" :class="mode === m.value ? 'btn-secondary' : 'btn-outline-secondary'"
            :title="m.title" @click="mode = m.value">
            <i class="bi" :class="m.icon"></i> {{ m.label }}
          </button>
        </div>
        <button v-if="phase === 'intro' || phase === 'error'" class="btn btn-link btn-sm ms-auto text-muted" @click="emit('done')">Skip</button>
      </div>

      <!-- before starting: no AI call until the learner wants it -->
      <template v-if="phase === 'intro'">
        <p class="text-muted small mb-3">
          {{ words.length }} word{{ words.length > 1 ? 's' : '' }} from this round in new sentences:
          {{ mode === 'type' ? 'type the missing word.' : mode === 'say' ? 'read each sentence aloud, saying the missing word.' : 'pick the right word or the right form (TOEIC Part 5).' }}
        </p>
        <button class="btn btn-primary" @click="load"><i class="bi bi-play-fill"></i> Start</button>
      </template>

      <div v-else-if="phase === 'loading'" class="text-muted small py-3">
        <span class="spinner-border spinner-border-sm me-2"></span> Preparing sentences… (AI may need up to ~30 s for new words)
      </div>

      <div v-else-if="phase === 'error'" class="alert alert-warning py-2 mb-0">
        {{ error }} <button class="btn btn-sm btn-outline-secondary ms-2" @click="load">Try again</button>
      </div>

      <template v-else>
        <div v-if="warning" class="alert alert-warning py-2 small">{{ warning }}</div>
        <div v-if="mode === 'choose' && formsLoading" class="small text-muted mb-2">
          <span class="spinner-border spinner-border-sm me-1"></span> Preparing word-form options (TOEIC Part 5)…
        </div>

        <!-- word bank (not needed when the options are shown) -->
        <div v-if="mode !== 'choose'" class="mb-3">
          <span class="small text-muted me-1">Words:</span>
          <span v-for="w in bank" :key="w" class="badge me-1 mb-1 bank-word" :class="usedWords.has(w) ? 'bg-success-subtle text-success text-decoration-line-through' : 'bg-light text-dark border'">
            {{ w }}
          </span>
        </div>

        <ol class="ps-3 mb-0">
          <li v-for="it in items" :key="it.example_id" :data-id="it.example_id" class="mb-3 item"
            :class="{ current: current === it.example_id }">
            <div class="sentence">
              <template v-for="(seg, k) in splitByWord(it.reference_en, it.word)" :key="k">
                <template v-if="seg.hit && !it.done">
                  <span v-if="seg.first && mode !== 'type'" class="blank-mc">_____</span>
                  <input v-else-if="seg.first" :ref="inputRef(it.example_id)" v-model="it.answer" class="blank form-control form-control-sm d-inline-block"
                    :class="{ 'is-invalid': it.wrong > 0 && !it.done }" :style="{ width: Math.max(6, seg.text.length + 2) + 'ch' }"
                    autocomplete="off" autocapitalize="off" spellcheck="false" aria-label="missing word"
                    @focus="current = it.example_id" @keydown.enter.prevent="check(it, seg.text, $event)" />
                  <span v-else class="text-muted">_____</span>
                </template>
                <strong v-else-if="seg.hit" :class="it.shown ? 'text-secondary' : 'text-success'">{{ seg.text }}</strong>
                <template v-else>{{ seg.text }}</template>
              </template>
            </div>

            <div class="small mt-1 d-flex flex-wrap align-items-center gap-2">
              <template v-if="!it.done && mode === 'choose'">
                <span class="q-kind" :title="it.qkind === 'form' ? 'Which form of the word fits the grammar? (TOEIC Part 5)' : 'Which word fits the meaning?'">
                  {{ it.qkind === 'form' ? 'Which form?' : 'Which word?' }}
                </span>
                <button v-for="(opt, k) in it.options" :key="opt" type="button" class="btn btn-sm opt"
                  :class="it.tried.includes(opt) ? 'btn-outline-danger text-decoration-line-through' : 'btn-outline-primary'"
                  :disabled="it.tried.includes(opt)" @click="pick(it, opt)">
                  <span class="opt-key">{{ 'ABCD'[k] }}</span> {{ opt }}
                </button>
                <button type="button" class="btn btn-sm btn-outline-secondary opt idk" title="Show the answer (0)" @click="giveUp(it)">
                  <span class="opt-key">0</span> ? I don't know
                </button>
                <span v-if="it.wrong" class="text-danger">Not this one, try again.</span>
              </template>
              <template v-else-if="!it.done && mode === 'say'">
                <button type="button" class="btn btn-sm" :class="sayingId === it.example_id ? 'btn-danger' : 'btn-outline-danger'"
                  :disabled="sayingId !== 0 && sayingId !== it.example_id" @click="sayItem(it)">
                  <i :class="sayingId === it.example_id ? 'bi bi-stop-circle-fill' : 'bi bi-mic-fill'"></i>
                  {{ sayingId === it.example_id ? 'Listening… tap to stop' : 'Say the whole sentence (S)' }}
                </button>
                <span v-if="sayingId === it.example_id && recTranscript" class="fst-italic text-muted">{{ recTranscript }}…</span>
                <span v-else-if="it.wrong" class="text-danger">Heard “{{ it.heard || '…' }}”: not the missing word, try again.</span>
                <span v-if="sayingId === it.example_id && recError" class="text-danger">{{ recError }}</span>
              </template>
              <template v-else-if="!it.done">
                <button class="btn btn-link btn-sm p-0" :disabled="!it.answer.trim()" @click="check(it, firstHit(it))">Check (Enter)</button>
                <span v-if="it.wrong" class="text-danger">Not this one{{ it.wrong >= 2 ? '' : ', try again' }}.</span>
              </template>
              <template v-else>
                <span v-if="it.shown" class="text-muted"><i class="bi bi-eye"></i> answer: {{ it.qkind === 'form' ? it.answerForm : it.word }}</span>
                <span v-else class="text-success"><i class="bi bi-check-circle-fill"></i>{{ it.note }}</span>
                <button class="btn btn-light btn-sm rounded-circle" title="Listen" @click="say(it.reference_en)"><i class="bi bi-volume-up-fill"></i></button>
                <button v-if="micOK" class="btn btn-sm rounded-circle" :class="practising === it.example_id ? 'btn-danger' : 'btn-light'"
                  title="Say it (S)" @click="practising = practising === it.example_id ? 0 : it.example_id"><i class="bi bi-mic-fill"></i></button>
                <span v-if="current === it.example_id" class="text-muted keys">
                  <kbd v-if="micOK">S</kbd><span v-if="micOK"> say it · </span><kbd>N</kbd> next · <kbd>P</kbd> previous
                </span>
              </template>
              <button class="btn btn-link btn-sm p-0 text-muted" @click="it.vi = !it.vi">{{ it.vi ? 'Hide' : 'Vietnamese' }}</button>
            </div>
            <div v-if="it.vi" class="small text-muted fst-italic">{{ it.prompt_vi }}</div>
            <SpeakCheck v-if="practising === it.example_id" :ref="setSpeak" :text="it.reference_en" :model-first="!it.done"
              label="Say it" class="mt-2" @scored="scored" />
          </li>
        </ol>

        <div class="d-flex flex-wrap align-items-center gap-2 mt-2 pt-3 border-top">
          <span class="fw-semibold">{{ correct }} / {{ items.length }} correct</span>
          <span v-if="remaining" class="small text-muted">{{ remaining }} left</span>
          <button class="btn ms-auto" :class="remaining ? 'btn-outline-secondary' : 'btn-success'" @click="emit('done')">
            {{ remaining ? 'Finish' : 'Done' }} <i class="bi bi-arrow-right"></i>
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { api, errorMessage } from '../../api'
import SpeakCheck from './SpeakCheck.vue'
import { speechRecognitionSupported, useSpeechRecognition } from '../../composables/useSpeechRecognition'
import { RATE_NORMAL, RATE_REPLAY, useSpeechSynthesis } from '../../composables/useSpeechSynthesis'
import { checkTypedWord, splitByWord, wordRegex } from '../../utils/words'
import { useBlankNav } from '../../composables/useBlankNav'

/**
 * End-of-round exercise (Study, Quiz): one new sentence per word of the round,
 * the word is a blank. Three ways to answer:
 *   Type    — type the word (exact form or base word accepted), word bank shown;
 *   Say     — read the whole sentence aloud, saying the missing word (recall + speaking);
 *   A B C D — "Which word?" (4 words of the same part of speech) or "Which form?"
 *             (4 forms of the same word family, like TOEIC Part 5).
 * 2 wrong tries show the answer.
 */
interface Item {
  card_id: number
  word: string
  example_id: number
  prompt_vi: string
  reference_en: string
  answer: string
  done: boolean
  shown: boolean // answer revealed after 2 wrong tries
  wrong: number
  note: string
  vi: boolean
  options: string[] // A B C D: the 4 options shown
  tried: string[] // wrong options already picked
  qkind: 'word' | 'form' // A B C D question type
  answerForm: string // the form in the sentence ("fluctuated")
  word_options: string[] // 3 other words, same part of speech when possible
  form_options: string[] | null // 3 other forms of the word family; null = not prepared yet
  heard: string // Say: what was recognised last
}

type Mode = 'type' | 'say' | 'choose'
const MODES = [
  { value: 'type' as Mode, label: 'Type', icon: 'bi-keyboard', title: 'Type the missing word' },
  ...(speechRecognitionSupported ? [{ value: 'say' as Mode, label: 'Say', icon: 'bi-mic', title: 'Read the whole sentence aloud, saying the missing word' }] : []),
  { value: 'choose' as Mode, label: 'A B C D', icon: 'bi-ui-checks', title: 'Pick the word or the right form out of 4 (keys 1–4)' },
]

const props = defineProps<{ words: { id: number; front: string }[] }>()
const emit = defineEmits<{ (e: 'done'): void }>()
const { speakEnglish } = useSpeechSynthesis()
const micOK = speechRecognitionSupported

const phase = ref<'intro' | 'loading' | 'error' | 'play'>('intro')
const items = ref<Item[]>([])
const error = ref('')
const warning = ref('')
const practising = ref(0)
// how to answer (remembered on this browser); phones start with A B C D
const savedMode = storageGet('elearn.cloze.mode') as Mode | null
const isPhone = typeof window !== 'undefined' && window.matchMedia?.('(pointer: coarse)').matches
const mode = ref<Mode>(savedMode && MODES.some((m) => m.value === savedMode) ? savedMode : isPhone ? 'choose' : 'type')
watch(mode, (v) => {
  storageSet('elearn.cloze.mode', v)
  recStop()
  if (v === 'choose') prepareChoices()
  const first = items.value.find((i) => !i.done)
  if (first) focusItem(first.example_id)
})
function storageGet(k: string) {
  try { return localStorage.getItem(k) } catch { return null }
}
function storageSet(k: string, v: string) {
  try { localStorage.setItem(k, v) } catch { /* private mode */ }
}

const bank = computed(() => [...items.value.map((i) => i.word)].sort((a, b) => a.localeCompare(b)))
const usedWords = computed(() => new Set(items.value.filter((i) => i.done && !i.shown).map((i) => i.word)))
const correct = computed(() => items.value.filter((i) => i.done && !i.shown).length)
const remaining = computed(() => items.value.filter((i) => !i.done).length)

function say(text: string, rate = RATE_REPLAY) {
  speakEnglish(text, rate).catch(() => {})
}

// keyboard: Enter checks, S says the filled sentence (≥ 80% → next blank), N / P next / previous
let speakRef: { toggle: () => void } | null = null
const setSpeak = (el: unknown) => { speakRef = (el as { toggle: () => void } | null) ?? null }
const { current, root, inputRef, focusItem, solved, wrong, scored, onKeydown } = useBlankNav({
  ids: () => items.value.map((i) => i.example_id),
  isOpen: (id) => !items.value.find((i) => i.example_id === id)?.done,
  speak: async (id) => {
    if (!micOK) return
    practising.value = id
    await nextTick()
    speakRef?.toggle()
  },
})

const firstHit = (it: Item) => splitByWord(it.reference_en, it.word).find((s) => s.hit)?.text ?? it.word

/** A B C D mode: the right option fills the blank; 2 wrong picks show the answer. */
function pick(it: Item, opt: string) {
  if (it.done || it.tried.includes(opt)) return
  current.value = it.example_id
  const right = it.qkind === 'form' ? it.answerForm : it.word
  if (opt.toLowerCase() !== right.toLowerCase()) {
    it.tried.push(opt)
    it.wrong++
    if (it.wrong < 2) return
    it.shown = true
  }
  finish(it, '')
}

/** "I don't know": the answer is shown (not counted as correct). */
function giveUp(it: Item) {
  if (it.done) return
  current.value = it.example_id
  it.shown = true
  finish(it, '')
}

/** A blank is answered (or shown): read the English sentence, show its Vietnamese (not read). */
function finish(it: Item, note: string) {
  it.done = true
  it.vi = true
  it.note = note
  say(it.reference_en, it.shown ? RATE_REPLAY : RATE_NORMAL)
  practising.value = 0
  solved(it.example_id)
  const nextOpen = items.value.find((i) => !i.done)
  if (nextOpen) current.value = nextOpen.example_id
}

// ---------------------------------------------------------------- Say
// The learner reads the whole sentence, saying the missing word. Recognition
// of a full sentence is much more reliable than of a single word; only the
// missing word is checked (any form of it). Practice only: no schedule change.
const { transcript: recTranscript, error: recError, listen: recListen, stop: recStop } = useSpeechRecognition()
const sayingId = ref(0)

async function sayItem(it: Item) {
  if (sayingId.value === it.example_id) {
    recStop()
    return
  }
  if (it.done || sayingId.value) return
  current.value = it.example_id
  sayingId.value = it.example_id
  const alts = await recListen()
  sayingId.value = 0
  if (!alts.length || it.done) return // nothing heard: the error says why
  const re = wordRegex(it.word)
  if (alts.some((a) => re.test(a))) {
    it.heard = alts[0]
    finish(it, '')
    return
  }
  it.heard = alts[0]
  it.wrong++
  if (it.wrong >= 2) {
    it.shown = true
    finish(it, '')
  }
}

// ---------------------------------------------------------------- A B C D
const formsLoading = ref(false)
let formsAsked = false

/** Builds the 4 options of each sentence; half "Which form?" when the word family is known. */
function buildOptions() {
  const bankWords = [...new Set(items.value.map((i) => i.word))]
  for (const it of items.value) {
    if (it.done || it.tried.length) continue
    const forms = it.form_options ?? []
    if (forms.length === 3 && it.answerForm && Math.random() < 0.5) {
      it.qkind = 'form'
      it.options = shuffle([it.answerForm, ...forms])
    } else {
      it.qkind = 'word'
      let others = it.word_options.filter((w) => w.toLowerCase() !== it.word.toLowerCase())
      if (others.length < 3) {
        const extra = shuffle(bankWords.filter((w) => w.toLowerCase() !== it.word.toLowerCase() && !others.includes(w)))
        others = [...others, ...extra].slice(0, 3)
      }
      it.options = shuffle([it.word, ...others.slice(0, 3)])
    }
  }
}

/** Word-form options are prepared by AI only when A B C D is used (one request for the round, kept). */
async function prepareChoices() {
  if (phase.value !== 'play') return
  const need = items.value.filter((i) => i.form_options === null && !i.done).map((i) => i.example_id)
  if (need.length && !formsAsked) {
    formsAsked = true
    formsLoading.value = true
    try {
      const { data } = await api.post('/practice/cloze/forms', { example_ids: need })
      for (const it of items.value) {
        const f = data.forms?.[it.example_id]
        if (f) {
          it.form_options = f.options ?? []
          if (f.answer) it.answerForm = f.answer
        }
      }
    } catch {
      // keep "Which word?" questions only
    } finally {
      formsLoading.value = false
    }
  }
  buildOptions()
}

/** Keys: 1–4 or A–D pick an option of the current sentence (A B C D mode), the rest as usual. */
function onKey(e: KeyboardEvent) {
  const tag = (e.target as HTMLElement)?.tagName
  if (mode.value === 'say' && tag !== 'INPUT' && tag !== 'TEXTAREA' && e.key.toLowerCase() === 's' && !e.ctrlKey && !e.metaKey) {
    const it = items.value.find((i) => i.example_id === current.value && !i.done) ?? items.value.find((i) => !i.done)
    if (it) {
      e.preventDefault()
      e.stopPropagation()
      sayItem(it)
      return
    }
  }
  if (mode.value === 'choose' && tag !== 'INPUT' && tag !== 'TEXTAREA' && !e.ctrlKey && !e.metaKey && !e.altKey) {
    const k = e.key.toLowerCase()
    const n = /^[1-4]$/.test(k) ? Number(k) - 1 : 'abcd'.indexOf(k.length === 1 ? k : '#')
    const it = items.value.find((i) => i.example_id === current.value) ?? items.value.find((i) => !i.done)
    if (n >= 0 && it && !it.done && it.options[n]) {
      e.preventDefault()
      e.stopPropagation()
      pick(it, it.options[n])
      return
    }
    if (k === '0' && it && !it.done) {
      e.preventDefault()
      e.stopPropagation()
      giveUp(it)
      return
    }
  }
  onKeydown(e)
}

function shuffle<T>(a: T[]): T[] {
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[a[i], a[j]] = [a[j], a[i]]
  }
  return a
}

/** Enter in a blank. Filled (or shown after 2 misses): the English sentence is read and its Vietnamese shown. */
function check(it: Item, expected: string, e?: Event) {
  if (it.done || !it.answer.trim()) return
  const r = checkTypedWord(it.answer, expected, it.word)
  if (r === 'wrong') {
    it.wrong++
    if (it.wrong < 2) {
      if (e) wrong(e) // select it: typing replaces it
      return
    }
    it.shown = true
  }
  it.done = true
  it.vi = true // Vietnamese shown, not read
  say(it.reference_en, it.shown ? RATE_REPLAY : RATE_NORMAL)
  it.note = r === 'base' ? ` right word (here: “${expected}”)` : ''
  practising.value = 0
  solved(it.example_id)
}

async function load() {
  phase.value = 'loading'
  error.value = ''
  try {
    const { data } = await api.post('/practice/cloze', { card_ids: props.words.map((w) => w.id).slice(0, 10) })
    type Raw = Pick<Item, 'card_id' | 'word' | 'example_id' | 'prompt_vi' | 'reference_en'> & {
      answer?: string; word_options?: string[]; form_options?: string[] | null
    }
    items.value = (data.items ?? []).map((p: Raw) => {
      const it: Item = {
        card_id: p.card_id, word: p.word, example_id: p.example_id, prompt_vi: p.prompt_vi, reference_en: p.reference_en,
        answer: '', done: false, shown: false, wrong: 0, note: '', vi: false, options: [], tried: [], qkind: 'word',
        answerForm: p.answer || '', word_options: p.word_options ?? [], form_options: p.form_options ?? null, heard: '',
      }
      if (!it.answerForm) it.answerForm = firstHit(it)
      return it
    })
    formsAsked = false
    buildOptions()
    warning.value = data.warning ?? ''
    if (!items.value.length) {
      phase.value = 'error'
      error.value = 'No sentences could be prepared for these words.'
      return
    }
    phase.value = 'play'
    if (mode.value === 'choose') prepareChoices()
    await nextTick()
    if (items.value[0]) focusItem(items.value[0].example_id)
  } catch (e) {
    phase.value = 'error'
    error.value = errorMessage(e)
  }
}
</script>

<style scoped>
.idk { border-style: dashed; }
.sentence { line-height: 2.1; }
.blank { padding: 0 0.3rem; height: 1.7rem; vertical-align: baseline; }
.blank-mc { font-weight: 600; color: var(--bs-primary); letter-spacing: 0.05em; }
.opt { padding: 0.1rem 0.55rem; }
.opt-key { font-size: 0.7rem; font-weight: 700; opacity: 0.6; margin-right: 0.15rem; }
.q-kind { font-size: 0.75rem; font-weight: 600; color: var(--bs-secondary-color); text-transform: uppercase; letter-spacing: 0.03em; }
.bank-word { font-size: 0.85rem; font-weight: 500; }
.cloze:focus { outline: none; }
.item.current { box-shadow: -10px 0 0 -7px var(--bs-primary); }
.keys kbd { font-size: 0.7rem; padding: 0.05rem 0.3rem; }
</style>
