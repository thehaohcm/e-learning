import { ref } from 'vue'
import { api } from '../api'

// A single, app-wide audio player. The HTMLAudioElement lives at module scope
// (not inside a component), so it keeps playing while the user navigates
// between pages, generates questions, etc. A small global player bar (mounted
// once in App.vue) binds to this reactive state for play/pause/seek/close.
//
// It also supports a PLAYLIST: a queue of tracks fetched from authed endpoints
// (note / daily audio). When one ends it auto-advances; next/prev and jumping
// to any track are exposed. Playback speed (1–2x) applies to every track.

const el: HTMLAudioElement | null = typeof Audio !== 'undefined' ? new Audio() : null

export interface QueueItem {
  key: string // unique, e.g. "note-5" / "daily-9" (for highlighting)
  id: number
  name: string
  endpoint: string // authed API path returning the audio bytes
}

export const curSrc = ref('')
export const curTitle = ref('')
export const curKey = ref('')
export const playing = ref(false)
export const currentTime = ref(0)
export const duration = ref(0)
export const loading = ref(false)
export const rate = ref(1)
export const queue = ref<QueueItem[]>([])
export const queueIndex = ref(-1)

// the object URL this player currently owns (created from an authed blob); it
// is revoked when a new track starts or the player is closed.
let ownedUrl = ''

function revokeOwned() {
  if (ownedUrl) { try { URL.revokeObjectURL(ownedUrl) } catch { /* ignore */ } ownedUrl = '' }
}

function applyRate() { if (el) el.playbackRate = rate.value }

// advance to the next queued track, or stop at the end.
function advance() {
  if (queue.value.length && queueIndex.value >= 0 && queueIndex.value < queue.value.length - 1) {
    void playQueueAt(queueIndex.value + 1)
  } else {
    playing.value = false
  }
}

if (el) {
  el.addEventListener('timeupdate', () => { currentTime.value = el.currentTime })
  el.addEventListener('durationchange', () => { duration.value = isFinite(el.duration) ? el.duration : 0 })
  el.addEventListener('loadedmetadata', () => { duration.value = isFinite(el.duration) ? el.duration : 0; applyRate() })
  el.addEventListener('ratechange', () => { if (el.playbackRate !== rate.value) rate.value = el.playbackRate })
  el.addEventListener('play', () => { playing.value = true; applyRate() })
  el.addEventListener('pause', () => { playing.value = false })
  el.addEventListener('ended', () => { advance() })
  el.addEventListener('waiting', () => { loading.value = true })
  el.addEventListener('playing', () => { loading.value = false })
}

// fetch (authed blob) and play the queue track at index i.
async function playQueueAt(i: number) {
  if (!el || i < 0 || i >= queue.value.length) return
  const item = queue.value[i]
  queueIndex.value = i
  curKey.value = item.key
  curTitle.value = item.name
  loading.value = true
  try {
    const { data } = await api.get(item.endpoint, { responseType: 'blob' })
    const url = URL.createObjectURL(data as Blob)
    if (ownedUrl && ownedUrl !== url) revokeOwned()
    ownedUrl = url
    curSrc.value = url
    currentTime.value = 0
    duration.value = 0
    el.src = url
    applyRate()
    await el.play().catch(() => { /* autoplay blocked / interrupted */ })
  } catch {
    loading.value = false
    playing.value = false
  }
}

export function useAudioPlayer() {
  // play a direct URL (static file / caller-made blob). owned=true hands a blob
  // URL's lifecycle to the player. This is a single track: it clears any queue.
  function play(url: string, title: string, owned = false) {
    if (!el || !url) return
    queue.value = []
    queueIndex.value = -1
    curKey.value = ''
    if (curSrc.value !== url) {
      if (ownedUrl && ownedUrl !== url) revokeOwned()
      ownedUrl = owned ? url : ''
      curSrc.value = url
      curTitle.value = title
      currentTime.value = 0
      duration.value = 0
      el.src = url
      applyRate()
    }
    el.play().catch(() => { /* autoplay blocked / interrupted */ })
  }
  // play a playlist of authed tracks, starting at startIndex.
  function playQueue(items: QueueItem[], startIndex = 0) {
    if (!items.length) return
    queue.value = items.slice()
    void playQueueAt(Math.max(0, Math.min(startIndex, items.length - 1)))
  }
  function nextTrack() { advance() }
  function prevTrack() { if (queueIndex.value > 0) void playQueueAt(queueIndex.value - 1) }
  function setRate(r: number) { rate.value = r; applyRate() }
  function toggle() {
    if (!el || !curSrc.value) return
    if (el.paused) el.play().catch(() => {})
    else el.pause()
  }
  function stop() {
    if (!el) return
    el.pause()
    el.removeAttribute('src')
    el.load()
    revokeOwned()
    curSrc.value = ''
    curTitle.value = ''
    curKey.value = ''
    queue.value = []
    queueIndex.value = -1
    currentTime.value = 0
    duration.value = 0
    playing.value = false
  }
  function seek(t: number) { if (el) el.currentTime = t }
  function isCurrent(url: string) { return curSrc.value === url }
  return {
    curSrc, curTitle, curKey, playing, currentTime, duration, loading, rate, queue, queueIndex,
    play, playQueue, nextTrack, prevTrack, setRate, toggle, stop, seek, isCurrent,
  }
}
