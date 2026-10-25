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
        <div class="card-body">
          <div class="d-flex justify-content-between align-items-center gap-2">
            <div>
              <div class="fw-semibold">{{ d.name }}</div>
              <small class="text-muted">
                Firmware {{ d.firmware || '?' }} ·
                {{ d.last_seen_at ? 'last seen ' + new Date(d.last_seen_at).toLocaleString() : 'never synced' }}
              </small>
              <div class="small mt-1">
                <i class="bi bi-funnel"></i> {{ topicSummary(d) }}
                <span class="text-muted">· {{ popupSummary(d) }}</span>
              </div>
            </div>
            <div class="d-flex gap-2 flex-shrink-0">
              <button class="btn btn-sm btn-outline-secondary" @click="toggleEdit(d)">
                <i class="bi bi-gear"></i> Settings
              </button>
              <button class="btn btn-sm btn-outline-danger" @click="unpair(d.id)"><i class="bi bi-x-lg"></i> Remove</button>
            </div>
          </div>

          <div v-if="editing[d.id]" class="border-top mt-3 pt-3">
            <label class="form-label fw-semibold">Topics on this device</label>
            <div class="d-flex flex-wrap gap-3 mb-1">
              <div v-for="s in subjects" :key="s.id" class="form-check">
                <input :id="`dev-${d.id}-sub-${s.id}`" v-model="editing[d.id].subject_ids" :value="s.id" type="checkbox" class="form-check-input" />
                <label :for="`dev-${d.id}-sub-${s.id}`" class="form-check-label">{{ s.name }}</label>
              </div>
            </div>
            <div class="form-text mb-3">
              None ticked = all topics. Cards of other topics still wait for you on the website.
            </div>

            <label class="form-label fw-semibold" :for="`dev-${d.id}-mode`">Pop-up shows</label>
            <select :id="`dev-${d.id}-mode`" v-model="editing[d.id].device_mode" class="form-select">
              <option v-for="m in DEVICE_MODES" :key="m.value" :value="m.value">{{ m.label }}</option>
            </select>
            <div class="form-text mb-3">
              {{ DEVICE_MODES.find((m) => m.value === editing[d.id].device_mode)?.help }}
              New words are always shown as flashcards first.
            </div>

            <label class="form-label fw-semibold" :for="`dev-${d.id}-cards`">Cards per pop-up</label>
            <select :id="`dev-${d.id}-cards`" v-model.number="editing[d.id].popup_cards" class="form-select">
              <option v-for="n in POPUP_CARDS" :key="n" :value="n">{{ n }} cards</option>
            </select>
            <div class="form-text mb-3">After the last one: “Keep studying” or “Done”.</div>

            <div class="row g-2 mb-1">
              <div class="col-4">
                <label class="form-label">Pop-up every (minutes)</label>
                <input v-model.number="editing[d.id].reminder_every_min" type="number" min="5" max="1440" class="form-control" />
              </div>
              <div class="col-4">
                <label class="form-label">From</label>
                <input v-model="editing[d.id].active_from" type="time" class="form-control" />
              </div>
              <div class="col-4">
                <label class="form-label">To</label>
                <input v-model="editing[d.id].active_to" type="time" class="form-control" />
              </div>
            </div>
            <div class="form-text mb-3">
              Pop-ups only come inside these hours, when the device was not used for that long and has cards to study.
              The device gets the changes within 3 minutes (or press Sync on it).
            </div>

            <button class="btn btn-sm btn-primary" :disabled="devSaving === d.id" @click="saveDevice(d)">Save</button>
            <span v-if="devMsg[d.id]" class="ms-2 small" :class="devMsg[d.id]!.ok ? 'text-success' : 'text-danger'">{{ devMsg[d.id]!.text }}</span>
          </div>
        </div>
      </div>
      <p v-if="!devices.length" class="text-muted small">No device linked yet.</p>
      <DeviceAudio v-if="auth.isAdmin" class="mt-3" />
    </div>

    <div class="col-lg-6">
      <h2 class="fw-bold mb-3"><i class="bi bi-sliders"></i> Study settings</h2>
      <p class="small text-muted">For your whole account (website and devices). Pop-ups and topics of a study device are in its <i class="bi bi-gear"></i> Settings.</p>
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
import { onMounted, reactive, ref } from 'vue'
import { api, errorMessage } from '../api'
import type { Subject } from '../types'
import AppNotifications from '../components/settings/AppNotifications.vue'
import ExtensionKeys from '../components/settings/ExtensionKeys.vue'
import DeviceAudio from '../components/settings/DeviceAudio.vue'
import { auth } from '../auth'

