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
            <ul v-if="isAws" class="small text-muted mb-3">
              <li>20 câu tình huống kiểu SAA-C03, trải 4 lĩnh vực: Secure ~30%, Resilient ~26%, Performance ~24%, Cost ~20%</li>
              <li>Mỗi câu hỏi "giải pháp nào MOST cost-effective/secure/available…"; có câu chọn 2 đáp án</li>
              <li>Câu hỏi do AI viết từ cheatsheet AWS của bạn; sau khi nộp có phân tích từng đáp án</li>
            </ul>
            <ul v-else-if="isPmp" class="small text-muted mb-3">
              <li>7 câu People, 8 câu Process, 5 câu Business Environment, trộn lẫn như đề thật</li>
              <li>Khoảng 40% predictive, 60% agile/hybrid; có câu chọn 2 đáp án và câu tính toán</li>
              <li>Đồng hồ ~27 phút (80 giây mỗi câu); sau khi nộp có phân tích từng đáp án</li>
            </ul>
            <ul v-else class="small text-muted mb-3">
              <li>20 câu trắc nghiệm tiếng Anh (TOEIC/IELTS) từ chính các từ bạn đang học</li>
              <li>Điền từ, chọn nghĩa gần nhất, hoặc chọn câu dùng từ đúng</li>
              <li>Sau khi nộp có giải thích từng đáp án và nghĩa tiếng Việt</li>
            </ul>
            <button class="btn btn-primary" :disabled="generating" @click="openGen">
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
            <div class="mt-3">
              <label class="form-label small mb-1"><i class="bi bi-pencil"></i> Ghi chú câu này</label>
              <textarea v-model="notes[cur]" class="form-control form-control-sm" rows="2" placeholder="Suy nghĩ / lý do chọn / điều cần nhớ…"></textarea>
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
              <button class="btn btn-sm btn-primary" :disabled="generating" @click="openGen">
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
              <button class="btn btn-sm" :class="player.isCurrent('') ? 'btn-outline-secondary' : 'btn-outline-secondary'" :disabled="audioLoading[a.id]" @click="playAudio(a)">
                <span v-if="audioLoading[a.id]" class="spinner-border spinner-border-sm"></span><template v-else><i class="bi bi-play-circle"></i> Nghe</template>
              </button>
              <button class="btn btn-sm btn-outline-danger" @click="deleteAudio(a.id)" title="Xoá"><i class="bi bi-trash"></i></button>
            </div>
          </div>
        </div>

        <div class="d-flex flex-wrap align-items-center gap-2 mb-3">
          <div class="btn-group btn-group-sm" role="group">
            <button v-for="f in filters" :key="f.id" class="btn" :class="filter === f.id ? 'btn-primary' : 'btn-outline-primary'" @click="filter = f.id">
              {{ f.label }} ({{ f.count }})
            </button>
          </div>
          <button class="btn btn-sm btn-outline-secondary ms-auto" @click="copyQuestions" :title="'Copy nội dung ' + reviewItems.length + ' câu đang hiện (đề + đáp án) — không gửi AI'">
            <i class="bi" :class="copied ? 'bi-check2' : 'bi-clipboard'"></i> {{ copied ? 'Đã copy' : 'Copy câu hỏi + đáp án' }}
          </button>
          <button class="btn btn-sm btn-outline-secondary" @click="downloadQuestions" title="Tải nội dung câu đang hiện ra file .txt">
            <i class="bi bi-download"></i> Tải .txt
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
            <div class="mt-2">
              <label class="form-label small mb-1"><i class="bi bi-pencil"></i> Ghi chú của bạn</label>
              <textarea v-model="notes[item.i]" class="form-control form-control-sm" rows="2" placeholder="Ghi chú câu này (tự lưu khi ngừng gõ)…" @input="scheduleSaveNotes" @blur="saveNotes"></textarea>
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
            <div class="d-flex flex-wrap align-items-center gap-2 mb-2">
              <h6 class="fw-bold mb-0"><i class="bi bi-robot"></i> AI đánh giá — phần cần nhớ để tránh lặp lỗi</h6>
              <button class="btn btn-sm btn-outline-secondary ms-auto" @click="copyAiReview">
                <i class="bi" :class="aiCopied ? 'bi-check2' : 'bi-clipboard'"></i> {{ aiCopied ? 'Đã copy' : 'Copy nội dung' }}
              </button>
            </div>
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

      <!-- generate options -->
      <div v-if="showGen" class="gen-overlay" @click.self="showGen = false">
        <div class="card border-0 shadow gen-card">
          <div class="card-body">
            <div class="d-flex align-items-center mb-3">
              <h6 class="fw-bold mb-0"><i class="bi bi-magic"></i> Tạo bộ 20 câu</h6>
              <button type="button" class="btn-close ms-auto" aria-label="Close" @click="showGen = false"></button>
            </div>
            <label class="form-label small fw-semibold mb-1">Nội dung muốn tập trung <span class="text-muted fw-normal">(không bắt buộc)</span></label>
            <textarea v-model="genFocus" class="form-control form-control-sm" rows="2" :placeholder="focusPlaceholder"></textarea>
            <div class="form-text mb-3">Để trống thì AI tự chọn chủ đề như thường.</div>
            <div class="d-flex gap-2">
              <button class="btn btn-primary" :disabled="generating" @click="confirmGen">
                <span v-if="generating" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-magic"></i>
                {{ generating ? 'AI đang viết… (1–2 phút)' : 'Tạo bộ' }}
              </button>
              <button class="btn btn-outline-secondary" :disabled="generating" @click="showGen = false">Huỷ</button>
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
import { useAudioPlayer } from '../composables/useAudioPlayer'

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
  reported: number[] | null; ai_review?: string; notes?: string[] | null
}
interface DSet { id: number; seq: number; date: string; questions: Q[]; submitted: boolean; score: number; result: Result }
interface Day { id: number; date: string; seq: number; score: number; total: number; submitted: boolean; has_review: boolean }
interface AudioFile { id: number; name: string; mime: string; bytes: number }

