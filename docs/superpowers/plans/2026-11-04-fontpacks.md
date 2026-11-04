# Fontpacks Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add opt-in `nimbleterminal.dev/ntfonts/packs` (seven Noto Sans script packages) and `nimbleterminal.dev/ntfonts/packs/emoji` (Noto Color Emoji, CBDT) modules that serve pinned font bytes through the existing `Face(family, variant)` contract and a new `Files()`/`WebFaces()` browser surface, plus the same surface on `cjk`.

**Architecture:** Two new nested Go modules mirror `cjk/`: standard-library only, embed-backed, hash-pinned, each with its own Taskfile, notices tool and CI gate wired into root `task ci`. `scripts/subset-fonts.py` gains `--packs` (full-feature subsets instanced at wght 400/700) and `--emoji` (hash-verified copy) modes without changing text or CJK output. The specimen viewer lists pack faces from the checkout as a coverage catalog only.

**Tech Stack:** Go 1.27.3 (module `go 1.26`), Go Task, golangci-lint v2.15.0, Python 3 with `fonttools==4.60.2` (via `uv venv`), GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-11-04-fontpacks-design.md`

## Global Constraints

- Every module is standard-library only; `go.mod` has no `require`. Never import nttypeset, even in tests.
- Never edit a font file by hand. Fonts change only through `scripts/subset-fonts.py`; the script regenerates each module's `subsets.sha256`.
- Text-mode and `--cjk` output must stay byte-identical: root `subsets.sha256`, `coverage.json` and `cjk/subsets.sha256` show no diff after running the extended script.
- Script sources pinned to google/fonts commit `bd8f81ddb5c74d5c8897b36ad88b440266245103`. Emoji pinned to googlefonts/noto-emoji tag `v2.051`, commit `6202fe7c20dd5e1727a4c3c01604edc176c576da`.
- Family names are API: `Noto Sans Arabic`, `Noto Sans Hebrew`, `Noto Sans Devanagari`, `Noto Sans Bengali`, `Noto Sans Tamil`, `Noto Sans Thai`, `Noto Sans Khmer`, `Noto Color Emoji`. Variants: `Regular`, `Bold`, `Italic`, `BoldItalic`; Italic aliases Regular, BoldItalic aliases Bold.
- Files on disk, in `//go:embed`, in `subsets.sha256` and in `Files()` use public names (`NotoSansArabic-Regular.ttf`). Only the font name table says `Gloss Arabic`.
- Error prefixes: `packs:` in the packs module, `emoji:` in the emoji module, `cjk:` in cjk. Errors name the offending family or variant with `%q`.
- Pack subsets retain all layout features (`layout_features=['*']`), strip hints, instance variable fonts at wght 400 and 700 at default width, fix `head` timestamps to 2082844800.
- All tasks run with `GOWORK=off` (Taskfiles set it). Run `task ci` at the root before finishing.
- Documentation states that automated checks do not establish visual quality, that the specimen checks coverage and not shaping, and whether a visual check happened (none is planned here).
- Commit after each task. Do not push.

## Review Focus

1. A caller that mutates the slice returned by `Face` must not corrupt later calls. Test in Task 4 (mutate then re-read).
2. `Files()` served by `http.FS` must reject invalid paths (`../x`, `/abs`, `arabic/..`) with an error rather than panicking or leaking. Test in Task 5 (union FS) and Task 4 (leaf FS via `fs.ValidPath`).
3. A caller that mutates the slice returned by `WebFaces()` or `Fallbacks()` must not change later calls. Test in Task 4 and Task 5.
4. `packs.Embedded.Face("Noto Color Emoji", ...)` must error, not silently serve nothing. Test in Task 5.
5. Variant casing (`"regular"`) and empty strings must error. Test in Task 4.

---

### Task 1: Subset script pack and emoji modes

**Files:**
- Modify: `scripts/subset-fonts.py`
- Create: `packs/sources.json`, `packs/emoji/sources.json`
- Generated: `packs/<script>/NotoSans<Script>-{Regular,Bold}.ttf` (7 scripts), `packs/subsets.sha256`, `packs/coverage.json`, `packs/internal/notice/notosans<script>.OFL.txt` (7 files), `packs/emoji/NotoColorEmoji.ttf`, `packs/emoji/notocoloremoji.OFL.txt`, `packs/emoji/subsets.sha256`

**Interfaces:**
- Produces: the font files and manifests every later task embeds and hashes. File name scheme `NotoSans<Script>-<Weight>.ttf` where `<Script>` is one of `Arabic Hebrew Devanagari Bengali Tamil Thai Khmer` and `<Weight>` is `Regular` or `Bold`. `packs/coverage.json` keys are `<Script>/<Weight>`, values the covered characters as a string (same shape as root `coverage.json`).

- [ ] **Step 1: Create a Python environment**

Run (in the scratchpad directory, not the repo):
```sh
uv venv /private/tmp/claude-501/-Users-evan-projects-ntfonts/cf44024f-be05-4de0-aae5-74514e08fbba/scratchpad/venv --python 3.12
uv pip install --python /private/tmp/claude-501/-Users-evan-projects-ntfonts/cf44024f-be05-4de0-aae5-74514e08fbba/scratchpad/venv/bin/python fonttools==4.60.2
```
Expected: `fonttools==4.60.2` installed. If Python 3.12 is unavailable use the newest version fonttools 4.60.2 supports; record the version used in the commit message.

- [ ] **Step 2: Write `packs/sources.json`**

