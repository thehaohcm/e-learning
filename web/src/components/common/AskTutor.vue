<template>
  <div class="ask-tutor">
    <button v-if="!open" type="button" class="btn btn-link btn-sm p-0 text-decoration-none" @click="openPanel">
      <i class="bi bi-chat-dots"></i> Ask about this
    </button>

    <div v-else class="panel rounded border p-2">
      <div class="d-flex align-items-center mb-1">
        <span class="small fw-semibold"><i class="bi bi-chat-dots text-primary"></i> Ask about this sentence</span>
        <button type="button" class="btn-close btn-sm ms-auto" aria-label="Close" @click="open = false"></button>
      </div>

      <div v-for="(t, i) in thread" :key="i" class="mb-2 small">
        <div class="q"><i class="bi bi-person-circle text-muted"></i> {{ t.q }}</div>
        <div class="a mt-1">{{ t.a }}</div>
        <div v-if="t.words.length" class="mt-1 d-flex flex-wrap align-items-center gap-1">
          <span class="text-muted">Save to notes:</span>
          <button v-for="w in t.words" :key="w" type="button" class="btn btn-sm py-0 px-2"
            :class="saved.has(w.toLowerCase()) ? 'btn-success' : 'btn-outline-secondary'" :disabled="saved.has(w.toLowerCase())"
            @click="saveWord(w, t)">
            <i :class="saved.has(w.toLowerCase()) ? 'bi bi-check2' : 'bi bi-plus'"></i> {{ w }}
          </button>
        </div>
      </div>

      <div v-if="!thread.length" class="small text-muted mb-1">
        e.g. <button v-for="ex in examples" :key="ex" type="button" class="btn btn-link btn-sm p-0 me-2 align-baseline" @click="question = ex">“{{ ex }}”</button>
      </div>

      <form class="d-flex gap-2" @submit.prevent="send">
        <input ref="input" v-model="question" class="form-control form-control-sm" maxlength="300"
          :placeholder="thread.length ? 'Ask a follow-up…' : 'Your question (Vietnamese or English)…'" :disabled="asking" @keydown.stop />
        <button class="btn btn-primary btn-sm flex-shrink-0" :disabled="asking || !question.trim()">
          <span v-if="asking" class="spinner-border spinner-border-sm"></span>
          <i v-else class="bi bi-send"></i>
        </button>
      </form>
      <div v-if="error" class="small text-danger mt-1">{{ error }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { api, errorMessage } from '../../api'

/**
 * "Ask about this": a short follow-up chat under an AI verdict (Write it,
 * Writing, forgotten-words practice). The exercise is sent as context, so the
 * learner only types the question. One AI request per question; words the
 * answer is about can be saved to the notes.
 */
interface TutorContext { task?: string; word?: string; learner?: string; corrected?: string; feedback?: string; sample?: string }
interface Turn { q: string; a: string; words: string[] }

const props = defineProps<{ context: TutorContext }>()

const open = ref(false)
const question = ref('')
const asking = ref(false)
const error = ref('')
const thread = ref<Turn[]>([])
const saved = reactive(new Set<string>())
const input = ref<HTMLInputElement | null>(null)

// a new sentence starts a new thread
watch(() => [props.context.task, props.context.learner], () => {
  thread.value = []
  question.value = ''
  error.value = ''
  open.value = false
})

const examples = computed(() => {
  const w = props.context.word
  const out = ['Why is my sentence wrong?']
  if (w) out.unshift(`Khi nào dùng “${w}”?`)
  return out
})

async function openPanel() {
  open.value = true
  await nextTick()
  input.value?.focus()
}

async function send() {
  const q = question.value.trim()
  if (!q || asking.value) return
  asking.value = true
  error.value = ''
  try {
    const { data } = await api.post('/practice/ask', {
      context: props.context,
      history: thread.value.slice(-4).map((t) => ({ q: t.q, a: t.a })),
      question: q,
    })
    thread.value.push({ q, a: data.answer, words: data.words ?? [] })
    question.value = ''
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    asking.value = false
    await nextTick()
    input.value?.focus()
  }
}

/** Adds the word to My notes (the notes panel updates itself). */
async function saveWord(word: string, t: Turn) {
  try {
    const { data } = await api.post('/notes', {
      text: word,
      note: t.a.slice(0, 300),
      context: props.context.corrected || props.context.sample || '',
      source: 'Ask about this',
    })
    saved.add(word.toLowerCase())
    window.dispatchEvent(new CustomEvent('elearn:note-saved', { detail: data }))
  } catch (e) {
    error.value = errorMessage(e)
  }
}
</script>

<style scoped>
.panel { background: var(--bs-body-bg); }
.q { font-weight: 600; }
.a { white-space: pre-wrap; line-height: 1.45; background: var(--bs-light); border-radius: 0.4rem; padding: 0.4rem 0.6rem; }
</style>
