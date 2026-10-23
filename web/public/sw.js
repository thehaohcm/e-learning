/* Service worker: makes the web app installable, opens fast (app shell cache)
 * and shows study reminders sent by the server (Web Push).
 * Plain JS, served as-is from /sw.js (not bundled by Vite). */
const SHELL = 'elearn-shell-v1'
const ASSETS = 'elearn-assets-v1'
const SHELL_FILES = ['/', '/manifest.webmanifest', '/icons/icon-192.png', '/icons/icon-512.png', '/icons/badge-96.png']
const MAX_ASSETS = 120

self.addEventListener('install', (event) => {
  event.waitUntil(caches.open(SHELL).then((c) => c.addAll(SHELL_FILES)).then(() => self.skipWaiting()))
})

self.addEventListener('activate', (event) => {
  event.waitUntil((async () => {
    for (const key of await caches.keys()) {
      if (key !== SHELL && key !== ASSETS) await caches.delete(key)
    }
    await trimAssets()
    await self.clients.claim()
  })())
})

async function trimAssets() {
  const cache = await caches.open(ASSETS)
  const keys = await cache.keys()
  for (const req of keys.slice(0, Math.max(0, keys.length - MAX_ASSETS))) await cache.delete(req)
}

self.addEventListener('fetch', (event) => {
  const req = event.request
  if (req.method !== 'GET') return
  const url = new URL(req.url)
  if (url.origin !== self.location.origin) return
  if (url.pathname.startsWith('/api/') || url.pathname === '/health') return // always live data

  // Pages: network first (always the latest version), the cached shell when offline.
  if (req.mode === 'navigate') {
    event.respondWith((async () => {
      try {
        const res = await fetch(req)
        if (res.ok) {
          const copy = res.clone()
          caches.open(SHELL).then((c) => c.put('/', copy))
        }
        return res
      } catch {
        return (await caches.match('/')) || Response.error()
      }
    })())
    return
  }

  // Built files have a hash in their name and never change: cache first.
  if (url.pathname.startsWith('/assets/')) {
    event.respondWith((async () => {
      const cached = await caches.match(req)
      if (cached) return cached
      const res = await fetch(req)
      if (res.ok) {
        const copy = res.clone()
        caches.open(ASSETS).then((c) => c.put(req, copy))
      }
      return res
    })())
    return
  }

  // Icons, manifest: cached copy, refreshed in the background.
  if (SHELL_FILES.includes(url.pathname)) {
    event.respondWith((async () => {
      const cached = await caches.match(req)
      const fresh = fetch(req).then((res) => {
        if (res.ok) {
          const copy = res.clone()
          caches.open(SHELL).then((c) => c.put(req, copy))
        }
        return res
      }).catch(() => cached)
      return cached || fresh
    })())
  }
})

// ---- study reminders (Web Push) -------------------------------------------
self.addEventListener('push', (event) => {
  let msg = {}
  try {
    msg = event.data ? event.data.json() : {}
  } catch {
    msg = { body: event.data ? event.data.text() : '' }
  }
  const title = msg.title || 'Time to study 📚'
  event.waitUntil(self.registration.showNotification(title, {
    body: msg.body || 'Your cards are waiting.',
    icon: '/icons/icon-192.png',
    badge: '/icons/badge-96.png',
    tag: msg.tag || 'study-reminder',
    renotify: true,
    data: { url: msg.url || '/review' },
  }))
})

self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  const target = new URL((event.notification.data && event.notification.data.url) || '/', self.location.origin).href
  event.waitUntil((async () => {
    const windows = await self.clients.matchAll({ type: 'window', includeUncontrolled: true })
    for (const w of windows) {
      if (new URL(w.url).origin === self.location.origin) {
        await w.focus()
        try {
          if ('navigate' in w) await w.navigate(target)
        } catch {
          // not controlled by this worker yet: the focused page stays as it is
        }
        return
      }
    }
    return self.clients.openWindow(target)
  })())
})
