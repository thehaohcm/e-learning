// A small, dependency-free Markdown → HTML renderer, safe by construction:
// every character is HTML-escaped FIRST, then a limited set of Markdown
// constructs is turned into tags. No raw HTML from the source is ever emitted,
// so notes (the user's own, or an AI summary of them) cannot inject scripts.
//
// Supported: headings (#..######), fenced code ```lang, tables (| a | b |),
// inline `code`, bold **x** / __x__, italic *x* / _x_, strikethrough ~~x~~,
// links [text](http…), unordered (-, *) and ordered (1.) lists,
// blockquotes (>), horizontal rules (---), paragraphs & line breaks.

function esc(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

/** Inline formatting on an already HTML-escaped string. */
function inline(s: string): string {
  // inline code first (protect its contents from other rules)
  const codes: string[] = []
  s = s.replace(/`([^`]+)`/g, (_m, c) => {
    codes.push(`<code>${c}</code>`)
    return `\u0000${codes.length - 1}\u0000`
  })
  // links [text](url) — only http(s), url already escaped
  s = s.replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g, (_m, t, u) =>
    `<a href="${u}" target="_blank" rel="noopener noreferrer">${t}</a>`)
  s = s.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  s = s.replace(/__([^_]+)__/g, '<strong>$1</strong>')
  s = s.replace(/~~([^~]+)~~/g, '<del>$1</del>')
  s = s.replace(/(^|[\s(])\*([^*\s][^*]*?)\*(?=[\s).,!?]|$)/g, '$1<em>$2</em>')
  s = s.replace(/(^|[\s(])_([^_\s][^_]*?)_(?=[\s).,!?]|$)/g, '$1<em>$2</em>')
  // restore inline code
  s = s.replace(/\u0000(\d+)\u0000/g, (_m, i) => codes[Number(i)])
  return s
}

function splitTableCells(line: string): string[] {
  let s = line.trim()
  if (s.startsWith('|')) s = s.slice(1)
  if (s.endsWith('|')) s = s.slice(0, -1)
  const cells: string[] = []
  let cur = ''
  let inCode = false
  for (let j = 0; j < s.length; j++) {
    const ch = s[j]
    if (ch === '`') inCode = !inCode
    if (ch === '|' && !inCode) {
      cells.push(cur.trim())
      cur = ''
    } else {
      cur += ch
    }
  }
  cells.push(cur.trim())
  return cells
}

function isTableDelimiter(line: string): boolean {
  if (!line || !line.includes('-')) return false
  const cells = splitTableCells(line)
  if (cells.length < 1) return false
  return cells.every(c => /^:?-{1,}:?$/.test(c.trim()))
}

function isTableStart(lines: string[], i: number): boolean {
  if (i + 1 >= lines.length) return false
  const cur = lines[i].trim()
  const next = lines[i + 1].trim()
  if (!cur.includes('|') || !next.includes('|')) return false
  return isTableDelimiter(next)
}

function getTableAlignments(delimiterLine: string): Array<'left' | 'center' | 'right' | null> {
  return splitTableCells(delimiterLine).map(c => {
    const t = c.trim()
    const left = t.startsWith(':')
    const right = t.endsWith(':')
    if (left && right) return 'center'
    if (right) return 'right'
    if (left) return 'left'
    return null
  })
}

