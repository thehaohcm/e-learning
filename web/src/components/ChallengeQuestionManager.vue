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

          <!-- Sync câu hỏi từ một mục cheatsheet vào challenge này -->
          <div class="border rounded p-2 mb-3 bg-body-secondary">
            <div class="d-flex flex-wrap align-items-center gap-2">
              <i class="bi bi-arrow-repeat text-primary"></i>
              <span class="fw-semibold small">Sync từ cheatsheet</span>
              <select v-model="syncTopic" class="form-select form-select-sm" style="max-width: 260px">
                <option value="">— Chọn mục cheatsheet —</option>
                <option v-for="t in topics" :key="t" :value="t">{{ t }}</option>
              </select>
              <div class="btn-group btn-group-sm" role="group" aria-label="Ngôn ngữ">
                <input id="sync-vi" v-model="syncLang" type="radio" class="btn-check" value="vi" />
                <label class="btn btn-outline-secondary" for="sync-vi">VI</label>
                <input id="sync-en" v-model="syncLang" type="radio" class="btn-check" value="en" />
                <label class="btn btn-outline-secondary" for="sync-en">EN</label>
              </div>
              <button type="button" class="btn btn-sm btn-primary" :disabled="syncBusy || !syncTopic" @click="syncFromCheatsheet">
                <span v-if="syncBusy" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-magic"></i>
                Lấy câu hỏi
              </button>
              <button type="button" class="btn btn-sm btn-outline-primary" :disabled="flowBusy || !syncTopic" @click="genThinkingFlow" title="AI sinh luồng tư duy tình huống (trắc nghiệm + lệnh)">
                <span v-if="flowBusy" class="spinner-border spinner-border-sm me-1"></span><i v-else class="bi bi-diagram-3"></i>
                Luồng tư duy (AI)
              </button>
            </div>
            <div class="form-text"><b>Lấy câu hỏi</b>: trích mọi lệnh trong mục (bỏ qua trùng). <b>Luồng tư duy (AI)</b>: sinh chuỗi tình huống "chọn bước đúng" + lệnh chủ chốt để luyện methodology. Xem lại rồi bấm Lưu.</div>
            <div v-if="syncMsg" class="small mt-1" :class="syncErr ? 'text-danger' : 'text-success'">{{ syncMsg }}</div>
          </div>

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
              <div class="mb-2 d-flex align-items-center gap-2">
                <label class="form-label small mb-0">Loại:</label>
                <select v-model="s.type" class="form-select form-select-sm" style="max-width: 170px">
                  <option value="command">Gõ lệnh</option>
                  <option value="choice">Trắc nghiệm (tư duy)</option>
                </select>
              </div>
              <div class="mb-2">
                <label class="form-label small mb-1">Câu hỏi / tình huống</label>
                <textarea v-model="s.prompt" class="form-control form-control-sm" rows="2" placeholder="Vd: Vừa có shell Windows, nên kiểm tra gì đầu tiên?"></textarea>
              </div>
              <div v-if="s.type !== 'choice'" class="mb-2">
                <label class="form-label small mb-1">Đáp án chấp nhận <span class="text-muted">(mỗi dòng 1 cách viết, regex không phân biệt hoa/thường)</span></label>
                <textarea v-model="s.patternsText" class="form-control form-control-sm font-monospace" rows="2" placeholder="smb&#10;enum4linux"></textarea>
              </div>
              <template v-else>
                <div class="mb-2">
                  <label class="form-label small mb-1">Các lựa chọn <span class="text-muted">(mỗi dòng 1 lựa chọn)</span></label>
                  <textarea v-model="s.choicesText" class="form-control form-control-sm" rows="4" placeholder="whoami /priv&#10;Tải winPEAS lên và chạy&#10;Reboot máy&#10;..."></textarea>
                </div>
                <div class="mb-2">
                  <label class="form-label small mb-1">Đáp án đúng</label>
                  <select v-model.number="s.answer" class="form-select form-select-sm">
                    <option v-for="(ln, li) in linesOf(s.choicesText)" :key="li" :value="li">{{ li + 1 }}. {{ ln }}</option>
                  </select>
                </div>
              </template>
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
import { onMounted, ref } from 'vue'
import { api, errorMessage } from '../api'
import type { Card } from '../types'
import { challengeKey, type Challenge } from '../data/pentestChallenges'
import { subjectId } from '../composables/useSubject'

const props = defineProps<{ challenge: Challenge; deckId: number | null; card: Card | null }>()
const emit = defineEmits<{ close: []; saved: [card: Card]; deleted: [cardId: number] }>()

interface EditStep { _key: number; type: 'command' | 'choice'; prompt: string; patternsText: string; choicesText: string; answer: number; explanation: string; hint: string }

