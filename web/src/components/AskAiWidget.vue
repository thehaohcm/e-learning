<template>
  <div class="ask-ai">
    <!-- panel -->
    <div v-if="open" class="panel card shadow-lg border-0">
      <div class="card-header bg-white d-flex align-items-center gap-2 py-2">
        <i class="bi bi-chat-dots text-primary"></i>
        <span class="fw-bold">Ask AI</span>
        <small class="text-muted text-truncate">· {{ topicName }}</small>
        <button type="button" class="btn-close ms-auto" aria-label="Close" @click="open = false"></button>
      </div>
      <div class="card-body p-2">
        <CourseChat />
      </div>
    </div>

    <!-- floating button -->
    <button type="button" class="btn btn-primary rounded-circle shadow fab" :title="open ? 'Close' : 'Ask AI about this course'"
      :aria-expanded="open" @click="open = !open">
      <i :class="open ? 'bi bi-x-lg' : 'bi bi-stars'"></i>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { shortName, subject } from '../composables/useSubject'
import CourseChat from './CourseChat.vue'

const open = ref(false)
const topicName = computed(() => (subject.value ? shortName(subject.value.slug, subject.value.name) : 'this course'))

function storageGet(k: string) {
  try { return localStorage.getItem(k) } catch { return null }
}
open.value = storageGet('elearn.askai.open') === 'true'
watch(open, (v) => { try { localStorage.setItem('elearn.askai.open', String(v)) } catch { /* ignore */ } })
</script>

<style scoped>
.fab {
  position: fixed;
  left: 20px;
  bottom: 20px;
  width: 54px;
  height: 54px;
  z-index: 1050;
  font-size: 1.35rem;
}
.panel {
  position: fixed;
  left: 20px;
  bottom: 86px;
  width: 380px;
  max-height: min(75vh, 620px);
  z-index: 1050;
  display: flex;
  flex-direction: column;
  border-radius: 16px !important;
  overflow: hidden;
}
.panel .card-body { overflow-y: auto; }
@media (max-width: 576px) {
  .panel { left: 8px; right: 8px; width: auto; bottom: 80px; }
  .fab { left: 14px; bottom: 14px; }
}
</style>
