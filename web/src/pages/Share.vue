<template>
  <div class="row justify-content-center">
    <div class="col-lg-7">
      <h2 class="fw-bold mb-3"><i class="bi bi-box-arrow-in-down"></i> Save a word</h2>

      <div class="card border-0 shadow-sm">
        <div class="card-body">
          <template v-if="!shared">
            <p class="text-muted mb-0">
              Nothing was shared. On Android, select a word in any app or browser, tap <b>Share</b> and choose <b>E-Learn</b>.
            </p>
          </template>

          <template v-else>
            <!-- a long text: tap the word(s) to save -->
            <template v-if="tokens.length > 1 && !single">
              <div class="small text-muted mb-2">Tap the word (or the words of a phrase) you want to save:</div>
              <p class="shared-text">
                <template v-for="(t, i) in tokens" :key="i">
                  <span v-if="t.word" class="tok" :class="{ on: picked.has(i) }" @click="toggle(i)">{{ t.text }}</span><template v-else>{{ t.text }}</template>
                </template>
              </p>
            </template>

            <label class="form-label small fw-semibold" for="w">Word or phrase</label>
            <input id="w" v-model="word" class="form-control mb-2" maxlength="60" autocomplete="off" />
            <div v-if="context" class="small text-muted mb-3"><i class="bi bi-quote"></i> {{ context }}</div>

            <div class="d-flex flex-wrap align-items-center gap-2">
              <button class="btn btn-primary" :disabled="!word.trim() || saving" @click="save">
                <span v-if="saving" class="spinner-border spinner-border-sm me-1"></span>
                <i v-else class="bi bi-journal-plus"></i> Save to My notes
              </button>
              <router-link to="/" class="btn btn-outline-secondary">Cancel</router-link>
              <span v-if="savedText" class="text-success small"><i class="bi bi-check-circle-fill"></i> Saved “{{ savedText }}”. Open 📝 to explain it or add it to Study.</span>
              <span v-if="error" class="text-danger small">{{ error }}</span>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, errorMessage } from '../api'

/**
 * PWA share target (Android): another app shares text to E-Learn → /share?text=…
 * A short text is the word itself; a longer one lets the learner tap the word(s)
 * and keeps the sentence they are in as context.
 */
const route = useRoute()
const str = (v: unknown) => (Array.isArray(v) ? String(v[0] ?? '') : String(v ?? ''))
const sharedUrl = str(route.query.url)
const raw = computed(() =>
  (str(route.query.text) || str(route.query.title)).replace(/https?:\/\/\S+/g, ' ').replace(/\s+/g, ' ').trim(),
)
const shared = computed(() => raw.value.length > 0)
const single = computed(() => raw.value.split(' ').length <= 6 && raw.value.length <= 60)

interface Tok { text: string; word: boolean }
const tokens = computed<Tok[]>(() => (raw.value.match(/[A-Za-z][A-Za-z'-]*|[^A-Za-z]+/g) ?? []).map((t) => ({ text: t, word: /^[A-Za-z]/.test(t) })))

const picked = reactive(new Set<number>())
const word = ref('')
const saving = ref(false)
const savedText = ref('')
const error = ref('')

if (single.value) word.value = raw.value.replace(/^[^A-Za-z]+|[^A-Za-z']+$/g, '')

function toggle(i: number) {
  if (picked.has(i)) picked.delete(i)
  else picked.add(i)
  const idx = [...picked].sort((a, b) => a - b)
  if (!idx.length) {
    word.value = ''
    return
  }
  // the words from the first to the last tapped one (a phrase), at most 6
  const parts = tokens.value.slice(idx[0], idx[idx.length - 1] + 1).map((t) => t.text).join('')
  word.value = parts.replace(/\s+/g, ' ').trim().split(' ').slice(0, 6).join(' ')
}

/** The sentence of the shared text that contains the word. */
const context = computed(() => {
  if (single.value || !word.value.trim()) return ''
  const sentences = raw.value.match(/[^.!?]+[.!?]*/g) ?? [raw.value]
  const hit = sentences.find((s) => s.toLowerCase().includes(word.value.trim().toLowerCase()))
  return (hit ?? '').trim().slice(0, 380)
})

watch(word, () => { savedText.value = ''; error.value = '' })

function siteOf(u: string) {
  try {
    return new URL(u).hostname.replace(/^www\./, '')
  } catch {
    return ''
  }
}

async function save() {
  saving.value = true
  error.value = ''
  try {
    const { data } = await api.post('/notes', { text: word.value.trim(), context: context.value, source: siteOf(sharedUrl) || 'Shared' })
    savedText.value = data.text
    window.dispatchEvent(new CustomEvent('elearn:note-saved', { detail: data }))
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.shared-text { line-height: 2; }
.tok { cursor: pointer; border-radius: 4px; padding: 1px 2px; }
.tok:hover { background: var(--bs-light); }
.tok.on { background: #ffe69c; font-weight: 600; }
</style>
