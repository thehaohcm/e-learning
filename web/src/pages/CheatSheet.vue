<template>
  <div class="row justify-content-center">
    <div class="col-12">
      <div class="d-flex flex-wrap justify-content-between align-items-end mb-3 gap-2">
        <div>
          <h2 class="fw-bold mb-0"><i class="bi bi-journal-text"></i> Cheat sheet</h2>
          <small class="text-muted">{{ topicName }} · your own notes. Write items, then let AI condense them for review.</small>
        </div>
        <div class="d-flex gap-2">
          <button v-if="isPentest && canStudy" class="btn btn-outline-warning" @click="genModal?.open()">
            <i class="bi bi-stars"></i> Call AI
          </button>
          <button v-if="isPentest && canStudy" class="btn btn-warning" :disabled="studying" @click="studyNotes">
            <span v-if="studying" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-arrow-repeat"></i> Study
          </button>
          <button class="btn btn-outline-primary" @click="startNew">
            <i class="bi bi-plus-lg"></i> New note
          </button>
        </div>
      </div>

      <!-- toolbar: search + course filter + expand/collapse all -->
      <div v-if="!loading && notes.length" class="d-flex flex-wrap align-items-center gap-2 mb-3">
        <div class="input-group input-group-sm" style="max-width: 300px">
          <span class="input-group-text"><i class="bi bi-search"></i></span>
          <input v-model="search" class="form-control" placeholder="Tìm keyword trong cheatsheet…" />
          <button v-if="search" class="btn btn-outline-secondary" @click="search = ''"><i class="bi bi-x-lg"></i></button>
        </div>
        <div class="btn-group btn-group-sm" role="group">
          <button v-for="c in courseChips" :key="c" type="button" class="btn"
            :class="courseFilter === c ? 'btn-primary' : 'btn-outline-primary'"
            @click="courseFilter = courseFilter === c ? '' : c">{{ c || 'Tất cả' }}</button>
        </div>
        <button class="btn btn-sm btn-outline-secondary ms-auto" @click="toggleAll">
          <i class="bi" :class="allOpen ? 'bi-chevron-bar-up' : 'bi-chevron-bar-down'"></i>
          {{ allOpen ? 'Thu gọn tất cả' : 'Mở tất cả' }}
        </button>
        <small class="text-muted w-100">Hiện {{ filteredNotes.length }}/{{ notes.length }} mục</small>
      </div>

      <GenerateQuestionsModal ref="genModal" @done="onGenerated" />

      <div v-if="msg" class="alert py-2" :class="msgOk ? 'alert-success' : 'alert-danger'">{{ msg }}</div>

      <!-- editor popup (new or edit the selected item in place) -->
      <div v-if="editing" class="cs-overlay" @click.self="cancel">
        <div class="card cs-dialog border-0 shadow-lg">
          <div class="card-header bg-white d-flex align-items-center gap-2">
            <i class="bi" :class="form.id ? 'bi-pencil' : 'bi-plus-lg'"></i>
            <span class="fw-bold">{{ form.id ? 'Sửa mục' : 'Mục mới' }}</span>
            <button type="button" class="btn-close ms-auto" aria-label="Close" @click="cancel"></button>
          </div>
          <div class="card-body">
            <input v-model="form.title" class="form-control mb-2" maxlength="200" placeholder="Tiêu đề (vd “nmap basics”)" />
            <textarea v-model="form.body" class="form-control font-monospace" rows="10"
              placeholder="Markdown: # headings, - lists, **bold**, `code`, ```code blocks```"></textarea>
            <div v-if="form.body.trim()" class="md mt-2 p-2 rounded bg-light cs-preview" v-html="render(form.body)"></div>
            <div class="mt-3">
              <label class="form-label small mb-1"><i class="bi bi-link-45deg"></i> Links tài liệu <span class="text-muted">(để track sau này)</span></label>
              <div v-for="(lnk, i) in form.links" :key="i" class="input-group input-group-sm mb-1">
                <input v-model="lnk.label" class="form-control" style="max-width: 36%" placeholder="Nhãn (vd: HTB writeup)" />
                <input v-model="lnk.url" class="form-control" placeholder="https://…" />
                <button class="btn btn-outline-danger" type="button" title="Xóa link" @click="removeLink(i)"><i class="bi bi-x-lg"></i></button>
              </div>
              <button class="btn btn-sm btn-outline-secondary" type="button" @click="addLink"><i class="bi bi-plus-lg"></i> Thêm link</button>
            </div>
          </div>
          <div class="card-footer bg-white d-flex gap-2">
            <button class="btn btn-primary btn-sm" :disabled="busy || (!form.title.trim() && !form.body.trim())" @click="save">
              <span v-if="busy" class="spinner-border spinner-border-sm me-1"></span>
              <i v-else class="bi bi-check-lg"></i> {{ form.id ? 'Lưu' : 'Thêm' }}
            </button>
            <button class="btn btn-link btn-sm" :disabled="busy" @click="cancel">Hủy</button>
          </div>
        </div>
      </div>

      <div v-if="loading" class="text-center py-5"><div class="spinner-border text-primary"></div></div>

      <!-- notes -->
      <div v-else>
        <div v-if="!notes.length && !editing" class="text-center text-muted py-4">
          <i class="bi bi-journal-plus display-5"></i>
          <p class="mt-2">No notes yet. Add cheat-sheet items for this course.</p>
        </div>
        <div v-else-if="!filteredNotes.length" class="text-center text-muted py-4">
          <i class="bi bi-search display-6"></i>
          <p class="mt-2">Không có mục nào khớp bộ lọc.</p>
        </div>

        <div v-for="n in filteredNotes" :key="n.id" class="card border-0 shadow-sm mb-2">
          <div class="card-body py-2">
            <div class="d-flex align-items-center gap-2">
              <button class="btn btn-sm btn-link p-0 text-muted flex-shrink-0" :title="isOpen(n.id) ? 'Thu gọn' : 'Mở rộng'" @click="toggle(n.id)">
                <i class="bi" :class="isOpen(n.id) ? 'bi-chevron-down' : 'bi-chevron-right'"></i>
              </button>
              <h6 class="fw-bold mb-0 flex-grow-1 min-w-0 text-truncate cs-title" @click="toggle(n.id)">
                {{ n.title || firstLine(n.body) || '(không tiêu đề)' }}
              </h6>
              <div class="d-flex gap-1 flex-shrink-0">
                <button v-if="!filtering" class="btn btn-sm btn-link p-0 text-muted" title="Move up" :disabled="idxOf(n) === 0 || busy" @click="move(n, -1)"><i class="bi bi-arrow-up"></i></button>
                <button v-if="!filtering" class="btn btn-sm btn-link p-0 text-muted" title="Move down" :disabled="idxOf(n) === notes.length - 1 || busy" @click="move(n, 1)"><i class="bi bi-arrow-down"></i></button>
                <button class="btn btn-sm btn-link p-0" title="Sửa" @click="startEdit(n)"><i class="bi bi-pencil"></i></button>
                <button class="btn btn-sm btn-link p-0 text-danger" title="Xóa" @click="remove(n)"><i class="bi bi-trash"></i></button>
              </div>
            </div>
            <div v-if="isOpen(n.id)" class="md mt-2" v-html="render(n.body)"></div>
            <div v-if="isOpen(n.id) && n.links && n.links.length" class="mt-2 d-flex flex-wrap gap-2">
              <a v-for="(lnk, i) in n.links" :key="i" :href="lnk.url" target="_blank" rel="noopener noreferrer"
                class="badge rounded-pill text-bg-light border text-decoration-none cs-link">
                <i class="bi bi-link-45deg"></i> {{ lnk.label || lnk.url }}
              </a>
            </div>
          </div>
        </div>

        <!-- AI cheat sheet -->
        <div v-if="notes.length" class="card border-0 shadow-sm mt-4">
          <div class="card-body">
            <div class="d-flex flex-wrap justify-content-between align-items-center gap-2 mb-2">
              <h6 class="fw-bold mb-0"><i class="bi bi-stars text-warning"></i> AI cheat sheet</h6>
              <div class="d-flex gap-2">
                <button v-if="summary" class="btn btn-sm btn-outline-secondary" @click="copySummary"><i class="bi bi-clipboard"></i> Copy</button>
                <button class="btn btn-sm btn-primary" :disabled="summarizing" @click="summarize">
                  <span v-if="summarizing" class="spinner-border spinner-border-sm me-1"></span>
                  <i v-else class="bi bi-magic"></i> {{ summary ? 'Regenerate' : 'Make a cheat sheet' }}
                </button>
              </div>
            </div>
            <p v-if="!summary" class="small text-muted mb-0">
              AI reads all your notes above and writes one condensed cheat sheet for quick review. Your notes stay the source of truth.
            </p>
            <template v-else>
              <div class="md summary p-2 rounded bg-light" v-html="render(summary.summary_md)"></div>
              <div class="d-flex flex-wrap justify-content-between align-items-center gap-2 mt-2">
                <div class="small text-muted">Generated {{ fromNow(summary.generated_at) }} from your {{ notes.length }} note(s).</div>
                <button class="btn btn-sm btn-outline-success" :disabled="busy" @click="adopt">
                  <i class="bi bi-arrow-left-right"></i> Use this as my notes
                </button>
              </div>
            </template>
          </div>
        </div>

        <!-- Ask AI about this course -->
        <div class="card border-0 shadow-sm mt-4">
          <div class="card-body">
            <h6 class="fw-bold mb-2"><i class="bi bi-chat-dots text-primary"></i> Ask AI about {{ topicName }}</h6>
            <CourseChat />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, errorMessage } from '../api'
