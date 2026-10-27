// A small, dependency-free Markdown → HTML renderer, safe by construction:
// every character is HTML-escaped FIRST, then a limited set of Markdown
// constructs is turned into tags. No raw HTML from the source is ever emitted,
// so notes (the user's own, or an AI summary of them) cannot inject scripts.
//
// Supported: headings (#..######), fenced code ```lang, inline `code`, bold
// **x**, italic *x* / _x_, links [text](http…), unordered (-, *) and ordered
// (1.) lists, blockquotes (>), horizontal rules (---), paragraphs & line breaks.

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
  s = s.replace(/(^|[\s(])\*([^*\s][^*]*?)\*(?=[\s).,!?]|$)/g, '$1<em>$2</em>')
  s = s.replace(/(^|[\s(])_([^_\s][^_]*?)_(?=[\s).,!?]|$)/g, '$1<em>$2</em>')
  // restore inline code
  s = s.replace(/\u0000(\d+)\u0000/g, (_m, i) => codes[Number(i)])
  return s
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
    const fence = line.match(/^```(\w+)?\s*$/)
    if (fence) {
      closeList()
      const body: string[] = []
      i++
      while (i < lines.length && !/^```\s*$/.test(lines[i])) {
        body.push(lines[i])
        i++
      }
      i++ // skip closing fence
      out.push(`<pre><code>${body.join('\n')}</code></pre>`)
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
      !/^```/.test(lines[i]) &&
      !/^(#{1,6})\s/.test(lines[i]) &&
      !/^\s*[-*]\s+/.test(lines[i]) &&
      !/^\s*\d+[.)]\s+/.test(lines[i]) &&
      !/^\s*&gt;\s?/.test(lines[i])
    ) {
      para.push(inline(lines[i].trim()))
      i++
    }
    out.push(`<p>${para.join('<br>')}</p>`)
  }
  closeList()
  return out.join('\n')
}
