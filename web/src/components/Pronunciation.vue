<template>
  <div v-if="p && p.syllables.length" class="pron" :class="{ compact, start }">
    <span class="syllables" :title="`Stress on “${p.syllables[p.stress]}”`">
      <template v-for="(s, i) in p.syllables" :key="i">
        <span v-if="i" class="dot">·</span><span :class="i === p.stress ? 'stressed' : 'weak'">{{ i === p.stress ? s.toUpperCase() : s }}</span>
      </template>
    </span>
    <button type="button" class="btn btn-link btn-sm p-0 ms-1 align-baseline" title="Syllable by syllable, then the whole word"
      :disabled="playing" @click="playSlow">🐢</button>
    <template v-if="p.tips_vi.length">
      <button v-if="compact" type="button" class="btn btn-link btn-sm p-0 ms-2 align-baseline small" @click="showTips = !showTips">
        <i class="bi bi-lightbulb"></i> {{ showTips ? 'Hide tips' : `Tips (${p.tips_vi.length})` }}
      </button>
      <ul v-if="!compact || showTips" class="tips list-unstyled small mb-0 mt-1">
        <li v-for="(t, i) in p.tips_vi" :key="i"><i class="bi bi-lightbulb text-warning"></i> {{ t }}</li>
      </ul>
    </template>
  </div>
  <!-- not explained yet: AI only when asked -->
  <div v-else-if="cache.has(cardId)" class="pron-ask" :class="{ compact }">
    <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none" :disabled="generating.has(cardId)"
      title="Syllables, stress and tips for Vietnamese speakers (AI)" @click="generate(cardId)">
      <span v-if="generating.has(cardId)" class="spinner-border spinner-border-sm me-1"></span>
      <i v-else class="bi bi-soundwave"></i> How to say it
    </button>
    <span v-if="errors.get(cardId)" class="small text-danger ms-1">{{ errors.get(cardId) }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { usePronunciation } from '../composables/usePronunciation'
import { RATE_NORMAL, RATE_SLOW, useSpeechSynthesis } from '../composables/useSpeechSynthesis'

/**
 * How to say the word: hy·POTH·e·sis (stressed syllable in capitals), 🐢 reads
 * it syllable by syllable then whole, and tips in Vietnamese for the hard sounds.
 */
// start: left-aligned layout (tips under the word, not centred)
const props = defineProps<{ cardId: number; word: string; compact?: boolean; start?: boolean }>()
const { cache, generating, errors, request, generate } = usePronunciation()
const { speakEnglish } = useSpeechSynthesis()
const showTips = ref(false)
const playing = ref(false)

const p = computed(() => cache.get(props.cardId) ?? null)
watch(() => props.cardId, (id) => { showTips.value = false; request(id) }, { immediate: true })

const pause = (ms: number) => new Promise((r) => setTimeout(r, ms))

async function playSlow() {
  const pr = p.value
  if (!pr) return
  playing.value = true
  try {
    for (const s of pr.syllables) {
      await speakEnglish(s, RATE_SLOW)
      await pause(250)
    }
    await pause(200)
    await speakEnglish(props.word, RATE_NORMAL)
  } catch {
    /* interrupted */
  } finally {
    playing.value = false
  }
}
</script>

<style scoped>
.syllables { letter-spacing: 0.02em; }
.stressed { color: var(--bs-primary); font-weight: 700; }
.weak { color: var(--bs-secondary); }
.dot { color: var(--bs-gray-400); margin: 0 0.15rem; }
.tips li { line-height: 1.4; margin-top: 0.15rem; }
/* tips on their own lines under the word: centred block in centred layouts, left in "start" */
.pron:not(.compact) .tips { display: block; width: fit-content; max-width: 100%; margin: 0.25rem auto 0; text-align: left; }
.pron.start .tips { margin-left: 0; }
</style>
