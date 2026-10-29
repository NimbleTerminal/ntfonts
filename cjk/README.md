# Optional CJK fonts

`nimbleterminal.dev/ntfonts/cjk` is an independent, standard-library-only Go
module. The parent ntfonts module does not import it and its module zip excludes
this directory. Import `scregular` by default to link only one weight. It does
not discover or download fonts at runtime.

```sh
go get nimbleterminal.dev/ntfonts/cjk/scregular
```

```go
import "nimbleterminal.dev/ntfonts/cjk/scregular"

// Register this source with your renderer's fallback mechanism.
font, err := scregular.Embedded.Face(scregular.Fallbacks()[0], "Regular")
```

| Import path | Embedded weights | Raw font bytes |
| --- | --- | ---: |
| `nimbleterminal.dev/ntfonts/cjk/scregular` | SC Regular only | 4,050,340 |
| `nimbleterminal.dev/ntfonts/cjk/scbold` | SC Bold only | 4,080,980 |
| `nimbleterminal.dev/ntfonts/cjk` | SC Regular + Bold | 8,131,320 |

Each leaf embeds exactly one font and serves that weight for all four variants
(Regular, Italic, Bold, BoldItalic). Regular-only hosts therefore render bold
requests as upright Regular, without needing variant fallback support.
Unknown families or variants remain errors. All three packages expose
`Embedded`, `Source`, `Fallbacks()` and `Licenses()`, with family
`"Noto Sans CJK SC"` and the same shared `License` type.

Import `cjk` itself when you want both weights: its existing API is unchanged,
with Regular/Italic using Regular and Bold/BoldItalic using Bold. It imports
the two leaf packages; importing either leaf never imports its sibling or the
combined package. The font files live only in the leaf directories. Shared
OFL text lives in `internal/notice/`, which imports no font package.

These sizes are raw font payloads added to native or js/wasm binaries, before
compression and other code/data overhead. The module download still includes
both fonts; the split saves linked binary size, not module download size.
Font internals remain Gloss Text CJKSC. Future regions use distinct flat leaf
packages (`jpregular`, `jpbold`, etc.) and their own regional family names.

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

The root `task ci` includes this module's gate, including js/wasm compilation and binary link-isolation tests.
`go test -short ./...` skips only the consumer-build isolation tests.
Automated cmap coverage checks do not establish visual quality. No visual
quality check has been performed for this expansion.

Source code: MIT (LICENSE). Fonts: OFL-1.1 ([notice](internal/notice/notosanscjksc.OFL.txt)).
Release tags for this nested module have the form `cjk/vX.Y.Z`.
