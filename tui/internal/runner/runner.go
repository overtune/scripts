// Package runner executes a script and streams its output.
package runner

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"sync"
)

// Line is a single line of script output.
type Line struct {
	Text  string
	IsErr bool
}

// Run executes path with args, streaming output on the returned lines channel
// (closed when the process ends), then sends the exit code on done.
func Run(ctx context.Context, path string, args []string) (<-chan Line, <-chan int) {
	lines := make(chan Line)
	done := make(chan int, 1)

	go func() {
		cmd := exec.CommandContext(ctx, path, args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			emitErr(ctx, lines, err.Error())
			close(lines)
			done <- 1
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			emitErr(ctx, lines, err.Error())
			close(lines)
			done <- 1
			return
		}
		if err := cmd.Start(); err != nil {
			emitErr(ctx, lines, err.Error())
			close(lines)
			done <- 1
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go streamPipe(ctx, stdout, false, lines, &wg)
		go streamPipe(ctx, stderr, true, lines, &wg)
		wg.Wait()

		code := 0
		if err := cmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				code = exitErr.ExitCode()
			} else {
				code = 1
			}
		}
		close(lines)
		done <- code
	}()

	return lines, done
}

// emitErr sends an error Line on lines, but abandons the send if ctx is
// cancelled before the consumer receives it, preventing a permanent block
// when the caller has stopped reading lines.
func emitErr(ctx context.Context, lines chan<- Line, msg string) {
	select {
	case lines <- Line{Text: msg, IsErr: true}:
	case <-ctx.Done():
	}
}

func streamPipe(ctx context.Context, r io.Reader, isErr bool, out chan<- Line, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case out <- Line{Text: scanner.Text(), IsErr: isErr}:
		case <-ctx.Done():
			return
		}
	}
}