Fourteen entries in the exact format of root `sources.json` (`file`, `url`, `sha256`). For each `<slug>` in `notosansarabic notosanshebrew notosansdevanagari notosansbengali notosanstamil notosansthai notosanskhmer` with `<Script>` the matching CamelCase name:
- `file`: `<slug>-NotoSans<Script>[wdth,wght].ttf`, `url`: `https://raw.githubusercontent.com/google/fonts/bd8f81ddb5c74d5c8897b36ad88b440266245103/ofl/<slug>/NotoSans<Script>%5Bwdth%2Cwght%5D.ttf`
- `file`: `<slug>-OFL.txt`, `url`: `.../ofl/<slug>/OFL.txt`

Compute each `sha256` by downloading with `curl -sL` into a fresh scratchpad directory and running `shasum -a 256`. Expected upstream sizes (bytes) to sanity-check the downloads: Arabic 844676, Hebrew 112640, Devanagari 641944, Bengali 463668, Tamil 340668, Thai 218652, Khmer 359788; every OFL.txt about 4.4 KB.

- [ ] **Step 3: Write `packs/emoji/sources.json`**

Two entries:
- `file`: `NotoColorEmoji.ttf`, `url`: `https://raw.githubusercontent.com/googlefonts/noto-emoji/6202fe7c20dd5e1727a4c3c01604edc176c576da/fonts/NotoColorEmoji.ttf` (expected 10,673,480 bytes)
- `file`: `notocoloremoji-OFL.txt`, `url`: `https://raw.githubusercontent.com/googlefonts/noto-emoji/6202fe7c20dd5e1727a4c3c01604edc176c576da/fonts/LICENSE`

- [ ] **Step 4: Extend the script**

In `scripts/subset-fonts.py`:
- Add `--packs` and `--emoji` flags (mutually exclusive with `--cjk`; use `add_mutually_exclusive_group`). `output` becomes `root/'packs'` for `--packs` and `root/'packs'/'emoji'` for `--emoji`.
- Give `save` two keyword parameters with defaults that reproduce today's behaviour exactly: `features=TEXT_FEATURES` where `TEXT_FEATURES=['ccmp','locl','mark','mkmk','kern']`, and `prefix='Gloss Text '`. Inside, `options.layout_features=features` and `familyname=prefix+ident`; name ID 6 becomes `familyname.replace(' ','')+'-'+variant` (equals today's `'GlossText'+ident+'-'+variant`). Do not change anything else in `save`.
- Module constants:
  ```python
  PACKS=[('arabic','Arabic',[(0x600,0x6ff),(0x750,0x77f),(0x8a0,0x8ff),(0xfb50,0xfdff),(0xfe70,0xfeff)]),
   ('hebrew','Hebrew',[(0x590,0x5ff),(0xfb1d,0xfb4f)]),
   ('devanagari','Devanagari',[(0x900,0x97f),(0x1cd0,0x1cff),(0xa8e0,0xa8ff)]),
   ('bengali','Bengali',[(0x980,0x9ff)]),
   ('tamil','Tamil',[(0xb80,0xbff),(0x11fc0,0x11fff)]),
   ('thai','Thai',[(0xe00,0xe7f)]),
   ('khmer','Khmer',[(0x1780,0x17ff),(0x19e0,0x19ff)])]
  CONTROLS={0x20,0xa0,0x200b,0x200c,0x200d,0x200e,0x200f,0x2066,0x2067,0x2068,0x2069,0xfffd}
  ```
- Factor the existing `fvar` instancing lines (settings from axis defaults, `wght`, optional `opsz`, `instantiateVariableFont`) into `def instance(font, weight, optical=None)` and call it from the text branch unchanged.
- `--packs` branch: for each pack and each `(variant, weight)` in `[('Regular',400),('Bold',700)]`: open `<slug>-NotoSans<Script>[wdth,wght].ttf`, `instance(font, weight)` (so `wdth` stays at its default), record `coverage[Script+'/'+variant]`, then `save(font, points(ranges)|CONTROLS, Script, variant, output/slug/('NotoSans'+Script+'-'+variant+'.ttf'), features=['*'], prefix='Gloss ')`. Copy each `<slug>-OFL.txt` from the cache to `output/'internal'/'notice'/(slug+'.OFL.txt')`. Write `output/'coverage.json'`.
- `--emoji` branch: no fontTools calls. Copy the cached `NotoColorEmoji.ttf` to `output/'NotoColorEmoji.ttf'` and `notocoloremoji-OFL.txt` to `output/'notocoloremoji.OFL.txt'`; add the font to `hashes` under its file name.
- The final `subsets.sha256` write and the byte-count print stay shared.

- [ ] **Step 5: Run pack and emoji modes**

Run from the repo root with the venv's python:
```sh
<venv>/bin/python scripts/subset-fonts.py <scratch>/ntfonts-sources --download --packs
<venv>/bin/python scripts/subset-fonts.py <scratch>/ntfonts-sources --download --emoji
```
Expected: 14 TTF files under `packs/<slug>/`, `packs/subsets.sha256` with 14 lines, `packs/coverage.json`, seven OFL texts under `packs/internal/notice/`, `packs/emoji/NotoColorEmoji.ttf`, `packs/emoji/subsets.sha256` with one line. Note the printed byte totals; Task 6 and Task 7 need them.

- [ ] **Step 6: Verify pack fonts keep shaping features and public names**

Run:
```sh
<venv>/bin/python -I - <<'EOF'
from fontTools.ttLib import TTFont
for s in ['Arabic','Hebrew','Devanagari','Bengali','Tamil','Thai','Khmer']:
  f=TTFont(f'packs/{s.lower()}/NotoSans{s}-Regular.ttf')
  gsub=[r.FeatureTag for r in f['GSUB'].table.FeatureList.FeatureRecord] if 'GSUB' in f else []
  gpos=[r.FeatureTag for r in f['GPOS'].table.FeatureList.FeatureRecord] if 'GPOS' in f else []
  cmap=f.getBestCmap()
  print(s, f['name'].getDebugName(1), f['maxp'].numGlyphs, len(cmap), sorted(set(gsub)), sorted(set(gpos)), [hex(c) for c in (0x200c,0x200d) if c in cmap], any(0x30<=c<=0x39 for c in cmap))
EOF
```
Expected: family name `Gloss <Script>`; Arabic GSUB includes `init medi fina rlig`; every GPOS includes `mark mkmk`; the digit check prints `False` for all; Arabic `numGlyphs` is well above its cmap count. Record the Devanagari GSUB tags and the per-face `0x200c/0x200d` presence; Task 4's tests use them.

