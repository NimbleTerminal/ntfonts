# Notices

Source code is MIT licensed; see LICENSE.

The embedded font subsets are licensed under the SIL Open Font License 1.1, not
MIT. They are derived from the upstream fonts below and renamed internally to
Gloss Text A through G and Gloss Text Symbols (the Reserved Font Names are not used). Each family's full
license is in `<family>.OFL.txt` and in THIRD_PARTY_NOTICES.md, generated with
`task notices`. Pinned sources, URLs and hashes are in `sources.json`.

- Literata: https://github.com/google/fonts/tree/bd8f81ddb5c74d5c8897b36ad88b440266245103/ofl/literata
- Newsreader: https://github.com/google/fonts/tree/bd8f81ddb5c74d5c8897b36ad88b440266245103/ofl/newsreader
- IBM Plex Sans: https://github.com/google/fonts/tree/bd8f81ddb5c74d5c8897b36ad88b440266245103/ofl/ibmplexsans
- Source Serif 4: https://github.com/google/fonts/tree/bd8f81ddb5c74d5c8897b36ad88b440266245103/ofl/sourceserif4
- Iosevka: https://github.com/be5invis/Iosevka/releases/tag/v34.9.0
- IBM Plex Mono: https://github.com/google/fonts/tree/bd8f81ddb5c74d5c8897b36ad88b440266245103/ofl/ibmplexmono
- JetBrains Mono: https://github.com/google/fonts/tree/bd8f81ddb5c74d5c8897b36ad88b440266245103/ofl/jetbrainsmono

The symbol fallback combines Noto Sans Symbols 2 with missing arrows/math from
Noto Sans Math, both pinned to the Google Fonts commit above. Their licenses
are `notosanssymbols2.OFL.txt` and `notosansmath.OFL.txt`; `Licenses()` includes
both texts in the fallback entry. The separate `cjk/` module ships its own OFL
and generated notices for Noto Sans CJK SC.

The Go reference fonts (`Go-*.ttf`, `GoMono-*.ttf`) are unused and carry their
own license in `LICENSE.gofont`.
