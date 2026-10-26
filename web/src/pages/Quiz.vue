<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-end mb-3 gap-2">
      <div>
        <h2 class="fw-bold text-success mb-0"><i class="bi bi-question-circle"></i> Quiz</h2>
        <small class="text-muted">Correct answers on due cards also count as a review.</small>
      </div>
      <div class="d-flex flex-wrap gap-2 align-items-center">
        <div v-if="isEnglish" class="btn-group btn-group-sm" role="group" aria-label="Quiz mode">
          <button class="btn" :class="qmode === 'choose' ? 'btn-success' : 'btn-outline-success'" @click="setMode('choose')"
            title="Pick the right meaning"><i class="bi bi-ui-checks"></i> Choose</button>
          <button class="btn" :class="qmode === 'write' ? 'btn-success' : 'btn-outline-success'" @click="setMode('write')"
            title="Write the English sentence for a Vietnamese one"><i class="bi bi-pencil"></i> Write it</button>
        </div>
        <div v-if="qmode === 'choose' && isEnglish" class="form-check form-switch mb-0 small"
          title="Say the word aloud before choosing its meaning. Forgot how to say it? Press “I forgot”: it comes back as a stress question.">
          <input id="sayfirst" v-model="sayFirst" class="form-check-input" type="checkbox" />
          <label class="form-check-label" for="sayfirst">Say it first</label>
        </div>
        <div v-if="qmode === 'choose' && isEnglish" class="form-check form-switch mb-0 small">
          <input id="listen" v-model="listening" class="form-check-input" type="checkbox" />
          <label class="form-check-label" for="listen">Listening mode</label>
        </div>
        <select v-model.number="deckId" class="form-select w-auto" @change="start">
          <option v-for="d in subject?.decks ?? []" :key="d.id" :value="d.id">{{ d.name }}</option>
        </select>
      </div>
    </div>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>
    <!-- stress questions need syllables: prepared by AI only when the learner asks -->
    <div v-if="qmode === 'choose' && pron.missing > 0 && pron.canGenerate" class="alert alert-light border small d-flex flex-wrap align-items-center gap-2 py-2">
      <span><i class="bi bi-soundwave text-primary"></i> Pronunciation questions (stress + sounds): {{ pron.total - pron.missing }} / {{ pron.total }} words of this deck are ready.</span>
      <span v-if="pron.error" class="text-danger">{{ pron.error }}</span>
      <button class="btn btn-sm btn-outline-primary ms-auto" :disabled="pron.busy" @click="preparePron">
        <span v-if="pron.busy" class="spinner-border spinner-border-sm me-1"></span>
        {{ pron.busy ? 'AI is preparing… (up to a minute)' : `Prepare ${Math.min(25, pron.missing)} more words (AI)` }}
      </button>
    </div>
    <WriteIt v-if="qmode === 'write'" :key="deckId" :deck-id="deckId" />
    <div v-else-if="loading" class="text-center py-5"><div class="spinner-border text-success"></div></div>

    <div v-else-if="!questions.length" class="text-center py-5 text-muted">
      <i class="bi bi-inbox display-4"></i>
      <p class="mt-2">Not enough cards in this deck for a quiz yet (needs at least 4).</p>
    </div>

    <div v-else-if="idx < questions.length" class="row justify-content-center">
      <div class="col-lg-8">
        <div class="d-flex justify-content-between small text-muted mb-2">
          <span>
            <button type="button" class="btn btn-sm btn-outline-secondary py-0 me-2" :disabled="idx === 0" title="Previous question (←)" @click="back">
              <i class="bi bi-chevron-left"></i> Back
            </button>
            Question {{ idx + 1 }} / {{ questions.length }}
          </span>
          <span>Score: {{ score }}</span>
        </div>
        <div class="card border-0 shadow-sm">
          <div class="card-body p-4">
            <h5 class="fw-bold mb-3">
              <template v-if="listening && q.speak && !solved && !isPron(q.mode)">
                Listen and choose the meaning
              </template>
              <template v-else>
                {{ q.prompt }}
                <small v-if="q.prompt_hint && solved" class="text-muted fw-normal">({{ q.prompt_hint }})</small>
                <ViHint v-else-if="q.prompt_hint" :text="q.prompt_hint" class="align-middle" />
              </template>
            </h5>

            <!-- "Which syllable is stressed?" (pronunciation has its own schedule) -->
            <template v-if="q.mode === 'stress'">
              <div class="text-center my-3">
                <div class="fs-2 fw-bold mb-3">{{ q.speak }}</div>
                <div class="d-flex flex-wrap justify-content-center gap-2">
                  <button v-for="(syl, i) in q.syllables" :key="i" type="button" class="btn btn-lg syl" :class="sylClass(i)"
                    :disabled="!!stress || checking" @click="pickStress(i)">
                    {{ stress && i === stress.pronunciation.stress ? syl.toUpperCase() : syl }}
                    <small class="d-block key-hint">{{ i + 1 }}</small>
                  </button>
                </div>
              </div>
              <div v-if="stress" class="mt-3">
                <div :class="stress.correct ? 'text-success' : 'text-danger'" class="fw-bold mb-2">
                  <i :class="stress.correct ? 'bi bi-check-circle-fill' : 'bi bi-x-circle-fill'"></i>
                  {{ stress.correct ? 'Right!' : `Not quite: the stress is on “${q.syllables![stress.pronunciation.stress]}”` }}
                  <span v-if="stress.applied && stress.next_in" class="fw-normal small text-muted ms-1">· asked again in {{ stress.next_in }}</span>
                </div>
                <Pronunciation :card-id="q.card_id" :word="q.speak!" start class="mb-2" />
                <SpeakCheck v-if="q.speak" :text="q.speak" label="Say it" model-first class="mb-3" />
                <button v-if="idx > 0" class="btn btn-outline-secondary me-2" @click="back"><i class="bi bi-chevron-left"></i> Back</button>
                <button class="btn btn-primary" @click="next">{{ idx < questions.length - 1 ? 'Next' : 'See result' }}</button>
              </div>
            </template>

            <!-- "How do you say the “u” in FLUC?" (vowels read wrongly from the spelling) -->
            <template v-else-if="q.mode === 'sound'">
              <div class="text-center my-3">
                <div class="fs-2 fw-bold mb-3">
                  <template v-for="(syl, i) in q.syllables" :key="i">
                    <span v-if="i" class="syl-dot">·</span>
                    <template v-for="(part, k) in focusParts(syl, i)" :key="k">
                      <mark v-if="part.hit" class="focus-letters">{{ part.text }}</mark><template v-else>{{ part.text }}</template>
                    </template>
                  </template>
                </div>
                <div class="d-grid gap-2 mx-auto sound-opts">
                  <button v-for="(c, i) in q.choices" :key="c.id" type="button" class="btn text-start d-flex align-items-center" :class="soundClass(c.id)"
                    :disabled="!!sound || checking" @click="pickSound(c.id)">
                    <span class="badge bg-light text-dark me-2">{{ i + 1 }}</span>
                    <span class="flex-grow-1">{{ c.text }} <small v-if="c.hint && sound" class="vi-after">({{ c.hint }})</small></span>
                    <ViHint v-if="c.hint && !sound" :text="c.hint" />
                  </button>
                </div>
              </div>
              <div v-if="sound" class="mt-3">
                <div :class="sound.correct ? 'text-success' : 'text-danger'" class="fw-bold mb-1">
                  <i :class="sound.correct ? 'bi bi-check-circle-fill' : 'bi bi-x-circle-fill'"></i>
                  {{ sound.correct ? 'Right!' : `Not quite: it is ${q.choices[sound.pronunciation.sound?.answer ?? 0]?.text}` }}
                  <span v-if="sound.applied && sound.next_in" class="fw-normal small text-muted ms-1">· asked again in {{ sound.next_in }}</span>
                </div>
                <div v-if="sound.pronunciation.sound?.note_vi" class="small mb-1">
                  <i class="bi bi-lightbulb text-warning"></i> {{ sound.pronunciation.sound.note_vi }}
                </div>
                <div v-if="sound.pronunciation.sound?.same?.length" class="small mb-2">
                  Same sound:
                  <button v-for="w in sound.pronunciation.sound.same" :key="w" type="button" class="btn btn-light btn-sm py-0 px-2 me-1"
                    title="Listen" @click="speakEnglish(w, RATE_REPLAY).catch(() => {})">
                    <i class="bi bi-volume-up small"></i> {{ w }}
                  </button>
                </div>
                <Pronunciation :card-id="q.card_id" :word="q.speak!" start class="mb-2" />
                <SpeakCheck v-if="q.speak" :text="q.speak" label="Say it" model-first class="mb-3" />
                <button v-if="idx > 0" class="btn btn-outline-secondary me-2" @click="back"><i class="bi bi-chevron-left"></i> Back</button>
                <button class="btn btn-primary" @click="next">{{ idx < questions.length - 1 ? 'Next' : 'See result' }}</button>
              </div>
            </template>

            <template v-else>
            <button v-if="q.speak && !sayPending" class="btn btn-sm btn-outline-secondary mb-3" @click="speakEnglish(q.speak, RATE_REPLAY).catch(() => {})">
              <i class="bi bi-volume-up"></i> Play
            </button>
            <!-- say the word first: no model audio, the result is shown only after answering -->
            <div v-if="sayPending" class="say-first text-center p-4 mb-2 rounded bg-light">
              <p class="mb-3">Say <strong>“{{ q.speak }}”</strong> aloud first, then choose its meaning.</p>
              <div class="d-flex flex-wrap justify-content-center gap-2">
                <button v-if="micOK" class="btn btn-lg" :class="recListening ? 'btn-danger' : 'btn-outline-danger'" @click="recListening ? recStop() : sayWord()">
                  <i :class="recListening ? 'bi bi-stop-circle-fill' : 'bi bi-mic-fill'"></i>
                  {{ recListening ? 'Listening… tap to stop' : 'Say it (S)' }}
                </button>
                <button v-else class="btn btn-lg btn-outline-success" @click="saidSelf"><i class="bi bi-check2"></i> I said it</button>
                <button class="btn btn-lg btn-outline-secondary" :disabled="recListening" @click="forgotSay">
                  <i class="bi bi-question-circle"></i> I forgot how to say it (F)
                </button>
              </div>
              <div v-if="recListening && recTranscript" class="small text-muted fst-italic mt-2">{{ recTranscript }}…</div>
              <div v-if="recError" class="small text-danger mt-2">{{ recError }}</div>
              <div class="mt-2">
                <button class="btn btn-link btn-sm text-muted" @click="saySkipped = true">Skip, I can't speak right now</button>
              </div>
            </div>
            <template v-else>
            <div v-if="said?.forgot && !solved" class="p-2 mb-3 rounded bg-light">
              <div class="small text-muted mb-1">Listen and say it after the model:</div>
              <Pronunciation :card-id="q.card_id" :word="q.speak!" compact />
              <button class="btn btn-light btn-sm rounded-circle ms-1" title="Listen again" @click="speakEnglish(q.speak!, RATE_SLOW).catch(() => {})">
                <i class="bi bi-volume-up-fill"></i>
              </button>
            </div>
            <button v-for="(c, i) in q.choices" :key="c.id" class="btn w-100 py-2 mb-2 text-start d-flex align-items-center" :class="choiceClass(c.id)"
              :disabled="solved || wrong.has(c.id) || checking" @click="answer(c.id)">
              <span class="badge bg-light text-dark me-2">{{ 'ABCD'[i] }}</span>
              <span class="flex-grow-1">
                {{ c.text }}
                <!-- after answering, the Vietnamese is shown so the learner can compare -->
                <small v-if="c.hint && (solved || wrong.has(c.id))" class="vi-after">({{ c.hint }})</small>
              </span>
              <ViHint v-if="c.hint && !solved && !wrong.has(c.id)" :text="c.hint" />
            </button>
            <!-- don't guess: see the answer now, the word goes back to learning and to practice -->
            <button v-if="!result" class="btn btn-outline-secondary w-100 py-2 mb-2 text-start d-flex align-items-center idk"
              :disabled="checking" @click="dontKnow">
              <span class="badge bg-light text-dark me-2">K</span><span class="flex-grow-1">? I don't know</span>
            </button>
            </template>

            <!-- wrong, still trying: the example sentence is the hint -->
            <div v-if="result && !solved" class="mt-3">
              <div class="d-flex flex-wrap align-items-center gap-2">
                <span class="text-danger fw-bold me-2"><i class="bi bi-x-circle-fill"></i> Not quite, listen to the example and try again</span>
                <template v-if="hintText">
                  <button class="btn btn-light btn-sm rounded-circle" title="Listen again" @click="speakHint(RATE_REPLAY)">
                    <i class="bi bi-volume-up-fill"></i>
                  </button>
                  <button class="btn btn-light btn-sm rounded-circle" title="Listen slowly" @click="speakHint(RATE_SLOW)">🐢</button>
                  <button class="btn btn-outline-secondary btn-sm" @click="showHint = !showHint">
                    <i :class="showHint ? 'bi bi-eye-slash' : 'bi bi-card-text'"></i> {{ showHint ? 'Hide text' : 'Show sentence' }}
                  </button>
                </template>
                <button class="btn btn-link btn-sm p-0 ms-auto" @click="reveal">Show answer</button>
              </div>
              <div v-if="showHint && hintText" class="mt-2 p-2 rounded bg-light fst-italic">“{{ hintText }}”</div>
            </div>

            <div v-if="result && solved" class="mt-3">
              <div :class="result.correct ? 'text-success' : 'text-warning'" class="fw-bold mb-1">
                <i :class="result.correct ? 'bi bi-check-circle-fill' : 'bi bi-info-circle-fill'"></i>
                {{ result.correct ? 'Correct!' : revealed ? 'Here is the answer' : `Found it on try ${wrong.size + 1} (counted as wrong)` }}
              </div>
              <!-- how to say it: stress always shown after answering -->
              <Pronunciation v-if="result.card.kind === 'vocab' || result.card.kind === 'phrasal'" :card-id="q.card_id" :word="result.card.front" compact class="mb-2" />
              <!-- the word said before answering (mic): for practice only, the schedule is not changed -->
              <div v-if="said && said.forgot" class="small text-muted mb-2">
                <i class="bi bi-arrow-repeat"></i> Marked “forgot how to say it”: it comes back as a stress question.
              </div>
              <div v-else-if="said && !said.self" class="small mb-2" :class="said.ok ? 'text-success' : 'text-danger'">
                <i :class="said.ok ? 'bi bi-mic-fill' : 'bi bi-mic-mute-fill'"></i>
                {{ said.ok ? `Mic check: heard “${said.heard}”.` : `Mic check: heard “${said.heard || '…'}”. Listen to the model and try again:` }}
              </div>
              <SpeakCheck v-if="said && !said.ok && !said.self && q.speak" :text="q.speak" label="Listen, then say the word" model-first class="mb-3" />
              <div v-if="result.explanation" class="small text-muted mb-2">{{ result.explanation }}</div>
              <div v-else-if="result.card.example_en" class="small mb-2 d-flex align-items-start gap-2">
                <div>
                  <em>{{ result.card.example_en }}</em><br /><span class="text-muted">{{ result.card.example_vi }}</span>
                </div>
                <button class="btn btn-light btn-sm rounded-circle" title="Listen again" @click="speakResolved(RATE_REPLAY)">
                  <i class="bi bi-volume-up-fill"></i>
                </button>
              </div>
              <SpeakCheck v-if="!result.explanation && result.card.example_en" :text="result.card.example_en" class="mb-3" />
              <button v-if="idx > 0" class="btn btn-outline-secondary me-2" @click="back"><i class="bi bi-chevron-left"></i> Back</button>
              <button class="btn btn-primary" @click="next">{{ idx < questions.length - 1 ? 'Next' : 'See result' }}</button>
            </div>
            </template>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="text-center">
      <div class="card border-0 shadow-sm">
        <div class="card-body py-5">
          <i class="bi bi-trophy-fill display-3 text-warning"></i>
          <h3 class="mt-3 fw-bold">{{ score }} / {{ questions.length }}</h3>
          <p class="text-muted">{{ Math.round((score / questions.length) * 100) }}% correct</p>
          <p v-if="stressCount.total || sayCount.forgot" class="text-muted">
            <i class="bi bi-soundwave"></i>
            <template v-if="stressCount.total">Pronunciation: {{ stressCount.ok }} / {{ stressCount.total }}</template>
            <template v-if="stressCount.total && sayCount.forgot"> · </template>
            <template v-if="sayCount.forgot">forgot how to say {{ sayCount.forgot }} word{{ sayCount.forgot > 1 ? 's' : '' }} (coming back as stress questions)</template>
          </p>
          <!-- next step: write sentences with the same words (web only; it is easy to forget to switch to Write it) -->
          <div v-if="roundWords.length && !writing" class="next-step mx-auto mb-4 p-3 rounded d-flex flex-wrap align-items-center gap-2 text-start" style="max-width: 720px">
            <div class="flex-grow-1">
              <div class="fw-bold"><i class="bi bi-pencil-square text-success"></i> Next: Write it with these {{ roundWords.length }} words</div>
              <div class="small text-muted">A Vietnamese sentence for each word: write it in English, AI checks it right away.</div>
            </div>
            <button class="btn btn-success" @click="writing = true"><i class="bi bi-arrow-right"></i> Write them now</button>
          </div>
          <WriteIt v-if="writing" :card-ids="roundWords.map((w) => w.id)" class="mx-auto mb-4 text-start" style="max-width: 820px" />
          <ForgottenWords :words="forgotten" class="mx-auto mb-4" style="max-width: 640px" />
          <ClozeRound v-if="roundWords.length >= 3 && !clozeDone" :words="roundWords" class="mx-auto mb-4" style="max-width: 720px"
            @done="clozeDone = true" />
          <button class="btn btn-outline-secondary me-2" @click="back"><i class="bi bi-chevron-left"></i> Back to questions</button>
          <button class="btn btn-primary me-2" @click="start"><i class="bi bi-arrow-repeat"></i> New quiz</button>
          <router-link to="/" class="btn btn-outline-secondary">Home</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, errorMessage } from '../api'
