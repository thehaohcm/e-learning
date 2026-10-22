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
