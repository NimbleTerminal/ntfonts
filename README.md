# ntfonts — Embedded fonts for Nimble Terminal

<p>
    <a href="https://pkg.go.dev/nimbleterminal.dev/ntfonts"><img src="https://pkg.go.dev/badge/nimbleterminal.dev/ntfonts.svg" alt="Go Reference"></a>
    <a href="https://github.com/NimbleTerminal/ntfonts/actions/workflows/ci.yml"><img src="https://github.com/NimbleTerminal/ntfonts/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
    <a href="./CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg" alt="Code Of Conduct"></a>
</p>

ntfonts is a Go module that embeds seven OFL-licensed font families as 28 text
subset faces (regular, bold, italic and bold italic of each), plus one symbol
fallback face, totaling 3,968,464 bytes (3.97 MB). It exists to give [nttypeset](https://nimbleterminal.dev/nttypeset) its
faces: `ntfonts.Embedded` implements nttypeset's `FaceSource` without importing
it. The module has no dependencies; it was extracted from
[Gloss](https://github.com/NimbleMarkets/gloss).

```sh
go get nimbleterminal.dev/ntfonts
```

```go
eng, err := nttypeset.NewEngine(nttypeset.WithFonts(ntfonts.Embedded))
```

## What is embedded

The families are stored under the internal names Gloss Text A through G, so the
subsets do not reuse the upstream reserved font names. `ntfonts.Families()`
returns the upstream family names, sorted, and `ntfonts.Embedded.Face(family,
variant)` returns the TTF bytes.

- Family: `Literata`, `Newsreader`, `IBM Plex Sans`, `Source Serif 4`,
  `Iosevka`, `IBM Plex Mono` or `JetBrains Mono`.
- Variant: `Regular`, `Bold`, `Italic` or `BoldItalic`.

`ntfonts.Fallbacks()` returns `[]string{"Noto Sans Symbols 2"}` in renderer
preference order. `Embedded.Face` accepts this name in all four variants;
all return the same upright Regular face because this source has no Bold or
Italic. Missing arrows/math (including →) are filled from pinned Noto Sans Math
inside this single face; its OFL is included in the fallback license entry.
The merged fallback supports horizontal text, not vertical or math layout. The caller implements glyph fallback; this package only returns bytes.
`Families()` remains the seven text families above.

Anything else is an error that names the offender. `ntfonts.Licenses()` returns
each family's OFL text, in the order of the list above followed by the fallback, for programs that must
ship the notices.

## Fonts

The faces are the `Gloss*.ttf` files at the module root. These are the looks
gloss builds from them:

| Look | Body | Code | Subset identifiers |
| --- | --- | --- | --- |
| Readable (default) | Literata | Iosevka | A / E |
| Editorial | Newsreader | Iosevka | B / E |
| Technical | IBM Plex Sans | IBM Plex Mono | C / F |
| Compact | Source Serif 4 | JetBrains Mono | D / G |

All seven families are OFL-1.1; each family's full license is alongside the
fonts as `<family>.OFL.txt`. Google Fonts sources are pinned to commit
`bd8f81ddb5c74d5c8897b36ad88b440266245103`; Iosevka is the official v34.9.0 TTF
release. `sources.json` records exact URLs and SHA-256 hashes.
`subsets.sha256` records the resulting embedded files, and `task fonts-check`
verifies them.

### Regenerating the subsets

Build-time generation only (Python with `fonttools==4.60.2`):

```sh
python scripts/subset-fonts.py /tmp/ntfonts-sources --download
```

The script verifies downloads, instances variable fonts at weight 400/700,
uses their default width, and fixes optical sizes: Literata 12, Newsreader 16,
Source Serif 4 16. It strips hints and discretionary ligature features, retaining
composition, mark positioning, localization and kerning. The text subsets retain ASCII, Latin-1, Latin Extended A/B (U+0100–024F),
spacing modifiers and combining marks (U+02B0–036F), Greek/Coptic and Cyrillic
(U+0370–052F), and Latin Extended Additional including Vietnamese (U+1E00–1EFF).
They also retain general punctuation (U+2000–206F), currency (U+20A0–20CF),
letterlike symbols, number forms, arrows and math (U+2100–22FF), box drawing,
blocks and geometric shapes (U+2500–25FF), check/cross marks U+2713/2714/2717/2718,
and U+FFFD. The symbol fallback retains arrows and math (U+2190–22FF), box drawing,
blocks, geometric shapes, miscellaneous symbols and dingbats (U+2500–27BF).
Coverage varies by upstream family; `coverage.json` records the source-derived
text coverage for every variant. It does not add glyphs absent upstream.
The internal families are renamed `Gloss Text A` through `Gloss Text G` to
avoid reserved font names. Runtime shaping also explicitly disables liga, clig,
dlig and calt in nttypeset. There is no runtime font discovery, download, or
system-font dependency.

### Size and optional CJK

| Default embedded set | Bytes |
| --- | ---: |
| Before (28 Latin subsets) | 1,202,648 |
| Wider 28 text subsets | 3,866,672 |
| Symbol fallback | 101,792 |
| After (29 faces) | 3,968,464 |

These are raw font bytes, not compressed module or binary sizes. No requested
ranges were dropped to meet the 4,000,000-byte budget, which is tested.

CJK is a separate nested module, `nimbleterminal.dev/ntfonts/cjk`, with its own
`go.mod`. It is excluded from the parent module zip and dependency graph.
Importing only ntfonts embeds no CJK data. Opt in with:

```sh
go get nimbleterminal.dev/ntfonts/cjk
```

Use `cjk.Fallbacks()` and `cjk.Embedded.Face(family, variant)` to register the
regional faces with your renderer. Initially only `"Noto Sans CJK SC"` is
available, with Regular and Bold (Italic/BoldItalic alias upright weights).
SC glyph shapes are intended for Simplified Chinese; JP, TC and KR will be
separate families, not aliases. See [the CJK README](cjk/README.md) for the Han
list, per-face sizes and wasm cost. No renderer behavior changes here.

### Go reference fonts

`Go-*.ttf` and `GoMono-*.ttf` are unmodified TTF bytes from
`golang.org/x/image v0.46.0/font/gofont`, created by Bigelow & Holmes for the Go
project. They are unused reference assets: no Go code in this module embeds
them, and they are not part of `Embedded`. Their BSD-style license is
`LICENSE.gofont`.

SHA-256:

```text
c18494baa7ea35b8dbfac3861787d360b9c0ab91232562371a09e6fbf09aaa40  Go-Bold.ttf
d6c33857c7ea7d740a8e1d81d37261782f698e9c17a2c9bf3bb0abbaf99e3c90  Go-BoldItalic.ttf
6efef00d080c8fe68bbe2bf08526294c83d52d7b90b16f938953025c0868cc57  Go-Italic.ttf
197d9f3703b4c00af609178876a8d73e396f64fe438b2c871778566632374be3  Go-Regular.ttf
b165f1a142dc88d3c388de2c89d3ab62dd04620ef21ad07f43fb51f77857edc3  GoMono-Bold.ttf
cb43aca5ab36631c0b4679831017cf249c33db4f721daad7379e04bba383ef1f  GoMono-BoldItalic.ttf
10d5f73519ea32b4882897a9bec6c42ba2ca27d6c06ba9a0e55ae38a383cf7e6  GoMono-Italic.ttf
8bc66a0154bbf69cd24e5bde41a12ec9495c8a242c5def2255e4a164900f1ed7  GoMono-Regular.ttf
```

## License

Source code is MIT; see [LICENSE](./LICENSE). The fonts are OFL-1.1; see
[NOTICE.md](./NOTICE.md) and the `*.OFL.txt` files.

## Specimen viewer

[Web demo](https://nimbleterminal.github.io/ntfonts/) (deployed by GitHub Actions).

Run `task specimen`, then open <http://127.0.0.1:8088>. Stop with Ctrl-C.
Use `task specimen -- -addr 127.0.0.1:8090` to choose another port.
The viewer has family/variant and size controls, editable text, multilingual
samples, and side-by-side comparisons with symbol and SC CJK fallbacks.
It reads each font's cmap and marks missing characters explicitly, preventing
browser system-font fallback from disguising missing coverage. Hover a glyph
for its source face or missing codepoint. Browser shaping is not nttypeset layout.

Core fonts come from `Embedded`; optional CJK files are read from the checkout
without adding a module dependency. Only selected faces are loaded by the
browser. All assets stay local. The viewer is a development tool; it does not
change the library API or embed CJK in the core library.

Build a static copy with `task specimen-build` (writes a new `dist/specimen/`
directory), or `task specimen-build OUT=/tmp/ntfonts-site` for another fresh
directory. Serve the output over HTTP; no Go server is needed in production.
The export includes the font files, catalog and OFL notices and works under
project paths such as `/ntfonts/`.
