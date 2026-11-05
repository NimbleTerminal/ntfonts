# Emoji sbix Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `nimbleterminal.dev/ntfonts/packs/emoji/sbix`, a Safari-compatible sbix build of the pinned Noto Color Emoji, with format metadata, tests, specimen support and docs.

**Architecture:** A new nested standard-library-only module mirrors `packs/emoji` (`Source`, `Embedded`, `Face`, `Files`, `WebFaces`, `Licenses`). `scripts/subset-fonts.py --emoji` derives the sbix font from the cached CBDT file deterministically. Both emoji `WebFace` types gain `Tech`, the CSS `font-tech()` keyword, so browsers select a format with `CSS.supports`.

**Tech Stack:** Go 1.26+ (stdlib only), Python 3 + fonttools 4.60.2 (build time), Go Task, GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-11-04-emoji-sbix-design.md`

> **Executed 2026-11-04.** Task 1's `originOffsetY = BearingY - height` and Task 2's matching assertion were superseded: offsets are measured from the `glyf` box corner (zero). See the spec's amended Decision 1 and the ledger ruling for Task 7.

## Global Constraints

- Standard library only; no `require` or `replace` in any go.mod.
- Fonts change only through `scripts/subset-fonts.py` with `fonttools==4.60.2`; hashes pinned in each module's `subsets.sha256`.
- The emoji CBDT file stays byte-identical (sha256 `72a635cb3d2f3524c51620cdde406b217204e8a6a06c6a096ff8ed4b5fd6e27b`).
- Public file name `NotoColorEmoji-sbix.ttf`; API family `Noto Color Emoji`; name table family `Noto Color Emoji sbix`.
- `WebFace.Tech` values: `color-CBDT` (emoji), `color-sbix` (sbix).
- Release harmony: five modules, tag `packs/emoji/sbix/vX.Y.Z`.
- Python cache and venv: `/private/tmp/claude-501/-Users-evan-projects-ntfonts/cf44024f-be05-4de0-aae5-74514e08fbba/scratchpad/{ntfonts-sources,venv}`.

## Review Focus

1. A consumer serving both files under one `font-family` in Chrome must not load both: the example and specimen pick the first supported tech only (Task 5 test: pick order).
2. A browser without `CSS.supports('font-tech(...)')` (old) must still get a chance to load: treat an unknown query as "unsupported tech unknown" and fall back to loading nothing rather than throwing (Task 5: `usable()` guards `CSS.supports` existence).
3. Glyphs without bitmaps (`.notdef`, space, ZWJ, tag characters) must stay empty in `glyf` so they render invisible (Task 2 `TestGlyfBoxes`).
4. The sbix origin offsets must place bitmaps on the baseline exactly as CBDT does (Task 2 per-glyph offset assertion against CBDT metrics).
5. `Files()` must refuse path traversal like every other module (Task 2 `TestFilesAndWebFaces`).

---

### Task 1: Script builds the sbix module files

**Files:**
- Modify: `scripts/subset-fonts.py` (emoji branch)
- Create: `packs/emoji/sbix/sources.json` (copy of `packs/emoji/sources.json`)
- Generated: `packs/emoji/sbix/NotoColorEmoji-sbix.ttf`, `packs/emoji/sbix/notocoloremoji.OFL.txt`, `packs/emoji/sbix/subsets.sha256`

**Interfaces:**
- Produces: font file with tables `sbix`, `glyf`, `loca`; no `CBDT`/`CBLC`; name IDs 1,4 = `Noto Color Emoji sbix`, 6 = `NotoColorEmoji-sbix`, 3 = `Noto Color Emoji sbix; ntfonts conversion of ` + original ID 5, 10 = `Color emoji font using sbix glyph data, converted by ntfonts from the Noto Color Emoji CBDT build.`; manifest line `<sha256>  NotoColorEmoji-sbix.ttf`.

- [ ] **Step 1: Extend the `--emoji` branch.** After copying the CBDT file, add function `sbix_font(src: Path, dest: Path)`: `TTFont(src, recalcTimestamp=False)`; strike ppem from `CBLC.strikes[0].bitmapSizeTable.ppemX`; `Strike(ppem=ppem, resolution=72)`; per CBDT glyph `SbixGlyph(glyphName, graphicType='png ', imageData, originOffsetX=BearingX, originOffsetY=BearingY-height)`; `glyf` built with `TTGlyphPen`: bitmap glyphs get two rectangles over the scaled box (`round(v*upem/ppem)`), first counter-clockwise `(x0,y0)->(x0,y1)->(x1,y1)->(x1,y0)`, second reversed; others `Glyph()`; `loca` new table; rename name IDs as in Interfaces; delete `CBDT`, `CBLC`; `font.save(dest)`. Record the hash in a second dict and write `dest.parent/'subsets.sha256'`. Copy the OFL text next to it. Update the module docstring usage line.
- [ ] **Step 2: Run** `venv/bin/python scripts/subset-fonts.py <cache> --emoji`. Expected: prints `Subset bytes: 10673480`; `git status` shows only new files under `packs/emoji/sbix/` and the script; `shasum -a 256 packs/emoji/NotoColorEmoji.ttf` unchanged.
- [ ] **Step 3: Determinism.** Run the command again. Expected: `git status` unchanged (no modified files).
- [ ] **Step 4: Commit** `feat(script): derive the sbix emoji font from the pinned CBDT build`.

### Task 2: The `sbix` module with tests

**Files:**
- Create: `packs/emoji/sbix/go.mod` (`module nimbleterminal.dev/ntfonts/packs/emoji/sbix`, `go 1.26`), `LICENSE`, `.gitignore`, `Taskfile.yml` (copy of emoji's; `fonts-check` desc "Verify the sbix emoji font against subsets.sha256"), `internal/tools/notices/main.go` (copy, import `sbix`), `THIRD_PARTY_NOTICES.md` (generated), `sbix.go`, `sbix_test.go`, `fontparse_test.go`

**Interfaces:**
- Produces: `type Source struct{}`; `var Embedded Source`; `func Fallbacks() []string`; `func (Source) Fallbacks() []string`; `func (Source) Face(name, variant string) ([]byte, error)` (errors `sbix: unknown variant %q`, `sbix: unknown family %q`); `func Files() fs.FS`; `type WebFace struct{ Family, File string; Weight int; Style, Tech string }`; `func WebFaces() []WebFace` = `[{"Noto Color Emoji","NotoColorEmoji-sbix.ttf",400,"normal","color-sbix"}]`; `type License struct{ Family, Text string }`; `func Licenses() []License`.

- [ ] **Step 1: Write `fontparse_test.go`** (package `sbix_test`): `tables`, `parseCmap` (copies of emoji's), plus `numGlyphs(data) int` (maxp offset 4), `cbdtBitmaps(data) (map[int]bitmap, error)` parsing CBLC strike 0 index subtable array (indexFormat 1, imageFormat 17: 5-byte small metrics `height,width,bearingX(int8),bearingY(int8),advance` + uint32 dataLen + PNG), `sbixBitmaps(data) (ppem, ppi int, glyphs map[int]bitmap, err)` (strike: ppem u16, ppi u16, glyphDataOffsets u32[numGlyphs+1]; glyph: originOffsetX int16, originOffsetY int16, graphicType 4 bytes, data), `glyfEntry(data, gid) (contours int, xMin,yMin,xMax,yMax int, empty bool)` via head.indexToLocFormat + loca, `nameString(data, id int) string` (platform 3 UTF-16BE), `cmapHasFormat14(data) bool`, `featureTags(data, tag string) []string` (as in packs tests). `type bitmap struct{ png []byte; x, y int }`.
- [ ] **Step 2: Write `sbix_test.go`:** `TestFormat` (has `sbix`,`glyf`,`loca`; lacks `CBDT`,`CBLC`,`COLR`,`CPAL`,`SVG `), `TestDerivedFromCBDT` (reads `../NotoColorEmoji.ttf`; identical bytes for `cmap`,`GSUB`,`hmtx`,`hhea`,`vhea`,`vmtx`,`OS/2`,`post`; equal numGlyphs; sbix ppem 109, ppi 72; `len(sbixGlyphs)==len(cbdtGlyphs)==3985`; per gid `bytes.Equal(png)` and `x==bearingX`, `y==bearingY-height`), `TestGlyfBoxes` (bitmap gids: contours 2, box == round(v*2048/109) of CBDT metrics; other gids empty), `TestNames` (ID 1 `Noto Color Emoji sbix`, 6 `NotoColorEmoji-sbix`, 10 contains `converted by ntfonts`), `TestShapingTables` (`featureTags(GSUB)` contains `ccmp`; `cmapHasFormat14`), `TestFaces`, `TestCorpusCoverage` (same runes as emoji), `TestFilesAndWebFaces` (want with `Tech: "color-sbix"`; `../x`, `/abs` refused), `TestLicenses`, `TestHashes` (`NotoColorEmoji-sbix.ttf`), `TestBudget` (12_000_000).
- [ ] **Step 3: Run** `cd packs/emoji/sbix && GOWORK=off go test ./...`. Expected: build failure, `package nimbleterminal.dev/ntfonts/packs/emoji/sbix` not found / undefined symbols.
- [ ] **Step 4: Write `sbix.go`** per Interfaces, embedding `NotoColorEmoji-sbix.ttf` and `notocoloremoji.OFL.txt`; package comment explains the derivation, Safari, and `Tech`.
- [ ] **Step 5: Run tests.** Expected: PASS. Then `task --dir packs/emoji/sbix notices` and `task --dir packs/emoji/sbix ci`. Expected: all green.
- [ ] **Step 6: Commit** `feat(sbix): add the Safari-compatible Noto Color Emoji sbix module`.

### Task 3: `Tech` on the emoji module

**Files:**
- Modify: `packs/emoji/emoji.go`, `packs/emoji/emoji_test.go`, `packs/emoji/CHANGELOG.md`

- [ ] **Step 1:** In `TestFilesAndWebFaces` change the want to include `Tech: "color-CBDT"`. Run: expected FAIL (unknown field).
- [ ] **Step 2:** Add `Tech string` to `WebFace` with doc comment, set `"color-CBDT"` in `WebFaces()`; update the package comment's Safari sentence to point at the sbix module. Run: PASS. `task --dir packs/emoji ci`: green.
- [ ] **Step 3:** CHANGELOG Unreleased: `WebFace.Tech` added. Commit `feat(emoji): expose the CSS font-tech keyword on WebFace`.

### Task 4: Repository wiring

**Files:**
- Modify: `Taskfile.yml` (`ci` adds `sbix-ci`; task `sbix-ci: task --dir packs/emoji/sbix ci`), `.github/workflows/pages.yml` (tags add `packs/emoji/sbix/v*`), `AGENTS.md` (overview, layout row, invariants: five modules, sbix invariants), `DEVELOP.md` (module list, regeneration paragraph, releases: five tags, push in two batches)

- [ ] **Step 1:** Make the edits. Run `task --list-all | grep sbix-ci`. Expected: listed. Run `grep -c "packs/emoji/sbix" AGENTS.md DEVELOP.md`. Expected: ≥2 each.
- [ ] **Step 2:** Commit `chore: wire the sbix module into CI, Pages and agent docs`.

### Task 5: Specimen viewer format selection

**Files:**
- Modify: `internal/tools/specimen/main.go`, `internal/tools/specimen/main_test.go`, `internal/tools/specimen/index.html`

**Interfaces:**
- `face` gains `Tech string json:"tech,omitempty"`; CBDT entry `Tech: "color-CBDT"`, sbix entry (`packs/emoji/sbix/NotoColorEmoji-sbix.ttf`, Family `Noto Color Emoji`, Variant `Regular`, Group `packs`) `Tech: "color-sbix"`; sbix notice appended only when the emoji notice is absent.

- [ ] **Step 1:** In `main_test.go` add fake `packs/emoji/sbix/NotoColorEmoji-sbix.ttf` = `fake sbix` and its OFL; expect 33 faces, a face with `tech == "color-sbix"` and one with `color-CBDT`, and `licenses.txt` containing `test emoji license` once. Run: FAIL (32 faces).
- [ ] **Step 2:** Implement discovery and `Tech`. Run: PASS.
- [ ] **Step 3:** `index.html`: `const techQuery=globalThis.CSS&&CSS.supports&&CSS.supports('font-tech(variations)')`; `function usable(f){return !f.tech||!techQuery||CSS.supports('font-tech('+f.tech+')')}`; `pick` searches `catalog.filter(usable)`; family dropdown unchanged; status line appends ` (`+tech+`)` for faces with tech; when `$('packs')` is checked and a pack family has no usable face, status gains `Noto Color Emoji: no supported color format in this browser`. Footer sentence: "Color emoji loads the first format this browser reports supporting (`font-tech`)."
- [ ] **Step 4:** `go run ./internal/tools/specimen -export <scratch>/site`, open `fonts.json`, confirm two `Noto Color Emoji` entries with tech. `GOWORK=off go test ./internal/...` PASS.
- [ ] **Step 5:** Commit `feat(specimen): pick the color emoji format by font-tech support`.

### Task 6: Documentation and example

**Files:**
- Create: `packs/emoji/sbix/README.md`, `packs/emoji/sbix/CHANGELOG.md`
- Modify: `packs/emoji/README.md`, `README.md`, `CHANGELOG.md`

- [ ] **Step 1:** sbix README: what/why, support matrix (verified vs untested), derivation and naming, sizes/budget, regeneration, "Serving to a browser" with the Go handler that merges `emoji.WebFaces()` and `sbix.WebFaces()` into JSON `{family,src,weight,style,tech}` and the JS `loadEmojiFont(faces)` that picks the first `CSS.supports('font-tech(...)')` entry, lazily creates a `FontFace` with `unicodeRange: 'U+0021-10FFFF'`, and returns `null` (system emoji) when none is supported; an "ntemoji" subsection mapping it onto `demo-assets` and `font-demo.js`; limitations; visual-check statement (filled in Task 7).
- [ ] **Step 2:** emoji README: Safari paragraph points to sbix; `Tech` in the JSON example. Root README: pack section mentions sbix and the `go get`. CHANGELOGs: root Unreleased (specimen, docs), sbix `## Unreleased` initial entry.
- [ ] **Step 3:** Commit `docs: describe the sbix emoji module and format selection`.

