<template>
  <div class="card border-0 shadow-sm mb-4">
    <div class="card-body">
      <div class="d-flex flex-wrap justify-content-between align-items-baseline gap-2 mb-2">
        <h6 class="fw-bold mb-0">
          <i class="bi bi-bar-chart-line text-primary"></i> How well you know your cards
          <span v-if="retention" class="badge ms-1 align-middle" :class="retention.cls">{{ retention.label }}</span>
        </h6>
        <small v-if="p && p.answers_30d" class="text-muted">
          Last 30 days: <strong :class="accuracy >= 80 ? 'text-success' : accuracy >= 60 ? 'text-primary' : 'text-danger'">{{ accuracy }}%</strong>
          right ({{ p.correct_30d }} / {{ p.answers_30d }} answers)
        </small>
      </div>

      <div v-if="loading" class="text-muted small"><span class="spinner-border spinner-border-sm me-1"></span>Loading…</div>
      <div v-else-if="error" class="small text-danger">{{ error }}</div>
      <div v-else-if="!p || !p.buckets.total" class="small text-muted">No cards in your list yet: start in Study.</div>
      <template v-else>
        <p class="mb-2">
          You have studied <strong>{{ studied }}</strong> of {{ p.buckets.total }} cards:
          <span class="text-success fw-semibold">{{ p.buckets.mastered + p.buckets.familiar }} known well ({{ pct(p.buckets.mastered + p.buckets.familiar, studied) }}%)</span>,
          <span class="text-danger fw-semibold">{{ p.buckets.struggling }} keep being forgotten ({{ pct(p.buckets.struggling, studied) }}%)</span>.
        </p>

        <!-- advice: learn new words, or focus on review? -->
        <div v-if="advice" class="advice d-flex align-items-start gap-2 p-2 rounded mb-3" :class="advice.cls">
          <i class="bi mt-1" :class="advice.icon"></i>
          <div class="flex-grow-1">
            <div class="fw-semibold">{{ advice.title }}</div>
            <div class="small">{{ advice.detail }}</div>
            <div class="mt-1 d-flex flex-wrap gap-2">
              <router-link v-if="p.due_now" class="btn btn-sm btn-warning py-0" :to="studyLink">
                <i class="bi bi-arrow-repeat"></i> Review {{ p.due_now }} now
              </router-link>
              <router-link class="btn btn-sm btn-outline-secondary py-0" to="/devices">
                <i class="bi bi-sliders"></i> New cards / day
              </router-link>
            </div>
          </div>
        </div>

        <!-- all cards in the list, by group -->
        <div class="progress stacked mb-2" role="img" :aria-label="summary">
          <div v-for="g in GROUPS" :key="g.key" class="progress-bar" :class="g.bar"
            :style="{ width: pct(p.buckets[g.key], p.buckets.total) + '%' }" :title="`${g.label}: ${p.buckets[g.key]}`"></div>
        </div>
        <div class="row g-2 small mb-3">
          <div v-for="g in GROUPS" :key="g.key" class="col-6 col-md">
            <div class="d-flex align-items-start gap-1">
              <span class="dot mt-1" :class="g.bar"></span>
              <div>
                <div><strong>{{ p.buckets[g.key] }}</strong> {{ g.label }} <span class="text-muted">· {{ pct(p.buckets[g.key], p.buckets.total) }}%</span></div>
                <div class="text-muted tiny">{{ g.help }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- per subject -->
        <div v-if="p.subjects.length > 1" class="mb-3">
          <div v-for="s in p.subjects" :key="s.subject_id" class="d-flex align-items-center gap-2 small mb-1">
            <span class="subject-name text-truncate" :title="s.name">{{ s.name }}</span>
            <div class="progress flex-grow-1 thin">
              <div v-for="g in GROUPS" :key="g.key" class="progress-bar" :class="g.bar"
                :style="{ width: pct(s.buckets[g.key], s.buckets.total) + '%' }" :title="`${g.label}: ${s.buckets[g.key]}`"></div>
            </div>
            <span class="text-muted text-nowrap">{{ pct(s.buckets.mastered + s.buckets.familiar, s.buckets.total - s.buckets.new) }}% known</span>
          </div>
        </div>

        <!-- the cards that keep being forgotten -->
        <template v-if="p.weakest.length">
          <div class="d-flex flex-wrap justify-content-between align-items-center gap-2 mb-1">
            <span class="fw-semibold small text-danger"><i class="bi bi-exclamation-circle"></i> Forgotten most often</span>
            <router-link v-if="wordIds.length" class="btn btn-sm btn-outline-danger" :to="`/practice/check?ids=${wordIds.join(',')}`">
              <i class="bi bi-ui-checks"></i> Check these words (A–D)
            </router-link>
          </div>
          <div class="d-flex flex-wrap gap-1">
            <span v-for="w in p.weakest" :key="w.card_id" class="badge weak text-start fw-normal"
              :title="`Forgotten ${w.lapses} time${w.lapses === 1 ? '' : 's'}`">
              <strong>{{ w.front }}</strong><span v-if="w.meaning" class="text-muted"> · {{ w.meaning }}</span>
              <span class="ms-1 text-danger">×{{ w.lapses }}</span>
            </span>
          </div>
        </template>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, errorMessage } from '../api'

