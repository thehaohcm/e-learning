<template>
  <div class="notes-widget">
    <!-- "+ Note" next to selected text -->
    <button v-if="pick" type="button" class="btn btn-dark btn-sm shadow pick-btn" :style="{ left: pick.x + 'px', top: pick.y + 'px' }"
      @mousedown.prevent @click="noteSelection">
      <i class="bi bi-journal-plus"></i> Note
    </button>

    <div v-if="toast" class="toast-note shadow-sm">{{ toast }}</div>

    <!-- panel -->
    <div v-if="open" class="panel card shadow-lg border-0" role="dialog" aria-label="My notes" @keydown.esc="open = false">
      <div class="card-header bg-white d-flex align-items-center gap-2 py-2">
        <i class="bi bi-journal-text text-warning"></i>
        <span class="fw-bold">My notes</span>
        <small class="text-muted">Alt+N</small>
        <button type="button" class="btn-close ms-auto" aria-label="Close" @click="open = false"></button>
      </div>

      <div class="card-body p-2 border-bottom">
        <form class="d-flex gap-1" @submit.prevent="add">
          <input ref="wordBox" v-model="word" class="form-control form-control-sm" maxlength="60"
            placeholder="New word or phrase… (Enter)" autocomplete="off" autocapitalize="off" spellcheck="false" />
          <button class="btn btn-sm btn-warning" :disabled="!word.trim() || saving"><i class="bi bi-plus-lg"></i></button>
        </form>
        <input v-if="word.trim()" v-model="remark" class="form-control form-control-sm mt-1" maxlength="300"
          placeholder="Your note (optional)" @keydown.enter.prevent="add" />
        <div class="small text-muted mt-1">Tip: select a word anywhere on the page, then press <b>Note</b>.</div>
      </div>

      <div class="list flex-grow-1">
        <div v-if="loading" class="p-3 text-center"><span class="spinner-border spinner-border-sm"></span></div>
        <div v-else-if="!notes.length" class="p-3 small text-muted text-center">No notes yet.</div>
        <div v-for="n in notes" :key="n.id" class="note px-2 py-2 border-bottom">
          <div class="d-flex align-items-start gap-1">
            <div class="flex-grow-1 min-w-0">
              <span class="fw-semibold">{{ n.text }}</span>
              <span v-if="n.info?.base && n.info.base.toLowerCase() !== n.text.toLowerCase()" class="small text-muted"> → {{ n.info.base }}</span>
              <span v-if="n.info?.ipa" class="small text-muted ms-1">{{ n.info.ipa }}</span>
              <button type="button" class="btn btn-link btn-sm p-0 ms-1 align-baseline" title="Listen" @click="say(n.info?.base || n.text)">
                <i class="bi bi-volume-up"></i>
              </button>
              <div v-if="n.info?.meaning_vi" class="small">{{ n.info.meaning_vi }}</div>
              <div v-else class="small text-muted fst-italic">no meaning yet: press Explain</div>
              <div v-if="n.note" class="small text-secondary"><i class="bi bi-chat-left-text"></i> {{ n.note }}</div>
              <div v-if="n.context" class="small text-muted fst-italic context">“{{ n.context }}”</div>
              <div v-if="n.source" class="tiny text-muted">{{ n.source }}</div>
            </div>
            <div class="d-flex flex-column align-items-end gap-1 flex-shrink-0">
              <span v-if="n.status === 'added'" class="badge bg-success-subtle text-success"><i class="bi bi-check2"></i> In Study</span>
              <button v-else type="button" class="btn btn-outline-success btn-sm py-0" :disabled="busy" title="Add to Study" @click="addToStudy([n.id])">
                + Study
              </button>
              <button type="button" class="btn btn-link btn-sm p-0 text-muted" title="Delete" @click="remove(n)"><i class="bi bi-trash"></i></button>
            </div>
          </div>
        </div>
      </div>

      <div class="card-footer bg-white p-2">
        <div v-if="error" class="small text-danger mb-1">{{ error }}</div>
        <div v-if="info" class="small text-success mb-1">{{ info }}</div>
        <div class="d-flex gap-1">
          <button type="button" class="btn btn-sm btn-outline-primary flex-fill" :disabled="busy || !newCount" @click="fill">
            <span v-if="busy === 'fill'" class="spinner-border spinner-border-sm me-1"></span>
            <i v-else class="bi bi-stars"></i> Explain {{ newCount || '' }}
          </button>
          <button type="button" class="btn btn-sm btn-success flex-fill" :disabled="busy || !pendingCount" @click="addToStudy([])">
            <span v-if="busy === 'add'" class="spinner-border spinner-border-sm me-1"></span>
            <i v-else class="bi bi-plus-circle"></i> Add {{ pendingCount || '' }} to Study
          </button>
        </div>
        <div class="tiny text-muted mt-1">New words go to your private deck “My words” (Browse).</div>
      </div>
    </div>

    <!-- floating button -->
    <button type="button" class="btn btn-warning rounded-circle shadow fab" :title="open ? 'Close notes (Alt+N)' : 'My notes (Alt+N)'"
      :aria-expanded="open" @click="toggle">
      <i :class="open ? 'bi bi-x-lg' : 'bi bi-journal-text'"></i>
      <span v-if="!open && pendingCount" class="badge rounded-pill bg-danger count">{{ pendingCount }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, errorMessage } from '../api'
