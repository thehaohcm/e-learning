<template>
  <div class="row justify-content-center">
    <div class="col-12">
      <div class="d-flex flex-wrap justify-content-between align-items-end mb-3 gap-2">
        <div>
          <h2 class="fw-bold mb-0"><i class="bi bi-calendar-check"></i> {{ subjectTitle }} Daily 20</h2>
          <small class="text-muted">{{ subjectBlurb }}</small>
        </div>
        <div v-if="streak > 0" class="badge text-bg-warning fs-6"><i class="bi bi-fire"></i> {{ streak }} ngày liên tiếp</div>
      </div>

      <div v-if="err" class="alert alert-danger py-2 small">{{ err }}</div>
      <div v-if="loading" class="text-center py-5"><div class="spinner-border text-primary"></div></div>

      <!-- no set loaded -->
      <template v-else-if="!set">
        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <p class="mb-2">Chưa có bộ nào hôm nay ({{ today }}). Tạo một bộ 20 câu để bắt đầu; bạn có thể tạo thêm bộ mới bất cứ lúc nào.</p>
            <ul v-if="isPmp" class="small text-muted mb-3">
              <li>7 câu People, 8 câu Process, 5 câu Business Environment, trộn lẫn như đề thật</li>
              <li>Khoảng 40% predictive, 60% agile/hybrid; có câu chọn 2 đáp án và câu tính toán</li>
              <li>Đồng hồ ~27 phút (80 giây mỗi câu); sau khi nộp có phân tích từng đáp án</li>
            </ul>
            <ul v-else class="small text-muted mb-3">
              <li>20 câu trắc nghiệm tiếng Anh (TOEIC/IELTS) từ chính các từ bạn đang học</li>
              <li>Điền từ, chọn nghĩa gần nhất, hoặc chọn câu dùng từ đúng</li>
              <li>Sau khi nộp có giải thích từng đáp án và nghĩa tiếng Việt</li>
            </ul>
            <button class="btn btn-primary" :disabled="generating" @click="generate">
              <span v-if="generating" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-magic"></i>
              {{ generating ? 'AI đang viết 20 câu… (có thể mất 1–2 phút)' : 'Tạo bộ 20 câu' }}
            </button>
          </div>
        </div>
      </template>

      <!-- taking the quiz -->
      <template v-else-if="!set.submitted && q">
        <div class="card border-0 shadow-sm mb-2">
          <div class="card-body py-2 d-flex flex-wrap align-items-center gap-2">
            <b>Câu {{ cur + 1 }}/{{ set.questions.length }}</b>
            <span class="badge text-bg-light border">{{ q.domain }}</span>
            <span class="badge text-bg-light border">{{ q.approach }}</span>
            <span class="ms-auto fw-bold" :class="remaining < 0 ? 'text-danger' : ''">
              <i class="bi bi-stopwatch"></i> {{ clock }}
            </span>
          </div>
        </div>

        <div class="card border-0 shadow-sm mb-2">
          <div class="card-body">
            <div class="mb-3 stem">
              {{ q.stem }}
              <button class="btn btn-sm btn-link p-0 ms-1" title="Nghe câu hỏi" @click="say(q.stem)"><i class="bi bi-volume-up"></i></button>
            </div>
            <div v-if="q.type === 'multi'" class="small text-primary mb-2"><i class="bi bi-info-circle"></i> Chọn <b>2</b> đáp án.</div>
            <div v-for="(o, oi) in q.options" :key="oi" class="form-check opt mb-2" :class="isPicked(cur, oi) ? 'picked' : ''" @click="choose(cur, oi)">
              <input class="form-check-input" :type="q.type === 'multi' ? 'checkbox' : 'radio'" :checked="isPicked(cur, oi)" @click.stop="choose(cur, oi)" />
              <label class="form-check-label"><b>{{ letters[oi] }}.</b> {{ o.text }}</label>
            </div>
          </div>
        </div>

        <div class="d-flex flex-wrap gap-2 mb-3">
          <button class="btn btn-outline-secondary" :disabled="cur === 0" @click="cur--"><i class="bi bi-chevron-left"></i></button>
          <button class="btn btn-outline-secondary" :disabled="cur === set.questions.length - 1" @click="cur++"><i class="bi bi-chevron-right"></i></button>
          <button class="btn" :class="flagged.includes(cur) ? 'btn-warning' : 'btn-outline-warning'" @click="toggleFlag(cur)">
            <i class="bi bi-flag"></i> Đánh dấu xem lại
          </button>
          <button class="btn btn-success ms-auto" :disabled="submitting" @click="confirmSubmit = true">
            <i class="bi bi-send-check"></i> Nộp bài
          </button>
        </div>

        <div v-if="confirmSubmit" class="alert alert-warning py-2 d-flex flex-wrap align-items-center gap-2">
          <span>Còn {{ unanswered }} câu chưa trả lời{{ flagged.length ? ', ' + flagged.length + ' câu đánh dấu' : '' }}. Nộp bài?</span>
          <button class="btn btn-sm btn-success" :disabled="submitting" @click="submit">
            <span v-if="submitting" class="spinner-border spinner-border-sm me-1"></span>Nộp
          </button>
          <button class="btn btn-sm btn-link" @click="confirmSubmit = false">Làm tiếp</button>
        </div>

        <div class="card border-0 shadow-sm">
          <div class="card-body py-2 d-flex flex-wrap gap-1">
            <button v-for="(_, i) in set.questions" :key="i" class="btn btn-sm pal"
              :class="[i === cur ? 'btn-primary' : answered(i) ? 'btn-secondary' : 'btn-outline-secondary', flagged.includes(i) ? 'flagged' : '']"
              @click="cur = i">{{ i + 1 }}</button>
          </div>
        </div>
      </template>

      <!-- result + review -->
      <template v-else-if="set.submitted">
        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <div class="d-flex flex-wrap align-items-center gap-3 mb-2">
              <div class="display-6 fw-bold">{{ set.score }}/{{ set.questions.length }}</div>
              <div class="text-muted">{{ pct }}% · {{ minutes(set.result.seconds) }}</div>
            </div>
            <div v-for="(v, d) in (set.result.by_domain || {})" :key="d" class="mb-1">
              <div class="small d-flex justify-content-between"><span>{{ d }}</span><span>{{ v[0] }}/{{ v[1] }}</span></div>
              <div class="progress" style="height: 6px"><div class="progress-bar" :style="{ width: (v[1] ? (100 * v[0]) / v[1] : 0) + '%' }"></div></div>
            </div>
            <div class="d-flex flex-wrap gap-2 mt-3">
              <button class="btn btn-sm btn-outline-primary" :disabled="retaking" @click="retake">
                <span v-if="retaking" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-arrow-repeat"></i>
                Làm lại bộ này
              </button>
              <button class="btn btn-sm btn-primary" :disabled="generating" @click="generate">
                <span v-if="generating" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-plus-circle"></i>
                {{ generating ? 'Đang tạo…' : 'Tạo bộ 20 câu mới' }}
              </button>
            </div>
          </div>
        </div>

        <!-- audio notes for this exam -->
        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <div class="d-flex flex-wrap align-items-center gap-2 mb-2">
              <h6 class="fw-bold mb-0"><i class="bi bi-music-note-beamed"></i> Audio ôn tập (của bạn)</h6>
              <small class="text-muted">mp3 / m4a · tối đa 25MB/file</small>
              <label class="btn btn-sm btn-outline-primary ms-auto mb-0">
                <span v-if="audioBusy" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-upload"></i>
                Tải audio lên
                <input type="file" accept="audio/*,.mp3,.m4a,.aac,.wav,.ogg" multiple class="d-none" :disabled="audioBusy" @change="uploadAudio" />
              </label>
            </div>
            <div v-if="audioErr" class="alert alert-danger py-2 small mb-2">{{ audioErr }}</div>
            <p v-if="!audioFiles.length" class="small text-muted mb-0">Chưa có file nào. Tải lên bản tóm tắt NotebookLM hay ghi âm của bạn để nghe lại.</p>
            <div v-for="a in audioFiles" :key="a.id" class="d-flex flex-wrap align-items-center gap-2 py-2 border-top">
              <div class="flex-grow-1">
                <div class="small fw-semibold text-truncate" style="max-width: 420px">{{ a.name }}</div>
                <div class="text-muted" style="font-size: 0.75rem">{{ sizeMB(a.bytes) }}</div>
              </div>
              <audio v-if="audioUrls[a.id]" :src="audioUrls[a.id]" controls preload="none" style="height: 36px"></audio>
              <button v-else class="btn btn-sm btn-outline-secondary" :disabled="audioLoading[a.id]" @click="loadBlob(a.id)">
                <span v-if="audioLoading[a.id]" class="spinner-border spinner-border-sm"></span><template v-else><i class="bi bi-play-circle"></i> Nghe</template>
              </button>
              <button class="btn btn-sm btn-outline-danger" @click="deleteAudio(a.id)" title="Xoá"><i class="bi bi-trash"></i></button>
            </div>
          </div>
        </div>

        <div class="btn-group btn-group-sm mb-3" role="group">
          <button v-for="f in filters" :key="f.id" class="btn" :class="filter === f.id ? 'btn-primary' : 'btn-outline-primary'" @click="filter = f.id">
            {{ f.label }} ({{ f.count }})
          </button>
        </div>

        <div v-for="item in reviewItems" :key="item.i" class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <div class="d-flex flex-wrap align-items-center gap-2 mb-2">
              <b>Câu {{ item.i + 1 }}</b>
              <span class="badge" :class="item.right ? 'text-bg-success' : 'text-bg-danger'">{{ item.right ? 'Đúng' : 'Sai' }}</span>
              <span class="badge text-bg-light border">{{ item.q.domain }}</span>
              <span class="badge text-bg-light border">{{ item.q.approach }}</span>
              <span v-if="wasFlagged(item.i)" class="badge text-bg-warning"><i class="bi bi-flag"></i></span>
            </div>
            <div class="mb-2 stem">
              {{ item.q.stem }}
              <button class="btn btn-sm btn-link p-0 ms-1" title="Nghe" @click="say(item.q.stem)"><i class="bi bi-volume-up"></i></button>
            </div>
            <div v-for="(o, oi) in item.q.options" :key="oi" class="rv mb-2 p-2 rounded"
              :class="item.q.correct.includes(oi) ? 'rv-ok' : pickOf(item.i).includes(oi) ? 'rv-bad' : ''">
              <div>
                <b>{{ letters[oi] }}.</b> {{ o.text }}
                <span v-if="item.q.correct.includes(oi)" class="ms-1 text-success"><i class="bi bi-check-circle-fill"></i></span>
                <span v-if="pickOf(item.i).includes(oi)" class="ms-1 small text-muted">(bạn chọn)</span>
              </div>
              <div class="small text-muted">{{ o.why }}</div>
            </div>
            <div class="bg-light rounded p-2 mt-2">
              <div class="small fw-bold mb-1">Explanation</div>
              <div class="small">{{ item.q.explanation }}</div>
              <div v-if="item.q.section" class="small text-muted mt-1"><i class="bi bi-journal-text"></i> {{ isPmp ? 'Cheatsheet' : 'Từ' }}: {{ item.q.section }}</div>
              <div v-if="item.q.terms && item.q.terms.length" class="small mt-1">
                <span v-for="(t, ti) in item.q.terms" :key="ti" class="me-3"><b>{{ t.en }}</b> = {{ t.vi }}</span>
              </div>
            </div>
            <div class="text-end mt-2">
              <button class="btn btn-sm btn-link text-muted" :disabled="isReported(item.i)" @click="report(item.i)">
                <i class="bi bi-exclamation-triangle"></i> {{ isReported(item.i) ? 'Đã báo' : 'Báo câu này có thể sai' }}
              </button>
            </div>
          </div>
        </div>

        <div class="text-center my-4">
          <button class="btn btn-warning" :disabled="aiReviewBusy" @click="runAiReview">
            <span v-if="aiReviewBusy" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-stars"></i>
            AI đánh giá bài làm (tiếng Việt)
          </button>
          <div v-if="aiReviewErr" class="text-danger small mt-2">{{ aiReviewErr }}</div>
        </div>
        <div v-if="aiReviewText" class="card border-0 shadow-sm mb-4">
          <div class="card-body">
            <h6 class="fw-bold mb-2"><i class="bi bi-robot"></i> AI đánh giá — phần cần nhớ để tránh lặp lỗi</h6>
            <div class="md" v-html="renderMarkdown(aiReviewText)"></div>
          </div>
        </div>
      </template>

      <!-- recent sets (clickable to review) -->
      <div v-if="!loading && history.length" class="card border-0 shadow-sm mt-4">
        <div class="card-body py-2">
          <div class="small fw-bold mb-1">Các bộ gần đây <span class="text-muted fw-normal">· bấm để xem lại</span></div>
          <div class="d-flex flex-wrap gap-2">
            <button v-for="h in history" :key="h.id" type="button" class="badge border btn-badge"
              :class="[h.id === (set && set.id) ? 'text-bg-primary' : h.submitted ? 'text-bg-light' : 'text-bg-secondary']"
              :title="h.submitted ? h.score + '/' + h.total : 'chưa nộp'" @click="openHistory(h.id)">
              {{ h.date.slice(5) }}<span v-if="h.seq > 1">#{{ h.seq }}</span> · {{ h.submitted ? h.score + '/' + h.total : '—' }}
              <i v-if="h.has_review" class="bi bi-stars text-warning ms-1"></i>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { api, errorMessage } from '../api'