- [ ] **Step 7: Verify text and CJK output are unchanged**

Run:
```sh
<venv>/bin/python scripts/subset-fonts.py <scratch>/ntfonts-sources --download
<venv>/bin/python scripts/subset-fonts.py <scratch>/ntfonts-sources --download --cjk
git status --short -- '*.ttf' '*.otf' subsets.sha256 coverage.json cjk/
```
Expected: no modified tracked files outside `packs/`. If any root or cjk file changed, the `save` refactor altered output; fix before continuing.

- [ ] **Step 8: Verify emoji table directory**

Run:
```sh
python3 -I - <<'EOF'
import struct;b=open('packs/emoji/NotoColorEmoji.ttf','rb').read(4096);n=struct.unpack('>H',b[4:6])[0]
print(sorted(b[12+16*i:16+16*i].decode() for i in range(n)))
EOF
```
Expected: includes `CBDT` and `CBLC`; excludes `COLR`, `CPAL`, `SVG `.

- [ ] **Step 9: Commit**

```bash
git add scripts/subset-fonts.py packs/
git commit -m "feat(packs): generate Noto Sans script subsets and pin Noto Color Emoji"
```

---

### Task 2: packs module scaffold and shared internals

**Files:**
- Create: `packs/go.mod`, `packs/LICENSE` (copy of `cjk/LICENSE`), `packs/.gitignore` (copy of `cjk/.gitignore`)
- Create: `packs/internal/notice/notice.go`, `packs/internal/notice/notice_test.go`
- Create: `packs/internal/pack/pack.go`, `packs/internal/pack/pack_test.go`

**Interfaces:**
- Produces:
  ```go
  // package notice
  type License struct{ Family, Text string }
  func Licenses(families ...string) []License   // one entry per family, in order; panics on unknown family
  func Families() []string                      // the seven script families in spec order
  // package pack
  type WebFace struct{ Family, File string; Weight int; Style string }
  type Pack struct{ Family, Base string; FS embed.FS }   // Base e.g. "NotoSansArabic"
  func (p Pack) Face(family, variant string) ([]byte, error)
  func (p Pack) Files() fs.FS
  func (p Pack) WebFaces() []WebFace
  ```

- [ ] **Step 1: Write `packs/go.mod`**

```
module nimbleterminal.dev/ntfonts/packs

go 1.26
```

- [ ] **Step 2: Write the failing notice test**

`packs/internal/notice/notice_test.go`, package `notice`:
- `TestLicenses`: `Licenses("Noto Sans Arabic","Noto Sans Khmer")` returns two entries in that order, each `Text` contains `SIL OPEN FONT LICENSE`; `Families()` equals `[]string{"Noto Sans Arabic","Noto Sans Hebrew","Noto Sans Devanagari","Noto Sans Bengali","Noto Sans Tamil","Noto Sans Thai","Noto Sans Khmer"}`; `Licenses(Families()...)` has 7 entries; a returned slice mutated (`got[0].Text = ""`) does not change a fresh call; `Licenses("Nope")` panics (use `defer func(){ if recover()==nil { t.Fatal() } }()`).

- [ ] **Step 3: Run it to verify it fails**

Run: `cd packs && GOWORK=off go test ./internal/notice/`
Expected: FAIL, package does not build.

- [ ] **Step 4: Implement `notice.go`**

`//go:embed *.OFL.txt` into an `embed.FS`; a table mapping family to file slug (`notosansarabic.OFL.txt` ...); `Licenses` reads each file (panic on error, as `ntfonts.Licenses` does: embedded at build time).

- [ ] **Step 5: Run it to verify it passes**

Run: `cd packs && GOWORK=off go test ./internal/notice/`
Expected: PASS

- [ ] **Step 6: Write the failing pack helper test**

`packs/internal/pack/pack_test.go`, package `pack`. `Pack.FS` is an `embed.FS`, so the test embeds its own fixtures: a test-only `//go:embed testdata/*` over two tiny files `testdata/Fake-Regular.ttf` and `testdata/Fake-Bold.ttf` with distinct bytes, and `p := Pack{Family: "Fake", Base: "testdata/Fake", FS: testFS}`.
- `TestFace`: Regular and Italic return the Regular bytes; Bold and BoldItalic return the Bold bytes; `("Fake","regular")`, `("Fake","")`, `("Other","Regular")` error with messages `packs: unknown variant "regular"`, `packs: unknown variant ""`, `packs: unknown family "Other"`; mutating a returned slice then re-reading yields original bytes.
- `TestWebFaces`: equals `[]WebFace{{"Fake","testdata/Fake-Regular.ttf",400,"normal"},{"Fake","testdata/Fake-Bold.ttf",700,"normal"}}`; mutating the returned slice does not affect a fresh call.
- `TestFiles`: `fs.ReadFile(p.Files(), "testdata/Fake-Bold.ttf")` equals the Bold bytes.

- [ ] **Step 7: Run it to verify it fails**

Run: `cd packs && GOWORK=off go test ./internal/pack/`
Expected: FAIL, package does not build.

- [ ] **Step 8: Implement `pack.go`**

