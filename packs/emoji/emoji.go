// Package emoji embeds Noto Color Emoji unmodified: the CBDT/CBLC bitmap
// color build from googlefonts/noto-emoji v2.051, which nttypeset can
// rasterize (it draws bitmap glyphs, not COLR or SVG). Chrome and Firefox
// render CBDT in the browser; Safari substitutes the platform emoji font.
// All four variants serve the same file.
package emoji

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed NotoColorEmoji.ttf
var files embed.FS

//go:embed notocoloremoji.OFL.txt
var license string

const family = "Noto Color Emoji"

// Source serves the single embedded emoji face.
type Source struct{}

// Embedded is the Source over the embedded file.
var Embedded Source

// Fallbacks returns the family served by this package.
func Fallbacks() []string { return []string{family} }

// Fallbacks returns the same family through the source itself.
func (Source) Fallbacks() []string { return Fallbacks() }

// Face returns the emoji font for Regular, Bold, Italic or BoldItalic.
// Unknown families and variants are errors. Returned bytes may be modified.
func (Source) Face(name, variant string) ([]byte, error) {
	switch variant {
	case "Regular", "Bold", "Italic", "BoldItalic":
	default:
		return nil, fmt.Errorf("emoji: unknown variant %q", variant)
	}
	if name != family {
		return nil, fmt.Errorf("emoji: unknown family %q", name)
	}
	return files.ReadFile("NotoColorEmoji.ttf")
}

// Files returns the embedded font file under its public name.
func Files() fs.FS { return files }

// WebFace describes the embedded file for a browser FontFace, mirroring the
// fields of ntwebterm's fontFaces option. File is a path inside Files().
type WebFace struct {
	Family string
	File   string
	Weight int
	Style  string
}

// WebFaces returns a fresh descriptor for the single embedded file.
func WebFaces() []WebFace {
	return []WebFace{{Family: family, File: "NotoColorEmoji.ttf", Weight: 400, Style: "normal"}}
}

// License is the license text shipped for a family.
type License struct{ Family, Text string }

// Licenses returns a fresh slice holding the OFL notice for Noto Color Emoji.
func Licenses() []License { return []License{{Family: family, Text: license}} }
