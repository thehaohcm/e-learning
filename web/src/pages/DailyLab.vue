<template>
  <div class="row justify-content-center">
    <div class="col-12">
      <div class="d-flex flex-wrap justify-content-between align-items-end mb-3 gap-2">
        <div>
          <h2 class="fw-bold mb-0"><i class="bi bi-terminal-plus"></i> {{ pageTitle }}</h2>
          <small class="text-muted">{{ pageBlurb }}</small>
        </div>
        <div v-if="streak > 0" class="badge text-bg-warning fs-6"><i class="bi bi-fire"></i> {{ streak }} ngày liên tiếp</div>
      </div>

      <div v-if="err" class="alert alert-danger py-2 small">{{ err }}</div>
      <div v-if="loading" class="text-center py-5"><div class="spinner-border text-primary"></div></div>

      <!-- no set -->
      <template v-else-if="!set">
        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <p class="mb-2">Chưa có bài lab hôm nay ({{ today }}). Tạo một bài để bắt đầu; có thể tạo thêm bài mới bất cứ lúc nào.</p>
            <ul class="small text-muted mb-3">
              <li>~10 bước: tình huống "nên làm gì tiếp theo?" (chọn) + lệnh chủ chốt (gõ) — dựa trên cheatsheet của bạn</li>
              <li>Chấm điểm theo bước; sau khi nộp có review từng bước + AI đánh giá + ghi chú</li>
            </ul>
            <button class="btn btn-primary" :disabled="generating" @click="openGen">
              <span v-if="generating" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-magic"></i>
              {{ generating ? 'AI đang dựng bài lab… (1–2 phút)' : 'Tạo bài lab hôm nay' }}
            </button>
          </div>
        </div>
      </template>

      <!-- drilling -->
      <template v-else-if="!set.submitted && cur < total">
        <div class="progress mb-2" style="height: 6px"><div class="progress-bar" :style="{ width: (100 * cur / total) + '%' }"></div></div>
        <div class="d-flex justify-content-between align-items-center mb-2">
          <span class="small text-muted">Bước {{ cur + 1 }} / {{ total }}</span>
          <span class="badge text-bg-light border">{{ step.type === 'choice' ? 'Tư duy' : 'Gõ lệnh' }}</span>
        </div>
        <div class="card border-0 shadow-sm mb-2">
          <div class="card-body">
            <p class="fw-semibold">{{ step.prompt }}</p>

            <!-- choice -->
            <template v-if="step.type === 'choice'">
              <div class="d-grid gap-2 mb-2">
                <button v-for="(opt, i) in (step.choices ?? [])" :key="i" class="btn text-start" :class="choiceClass(i)" :disabled="revealed" @click="pick(i)">
                  <b>{{ letters[i] }}.</b> {{ opt }}
                </button>
              </div>
            </template>

            <!-- command -->
            <template v-else>
              <div class="input-group mb-2">
                <span class="input-group-text font-monospace">$</span>
                <input ref="cmdInput" v-model="input" type="text" class="form-control font-monospace" placeholder="Gõ lệnh ở đây…"
                  :class="{ 'is-invalid': feedback === 'wrong' }" :readonly="revealed" @keydown.enter.prevent="onCheck" autocomplete="off" autocapitalize="off" spellcheck="false" />
              </div>
              <button v-if="!revealed" class="btn btn-primary" :disabled="!input.trim()" @click="onCheck">Kiểm tra</button>
            </template>

            <button v-if="step.hint && !revealed" class="btn btn-outline-secondary btn-sm ms-2" @click="showHint = !showHint">{{ showHint ? 'Ẩn gợi ý' : 'Xem gợi ý' }}</button>
            <div v-if="showHint && step.hint" class="alert alert-light border small mt-2 mb-0"><i class="bi bi-lightbulb text-warning"></i> {{ step.hint }}</div>

            <!-- feedback -->
            <div v-if="revealed" class="mt-3">
              <div class="alert py-2 small mb-2" :class="feedback === 'correct' ? 'alert-success' : 'alert-danger'">
                <b>{{ feedback === 'correct' ? 'Chính xác!' : 'Chưa đúng.' }}</b>
                <span v-if="step.type === 'command' && feedback !== 'correct'"> Lệnh đúng: <code>{{ modelCommand(step.patterns) }}</code>.</span>
                <span v-if="step.type === 'choice' && feedback !== 'correct'"> Đáp án đúng: <b>{{ (step.choices ?? [])[step.answer ?? 0] }}</b>.</span>
                <div v-if="step.explanation" class="mt-1">{{ step.explanation }}</div>
              </div>
              <div class="mb-2">
                <label class="form-label small mb-1"><i class="bi bi-pencil"></i> Ghi chú bước này</label>
                <textarea v-model="notes[cur]" class="form-control form-control-sm" rows="2" placeholder="Điều cần nhớ ở bước này…"></textarea>
              </div>
              <button class="btn btn-success" @click="next">
                {{ cur + 1 < total ? 'Bước tiếp theo' : 'Nộp bài' }} <i class="bi bi-arrow-right"></i>
              </button>
            </div>
          </div>
        </div>
      </template>

      <!-- result + review -->
      <template v-else-if="set.submitted">
        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <div class="d-flex flex-wrap align-items-center gap-3 mb-2">
              <div class="display-6 fw-bold">{{ set.score }}/{{ total }}</div>
              <div class="text-muted">{{ pct }}%</div>
            </div>
            <div class="d-flex flex-wrap gap-2 mt-2">
              <button class="btn btn-sm btn-outline-primary" :disabled="retaking" @click="retake">
                <span v-if="retaking" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-arrow-repeat"></i> Làm lại bài này
              </button>
              <button class="btn btn-sm btn-primary" :disabled="generating" @click="openGen">
                <span v-if="generating" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-plus-circle"></i>
                {{ generating ? 'Đang tạo…' : 'Tạo bài lab mới' }}
              </button>
              <button class="btn btn-sm btn-outline-secondary" @click="copyAll"><i class="bi" :class="copied ? 'bi-check2' : 'bi-clipboard'"></i> {{ copied ? 'Đã copy' : 'Copy' }}</button>
              <button class="btn btn-sm btn-outline-secondary" @click="downloadTxt"><i class="bi bi-download"></i> Tải .txt</button>
            </div>
          </div>
        </div>

        <div v-for="(st, i) in set.questions" :key="i" class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <div class="d-flex flex-wrap align-items-center gap-2 mb-2">
              <b>Bước {{ i + 1 }}</b>
              <span class="badge" :class="isRight(i) ? 'text-bg-success' : 'text-bg-danger'">{{ isRight(i) ? 'Đúng' : 'Sai' }}</span>
              <span class="badge text-bg-light border">{{ st.type === 'choice' ? 'Tư duy' : 'Gõ lệnh' }}</span>
            </div>
            <div class="mb-2 fw-semibold">{{ st.prompt }}</div>
            <div class="small mb-1"><span class="text-success"><i class="bi bi-check-circle-fill"></i> Đáp án đúng:</span>
              <code v-if="st.type === 'command'">{{ modelCommand(st.patterns) }}</code>
              <b v-else>{{ (st.choices ?? [])[st.answer ?? 0] }}</b>
            </div>
            <div class="small mb-1 text-muted">Bạn trả lời: <span :class="isRight(i) ? '' : 'text-danger'">{{ answerOf(i) || '(bỏ trống)' }}</span></div>
            <div v-if="st.explanation" class="bg-light rounded p-2 small mt-1">{{ st.explanation }}</div>
            <div class="mt-2">
              <label class="form-label small mb-1"><i class="bi bi-pencil"></i> Ghi chú của bạn</label>
              <textarea v-model="notes[i]" class="form-control form-control-sm" rows="2" placeholder="Ghi chú bước này (tự lưu khi ngừng gõ)…" @input="scheduleSaveNotes" @blur="saveNotes"></textarea>
            </div>
          </div>
        </div>

        <div class="text-center my-4">
          <button class="btn btn-warning" :disabled="aiReviewBusy" @click="runAiReview">
            <span v-if="aiReviewBusy" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-stars"></i>
            AI đánh giá bài lab (tiếng Việt)
          </button>
          <div v-if="aiReviewErr" class="text-danger small mt-2">{{ aiReviewErr }}</div>
        </div>
        <div v-if="aiReviewText" class="card border-0 shadow-sm mb-4">
          <div class="card-body">
            <div class="d-flex flex-wrap align-items-center gap-2 mb-2">
              <h6 class="fw-bold mb-0"><i class="bi bi-robot"></i> AI đánh giá</h6>
              <button class="btn btn-sm btn-outline-secondary ms-auto" @click="copyAiReview"><i class="bi" :class="aiCopied ? 'bi-check2' : 'bi-clipboard'"></i> {{ aiCopied ? 'Đã copy' : 'Copy nội dung' }}</button>
            </div>
            <div class="md" v-html="renderMarkdown(aiReviewText)"></div>
          </div>
        </div>
      </template>

      <!-- recent sets -->
      <div v-if="!loading && history.length" class="card border-0 shadow-sm mt-4">
        <div class="card-body py-2">
          <div class="small fw-bold mb-1">Các bài gần đây <span class="text-muted fw-normal">· bấm để xem lại</span></div>
          <div class="d-flex flex-wrap gap-2">
            <button v-for="h in history" :key="h.id" type="button" class="badge border btn-badge"
              :class="[h.id === (set && set.id) ? 'text-bg-primary' : h.submitted ? 'text-bg-light' : 'text-bg-secondary']"
              @click="openHistory(h.id)">
              {{ h.date.slice(5) }}<span v-if="h.seq > 1">#{{ h.seq }}</span> · {{ h.submitted ? h.score + '/' + h.total : '—' }}
              <i v-if="h.has_review" class="bi bi-stars text-warning ms-1"></i>
            </button>
          </div>
        </div>
      </div>

      <!-- generate options -->
      <div v-if="showGen" class="gen-overlay" @click.self="showGen = false">
        <div class="card border-0 shadow gen-card">
          <div class="card-body">
            <div class="d-flex align-items-center mb-3">
              <h6 class="fw-bold mb-0"><i class="bi bi-magic"></i> Tạo bài lab {{ isPentest ? 'pentest' : 'AWS' }}</h6>
              <button type="button" class="btn-close ms-auto" aria-label="Close" @click="showGen = false"></button>
            </div>
            <template v-if="isPentest">
              <label class="form-label small fw-semibold mb-1">Kiểu bài (giống challenge THM)</label>
              <div class="btn-group w-100 mb-2" role="group">
                <template v-for="t in tracks" :key="t.id">
                  <input :id="'trk-' + t.id" v-model="genTrack" type="radio" class="btn-check" :value="t.id" />
                  <label :for="'trk-' + t.id" class="btn btn-outline-primary btn-sm">{{ t.id }}</label>
                </template>
              </div>
              <div class="small text-muted mb-3">{{ trackDesc }}</div>
            </template>
            <label class="form-label small fw-semibold mb-1">Bước / kỹ năng muốn tập trung <span class="text-muted fw-normal">(không bắt buộc)</span></label>
            <select v-model="genChoice" class="form-select form-select-sm mb-2">
              <option value="">Tự động — đi đủ các bước của một bài lab</option>
              <option v-for="p in phaseOptions" :key="p" :value="p">{{ p }}</option>
              <option value="__custom__">Khác — nhập chủ đề cụ thể…</option>
            </select>
            <textarea v-if="genChoice === '__custom__'" v-model="genCustom" class="form-control form-control-sm" rows="2" :placeholder="focusPlaceholder"></textarea>
            <div class="form-text mb-3">Mỗi bài lab gồm nhiều bước, mỗi bước là một câu hỏi. Chọn một bước để luyện sâu, hoặc "Khác" để gõ chủ đề.</div>
            <div class="d-flex gap-2">
              <button class="btn btn-primary" :disabled="generating" @click="confirmGen">
                <span v-if="generating" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-magic"></i>
                {{ generating ? 'AI đang dựng… (1–2 phút)' : 'Tạo bài' }}
              </button>
              <button class="btn btn-outline-secondary" :disabled="generating" @click="showGen = false">Huỷ</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { api, errorMessage } from '../api'
