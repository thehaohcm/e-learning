// Installable app (PWA) and study reminders (Web Push).
// Imported once from main.ts so the install prompt is caught early.
import { computed, ref } from 'vue'
import { api } from '../api'

// ---- install --------------------------------------------------------------
interface InstallPrompt extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

const deferred = ref<InstallPrompt | null>(null)

function standalone(): boolean {
  return window.matchMedia?.('(display-mode: standalone)').matches || (navigator as unknown as { standalone?: boolean }).standalone === true
}

/** Running as the installed app (home screen), not in a browser tab. */
export const isInstalled = ref(standalone())
/** iPhone / iPad (also iPadOS that reports itself as a Mac). */
export const isIOS = /iphone|ipad|ipod/i.test(navigator.userAgent) || (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1)
/** The browser offers an install dialog (Chrome, Edge, Android). */
export const canInstall = computed(() => !!deferred.value && !isInstalled.value)

window.addEventListener('beforeinstallprompt', (e) => {
  e.preventDefault() // show our own button instead of the mini bar
  deferred.value = e as InstallPrompt
})
window.addEventListener('appinstalled', () => {
  deferred.value = null
  isInstalled.value = true
})

export async function installApp(): Promise<boolean> {
  const p = deferred.value
  if (!p) return false
  await p.prompt()
  const { outcome } = await p.userChoice
  deferred.value = null
  return outcome === 'accepted'
}

/** Registers /sw.js (production build only: the dev server has no sw.js build step). */
export function registerServiceWorker() {
  if (!('serviceWorker' in navigator) || !import.meta.env.PROD) return
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').catch((e) => console.warn('service worker:', e))
  })
}

// ---- push notifications -----------------------------------------------------
/** The browser can receive push notifications (iOS: only the installed app, iOS 16.4+). */
export const pushSupported = 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window

export function notificationPermission(): NotificationPermission | 'unsupported' {
  return pushSupported ? Notification.permission : 'unsupported'
}

async function registration(): Promise<ServiceWorkerRegistration> {
  const reg = await navigator.serviceWorker.getRegistration()
  if (reg) return reg
  if (!import.meta.env.PROD) throw new Error('Notifications need the built app (the dev server has no service worker).')
  return navigator.serviceWorker.register('/sw.js')
}

function keyBytes(b64url: string): Uint8Array {
  const b64 = b64url.replace(/-/g, '+').replace(/_/g, '/') + '='.repeat((4 - (b64url.length % 4)) % 4)
  const raw = atob(b64)
  const out = new Uint8Array(raw.length)
  for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i)
  return out
}

function sameKey(a: ArrayBuffer | null | undefined, b: Uint8Array): boolean {
  if (!a) return false
  const x = new Uint8Array(a)
  return x.length === b.length && x.every((v, i) => v === b[i])
}

/** This browser's current subscription, if any. */
export async function currentSubscription(): Promise<PushSubscription | null> {
  if (!pushSupported) return null
  const reg = await navigator.serviceWorker.getRegistration()
  return reg ? reg.pushManager.getSubscription() : null
}

/** Is this browser subscribed AND known to the server for this account? */
export async function pushEnabledHere(): Promise<boolean> {
  if (!pushSupported || Notification.permission !== 'granted') return false
  const sub = await currentSubscription()
  if (!sub) return false
  const { data } = await api.get('/push/key', { params: { endpoint: sub.endpoint } })
  return !!data.subscribed
}

/** Asks for permission (must run from a click), subscribes and tells the server. */
export async function enablePush(): Promise<void> {
  if (!pushSupported) throw new Error('This browser cannot receive notifications.')
  const perm = await Notification.requestPermission()
  if (perm !== 'granted') {
    throw new Error(perm === 'denied'
      ? 'Notifications are blocked for this site. Allow them in the browser’s site settings, then try again.'
      : 'Notifications were not allowed.')
  }
  const reg = await registration()
  const { data } = await api.get('/push/key')
  const key = keyBytes(data.public_key)
  let sub = await reg.pushManager.getSubscription()
  if (sub && !sameKey(sub.options?.applicationServerKey, key)) {
    await sub.unsubscribe() // made for another server key
    sub = null
  }
  if (!sub) sub = await reg.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: key.buffer as ArrayBuffer })
  await api.post('/push/subscribe', sub.toJSON())
}

/** Stops reminders on this browser. */
export async function disablePush(): Promise<void> {
  const sub = await currentSubscription()
  if (!sub) return
  await api.post('/push/unsubscribe', { endpoint: sub.endpoint }).catch(() => {})
  await sub.unsubscribe().catch(() => false)
}

/** Sends a reminder right now to this browser. */
export async function testPush(): Promise<string> {
  const sub = await currentSubscription()
  if (!sub) throw new Error('Notifications are off on this device.')
  const { data } = await api.post('/push/test', { endpoint: sub.endpoint })
  return data.message?.body ?? ''
}

/**
 * On app start: browsers sometimes renew a subscription; send the current one
 * again so reminders keep arriving. Quiet, does nothing without permission.
 */
export async function syncPush(): Promise<void> {
  try {
    if (!pushSupported || Notification.permission !== 'granted') return
    const sub = await currentSubscription()
    if (sub) await api.post('/push/subscribe', sub.toJSON())
  } catch {
    // offline or server without push: try again next start
  }
}
