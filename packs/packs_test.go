package packs_test

import (
	"bytes"
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"nimbleterminal.dev/ntfonts/packs"
)

var variants = []string{"Regular", "Bold", "Italic", "BoldItalic"}

func TestRootDispatch(t *testing.T) {
	want := []string{"Noto Sans Arabic", "Noto Sans Hebrew", "Noto Sans Devanagari", "Noto Sans Bengali", "Noto Sans Tamil", "Noto Sans Thai", "Noto Sans Khmer"}
	if !slices.Equal(packs.Fallbacks(), want) || !slices.Equal(packs.Embedded.Fallbacks(), want) {
		t.Fatal(packs.Fallbacks())
	}
	got := packs.Fallbacks()
	got[0] = "changed"
	if packs.Fallbacks()[0] != want[0] {
		t.Fatal("fallback slice is shared")
	}
	for _, s := range scripts {
		for _, v := range variants {
			root, err := packs.Embedded.Face(s.family, v)
			if err != nil {
				t.Fatal(s.family, v, err)
			}
			leaf, _ := s.face(s.family, v)
			if !bytes.Equal(root, leaf) {
				t.Fatalf("%s/%s: root bytes differ from leaf", s.family, v)
			}
		}
	}
	for _, tc := range []struct{ family, variant, msg string }{
		{"Noto Color Emoji", "Regular", `packs: unknown family "Noto Color Emoji"`},
		{"Noto Sans Arabic", "Heavy", `packs: unknown variant "Heavy"`},
	} {
		_, err := packs.Embedded.Face(tc.family, tc.variant)
		if err == nil || err.Error() != tc.msg {
			t.Fatalf("%s/%s: %v", tc.family, tc.variant, err)
		}
	}
}

func TestRootWebFaces(t *testing.T) {
	got := packs.WebFaces()
	if len(got) != 14 {
		t.Fatalf("%d web faces", len(got))
	}
	for i, s := range scripts {
		for j, w := range weights {
			wf := got[2*i+j]
			want := packs.WebFace{Family: s.family, File: s.base + "-" + w + ".ttf", Weight: 400 + 300*j, Style: "normal"}
			if wf != want {
				t.Fatalf("entry %d: %+v, want %+v", 2*i+j, wf, want)
			}
			data, err := fs.ReadFile(packs.Files(), wf.File)
			if err != nil {
				t.Fatal(err)
			}
			face, _ := packs.Embedded.Face(s.family, w)
			if !bytes.Equal(data, face) {
				t.Fatalf("%s: Files bytes differ from Face", wf.File)
			}
		}
	}
	got[0].Family = "changed"
	if packs.WebFaces()[0].Family != scripts[0].family {
		t.Fatal("WebFaces slice is shared")
	}
}

func TestRootFiles(t *testing.T) {
	var names []string
	for _, wf := range packs.WebFaces() {
		names = append(names, wf.File)
	}
	if err := fstest.TestFS(packs.Files(), names...); err != nil {
		t.Fatal(err)
	}
	_, err := packs.Files().Open("missing.ttf")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing file: %v", err)
	}
	for _, bad := range []string{"../x", "/abs", "arabic/.."} {
		if _, err := packs.Files().Open(bad); err == nil {
			t.Fatalf("opened %q", bad)
		}
	}
}

func TestRootLicenses(t *testing.T) {
	got := packs.Licenses()
	if len(got) != 7 {
		t.Fatalf("%d licenses", len(got))
	}
	for i, family := range packs.Fallbacks() {
		if got[i].Family != family || !strings.Contains(got[i].Text, "SIL OPEN FONT LICENSE") {
			t.Fatalf("entry %d: %s", i, got[i].Family)
		}
	}
}
