package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"scriptstui/internal/meta"
)

var detailStyle = lipgloss.NewStyle().Padding(0, 2).Border(lipgloss.RoundedBorder(), false, false, false, true)

type appState int

const (
	stateList appState = iota
	stateArgs
	stateRunning
)

type model struct {
	list    list.Model
	width   int
	height  int
	state   appState
	session *runSession
	output  []string
	exited  bool
	code    int
	form    argForm
	pending meta.ScriptMeta // script awaiting args
	formErr string
}

func newModel(items []list.Item) model {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Scripts"
	return model{list: l}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width/2, msg.Height-1)
		return m, nil

	case outputLineMsg:
		m.output = append(m.output, msg.Text)
		return m, m.session.readNext()

	case doneMsg:
		m.exited = true
		m.code = msg.code
		return m, nil

	case tea.KeyMsg:
		if m.state == stateArgs {
			switch msg.String() {
			case "esc":
				m.state = stateList
				return m, nil
			case "enter":
				vals := m.form.values()
				if miss := missingRequired(m.pending.Args, vals); len(miss) > 0 {
					m.formErr = "Required: " + strings.Join(miss, ", ")
					return m, nil
				}
				m.state = stateRunning
				m.output = nil
				m.exited = false
				return m, startRunFor(&m, m.pending.Path, vals)
			}
			var cmd tea.Cmd
			m.form, cmd = m.form.update(msg)
			return m, cmd
		}
		if m.state == stateRunning {
			if msg.String() == "esc" && m.exited {
				m.state = stateList
				m.output = nil
				m.exited = false
				return m, nil
			}
			return m, nil
		}
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				if sel, ok := m.list.SelectedItem().(scriptItem); ok {
					if len(sel.m.Args) == 0 {
						m.state = stateRunning
						m.output = nil
						m.exited = false
						return m, startRunFor(&m, sel.m.Path, nil)
					}
					m.state = stateArgs
					m.pending = sel.m
					m.form = newArgForm(sel.m.Args)
					m.formErr = ""
					return m, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// startRunFor launches path with args, storing the resulting session on m so
// the model can keep pumping reads via session.readNext().
func startRunFor(m *model, path string, args []string) tea.Cmd {
	lines, done := runnerRun(path, args)
	m.session = &runSession{lines: lines, done: done}
	return m.session.readNext()
}

func (m model) View() string {
	if m.state == stateArgs {
		v := m.form.view()
		if m.formErr != "" {
			v += "\n" + m.formErr + "\n"
		}
		return v
	}
	if m.state == stateRunning {
		return m.runningView()
	}
	left := m.list.View()
	right := detailStyle.Width(m.width/2 - 4).Render(m.detailView())
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m model) runningView() string {
	var b strings.Builder
	b.WriteString("Running...\n\n")
	for _, line := range m.output {
		b.WriteString(line)
		b.WriteString("\n")
	}
	if m.exited {
		status := "success"
		if m.code != 0 {
			status = "failed"
		}
		fmt.Fprintf(&b, "\nExit code: %d (%s)\nPress esc to return.\n", m.code, status)
	}
	return b.String()
}

func (m model) detailView() string {
	sel, ok := m.list.SelectedItem().(scriptItem)
	if !ok {
		return "No script selected."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", sel.m.Name)
	if sel.m.Description != "" {
		fmt.Fprintf(&b, "%s\n\n", sel.m.Description)
	}
	fmt.Fprintf(&b, "Category: %s\n", sel.m.Category)
	fmt.Fprintf(&b, "Path: %s\n", sel.m.Path)
	if len(sel.m.Args) > 0 {
		b.WriteString("\nArguments:\n")
		for _, a := range sel.m.Args {
			req := ""
			if a.Required {
				req = " (required)"
			}
			fmt.Fprintf(&b, "  - %s%s: %s\n", a.Name, req, a.Help)
		}
	}
	if preview := sourcePreview(sel.m.Path, 12); preview != "" {
		fmt.Fprintf(&b, "\nSource preview:\n%s", preview)
	}
	return b.String()
}

// sourcePreview returns up to n lines from the top of the file at path.
func sourcePreview(path string, n int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	var b strings.Builder
	scanner := bufio.NewScanner(f)
	for i := 0; i < n && scanner.Scan(); i++ {
		b.WriteString("  ")
		b.WriteString(scanner.Text())
		b.WriteString("\n")
	}
	return b.String()
}
