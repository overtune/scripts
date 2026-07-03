package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var detailStyle = lipgloss.NewStyle().Padding(0, 2).Border(lipgloss.RoundedBorder(), false, false, false, true)

type model struct {
	list   list.Model
	width  int
	height int
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
	case tea.KeyMsg:
		if m.list.FilterState() != list.Filtering {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	left := m.list.View()
	right := detailStyle.Width(m.width/2 - 4).Render(m.detailView())
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
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