import { RATE_REPLAY, RATE_SLOW, useSpeechSynthesis } from '../composables/useSpeechSynthesis'
import SpeakCheck from '../components/common/SpeakCheck.vue'
import ForgottenWords from '../components/common/ForgottenWords.vue'
import WriteIt from '../components/quiz/WriteIt.vue'
import ClozeRound from '../components/common/ClozeRound.vue'
import ViHint from '../components/common/ViHint.vue'
import Pronunciation from '../components/Pronunciation.vue'
import { usePronunciation } from '../composables/usePronunciation'
import { scoreSpeech, speechRecognitionSupported, useSpeechRecognition } from '../composables/useSpeechRecognition'
import { isEnglish, loadSubjects, subject, subjectId } from '../composables/useSubject'

interface Question {
  card_id: number; kind: string; mode: string; prompt: string; prompt_hint?: string; speak?: string
  choices: { id: number; text: string; hint?: string }[]
  syllables?: string[] // modes "stress", "sound"
  focus?: { syllable: number; letters: string } // mode "sound"
}
interface SoundRes {
  correct: boolean
  word: string
  pronunciation: {
    syllables: string[]; stress: number; tips_vi: string[]
    sound?: { syllable: number; letters: string; answer: number; same: string[]; note_vi: string }
  }
  applied: boolean
  next_in?: string
}
interface StressRes {
  correct: boolean
  word: string
  pronunciation: { syllables: string[]; stress: number; tips_vi: string[] }
  applied: boolean
  next_in?: string
}
interface Result {
  correct: boolean
  correct_choice_id: number
  explanation?: string
  card: { front: string; kind: string; example_en?: string; example_vi?: string }
}

