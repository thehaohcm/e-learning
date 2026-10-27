<template>
  <div class="row justify-content-center">
    <div class="col-lg-9">
      <div class="d-flex flex-wrap justify-content-between align-items-end mb-3 gap-2">
        <div>
          <h2 class="fw-bold mb-0"><i class="bi bi-mortarboard"></i> Courses</h2>
          <small class="text-muted">Create and manage your study subjects, their decks and cards (admin).</small>
        </div>
        <button class="btn btn-primary" @click="showNew = !showNew">
          <i class="bi bi-plus-lg"></i> New course
        </button>
      </div>

      <div v-if="msg" class="alert py-2" :class="msgOk ? 'alert-success' : 'alert-danger'">{{ msg }}</div>

      <!-- new course -->
      <div v-if="showNew" class="card border-0 shadow-sm mb-3 border-start border-primary border-4">
        <div class="card-body">
          <h6 class="fw-bold mb-2">New course</h6>
          <div class="row g-2">
            <div class="col-md-5">
              <input v-model="nc.name" class="form-control" placeholder="Name, e.g. “German A1”" @input="nc.slug = slugify(nc.name)" />
            </div>
            <div class="col-md-4">
              <div class="input-group">
                <span class="input-group-text">slug</span>
                <input v-model="nc.slug" class="form-control" placeholder="german-a1" />
              </div>
            </div>
            <div class="col-md-3">
              <div class="input-group">
                <span class="input-group-text">Order</span>
                <input v-model.number="nc.sort_order" type="number" class="form-control" />
              </div>
            </div>
            <div class="col-12">
              <input v-model="nc.description" class="form-control" placeholder="Short description (optional)" />
            </div>
          </div>
          <button class="btn btn-sm btn-primary mt-2" :disabled="busy || !nc.name.trim() || !nc.slug.trim()" @click="createCourse">
            Create course
          </button>
          <div class="form-text">The slug is the permanent id (lowercase, no spaces). The name and order can be changed later.</div>
        </div>
      </div>

      <div v-if="loading" class="text-center py-5"><div class="spinner-border text-primary"></div></div>

      <!-- each course -->
      <div v-for="s in subjects" :key="s.id" class="card border-0 shadow-sm mb-3">
        <div class="card-body">
          <div class="d-flex flex-wrap justify-content-between align-items-start gap-2">
            <div class="flex-grow-1 min-w-0">
              <div v-if="edit.id === s.id" class="row g-2">
                <div class="col-md-6"><input v-model="edit.name" class="form-control form-control-sm" /></div>
                <div class="col-md-2">
                  <div class="input-group input-group-sm"><span class="input-group-text">Order</span>
                    <input v-model.number="edit.sort_order" type="number" class="form-control" /></div>
                </div>
                <div class="col-12"><input v-model="edit.description" class="form-control form-control-sm" placeholder="Description" /></div>
                <div class="col-12">
                  <button class="btn btn-sm btn-success" :disabled="busy" @click="saveCourse(s)">Save</button>
                  <button class="btn btn-sm btn-link" @click="edit.id = 0">Cancel</button>
                </div>
              </div>
              <template v-else>
                <h5 class="fw-bold mb-0">{{ s.name }} <small class="text-muted fw-normal">/{{ s.slug }}</small></h5>
                <div class="small text-muted">{{ s.description || '—' }} · order {{ s.sort_order }} · {{ s.decks.length }} deck(s)</div>
              </template>
            </div>
            <button v-if="edit.id !== s.id" class="btn btn-sm btn-outline-secondary" @click="startEdit(s)"><i class="bi bi-pencil"></i></button>
          </div>

          <!-- decks: a flex list that wraps on narrow screens (no fixed-width table) -->
          <div class="decks mt-3 mb-2">
            <div v-for="d in s.decks" :key="d.id" class="deck-row d-flex flex-wrap align-items-center gap-2 py-2 border-top">
              <template v-if="dedit.id === d.id">
                <div class="flex-grow-1 min-w-0 d-flex flex-column gap-1">
                  <input v-model="dedit.name" class="form-control form-control-sm" placeholder="Deck name" />
                  <input v-model="dedit.description" class="form-control form-control-sm" placeholder="Description" />
                  <select v-model="dedit.default_kind" class="form-select form-select-sm w-auto">
                    <option v-for="k in KINDS" :key="k" :value="k">{{ k }}</option>
                  </select>
                </div>
                <div class="d-flex gap-1 flex-shrink-0">
                  <button class="btn btn-sm btn-success" :disabled="busy" @click="saveDeck(s, d)">Save</button>
                  <button class="btn btn-sm btn-link" @click="dedit.id = 0">Cancel</button>
                </div>
              </template>
              <template v-else>
                <div class="flex-grow-1 min-w-0">
                  <div><span class="fw-semibold">{{ d.name }}</span> <small class="text-muted">/{{ d.slug }}</small></div>
                  <div v-if="d.description" class="small text-muted">{{ d.description }}</div>
                </div>
                <span class="badge text-bg-light border flex-shrink-0">{{ d.default_kind }}</span>
                <span class="small text-muted text-nowrap flex-shrink-0">{{ d.card_count }} cards</span>
                <div class="d-flex align-items-center gap-1 flex-shrink-0 ms-auto">
                  <div class="input-group input-group-sm w-auto">
                    <input v-model.number="genCount[d.id]" type="number" min="1" max="20" class="form-control" style="width: 3.5rem" />
                    <button class="btn btn-outline-primary" :disabled="genBusy === d.id" title="Generate cards with AI" @click="generate(s, d)">
                      <span v-if="genBusy === d.id" class="spinner-border spinner-border-sm"></span>
                      <template v-else><i class="bi bi-stars"></i> AI</template>
                    </button>
                  </div>
                  <router-link class="btn btn-sm btn-outline-secondary" :to="`/learn?deck_id=${d.id}`" title="Browse / add cards"><i class="bi bi-collection"></i></router-link>
                  <button class="btn btn-sm btn-outline-secondary" title="Edit deck" @click="startDeckEdit(d)"><i class="bi bi-pencil"></i></button>
                </div>
              </template>
            </div>
            <div v-if="!s.decks.length" class="text-muted small py-2 border-top">No decks yet — add one below.</div>
          </div>

          <!-- AI: fill the whole course -->
          <div v-if="auth.isAdmin && s.decks.length" class="border-top pt-2 mb-2">
            <div class="d-flex flex-wrap align-items-center gap-2">
              <button class="btn btn-sm btn-warning" :disabled="fill.running" @click="fillCourse(s)">
                <i class="bi bi-stars"></i> AI: gen toàn bộ course
              </button>
              <template v-if="!fill.running">
                <select v-if="fillModels.length" v-model="fill.model" class="form-select form-select-sm w-auto">
                  <option value="">Model mặc định</option>
                  <option v-for="m in fillModels" :key="m" :value="m">{{ m }}</option>
                </select>
                <div class="input-group input-group-sm w-auto">
                  <span class="input-group-text">Trần/deck</span>
                  <input v-model.number="fill.ceiling" type="number" min="10" max="400" class="form-control" style="width: 5rem" />
                </div>
              </template>
              <button v-if="fill.running && fill.subjectId === s.id" class="btn btn-sm btn-outline-danger" @click="fill.stop = true">
                <i class="bi bi-stop-circle"></i> Dừng
              </button>
            </div>
            <p class="small text-muted mb-0 mt-1">
              AI liệt kê toàn bộ khái niệm của từng deck rồi tự tạo đến khi phủ hết (hoặc đạt trần). Chạy lại chỉ tạo phần còn thiếu. Nên xem lại ở Browse và sửa thẻ sai.
            </p>
            <div v-if="fill.subjectId === s.id && (fill.running || fill.log.length)" class="mt-2 small">
              <div v-if="fill.running" class="d-flex align-items-center gap-2 text-primary">
                <span class="spinner-border spinner-border-sm"></span>
                <span>{{ fill.status }}</span>
                <span class="text-muted">· đã tạo {{ fill.created }} thẻ</span>
              </div>
              <div v-else class="fw-semibold">{{ fill.status }} · {{ fill.created }} thẻ mới</div>
              <ul v-if="fill.log.length" class="mb-0 mt-1 ps-3">
                <li v-for="(l, i) in fill.log" :key="i" class="text-muted">{{ l }}</li>
              </ul>
            </div>
          </div>

          <!-- new deck -->
          <div v-if="newDeckFor === s.id" class="row g-2 border-top pt-2">
            <div class="col-md-5"><input v-model="nd.name" class="form-control form-control-sm" placeholder="Deck name" @input="nd.slug = slugify(nd.name)" /></div>
            <div class="col-md-3"><input v-model="nd.slug" class="form-control form-control-sm" placeholder="slug" /></div>
            <div class="col-md-2">
              <select v-model="nd.default_kind" class="form-select form-select-sm">
                <option v-for="k in KINDS" :key="k" :value="k">{{ k }}</option>
              </select>
            </div>
            <div class="col-md-2">
              <button class="btn btn-sm btn-primary w-100" :disabled="busy || !nd.name.trim()" @click="createDeck(s)">Add</button>
            </div>
            <div class="col-12"><input v-model="nd.description" class="form-control form-control-sm" placeholder="Description (helps the AI generate on-topic cards)" /></div>
          </div>
          <button v-else class="btn btn-sm btn-outline-primary" @click="openNewDeck(s)"><i class="bi bi-plus-lg"></i> Add deck</button>
        </div>
      </div>

      <p class="small text-muted">
        <i class="bi bi-info-circle"></i> Card kinds: <b>vocab</b>/<b>phrasal</b> (English words), <b>concept</b> (term → definition, for PMP/Pentest/notes),
        <b>mcq</b> (multiple-choice question). “AI” appends generated cards; use Browse to add or edit cards by hand. Deleting a course/deck isn't available here yet.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, errorMessage } from '../api'
