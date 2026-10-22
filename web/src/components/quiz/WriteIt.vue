<template>
  <div class="row justify-content-center">
    <div class="col-lg-8">
      <div v-if="warning" class="alert alert-warning py-2 small">{{ warning }}</div>

      <div v-if="phase === 'loading'" class="text-center py-5">
        <div class="spinner-border text-success"></div>
        <p class="mt-3 text-muted small">Preparing sentences… (AI may need up to ~40 s for words without sentences yet)</p>
      </div>

      <div v-else-if="phase === 'error'" class="alert alert-danger">
        {{ error }}
        <button class="btn btn-sm btn-outline-danger ms-2" @click="load">Try again</button>
      </div>

      <!-- writing -->
      <template v-else-if="phase === 'write' && item">
        <div class="small text-muted mb-2">Sentence {{ idx + 1 }} / {{ items.length }}</div>

        <div class="card border-0 shadow-sm">
          <div class="card-body p-4">
            <div class="small text-muted mb-1">Write in English:</div>
            <p class="fs-5 fw-semibold mb-3">“{{ item.prompt_vi }}”</p>

            <div class="mb-3 small">
              Use the English word for: <strong>{{ item.meaning_vi }}</strong>
              <template v-if="item.hinted || item.submitted">
                → <strong class="text-primary">{{ item.word }}</strong>
              </template>
              <button v-else class="btn btn-outline-secondary btn-sm rounded-circle ms-1 hint-btn" title="Show the word to use"
                @click="item.hinted = true">?</button>
            </div>

            <textarea ref="box" v-model="item.answer" class="form-control mb-2" rows="2" :readonly="item.submitted"
              placeholder="Your English sentence…" @keydown.enter.exact.prevent="item.submitted ? next() : submit()"></textarea>

            <div v-if="!item.submitted" class="d-flex align-items-center gap-2">
              <button class="btn btn-success btn-sm" :disabled="!item.answer.trim()" @click="submit">
                Check <small class="opacity-75 d-none d-md-inline">(Enter)</small>
              </button>
              <button class="btn btn-link btn-sm text-muted" @click="giveUp">I don't know</button>
            </div>

            <template v-else>
              <!-- forgot the word? (in Hard mode, remembering it is worth a tick) -->
              <div v-if="item.answer.trim() && !item.usedWord" class="small mb-2 text-warning">
                <i class="bi bi-exclamation-triangle-fill"></i> You didn't use “{{ item.word }}”.
                <button class="btn btn-link btn-sm p-0 ms-1 align-baseline" @click="edit">Edit my sentence</button>
              </div>
              <div v-else-if="item.answer.trim() && !item.hinted" class="small mb-2 text-success">
                <i class="bi bi-check-circle-fill"></i> You remembered “{{ item.word }}”.
              </div>

              <!-- AI verdict for this sentence -->
              <div v-if="item.answer.trim()" class="mb-2">
                <div v-if="item.checking" class="small text-muted">
                  <span class="spinner-border spinner-border-sm me-1"></span> AI is checking your sentence…
                </div>
                <div v-else-if="item.result?.graded" class="p-2 rounded small"
                  :class="item.result.correct ? 'bg-success-subtle' : 'bg-danger-subtle'">
                  <span class="badge me-1" :class="item.result.correct ? 'bg-success' : 'bg-danger'">
                    {{ item.result.correct ? '✓ Correct' : '✗ Not quite' }} · {{ item.result.score }}
                  </span>
                  {{ item.result.feedback_vi }}
                  <div v-if="item.result.corrected && item.result.corrected !== item.answer.trim()" class="mt-1">
                    <span class="text-muted">Better:</span> <span class="fw-semibold">{{ item.result.corrected }}</span>
                  </div>
                  <div v-if="item.result.scheduled" class="text-muted mt-1"><i class="bi bi-calendar-check"></i> Counted as a review of this word.</div>
                </div>
                <div v-else-if="item.checkError" class="small text-warning">
                  <i class="bi bi-wifi-off"></i> {{ item.checkError }}
                  <button class="btn btn-link btn-sm p-0 ms-1 align-baseline" @click="check(item)">Try again</button>
                </div>
              </div>
              <div class="p-2 rounded bg-light small d-flex align-items-start gap-2">
                <div class="flex-grow-1"><span class="text-muted">Sample:</span> {{ item.reference_en }}</div>
                <button class="btn btn-light btn-sm rounded-circle flex-shrink-0" title="Listen" @click="say(item.reference_en)">
                  <i class="bi bi-volume-up-fill"></i>
                </button>
              </div>
              <div class="d-flex mt-3">
                <button ref="nextBtn" class="btn btn-primary ms-auto" @click="next">
                  {{ idx < items.length - 1 ? 'Next' : 'Finish' }} <i class="bi bi-arrow-right"></i>
                </button>
              </div>
            </template>
          </div>
        </div>
      </template>

      <div v-else-if="phase === 'grading'" class="text-center py-5">
        <div class="spinner-border text-success"></div>
        <p class="mt-3 text-muted">Waiting for the last checks…</p>
      </div>

      <!-- summary -->
      <template v-else-if="phase === 'done'">
        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body text-center py-4">
            <i class="bi bi-trophy-fill display-5 text-warning"></i>
            <h4 class="fw-bold mt-2 mb-1">
              <template v-if="gradedCount">{{ correctCount }} / {{ gradedCount }} sentences correct</template>
              <template v-else>Round finished</template>
            </h4>
            <div v-if="scheduledCount" class="small text-muted">{{ scheduledCount }} due word{{ scheduledCount > 1 ? 's' : '' }} counted as reviewed.</div>
            <div v-if="gradeError" class="alert alert-warning py-2 small mt-3 mb-0">
              {{ gradeError }} <button class="btn btn-sm btn-outline-secondary ms-2" @click="finish">Check again</button>
            </div>
          </div>
        </div>

        <ForgottenWords :words="toPractise" />

        <div v-for="it in items" :key="it.example_id" class="card border-0 shadow-sm mb-2">
          <div class="card-body py-3 small">
            <div class="d-flex justify-content-between gap-2">
              <span class="fw-semibold">{{ it.word }} <span class="text-muted fw-normal">— {{ it.meaning_vi }}</span></span>
              <span v-if="it.result?.graded" class="badge align-self-start" :class="it.result.correct ? 'bg-success' : 'bg-danger'">
                {{ it.result.correct ? '✓' : '✗' }} {{ it.result.score }}
              </span>
              <span v-else-if="!it.answer.trim()" class="badge bg-secondary align-self-start">skipped</span>
            </div>
            <div class="text-muted mt-1">“{{ it.prompt_vi }}”</div>
            <div v-if="it.answer.trim()" class="mt-1"><span class="text-muted">You:</span> {{ it.answer }}</div>
            <div v-if="it.result?.feedback_vi" :class="it.result.correct ? 'text-success' : 'text-danger'">{{ it.result.feedback_vi }}</div>
            <div v-if="it.result?.corrected && it.result.corrected !== it.answer.trim()" class="mt-1">
              <span class="text-muted">Corrected:</span> <span class="text-success">{{ it.result.corrected }}</span>
            </div>
            <div class="mt-1 d-flex align-items-start gap-2">
              <span class="flex-grow-1"><span class="text-muted">Sample:</span> {{ it.reference_en }}</span>
              <button class="btn btn-light btn-sm rounded-circle flex-shrink-0" title="Listen" @click="say(it.reference_en)">
                <i class="bi bi-volume-up-fill"></i>
              </button>
            </div>
          </div>
        </div>

        <div class="text-center my-3">
          <button class="btn btn-success" @click="load"><i class="bi bi-arrow-repeat"></i> New round</button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { api, errorMessage } from '../../api'
