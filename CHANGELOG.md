# ntfonts CHANGELOG

## Unreleased

- Export a static specimen site with `task specimen-build`; publish GitHub
  Pages from `main`, `v*`, and `cjk/v*` pushes after CI passes.

- Add `task specimen`: a local browser viewer with editable samples, script
  coverage, variant/size controls, and explicit symbol/CJK fallback comparisons.

- Widen text subsets to upstream Latin Extended/Vietnamese, Greek, Cyrillic,
  arrows, math, box drawing, geometric shapes and check marks. Default font
  bytes grow from 1,202,648 to 3,968,464, including the symbol fallback, below 4 MB.
- Add `Fallbacks()` and Noto Sans Symbols 2 to `Embedded.Face` and `Licenses()`;
  all variants alias its Regular face, supplemented with missing Noto Sans Math
  arrows/operators. `Families()` is unchanged.
- Add independent opt-in module `nimbleterminal.dev/ntfonts/cjk` with regional
  Noto Sans CJK SC Regular/Bold, GB 2312 Han plus 漢, kana and Hangul, totaling
  8,131,320 font bytes. Core imports and module downloads exclude CJK.

## v0.1.0 - 2026-10-18

- Extracted from gloss: seven OFL font families as 28 Latin subset faces, served by `Embedded`, with `Families` and `Licenses`.
