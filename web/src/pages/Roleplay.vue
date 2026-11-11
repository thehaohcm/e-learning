<template>
  <div>
      <div class="mb-3">
        <h2 class="fw-bold mb-0"><i class="bi bi-chat-quote"></i> Roleplay: PM &amp; client</h2>
        <small class="text-muted">Bạn là Project Manager, AI đóng vai client khó tính. Luyện nghe và nói tiếng Anh trong tình huống PMP thật.</small>
      </div>

      <div v-if="err" class="alert alert-danger py-2 small d-flex justify-content-between align-items-center">
        <span>{{ err }}</span>
        <button v-if="retryable" class="btn btn-sm btn-outline-danger" @click="retry">Thử lại</button>
      </div>

      <div v-if="loading" class="text-center py-5"><div class="spinner-border text-primary"></div></div>

      <!-- 1. choose a scenario -->
      <template v-else-if="!started">
        <div class="row g-2 mb-3">
          <div v-for="s in scenarios" :key="s.id" class="col-12 col-md-6">
            <div class="card h-100 border-2 sc-card" :class="scenarioId === s.id ? 'border-primary' : 'border-0 shadow-sm'" @click="scenarioId = s.id">
              <div class="card-body py-2">
                <div class="fw-bold">{{ s.title }}</div>
                <div class="small text-muted">{{ s.topics }}</div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="scenario" class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <h6 class="fw-bold">{{ scenario.title }}</h6>
            <p class="mb-1">{{ scenario.setup }}</p>
            <p class="small text-muted mb-2">{{ scenario.setup_vi }}</p>
            <p class="small mb-0"><b>Mục tiêu của PM:</b> {{ scenario.goal }}</p>
          </div>
        </div>

        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <label class="form-label small mb-1">Kiểu client</label>
            <select v-model="personaId" class="form-select mb-1">
              <option v-for="p in personas" :key="p.id" :value="p.id">{{ p.label }}</option>
            </select>
            <div class="small text-muted mb-3">{{ personaDesc }}</div>

            <div class="form-check form-switch mb-1">
              <input id="rp-voice" v-model="voiceOn" class="form-check-input" type="checkbox" />
              <label class="form-check-label" for="rp-voice">Client nói thành tiếng (nghe)</label>
            </div>
            <div class="form-check form-switch mb-1">
              <input id="rp-hide" v-model="hideText" class="form-check-input" type="checkbox" :disabled="!voiceOn" />
              <label class="form-check-label" for="rp-hide">Ẩn chữ của client, nghe trước rồi mới bấm xem (luyện nghe)</label>
            </div>
            <div class="form-check form-switch mb-2">
              <input id="rp-auto" v-model="autoSend" class="form-check-input" type="checkbox" />
              <label class="form-check-label" for="rp-auto">Tự gửi ngay sau khi nói (giống hội thoại thật, không sửa được)</label>
            </div>
            <div class="d-flex align-items-center gap-2 mb-3">
              <label class="small text-muted mb-0">Tốc độ giọng client</label>
              <select v-model.number="rate" class="form-select form-select-sm w-auto">
                <option :value="1">Bình thường</option>
                <option :value="0.8">Chậm</option>
                <option :value="0.65">Rất chậm</option>
              </select>
            </div>
            <p v-if="!sttSupported" class="small text-warning mb-2">
              <i class="bi bi-exclamation-triangle"></i> Trình duyệt này không có nhận dạng giọng nói (dùng Chrome hoặc Edge để nói). Bạn vẫn có thể gõ câu trả lời.
            </p>
            <button class="btn btn-primary" :disabled="!scenario" @click="start"><i class="bi bi-play-fill"></i> Bắt đầu cuộc họp</button>
          </div>
        </div>
      </template>

      <!-- 2. the conversation -->
      <template v-else>
        <div class="card border-0 shadow-sm mb-2">
          <div class="card-body py-2 small">
            <b>{{ scenario?.title }}</b> · {{ personaLabel }}
            <details class="mt-1">
              <summary class="text-muted">Tình huống &amp; mục tiêu</summary>
              <p class="mb-1 mt-1">{{ scenario?.setup }}</p>
              <p class="mb-0"><b>Goal:</b> {{ scenario?.goal }}</p>
            </details>
          </div>
        </div>

        <div ref="chatBox" class="chat mb-2">
          <div v-for="(t, i) in turns" :key="i" class="mb-2 d-flex" :class="t.role === 'pm' ? 'justify-content-end' : ''">
            <div class="bubble" :class="t.role === 'pm' ? 'you' : 'ai'">
              <div class="small fw-bold opacity-75 mb-1">{{ t.role === 'pm' ? 'You (PM)' : 'Client' }}</div>
              <template v-if="t.role === 'client' && hideText && !revealed.has(i)">
                <button class="btn btn-sm btn-outline-secondary" @click="reveal(i)"><i class="bi bi-eye"></i> Hiện chữ</button>
              </template>
              <div v-else>{{ t.text }}</div>
              <div v-if="t.role === 'client'" class="mt-1">
                <button class="btn btn-sm btn-link p-0 me-2" title="Nghe lại" @click="say(t.text, RATE_REPLAY)"><i class="bi bi-volume-up"></i></button>
                <button class="btn btn-sm btn-link p-0" title="Nghe chậm" @click="say(t.text, RATE_SLOW)">🐢</button>
              </div>
            </div>
          </div>
          <div v-if="waiting" class="text-muted small"><span class="spinner-border spinner-border-sm me-1"></span>Client is thinking…</div>
        </div>

        <!-- composer -->
        <div v-if="!feedback" class="card border-0 shadow-sm mb-3">
          <div class="card-body py-2">
            <div v-if="listening" class="small text-danger mb-1"><i class="bi bi-mic-fill"></i> Đang nghe… {{ liveText }}</div>
            <div v-if="sttError" class="small text-danger mb-1">{{ sttError }}</div>
            <form class="d-flex gap-2" @submit.prevent="sendTyped">
              <button v-if="sttSupported" type="button" class="btn" :class="listening ? 'btn-danger' : 'btn-outline-danger'"
                :disabled="waiting || reviewing" :title="listening ? 'Dừng' : 'Nói'" @click="toggleMic">
                <i class="bi" :class="listening ? 'bi-stop-fill' : 'bi-mic-fill'"></i>
              </button>
              <input v-model="input" class="form-control" placeholder="Nói bằng mic hoặc gõ câu trả lời bằng tiếng Anh…" maxlength="700"
                :disabled="waiting || reviewing" />
              <button class="btn btn-primary" :disabled="waiting || reviewing || !input.trim()"><i class="bi bi-send"></i></button>
            </form>
            <div class="d-flex flex-wrap justify-content-between align-items-center mt-2 gap-2">
              <small class="text-muted">
                {{ pmTurns }} lượt của bạn<span v-if="pmTurns >= 10"> · đủ dài rồi, nên kết thúc để xem nhận xét</span>
              </small>
              <div class="d-flex gap-2">
                <button class="btn btn-sm btn-outline-secondary" :disabled="reviewing" @click="resetAll">Thoát</button>
                <button class="btn btn-sm btn-success" :disabled="pmTurns === 0 || waiting || reviewing" @click="finish">
                  <span v-if="reviewing" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-clipboard-check"></i>
                  Kết thúc &amp; xem nhận xét
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 3. feedback -->
        <div v-else class="card border-0 shadow-sm mb-4">
          <div class="card-body">
            <h5 class="fw-bold"><i class="bi bi-clipboard-check"></i> Nhận xét</h5>
            <p>{{ feedback.summary_vi }}</p>

            <div class="row g-3 mb-3">
              <div class="col-6">
                <div class="small text-muted">Xử lý tình huống (PMI)</div>
                <div class="fs-5 stars">{{ stars(feedback.pmi_score) }}</div>
              </div>
              <div class="col-6">
                <div class="small text-muted">Tiếng Anh</div>
                <div class="fs-5 stars">{{ stars(feedback.english_score) }}</div>
              </div>
            </div>

            <div v-if="feedback.pmi_good_vi?.length" class="mb-2">
              <div class="fw-bold small text-success">Làm tốt</div>
              <ul class="mb-1"><li v-for="(x, i) in feedback.pmi_good_vi" :key="'g' + i">{{ x }}</li></ul>
            </div>
            <div v-if="feedback.pmi_improve_vi?.length" class="mb-3">
              <div class="fw-bold small text-warning">Nên làm khác đi (theo PMI)</div>
              <ul class="mb-1"><li v-for="(x, i) in feedback.pmi_improve_vi" :key="'i' + i">{{ x }}</li></ul>
            </div>

            <div v-if="feedback.corrections?.length" class="mb-3">
              <div class="fw-bold small mb-1">Sửa tiếng Anh</div>
              <div v-for="(c, i) in feedback.corrections" :key="'c' + i" class="border rounded p-2 mb-2">
                <div class="text-danger small"><i class="bi bi-x-circle"></i> {{ c.you }}</div>
                <div class="text-success">
                  <i class="bi bi-check-circle"></i> {{ c.better }}
                  <button class="btn btn-sm btn-link p-0 ms-1" title="Nghe" @click="say(c.better, RATE_REPLAY)"><i class="bi bi-volume-up"></i></button>
                </div>
                <div class="small text-muted">{{ c.why_vi }}</div>
              </div>
            </div>

            <div v-if="feedback.model_answer" class="mb-3">
              <div class="fw-bold small mb-1">Ví dụ câu trả lời tốt</div>
              <div class="bg-light rounded p-2">
                {{ feedback.model_answer }}
                <button class="btn btn-sm btn-link p-0 ms-1" title="Nghe" @click="say(feedback.model_answer, RATE_REPLAY)"><i class="bi bi-volume-up"></i></button>
              </div>
            </div>

            <div v-if="feedback.phrases?.length" class="mb-3">
              <div class="fw-bold small mb-1">Cụm từ nên nhớ</div>
              <ul class="list-unstyled mb-0">
                <li v-for="(p, i) in feedback.phrases" :key="'p' + i" class="mb-1">
                  <b>{{ p.en }}</b> <span class="text-muted">· {{ p.vi }}</span>
                  <button class="btn btn-sm btn-link p-0 ms-1" title="Nghe" @click="say(p.en, RATE_REPLAY)"><i class="bi bi-volume-up"></i></button>
                </li>
              </ul>
            </div>

            <p v-if="feedback.next_step_vi" class="small"><b>Luyện tiếp:</b> {{ feedback.next_step_vi }}</p>

            <div class="d-flex gap-2">
              <button class="btn btn-primary" @click="start"><i class="bi bi-arrow-repeat"></i> Làm lại kịch bản này</button>
              <button class="btn btn-outline-secondary" @click="resetAll">Chọn kịch bản khác</button>
            </div>
          </div>
        </div>
      </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { api, errorMessage } from '../api'
