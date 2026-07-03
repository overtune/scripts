package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"scriptstui/internal/registry"
)

func discoverRoot() string {
	rootFlag := flag.String("root", "", "path to the scripts collection root")
	flag.Parse()
	if *rootFlag != "" {
		return *rootFlag
	}
	if env := os.Getenv("SCRIPTS_HOME"); env != "" {
		return env
	}
	// Default: parent of the tui/ working directory.
	return ".."
}

func main() {
	root := discoverRoot()
	metas, err := registry.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to scan %q: %v\n", root, err)
		os.Exit(1)
	}

	p := tea.NewProgram(newModel(newItems(metas)), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
