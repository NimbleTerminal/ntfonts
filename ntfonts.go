package ntfonts

import (
	"embed"
	"fmt"
	"sort"
)

//go:embed Gloss*.ttf
var faces embed.FS

//go:embed *.OFL.txt
var licenseFiles embed.FS

type family struct {
	name, letter, slug string
}

// table lists the families in license order.
var table = []family{
	{"Literata", "A", "literata"},
	{"Newsreader", "B", "newsreader"},
	{"IBM Plex Sans", "C", "ibmplexsans"},
	{"Source Serif 4", "D", "sourceserif4"},
	{"Iosevka", "E", "iosevka"},
	{"IBM Plex Mono", "F", "ibmplexmono"},
	{"JetBrains Mono", "G", "jetbrainsmono"},
}

var variants = map[string]bool{"Regular": true, "Bold": true, "Italic": true, "BoldItalic": true}

// Source serves the embedded font subsets.
type Source struct{}

// Embedded is the Source over the embedded subsets.
var Embedded Source

// Face returns the font bytes for a family and a variant (Regular, Bold,
// Italic or BoldItalic). It errors, naming the offender, for anything else.
func (Source) Face(family, variant string) ([]byte, error) {
	if !variants[variant] {
		return nil, fmt.Errorf("ntfonts: unknown variant %q", variant)
	}
	if family == "Noto Sans Symbols 2" {
		return faces.ReadFile("GlossSymbols-Regular.ttf")
	}
	for _, f := range table {
		if f.name == family {
			return faces.ReadFile("Gloss" + f.letter + "-" + variant + ".ttf")
		}
	}
	return nil, fmt.Errorf("ntfonts: unknown family %q", family)
}

// Fallbacks returns glyph-fallback families in renderer preference order.
// Embedded.Face serves these families. Noto Sans Symbols 2 is upright Regular
// for all four variants; no synthetic weight or slant is applied. The single
// face includes missing arrows/math from Noto Sans Math.
func Fallbacks() []string { return []string{"Noto Sans Symbols 2"} }

// Fallbacks returns the same ordered glyph-fallback families as the package
// function, allowing renderers to discover them through the source itself.
func (Source) Fallbacks() []string { return Fallbacks() }

// Families returns the available family names, sorted.
func Families() []string {
	names := make([]string, len(table))
	for i, f := range table {
		names[i] = f.name
	}
	sort.Strings(names)
	return names
}

// License is the license text shipped for a family.
type License struct{ Family, Text string }

// Licenses returns the OFL text of every text family followed by the fallback families.
func Licenses() []License {
	all := append(append([]family(nil), table...), family{"Noto Sans Symbols 2", "Symbols", "notosanssymbols2"})
	out := make([]License, len(all))
	for i, f := range all {
		text, err := licenseFiles.ReadFile(f.slug + ".OFL.txt")
		if err != nil {
			panic(err) // embedded at build time
		}
		if f.name == "Noto Sans Symbols 2" {
			math, err := licenseFiles.ReadFile("notosansmath.OFL.txt")
			if err != nil {
				panic(err)
			}
			text = append(append(text, '\n'), math...)
		}
		out[i] = License{Family: f.name, Text: string(text)}
	}
	return out
}
