// Shared by the service worker (importScripts), the popup and the options page.
/* exported DEFAULT_SERVER, getConfig, setConfig, apiFetch, cleanSelection */

const DEFAULT_SERVER = 'https://elearn.tradesign.io.vn'

/** { server, key, sendSite } from chrome.storage.sync */
async function getConfig() {
  const c = await chrome.storage.sync.get({ server: DEFAULT_SERVER, key: '', sendSite: true })
  c.server = (c.server || DEFAULT_SERVER).replace(/\/+$/, '')
  return c
}

async function setConfig(patch) {
  await chrome.storage.sync.set(patch)
}

/** Calls the E-Learning API with the extension key. Throws Error(message) on failure. */
async function apiFetch(path, options = {}, cfg) {
  const c = cfg || (await getConfig())
  if (!c.key) throw new Error('Not set up yet: open the extension options and paste your key.')
  let res
  try {
    res = await fetch(c.server + '/api' + path, {
      ...options,
      headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + c.key, ...(options.headers || {}) },
    })
  } catch {
    throw new Error('Cannot reach ' + c.server + ' (offline or wrong address).')
  }
  let data = null
  try {
    data = await res.json()
  } catch {
    /* empty body */
  }
  if (!res.ok) throw new Error((data && data.error) || 'Server error ' + res.status)
  return data
}

/** A word or short phrase worth saving, or '' (sentences, numbers, other languages...). */
function cleanSelection(text) {
  const t = (text || '').replace(/[‘’]/g, "'").replace(/^[^A-Za-z]+|[^A-Za-z']+$/g, '').replace(/\s+/g, ' ').trim()
  if (!t || t.length > 60 || t.split(' ').length > 6) return ''
  return /^[A-Za-z][A-Za-z' -]*$/.test(t) ? t : ''
}
