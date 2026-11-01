<template>
  <div>
    <!-- list view -->
    <div v-if="!activeCard">
      <div class="mb-3">
        <h2 class="fw-bold mb-0"><i class="bi bi-terminal"></i> Thực hành gõ lệnh (Labs)</h2>
        <small class="text-muted">Làm theo từng bước, gõ đúng lệnh mới qua bước tiếp theo — luyện phản xạ gõ lệnh trước khi lên THM/HTB.</small>
      </div>

      <div v-if="loading" class="placeholder-glow"><span class="placeholder col-6"></span></div>
      <template v-else>
        <ul class="nav nav-pills mb-3">
          <li class="nav-item">
            <button class="nav-link" :class="{ active: activeTab === 'lab' }" @click="activeTab = 'lab'">
              <i class="bi bi-terminal"></i> Lab <span class="badge bg-light text-dark ms-1">{{ labOnlyCards.length }}</span>
            </button>
          </li>
          <li class="nav-item">
            <button class="nav-link" :class="{ active: activeTab === 'assignment' }" @click="activeTab = 'assignment'">
              <i class="bi bi-flag"></i> Assignment <span class="badge bg-light text-dark ms-1">{{ assignmentCards.length }}</span>
            </button>
          </li>
        </ul>

        <div v-if="!visibleCards.length" class="small text-muted">Chưa có bài nào ở mục này.</div>

        <div class="row g-3">
          <div v-for="c in visibleCards" :key="c.id" class="col-md-6">
            <div class="card border-0 shadow-sm h-100 hoverable">
              <div class="card-body d-flex flex-column">
                <h5 class="fw-bold mb-1">{{ c.front }}</h5>
                <p class="small text-muted mb-2">{{ c.payload.scenario }}</p>
              <div class="small mb-2">
                <span class="badge" :class="progressBadgeClass(c)">
                  {{ progressLabel(c) }}
                </span>
                <span v-if="c.payload.difficulty" class="badge ms-1" :class="difficultyBadgeClass(c.payload.difficulty)">
                  {{ difficultyLabel(c.payload.difficulty) }}
                </span>
              </div>
                <router-link :to="`/labs?id=${c.id}`" class="btn mt-auto align-self-start"
                  :class="isDone(c) ? 'btn-outline-success' : 'btn-primary'">
                  {{ isDone(c) ? 'Làm lại' : (stepIndex(c.id) > 0 ? 'Tiếp tục' : 'Bắt đầu') }}
                </router-link>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- runner view -->
    <div v-else>
      <router-link to="/labs" class="btn btn-sm btn-outline-secondary mb-3"><i class="bi bi-arrow-left"></i> Danh sách lab</router-link>

      <h3 class="fw-bold mb-1">
        {{ activeCard.front }}
        <span v-if="activeCard.payload.mode === 'assignment'" class="badge bg-dark ms-1">Assignment</span>
        <span v-if="activeCard.payload.difficulty" class="badge ms-1" :class="difficultyBadgeClass(activeCard.payload.difficulty)">
          {{ difficultyLabel(activeCard.payload.difficulty) }}
        </span>
      </h3>
      <p class="text-muted">{{ activeCard.payload.scenario }}</p>

      <div class="progress mb-3" style="height: 8px;">
        <div class="progress-bar bg-success" :style="{ width: progressPct + '%' }"></div>
      </div>
      <div class="d-flex justify-content-between align-items-center mb-1">
        <p class="small text-muted mb-0">Bước {{ Math.min(curStep + 1, totalSteps) }} / {{ totalSteps }}</p>
        <button v-if="curStep > 0" class="btn btn-sm btn-outline-secondary" @click="prev">
          <i class="bi bi-arrow-left"></i> Bước trước
        </button>
      </div>

      <!-- finished -->
      <div v-if="finished" class="alert alert-success">
        <i class="bi bi-trophy"></i> <strong>Hoàn thành lab!</strong> Bạn đã gõ đúng toàn bộ {{ totalSteps }} bước.
        <div class="mt-2">
          <router-link to="/labs" class="btn btn-sm btn-success">Quay lại danh sách</router-link>
          <button class="btn btn-sm btn-outline-secondary ms-2" @click="restart">Làm lại từ đầu</button>
        </div>
      </div>

      <!-- current step -->
      <div v-else class="card border-0 shadow-sm">
        <div class="card-body">
          <pre v-if="currentStep.output" class="bg-dark text-light p-2 rounded small mb-2" style="white-space: pre-wrap;">{{ currentStep.output }}</pre>

          <p class="fw-semibold">{{ currentStep.prompt }}</p>

          <div v-if="showHint && currentStep.hint" class="alert alert-light border small mb-2">
            <i class="bi bi-lightbulb text-warning"></i> {{ currentStep.hint }}
          </div>

          <!-- choice step: pick the right next move -->
          <div v-if="currentStep.type === 'choice'">
            <div class="d-grid gap-2 mb-2">
              <button v-for="(opt, i) in currentStep.choices" :key="i"
                class="btn text-start"
                :class="choiceBtnClass(i)"
                :disabled="feedback === 'correct'"
                @click="checkChoice(i)">
                {{ opt }}
              </button>
            </div>
            <button v-if="currentStep.hint" class="btn btn-outline-secondary btn-sm mb-2" @click="showHint = !showHint">
              {{ showHint ? 'Ẩn gợi ý' : 'Xem gợi ý' }}
            </button>
          </div>

          <!-- command step: type the command -->
          <template v-else>
            <div class="input-group mb-2">
              <span class="input-group-text font-monospace">$</span>
              <input v-model="answer" type="text" class="form-control font-monospace" placeholder="Gõ lệnh ở đây..."
                :class="{ 'is-invalid': feedback === 'wrong' }"
                @keyup.enter="onEnter" autocomplete="off" spellcheck="false" />
            </div>

            <div class="d-flex gap-2 mb-2">
              <button class="btn btn-primary" :disabled="!answer.trim()" @click="check">Kiểm tra</button>
              <button v-if="currentStep.hint" class="btn btn-outline-secondary" @click="showHint = !showHint">
                {{ showHint ? 'Ẩn gợi ý' : 'Xem gợi ý' }}
              </button>
            </div>
          </template>

          <div v-if="feedback === 'wrong'" class="alert alert-danger py-2 small mb-0">
            <i class="bi bi-x-circle"></i> Chưa đúng, thử lại (gợi ý: kiểm tra lại cú pháp/cờ của lệnh).
          </div>
          <div v-else-if="feedback === 'correct'" class="alert alert-success py-2 small mb-0">
            <i class="bi bi-check-circle"></i> {{ currentStep.explanation || 'Chính xác!' }}
            <div class="mt-2">
              <button class="btn btn-sm btn-success" @click="next">Bước tiếp theo <i class="bi bi-arrow-right"></i></button>
            </div>
          </div>
        </div>
      </div>

      <!-- reference / study notes, for review -->
      <div v-if="activeCard.payload.reference" class="mt-3">
        <button class="btn btn-sm btn-outline-secondary" @click="showReference = !showReference">
          <i class="bi bi-journal-text"></i> {{ showReference ? 'Ẩn tài liệu ôn tập' : 'Xem tài liệu ôn tập' }}
        </button>
        <div v-if="showReference" class="card border-0 shadow-sm mt-2">
          <div class="card-body reference-notes" v-html="referenceHtml"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../api'