import { auth } from '../auth'
import { currentPrefix } from '../router'
import { loadSubjects, shortName, subject, subjectId, subjectSlug } from '../composables/useSubject'
import { renderMarkdown } from '../utils/markdown'
import CourseChat from '../components/CourseChat.vue'
import GenerateQuestionsModal from '../components/GenerateQuestionsModal.vue'

const router = useRouter()
const isPentest = computed(() => subjectSlug.value === 'pentest')
const canStudy = computed(() => auth.isAdmin) // studying materializes notes into shared cards (admin)
const studying = ref(false)
const genModal = ref<InstanceType<typeof GenerateQuestionsModal> | null>(null)

function onGenerated(p: { deckId: number; created: number }) {
  router.push(`/${currentPrefix()}/review?deck_id=${p.deckId}`)
}

interface NoteLink { url: string; label?: string }
interface Note { id: number; subject_id: number; title: string; body: string; sort_order: number; links?: NoteLink[] }
interface Summary { summary_md: string; generated_at: string }

const notes = ref<Note[]>([])
const summary = ref<Summary | null>(null)
const loading = ref(true)
const busy = ref(false)
const summarizing = ref(false)
const msg = ref('')
const msgOk = ref(false)

const editing = ref(false)
const form = reactive({ id: 0, title: '', body: '', links: [] as NoteLink[] })

