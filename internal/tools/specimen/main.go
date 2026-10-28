// specimen serves or exports a browser preview of the repository's font subsets.
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func siteFiles(root string) (map[string][]byte, error) {
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

	files := make(map[string][]byte)
	for i := range faces {
		f := &faces[i]
		ext := ".ttf"
		if f.Group == "cjk" {
			ext = ".otf"
		}
		f.URL = fmt.Sprintf("fonts/%d%s", i, ext)
		files[f.URL] = f.data
	}
	catalog, err := json.Marshal(faces)
	if err != nil {
		return nil, err
	}
	files["fonts.json"] = catalog
	html, err := fs.ReadFile(page, "index.html")
	if err != nil {
		return nil, err
	}
	files["index.html"] = html
	var licenses strings.Builder
	for _, license := range ntfonts.Licenses() {
		fmt.Fprintf(&licenses, "%s\n\n%s\n\n", license.Family, license.Text)
	}
	for _, f := range faces {
		if f.Group != "cjk" {
			continue
		}
		license, err := os.ReadFile(filepath.Join(root, "cjk", "notosanscjksc.OFL.txt"))
		if err != nil {
			return nil, fmt.Errorf("CJK license: %w", err)
		}
		fmt.Fprintf(&licenses, "Noto Sans CJK SC\n\n%s\n", license)
		break
	}
	files["licenses.txt"] = []byte(licenses.String())
	files[".nojekyll"] = []byte{}
	return files, nil
}

func viewer(root string) (http.Handler, error) {
	files, err := siteFiles(root)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	for name, data := range files {
		route := "/" + name
		if name == "index.html" {
			route = "/{$}"
		}
		mux.HandleFunc("GET "+route, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(name)))
			w.Header().Set("Cache-Control", "no-cache")
			_, _ = w.Write(data)
		})
	}
	return mux, nil
}

// exportSite writes into a new directory, avoiding stale files from prior builds.
func exportSite(root, destination string) error {
	files, err := siteFiles(root)
	if err != nil {
		return err
	}
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("output directory already exists: %s (choose a fresh directory)", destination)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Join(destination, "fonts"), 0755); err != nil {
		return err
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(destination, filepath.FromSlash(name)), data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8088", "local listen address")
	root := flag.String("root", ".", "repository root containing optional cjk assets")
	output := flag.String("out", "", "export static site to a new directory instead of serving")
	flag.Parse()
	if *output != "" {
		if err := exportSite(*root, *output); err != nil {
			log.Fatal(err)
		}
		log.Printf("Static specimen exported to %s", *output)
		return
	}
	handler, err := viewer(*root)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Specimen viewer: http://%s (Ctrl-C to stop)", *addr)
	log.Fatal(server.ListenAndServe())
}
