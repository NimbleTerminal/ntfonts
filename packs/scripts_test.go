package packs_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"

	"nimbleterminal.dev/ntfonts/packs/arabic"
	"nimbleterminal.dev/ntfonts/packs/bengali"
	"nimbleterminal.dev/ntfonts/packs/devanagari"
	"nimbleterminal.dev/ntfonts/packs/hebrew"
	"nimbleterminal.dev/ntfonts/packs/khmer"
	"nimbleterminal.dev/ntfonts/packs/tamil"
	"nimbleterminal.dev/ntfonts/packs/thai"
)

// budget bounds the fourteen embedded files; measured 1,068,724 bytes at the
// pinned sources, rounded up from 1.25x.
const budget = 1_400_000

type span struct{ first, last rune }

// scripts drives every per-package test. Every package's WebFace and License
// alias the same underlying types, so arabic's names stand for all.
// row is the ntwebterm multilingual corpus text, verbatim from
// ntwebterm/demo/multilingual-corpus.js at 0.4.2: the ecosystem's agreed sample.
var scripts = []struct {
	slug, family, base string
	face               func(string, string) ([]byte, error)
	fallbacks          func() []string
	files              func() fs.FS
	webFaces           func() []arabic.WebFace
	licenses           func() []arabic.License
	row                string
	ranges             []span
	gsub, gpos         []string // feature tags that must survive subsetting
}{
	{"arabic", "Noto Sans Arabic", "NotoSansArabic", arabic.Embedded.Face, arabic.Fallbacks, arabic.Files, arabic.WebFaces, arabic.Licenses,
		"Arabic: مرحبا بالعالم · اللغة العربية · لا · تَجْرِبَة · فارسی",
		[]span{{0x600, 0x6ff}, {0x750, 0x77f}, {0x8a0, 0x8ff}, {0xfb50, 0xfdff}, {0xfe70, 0xfeff}},
		[]string{"init", "medi", "fina", "rlig"}, []string{"mark", "mkmk"}},
	{"hebrew", "Noto Sans Hebrew", "NotoSansHebrew", hebrew.Embedded.Face, hebrew.Fallbacks, hebrew.Files, hebrew.WebFaces, hebrew.Licenses,
		"Hebrew: שלום עולם · עברית",
		[]span{{0x590, 0x5ff}, {0xfb1d, 0xfb4f}},
		nil, []string{"mark"}},
	{"devanagari", "Noto Sans Devanagari", "NotoSansDevanagari", devanagari.Embedded.Face, devanagari.Fallbacks, devanagari.Files, devanagari.WebFaces, devanagari.Licenses,
		"Devanagari: नमस्ते दुनिया · क्ष त्र ज्ञ",
		[]span{{0x900, 0x97f}, {0x1cd0, 0x1cff}, {0xa8e0, 0xa8ff}},
		[]string{"akhn", "abvs", "blwf", "blws", "cjct", "half", "locl", "nukt", "pres", "psts", "rkrf", "rphf"}, []string{"abvm", "blwm"}},
	{"bengali", "Noto Sans Bengali", "NotoSansBengali", bengali.Embedded.Face, bengali.Fallbacks, bengali.Files, bengali.WebFaces, bengali.Licenses,
		"Bengali: বাংলা ভাষা · শুভ সন্ধ্যা",
		[]span{{0x980, 0x9ff}},
		[]string{"akhn", "half", "pstf", "vatu"}, []string{"abvm", "blwm"}},
	{"tamil", "Noto Sans Tamil", "NotoSansTamil", tamil.Embedded.Face, tamil.Fallbacks, tamil.Files, tamil.WebFaces, tamil.Licenses,
		"Tamil: தமிழ் மொழி · வணக்கம்",
		[]span{{0xb80, 0xbff}, {0x11fc0, 0x11fff}},
		[]string{"akhn", "haln"}, []string{"abvm"}},
	{"thai", "Noto Sans Thai", "NotoSansThai", thai.Embedded.Face, thai.Fallbacks, thai.Files, thai.WebFaces, thai.Licenses,
		"Thai: สวัสดีชาวโลก · ภาษาไทย",
		[]span{{0xe00, 0xe7f}},
		[]string{"ccmp"}, []string{"mark", "mkmk"}},
	{"khmer", "Noto Sans Khmer", "NotoSansKhmer", khmer.Embedded.Face, khmer.Fallbacks, khmer.Files, khmer.WebFaces, khmer.Licenses,
		"Khmer: ជម្រាបសួរ ពិភពលោក · ភាសាខ្មែរ",
		[]span{{0x1780, 0x17ff}, {0x19e0, 0x19ff}},
		[]string{"abvf", "blwf", "pref", "pres", "pstf", "psts"}, []string{"abvm", "blwm"}},
}

