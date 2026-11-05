# packs/emoji/sbix CHANGELOG

## Unreleased

## v0.3.1 - 2026-11-04

- Initial `nimbleterminal.dev/ntfonts/packs/emoji/sbix` module embedding
  `NotoColorEmoji-sbix.ttf` (10,845,476 bytes): the pinned Noto Color Emoji
  v2.051 bitmaps as an sbix font for WebKit, derived deterministically by
  `scripts/subset-fonts.py --emoji`, served for all four variants as
  `Noto Color Emoji` with `Files()`, `WebFaces()` (`Tech: "color-sbix"`) and
  `Licenses()`.
- First release is v0.3.1, matching the core module: all five ntfonts modules
  release together under one version. The repository tag is
  `packs/emoji/sbix/v0.3.1`.
