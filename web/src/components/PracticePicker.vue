<template>
  <div class="card border-0 shadow-sm mb-3 practice-picker">
    <div class="card-body py-3">
      <div class="d-flex flex-wrap align-items-end gap-2">
        <div>
          <label class="form-label small mb-1 text-muted">Dạng</label>
          <div class="btn-group d-block" role="group">
            <input v-if="isCommandSubject" id="pp-lab" v-model="kind" type="radio" class="btn-check" value="lab" />
            <label v-if="isCommandSubject" class="btn btn-sm btn-outline-primary" for="pp-lab"><i class="bi bi-terminal"></i> Gõ lệnh</label>
            <input id="pp-mcq" v-model="kind" type="radio" class="btn-check" value="mcq" />
            <label class="btn btn-sm btn-outline-primary" for="pp-mcq"><i class="bi bi-ui-checks"></i> Trắc nghiệm</label>
          </div>
        </div>

        <div class="flex-grow-1" style="min-width: 220px" ref="topicMenuRoot">
          <label class="form-label small mb-1 text-muted">Mục (từ cheatsheet)</label>
          <div class="dropdown topic-menu">
            <button type="button" class="btn btn-sm btn-outline-secondary dropdown-toggle w-100 text-start text-truncate"
              @click="topicMenuOpen = !topicMenuOpen">
              {{ topicSummaryLabel }}
            </button>
            <ul class="dropdown-menu shadow-sm w-100" :class="{ show: topicMenuOpen }">
              <li>
                <button type="button" class="dropdown-item" :class="{ active: selectedTopics.length === 0 }" @click="selectedTopics = []">
                  Tất cả{{ totalForKind ? ` (${totalForKind})` : '' }}
                </button>
              </li>
              <li v-if="topicsForKind.length"><hr class="dropdown-divider" /></li>
              <li class="topic-list">
                <label v-for="t in topicsForKind" :key="t.topic" class="dropdown-item">
                  <input type="checkbox" class="form-check-input flex-shrink-0" :checked="selectedTopics.includes(t.topic)" @change="toggleTopic(t.topic)" />
                  <span class="text-truncate">{{ t.topic || '(không mục)' }} ({{ kind === 'lab' ? t.lab : t.mcq }})</span>
                </label>
              </li>
            </ul>
          </div>
        </div>

        <div>
          <button class="btn btn-sm btn-warning" :disabled="busy || totalForKind === 0" @click="drill">
            <span v-if="busy" class="spinner-border spinner-border-sm me-1"></span>
            <i v-else class="bi bi-play-fill"></i> Drill ngay
          </button>
        </div>
        <button v-if="canEdit && selectedTopics.length" class="btn btn-sm btn-outline-danger" :disabled="deleting || totalForKind === 0"
          title="Xóa toàn bộ câu hỏi AI đã tạo cho các mục đã chọn" @click="deleteTopic">
          <span v-if="deleting" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-trash3"></i> Xóa mục
        </button>
        <button class="btn btn-sm btn-link text-muted px-1" :disabled="loading" title="Tải lại danh sách mục" @click="loadTopics">
          <i class="bi bi-arrow-clockwise"></i>
        </button>
      </div>

      <div v-if="err" class="text-danger small mt-2">{{ err }}</div>
      <div v-else-if="totalForKind === 0 && !loading" class="text-muted small mt-2">
        Chưa có thẻ {{ kind === 'lab' ? 'gõ lệnh' : 'trắc nghiệm' }} nào. Bấm <b>Call AI</b> / <b>Tạo câu hỏi từ cheatsheet</b> để tạo trước.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { api, errorMessage } from '../api'
import { auth } from '../auth'
import { isCommandSubject, subjectId } from '../composables/useSubject'
import type { QueueItem } from '../types'

// "Study a chosen section + format now" picker (Pentest). Lists the cheatsheet
// sections that have AI-generated drills, lets the user pick a format
// (typed-command / multiple-choice) and a section, and loads those cards as a
// drill set to study immediately (no spaced-repetition schedule).
const emit = defineEmits<{ drill: [payload: { items: QueueItem[]; kind: 'lab' | 'mcq'; topic: string; label: string }] }>()

type TopicCount = { topic: string; lab: number; mcq: number }

// PMP has no shell commands, so it only offers multiple-choice drills
const kind = ref<'lab' | 'mcq'>(isCommandSubject.value ? 'lab' : 'mcq')
// multiple sections can be picked at once (checkboxes); empty = "Tất cả"
const selectedTopics = ref<string[]>([])
const topicMenuOpen = ref(false)
const topicMenuRoot = ref<HTMLElement | null>(null)
const topics = ref<TopicCount[]>([])
const totals = ref<{ lab: number; mcq: number }>({ lab: 0, mcq: 0 })
const loading = ref(false)
const busy = ref(false)
const deleting = ref(false)
const err = ref('')
const canEdit = computed(() => auth.isAdmin)

const totalForKind = computed(() => (kind.value === 'lab' ? totals.value.lab : totals.value.mcq))
const topicsForKind = computed(() => topics.value.filter((t) => (kind.value === 'lab' ? t.lab : t.mcq) > 0))

