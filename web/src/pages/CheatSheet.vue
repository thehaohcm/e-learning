<template>
  <div class="row justify-content-center">
    <div class="col-lg-8">
      <div class="d-flex flex-wrap justify-content-between align-items-end mb-3 gap-2">
        <div>
          <h2 class="fw-bold mb-0"><i class="bi bi-journal-text"></i> Cheat sheet</h2>
          <small class="text-muted">{{ topicName }} · your own notes. Write items, then let AI condense them for review.</small>
        </div>
        <button class="btn btn-outline-primary" @click="startNew">
          <i class="bi bi-plus-lg"></i> New note
        </button>
      </div>

      <div v-if="msg" class="alert py-2" :class="msgOk ? 'alert-success' : 'alert-danger'">{{ msg }}</div>

      <!-- editor (new or edit) -->
      <div v-if="editing" class="card border-0 shadow-sm mb-3 border-start border-primary border-4">
        <div class="card-body">
          <input v-model="form.title" class="form-control mb-2" maxlength="200" placeholder="Title (e.g. “nmap basics”)" />
          <textarea v-model="form.body" class="form-control font-monospace" rows="8"
            placeholder="Markdown supported: # headings, - lists, **bold**, `code`, ```code blocks```"></textarea>
          <div class="d-flex gap-2 mt-2">
            <button class="btn btn-primary btn-sm" :disabled="busy || (!form.title.trim() && !form.body.trim())" @click="save">
              <i class="bi bi-check-lg"></i> {{ form.id ? 'Save' : 'Add' }}
            </button>
            <button class="btn btn-link btn-sm" @click="cancel">Cancel</button>
            <span class="ms-auto small text-muted align-self-center">Preview below as you type</span>
          </div>
          <div v-if="form.body.trim()" class="md mt-2 p-2 rounded bg-light" v-html="render(form.body)"></div>
        </div>
      </div>

      <div v-if="loading" class="text-center py-5"><div class="spinner-border text-primary"></div></div>

      <!-- notes -->
      <div v-else>
        <div v-if="!notes.length && !editing" class="text-center text-muted py-4">
          <i class="bi bi-journal-plus display-5"></i>
          <p class="mt-2">No notes yet. Add cheat-sheet items for this course.</p>
        </div>

        <div v-for="(n, idx) in notes" :key="n.id" class="card border-0 shadow-sm mb-2">
          <div class="card-body py-2">
            <div class="d-flex align-items-start gap-2">
              <div class="flex-grow-1 min-w-0">
                <h6 v-if="n.title" class="fw-bold mb-1">{{ n.title }}</h6>
                <div class="md" v-html="render(n.body)"></div>
              </div>
              <div class="d-flex flex-column gap-1 flex-shrink-0">
                <button class="btn btn-sm btn-link p-0 text-muted" title="Move up" :disabled="idx === 0 || busy" @click="move(idx, -1)"><i class="bi bi-arrow-up"></i></button>
                <button class="btn btn-sm btn-link p-0 text-muted" title="Move down" :disabled="idx === notes.length - 1 || busy" @click="move(idx, 1)"><i class="bi bi-arrow-down"></i></button>
                <button class="btn btn-sm btn-link p-0" title="Edit" @click="startEdit(n)"><i class="bi bi-pencil"></i></button>
                <button class="btn btn-sm btn-link p-0 text-danger" title="Delete" @click="remove(n)"><i class="bi bi-trash"></i></button>
              </div>
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
import { api, errorMessage } from '../api'
import { loadSubjects, shortName, subject, subjectId } from '../composables/useSubject'
import { renderMarkdown } from '../utils/markdown'
import CourseChat from '../components/CourseChat.vue'

interface Note { id: number; subject_id: number; title: string; body: string; sort_order: number }
interface Summary { summary_md: string; generated_at: string }

const notes = ref<Note[]>([])
const summary = ref<Summary | null>(null)
const loading = ref(true)
const busy = ref(false)
const summarizing = ref(false)
const msg = ref('')
const msgOk = ref(false)

const editing = ref(false)
const form = reactive({ id: 0, title: '', body: '' })

const topicName = computed(() => (subject.value ? shortName(subject.value.slug, subject.value.name) : ''))
const render = (s: string) => renderMarkdown(s)

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

function startNew() {
  Object.assign(form, { id: 0, title: '', body: '' })
  editing.value = true
}
function startEdit(n: Note) {
  Object.assign(form, { id: n.id, title: n.title, body: n.body })
  editing.value = true
}
function cancel() {
  editing.value = false
}

async function save() {
  busy.value = true
  try {
    if (form.id) {
      await api.patch(`/course-notes/${form.id}`, { title: form.title, body: form.body })
    } else {
      await api.post('/course-notes', { subject_id: subjectId.value, title: form.title, body: form.body })
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
async function move(idx: number, dir: -1 | 1) {
  const j = idx + dir
  if (j < 0 || j >= notes.value.length) return
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
</style>
