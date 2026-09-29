# Local Markdown sentences-per-line formatter

This wrapper adds two corrections to the upstream `prettier-plugin-sentences-per-line` plugin and delegates everything else, including its options, to it.
The dependency floor is `^0.2.5` because the wrapper no longer repairs table rows itself.
Upstream leaves table rows intact since 0.2.5, and both corrections below skip tables, because a Markdown table requires every row to occupy one physical source line.

The wrapper inserts a break after any sentence that ends in a number, because upstream skips every word node matching `/^\s*\d+\./` in order to protect list markers.
A list marker is structural and never reaches the word stream, so that test also swallows genuine sentence ends such as `Week 1.` and leaves them unsplit.
Interior decimals such as `Apache 2.0` and `version 14.14` are unaffected, because the trailing period they need is not the end of the word.
Without this correction the formatter and the `sentences-per-line/one` lint rule disagree on every sentence that ends in a number, because the lint rule has no such guard and reports those lines.

The wrapper also inserts a break where a sentence ends at the edge of an inline structure, such as directly before a link or directly after emphasis.
Upstream scans the `word` children of a single `sentence` node, and Prettier's Markdown preprocess turns each `text` run into its own `sentence` node, so a sentence boundary that falls between two runs is invisible to that scan.
To find those boundaries the wrapper re-segments the raw source of each paragraph and heading with `Intl.Segmenter` and masks the ranges spanned by inline structures, so that brackets, emphasis markers, and URLs do not hide the sentence end.
Abbreviation suppression at those edges comes from the upstream `sentencesPerLineAdditionalAbbreviations` option, which the wrapper passes through unchanged.

`scripts/markdown.mjs` invokes Prettier through its JavaScript API because Prettier's CLI loads its built-in Markdown printer after configured plugins and ignores sentence-per-line printer overrides.
