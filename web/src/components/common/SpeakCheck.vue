<template>
  <div v-if="supported" class="speak-check">
    <button type="button" class="btn btn-sm" :class="listening ? 'btn-danger' : 'btn-outline-danger'"
      :title="listening ? 'Stop' : result ? 'Listen to it slowly, then say it again (R)' : 'Read the sentence aloud and check your pronunciation (R)'"
      :disabled="playing" @click="toggle">
      <i :class="listening ? 'bi bi-stop-circle-fill' : playing ? 'bi bi-volume-up-fill' : 'bi bi-mic-fill'"></i>
      {{ listening ? 'Your turn… tap to stop' : playing ? 'Listen first…' : result ? 'Try again' : (label || 'Read it aloud') }}
    </button>

    <div v-if="listening && transcript" class="small text-muted mt-2 fst-italic">{{ transcript }}…</div>
    <div v-if="error" class="small text-danger mt-2">{{ error }}</div>

    <div v-if="result && !listening" class="mt-2 small">
      <div class="d-flex align-items-center gap-2 mb-1 justify-content-center">
        <span class="badge" :class="badgeClass">{{ result.pct }}%</span>
        <span :class="textClass" class="fw-semibold">{{ verdict }}</span>
      </div>
      <div class="words">
        <span v-for="(w, i) in result.words" :key="i" :class="w.ok ? 'ok' : 'miss'">{{ w.word }} </span>
      </div>
      <div v-if="result.heard" class="text-muted mt-1">You said: “{{ result.heard }}”</div>
      <div v-else class="text-muted mt-1">Nothing recognised.</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { scoreSpeech, useSpeechRecognition } from '../../composables/useSpeechRecognition'
import { RATE_REPLAY, useSpeechSynthesis } from '../../composables/useSpeechSynthesis'
import type { SpeechScore } from '../../composables/useSpeechRecognition'

const props = defineProps<{
  text: string
  label?: string // text of the first button
  modelFirst?: boolean // play the sentence before the first attempt too (the learner has not heard it yet)
}>()
const emit = defineEmits<{ (e: 'scored', pct: number): void }>()
const { supported, listening, transcript, error, listen, stop } = useSpeechRecognition()
const { speakEnglish } = useSpeechSynthesis()
const result = ref<SpeechScore | null>(null)
const playing = ref(false) // reading the model sentence before "Try again"

// a new sentence (next card) clears the previous attempt
watch(() => props.text, () => {
  if (listening.value) stop()
  result.value = null
  error.value = ''
})

const verdict = computed(() => {
  const p = result.value?.pct ?? 0
  if (p >= 90) return 'Excellent!'
  if (p >= 70) return 'Good, check the red words'
  if (p >= 40) return 'Getting there, listen and try again'
  return 'Try again slowly'
})
const badgeClass = computed(() => {
  const p = result.value?.pct ?? 0
  return p >= 90 ? 'bg-success' : p >= 70 ? 'bg-primary' : p >= 40 ? 'bg-warning text-dark' : 'bg-danger'
})
const textClass = computed(() => {
  const p = result.value?.pct ?? 0
  return p >= 90 ? 'text-success' : p >= 70 ? 'text-primary' : p >= 40 ? 'text-warning' : 'text-danger'
})

/**
 * First attempt: listen straight away (the sentence was just read), unless modelFirst.
 * "Try again": read the sentence slowly first so it can be imitated, then listen.
 */
async function start() {
  if (!props.text || listening.value || playing.value) return
  const target = props.text
  if (result.value || props.modelFirst) {
    playing.value = true
    try {
      await speakEnglish(target, RATE_REPLAY)
      await new Promise((r) => setTimeout(r, 250))
    } catch {
      /* interrupted: still let the user try */
    } finally {
      playing.value = false
    }
    if (target !== props.text) return
  }
  const heard = await listen('en-US')
  if (target !== props.text) return // card changed while listening
  if (!heard.length && error.value) return
  result.value = scoreSpeech(target, heard)
  emit('scored', result.value.pct)
}

function toggle() {
  if (listening.value) stop()
  else start()
}

defineExpose({ toggle })
</script>

<style scoped>
.words { line-height: 1.8; }
.words .ok { color: var(--bs-success); }
.words .miss { color: var(--bs-danger); text-decoration: underline wavy; text-underline-offset: 3px; }
</style>
