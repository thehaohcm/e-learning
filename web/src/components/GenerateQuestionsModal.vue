<template>
  <div v-if="show" class="gq-overlay" @click.self="close">
    <div class="card border-0 shadow-lg gq-dialog">
      <div class="card-header bg-white d-flex align-items-center gap-2">
        <i class="bi bi-stars text-warning"></i>
        <span class="fw-bold">Tạo câu hỏi bằng AI</span>
        <small class="text-muted">· {{ topicName() }}</small>
        <button type="button" class="btn-close ms-auto" aria-label="Close" @click="close"></button>
      </div>
      <div class="card-body">
        <p class="small text-muted">AI đọc <b>cheatsheet (ghi chú)</b> của bạn và tạo câu hỏi trắc nghiệm để ôn. Chỉ dựa trên ghi chú, không bịa.</p>

        <label class="form-label small mb-1">Dạng</label>
        <div class="btn-group w-100 mb-2" role="group">
          <input id="fmt-lab" v-model="format" type="radio" class="btn-check" value="lab" />
          <label class="btn btn-outline-primary" for="fmt-lab"><i class="bi bi-terminal"></i> Gõ lệnh (Labs)</label>
          <input id="fmt-mcq" v-model="format" type="radio" class="btn-check" value="mcq" />
          <label class="btn btn-outline-primary" for="fmt-mcq"><i class="bi bi-ui-checks"></i> Trắc nghiệm</label>
        </div>
        <p class="small text-muted mb-2">
          {{ format === 'lab' ? 'Trong Study bạn gõ lệnh, hệ thống so khớp (cờ không cần đúng thứ tự, vd ss -tunlp = ss -utpln).' : 'Chọn 1 trong 4 đáp án, ôn trong Study.' }}
        </p>

        <label class="form-label small mb-1">Chủ đề / mục cần ôn <span class="text-muted">(để trống = toàn bộ cheatsheet)</span></label>
        <input v-model="topic" class="form-control mb-2" maxlength="200" placeholder="VD: Active Directory, Recon, privilege escalation…" />

        <label class="form-label small mb-1">Model</label>
        <select v-model="model" class="form-select mb-2">
          <option value="">Mặc định</option>
          <option v-for="m in models" :key="m" :value="m">{{ m }}</option>
        </select>
        <p class="small text-muted mb-2"><i class="bi bi-info-circle"></i> AI tự chọn số câu phù hợp với lượng cheatsheet (tối đa 25). Tạo xong, vào Study chọn mục để drill lại nhiều lần cho nhớ.</p>

        <div v-if="err" class="alert alert-danger py-2 small mb-2">{{ err }}</div>
        <div v-if="okMsg" class="alert alert-success py-2 small mb-2">{{ okMsg }}</div>

        <div class="d-flex gap-2">
          <button class="btn btn-warning" :disabled="busy" @click="generate">
            <span v-if="busy" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-magic"></i>
            {{ busy ? 'Đang tạo… (có thể ~30–60s)' : 'Tạo & học' }}
          </button>
          <button class="btn btn-link" :disabled="busy" @click="close">Đóng</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { api, errorMessage } from '../api'
import { shortName, subject, subjectId } from '../composables/useSubject'

// Reusable "generate practice from the cheatsheet" dialog. Opened via the
// exposed open() method. "lab" = typed-command cards, "mcq" = multiple-choice;
// both are enrolled into the cheatsheet deck and studied (SRS). On success it
// emits 'done' so the parent can open Study.
const emit = defineEmits<{ done: [payload: { deckId: number; created: number }] }>()

const show = ref(false)
const format = ref<'lab' | 'mcq'>('lab')
const topic = ref('')
const model = ref('')
const models = ref<string[]>([])
const busy = ref(false)
const err = ref('')
const okMsg = ref('')

const topicName = () => (subject.value ? shortName(subject.value.slug, subject.value.name) : '')

async function loadModels() {
  if (models.value.length) return
  try { models.value = (await api.get('/chat-models')).data.models ?? [] } catch { models.value = [] }
}

function open() {
  err.value = ''
  okMsg.value = ''
  show.value = true
  loadModels()
}
function close() {
  if (busy.value) return
  show.value = false
}

async function generate() {
  if (!subjectId.value) return
  busy.value = true
  err.value = ''
  okMsg.value = ''
  try {
    const { data } = await api.post('/course-notes/gen-questions',
      { subject_id: subjectId.value, topic: topic.value.trim(), model: model.value || undefined, format: format.value },
      { timeout: 180000 })
    if (!data.created) {
      okMsg.value = 'AI không tạo được gì từ ghi chú hiện có. Thử chủ đề khác hoặc thêm ghi chú.'
      return
    }
    show.value = false
    // both formats are studied in the cheatsheet deck (SRS): enroll, then raise
    // today's new-card room so the just-made cards are available to study now
    const ids: number[] = data.card_ids ?? []
    if (ids.length) await api.post('/study/enroll', { card_ids: ids })
    try { await api.post('/study/extra-new', { count: data.created, subject_id: subjectId.value }) } catch { /* best effort */ }
    emit('done', { deckId: data.deck_id, created: data.created })
  } catch (e) {
    err.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.gq-overlay {
  position: fixed;
  inset: 0;
  z-index: 1060;
  background: rgba(15, 23, 42, 0.5);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 1.5rem 1rem;
  overflow-y: auto;
}
.gq-dialog { width: 100%; max-width: 520px; border-radius: 14px; }
</style>
