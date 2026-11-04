// Package packs serves every optional Noto Sans script pack from one source.
// Importing it links all seven script fonts; import a single package such as
// packs/arabic to link only that script. Noto Color Emoji is the separate
// nimbleterminal.dev/ntfonts/packs/emoji module and is not served here.
package packs

import (
	"fmt"
	"io/fs"

	"nimbleterminal.dev/ntfonts/packs/arabic"
	"nimbleterminal.dev/ntfonts/packs/bengali"
	"nimbleterminal.dev/ntfonts/packs/devanagari"
	"nimbleterminal.dev/ntfonts/packs/hebrew"
	"nimbleterminal.dev/ntfonts/packs/internal/notice"
	"nimbleterminal.dev/ntfonts/packs/internal/pack"
	"nimbleterminal.dev/ntfonts/packs/khmer"
	"nimbleterminal.dev/ntfonts/packs/tamil"
	"nimbleterminal.dev/ntfonts/packs/thai"
)

// script is one leaf package's public surface.
type script struct {
	family   string
	face     func(family, variant string) ([]byte, error)
	files    func() fs.FS
	webFaces func() []WebFace
}

// all lists the packs in spec order; Fallbacks, WebFaces and Licenses follow it.
var all = []script{
	{"Noto Sans Arabic", arabic.Embedded.Face, arabic.Files, arabic.WebFaces},
	{"Noto Sans Hebrew", hebrew.Embedded.Face, hebrew.Files, hebrew.WebFaces},
	{"Noto Sans Devanagari", devanagari.Embedded.Face, devanagari.Files, devanagari.WebFaces},
	{"Noto Sans Bengali", bengali.Embedded.Face, bengali.Files, bengali.WebFaces},
	{"Noto Sans Tamil", tamil.Embedded.Face, tamil.Files, tamil.WebFaces},
	{"Noto Sans Thai", thai.Embedded.Face, thai.Files, thai.WebFaces},
	{"Noto Sans Khmer", khmer.Embedded.Face, khmer.Files, khmer.WebFaces},
}

// Source serves all seven script families.
type Source struct{}

// Embedded is the Source over every script pack.
var Embedded Source

// Fallbacks returns the seven script families in renderer preference order.
func Fallbacks() []string {
	out := make([]string, len(all))
	for i, s := range all {
		out[i] = s.family
	}
	return out
}

// Fallbacks returns the same families through the source itself.
func (Source) Fallbacks() []string { return Fallbacks() }

// Face dispatches to the pack serving family. Regular and Italic return the
// Regular file, Bold and BoldItalic the Bold file. Unknown variants and
// families are errors naming the offender.
func (Source) Face(family, variant string) ([]byte, error) {
	switch variant {
	case "Regular", "Bold", "Italic", "BoldItalic":
	default:
		return nil, fmt.Errorf("packs: unknown variant %q", variant)
	}
	for _, s := range all {
		if s.family == family {
			return s.face(family, variant)
		}
	}
	return nil, fmt.Errorf("packs: unknown family %q", family)
}

// Files returns every pack's font files under their public names.
func Files() fs.FS {
	u := make(union, len(all))
	for i, s := range all {
		u[i] = s.files()
	}
	return u
}

// WebFace describes one embedded file for a browser FontFace.
type WebFace = pack.WebFace

// WebFaces returns a fresh descriptor per embedded file: Regular then Bold
// for each family, in Fallbacks order.
func WebFaces() []WebFace {
	out := make([]WebFace, 0, 2*len(all))
	for _, s := range all {
		out = append(out, s.webFaces()...)
	}
	return out
}

// License is the license text shipped for a family.
type License = notice.License

// Licenses returns one OFL notice per family, in Fallbacks order.
func Licenses() []License { return notice.Licenses(Fallbacks()...) }
