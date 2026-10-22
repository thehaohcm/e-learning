<template>
  <div class="card border-danger-subtle shadow-sm mb-3 practice">
    <div class="card-body">
      <div class="d-flex flex-wrap align-items-center gap-2 mb-2">
        <span class="fw-semibold">
          <template v-if="mode === 'shadow'"><i class="bi bi-mic"></i> Shadowing</template>
          <template v-else><i class="bi bi-people"></i> Role-play as <b>{{ role }}</b> · {{ level === 'hard' ? 'Hard' : 'Easy' }}</template>
        </span>
        <span v-if="!finished" class="small text-muted">{{ mode === 'shadow' ? 'sentence' : 'line' }} {{ pos + 1 }} / {{ units.length }}</span>
        <button class="btn btn-sm btn-outline-secondary ms-auto" @click="close">
          <i class="bi bi-x-lg"></i> {{ finished ? 'Close' : 'Stop' }}
        </button>
      </div>
      <div class="progress mb-3" style="height: 4px">
        <div class="progress-bar bg-danger" :style="{ width: progressPct + '%' }"></div>
      </div>

      <!-- current unit -->
      <template v-if="!finished && unit">
        <!-- partner line (role-play): the app speaks -->
        <div v-if="!unit.mine" class="text-muted">
          <span class="fw-semibold">{{ unit.speaker }}:</span> <span class="fst-italic">{{ unit.text }}</span>
        </div>

        <template v-else>
          <div v-if="unit.speaker" class="small fw-semibold mb-1">{{ unit.speaker }} (you)</div>
          <!-- what to say: hidden in hard role-play until solved or after 3 tries -->
          <div v-if="showText" class="fs-5 mb-1 words">
            <template v-if="score">
              <span v-for="(w, k) in score.words" :key="k" :class="w.ok ? 'ok' : 'miss'">{{ w.word }} </span>
            </template>
            <template v-else>{{ unit.text }}</template>
          </div>
          <div v-else class="mb-1">
            <div class="fs-5 text-primary">{{ unit.vi || '(no translation)' }}</div>
            <div class="small text-muted">Say it in English — your own words are fine if the meaning is the same.</div>
          </div>
          <div v-if="unit.vi && showText && level === 'hard'" class="small text-muted mb-1">{{ unit.vi }}</div>

          <!-- status -->
          <div class="d-flex flex-wrap align-items-center gap-2 mt-2">
            <span v-if="phase === 'playing'" class="text-muted small"><i class="bi bi-volume-up-fill"></i> Listen…</span>
            <span v-else-if="phase === 'listening'" class="text-danger small fw-semibold">
              <span class="rec-dot"></span> Your turn: speak now
              <button class="btn btn-link btn-sm p-0 ms-1" @click="stopListening">done</button>
            </span>
            <template v-else-if="score">
              <span class="badge" :class="score.pct >= pass ? 'bg-success' : score.pct >= 40 ? 'bg-warning text-dark' : 'bg-danger'">{{ score.pct }}%</span>
              <span v-if="score.pct >= pass" class="text-success small fw-semibold">Great!</span>
              <span v-else class="small text-muted">try {{ tries }}/{{ MAX_TRIES }}<template v-if="score.heard"> · you said “{{ score.heard }}”</template></span>
            </template>
            <span v-if="phase === 'listening' && transcript" class="small fst-italic text-muted">{{ transcript }}…</span>
          </div>
          <div v-if="recError" class="small text-danger mt-1">{{ recError }}</div>

          <div v-if="phase === 'result'" class="d-flex flex-wrap gap-2 mt-3">
            <template v-if="!score || score.pct < pass">
              <button class="btn btn-danger btn-sm" @click="tryAgain">
                <i class="bi bi-arrow-repeat"></i> {{ showText ? 'Listen slowly & try again' : 'Try again' }}
              </button>
              <button v-if="!showText" class="btn btn-outline-secondary btn-sm" @click="hear">
                <i class="bi bi-volume-up"></i> Hear the line
              </button>
            </template>
            <button class="btn btn-sm ms-auto" :class="canGoOn ? 'btn-success' : 'btn-outline-secondary'" @click="advance">
              {{ canGoOn ? 'Next' : 'Skip' }} <i class="bi bi-arrow-right"></i>
            </button>
          </div>
        </template>
      </template>

      <!-- summary -->
      <template v-if="finished">
        <div class="text-center mb-3">
          <div class="fs-4 fw-bold">{{ avg }}%</div>
          <div class="small text-muted">average over {{ spoken.length }} {{ mode === 'shadow' ? 'sentences' : 'lines' }} you said ·
            {{ spoken.filter((u) => u.best >= pass).length }} reached {{ pass }}%</div>
        </div>
        <ul class="list-unstyled small mb-3">
          <li v-for="(u, k) in spoken" :key="k" class="d-flex gap-2 py-1 border-bottom">
            <span class="badge align-self-start" :class="u.best >= pass ? 'bg-success' : u.tries ? 'bg-warning text-dark' : 'bg-secondary'">
              {{ u.tries ? u.best + '%' : 'skipped' }}
            </span>
            <span>{{ u.text }}</span>
          </li>
        </ul>
        <div class="d-flex flex-wrap gap-2">
          <button v-if="weak.length" class="btn btn-danger btn-sm" @click="restart(true)">
            <i class="bi bi-arrow-repeat"></i> Practise the {{ weak.length }} weak one{{ weak.length > 1 ? 's' : '' }} again
          </button>
          <button class="btn btn-outline-danger btn-sm" @click="restart(false)">Start over</button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { scoreSpeech, useSpeechRecognition } from '../../composables/useSpeechRecognition'
