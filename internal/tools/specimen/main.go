// specimen serves a local browser preview of the repository's font subsets.
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"nimbleterminal.dev/ntfonts"
)

//go:embed index.html
var page embed.FS

type face struct {
	Family  string `json:"family"`
	Variant string `json:"variant"`
	Group   string `json:"group"`
	URL     string `json:"url"`
	Bytes   int    `json:"bytes"`
	data    []byte
}

func viewer(root string) (http.Handler, error) {
	var faces []face
	variants := []string{"Regular", "Bold", "Italic", "BoldItalic"}
	for _, family := range append(ntfonts.Families(), ntfonts.Fallbacks()...) {
		group := "text"
		names := variants
		if family == "Noto Sans Symbols 2" {
			group = "symbols"
			names = variants[:1]
		}
		for _, variant := range names {
			data, err := ntfonts.Embedded.Face(family, variant)
			if err != nil {
				return nil, err
			}
			faces = append(faces, face{Family: family, Variant: variant, Group: group, Bytes: len(data), data: data})
		}
	}
	// Read optional CJK assets from the checkout; never import/embed the CJK module.
	for _, variant := range variants[:2] {
		name := filepath.Join(root, "cjk", "GlossCJKSC-"+variant+".otf")
		data, err := os.ReadFile(name)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		faces = append(faces, face{Family: "Noto Sans CJK SC", Variant: variant, Group: "cjk", Bytes: len(data), data: data})
	}
	mux := http.NewServeMux()
	for i := range faces {
		f := &faces[i]
		f.URL = fmt.Sprintf("/fonts/%d", i)
		mux.HandleFunc("GET "+f.URL, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "font/ttf")
			if f.Group == "cjk" {
				w.Header().Set("Content-Type", "font/otf")
			}
			w.Header().Set("Cache-Control", "no-cache")
			_, _ = w.Write(f.data)
		})
	}
	mux.HandleFunc("GET /api/fonts", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(faces)
	})
	html, err := fs.ReadFile(page, "index.html")
	if err != nil {
		return nil, err
	}
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(html)
	})
	return mux, nil
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8088", "local listen address")
	root := flag.String("root", ".", "repository root containing optional cjk assets")
	flag.Parse()
	handler, err := viewer(*root)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Specimen viewer: http://%s (Ctrl-C to stop)", *addr)
	log.Fatal(server.ListenAndServe())
}
