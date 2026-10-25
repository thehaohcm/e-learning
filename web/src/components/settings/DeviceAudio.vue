<template>
  <div class="card border-0 shadow-sm mb-3">
    <div class="card-body">
      <h6 class="fw-bold"><i class="bi bi-volume-up"></i> Words read aloud (device)</h6>
      <p class="small text-muted mb-2">
        The device reads English words aloud. Each word is made once and kept on the server
        (about 30 KB). Words not played for {{ st?.keep_days ?? 60 }} days are deleted automatically
        and made again when they come back.
      </p>
      <div v-if="st" class="small mb-2">
        <span v-if="!st.enabled" class="text-danger">Text-to-speech is off (TTS_URL is empty).</span>
        <template v-else>
          <strong>{{ st.words }}</strong> words · <strong>{{ size(st.bytes) }}</strong>
          <span v-if="st.oldest_used_at" class="text-muted">
            · oldest played {{ new Date(st.oldest_used_at).toLocaleDateString() }}
          </span>
          · voice {{ st.voice }}
        </template>
      </div>
      <button class="btn btn-sm btn-outline-primary me-2" :disabled="testing" @click="test">
        <i class="bi bi-play-circle"></i> {{ testing ? 'Testing…' : 'Test voice' }}
      </button>
      <button class="btn btn-sm btn-outline-danger" :disabled="busy || !st?.words" @click="purge">
        <i class="bi bi-trash"></i> Delete all audio
      </button>
      <span v-if="msg" class="ms-2 small text-muted">{{ msg }}</span>
      <div v-if="testMsg" class="small mt-2" :class="testOk ? 'text-success' : 'text-danger'">{{ testMsg }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, errorMessage } from '../../api'

interface AudioStats { words: number; bytes: number; oldest_used_at: string | null; keep_days: number; enabled: boolean; voice: string }

const st = ref<AudioStats | null>(null)
const busy = ref(false)
const msg = ref('')
const testing = ref(false)
const testMsg = ref('')
const testOk = ref(true)

// makes "hello" on the server (like the device would get it) and plays it here
async function test() {
  testing.value = true
  testMsg.value = ''
  try {
    const { data } = await api.get('/admin/audio/test', { params: { text: 'hello' } })
    testOk.value = data.ok
    if (data.ok) {
      testMsg.value = 'The voice works: you should hear “hello”.'
      new Audio(data.url).play().catch(() => {})
      await load()
    } else {
      testMsg.value = 'Not working: ' + data.error
    }
  } catch (e) {
    testOk.value = false
    testMsg.value = errorMessage(e)
  } finally {
    testing.value = false
  }
}

function size(b: number) {
  return b >= 1 << 20 ? (b / (1 << 20)).toFixed(1) + ' MB' : Math.ceil(b / 1024) + ' KB'
}

async function load() {
  const { data } = await api.get('/admin/audio')
  st.value = data
}

async function purge() {
  if (!confirm('Delete all stored audio? Words are made again when they are studied.')) return
  busy.value = true
  msg.value = ''
  try {
    const { data } = await api.delete('/admin/audio')
    msg.value = `Deleted ${data.deleted} words`
    await load()
  } catch (e) {
    msg.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

onMounted(() => load().catch(() => {}))
</script>
