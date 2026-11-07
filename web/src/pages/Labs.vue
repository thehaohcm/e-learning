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
            <button class="nav-link" :class="{ active: activeTab === 'mistakes' }" @click="activeTab = 'mistakes'">
              <i class="bi bi-arrow-repeat"></i> Ôn lại câu sai
              <span class="badge ms-1" :class="dueMistakes.length ? 'bg-danger' : 'bg-light text-dark'">{{ dueMistakes.length }}</span>
            </button>
          </li>
          <li class="nav-item">
            <button class="nav-link" :class="{ active: activeTab === 'assignment' }" @click="activeTab = 'assignment'">
              <i class="bi bi-flag"></i> Assignment <span class="badge bg-light text-dark ms-1">{{ assignmentCards.length }}</span>
            </button>
          </li>
          <li class="nav-item">
            <button class="nav-link" :class="{ active: activeTab === 'thm' }" @click="activeTab = 'thm'">
              <i class="bi bi-box-arrow-up-right"></i> THM Challenge <span class="badge bg-light text-dark ms-1">{{ thmChallenges.length }}</span>
            </button>
          </li>
          <li class="nav-item">
            <button class="nav-link" :class="{ active: activeTab === 'htb' }" @click="activeTab = 'htb'">
              <i class="bi bi-box-arrow-up-right"></i> HTB Challenge <span class="badge bg-light text-dark ms-1">{{ htbChallenges.length }}</span>
            </button>
          </li>
          <li class="nav-item">
            <button class="nav-link" :class="{ active: activeTab === 'cheatsheet' }" @click="activeTab = 'cheatsheet'">
              <i class="bi bi-journal-code"></i> Cheatsheet
            </button>
          </li>
        </ul>

        <!-- THM/HTB challenge tabs: static, forum-recommended external links -->
        <template v-if="activeTab === 'thm' || activeTab === 'htb'">
          <p class="small text-muted">
            Danh sách phòng/box được cộng đồng (r/tryhackme, r/hackthebox, le0pard, Viblo...) khuyên luyện trước khi thi
            {{ activeTab === 'thm' ? 'PT1/WEB1' : 'CPTS' }}. Một số link trỏ thẳng tới phòng (đã kiểm chứng URL), số còn lại
            trỏ qua tìm kiếm theo đúng tên (vì không chắc chắn được slug URL chính xác) để tránh link sai/gãy.
          </p>
          <div class="row g-3">
            <div v-for="c in (activeTab === 'thm' ? thmChallenges : htbChallenges)" :key="c.name" class="col-md-6">
              <div class="card border-0 shadow-sm h-100 hoverable">
                <div class="card-body d-flex flex-column">
                  <h6 class="fw-bold mb-1">{{ c.name }}</h6>
                  <p class="small text-muted mb-2">{{ c.category }}<span v-if="c.note"> — {{ c.note }}</span></p>
                  <div class="small mb-2 d-flex flex-wrap gap-1 align-items-center">
                    <span class="badge" :class="difficultyBadgeClass(c.difficulty)">{{ difficultyLabel(c.difficulty) }}</span>
                    <span v-for="tag in c.cert_tags" :key="tag" class="badge" :class="certBadgeClass(tag)">{{ tag }}</span>
                    <span v-if="c.search" class="badge bg-light text-dark border"><i class="bi bi-search"></i> link qua tìm kiếm</span>
                  </div>
                  <div class="d-flex flex-wrap gap-2 mt-auto">
                    <a :href="challengeUrl(c)" target="_blank" rel="noopener" class="btn btn-outline-dark">
                      Mở {{ c.platform }} <i class="bi bi-box-arrow-up-right"></i>
                    </a>
                    <a v-if="c.writeup_url" :href="c.writeup_url" target="_blank" rel="noopener" class="btn btn-outline-info">
                      <i class="bi bi-file-text"></i> Bài giải
                    </a>
                    <a v-if="c.youtube_url" :href="c.youtube_url" target="_blank" rel="noopener" class="btn btn-outline-danger">
                      <i class="bi bi-youtube"></i> YouTube
                    </a>
                    <div v-else-if="c.youtube_urls && c.youtube_urls.length" class="dropdown">
                      <button type="button" class="btn btn-outline-danger dropdown-toggle" data-bs-toggle="dropdown" aria-expanded="false">
                        <i class="bi bi-youtube"></i> YouTube
                      </button>
                      <ul class="dropdown-menu">
                        <li v-for="(yt, i) in c.youtube_urls" :key="i">
                          <a class="dropdown-item" :href="yt.url" target="_blank" rel="noopener">{{ yt.label }}</a>
                        </li>
                      </ul>
                    </div>
                    <router-link v-if="simCardFor(c)" :to="`/labs?id=${simCardFor(c)!.id}`" class="btn btn-success">
                      <i class="bi bi-joystick"></i> Làm Simulation
                      <span class="badge bg-light text-dark ms-1">{{ simCardFor(c)!.payload.steps?.length ?? 0 }}</span>
                    </router-link>
                    <button type="button" class="btn btn-outline-secondary" @click="manageTarget = c">
                      <i class="bi bi-pencil-square"></i> Quản lý câu hỏi
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- mistakes tab: lab steps answered wrong or given up on, resurfaced on a Leitner
             schedule (10 min -> 1 day -> 3 -> 7 -> 16 days) like vocab's spaced repetition,
             except local to this browser (lab cards never go into the server-side FSRS). -->
        <template v-else-if="activeTab === 'mistakes'">
          <p class="small text-muted">
            Câu đã gõ sai hoặc bấm "Không nhớ" quay lại đây theo lịch ôn tập ngắn → dài, giống cách ôn từ vựng.
            Chỉ lưu trên trình duyệt này.
          </p>
          <p v-if="!dueMistakes.length" class="text-muted small text-center py-4">
            <i class="bi bi-emoji-smile display-6 d-block mb-2"></i> Không có câu nào cần ôn lại bây giờ 🎉
          </p>
          <div v-else class="d-flex flex-column gap-2">
            <div v-for="m in dueMistakes" :key="mistakeKey(m)" class="card border-0 shadow-sm">
              <div class="card-body py-2">
                <div class="small text-muted mb-1">{{ mistakeCard(m)?.front }}</div>
                <p class="fw-semibold mb-2">{{ mistakeStep(m)?.prompt }}</p>
                <div class="input-group input-group-sm mb-1">
                  <span class="input-group-text font-monospace">$</span>
                  <input v-model="mistakeAnswers[mistakeKey(m)]" class="form-control font-monospace" autocomplete="off" autocapitalize="off" spellcheck="false"
                    placeholder="gõ lệnh…" :class="{ 'is-invalid': mistakeFeedback[mistakeKey(m)] === 'wrong' }"
                    @keydown.enter.prevent="checkMistake(m)" />
                  <button class="btn btn-outline-primary" type="button" @click="checkMistake(m)">Kiểm tra</button>
                </div>
                <div v-if="mistakeFeedback[mistakeKey(m)] === 'wrong'" class="small text-danger">
                  <i class="bi bi-x-circle"></i> Chưa đúng, thử lại.
                  <button class="btn btn-sm btn-link p-0 ms-1" @click="mistakeShowAnswer[mistakeKey(m)] = true">Xem đáp án</button>
                </div>
                <div v-if="mistakeShowAnswer[mistakeKey(m)]" class="small text-muted mt-1">
                  Đáp án mẫu: <code>{{ mistakeStep(m)?.explanation || mistakeStep(m)?.patterns?.[0] }}</code>
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- cheatsheet tab: the learner's own notes (shared with the Cheat sheet page + floating Notes widget) -->
        <template v-else-if="activeTab === 'cheatsheet'">
          <div class="d-flex flex-wrap justify-content-between align-items-center gap-2 mb-2">
            <p class="small text-muted mb-0">
              Ghi chú lệnh của bạn — dùng chung với trang <b>Cheat sheet</b> và nút <b>Notes</b> nổi. Sửa/thêm ở Cheat sheet.
            </p>
            <router-link :to="`/${prefix}/cheatsheet`" class="btn btn-sm btn-outline-primary">
              <i class="bi bi-pencil-square"></i> Sửa / thêm · Study
            </router-link>
          </div>
          <div v-if="notesLoading" class="text-muted small"><span class="spinner-border spinner-border-sm me-1"></span>Đang tải…</div>
          <p v-else-if="!notes.length" class="text-muted small">
            Chưa có ghi chú nào. Thêm ở <router-link :to="`/${prefix}/cheatsheet`">Cheat sheet</router-link> hoặc nút Notes nổi (góc dưới-phải).
          </p>
          <div v-else>
            <div class="d-flex flex-wrap align-items-center gap-2 mb-2">
              <div class="input-group input-group-sm" style="max-width: 280px">
                <span class="input-group-text"><i class="bi bi-search"></i></span>
                <input v-model="noteSearch" class="form-control" placeholder="Tìm keyword…" />
                <button v-if="noteSearch" class="btn btn-outline-secondary" @click="noteSearch = ''"><i class="bi bi-x-lg"></i></button>
              </div>
              <div class="btn-group btn-group-sm" role="group">
                <button v-for="c in noteCourseChips" :key="c" type="button" class="btn"
                  :class="noteCourse === c ? 'btn-primary' : 'btn-outline-primary'"
                  @click="noteCourse = noteCourse === c ? '' : c">{{ c }}</button>
              </div>
              <button class="btn btn-sm btn-outline-secondary ms-auto" @click="toggleAllNotes">
                <i class="bi" :class="allNotesOpen ? 'bi-chevron-bar-up' : 'bi-chevron-bar-down'"></i>
                {{ allNotesOpen ? 'Thu gọn tất cả' : 'Mở tất cả' }}
              </button>
              <small class="text-muted w-100">Hiện {{ filteredNotes.length }}/{{ notes.length }} mục</small>
            </div>
            <p v-if="!filteredNotes.length" class="text-muted small text-center py-3">Không có mục nào khớp bộ lọc.</p>
            <div class="d-flex flex-column gap-2">
              <div v-for="n in filteredNotes" :key="n.id" class="card border-0 shadow-sm">
                <div class="card-body py-2">
                  <div class="d-flex align-items-center gap-2">
                    <button class="btn btn-sm btn-link p-0 text-muted flex-shrink-0" @click="toggleNote(n.id)">
                      <i class="bi" :class="isNoteOpen(n.id) ? 'bi-chevron-down' : 'bi-chevron-right'"></i>
                    </button>
                    <h6 class="fw-bold mb-0 flex-grow-1 text-truncate cs-title" @click="toggleNote(n.id)">
                      {{ n.title || noteFirstLine(n.body) || '(không tiêu đề)' }}
                    </h6>
                  </div>
                  <div v-if="isNoteOpen(n.id)" class="reference-notes md mt-2" v-html="renderMarkdown(n.body)"></div>
                  <div v-if="isNoteOpen(n.id) && n.links && n.links.length" class="mt-2 d-flex flex-wrap gap-2">
                    <a v-for="(lnk, i) in n.links" :key="i" :href="lnk.url" target="_blank" rel="noopener noreferrer"
                      class="badge rounded-pill text-bg-light border text-decoration-none">
                      <i class="bi bi-link-45deg"></i> {{ lnk.label || lnk.url }}
                    </a>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>

        <template v-else>
          <div v-if="!visibleCards.length" class="small text-muted">Chưa có bài nào ở mục này.</div>

          <div class="row g-3">
            <div v-for="c in visibleCards" :key="c.id" class="col-md-6">
              <div class="card border-0 shadow-sm h-100 hoverable">
                <div class="card-body d-flex flex-column">
                  <h5 class="fw-bold mb-1">{{ c.front }}</h5>
                  <p class="small text-muted mb-2">{{ c.payload.scenario }}</p>
                <div class="small mb-2 d-flex flex-wrap gap-1 align-items-center">
                  <span class="badge" :class="progressBadgeClass(c)">
                    {{ progressLabel(c) }}
                  </span>
                  <span v-if="c.payload.difficulty" class="badge" :class="difficultyBadgeClass(c.payload.difficulty)">
                    {{ difficultyLabel(c.payload.difficulty) }}
                  </span>
                  <span v-for="tag in (c.payload.cert_tags ?? [])" :key="tag" class="badge" :class="certBadgeClass(tag)">
                    {{ tag }}
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
      </template>
    </div>

    <!-- runner view -->
    <div v-else>
      <router-link to="/labs" class="btn btn-sm btn-outline-secondary mb-3"><i class="bi bi-arrow-left"></i> Danh sách lab</router-link>

      <h3 class="fw-bold mb-1">
        {{ activeCard.front }}
        <span v-if="activeCard.payload.mode === 'assignment'" class="badge bg-dark ms-1">Assignment</span>
        <span v-if="activeCard.payload.mode === 'challenge_sim'" class="badge bg-success ms-1"><i class="bi bi-joystick"></i> Simulation</span>
        <span v-if="activeCard.payload.difficulty" class="badge ms-1" :class="difficultyBadgeClass(activeCard.payload.difficulty)">
          {{ difficultyLabel(activeCard.payload.difficulty) }}
        </span>
        <span v-for="tag in (activeCard.payload.cert_tags ?? [])" :key="tag" class="badge ms-1" :class="certBadgeClass(tag)">
          {{ tag }}
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
          <div v-if="contextOutput" class="mb-2">
            <div class="small text-muted mb-1"><i class="bi bi-terminal"></i> Kết quả bước trước:</div>
            <pre class="bg-dark text-light p-2 rounded small mb-0" style="white-space: pre-wrap;">{{ contextOutput }}</pre>
          </div>

          <p class="fw-semibold">{{ currentStep.prompt }}</p>

          <div v-if="showHint && currentStep.hint" class="alert alert-light border small mb-2" style="white-space: pre-wrap;">
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
          <div v-else-if="feedback === 'correct'">
            <div v-if="currentStep.type !== 'choice' && currentStep.output" class="mb-2">
              <div class="small text-muted mb-1"><i class="bi bi-terminal-fill text-success"></i> Kết quả thực thi lệnh:</div>
              <pre class="bg-dark text-light p-2 rounded small mb-0" style="white-space: pre-wrap;">{{ currentStep.output }}</pre>
            </div>
            <div class="alert alert-success py-2 small mb-0">
              <i class="bi bi-check-circle"></i> {{ currentStep.explanation || 'Chính xác!' }}
              <div class="mt-2">
                <button ref="nextStepBtn" class="btn btn-sm btn-success" @click="next">Bước tiếp theo <i class="bi bi-arrow-right"></i></button>
              </div>
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
          <div class="card-body reference-notes md" v-html="referenceHtml"></div>
        </div>
      </div>
    </div>

    <ChallengeQuestionManager v-if="manageTarget" :challenge="manageTarget" :deck-id="labsDeckId"
      :card="simCardFor(manageTarget)" @close="manageTarget = null" @saved="onSimSaved" @deleted="onSimDeleted" />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../api'