import type { SpeechScore } from '../../composables/useSpeechRecognition'
import { RATE_NORMAL, RATE_REPLAY } from '../../composables/useSpeechSynthesis'
import type { PracticeItem } from './practice'


const props = defineProps<{
  items: PracticeItem[]
  mode: 'shadow' | 'roleplay'
  role?: string // role-play: the speaker the learner plays
  level?: 'easy' | 'hard' // role-play: hard = only the Vietnamese meaning is shown
  /** Reads a text aloud (the page picks the voice of the speaker). */
  say: (text: string, speaker: string | undefined, rate: number) => Promise<void>
}>()
const emit = defineEmits<{ (e: 'highlight', index: number): void; (e: 'close'): void }>()

interface Unit { index: number; text: string; vi?: string; speaker?: string; mine: boolean; best: number; tries: number }

const MAX_TRIES = 3
const LONG = 20 // words: longer sentences are split at commas for shadowing

const { listening, transcript, error: recError, listen, stop } = useSpeechRecognition()

const units = ref<Unit[]>([])
const pos = ref(0)
const phase = ref<'playing' | 'listening' | 'result' | 'partner'>('playing')
const score = ref<SpeechScore | null>(null)
const tries = ref(0)
const finished = ref(false)
const solvedHard = ref(false) // hard role-play: text revealed (passed or gave up)
let token = 0

const pass = computed(() => (props.mode === 'roleplay' && props.level === 'hard' ? 60 : 80))
const unit = computed(() => units.value[pos.value])
const showText = computed(() => !(props.mode === 'roleplay' && props.level === 'hard') || solvedHard.value)
const canGoOn = computed(() => (score.value?.pct ?? 0) >= pass.value || tries.value >= MAX_TRIES)
const progressPct = computed(() => (units.value.length ? Math.round((pos.value / units.value.length) * 100) : 0))
const spoken = computed(() => units.value.filter((u) => u.mine))
const weak = computed(() => spoken.value.filter((u) => u.best < pass.value))
const avg = computed(() => {
  const s = spoken.value.filter((u) => u.tries)
  return s.length ? Math.round(s.reduce((a, u) => a + u.best, 0) / s.length) : 0
})

/** Long sentences are hard to repeat in one breath: split them at , ; : into parts of 6+ words. */
function chunks(text: string): string[] {
  if (text.split(/\s+/).length <= LONG) return [text]
  const parts = text.split(/(?<=[,;:])\s+/)
  const out: string[] = []
  let cur = ''
  for (const p of parts) {
    cur = cur ? `${cur} ${p}` : p
    if (cur.split(/\s+/).length >= 6) {
      out.push(cur)
      cur = ''
    }
  }
  if (cur) {
    if (out.length && cur.split(/\s+/).length < 6) out[out.length - 1] += ' ' + cur
    else out.push(cur)
  }
  return out
}