import { loadSubjects, subject } from '../composables/useSubject'
import type { Card, LabStep } from '../types'

const route = useRoute()

const loading = ref(true)
const labCards = ref<Card[]>([])

const activeTab = ref<'lab' | 'assignment'>('lab')
const labOnlyCards = computed(() => labCards.value.filter((c) => c.payload.mode !== 'assignment'))
const assignmentCards = computed(() => labCards.value.filter((c) => c.payload.mode === 'assignment'))
const visibleCards = computed(() => (activeTab.value === 'assignment' ? assignmentCards.value : labOnlyCards.value))

// --- local practice progress (per card, this browser only — this is
// self-practice muscle-memory drilling, not graded/synced progress) ---
const PROGRESS_KEY = 'elearn.labs.progress'

function loadProgress(): Record<number, number> {
  try { return JSON.parse(localStorage.getItem(PROGRESS_KEY) || '{}') } catch { return {} }
}
function saveProgress(p: Record<number, number>) {
  try { localStorage.setItem(PROGRESS_KEY, JSON.stringify(p)) } catch { /* ignore */ }
}
const progress = ref<Record<number, number>>(loadProgress())

function stepIndex(cardId: number): number { return progress.value[cardId] ?? 0 }
function totalStepsOf(c: Card): number { return c.payload.steps?.length ?? 0 }
function isDone(c: Card): boolean { return stepIndex(c.id) >= totalStepsOf(c) && totalStepsOf(c) > 0 }
function progressLabel(c: Card): string {
  const total = totalStepsOf(c)
  if (isDone(c)) return `Hoàn thành ${total}/${total}`
  const cur = stepIndex(c.id)
  return cur > 0 ? `${cur}/${total} bước` : `${total} bước`
}
function progressBadgeClass(c: Card): string {
  return isDone(c) ? 'bg-success' : (stepIndex(c.id) > 0 ? 'bg-warning text-dark' : 'bg-light text-dark border')
}