const route = useRoute()
const router = useRouter()
const { speakEnglish } = useSpeechSynthesis()
const { cache: pronCache, generate: pronGenerate } = usePronunciation()

const deckId = ref<number>(Number(route.query.deck_id) || 0)
// ?count=5 from the "Quick quiz" reminder / the empty Review page
const quizCount = Math.min(50, Math.max(5, Number(route.query.count) || 10))
const questions = ref<Question[]>([])
const idx = ref(0)
const score = ref(0)
const result = ref<Result | null>(null) // result of the FIRST attempt (this is what counts)
const wrong = reactive(new Set<number>()) // options already tried and wrong
const solved = ref(false) // correct option found or answer revealed
const revealed = ref(false)
const showHint = ref(false)
const checking = ref(false)
const loading = ref(true)
const listening = ref(false)
const error = ref('')
let speakToken = 0
// English words of this quiz: a fill-in-the-blank check is offered at the end
const roundWords = ref<{ id: number; front: string }[]>([])
const clozeDone = ref(false)
const writing = ref(false) // "Write it" with the words of this quiz, at the end
function trackRound(r: Result) {
  if (r.card.kind !== 'vocab' && r.card.kind !== 'phrasal') return
  const id = q.value.card_id
  if (!roundWords.value.some((w) => w.id === id)) roundWords.value.push({ id, front: r.card.front })
}