import { renderMarkdown } from '../utils/markdown'
import { useSpeechSynthesis } from '../composables/useSpeechSynthesis'
import { subjectSlug } from '../composables/useSubject'

// "Daily 20": a stored set of 20 AI-written, exam-style questions per subject.
// PMP uses the exam-domain weighting; English draws from the learner's vocab.
// More than one set can exist per day (generate another, or re-take a saved
// one). Progress is kept in localStorage per set while it is open.

interface Opt { text: string; why: string }
interface Term { en: string; vi: string }
interface Q {
  domain: string; approach: string; type: 'single' | 'multi'
  stem: string; options: Opt[]; correct: number[]
  explanation: string; section: string; terms: Term[]
}
interface Result {
  picks: number[][] | null; flagged: number[] | null; seconds: number
  by_domain: Record<string, [number, number]> | null; correct: boolean[] | null
  reported: number[] | null; ai_review?: string
}
interface DSet { id: number; seq: number; date: string; questions: Q[]; submitted: boolean; score: number; result: Result }
interface Day { id: number; date: string; seq: number; score: number; total: number; submitted: boolean; has_review: boolean }
interface AudioFile { id: number; name: string; mime: string; bytes: number }

const SECONDS_PER_QUESTION = 80
const letters = ['A', 'B', 'C', 'D', 'E']

