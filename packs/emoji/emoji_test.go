package emoji_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"

	"nimbleterminal.dev/ntfonts/packs/emoji"
)

const family = "Noto Color Emoji"

var variants = []string{"Regular", "Bold", "Italic", "BoldItalic"}

func TestFormat(t *testing.T) {
	data := face(t, "Regular")
	tbl, err := tables(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"CBDT", "CBLC"} {
		if tbl[want] == nil {
			t.Errorf("missing %s table: not the bitmap build", want)
		}
	}
	for _, bad := range []string{"COLR", "CPAL", "SVG "} {
		if tbl[bad] != nil {
			t.Errorf("unexpected %q table: nttypeset cannot draw it", bad)
		}
	}
}

func TestFaces(t *testing.T) {
	if !slices.Equal(emoji.Fallbacks(), []string{family}) || !slices.Equal(emoji.Embedded.Fallbacks(), []string{family}) {
		t.Fatal(emoji.Fallbacks())
	}
	got := emoji.Fallbacks()
	got[0] = "changed"
	if emoji.Fallbacks()[0] != family {
		t.Fatal("fallback slice is shared")
	}
	regular := face(t, "Regular")
	for _, v := range variants[1:] {
		if !bytes.Equal(face(t, v), regular) {
			t.Fatalf("%s must alias Regular", v)
		}
	}
	for _, tc := range []struct{ family, variant, msg string }{
		{family, "regular", `emoji: unknown variant "regular"`},
		{family, "", `emoji: unknown variant ""`},
		{"Noto Sans Arabic", "Regular", `emoji: unknown family "Noto Sans Arabic"`},
	} {
		_, err := emoji.Embedded.Face(tc.family, tc.variant)
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
// `Emoji: 😀 🚀 🐱 👨‍👩‍👧‍👦 🇺🇳 👍🏽` (demo/multilingual-corpus.js at 0.4.2).
func TestCorpusCoverage(t *testing.T) {
	has, err := parseCmap(face(t, "Regular"))
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
	want := []emoji.WebFace{{Family: family, File: "NotoColorEmoji.ttf", Weight: 400, Style: "normal"}}
	got := emoji.WebFaces()
	if !slices.Equal(got, want) {
		t.Fatalf("%+v", got)
	}
	data, err := fs.ReadFile(emoji.Files(), "NotoColorEmoji.ttf")
	if err != nil || !bytes.Equal(data, face(t, "Regular")) {
		t.Fatal("Files bytes differ from Face", err)
	}
	for _, bad := range []string{"../x", "/abs"} {
		if _, err := emoji.Files().Open(bad); err == nil {
			t.Fatalf("opened %q", bad)
		}
	}
	got[0].Family = "changed"
	if emoji.WebFaces()[0].Family != family {
		t.Fatal("WebFaces slice is shared")
	}
}

func TestLicenses(t *testing.T) {
	got := emoji.Licenses()
	if len(got) != 1 || got[0].Family != family || !strings.Contains(got[0].Text, "SIL OPEN FONT LICENSE") {
		t.Fatalf("%+v", got)
	}
	got[0].Text = ""
	if emoji.Licenses()[0].Text == "" {
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
	if len(fields) != 2 || fields[1] != "NotoColorEmoji.ttf" {
		t.Fatal("invalid manifest", lines[0])
	}
	if fmt.Sprintf("%x", sha256.Sum256(face(t, "Regular"))) != fields[0] {
		t.Fatal("embedded font hash differs from subsets.sha256")
	}
}

func TestBudget(t *testing.T) {
	if n := len(face(t, "Regular")); n > 12_000_000 {
		t.Fatalf("emoji font is %d bytes", n)
	}
}

func face(t *testing.T, variant string) []byte {
	t.Helper()
	data, err := emoji.Embedded.Face(family, variant)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// tables returns each table's bytes by tag. The module is standard-library
// only, so this and parseCmap stand in for a font parser.
func tables(data []byte) (map[string][]byte, error) {
	if len(data) < 12 || (binary.BigEndian.Uint32(data) != 0x00010000 && string(data[:4]) != "OTTO") {
		return nil, fmt.Errorf("not an sfnt font")
	}
	out := map[string][]byte{}
	for i, n := 0, int(binary.BigEndian.Uint16(data[4:])); i < n; i++ {
		rec := data[12+16*i:]
		off, length := int(binary.BigEndian.Uint32(rec[8:])), int(binary.BigEndian.Uint32(rec[12:]))
		if off+length > len(data) {
			return nil, fmt.Errorf("table %q exceeds font", rec[:4])
		}
		out[string(rec[:4])] = data[off : off+length]
	}
	return out, nil
}

// parseCmap returns a lookup over the font's Unicode cmap (format 4 or 12).
func parseCmap(data []byte) (func(rune) bool, error) {
	t, err := tables(data)
	if err != nil {
		return nil, err
	}
	cmap := t["cmap"]
	if cmap == nil {
		return nil, fmt.Errorf("no cmap table")
	}
	u16 := func(b []byte, off int) int { return int(binary.BigEndian.Uint16(b[off:])) }
	u32 := func(b []byte, off int) int { return int(binary.BigEndian.Uint32(b[off:])) }
	var ranges []func(rune) bool
	for i, n := 0, u16(cmap, 2); i < n; i++ {
		sub := cmap[u32(cmap, 4+8*i+4):]
		switch u16(sub, 0) {
		case 4:
			segs := u16(sub, 6) / 2
			ends, starts := 14, 16+2*segs
			ranges = append(ranges, func(r rune) bool {
				if r > 0xFFFF {
					return false
				}
				for s := range segs {
					end, start := u16(sub, ends+2*s), u16(sub, starts+2*s)
					if int(r) < start || int(r) > end {
						continue
					}
					delta, offset := u16(sub, starts+2*segs+2*s), u16(sub, starts+4*segs+2*s)
					if offset == 0 {
						return (int(r)+delta)&0xFFFF != 0
					}
					glyph := u16(sub, starts+4*segs+2*s+offset+2*(int(r)-start))
					return glyph != 0
				}
				return false
			})
		case 12:
			groups := u32(sub, 12)
			ranges = append(ranges, func(r rune) bool {
				for g := range groups {
					start, end := u32(sub, 16+12*g), u32(sub, 20+12*g)
					if int(r) >= start && int(r) <= end {
						return u32(sub, 24+12*g)+int(r)-start != 0
					}
				}
				return false
			})
		}
	}
	if len(ranges) == 0 {
		return nil, fmt.Errorf("no Unicode cmap subtable")
	}
	return func(r rune) bool {
		for _, in := range ranges {
			if in(r) {
				return true
			}
		}
		return false
	}, nil
}
