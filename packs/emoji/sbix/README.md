# Optional color emoji for Safari (sbix)

`nimbleterminal.dev/ntfonts/packs/emoji/sbix` is an independent,
standard-library-only Go module that embeds `NotoColorEmoji-sbix.ttf`: the
pinned Noto Color Emoji artwork repackaged in Apple's `sbix` bitmap format.
WebKit (Safari on macOS and iOS) rejects the CBDT build that
`nimbleterminal.dev/ntfonts/packs/emoji` ships; it renders this file. Chrome
renders both. Neither ntfonts, `packs` nor `packs/emoji` imports this module,
and each excludes this directory from its module zip. Importing it adds
10,845,476 raw font bytes to a binary; a test caps the file at 12,000,000
bytes. Nothing is discovered or downloaded at runtime.

```sh
go get nimbleterminal.dev/ntfonts/packs/emoji/sbix
```

The package exposes `Embedded`, `Source`, `Fallbacks()`, `Face(family, variant)`,
`Licenses()`, `Files()` and `WebFaces()`, with its own `License` and `WebFace`
types, exactly like `packs/emoji`. The API family is `Noto Color Emoji` and all
four variants return the same file, so a consumer swaps formats without
changing family strings. `WebFace.Tech` is `color-sbix`; the emoji module's is
`color-CBDT`.

Native renderers should keep the CBDT module: nttypeset draws bitmap glyphs
from it already, and this file exists for browsers.

## How the file is made

`scripts/subset-fonts.py --emoji` derives it from the hash-verified
`NotoColorEmoji.ttf` of googlefonts/noto-emoji v2.051 (commit
`8998f5dd683424a73e2314a8c1f1e359c19e8742`) with `fonttools==4.60.2`; the
build is deterministic and needs no other input:

- One `sbix` strike at the CBDT strike's 109 ppem, 72 ppi. Every glyph with a
  CBDT bitmap gets the identical PNG bytes. No resampling.
- Each bitmap glyph gets a `glyf` entry: two coincident rectangles of opposite
  winding over the bitmap's box, so the glyph has a bounding box (Chrome draws
  no sbix bitmap without one) but fills nothing where sbix is unsupported,
  reading as tofu rather than a black box. Apple Color Emoji uses the same
  two-contour box. Glyphs without bitmaps (`.notdef`, space, ZWJ, tag
  characters) stay empty.
- The origin offsets are measured from that box's lower-left corner, as
  CoreText and FreeType do when a glyph has contours, so they are zero and the
  bitmap lands exactly where the CBDT renderer puts it.
- `CBDT` and `CBLC` are removed. `cmap`, `GSUB`, `hmtx`, `vhea`, `vmtx`,
  `OS/2` and `post` are byte-identical to the CBDT font, so ZWJ sequences,
  skin tones, flags, keycaps and variation selectors shape the same way.
- The name table says `Noto Color Emoji sbix` (family and full name),
  `NotoColorEmoji-sbix` (PostScript name), and records the conversion in the
  unique ID and description. Noto's OFL declares no Reserved Font Name, so the
  rename is not required; it keeps the modified file distinguishable from
  Google's. The public file name is `NotoColorEmoji-sbix.ttf`.

Tests compare the embedded file with `../NotoColorEmoji.ttf` table by table
and glyph by glyph (PNG bytes, offsets, `glyf` boxes), check the name table,
`ccmp` and the format 14 cmap, the API, `Files()` traversal refusal, the OFL
notice, the pinned hash and the size budget.

## Serving to a browser

Serve both files and let the page pick the first format it supports with
`CSS.supports("font-tech(...)")`. No user-agent sniffing; when neither format
is supported nothing is downloaded and the platform emoji font renders.

```go
import (
	"nimbleterminal.dev/ntfonts/packs/emoji"
	"nimbleterminal.dev/ntfonts/packs/emoji/sbix"
)

type webFace struct {
	Family string `json:"family"`
	Src    string `json:"src"`
	Weight int    `json:"weight"`
	Style  string `json:"style"`
	Tech   string `json:"tech"`
}

// CBDT first: Chrome and Firefox get Google's unmodified file, Safari the sbix one.
var faces []webFace
for _, f := range emoji.WebFaces() {
	faces = append(faces, webFace{f.Family, "/fonts/" + f.File, f.Weight, f.Style, f.Tech})
}
for _, f := range sbix.WebFaces() {
	faces = append(faces, webFace{f.Family, "/fonts/" + f.File, f.Weight, f.Style, f.Tech})
}
mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.FS(emoji.Files()))))
mux.Handle("/fonts/"+sbix.WebFaces()[0].File, http.StripPrefix("/fonts/", http.FileServer(http.FS(sbix.Files()))))
// Emit faces as JSON for the page, e.g. {"emojiFaces": faces}.
```

