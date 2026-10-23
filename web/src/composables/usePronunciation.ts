// Pronunciation help (syllables, stress, tips in Vietnamese) for English cards.
// Showing a card only reads what is already stored (batched: one call for all
// cards on screen, no AI). AI is called only when the learner presses
// "How to say it" on a word (generate), and the result is kept for everyone.
import { reactive } from 'vue'
import { api, errorMessage } from '../api'

export interface Pronunciation { syllables: string[]; stress: number; tips_vi: string[] }

const cache = reactive(new Map<number, Pronunciation | null>()) // null = not explained yet
const generating = reactive(new Set<number>())
const errors = reactive(new Map<number, string>())
const queued = new Set<number>()
let timer: number | undefined

async function flush() {
  timer = undefined
  const ids = [...queued]
  queued.clear()
  for (let i = 0; i < ids.length; i += 20) {
    const chunk = ids.slice(i, i + 20)
    try {
      const { data } = await api.post('/cards/pronunciations', { card_ids: chunk })
      for (const id of chunk) cache.set(id, data.items?.[id] ?? null)
    } catch {
      for (const id of chunk) cache.set(id, null)
    }
  }
}

export function usePronunciation() {
  /** Stored pronunciation of the card, if any (once per session). */
  function request(cardId: number) {
    if (cache.has(cardId) || queued.has(cardId)) return
    queued.add(cardId)
    if (timer === undefined) timer = window.setTimeout(flush, 60)
  }

  /** "How to say it": one AI request for this word. */
  async function generate(cardId: number) {
    if (generating.has(cardId)) return
    generating.add(cardId)
    errors.delete(cardId)
    try {
      const { data } = await api.post(`/cards/${cardId}/pronunciation`)
      cache.set(cardId, data)
    } catch (e) {
      errors.set(cardId, errorMessage(e))
    } finally {
      generating.delete(cardId)
    }
  }

  return { cache, generating, errors, request, generate }
}