import { RATE_REPLAY, useSpeechSynthesis } from '../composables/useSpeechSynthesis'

interface NoteInfo { base?: string; ipa?: string; pos?: string; meaning_vi?: string; meaning_en?: string }
interface Note { id: number; text: string; note: string; context: string; source: string; card_id: number | null; status: 'new' | 'ready' | 'added'; info: NoteInfo }

const route = useRoute()
const { speakEnglish } = useSpeechSynthesis()

const open = ref(storageGet('elearn.notes.open') === 'true')
watch(open, (v) => {
  storageSet('elearn.notes.open', String(v))
  if (v) {
    load()
    focusBox()
  }
})
const notes = ref<Note[]>([])
const loading = ref(false)
const loaded = ref(false)
const word = ref('')
const remark = ref('')
const saving = ref(false)
const busy = ref<'' | 'fill' | 'add'>('')
const error = ref('')
const info = ref('')
const toast = ref('')
const wordBox = ref<HTMLInputElement | null>(null)

const newCount = computed(() => notes.value.filter((n) => n.status === 'new').length)
const pendingCount = computed(() => notes.value.filter((n) => n.status !== 'added').length)

function storageGet(k: string) {
  try { return localStorage.getItem(k) } catch { return null }
}
function storageSet(k: string, v: string) {
  try { localStorage.setItem(k, v) } catch { /* ignore */ }
}

/** Where the word was met, from the current page. */
function sourceLabel(): string {
  const p = route.path
  const label =
    p.startsWith('/review') ? 'Study' : p.startsWith('/quiz') ? 'Quiz' : p.startsWith('/reading') ? 'Reading'
      : p.startsWith('/writing') ? 'Writing' : p.startsWith('/practice') ? 'Practice' : p.startsWith('/learn') ? 'Browse' : 'Home'
  const title = p.startsWith('/reading') ? document.querySelector('h3')?.textContent?.trim() : ''
  return title ? `${label}: ${title}`.slice(0, 80) : label
}

function say(text: string) {
  speakEnglish(text, RATE_REPLAY).catch(() => {})
}

async function focusBox() {
  await nextTick()
  wordBox.value?.focus()
}

function toggle() {
  open.value = !open.value
}

function flash(msg: string) {
  toast.value = msg
  setTimeout(() => { if (toast.value === msg) toast.value = '' }, 2200)
}