import { auth } from '../auth'
import { loadSubjects } from '../composables/useSubject'
import type { CardKind, Subject } from '../types'

const KINDS: CardKind[] = ['vocab', 'phrasal', 'concept', 'mcq']

const subjects = ref<Subject[]>([])
const loading = ref(true)
const busy = ref(false)
const genBusy = ref(0)
const msg = ref('')
const msgOk = ref(false)

const showNew = ref(false)
const nc = reactive({ name: '', slug: '', description: '', sort_order: 0 })
const newDeckFor = ref(0)
const nd = reactive({ name: '', slug: '', description: '', default_kind: 'concept' as CardKind })
const genCount = reactive<Record<number, number>>({})

const edit = reactive({ id: 0, name: '', description: '', sort_order: 0 })
const dedit = reactive({ id: 0, name: '', description: '', default_kind: 'concept' as CardKind })

// "Fill whole course": AI lays out the deck's full scope, then generates until
// covered. One run at a time, across every deck of the chosen course.
const fillModels = ref<string[]>([])
const fill = reactive({
  subjectId: 0, running: false, stop: false,
  model: '', ceiling: 150, deck: '', created: 0, status: '', log: [] as string[],
})

function slugify(s: string): string {
  return s.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '').replace(/đ/g, 'd')
    .replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '').slice(0, 64)
}

