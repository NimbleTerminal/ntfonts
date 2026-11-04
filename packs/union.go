package packs

import (
	"io"
	"io/fs"
	"time"
)

// union serves several read-only file systems as one. File names are unique
// across packs, so the first system that opens a name wins; "." lists every
// root entry so http.FS and fstest.TestFS see a complete directory.
type union []fs.FS

func (u union) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	if name == "." {
		var entries []fs.DirEntry
		for _, sub := range u {
			found, err := fs.ReadDir(sub, ".")
			if err != nil {
				return nil, err
			}
			entries = append(entries, found...)
		}
		return &unionDir{entries: entries}, nil
	}
	for _, sub := range u {
		f, err := sub.Open(name)
		if err == nil {
			return f, nil
		}
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

// unionDir is the synthetic root directory of a union.
type unionDir struct {
	entries []fs.DirEntry
	offset  int
}

func (d *unionDir) Stat() (fs.FileInfo, error) { return rootInfo{}, nil }
func (d *unionDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: ".", Err: fs.ErrInvalid}
}
func (d *unionDir) Close() error { return nil }

func (d *unionDir) ReadDir(n int) ([]fs.DirEntry, error) {
	rest := d.entries[d.offset:]
	if n <= 0 {
		d.offset = len(d.entries)
		return rest, nil
	}
	if len(rest) == 0 {
		return nil, io.EOF
	}
	if n > len(rest) {
		n = len(rest)
	}
	d.offset += n
	return rest[:n], nil
}

// rootInfo describes the union's root directory.
type rootInfo struct{}

func (rootInfo) Name() string       { return "." }
func (rootInfo) Size() int64        { return 0 }
func (rootInfo) Mode() fs.FileMode  { return fs.ModeDir | 0555 }
func (rootInfo) ModTime() time.Time { return time.Time{} }
func (rootInfo) IsDir() bool        { return true }
func (rootInfo) Sys() any           { return nil }
