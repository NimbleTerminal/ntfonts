# CJK CHANGELOG

## Unreleased

## v0.3.1 - 2026-11-04

- Add `Source.Fallbacks()` on `cjk`, `scregular` and `scbold` so renderers can
  discover regional fallback families through the source value, matching the
  core, `packs`, `emoji` and `sbix` modules. Font payloads are unchanged.

## v0.3.0 - 2026-11-04

- Add `Files()` and `WebFaces()` to `cjk`, `scregular` and `scbold` so hosts
  can serve the SC faces to a browser terminal through the same path as the
  script packs (`nimbleterminal.dev/ntfonts/packs`).
- Version jumps from v0.1.0 to v0.3.0 to match the core module: all ntfonts
  modules now release together under one version. Font payloads are unchanged.

## v0.1.0 - 2026-10-29

- Initial release of the independent `nimbleterminal.dev/ntfonts/cjk` module.
- Provide SC Regular and Bold as isolated `scregular` and `scbold` packages,
  embedding 4,050,340 and 4,080,980 font bytes respectively. Each leaf serves
  its single weight for all four variants. The combined `cjk` package serves
  both weights (8,131,320 bytes), with upright italic aliases.
- Retain GB 2312's 6,763 Han characters plus 漢, kana, Hangul, and CJK
  punctuation from pinned Noto Sans CJK SC sources. Share OFL notices without
  linking extra fonts, and verify coverage, hashes, and binary link isolation.

The repository tag for this release is `cjk/v0.1.0`; the core module's
`v0.2.0` tag is independent. Font payloads are unchanged from the pre-release
subsets. SC regional shapes do not substitute for future JP, TC, or KR faces.
