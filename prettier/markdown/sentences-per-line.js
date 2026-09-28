import * as base from 'prettier-plugin-sentences-per-line'

// The two passes below are additive to upstream, which already leaves table
// rows alone, so they have to skip tables themselves.
const tableNodeTypes = new Set(['table', 'tableRow', 'tableCell'])
const gapNodeTypes = new Set(['paragraph', 'heading'])
const breakNodeType = 'break'
const textLikeNodeTypes = new Set(['text', 'sentence'])
const inlineStructurePlaceholder = '\uE000'
const nextLetterPattern = /^[A-Z]/
const whitespacePattern = /\s/
const closingPairsPattern = /[)"'”’\]]+$/
const trailingWhitespacePattern = /\s+$/
const digitPeriodPattern = /\d\.$/
// Deliberately identical to digitPeriodPattern but with the opposite meaning:
// digitPeriodPattern suppresses a break, this one requires one. Keep them
// separate so that narrowing either does not silently change the other.
const numericSentenceEndPattern = /\d\.$/
const upstreamIgnoredWords = [
  'eg.',
  'e.g.',
  'etc.',
  'ex.',
  'ie.',
  'i.e.',
  'vs.',
]

let sentenceSegmenter

function getSentenceSegmenter() {
  if (sentenceSegmenter === undefined) {
    sentenceSegmenter =
      typeof Intl !== 'undefined' && typeof Intl.Segmenter === 'function'
        ? new Intl.Segmenter('en', { granularity: 'sentence' })
        : null
  }
  return sentenceSegmenter
}

function hasPosition(node) {
  const { position } = node
  return (
    Number.isInteger(position?.start?.offset) &&
    Number.isInteger(position?.end?.offset) &&
    position.end.offset >= position.start.offset
  )
}

function collectBlankRanges(node, ranges) {
  if (!hasPosition(node)) {
    for (const child of node.children ?? []) collectBlankRanges(child, ranges)
    return
  }
  if (textLikeNodeTypes.has(node.type)) return

  const { position } = node
  const start = position.start.offset
  const end = position.end.offset
  const children = Array.isArray(node.children) ? node.children : []
  if (children.length === 0) {
    ranges.push([start, end])
    return
  }

  let firstChildStart
  let lastChildEnd
  for (const child of children) {
    if (hasPosition(child)) {
      if (
        firstChildStart === undefined ||
        child.position.start.offset < firstChildStart
      ) {
        firstChildStart = child.position.start.offset
      }
      if (
        lastChildEnd === undefined ||
        child.position.end.offset > lastChildEnd
      ) {
        lastChildEnd = child.position.end.offset
      }
    }
    collectBlankRanges(child, ranges)
  }
  if (firstChildStart === undefined) {
    ranges.push([start, end])
    return
  }
  if (firstChildStart > start) ranges.push([start, firstChildStart])
  if (end > lastChildEnd) ranges.push([lastChildEnd, end])
}

function endsWithSuppressedTail(tail, customAbbreviations) {
  if (digitPeriodPattern.test(tail)) return true
  const lowered = tail.toLowerCase()
  return [...upstreamIgnoredWords, ...customAbbreviations].some((word) =>
    lowered.endsWith(word.toLowerCase()),
  )
}

function trimBoundaryWhitespace(previous, next) {
  if (previous.type === 'sentence' && Array.isArray(previous.children)) {
    while (
      previous.children.length > 0 &&
      previous.children[previous.children.length - 1].type === 'whitespace'
    ) {
      previous.children.pop()
    }
  }
  if (next.type === 'sentence' && Array.isArray(next.children)) {
    while (next.children.length > 0 && next.children[0].type === 'whitespace') {
      next.children.shift()
    }
  }
}

function clamp(value, minimum, maximum) {
  return Math.min(Math.max(value, minimum), maximum)
}

