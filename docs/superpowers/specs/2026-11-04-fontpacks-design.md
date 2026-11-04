# Fontpacks: optional per-script and emoji font modules

Date: 2026-11-04. Status: builder check done; awaiting approval.

## Intent

The Nimble Terminal ecosystem should render the same glyphs everywhere. Today
ntfonts supplies seven Latin/Greek/Cyrillic text families to nttypeset, and the
optional `cjk` module supplies Simplified Chinese. The browser side, ntwebterm
0.4.2 and booba 0.12.x, now shapes Arabic and other joining scripts and ships a
multilingual check corpus built from twelve Noto Sans faces fetched from Google
Fonts at demo build time. Nothing in the ecosystem embeds those faces, so a Go
renderer has no Arabic, Hebrew, Indic, Thai, Khmer or emoji glyphs, and a booba
host has nothing local to serve through `fontFaces`.

Fontpacks close that gap: small opt-in Go modules, one per script plus one for
color emoji, each serving pinned, reproducible font bytes through the same
`Face(family, variant)` contract the core module and `cjk` already use, and
each able to hand the identical bytes to a browser.

Why packs rather than growing the core module: the core embed is budgeted at
4,000,000 bytes and is linked by every consumer. Script coverage is only wanted
by some hosts, and color emoji alone is around ten megabytes. The `cjk` module
proved the opt-in pattern; packs reuse it.

## Goals

- Importing ntfonts alone embeds nothing new. Each pack is opt-in.
- One package per script: Arabic, Hebrew, Devanagari, Bengali, Tamil, Thai,
  Khmer. These are the scripts in ntwebterm's multilingual corpus, minus the
  ones already handled elsewhere (Latin, Greek, Cyrillic, Vietnamese in core;
  Simplified Chinese in `cjk`; Japanese and Korean are future `cjk` leaves).
- A separate emoji pack carrying Noto Color Emoji.
- Family names match the upstream names ntwebterm's corpus uses, so the same
  `fontFamily` string selects the same face in a browser and in nttypeset.
- Each pack serves its bytes two ways: `Embedded.Face(family, variant)` for Go
  renderers, and an `fs.FS` plus face descriptors for hosts that serve fonts
  to a browser terminal.
- Shaping works. Pack subsets retain every OpenType layout feature. Note for
  the READMEs: nttypeset disables `liga`, `clig`, `dlig` and `calt` at runtime,
  so in Go the guarantee covers `init`/`medi`/`fina`/`rlig` and the Indic and
  USE features, not `calt`-dependent forms. Browsers apply everything.
- Reproducible: pinned upstream URLs and hashes, pinned output hashes, one
  script regenerates everything, `task ci` verifies.