```js
// Lazy: call when the user opts into Noto, not on page load (10.7 MB).
async function loadEmojiFont(faces, cssFamily = 'Noto Color Emoji') {
  // Trust a negative answer only when the browser understands font-tech at all.
  const canQuery = globalThis.CSS?.supports?.('font-tech(variations)') === true;
  const face = faces.find(f => !canQuery || CSS.supports(`font-tech(${f.tech})`));
  if (!face) return null; // neither format: keep the platform emoji font
  const font = new FontFace(cssFamily, `url("${face.src}")`, {
    weight: String(face.weight), style: face.style,
    unicodeRange: 'U+0021-10FFFF', // Noto's space is 1.25 em wide; leave U+0020 to the text font
  });
  await font.load();
  document.fonts.add(font);
  return face.tech; // "color-CBDT" or "color-sbix"
}
```

Use the result as `font-family: <text font>, "Noto Color Emoji"`, or declare
one `@font-face` with both sources and let CSS choose:

```css
@font-face {
  font-family: "Noto Color Emoji";
  src: url(/fonts/NotoColorEmoji.ttf) format("truetype") tech(color-CBDT),
       url(/fonts/NotoColorEmoji-sbix.ttf) format("truetype") tech(color-sbix);
}
```

### ntemoji

ntemoji's `examples/cmd/demo-assets` stages `emoji.WebFaces()[0]` into
`site/fonts/` and writes a manifest with `family`, `file` and `bytes`;
`site/font-demo.js` disables the Noto option when `font-tech(color-CBDT)` is
unsupported. To get Noto in Safari: import this module too, stage both files,
make the manifest's `file` a list of `{file, bytes, tech}` entries in the order
above, and replace the CBDT-only check with `loadEmojiFont`: pick the first
entry whose tech passes `CSS.supports`, keep the existing lazy `FontFace`
creation and `unicodeRange`, and disable the option only when no entry passes.
The status line can then say which format loaded.

## Visual check (2026-11-04)

Performed on macOS 27.2 with the exported specimen site served locally, the
same sequences in all renderers: 👋 😀 👍🏽 🇯🇵 👨‍👩‍👧‍👦 1️⃣ ❤️ 🏴󠁧󠁢󠁳󠁣󠁴󠁿 🧑‍💻 ✈️ #️⃣ 🏳️‍🌈
at 16, 28 and 48 px, next to Apple Color Emoji, plus a 109 px baseline test.

- WebKit 605.1.15 (WKWebView, the Safari engine; Safari itself was not
  driven): `font-tech(color-sbix)` true, `font-tech(color-CBDT)` false; the
  CBDT `FontFace` errors, the sbix one loads. Skin tone, flags, the family ZWJ
  sequence, keycaps, VS16 and the tag-sequence flag shape correctly. Bitmaps
  sit where the CBDT file's metrics put them: bottom edge 27/109 em below the
  baseline, the same as Chrome's CBDT rendering. Noto glyphs are about 1.25 em
  wide; Apple's are narrower, so lines run longer than with the platform font.
- Chrome 157 (macOS): both files load and render identically, including
  keycaps, ❤️, ✈️ and 🏳️‍🌈 after a text font.
- FreeType 2.15.0 (Homebrew, with libpng; the engine Chrome and Firefox use
  on Linux and Windows): `FT_LOAD_COLOR` renders of U+1F600, U+1F44D and
  U+1F1EF from both files give identical bitmaps, `bitmap_left`, `bitmap_top`
  and advances. The first build, with offsets measured from the glyph origin,
  rendered 27 px too low in every engine; that is why the offsets are measured
  from the `glyf` box.
- The specimen viewer picked `color-sbix` in WebKit and `color-CBDT` in Chrome
  on its own.

Not verified: Safari.app itself, iOS, Firefox, and Chrome on Linux or Windows
in a browser (FreeType was exercised directly).

## Limitations

- WebKit sends sequences whose base character the preceding text font covers
  (keycap digits, `❤️`, `✈️`, `🏳️‍🌈`) to the platform emoji font when
  `Noto Color Emoji` is listed after a text font; Chrome routes them to Noto.
  Listing Noto first renders everything through Noto but then its 1.25 em
  space applies unless `unicodeRange` excludes U+0020.
- 10.8 MB; load it lazily. PNG data does not benefit from WOFF2.
- Where sbix is unsupported but the file loads anyway, glyphs show as hairline
  boxes; feature detection should prevent the download in the first place.
- Automated checks establish structure, mappings and reproducibility, not
  visual quality beyond the check above.

## Regenerating

From the repository root:

```sh
python scripts/subset-fonts.py /tmp/ntfonts-sources --download --emoji
task --dir packs/emoji/sbix notices
task --dir packs/emoji/sbix ci
```

`sources.json` records the URL and SHA-256 of the CBDT source and its license
(the same entries as `packs/emoji/sources.json`); `subsets.sha256` records the
generated file, and `task fonts-check` verifies it. The root `task ci` includes
this module's gate.

Source code: MIT (LICENSE). Font: OFL-1.1 (`notocoloremoji.OFL.txt`).
Release tags for this nested module have the form `packs/emoji/sbix/vX.Y.Z`.
