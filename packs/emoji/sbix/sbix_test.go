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

// TestGlyfBoxes checks the embedded sbix strikes and the bounding boxes Chrome
// needs without reading outside the module: 3,985 136x128 PNGs at 109 ppem/72
// ppi with zero origin offsets relative to the box corner, two opposite-winding
// contours over the scaled strike box for bitmap glyphs, nothing for the rest,
// and maxp outline limits large enough for them. Cross-font identity with the
// parent CBDT file is verified by TestSbixDerivedFromCBDT at the repo root.
func TestGlyfBoxes(t *testing.T) {
	got := parsed(t, face(t, "Regular"))
	if maxPoints, maxContours := u16(got["maxp"], 6), u16(got["maxp"], 8); maxPoints < 8 || maxContours < 2 {
		t.Fatalf("maxp maxPoints=%d maxContours=%d, want at least 8 and 2", maxPoints, maxContours)
	}
	ppem, ppi, have, err := sbixBitmaps(got)
	if err != nil {
		t.Fatal(err)
	}
	if ppem != 109 || ppi != 72 || len(have) != bitmaps {
		t.Fatalf("strike %d ppem %d ppi with %d bitmaps, want 109 ppem 72 ppi with %d", ppem, ppi, len(have), bitmaps)
	}
	upem := u16(got["head"], 18)
	scale := func(v int) int { return int(math.Round(float64(v) * float64(upem) / float64(ppem))) }
	for gid := range numGlyphs(got) {
		contours, xMin, yMin, xMax, yMax, empty := glyfBox(got, gid)
		h, ok := have[gid]
		if !ok {
			if !empty {
				t.Fatalf("glyph %d without bitmap has an outline", gid)
			}
			continue
		}
		if !bytes.HasPrefix(h.png, []byte("\x89PNG\r\n\x1a\n")) || len(h.png) < 24 || u32(h.png, 16) != 136 || u32(h.png, 20) != 128 {
			t.Fatalf("glyph %d is not a 136x128 PNG", gid)
		}
		if h.x != 0 || h.y != 0 {
			t.Fatalf("glyph %d offset (%d,%d), want (0,0) relative to box corner", gid, h.x, h.y)
		}
		if empty || contours != 2 {
			t.Fatalf("glyph %d: %d contours, empty %v", gid, contours, empty)
		}
		if xMax-xMin != scale(136) || yMax-yMin != scale(128) {
			t.Fatalf("glyph %d box (%d,%d)-(%d,%d) size (%d,%d), want (%d,%d)", gid, xMin, yMin, xMax, yMax, xMax-xMin, yMax-yMin, scale(136), scale(128))
		}
		if areas := contourAreas(got, gid); len(areas) != 2 || areas[0] == 0 || areas[0]+areas[1] != 0 {
			t.Fatalf("glyph %d contour areas %v do not cancel", gid, areas)
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

func parsed(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	tbl, err := tables(data)
	if err != nil {
		t.Fatal(err)
	}
	return tbl
}
