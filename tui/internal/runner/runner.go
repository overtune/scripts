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
		defer close(lines)

		cmd := exec.CommandContext(ctx, path, args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			lines <- Line{Text: err.Error(), IsErr: true}
			done <- 1
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			lines <- Line{Text: err.Error(), IsErr: true}
			done <- 1
			return
		}
		if err := cmd.Start(); err != nil {
			lines <- Line{Text: err.Error(), IsErr: true}
			done <- 1
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go streamPipe(stdout, false, lines, &wg)
		go streamPipe(stderr, true, lines, &wg)
		wg.Wait()

		code := 0
		if err := cmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				code = exitErr.ExitCode()
			} else {
				code = 1
			}
		}
		done <- code
	}()

	return lines, done
}

func streamPipe(r io.Reader, isErr bool, out chan<- Line, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		out <- Line{Text: scanner.Text(), IsErr: isErr}
	}
}