- The specimen viewer lists pack faces and marks missing code points for its
  multilingual samples (catalog only, the user's choice). The viewer renders
  each code point in isolation with `calt` off, so this is a coverage check,
  not a shaping check, and the READMEs must say so. A shaped-run viewer is the
  next project after this one. Automated tests do not establish visual quality.

## Non-goals

- No change to the core seven families, their coverage or their API.
- No move of `cjk` under `packs`. Its import paths are API. Adding `Files()`
  and `WebFaces()` to it is in scope; nothing else in cjk changes.
- No web delivery surface for the core module in this design. The user is still
  choosing gloss's body and code fonts; that comes later as its own change.
- No WOFF2 output. Browsers load TTF/OTF through FontFace and gloss already
  serves TTF. WOFF2 can follow once brotli is in the subset pipeline.
- No monochrome emoji, no Japanese, Korean or Traditional Chinese, no Powerline
  or Nerd Font glyphs. Each would be a later pack or leaf.
- No bidi, script segmentation or fallback logic. Packs return bytes; the
  caller (nttypeset, ntwebterm) does layout.
- No changes in booba, gloss or ntwebterm. Wiring packs into booba-assets and
  gloss's handoff server, and moving consumers from go-booba to
  `nimbleterminal.dev/booba`, are follow-ups in those repositories.

## Anti-goals

- A pack that silently falls back or substitutes. Unknown family or variant is
  an error naming the offender, as in core and `cjk`.
- Adding a dependency. Every module here stays standard-library only and never
  imports nttypeset.
- Hand-edited font files. Fonts change only through `scripts/subset-fonts.py`.
- Breaking `go get nimbleterminal.dev/ntfonts/cjk/...` or the core module's
  pinned hashes.

## Constraints

- Go 1.27.3 toolchain, module minimum Go 1.26, `GOWORK=off`, Go Task.
- fonttools 4.60.2 for subsetting; script sources pinned to the same
  google/fonts commit the core uses, `bd8f81ddb5c74d5c8897b36ad88b440266245103`,
  unless a needed file is absent or in the wrong format there, in which case
  pin one other commit or release and record it in the pack README. The emoji
  pack uses this exception (see Coverage).
- All fonts OFL-1.1. Noto declares no Reserved Font Name, but internal names
  are still renamed to `Gloss <Script>` for consistency with core and `cjk`.
- `task ci` at the root runs every module's gate, as it does for `cjk` today.

## Layout

```
packs/                      module nimbleterminal.dev/ntfonts/packs
  go.mod                    standard library only
  packs.go                  package packs: combined Source over all scripts
  arabic/                   package arabic: NotoSansArabic-Regular.ttf, -Bold.ttf, arabic.go
  hebrew/                   ...
  devanagari/
  bengali/
  tamil/
  thai/
  khmer/
  internal/notice/          shared License type, OFL texts, Licenses() helper
  internal/webface/         shared WebFace type (see API)
  internal/tools/notices/   generates packs/THIRD_PARTY_NOTICES.md
  sources.json, subsets.sha256, Taskfile.yml, README.md, CHANGELOG.md, LICENSE
  emoji/                    module nimbleterminal.dev/ntfonts/packs/emoji (nested)
    go.mod
    emoji.go                NotoColorEmoji-Regular.ttf embedded unmodified
    sources.json, subsets.sha256, Taskfile.yml, README.md, CHANGELOG.md, LICENSE
    internal/tools/notices/
```

Decisions behind the layout, all mine unless marked:

- **One `packs` module for the seven scripts, emoji as its own nested module.**
  Script subsets are small (estimate 50 to 300 KB per face), so one module
  download is acceptable and keeps one go.mod, one Taskfile and one CI job for
  all scripts. Emoji is roughly ten megabytes and is downloaded only by hosts
  that import it. The Go module zip for `packs` excludes the nested `emoji`
  directory automatically.
- **One package per script, both weights in it.** `cjk` split weights into leaf
  packages because each CJK weight is four megabytes. Script faces are two
  orders of magnitude smaller; a per-weight split would add fourteen packages
  for little linked-size benefit. Regular and Italic serve the Regular file,
  Bold and BoldItalic serve the Bold file, as `cjk` does.
- **Granularity per script, not per region** (the user accepted this; the
  specific choice is mine). Import paths are API, and a host that needs Arabic
  should not link Hebrew.
- **`packs` root package serves all seven.** Importing it links every script
  font, mirroring how `cjk` serves both weights. Hosts that want everything
  register one source.

## API

Every script package and the `packs` root package export the same shape as
`cjk/scregular`:

```go
type Source struct{}
var Embedded Source
func Fallbacks() []string                              // e.g. {"Noto Sans Arabic"}
func (Source) Fallbacks() []string
func (Source) Face(family, variant string) ([]byte, error)
type License = notice.License                          // {Family, Text string}
func Licenses() []License
```

Family names are the upstream Google Fonts names: `Noto Sans Arabic`,
`Noto Sans Hebrew`, `Noto Sans Devanagari`, `Noto Sans Bengali`,
`Noto Sans Tamil`, `Noto Sans Thai`, `Noto Sans Khmer`, `Noto Color Emoji`.
The `packs` root `Fallbacks()` returns all seven script families in that order.
Variant handling: `Regular`/`Italic` return the Regular file, `Bold`/`BoldItalic`
return the Bold file; the emoji pack returns its single file for all four.
Unknown family or variant: `fmt.Errorf("packs: unknown family %q", family)`
style errors, prefix `packs:` for scripts and `emoji:` for emoji.

Web delivery, new in this design and present in every pack package:

```go
// WebFace describes one embedded file for a browser FontFace, mirroring the
// fields of ntwebterm's fontFaces option.
type WebFace struct {
    Family string // "Noto Sans Arabic"
    File   string // "NotoSansArabic-Regular.ttf", a path inside Files()
    Weight int    // 400 or 700
    Style  string // "normal"
}
func Files() fs.FS            // the embedded font files under their public names
func WebFaces() []WebFace     // one entry per embedded file
```

A booba host does `http.FileServer(http.FS(arabic.Files()))` under `/fonts/`
and builds its `fontFaces` array from `WebFaces()` with `src` set to
`/fonts/<File>`. Nothing in the pack knows about HTTP or JSON. The `packs`
root package aggregates `Files()` (a read-only union over the seven embed
file systems; file names are unique across packs) and `WebFaces()` the same
way it aggregates `Face`, so a host that imports the root serves everything
from one mount. The emoji module exports the identical surface (`Embedded`,
`Fallbacks`, `Face`, `Licenses`, `Files`, `WebFaces`) with its own copies of
`WebFace` and `License`; `packs.Embedded` does not serve `Noto Color Emoji`.
`cjk`, `cjk/scregular` and `cjk/scbold` gain `Files()` and `WebFaces()` in
this change too (my call: two small functions and their tests, shipped as a
cjk minor release), so the Simplified Chinese corpus row is loadable through
the same browser path. `WebFace` is
defined once in `packs/internal/webface` and aliased from each package; the
emoji module defines its own identical type because it cannot import the
`packs` module's internal package. Files on disk, in `//go:embed`, in `subsets.sha256` and in
`Files()` all use the public names (`NotoSansArabic-Regular.ttf`), so URLs read
well and `Files()` is a plain `embed.FS` with no renaming layer. Only the
font's internal name table says `Gloss Arabic`. (Builder check resolved this;
the earlier draft's layout line said `GlossArabic-*.ttf`, which was wrong.)

## Coverage

Each script subset keeps only glyphs present upstream within these ranges,
plus U+0020, U+00A0, U+200B to U+200F, U+2066 to U+2069 and U+FFFD so joiners
and directional controls never fall out of the face. Latin letters and digits
are not retained; the primary face supplies them, and keeping them out stops a
pack face from capturing Latin when it precedes the monospace fallback in a
browser font list. Arabic-Indic digits live inside the Arabic block and stay.

| Pack | Blocks |
| --- | --- |
| arabic | U+0600–06FF, U+0750–077F, U+08A0–08FF, U+FB50–FDFF, U+FE70–FEFF |
| hebrew | U+0590–05FF, U+FB1D–FB4F |
| devanagari | U+0900–097F, U+1CD0–1CFF, U+A8E0–A8FF |
| bengali | U+0980–09FF |
| tamil | U+0B80–0BFF, U+11FC0–11FFF |
| thai | U+0E00–0E7F |
| khmer | U+1780–17FF, U+19E0–19FF |

The builder may widen a range when the upstream face covers an adjacent block
that the ntwebterm corpus rows need; narrowing is not allowed. Record the final
ranges in `packs/README.md` and a `packs/coverage.json` like the core one.

Emoji is embedded unmodified, hash-verified, no subsetting and no rename.
The source is **not** google/fonts: at the pinned commit
`ofl/notocoloremoji/NotoColorEmoji-Regular.ttf` is a 25,332,736-byte COLRv1
font with a 20 MB SVG table and no CBDT/CBLC (verified 2026-11-04 by reading
its table directory). The CBDT build lives in the googlefonts/noto-emoji
repository as `fonts/NotoColorEmoji.ttf`; at tag `v2.047` it is 10,643,852
bytes with CBDT, CBLC, GSUB and cmap tables and an OFL-1.1 `fonts/LICENSE`
(verified the same way). Pin a noto-emoji release tag, `v2.047` or a newer tag
whose `fonts/` directory still carries the CBDT file, and record the tag, URL
and SHA-256 in `packs/emoji/sources.json` and the README.
Reasons for CBDT: nttypeset's rasterizer draws bitmap glyph formats through
go-text and does not draw COLR or SVG, so CBDT is the format that renders in
Go; fonttools rewriting of CBDT is a risk the project does not need to take;
and the OFL permits unmodified redistribution with the notice. Safari does not
render CBDT and will use the platform emoji font instead; Chrome and Firefox
render it. The emoji README states this, the raw byte size, and the js/wasm
cost in the same words the CJK README uses.

## Subsetting

`scripts/subset-fonts.py` gains `--packs` and `--emoji` modes next to `--cjk`,
writing into `packs/` and `packs/emoji/` and regenerating each module's
`subsets.sha256`. Pack mode differs from text mode in two ways and the script
must make that explicit in code and comments:

- A separate `subset.Options` for pack mode with `layout_features=['*']`, so
  GSUB, GPOS and GDEF closure keep unencoded alternates; the text-mode options
  object is untouched and text output stays byte-identical. Every GSUB/GPOS
  feature is retained. Arabic needs
  `init`, `medi`, `fina`, `rlig`, `calt`; Indic scripts need `akhn`, `rphf`,
  `blwf`, `half`, `pstf`, `vatu`, `pres`, `abvs`, `blws`, `psts`, `haln`,
  `cjct`, `nukt`; Khmer and Thai need marks and the Universal Shaping Engine
  features. Stripping them, as the text subsets do, would break shaping.
- Variable sources are instanced at wght 400 and 700 at default width, as the
  core already does for IBM Plex Sans; hints are stripped; timestamps fixed.

Upstream files, to be confirmed by the builder at the pinned commit, are the
`ofl/notosans<script>/NotoSans<Script>[wdth,wght].ttf` variable fonts and
`ofl/notocoloremoji/NotoColorEmoji-Regular.ttf`, with each directory's
`OFL.txt`. `packs/sources.json` and `packs/emoji/sources.json` record URL and
SHA-256 for every download, in the format of the existing files.

## Tests

Standard library only, in each module, modelled on `ntfonts_test.go` and
`cjk/*_test.go`:

- Pinned hashes of every embedded file (`subsets.sha256`), also checked by
  `task fonts-check` in each module.
- Family and variant table: every family in `Fallbacks()` serves all four
  variants; unknown names error and name the offender; Regular and Italic
  bytes are identical, Bold and BoldItalic bytes are identical.
- Coverage: parse the cmap of each face and assert every code point of the
  matching ntwebterm corpus row (copied verbatim into the test as the
  ecosystem's agreed sample text, with a comment naming its source) is
  present, ignoring the Latin label prefix, spaces and punctuation outside the
  pack's ranges. Also assert that no Latin letter is present in a script face.
- Layout features, by parsing the GSUB and GPOS feature lists with the same
  minimal table parser the cmap tests use: the Arabic face has `init`, `medi`,
  `fina` and `rlig`; the Devanagari face has `akhn` and whichever Indic tags
  the first build shows it actually carries (`half` is a hypothesis until
  then); every script face has `mark` and `mkmk` in GPOS, since Hebrew, Thai,
  Khmer and Arabic vowels depend on mark positioning rather than substitution;
  and the Arabic face's glyph count (`maxp.numGlyphs`) is well above its cmap
  entry count, proving the unencoded contextual forms survived the subset.
  Together these guard the one way the pipeline can regress silently.
- Controls and exclusions: ASCII digits and Latin letters are absent from every
  script face; U+200C and U+200D are present in each face whose upstream has
  them (the builder records which, in `coverage.json`).
- Emoji format: the embedded file has CBDT and CBLC tables and no COLR, CPAL
  or SVG table, so a wrong upstream can never land silently.
- `Files()` and `WebFaces()` agree: every `WebFace.File` opens, every file in
  `Files()` is described exactly once, and bytes equal `Face()` output.
- `Licenses()` returns one OFL entry per family with non-empty text.
- Size budgets per module, asserted in a test and documented: builder measures
  and sets them at roughly 25 percent above the measured total. Record the
  final numbers in the READMEs.
- Link isolation, as `cjk/isolation_test.go` does: a consumer that imports only
  `packs/arabic` links no other pack's bytes; skipped under `-short`.
- `packs` root: `Face` dispatches to the right package for every family.

## Tooling and CI

- `packs/Taskfile.yml` and `packs/emoji/Taskfile.yml` copied from
  `cjk/Taskfile.yml`, including lint, vet, fmt, tidy, verify, fonts-check,
  test, test-race, cross-windows, cross-wasm, vulncheck, notices and ci. Root
  `Taskfile.yml`'s `ci` task gains both modules the way it includes `cjk`.
- `.github/workflows/ci.yml` runs both new gates.
- `internal/tools/notices` equivalents produce `THIRD_PARTY_NOTICES.md` for
  each module; `notices-check` enforces it.
- `.gitignore` in each module matches `cjk/.gitignore`.
- The specimen tool reads pack faces from the checkout, as it already does for
  CJK, without adding a module dependency, and offers them as fallbacks in the
  multilingual samples. No change to how it renders text. The exported Pages
  site includes every pack file, including the emoji font; the browser fetches
  a face only when it is selected.

## Documentation

- `packs/README.md`: purpose, import paths, per-pack sizes table, family
  names, coverage table, regeneration commands, the shaping-features note, the
  web delivery example with ntwebterm `fontFaces`, release tag form
  `packs/vX.Y.Z`, and the CJK README's sentence that one module zip contains
  every script face while a per-script import saves linked binary size. `packs/emoji/README.md` likewise with tag `packs/emoji/vX.Y.Z`
  and the Safari CBDT note.
- Root `README.md`: a short section pointing to the packs, parallel to the
  CJK section. `AGENTS.md`: module layout table and invariants updated to name
  the new modules, the feature-retention difference, and the unmodified emoji
  file. `DEVELOP.md`: regeneration commands. `CHANGELOG.md` entries under
  Unreleased in root and in each new module.
- Every README states that automated cmap and feature checks do not establish
  visual quality, that the specimen checks coverage and not shaping, and
  reports whether a visual check happened.

## Left to the builder

- Exact upstream paths and hashes at the pinned commit; whether any script's
  variable font lacks a `wdth` axis (then instance on `wght` only); which
  noto-emoji tag to pin.
- Final size budgets and the measured per-pack sizes.
- Whether to widen any coverage range so a corpus row renders fully.
- Internal structure of the subset script's pack mode, as long as text mode
  output is byte-identical to today (core and cjk hashes unchanged).
- How the specimen tool discovers pack files on disk.

## Follow-ups outside this design

- booba: let `booba-assets` copy pack faces and emit `fontFaces`; gloss: serve
  packs from the handoff server; both move to `nimbleterminal.dev/booba`.
- Core module web delivery once gloss's fonts are chosen.
- Shaped-run specimen viewer: render same-face runs as one span with bidi
  direction and layout features on, per-pack toggles, updated catalog tests.
  Agreed as the project that follows this one.
- WOFF2 output; monochrome emoji; JP, TC and KR leaves under `cjk`.
