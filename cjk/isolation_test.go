package cjk_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"nimbleterminal.dev/ntfonts/cjk"
	"nimbleterminal.dev/ntfonts/cjk/scbold"
	"nimbleterminal.dev/ntfonts/cjk/scregular"
)

func TestLeafFaces(t *testing.T) {
	for _, tc := range []struct {
		name, weight string
		source       interface {
			Face(string, string) ([]byte, error)
		}
		families []string
		licenses []cjk.License
	}{
		{"scregular", "Regular", scregular.Embedded, scregular.Fallbacks(), scregular.Licenses()},
		{"scbold", "Bold", scbold.Embedded, scbold.Fallbacks(), scbold.Licenses()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.families) != 1 || tc.families[0] != "Noto Sans CJK SC" {
				t.Fatal(tc.families)
			}
			if len(tc.licenses) != 1 || tc.licenses[0] != cjk.Licenses()[0] {
				t.Fatal("inconsistent shared notice")
			}
			want, err := cjk.Embedded.Face(tc.families[0], tc.weight)
			if err != nil {
				t.Fatal(err)
			}
			for _, variant := range []string{"Regular", "Italic", "Bold", "BoldItalic"} {
				data, err := tc.source.Face(tc.families[0], variant)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(data, want) {
					t.Fatalf("%s did not alias %s", variant, tc.weight)
				}
				cmap, err := parseCmap(data)
				if err != nil {
					t.Fatal(err)
				}
				for _, r := range "漢字かなカナ한글。" {
					if !cmap(r) {
						t.Fatalf("missing %U", r)
					}
				}
				if cmap('𠀀') {
					t.Fatal("out-of-scope Han included")
				}
				// Keep the existing fresh-byte behavior of embed.FS.ReadFile.
				data[0] ^= 0xff
				fresh, err := tc.source.Face(tc.families[0], variant)
				if err != nil || !bytes.Equal(fresh, want) {
					t.Fatal("caller modified embedded font")
				}
			}
			for _, args := range [][2]string{{"Nope", "Regular"}, {tc.families[0], "Heavy"}} {
				if _, err := tc.source.Face(args[0], args[1]); err == nil {
					t.Fatal("invalid face accepted")
				}
			}
		})
	}
	regular, _ := cjk.Embedded.Face("Noto Sans CJK SC", "Regular")
	bold, _ := cjk.Embedded.Face("Noto Sans CJK SC", "Bold")
	if bytes.Equal(regular, bold) {
		t.Fatal("combined source lost separate weights")
	}
}

func TestLinkIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("builds two tiny consumers")
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	regular, err := os.ReadFile("scregular/GlossCJKSC-Regular.otf")
	if err != nil {
		t.Fatal(err)
	}
	bold, err := os.ReadFile("scbold/GlossCJKSC-Bold.otf")
	if err != nil {
		t.Fatal(err)
	}
	// Find a 64-byte signature near the middle, proven unique to that weight.
	signature := func(data, other []byte) []byte {
		for offset := len(data) / 2; offset+64 <= len(data); offset += 64 {
			window := data[offset : offset+64]
			if !bytes.Contains(other, window) {
				return window
			}
		}
		t.Fatal("no distinct font signature")
		return nil
	}
	regularSig, boldSig := signature(regular, bold), signature(bold, regular)
	for _, tc := range []struct {
		pkg, other      string
		present, absent []byte
	}{
		{"scregular", "scbold", regularSig, boldSig},
		{"scbold", "scregular", boldSig, regularSig},
	} {
		t.Run(tc.pkg, func(t *testing.T) {
			run := func(dir string, args ...string) []byte {
				cmd := exec.Command("go", args...)
				cmd.Dir = dir
				cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=")
				output, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("go %v: %v\n%s", args, err, output)
				}
				return output
			}
			deps := strings.Fields(string(run(root, "list", "-deps", "./"+tc.pkg)))
			for _, dep := range deps {
				if dep == "nimbleterminal.dev/ntfonts/cjk" || dep == "nimbleterminal.dev/ntfonts/cjk/"+tc.other {
					t.Fatalf("leaf pulls in %s", dep)
				}
			}
			dir := t.TempDir()
			module := fmt.Sprintf("module isolation\n\ngo 1.26\n\nrequire nimbleterminal.dev/ntfonts/cjk v0.0.0\nreplace nimbleterminal.dev/ntfonts/cjk => %s\n", filepath.ToSlash(root))
			if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(module), 0644); err != nil {
				t.Fatal(err)
			}
			main := fmt.Sprintf(`package main
import ("os"; "nimbleterminal.dev/ntfonts/cjk/%s")
func main() {
 data,err:=%s.Embedded.Face("Noto Sans CJK SC","Regular")
 if err!=nil {panic(err)}
 if _,err:=os.Stdout.Write(data);err!=nil {panic(err)}
}
`, tc.pkg, tc.pkg)
			if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(main), 0644); err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(dir, "consumer")
			run(dir, "build", "-mod=readonly", "-o", binary, ".")
			data, err := os.ReadFile(binary)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(data, tc.present) {
				t.Fatal("selected font not linked")
			}
			if bytes.Contains(data, tc.absent) {
				t.Fatal("other weight linked into leaf consumer")
			}
		})
	}
}