// English words answered wrong on the first try: offered for practice at the end
const forgotten = ref<{ id: number; front: string }[]>([])
function remember(r: Result | null) {
  if (!r || (r.card.kind !== 'vocab' && r.card.kind !== 'phrasal')) return
  const id = q.value.card_id
  if (!forgotten.value.some((w) => w.id === id)) forgotten.value.push({ id, front: r.card.front })
}

const q = computed(() => questions.value[idx.value])

// ---------------------------------------------------------------- say it first
// Before choosing, the learner says the word (no model audio first). The mic
// result is only for practice (speech recognition is not a reliable judge, so
// the schedule is never changed by it). "I forgot how to say it" puts the
// word on the pronunciation schedule: it comes back as a stress question.
const micOK = speechRecognitionSupported
const SAY_PASS = 80
const sayFirst = ref(storageGet('elearn.quiz.sayFirst') !== 'false')
watch(sayFirst, (v) => storageSet('elearn.quiz.sayFirst', String(v)))
const { listening: recListening, transcript: recTranscript, error: recError, listen: recListen, stop: recStop } = useSpeechRecognition()
interface Said { ok: boolean; heard: string; forgot?: boolean; self?: boolean }
const said = ref<Said | null>(null)
const saySkipped = ref(false)
const sayLog = reactive(new Map<number, 'ok' | 'miss' | 'forgot' | 'self'>()) // question index -> outcome
const sayCount = computed(() => ({ forgot: [...sayLog.values()].filter((v) => v === 'forgot').length }))