import { renderMarkdown } from '../utils/markdown'
import { matchAny, modelCommand } from '../utils/labMatch'
import { subjectSlug } from '../composables/useSubject'

interface FlowStep { type: 'choice' | 'command'; prompt: string; choices?: string[]; answer?: number; patterns?: string[]; explanation?: string; hint?: string }
interface LabResult { correct?: boolean[] | null; notes?: string[] | null; answers?: string[] | null; by_domain?: Record<string, [number, number]> | null; ai_review?: string }
interface DSet { id: number; seq: number; date: string; questions: FlowStep[]; submitted: boolean; score: number; result: LabResult }
interface Day { id: number; date: string; seq: number; score: number; total: number; submitted: boolean; has_review: boolean }

const letters = ['A', 'B', 'C', 'D', 'E', 'F']
const today = new Date().toLocaleDateString('en-CA')

const isPentest = computed(() => subjectSlug.value === 'pentest')
const tracks = [
  { id: 'PT1', desc: 'PT1 — box mạng/host: recon → enum dịch vụ → foothold → leo quyền Linux/Windows (giống TryHackMe PT1).' },
  { id: 'WEB1', desc: 'WEB1 — mục tiêu web app: recon, auth/session, injection (SQLi/command/SSTI), IDOR, upload (giống TryHackMe WEB1).' },
  { id: 'CPTS', desc: 'CPTS — engagement đầy đủ theo blueprint HTB: recon → enum → web → Active Directory → leo quyền → lateral/pivoting → post-exploitation.' },
  { id: 'OPTS', desc: 'OPTS — kịch bản tổng hợp: kết hợp host + web + Active Directory như một phòng CTF/boot-to-root.' },
]
const showGen = ref(false)
const genFocus = ref('')
const genChoice = ref('') // '' = auto, a phase name, or '__custom__'
const genCustom = ref('')
const genTrack = ref('PT1')
const trackDesc = computed(() => tracks.find((t) => t.id === genTrack.value)?.desc ?? '')
// Methodology phases to drill, matched to the chosen track (each lab step = one question).
const phaseOptions = computed<string[]>(() => {
  if (!isPentest.value) return []
  switch (genTrack.value) {
    case 'WEB1':
      return ['Recon', 'Auth / session', 'Injection (SQLi/SSTI/cmd)', 'Access control / IDOR', 'File upload', 'Foothold']
    case 'CPTS':
    case 'OPTS':
      return ['Recon', 'Enumeration', 'Web', 'Active Directory', 'Privilege escalation', 'Lateral movement', 'Pivoting / tunnelling', 'Post-exploitation']
    default: // PT1
      return ['Recon', 'Enumeration', 'Foothold', 'Privilege escalation', 'Active Directory']
  }
})
const pageTitle = computed(() => (isPentest.value ? 'Pentest Daily Lab' : 'AWS Daily Lab'))
const pageBlurb = computed(() => isPentest.value
  ? 'Mỗi ngày một kịch bản methodology do AI dựng: xen câu tư duy (trắc nghiệm) và gõ lệnh thật, chấm điểm, ghi chú và AI đánh giá.'
  : 'Mỗi ngày một kịch bản vận hành AWS do AI dựng: xen câu tư duy và lệnh AWS CLI thật, chấm điểm, ghi chú và AI đánh giá.')