const topicName = computed(() => (subject.value ? shortName(subject.value.slug, subject.value.name) : ''))
const render = (s: string) => renderMarkdown(s)

// ---- search + course filter + collapse (collapsed by default) ----
const search = ref('')
const courseFilter = ref('')
const courseChips = ['PT1', 'WEB1', 'CPTS']
const expanded = ref<Set<number>>(new Set())
const isOpen = (id: number) => expanded.value.has(id)
function toggle(id: number) {
  const nx = new Set(expanded.value)
  nx.has(id) ? nx.delete(id) : nx.add(id)
  expanded.value = nx
}
const filtering = computed(() => !!search.value.trim() || !!courseFilter.value)
const filteredNotes = computed(() => {
  const q = search.value.trim().toLowerCase()
  const c = courseFilter.value.toLowerCase()
  return notes.value.filter((n) => {
    const hay = (n.title + ' ' + n.body).toLowerCase()
    if (c && !hay.includes(c)) return false
    if (q && !hay.includes(q)) return false
    return true
  })
})
const allOpen = computed(() => filteredNotes.value.length > 0 && filteredNotes.value.every((n) => expanded.value.has(n.id)))
function toggleAll() {
  if (allOpen.value) { expanded.value = new Set() }
  else { expanded.value = new Set(filteredNotes.value.map((n) => n.id)) }
}
const idxOf = (n: Note) => notes.value.findIndex((x) => x.id === n.id)
function addLink() { form.links.push({ url: '', label: '' }) }
function removeLink(i: number) { form.links.splice(i, 1) }
function cleanLinks(): NoteLink[] {
  return form.links.map((l) => ({ url: (l.url || '').trim(), label: (l.label || '').trim() })).filter((l) => l.url)
}

