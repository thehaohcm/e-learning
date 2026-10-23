<template>
  <div class="row g-4">
    <div class="col-lg-6">
      <h2 class="fw-bold mb-3"><i class="bi bi-cpu"></i> Study device</h2>

      <div class="card border-0 shadow-sm mb-3">
        <div class="card-body">
          <h6 class="fw-bold">Link a device</h6>
          <p class="small text-muted">Turn on the device and connect it to Wi-Fi. It shows a 6-digit code: type it here.</p>
          <form class="row g-2" @submit.prevent="pair">
            <div class="col-5"><input v-model="code" class="form-control form-control-lg text-center" inputmode="numeric" maxlength="6" placeholder="123456" pattern="\d{6}" required /></div>
            <div class="col-7"><input v-model="name" class="form-control form-control-lg" placeholder="Name (optional)" maxlength="100" /></div>
            <div class="col-12"><button class="btn btn-primary" :disabled="busy">Link device</button></div>
          </form>
          <div v-if="pairMsg" class="alert mt-3 mb-0 py-2" :class="pairOk ? 'alert-success' : 'alert-danger'">{{ pairMsg }}</div>
        </div>
      </div>

      <div v-for="d in devices" :key="d.id" class="card border-0 shadow-sm mb-2">
        <div class="card-body d-flex justify-content-between align-items-center">
          <div>
            <div class="fw-semibold">{{ d.name }}</div>
            <small class="text-muted">
              Firmware {{ d.firmware || '?' }} ·
              {{ d.last_seen_at ? 'last seen ' + new Date(d.last_seen_at).toLocaleString() : 'never synced' }}
            </small>
          </div>
          <button class="btn btn-sm btn-outline-danger" @click="unpair(d.id)"><i class="bi bi-x-lg"></i> Remove</button>
        </div>
      </div>
      <p v-if="!devices.length" class="text-muted small">No device linked yet.</p>
    </div>

    <div class="col-lg-6">
      <h2 class="fw-bold mb-3"><i class="bi bi-sliders"></i> Study settings</h2>
      <AppNotifications v-if="settings" v-model:times="settings.push_times" :timezone="timezone" />
      <ExtensionKeys />
      <div v-if="settings" class="card border-0 shadow-sm">
        <div class="card-body">
          <div class="mb-3">
            <label class="form-label">New cards per day</label>
            <input v-model.number="settings.daily_new" type="number" min="0" max="200" class="form-control" />
          </div>
          <div class="mb-3">
            <label class="form-label">Decks that feed new cards automatically</label>
            <div v-for="s in subjects" :key="s.id" class="mb-1">
              <small class="text-muted">{{ s.name }}</small>
              <div v-for="d in s.decks" :key="d.id" class="form-check">
                <input :id="`deck-${d.id}`" v-model="settings.active_deck_ids" :value="d.id" type="checkbox" class="form-check-input" />
                <label :for="`deck-${d.id}`" class="form-check-label">{{ d.name }}</label>
              </div>
            </div>
            <div class="form-text">None selected = Core Vocabulary.</div>
          </div>
          <div class="mb-3">
            <label class="form-label" for="devmode">Device pop-up shows</label>
            <select id="devmode" v-model="settings.device_mode" class="form-select">
              <option v-for="m in DEVICE_MODES" :key="m.value" :value="m.value">{{ m.label }}</option>
            </select>
            <div class="form-text">
              {{ DEVICE_MODES.find((m) => m.value === settings!.device_mode)?.help }}
              New words are always shown as flashcards first.
            </div>
          </div>
          <div class="row g-2 mb-3">
            <div class="col-4">
              <label class="form-label">Device pop-up every (min)</label>
              <input v-model.number="settings.reminder_every_min" type="number" min="5" max="1440" class="form-control" />
            </div>
            <div class="col-4">
              <label class="form-label">From</label>
              <input v-model="settings.active_from" type="time" class="form-control" />
            </div>
            <div class="col-4">
              <label class="form-label">To</label>
              <input v-model="settings.active_to" type="time" class="form-control" />
            </div>
          </div>
          <div class="mb-3">
            <label class="form-label">Time zone</label>
            <input v-model="timezone" class="form-control" placeholder="Asia/Ho_Chi_Minh" />
          </div>
          <button class="btn btn-primary" :disabled="saving" @click="save">Save</button>
          <span v-if="saveMsg" class="ms-2 small" :class="saveOk ? 'text-success' : 'text-danger'">{{ saveMsg }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, errorMessage } from '../api'
import type { Subject } from '../types'
import AppNotifications from '../components/settings/AppNotifications.vue'
import ExtensionKeys from '../components/settings/ExtensionKeys.vue'

interface Device { id: number; name: string; firmware: string; last_seen_at: string | null }
interface Settings { daily_new: number; active_deck_ids: number[]; reminder_every_min: number; active_from: string; active_to: string; tts_speed: number; device_mode: string; push_times: string[] }

const DEVICE_MODES = [
  { value: 'flashcard', label: 'Flashcard', help: 'The word, tap to see the meaning, then rate how well you knew it.' },
  { value: 'quiz', label: 'Quiz (4 options)', help: 'The word with 4 meanings: tap the right one.' },
  { value: 'listening', label: 'Listening', help: 'The device plays the word/sentence, you pick the meaning. Needs audio files on the server (coming soon): until then these cards are shown as a quiz.' },
  { value: 'cloze', label: 'Fill in the blank (pick the word)', help: 'An example sentence with the word missing: tap the right word out of 4, or the right form of the word (TOEIC Part 5) when it was prepared on the web. No typing on the device. Words without a suitable sentence are shown as a quiz.' },
  { value: 'mix', label: 'Mix (flashcard · quiz · fill in the blank · listening)', help: 'Rotates the modes.' },
]

const devices = ref<Device[]>([])
const subjects = ref<Subject[]>([])
const settings = ref<Settings | null>(null)
const timezone = ref('')
const code = ref('')
const name = ref('')
const busy = ref(false)
const pairMsg = ref('')
const pairOk = ref(true)
const saving = ref(false)
const saveMsg = ref('')
const saveOk = ref(true)

async function loadDevices() {
  const { data } = await api.get('/devices')
  devices.value = data
}

async function pair() {
  busy.value = true
  pairMsg.value = ''
  try {
    const { data } = await api.post('/devices/pair', { code: code.value.trim(), name: name.value })
    pairOk.value = true
    pairMsg.value = `“${data.name}” is linked. It will start syncing within a minute.`
    code.value = ''
    name.value = ''
    await loadDevices()
  } catch (e) {
    pairOk.value = false
    pairMsg.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

async function unpair(id: number) {
  if (!confirm('Remove this device? It will need to be linked again.')) return
  await api.delete(`/devices/${id}`)
  await loadDevices()
}

async function save() {
  saving.value = true
  saveMsg.value = ''
  try {
    const { data } = await api.put('/settings', { settings: settings.value, timezone: timezone.value })
    settings.value = data.settings
    timezone.value = data.timezone
    saveOk.value = true
    saveMsg.value = 'Saved'
  } catch (e) {
    saveOk.value = false
    saveMsg.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  const [s, subs] = await Promise.all([api.get('/settings'), api.get('/subjects'), loadDevices()])
  settings.value = { ...s.data.settings, push_times: s.data.settings.push_times ?? [] }
  timezone.value = s.data.timezone
  subjects.value = subs.data
})
</script>
