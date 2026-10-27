<template>
  <div>
    <div v-if="chat.length" ref="chatBox" class="chat mb-2">
      <div v-for="(m, i) in chat" :key="i" class="mb-2 d-flex" :class="m.role === 'you' ? 'justify-content-end' : ''">
        <div class="bubble" :class="m.role === 'you' ? 'you' : 'ai'">
          <div v-if="m.role === 'you'">{{ m.text }}</div>
          <div v-else class="md" v-html="render(m.text)"></div>
        </div>
      </div>
      <div v-if="asking" class="text-muted small"><span class="spinner-border spinner-border-sm me-1"></span>Thinking…</div>
    </div>
    <p v-else class="small text-muted mb-2">Ask anything about <strong>{{ topicName }}</strong> — concepts, commands, examples. The AI also sees your notes for this course.</p>

    <div v-if="models.length" class="mb-2 d-flex align-items-center gap-2 small">
      <label class="text-muted mb-0">Model</label>
      <select v-model="model" class="form-select form-select-sm w-auto">
        <option value="">Default</option>
        <option v-for="m in models" :key="m" :value="m">{{ m }}</option>
      </select>
    </div>
    <form class="d-flex gap-2" @submit.prevent="ask">
      <input v-model="question" class="form-control" :placeholder="`Ask about ${topicName}…`" :disabled="asking" maxlength="1000" />
      <button class="btn btn-primary" :disabled="asking || !question.trim()"><i class="bi bi-send"></i></button>
      <button v-if="chat.length" type="button" class="btn btn-outline-secondary" :disabled="asking" title="Clear chat" @click="chat = []"><i class="bi bi-trash"></i></button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { api, errorMessage } from '../api'
import { shortName, subject, subjectId } from '../composables/useSubject'
import { renderMarkdown } from '../utils/markdown'

// Reusable "Ask AI about this course" chat. Follows the topic chosen in the
// menu and is grounded (server-side) in the user's notes for that course.
// Used inline on the Cheat sheet page and inside the floating AskAiWidget.

interface ChatMsg { role: 'you' | 'ai'; text: string }

const chat = ref<ChatMsg[]>([])
const question = ref('')
const asking = ref(false)
const chatBox = ref<HTMLElement | null>(null)
const models = ref<string[]>([])
const model = ref('')

const topicName = computed(() => (subject.value ? shortName(subject.value.slug, subject.value.name) : 'this course'))
const render = (s: string) => renderMarkdown(s)

async function loadModels() {
  try {
    models.value = (await api.get('/chat-models')).data.models ?? []
  } catch {
    models.value = []
  }
}

async function ask() {
  const q = question.value.trim()
  if (!q || asking.value || !subjectId.value) return
  chat.value.push({ role: 'you', text: q })
  question.value = ''
  asking.value = true
  await scrollChat()
  try {
    const { data } = await api.post('/course-notes/ask',
      { subject_id: subjectId.value, question: q, model: model.value || undefined, history: pairs() },
      { timeout: 120000 })
    chat.value.push({ role: 'ai', text: data.answer ?? '' })
  } catch (e) {
    chat.value.push({ role: 'ai', text: '⚠️ ' + errorMessage(e) })
  } finally {
    asking.value = false
    await scrollChat()
  }
}

function pairs(): { q: string; a: string }[] {
  const out: { q: string; a: string }[] = []
  for (let i = 0; i < chat.value.length - 1; i++) {
    if (chat.value[i].role === 'you' && chat.value[i + 1]?.role === 'ai') {
      out.push({ q: chat.value[i].text, a: chat.value[i + 1].text })
    }
  }
  return out.slice(-6)
}

async function scrollChat() {
  await nextTick()
  if (chatBox.value) chatBox.value.scrollTop = chatBox.value.scrollHeight
}

// a different course = a fresh conversation
watch(subjectId, (id, old) => { if (id !== old) chat.value = [] })
onMounted(loadModels)
</script>

<style scoped>
.chat { max-height: 360px; overflow-y: auto; }
.bubble { max-width: 85%; padding: 0.5rem 0.75rem; border-radius: 12px; }
.bubble.you { background: #4f46e5; color: #fff; border-bottom-right-radius: 3px; }
.bubble.ai { background: #f1f5f9; color: #1e293b; border-bottom-left-radius: 3px; }
.bubble.ai :deep(p:last-child) { margin-bottom: 0; }
.md :deep(pre) { background: #0f172a; color: #e2e8f0; padding: 0.6rem 0.8rem; border-radius: 8px; overflow-x: auto; font-size: 0.85rem; }
.md :deep(code) { background: rgba(2, 6, 23, 0.07); padding: 0.1em 0.3em; border-radius: 4px; font-size: 0.9em; }
.md :deep(pre code) { background: none; padding: 0; }
.md :deep(h1), .md :deep(h2), .md :deep(h3) { font-size: 1.05rem; font-weight: 700; margin: 0.5rem 0 0.3rem; }
.md :deep(ul), .md :deep(ol) { margin: 0.2rem 0 0.4rem 1.1rem; padding: 0; }
.md :deep(p) { margin: 0.3rem 0; }
.md :deep(a) { color: #4f46e5; }
</style>
