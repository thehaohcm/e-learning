<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-end mb-2 gap-2">
      <div>
        <h2 class="fw-bold mb-0"><i class="bi bi-journal-code"></i> Cheatsheet</h2>
        <small class="text-muted">{{ topicName }} · mỗi mục: <b>muốn làm gì → lệnh</b>. Sửa được, gom theo nhóm, và học được bằng Study.</small>
      </div>
      <div class="d-flex gap-2">
        <button v-if="entries.length" class="btn btn-warning" :disabled="busy" @click="study">
          <span v-if="busy" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-arrow-repeat"></i> Study
        </button>
        <button v-if="canEdit" class="btn btn-outline-primary" @click="startNew()">
          <i class="bi bi-plus-lg"></i> New entry
        </button>
      </div>
    </div>

    <div v-if="msg" class="alert py-2" :class="msgOk ? 'alert-success' : 'alert-danger'">{{ msg }}</div>

    <!-- editor -->
    <div v-if="editing" class="card border-0 shadow-sm mb-3 border-start border-primary border-4">
      <div class="card-body">
        <div class="row g-2">
          <div class="col-12">
            <label class="form-label small mb-0">Muốn làm gì / khi nào dùng</label>
            <input v-model="form.front" class="form-control" maxlength="300" placeholder="VD: Liệt kê user hợp lệ trên Domain Controller" />
          </div>
          <div class="col-12">
            <label class="form-label small mb-0">Lệnh</label>
            <textarea v-model="form.back" class="form-control font-monospace" rows="2" placeholder="VD: kerbrute userenum --dc 10.10.3.2 -d DOMAIN.local users.txt"></textarea>
          </div>
          <div class="col-md-8">
            <label class="form-label small mb-0">Ghi chú (cờ, biến thể, lưu ý) — tùy chọn</label>
            <input v-model="form.note" class="form-control" maxlength="500" placeholder="VD: -o để lưu file; cần wordlist user" />
          </div>
          <div class="col-md-4">
            <label class="form-label small mb-0">Nhóm chủ đề</label>
            <input v-model="form.group" class="form-control" maxlength="80" list="cs-groups" placeholder="VD: Active Directory" />
            <datalist id="cs-groups"><option v-for="g in groupNames" :key="g" :value="g" /></datalist>
          </div>
        </div>
        <div class="d-flex gap-2 mt-2">
          <button class="btn btn-primary btn-sm" :disabled="busy || !form.front.trim() || !form.back.trim()" @click="save">
            <i class="bi bi-check-lg"></i> {{ form.id ? 'Save' : 'Add' }}
          </button>
          <button class="btn btn-link btn-sm" @click="editing = false">Cancel</button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="text-center py-5"><div class="spinner-border text-primary"></div></div>

    <template v-else>
      <p v-if="!entries.length && !editing" class="text-muted text-center py-4">
        <i class="bi bi-journal-plus display-5 d-block mb-2"></i>
        Chưa có lệnh nào. {{ canEdit ? 'Bấm “New entry” để thêm.' : '' }}
      </p>

      <!-- grouped entries -->
      <div v-for="g in groups" :key="g.name" class="mb-3">
        <h6 class="fw-bold text-uppercase text-muted small mt-3 mb-2">{{ g.name }} <span class="text-muted">({{ g.items.length }})</span></h6>
        <div class="card border-0 shadow-sm">
          <div class="list-group list-group-flush">
            <div v-for="c in g.items" :key="c.id" class="list-group-item">
              <div class="d-flex align-items-start gap-2">
                <div class="flex-grow-1 min-w-0">
                  <div class="fw-semibold">{{ c.front }}</div>
                  <div class="cmd-row d-flex align-items-start gap-1 mt-1">
                    <pre class="cmd mb-0 flex-grow-1">{{ c.payload.back }}</pre>
                    <button class="btn btn-sm btn-outline-secondary flex-shrink-0 copy-btn" :title="copiedId === c.id ? 'Đã chép' : 'Chép lệnh'" @click="copy(c)">
                      <i :class="copiedId === c.id ? 'bi bi-check2' : 'bi bi-clipboard'"></i>
                    </button>
                  </div>
                  <div v-if="c.payload.note_vi" class="small text-muted mt-1">{{ c.payload.note_vi }}</div>
                </div>
                <div v-if="canEdit" class="d-flex flex-column gap-1 flex-shrink-0">
                  <button class="btn btn-sm btn-link p-0" title="Sửa" @click="startEdit(c)"><i class="bi bi-pencil"></i></button>
                  <button class="btn btn-sm btn-link p-0 text-danger" title="Xóa" @click="remove(c)"><i class="bi bi-trash"></i></button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, errorMessage } from '../api'