const focusPlaceholder = computed(() => isPentest.value
  ? 'VD: Windows privilege escalation, SMB enumeration, SQL injection…'
  : 'VD: sự cố S3 bị public, CloudFront 5xx, xoay vòng key KMS, scale ECS…')

const set = ref<DSet | null>(null)
const history = ref<Day[]>([])
const loading = ref(true)
const generating = ref(false)
const retaking = ref(false)
const submitting = ref(false)
const err = ref('')

const cur = ref(0)
const answers = ref<string[]>([])
const correct = ref<boolean[]>([])
const notes = ref<string[]>([])
const input = ref('')
const chosen = ref<number | null>(null)
const feedback = ref<'idle' | 'correct' | 'wrong'>('idle')
const revealed = ref(false)
const showHint = ref(false)
const startedAt = ref(0)
const cmdInput = ref<HTMLInputElement | null>(null)

const aiReviewText = ref('')
const aiReviewBusy = ref(false)
const aiReviewErr = ref('')
const aiCopied = ref(false)
const copied = ref(false)

const total = computed(() => set.value?.questions.length ?? 0)
const step = computed<FlowStep>(() => set.value?.questions[cur.value] ?? { type: 'choice', prompt: '', choices: [], answer: 0 })
const pct = computed(() => (total.value ? Math.round(100 * (set.value?.score ?? 0) / total.value) : 0))
const streak = computed(() => {
  const done = new Set(history.value.filter((h) => h.submitted).map((h) => h.date))
  let n = 0
  const d = new Date()
  if (!done.has(d.toLocaleDateString('en-CA'))) d.setDate(d.getDate() - 1)
  while (done.has(d.toLocaleDateString('en-CA'))) { n++; d.setDate(d.getDate() - 1) }
  return n
})

