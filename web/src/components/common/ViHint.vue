<template>
  <!-- A small "?" that shows the Vietnamese meaning on hover (tap on phones).
       Clicks stop here, so it can sit inside an answer button without answering. -->
  <span class="vi-hint" role="button" tabindex="0" :aria-label="`Vietnamese: ${text}`" :class="{ open }"
    @mouseenter="hover = true" @mouseleave="hover = false"
    @click.stop.prevent="pinned = !pinned" @keydown.enter.stop.prevent="pinned = !pinned" @keydown.space.stop.prevent="pinned = !pinned"
    @blur="pinned = false">
    ?
    <span v-if="open" class="vi-tip" role="tooltip">{{ text }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

defineProps<{ text: string }>()
const hover = ref(false)
const pinned = ref(false)
const open = computed(() => hover.value || pinned.value)
</script>

<style scoped>
.vi-hint {
  position: relative;
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.6rem;
  height: 1.6rem;
  margin-left: 0.5rem;
  border-radius: 50%;
  border: 1px solid var(--bs-border-color);
  background: var(--bs-body-bg);
  color: var(--bs-secondary-color);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: help;
  user-select: none;
}
.vi-hint.open,
.vi-hint:focus-visible {
  border-color: var(--bs-warning);
  color: var(--bs-body-color);
  outline: none;
}
.vi-tip {
  position: absolute;
  right: 0;
  bottom: calc(100% + 6px);
  z-index: 20;
  width: max-content;
  max-width: min(280px, 70vw);
  padding: 0.35rem 0.6rem;
  border-radius: 0.4rem;
  background: #212529;
  color: #fff;
  font-size: 0.85rem;
  font-weight: 400;
  line-height: 1.35;
  text-align: left;
  white-space: normal;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}
</style>
