// Package arabic embeds Noto Sans Arabic Regular and Bold subsets with every
// OpenType layout feature retained. Italic variants use the upright weights.
package arabic

import (
	"embed"
	"io/fs"

	"nimbleterminal.dev/ntfonts/packs/internal/notice"
	"nimbleterminal.dev/ntfonts/packs/internal/pack"
)

//go:embed NotoSansArabic-Regular.ttf NotoSansArabic-Bold.ttf
var files embed.FS

var p = pack.Pack{Family: "Noto Sans Arabic", Base: "NotoSansArabic", FS: files}

// Source serves the embedded Noto Sans Arabic faces.
type Source struct{}

// Embedded is the Source over this package's two files.
var Embedded Source

// Fallbacks returns the family served by this package.
func Fallbacks() []string { return []string{p.Family} }

// Fallbacks returns the same family through the source itself.
func (Source) Fallbacks() []string { return Fallbacks() }

// Face returns Regular for Regular/Italic and Bold for Bold/BoldItalic.
// Unknown families and variants are errors. Returned bytes may be modified.
func (Source) Face(family, variant string) ([]byte, error) { return p.Face(family, variant) }

// Files returns the embedded font files under their public names.
func Files() fs.FS { return p.Files() }

// WebFace describes one embedded file for a browser FontFace.
type WebFace = pack.WebFace

// WebFaces returns one descriptor per embedded file, Regular first.
func WebFaces() []WebFace { return p.WebFaces() }

// License is the license text shipped for a family.
type License = notice.License

// Licenses returns this family's OFL notice.
func Licenses() []License { return notice.Licenses(p.Family) }