import { useSpeechRecognition } from '../composables/useSpeechRecognition'
import { RATE_REPLAY, RATE_SLOW, useSpeechSynthesis } from '../composables/useSpeechSynthesis'

// PMP roleplay: the learner is the project manager, the AI plays a difficult
// client. Speaking uses the browser's speech recognition (Chrome/Edge), the
// client is read aloud with the browser's text-to-speech. Nothing is stored:
// the conversation lives in this page and is sent to the server each turn.

interface Scenario { id: string; title: string; setup: string; setup_vi: string; goal: string; topics: string }
interface Persona { id: string; label: string; desc: string }
interface Turn { role: 'client' | 'pm'; text: string }
interface Correction { you: string; better: string; why_vi: string }
interface Phrase { en: string; vi: string }
interface Feedback {
  summary_vi: string
  pmi_score: number
  pmi_good_vi: string[]
  pmi_improve_vi: string[]
  english_score: number
  corrections: Correction[]
  phrases: Phrase[]
  model_answer: string
  next_step_vi: string
}

const scenarios = ref<Scenario[]>([])
const personas = ref<Persona[]>([])
const loading = ref(true)
const err = ref('')
const retryable = ref<null | (() => void)>(null)

const scenarioId = ref('')
const personaId = ref('firm')
const voiceOn = ref(true)
const hideText = ref(false)
const autoSend = ref(false)
const rate = ref(1)

