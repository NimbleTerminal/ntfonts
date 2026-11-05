# packs CHANGELOG

## Unreleased

## v0.3.1 - 2026-11-04

- Release harmony with `ntfonts` v0.3.1 (`packs/v0.3.1`). No code or font
  changes since v0.3.0.

## v0.3.0 - 2026-11-04

- Initial `nimbleterminal.dev/ntfonts/packs` module: seven opt-in script
  packages (`arabic`, `hebrew`, `devanagari`, `bengali`, `tamil`, `thai`,
  `khmer`), each embedding Noto Sans Regular and Bold subsets with every
  OpenType layout feature retained, 1,068,724 font bytes in all. Italic
  variants alias the upright weights.
- Combined `packs` source serving all seven families, with `Fallbacks()`,
  `Face()`, `Licenses()` and the shared `License` type.
- `Files()` and `WebFaces()` on every package so hosts can serve the same
  bytes to a browser terminal through ntwebterm's `fontFaces`.
- Sources pinned to google/fonts commit `bd8f81ddb5c74d5c8897b36ad88b440266245103`;
  `subsets.sha256`, `coverage.json` and link-isolation tests verify the embed.
- First release is v0.3.0, matching the core module: all ntfonts modules
  release together under one version. The repository tag is `packs/v0.3.0`.
