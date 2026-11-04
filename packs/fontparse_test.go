package packs_test

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"
)

// The module stays standard-library only, so these minimal readers stand in
// for a font parser: table directory, Unicode cmap, feature tags, glyph count.

// tables returns each table's bytes by tag for a TrueType or CFF-flavoured font.
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

// featureTags lists the FeatureList tags of a GSUB or GPOS table.
func featureTags(table []byte) map[string]bool {
	out := map[string]bool{}
	if len(table) < 8 {
		return out
	}
	list := table[binary.BigEndian.Uint16(table[6:]):]
	for i, n := 0, int(binary.BigEndian.Uint16(list)); i < n; i++ {
		out[string(list[2+6*i:6+6*i])] = true
	}
	return out
}

// numGlyphs reads maxp.numGlyphs.
func numGlyphs(data []byte) int {
	t, err := tables(data)
	if err != nil || len(t["maxp"]) < 6 {
		return 0
	}
	return int(binary.BigEndian.Uint16(t["maxp"][4:]))
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

// TestParsersOnCore checks the readers against a core subset in the checkout.
func TestParsersOnCore(t *testing.T) {
	data, err := os.ReadFile("../GlossA-Regular.ttf")
	if err != nil {
		t.Skip("core subset not in checkout:", err)
	}
	tbl, err := tables(data)
	if err != nil {
		t.Fatal(err)
	}
	if tbl["cmap"] == nil || tbl["GPOS"] == nil {
		t.Fatal("missing cmap or GPOS")
	}
	has, err := parseCmap(data)
	if err != nil || !has('A') {
		t.Fatal("cmap lookup failed", err)
	}
	if !featureTags(tbl["GPOS"])["kern"] {
		t.Fatal("GPOS kern not found")
	}
	if numGlyphs(data) == 0 {
		t.Fatal("maxp not read")
	}
}