const started = ref(false)
const turns = ref<Turn[]>([])
const revealed = ref<Set<number>>(new Set())
const input = ref('')
const waiting = ref(false)
const reviewing = ref(false)
const feedback = ref<Feedback | null>(null)
const chatBox = ref<HTMLElement | null>(null)

const tts = useSpeechSynthesis()
const stt = useSpeechRecognition()
const sttSupported = stt.supported
const listening = stt.listening
const liveText = stt.transcript
const sttError = stt.error

const scenario = computed(() => scenarios.value.find((s) => s.id === scenarioId.value) ?? null)
const personaDesc = computed(() => personas.value.find((p) => p.id === personaId.value)?.desc ?? '')
const personaLabel = computed(() => personas.value.find((p) => p.id === personaId.value)?.label ?? '')
const pmTurns = computed(() => turns.value.filter((t) => t.role === 'pm').length)

function stars(n: number): string {
  const k = Math.max(0, Math.min(5, Math.round(n || 0)))
  return '★'.repeat(k) + '☆'.repeat(5 - k)
}

async function load() {
  loading.value = true
  err.value = ''
  try {
    const { data } = await api.get('/roleplay/scenarios')
    scenarios.value = data.scenarios ?? []
    personas.value = data.personas ?? []
    if (!scenarioId.value && scenarios.value.length) scenarioId.value = scenarios.value[0].id
  } catch (e) {
    err.value = errorMessage(e)
    retryable.value = load
  } finally {
    loading.value = false
  }
}

