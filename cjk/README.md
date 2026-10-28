# Optional CJK fonts

`nimbleterminal.dev/ntfonts/cjk` is an independent, standard-library-only Go
module. The parent ntfonts module does not import it and its module zip excludes
this directory. Importing this module embeds both faces below, adding about
8.13 MB of raw font data to a native or js/wasm binary (compression may reduce
transfer size). It does not discover or download fonts at runtime.

```sh
go get nimbleterminal.dev/ntfonts/cjk
```

```go
import "nimbleterminal.dev/ntfonts/cjk"

// Register the regional family with your renderer's fallback mechanism.
// Handle err in application code.
font, err := cjk.Embedded.Face(cjk.Fallbacks()[0], "Regular")
```

| Family | Variant | Bytes |
| --- | --- | ---: |
| Noto Sans CJK SC | Regular | 4,050,340 |
| Noto Sans CJK SC | Bold | 4,080,980 |
| Total | | 8,131,320 |

Italic aliases Regular; BoldItalic aliases Bold. These are upright fonts, not
synthetic italics. `Licenses()` returns the bundled OFL text. Font internals
are renamed to Gloss Text CJKSC.

SC uses Simplified Chinese regional glyph shapes, including for shared Han
characters. JP, TC and KR are future, separate families; they must not alias
SC. Hosts should choose a region matching their content. Kana and Hangul
coverage does not make SC a Japanese or Korean regional face.

## Coverage and reproducibility

Han is restricted to the 6,763 GB 2312 characters in
[data/gb2312-han.txt](data/gb2312-han.txt), plus the explicit extra 漢 (U+6F22)
for the mixed-language ruby example in the upstream request. U+20000 is absent.
Also retained where present upstream: ASCII, CJK punctuation U+3000–303F,
hiragana/katakana U+3040–30FF, halfwidth/fullwidth forms U+FF00–FFEF, Hangul
syllables U+AC00–D7A3, and jamo U+1100–11FF, U+3130–318F, U+A960–A97F,
U+D7B0–D7FF. Hangul remains in SC; the combined size is reported above.

Sources are Noto Sans CJK Sans2.004, pinned to commit
`523d033d6cb47f4a80c58a35753646f5c3608a78`. `sources.json` records each upstream
URL and SHA-256; `subsets.sha256` records generated OTF hashes. The shared
script uses `fonttools==4.60.2`, strips hints, preserves ccmp/locl/mark/mkmk/kern,
renames the fonts, and fixes timestamps. From the repository root:

```sh
python scripts/subset-fonts.py /tmp/ntfonts-sources --download --cjk
task --dir cjk notices
task --dir cjk ci
```

The root `task ci` includes this module's gate, including js/wasm compilation.
Automated cmap coverage checks do not establish visual quality. No visual
quality check has been performed for this expansion.

Source code: MIT (LICENSE). Fonts: OFL-1.1 (notosanscjksc.OFL.txt).
Release tags for this nested module have the form `cjk/vX.Y.Z`.