`Face` validates the variant first (switch on the four names), then the family, then `p.FS.ReadFile(p.Base + "-" + weight + ".ttf")` where weight is `Regular` for Regular/Italic and `Bold` otherwise. `WebFaces` builds a fresh two-element slice each call. `Files` returns `p.FS`.

- [ ] **Step 9: Run it to verify it passes**

Run: `cd packs && GOWORK=off go test ./internal/...`
Expected: PASS

- [ ] **Step 10: Commit**

```bash
git add packs/go.mod packs/LICENSE packs/.gitignore packs/internal/
git commit -m "feat(packs): module scaffold with shared notice and pack helpers"
```

---

### Task 3: Seven script packages

**Files:**
- Create: `packs/arabic/arabic.go`, `packs/hebrew/hebrew.go`, `packs/devanagari/devanagari.go`, `packs/bengali/bengali.go`, `packs/tamil/tamil.go`, `packs/thai/thai.go`, `packs/khmer/khmer.go`
- Test: `packs/scripts_test.go` (package `packs_test`, written in Task 4; this task adds a build-only check)

**Interfaces:**
- Consumes: `pack.Pack`, `pack.WebFace`, `notice.License`, `notice.Licenses` from Task 2.
- Produces, identically in each package `<slug>`:
  ```go
  type Source struct{}
  var Embedded Source
  func Fallbacks() []string              // {"Noto Sans <Script>"}
  func (Source) Fallbacks() []string
  func (Source) Face(family, variant string) ([]byte, error)
  func Files() fs.FS
  type WebFace = pack.WebFace
  func WebFaces() []WebFace
  type License = notice.License
  func Licenses() []License              // notice.Licenses("Noto Sans <Script>")
  ```

- [ ] **Step 1: Write `packs/arabic/arabic.go`**

Package comment: `// Package arabic embeds Noto Sans Arabic Regular and Bold subsets. Italic variants use the upright weights.` Then `//go:embed NotoSansArabic-Regular.ttf NotoSansArabic-Bold.ttf` into `var files embed.FS`; `var p = pack.Pack{Family: "Noto Sans Arabic", Base: "NotoSansArabic", FS: files}`; the functions above delegate to `p`. `Fallbacks` returns a fresh slice each call.

- [ ] **Step 2: Create the other six by the same pattern**

Change only the package name, the comment, the embed names, `Family` and `Base`.

- [ ] **Step 3: Build**

Run: `cd packs && GOWORK=off go build ./... && GOWORK=off go vet ./...`
Expected: success, no output.

- [ ] **Step 4: Commit**

```bash
git add packs/*/
git commit -m "feat(packs): add seven Noto Sans script packages"
```

---

### Task 4: Script package tests with font-table parsers

**Files:**
- Create: `packs/fontparse_test.go`, `packs/scripts_test.go` (both package `packs_test`)

**Interfaces:**
- Produces (test helpers reused by Task 5 and copied by Task 7):
  ```go
  func tables(data []byte) (map[string][]byte, error)        // tag -> table bytes; errors unless sfnt 0x00010000 or "OTTO"
  func parseCmap(data []byte) (func(rune) bool, error)       // copied verbatim from cjk/cjk_test.go
  func featureTags(table []byte) map[string]bool              // GSUB/GPOS FeatureList tags
  func numGlyphs(data []byte) int                             // maxp uint16 at offset 4
  ```

- [ ] **Step 1: Write `fontparse_test.go`**

Copy `parseCmap` from `cjk/cjk_test.go`. Add `tables` (iterate the table directory: numTables at offset 4, records at 12+16i with tag, checksum, offset, length). Add `featureTags`: header `majorVersion u16, minorVersion u16, scriptListOffset u16, featureListOffset u16`; FeatureList at that offset: `featureCount u16` then `featureCount` records of `tag[4] + offset u16`. Add `numGlyphs`.

Self-test `TestParsersOnCore`: open `../GlossG-Regular.ttf` (JetBrains Mono subset, present in the checkout) and assert `tables` has `cmap` and `GPOS`, `parseCmap` finds `'A'`, `featureTags(tables["GPOS"])["kern"]` is true, `numGlyphs > 0`.

- [ ] **Step 2: Write `scripts_test.go`**

A table `var scripts = []struct{ slug, family, base string; face func(string,string)([]byte,error); fallbacks func() []string; files func() fs.FS; webFaces func() []arabic.WebFace; licenses func() []arabic.License; row string }`. `pack` is internal, so the test names the types through `arabic`'s public aliases; every package's alias is the same underlying type, so the other packages' functions assign directly. Rows: one per package, with `row string` holding the ntwebterm corpus text verbatim (source: `ntwebterm/demo/multilingual-corpus.js` at 0.4.2; cite in a comment):

| slug | row |
| --- | --- |
| arabic | `Arabic: مرحبا بالعالم · اللغة العربية · لا · تَجْرِبَة · فارسی` |
| hebrew | `Hebrew: שלום עולם · עברית` |
| devanagari | `Devanagari: नमस्ते दुनिया · क्ष त्र ज्ञ` |
| bengali | `Bengali: বাংলা ভাষা · শুভ সন্ধ্যা` |
| tamil | `Tamil: தமிழ் மொழி · வணக்கம்` |
| thai | `Thai: สวัสดีชาวโลก · ภาษาไทย` |
| khmer | `Khmer: ជម្រាបសួរ ពិភពលោក · ភាសាខ្មែរ` |