type BucketKey = 'mastered' | 'familiar' | 'learning' | 'struggling' | 'new'
interface Buckets { total: number; new: number; struggling: number; learning: number; familiar: number; mastered: number }
interface Progress {
  buckets: Buckets
  subjects: { subject_id: number; name: string; buckets: Buckets }[]
  answers_30d: number
  correct_30d: number
  due_now: number
  weakest: { card_id: number; front: string; kind: string; meaning: string; lapses: number }[]
}

// same rules as the server (api/internal/study/progress.go)
const GROUPS: { key: BucketKey; label: string; help: string; bar: string }[] = [
  { key: 'mastered', label: 'mastered', help: 'remembered for 3+ weeks', bar: 'bg-success' },
  { key: 'familiar', label: 'familiar', help: 'remembered for 1–3 weeks', bar: 'bg-info' },
  { key: 'learning', label: 'learning', help: 'remembered for less than a week', bar: 'bg-warning' },
  { key: 'struggling', label: 'struggling', help: 'forgotten 2+ times, not solid yet', bar: 'bg-danger' },
  { key: 'new', label: 'not started', help: 'in your list, never studied', bar: 'bg-secondary-subtle' },
]

const props = defineProps<{ subjectId?: number }>()

interface Advice { title: string; detail: string; cls: string; icon: string }

const p = ref<Progress | null>(null)
const loading = ref(true)
const error = ref('')

const pct = (n: number, of: number) => (of > 0 ? Math.round((n / of) * 100) : 0)
const studied = computed(() => (p.value ? p.value.buckets.total - p.value.buckets.new : 0))
const accuracy = computed(() => (p.value ? pct(p.value.correct_30d, p.value.answers_30d) : 0))
// the AI check page is for English words only
const wordIds = computed(() => (p.value?.weakest ?? []).filter((w) => w.kind === 'vocab' || w.kind === 'phrasal').map((w) => w.card_id))
const summary = computed(() => (p.value ? GROUPS.map((g) => `${g.label} ${p.value!.buckets[g.key]}`).join(', ') : ''))
const studyLink = computed(() => (props.subjectId ? `/review?subject_id=${props.subjectId}` : '/review'))

/**
 * Should the learner add new words, or focus on review? Uses the backlog of
 * due cards, the share of studied cards that keep being forgotten, and the
 * 30-day accuracy. New cards are cheap to add but each one brings many future
 * reviews, so when memory is under strain the advice is to pause them.
 */