function difficultyLabel(d: string): string {
  switch (d) {
    case 'co_ban': return 'Cơ bản'
    case 'trung_binh': return 'Trung bình'
    case 'nang_cao': return 'Nâng cao'
    default: return d
  }
}
function difficultyBadgeClass(d: string): string {
  switch (d) {
    case 'co_ban': return 'bg-info text-dark'
    case 'trung_binh': return 'bg-warning text-dark'
    case 'nang_cao': return 'bg-danger'
    default: return 'bg-light text-dark border'
  }
}

async function loadLabs() {
  await loadSubjects()
  const deck = subject.value?.decks.find((d) => d.slug === 'labs')
  if (!deck) { labCards.value = []; loading.value = false; return }
  const { data } = await api.get('/cards', { params: { deck_id: deck.id, limit: 100 } })
  labCards.value = (data?.data ?? []).filter((c: Card) => c.kind === 'lab')
  loading.value = false
}

// --- runner state ---
const activeId = computed(() => {
  const raw = route.query.id
  const n = Number(Array.isArray(raw) ? raw[0] : raw)
  return Number.isFinite(n) && n > 0 ? n : null
})
const activeCard = computed(() => labCards.value.find((c) => c.id === activeId.value) ?? null)
const totalSteps = computed(() => activeCard.value ? totalStepsOf(activeCard.value) : 0)
const curStep = computed(() => activeCard.value ? stepIndex(activeCard.value.id) : 0)
const finished = computed(() => totalSteps.value > 0 && curStep.value >= totalSteps.value)
const currentStep = computed<LabStep>(() => activeCard.value?.payload.steps?.[curStep.value] ?? { prompt: '', patterns: [] })
const chosenIndex = ref<number | null>(null)

const answer = ref('')
const feedback = ref<'idle' | 'correct' | 'wrong'>('idle')
const showHint = ref(false)
const showReference = ref(false)
const progressPct = computed(() => totalSteps.value ? Math.round((curStep.value / totalSteps.value) * 100) : 0)

watch(activeId, () => { answer.value = ''; feedback.value = 'idle'; showHint.value = false; showReference.value = false; chosenIndex.value = null })

function normalize(raw: string): string {
  return raw.trim().replace(/\s+/g, ' ')
}

function check() {
  if (!activeCard.value) return
  const norm = normalize(answer.value)
  const ok = (currentStep.value.patterns ?? []).some((p) => {
    try { return new RegExp(p, 'i').test(norm) } catch { return false }
  })
  feedback.value = ok ? 'correct' : 'wrong'
}

function checkChoice(i: number) {
  if (!activeCard.value || feedback.value === 'correct') return
  chosenIndex.value = i
  feedback.value = i === currentStep.value.answer ? 'correct' : 'wrong'
}

