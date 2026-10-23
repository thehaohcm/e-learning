// Service worker: right-click menu + Alt+Shift+N → save the selected word,
// with the sentence it is in, to the E-Learning notes.
importScripts('shared.js')

chrome.runtime.onInstalled.addListener((details) => {
  chrome.contextMenus.create({ id: 'save-word', title: 'Add “%s” to E-Learn notes', contexts: ['selection'] })
  if (details.reason === 'install') chrome.runtime.openOptionsPage()
})

chrome.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId === 'save-word') saveFromTab(tab, info.selectionText)
})

chrome.commands.onCommand.addListener(async (command, tab) => {
  if (command !== 'save-selection') return
  const t = tab || (await chrome.tabs.query({ active: true, currentWindow: true }))[0]
  saveFromTab(t, '')
})

/** Runs in the page: the selection, its sentence, and the page's site. */
function grabSelection() {
  const sel = window.getSelection()
  const text = (sel ? sel.toString() : '').trim()
  if (!text) return { text: '' }
  let el = sel.anchorNode
  if (el && el.nodeType !== 1) el = el.parentElement
  const block = el && el.closest('p,li,td,th,dd,dt,blockquote,figcaption,h1,h2,h3,h4,h5,h6,article,section,div')
  const full = ((block && block.innerText) || '').replace(/\s+/g, ' ').trim()
  let sentence = ''
  if (full) {
    const parts = full.match(/[^.!?]+[.!?]*["'”’)\]]*\s*/g) || [full]
    sentence = (parts.find((p) => p.toLowerCase().includes(text.toLowerCase())) || '').trim()
    if (sentence.length > 380) {
      const at = sentence.toLowerCase().indexOf(text.toLowerCase())
      const start = Math.max(0, at - 180)
      sentence = (start ? '…' : '') + sentence.slice(start, start + 360).trim() + '…'
    }
  }
  return { text, sentence, site: location.origin }
}

/** Runs in the page: a small message in the corner (shadow DOM, so the page CSS cannot break it). */
function showToast(message, ok) {
  const id = 'elearn-save-toast'
  document.getElementById(id)?.remove()
  const host = document.createElement('div')
  host.id = id
  host.style.cssText = 'position:fixed;z-index:2147483647;right:16px;bottom:16px'
  const root = host.attachShadow({ mode: 'open' })
  const box = document.createElement('div')
  box.textContent = message
  box.style.cssText = `font:14px/1.4 system-ui,-apple-system,"Segoe UI",sans-serif;color:#fff;max-width:320px;
    padding:10px 14px;border-radius:10px;box-shadow:0 6px 20px rgba(0,0,0,.25);background:${ok ? '#4f46e5' : '#b91c1c'}`
  root.appendChild(box)
  document.documentElement.appendChild(host)
  setTimeout(() => host.remove(), 2800)
}

async function inPage(tabId, func, args) {
  const [res] = await chrome.scripting.executeScript({ target: { tabId }, func, args: args || [] })
  return res && res.result
}

async function notify(tab, message, ok) {
  try {
    await inPage(tab.id, showToast, [message, ok])
  } catch {
    // pages the extension cannot touch (chrome://, the web store, PDF viewer): use the badge
    chrome.action.setBadgeBackgroundColor({ color: ok ? '#4f46e5' : '#b91c1c' })
    chrome.action.setBadgeText({ text: ok ? '✓' : '!', tabId: tab.id })
    chrome.action.setTitle({ title: message, tabId: tab.id })
    setTimeout(() => chrome.action.setBadgeText({ text: '', tabId: tab.id }), 4000)
  }
}

async function saveFromTab(tab, fallbackText) {
  if (!tab || tab.id === undefined) return
  let grabbed = { text: fallbackText || '', sentence: '', site: '' }
  try {
    const g = await inPage(tab.id, grabSelection)
    if (g && g.text) grabbed = g
  } catch {
    /* cannot run in this page: use the context-menu text only */
  }
  const word = cleanSelection(grabbed.text)
  if (!word) {
    notify(tab, grabbed.text ? 'Select an English word or a short phrase (up to 6 words).' : 'Select a word first.', false)
    return
  }
  const cfg = await getConfig()
  try {
    const note = await apiFetch('/ext/notes', {
      method: 'POST',
      body: JSON.stringify({ text: word, context: grabbed.sentence, url: cfg.sendSite ? grabbed.site : '' }),
    }, cfg)
    notify(tab, `✓ Saved “${note.text}” to E-Learn notes`, true)
  } catch (e) {
    notify(tab, e.message, false)
    if (!cfg.key) chrome.runtime.openOptionsPage()
  }
}
