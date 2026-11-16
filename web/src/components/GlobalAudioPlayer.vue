<template>
  <div v-if="curSrc" class="gap-bar shadow-lg">
    <div class="d-flex align-items-center gap-2 px-3 py-2">
      <button v-if="hasQueue" class="btn btn-sm btn-outline-secondary gap-skip" :disabled="queueIndex <= 0" title="Bài trước" @click="prevTrack">
        <i class="bi bi-skip-start-fill"></i>
      </button>
      <button class="btn btn-sm btn-primary rounded-circle gap-play" :disabled="loading" @click="toggle">
        <span v-if="loading" class="spinner-border spinner-border-sm"></span>
        <i v-else :class="playing ? 'bi bi-pause-fill' : 'bi bi-play-fill'"></i>
      </button>
      <button v-if="hasQueue" class="btn btn-sm btn-outline-secondary gap-skip" :disabled="queueIndex >= queue.length - 1" title="Bài tiếp" @click="nextTrack">
        <i class="bi bi-skip-end-fill"></i>
      </button>
      <div class="flex-grow-1" style="min-width: 0">
        <div class="small fw-semibold text-truncate">
          {{ curTitle || 'Audio' }}
          <span v-if="hasQueue" class="text-muted fw-normal">· {{ queueIndex + 1 }}/{{ queue.length }}</span>
        </div>
        <div class="d-flex align-items-center gap-2">
          <span class="gap-time">{{ fmt(currentTime) }}</span>
          <input type="range" class="form-range gap-range" min="0" :max="duration || 0" step="0.1"
            :value="currentTime" @input="onSeek" />
          <span class="gap-time">{{ fmt(duration) }}</span>
        </div>
      </div>
      <button class="btn btn-sm btn-outline-secondary gap-rate" :title="'Tốc độ phát (' + rate + 'x)'" @click="cycleRate">{{ rate }}×</button>
      <button class="btn btn-sm btn-outline-secondary" title="Đóng" @click="stop"><i class="bi bi-x-lg"></i></button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAudioPlayer } from '../composables/useAudioPlayer'
const { curSrc, curTitle, playing, currentTime, duration, loading, rate, queue, queueIndex, toggle, stop, seek, setRate, nextTrack, prevTrack } = useAudioPlayer()

const RATES = [1, 1.25, 1.5, 1.75, 2]
const hasQueue = computed(() => queue.value.length > 1)
function onSeek(e: Event) { seek(Number((e.target as HTMLInputElement).value)) }
function cycleRate() {
  const i = RATES.indexOf(rate.value)
  setRate(RATES[(i + 1) % RATES.length])
}
function fmt(s: number): string {
  if (!s || !isFinite(s)) return '0:00'
  const m = Math.floor(s / 60)
  return `${m}:${String(Math.floor(s % 60)).padStart(2, '0')}`
}
</script>

<style scoped>
.gap-bar {
  position: fixed;
  left: 50%;
  transform: translateX(-50%);
  bottom: 1rem;
  z-index: 1055;
  width: min(680px, calc(100vw - 2rem));
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
}
.gap-play { width: 2.4rem; height: 2.4rem; display: flex; align-items: center; justify-content: center; }
.gap-skip { width: 2rem; height: 2rem; display: flex; align-items: center; justify-content: center; padding: 0; }
.gap-rate { min-width: 2.9rem; font-variant-numeric: tabular-nums; font-weight: 600; }
.gap-range { flex: 1; }
.gap-time { font-size: 0.72rem; color: #64748b; white-space: nowrap; font-variant-numeric: tabular-nums; }
</style>
