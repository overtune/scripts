package registry

import (
	"os"
	"path/filepath"
	"testing"
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