### Task 7: Rendering validation

**Files:**
- Scratch only: `<scratch>/validate/` page + `wksnap` snapshots; results recorded in `packs/emoji/sbix/README.md` "Visual check" and the final report.

- [ ] **Step 1:** Build a comparison page from the exported specimen site's files (sbix and CBDT URLs): rows for Noto sbix, Noto CBDT, platform emoji, in text context (`font-family: Helvetica, X`) and emoji-first, with the sequence set `👋 😀 👍🏽 🇯🇵 👨‍👩‍👧‍👦 1️⃣ ❤️ 🏴󠁧󠁢󠁳󠁣󠁴󠁿 🧑‍💻 ✈️ #️⃣ 🏳️‍🌈` at 16/28/48 px, plus baseline guides. Render with `wksnap` and in Chrome; zoom and compare alignment, sizing, spacing, shaping.
- [ ] **Step 2:** Load the exported specimen site in WebKit with packs on and confirm the status names `color-sbix`; in Chrome confirm `color-CBDT`.
- [ ] **Step 3:** Optional FreeType check: `venv/bin/pip install freetype-py`, load the sbix font, `FT_LOAD_COLOR` render of U+1F600 at 109 px, assert bitmap non-empty. Record result or "not run".
- [ ] **Step 4:** Write findings into the README's "Visual check" and commit `docs(sbix): record the WebKit and Chrome visual check`.

### Task 8: Final gate

- [ ] `task ci` at the root (includes all five module gates). Expected: green.
- [ ] Final whole-branch review per executing-plans; fix pass; finishing-a-development-branch.
