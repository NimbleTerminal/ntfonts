# Optional color emoji

`nimbleterminal.dev/ntfonts/packs/emoji` is an independent, standard-library-only
Go module that embeds Noto Color Emoji unmodified. Neither ntfonts nor the
`packs` module imports it, and both exclude this directory from their module
zips. Importing it adds 10,673,480 raw font bytes to a native or js/wasm
binary, before compression and other code/data overhead; a test caps the file
at 12,000,000 bytes. Nothing is discovered
or downloaded at runtime.

```sh
go get nimbleterminal.dev/ntfonts/packs/emoji
```

```go
import "nimbleterminal.dev/ntfonts/packs/emoji"

// Register this source with your renderer's fallback mechanism.
font, err := emoji.Embedded.Face(emoji.Fallbacks()[0], "Regular")
```

The package exposes `Embedded`, `Source`, `Fallbacks()`, `Face(family, variant)`,
`Licenses()`, `Files()` and `WebFaces()`, with its own `License` and `WebFace`
types. The family is `Noto Color Emoji`; all four variants return the same
file. Unknown families and variants are errors naming the offender.
`packs.Embedded` does not serve this family.

## Why this file

The embedded `NotoColorEmoji.ttf` is the CBDT/CBLC bitmap color build from
[googlefonts/noto-emoji](https://github.com/googlefonts/noto-emoji) at tag
`v2.051` (commit `8998f5dd683424a73e2314a8c1f1e359c19e8742`), copied byte for
byte: no subsetting, no rename, no fontTools pass. nttypeset rasterizes bitmap
glyph formats through go-text and does not draw COLR or SVG, so CBDT is the
format that renders in Go; rewriting a CBDT font is a risk the project does not
need; and the OFL permits unmodified redistribution with its notice.

The `ofl/notocoloremoji` font in google/fonts is a 25 MB COLRv1 build with an
SVG table and no CBDT. It is not suitable here, and a test asserts that the
embedded file has CBDT and CBLC tables and no COLR, CPAL or SVG table.

Browsers: Chrome and Firefox render CBDT color fonts. Safari does not. The
companion module `nimbleterminal.dev/ntfonts/packs/emoji/sbix` carries the
same bitmaps as an sbix font that Safari renders; `WebFace.Tech` names each
file's format (`color-CBDT` here) so a page can choose with
`CSS.supports("font-tech(...)")`. See [the sbix README](sbix/README.md) for
the format-selection and lazy-loading example.

## Serving the face to a browser terminal

```go
mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.FS(emoji.Files()))))
for _, f := range emoji.WebFaces() {
	// emit {family: f.Family, src: "/fonts/" + f.File, weight: f.Weight, style: f.Style, tech: f.Tech}
}
```

```json
{"fontFaces": [{"family": "Noto Color Emoji", "src": "/fonts/NotoColorEmoji.ttf", "weight": 400, "style": "normal", "tech": "color-CBDT"}]}
```

## Regenerating

From the repository root:

```sh
python scripts/subset-fonts.py /tmp/ntfonts-sources --download --emoji
task --dir packs/emoji notices
task --dir packs/emoji ci
```

`sources.json` records the URL and SHA-256 of the font and its license;
`subsets.sha256` records the embedded file, and `task fonts-check` verifies it.
The root `task ci` includes this module's gate, including js/wasm compilation.

Automated cmap and table checks do not establish visual quality. The specimen
viewer lists this face and marks missing code points, but renders one code
point at a time, so ZWJ sequences, flags and skin-tone modifiers appear as
separate glyphs there: it checks coverage, not shaping. No visual check has
been performed for this module.

Source code: MIT (LICENSE). Font: OFL-1.1 (`notocoloremoji.OFL.txt`).
Release tags for this nested module have the form `packs/emoji/vX.Y.Z`.
