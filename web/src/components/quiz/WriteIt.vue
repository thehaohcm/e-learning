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
        <div class="d-flex justify-content-between align-items-center small text-muted mb-2">
          <span>Sentence {{ idx + 1 }} / {{ items.length }}</span>
          <div class="btn-group btn-group-sm">
            <button class="btn" :class="level === 'easy' ? 'btn-secondary' : 'btn-outline-secondary'" @click="level = 'easy'"
              title="The English word to use is shown">Easy</button>
            <button class="btn" :class="level === 'hard' ? 'btn-secondary' : 'btn-outline-secondary'" @click="level = 'hard'"
              title="Only its Vietnamese meaning is shown: recall the English word">Hard</button>
          </div>
        </div>

        <div class="card border-0 shadow-sm">
          <div class="card-body p-4">
            <div class="small text-muted mb-1">Write in English:</div>
            <p class="fs-5 fw-semibold mb-3">“{{ item.prompt_vi }}”</p>

            <div class="mb-3 small">
              <template v-if="level === 'easy' || item.submitted || hints >= 2">
                Use: <strong class="text-primary">{{ item.word }}</strong>
                <span class="text-muted">({{ item.meaning_vi }})</span>
              </template>
              <template v-else>
                Use the English word for: <strong>{{ item.meaning_vi }}</strong>
                <span v-if="hints === 1" class="ms-2 font-monospace text-primary">{{ letterHint }}</span>
                <button class="btn btn-link btn-sm py-0" @click="hints++">
                  <i class="bi bi-lightbulb"></i> {{ hints === 0 ? 'First letter' : 'Show the word' }}
                </button>
              </template>
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
              <div v-if="item.answer.trim()" class="small mb-2" :class="item.usedWord ? 'text-success' : 'text-warning'">
                <i :class="item.usedWord ? 'bi bi-check-circle-fill' : 'bi bi-exclamation-triangle-fill'"></i>
                {{ item.usedWord ? `You used “${item.word}”.` : `You didn't use “${item.word}”.` }}
                <button v-if="!item.usedWord" class="btn btn-link btn-sm p-0 ms-1 align-baseline" @click="edit">Edit my sentence</button>
              </div>
              <div class="p-2 rounded bg-light small d-flex align-items-start gap-2">
                <div class="flex-grow-1"><span class="text-muted">Sample:</span> {{ item.reference_en }}</div>
                <button class="btn btn-light btn-sm rounded-circle flex-shrink-0" title="Listen" @click="say(item.reference_en)">
                  <i class="bi bi-volume-up-fill"></i>
                </button>
              </div>
              <div class="small text-muted mt-1">AI checks your sentences at the end.</div>
              <div class="d-flex mt-3">
                <button ref="nextBtn" class="btn btn-primary ms-auto" @click="next">
                  {{ idx < items.length - 1 ? 'Next' : 'Finish & check' }} <i class="bi bi-arrow-right"></i>
                </button>
              </div>
            </template>
          </div>
        </div>
      </template>

      <div v-else-if="phase === 'grading'" class="text-center py-5">
        <div class="spinner-border text-success"></div>
        <p class="mt-3 text-muted">AI is checking your sentences…</p>
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
              {{ gradeError }} <button class="btn btn-sm btn-outline-secondary ms-2" @click="grade">Check again</button>
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
  result?: Result
}

const props = defineProps<{ deckId: number }>()
const { speakEnglish } = useSpeechSynthesis()

const phase = ref<'loading' | 'error' | 'write' | 'grading' | 'done'>('loading')
const items = ref<Item[]>([])
const idx = ref(0)
const hints = ref(0)
const error = ref('')
const warning = ref('')
const gradeError = ref('')
const box = ref<HTMLTextAreaElement | null>(null)
const nextBtn = ref<HTMLButtonElement | null>(null)
const level = ref<'easy' | 'hard'>(storageGet('elearn.writeit.level') === 'hard' ? 'hard' : 'easy')
watch(level, (v) => storageSet('elearn.writeit.level', v))

const item = computed(() => items.value[idx.value])
const letterHint = computed(() => {
  const w = item.value?.word ?? ''
  return w.split(' ').map((p) => p[0] + ' _'.repeat(Math.max(0, p.length - 1))).join('   ')
})
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

function storageGet(k: string) {
  try { return localStorage.getItem(k) } catch { return null }
}
function storageSet(k: string, v: string) {
  try { localStorage.setItem(k, v) } catch { /* ignore */ }
}

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
  focus('next')
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
  focus('box')
}

function next() {
  hints.value = 0
  if (idx.value < items.value.length - 1) {
    idx.value++
    focus('box')
  } else {
    grade()
  }
}

async function load() {
  phase.value = 'loading'
  error.value = ''
  warning.value = ''
  gradeError.value = ''
  try {
    const { data } = await api.post('/practice/write/prompts', { deck_id: props.deckId, count: 10 })
    items.value = (data.prompts ?? []).map((p: Omit<Item, 'answer' | 'submitted' | 'usedWord'>) => ({
      ...p, answer: '', submitted: false, usedWord: false,
    }))
    warning.value = data.warning ?? ''
    if (!items.value.length) {
      phase.value = 'error'
      error.value = 'No sentences could be prepared for this deck.'
      return
    }
    idx.value = 0
    hints.value = 0
    phase.value = 'write'
    focus('box')
  } catch (e) {
    phase.value = 'error'
    error.value = errorMessage(e)
  }
}

/** One AI request checks every sentence of the round. */
async function grade() {
  const payload = items.value
    .filter((i) => i.answer.trim())
    .map((i) => ({ example_id: i.example_id, answer: i.answer.trim() }))
  gradeError.value = ''
  if (!payload.length) {
    phase.value = 'done'
    return
  }
  phase.value = 'grading'
  try {
    const { data } = await api.post('/practice/write/grade', { items: payload })
    const byId = new Map<number, Result>((data.results ?? []).map((r: Result) => [r.example_id, r]))
    for (const it of items.value) it.result = byId.get(it.example_id)
  } catch (e) {
    gradeError.value = `Could not check the sentences: ${errorMessage(e)}`
  } finally {
    phase.value = 'done'
  }
}

watch(() => props.deckId, load)
onMounted(load)
</script>
