// Shared command-matching + model-command helpers for lab/flow drills
// (Labs challenge sim and Daily Lab).

export function normalizeCmd(raw: string): string {
  return raw.trim().replace(/\s+/g, ' ')
}

function flagSpecMatch(spec: string, input: string): boolean {
  const parts = spec.split(':')
  const cmd = (parts[1] ?? '').toLowerCase()
  const letters = (parts[2] ?? '').toLowerCase()
  const toks = input.trim().toLowerCase().split(/\s+/)
  if (!cmd || toks[0] !== cmd) return false
  const have = new Set<string>()
  for (const t of toks.slice(1)) {
    if (/^-[a-z0-9]+$/.test(t)) for (const ch of t.slice(1)) have.add(ch)
  }
  for (const ch of letters) if (!have.has(ch)) return false
  return true
}

export function matchPattern(p: string, norm: string): boolean {
  if (p.startsWith('@flags:')) return flagSpecMatch(p, norm)
  try { return new RegExp(p, 'i').test(norm) } catch { return false }
}

export function matchAny(patterns: string[] | undefined, input: string): boolean {
  const n = normalizeCmd(input)
  return (patterns ?? []).some((p) => matchPattern(p, n))
}

// Reconstruct a readable model command from the first regex pattern (for the
// "type this" hint / review), stripping regex syntax.
export function modelCommand(patterns: string[] | undefined): string {
  const p = patterns?.[0]
  if (!p) return ''
  if (p.startsWith('@flags:')) {
    const [, cmd = '', letters = ''] = p.split(':')
    return letters ? `${cmd} -${letters}` : cmd
  }
  let t = p.replace(/^\^/, '').replace(/\$$/, '')
  for (let i = 0; i < 6; i++) t = t.replace(/\((?:\?:)?([^()]*)\)[?*+]?/g, (_m, g: string) => g.split('|')[0])
  t = t.replace(/\[([^\]]*)\]/g, (_m, g: string) => (g.replace(/\\(.)/g, '$1')[0] ?? ''))
  t = t.replace(/\\s[+*?]?/g, ' ')
  t = t.replace(/\.[*+]/g, ' … ')
  t = t.replace(/\\(.)/g, '$1')
  t = t.replace(/([^\s…])[?*+]/g, '$1')
  t = t.replace(/\s+/g, ' ').trim()
  return t
}