Tests, each iterating the table with `t.Run(slug, ...)`:
- `TestFacesAndAliases`: `fallbacks()` equals `[]string{family}` and mutating the result does not change a fresh call; all four variants succeed; Italic bytes equal Regular, BoldItalic equal Bold, Regular not equal Bold; `face(family,"regular")`, `face(family,"")`, `face("Nope","Regular")` error and the message contains the offending `%q`; mutate a returned slice then re-read equals original.
- `TestCorpusCoverage`: for Regular and Bold, `parseCmap` holds every rune of `row` with `r >= 0x80 && r != '·'`.
- `TestExclusions`: for Regular and Bold, no rune in `0123456789AaZz` is in the cmap.
- `TestControls`: read `coverage.json`; for each face the presence of U+200C and U+200D in the cmap equals their presence in the `coverage.json` string for `<Script>/<Weight>`.
- `TestLayoutFeatures`: Regular and Bold of every face have `mark` and `mkmk` in GPOS; Arabic has `init`, `medi`, `fina`, `rlig` in GSUB; Devanagari has `akhn` plus the tags recorded in Task 1 Step 6 (write them as a literal list); Arabic `numGlyphs(data) > cmapCount*3/2` where `cmapCount` counts mapped code points across the Arabic ranges from the spec's coverage table. If the measured ratio is lower, lower the factor to the largest tenth below the measurement that is still above 1.0 and record both numbers in a comment.
- `TestFilesAndWebFaces`: `webFaces()` equals `{{family, base+"-Regular.ttf", 400, "normal"},{family, base+"-Bold.ttf", 700, "normal"}}`; every `File` opens through `files()` and its bytes equal `face(family, weight)`; `fs.ReadDir(files(), ".")` lists exactly those two names; `files().Open("../x")` and `Open("/abs")` return errors; mutating `webFaces()` output does not affect a fresh call.
- `TestLicenses`: each package's `Licenses()` is one entry, `Family == family`, text contains `SIL OPEN FONT LICENSE`.
- `TestHashes`: read `subsets.sha256`, expect 14 lines, each `sha256(os.ReadFile(path))` matches (same shape as `cjk.TestHashes`).
- `TestBudget`: sum of all 14 embedded files `<= budget` where `const budget = <Task 1 Step 5 packs total × 1.25, rounded up to the next 100,000>`; put the measured total in a comment.

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd packs && GOWORK=off go test ./...`
Expected: FAIL (compile errors until `scripts_test.go` references resolve, then any assertion failures from real data). Fix only test bugs here; data failures mean Task 1 is wrong and must be regenerated.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd packs && GOWORK=off go test ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add packs/fontparse_test.go packs/scripts_test.go
git commit -m "test(packs): coverage, feature, hash and web surface tests for script packs"
```

---

### Task 5: packs root package with union file system and link isolation

**Files:**
- Create: `packs/packs.go`, `packs/union.go`, `packs/packs_test.go`, `packs/isolation_test.go`

**Interfaces:**
- Consumes: the seven packages from Task 3.
- Produces:
  ```go
  // package packs
  type Source struct{}
  var Embedded Source
  func Fallbacks() []string                 // seven families in spec order
  func (Source) Fallbacks() []string
  func (Source) Face(family, variant string) ([]byte, error)
  func Files() fs.FS                        // union of the seven Files()
  type WebFace = pack.WebFace
  func WebFaces() []WebFace                 // 14 entries, Regular then Bold per family, spec order
  type License = notice.License
  func Licenses() []License                 // notice.Licenses(Fallbacks()...)
  // union.go (unexported)
  type union []fs.FS                        // Open(name): "." lists every entry; other names try each FS in order
  ```

- [ ] **Step 1: Write the failing tests**

`packs_test.go`:
- `TestRootDispatch`: `packs.Fallbacks()` equals the seven-family list; for each family and variant, `packs.Embedded.Face` bytes equal the corresponding leaf package's `Face`; `Face("Noto Color Emoji","Regular")` errors with `packs: unknown family "Noto Color Emoji"`; `Face("Noto Sans Arabic","Heavy")` errors.
- `TestRootWebFaces`: `len(packs.WebFaces()) == 14`; order is arabic Regular, arabic Bold, hebrew Regular, ...; every `File` opens via `packs.Files()` with bytes equal to `Face`.
- `TestRootFiles`: `fstest.TestFS(packs.Files(), <all 14 names>...)` passes; `Open("missing.ttf")` returns an error satisfying `errors.Is(err, fs.ErrNotExist)`; `Open("../x")` returns an error.
- `TestRootLicenses`: 7 entries, families in `Fallbacks()` order.

`isolation_test.go`, adapted from `cjk/isolation_test.go`:
- `TestLinkIsolation` (skipped under `-short`): signature bytes from `arabic/NotoSansArabic-Regular.ttf` not present in `hebrew/NotoSansHebrew-Regular.ttf` and vice versa; `go list -deps ./arabic` contains neither `nimbleterminal.dev/ntfonts/packs` nor `.../packs/hebrew`; a temp consumer module (`require nimbleterminal.dev/ntfonts/packs v0.0.0` + `replace` to the checkout) importing only `packs/arabic` builds, its binary contains the Arabic signature and not the Hebrew one.

- [ ] **Step 2: Run to verify failure**

Run: `cd packs && GOWORK=off go test -short ./...`
Expected: FAIL, package `packs` does not exist.

- [ ] **Step 3: Implement `union.go` and `packs.go`**

`union.Open`: reject `!fs.ValidPath(name)` with `&fs.PathError{Op:"open", Path:name, Err:fs.ErrInvalid}`; for `"."` return a `unionDir` whose `ReadDir(n)` returns the concatenated `fs.ReadDir(sub, ".")` entries (n <= 0 returns all; n > 0 pages), `Stat` returns a synthetic `fs.FileInfo` (name ".", `fs.ModeDir`), `Read` returns an error, `Close` nil; otherwise try each sub FS and return the first success, else `&fs.PathError{Op:"open", Path:name, Err:fs.ErrNotExist}`.

