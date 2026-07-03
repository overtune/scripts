// Package registry discovers scripts under a collection root.
package registry

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"scriptstui/internal/meta"
)

var scriptExts = map[string]bool{".sh": true, ".py": true, ".js": true}
var skipDirs = map[string]bool{".git": true, "tui": true, "docs": true, "templates": true}

// Scan walks root and returns metadata for every recognized script file.
func Scan(root string) ([]meta.ScriptMeta, error) {
	var out []meta.ScriptMeta
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && skipDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if !scriptExts[strings.ToLower(filepath.Ext(d.Name()))] {
			return nil
		}
		content, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		m, perr := meta.Parse(content)
		if perr != nil {
			m = meta.ScriptMeta{Name: d.Name()}
			if !errors.Is(perr, meta.ErrNoMetaBlock) {
				m.Warning = "metadata parse error: " + perr.Error()
			}
		}
		if m.Category == "" {
			m.Category = filepath.Base(filepath.Dir(path))
		}
		m.Path = path
		out = append(out, m)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