const isRight = (i: number) => !!set.value?.result.correct?.[i]
const answerOf = (i: number) => set.value?.result.answers?.[i] ?? ''

// ---- progress (localStorage) ----
const storeKey = () => `daily-lab-${set.value?.id ?? 0}`
function saveProgress() {
  if (!set.value || set.value.submitted) return
  try { localStorage.setItem(storeKey(), JSON.stringify({ cur: cur.value, answers: answers.value, correct: correct.value, notes: notes.value, startedAt: startedAt.value })) } catch { /* ignore */ }
}
function restoreProgress() {
  const n = total.value
  answers.value = Array.from({ length: n }, () => '')
  correct.value = Array.from({ length: n }, () => false)
  notes.value = Array.from({ length: n }, () => '')
  cur.value = 0
  startedAt.value = Date.now()
  try {
    const raw = localStorage.getItem(storeKey())
    if (!raw) return
    const p = JSON.parse(raw)
    if (Array.isArray(p.answers)) p.answers.forEach((x: unknown, i: number) => { if (i < n) answers.value[i] = typeof x === 'string' ? x : '' })
    if (Array.isArray(p.correct)) p.correct.forEach((x: unknown, i: number) => { if (i < n) correct.value[i] = !!x })
    if (Array.isArray(p.notes)) p.notes.forEach((x: unknown, i: number) => { if (i < n) notes.value[i] = typeof x === 'string' ? x : '' })
    if (typeof p.cur === 'number' && p.cur >= 0 && p.cur <= n) cur.value = p.cur
    if (typeof p.startedAt === 'number' && p.startedAt > 0) startedAt.value = p.startedAt
  } catch { /* ignore */ }
}
watch([answers, correct, notes, cur], saveProgress, { deep: true })

