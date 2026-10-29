// Package scregular embeds only the Simplified Chinese Regular face.
// All supported variants use this upright weight.
package scregular

import (
	"embed"
	"fmt"
	"nimbleterminal.dev/ntfonts/cjk/internal/notice"
)

//go:embed GlossCJKSC-Regular.otf
var face embed.FS

// Source serves the single embedded SC Regular face.
type Source struct{}

// Embedded is the source over this single face.
var Embedded Source

// Fallbacks returns the regional family served by this package.
func Fallbacks() []string { return []string{"Noto Sans CJK SC"} }

// Face returns SC Regular bytes for Regular, Italic, Bold, or BoldItalic.
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
	return face.ReadFile("GlossCJKSC-Regular.otf")
}

// License is the license text shipped for a regional family.
type License = notice.License

// Licenses returns the shared OFL notice for this face.
func Licenses() []License { return notice.Licenses() }
