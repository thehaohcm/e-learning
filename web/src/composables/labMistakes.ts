// Local (per-browser) spaced-review queue for pentest command-typing drills.
// "lab" cards are intentionally excluded from the server-side FSRS engine
// (see api/internal/study/service.go) because they're muscle-memory drills,
// not quiz material. This gives them a lightweight Leitner-style schedule of
// their own so a wrong/"Không nhớ" command comes back for another try later,
// the same spirit as vocab's spaced repetition — just local to this browser,
// not synced across devices.

export interface LabMistake {
  cardId: number
  stepIndex: number
  box: number // 1..INTERVALS.length; resets to 1 on a miss, grows by one per correct review
  dueAt: number // epoch ms
  misses: number
}

const KEY = 'elearn.labs.mistakes'
// 10 min -> 1 day -> 3 days -> 7 days -> 16 days, then "mastered" (dropped).
const INTERVALS = [10 * 60_000, 24 * 3600_000, 3 * 24 * 3600_000, 7 * 24 * 3600_000, 16 * 24 * 3600_000]

function load(): LabMistake[] {
  try { return JSON.parse(localStorage.getItem(KEY) || '[]') } catch { return [] }
}
function save(list: LabMistake[]) {
  try { localStorage.setItem(KEY, JSON.stringify(list)) } catch { /* ignore */ }
}
function keyOf(cardId: number, stepIndex: number): string { return `${cardId}:${stepIndex}` }

/** Wrong answer or "Không nhớ": (re)schedule this step for review soon, from box 1. */
export function recordLabMistake(cardId: number, stepIndex: number): void {
  const list = load()
  const k = keyOf(cardId, stepIndex)
  const i = list.findIndex((m) => keyOf(m.cardId, m.stepIndex) === k)
  const now = Date.now()
  if (i >= 0) list[i] = { ...list[i], box: 1, dueAt: now + INTERVALS[0], misses: list[i].misses + 1 }
  else list.push({ cardId, stepIndex, box: 1, dueAt: now + INTERVALS[0], misses: 1 })
  save(list)
}

/** Correct answer: if this step is tracked as a mistake, push it to the next
 *  box (further out), or drop it once it clears the last box (mastered).
 *  No-op if the step was never missed — never creates an entry. */
export function resolveLabMistake(cardId: number, stepIndex: number): void {
  const list = load()
  const k = keyOf(cardId, stepIndex)
  const i = list.findIndex((m) => keyOf(m.cardId, m.stepIndex) === k)
  if (i < 0) return
  const nextBox = list[i].box + 1
  if (nextBox > INTERVALS.length) {
    list.splice(i, 1)
  } else {
    list[i] = { ...list[i], box: nextBox, dueAt: Date.now() + INTERVALS[nextBox - 1] }
  }
  save(list)
}

/** Mistakes whose review time has come, soonest first. */
export function dueLabMistakes(): LabMistake[] {
  const now = Date.now()
  return load().filter((m) => m.dueAt <= now).sort((a, b) => a.dueAt - b.dueAt)
}

export function dueLabMistakeCount(): number {
  return dueLabMistakes().length
}
