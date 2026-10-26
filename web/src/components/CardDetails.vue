<template>
  <div class="card-details">
    <!-- vocab / phrasal: senses with level tabs -->
    <template v-if="senses.length">
      <ul v-if="senses.length > 1" class="nav nav-pills nav-fill mb-2 small">
        <li class="nav-item" v-for="s in senses" :key="s.level">
          <button class="nav-link py-1" type="button"
            :class="{ active: active === s.level }"
            :style="active === s.level ? `background:${color(s.level)};color:white` : ''"
            @click="active = s.level">
            {{ levelLabel(s.level) }}
          </button>
        </li>
      </ul>
      <div v-if="current" class="p-2 rounded bg-light">
        <div class="mb-1"><span class="badge bg-danger me-1">VI</span>{{ current.meaning_vi }}</div>
        <div v-if="current.meaning_en" class="mb-2 text-muted small"><span class="badge bg-primary me-1">EN</span>{{ current.meaning_en }}</div>
        <div v-if="current.example_en" class="border-start border-3 border-primary ps-2 mb-1">
          <em>{{ current.example_en }}</em>
          <button v-if="speakable" class="btn btn-link btn-sm p-0 ms-1" type="button" title="Listen"
            @click="$emit('speak', current.example_en)">
            <i class="bi bi-volume-up"></i>
          </button>
        </div>
        <div v-if="current.example_vi" class="border-start border-3 border-warning ps-2 text-muted"><em>{{ current.example_vi }}</em></div>
        <small v-if="current.usage_note" class="text-muted d-block mt-2"><i class="bi bi-info-circle"></i> {{ current.usage_note }}</small>
      </div>
      <small v-if="card.kind === 'phrasal'" class="text-muted d-block mt-1">
        {{ card.payload.separable ? 'Separable: “give it up” is OK' : 'Inseparable: keep verb and particle together' }}
      </small>

      <!-- more example sentences (AI, different situations): fill the blank, listen, say it -->
      <div v-if="speakable" class="mt-2">
        <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none" @click="toggleMore">
          <i :class="moreOpen ? 'bi bi-chevron-down' : 'bi bi-chevron-right'"></i>
          <i class="bi bi-journal-text ms-1"></i> More examples
        </button>
        <WordExamples v-if="moreOpen" :card-id="card.id" :word="card.front" :exclude-id="current?.example_id" class="mt-1" />
      </div>
    </template>

    <!-- concept -->
    <div v-else-if="card.kind === 'concept'" class="p-2 rounded bg-light">
      <div class="fw-semibold fs-5">{{ card.payload.back }}</div>
      <div v-if="card.payload.note_vi" class="text-muted mt-1">{{ card.payload.note_vi }}</div>
      <template v-if="card.payload.group">
        <div v-if="card.payload.example" class="small mt-1 fst-italic" style="white-space: pre-line">“{{ card.payload.example }}”</div>
        <div v-if="card.payload.example_vi" class="small text-muted" style="white-space: pre-line">{{ card.payload.example_vi }}</div>
      </template>
      <div v-else-if="card.payload.example" class="small mt-1"><i class="bi bi-calculator"></i> {{ card.payload.example }}</div>
    </div>

    <!-- mcq -->
    <div v-else-if="card.kind === 'mcq'" class="p-2 rounded bg-light">
      <ol type="A" class="mb-2 ps-3">
        <li v-for="(c, i) in card.payload.choices" :key="i" :class="{ 'fw-bold text-success': i === card.payload.answer }">{{ c }}</li>
      </ol>
      <div v-if="card.payload.explanation" class="small text-muted">{{ card.payload.explanation }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Card } from '../types'
import { LEVELS } from '../types'
import WordExamples from './WordExamples.vue'

const props = defineProps<{ card: Card; speakable?: boolean }>()
defineEmits<{ speak: [en: string] }>() // English only: the Vietnamese is shown, not read

const senses = computed(() => [...(props.card.payload.senses ?? [])].sort((a, b) => a.level - b.level))
const active = ref(1)
watch(() => props.card.id, () => { active.value = senses.value[0]?.level ?? 1 }, { immediate: true })
const current = computed(() => senses.value.find((s) => s.level === active.value) ?? senses.value[0])

// "More examples" stays open/closed as the learner left it
const moreOpen = ref(storageGet('elearn.moreExamples') === 'true')
function toggleMore() {
  moreOpen.value = !moreOpen.value
  try { localStorage.setItem('elearn.moreExamples', String(moreOpen.value)) } catch { /* ignore */ }
}
function storageGet(k: string) {
  try { return localStorage.getItem(k) } catch { return null }
}

const color = (l: number) => LEVELS.find((x) => x.level === l)?.color ?? '#6c757d'
const levelLabel = (l: number) => LEVELS.find((x) => x.level === l)?.label ?? `L${l}`
</script>
