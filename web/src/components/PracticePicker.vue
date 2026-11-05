<template>
  <div class="card border-0 shadow-sm mb-3 practice-picker">
    <div class="card-body py-3">
      <div class="d-flex flex-wrap align-items-end gap-2">
        <div>
          <label class="form-label small mb-1 text-muted">Dạng</label>
          <div class="btn-group d-block" role="group">
            <input id="pp-lab" v-model="kind" type="radio" class="btn-check" value="lab" />
            <label class="btn btn-sm btn-outline-primary" for="pp-lab"><i class="bi bi-terminal"></i> Gõ lệnh</label>
            <input id="pp-mcq" v-model="kind" type="radio" class="btn-check" value="mcq" />
            <label class="btn btn-sm btn-outline-primary" for="pp-mcq"><i class="bi bi-ui-checks"></i> Trắc nghiệm</label>
          </div>
        </div>

        <div class="flex-grow-1" style="min-width: 180px">
          <label class="form-label small mb-1 text-muted">Mục (từ cheatsheet)</label>
          <select v-model="topic" class="form-select form-select-sm">
            <option value="">Tất cả{{ totalForKind ? ` (${totalForKind})` : '' }}</option>
            <option v-for="t in topicsForKind" :key="t.topic" :value="t.topic">
              {{ t.topic || '(không mục)' }} ({{ kind === 'lab' ? t.lab : t.mcq }})
            </option>
          </select>
        </div>

        <div>
          <button class="btn btn-sm btn-warning" :disabled="busy || totalForKind === 0" @click="drill">
            <span v-if="busy" class="spinner-border spinner-border-sm me-1"></span>
            <i v-else class="bi bi-play-fill"></i> Drill ngay
          </button>
        </div>
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
import { computed, onMounted, ref, watch } from 'vue'
import { api, errorMessage } from '../api'
import { subjectId } from '../composables/useSubject'
import type { QueueItem } from '../types'

// "Study a chosen section + format now" picker (Pentest). Lists the cheatsheet
// sections that have AI-generated drills, lets the user pick a format
// (typed-command / multiple-choice) and a section, and loads those cards as a
// drill set to study immediately (no spaced-repetition schedule).
const emit = defineEmits<{ drill: [payload: { items: QueueItem[]; kind: 'lab' | 'mcq'; topic: string; label: string }] }>()

type TopicCount = { topic: string; lab: number; mcq: number }

const kind = ref<'lab' | 'mcq'>('lab')
const topic = ref('')
const topics = ref<TopicCount[]>([])
const totals = ref<{ lab: number; mcq: number }>({ lab: 0, mcq: 0 })
const loading = ref(false)
const busy = ref(false)
const err = ref('')

const totalForKind = computed(() => (kind.value === 'lab' ? totals.value.lab : totals.value.mcq))
const topicsForKind = computed(() => topics.value.filter((t) => (kind.value === 'lab' ? t.lab : t.mcq) > 0))

// if the chosen section has nothing in the new format, fall back to "Tất cả"
watch(kind, () => {
  if (topic.value && !topicsForKind.value.some((t) => t.topic === topic.value)) topic.value = ''
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
    const { data } = await api.get('/study/practice', {
      params: { subject_id: subjectId.value, kind: kind.value, topic: topic.value || undefined },
    })
    const items: QueueItem[] = data ?? []
    const label = `${kind.value === 'lab' ? 'Gõ lệnh' : 'Trắc nghiệm'}${topic.value ? ' · ' + topic.value : ''}`
    emit('drill', { items, kind: kind.value, topic: topic.value, label })
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    busy.value = false
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
</style>
