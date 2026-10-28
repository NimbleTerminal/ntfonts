package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestViewer(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "cjk"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cjk", "GlossCJKSC-Regular.otf"), []byte("optional font"), 0644); err != nil {
		t.Fatal(err)
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
	w := request("/api/fonts")
	var fonts []face
	if err := json.Unmarshal(w.Body.Bytes(), &fonts); err != nil {
		t.Fatal(err)
	}
	if len(fonts) != 30 {
		t.Fatalf("got %d faces", len(fonts))
	}
	for _, font := range fonts {
		result := request(font.URL)
		if result.Code != 200 || result.Body.Len() != font.Bytes {
			t.Fatalf("invalid font response: %+v", font)
		}
		if font.Group == "cjk" && result.Body.String() != "optional font" {
			t.Fatal("CJK must come from checkout")
		}
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
	w = request("/api/fonts")
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
