package packs_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestLinkIsolation builds a consumer that imports only packs/arabic and
// checks that no other pack's bytes reach its binary.
func TestLinkIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a tiny consumer")
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	arabicData, err := os.ReadFile("arabic/NotoSansArabic-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}
	hebrewData, err := os.ReadFile("hebrew/NotoSansHebrew-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}
	// Find a 64-byte signature near the middle, proven unique to that face.
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
	arabicSig, hebrewSig := signature(arabicData, hebrewData), signature(hebrewData, arabicData)
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
	for dep := range strings.FieldsSeq(string(run(root, "list", "-deps", "./arabic"))) {
		if dep == "nimbleterminal.dev/ntfonts/packs" || dep == "nimbleterminal.dev/ntfonts/packs/hebrew" {
			t.Fatalf("leaf pulls in %s", dep)
		}
	}
	dir := t.TempDir()
	module := fmt.Sprintf("module isolation\n\ngo 1.26\n\nrequire nimbleterminal.dev/ntfonts/packs v0.0.0\nreplace nimbleterminal.dev/ntfonts/packs => %s\n", filepath.ToSlash(root))
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(module), 0644); err != nil {
		t.Fatal(err)
	}
	main := `package main
import ("os"; "nimbleterminal.dev/ntfonts/packs/arabic")
func main() {
 data,err:=arabic.Embedded.Face("Noto Sans Arabic","Regular")
 if err!=nil {panic(err)}
 if _,err:=os.Stdout.Write(data);err!=nil {panic(err)}
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(main), 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "consumer")
	run(dir, "build", "-mod=readonly", "-o", binary, ".")
	data, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, arabicSig) {
		t.Fatal("selected font not linked")
	}
	if bytes.Contains(data, hebrewSig) {
		t.Fatal("other pack linked into leaf consumer")
	}
}