function flash(m: string, ok = true) {
  msg.value = m
  msgOk.value = ok
  setTimeout(() => { if (msg.value === m) msg.value = '' }, 4000)
}

async function load() {
  loading.value = true
  try {
    subjects.value = (await api.get('/subjects')).data ?? []
    for (const s of subjects.value) for (const d of s.decks) if (!(d.id in genCount)) genCount[d.id] = 10
    if (!fillModels.value.length) fillModels.value = (await api.get('/chat-models').catch(() => ({ data: {} }))).data.models ?? []
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    loading.value = false
  }
}

async function createCourse() {
  busy.value = true
  try {
    await api.post('/subjects', { name: nc.name.trim(), slug: slugify(nc.slug || nc.name), description: nc.description, sort_order: nc.sort_order })
    Object.assign(nc, { name: '', slug: '', description: '', sort_order: 0 })
    showNew.value = false
    await refresh()
    flash('Course created.')
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    busy.value = false
  }
}

function startEdit(s: Subject) {
  Object.assign(edit, { id: s.id, name: s.name, description: s.description, sort_order: s.sort_order })
}
async function saveCourse(s: Subject) {
  busy.value = true
  try {
    await api.put(`/subjects/${s.id}`, { name: edit.name.trim(), description: edit.description, sort_order: edit.sort_order })
    edit.id = 0
    await refresh()
    flash('Course saved.')
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    busy.value = false
  }
}