import ForgottenWords from '../common/ForgottenWords.vue'
import { RATE_REPLAY, useSpeechSynthesis } from '../../composables/useSpeechSynthesis'
import { usesWord } from '../../utils/words'

interface Result { example_id: number; card_id: number; used_word: boolean; graded: boolean; correct: boolean; score: number; feedback_vi: string; corrected: string; scheduled: boolean }
interface Item {
  card_id: number
  word: string
  meaning_vi: string
  example_id: number
  prompt_vi: string
  reference_en: string
  answer: string
  submitted: boolean
  usedWord: boolean
  hinted: boolean // "?" pressed: the English word was shown before writing
  checking: boolean
  checkError: string
  result?: Result
}

const props = defineProps<{ deckId: number }>()
const { speakEnglish } = useSpeechSynthesis()

const phase = ref<'loading' | 'error' | 'write' | 'grading' | 'done'>('loading')
const items = ref<Item[]>([])
const idx = ref(0)
const error = ref('')
const warning = ref('')
const gradeError = ref('')
const box = ref<HTMLTextAreaElement | null>(null)
const nextBtn = ref<HTMLButtonElement | null>(null)

const item = computed(() => items.value[idx.value])
const graded = computed(() => items.value.filter((i) => i.result?.graded))
const gradedCount = computed(() => graded.value.length)
const correctCount = computed(() => graded.value.filter((i) => i.result!.correct).length)
const scheduledCount = computed(() => graded.value.filter((i) => i.result!.scheduled).length)
// wrong or skipped words: offered in the "forgotten words" drill
const toPractise = computed(() =>
  phase.value !== 'done' || gradeError.value
    ? []
    : items.value.filter((i) => !i.answer.trim() || (i.result?.graded && !i.result.correct)).map((i) => ({ id: i.card_id, front: i.word })),
)