var weights = []string{"Regular", "Bold"}

func TestFacesAndAliases(t *testing.T) {
	for _, s := range scripts {
		t.Run(s.slug, func(t *testing.T) {
			got := s.fallbacks()
			if !slices.Equal(got, []string{s.family}) {
				t.Fatal(got)
			}
			got[0] = "changed"
			if s.fallbacks()[0] != s.family {
				t.Fatal("fallback slice is shared")
			}
			face := func(v string) []byte {
				data, err := s.face(s.family, v)
				if err != nil {
					t.Fatal(v, err)
				}
				return data
			}
			regular, bold := face("Regular"), face("Bold")
			if !bytes.Equal(face("Italic"), regular) || !bytes.Equal(face("BoldItalic"), bold) {
				t.Fatal("italic variants must alias upright weights")
			}
			if bytes.Equal(regular, bold) {
				t.Fatal("weights must differ")
			}
			for _, tc := range []struct{ family, variant, offender string }{
				{s.family, "regular", `"regular"`}, {s.family, "", `""`}, {"Nope", "Regular", `"Nope"`},
			} {
				_, err := s.face(tc.family, tc.variant)
				if err == nil || !strings.HasPrefix(err.Error(), "packs: ") || !strings.Contains(err.Error(), tc.offender) {
					t.Fatalf("%s/%s: %v", tc.family, tc.variant, err)
				}
			}
			regular[0] ^= 0xff
			if !bytes.Equal(face("Regular"), face("Italic")) {
				t.Fatal("caller modified embedded font")
			}
		})
	}
}

func TestCorpusCoverage(t *testing.T) {
	for _, s := range scripts {
		for _, w := range weights {
			t.Run(s.slug+"/"+w, func(t *testing.T) {
				has := cmapOf(t, s.face, s.family, w)
				for _, r := range s.row {
					if r >= 0x80 && r != '·' && !has(r) {
						t.Errorf("missing %U %q", r, r)
					}
				}
			})
		}
	}
}

func TestExclusions(t *testing.T) {
	for _, s := range scripts {
		for _, w := range weights {
			has := cmapOf(t, s.face, s.family, w)
			for _, r := range "0123456789AaZz" {
				if has(r) {
					t.Errorf("%s/%s: Latin %q must not be in a script face", s.slug, w, r)
				}
			}
		}
	}
}