const isPmp = computed(() => subjectSlug.value !== 'english')
const subjectTitle = computed(() => (isPmp.value ? 'PMP' : 'English'))
const subjectBlurb = computed(() =>
  isPmp.value
    ? '20 câu tình huống bằng tiếng Anh mỗi ngày, theo tỷ trọng domain của đề thật. Câu hỏi do AI viết từ cheatsheet của bạn.'
    : '20 câu trắc nghiệm tiếng Anh (TOEIC/IELTS) mỗi ngày, do AI viết từ chính các từ bạn đang học.',
)

const today = new Date().toLocaleDateString('en-CA') // YYYY-MM-DD local
const set = ref<DSet | null>(null)
const history = ref<Day[]>([])
const loading = ref(true)
const generating = ref(false)
const retaking = ref(false)
const submitting = ref(false)
const confirmSubmit = ref(false)
const err = ref('')

const cur = ref(0)
const picks = ref<number[][]>([])
const flagged = ref<number[]>([])
const startedAt = ref(0)
const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | null = null

const filter = ref<'all' | 'wrong' | 'flagged'>('all')
const aiReviewText = ref('')
const aiReviewBusy = ref(false)
const aiReviewErr = ref('')
const tts = useSpeechSynthesis()

// audio notes
const audioFiles = ref<AudioFile[]>([])
const audioUrls = ref<Record<number, string>>({})
const audioLoading = ref<Record<number, boolean>>({})
const audioBusy = ref(false)
const audioErr = ref('')