interface DeviceSettings { subject_ids: number[]; device_mode: string; reminder_every_min: number | ''; active_from: string; active_to: string; popup_cards: number }
interface Device { id: number; name: string; firmware: string; last_seen_at: string | null; settings: Partial<DeviceSettings> | null }
interface Settings { daily_new: number; active_deck_ids: number[]; reminder_every_min: number; active_from: string; active_to: string; tts_speed: number; device_mode: string; push_times: string[]; popup_cards: number }

const POPUP_CARDS = [5, 10, 20]

const DEVICE_MODES = [
  { value: 'flashcard', label: 'Flashcard', help: 'The word, tap to see the meaning, then rate how well you knew it.' },
  { value: 'quiz', label: 'Quiz (4 options)', help: 'The word with 4 meanings: tap the right one.' },
  { value: 'listening', label: 'Listening', help: 'The device says the English word, you pick the meaning. Words whose audio is not ready yet are shown as a quiz.' },
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

const editing = reactive<Record<number, DeviceSettings>>({})
const devMsg = reactive<Record<number, { ok: boolean; text: string } | undefined>>({})
const devSaving = ref<number | null>(null)

// The values the device uses: its own (the defaults for a device saved before
// these settings moved here come from the account).
function devSettings(d: Device): DeviceSettings {
  const x = d.settings ?? {}
  const acc = settings.value
  return {
    subject_ids: [...(x.subject_ids ?? [])],
    device_mode: x.device_mode || acc?.device_mode || 'mix',
    popup_cards: x.popup_cards || acc?.popup_cards || 10,
    reminder_every_min: x.reminder_every_min || acc?.reminder_every_min || '',
    active_from: x.active_from || acc?.active_from || '',
    active_to: x.active_to || acc?.active_to || '',
  }
}

// The device's own values (all of them: what the form shows is what it uses).
function overrides(e: DeviceSettings) {
  return {
    subject_ids: e.subject_ids,
    device_mode: e.device_mode,
    reminder_every_min: Number(e.reminder_every_min) || 0,
    active_from: e.active_from,
    active_to: e.active_to,
    popup_cards: e.popup_cards,
  }
}

function toggleEdit(d: Device) {
  if (editing[d.id]) delete editing[d.id]
  else editing[d.id] = devSettings(d)
  devMsg[d.id] = undefined
}

function topicSummary(d: Device) {
  const ids = d.settings?.subject_ids ?? []
  if (!ids.length) return 'All topics'
  return subjects.value.filter((s) => ids.includes(s.id)).map((s) => s.name).join(', ') || 'All topics'
}

function popupSummary(d: Device) {
  const e = devSettings(d)
  return `pop-up ${e.popup_cards} cards every ${e.reminder_every_min} min, ${e.active_from}–${e.active_to}`
}

async function putDevice(d: Device, body: DeviceSettings) {
  devSaving.value = d.id
  devMsg[d.id] = undefined
  try {
    const { data } = await api.put(`/devices/${d.id}/settings`, overrides(body))
    d.settings = data
    editing[d.id] = devSettings(d)
    devMsg[d.id] = { ok: true, text: 'Saved' }
  } catch (e) {
    devMsg[d.id] = { ok: false, text: errorMessage(e) }
  } finally {
    devSaving.value = null
  }
}

function saveDevice(d: Device) {
  const e = editing[d.id]
  if (e) putDevice(d, e)
}


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
  settings.value = { ...s.data.settings, push_times: s.data.settings.push_times ?? [], popup_cards: s.data.settings.popup_cards || 10 }
  timezone.value = s.data.timezone
  subjects.value = subs.data
})
</script>