// ---- Ask AI chat (ephemeral, in this page session) ----
function flash(m: string, ok = true) {
  msg.value = m
  msgOk.value = ok
  setTimeout(() => { if (msg.value === m) msg.value = '' }, 4000)
}

function fromNow(iso: string): string {
  const d = new Date(iso)
  const min = Math.round((Date.now() - d.getTime()) / 60000)
  if (min < 1) return 'just now'
  if (min < 60) return `${min} min ago`
  if (min < 1440) return `${Math.round(min / 60)} h ago`
  return d.toLocaleDateString()
}

async function load() {
  loading.value = true
  try {
    await loadSubjects()
    if (!subjectId.value) return
    const { data } = await api.get('/course-notes', { params: { subject_id: subjectId.value } })
    notes.value = data.notes ?? []
    summary.value = data.summary ?? null
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    loading.value = false
  }
}

function firstLine(s: string): string {
  return (s.split('\n').find((l) => l.trim()) ?? '').trim()
}

/**
 * Turn the notes into studyable flashcards (front = title or first line,
 * back = body) in this course's "cheatsheet" deck, then open Study for it.
 * Re-running syncs: new notes are added, changed bodies updated, SRS kept.
 */
async function studyNotes() {
  // nothing to study yet → offer to generate questions from the cheatsheet
  if (!notes.value.length) { genModal.value?.open(); return }
  const deck = subject.value?.decks?.find((d) => d.slug === 'cheatsheet')
  if (!deck) { flash('Chưa có deck Cheatsheet — deploy bản mới (migration 0052) để tạo.', false); return }
  studying.value = true
  try {
    const { data } = await api.get('/cards', { params: { deck_id: deck.id, kind: 'concept', limit: 500 } })
    const existing = new Map<string, { id: number; payload?: { back?: string } }>()
    for (const c of (data.data ?? [])) existing.set(String(c.front).trim().toLowerCase(), c)

    const ids: number[] = []
    for (const n of notes.value) {
      const front = (n.title.trim() || firstLine(n.body)).slice(0, 300)
      const back = n.body.trim()
      if (!front || !back) continue
      const cur = existing.get(front.toLowerCase())
      try {
        if (!cur) {
          const r = await api.post('/cards', { deck_id: deck.id, kind: 'concept', front, payload: { back } })
          ids.push(r.data.id)
        } else {
          if ((cur.payload?.back ?? '') !== back) {
            await api.put(`/cards/${cur.id}`, { front, kind: 'concept', payload: { back } })
          }
          ids.push(cur.id)
        }
      } catch { /* skip a single bad entry, keep going */ }
    }
    if (!ids.length) { flash('Không có mục nào để học (cần tiêu đề hoặc nội dung).', false); return }
    await api.post('/study/enroll', { card_ids: ids })
    router.push(`/${currentPrefix()}/review?deck_id=${deck.id}`)
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    studying.value = false
  }
}

function startNew() {
  Object.assign(form, { id: 0, title: '', body: '', links: [] })
  editing.value = true
}
function startEdit(n: Note) {
  Object.assign(form, { id: n.id, title: n.title, body: n.body, links: (n.links ?? []).map((l) => ({ ...l })) })
  editing.value = true
}
function cancel() {
  editing.value = false
}