const q = computed(() => set.value?.questions[cur.value] ?? null)
const total = computed(() => set.value?.questions.length ?? 0)
const budget = computed(() => total.value * SECONDS_PER_QUESTION)
const elapsed = computed(() => (startedAt.value ? Math.floor((now.value - startedAt.value) / 1000) : 0))
const remaining = computed(() => budget.value - elapsed.value)
const clock = computed(() => {
  const r = Math.abs(remaining.value)
  const s = `${String(Math.floor(r / 60)).padStart(2, '0')}:${String(r % 60).padStart(2, '0')}`
  return remaining.value < 0 ? '+' + s : s
})
const unanswered = computed(() => picks.value.filter((p) => !p || p.length === 0).length + Math.max(0, total.value - picks.value.length))
const pct = computed(() => (set.value && total.value ? Math.round((100 * set.value.score) / total.value) : 0))

const streak = computed(() => {
  const done = new Set(history.value.filter((h) => h.submitted).map((h) => h.date))
  let n = 0
  const d = new Date()
  if (!done.has(d.toLocaleDateString('en-CA'))) d.setDate(d.getDate() - 1)
  while (done.has(d.toLocaleDateString('en-CA'))) { n++; d.setDate(d.getDate() - 1) }
  return n
})

function minutes(sec: number): string {
  const m = Math.floor((sec || 0) / 60)
  return `${m} phút ${String((sec || 0) % 60).padStart(2, '0')} giây`
}
function sizeMB(b: number): string {
  return b >= 1024 * 1024 ? (b / 1024 / 1024).toFixed(1) + ' MB' : Math.round(b / 1024) + ' KB'
}

