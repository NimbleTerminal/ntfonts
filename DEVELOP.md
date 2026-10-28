# Development

Use Go 1.27.2 for development and CI (the module minimum is Go 1.26). The root
module and optional `cjk/` nested module have no dependencies.
The root `task ci` also runs the CJK module gate. Installed QA tools live in
`.task/bin/` and generated output in
`bin/` and `dist/`; all are ignored.

```sh
task ci                  # same gate as GitHub Actions; includes network scans
task test                # tests
task test-race           # tests under the race detector
task lint                # golangci-lint v2.14.0, installed into .task/bin
task fonts-check         # shasum -a 256 -c subsets.sha256 (sha256sum as a fallback)
task cross-windows       # check Windows compilation
task cross-wasm          # check js/wasm compilation
task notices             # regenerate THIRD_PARTY_NOTICES.md
task clean               # remove generated artifacts and installed QA tools
```

Install [Go Task](https://taskfile.dev) and a C compiler for the race detector.
`task --list-all` lists maintenance tasks. CI runs on Linux and macOS. Tasks
explicitly use `GOWORK=off`.

`task vulncheck` installs govulncheck v1.8.0 into `.task/bin` and scans the
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
that uses `cjk/sources.json`, writes its two OTFs and `cjk/subsets.sha256`,
and reads the checked-in `cjk/data/gb2312-han.txt`. CJK is pinned to Noto Sans
CJK Sans2.004; symbols and supplemental math glyphs use the same Google Fonts
commit as the text faces. After regenerating, run `task fonts-check` and `task ci`,
update CHANGELOG.md, and review the diff of `subsets.sha256`. The recipe is
described in the Fonts section of README.md.

Run `task notices` and `task --dir cjk notices` after font or license changes,
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

Release the optional module with `cjk/vX.Y.Z` tags and configure its vanity
import path separately. Parent tags do not release the nested module.

## Specimen viewer

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

## GitHub Pages

`.github/workflows/pages.yml` runs the full CI gate, exports the specimen, and
deploys it on pushes to `main`, core version tags (`v*`), and CJK version tags
(`cjk/v*`). It also supports manual runs. All triggers update the same Pages
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