import { loadSubjects, subject, subjectId } from '../composables/useSubject'
import { currentPrefix } from '../router'
import { renderMarkdown } from '../utils/markdown'
import type { Card, LabStep } from '../types'
import { challengeKey, challengeUrl, htbChallenges, thmChallenges, type Challenge } from '../data/pentestChallenges'
import ChallengeQuestionManager from '../components/ChallengeQuestionManager.vue'
import { dueLabMistakes, recordLabMistake, resolveLabMistake, type LabMistake } from '../composables/labMistakes'

const route = useRoute()

const loading = ref(true)
const labCards = ref<Card[]>([])

const activeTab = ref<'lab' | 'mistakes' | 'assignment' | 'thm' | 'htb' | 'cheatsheet'>(route.query.tab === 'mistakes' ? 'mistakes' : 'lab')

// Cheatsheet tab = the learner's own course notes (shared with the Cheat sheet
// page and the floating Notes widget), so all three stay in sync.
const prefix = currentPrefix()
interface CheatNote { id: number; title: string; body: string; links?: { url: string; label?: string }[] }
const notes = ref<CheatNote[]>([])
// cheatsheet tab UX (same as the Cheat sheet page): search + course filter + collapse
const noteSearch = ref('')
const noteCourse = ref('')
const noteCourseChips = ['PT1', 'WEB1', 'CPTS']
const noteOpen = ref<Set<number>>(new Set())
const isNoteOpen = (id: number) => noteOpen.value.has(id)
function toggleNote(id: number) {
  const nx = new Set(noteOpen.value)
  nx.has(id) ? nx.delete(id) : nx.add(id)
  noteOpen.value = nx
}
const noteFirstLine = (t: string) => (t.split('\n').find((l) => l.trim()) ?? '').trim()
const filteredNotes = computed(() => {
  const q = noteSearch.value.trim().toLowerCase()
  const c = noteCourse.value.toLowerCase()
  return notes.value.filter((n) => {
    const hay = (n.title + ' ' + n.body).toLowerCase()
    if (c && !hay.includes(c)) return false
    if (q && !hay.includes(q)) return false
    return true
  })
})
const allNotesOpen = computed(() => filteredNotes.value.length > 0 && filteredNotes.value.every((n) => noteOpen.value.has(n.id)))
function toggleAllNotes() {
  noteOpen.value = allNotesOpen.value ? new Set() : new Set(filteredNotes.value.map((n) => n.id))
}
const notesLoading = ref(false)
let notesLoaded = false
async function loadNotes() {
  if (!subjectId.value) return
  notesLoading.value = true
  try {
    const { data } = await api.get('/course-notes', { params: { subject_id: subjectId.value } })
    notes.value = data.notes ?? []
    notesLoaded = true
  } catch { /* keep what we have */ } finally {
    notesLoading.value = false
  }
}
watch(activeTab, (t) => { if (t === 'cheatsheet' && !notesLoaded) loadNotes() })
// "challenge_sim" cards are the user's own self-quiz per THM/HTB challenge (see
// ChallengeQuestionManager) — they live in the same "labs" deck but are kept
// out of the Lab/Assignment tabs and surfaced only on their own challenge card.
const labOnlyCards = computed(() => labCards.value.filter((c) => c.payload.mode !== 'assignment' && c.payload.mode !== 'challenge_sim'))
const assignmentCards = computed(() => labCards.value.filter((c) => c.payload.mode === 'assignment'))
const visibleCards = computed(() => (activeTab.value === 'assignment' ? assignmentCards.value : labOnlyCards.value))

