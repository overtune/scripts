package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"scriptstui/internal/meta"
)

type argForm struct {
	args   []meta.Arg
	inputs []textinput.Model
	focus  int
}

func newArgForm(args []meta.Arg) argForm {
	inputs := make([]textinput.Model, len(args))
	for i, a := range args {
		ti := textinput.New()
		ti.Placeholder = a.Help
		ti.Prompt = a.Name + ": "
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}
	return argForm{args: args, inputs: inputs}
}

func (f argForm) values() []string {
	vals := make([]string, len(f.inputs))
	for i, in := range f.inputs {
		vals[i] = strings.TrimSpace(in.Value())
	}
	return vals
}

// update handles focus movement and typing; returns the updated form.
func (f argForm) update(msg tea.Msg) (argForm, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "tab", "down":
			f.setFocus(f.focus + 1)
			return f, nil
		case "shift+tab", "up":
			f.setFocus(f.focus - 1)
			return f, nil
		}
	}
	var cmd tea.Cmd
	f.inputs[f.focus], cmd = f.inputs[f.focus].Update(msg)
	return f, cmd
}

func (f *argForm) setFocus(i int) {
	if i < 0 {
		i = len(f.inputs) - 1
	}
	if i >= len(f.inputs) {
		i = 0
	}
	for j := range f.inputs {
		if j == i {
			f.inputs[j].Focus()
		} else {
			f.inputs[j].Blur()
		}
	}
	f.focus = i
}

func (f argForm) view() string {
	var b strings.Builder
	b.WriteString("Arguments (enter to run, esc to cancel):\n\n")
	for _, in := range f.inputs {
		b.WriteString(in.View())
		b.WriteString("\n")
	}
	return b.String()
}

// missingRequired returns the names of required args with empty values.
func missingRequired(args []meta.Arg, values []string) []string {
	var missing []string
	for i, a := range args {
		if a.Required && (i >= len(values) || strings.TrimSpace(values[i]) == "") {
			missing = append(missing, a.Name)
		}
	}
	return missing
}
