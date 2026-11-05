package ntfonts_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"testing"
)

// TestSbixDerivedFromCBDT is a repository integration check proving that the
// sbix emoji module's font is derived from the pinned CBDT font in packs/emoji:
// identical layout and metric tables, identical PNG bytes for all 3,985 bitmap
// glyphs, glyf boxes matching each glyph's scaled CBDT metrics, and sbix
// origin offsets measured relative to the lower-left corner of the glyf box.
// Kept at the root so packs/emoji/sbix's own tests remain self-contained when
// tested outside the repository checkout.
func TestSbixDerivedFromCBDT(t *testing.T) {
	cbdtBytes, err := os.ReadFile("packs/emoji/NotoColorEmoji.ttf")
	if os.IsNotExist(err) {
		t.Skip("optional emoji modules not in checkout")
	}
	if err != nil {
		t.Fatal(err)
	}
	sbixBytes, err := os.ReadFile("packs/emoji/sbix/NotoColorEmoji-sbix.ttf")
	if err != nil {
		t.Fatal(err)
	}
	src, err := sfntTables(cbdtBytes)
	if err != nil {
		t.Fatal(err)
	}
	got, err := sfntTables(sbixBytes)
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range []string{"cmap", "GSUB", "hmtx", "vhea", "vmtx", "OS/2", "post"} {
		if !bytes.Equal(got[tag], src[tag]) {
			t.Errorf("%s differs from the CBDT font", tag)
		}
	}
	// hhea's outline extents are recomputed for the new glyf boxes; the line
	// metrics and the hmtx count must not change.
	for _, off := range []int{4, 6, 8, 34} {
		if beI16(got["hhea"], off) != beI16(src["hhea"], off) {
			t.Errorf("hhea field at %d differs from the CBDT font", off)
		}
	}
	nGlyphs := beU16(got["maxp"], 4)
	if nGlyphs != beU16(src["maxp"], 4) {
		t.Fatalf("numGlyphs %d, CBDT %d", nGlyphs, beU16(src["maxp"], 4))
	}
	srcPpem, want, err := parseCBDTBitmaps(src)
	if err != nil {
		t.Fatal(err)
	}
	ppem, ppi, have, err := parseSbixBitmaps(got, nGlyphs)
	if err != nil {
		t.Fatal(err)
	}
	if ppem != srcPpem || ppem != 109 || ppi != 72 {
		t.Fatalf("strike %d ppem %d ppi, CBDT %d ppem", ppem, ppi, srcPpem)
	}
	const bitmaps = 3985
	if len(want) != bitmaps || len(have) != bitmaps {
		t.Fatalf("%d sbix bitmaps, %d CBDT bitmaps, want %d", len(have), len(want), bitmaps)
	}
	upem := beU16(got["head"], 18)
	px := func(units int) int { return int(math.Round(float64(units) * float64(ppem) / float64(upem))) }
	scale := func(v int) int { return int(math.Round(float64(v) * float64(upem) / float64(ppem))) }
	for gid := range nGlyphs {
		contours, xMin, yMin, xMax, yMax, empty := parseGlyfBox(got, gid)
		w, ok := want[gid]
		if !ok {
			if !empty {
				t.Fatalf("glyph %d without CBDT bitmap has a glyf outline", gid)
			}
			continue
		}
		h, ok := have[gid]
		if !ok {
			t.Fatalf("glyph %d has no sbix bitmap", gid)
		}
		if !bytes.Equal(h.png, w.png) || !bytes.HasPrefix(h.png, []byte("\x89PNG")) {
			t.Fatalf("glyph %d PNG differs", gid)
		}
		if empty || contours != 2 {
			t.Fatalf("glyph %d: %d contours, empty %v", gid, contours, empty)
		}
		if xMin != scale(w.x) || yMin != scale(w.y) || xMax != scale(w.x+136) || yMax != scale(w.y+128) {
			t.Fatalf("glyph %d box (%d,%d)-(%d,%d) differs from scaled CBDT metrics", gid, xMin, yMin, xMax, yMax)
		}
		if h.x != w.x-px(xMin) || h.y != w.y-px(yMin) {
			t.Fatalf("glyph %d offset (%d,%d), want CBDT origin (%d,%d) relative to box corner (%d,%d)", gid, h.x, h.y, w.x, w.y, px(xMin), px(yMin))
		}
	}
}

type emojiBitmap struct {
	png  []byte
	x, y int
}