`packs.go`: a table `var all = []struct{ family string; face func(string,string)([]byte,error); files func() fs.FS; webFaces func() []WebFace }` in spec order; `Face` validates the variant with the same switch as `pack.Pack.Face` before looking up the family so `("Noto Sans Arabic","Heavy")` reports the variant; `Files()` returns `union{...}` built from the table; `WebFaces` concatenates.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd packs && GOWORK=off go test ./...` (not `-short`, so isolation runs)
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add packs/packs.go packs/union.go packs/packs_test.go packs/isolation_test.go
git commit -m "feat(packs): combined source with union file system and link isolation tests"
```

---

### Task 6: packs tooling, notices, docs and root CI wiring

**Files:**
- Create: `packs/Taskfile.yml`, `packs/internal/tools/notices/main.go`, `packs/THIRD_PARTY_NOTICES.md`, `packs/README.md`, `packs/CHANGELOG.md`
- Modify: `Taskfile.yml` (root), `.github/workflows/pages.yml`

- [ ] **Step 1: Copy `cjk/Taskfile.yml` to `packs/Taskfile.yml`**

Change only descriptions that say CJK. Keep every task name.

- [ ] **Step 2: Copy `cjk/internal/tools/notices/main.go` to `packs/internal/tools/notices/main.go`**

Change the import to `packs "nimbleterminal.dev/ntfonts/packs"`, the header paragraph to name `ntfonts/packs`, and the section note to: `OFL-1.1; pinned sources and subset recipe in sources.json and ../scripts/subset-fonts.py; all OpenType layout features retained; embedded under distinct Gloss <Script> internal names`.

- [ ] **Step 3: Generate notices and run the gate**

Run: `cd packs && task notices && task ci`
Expected: `THIRD_PARTY_NOTICES.md` written with the Go notice and seven OFL sections; `task ci` passes (vulncheck needs network).

- [ ] **Step 4: Write `packs/README.md`**

Sections, in order: purpose and the `go get nimbleterminal.dev/ntfonts/packs/arabic` example with `arabic.Embedded.Face(arabic.Fallbacks()[0], "Regular")`; the import-path table with per-pack raw byte sizes (Regular + Bold) and the module total, followed by the sentence `One module zip contains every script face; a per-script import saves linked binary size, not module download size.`; family names; the coverage table from the spec plus the controls set; the shaping-features paragraph (all features retained; nttypeset disables `liga`, `clig`, `dlig` and `calt` at runtime so Go shaping relies on `init`/`medi`/`fina`/`rlig`, Indic and USE features; browsers apply everything); the web delivery example:
```go
mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.FS(arabic.Files()))))
for _, f := range arabic.WebFaces() { /* emit {family: f.Family, src: "/fonts/"+f.File, weight: f.Weight, style: f.Style} */ }
```
with the matching ntwebterm `fontFaces` JSON; regeneration (`python scripts/subset-fonts.py /tmp/ntfonts-sources --download --packs`, `task --dir packs notices`, `task --dir packs ci`); the sentence that automated cmap and feature checks do not establish visual quality, the specimen checks coverage and not shaping, and no visual check has been performed; release tags `packs/vX.Y.Z`; licenses (MIT code, OFL fonts).

- [ ] **Step 5: Write `packs/CHANGELOG.md`**

`# packs CHANGELOG`, `## Unreleased`, one bullet per: initial module with the seven packages and sizes; combined `packs` source; `Files()`/`WebFaces()` surface; full layout-feature retention; pinned sources commit.

- [ ] **Step 6: Wire root tasks and Pages**

Root `Taskfile.yml`: add
```yaml
  packs-ci:
    desc: Run the script packs module gate including link-isolation tests
    cmds: ['task --dir packs ci']
```
and `- task: packs-ci` after `cjk-ci` in `ci`. In `.github/workflows/pages.yml`, extend `tags:` to `['v*', 'cjk/v*', 'packs/v*', 'packs/emoji/v*']`.

- [ ] **Step 7: Run the root gate**

Run: `task ci`
Expected: PASS including `packs-ci`.

- [ ] **Step 8: Commit**

```bash
git add packs/Taskfile.yml packs/internal/tools packs/THIRD_PARTY_NOTICES.md packs/README.md packs/CHANGELOG.md Taskfile.yml .github/workflows/pages.yml
git commit -m "chore(packs): task gate, notices, docs and root CI wiring"
```

---

### Task 7: emoji module

**Files:**
- Create: `packs/emoji/go.mod`, `packs/emoji/LICENSE`, `packs/emoji/.gitignore`, `packs/emoji/emoji.go`, `packs/emoji/emoji_test.go`, `packs/emoji/Taskfile.yml`, `packs/emoji/internal/tools/notices/main.go`, `packs/emoji/THIRD_PARTY_NOTICES.md`, `packs/emoji/README.md`, `packs/emoji/CHANGELOG.md`
- Modify: `Taskfile.yml` (root)

**Interfaces:**
- Produces:
  ```go
  // package emoji (module nimbleterminal.dev/ntfonts/packs/emoji)
  type Source struct{}
  var Embedded Source
  func Fallbacks() []string                  // {"Noto Color Emoji"}
  func (Source) Fallbacks() []string
  func (Source) Face(family, variant string) ([]byte, error)   // same file for all four variants; errors "emoji: unknown ..."
  func Files() fs.FS
  type WebFace struct{ Family, File string; Weight int; Style string }
  func WebFaces() []WebFace                  // {{"Noto Color Emoji","NotoColorEmoji.ttf",400,"normal"}}
  type License struct{ Family, Text string }
  func Licenses() []License                  // one entry from notocoloremoji.OFL.txt
  ```

- [ ] **Step 1: Write `go.mod`, copy `LICENSE` and `.gitignore` from `cjk/`**

```
module nimbleterminal.dev/ntfonts/packs/emoji

go 1.26
```

