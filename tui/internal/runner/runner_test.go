package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunStreamsOutputAndExitCode(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "s.sh")
	body := "#!/bin/bash\necho hello\necho oops >&2\nexit 3\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}

	lines, done := Run(context.Background(), script, nil)

	var stdoutSeen, stderrSeen bool
	for l := range lines {
		if l.Text == "hello" && !l.IsErr {
			stdoutSeen = true
		}
		if l.Text == "oops" && l.IsErr {
			stderrSeen = true
		}
	}
	code := <-done

	if !stdoutSeen || !stderrSeen {
		t.Fatalf("missing output: stdout=%v stderr=%v", stdoutSeen, stderrSeen)
	}
	if code != 3 {
		t.Fatalf("expected exit code 3, got %d", code)
	}
}

func TestRunMissingScript(t *testing.T) {
	lines, done := Run(context.Background(), "/nonexistent/xyz.sh", nil)
	for range lines {
	}
	if code := <-done; code == 0 {
		t.Fatalf("expected non-zero exit code for missing script")
	}
}
