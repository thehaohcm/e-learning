// Keyboard flow for a list of fill-in-the-blank sentences (More examples, Round check):
//   Enter in a blank  → check (the component does it); a wrong answer is selected so it can be retyped
//   S                 → say the current (filled) sentence and get a score; ≥ 80% jumps to the next blank
//   N / P             → next blank / previous sentence
// S, N and P only act when the focus is not in a text box, and the keys handled
// here do not reach the page's own shortcuts (e.g. Study: P = pronounce, Enter = grade).
import { nextTick, ref } from 'vue'

export const SPEAK_PASS = 80

export function useBlankNav(opts: {
  ids: () => number[] // sentence ids in display order
  isOpen: (id: number) => boolean // the sentence still has a blank to fill
  speak: (id: number) => void // open the 🎤 panel for this sentence and start listening
}) {
  const current = ref(0) // id of the active sentence
  const root = ref<HTMLElement | null>(null)
  const inputs = new Map<number, HTMLInputElement>()

  /** :ref="inputRef(id)" on each blank */
  function inputRef(id: number) {
    return (el: unknown) => {
      if (el instanceof HTMLInputElement) inputs.set(id, el)
      else inputs.delete(id)
    }
  }

  async function focusItem(id: number) {
    current.value = id
    await nextTick()
    const input = inputs.get(id)
    if (opts.isOpen(id) && input) {
      input.focus()
      input.select()
    } else {
      root.value?.focus({ preventScroll: true })
      root.value?.querySelector<HTMLElement>(`[data-id="${id}"]`)?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
    }
  }

  /** Next sentence with a blank (after the current one), else simply the next sentence. */
  function next() {
    const ids = opts.ids()
    const i = ids.indexOf(current.value)
    const after = ids.slice(i + 1)
    const target = after.find((id) => opts.isOpen(id)) ?? after[0]
    if (target !== undefined) focusItem(target)
  }

  function prev() {
    const ids = opts.ids()
    const i = ids.indexOf(current.value)
    if (i > 0) focusItem(ids[i - 1])
  }

  /** After a blank is solved: its input disappears, keep the keyboard on this list. */
  function solved(id: number) {
    focusItem(id)
  }

  /** Wrong answer: select it so typing replaces it. */
  function wrong(e: Event) {
    const t = e.target
    if (t instanceof HTMLInputElement) t.select()
  }

  /** Score of the 🎤 check of the current sentence. */
  function scored(pct: number) {
    if (pct >= SPEAK_PASS) next()
  }

  function onKeydown(e: KeyboardEvent) {
    const tag = (e.target as HTMLElement)?.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA') return // typing: Enter is handled by the blank itself
    const k = e.key.toLowerCase()
    if (k === 's' && current.value && !opts.isOpen(current.value)) opts.speak(current.value)
    else if (k === 'n') next()
    else if (k === 'p') prev()
    else if (k !== 'enter') return
    e.preventDefault()
    e.stopPropagation()
  }

  return { current, root, inputRef, focusItem, next, prev, solved, wrong, scored, onKeydown }
}