const SECONDS_PER_QUESTION = 80
const letters = ['A', 'B', 'C', 'D', 'E']

const isPmp = computed(() => subjectSlug.value !== 'english') // non-English subjects ground questions in the cheatsheet (section = note title)
const isAws = computed(() => subjectSlug.value === 'aws')
const subjectTitle = computed(() => (subjectSlug.value === 'english' ? 'English' : isAws.value ? 'AWS' : 'PMP'))
const subjectBlurb = computed(() =>
  isAws.value
    ? '20 câu trắc nghiệm kiểu SAA-C03 mỗi ngày, trải 4 lĩnh vực (Secure/Resilient/Performance/Cost). Câu hỏi do AI viết từ cheatsheet AWS của bạn.'
    : subjectSlug.value === 'english'
    ? '20 câu trắc nghiệm tiếng Anh (TOEIC/IELTS) mỗi ngày, do AI viết từ chính các từ bạn đang học.'
    : '20 câu tình huống bằng tiếng Anh mỗi ngày, theo tỷ trọng domain của đề thật. Câu hỏi do AI viết từ cheatsheet của bạn.',
)

const today = new Date().toLocaleDateString('en-CA') // YYYY-MM-DD local
const set = ref<DSet | null>(null)
const history = ref<Day[]>([])
const loading = ref(true)
const generating = ref(false)
const showGen = ref(false)
const genFocus = ref('')
const focusPlaceholder = computed(() => (subjectSlug.value === 'english'
  ? 'VD: business emails, synonyms, phrasal verbs…'
  : 'VD: Risk, EVM, Agile/Scrum, Change Control…'))
const retaking = ref(false)
const submitting = ref(false)
const confirmSubmit = ref(false)
const err = ref('')

const cur = ref(0)
const picks = ref<number[][]>([])
const flagged = ref<number[]>([])
const notes = ref<string[]>([])
const startedAt = ref(0)
const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | null = null

const filter = ref<'all' | 'wrong' | 'flagged'>('all')
const aiReviewText = ref('')
const aiReviewBusy = ref(false)
const aiReviewErr = ref('')
const copied = ref(false)
const aiCopied = ref(false)

