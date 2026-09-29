const knownAbbreviations = new Set([
  'mr',
  'mrs',
  'ms',
  'dr',
  'prof',
  'sr',
  'jr',
  'st',
  'mt',
  'vs',
  'etc',
  'cf',
  'al',
  'approx',
  'fig',
  'inc',
  'ltd',
])

const closingPairsPattern = /[)"'”’\]*_]+$/
const terminalPunctuationPattern = /[.!?]$/
const digitPeriodPattern = /\d\.$/
const initialismPattern = /(?:^|[\s("'])(?:[A-Za-z]\.)+$/
const finalWordPeriodPattern = /([A-Za-z]+)\.$/
const trailingPaddingPattern = /[ \t]+$/
const leadingPaddingPattern = /^[ \t]+/
const backslashHardBreakPattern = /\\$/
const spacingHardBreakPattern = /[ \t]{2,}$/
const hyphenTailPattern = /-$/
const pipeTailPattern = /\|$/
const pipeHeadPattern = /^\|/

function isHardBreak(rawTail) {
  return backslashHardBreakPattern.test(rawTail) || spacingHardBreakPattern.test(rawTail)
}

function analyzeTail(rawTail) {
  if (!rawTail || isHardBreak(rawTail)) {
    return null
  }
  const tail = rawTail.replace(trailingPaddingPattern, '')
  if (!tail) {
    return null
  }
  const stripped = tail.replace(closingPairsPattern, '')
  if (!terminalPunctuationPattern.test(stripped)) {
    return { tailLength: tail.length }
  }
  if (digitPeriodPattern.test(stripped)) {
    return null
  }
  if (initialismPattern.test(stripped)) {
    return { tailLength: tail.length }
  }
  const wordMatch = stripped.match(finalWordPeriodPattern)
  if (wordMatch && knownAbbreviations.has(wordMatch[1].toLowerCase())) {
    return { tailLength: tail.length }
  }
  return null
}

export default {
  meta: {
    type: 'layout',
    docs: {
      description:
        'Disallow Markdown prose sentences that continue across physical source lines',
      recommended: false,
    },
    fixable: 'whitespace',
    messages: {
      sentenceContinues:
        'Expected the sentence to end on this line, but it continues on the next physical line.',
    },
    schema: [],
  },
  create(context) {
    const { sourceCode } = context
    return {
      paragraph(node, parent) {
        if (!parent || parent.type !== 'root') {
          return
        }
        const base = node.position.start.offset
        const raw = sourceCode.text.slice(base, node.position.end.offset)
        let lineStart = 0
        for (;;) {
          const newline = raw.indexOf('\n', lineStart)
          if (newline === -1) {
            break
          }
          const nextNewline = raw.indexOf('\n', newline + 1)
          const rawTail = raw.slice(lineStart, newline)
          const head =
            nextNewline === -1 ? raw.slice(newline + 1) : raw.slice(newline + 1, nextNewline)
          const analysis = analyzeTail(rawTail)
          if (analysis) {
            const trimmedTail = rawTail.replace(trailingPaddingPattern, '')
            const trimmedHead = head.trim()
            if (trimmedTail && trimmedHead) {
              const tailEnd = base + lineStart + trimmedTail.length
              const headLead = head.match(leadingPaddingPattern)?.[0].length ?? 0
              const headStart = base + newline + 1 + headLead
              const fixable =
                !hyphenTailPattern.test(trimmedTail) &&
                !pipeTailPattern.test(trimmedTail) &&
                !pipeHeadPattern.test(trimmedHead)
              context.report({
                loc: {
                  start: sourceCode.getLocFromIndex(tailEnd),
                  end: sourceCode.getLocFromIndex(Math.min(headStart, tailEnd + 1)),
                },
                messageId: 'sentenceContinues',
                ...(fixable
                  ? {
                      fix(fixer) {
                        return fixer.replaceTextRange([tailEnd, headStart], ' ')
                      },
                    }
                  : {}),
              })
            }
          }
          lineStart = newline + 1
        }
      },
    }
  },
}