async function save() {
  busy.value = true
  try {
    const links = cleanLinks()
    if (form.id) {
      await api.patch(`/course-notes/${form.id}`, { title: form.title, body: form.body, links })
    } else {
      await api.post('/course-notes', { subject_id: subjectId.value, title: form.title, body: form.body, links })
    }
    editing.value = false
    await reload()
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    busy.value = false
  }
}

async function remove(n: Note) {
  if (!confirm('Delete this note?')) return
  busy.value = true
  try {
    await api.delete(`/course-notes/${n.id}`)
    await reload()
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    busy.value = false
  }
}

/** Swap order with the neighbour and persist both. */
async function move(n: Note, dir: -1 | 1) {
  const idx = idxOf(n)
  const j = idx + dir
  if (idx < 0 || j < 0 || j >= notes.value.length) return
  const a = notes.value[idx]
  const b = notes.value[j]
  busy.value = true
  try {
    await Promise.all([
      api.patch(`/course-notes/${a.id}`, { sort_order: b.sort_order }),
      api.patch(`/course-notes/${b.id}`, { sort_order: a.sort_order }),
    ])
    await reload()
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    busy.value = false
  }
}

async function summarize() {
  summarizing.value = true
  try {
    const { data } = await api.post('/course-notes/summarize', { subject_id: subjectId.value }, { timeout: 180000 })
    summary.value = data
    flash('Cheat sheet ready.')
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    summarizing.value = false
  }
}

/** Replace the messy notes with the tidy AI version (split into items by heading). */
async function adopt() {
  if (!confirm('Replace your current notes with this cheat sheet? Your existing notes for this course will be removed.')) return
  busy.value = true
  try {
    const { data } = await api.post('/course-notes/adopt-summary', { subject_id: subjectId.value })
    notes.value = data.notes ?? []
    flash('Your notes now use the AI cheat sheet. You can keep editing each item.')
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    busy.value = false
  }
}

async function copySummary() {
  try {
    await navigator.clipboard.writeText(summary.value?.summary_md ?? '')
    flash('Copied to clipboard.')
  } catch {
    flash('Could not copy.', false)
  }
}

async function reload() {
  const { data } = await api.get('/course-notes', { params: { subject_id: subjectId.value } })
  notes.value = data.notes ?? []
  summary.value = data.summary ?? null
}

// switching topic in the menu reloads this course's notes
watch(subjectId, (id, old) => { if (id && id !== old) { editing.value = false; load() } })
onMounted(load)
</script>

<style scoped>
.min-w-0 { min-width: 0; }
.md :deep(pre) { background: #0f172a; color: #e2e8f0; padding: 0.6rem 0.8rem; border-radius: 8px; overflow-x: auto; font-size: 0.85rem; }
.md :deep(code) { background: rgba(2, 6, 23, 0.07); padding: 0.1em 0.3em; border-radius: 4px; font-size: 0.9em; }
.md :deep(pre code) { background: none; padding: 0; }
.md :deep(h1), .md :deep(h2), .md :deep(h3) { font-size: 1.1rem; font-weight: 700; margin: 0.6rem 0 0.3rem; }
.md :deep(ul), .md :deep(ol) { margin: 0.2rem 0 0.4rem 1.1rem; padding: 0; }
.md :deep(blockquote) { border-left: 3px solid #cbd5e1; margin: 0.3rem 0; padding-left: 0.7rem; color: #475569; }
.md :deep(p) { margin: 0.3rem 0; }
.md :deep(a) { color: #4f46e5; }
.md.summary :deep(hr) { margin: 0.6rem 0; }
.cs-title { cursor: pointer; }
.cs-title:hover { color: #4f46e5; }
.cs-link { font-weight: 500; }
.cs-overlay { position: fixed; inset: 0; z-index: 1060; background: rgba(15,23,42,0.5); display: flex; align-items: flex-start; justify-content: center; padding: 1.5rem 1rem; overflow-y: auto; }
.cs-dialog { width: 100%; max-width: 640px; border-radius: 14px; }
.cs-preview { max-height: 220px; overflow-y: auto; }
</style>
