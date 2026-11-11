<template>
  <div class="row justify-content-center">
    <div class="col-12 col-lg-9 col-xl-8">
      <div class="d-flex flex-wrap justify-content-between align-items-end mb-3 gap-2">
        <div>
          <h2 class="fw-bold mb-0"><i class="bi bi-calendar-check"></i> PMP Daily 20</h2>
          <small class="text-muted">20 câu tình huống bằng tiếng Anh mỗi ngày, theo tỷ trọng domain của đề thật. Câu hỏi do AI viết từ cheatsheet của bạn.</small>
        </div>
        <div v-if="streak > 0" class="badge text-bg-warning fs-6"><i class="bi bi-fire"></i> {{ streak }} ngày liên tiếp</div>
      </div>

      <div v-if="err" class="alert alert-danger py-2 small">{{ err }}</div>
      <div v-if="loading" class="text-center py-5"><div class="spinner-border text-primary"></div></div>

      <!-- no set yet today -->
      <template v-else-if="!set">
        <div class="card border-0 shadow-sm mb-3">
          <div class="card-body">
            <p class="mb-2">Bộ hôm nay ({{ today }}) chưa được tạo. Mỗi ngày có một bộ; mở lại trong ngày sẽ thấy đúng bộ đó.</p>
            <ul class="small text-muted mb-3">
              <li>7 câu People, 8 câu Process, 5 câu Business Environment, trộn lẫn như đề thật</li>
              <li>Khoảng 40% predictive, 60% agile/hybrid; có câu chọn 2 đáp án và câu tính toán</li>
              <li>Đồng hồ ~27 phút (80 giây mỗi câu); sau khi nộp có phân tích từng đáp án</li>
            </ul>
            <button class="btn btn-primary" :disabled="generating" @click="generate">
              <span v-if="generating" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-magic"></i>
              {{ generating ? 'AI đang viết 20 câu… (có thể mất 1–2 phút)' : 'Tạo bộ 20 câu hôm nay' }}
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
            <p class="small text-muted mt-2 mb-0">Bộ mới sẽ có vào ngày mai. Xem lại các câu sai bên dưới và đọc phần giải thích từng đáp án.</p>
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
              <div v-if="item.q.section" class="small text-muted mt-1"><i class="bi bi-journal-text"></i> Cheatsheet: {{ item.q.section }}</div>
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
      </template>

      <!-- last days -->
      <div v-if="!loading && history.length" class="card border-0 shadow-sm mt-4">
        <div class="card-body py-2">
          <div class="small fw-bold mb-1">Các ngày gần đây</div>
          <div class="d-flex flex-wrap gap-2">
            <span v-for="h in history" :key="h.date" class="badge border" :class="h.submitted ? 'text-bg-light' : 'text-bg-secondary'"
              :title="h.submitted ? h.score + '/' + h.total : 'chưa nộp'">
              {{ h.date.slice(5) }} · {{ h.submitted ? h.score + '/' + h.total : '—' }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { api, errorMessage } from '../api'
import { useSpeechSynthesis } from '../composables/useSpeechSynthesis'

// PMP Daily 20: one stored set of 20 AI-written, exam-style questions per day.
// Progress (answers, marks, start time) is kept in localStorage while the set
// is open; the server scores it on submit.

interface Opt { text: string; why: string }
interface Term { en: string; vi: string }
interface Q {
  domain: string; approach: string; type: 'single' | 'multi'
  stem: string; options: Opt[]; correct: number[]
  explanation: string; section: string; terms: Term[]
}
interface Result {
  picks: number[][] | null; flagged: number[] | null; seconds: number
  by_domain: Record<string, [number, number]> | null; correct: boolean[] | null; reported: number[] | null
}
interface DSet { date: string; questions: Q[]; submitted: boolean; score: number; result: Result }
interface Day { date: string; score: number; total: number; submitted: boolean }

const SECONDS_PER_QUESTION = 80
const letters = ['A', 'B', 'C', 'D', 'E']

const today = new Date().toLocaleDateString('en-CA') // YYYY-MM-DD in the local time zone
const set = ref<DSet | null>(null)
const history = ref<Day[]>([])
const loading = ref(true)
const generating = ref(false)
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
const tts = useSpeechSynthesis()

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
  // consecutive submitted days ending today (or yesterday if today is not done yet)
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

// ---- progress in localStorage (best effort) ----
const storeKey = () => `pmp-daily-${today}`
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

// ---- loading ----
async function loadHistory() {
  try {
    const { data } = await api.get('/pmp-daily/history')
    history.value = data.days ?? []
  } catch { /* history is optional */ }
}

function adopt(s: DSet | null) {
  set.value = s
  confirmSubmit.value = false
  if (s && !s.submitted) {
    cur.value = 0
    restoreProgress()
    if (!timer) timer = setInterval(() => { now.value = Date.now() }, 1000)
  } else if (timer) {
    clearInterval(timer)
    timer = null
  }
}

async function load() {
  loading.value = true
  err.value = ''
  try {
    const { data } = await api.get('/pmp-daily', { params: { date: today } })
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
    const { data } = await api.post('/pmp-daily/generate', { date: today }, { timeout: 180000 })
    adopt(data.set)
    await loadHistory()
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    generating.value = false
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
    else { list.shift(); list.push(oi) } // a third choice replaces the oldest
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
      date: set.value.date,
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

async function report(i: number) {
  if (!set.value) return
  try {
    const { data } = await api.post('/pmp-daily/report', { date: set.value.date, index: i })
    set.value = data.set
  } catch (e) {
    err.value = errorMessage(e)
  }
}

async function say(text: string) {
  try { await tts.speakEnglish(text, 0.9) } catch { /* speech unavailable or interrupted */ }
}

onMounted(load)
onUnmounted(() => {
  if (timer) clearInterval(timer)
  tts.stop()
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
</style>
