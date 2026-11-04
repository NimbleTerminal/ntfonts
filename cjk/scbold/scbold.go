// Package scbold embeds only the Simplified Chinese Bold face.
// All supported variants use this upright weight.
package scbold

import (
	"embed"
	"fmt"
	"io/fs"
	"nimbleterminal.dev/ntfonts/cjk/internal/notice"
)

//go:embed GlossCJKSC-Bold.otf
var face embed.FS

// Source serves the single embedded SC Bold face.
type Source struct{}

// Embedded is the source over this single face.
var Embedded Source

// Fallbacks returns the regional family served by this package.
func Fallbacks() []string { return []string{"Noto Sans CJK SC"} }

// Face returns SC Bold bytes for Regular, Italic, Bold, or BoldItalic.
// Unknown families and variants are errors. Returned bytes may be modified.
func (Source) Face(family, variant string) ([]byte, error) {
	switch variant {
	case "Regular", "Bold", "Italic", "BoldItalic":
	default:
		return nil, fmt.Errorf("cjk: unknown variant %q", variant)
	}
	if family != "Noto Sans CJK SC" {
		return nil, fmt.Errorf("cjk: unknown family %q", family)
	}
	return face.ReadFile("GlossCJKSC-Bold.otf")
}

// Files returns the embedded font file under its public name.
func Files() fs.FS { return face }

// WebFace describes one embedded file for a browser FontFace.
type WebFace = notice.WebFace

// WebFaces returns a fresh descriptor for the single embedded file.
func WebFaces() []WebFace {
	return []WebFace{{Family: "Noto Sans CJK SC", File: "GlossCJKSC-Bold.otf", Weight: 700, Style: "normal"}}
}

// License is the license text shipped for a regional family.
type License = notice.License

// Licenses returns the shared OFL notice for this face.
func Licenses() []License { return notice.Licenses() }
