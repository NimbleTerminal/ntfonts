package sbix_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"math"
	"os"
	"slices"
	"strings"
	"testing"

	"nimbleterminal.dev/ntfonts/packs/emoji/sbix"
)

const (
	family   = "Noto Color Emoji"
	fileName = "NotoColorEmoji-sbix.ttf"
	// The CBDT source this file is derived from, read from the parent module's
	// checkout; the two modules never import each other.
	cbdtPath = "../NotoColorEmoji.ttf"
	bitmaps  = 3985
)

var variants = []string{"Regular", "Bold", "Italic", "BoldItalic"}

func TestFormat(t *testing.T) {
	tbl := parsed(t, face(t, "Regular"))
	for _, want := range []string{"sbix", "glyf", "loca"} {
		if tbl[want] == nil {
			t.Errorf("missing %s table", want)
		}
	}
	for _, bad := range []string{"CBDT", "CBLC", "COLR", "CPAL", "SVG "} {
		if tbl[bad] != nil {
			t.Errorf("unexpected %q table", bad)
		}
	}
}

// TestDerivedFromCBDT proves the artwork and shaping data are the pinned
// CBDT font's: identical layout tables, and the same PNG per glyph placed at
// the same origin.
func TestDerivedFromCBDT(t *testing.T) {
	got, src := parsed(t, face(t, "Regular")), parsed(t, cbdt(t))
	for _, tag := range []string{"cmap", "GSUB", "hmtx", "vhea", "vmtx", "OS/2", "post"} {
		if !bytes.Equal(got[tag], src[tag]) {
			t.Errorf("%s differs from the CBDT font", tag)
		}
	}
	// hhea's outline extents are recomputed for the new glyf boxes; the line
	// metrics and the hmtx count must not change.
	for _, off := range []int{4, 6, 8, 34} {
		if i16(got["hhea"], off) != i16(src["hhea"], off) {
			t.Errorf("hhea field at %d differs from the CBDT font", off)
		}
	}
	if numGlyphs(got) != numGlyphs(src) {
		t.Fatalf("numGlyphs %d, CBDT %d", numGlyphs(got), numGlyphs(src))
	}
	srcPpem, want, err := cbdtBitmaps(src)
	if err != nil {
		t.Fatal(err)
	}
	ppem, ppi, have, err := sbixBitmaps(got)
	if err != nil {
		t.Fatal(err)
	}
	if ppem != srcPpem || ppem != 109 || ppi != 72 {
		t.Fatalf("strike %d ppem %d ppi, CBDT %d ppem", ppem, ppi, srcPpem)
	}
	if len(want) != bitmaps || len(have) != bitmaps {
		t.Fatalf("%d sbix bitmaps, %d CBDT bitmaps, want %d", len(have), len(want), bitmaps)
	}
	for gid, w := range want {
		h, ok := have[gid]
		if !ok {
			t.Fatalf("glyph %d has no sbix bitmap", gid)
		}
		if !bytes.Equal(h.png, w.png) || !bytes.HasPrefix(h.png, []byte("\x89PNG")) {
			t.Fatalf("glyph %d PNG differs", gid)
		}
		if h.x != w.x || h.y != w.y {
			t.Fatalf("glyph %d origin (%d,%d), CBDT (%d,%d)", gid, h.x, h.y, w.x, w.y)
		}
	}
}

// TestGlyfBoxes checks the bounding boxes Chrome needs: two contours over the
// bitmap's design-space box for bitmap glyphs, nothing for the rest.
func TestGlyfBoxes(t *testing.T) {
	got, src := parsed(t, face(t, "Regular")), parsed(t, cbdt(t))
	ppem, want, err := cbdtBitmaps(src)
	if err != nil {
		t.Fatal(err)
	}
	upem := u16(got["head"], 18)
	scale := func(v int) int { return int(math.Round(float64(v) * float64(upem) / float64(ppem))) }
	for gid := range numGlyphs(got) {
		contours, xMin, yMin, xMax, yMax, empty := glyfBox(got, gid)
		w, ok := want[gid]
		if !ok {
			if !empty {
				t.Fatalf("glyph %d without bitmap has an outline", gid)
			}
			continue
		}
		if empty || contours != 2 {
			t.Fatalf("glyph %d: %d contours, empty %v", gid, contours, empty)
		}
		// All Noto strikes are 136x128 at 109 ppem; derive the box from the
		// CBDT metrics rather than hard-coding it.
		if xMin != scale(w.x) || yMin != scale(w.y) || xMax != scale(w.x+136) || yMax != scale(w.y+128) {
			t.Fatalf("glyph %d box (%d,%d)-(%d,%d)", gid, xMin, yMin, xMax, yMax)
		}
	}
}

func TestNames(t *testing.T) {
	tbl := parsed(t, face(t, "Regular"))
	for id, want := range map[int]string{1: "Noto Color Emoji sbix", 4: "Noto Color Emoji sbix", 6: "NotoColorEmoji-sbix"} {
		if got := nameString(tbl, id); got != want {
			t.Errorf("name %d = %q, want %q", id, got, want)
		}
	}
	if got := nameString(tbl, 10); !strings.Contains(got, "converted by ntfonts") {
		t.Errorf("name 10 = %q", got)
	}
	if got := nameString(tbl, 3); !strings.HasPrefix(got, "Noto Color Emoji sbix; ntfonts conversion of Version 2.051") {
		t.Errorf("name 3 = %q", got)
	}
}