const needsSay = computed(() => {
  const cur = q.value
  return sayFirst.value && !listening.value && !!cur && cur.mode === 'meaning' && !!cur.speak && (cur.kind === 'vocab' || cur.kind === 'phrasal')
})
const sayPending = computed(() => needsSay.value && !said.value && !saySkipped.value && !result.value)

async function sayWord() {
  const target = q.value?.speak
  if (!target || recListening.value) return
  const at = idx.value
  const alts = await recListen()
  if (!alts.length || idx.value !== at) return // nothing heard: try again (the error says why)
  const sc = scoreSpeech(target, alts)
  said.value = { ok: sc.pct >= SAY_PASS, heard: sc.heard }
  sayLog.set(at, said.value.ok ? 'ok' : 'miss')
}

function saidSelf() {
  said.value = { ok: true, heard: '', self: true }
  sayLog.set(idx.value, 'self')
}

/** "I forgot how to say it": model audio + stress now, and the word's pronunciation comes back later. */
async function forgotSay() {
  const cur = q.value
  if (!cur?.speak) return
  recStop()
  said.value = { ok: false, heard: '', forgot: true }
  sayLog.set(idx.value, 'forgot')
  api.post('/pron/forgot', { card_id: cur.card_id }).catch(() => {})
  if (!pronCache.get(cur.card_id)) pronGenerate(cur.card_id) // returns the stored one, or asks AI once
  const token = ++speakToken
  try {
    await speakEnglish(cur.speak, RATE_SLOW)
    await new Promise((r) => setTimeout(r, 300))
    if (token === speakToken) await speakEnglish(cur.speak, RATE_REPLAY)
  } catch {
    /* interrupted */
  }
}

