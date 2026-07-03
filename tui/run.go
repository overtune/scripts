package main

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"scriptstui/internal/runner"
)

type outputLineMsg runner.Line
type doneMsg struct{ code int }

type runSession struct {
	lines <-chan runner.Line
	done  <-chan int
}

// runnerRun is a seam over runner.Run so the model can start a script and
// hold onto the resulting channels via a runSession.
func runnerRun(path string, args []string) (<-chan runner.Line, <-chan int) {
	return runner.Run(context.Background(), path, args)
}

func (s *runSession) readNext() tea.Cmd {
	return func() tea.Msg {
		select {
		case l, ok := <-s.lines:
			if ok {
				return outputLineMsg(l)
			}
			return doneMsg{code: <-s.done}
		}
	}
}