async function load() {
  if (loading.value) return
  loading.value = !loaded.value
  try {
    notes.value = (await api.get('/notes')).data ?? []
    loaded.value = true
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

/** A note saved by another part of the app ("Ask about this"): show it at the top. */
function onSavedElsewhere(e: Event) {
  const n = (e as CustomEvent<Note>).detail
  if (n?.id) notes.value = [n, ...notes.value.filter((x) => x.id !== n.id)]
}

/** Saves a note; noting the same word again just moves it to the top. */
async function save(text: string, extra: { note?: string; context?: string }) {
  const { data } = await api.post('/notes', { text, source: sourceLabel(), ...extra })
  notes.value = [data, ...notes.value.filter((n) => n.id !== data.id)]
  return data as Note
}

async function add() {
  const text = word.value.trim()
  if (!text || saving.value) return
  saving.value = true
  error.value = ''
  info.value = ''
  try {
    await save(text, { note: remark.value })
    word.value = ''
    remark.value = ''
    focusBox()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

async function remove(n: Note) {
  notes.value = notes.value.filter((x) => x.id !== n.id)
  try {
    await api.delete(`/notes/${n.id}`)
  } catch (e) {
    error.value = errorMessage(e)
    load()
  }
}

/** One AI request explains every note without a meaning. */
async function fill() {
  busy.value = 'fill'
  error.value = ''
  info.value = ''
  try {
    const { data } = await api.post('/notes/fill')
    notes.value = data.notes ?? notes.value
    info.value = removedText(data.removed)
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    busy.value = ''
  }
}

/** Words the AI said are not English: they were deleted. */
function removedText(removed?: string[]): string {
  return removed?.length ? `Removed (not English words): ${removed.join(', ')}.` : ''
}

/** Enroll notes in Study (all pending ones when ids is empty); unexplained ones are explained first. */
async function addToStudy(ids: number[]) {
  busy.value = 'add'
  error.value = ''
  info.value = ''
  try {
    const { data } = await api.post('/notes/add', { ids })
    notes.value = data.notes ?? notes.value
    info.value = [
      data.added
        ? `${data.added} word${data.added > 1 ? 's' : ''} added to Study${data.created ? ` (${data.created} new in “My words”)` : ''}.`
        : data.removed?.length ? '' : 'Nothing added yet: try again in a moment.',
      removedText(data.removed),
    ].filter(Boolean).join(' ')
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    busy.value = ''
  }
}

// ---------------------------------------------------------------- select a word → "Note"

const pick = ref<{ x: number; y: number; text: string; context: string } | null>(null)

/** The sentence around the selection, from the text block that contains it. */
function sentenceAround(range: Range, text: string): string {
  let el: Node | null = range.commonAncestorContainer
  while (el && el.nodeType !== Node.ELEMENT_NODE) el = el.parentNode
  const block = (el as HTMLElement | null)?.closest('p, li, td, blockquote, .bubble, .sentence, .example, h5, div') ?? el
  const full = (block?.textContent ?? '').replace(/\s+/g, ' ').trim()
  const i = full.toLowerCase().indexOf(text.toLowerCase())
  if (i < 0) return ''
  const before = full.slice(0, i)
  const start = Math.max(before.search(/[^.!?]*$/), 0)
  const rest = full.slice(i)
  const endMatch = rest.match(/[.!?](\s|$)/)
  const end = endMatch ? i + (endMatch.index ?? 0) + 1 : full.length
  const s = full.slice(start, end).trim()
  return s.length > 400 ? '' : s
}

function onSelect(e: Event) {
  const target = e.target as HTMLElement | null
  if (target?.closest?.('.notes-widget, input, textarea, [contenteditable]')) return
  setTimeout(() => {
    const sel = window.getSelection()
    const text = sel?.toString().replace(/\s+/g, ' ').trim() ?? ''
    if (!sel || sel.rangeCount === 0 || !text || text.length > 60 || text.split(' ').length > 5 || !/[a-z]/i.test(text)) {
      pick.value = null
      return
    }
    const range = sel.getRangeAt(0)
    const r = range.getBoundingClientRect()
    if (!r.width && !r.height) return
    pick.value = {
      x: Math.min(Math.max(r.left + r.width / 2 - 34, 8), window.innerWidth - 90),
      y: r.top > 48 ? r.top - 40 : r.bottom + 8,
      text: text.replace(/^[^a-z0-9]+|[^a-z0-9]+$/gi, ''),
      context: sentenceAround(range, text),
    }
  }, 10)
}

async function noteSelection() {
  const p = pick.value
  pick.value = null
  if (!p?.text) return
  try {
    const n = await save(p.text, { context: p.context })
    window.getSelection()?.removeAllRanges()
    flash(n.status === 'added' ? `“${n.text}” is already in Study` : `Noted “${n.text}” ✓`)
  } catch (e) {
    flash(errorMessage(e))
  }
}

function hidePick() {
  pick.value = null
}

function onKey(e: KeyboardEvent) {
  if (e.altKey && (e.key === 'n' || e.key === 'N' || e.code === 'KeyN')) {
    e.preventDefault()
    open.value = !open.value
    if (open.value) focusBox()
  }
}

onMounted(() => {
  window.addEventListener('elearn:note-saved', onSavedElsewhere)
  document.addEventListener('mouseup', onSelect)
  document.addEventListener('touchend', onSelect)
  document.addEventListener('keyup', onSelect)
  window.addEventListener('scroll', hidePick, true)
  window.addEventListener('keydown', onKey)
  if (open.value) load()
  else api.get('/notes').then((r: { data: Note[] }) => { notes.value = r.data ?? []; loaded.value = true }).catch(() => {}) // for the badge
})
onUnmounted(() => {
  document.removeEventListener('mouseup', onSelect)
  document.removeEventListener('touchend', onSelect)
  document.removeEventListener('keyup', onSelect)
  window.removeEventListener('scroll', hidePick, true)
  window.removeEventListener('keydown', onKey)
  window.removeEventListener('elearn:note-saved', onSavedElsewhere)
})
</script>

<style scoped>
.fab {
  position: fixed;
  right: 20px;
  bottom: 20px;
  width: 54px;
  height: 54px;
  z-index: 1050;
  font-size: 1.35rem;
}
.fab .count {
  position: absolute;
  top: -4px;
  right: -4px;
  font-size: 0.7rem;
}
.panel {
  position: fixed;
  right: 20px;
  bottom: 86px;
  width: 360px;
  max-height: min(75vh, 640px);
  z-index: 1050;
  display: flex;
  flex-direction: column;
  border-radius: 16px !important;
  overflow: hidden;
}
.list {
  overflow-y: auto;
  min-height: 80px;
}
.note:last-child { border-bottom: 0 !important; }
.context {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.tiny { font-size: 0.72rem; }
.min-w-0 { min-width: 0; }
.pick-btn {
  position: fixed;
  z-index: 1060;
  border-radius: 999px;
  padding: 0.15rem 0.6rem;
}
.toast-note {
  position: fixed;
  right: 86px;
  bottom: 30px;
  z-index: 1060;
  background: #212529;
  color: #fff;
  border-radius: 8px;
  padding: 0.35rem 0.75rem;
  font-size: 0.85rem;
}
@media (max-width: 576px) {
  .panel { right: 8px; left: 8px; width: auto; bottom: 80px; }
  .fab { right: 14px; bottom: 14px; }
}
</style>
