<template>
  <div class="card border-0 shadow-sm mb-3">
    <div class="card-body">
      <h6 class="fw-bold"><i class="bi bi-puzzle"></i> Browser extension</h6>
      <p class="small text-muted mb-2">
        Save words from any website: select a word → right-click → <b>Add to E-Learn notes</b> (or <kbd>Alt</kbd>+<kbd>Shift</kbd>+<kbd>N</kbd>).
        The word is saved with its sentence in <b>My notes</b>.
      </p>
      <details class="small mb-2">
        <summary>How to install (Chrome, Edge, Brave, Cốc Cốc)</summary>
        <ol class="mt-2 mb-0 ps-3">
          <li>Open <code>chrome://extensions</code> and turn on <b>Developer mode</b> (top right).</li>
          <li><b>Load unpacked</b> → choose the <code>extension</code> folder of this project.</li>
          <li>Create a key below and paste it in the extension options (they open by themselves after installing).</li>
        </ol>
      </details>

      <div v-if="newKey" class="alert alert-success py-2 small">
        <div class="fw-semibold mb-1">Your new key (shown only once):</div>
        <div class="input-group input-group-sm">
          <input class="form-control font-monospace" :value="newKey" readonly @focus="($event.target as HTMLInputElement).select()" />
          <button class="btn btn-outline-success" @click="copy">{{ copied ? 'Copied' : 'Copy' }}</button>
        </div>
        <div class="mt-1">Paste it in the extension: right-click the extension icon → <b>Options</b>.</div>
      </div>

      <ul v-if="keys.length" class="list-unstyled small mb-2">
        <li v-for="k in keys" :key="k.id" class="d-flex align-items-center gap-2 py-1 border-bottom">
          <i class="bi bi-key text-muted"></i>
          <span class="fw-semibold">{{ k.name }}</span>
          <code class="text-muted">{{ k.prefix }}…</code>
          <span class="text-muted">{{ k.last_used_at ? 'used ' + new Date(k.last_used_at).toLocaleDateString() : 'never used' }}</span>
          <button class="btn btn-link btn-sm text-danger ms-auto p-0" @click="remove(k.id)">Remove</button>
        </li>
      </ul>

      <form class="d-flex gap-2" @submit.prevent="create">
        <input v-model="name" class="form-control form-control-sm" maxlength="60" placeholder="Name, e.g. Chrome at work" />
        <button class="btn btn-sm btn-primary flex-shrink-0" :disabled="busy">Create key</button>
      </form>
      <div v-if="error" class="small text-danger mt-1">{{ error }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, errorMessage } from '../../api'

interface Key { id: number; name: string; prefix: string; created_at: string; last_used_at: string | null }

const keys = ref<Key[]>([])
const name = ref('')
const newKey = ref('')
const copied = ref(false)
const busy = ref(false)
const error = ref('')

async function load() {
  try {
    keys.value = (await api.get('/extension/keys')).data ?? []
  } catch (e) {
    error.value = errorMessage(e)
  }
}

async function create() {
  busy.value = true
  error.value = ''
  try {
    const { data } = await api.post('/extension/keys', { name: name.value })
    newKey.value = data.key
    copied.value = false
    name.value = ''
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

async function remove(id: number) {
  if (!confirm('Remove this key? The extension using it stops working.')) return
  await api.delete(`/extension/keys/${id}`)
  await load()
}

async function copy() {
  try {
    await navigator.clipboard.writeText(newKey.value)
    copied.value = true
  } catch {
    copied.value = false
  }
}

onMounted(load)
</script>