/**
 * Units to go through. `only` (practise weak ones again) keeps just those of
 * the learner's sentences; partner lines stay in role-play for context.
 */
function build(only?: Set<string>) {
  const list: Unit[] = []
  for (const it of props.items) {
    const mine = props.mode === 'shadow' || it.speaker === props.role
    const texts = props.mode === 'shadow' ? chunks(it.text) : [it.text]
    for (const t of texts) {
      if (only && mine && !only.has(t)) continue
      list.push({ index: it.index, text: t, vi: it.vi, speaker: it.speaker, mine, best: 0, tries: 0 })
    }
  }
  units.value = list
}

const pause = (ms: number) => new Promise((r) => setTimeout(r, ms))

/** Play (unless hidden), then listen and score the current unit. */
async function run(opts: { slow?: boolean; playFirst?: boolean } = {}) {
  const t = ++token
  const u = unit.value
  if (!u) return finish()
  score.value = null
  emit('highlight', u.index)

  if (!u.mine) {
    phase.value = 'partner'
    try { await props.say(u.text, u.speaker, RATE_NORMAL) } catch { /* stopped */ }
    if (t !== token) return
    await pause(300)
    if (t === token) advance()
    return
  }

  // shadowing always plays the model first; role-play easy only on retries; hard only on request
  if (opts.playFirst ?? props.mode === 'shadow') {
    phase.value = 'playing'
    try { await props.say(u.text, u.speaker, opts.slow ? RATE_REPLAY : RATE_NORMAL) } catch { /* stopped */ }
    if (t !== token) return
    await pause(200)
    if (t !== token) return
  }

  phase.value = 'listening'
  const heard = await listen('en-US')
  if (t !== token) return
  tries.value++
  u.tries++
  const s = scoreSpeech(u.text, heard)
  score.value = s
  u.best = Math.max(u.best, s.pct)
  phase.value = 'result'
  if (props.level === 'hard' && (s.pct >= pass.value || tries.value >= MAX_TRIES)) solvedHard.value = true
  if (s.pct >= pass.value) {
    await pause(1200)
    if (t === token) advance()
  }
}

function tryAgain() {
  // hard role-play keeps the line hidden: just listen again (use "Hear the line" for help)
  run({ slow: true, playFirst: showText.value })
}

async function hear() {
  const u = unit.value
  if (!u) return
  const t = ++token
  phase.value = 'playing'
  try { await props.say(u.text, u.speaker, RATE_REPLAY) } catch { /* stopped */ }
  if (t === token) phase.value = 'result'
}

function stopListening() {
  stop() // whatever was heard is scored
}

function advance() {
  token++
  stop()
  tries.value = 0
  solvedHard.value = false
  if (pos.value < units.value.length - 1) {
    pos.value++
    run()
  } else {
    finish()
  }
}

function finish() {
  token++
  stop()
  finished.value = true
  emit('highlight', -1)
}

function restart(onlyWeak: boolean) {
  const only = onlyWeak ? new Set(weak.value.map((u) => u.text)) : undefined
  build(only)
  pos.value = 0
  tries.value = 0
  solvedHard.value = false
  finished.value = false
  run()
}

function close() {
  token++
  stop()
  if (typeof speechSynthesis !== 'undefined') speechSynthesis.cancel()
  emit('highlight', -1)
  emit('close')
}

onMounted(() => {
  build()
  run()
})
onUnmounted(() => {
  token++
  stop()
  if (typeof speechSynthesis !== 'undefined') speechSynthesis.cancel()
})

defineExpose({ listening })
</script>

<style scoped>
.practice { border-width: 2px; }
.words { line-height: 1.7; }
.words .ok { color: var(--bs-success); }
.words .miss { color: var(--bs-danger); text-decoration: underline wavy; text-underline-offset: 3px; }
.rec-dot {
  display: inline-block;
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: var(--bs-danger);
  margin-right: 4px;
  animation: blink 1s infinite;
}
@keyframes blink { 50% { opacity: 0.2; } }
</style>