// ---- progress in localStorage (best effort), keyed by set id ----
const storeKey = () => `daily-${subjectSlug.value}-${set.value?.id ?? 0}`
function saveProgress() {
  if (!set.value || set.value.submitted) return
  try {
    localStorage.setItem(storeKey(), JSON.stringify({ picks: picks.value, flagged: flagged.value, startedAt: startedAt.value, cur: cur.value }))
  } catch { /* storage unavailable */ }
}
function restoreProgress() {
  picks.value = Array.from({ length: total.value }, () => [])
  flagged.value = []
  startedAt.value = Date.now()
  try {
    const raw = localStorage.getItem(storeKey())
    if (!raw) return
    const p = JSON.parse(raw)
    if (Array.isArray(p.picks)) picks.value = Array.from({ length: total.value }, (_, i) => (Array.isArray(p.picks[i]) ? p.picks[i] : []))
    if (Array.isArray(p.flagged)) flagged.value = p.flagged
    if (typeof p.startedAt === 'number' && p.startedAt > 0) startedAt.value = p.startedAt
    if (typeof p.cur === 'number' && p.cur >= 0 && p.cur < total.value) cur.value = p.cur
  } catch { /* ignore a corrupt value */ }
}
watch([picks, flagged, cur], saveProgress, { deep: true })

function startTimer() { if (!timer) timer = setInterval(() => { now.value = Date.now() }, 1000) }
function stopTimer() { if (timer) { clearInterval(timer); timer = null } }

// ---- loading ----
async function loadHistory() {
  try {
    const { data } = await api.get('/pmp-daily/history', { params: { subject: subjectSlug.value } })
    history.value = data.days ?? []
  } catch { /* history is optional */ }
}

function clearAudio() {
  for (const k in audioUrls.value) { try { URL.revokeObjectURL(audioUrls.value[+k]) } catch { /* ignore */ } }
  audioUrls.value = {}
  audioLoading.value = {}
  audioFiles.value = []
  audioErr.value = ''
}