// Copy the currently-shown questions (built locally — NOT sent to the AI, so no
// token cost): stem + options (marking the correct one and the learner's pick)
// + explanation.
function questionsText(): string {
  const lines: string[] = []
  for (const it of reviewItems.value) {
    const qq = it.q, i = it.i
    lines.push(`Câu ${i + 1} (${qq.domain} · ${qq.approach})${it.right ? '' : ' — SAI'}`)
    lines.push(qq.stem)
    qq.options.forEach((o, oi) => {
      const mark = qq.correct.includes(oi) ? '  ✅ (đáp án đúng)' : (pickOf(i).includes(oi) ? '  ❌ (bạn chọn)' : '')
      lines.push(`  ${letters[oi]}. ${o.text}${mark}`)
    })
    if (qq.explanation) lines.push(`Giải thích: ${qq.explanation}`)
    if (qq.section) lines.push(`Cheatsheet: ${qq.section}`)
    if ((notes.value[i] ?? '').trim()) lines.push(`Ghi chú của bạn: ${notes.value[i].trim()}`)
    lines.push('')
  }
  return lines.join('\n').trim()
}
async function copyAiReview() {
  try {
    await navigator.clipboard.writeText(aiReviewText.value)
    aiCopied.value = true
    setTimeout(() => { aiCopied.value = false }, 1800)
  } catch { /* clipboard unavailable */ }
}
async function copyQuestions() {
  try {
    await navigator.clipboard.writeText(questionsText())
    copied.value = true
    setTimeout(() => { copied.value = false }, 1800)
  } catch { /* clipboard unavailable */ }
}
function downloadQuestions() {
  const blob = new Blob([questionsText()], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `daily-${subjectSlug.value}-${set.value?.date ?? today}.txt`
  document.body.appendChild(a)
  a.click()
  a.remove()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
const tts = useSpeechSynthesis()
const player = useAudioPlayer()

// audio notes
const audioFiles = ref<AudioFile[]>([])
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
    localStorage.setItem(storeKey(), JSON.stringify({ picks: picks.value, flagged: flagged.value, startedAt: startedAt.value, cur: cur.value, notes: notes.value }))
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
    if (Array.isArray(p.notes)) p.notes.forEach((n: unknown, i: number) => { if (typeof n === 'string' && i < notes.value.length) notes.value[i] = n })
  } catch { /* ignore a corrupt value */ }
}
watch([picks, flagged, cur, notes], saveProgress, { deep: true })

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
  audioLoading.value = {}
  audioFiles.value = []
  audioErr.value = ''
}

function adopt(s: DSet | null) {
  set.value = s
  notes.value = Array.from({ length: s?.questions.length ?? 0 }, (_, i) => (s?.result?.notes?.[i] ?? ''))
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

function openGen() { genFocus.value = ''; showGen.value = true }
async function confirmGen() { await generate(); if (!err.value) showGen.value = false }
async function generate() {
  generating.value = true
  err.value = ''
  try {
    const { data } = await api.post('/pmp-daily/generate', { date: today, subject: subjectSlug.value, focus: genFocus.value.trim() }, { timeout: 180000 })
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
      notes: notes.value,
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
  const build = (qq: Q, i: number, right: boolean) => ({
    stem: qq.stem,
    correct: qq.correct.map((ci) => qq.options[ci]?.text ?? '').filter(Boolean),
    picked: pickOf(i).map((pi) => qq.options[pi]?.text ?? '').filter(Boolean),
    explanation: qq.explanation,
    section: qq.section,
    terms: (qq.terms ?? []).map((t) => `${t.en} = ${t.vi}`),
    note: notes.value[i] ?? '',
    right,
  })
  // all wrong questions + correct questions that have a note (so AI can respond to them)
  const wrong = [
    ...sv.questions.map((qq, i) => ({ qq, i })).filter(({ i }) => !isRight(i)).map(({ qq, i }) => build(qq, i, false)),
    ...sv.questions.map((qq, i) => ({ qq, i })).filter(({ i }) => isRight(i) && (notes.value[i] ?? '').trim()).map(({ qq, i }) => build(qq, i, true)),
  ]
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

let notesTimer: ReturnType<typeof setTimeout> | null = null
function scheduleSaveNotes() {
  if (notesTimer) clearTimeout(notesTimer)
  notesTimer = setTimeout(() => { saveNotes() }, 900)
}
async function saveNotes() {
  if (notesTimer) { clearTimeout(notesTimer); notesTimer = null }
  if (!set.value) return
  try { await api.post('/pmp-daily/notes', { id: set.value.id, notes: notes.value }) } catch { /* best effort */ }
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

async function playAudio(a: AudioFile) {
  audioLoading.value = { ...audioLoading.value, [a.id]: true }
  try {
    const { data } = await api.get(`/pmp-daily/audio/${a.id}`, { responseType: 'blob' })
    player.play(URL.createObjectURL(data), a.name, true)
  } catch (e) {
    audioErr.value = errorMessage(e)
  } finally {
    audioLoading.value = { ...audioLoading.value, [a.id]: false }
  }
}

async function deleteAudio(id: number) {
  try {
    await api.delete(`/pmp-daily/audio/${id}`)
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
.gen-overlay { position: fixed; inset: 0; background: rgba(0,0,0,.45); display: flex; align-items: center; justify-content: center; z-index: 1060; padding: 1rem; }
.gen-card { width: 100%; max-width: 460px; }
</style>
