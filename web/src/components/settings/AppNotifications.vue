<template>
  <div class="card border-0 shadow-sm mb-3">
    <div class="card-body">
      <h6 class="fw-bold"><i class="bi bi-phone"></i> Phone app &amp; study reminders</h6>

      <!-- install -->
      <div class="d-flex flex-wrap align-items-center gap-2 mb-3">
        <template v-if="isInstalled">
          <span class="text-success small"><i class="bi bi-check-circle-fill"></i> Running as an installed app.</span>
        </template>
        <template v-else-if="canInstall">
          <button class="btn btn-outline-primary btn-sm" @click="install"><i class="bi bi-download"></i> Install app</button>
          <span class="small text-muted">Adds an icon to your home screen / desktop, opens full screen.</span>
        </template>
        <div v-else-if="isIOS" class="small text-muted">
          <i class="bi bi-apple"></i> On iPhone/iPad: open this site in <strong>Safari</strong>, tap
          <i class="bi bi-box-arrow-up"></i> <strong>Share</strong> → <strong>Add to Home Screen</strong>, then open the app from its icon.
        </div>
        <div v-else class="small text-muted">
          To install: browser menu (⋮) → <strong>Install app</strong> or <strong>Add to Home screen</strong>.
        </div>
      </div>

      <!-- notifications on this device -->
      <div v-if="!pushSupported" class="alert alert-light border small py-2 mb-3">
        <template v-if="isIOS && !isInstalled">
          iPhone/iPad can show reminders only from the installed app (iOS 16.4 or newer): add it to the Home Screen first (above), open it from the icon, then come back here.
        </template>
        <template v-else>This browser cannot show notifications. Try Chrome, Edge, Firefox or Safari.</template>
      </div>
      <template v-else>
        <div class="form-check form-switch mb-1">
          <input id="push-on" class="form-check-input" type="checkbox" role="switch" :checked="enabled" :disabled="busy || checking" @change="toggle" />
          <label class="form-check-label" for="push-on">Study reminders on this device</label>
          <span v-if="busy || checking" class="spinner-border spinner-border-sm ms-2 text-muted"></span>
        </div>
        <div v-if="permission === 'denied' && !enabled" class="small text-danger mb-2">
          Notifications are blocked for this site: allow them in the browser’s site settings (🔒 next to the address), then switch this on.
        </div>
        <div v-if="enabled" class="mb-2">
          <button class="btn btn-link btn-sm p-0" :disabled="busy" @click="test"><i class="bi bi-send"></i> Send a test now</button>
        </div>
      </template>

      <!-- reminder times (per account, every device that has reminders on) -->
      <div class="mt-2">
        <label class="form-label small fw-semibold mb-1">Reminder times</label>
        <div class="d-flex flex-wrap align-items-center gap-2">
          <span v-for="t in times" :key="t" class="badge rounded-pill text-bg-light border fs-6 fw-normal">
            {{ t }}
            <button class="btn btn-link btn-sm p-0 ms-1 text-muted align-baseline" :title="`Remove ${t}`" :disabled="busy" @click="removeTime(t)"><i class="bi bi-x"></i></button>
          </span>
          <span v-if="!times.length" class="small text-muted">none</span>
          <form v-if="times.length < MAX_TIMES" class="d-flex gap-1" @submit.prevent="addTime">
            <input v-model="newTime" type="time" class="form-control form-control-sm" style="width: 7.5rem" required />
            <button class="btn btn-sm btn-outline-secondary" :disabled="busy || !newTime">Add</button>
          </form>
        </div>
        <div class="form-text">
          A reminder says how many cards are due and new words are left. It is sent only when there is something to study
          (time zone: {{ timezone || 'your account time zone' }}). The times apply to every phone/browser with reminders on.
        </div>
      </div>

      <div v-if="msg" class="alert mt-2 mb-0 py-2 small" :class="ok ? 'alert-success' : 'alert-danger'">{{ msg }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, errorMessage } from '../../api'
import {
  canInstall, disablePush, enablePush, installApp, isInstalled, isIOS, notificationPermission, pushEnabledHere, pushSupported, testPush,
} from '../../composables/usePwa'

const props = defineProps<{ times: string[]; timezone: string }>()
const emit = defineEmits<{ (e: 'update:times', v: string[]): void }>()

const MAX_TIMES = 6
const DEFAULT_TIME = '20:00'

const enabled = ref(false)
const checking = ref(true)
const busy = ref(false)
const permission = ref(notificationPermission())
const newTime = ref('')
const msg = ref('')
const ok = ref(true)

function say(text: string, good = true) {
  msg.value = text
  ok.value = good
}

async function install() {
  if (await installApp()) say('Installed. You can open it from the home screen / app list.')
}

/** Saves only push_times (a partial settings update). */
async function saveTimes(list: string[]) {
  const { data } = await api.put('/settings', { settings: { push_times: list } })
  emit('update:times', data.settings.push_times ?? [])
}

async function toggle(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  busy.value = true
  msg.value = ''
  try {
    if (on) {
      await enablePush()
      enabled.value = true
      if (!props.times.length) await saveTimes([DEFAULT_TIME])
      say(`Reminders are on for this device at ${(props.times.length ? props.times : [DEFAULT_TIME]).join(', ')}.`)
    } else {
      await disablePush()
      enabled.value = false
      say('Reminders are off for this device.')
    }
  } catch (err) {
    enabled.value = false
    ;(e.target as HTMLInputElement).checked = false
    say(err instanceof Error && !('response' in err) ? err.message : errorMessage(err), false)
  } finally {
    permission.value = notificationPermission()
    busy.value = false
  }
}

async function test() {
  busy.value = true
  msg.value = ''
  try {
    const body = await testPush()
    say(`Sent: “${body}”. It should appear in a few seconds.`)
  } catch (err) {
    say(err instanceof Error && !('response' in err) ? err.message : errorMessage(err), false)
  } finally {
    busy.value = false
  }
}

async function addTime() {
  if (!newTime.value || props.times.includes(newTime.value)) {
    newTime.value = ''
    return
  }
  busy.value = true
  try {
    await saveTimes([...props.times, newTime.value])
    newTime.value = ''
    say('Saved.')
  } catch (err) {
    say(errorMessage(err), false)
  } finally {
    busy.value = false
  }
}

async function removeTime(t: string) {
  busy.value = true
  try {
    await saveTimes(props.times.filter((x) => x !== t))
    say('Saved.')
  } catch (err) {
    say(errorMessage(err), false)
  } finally {
    busy.value = false
  }
}

onMounted(async () => {
  try {
    enabled.value = await pushEnabledHere()
  } catch {
    enabled.value = false
  } finally {
    checking.value = false
  }
})
</script>