func TestControls(t *testing.T) {
	raw, err := os.ReadFile("coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var coverage map[string]string
	if err := json.Unmarshal(raw, &coverage); err != nil {
		t.Fatal(err)
	}
	for _, s := range scripts {
		for _, w := range weights {
			key := strings.TrimPrefix(s.base, "NotoSans") + "/" + w
			recorded, ok := coverage[key]
			if !ok {
				t.Fatalf("coverage.json lacks %s", key)
			}
			has := cmapOf(t, s.face, s.family, w)
			for _, r := range []rune{0x200c, 0x200d} {
				if has(r) != strings.ContainsRune(recorded, r) {
					t.Errorf("%s: %U presence %v disagrees with coverage.json", key, r, has(r))
				}
			}
		}
	}
}

func TestLayoutFeatures(t *testing.T) {
	for _, s := range scripts {
		for _, w := range weights {
			t.Run(s.slug+"/"+w, func(t *testing.T) {
				data, err := s.face(s.family, w)
				if err != nil {
					t.Fatal(err)
				}
				tbl, err := tables(data)
				if err != nil {
					t.Fatal(err)
				}
				gsub, gpos := featureTags(tbl["GSUB"]), featureTags(tbl["GPOS"])
				for _, tag := range s.gsub {
					if !gsub[tag] {
						t.Errorf("GSUB lacks %s", tag)
					}
				}
				for _, tag := range s.gpos {
					if !gpos[tag] {
						t.Errorf("GPOS lacks %s", tag)
					}
				}
				// Unencoded glyphs (contextual forms, conjuncts) must survive the
				// subset. Arabic encodes most forms as Presentation Forms, so its
				// margin is small (measured 1263 glyphs vs 1179 mapped); Devanagari
				// conjuncts are unencoded (742 vs 206).
				has := cmapOf(t, s.face, s.family, w)
				mapped := 0
				for _, sp := range s.ranges {
					for r := sp.first; r <= sp.last; r++ {
						if has(r) {
							mapped++
						}
					}
				}
				glyphs := numGlyphs(data)
				switch s.slug {
				case "arabic":
					if glyphs <= mapped {
						t.Errorf("Arabic has %d glyphs for %d code points: contextual forms dropped", glyphs, mapped)
					}
				case "devanagari":
					if glyphs <= 3*mapped {
						t.Errorf("Devanagari has %d glyphs for %d code points: conjuncts dropped", glyphs, mapped)
					}
				}
			})
		}
	}
}

func TestFilesAndWebFaces(t *testing.T) {
	for _, s := range scripts {
		t.Run(s.slug, func(t *testing.T) {
			want := []arabic.WebFace{{Family: s.family, File: s.base + "-Regular.ttf", Weight: 400, Style: "normal"}, {Family: s.family, File: s.base + "-Bold.ttf", Weight: 700, Style: "normal"}}
			got := s.webFaces()
			if !slices.Equal(got, want) {
				t.Fatalf("%+v", got)
			}
			for i, wf := range got {
				data, err := fs.ReadFile(s.files(), wf.File)
				if err != nil {
					t.Fatal(err)
				}
				face, _ := s.face(s.family, weights[i])
				if !bytes.Equal(data, face) {
					t.Fatalf("%s differs from Face", wf.File)
				}
			}
			entries, err := fs.ReadDir(s.files(), ".")
			if err != nil {
				t.Fatal(err)
			}
			var names []string
			for _, e := range entries {
				names = append(names, e.Name())
			}
			if !slices.Equal(names, []string{s.base + "-Bold.ttf", s.base + "-Regular.ttf"}) {
				t.Fatalf("unexpected files %v", names)
			}
			for _, bad := range []string{"../x", "/abs"} {
				if _, err := s.files().Open(bad); err == nil {
					t.Fatalf("opened %q", bad)
				}
			}
			got[0].Family = "changed"
			if s.webFaces()[0].Family != s.family {
				t.Fatal("WebFaces slice is shared")
			}
		})
	}
}

func TestLicenses(t *testing.T) {
	for _, s := range scripts {
		got := s.licenses()
		if len(got) != 1 || got[0].Family != s.family || !strings.Contains(got[0].Text, "SIL OPEN FONT LICENSE") {
			t.Fatalf("%s: %+v", s.slug, got)
		}
	}
}

func TestHashes(t *testing.T) {
	manifest, err := os.ReadFile("subsets.sha256")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(manifest)), "\n")
	if len(lines) != 14 {
		t.Fatalf("expected fourteen faces, got %d", len(lines))
	}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatal("invalid manifest line", line)
		}
		data, err := os.ReadFile(fields[1])
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != fields[0] {
			t.Fatal("hash mismatch", fields[1])
		}
	}
}

func TestBudget(t *testing.T) {
	total := 0
	for _, s := range scripts {
		for _, w := range weights {
			data, err := s.face(s.family, w)
			if err != nil {
				t.Fatal(err)
			}
			total += len(data)
		}
	}
	if total > budget {
		t.Fatalf("embedded packs total %d bytes, budget %d", total, budget)
	}
	t.Logf("packs total %d bytes", total)
}

func cmapOf(t *testing.T, face func(string, string) ([]byte, error), family, variant string) func(rune) bool {
	t.Helper()
	data, err := face(family, variant)
	if err != nil {
		t.Fatal(err)
	}
	has, err := parseCmap(data)
	if err != nil {
		t.Fatal(err)
	}
	return has
}