// Enter in the command input: check the answer, or — if already correct —
// advance to the next step (so the learner can keep hitting Enter).
function onEnter() {
  if (!activeCard.value) return
  if (feedback.value === 'correct') { next(); return }
  check()
}

// Enter anywhere (no input to focus on choice steps): advance once correct.
function handleGlobalKeydown(e: KeyboardEvent) {
  if (e.key !== 'Enter') return
  if (!activeCard.value || finished.value) return
  if (currentStep.value.type === 'choice' && feedback.value === 'correct') {
    e.preventDefault()
    next()
  }
}
onMounted(() => window.addEventListener('keydown', handleGlobalKeydown))
onUnmounted(() => window.removeEventListener('keydown', handleGlobalKeydown))

function choiceBtnClass(i: number): string {
  if (feedback.value === 'correct' && i === currentStep.value.answer) return 'btn-success'
  if (chosenIndex.value === i && feedback.value === 'wrong') return 'btn-outline-danger'
  return 'btn-outline-dark'
}

function next() {
  if (!activeCard.value) return
  const p = { ...progress.value, [activeCard.value.id]: curStep.value + 1 }
  progress.value = p
  saveProgress(p)
  answer.value = ''
  feedback.value = 'idle'
  showHint.value = false
  chosenIndex.value = null
}

function prev() {
  if (!activeCard.value || curStep.value === 0) return
  const p = { ...progress.value, [activeCard.value.id]: curStep.value - 1 }
  progress.value = p
  saveProgress(p)
  answer.value = ''
  feedback.value = 'idle'
  showHint.value = false
  chosenIndex.value = null
}

function restart() {
  if (!activeCard.value) return
  const p = { ...progress.value, [activeCard.value.id]: 0 }
  progress.value = p
  saveProgress(p)
  answer.value = ''
  feedback.value = 'idle'
  showHint.value = false
  chosenIndex.value = null
}

// Minimal Markdown-ish renderer for lab reference notes (headings, bullet
// lists, bold, inline code, paragraphs). Not a full Markdown parser -- just
// enough for the short study-note blocks attached to each lab. Input is our
// own authored content (not untrusted user input).
function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}
function inlineMd(s: string): string {
  let out = escapeHtml(s)
  out = out.replace(/`([^`]+)`/g, '<code>$1</code>')
  out = out.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  return out
}
const referenceHtml = computed(() => {
  const md = activeCard.value?.payload.reference ?? ''
  const lines = md.split('\n')
  let html = ''
  let inList = false
  const closeList = () => { if (inList) { html += '</ul>'; inList = false } }
  for (const raw of lines) {
    const line = raw.trim()
    if (!line) { closeList(); continue }
    const h = line.match(/^(#{1,4})\s+(.*)$/)
    if (h) { closeList(); const lvl = Math.min(h[1].length + 3, 6); html += `<h${lvl}>${inlineMd(h[2])}</h${lvl}>`; continue }
    const li = line.match(/^[-*]\s+(.*)$/)
    if (li) { if (!inList) { html += '<ul>'; inList = true }; html += `<li>${inlineMd(li[1])}</li>`; continue }
    closeList()
    html += `<p>${inlineMd(line)}</p>`
  }
  closeList()
  return html
})

onMounted(loadLabs)
</script>

<style scoped>
.reference-notes :deep(h4),
.reference-notes :deep(h5),
.reference-notes :deep(h6) {
  font-weight: 700;
  margin-top: 0.75rem;
}
.reference-notes :deep(h4:first-child),
.reference-notes :deep(h5:first-child) {
  margin-top: 0;
}
.reference-notes :deep(p) { margin-bottom: 0.5rem; }
.reference-notes :deep(ul) { padding-left: 1.25rem; margin-bottom: 0.5rem; }
.reference-notes :deep(code) {
  background: #f1f3f5;
  padding: 0.1rem 0.3rem;
  border-radius: 0.25rem;
  font-size: 0.9em;
}
</style>
