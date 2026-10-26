<template>
  <div>
    <div class="d-flex flex-wrap justify-content-between align-items-end mb-4 gap-2">
      <div>
        <h1 class="fw-bold mb-1">Hi{{ auth.user?.display_name ? `, ${auth.user.display_name}` : '' }} 👋</h1>
        <p class="text-muted mb-0">{{ headline }}</p>
      </div>
      <router-link :to="studyLink" class="btn btn-warning btn-lg shadow-sm" :class="{ disabled: !stats }">
        <i class="bi bi-arrow-repeat me-1"></i> Study now
        <span v-if="stats" class="badge bg-dark ms-1">{{ stats.due_now + stats.new_left_today }}</span>
      </router-link>
    </div>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <div v-if="stats" class="row g-3 mb-4">
      <div class="col-6 col-md-3" v-for="t in tiles" :key="t.label">
        <div class="card border-0 shadow-sm h-100 text-center">
          <div class="card-body">
            <i :class="`bi ${t.icon} fs-3 text-${t.color}`"></i>
            <h3 class="mt-1 mb-0 fw-bold">{{ t.value }}</h3>
            <small class="text-muted">{{ t.label }}</small>
          </div>
        </div>
      </div>
    </div>

    <ProgressCard v-if="stats && subjectId" :key="subjectId" :subject-id="subjectId" />

    <div v-if="stats" class="card border-0 shadow-sm mb-4">
      <div class="card-body">
        <h6 class="fw-bold mb-3">Reviews in the last 14 days</h6>
        <div class="d-flex align-items-end gap-1" style="height: 90px">
          <div v-for="d in stats.reviews_last_14_days" :key="d.day" class="flex-fill text-center" :title="`${d.day}: ${d.count}`">
            <div class="bg-primary rounded-top mx-auto" :style="{ height: barHeight(d.count) + 'px', width: '70%', opacity: d.count ? 1 : 0.15 }"></div>
            <small class="text-muted" style="font-size: 0.65rem">{{ d.day.slice(8) }}</small>
          </div>
        </div>
      </div>
    </div>

    <div v-for="s in shown" :key="s.id" class="mb-4">
      <div class="d-flex justify-content-between align-items-center mb-2">
        <div>
          <h4 class="fw-bold mb-0">{{ s.name }}</h4>
          <small class="text-muted">{{ s.description }}</small>
        </div>
        <span v-if="subjectStats(s.id)" class="small text-muted text-end">
          {{ subjectStats(s.id)!.due }} due · {{ subjectStats(s.id)!.learning }} learning · {{ subjectStats(s.id)!.mastered }} mastered
        </span>
      </div>
      <div class="row g-3">
        <div v-for="d in s.decks" :key="d.id" class="col-md-6 col-lg-3">
          <div class="card border-0 shadow-sm h-100 hoverable">
            <div class="card-body d-flex flex-column">
              <h6 class="fw-bold mb-1">{{ d.name }}</h6>
              <small class="text-muted flex-grow-1">{{ d.description }}</small>
              <small class="text-muted my-2">{{ d.learning_count }} / {{ d.card_count }} in your list</small>
              <div class="d-flex gap-1">
                <router-link :to="`/review?deck_id=${d.id}`" class="btn btn-sm btn-warning flex-fill">Study</router-link>
                <router-link :to="`/learn?deck_id=${d.id}`" class="btn btn-sm btn-outline-primary flex-fill">Browse</router-link>
                <router-link :to="`/quiz?deck_id=${d.id}`" class="btn btn-sm btn-outline-success flex-fill">Quiz</router-link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api, errorMessage } from '../api'
import { auth } from '../auth'
import type { Subject } from '../types'
import ProgressCard from '../components/ProgressCard.vue'
import { loadSubjects, shortName, subject, subjectId, subjects as allSubjects } from '../composables/useSubject'

interface Stats {
  due_now: number
  new_left_today: number
  reviews_today: number
  correct_today: number
  streak_days: number
  learning: number
  mastered: number
  subjects: { subject_id: number; due: number; learning: number; mastered: number }[]
  reviews_last_14_days: { day: string; count: number }[]
}

const stats = ref<Stats | null>(null)
const error = ref('')
// only the topic chosen in the menu
const shown = computed<Subject[]>(() => allSubjects.value.filter((x) => x.id === subjectId.value))
const studyLink = computed(() => (subjectId.value ? `/review?subject_id=${subjectId.value}` : '/review'))

const headline = computed(() => {
  if (!stats.value) return 'Loading your progress…'
  const s = stats.value
  const topic = subject.value ? shortName(subject.value.slug, subject.value.name) + ': ' : ''
  if (s.due_now === 0 && s.new_left_today === 0) return `${topic}all done for today. Nice work!`
  return `${topic}${s.due_now} card(s) to review and ${s.new_left_today} new card(s) left today.`
})

const tiles = computed(() => {
  const s = stats.value!
  return [
    { label: 'Due now', value: s.due_now, icon: 'bi-clock', color: 'warning' },
    { label: 'Reviewed today', value: `${s.reviews_today}`, icon: 'bi-check2-circle', color: 'success' },
    { label: 'Day streak', value: s.streak_days, icon: 'bi-fire', color: 'danger' },
    { label: 'Mastered', value: `${s.mastered} / ${s.learning}`, icon: 'bi-trophy', color: 'primary' },
  ]
})

const maxDay = computed(() => Math.max(1, ...(stats.value?.reviews_last_14_days.map((d) => d.count) ?? [1])))
const barHeight = (n: number) => (n ? Math.max(4, Math.round((n / maxDay.value) * 70)) : 4)
const subjectStats = (id: number) => stats.value?.subjects.find((x) => x.subject_id === id)

async function load() {
  await loadSubjects(true) // fresh deck counts
  if (!subjectId.value) return
  try {
    stats.value = (await api.get('/stats', { params: { subject_id: subjectId.value } })).data
  } catch (e) {
    error.value = errorMessage(e)
  }
}
watch(subjectId, (id, old) => { if (id && id !== old) load() })
onMounted(load)
</script>