const labsDeckId = computed<number | null>(() => subject.value?.decks.find((d) => d.slug === 'labs')?.id ?? null)
const simCardByKey = computed<Record<string, Card>>(() => {
  const out: Record<string, Card> = {}
  for (const c of labCards.value) {
    if (c.payload.mode === 'challenge_sim' && c.payload.challenge_key) out[c.payload.challenge_key] = c
  }
  return out
})
function simCardFor(c: Challenge): Card | undefined {
  return simCardByKey.value[challengeKey(c)]
}

const manageTarget = ref<Challenge | null>(null)
function onSimSaved(card: Card) {
  const idx = labCards.value.findIndex((c) => c.id === card.id)
  if (idx >= 0) labCards.value[idx] = card
  else labCards.value.push(card)
  manageTarget.value = null
}
function onSimDeleted(cardId: number) {
  labCards.value = labCards.value.filter((c) => c.id !== cardId)
  manageTarget.value = null
}

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

function certBadgeClass(cert: string): string {
  switch (cert) {
    case 'THM PT1': return 'bg-primary'
    case 'THM WEB1': return 'bg-success'
    case 'HTB CPTS': return 'bg-dark'
    default: return 'bg-secondary'
  }
}

async function loadLabs() {
  await loadSubjects()
  const deck = subject.value?.decks.find((d) => d.slug === 'labs')
  if (!deck) { labCards.value = []; loading.value = false; return }
  const all: Card[] = []
  let page = 1
  const limit = 100
  while (true) {
    const { data } = await api.get('/cards', { params: { deck_id: deck.id, kind: 'lab', page, limit } })
    const rows: Card[] = data?.data ?? []
    all.push(...rows)
    const total = Number(data?.total ?? 0)
    if (!rows.length || all.length >= total || page >= 10) break
    page++
  }
  labCards.value = all.filter((c: Card) => c.kind === 'lab')
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
const contextOutput = computed<string>(() => {
  if (currentStep.value.type === 'choice' && currentStep.value.output) {
    return currentStep.value.output
  }
  if (curStep.value > 0) {
    return activeCard.value?.payload.steps?.[curStep.value - 1]?.output ?? ''
  }
  return ''
})
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

/**
 * Order-independent flag matching. A pattern written as `@flags:<cmd>:<letters>`
 * is correct when the input starts with <cmd> and contains every short-flag
 * letter, in ANY order or grouping — so `ss -tunlp`, `ss -utpln`, `ss -t -u -n -l -p`
 * and `ss -tun -lp` all pass `@flags:ss:tunlp`. Extra flags are allowed.
 */
function flagSpecMatch(spec: string, input: string): boolean {
  const parts = spec.split(':')
  const cmd = (parts[1] ?? '').toLowerCase()
  const letters = (parts[2] ?? '').toLowerCase()
  const toks = input.trim().toLowerCase().split(/\s+/)
  if (!cmd || toks[0] !== cmd) return false
  const have = new Set<string>()
  for (const t of toks.slice(1)) {
    if (/^-[a-z0-9]+$/.test(t)) for (const ch of t.slice(1)) have.add(ch)
  }
  for (const ch of letters) if (!have.has(ch)) return false
  return true
}

function matchPattern(p: string, norm: string): boolean {
  if (p.startsWith('@flags:')) return flagSpecMatch(p, norm)
  try { return new RegExp(p, 'i').test(norm) } catch { return false }
}

const nextStepBtn = ref<HTMLButtonElement | null>(null)
/** After a correct answer, move focus to "Bước tiếp theo" so Enter/Space
 *  advances without reaching for the mouse. */
function focusNextStep() {
  nextTick(() => nextStepBtn.value?.focus())
}

function check() {
  if (!activeCard.value) return
  const norm = normalize(answer.value)
  const ok = (currentStep.value.patterns ?? []).some((p) => matchPattern(p, norm))
  feedback.value = ok ? 'correct' : 'wrong'
  if (ok) {
    resolveLabMistake(activeCard.value.id, curStep.value)
    focusNextStep()
  } else {
    recordLabMistake(activeCard.value.id, curStep.value)
  }
  refreshMistakes()
}

function checkChoice(i: number) {
  if (!activeCard.value || feedback.value === 'correct') return
  chosenIndex.value = i
  const ok = i === currentStep.value.answer
  feedback.value = ok ? 'correct' : 'wrong'
  if (ok) {
    resolveLabMistake(activeCard.value.id, curStep.value)
    focusNextStep()
  } else {
    recordLabMistake(activeCard.value.id, curStep.value)
  }
  refreshMistakes()
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

// --- "Ôn lại câu sai" tab: local Leitner review queue for wrong/given-up steps ---
const dueMistakes = ref<LabMistake[]>([])
function refreshMistakes() { dueMistakes.value = dueLabMistakes() }
const mistakeAnswers = reactive<Record<string, string>>({})
const mistakeFeedback = reactive<Record<string, 'wrong' | undefined>>({})
const mistakeShowAnswer = reactive<Record<string, boolean>>({})
function mistakeKey(m: LabMistake): string { return `${m.cardId}:${m.stepIndex}` }
function mistakeCard(m: LabMistake): Card | undefined { return labCards.value.find((c) => c.id === m.cardId) }
function mistakeStep(m: LabMistake) { return mistakeCard(m)?.payload.steps?.[m.stepIndex] }
function checkMistake(m: LabMistake) {
  const k = mistakeKey(m)
  const step = mistakeStep(m)
  const norm = normalize(mistakeAnswers[k] ?? '')
  const ok = !!step && (step.patterns ?? []).some((p) => matchPattern(p, norm))
  if (ok) {
    resolveLabMistake(m.cardId, m.stepIndex)
    mistakeFeedback[k] = undefined
    mistakeShowAnswer[k] = false
    delete mistakeAnswers[k]
    refreshMistakes()
  } else {
    recordLabMistake(m.cardId, m.stepIndex)
    mistakeFeedback[k] = 'wrong'
  }
}
let mistakesTimer: number | undefined

const referenceHtml = computed(() => renderMarkdown(activeCard.value?.payload.reference ?? ''))

onMounted(loadLabs)
onMounted(() => {
  refreshMistakes()
  mistakesTimer = window.setInterval(refreshMistakes, 30_000)
})
onUnmounted(() => clearInterval(mistakesTimer))
</script>

<style scoped>
.reference-notes :deep(h1),
.reference-notes :deep(h2),
.reference-notes :deep(h3),
.reference-notes :deep(h4),
.reference-notes :deep(h5),
.reference-notes :deep(h6) {
  font-weight: 700;
  margin-top: 1.25rem;
  margin-bottom: 0.5rem;
  color: #0f172a;
}
.cs-title { cursor: pointer; }
.cs-title:hover { color: #4f46e5; }
.reference-notes :deep(h1) { font-size: 1.25rem; }
.reference-notes :deep(h2) { font-size: 1.15rem; }
.reference-notes :deep(h3) { font-size: 1.05rem; }
.reference-notes :deep(h4),
.reference-notes :deep(h5),
.reference-notes :deep(h6) { font-size: 0.95rem; }
.reference-notes :deep(h1:first-child),
.reference-notes :deep(h2:first-child),
.reference-notes :deep(h3:first-child),
.reference-notes :deep(h4:first-child) {
  margin-top: 0;
}
.reference-notes :deep(p) {
  margin-bottom: 0.6rem;
  line-height: 1.6;
  color: #334155;
}
.reference-notes :deep(ul),
.reference-notes :deep(ol) {
  padding-left: 1.35rem;
  margin-bottom: 0.6rem;
  color: #334155;
}
.reference-notes :deep(li) {
  margin-bottom: 0.25rem;
  line-height: 1.55;
}
.reference-notes :deep(strong) {
  font-weight: 600;
  color: #0f172a;
}
.reference-notes :deep(code) {
  background: #f1f5f9;
  color: #d63384;
  padding: 0.15em 0.4em;
  border-radius: 4px;
  font-size: 0.875em;
  font-family: SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}
.reference-notes :deep(pre) {
  background: #0f172a;
  color: #e2e8f0;
  padding: 0.85rem 1rem;
  border-radius: 8px;
  overflow-x: auto;
  font-size: 0.85rem;
  line-height: 1.55;
  margin: 0.6rem 0 0.85rem;
  font-family: SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.2);
}
.reference-notes :deep(pre code) {
  background: transparent !important;
  color: #f1f5f9 !important;
  padding: 0 !important;
  border-radius: 0;
  font-size: inherit;
  font-family: inherit;
}
.reference-notes :deep(table) {
  width: 100%;
  margin: 0.75rem 0;
  border-collapse: collapse;
  font-size: 0.875rem;
  background: #fff;
}
.reference-notes :deep(th),
.reference-notes :deep(td) {
  border: 1px solid #e2e8f0;
  padding: 0.5rem 0.75rem;
  vertical-align: middle;
}
.reference-notes :deep(th) {
  background: #f8fafc;
  font-weight: 600;
  color: #1e293b;
  border-bottom: 2px solid #cbd5e1;
}
.reference-notes :deep(tbody tr:nth-child(even)) {
  background: #f8fafc;
}
.reference-notes :deep(tbody tr:hover) {
  background: #f1f5f9;
}
.reference-notes :deep(blockquote) {
  border-left: 3px solid #6366f1;
  margin: 0.6rem 0;
  padding: 0.4rem 0.8rem;
  background: #f8fafc;
  color: #475569;
  border-radius: 0 4px 4px 0;
}
.reference-notes :deep(hr) {
  border: 0;
  border-top: 1px solid #e2e8f0;
  margin: 1rem 0;
}
</style>