// TestShapingTables checks the tables sequences depend on survived: the ccmp
// ligatures for ZWJ sequences, flags, keycaps and skin tones, and the
// format 14 cmap for variation selectors.
func TestShapingTables(t *testing.T) {
	tbl := parsed(t, face(t, "Regular"))
	if !slices.Contains(featureTags(tbl["GSUB"]), "ccmp") {
		t.Errorf("GSUB features %v lack ccmp", featureTags(tbl["GSUB"]))
	}
	if !cmapHasFormat14(tbl) {
		t.Error("no format 14 cmap subtable")
	}
}

func TestFaces(t *testing.T) {
	if !slices.Equal(sbix.Fallbacks(), []string{family}) || !slices.Equal(sbix.Embedded.Fallbacks(), []string{family}) {
		t.Fatal(sbix.Fallbacks())
	}
	got := sbix.Fallbacks()
	got[0] = "changed"
	if sbix.Fallbacks()[0] != family {
		t.Fatal("fallback slice is shared")
	}
	regular := face(t, "Regular")
	for _, v := range variants[1:] {
		if !bytes.Equal(face(t, v), regular) {
			t.Fatalf("%s must alias Regular", v)
		}
	}
	for _, tc := range []struct{ family, variant, msg string }{
		{family, "regular", `sbix: unknown variant "regular"`},
		{family, "", `sbix: unknown variant ""`},
		{"Noto Sans Arabic", "Regular", `sbix: unknown family "Noto Sans Arabic"`},
	} {
		_, err := sbix.Embedded.Face(tc.family, tc.variant)
		if err == nil || err.Error() != tc.msg {
			t.Fatalf("%s/%s: %v", tc.family, tc.variant, err)
		}
	}
	regular[0] ^= 0xff
	if !bytes.Equal(face(t, "Regular"), face(t, "Italic")) {
		t.Fatal("caller modified embedded font")
	}
}

// TestCorpusCoverage checks the code points of the ntwebterm corpus row
// `Emoji: 😀 🚀 🐱 👨‍👩‍👧‍👦 🇺🇳 👍🏽`, as the emoji module does.
func TestCorpusCoverage(t *testing.T) {
	has, err := parseCmap(parsed(t, face(t, "Regular")))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range []rune{0x1F600, 0x1F680, 0x1F431, 0x1F468, 0x1F469, 0x1F467, 0x1F466, 0x1F1FA, 0x1F1F3, 0x1F44D, 0x1F3FD} {
		if !has(r) {
			t.Errorf("missing %U", r)
		}
	}
}

func TestFilesAndWebFaces(t *testing.T) {
	want := []sbix.WebFace{{Family: family, File: fileName, Weight: 400, Style: "normal", Tech: "color-sbix"}}
	got := sbix.WebFaces()
	if !slices.Equal(got, want) {
		t.Fatalf("%+v", got)
	}
	data, err := fs.ReadFile(sbix.Files(), fileName)
	if err != nil || !bytes.Equal(data, face(t, "Regular")) {
		t.Fatal("Files bytes differ from Face", err)
	}
	for _, bad := range []string{"../x", "/abs", "NotoColorEmoji.ttf"} {
		if _, err := sbix.Files().Open(bad); err == nil {
			t.Fatalf("opened %q", bad)
		}
	}
	got[0].Family = "changed"
	if sbix.WebFaces()[0].Family != family {
		t.Fatal("WebFaces slice is shared")
	}
}

func TestLicenses(t *testing.T) {
	got := sbix.Licenses()
	if len(got) != 1 || got[0].Family != family || !strings.Contains(got[0].Text, "SIL OPEN FONT LICENSE") {
		t.Fatalf("%+v", got)
	}
	got[0].Text = ""
	if sbix.Licenses()[0].Text == "" {
		t.Fatal("license slice is shared")
	}
}

func TestHashes(t *testing.T) {
	manifest, err := os.ReadFile("subsets.sha256")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(manifest)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected one file, got %d", len(lines))
	}
	fields := strings.Fields(lines[0])
	if len(fields) != 2 || fields[1] != fileName {
		t.Fatal("invalid manifest", lines[0])
	}
	if fmt.Sprintf("%x", sha256.Sum256(face(t, "Regular"))) != fields[0] {
		t.Fatal("embedded font hash differs from subsets.sha256")
	}
}

func TestBudget(t *testing.T) {
	if n := len(face(t, "Regular")); n > 12_000_000 {
		t.Fatalf("sbix font is %d bytes", n)
	}
}

func face(t *testing.T, variant string) []byte {
	t.Helper()
	data, err := sbix.Embedded.Face(family, variant)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func cbdt(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(cbdtPath)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func parsed(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	tbl, err := tables(data)
	if err != nil {
		t.Fatal(err)
	}
	return tbl
}
