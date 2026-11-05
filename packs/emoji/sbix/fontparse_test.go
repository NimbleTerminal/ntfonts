package sbix_test

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// The module is standard-library only, so these small parsers stand in for a
// font library. They read only the trusted embedded files.

// tables returns each table's bytes by tag.
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

func u16(b []byte, off int) int { return int(binary.BigEndian.Uint16(b[off:])) }
func u32(b []byte, off int) int { return int(binary.BigEndian.Uint32(b[off:])) }
func i16(b []byte, off int) int { return int(int16(binary.BigEndian.Uint16(b[off:]))) }

// numGlyphs reads maxp.numGlyphs.
func numGlyphs(t map[string][]byte) int { return u16(t["maxp"], 4) }

// bitmap is one glyph's PNG and its origin offset in strike pixels: x is the
// left edge and y the bottom edge relative to the glyph's bounding box corner.
type bitmap struct {
	png  []byte
	x, y int
}

// sbixBitmaps reads the single strike of an sbix font.
func sbixBitmaps(t map[string][]byte) (ppem, ppi int, glyphs map[int]bitmap, err error) {
	s := t["sbix"]
	if s == nil {
		return 0, 0, nil, fmt.Errorf("no sbix")
	}
	if n := u32(s, 4); n != 1 {
		return 0, 0, nil, fmt.Errorf("%d strikes", n)
	}
	strike := s[u32(s, 8):]
	ppem, ppi = u16(strike, 0), u16(strike, 2)
	glyphs = map[int]bitmap{}
	for gid := range numGlyphs(t) {
		start, end := u32(strike, 4+4*gid), u32(strike, 8+4*gid)
		if start == end {
			continue
		}
		g := strike[start:end]
		if string(g[4:8]) != "png " {
			return 0, 0, nil, fmt.Errorf("glyph %d graphic type %q", gid, g[4:8])
		}
		glyphs[gid] = bitmap{png: g[8:], x: i16(g, 0), y: i16(g, 2)}
	}
	return ppem, ppi, glyphs, nil
}

// glyfBox returns a glyph's contour count and bounding box; empty is true for
// a zero-length glyf entry.
func glyfBox(t map[string][]byte, gid int) (contours, xMin, yMin, xMax, yMax int, empty bool) {
	loca, glyf := t["loca"], t["glyf"]
	var start, end int
	if i16(t["head"], 50) == 0 {
		start, end = 2*u16(loca, 2*gid), 2*u16(loca, 2*gid+2)
	} else {
		start, end = u32(loca, 4*gid), u32(loca, 4*gid+4)
	}
	if start == end {
		return 0, 0, 0, 0, 0, true
	}
	g := glyf[start:end]
	return i16(g, 0), i16(g, 2), i16(g, 4), i16(g, 6), i16(g, 8), false
}

// contourAreas returns twice the signed area of each contour in a simple glyph.
func contourAreas(t map[string][]byte, gid int) []int {
	loca, glyf := t["loca"], t["glyf"]
	var start, end int
	if i16(t["head"], 50) == 0 {
		start, end = 2*u16(loca, 2*gid), 2*u16(loca, 2*gid+2)
	} else {
		start, end = u32(loca, 4*gid), u32(loca, 4*gid+4)
	}
	if start == end {
		return nil
	}
	g := glyf[start:end]
	nc := i16(g, 0)
	if nc <= 0 {
		return nil
	}
	npts := u16(g, 10+2*(nc-1)) + 1
	p := 12 + 2*nc + u16(g, 10+2*nc)
	flags := make([]byte, 0, npts)
	for len(flags) < npts {
		f := g[p]
		p++
		flags = append(flags, f)
		if f&8 != 0 {
			for range g[p] {
				flags = append(flags, f)
			}
			p++
		}
	}
	coords := func(shortBit, sameBit byte) []int {
		out, v := make([]int, npts), 0
		for i, f := range flags {
			switch {
			case f&shortBit != 0 && f&sameBit != 0:
				v += int(g[p])
				p++
			case f&shortBit != 0:
				v -= int(g[p])
				p++
			case f&sameBit == 0:
				v += i16(g, p)
				p += 2
			}
			out[i] = v
		}
		return out
	}
	xs, ys := coords(2, 16), coords(4, 32)
	areas, first := make([]int, nc), 0
	for c := range nc {
		last := u16(g, 10+2*c)
		for i := first; i <= last; i++ {
			j := i + 1
			if i == last {
				j = first
			}
			areas[c] += xs[i]*ys[j] - xs[j]*ys[i]
		}
		first = last + 1
	}
	return areas
}

// nameString returns the Windows Unicode string for a name ID.
func nameString(t map[string][]byte, id int) string {
	name := t["name"]
	count, strings := u16(name, 2), u16(name, 4)
	for i := range count {
		rec := name[6+12*i:]
		if u16(rec, 0) != 3 || u16(rec, 6) != id {
			continue
		}
		raw := name[strings+u16(rec, 10):][:u16(rec, 8)]
		codes := make([]uint16, len(raw)/2)
		for j := range codes {
			codes[j] = binary.BigEndian.Uint16(raw[2*j:])
		}
		return string(utf16.Decode(codes))
	}
	return ""
}

// cmapHasFormat14 reports whether a Unicode Variation Sequences subtable exists.
func cmapHasFormat14(t map[string][]byte) bool {
	cmap := t["cmap"]
	for i, n := 0, u16(cmap, 2); i < n; i++ {
		if u16(cmap, u32(cmap, 4+8*i+4)) == 14 {
			return true
		}
	}
	return false
}

// featureTags lists the feature tags of a GSUB or GPOS table.
func featureTags(table []byte) []string {
	if table == nil {
		return nil
	}
	list := table[u16(table, 6):]
	var tags []string
	for i, n := 0, u16(list, 0); i < n; i++ {
		tags = append(tags, string(list[2+6*i:6+6*i]))
	}
	return tags
}

// parseCmap returns a lookup over the font's Unicode cmap (format 4 or 12).
func parseCmap(t map[string][]byte) (func(rune) bool, error) {
	cmap := t["cmap"]
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
				for s := range segs {
					end, start := u16(sub, ends+2*s), u16(sub, starts+2*s)
					if int(r) < start || int(r) > end {
						continue
					}
					delta, offset := u16(sub, starts+2*segs+2*s), u16(sub, starts+4*segs+2*s)
					if offset == 0 {
						return (int(r)+delta)&0xFFFF != 0
					}
					return u16(sub, starts+4*segs+2*s+offset+2*(int(r)-start)) != 0
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