const topicSummaryLabel = computed(() => {
  if (selectedTopics.value.length === 0) return `Tất cả${totalForKind.value ? ` (${totalForKind.value})` : ''}`
  if (selectedTopics.value.length === 1) return selectedTopics.value[0] || '(không mục)'
  return `${selectedTopics.value.length} mục đã chọn`
})

function toggleTopic(t: string) {
  const i = selectedTopics.value.indexOf(t)
  if (i === -1) selectedTopics.value = [...selectedTopics.value, t]
  else selectedTopics.value = selectedTopics.value.filter((x) => x !== t)
}

function onDocClick(e: MouseEvent) {
  if (topicMenuOpen.value && topicMenuRoot.value && !topicMenuRoot.value.contains(e.target as Node)) topicMenuOpen.value = false
}
onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))

// switching Pentest <-> PMP in the menu: PMP has no typed-command drills
watch(isCommandSubject, (cmd) => { if (!cmd) kind.value = 'mcq' })

// if the chosen sections no longer exist in the new format, drop them
watch(kind, () => {
  selectedTopics.value = selectedTopics.value.filter((t) => topicsForKind.value.some((x) => x.topic === t))
})

async function loadTopics() {
  if (!subjectId.value) return
  loading.value = true
  err.value = ''
  try {
    const { data } = await api.get('/study/practice-topics', { params: { subject_id: subjectId.value } })
    topics.value = data.topics ?? []
    totals.value = { lab: data.lab ?? 0, mcq: data.mcq ?? 0 }
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

async function drill() {
  if (!subjectId.value) return
  busy.value = true
  err.value = ''
  try {
    // always fetch the full set (no server-side topic filter) and, when one or
    // more sections are checked, keep only cards from those sections here —
    // lets the picker combine several sections in one drill with one request.
    const { data } = await api.get('/study/practice', {
      params: { subject_id: subjectId.value, kind: kind.value, limit: 200 },
    })
    let items: QueueItem[] = data ?? []
    if (selectedTopics.value.length > 0) {
      const set = new Set(selectedTopics.value)
      items = items.filter((it) => set.has(it.card.payload.topic ?? ''))
    }
    const topicLabel = selectedTopics.value.map((t) => t || '(không mục)').join(', ')
    const label = `${kind.value === 'lab' ? 'Gõ lệnh' : 'Trắc nghiệm'}${topicLabel ? ' · ' + topicLabel : ''}`
    emit('drill', { items, kind: kind.value, topic: selectedTopics.value.join(','), label })
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

/** Delete every AI-generated card of the current kind + selected topic (the
 *  same set "Drill ngay" would study) — fetches the live list so counts can
 *  never drift, then removes them one by one via the existing card-delete
 *  endpoint (cascades progress/logs server-side). */
async function deleteTopic() {
  if (!subjectId.value || selectedTopics.value.length === 0) return
  const set = new Set(selectedTopics.value)
  const n = topicsForKind.value.filter((t) => set.has(t.topic)).reduce((sum, t) => sum + (kind.value === 'lab' ? t.lab : t.mcq), 0)
  const names = selectedTopics.value.map((t) => t || '(không mục)').join(', ')
  if (!confirm(`Xóa toàn bộ ${n || totalForKind.value} câu ${kind.value === 'lab' ? 'gõ lệnh' : 'trắc nghiệm'} đã tạo cho mục "${names}"? Không thể hoàn tác.`)) return
  deleting.value = true
  err.value = ''
  try {
    const { data } = await api.get('/study/practice', {
      params: { subject_id: subjectId.value, kind: kind.value, limit: 200 },
    })
    const items: QueueItem[] = (data ?? []).filter((it: QueueItem) => set.has(it.card.payload.topic ?? ''))
    await Promise.all(items.map((it) => api.delete(`/cards/${it.card.id}`)))
    selectedTopics.value = []
    await loadTopics()
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    deleting.value = false
  }
}

watch(subjectId, () => loadTopics())
onMounted(loadTopics)
defineExpose({ loadTopics })
</script>

<style scoped>
.practice-picker .form-label {
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.topic-menu .dropdown-menu {
  padding: 0.35rem;
  border-radius: 0.6rem;
  font-size: 0.85rem;
}
.topic-menu .dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.6rem;
  border-radius: 0.4rem;
  cursor: pointer;
}
.topic-menu .dropdown-item:hover,
.topic-menu .dropdown-item:focus-visible {
  background-color: var(--bs-tertiary-bg, #f1f3f5);
}
.topic-menu .dropdown-item.active {
  border-radius: 0.4rem;
}
.topic-menu .dropdown-divider {
  margin: 0.3rem 0.2rem;
}
.topic-menu .topic-list {
  display: block;
  max-height: 240px;
  overflow-y: auto;
}
.topic-menu .topic-list::-webkit-scrollbar {
  width: 6px;
}
.topic-menu .topic-list::-webkit-scrollbar-thumb {
  background-color: rgba(0, 0, 0, 0.18);
  border-radius: 3px;
}
.topic-menu .form-check-input {
  margin-top: 0;
}
</style>