import { auth } from '../auth'
import { currentPrefix } from '../router'
import { loadSubjects, shortName, subject, subjectId } from '../composables/useSubject'
import type { CardWithProgress } from '../types'

// The command cheatsheet for a course: one editable entry per command
// (front = goal/when, payload.back = the command), grouped by payload.group.
// The same cards are studied with the normal spaced-repetition engine.

const router = useRouter()
const canEdit = computed(() => auth.isAdmin)
const topicName = computed(() => (subject.value ? shortName(subject.value.slug, subject.value.name) : ''))

const entries = ref<CardWithProgress[]>([])
const loading = ref(true)
const busy = ref(false)
const copiedId = ref(0)
const msg = ref('')
const msgOk = ref(false)

const editing = ref(false)
const form = reactive({ id: 0, front: '', back: '', note: '', group: '' })

/** The cheatsheet deck of the current course (created by migration 0052). */
const deck = computed(() => subject.value?.decks?.find((d) => d.slug === 'cheatsheet'))

const groupNames = computed(() => {
  const seen: string[] = []
  for (const c of entries.value) {
    const g = (c.payload.group ?? '').trim()
    if (g && !seen.includes(g)) seen.push(g)
  }
  return seen
})

const groups = computed(() => {
  const map = new Map<string, CardWithProgress[]>()
  for (const c of entries.value) {
    const g = (c.payload.group ?? '').trim() || 'Khác'
    if (!map.has(g)) map.set(g, [])
    map.get(g)!.push(c)
  }
  // "Khác" (ungrouped) always last; others in first-seen order
  const names = [...map.keys()].filter((n) => n !== 'Khác')
  if (map.has('Khác')) names.push('Khác')
  return names.map((name) => ({ name, items: map.get(name)! }))
})

function flash(m: string, ok = true) {
  msg.value = m
  msgOk.value = ok
  setTimeout(() => { if (msg.value === m) msg.value = '' }, 4000)
}

async function load() {
  loading.value = true
  try {
    await loadSubjects()
    if (!deck.value) { entries.value = []; return }
    const { data } = await api.get('/cards', { params: { deck_id: deck.value.id, kind: 'concept', limit: 500 } })
    entries.value = data.data ?? []
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    loading.value = false
  }
}

function startNew() {
  Object.assign(form, { id: 0, front: '', back: '', note: '', group: groupNames.value[0] ?? '' })
  editing.value = true
}
function startEdit(c: CardWithProgress) {
  Object.assign(form, { id: c.id, front: c.front, back: c.payload.back ?? '', note: c.payload.note_vi ?? '', group: c.payload.group ?? '' })
  editing.value = true
}

async function save() {
  if (!deck.value || !form.front.trim() || !form.back.trim()) return
  busy.value = true
  try {
    const payload = { back: form.back.trim(), note_vi: form.note.trim(), group: form.group.trim() }
    if (form.id) {
      await api.put(`/cards/${form.id}`, { front: form.front.trim(), kind: 'concept', payload })
    } else {
      await api.post('/cards', { deck_id: deck.value.id, kind: 'concept', front: form.front.trim(), payload })
    }
    editing.value = false
    await load()
    flash('Đã lưu.')
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    busy.value = false
  }
}

async function remove(c: CardWithProgress) {
  if (!confirm(`Xóa mục “${c.front}”?`)) return
  busy.value = true
  try {
    await api.delete(`/cards/${c.id}`)
    entries.value = entries.value.filter((x) => x.id !== c.id)
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    busy.value = false
  }
}

async function copy(c: CardWithProgress) {
  try {
    await navigator.clipboard.writeText(c.payload.back ?? '')
    copiedId.value = c.id
    setTimeout(() => { if (copiedId.value === c.id) copiedId.value = 0 }, 1500)
  } catch {
    flash('Không chép được.', false)
  }
}

/** Enroll every entry (new ones only, server-side) then open Study for this deck. */
async function study() {
  if (!deck.value) return
  busy.value = true
  try {
    const ids = entries.value.filter((c) => !c.progress).map((c) => c.id)
    if (ids.length) await api.post('/study/enroll', { card_ids: ids })
    router.push(`/${currentPrefix()}/review?deck_id=${deck.value.id}`)
  } catch (e) {
    flash(errorMessage(e), false)
    busy.value = false
  }
}

watch(subjectId, (id, old) => { if (id && id !== old) { editing.value = false; load() } })
onMounted(load)
</script>

<style scoped>
.min-w-0 { min-width: 0; }
.cmd {
  white-space: pre-wrap;
  word-break: break-word;
  font-family: var(--bs-font-monospace);
  background: #0f172a;
  color: #e2e8f0;
  padding: 0.5rem 0.7rem;
  border-radius: 8px;
  font-size: 0.85rem;
  margin: 0;
}
.copy-btn { align-self: stretch; }
</style>
