# Safari-compatible color emoji: the `packs/emoji/sbix` module

Date: 2026-11-04. Status: approved by the brief; format and placement are the
author's calls after a verified spike (below).

## Intent

`nimbleterminal.dev/ntfonts/packs/emoji` v0.3.0 embeds the unmodified Noto
Color Emoji CBDT/CBLC font. Chrome and Firefox render it; Safari (WebKit on
macOS and iOS) rejects CBDT, so the ntemoji browser font demo can only disable
the Noto option there. We want actual Noto artwork in Safari, from the same
pinned source, selected by feature detection rather than user-agent sniffing.

## Spike findings (verified 2026-11-04 on macOS 27.2, WebKit 605.1.15 via WKWebView, Chrome 157)

- An `sbix` font derived from the pinned CBDT file (same PNG bitmaps, same
  cmap and GSUB) renders in WebKit and in Chrome, with skin tones, flags, ZWJ
  families, keycaps, VS16 and tag-sequence flags shaped correctly.
- WebKit reports `CSS.supports("font-tech(color-sbix)")` true and
  `font-tech(color-CBDT)` false; Chrome reports both true. The CBDT face
  errors in WebKit (`FontFace.status === "error"`). This is the selection
  signal.
- Chrome renders nothing for sbix glyphs whose `glyf` entry is empty; it needs
  a bounding box. WebKit renders either. A filled rectangle makes renderers
  without sbix draw black boxes; two coincident rectangles with opposite
  winding give the bounding box with zero fill and read as tofu (hairline box)
  where sbix is unsupported. That is the chosen `glyf` shape.
- The build is deterministic: two runs produced identical bytes.
- The CBDT font maps U+0020 to a 2550-unit advance; that is upstream and
  identical in both formats. Consumers put the emoji family after a text font
  or restrict `unicodeRange`.

## Decisions

1. **Format: sbix** (Apple's bitmap format, PNG strikes). One strike at the
   CBDT strike's ppem (109) and 72 ppi. Every glyph with a CBDT bitmap gets
   the identical PNG bytes. No resampling: artwork is byte-identical.
   *Amended during execution:* the origin offsets are measured from the
   lower-left corner of the glyph's `glyf` box, not from the glyph origin.
   FreeType (`ttgload.c`, sbix branch) and CoreText shift sbix bitmaps by the
   box corner when a glyph has contours, and Apple Color Emoji ships zero
   offsets over a baseline box. Because the box is the bitmap's own rectangle,
   the offsets are `BearingX - px(xMin)` and `(BearingY - height) - px(yMin)`,
   both zero for Noto. The first build with origin-relative offsets rendered
   27 px too low in WebKit, Chrome and FreeType alike.
2. **Separate module** `nimbleterminal.dev/ntfonts/packs/emoji/sbix`, package
   `sbix`. Keeping it inside `packs/emoji` would double that module's zip for
   every consumer. The emoji module and its API are unchanged except for one
   additive field.
3. **File name** `NotoColorEmoji-sbix.ttf` on disk and in `Files()`/`WebFaces()`.
   The name table says `Noto Color Emoji sbix` (IDs 1, 4), PostScript
   `NotoColorEmoji-sbix` (6), unique ID and description note the conversion
   (3, 10). Noto's OFL declares no Reserved Font Names, so the rename is a
   courtesy that keeps the modified file distinguishable, not a requirement.
   The API family key stays `Noto Color Emoji` so consumers select the same
   family in either format.
4. **Format metadata**: `WebFace` in both `emoji` and `sbix` gains
   `Tech string`, the CSS `font-tech()` keyword (`color-CBDT`, `color-sbix`).
   Consumers test `CSS.supports("font-tech(" + tech + ")")`, or declare
   `src: url(...) format("truetype") tech(...)`. Outline-font modules keep
   their `WebFace` types unchanged.
5. **Generation**: `scripts/subset-fonts.py --emoji` builds both modules from
   the hash-verified cached `NotoColorEmoji.ttf` with `fonttools==4.60.2`,
   writing `packs/emoji/sbix/{NotoColorEmoji-sbix.ttf, notocoloremoji.OFL.txt,
   subsets.sha256}`. `packs/emoji/sbix/sources.json` is a checked-in copy of
   the emoji sources. No runtime downloads; no other inputs.
6. **Tests** (standard library only): in `packs/emoji/sbix` (self-contained, no
   parent checkout required): table presence (`sbix`, `glyf`, `loca`; no
   `CBDT`, `CBLC`, `COLR`, `CPAL`, `SVG `); 3,985 136x128 PNG strikes at 109
   ppem/72 ppi with zero offsets; `glyf` boxes with two opposite-winding
   contours and `maxp` limits; name table; `ccmp` and a format 14 cmap
   subtable; faces/aliases/errors; `Files()`, `WebFaces()` with `Tech`;
   licenses; pinned hash; 12,000,000-byte budget. At the repository root:
   cross-font derivation check against `packs/emoji/NotoColorEmoji.ttf`
   (identical `cmap`, `GSUB`, `hmtx`, `hhea` line metrics, `vhea`, `vmtx`,
   `OS/2`, `post`; same glyph count; per-glyph PNG identity, scaled `glyf`
   boxes and origin offsets).
7. **Specimen viewer** catalogs the sbix file as a second `Noto Color Emoji`
   face with a `tech` field, picks the first face whose tech the browser
   supports, loads it lazily (existing behavior), and reports when no format
   is supported. This makes the deployed Pages site a live Safari check.
8. **Release harmony** extends to five modules: tag `packs/emoji/sbix/vX.Y.Z`.
9. **Fallback** when neither format is supported: load nothing and let the
   platform emoji font render. The integration example does exactly that.

## Non-goals

- COLRv1 or OT-SVG builds (WebKit lacks COLRv1; Chrome lacks OT-SVG; neither
  uses the bitmap artwork).
- Resized strikes or WOFF2 packaging (PNG data does not compress; resampling
  would alter the artwork).
- Native use of the sbix file. go-text parses sbix, but nttypeset consumers
  should keep the unmodified CBDT module.

## Known limitations to document

- Tested only in WebKit (WKWebView, the Safari engine) and Chrome on macOS.
  Firefox and FreeType-based Chrome builds (Linux, Windows) are untested
  here unless the FreeType check in the plan runs.
- WebKit's fallback picks the system emoji font for sequences whose base
  character the preceding text font covers (keycap digits, `❤️`, `✈️`) when
  the emoji family is listed after a text font. Chrome routes them to Noto.
- Noto glyphs advance 2550/2048 = 1.245 em; Apple Color Emoji is narrower.
  Line layout differs from the platform font by design.
- The file is 10.8 MB; consumers must load it lazily.