function adopt(s: DSet | null) {
  set.value = s
  aiReviewErr.value = ''
  aiReviewText.value = s?.result?.ai_review ?? ''
  feedback.value = 'idle'; revealed.value = false; chosen.value = null; input.value = ''; showHint.value = false
  if (s && !s.submitted) restoreProgress()
  else if (s) notes.value = Array.from({ length: s.questions.length }, (_, i) => s.result.notes?.[i] ?? '')
}

async function loadHistory() {
  try { const { data } = await api.get('/pmp-daily/history', { params: { subject: subjectSlug.value } }); history.value = data.days ?? [] } catch { /* optional */ }
}
async function load() {
  loading.value = true; err.value = ''
  try {
    const { data } = await api.get('/pmp-daily', { params: { subject: subjectSlug.value, date: today } })
    adopt(data.set ?? null)
    await loadHistory()
  } catch (e) { err.value = errorMessage(e) } finally { loading.value = false }
}
function openGen() { genChoice.value = ''; genCustom.value = ''; genFocus.value = ''; showGen.value = true }
async function confirmGen() {
  genFocus.value = genChoice.value === '__custom__' ? genCustom.value.trim() : genChoice.value
  await generate()
  if (!err.value) showGen.value = false
}
async function generate() {
  generating.value = true; err.value = ''
  try {
    const payload: Record<string, unknown> = { date: today, subject: subjectSlug.value, focus: genFocus.value.trim() }
    if (isPentest.value) payload.track = genTrack.value
    const { data } = await api.post('/pmp-daily/generate', payload, { timeout: 180000 })
    adopt(data.set); await loadHistory(); window.scrollTo({ top: 0 })
  } catch (e) { err.value = errorMessage(e) } finally { generating.value = false }
}
async function retake() {
  if (!set.value) return
  retaking.value = true; err.value = ''
  try { const { data } = await api.post('/pmp-daily/retake', { id: set.value.id, date: today }); adopt(data.set); await loadHistory(); window.scrollTo({ top: 0 }) }
  catch (e) { err.value = errorMessage(e) } finally { retaking.value = false }
}
async function openHistory(id: number) {
  if (set.value && set.value.id === id) return
  loading.value = true; err.value = ''
  try { const { data } = await api.get('/pmp-daily', { params: { id } }); adopt(data.set ?? null); window.scrollTo({ top: 0 }) }
  catch (e) { err.value = errorMessage(e) } finally { loading.value = false }
}

