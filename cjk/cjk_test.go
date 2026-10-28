package cjk_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"nimbleterminal.dev/ntfonts/cjk"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestFaces(t *testing.T) {
	for _, family := range cjk.Fallbacks() {
		for _, variant := range []string{"Regular", "Bold", "Italic", "BoldItalic"} {
			data, err := cjk.Embedded.Face(family, variant)
			if err != nil {
				t.Fatal(err)
			}
			cmap, err := parseCmap(data)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range "漢字かなカナ한글。" {
				if !cmap(r) {
					t.Fatalf("%s lacks %U", variant, r)
				}
			}
			if cmap('𠀀') {
				t.Fatal("out-of-scope Han included")
			}
			upright := "Regular"
			if variant == "Bold" || variant == "BoldItalic" {
				upright = "Bold"
			}
			want, _ := cjk.Embedded.Face(family, upright)
			if !bytes.Equal(data, want) {
				t.Fatal("incorrect variant alias")
			}
		}
	}
	for _, args := range [][2]string{{"nope", "Regular"}, {"Noto Sans CJK SC", "Heavy"}} {
		if _, err := cjk.Embedded.Face(args[0], args[1]); err == nil {
			t.Fatal("invalid face accepted")
		}
	}
	licenses := cjk.Licenses()
	if len(licenses) != 1 || licenses[0].Family != "Noto Sans CJK SC" || !strings.Contains(licenses[0].Text, "SIL OPEN FONT LICENSE") {
		t.Fatal("missing license")
	}
}
func TestHashes(t *testing.T) {
	manifest, err := os.ReadFile("subsets.sha256")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(manifest)), "\n")
	if len(lines) != 2 {
		t.Fatal("expected two faces")
	}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatal("invalid manifest")
		}
		data, err := os.ReadFile(fields[1])
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != fields[0] {
			t.Fatal("hash mismatch", fields[1])
		}
	}
}

// parseCmap checks that data is a TrueType font and returns a lookup over its
// Unicode character map (format 4 or 12). The module stays standard-library
// only, so this stands in for a font parser.
func parseCmap(data []byte) (func(rune) bool, error) {
	if len(data) < 12 || (binary.BigEndian.Uint32(data) != 0x00010000 && string(data[:4]) != "OTTO") {
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

func TestCommonUseHan(t *testing.T) {
	raw, err := os.ReadFile("data/gb2312-han.txt")
	if err != nil {
		t.Fatal(err)
	}
	chars := map[rune]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cp, err := strconv.ParseInt(line, 16, 32)
		if err != nil {
			t.Fatal(err)
		}
		if chars[rune(cp)] {
			t.Fatal("duplicate Han codepoint")
		}
		chars[rune(cp)] = true
	}
	if len(chars) != 6763 {
		t.Fatalf("Han list contains %d characters", len(chars))
	}
	for _, variant := range []string{"Regular", "Bold"} {
		data, err := cjk.Embedded.Face("Noto Sans CJK SC", variant)
		if err != nil {
			t.Fatal(err)
		}
		cmap, err := parseCmap(data)
		if err != nil {
			t.Fatal(err)
		}
		for r := range chars {
			if !cmap(r) {
				t.Fatalf("%s lacks common-use %U", variant, r)
			}
		}
	}
}
