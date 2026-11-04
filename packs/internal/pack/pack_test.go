package pack

import (
	"bytes"
	"embed"
	"io/fs"
	"slices"
	"testing"
)

//go:embed testdata/*
var testFS embed.FS

var p = Pack{Family: "Fake", Base: "testdata/Fake", FS: testFS}

func TestFace(t *testing.T) {
	regular, _ := fs.ReadFile(testFS, "testdata/Fake-Regular.ttf")
	bold, _ := fs.ReadFile(testFS, "testdata/Fake-Bold.ttf")
	for variant, want := range map[string][]byte{"Regular": regular, "Italic": regular, "Bold": bold, "BoldItalic": bold} {
		got, err := p.Face("Fake", variant)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("%s: %v %q", variant, err, got)
		}
	}
	for _, tc := range []struct{ family, variant, msg string }{
		{"Fake", "regular", `packs: unknown variant "regular"`},
		{"Fake", "", `packs: unknown variant ""`},
		{"Other", "Regular", `packs: unknown family "Other"`},
	} {
		_, err := p.Face(tc.family, tc.variant)
		if err == nil || err.Error() != tc.msg {
			t.Fatalf("%s/%s: got %v, want %s", tc.family, tc.variant, err, tc.msg)
		}
	}
	data, _ := p.Face("Fake", "Regular")
	data[0] ^= 0xff
	again, _ := p.Face("Fake", "Regular")
	if !bytes.Equal(again, regular) {
		t.Fatal("caller modified embedded font")
	}
}

func TestWebFaces(t *testing.T) {
	want := []WebFace{{"Fake", "testdata/Fake-Regular.ttf", 400, "normal"}, {"Fake", "testdata/Fake-Bold.ttf", 700, "normal"}}
	got := p.WebFaces()
	if !slices.Equal(got, want) {
		t.Fatalf("%+v", got)
	}
	got[0].Family = "changed"
	if p.WebFaces()[0].Family != "Fake" {
		t.Fatal("returned slice is shared")
	}
}

func TestFiles(t *testing.T) {
	bold, _ := fs.ReadFile(testFS, "testdata/Fake-Bold.ttf")
	got, err := fs.ReadFile(p.Files(), "testdata/Fake-Bold.ttf")
	if err != nil || !bytes.Equal(got, bold) {
		t.Fatal(err)
	}
}