function insertParagraphGapBreaks(node, options) {
  const segmenter = getSentenceSegmenter()
  if (!segmenter) return
  const children = node.children
  if (!Array.isArray(children) || children.length < 2) return
  if (!hasPosition(node) || typeof options.originalText !== 'string') return

  const nodeStart = node.position.start.offset
  const nodeEnd = node.position.end.offset
  const raw = options.originalText.slice(
    clamp(nodeStart, 0, options.originalText.length),
    clamp(nodeEnd, 0, options.originalText.length),
  )
  const blankRanges = []
  for (const child of children) collectBlankRanges(child, blankRanges)
  const placeholderMask = new Uint8Array(raw.length)
  let source = raw
  for (const [rangeStart, rangeEnd] of blankRanges) {
    const localStart = clamp(rangeStart - nodeStart, 0, source.length)
    const localEnd = clamp(rangeEnd - nodeStart, 0, source.length)
    if (localEnd <= localStart) continue
    placeholderMask.fill(1, localStart, localEnd)
    source =
      source.slice(0, localStart) +
      inlineStructurePlaceholder.repeat(localEnd - localStart) +
      source.slice(localEnd)
  }

  const segments = []
  for (const segment of segmenter.segment(source)) {
    segments.push({
      start: segment.index,
      end: segment.index + segment.segment.length,
    })
  }
  const nextRealStart = (from) => {
    let index = from
    while (index < source.length && placeholderMask[index]) index += 1
    return index < source.length ? index : -1
  }
  const nextLetterAfterBoundary = (from) => {
    const realStart = nextRealStart(from)
    if (realStart === -1) return ''
    let index = realStart
    while (index < source.length && whitespacePattern.test(source[index]))
      index += 1
    return index < source.length ? source[index] : ''
  }
  const proseTailBefore = (from) => {
    let tail = ''
    for (let index = 0; index < from; index += 1) {
      if (!placeholderMask[index]) tail += source[index]
    }
    return tail
      .replace(trailingWhitespacePattern, '')
      .replace(closingPairsPattern, '')
  }

  const customAbbreviations = Array.isArray(
    options.sentencesPerLineAdditionalAbbreviations,
  )
    ? options.sentencesPerLineAdditionalAbbreviations
    : []
  const insertionIndexes = []
  for (let index = 0; index < segments.length - 1; index += 1) {
    const boundaryEnd = segments[index].end
    const nextSegmentStart = nextRealStart(segments[index + 1].start)
    if (nextSegmentStart === -1) continue
    let childIndex = -1
    for (let i = children.length - 1; i >= 0; i -= 1) {
      const childStart = children[i].position?.start?.offset
      if (
        Number.isInteger(childStart) &&
        childStart - nodeStart < boundaryEnd
      ) {
        childIndex = i
        break
      }
    }
    if (childIndex === -1 || childIndex === children.length - 1) continue
    const previous = children[childIndex]
    const next = children[childIndex + 1]
    if (previous.type === breakNodeType || next.type === breakNodeType) continue
    if (!hasPosition(next)) continue
    const nextLocalStart = next.position.start.offset - nodeStart
    if (boundaryEnd > nextLocalStart || nextSegmentStart < nextLocalStart)
      continue
    if (
      !nextLetterPattern.test(
        nextLetterAfterBoundary(segments[index + 1].start),
      )
    )
      continue
    if (raw.slice(boundaryEnd, nextSegmentStart).includes('\n')) continue
    if (
      endsWithSuppressedTail(proseTailBefore(boundaryEnd), customAbbreviations)
    )
      continue
    insertionIndexes.push(childIndex)
  }

  for (const childIndex of insertionIndexes.reverse()) {
    trimBoundaryWhitespace(children[childIndex], children[childIndex + 1])
    children.splice(childIndex + 1, 0, { type: 'sentenceBreak' })
  }
}

function splitNumericSentenceEnds(node) {
  if (tableNodeTypes.has(node.type)) return
  if (node.type === 'sentence') splitNumericSentenceEndsInSentence(node)
  for (const child of node.children ?? []) splitNumericSentenceEnds(child)
}

function splitNumericSentenceEndsInSentence(sentence) {
  const children = sentence.children
  if (!Array.isArray(children)) return

  for (let index = children.length - 2; index >= 1; index -= 1) {
    const child = children[index]
    // Upstream skips any word node matching /^\s*\d+\./ to avoid breaking list
    // markers, but a list marker is structural and never reaches this stream, so
    // the same test also swallows every genuine sentence that ends in a number,
    // such as "Week 1.". Split those here instead.
    if (child.type !== 'word' || !numericSentenceEndPattern.test(child.value))
      continue
    const whitespace = children[index + 1]
    if (whitespace?.type !== 'whitespace') continue
    const next = children[index + 2]
    if (next?.type !== 'word' || !nextLetterPattern.test(next.value)) continue

    children.splice(index + 1, 1)
    children.splice(index + 1, 0, { type: 'sentenceBreak' })
  }
}

function insertGapSentenceBreaks(node, options) {
  if (tableNodeTypes.has(node.type)) return
  if (gapNodeTypes.has(node.type)) insertParagraphGapBreaks(node, options)
  for (const child of node.children ?? [])
    insertGapSentenceBreaks(child, options)
}

export const options = base.options
export const parsers = base.parsers
export const printers = {
  mdast: {
    ...base.printers.mdast,
    async preprocess(ast, options) {
      const processed = await base.printers.mdast.preprocess(ast, options)
      insertGapSentenceBreaks(processed, options)
      splitNumericSentenceEnds(processed)
      return processed
    },
  },
}
