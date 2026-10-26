<template>
  <!-- gentle in-page reminder: no sound, no pop-up; hidden for one interval when closed -->
  <div v-if="show" class="nudge alert alert-warning d-flex align-items-center gap-2 py-2 px-3 mb-3 shadow-sm" role="status">
    <i class="bi bi-alarm"></i>
    <span class="flex-grow-1">
      <strong>{{ due }} {{ topic }} card{{ due > 1 ? 's' : '' }} due</strong>
      <span class="text-muted small ms-1">· a few minutes now keeps them fresh</span>
    </span>
    <button type="button" class="btn btn-sm btn-warning" @click="study">
      <i class="bi bi-play-fill"></i> Study now
    </button>
    <button type="button" class="btn-close" aria-label="Later" title="Remind me later" @click="later"></button>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api'
import { shortName, subject, subjectId } from '../composables/useSubject'

// Web counterpart of the device pop-up: while the site is open (and visible),
// once per "pop-up every N min" (the device's setting) inside its active hours, a
// banner says how many cards are due. Nothing if the user studied (web or
// device) during the last interval, or is already studying.

const route = useRoute()
const router = useRouter()
const show = ref(false)
const due = ref(0)
const topic = ref('') // the topic shown in the menu (English / PMP)

const KEY = 'elearn.nudge.at' // last time the banner was closed / used
const QUIET_PATHS = ['/', '/review', '/quiz', '/practice/check', '/login', '/register']

let everyMin = 45
let activeFrom = '07:30'
let activeTo = '22:00'
let settingsAt = 0
let statsAt = 0
let timer: number | undefined

function storedAt(): number {
  try { return Number(localStorage.getItem(KEY)) || 0 } catch { return 0 }
}
function storeAt(t: number) {
  try { localStorage.setItem(KEY, String(t)) } catch { /* ignore */ }
}

/** "07:30" <= now < "22:00" in the browser's time (also handles ranges over midnight). */
function inActiveHours(d: Date): boolean {
  const hm = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  return activeFrom <= activeTo ? hm >= activeFrom && hm < activeTo : hm >= activeFrom || hm < activeTo
}

const hhmm = (v: unknown) => typeof v === 'string' && /^\d\d:\d\d$/.test(v)

/** The same interval and hours as the study device (its own settings, else the account's). */
async function loadSettings() {
  try {
    const [acc, devs] = await Promise.all([api.get('/settings'), api.get('/devices').catch(() => ({ data: [] }))])
    const a = acc.data.settings ?? {}
    const d: Record<string, unknown> = (Array.isArray(devs.data) ? devs.data : [])
      .map((x: { settings?: Record<string, unknown> }) => x.settings ?? {})
      .sort((x: Record<string, unknown>, y: Record<string, unknown>) =>
        (Number(x.reminder_every_min) || 9999) - (Number(y.reminder_every_min) || 9999))[0] ?? {}
    const every = Number(d.reminder_every_min) || Number(a.reminder_every_min)
    if (every >= 5) everyMin = every
    activeFrom = hhmm(d.active_from) ? (d.active_from as string) : hhmm(a.active_from) ? a.active_from : activeFrom
    activeTo = hhmm(d.active_to) ? (d.active_to as string) : hhmm(a.active_to) ? a.active_to : activeTo
  } catch { /* keep the defaults */ }
  settingsAt = Date.now()
}

async function tick() {
  const now = Date.now()
  if (document.hidden || QUIET_PATHS.includes(route.path) || !inActiveHours(new Date())) {
    show.value = false
    return
  }
  if (show.value) return
  if (now - settingsAt > 60 * 60_000) await loadSettings()
  const every = everyMin * 60_000
  // closed or used during the last interval, or asked the server not long ago
  if (now - storedAt() < every || now - statsAt < Math.min(every, 10 * 60_000)) return
  statsAt = now
  try {
    const { data } = await api.get('/stats', { params: { subject_id: subjectId.value || undefined } })
    const last = data.last_review_at ? Date.parse(data.last_review_at) : 0
    if (now - last < every) return // studied recently, on the web or a device
    due.value = data.due_now ?? 0
    topic.value = subject.value ? shortName(subject.value.slug, subject.value.name) : ''
    show.value = due.value > 0
  } catch { /* offline: try again later */ }
}

function later() {
  show.value = false
  storeAt(Date.now())
}

function study() {
  later()
  router.push(subjectId.value ? `/review?subject_id=${subjectId.value}` : '/review')
}

watch(() => route.path, () => {
  if (QUIET_PATHS.includes(route.path)) show.value = false
})
// another topic chosen: its own count (asked again at the next check)
watch(subjectId, () => {
  show.value = false
  statsAt = 0
})

function onVisible() {
  if (!document.hidden) tick()
}

onMounted(async () => {
  await loadSettings()
  tick()
  timer = window.setInterval(tick, 60_000)
  document.addEventListener('visibilitychange', onVisible)
})
onUnmounted(() => {
  clearInterval(timer)
  document.removeEventListener('visibilitychange', onVisible)
})
</script>

<style scoped>
.nudge { border-radius: 12px; }
</style>
