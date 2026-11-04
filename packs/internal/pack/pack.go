// Package pack implements one script pack over an embedded Regular and Bold
// file, so the seven script packages share one Face and WebFaces definition.
package pack

import (
	"embed"
	"fmt"
	"io/fs"
)

// WebFace describes one embedded file for a browser FontFace, mirroring the
// fields of ntwebterm's fontFaces option. File is a path inside Files().
type WebFace struct {
	Family string // upstream family name, e.g. "Noto Sans Arabic"
	File   string // e.g. "NotoSansArabic-Regular.ttf"
	Weight int    // 400 or 700
	Style  string // "normal"
}

// Pack serves one family's Regular and Bold files under public names
// Base+"-Regular.ttf" and Base+"-Bold.ttf".
type Pack struct {
	Family, Base string
	FS           embed.FS
}

// Face returns the Regular file for Regular and Italic and the Bold file for
// Bold and BoldItalic. Unknown variants and families are errors naming the
// offender. Returned bytes are fresh and may be modified by the caller.
func (p Pack) Face(family, variant string) ([]byte, error) {
	weight, ok := weights[variant]
	if !ok {
		return nil, fmt.Errorf("packs: unknown variant %q", variant)
	}
	if family != p.Family {
		return nil, fmt.Errorf("packs: unknown family %q", family)
	}
	return p.FS.ReadFile(p.Base + "-" + weight + ".ttf")
}

var weights = map[string]string{"Regular": "Regular", "Italic": "Regular", "Bold": "Bold", "BoldItalic": "Bold"}

// Files returns the embedded font files under their public names.
func (p Pack) Files() fs.FS { return p.FS }

// WebFaces returns a fresh descriptor per embedded file, Regular first.
func (p Pack) WebFaces() []WebFace {
	return []WebFace{
		{Family: p.Family, File: p.Base + "-Regular.ttf", Weight: 400, Style: "normal"},
		{Family: p.Family, File: p.Base + "-Bold.ttf", Weight: 700, Style: "normal"},
	}
}
