import { ref } from 'vue'

// A single, app-wide audio player. The HTMLAudioElement lives at module scope
// (not inside a component), so it keeps playing while the user navigates
// between pages, generates questions, etc. A small global player bar (mounted
// once in App.vue) binds to this reactive state for play/pause/seek/close.

const el: HTMLAudioElement | null = typeof Audio !== 'undefined' ? new Audio() : null

export const curSrc = ref('')
export const curTitle = ref('')
export const playing = ref(false)
export const currentTime = ref(0)
export const duration = ref(0)
export const loading = ref(false)

// the object URL this player currently owns (created from an authed blob); it
// is revoked when a new track starts or the player is closed.
let ownedUrl = ''

if (el) {
  el.addEventListener('timeupdate', () => { currentTime.value = el.currentTime })
  el.addEventListener('durationchange', () => { duration.value = isFinite(el.duration) ? el.duration : 0 })
  el.addEventListener('loadedmetadata', () => { duration.value = isFinite(el.duration) ? el.duration : 0 })
  el.addEventListener('play', () => { playing.value = true })
  el.addEventListener('pause', () => { playing.value = false })
  el.addEventListener('ended', () => { playing.value = false })
  el.addEventListener('waiting', () => { loading.value = true })
  el.addEventListener('playing', () => { loading.value = false })
}

function revokeOwned() {
  if (ownedUrl) { try { URL.revokeObjectURL(ownedUrl) } catch { /* ignore */ } ownedUrl = '' }
}

export function useAudioPlayer() {
  // play a direct URL (static file). owned=true hands a blob URL's lifecycle to
  // the player (it will revoke it when replaced or closed) so playback survives
  // the component that created it unmounting.
  function play(url: string, title: string, owned = false) {
    if (!el || !url) return
    if (curSrc.value !== url) {
      if (ownedUrl && ownedUrl !== url) revokeOwned()
      ownedUrl = owned ? url : ''
      curSrc.value = url
      curTitle.value = title
      currentTime.value = 0
      duration.value = 0
      el.src = url
    }
    el.play().catch(() => { /* autoplay blocked / interrupted */ })
  }
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
    currentTime.value = 0
    duration.value = 0
    playing.value = false
  }
  function seek(t: number) { if (el) el.currentTime = t }
  function isCurrent(url: string) { return curSrc.value === url }
  return { curSrc, curTitle, playing, currentTime, duration, loading, play, toggle, stop, seek, isCurrent }
}
