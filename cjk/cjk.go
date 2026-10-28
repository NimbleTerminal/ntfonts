// Package cjk serves optional regional CJK subsets without enlarging ntfonts.
package cjk

import (
	"embed"
	"fmt"
)

//go:embed GlossCJKSC-*.otf
var faces embed.FS

//go:embed notosanscjksc.OFL.txt
var licenseText string

// Source serves regional CJK faces. Italic variants use upright faces.
type Source struct{}

// Embedded is the source over the optional CJK subsets.
var Embedded Source

// Fallbacks returns available regional families. Hosts choose the region that
// matches their content; SC shapes are not a substitute for JP, TC, or KR shapes.
func Fallbacks() []string { return []string{"Noto Sans CJK SC"} }

// Face returns font bytes for the named regional family and variant.
func (Source) Face(family, variant string) ([]byte, error) {
	switch variant {
	case "Regular", "Italic":
		variant = "Regular"
	case "Bold", "BoldItalic":
		variant = "Bold"
	default:
		return nil, fmt.Errorf("cjk: unknown variant %q", variant)
	}
	switch family {
	case "Noto Sans CJK SC":
		return faces.ReadFile("GlossCJKSC-" + variant + ".otf")
	default:
		return nil, fmt.Errorf("cjk: unknown family %q", family)
	}
}

// License is the license text shipped for a regional family.
type License struct{ Family, Text string }

// Licenses returns the OFL notices for the embedded regional families.
func Licenses() []License { return []License{{"Noto Sans CJK SC", licenseText}} }