export function renderMarkdown(src: string): string {
  const lines = esc(src ?? '').replace(/\r\n?/g, '\n').split('\n')
  const out: string[] = []
  let i = 0
  let listType: 'ul' | 'ol' | null = null

  const closeList = () => {
    if (listType) {
      out.push(`</${listType}>`)
      listType = null
    }
  }

  while (i < lines.length) {
    const line = lines[i]

    // fenced code block
    const fence = line.match(/^\s*```([a-zA-Z0-9_+-]+)?\s*$/)
    if (fence) {
      closeList()
      const body: string[] = []
      i++
      while (i < lines.length && !/^\s*```\s*$/.test(lines[i])) {
        body.push(lines[i])
        i++
      }
      if (i < lines.length) i++ // skip closing fence
      const lang = fence[1] ? ` class="language-${fence[1]}"` : ''
      out.push(`<pre><code${lang}>${body.join('\n')}</code></pre>`)
      continue
    }

    // blank line
    if (/^\s*$/.test(line)) {
      closeList()
      i++
      continue
    }

    // horizontal rule
    if (/^\s*(-{3,}|\*{3,}|_{3,})\s*$/.test(line)) {
      closeList()
      out.push('<hr>')
      i++
      continue
    }

    // heading
    const h = line.match(/^(#{1,6})\s+(.*)$/)
    if (h) {
      closeList()
      const level = h[1].length
      out.push(`<h${level}>${inline(h[2].trim())}</h${level}>`)
      i++
      continue
    }

    // table
    if (isTableStart(lines, i)) {
      closeList()
      const headerCells = splitTableCells(lines[i])
      const alignments = getTableAlignments(lines[i + 1])
      i += 2
      const bodyRows: string[][] = []
      while (i < lines.length) {
        const rowLine = lines[i].trim()
        if (!rowLine || !rowLine.includes('|')) break
        if (/^\s*```/.test(rowLine) || /^(#{1,6})\s/.test(rowLine)) break
        bodyRows.push(splitTableCells(rowLine))
        i++
      }
      const ths = headerCells.map((c, idx) => {
        const align = alignments[idx] ? ` style="text-align:${alignments[idx]};"` : ''
        return `<th${align}>${inline(c)}</th>`
      }).join('')
      const trs = bodyRows.map(row => {
        const tds = headerCells.map((_, idx) => {
          const c = row[idx] ?? ''
          const align = alignments[idx] ? ` style="text-align:${alignments[idx]};"` : ''
          return `<td${align}>${inline(c)}</td>`
        }).join('')
        return `<tr>${tds}</tr>`
      }).join('\n')
      out.push(`<div class="table-responsive my-2"><table class="table table-sm table-bordered"><thead><tr>${ths}</tr></thead><tbody>${trs}</tbody></table></div>`)
      continue
    }

    // blockquote (single level)
    if (/^\s*&gt;\s?/.test(line)) {
      closeList()
      const quote: string[] = []
      while (i < lines.length && /^\s*&gt;\s?/.test(lines[i])) {
        quote.push(inline(lines[i].replace(/^\s*&gt;\s?/, '')))
        i++
      }
      out.push(`<blockquote>${quote.join('<br>')}</blockquote>`)
      continue
    }

    // unordered list
    const ul = line.match(/^\s*[-*]\s+(.*)$/)
    if (ul) {
      if (listType !== 'ul') {
        closeList()
        out.push('<ul>')
        listType = 'ul'
      }
      out.push(`<li>${inline(ul[1])}</li>`)
      i++
      continue
    }

    // ordered list
    const ol = line.match(/^\s*\d+[.)]\s+(.*)$/)
    if (ol) {
      if (listType !== 'ol') {
        closeList()
        out.push('<ol>')
        listType = 'ol'
      }
      out.push(`<li>${inline(ol[1])}</li>`)
      i++
      continue
    }

    // paragraph (join following non-blank, non-special lines with <br>)
    closeList()
    const para: string[] = [inline(line.trim())]
    i++
    while (
      i < lines.length &&
      !/^\s*$/.test(lines[i]) &&
      !/^\s*```/.test(lines[i]) &&
      !/^(#{1,6})\s/.test(lines[i]) &&
      !/^\s*[-*]\s+/.test(lines[i]) &&
      !/^\s*\d+[.)]\s+/.test(lines[i]) &&
      !/^\s*&gt;\s?/.test(lines[i]) &&
      !/^\s*(-{3,}|\*{3,}|_{3,})\s*$/.test(lines[i]) &&
      !isTableStart(lines, i)
    ) {
      para.push(inline(lines[i].trim()))
      i++
    }
    out.push(`<p>${para.join('<br>')}</p>`)
  }
  closeList()
  return out.join('\n')
}