function say(text: string) {
  speakEnglish(text, RATE_REPLAY).catch(() => {})
}

async function focus(el: 'box' | 'next') {
  await nextTick()
  ;(el === 'box' ? box.value : nextBtn.value)?.focus()
}

function submit() {
  const it = item.value
  if (!it || it.submitted || !it.answer.trim()) return
  it.usedWord = usesWord(it.answer, it.word)
  it.submitted = true
  say(it.reference_en)
  check(it) // AI verdict arrives while the sample is read
  focus('next')
}

// one request per sentence, so the verdict comes right after "Check"
const pending = new Map<number, Promise<void>>()
function check(it: Item): Promise<void> {
  const answer = it.answer.trim()
  const p = (async () => {
    it.checking = true
    it.checkError = ''
    it.result = undefined
    try {
      const { data } = await api.post('/practice/write/grade', { items: [{ example_id: it.example_id, answer, hinted: it.hinted }] })
      if (it.answer.trim() === answer) it.result = data.results?.[0]
    } catch (e) {
      it.checkError = `Could not check this sentence: ${errorMessage(e)}`
    } finally {
      it.checking = false
    }
  })()
  pending.set(it.example_id, p)
  return p
}

function giveUp() {
  const it = item.value
  if (!it) return
  it.answer = ''
  it.submitted = true
  say(it.reference_en)
  focus('next')
}

function edit() {
  const it = item.value
  if (!it) return
  it.submitted = false
  it.result = undefined
  focus('box')
}

function next() {
  if (idx.value < items.value.length - 1) {
    idx.value++
    focus('box')
  } else {
    finish()
  }
}

async function load() {
  pending.clear()
  phase.value = 'loading'
  error.value = ''
  warning.value = ''
  gradeError.value = ''
  try {
    const { data } = await api.post('/practice/write/prompts', { deck_id: props.deckId, count: 10 })
    items.value = (data.prompts ?? []).map((p: Omit<Item, 'answer' | 'submitted' | 'usedWord'>) => ({
      ...p, answer: '', submitted: false, usedWord: false, hinted: false, checking: false, checkError: '',
    }))
    warning.value = data.warning ?? ''
    if (!items.value.length) {
      phase.value = 'error'
      error.value = 'No sentences could be prepared for this deck.'
      return
    }
    idx.value = 0
      phase.value = 'write'
    focus('box')
  } catch (e) {
    phase.value = 'error'
    error.value = errorMessage(e)
  }
}

/**
 * End of the round: wait for the checks still running, then check in one
 * request the sentences whose check failed (network, AI busy...).
 */
async function finish() {
  gradeError.value = ''
  phase.value = 'grading'
  await Promise.allSettled([...pending.values()])
  const missing = items.value.filter((i) => i.answer.trim() && !i.result?.graded)
  if (!missing.length) {
    phase.value = 'done'
    return
  }
  const payload = missing.map((i) => ({ example_id: i.example_id, answer: i.answer.trim(), hinted: i.hinted }))
  try {
    const { data } = await api.post('/practice/write/grade', { items: payload })
    const byId = new Map<number, Result>((data.results ?? []).map((r: Result) => [r.example_id, r]))
    for (const it of missing) {
      it.result = byId.get(it.example_id)
      it.checkError = ''
    }
  } catch (e) {
    gradeError.value = `Could not check the sentences: ${errorMessage(e)}`
  } finally {
    phase.value = 'done'
  }
}

watch(() => props.deckId, load)
onMounted(load)
</script>

<style scoped>
.hint-btn { width: 1.9rem; height: 1.9rem; padding: 0; font-weight: 700; line-height: 1; }
</style>