function adopt(s: DSet | null) {
  set.value = s
  confirmSubmit.value = false
  aiReviewErr.value = ''
  aiReviewText.value = s?.result?.ai_review ?? ''
  clearAudio()
  if (s && !s.submitted) {
    cur.value = 0
    restoreProgress()
    startTimer()
  } else {
    stopTimer()
    if (s && s.submitted) loadAudio(s.id)
  }
}

async function load() {
  loading.value = true
  err.value = ''
  try {
    const { data } = await api.get('/pmp-daily', { params: { subject: subjectSlug.value, date: today } })
    adopt(data.set ?? null)
    await loadHistory()
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

async function generate() {
  generating.value = true
  err.value = ''
  try {
    const { data } = await api.post('/pmp-daily/generate', { date: today, subject: subjectSlug.value }, { timeout: 180000 })
    adopt(data.set)
    await loadHistory()
    window.scrollTo({ top: 0 })
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    generating.value = false
  }
}

async function retake() {
  if (!set.value) return
  retaking.value = true
  err.value = ''
  try {
    const { data } = await api.post('/pmp-daily/retake', { id: set.value.id, date: today })
    adopt(data.set)
    await loadHistory()
    window.scrollTo({ top: 0 })
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    retaking.value = false
  }
}

async function openHistory(id: number) {
  if (set.value && set.value.id === id) return
  loading.value = true
  err.value = ''
  try {
    const { data } = await api.get('/pmp-daily', { params: { id } })
    adopt(data.set ?? null)
    window.scrollTo({ top: 0 })
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

// ---- answering ----
const isPicked = (qi: number, oi: number) => (picks.value[qi] ?? []).includes(oi)
const answered = (qi: number) => (picks.value[qi] ?? []).length > 0

function choose(qi: number, oi: number) {
  const question = set.value?.questions[qi]
  if (!question) return
  const list = [...(picks.value[qi] ?? [])]
  if (question.type === 'multi') {
    const at = list.indexOf(oi)
    if (at >= 0) list.splice(at, 1)
    else if (list.length < 2) list.push(oi)
    else { list.shift(); list.push(oi) }
  } else {
    list.splice(0, list.length, oi)
  }
  picks.value[qi] = list
}

function toggleFlag(qi: number) {
  const at = flagged.value.indexOf(qi)
  if (at >= 0) flagged.value.splice(at, 1)
  else flagged.value.push(qi)
}

async function submit() {
  if (!set.value) return
  submitting.value = true
  err.value = ''
  try {
    const { data } = await api.post('/pmp-daily/submit', {
      id: set.value.id,
      picks: picks.value,
      flagged: flagged.value,
      seconds: elapsed.value,
    })
    try { localStorage.removeItem(storeKey()) } catch { /* ignore */ }
    adopt(data.set)
    filter.value = 'all'
    await loadHistory()
    window.scrollTo({ top: 0 })
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    submitting.value = false
  }
}

// ---- review ----
const pickOf = (i: number): number[] => set.value?.result.picks?.[i] ?? []
const wasFlagged = (i: number) => !!set.value?.result.flagged?.includes(i)
const isReported = (i: number) => !!set.value?.result.reported?.includes(i)
const isRight = (i: number) => !!set.value?.result.correct?.[i]

const filters = computed(() => {
  const n = total.value
  const wrong = Array.from({ length: n }, (_, i) => i).filter((i) => !isRight(i)).length
  const fl = set.value?.result.flagged?.length ?? 0
  return [
    { id: 'all' as const, label: 'Tất cả', count: n },
    { id: 'wrong' as const, label: 'Câu sai', count: wrong },
    { id: 'flagged' as const, label: 'Đã đánh dấu', count: fl },
  ]
})

const reviewItems = computed(() => {
  const s = set.value
  if (!s) return []
  return s.questions
    .map((qq, i) => ({ q: qq, i, right: isRight(i) }))
    .filter((it) => filter.value === 'all' || (filter.value === 'wrong' && !it.right) || (filter.value === 'flagged' && wasFlagged(it.i)))
})

async function runAiReview() {
  const sv = set.value
  if (!sv) return
  const wrong = sv.questions
    .map((qq, i) => ({ qq, i }))
    .filter(({ i }) => !isRight(i))
    .map(({ qq, i }) => ({
      stem: qq.stem,
      correct: qq.correct.map((ci) => qq.options[ci]?.text ?? '').filter(Boolean),
      picked: pickOf(i).map((pi) => qq.options[pi]?.text ?? '').filter(Boolean),
      explanation: qq.explanation,
      section: qq.section,
      terms: (qq.terms ?? []).map((t) => `${t.en} = ${t.vi}`),
    }))
  aiReviewBusy.value = true
  aiReviewErr.value = ''
  try {
    const { data } = await api.post('/pmp-daily/ai-review', { id: sv.id, wrong }, { timeout: 120000 })
    aiReviewText.value = data.review ?? ''
    await loadHistory()
  } catch (e) {
    aiReviewErr.value = errorMessage(e)
  } finally {
    aiReviewBusy.value = false
  }
}

async function report(i: number) {
  if (!set.value) return
  try {
    const { data } = await api.post('/pmp-daily/report', { id: set.value.id, index: i })
    set.value = data.set
  } catch (e) {
    err.value = errorMessage(e)
  }
}

// ---- audio notes ----
async function loadAudio(setId: number) {
  try {
    const { data } = await api.get('/pmp-daily/audio', { params: { set_id: setId } })
    audioFiles.value = data.files ?? []
  } catch { audioFiles.value = [] }
}

async function uploadAudio(e: Event) {
  const input = e.target as HTMLInputElement
  const files = input.files
  if (!files || !files.length || !set.value) return
  audioBusy.value = true
  audioErr.value = ''
  try {
    const fd = new FormData()
    fd.append('set_id', String(set.value.id))
    for (const f of Array.from(files)) fd.append('files', f)
    await api.post('/pmp-daily/audio', fd, { timeout: 300000 })
    await loadAudio(set.value.id)
  } catch (err2) {
    audioErr.value = errorMessage(err2)
  } finally {
    audioBusy.value = false
    input.value = ''
  }
}

async function loadBlob(id: number) {
  audioLoading.value = { ...audioLoading.value, [id]: true }
  try {
    const { data } = await api.get(`/pmp-daily/audio/${id}`, { responseType: 'blob' })
    audioUrls.value = { ...audioUrls.value, [id]: URL.createObjectURL(data) }
  } catch (e) {
    audioErr.value = errorMessage(e)
  } finally {
    audioLoading.value = { ...audioLoading.value, [id]: false }
  }
}

async function deleteAudio(id: number) {
  try {
    await api.delete(`/pmp-daily/audio/${id}`)
    if (audioUrls.value[id]) { try { URL.revokeObjectURL(audioUrls.value[id]) } catch { /* ignore */ } }
    audioFiles.value = audioFiles.value.filter((a) => a.id !== id)
  } catch (e) {
    audioErr.value = errorMessage(e)
  }
}

async function say(text: string) {
  try { await tts.speakEnglish(text, 0.9) } catch { /* speech unavailable or interrupted */ }
}

onMounted(load)
onUnmounted(() => {
  stopTimer()
  tts.stop()
  clearAudio()
})
</script>

<style scoped>
.stem { line-height: 1.55; }
.opt { cursor: pointer; padding: 0.55rem 0.75rem 0.55rem 2rem; border: 1px solid #e2e8f0; border-radius: 8px; }
.opt.picked { border-color: #4f46e5; background: #eef2ff; }
.pal { min-width: 2.2rem; }
.pal.flagged { box-shadow: 0 0 0 2px #f59e0b; }
.rv { border: 1px solid #e2e8f0; }
.rv-ok { background: #ecfdf5; border-color: #10b981; }
.rv-bad { background: #fef2f2; border-color: #ef4444; }
.btn-badge { cursor: pointer; border: 1px solid #cbd5e1; }
.btn-badge:hover { filter: brightness(0.97); }
</style>
