# packs CHANGELOG

## Unreleased

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
