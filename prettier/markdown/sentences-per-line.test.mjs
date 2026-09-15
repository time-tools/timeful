import { describe, expect, it } from 'vitest'
import * as prettier from 'prettier'
import * as sentencesPerLine from './sentences-per-line.js'

function format(markdown, options = {}) {
  return prettier.format(markdown, {
    parser: 'markdown',
    plugins: [sentencesPerLine],
    ...options,
  })
}

describe('table-safe sentences-per-line formatter', () => {
  it('splits prose sentences onto separate physical lines', async () => {
    await expect(
      format('The first sentence ends here. The second sentence follows.\n'),
    ).resolves.toBe(
      'The first sentence ends here.\nThe second sentence follows.\n',
    )
  })

  it('keeps a multi-sentence table cell on its row', async () => {
    const formatted = await format(
      '| Scenario | Requirement |\n| --- | --- |\n| Response | The system shall issue access only after approval. It shall reject every invalid state. |\n',
    )

    expect(formatted).toContain(
      '| Response | The system shall issue access only after approval. It shall reject every invalid state. |',
    )
    expect(
      formatted.split('\n').filter((line) => line.startsWith('|')),
    ).toHaveLength(3)
  })

  it('round-trips the QR-011 table without splitting its response rows', async () => {
    const source = [
      '| Scenario element | Requirement |',
      '| --- | --- |',
      '| Response | The system shall issue target-browser access only after the source browser approves the exact matching code. It shall accept a transfer at most once and only during its five-minute validity period. It shall reject expired, cancelled, reused, code-mismatched, unapproved, and unauthorized transfers, and reject use of a revoked Granted EVCC. Anonymous initiation shall require a valid [Event Visitor Control Credential](../../../terminology/glossary.md#event-visitor-control-credential-evcc), plus an [Event Owner Edit Token](../../../terminology/glossary.md#event-owner-edit-token) when owner authority is requested. |',
      '| Response measure | Automated route and browser tests demonstrate a successful approved transfer and rejection for every invalid transfer state, including source revocation. |',
      '',
    ].join('\n')

    const once = await format(source)
    const twice = await format(once)

    expect(twice).toBe(once)
    expect(
      once.split('\n').filter((line) => line.startsWith('|')),
    ).toHaveLength(4)
    expect(once).toContain(
      'five-minute validity period. It shall reject expired',
    )
    expect(once).toMatch(/including source revocation\. +\|/)
  })

  it('passes custom abbreviations through to the upstream formatter', async () => {
    await expect(
      format('The memo cites Dr. Wu. It records the decision.\n', {
        sentencesPerLineAdditionalAbbreviations: ['Dr.'],
      }),
    ).resolves.toBe('The memo cites Dr. Wu.\nIt records the decision.\n')
  })

  it('splits prose when a sentence ends before an inline link', async () => {
    await expect(
      format('First sentence ends here. [The link](#x) starts the next.\n'),
    ).resolves.toBe(
      'First sentence ends here.\n[The link](#x) starts the next.\n',
    )
  })

  it('splits the joined Timed Domain Mode glossary entry and stays idempotent', async () => {
    const source =
      "The persisted configuration that determines how a [Timed Event's](#timed-event) set of [Active Slots](#active-slots) is maintained. [Ranged Domain Mode](#ranged-domain-mode) configures it from an [Active Slot Range](#active-slot-range); [Custom Domain Mode](#custom-domain-mode) configures it through [Custom Domain Editing](#custom-domain-editing).\n"
    const once = await format(source)
    const twice = await format(once)

    expect(once).toBe(
      "The persisted configuration that determines how a [Timed Event's](#timed-event) set of [Active Slots](#active-slots) is maintained.\n[Ranged Domain Mode](#ranged-domain-mode) configures it from an [Active Slot Range](#active-slot-range); [Custom Domain Mode](#custom-domain-mode) configures it through [Custom Domain Editing](#custom-domain-editing).\n",
    )
    expect(twice).toBe(once)
  })

  it('splits prose when a sentence ends inside a link label', async () => {
    await expect(format('See [the docs.](#d) Then go.\n')).resolves.toBe(
      'See [the docs.](#d)\nThen go.\n',
    )
  })

  it('splits prose around emphasis boundaries on both sides', async () => {
    await expect(
      format('First sentence ends here. **Bold start** of the next.\n'),
    ).resolves.toBe('First sentence ends here.\n**Bold start** of the next.\n')
    await expect(
      format('Bold end here. **Bold end.** Next sentence follows.\n'),
    ).resolves.toBe('Bold end here.\n**Bold end.**\nNext sentence follows.\n')
  })

  it('handles multiple gap and inline boundaries in one paragraph', async () => {
    await expect(
      format(
        'Alpha ends. Beta continues. [Gamma](#g) starts. Delta follows.\n',
      ),
    ).resolves.toBe(
      'Alpha ends.\nBeta continues.\n[Gamma](#g) starts.\nDelta follows.\n',
    )
  })

  it('keeps table rows intact when a cell holds an inline-structure boundary', async () => {
    const formatted = await format(
      '| Scenario | Requirement |\n| --- | --- |\n| Response | First sentence ends here. [Ranged Domain Mode](#ranged) configures the rest. |\n',
    )

    expect(
      formatted.split('\n').filter((line) => line.startsWith('|')),
    ).toHaveLength(3)
    expect(formatted).toContain(
      'First sentence ends here. [Ranged Domain Mode](#ranged) configures the rest.',
    )
  })

  it('does not split after digit periods', async () => {
    await expect(format('It costs 3. Next sentence.\n')).resolves.toBe(
      'It costs 3. Next sentence.\n',
    )
    await expect(
      format('Version 3. **Next section** starts here.\n'),
    ).resolves.toBe('Version 3. **Next section** starts here.\n')
  })

  it('does not split after known or custom abbreviations at structure gaps', async () => {
    await expect(format('Cite Dr. [Wu](#w) said.\n')).resolves.toBe(
      'Cite Dr. [Wu](#w) said.\n',
    )
    await expect(
      format('Cite Xu. [Wang](#w) said.\n', {
        sentencesPerLineAdditionalAbbreviations: ['Xu.'],
      }),
    ).resolves.toBe('Cite Xu. [Wang](#w) said.\n')
  })

  it('does not split at semicolons or lowercase continuations', async () => {
    await expect(format('He said; She said.\n')).resolves.toBe(
      'He said; She said.\n',
    )
    await expect(format('First. mid sentence continues.\n')).resolves.toBe(
      'First. mid sentence continues.\n',
    )
    await expect(format('First. [link](#x) more words.\n')).resolves.toBe(
      'First. [link](#x) more words.\n',
    )
  })

  it('never splits on punctuation inside inline code but splits at code edges', async () => {
    await expect(format('Use `a. Then` mode.\n')).resolves.toBe(
      'Use `a. Then` mode.\n',
    )
    await expect(format('Passes here. `npm test` fails.\n')).resolves.toBe(
      'Passes here. `npm test` fails.\n',
    )
    await expect(format('Passes here. `npm test` Fails hard.\n')).resolves.toBe(
      'Passes here.\n`npm test` Fails hard.\n',
    )
  })

  it('does not split at hard-break siblings', async () => {
    const source = 'First sentence.  \nSecond starts here.\n'
    await expect(format(source)).resolves.toBe(source)
  })
})
