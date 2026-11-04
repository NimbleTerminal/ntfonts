// Package notice shares CJK license text without importing any font package.
package notice

import _ "embed"

//go:embed notosanscjksc.OFL.txt
var text string

// License is the license text shipped for a regional family.
type License struct{ Family, Text string }

// WebFace describes one embedded file for a browser FontFace, mirroring the
// fields of ntwebterm's fontFaces option. File is a path inside Files().
type WebFace struct {
	Family string
	File   string
	Weight int
	Style  string
}

// Licenses returns a fresh slice of notices for the SC faces.
func Licenses() []License { return []License{{"Noto Sans CJK SC", text}} }
