import { ref } from 'vue'
import { api } from '../api'

// A single, app-wide audio player. The HTMLAudioElement lives at module scope
// (not inside a component), so it keeps playing while the user navigates
// between pages, generates questions, etc. A small global player bar (mounted
// once in App.vue) binds to this reactive state for play/pause/seek/close.
//
// Playlist + iOS background playback:
//  - tracks are PRELOADED as blobs up front, so advancing to the next track
//    needs no network. iOS suspends JS/network when the screen locks, so a
//    fetch-on-advance would stall until you wake the device; a cached blob lets
//    playback continue in the background.
//  - the Media Session API publishes title + next/prev to the lock screen and
//    keeps the audio session alive.

const el: HTMLAudioElement | null = typeof Audio !== 'undefined' ? new Audio() : null
if (el) { el.preload = 'auto' }

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

// object URL for a single (non-queue) play; revoked when replaced/closed.
let ownedUrl = ''
// preloaded queue blobs: track key -> object URL. Revoked when the queue is
// replaced or the player is closed.
const blobCache = new Map<string, string>()
let prefetchRun = 0

function revokeOwned() {
  if (ownedUrl) { try { URL.revokeObjectURL(ownedUrl) } catch { /* ignore */ } ownedUrl = '' }
}
function clearCache() {
  prefetchRun++ // cancel any running prefetch loop
  for (const url of blobCache.values()) { try { URL.revokeObjectURL(url) } catch { /* ignore */ } }
  blobCache.clear()
}
function applyRate() { if (el) el.playbackRate = rate.value }

async function fetchBlobUrl(item: QueueItem): Promise<string> {
  const cached = blobCache.get(item.key)
  if (cached) return cached
  const { data } = await api.get(item.endpoint, { responseType: 'blob' })
  const url = URL.createObjectURL(data as Blob)
  blobCache.set(item.key, url)
  return url
}

// preload every queued track (sequentially) so later advances — including while
// the screen is locked — don't need the network.
async function prefetchAll() {
  const run = ++prefetchRun
  for (const item of queue.value) {
    if (run !== prefetchRun) return // a newer queue/prefetch superseded this one
    if (!blobCache.has(item.key)) {
      try { await fetchBlobUrl(item) } catch { /* ignore: fetched on demand later */ }
    }
  }
}

function setMediaSession(title: string) {
  const ms = (navigator as unknown as { mediaSession?: MediaSession }).mediaSession
  if (!ms) return
  try {
    const MM = (window as unknown as { MediaMetadata?: typeof MediaMetadata }).MediaMetadata
    if (MM) ms.metadata = new MM({ title, artist: 'E-Learning' })
    ms.setActionHandler('play', () => { el?.play().catch(() => {}) })
    ms.setActionHandler('pause', () => { el?.pause() })
    const multi = queue.value.length > 1
    ms.setActionHandler('previoustrack', multi ? () => { if (queueIndex.value > 0) void playQueueAt(queueIndex.value - 1) } : null)
    ms.setActionHandler('nexttrack', multi ? () => advance() : null)
  } catch { /* some actions unsupported on this browser */ }
}

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

// play the queue track at index i, using the preloaded blob when available so
// no network is needed (critical for iOS background advance).
async function playQueueAt(i: number) {
  if (!el || i < 0 || i >= queue.value.length) return
  const item = queue.value[i]
  queueIndex.value = i
  curKey.value = item.key
  curTitle.value = item.name
  const cached = blobCache.get(item.key)
  if (!cached) loading.value = true
  try {
    const url = cached ?? await fetchBlobUrl(item) // no await when already preloaded
    revokeOwned() // leaving single-play mode; queue blobs are owned by the cache
    curSrc.value = url
    currentTime.value = 0
    duration.value = 0
    el.src = url
    applyRate()
    setMediaSession(item.name)
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
    clearCache()
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
      setMediaSession(title)
    }
    el.play().catch(() => { /* autoplay blocked / interrupted */ })
  }
  // play a playlist of authed tracks, starting at startIndex. Preloads the rest
  // so it keeps advancing with the screen locked.
  function playQueue(items: QueueItem[], startIndex = 0) {
    if (!items.length) return
    revokeOwned()
    clearCache()
    queue.value = items.slice()
    void playQueueAt(Math.max(0, Math.min(startIndex, items.length - 1)))
    void prefetchAll()
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
    clearCache()
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