// ---------------------------------------------------------------- stress questions
const stress = ref<StressRes | null>(null)
const stressLog = reactive(new Map<number, boolean>())
const stressCount = computed(() => ({ total: stressLog.size, ok: [...stressLog.values()].filter(Boolean).length }))
const picked = ref(-1)

function sylClass(i: number) {
  const r = stress.value
  if (!r) return 'btn-outline-primary'
  if (i === r.pronunciation.stress) return 'btn-success fw-bold'
  if (i === picked.value) return 'btn-outline-danger text-decoration-line-through'
  return 'btn-outline-secondary opacity-50'
}

// ---------------------------------------------------------------- sound questions
const sound = ref<SoundRes | null>(null)

/** The syllable split around the letters asked about (highlighted). */
function focusParts(syl: string, i: number): { text: string; hit: boolean }[] {
  const f = q.value?.focus
  if (!f || i !== f.syllable) return [{ text: syl, hit: false }]
  const at = syl.toLowerCase().indexOf(f.letters.toLowerCase())
  if (at < 0) return [{ text: syl, hit: false }]
  return [
    { text: syl.slice(0, at), hit: false },
    { text: syl.slice(at, at + f.letters.length), hit: true },
    { text: syl.slice(at + f.letters.length), hit: false },
  ].filter((p) => p.text)
}

function soundClass(id: number) {
  const r = sound.value
  if (!r) return 'btn-outline-primary'
  if (id === r.pronunciation.sound?.answer) return 'btn-success'
  if (id === picked.value) return 'btn-outline-danger text-decoration-line-through'
  return 'btn-outline-secondary opacity-50'
}

async function pickSound(i: number) {
  const cur = q.value
  if (!cur || sound.value || checking.value) return
  checking.value = true
  picked.value = i
  try {
    const { data } = await api.post('/quiz/sound', { card_id: cur.card_id, option: i })
    sound.value = data
    solved.value = true
    stressLog.set(idx.value, data.correct)
    if (data.correct) score.value++
    speakEnglish(cur.speak ?? data.word, RATE_REPLAY).catch(() => {})
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    checking.value = false
  }
}

async function pickStress(i: number) {
  const cur = q.value
  if (!cur || stress.value || checking.value) return
  checking.value = true
  picked.value = i
  try {
    const { data } = await api.post('/quiz/stress', { card_id: cur.card_id, syllable: i })
    stress.value = data
    solved.value = true
    stressLog.set(idx.value, data.correct)
    if (data.correct) score.value++
    speakEnglish(cur.speak ?? data.word, RATE_REPLAY).catch(() => {})
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    checking.value = false
  }
}

function storageGet(k: string) {
  try { return localStorage.getItem(k) } catch { return null }
}
function storageSet(k: string, v: string) {
  try { localStorage.setItem(k, v) } catch { /* private mode */ }
}

// don't re-read questions that were already answered (going back)
const isPron = (m?: string) => m === 'stress' || m === 'sound'
watch(q, (cur) => { if (cur?.speak && !isPron(cur.mode) && listening.value && !solved.value) speakEnglish(cur.speak).catch(() => {}) })

