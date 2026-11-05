package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestViewer(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cjk", "scregular"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cjk", "scregular", "GlossCJKSC-Regular.otf"), []byte("optional font"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "cjk", "internal", "notice"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cjk", "internal", "notice", "notosanscjksc.OFL.txt"), []byte("test CJK license"), 0644); err != nil {
		t.Fatal(err)
	}
	// Optional packs: one script face and the emoji font, with their notices.
	fakes := map[string]string{
		"packs/arabic/NotoSansArabic-Regular.ttf":      "fake arabic",
		"packs/internal/notice/notosansarabic.OFL.txt": "test arabic license",
		"packs/emoji/NotoColorEmoji.ttf":               "fake emoji",
		"packs/emoji/notocoloremoji.OFL.txt":           "test emoji license",
		"packs/emoji/sbix/NotoColorEmoji-sbix.ttf":     "fake sbix",
		"packs/emoji/sbix/notocoloremoji.OFL.txt":      "test emoji license",
	}
	for name, content := range fakes {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	handler, err := viewer(root)
	if err != nil {
		t.Fatal(err)
	}
	request := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w
	}
	w := request("/fonts.json")
	var fonts []face
	if err := json.Unmarshal(w.Body.Bytes(), &fonts); err != nil {
		t.Fatal(err)
	}
	if len(fonts) != 33 {
		t.Fatalf("got %d faces", len(fonts))
	}
	// Both color emoji files share the family; tech tells them apart so the
	// page can pick the format this browser supports. CBDT is listed first.
	packFaces := map[string]string{"Noto Sans Arabic": "fake arabic", "Noto Color Emoji/color-CBDT": "fake emoji", "Noto Color Emoji/color-sbix": "fake sbix"}
	for _, font := range fonts {
		key := font.Family
		if font.Tech != "" {
			key += "/" + font.Tech
		}
		result := request("/" + font.URL)
		if result.Code != 200 || result.Body.Len() != font.Bytes {
			t.Fatalf("invalid font response: %+v", font)
		}
		if font.Group == "cjk" && result.Body.String() != "optional font" {
			t.Fatal("CJK must come from checkout")
		}
		if font.Group == "packs" {
			if want, ok := packFaces[key]; !ok || result.Body.String() != want || !strings.HasSuffix(font.URL, ".ttf") {
				t.Fatalf("pack face must come from checkout: %+v", font)
			}
			delete(packFaces, key)
		} else if font.Tech != "" {
			t.Fatalf("tech on a non-pack face: %+v", font)
		}
	}
	if len(packFaces) != 0 {
		t.Fatalf("pack faces not catalogued: %v", packFaces)
	}
	if i, j := slices.IndexFunc(fonts, func(f face) bool { return f.Tech == "color-CBDT" }), slices.IndexFunc(fonts, func(f face) bool { return f.Tech == "color-sbix" }); i > j {
		t.Fatal("CBDT must precede sbix so Chrome keeps the unmodified file")
	}
	licenses := request("/licenses.txt").Body.String()
	for _, want := range []string{"test CJK license", "test arabic license", "test emoji license"} {
		if !strings.Contains(licenses, want) {
			t.Fatalf("licenses.txt lacks %q", want)
		}
	}
	if strings.Count(licenses, "test emoji license") != 1 {
		t.Fatal("the shared emoji notice must appear once")
	}
	for _, path := range []string{"/sources.json", "/cjk/go.mod", "/fonts/999", "/index.html"} {
		if result := request(path); result.Code != 404 {
			t.Fatalf("%s: status %d", path, result.Code)
		}
	}
	if result := request("/"); result.Code != 200 {
		t.Fatal("missing viewer")
	}
	// Optional module may be absent; the core viewer still works.
	handler, err = viewer(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	w = request("/fonts.json")
	data, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fonts); err != nil {
		t.Fatal(err)
	}
	if len(fonts) != 29 {
		t.Fatalf("core has %d faces", len(fonts))
	}
}

// A project Pages site lives under /ntfonts/, not the domain root. Exercise
// the actual exported files at that prefix, including every font and license.
func TestStaticExport(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(t.TempDir(), "site")
	if err := exportSite(root, destination); err != nil {
		t.Fatal(err)
	}
	if err := exportSite(root, destination); err == nil {
		t.Fatal("must not mix builds in an existing directory")
	}
	handler := http.StripPrefix("/ntfonts/", http.FileServer(http.Dir(destination)))
	request := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ntfonts/"+path, nil))
		return w
	}
	index := request("")
	if index.Code != 200 || !strings.Contains(index.Body.String(), "fetch('fonts.json')") {
		t.Fatal("index must load relative catalog")
	}
	if !strings.Contains(index.Body.String(), `href="licenses.txt"`) {
		t.Fatal("missing license link")
	}
	catalog := request("fonts.json")
	var fonts []face
	if err := json.Unmarshal(catalog.Body.Bytes(), &fonts); err != nil {
		t.Fatal(err)
	}
	if len(fonts) != 29 {
		t.Fatal("incomplete export")
	}
	for _, font := range fonts {
		if strings.HasPrefix(font.URL, "/") {
			t.Fatal("font URL escapes project prefix")
		}
		response := request(font.URL)
		if response.Code != 200 || response.Body.Len() != font.Bytes {
			t.Fatal("broken exported font", font.URL)
		}
	}
	licenses := request("licenses.txt")
	if licenses.Code != 200 || !strings.Contains(licenses.Body.String(), "SIL OPEN FONT LICENSE") {
		t.Fatal("missing OFL notices")
	}
	if _, err := os.Stat(filepath.Join(destination, ".nojekyll")); err != nil {
		t.Fatal(err)
	}
}
