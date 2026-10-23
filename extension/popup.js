/* global getConfig, apiFetch, cleanSelection */
const $ = (id) => document.getElementById(id)
let cfg

function msg(text, ok) {
  $('msg').textContent = text
  $('msg').className = 'small ' + (ok ? 'ok' : 'err')
}

function render(notes) {
  const ul = $('recent')
  ul.textContent = ''
  if (!notes.length) {
    const li = document.createElement('li')
    li.className = 'muted'
    li.textContent = 'Nothing yet.'
    ul.appendChild(li)
    return
  }
  for (const n of notes) {
    const li = document.createElement('li')
    const b = document.createElement('b')
    b.textContent = n.text
    li.appendChild(b)
    const meaning = n.info && n.info.meaning_vi
    const extra = document.createElement('span')
    extra.className = 'muted'
    extra.textContent = meaning ? ' — ' + meaning : n.status === 'added' ? ' — in Study' : ''
    li.appendChild(extra)
    ul.appendChild(li)
  }
}

async function loadRecent() {
  try {
    render(await apiFetch('/ext/notes', {}, cfg))
  } catch (e) {
    msg(e.message, false)
  }
}

async function init() {
  cfg = await getConfig()
  $('open').addEventListener('click', (e) => {
    e.preventDefault()
    chrome.tabs.create({ url: cfg.server + '/' })
  })
  if (!cfg.key) {
    $('setup').hidden = false
    $('options').addEventListener('click', () => chrome.runtime.openOptionsPage())
    return
  }
  $('main').hidden = false
  $('word').focus()
  loadRecent()
}

$('add').addEventListener('submit', async (e) => {
  e.preventDefault()
  const word = cleanSelection($('word').value)
  if (!word) {
    msg('Type an English word or a short phrase.', false)
    return
  }
  try {
    const note = await apiFetch('/ext/notes', { method: 'POST', body: JSON.stringify({ text: word }) }, cfg)
    msg('✓ Saved “' + note.text + '”', true)
    $('word').value = ''
    loadRecent()
  } catch (err) {
    msg(err.message, false)
  }
})

init()
