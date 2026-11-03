<template>
  <div class="modal-backdrop fade show"></div>
  <div class="modal fade show d-block" tabindex="-1" role="dialog" @mousedown.self="$emit('close')">
    <div class="modal-dialog modal-lg modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">
            <i class="bi bi-journal-plus"></i> Câu hỏi tự luyện — {{ challenge.name }}
            <span class="badge bg-secondary ms-1">{{ challenge.platform }}</span>
          </h5>
          <button type="button" class="btn-close" @click="$emit('close')"></button>
        </div>

        <div class="modal-body">
          <p class="small text-muted">
            Soạn câu hỏi/đáp án riêng cho phòng/box này để tự luyện lại trước khi làm thật (không tính vào điểm ôn tập/SR).
            Mỗi câu gồm: câu hỏi, (các) đáp án chấp nhận, giải thích và ghi chú riêng của bạn.
          </p>

          <div v-if="!steps.length" class="text-center text-muted small py-3">Chưa có câu hỏi nào. Bấm "Thêm câu hỏi" bên dưới.</div>

          <div v-for="(s, i) in steps" :key="s._key" class="card border-0 shadow-sm mb-2">
            <div class="card-body py-2">
              <div class="d-flex justify-content-between align-items-start mb-1">
                <span class="fw-semibold small text-muted">Câu {{ i + 1 }}</span>
                <div class="btn-group btn-group-sm">
                  <button type="button" class="btn btn-outline-secondary" :disabled="i === 0" @click="move(i, -1)" title="Lên"><i class="bi bi-arrow-up"></i></button>
                  <button type="button" class="btn btn-outline-secondary" :disabled="i === steps.length - 1" @click="move(i, 1)" title="Xuống"><i class="bi bi-arrow-down"></i></button>
                  <button type="button" class="btn btn-outline-danger" @click="removeStep(i)" title="Xoá câu này"><i class="bi bi-trash"></i></button>
                </div>
              </div>
              <div class="mb-2">
                <label class="form-label small mb-1">Câu hỏi</label>
                <textarea v-model="s.prompt" class="form-control form-control-sm" rows="2" placeholder="Vd: Port 445 mở dịch vụ gì và nên enum bằng công cụ nào?"></textarea>
              </div>
              <div class="mb-2">
                <label class="form-label small mb-1">Đáp án chấp nhận <span class="text-muted">(mỗi dòng 1 cách viết, regex không phân biệt hoa/thường)</span></label>
                <textarea v-model="s.patternsText" class="form-control form-control-sm font-monospace" rows="2" placeholder="smb&#10;enum4linux"></textarea>
              </div>
              <div class="mb-2">
                <label class="form-label small mb-1">Giải thích <span class="text-muted">(hiện sau khi trả lời đúng)</span></label>
                <textarea v-model="s.explanation" class="form-control form-control-sm" rows="2"></textarea>
              </div>
              <div>
                <label class="form-label small mb-1">Ghi chú / gợi ý riêng <span class="text-muted">(hiện khi bấm "Xem gợi ý")</span></label>
                <textarea v-model="s.hint" class="form-control form-control-sm" rows="2"></textarea>
              </div>
            </div>
          </div>

          <button type="button" class="btn btn-sm btn-outline-primary" @click="addStep">
            <i class="bi bi-plus-lg"></i> Thêm câu hỏi
          </button>

          <div v-if="error" class="alert alert-danger small mt-3 mb-0">{{ error }}</div>
        </div>

        <div class="modal-footer justify-content-between">
          <button v-if="card" type="button" class="btn btn-outline-danger btn-sm" :disabled="saving" @click="removeAll">
            <i class="bi bi-trash"></i> Xoá toàn bộ mô phỏng
          </button>
          <span v-else></span>
          <div>
            <button type="button" class="btn btn-outline-secondary me-2" :disabled="saving" @click="$emit('close')">Đóng</button>
            <button type="button" class="btn btn-primary" :disabled="saving" @click="save">
              <span v-if="saving" class="spinner-border spinner-border-sm me-1"></span> Lưu
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { api, errorMessage } from '../api'
import type { Card } from '../types'
import { challengeKey, type Challenge } from '../data/pentestChallenges'

const props = defineProps<{ challenge: Challenge; deckId: number | null; card: Card | null }>()
const emit = defineEmits<{ close: []; saved: [card: Card]; deleted: [cardId: number] }>()

interface EditStep { _key: number; prompt: string; patternsText: string; explanation: string; hint: string }

let keySeq = 0
function toEditStep(s: { prompt: string; patterns?: string[]; explanation?: string; hint?: string }): EditStep {
  return { _key: keySeq++, prompt: s.prompt, patternsText: (s.patterns ?? []).join('\n'), explanation: s.explanation ?? '', hint: s.hint ?? '' }
}

const steps = ref<EditStep[]>((props.card?.payload.steps ?? []).map(toEditStep))
const saving = ref(false)
const error = ref('')

function addStep() {
  steps.value.push({ _key: keySeq++, prompt: '', patternsText: '', explanation: '', hint: '' })
}
function removeStep(i: number) {
  steps.value.splice(i, 1)
}
function move(i: number, dir: 1 | -1) {
  const j = i + dir
  if (j < 0 || j >= steps.value.length) return
  const arr = steps.value
  ;[arr[i], arr[j]] = [arr[j], arr[i]]
}

async function save() {
  error.value = ''
  const payloadSteps = steps.value
    .map((s) => ({
      prompt: s.prompt.trim(),
      patterns: s.patternsText.split('\n').map((p) => p.trim()).filter(Boolean),
      explanation: s.explanation.trim(),
      hint: s.hint.trim(),
    }))
    .filter((s) => s.prompt && s.patterns.length)
  if (!payloadSteps.length) {
    error.value = 'Cần ít nhất 1 câu hỏi có đủ câu hỏi và đáp án.'
    return
  }
  saving.value = true
  try {
    const body = {
      deck_id: props.deckId,
      kind: 'lab',
      front: `[Simulation] ${props.challenge.platform} — ${props.challenge.name}`,
      payload: {
        mode: 'challenge_sim',
        scenario: `Tự luyện câu hỏi/đáp án riêng cho "${props.challenge.name}" (${props.challenge.platform}) trước khi làm thật — không tính vào ôn tập ngắt quãng.`,
        steps: payloadSteps,
        difficulty: props.challenge.difficulty,
        cert_tags: props.challenge.cert_tags,
        challenge_key: challengeKey(props.challenge),
      },
    }
    const { data } = props.card
      ? await api.put(`/cards/${props.card.id}`, body)
      : await api.post('/cards', body)
    emit('saved', data as Card)
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

async function removeAll() {
  if (!props.card) return
  if (!confirm('Xoá toàn bộ câu hỏi mô phỏng cho challenge này?')) return
  saving.value = true
  try {
    await api.delete(`/cards/${props.card.id}`)
    emit('deleted', props.card.id)
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}
</script>