/** Overall memory health, from recall accuracy and how much is solid vs slipping. */
const retention = computed<{ label: string; cls: string } | null>(() => {
  const v = p.value
  if (!v || !v.buckets.total) return null
  const done = v.buckets.total - v.buckets.new
  if (done < 5) return { label: 'Just started', cls: 'text-bg-secondary' }
  const knownWell = done ? (v.buckets.mastered + v.buckets.familiar) / done : 0
  const struggleRatio = done ? v.buckets.struggling / done : 0
  const acc = v.answers_30d ? v.correct_30d / v.answers_30d : null
  if (struggleRatio >= 0.2 || knownWell < 0.3 || (acc !== null && acc < 0.65)) {
    return { label: 'Needs focus', cls: 'text-bg-danger' }
  }
  if (knownWell >= 0.6 && struggleRatio < 0.1 && (acc === null || acc >= 0.8)) {
    return { label: 'Strong', cls: 'text-bg-success' }
  }
  return { label: 'Okay', cls: 'text-bg-warning' }
})

const advice = computed<Advice | null>(() => {
  const v = p.value
  if (!v || !v.buckets.total) return null
  const done = v.buckets.total - v.buckets.new
  const due = v.due_now
  const struggleRatio = done ? v.buckets.struggling / done : 0
  const acc = v.answers_30d ? v.correct_30d / v.answers_30d : null // null = too little data

  if (done < 10) {
    return { title: 'Just getting started', icon: 'bi-rocket-takeoff', cls: 'ad-ok',
      detail: 'Learn a small batch (about 5–10 a day) and watch how the reviews feel over the next few days before adding more.' }
  }
  // under strain → review, don't add new
  if (due >= 30 || struggleRatio >= 0.25 || (acc !== null && acc < 0.65)) {
    const why = due >= 30 ? `${due} cards are due`
      : struggleRatio >= 0.25 ? `${Math.round(struggleRatio * 100)}% of what you've studied keeps slipping`
        : `your recent accuracy is ${Math.round((acc ?? 0) * 100)}%`
    return { title: 'Focus on review — hold off on new words', icon: 'bi-exclamation-triangle', cls: 'ad-stop',
      detail: `${cap(why)}. Adding new words now would pile on more reviews. Clear the backlog and let it settle first.` }
  }
  // some pressure → keep new words light
  if (due >= 15 || struggleRatio >= 0.12 || (acc !== null && acc < 0.8)) {
    return { title: 'Keep new words light', icon: 'bi-cup-hot', cls: 'ad-slow',
      detail: 'Do your reviews first, then add only a few new words a day (around 5) until more of them move to “familiar”.' }
  }
  // healthy → safe to add more
  return { title: 'Memory looks healthy — you can add more', icon: 'bi-check-circle', cls: 'ad-ok',
    detail: 'Reviews are under control and recall is strong. It’s a good time to raise your new-cards-per-day (e.g. to 10–15) if you want to move faster.' }
})

function cap(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1)
}

onMounted(async () => {
  try {
    p.value = (await api.get('/stats/progress', { params: { subject_id: props.subjectId || undefined } })).data
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.stacked { height: 14px; }
.thin { height: 8px; }
.dot { display: inline-block; width: 10px; height: 10px; border-radius: 3px; flex-shrink: 0; }
.tiny { font-size: 0.72rem; }
.subject-name { width: 9rem; }
.weak { background: #fef2f2; color: #1e293b; border: 1px solid #fecaca; white-space: normal; }
.advice.ad-ok { background: #f0fdf4; border: 1px solid #bbf7d0; }
.advice.ad-ok .bi { color: #16a34a; }
.advice.ad-slow { background: #fffbeb; border: 1px solid #fde68a; }
.advice.ad-slow .bi { color: #d97706; }
.advice.ad-stop { background: #fef2f2; border: 1px solid #fecaca; }
.advice.ad-stop .bi { color: #dc2626; }
</style>