- [ ] **Step 2: Write the failing test `emoji_test.go`** (package `emoji_test`)

Copy `tables` and `parseCmap` from `packs/fontparse_test.go` (separate module; duplication is required).
- `TestFormat`: `tables(data)` has `CBDT` and `CBLC`; lacks `COLR`, `CPAL`, `SVG `.
- `TestFaces`: all four variants return identical bytes; `Fallbacks()` equals `{"Noto Color Emoji"}`; `Face("Noto Color Emoji","regular")`, `Face("Noto Color Emoji","")`, `Face("Noto Sans Arabic","Regular")` error with messages containing the `%q` offender and prefix `emoji:`; mutate then re-read equals original.
- `TestCorpusCoverage`: cmap holds each of U+1F600, U+1F680, U+1F431, U+1F468, U+1F469, U+1F467, U+1F466, U+1F1FA, U+1F1F3, U+1F44D, U+1F3FD (the code points of the corpus row `Emoji: 😀 🚀 🐱 👨‍👩‍👧‍👦 🇺🇳 👍🏽`).
- `TestFilesAndWebFaces`: `WebFaces()` equals the single entry above; `fs.ReadFile(Files(),"NotoColorEmoji.ttf")` equals `Face` bytes; `Open("../x")` errors; mutated `WebFaces()` output does not affect a fresh call.
- `TestLicenses`: one entry, `Family == "Noto Color Emoji"`, text contains `SIL OPEN FONT LICENSE`.
- `TestHashes`: `subsets.sha256` has exactly one line and it matches.
- `TestBudget`: `len(data) <= 12_000_000`.

- [ ] **Step 3: Run to verify failure**

Run: `cd packs/emoji && GOWORK=off go test ./...`
Expected: FAIL, package does not build.

- [ ] **Step 4: Implement `emoji.go`**

`//go:embed NotoColorEmoji.ttf` and `//go:embed notocoloremoji.OFL.txt` (two `embed.FS` or one plus a `string`). Package comment states: unmodified CBDT build from googlefonts/noto-emoji v2.051; Safari does not render CBDT and substitutes the platform emoji font; Chrome and Firefox render it.

- [ ] **Step 5: Run to verify pass**

Run: `cd packs/emoji && GOWORK=off go test ./...`
Expected: PASS

- [ ] **Step 6: Tooling and notices**

Copy `cjk/Taskfile.yml` to `packs/emoji/Taskfile.yml` (descriptions only). Copy the notices tool with import `emoji "nimbleterminal.dev/ntfonts/packs/emoji"`, header naming `ntfonts/packs/emoji`, section note: `OFL-1.1; unmodified NotoColorEmoji.ttf (CBDT) from googlefonts/noto-emoji v2.051, pinned in sources.json`. Run `cd packs/emoji && task notices && task ci`. Expected: PASS.

- [ ] **Step 7: README and CHANGELOG**

`packs/emoji/README.md`: purpose; `go get nimbleterminal.dev/ntfonts/packs/emoji`; raw byte size 10,673,480 and the sentence `These sizes are raw font payloads added to native or js/wasm binaries, before compression and other code/data overhead.`; why CBDT (nttypeset draws bitmap glyphs, not COLR); the Safari note; why unmodified (no fontTools rewrite, OFL permits redistribution with notice); the source tag/commit and that google/fonts' `notocoloremoji` is COLRv1 and unsuitable; regeneration `--emoji`; web delivery example mirroring the packs README; the visual-quality sentence; release tag `packs/emoji/vX.Y.Z`. `packs/emoji/CHANGELOG.md` with `## Unreleased` and the initial bullet.

- [ ] **Step 8: Root wiring**

Root `Taskfile.yml`: add `emoji-ci` (`task --dir packs/emoji ci`) and `- task: emoji-ci` after `packs-ci` in `ci`. Run `task ci`. Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add packs/emoji Taskfile.yml
git commit -m "feat(emoji): add Noto Color Emoji CBDT module"
```

---

### Task 8: cjk Files() and WebFaces()

**Files:**
- Modify: `cjk/scregular/scregular.go`, `cjk/scbold/scbold.go`, `cjk/cjk.go`, `cjk/internal/notice/notice.go`, `cjk/CHANGELOG.md`, `cjk/README.md`
- Create: `cjk/union.go`
- Test: `cjk/isolation_test.go` (extend `TestLeafFaces`), `cjk/cjk_test.go` (add `TestFilesAndWebFaces`)

**Interfaces:**
- Produces:
  ```go
  // notice
  type WebFace struct{ Family, File string; Weight int; Style string }
  // scregular, scbold, cjk
  func Files() fs.FS
  type WebFace = notice.WebFace
  func WebFaces() []WebFace   // scregular: {{"Noto Sans CJK SC","GlossCJKSC-Regular.otf",400,"normal"}}; scbold: 700 Bold; cjk: both, Regular first
  ```

- [ ] **Step 1: Write the failing tests**

In `cjk/cjk_test.go` add `TestFilesAndWebFaces`: `cjk.WebFaces()` equals the two entries above; each `File` read through `cjk.Files()` equals `cjk.Embedded.Face("Noto Sans CJK SC", weight)`; `fstest.TestFS(cjk.Files(), "GlossCJKSC-Regular.otf", "GlossCJKSC-Bold.otf")` passes; `Open("../x")` errors. In `TestLeafFaces` add per leaf: `WebFaces()` is the single expected entry, `fs.ReadFile(Files(), entry.File)` equals the face bytes.

- [ ] **Step 2: Run to verify failure**

Run: `cd cjk && GOWORK=off go test -short ./...`
Expected: FAIL to build.

- [ ] **Step 3: Implement**

Add `WebFace` to `cjk/internal/notice/notice.go`. In each leaf: `func Files() fs.FS { return face }` and `WebFaces()` returning a fresh one-element slice. `cjk/union.go`: copy `packs/union.go` with package `cjk`. `cjk.Files()` returns `union{scregular.Files(), scbold.Files()}`; `cjk.WebFaces()` concatenates.

- [ ] **Step 4: Run to verify pass**

Run: `cd cjk && GOWORK=off go test ./...` then `task --dir cjk ci`
Expected: PASS

- [ ] **Step 5: Docs**

`cjk/CHANGELOG.md` under `## Unreleased`: `Add Files() and WebFaces() to cjk, scregular and scbold so hosts can serve the SC faces to a browser terminal through the same path as the script packs.` `cjk/README.md`: one paragraph after the API list describing the two functions with the `http.FS` example.

