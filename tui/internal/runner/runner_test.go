package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
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

	var sawErrLine bool
	for l := range lines {
		if l.IsErr {
			sawErrLine = true
		}
	}
	if !sawErrLine {
		t.Fatalf("expected at least one error Line for missing script")
	}
	if code := <-done; code == 0 {
		t.Fatalf("expected non-zero exit code for missing script")
	}
}

// TestRunCancelUnblocksOnAbandonedConsumer verifies that cancelling the
// context releases blocked sends in streamPipe when the consumer stops
// reading before EOF, so done still fires instead of leaking goroutines.
func TestRunCancelUnblocksOnAbandonedConsumer(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "many.sh")
	body := "#!/bin/bash\nfor i in $(seq 1 100000); do echo line$i; done\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	lines, done := Run(ctx, script, nil)

	// Read only a couple of lines, then abandon the consumer and cancel.
	<-lines
	<-lines
	cancel()

	select {
	case <-done:
		// success: process reaped and done fired despite abandoned reads.
	case <-time.After(2 * time.Second):
		t.Fatal("leaked/hung: done did not fire after context cancellation")
	}
}
