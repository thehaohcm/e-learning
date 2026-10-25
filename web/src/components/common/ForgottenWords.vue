<template>
  <div v-if="words.length" class="card border-warning shadow-sm text-start" :class="compact ? '' : 'mb-3'">
    <div class="card-body" :class="compact ? 'py-2 px-3' : ''">
      <div class="d-flex flex-wrap align-items-center gap-2">
        <span class="fw-semibold"><i class="bi bi-lightbulb text-warning"></i> {{ words.length }} word{{ words.length > 1 ? 's' : '' }} to practise:</span>
        <span v-for="w in words.slice(0, 10)" :key="w.id" class="badge bg-warning-subtle text-dark border">{{ w.front }}</span>
        <span v-if="words.length > 10" class="small text-muted">+{{ words.length - 10 }} more (next time)</span>
        <button class="btn btn-outline-warning btn-sm ms-auto" title="One multiple-choice question per word (usage, closest meaning, collocation)"
          @click="check">
          <i class="bi bi-ui-checks"></i> Check them (A–D)
        </button>
        <button class="btn btn-warning btn-sm" @click="go">
          <i class="bi bi-translate"></i> Practise them
        </button>
      </div>
      <div v-if="!compact" class="small text-muted mt-2">
        3 new example sentences per word: translate each into Vietnamese, then read it aloud.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const props = defineProps<{ words: { id: number; front: string }[]; compact?: boolean }>()
const router = useRouter()

function check() {
  router.push({ path: '/practice/check', query: { ids: props.words.slice(0, 10).map((w) => w.id).join(',') } })
}

function go() {
  router.push({ path: '/practice/forgotten', query: { ids: props.words.slice(0, 10).map((w) => w.id).join(',') } })
}
</script>
