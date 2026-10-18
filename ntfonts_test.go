package ntfonts_test

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"nimbleterminal.dev/ntfonts"
)

var variants = []string{"Regular", "Bold", "Italic", "BoldItalic"}

func TestFontSubsetHashes(t *testing.T) {
	manifest, err := os.ReadFile("subsets.sha256")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(manifest)), "\n")
	if len(lines) != 28 {
		t.Fatal("expected 28 font faces")
	}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatal("invalid hash manifest")
		}
		data, err := os.ReadFile(fields[1])
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != fields[0] {
			t.Fatalf("subset changed: %s", fields[1])
		}
	}
}

func TestSourceFamilies(t *testing.T) {
	families := ntfonts.Families()
	if len(families) != 7 || !slices.IsSorted(families) {
		t.Fatalf("families %v", families)
	}
	for _, family := range families {
		for _, variant := range variants {
			data, err := ntfonts.Embedded.Face(family, variant)
			if err != nil {
				t.Fatal(family, variant, err)
			}
			if _, err := parseCmap(data); err != nil {
				t.Fatal(family, variant, err)
			}
		}
	}
	if _, err := ntfonts.Embedded.Face("Nope", "Regular"); err == nil || !strings.Contains(err.Error(), "Nope") {
		t.Fatal("unknown family accepted", err)
	}
	if _, err := ntfonts.Embedded.Face("Literata", "Heavy"); err == nil || !strings.Contains(err.Error(), "Heavy") {
		t.Fatal("unknown variant accepted", err)
	}
}

func TestSubsetCoverage(t *testing.T) {
	for _, family := range ntfonts.Families() {
		for _, variant := range variants {
			data, err := ntfonts.Embedded.Face(family, variant)
			if err != nil {
				t.Fatal(err)
			}
			cmap, err := parseCmap(data)
			if err != nil {
				t.Fatal(family, variant, err)
			}
			for _, r := range "Café naïve résumé é — “quotes”…" {
				if !cmap(r) {
					t.Fatalf("%s %s lacks %U", family, variant, r)
				}
			}
			for _, r := range "漢🙂" {
				if cmap(r) {
					t.Fatalf("%s %s ships out-of-scope glyph %U", family, variant, r)
				}
			}
		}
	}
}

func TestLicenses(t *testing.T) {
	licenses := ntfonts.Licenses()
	if len(licenses) != 7 {
		t.Fatalf("%d licenses", len(licenses))
	}
	want := []string{"Literata", "Newsreader", "IBM Plex Sans", "Source Serif 4", "Iosevka", "IBM Plex Mono", "JetBrains Mono"}
	for i, l := range licenses {
		if l.Family != want[i] {
			t.Fatalf("license %d is %q, want %q", i, l.Family, want[i])
		}
		if !strings.Contains(strings.ToUpper(l.Text), "SIL OPEN FONT LICENSE") {
			t.Fatalf("%s license lacks the OFL text", l.Family)
		}
	}
}

// parseCmap checks that data is a TrueType font and returns a lookup over its
// Unicode character map (format 4 or 12). The module stays standard-library
// only, so this stands in for a font parser.
func parseCmap(data []byte) (func(rune) bool, error) {
	if len(data) < 12 || binary.BigEndian.Uint32(data) != 0x00010000 {
		return nil, fmt.Errorf("not a TrueType font")
	}
	u16 := func(b []byte, off int) int { return int(binary.BigEndian.Uint16(b[off:])) }
	u32 := func(b []byte, off int) int { return int(binary.BigEndian.Uint32(b[off:])) }
	var cmap []byte
	for i, n := 0, u16(data, 4); i < n; i++ {
		rec := data[12+16*i:]
		if string(rec[:4]) == "cmap" {
			cmap = data[u32(rec, 8):][:u32(rec, 12)]
		}
	}
	if cmap == nil {
		return nil, fmt.Errorf("no cmap table")
	}
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
				for s := 0; s < segs; s++ {
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
				for g := 0; g < groups; g++ {
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
