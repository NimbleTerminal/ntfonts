package notice

import (
	"slices"
	"strings"
	"testing"
)

func TestLicenses(t *testing.T) {
	want := []string{"Noto Sans Arabic", "Noto Sans Hebrew", "Noto Sans Devanagari", "Noto Sans Bengali", "Noto Sans Tamil", "Noto Sans Thai", "Noto Sans Khmer"}
	if !slices.Equal(Families(), want) {
		t.Fatalf("families %v", Families())
	}
	got := Licenses("Noto Sans Arabic", "Noto Sans Khmer")
	if len(got) != 2 || got[0].Family != "Noto Sans Arabic" || got[1].Family != "Noto Sans Khmer" {
		t.Fatalf("order %+v", got)
	}
	for _, l := range got {
		if !strings.Contains(l.Text, "SIL OPEN FONT LICENSE") {
			t.Fatalf("%s: not an OFL text", l.Family)
		}
	}
	if len(Licenses(Families()...)) != 7 {
		t.Fatal("expected seven notices")
	}
	got[0].Text = ""
	if Licenses("Noto Sans Arabic")[0].Text == "" {
		t.Fatal("returned slice is shared")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("unknown family must panic")
		}
	}()
	Licenses("Nope")
}
