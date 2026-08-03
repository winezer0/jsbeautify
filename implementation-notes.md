# Implementation Notes

## Formatter scope

- The formatter is a token-preserving readability tool for minified JavaScript, not a parser or validator.
- It preserves token order and literal text. It does not evaluate, unpack, rename, or add semicolons.
- Unterminated literals and mismatched delimiters fail without producing formatted output.

## Design choices

- The implementation uses a scanner plus printer state machine so memory remains proportional to source and output, without an AST.
- Template literals are kept as opaque tokens. This avoids changing interpolation semantics at the cost of not reindenting their contents.
- The default width is 120 columns. Wrapping occurs only at safe formatting boundaries already represented by punctuation or operators.

## Real asset validation

- The integration test checks both original and formatted jQuery assets from `testdata` with `node --check`.
- This detects truncation and syntax damage from formatting. It does not claim browser DOM behavior equivalence.

## Documentation

- `README.md` is the English default README and `README.zh-CN.md` contains the Chinese documentation.
