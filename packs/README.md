# Optional script packs

`nimbleterminal.dev/ntfonts/packs` is an independent, standard-library-only Go
module of opt-in script fonts. The parent ntfonts module does not import it and
its module zip excludes this directory. Import one script package to link only
that script. Nothing is discovered or downloaded at runtime.

```sh
go get nimbleterminal.dev/ntfonts/packs/arabic
```

```go
import "nimbleterminal.dev/ntfonts/packs/arabic"

// Register this source with your renderer's fallback mechanism.
font, err := arabic.Embedded.Face(arabic.Fallbacks()[0], "Regular")
```

| Import path | Family | Raw font bytes (Regular + Bold) |
| --- | --- | ---: |
| `nimbleterminal.dev/ntfonts/packs/arabic` | Noto Sans Arabic | 244,380 |
| `nimbleterminal.dev/ntfonts/packs/hebrew` | Noto Sans Hebrew | 30,260 |
| `nimbleterminal.dev/ntfonts/packs/devanagari` | Noto Sans Devanagari | 343,112 |
| `nimbleterminal.dev/ntfonts/packs/bengali` | Noto Sans Bengali | 193,456 |
| `nimbleterminal.dev/ntfonts/packs/tamil` | Noto Sans Tamil | 77,140 |
| `nimbleterminal.dev/ntfonts/packs/thai` | Noto Sans Thai | 37,944 |
| `nimbleterminal.dev/ntfonts/packs/khmer` | Noto Sans Khmer | 142,432 |
| `nimbleterminal.dev/ntfonts/packs` | all seven | 1,068,724 |

One module zip contains every script face; a per-script import saves linked
binary size, not module download size. These sizes are raw font payloads added
to native or js/wasm binaries, before compression and other code/data overhead.

Every package exposes the same surface: `Embedded`, `Source`, `Fallbacks()`,
`Face(family, variant)`, `Licenses()`, `Files()` and `WebFaces()`, with the
shared `License` and `WebFace` types. `Regular` and `Italic` return the Regular
file; `Bold` and `BoldItalic` return the Bold file. Unknown families and
variants are errors naming the offender. The combined `packs` package serves
all seven families from one source and aggregates `Files()` and `WebFaces()`.
Noto Color Emoji is the separate [`packs/emoji`](emoji/README.md) module and is
not served by `packs.Embedded`.

Family names are the upstream names, so the same string selects the same face
in nttypeset and in a browser: `Noto Sans Arabic`, `Noto Sans Hebrew`,
`Noto Sans Devanagari`, `Noto Sans Bengali`, `Noto Sans Tamil`,
`Noto Sans Thai`, `Noto Sans Khmer`. The embedded files use public names such
as `NotoSansArabic-Regular.ttf`; only the font's internal name table says
`Gloss Arabic`.

## Coverage

Each subset keeps the glyphs present upstream within its script blocks, plus
U+0020, U+00A0, U+200B–U+200F, U+2066–U+2069 and U+FFFD so joiners and
directional controls never fall out of the face. Latin letters and digits are
not retained: the primary face supplies them, and leaving them out stops a pack
face from capturing Latin when it precedes the monospace fallback in a browser
font list. `coverage.json` records the covered characters per face.

| Pack | Blocks |
| --- | --- |
| arabic | U+0600–06FF, U+0750–077F, U+08A0–08FF, U+FB50–FDFF, U+FE70–FEFF |
| hebrew | U+0590–05FF, U+FB1D–FB4F |
| devanagari | U+0900–097F, U+1CD0–1CFF, U+A8E0–A8FF |
| bengali | U+0980–09FF |
| tamil | U+0B80–0BFF, U+11FC0–11FFF |
| thai | U+0E00–0E7F |
| khmer | U+1780–17FF, U+19E0–19FF |

## Shaping features

Unlike the core text subsets, pack subsets retain every OpenType layout
feature: Arabic needs `init`, `medi`, `fina` and `rlig`; the Indic scripts need
`akhn`, `rphf`, `blwf`, `half`, `pres`, `abvs`, `blws`, `psts`, `cjct` and
their mark positioning in `abvm`/`blwm`; Thai and Khmer need `mark`/`mkmk` and
the Universal Shaping Engine features. Tests assert the tags each face
carries. nttypeset disables `liga`, `clig`, `dlig` and `calt` at runtime, so in
Go the guarantee covers joining, reordering and mark positioning rather than
`calt`-dependent forms; browsers apply everything.

## Serving faces to a browser terminal

`Files()` is an `fs.FS` of the embedded files and `WebFaces()` describes each
one with the fields of ntwebterm's `fontFaces` option:

```go
mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.FS(arabic.Files()))))
for _, f := range arabic.WebFaces() {
	// emit {family: f.Family, src: "/fonts/" + f.File, weight: f.Weight, style: f.Style}
}
```

```json
{"fontFaces": [{"family": "Noto Sans Arabic", "src": "/fonts/NotoSansArabic-Regular.ttf", "weight": 400, "style": "normal"}]}
```

Nothing in the packs knows about HTTP or JSON.

## Regenerating

From the repository root, with Python and `fonttools==4.60.2`:

```sh
python scripts/subset-fonts.py /tmp/ntfonts-sources --download --packs
task --dir packs notices
task --dir packs ci
```

Sources are pinned to google/fonts commit
`bd8f81ddb5c74d5c8897b36ad88b440266245103`; `sources.json` records each URL
and SHA-256, `subsets.sha256` records the generated files, and `task fonts-check`
verifies them. The script instances the variable fonts at weight 400 and 700
at default width, strips hints, keeps every layout feature, renames the fonts
and fixes timestamps. The root `task ci` includes this module's gate, including
js/wasm compilation and the link-isolation test.

Automated cmap and feature checks do not establish visual quality. The
specimen viewer lists these faces and marks missing code points, but renders
one code point at a time with `calt` off: it checks coverage, not shaping. No
visual check has been performed for this module.

Source code: MIT (LICENSE). Fonts: OFL-1.1 ([notices](internal/notice/)).
Release tags for this nested module have the form `packs/vX.Y.Z`.