function choiceClass(id: number) {
  if (!result.value) return 'btn-outline-primary'
  if (solved.value && id === result.value.correct_choice_id) return 'btn-success'
  if (wrong.has(id)) return 'btn-outline-danger text-decoration-line-through'
  return solved.value ? 'btn-outline-secondary opacity-50' : 'btn-outline-primary'
}

/**
 * Read the word, then its example sentence. Used once the question is solved,
 * and after a wrong first try on "What does X mean?" questions (reading X does
 * not give the answer away there; on "Which one means" it would).
 */
async function speakResolved(rate = 1) {
  const r = result.value
  if (!r || r.card.kind === 'mcq' || r.card.kind === 'concept') return
  const token = ++speakToken
  try {
    await speakEnglish(r.card.front, rate)
    if (!r.card.example_en) return
    await new Promise((res) => setTimeout(res, 400))
    if (token === speakToken) await speakEnglish(r.card.example_en, rate)
  } catch {
    /* interrupted */
  }
}

/** Regexp matching the studied word in common forms (negotiate→negotiated, carry out→carried out). */
function wordRegex(word: string): RegExp {
  const parts = word.toLowerCase().trim().split(/\s+/).map((p, i) => {
    let x = p.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    if (i === 0 && x.length > 3 && /[ey]$/.test(x)) x = x.slice(0, -1)
    return x + '\\w*'
  })
  return new RegExp('\\b' + parts.join('(?:\\W+\\w+){0,3}?\\W+') + '\\b', 'gi')
}

/**
 * Example sentence used as a hint after a wrong answer. On "Which one means…"
 * questions the options are words, so the word itself is hidden.
 */
const hintText = computed(() => {
  const r = result.value
  const ex = r?.card.example_en
  if (!r || !ex) return ''
  return q.value?.mode === 'term' ? ex.replace(wordRegex(r.card.front), '_____') : ex
})

async function speakHint(rate: number) {
  const text = hintText.value.replace(/_____/g, 'blank')
  if (!text) return
  ++speakToken
  try {
    await speakEnglish(text, rate)
  } catch {
    /* interrupted */
  }
}

async function answer(id: number) {
  if (solved.value || wrong.has(id) || checking.value) return
  // later attempts are checked locally: only the first one is scored and sent to the server
  if (result.value) {
    if (id === result.value.correct_choice_id) {
      solved.value = true
      speakResolved()
    } else {
      wrong.add(id)
      speakHint(1)
    }
    return
  }
  checking.value = true
  try {
    const { data } = await api.post('/quiz/answer', { card_id: q.value.card_id, choice_id: id })
    result.value = data
    trackRound(data)
    if (data.correct) {
      score.value++
      solved.value = true
      speakResolved()
    } else {
      wrong.add(id)
      remember(data)
      speakHint(1)
    }
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    checking.value = false
  }
}

/** "I don't know this word": counted as forgotten (Again if due), answer shown, word added to practice. */
async function dontKnow() {
  if (result.value || checking.value) return
  checking.value = true
  try {
    const { data } = await api.post('/quiz/answer', { card_id: q.value.card_id, choice_id: -1 })
    result.value = data
    trackRound(data)
    remember(data)
    reveal()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    checking.value = false
  }
}

function reveal() {
  revealed.value = true
  solved.value = true
  speakResolved()
}

function resetQuestion() {
  speakToken++
  recStop()
  said.value = null
  saySkipped.value = false
  stress.value = null
  sound.value = null
  picked.value = -1
  showHint.value = false
  result.value = null
  wrong.clear()
  solved.value = false
  revealed.value = false
}

/** State of questions already visited, so "Back" shows them as they were (score is not changed). */
interface Saved { result: Result | null; wrong: number[]; solved: boolean; revealed: boolean; said: Said | null; saySkipped: boolean; stress: StressRes | null; sound: SoundRes | null; picked: number }
let saved: Record<number, Saved> = {}

function saveState() {
  saved[idx.value] = { result: result.value, wrong: [...wrong], solved: solved.value, revealed: revealed.value, said: said.value, saySkipped: saySkipped.value, stress: stress.value, sound: sound.value, picked: picked.value }
}

function go(i: number) {
  if (i < 0 || i > questions.value.length) return
  if (idx.value < questions.value.length) saveState()
  resetQuestion()
  const st = saved[i]
  if (st) {
    result.value = st.result
    st.wrong.forEach((w) => wrong.add(w))
    solved.value = st.solved
    revealed.value = st.revealed
    said.value = st.said
    saySkipped.value = st.saySkipped
    stress.value = st.stress
    sound.value = st.sound
    picked.value = st.picked
  }
  idx.value = i
}

function next() {
  go(idx.value + 1)
}