let keySeq = 0
function toEditStep(s: any): EditStep {
  const type = s.type === 'choice' ? 'choice' : 'command'
  return {
    _key: keySeq++, type,
    prompt: s.prompt ?? '',
    patternsText: (s.patterns ?? []).join('\n'),
    choicesText: (s.choices ?? []).join('\n'),
    answer: typeof s.answer === 'number' ? s.answer : 0,
    explanation: s.explanation ?? '', hint: s.hint ?? '',
  }
}
function linesOf(text: string): string[] { return text.split('\n').map((x) => x.trim()).filter(Boolean) }

const steps = ref<EditStep[]>((props.card?.payload.steps ?? []).map(toEditStep))
const saving = ref(false)
const error = ref('')

// --- sync questions from a cheatsheet section into this challenge ---
const topics = ref<string[]>([])
const syncTopic = ref('')
const syncLang = ref<'vi' | 'en'>('vi')
const syncBusy = ref(false)
const syncMsg = ref('')
const syncErr = ref(false)

async function loadTopics() {
  if (!subjectId.value) return
  try {
    const { data } = await api.get('/course-notes', { params: { subject_id: subjectId.value } })
    const seen = new Set<string>()
    topics.value = []
    for (const n of (data.notes ?? [])) {
      const t = (n.title ?? '').trim()
      if (t && !seen.has(t)) { seen.add(t); topics.value.push(t) }
    }
  } catch { topics.value = [] }
}

interface SyncStep { prompt: string; patterns: string[]; explanation?: string; note?: string; lesson?: string; hint?: string }

async function syncFromCheatsheet() {
  if (!syncTopic.value || !subjectId.value) return
  syncBusy.value = true; syncMsg.value = ''; syncErr.value = false
  try {
    const { data } = await api.post('/course-notes/extract-steps',
      { subject_id: subjectId.value, topic: syncTopic.value, language: syncLang.value },
      { timeout: 180000 })
    const incoming = (data.steps ?? []) as SyncStep[]
    const seen = new Set(steps.value.map((s) => (s.prompt.trim() + '|' + s.patternsText.trim()).toLowerCase()))
    let added = 0
    for (const s of incoming) {
      const patternsText = (s.patterns ?? []).join('\n')
      const key = (s.prompt.trim() + '|' + patternsText.trim()).toLowerCase()
      if (!s.prompt || !patternsText || seen.has(key)) continue
      seen.add(key)
      const hint = s.hint || [s.note, s.lesson].filter(Boolean).join(' ')
      steps.value.push({ _key: keySeq++, type: 'command', prompt: s.prompt, patternsText, choicesText: '', answer: 0, explanation: s.explanation ?? '', hint })
      added++
    }
    syncMsg.value = `Đã thêm ${added} câu từ mục "${syncTopic.value}" (${data.extracted ?? incoming.length} lệnh trích được). Xem lại rồi bấm Lưu.`
  } catch (e) {
    syncErr.value = true; syncMsg.value = errorMessage(e)
  } finally {
    syncBusy.value = false
  }
}

const flowBusy = ref(false)
async function genThinkingFlow() {
  if (!syncTopic.value || !subjectId.value) return
  flowBusy.value = true; syncMsg.value = ''; syncErr.value = false
  try {
    const { data } = await api.post('/course-notes/thinking-flow',
      { subject_id: subjectId.value, topic: syncTopic.value, language: syncLang.value },
      { timeout: 180000 })
    const incoming = (data.steps ?? []) as any[]
    let added = 0
    for (const s of incoming) {
      if (!s.prompt) continue
      const type = s.type === 'choice' ? 'choice' : 'command'
      steps.value.push({
        _key: keySeq++, type,
        prompt: s.prompt,
        patternsText: (s.patterns ?? []).join('\n'),
        choicesText: (s.choices ?? []).join('\n'),
        answer: typeof s.answer === 'number' ? s.answer : 0,
        explanation: s.explanation ?? '', hint: s.hint ?? '',
      })
      added++
    }
    syncMsg.value = `Đã thêm ${added} bước luồng tư duy từ mục "${syncTopic.value}". Xem lại/sửa rồi bấm Lưu.`
  } catch (e) {
    syncErr.value = true; syncMsg.value = errorMessage(e)
  } finally {
    flowBusy.value = false
  }
}

onMounted(loadTopics)

function addStep() {
  steps.value.push({ _key: keySeq++, type: 'command', prompt: '', patternsText: '', choicesText: '', answer: 0, explanation: '', hint: '' })
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
    .map((s) => {
      const base: any = { type: s.type, prompt: s.prompt.trim(), explanation: s.explanation.trim(), hint: s.hint.trim() }
      if (s.type === 'choice') {
        base.choices = linesOf(s.choicesText)
        base.answer = Math.max(0, Math.min(s.answer, base.choices.length - 1))
      } else {
        base.patterns = s.patternsText.split('\n').map((p) => p.trim()).filter(Boolean)
      }
      return base
    })
    .filter((s) => s.prompt && (s.type === 'choice' ? s.choices.length >= 2 : s.patterns.length))
  if (!payloadSteps.length) {
    error.value = 'Cần ít nhất 1 câu hợp lệ (câu hỏi + đáp án regex, hoặc ≥2 lựa chọn cho câu trắc nghiệm).'
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