// ---- drilling ----
function choiceClass(i: number): string {
  if (!revealed.value) return chosen.value === i ? 'btn-primary' : 'btn-outline-primary'
  if (i === (step.value.answer ?? 0)) return 'btn-success'
  if (i === chosen.value) return 'btn-danger'
  return 'btn-outline-secondary'
}
function pick(i: number) {
  if (revealed.value) return
  chosen.value = i
  const ok = i === (step.value.answer ?? 0)
  correct.value[cur.value] = ok
  answers.value[cur.value] = (step.value.choices ?? [])[i] ?? ''
  feedback.value = ok ? 'correct' : 'wrong'
  revealed.value = true
}
function onCheck() {
  if (revealed.value || !input.value.trim()) return
  const ok = matchAny(step.value.patterns, input.value)
  correct.value[cur.value] = ok
  answers.value[cur.value] = input.value.trim()
  feedback.value = ok ? 'correct' : 'wrong'
  revealed.value = true
}
function next() {
  if (cur.value + 1 < total.value) {
    cur.value++
    feedback.value = 'idle'; revealed.value = false; chosen.value = null; input.value = ''; showHint.value = false
    nextTick(() => cmdInput.value?.focus())
  } else {
    submit()
  }
}
async function submit() {
  if (!set.value) return
  submitting.value = true; err.value = ''
  try {
    const seconds = startedAt.value ? Math.floor((Date.now() - startedAt.value) / 1000) : 0
    const { data } = await api.post('/pmp-daily/submit', { id: set.value.id, correct: correct.value, answers: answers.value, notes: notes.value, seconds })
    try { localStorage.removeItem(storeKey()) } catch { /* ignore */ }
    adopt(data.set); await loadHistory(); window.scrollTo({ top: 0 })
  } catch (e) { err.value = errorMessage(e) } finally { submitting.value = false }
}

