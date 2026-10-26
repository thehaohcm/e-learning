import { computed, reactive } from 'vue'
import { api } from '../api'
import type { Subject } from '../types'

// The topic the web app shows: English or PMP. Chosen with the switch in the
// menu and remembered in this browser; every page (Home, Study, Practice,
// Quiz, Browse) shows only this topic.

const KEY = 'elearn.subject'

function stored(): string {
  try { return localStorage.getItem(KEY) || 'english' } catch { return 'english' }
}

const state = reactive({
  slug: stored(),
  subjects: [] as Subject[],
  loading: null as Promise<void> | null,
})

/** Loads the topics once (with their decks). */
export function loadSubjects(force = false): Promise<void> {
  if (state.loading && !force) return state.loading
  state.loading = api.get('/subjects').then(({ data }) => {
    state.subjects = data ?? []
  }).catch(() => {
    state.loading = null // try again next time
  })
  return state.loading
}

export const subjects = computed(() => state.subjects)

/** The chosen topic (null until the topics are loaded). */
export const subject = computed<Subject | null>(
  () => state.subjects.find((s) => s.slug === state.slug) ?? state.subjects[0] ?? null,
)

export const subjectSlug = computed(() => subject.value?.slug ?? state.slug)
export const subjectId = computed(() => subject.value?.id ?? 0)
export const isEnglish = computed(() => subjectSlug.value === 'english')

/** Short name for the menu: "English", "PMP". */
export function shortName(slug: string, name = ''): string {
  return slug === 'english' ? 'English' : slug === 'pmp' ? 'PMP' : slug === 'pentest' ? 'Pentest' : name || slug
}

export function setSubject(slug: string) {
  state.slug = slug
  try { localStorage.setItem(KEY, slug) } catch { /* ignore */ }
}

/** The topic of a deck (to follow a link like /review?deck_id=5 into the right topic). */
export function subjectOfDeck(deckId: number): Subject | undefined {
  return state.subjects.find((s) => s.decks?.some((d) => d.id === deckId))
}
