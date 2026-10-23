/**
 * Regexp matching a word or phrasal verb in common forms:
 * negotiate → negotiated / negotiating, carry out → carried out / carry it out.
 */
export function wordRegex(word: string, flags = 'i'): RegExp {
  const parts = word.toLowerCase().trim().split(/\s+/).map((p, i) => {
    let x = p.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    if (i === 0 && x.length > 3 && /[ey]$/.test(x)) x = x.slice(0, -1)
    return x + '\\w*'
  })
  return new RegExp('\\b' + parts.join('(?:\\W+\\w+){0,3}?\\W+') + '\\b', flags)
}

/** Does the text use the word in some form? */
export function usesWord(text: string, word: string): boolean {
  return wordRegex(word).test(text)
}

export interface WordSegment { text: string; hit: boolean; first?: boolean }

/** Text split around the word (any form), e.g. for highlighting or a blank. */
export function splitByWord(text: string, word: string): WordSegment[] {
  const out: WordSegment[] = []
  let last = 0
  for (const m of text.matchAll(wordRegex(word, 'gi'))) {
    const i = m.index ?? 0
    if (i > last) out.push({ text: text.slice(last, i), hit: false })
    out.push({ text: m[0], hit: true, first: !out.some((s) => s.hit) })
    last = i + m[0].length
  }
  if (last < text.length) out.push({ text: text.slice(last), hit: false })
  return out
}

const normalize = (s: string) => s.toLowerCase().replace(/[^a-z' ]/g, ' ').replace(/\s+/g, ' ').trim()

/**
 * Checks a word typed in a blank: "exact" = the form used in the sentence,
 * "base" = the dictionary form (accepted, the right form should be shown),
 * "wrong" otherwise. Spelling must be right.
 */
export function checkTypedWord(typed: string, expected: string, word: string): 'exact' | 'base' | 'wrong' {
  const t = normalize(typed)
  if (t && t === normalize(expected)) return 'exact'
  if (t && t === normalize(word)) return 'base'
  return 'wrong'
}