- [ ] **Step 6: Commit**

```bash
git add cjk/
git commit -m "feat(cjk): expose Files() and WebFaces() for browser delivery"
```

---

### Task 9: Specimen catalog for packs

**Files:**
- Modify: `internal/tools/specimen/main.go`, `internal/tools/specimen/index.html`, `internal/tools/specimen/main_test.go`

- [ ] **Step 1: Extend the failing test**

In `TestViewer`, after creating the fake CJK files, create `root/packs/arabic/NotoSansArabic-Regular.ttf` (`"fake arabic"`), `root/packs/internal/notice/notosansarabic.OFL.txt` (`"test arabic license"`), `root/packs/emoji/NotoColorEmoji.ttf` (`"fake emoji"`), `root/packs/emoji/notocoloremoji.OFL.txt` (`"test emoji license"`). Expect `len(fonts) == 32`; faces with `Group == "packs"` must have `Family` in `{"Noto Sans Arabic","Noto Color Emoji"}`, URLs ending `.ttf`, and bodies equal to the fake bytes. `request("/licenses.txt")` body contains `test arabic license` and `test emoji license`. The absent-module path still yields 29.

- [ ] **Step 2: Run to verify failure**

Run: `GOWORK=off go test ./internal/tools/specimen/`
Expected: FAIL on the 32 count.

- [ ] **Step 3: Implement discovery in `siteFiles`**

After the CJK loop: a table of `{slug, script, family}` for the seven scripts; for each and for `Regular`, `Bold`: read `filepath.Join(root,"packs",slug,"NotoSans"+script+"-"+variant+".ttf")`, skip `os.IsNotExist`, append `face{Family: family, Variant: variant, Group: "packs"}`. Then `packs/emoji/NotoColorEmoji.ttf` as `Family: "Noto Color Emoji", Variant: "Regular", Group: "packs"`. Licenses: for each pack family present, append `packs/internal/notice/<slug>.OFL.txt`; for emoji, `packs/emoji/notocoloremoji.OFL.txt`. Extension stays `.ttf` for `packs`.

- [ ] **Step 4: Page checkbox**

In `index.html`: add `<label class="check"><input id="packs" type="checkbox"> Script and emoji packs</label>` after the CJK label. In `render`, when `$('packs').checked`, push `pick(f.family, variant)` for every distinct `catalog` family with `group==='packs'`. In `init`, disable the checkbox with title `Optional script and emoji packs were not found in this checkout` when no `packs` faces exist, and add `'packs'` to the change-listener list. Update the footer sentence to add: `Pack faces are rendered one code point at a time: this checks coverage, not shaping.`

- [ ] **Step 5: Run to verify pass**

Run: `GOWORK=off go test ./internal/tools/specimen/ && task specimen-build OUT=<scratch>/specimen-site`
Expected: tests PASS; export contains 29 + 14 + 1 = 44 catalog entries (check `jq length <scratch>/specimen-site/fonts.json`).

- [ ] **Step 6: Commit**

```bash
git add internal/tools/specimen/
git commit -m "feat(specimen): catalog script and emoji pack faces from the checkout"
```

---

### Task 10: Root documentation and final gate

**Files:**
- Modify: `README.md`, `AGENTS.md`, `DEVELOP.md`, `CHANGELOG.md`

- [ ] **Step 1: README**

After the CJK subsection add `### Script and emoji packs`: what the two modules are, the `go get` lines, the family list, the sentence about shaping features, pointers to `packs/README.md` and `packs/emoji/README.md`, and the coverage-not-shaping note for the specimen. Update the Specimen viewer section's fallback sentence to mention packs.

- [ ] **Step 2: AGENTS.md**

Module layout table: add rows for `packs/` and `packs/emoji/`. Commands: add `task packs-ci emoji-ci`. Invariants: add `Pack subsets keep every layout feature (unlike text subsets); the emoji file is embedded unmodified and must have CBDT/CBLC, never COLR or SVG.` and `Files on disk use public Noto names; only name tables say Gloss.` Update the opening paragraph's description of what is embedded.

- [ ] **Step 3: DEVELOP.md**

Regeneration section: add the `--packs` and `--emoji` commands, the noto-emoji pin, and `task --dir packs notices`, `task --dir packs/emoji notices`. Releases: tags `packs/vX.Y.Z` and `packs/emoji/vX.Y.Z`. Specimen section: packs are catalogued from the checkout, coverage only.

- [ ] **Step 4: CHANGELOG.md**

Under `## Unreleased` add bullets: new `packs` module; new `packs/emoji` module; subset script `--packs`/`--emoji` modes; specimen catalogs pack faces; root `task ci` runs both new gates.

- [ ] **Step 5: Final gate**

Run: `task ci && git status --short`
Expected: PASS; only intended files modified or added; no untracked font caches inside the repo.

- [ ] **Step 6: Commit**

```bash
git add README.md AGENTS.md DEVELOP.md CHANGELOG.md
git commit -m "docs: describe script and emoji packs"
```