async function scrollDown() {
  await nextTick()
  if (chatBox.value) chatBox.value.scrollTop = chatBox.value.scrollHeight
}

function history() {
  return turns.value.map((t) => ({ role: t.role, text: t.text }))
}

async function say(text: string, r: number = rate.value) {
  try { await tts.speakEnglish(text, r) } catch { /* speech unavailable or interrupted */ }
}

function reveal(i: number) {
  const n = new Set(revealed.value)
  n.add(i)
  revealed.value = n
}

/** Ask the AI for the client's next line (the opening line when there are no turns yet). */
async function clientTurn() {
  waiting.value = true
  err.value = ''
  retryable.value = null
  try {
    const { data } = await api.post('/roleplay/turn',
      { scenario_id: scenarioId.value, persona: personaId.value, history: history() },
      { timeout: 100000 })
    const text: string = data.reply ?? ''
    turns.value.push({ role: 'client', text })
    await scrollDown()
    if (voiceOn.value) say(text)
  } catch (e) {
    err.value = errorMessage(e)
    retryable.value = clientTurn
  } finally {
    waiting.value = false
  }
}

async function start() {
  if (!scenarioId.value) return
  tts.stop()
  turns.value = []
  revealed.value = new Set()
  feedback.value = null
  input.value = ''
  started.value = true
  await clientTurn()
}

async function sendText(text: string) {
  const t = text.trim()
  if (!t || waiting.value) return
  turns.value.push({ role: 'pm', text: t })
  input.value = ''
  await scrollDown()
  await clientTurn()
}

function sendTyped() {
  return sendText(input.value)
}

async function toggleMic() {
  if (listening.value) { stt.stop(); return }
  const alts = await stt.listen('en-US')
  const best = alts[0]
  if (!best) return
  if (autoSend.value) await sendText(best)
  else input.value = best
}

async function finish() {
  if (pmTurns.value === 0) return
  tts.stop()
  reviewing.value = true
  err.value = ''
  retryable.value = null
  try {
    const { data } = await api.post('/roleplay/feedback',
      { scenario_id: scenarioId.value, persona: personaId.value, history: history() },
      { timeout: 160000 })
    feedback.value = data.feedback
    await scrollDown()
  } catch (e) {
    err.value = errorMessage(e)
    retryable.value = finish
  } finally {
    reviewing.value = false
  }
}

function resetAll() {
  tts.stop()
  stt.stop()
  started.value = false
  turns.value = []
  feedback.value = null
  err.value = ''
  retryable.value = null
}

function retry() {
  const f = retryable.value
  if (f) f()
}

onMounted(load)
onUnmounted(() => tts.stop())
</script>

<style scoped>
.sc-card { cursor: pointer; }
.chat { max-height: 55vh; overflow-y: auto; padding: 0.25rem; }
.bubble { max-width: 85%; padding: 0.5rem 0.75rem; border-radius: 12px; }
.bubble.you { background: #4f46e5; color: #fff; border-bottom-right-radius: 3px; }
.bubble.ai { background: #f1f5f9; color: #1e293b; border-bottom-left-radius: 3px; }
.bubble.you .btn-link { color: #fff; }
.stars { color: #f59e0b; letter-spacing: 2px; }
</style>
