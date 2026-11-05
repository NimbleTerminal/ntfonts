# packs/emoji CHANGELOG

## Unreleased

- `WebFace` gains `Tech`, the CSS `font-tech()` keyword (`color-CBDT`), so a
  page can pick between this file and the sbix companion module with
  `CSS.supports` instead of user-agent sniffing.
- Point to `nimbleterminal.dev/ntfonts/packs/emoji/sbix` for Safari.

## v0.3.0 - 2026-11-04

- Initial `nimbleterminal.dev/ntfonts/packs/emoji` module embedding Noto Color
  Emoji unmodified (CBDT bitmap color, 10,673,480 bytes) from
  googlefonts/noto-emoji v2.051, served for all four variants as
  `Noto Color Emoji`, with `Files()` and `WebFaces()` for browser delivery.
- First release is v0.3.0, matching the core module: all ntfonts modules
  release together under one version. The repository tag is `packs/emoji/v0.3.0`.
