import { describe, expect, it } from 'vitest'
import { Linter } from 'eslint'
import markdown from '@eslint/markdown'
import noSplitSentence from './no-split-sentence.js'

const linter = new Linter({ configType: 'flat' })

const config = [
  {
    files: ['**/*.md'],
    plugins: {
      markdown,
      local: {
        rules: {
          'no-split-sentence': noSplitSentence,
        },
      },
    },
    language: 'markdown/gfm',
    languageOptions: {
      frontmatter: 'yaml',
    },
    rules: {
      'local/no-split-sentence': 'error',
    },
  },
]

function verify(code) {
  return linter.verify(code, config, 'fixture.md')
}

function fix(code) {
  return linter.verifyAndFix(code, config, 'fixture.md')
}

describe('no-split-sentence', () => {
  it('reports a prose sentence that continues across a physical line and fixes it', () => {
    const code =
      'The deployment finished early\nand the queue drained cleanly.\n'

    const messages = verify(code)
    expect(messages).toHaveLength(1)
    expect(messages[0].ruleId).toBe('local/no-split-sentence')
    expect(messages[0].messageId).toBe('sentenceContinues')
    expect(messages[0].fix).toBeDefined()

    const result = fix(code)
    expect(result.fixed).toBe(true)
    expect(result.messages).toHaveLength(0)
    expect(result.output).toBe(
      'The deployment finished early and the queue drained cleanly.\n',
    )
  })

  it('does not report lines that end with sentence-terminal punctuation', () => {
    const code =
      'First line ends here.\nSecond ends with energy!\nThird asks nicely?\nFourth quotes "the term."\n'

    expect(verify(code)).toHaveLength(0)
  })

  it('reports abbreviation endings as continuations and fixes them', () => {
    const code = 'The memo cites Dr.\nWu and the appendix.\n'

    const messages = verify(code)
    expect(messages).toHaveLength(1)
    expect(messages[0].fix).toBeDefined()

    const result = fix(code)
    expect(result.output).toBe('The memo cites Dr. Wu and the appendix.\n')
  })

  it('reports initialism endings as continuations and fixes them', () => {
    const code = 'The spec references U.S.\nfederal guidance today.\n'

    const messages = verify(code)
    expect(messages).toHaveLength(1)

    const result = fix(code)
    expect(result.output).toBe('The spec references U.S. federal guidance today.\n')
  })

  it('treats digit periods as sentence ends without reporting', () => {
    const code = 'Version 2.\nShips today.\n'

    expect(verify(code)).toHaveLength(0)
  })

  it('reports one diagnostic per soft-wrap boundary and joins all of them', () => {
    const code = 'Alpha began here\nbeta middle chunk\ngamma wrapped up.\n'

    const messages = verify(code)
    expect(messages).toHaveLength(2)

    const result = fix(code)
    expect(result.output).toBe('Alpha began here beta middle chunk gamma wrapped up.\n')
  })

  it('reports but does not autofix boundaries with unsafe edge characters', () => {
    const code = 'A well-\nknown fact appears.\n'

    const messages = verify(code)
    expect(messages).toHaveLength(1)
    expect(messages[0].fix).toBeUndefined()

    const result = fix(code)
    expect(result.fixed).toBe(false)
    expect(result.output).toBe(code)
  })

  it('excludes structural Markdown content from reporting', () => {
    const structural = [
      '---',
      'title: Wrapped',
      'prose sample',
      '---',
      '',
      '# Heading text that',
      'continues below',
      '',
      'Fenced code sample:',
      '',
      '```js',
      'const first = 1',
      'const second = 2',
      '```',
      '',
      '    Indented code line one',
      '    indented code line two continues',
      '',
      '| Column | Value |',
      '| ------ | ----- |',
      '| alpha wrapped | one |',
      '',
      '- Wrapped bullet item',
      '  continuing on the next line',
      '- Parent item',
      '  - Nested wrapped item',
      '    continuing deeper',
      '',
      '1. Ordered wrapped item',
      '   continues here',
      '',
      '- [ ] Task item wrapped',
      '  continues too',
      'Lazy unindented continuation',
      'of the preceding list item stays quiet',
      '',
      '> Quoted prose wrapped',
      '> across physical lines',
      '',
      'A definition style term',
      '---',
      '',
      'Double space hard break next  ',
      'still excluded here',
      '',
      'Backslash hard break next\\',
      'also excluded entirely',
      '',
    ].join('\n')

    expect(verify(structural)).toHaveLength(0)
  })

  it('reports a tail ending before a link head and fixes it', () => {
    const code =
      'The outcomes are defined by\n[FR-070](https://example.com/fr-070), and [FR-071](https://example.com/fr-071).\n'

    const messages = verify(code)
    expect(messages).toHaveLength(1)
    expect(messages[0].fix).toBeDefined()

    const result = fix(code)
    expect(result.fixed).toBe(true)
    expect(result.output).toBe(
      'The outcomes are defined by [FR-070](https://example.com/fr-070), and [FR-071](https://example.com/fr-071).\n',
    )
  })

  it('reports a boundary inside a wrapped link label and fixes it', () => {
    const code =
      'See the [minimal template\nfor decisions](https://example.com) before writing.\n'

    const messages = verify(code)
    expect(messages).toHaveLength(1)

    const result = fix(code)
    expect(result.output).toBe(
      'See the [minimal template for decisions](https://example.com) before writing.\n',
    )
  })

  it('reports a boundary inside a wrapped inline code span and fixes it', () => {
    const code = 'Stalwart joins the `timeful-edge\nnetwork` during setup.\n'

    const messages = verify(code)
    expect(messages).toHaveLength(1)

    const result = fix(code)
    expect(result.output).toBe('Stalwart joins the `timeful-edge network` during setup.\n')
  })

  it('reports a closing-paren continuation after a link and fixes it', () => {
    const code =
      'The visitor proves a valid [token](https://example.com/t)\nand signs in safely afterward.\n'

    const messages = verify(code)
    expect(messages).toHaveLength(1)

    const result = fix(code)
    expect(result.output).toBe(
      'The visitor proves a valid [token](https://example.com/t) and signs in safely afterward.\n',
    )
  })

  it('reports a boundary before bold emphasis structure and fixes it', () => {
    const code = '**Available** and\n**If needed** count equally here.\n'

    const messages = verify(code)
    expect(messages).toHaveLength(1)

    const result = fix(code)
    expect(result.output).toBe('**Available** and **If needed** count equally here.\n')
  })

  it('does not report a completed sentence that closes with emphasis markers', () => {
    const code = [
      'Before you look at any product, write one sentence: **whose problem are we solving, and what are they trying to do?**',
      'Not "what are we building" — that comes later, and it is constrained by this sentence.',
      '',
      'The **strong** answer follows.',
      'The *emphasised* answer follows.',
      'The _underlined_ answer follows.',
      '',
    ].join('\n')

    expect(verify(code)).toHaveLength(0)
  })

  it('does not report a completed sentence that closes with a bracket or a quote', () => {
    const code = 'The team agreed on (three options.)\nand ranked them all by cost.\n'

    expect(verify(code)).toHaveLength(0)
  })

  it('reports but never fixes boundaries involving pipe table fragments', () => {
    const code = 'Anonymous initiation requires proof of authority. |\n| Response measure | tests apply.\n'

    const messages = verify(code)
    expect(messages).toHaveLength(1)
    expect(messages[0].fix).toBeUndefined()

    const result = fix(code)
    expect(result.fixed).toBe(false)
    expect(result.output).toBe(code)
  })

  it('reports one diagnostic per boundary across a chain mixing inline structures and fixes all of them', () => {
    const code =
      'Alpha cites [one doc](https://example.com/1)\nthen references `two code`\nand closes with **three** words.\n'

    const messages = verify(code)
    expect(messages).toHaveLength(2)
    for (const message of messages) {
      expect(message.fix).toBeDefined()
    }

    const result = fix(code)
    expect(result.output).toBe(
      'Alpha cites [one doc](https://example.com/1) then references `two code` and closes with **three** words.\n',
    )
  })
})
