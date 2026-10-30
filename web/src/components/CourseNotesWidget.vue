<template>
  <div class="course-notes">
    <!-- panel -->
    <div v-if="open" class="panel card shadow-lg border-0" :class="{ raised }" role="dialog" aria-label="Course notes">
      <div class="card-header bg-white d-flex align-items-center gap-2 py-2">
        <i class="bi bi-journal-text text-success"></i>
        <span class="fw-bold">Notes</span>
        <small class="text-muted text-truncate">· {{ topicName }}</small>
        <button type="button" class="btn-close ms-auto" aria-label="Close" @click="open = false"></button>
      </div>

      <!-- quick add -->
      <div class="card-body p-2 border-bottom">
        <input v-model="title" class="form-control form-control-sm mb-1" maxlength="200" placeholder="Title (optional, e.g. “AD commands”)" />
        <textarea v-model="body" class="form-control form-control-sm font-monospace" rows="4"
          placeholder="Paste commands or notes… Markdown ok (```code```). Ctrl/⌘+Enter to add." @keydown="onKey"></textarea>
        <div class="d-flex align-items-center gap-2 mt-1">
          <button class="btn btn-sm btn-success" :disabled="busy || !body.trim()" @click="add">
            <span v-if="busy" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-plus-lg"></i> Add note
          </button>
          <router-link class="btn btn-sm btn-link ms-auto p-0" to="/cheatsheet" @click="open = false">Open Cheat sheet →</router-link>
        </div>
        <div v-if="err" class="small text-danger mt-1">{{ err }}</div>
      </div>

      <!-- recent notes -->
      <div class="card-body p-2 notes-list">
        <div v-if="loading" class="text-center py-3"><span class="spinner-border spinner-border-sm"></span></div>
        <p v-else-if="!notes.length" class="small text-muted mb-0 text-center py-2">No notes yet for {{ topicName }}. Add the first above.</p>
        <div v-else>
          <div v-for="n in notes" :key="n.id" class="note border-bottom pb-1 mb-1">
            <div class="d-flex align-items-start gap-1">
              <div class="flex-grow-1 min-w-0">
                <div v-if="n.title" class="fw-semibold small text-truncate">{{ n.title }}</div>
                <pre class="body small mb-0">{{ n.body }}</pre>
              </div>
              <button class="btn btn-sm btn-link p-0 text-danger flex-shrink-0" title="Delete" @click="remove(n)"><i class="bi bi-trash"></i></button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- floating button -->
    <button type="button" class="btn btn-success rounded-circle shadow fab" :class="{ raised }"
      :title="open ? 'Close notes' : 'Quick notes for this course'" :aria-expanded="open" @click="toggle">
      <i :class="open ? 'bi bi-x-lg' : 'bi bi-journal-text'"></i>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { api, errorMessage } from '../api'
import { isEnglish, shortName, subject, subjectId } from '../composables/useSubject'

// A floating quick-notes panel (like the Ask AI widget) that writes to the
// current course's notes — the same store as the Cheat sheet page. Lets the
// learner jot commands/notes on any page without leaving it.

interface Note { id: number; title: string; body: string; sort_order: number }

const KEY = 'elearn.notes.open'
function storageGet(k: string) { try { return localStorage.getItem(k) } catch { return null } }
function storageSet(k: string, v: string) { try { localStorage.setItem(k, v) } catch { /* ignore */ } }

const open = ref(storageGet(KEY) === 'true')
const notes = ref<Note[]>([])
const title = ref('')
const body = ref('')
const loading = ref(false)
const busy = ref(false)
const err = ref('')

const topicName = computed(() => (subject.value ? shortName(subject.value.slug, subject.value.name) : 'this course'))
// the English-only word-notes widget sits bottom-right too: lift this one above it
const raised = computed(() => isEnglish.value)

watch(open, (v) => storageSet(KEY, String(v)))

async function toggle() {
  open.value = !open.value
  if (open.value) await load()
}

async function load() {
  if (!subjectId.value) return
  loading.value = true
  try {
    const { data } = await api.get('/course-notes', { params: { subject_id: subjectId.value } })
    notes.value = data.notes ?? []
  } catch { /* offline: keep what we have */ } finally {
    loading.value = false
  }
}

function onKey(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') { e.preventDefault(); add() }
}

async function add() {
  if (!body.value.trim() || !subjectId.value) return
  busy.value = true
  err.value = ''
  try {
    await api.post('/course-notes', { subject_id: subjectId.value, title: title.value.trim(), body: body.value })
    title.value = ''
    body.value = ''
    await load()
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

async function remove(n: Note) {
  if (!confirm('Delete this note?')) return
  try {
    await api.delete(`/course-notes/${n.id}`)
    notes.value = notes.value.filter((x) => x.id !== n.id)
  } catch (e) {
    err.value = errorMessage(e)
  }
}

// switching course: reload its notes if the panel is open
watch(subjectId, (id, old) => { if (id !== old && open.value) load() })
</script>

<style scoped>
.fab {
  position: fixed;
  right: 20px;
  bottom: 20px;
  width: 54px;
  height: 54px;
  z-index: 1050;
  font-size: 1.3rem;
}
.fab.raised { bottom: 84px; } /* clear the word-notes widget on English */
.panel {
  position: fixed;
  right: 20px;
  bottom: 86px;
  width: 360px;
  max-height: min(78vh, 640px);
  z-index: 1050;
  display: flex;
  flex-direction: column;
  border-radius: 16px !important;
  overflow: hidden;
}
.panel.raised { bottom: 150px; }
.notes-list { overflow-y: auto; }
.note:last-child { border-bottom: 0 !important; margin-bottom: 0 !important; }
.body { white-space: pre-wrap; word-break: break-word; font-family: var(--bs-font-monospace); margin: 0; color: #334155; max-height: 7.5em; overflow: hidden; }
@media (max-width: 576px) {
  .panel { right: 8px; left: 8px; width: auto; bottom: 80px; }
  .panel.raised { bottom: 140px; }
  .fab { right: 14px; bottom: 14px; }
  .fab.raised { bottom: 78px; }
}
</style>
