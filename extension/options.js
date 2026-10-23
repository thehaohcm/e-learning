/* global getConfig, setConfig, apiFetch */
const $ = (id) => document.getElementById(id)

function status(text, ok) {
  $('status').textContent = text
  $('status').className = 'small ' + (ok ? 'ok' : 'err')
}

async function init() {
  const c = await getConfig()
  $('server').value = c.server
  $('key').value = c.key
  $('sendSite').checked = c.sendSite
  if (c.key) test(c)
}

async function test(c) {
  status('Checking…', true)
  try {
    const me = await apiFetch('/ext/me', {}, c)
    status('✓ Connected as ' + (me.display_name || me.email), true)
  } catch (e) {
    status(e.message, false)
  }
}

$('form').addEventListener('submit', async (e) => {
  e.preventDefault()
  const server = $('server').value.trim().replace(/\/+$/, '')
  if (!/^https?:\/\//.test(server)) {
    status('The address must start with https://', false)
    return
  }
  const patch = { server, key: $('key').value.trim(), sendSite: $('sendSite').checked }
  await setConfig(patch)
  if (!patch.key) {
    status('Saved. Paste your key to connect.', false)
    return
  }
  test({ ...patch })
})

init()
