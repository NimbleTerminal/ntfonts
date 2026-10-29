# CJK CHANGELOG

## Unreleased

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
