// Package ntfonts serves seven OFL-licensed font families, embedded in the
// binary as 28 Latin subset faces. Source implements the FaceSource interface
// of nimbleterminal.dev/nttypeset without importing it, so a program passes
// ntfonts.Embedded to nttypeset.WithFonts.
package ntfonts
