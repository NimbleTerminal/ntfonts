// Package notice shares the script packs' OFL texts without importing any
// font package, so a host that links one pack embeds only that pack's font.
package notice

import "embed"

//go:embed *.OFL.txt
var texts embed.FS

// License is the license text shipped for a family.
type License struct{ Family, Text string }

// families lists the script families in spec order with their notice files.
var families = []struct{ name, file string }{
	{"Noto Sans Arabic", "notosansarabic.OFL.txt"},
	{"Noto Sans Hebrew", "notosanshebrew.OFL.txt"},
	{"Noto Sans Devanagari", "notosansdevanagari.OFL.txt"},
	{"Noto Sans Bengali", "notosansbengali.OFL.txt"},
	{"Noto Sans Tamil", "notosanstamil.OFL.txt"},
	{"Noto Sans Thai", "notosansthai.OFL.txt"},
	{"Noto Sans Khmer", "notosanskhmer.OFL.txt"},
}

// Families returns the seven script families in the order packs serves them.
func Families() []string {
	out := make([]string, len(families))
	for i, f := range families {
		out[i] = f.name
	}
	return out
}

// Licenses returns a fresh notice per requested family, in the given order.
// Unknown families panic: the set is fixed at build time.
func Licenses(names ...string) []License {
	out := make([]License, 0, len(names))
	for _, name := range names {
		var file string
		for _, f := range families {
			if f.name == name {
				file = f.file
			}
		}
		if file == "" {
			panic("packs: no notice for family " + name)
		}
		text, err := texts.ReadFile(file)
		if err != nil {
			panic(err) // embedded at build time
		}
		out = append(out, License{Family: name, Text: string(text)})
	}
	return out
}
