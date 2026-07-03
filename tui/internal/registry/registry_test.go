package registry

import (
	"os"
	"path/filepath"
	"testing"

	"scriptstui/internal/meta"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestScanFindsDocumentedAndUndocumented(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "net", "portkill.sh"), `#!/bin/bash
# @meta
# name: portkill
# description: kill a port
# category: net
# @end
echo hi
`)
	writeFile(t, filepath.Join(root, "dev", "raw.py"), "print('hi')\n")
	// Should be skipped:
	writeFile(t, filepath.Join(root, "tui", "main.go"), "package main\n")
	writeFile(t, filepath.Join(root, "docs", "note.sh"), "echo skip\n")

	got, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 scripts, got %d: %+v", len(got), got)
	}

	byName := map[string]bool{}
	for _, m := range got {
		byName[m.Name] = m.Documented
	}
	if documented, ok := byName["portkill"]; !ok || !documented {
		t.Fatalf("portkill missing or not documented: %+v", got)
	}
	if documented, ok := byName["raw.py"]; !ok || documented {
		t.Fatalf("expected undocumented raw.py, got: %+v", got)
	}
}

func TestScanDistinguishesMalformedMetaFromNoMeta(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "net", "broken.sh"), `#!/bin/bash
# @meta
# name: : :
# @end
echo hi
`)
	writeFile(t, filepath.Join(root, "dev", "plain.sh"), "echo hi\n")

	got, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 scripts, got %d: %+v", len(got), got)
	}

	byName := map[string]meta.ScriptMeta{}
	for _, m := range got {
		byName[m.Name] = m
	}

	broken, ok := byName["broken.sh"]
	if !ok {
		t.Fatalf("broken.sh missing from scan: %+v", got)
	}
	if broken.Documented {
		t.Fatalf("expected broken.sh to be undocumented, got: %+v", broken)
	}
	if broken.Warning == "" {
		t.Fatalf("expected broken.sh to have a non-empty parse warning, got: %+v", broken)
	}

	plain, ok := byName["plain.sh"]
	if !ok {
		t.Fatalf("plain.sh missing from scan: %+v", got)
	}
	if plain.Documented {
		t.Fatalf("expected plain.sh to be undocumented, got: %+v", plain)
	}
	if plain.Warning != "" {
		t.Fatalf("expected plain.sh to have no parse warning, got: %+v", plain)
	}
}

func TestScanUndocumentedCategoryFromFolder(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "files", "cleanup.sh"), "echo hi\n")
	got, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Category != "files" {
		t.Fatalf("expected category files, got: %+v", got)
	}
}
