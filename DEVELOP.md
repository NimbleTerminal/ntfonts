# Development

Use Go 1.27.3 for development and CI (the module minimum is Go 1.26). The root
module and the optional nested modules `cjk/`, `packs/` and `packs/emoji/` have
no dependencies. The root `task ci` also runs the three nested module gates
(`cjk-ci`, `packs-ci`, `emoji-ci`). Installed QA tools live in
`.task/bin/` and generated output in
`bin/` and `dist/`; all are ignored.

```sh
task ci                  # same gate as GitHub Actions; includes network scans
task test                # tests
task test-race           # tests under the race detector
task lint                # golangci-lint v2.15.0, installed into .task/bin
task fonts-check         # shasum -a 256 -c subsets.sha256 (sha256sum as a fallback)
task cross-windows       # check Windows compilation
task cross-wasm          # check js/wasm compilation
task notices             # regenerate THIRD_PARTY_NOTICES.md
task clean               # remove generated artifacts and installed QA tools
```

Install [Go Task](https://taskfile.dev) and a C compiler for the race detector.
`task --list-all` lists maintenance tasks. CI runs on Linux and macOS. Tasks
explicitly use `GOWORK=off`.

`task vulncheck` installs govulncheck v1.9.0 into `.task/bin` and scans the
module. It requires network access to the Go vulnerability database; failures
stop CI.

## Dependencies

None. The module is standard-library only and never depends on nttypeset,
ntpicture, Bubble Tea, or gloss, even in tests.

## Regenerating the font subsets

Subsets change only through the script, with Python and `fonttools==4.60.2`:

```sh
python scripts/subset-fonts.py /tmp/ntfonts-sources --download
```

Google Fonts sources are pinned to commit
`bd8f81ddb5c74d5c8897b36ad88b440266245103` and Iosevka to the v34.9.0 TTF
release; `sources.json` records exact URLs and SHA-256 hashes, and the script
verifies every download. It rewrites the 29 `Gloss*.ttf` files, the source-derived
`coverage.json`, and
`subsets.sha256`. Build CJK separately with the same command plus `--cjk`;
that uses `cjk/sources.json`, writes one OTF each under `cjk/scregular/` and `cjk/scbold/`, plus
`cjk/subsets.sha256`,
and reads the checked-in `cjk/data/gb2312-han.txt`. CJK is pinned to Noto Sans
CJK Sans2.004; symbols and supplemental math glyphs use the same Google Fonts
commit as the text faces. After regenerating, run `task fonts-check` and `task ci`,
update CHANGELOG.md, and review the diff of `subsets.sha256`. The recipe is
described in the Fonts section of README.md.

Build the script packs with `--packs` (uses `packs/sources.json`, writes
`NotoSans<Script>-{Regular,Bold}.ttf` under `packs/<script>/`, the OFL texts
under `packs/internal/notice/`, `packs/coverage.json` and `packs/subsets.sha256`,
retaining every layout feature) and the emoji module with `--emoji` (copies the
hash-verified CBDT `NotoColorEmoji.ttf` from googlefonts/noto-emoji v2.051 into
`packs/emoji/` unmodified). The three modes are mutually exclusive; text mode
output must stay byte-identical when the script changes.

Run `task notices`, `task --dir cjk notices`, `task --dir packs notices` and
`task --dir packs/emoji notices` after font or license changes,
then review and commit
THIRD_PARTY_NOTICES.md. `task notices-check` enforces that generated text in CI.

## Releases

This is a Go library: CI runs for branches, pull requests, and tags. Release
publication is manual, and there is no binary packaging.

Before the first release, confirm that the `nimbleterminal.dev/ntfonts`
`go-import` metadata resolves to `github.com/NimbleTerminal/ntfonts`, and
confirm module discovery from a fresh consumer without workspace or local
replacements. Update CHANGELOG.md, run `task ci`, and review font licenses. Then
commit the release preparation, tag the intended version, wait for tag CI to
pass, and publish the GitHub release with its changelog.

All four modules release together under one version number (release
harmony): every release creates the tags `vX.Y.Z`, `cjk/vX.Y.Z`,
`packs/vX.Y.Z` and `packs/emoji/vX.Y.Z` on the same commit, and every module's
CHANGELOG gets a heading for that version even when its code did not change.
Parent tags do not release the nested modules on their own, and the vanity
redirect for the first path segment covers them all. Nested modules have no
`require` on each other, so no go.mod edits are needed to bump.

## Specimen viewer

Run `task specimen`, then open <http://127.0.0.1:8088>. Stop with Ctrl-C.
Use `task specimen -- -addr 127.0.0.1:8090` to choose another port.
The viewer has family/variant and size controls, editable text, multilingual
samples, and side-by-side comparisons with symbol and SC CJK fallbacks.
It reads each font's cmap and marks missing characters explicitly, preventing
browser system-font fallback from disguising missing coverage. Hover a glyph
for its source face or missing codepoint. Browser shaping is not nttypeset layout.

Core fonts come from `Embedded`; optional CJK, script pack and emoji files are
read from the checkout without adding a module dependency. Pack faces are
rendered one code point at a time, so the viewer checks their coverage, not
their shaping. Only selected faces are loaded by the
browser. All assets stay local. The viewer is a development tool; it does not
change the library API or embed CJK in the core library.

## GitHub Pages

`.github/workflows/pages.yml` runs the full CI gate, exports the specimen, and
deploys it on pushes to `main`, core version tags (`v*`), CJK version tags
(`cjk/v*`), and pack version tags (`packs/v*`, `packs/emoji/v*`). It also supports manual runs. All triggers update the same Pages
site; the most recent successful deployment replaces it (no per-tag archive).

Repository setup: select **Settings → Pages → Build and deployment → Source →
GitHub Actions**. If the `github-pages` environment restricts deployment refs,
allow `main` and the version tag patterns too. The workflow does not change
repository settings. After enabling Pages, rerun the workflow if an earlier
deployment failed before setup was complete.

`task specimen-build` exports to a fresh `dist/specimen/` directory. Use
`OUT=/path/to/new-directory` for repeat builds without deleting prior output.
Assets and the catalog use relative URLs, so the same build works at a project
Pages path or a custom domain. Font licenses are linked from the viewer.

CJK leaf packages embed one font each. Keep shared notices in
`cjk/internal/notice` free of dependencies on public font packages. Root
`task cjk-ci` runs the nested gate, including dependency and binary link-isolation
tests for both leaves. `task --dir cjk notices-check` also regenerates
`cjk/dist/THIRD_PARTY_NOTICES.md` for review.
