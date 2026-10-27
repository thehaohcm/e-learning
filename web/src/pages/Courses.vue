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

          <!-- decks -->
          <table class="table table-sm align-middle mt-3 mb-2">
            <thead><tr class="small text-muted"><th>Deck</th><th>Kind</th><th class="text-end">Cards</th><th></th></tr></thead>
            <tbody>
              <tr v-for="d in s.decks" :key="d.id">
                <td>
                  <template v-if="dedit.id === d.id">
                    <input v-model="dedit.name" class="form-control form-control-sm d-inline-block" style="width: 14rem" />
                    <input v-model="dedit.description" class="form-control form-control-sm mt-1" placeholder="Description" />
                  </template>
                  <template v-else>
                    <span class="fw-semibold">{{ d.name }}</span> <small class="text-muted">/{{ d.slug }}</small>
                    <div class="small text-muted">{{ d.description }}</div>
                  </template>
                </td>
                <td>
                  <select v-if="dedit.id === d.id" v-model="dedit.default_kind" class="form-select form-select-sm">
                    <option v-for="k in KINDS" :key="k" :value="k">{{ k }}</option>
                  </select>
                  <span v-else class="badge text-bg-light border">{{ d.default_kind }}</span>
                </td>
                <td class="text-end">{{ d.card_count }}</td>
                <td class="text-end text-nowrap">
                  <template v-if="dedit.id === d.id">
                    <button class="btn btn-sm btn-success py-0" :disabled="busy" @click="saveDeck(s, d)">Save</button>
                    <button class="btn btn-sm btn-link py-0" @click="dedit.id = 0">Cancel</button>
                  </template>
                  <template v-else>
                    <div class="input-group input-group-sm d-inline-flex w-auto align-middle me-1">
                      <input v-model.number="genCount[d.id]" type="number" min="1" max="20" class="form-control" style="width: 4rem" />
                      <button class="btn btn-outline-primary" :disabled="genBusy === d.id" title="Generate cards with AI" @click="generate(s, d)">
                        <span v-if="genBusy === d.id" class="spinner-border spinner-border-sm"></span>
                        <template v-else><i class="bi bi-stars"></i> AI</template>
                      </button>
                    </div>
                    <router-link class="btn btn-sm btn-outline-secondary py-0" :to="`/learn?deck_id=${d.id}`" title="Browse / add cards"><i class="bi bi-collection"></i></router-link>
                    <button class="btn btn-sm btn-outline-secondary py-0 ms-1" @click="startDeckEdit(d)"><i class="bi bi-pencil"></i></button>
                  </template>
                </td>
              </tr>
              <tr v-if="!s.decks.length"><td colspan="4" class="text-muted small">No decks yet — add one below.</td></tr>
            </tbody>
          </table>

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

/** Reload here and refresh the topic switch used by the rest of the app. */
async function refresh() {
  await Promise.all([load(), loadSubjects(true)])
}

onMounted(load)
</script>

<style scoped>
.min-w-0 { min-width: 0; }
</style>