func beU16(b []byte, off int) int { return int(binary.BigEndian.Uint16(b[off:])) }
func beU32(b []byte, off int) int { return int(binary.BigEndian.Uint32(b[off:])) }
func beI16(b []byte, off int) int { return int(int16(binary.BigEndian.Uint16(b[off:]))) }

func sfntTables(data []byte) (map[string][]byte, error) {
	if len(data) < 12 || (binary.BigEndian.Uint32(data) != 0x00010000 && string(data[:4]) != "OTTO") {
		return nil, fmt.Errorf("not an sfnt font")
	}
	out := map[string][]byte{}
	for i, n := 0, beU16(data, 4); i < n; i++ {
		rec := data[12+16*i:]
		off, length := beU32(rec, 8), beU32(rec, 12)
		if off+length > len(data) {
			return nil, fmt.Errorf("table %q exceeds font", rec[:4])
		}
		out[string(rec[:4])] = data[off : off+length]
	}
	return out, nil
}

func parseCBDTBitmaps(t map[string][]byte) (ppem int, glyphs map[int]emojiBitmap, err error) {
	cblc, cbdt := t["CBLC"], t["CBDT"]
	if cblc == nil || cbdt == nil {
		return 0, nil, fmt.Errorf("no CBLC/CBDT")
	}
	size := cblc[8:]
	arrayOff, count, ppem := beU32(size, 0), beU32(size, 8), int(size[44])
	glyphs = map[int]emojiBitmap{}
	for i := range count {
		entry := cblc[arrayOff+8*i:]
		first, last, sub := beU16(entry, 0), beU16(entry, 2), cblc[arrayOff+beU32(entry, 4):]
		indexFormat, imageFormat, dataOff := beU16(sub, 0), beU16(sub, 2), beU32(sub, 4)
		if imageFormat != 17 {
			return 0, nil, fmt.Errorf("image format %d", imageFormat)
		}
		offset := func(k int) int {
			switch indexFormat {
			case 1:
				return beU32(sub, 8+4*k)
			case 3:
				return beU16(sub, 8+2*k)
			}
			return -1
		}
		for gid := first; gid <= last; gid++ {
			start, end := offset(gid-first), offset(gid-first+1)
			if start < 0 {
				return 0, nil, fmt.Errorf("index format %d", indexFormat)
			}
			if start == end {
				continue
			}
			g := cbdt[dataOff+start : dataOff+end]
			height, bearingX, bearingY := int(g[0]), int(int8(g[2])), int(int8(g[3]))
			n := beU32(g, 5)
			glyphs[gid] = emojiBitmap{png: g[9 : 9+n], x: bearingX, y: bearingY - height}
		}
	}
	return ppem, glyphs, nil
}

func parseSbixBitmaps(t map[string][]byte, numGlyphs int) (ppem, ppi int, glyphs map[int]emojiBitmap, err error) {
	s := t["sbix"]
	if s == nil {
		return 0, 0, nil, fmt.Errorf("no sbix")
	}
	if n := beU32(s, 4); n != 1 {
		return 0, 0, nil, fmt.Errorf("%d strikes", n)
	}
	strike := s[beU32(s, 8):]
	ppem, ppi = beU16(strike, 0), beU16(strike, 2)
	glyphs = map[int]emojiBitmap{}
	for gid := range numGlyphs {
		start, end := beU32(strike, 4+4*gid), beU32(strike, 8+4*gid)
		if start == end {
			continue
		}
		g := strike[start:end]
		if string(g[4:8]) != "png " {
			return 0, 0, nil, fmt.Errorf("glyph %d graphic type %q", gid, g[4:8])
		}
		glyphs[gid] = emojiBitmap{png: g[8:], x: beI16(g, 0), y: beI16(g, 2)}
	}
	return ppem, ppi, glyphs, nil
}

func parseGlyfBox(t map[string][]byte, gid int) (contours, xMin, yMin, xMax, yMax int, empty bool) {
	loca, glyf := t["loca"], t["glyf"]
	var start, end int
	if beI16(t["head"], 50) == 0 {
		start, end = 2*beU16(loca, 2*gid), 2*beU16(loca, 2*gid+2)
	} else {
		start, end = beU32(loca, 4*gid), beU32(loca, 4*gid+4)
	}
	if start == end {
		return 0, 0, 0, 0, 0, true
	}
	g := glyf[start:end]
	return beI16(g, 0), beI16(g, 2), beI16(g, 4), beI16(g, 6), beI16(g, 8), false
}