function openNewDeck(s: Subject) {
  newDeckFor.value = s.id
  Object.assign(nd, { name: '', slug: '', description: '', default_kind: 'concept' })
}
async function createDeck(s: Subject) {
  busy.value = true
  try {
    await api.post('/decks', { subject_id: s.id, name: nd.name.trim(), slug: slugify(nd.slug || nd.name), description: nd.description, default_kind: nd.default_kind })
    newDeckFor.value = 0
    await refresh()
    flash('Deck added.')
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    busy.value = false
  }
}

function startDeckEdit(d: Subject['decks'][number]) {
  Object.assign(dedit, { id: d.id, name: d.name, description: d.description, default_kind: d.default_kind })
}
async function saveDeck(_s: Subject, d: Subject['decks'][number]) {
  busy.value = true
  try {
    await api.put(`/decks/${d.id}`, { name: dedit.name.trim(), description: dedit.description, default_kind: dedit.default_kind })
    dedit.id = 0
    await refresh()
    flash('Deck saved.')
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    busy.value = false
  }
}

async function generate(_s: Subject, d: Subject['decks'][number]) {
  genBusy.value = d.id
  try {
    const { data } = await api.post(`/decks/${d.id}/generate`, { count: genCount[d.id] || 10 }, { timeout: 180000 })
    await refresh()
    flash(`${d.name}: added ${data.created} card(s)${data.skipped ? `, skipped ${data.skipped} duplicate/invalid` : ''}. Review them in Browse.`)
  } catch (e) {
    flash(errorMessage(e), false)
  } finally {
    genBusy.value = 0
  }
}

/**
 * Fill an entire course with AI, deck by deck: ask the AI for the deck's full
 * scope (a finite list of concepts), then generate cards in small batches until
 * the scope is covered or the per-deck ceiling is reached. Resumable: it only
 * generates what's missing, and stops a deck when a round adds nothing new.
 */
async function fillCourse(s: Subject) {
  if (fill.running) return
  Object.assign(fill, { subjectId: s.id, running: true, stop: false, created: 0, deck: '', status: 'Bắt đầu…', log: [] as string[] })
  const model = fill.model || undefined
  const ceiling = Math.max(10, Math.min(400, fill.ceiling || 150))
  try {
    for (const d of s.decks) {
      if (fill.stop) break
      fill.deck = d.name
      let rounds = 0
      while (!fill.stop && rounds < 40) {
        rounds++
        fill.status = `${d.name}: đang lập danh mục…`
        let data: { terms?: string[]; existing?: number; ceiling?: number; capped?: boolean; scope_complete?: boolean }
        try {
          data = (await api.post(`/decks/${d.id}/outline`, { ceiling, model }, { timeout: 180000 })).data
        } catch (e) {
          fill.log.push(`⚠️ ${d.name}: lập danh mục lỗi — ${errorMessage(e)}`)
          break
        }
        if (data.scope_complete) { fill.log.push(`✓ ${d.name}: đã phủ hết scope (${data.existing} thẻ)`); break }
        if (data.capped) { fill.log.push(`• ${d.name}: đạt trần ${data.ceiling} thẻ — tăng trần nếu muốn thêm`); break }
        const terms = data.terms ?? []
        if (!terms.length) { fill.log.push(`✓ ${d.name}: không còn mục mới`); break }

        const before = fill.created
        for (let i = 0; i < terms.length && !fill.stop; i += 12) {
          const batch = terms.slice(i, i + 12)
          fill.status = `${d.name}: đang tạo thẻ (${Math.min(i + batch.length, terms.length)}/${terms.length} mục vòng này)…`
          try {
            const r = await api.post(`/decks/${d.id}/generate`, { terms: batch, model }, { timeout: 180000 })
            fill.created += r.data.created ?? 0
          } catch (e) {
            fill.log.push(`⚠️ ${d.name}: tạo lỗi — ${errorMessage(e)}`)
          }
        }
        if (fill.created === before) { fill.log.push(`• ${d.name}: không tạo thêm được (có thể trùng/không hợp lệ) — dừng deck này`); break }
      }
    }
    fill.status = fill.stop ? 'Đã dừng.' : 'Hoàn tất toàn bộ course.'
    await refresh()
  } finally {
    fill.running = false
  }
}

/** Reload here and refresh the topic switch used by the rest of the app. */
async function refresh() {
  await Promise.all([load(), loadSubjects(true)])
}

onMounted(load)
</script>

<style scoped>
.min-w-0 { min-width: 0; }
</style>
