// Package sbix embeds a Safari-compatible build of Noto Color Emoji: the same
// PNG bitmaps, cmap and GSUB as the CBDT font in nimbleterminal.dev/ntfonts/packs/emoji,
// repackaged as an sbix table by scripts/subset-fonts.py. WebKit rejects
// CBDT but renders sbix; Chrome renders both. Browsers choose a format with
// CSS.supports("font-tech(color-sbix)") and the Tech field of WebFace.
//
// The file on disk is NotoColorEmoji-sbix.ttf and its name table says
// "Noto Color Emoji sbix", marking it as a modified version; the API family
// stays "Noto Color Emoji" so consumers select the same family in either
// format. Native renderers should keep using the unmodified CBDT module.
package sbix

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed NotoColorEmoji-sbix.ttf
var files embed.FS

//go:embed notocoloremoji.OFL.txt
var license string

const (
	family   = "Noto Color Emoji"
	fileName = "NotoColorEmoji-sbix.ttf"
)

// Source serves the single embedded emoji face.
type Source struct{}

// Embedded is the Source over the embedded file.
var Embedded Source

// Fallbacks returns the family served by this package.
func Fallbacks() []string { return []string{family} }

// Fallbacks returns the same family through the source itself.
func (Source) Fallbacks() []string { return Fallbacks() }

// Face returns the sbix font for Regular, Bold, Italic or BoldItalic.
// Unknown families and variants are errors. Returned bytes may be modified.
func (Source) Face(name, variant string) ([]byte, error) {
	switch variant {
	case "Regular", "Bold", "Italic", "BoldItalic":
	default:
		return nil, fmt.Errorf("sbix: unknown variant %q", variant)
	}
	if name != family {
		return nil, fmt.Errorf("sbix: unknown family %q", name)
	}
	return files.ReadFile(fileName)
}

// Files returns the embedded font file under its public name.
func Files() fs.FS { return files }

// WebFace describes the embedded file for a browser FontFace, mirroring the
// fields of ntwebterm's fontFaces option. File is a path inside Files().
// Tech is the CSS font-tech() keyword the browser must support to render the
// file; test it with CSS.supports("font-tech(" + Tech + ")") or declare it
// with tech() in a @font-face src list.
type WebFace struct {
	Family string
	File   string
	Weight int
	Style  string
	Tech   string
}

// WebFaces returns a fresh descriptor for the single embedded file.
func WebFaces() []WebFace {
	return []WebFace{{Family: family, File: fileName, Weight: 400, Style: "normal", Tech: "color-sbix"}}
}

// License is the license text shipped for a family.
type License struct{ Family, Text string }

// Licenses returns a fresh slice holding the OFL notice for Noto Color Emoji.
func Licenses() []License { return []License{{Family: family, Text: license}} }
