// Package cjk serves both optional SC weights. Import cjk/scregular or
// cjk/scbold instead to link only one weight. No CJK data is linked by ntfonts.
package cjk

import (
	"fmt"
	"nimbleterminal.dev/ntfonts/cjk/internal/notice"
	"nimbleterminal.dev/ntfonts/cjk/scbold"
	"nimbleterminal.dev/ntfonts/cjk/scregular"
)

// Source serves regional CJK faces. Italic variants use upright faces.
type Source struct{}

// Embedded serves both SC weights; importing this package links both fonts.
var Embedded Source

// Fallbacks returns available regional families. Hosts choose the region that
// matches their content; SC shapes are not a substitute for JP, TC, or KR shapes.
func Fallbacks() []string { return scregular.Fallbacks() }

// Face returns Regular for Regular/Italic and Bold for Bold/BoldItalic.
func (Source) Face(family, variant string) ([]byte, error) {
	switch variant {
	case "Regular", "Italic":
		return scregular.Embedded.Face(family, variant)
	case "Bold", "BoldItalic":
		return scbold.Embedded.Face(family, variant)
	default:
		return nil, fmt.Errorf("cjk: unknown variant %q", variant)
	}
}

// License is the license text shipped for a regional family.
type License = notice.License

// Licenses returns one shared OFL notice for the SC family.
func Licenses() []License { return notice.Licenses() }