// ---- notes autosave (review) ----
let notesTimer: ReturnType<typeof setTimeout> | null = null
function scheduleSaveNotes() { if (notesTimer) clearTimeout(notesTimer); notesTimer = setTimeout(() => saveNotes(), 900) }
async function saveNotes() {
  if (notesTimer) { clearTimeout(notesTimer); notesTimer = null }
  if (!set.value) return
  try { await api.post('/pmp-daily/notes', { id: set.value.id, notes: notes.value }) } catch { /* best effort */ }
}

// ---- AI review ----
function correctText(st: FlowStep): string {
  return st.type === 'command' ? (modelCommand(st.patterns) || st.explanation || '') : ((st.choices ?? [])[st.answer ?? 0] ?? '')
}
async function runAiReview() {
  const sv = set.value
  if (!sv) return
  const build = (st: FlowStep, i: number, right: boolean) => ({
    stem: st.prompt,
    correct: [correctText(st)].filter(Boolean),
    picked: [answerOf(i)].filter(Boolean),
    explanation: st.explanation ?? '',
    section: '',
    terms: [] as string[],
    note: notes.value[i] ?? '',
    right,
  })
  const wrong = [
    ...sv.questions.map((st, i) => ({ st, i })).filter(({ i }) => !isRight(i)).map(({ st, i }) => build(st, i, false)),
    ...sv.questions.map((st, i) => ({ st, i })).filter(({ i }) => isRight(i) && (notes.value[i] ?? '').trim()).map(({ st, i }) => build(st, i, true)),
  ]
  aiReviewBusy.value = true; aiReviewErr.value = ''
  try {
    const { data } = await api.post('/pmp-daily/ai-review', { id: sv.id, wrong }, { timeout: 120000 })
    aiReviewText.value = data.review ?? ''
    await loadHistory()
  } catch (e) { aiReviewErr.value = errorMessage(e) } finally { aiReviewBusy.value = false }
}

// ---- copy / download ----
function allText(): string {
  const sv = set.value
  if (!sv) return ''
  const lines: string[] = []
  sv.questions.forEach((st, i) => {
    lines.push(`Bước ${i + 1} (${st.type === 'choice' ? 'Tư duy' : 'Gõ lệnh'})${isRight(i) ? '' : ' — SAI'}`)
    lines.push(st.prompt)
    lines.push(`Đáp án đúng: ${correctText(st)}`)
    lines.push(`Bạn trả lời: ${answerOf(i) || '(bỏ trống)'}`)
    if (st.explanation) lines.push(`Giải thích: ${st.explanation}`)
    if ((notes.value[i] ?? '').trim()) lines.push(`Ghi chú của bạn: ${notes.value[i].trim()}`)
    lines.push('')
  })
  return lines.join('\n').trim()
}
async function copyAll() { try { await navigator.clipboard.writeText(allText()); copied.value = true; setTimeout(() => { copied.value = false }, 1800) } catch { /* ignore */ } }
function downloadTxt() {
  const blob = new Blob([allText()], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a'); a.href = url; a.download = `daily-lab-${set.value?.date ?? today}.txt`
  document.body.appendChild(a); a.click(); a.remove(); setTimeout(() => URL.revokeObjectURL(url), 1000)
}
async function copyAiReview() { try { await navigator.clipboard.writeText(aiReviewText.value); aiCopied.value = true; setTimeout(() => { aiCopied.value = false }, 1800) } catch { /* ignore */ } }

onMounted(load)
</script>

<style scoped>
.btn-badge { cursor: pointer; border: 1px solid #cbd5e1; }
.btn-badge:hover { filter: brightness(0.97); }
.gen-overlay { position: fixed; inset: 0; background: rgba(0,0,0,.45); display: flex; align-items: center; justify-content: center; z-index: 1060; padding: 1rem; }
.gen-card { width: 100%; max-width: 460px; }
</style>