function back() {
  go(idx.value - 1)
}

function onKey(e: KeyboardEvent) {
  if (qmode.value === 'write') return
  const tag = (e.target as HTMLElement)?.tagName
  if (tag === 'INPUT' || tag === 'SELECT' || tag === 'TEXTAREA') return
  if (e.ctrlKey || e.metaKey || e.altKey) return
  if (sayPending.value && (e.key === 's' || e.key === 'S') && micOK) {
    e.preventDefault()
    if (recListening.value) recStop()
    else sayWord()
    return
  }
  if (sayPending.value && (e.key === 'f' || e.key === 'F')) {
    e.preventDefault()
    forgotSay()
    return
  }
  if ((e.key === 'k' || e.key === 'K') && !isPron(q.value?.mode) && !result.value && !sayPending.value) {
    e.preventDefault()
    dontKnow()
    return
  }
  if (q.value?.mode === 'sound' && !sound.value && /^[1-4]$/.test(e.key)) {
    const i = Number(e.key) - 1
    if (i < q.value.choices.length) pickSound(i)
    return
  }
  if (q.value?.mode === 'stress' && !stress.value && /^[1-9]$/.test(e.key)) {
    const i = Number(e.key) - 1
    if (i < (q.value.syllables?.length ?? 0)) pickStress(i)
    return
  }
  if (e.key === 'ArrowLeft' && idx.value > 0) back()
  else if (e.key === 'ArrowRight' && solved.value) next()
}
onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))

// "choose" = multiple choice (+ listening), "write" = Write it (Vietnamese -> English sentence)
const qmode = ref<'choose' | 'write'>(route.query.mode === 'write' ? 'write' : 'choose')

function setMode(m: 'choose' | 'write') {
  if (qmode.value === m) return
  qmode.value = m
  start()
}

async function start() {
  if (!deckId.value) return
  router.replace({ query: { deck_id: deckId.value, ...(qmode.value === 'write' ? { mode: 'write' } : {}) } })
  if (qmode.value === 'write') {
    loading.value = false // WriteIt loads its own sentences
    return
  }
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.get('/quiz', { params: { deck_id: deckId.value, count: quizCount } })
    questions.value = data
    saved = {}
    forgotten.value = []
    roundWords.value = []
    clozeDone.value = false
    writing.value = false
    idx.value = 0
    score.value = 0
    sayLog.clear()
    stressLog.clear()
    resetQuestion()
    loadPronStatus()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

// ---------------------------------------------------------------- stress data for the deck
const pron = reactive({ missing: 0, total: 0, canGenerate: false, busy: false, error: '' })

async function loadPronStatus() {
  if (!deckId.value) return
  try {
    const { data } = await api.get('/quiz/pron-status', { params: { deck_id: deckId.value } })
    Object.assign(pron, { missing: data.missing, total: data.total, canGenerate: data.can_generate, error: '' })
  } catch {
    pron.missing = 0
  }
}

/** One AI request: syllables + stress for up to 25 words of the deck (kept for everyone). */
async function preparePron() {
  pron.busy = true
  pron.error = ''
  try {
    const { data } = await api.post('/quiz/pron-generate', { deck_id: deckId.value })
    pron.missing = data.missing
    pron.total = data.total
  } catch (e) {
    pron.error = errorMessage(e)
  } finally {
    pron.busy = false
  }
}

/** A deck of the topic chosen in the menu (the one asked for when it belongs to it). */
function deckOfTopic(want: number): number {
  const decks = subject.value?.decks ?? []
  return decks.some((d) => d.id === want) ? want : decks[0]?.id ?? 0
}

watch(subjectId, (id, old) => {
  if (!old || id === old) return
  if (!isEnglish.value) qmode.value = 'choose' // Write it is for English
  deckId.value = deckOfTopic(0)
  start()
})

onMounted(async () => {
  try {
    await loadSubjects()
    deckId.value = deckOfTopic(deckId.value)
    await start()
  } catch (e) {
    error.value = errorMessage(e)
    loading.value = false
  }
})
</script>

<style scoped>
.idk { border-style: dashed; }
.vi-after { opacity: 0.75; font-size: 0.85em; }
.next-step { background: #ecfdf5; border: 1px solid #a7f3d0; }
.syl { min-width: 4rem; font-size: 1.35rem; line-height: 1.2; }
.syl-dot { color: var(--bs-gray-400); margin: 0 0.1rem; }
.focus-letters { background: #ffe69c; padding: 0 0.1em; border-radius: 4px; text-decoration: underline; text-underline-offset: 4px; }
.sound-opts { max-width: 420px; }
.key-hint { font-size: 0.65rem; opacity: 0.6; }
</style>
