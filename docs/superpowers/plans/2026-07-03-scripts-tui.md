# Scripts Collection + TUI Launcher Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an organized home for reusable scripts plus a Go/Bubble Tea TUI that discovers, browses, and runs them, prompting for declared arguments.

**Architecture:** Scripts live in category folders, each carrying an embedded `@meta`/`@end` metadata block. Three small Go packages handle parsing (`meta`), discovery (`registry`), and execution (`runner`); a Bubble Tea `main` package wires them into a list → detail → arg-form → streamed-output flow.

**Tech Stack:** Go, Bubble Tea (`github.com/charmbracelet/bubbletea`), Bubbles (`github.com/charmbracelet/bubbles`), Lip Gloss (`github.com/charmbracelet/lipgloss`), YAML (`gopkg.in/yaml.v3`).

## Global Constraints

- Go 1.22 or newer.
- The Go module lives in `tui/` with module path `scriptstui`. Internal packages are under `scriptstui/internal/...`.
- Metadata is the source of truth for a script's category; when absent, fall back to the containing folder name.
- A script with no valid metadata block is still listed as "undocumented" (name = filename), never skipped and never a hard error.
- Development platform is macOS (darwin) with Homebrew available.
- The metadata block must appear within the first 20 lines of a script file.
- Recognized script extensions: `.sh`, `.py`, `.js`. Comment prefixes stripped when parsing metadata: `#` (bash/python) and `//` (node).

---

### Task 0: Project scaffolding, script migration, and Go module

**Files:**
- Create: `net/portkill.sh` (moved from `portkill.sh`, with metadata added)
- Delete: `portkill.sh`
- Create: `templates/example.sh`
- Create: `git/.gitkeep`, `files/.gitkeep`, `dev/.gitkeep`, `media/.gitkeep`
- Create: `tui/go.mod`
- Modify: `README.md`

**Interfaces:**
- Produces: the folder layout and a documented `net/portkill.sh` used as a fixture reference; module `scriptstui` that later tasks add packages to.

- [ ] **Step 1: Install Go (if missing)**

Run:
```bash
go version || brew install go
go version
```
Expected: prints `go version go1.22...` (or newer). If Homebrew is missing, install from https://brew.sh first.

- [ ] **Step 2: Create the folder structure**

```bash
cd /Users/johanrunesson/Development/private/scripts
mkdir -p net git files dev media templates tui
touch git/.gitkeep files/.gitkeep dev/.gitkeep media/.gitkeep
```

- [ ] **Step 3: Move and document portkill**

```bash
git mv portkill.sh net/portkill.sh
```

Then edit `net/portkill.sh` to add the metadata block directly after the shebang so the whole file reads:

```bash
#!/bin/bash
# @meta
# name: portkill
# description: Kill the process listening on a given TCP port
# category: net
# args:
#   - name: port
#     required: true
#     help: TCP port number to kill
# @end

PORT=$1
if [ "$PORT" ]; then
	lsof -n -i4TCP:$PORT -sTCP:LISTEN -t | xargs kill
	echo "$PORT is dead, long live $PORT!"
else
	echo "Must specify a port! Else I don't know who to kill..."
	exit 1
fi
```

- [ ] **Step 4: Create the template script**

Create `templates/example.sh`:

```bash
#!/bin/bash
# @meta
# name: example
# description: A starter script showing the metadata convention
# category: templates
# args:
#   - name: message
#     required: false
#     help: Text to echo back
# @end

echo "Hello from the example script. You said: ${1:-nothing}"
```

Then: `chmod +x templates/example.sh`

- [ ] **Step 5: Initialize the Go module**

```bash
cd /Users/johanrunesson/Development/private/scripts/tui
go mod init scriptstui
```
Expected: creates `tui/go.mod` containing `module scriptstui` and a `go 1.2x` line.

- [ ] **Step 6: Update README**

Replace `README.md` contents with:

```markdown
# scripts

A collection of reusable scripts (bash, node, python), organized by category,
each carrying an embedded metadata block. A terminal UI in `tui/` discovers and
runs them.

## Layout

- `net/`, `git/`, `files/`, `dev/`, `media/` — scripts grouped by domain
- `templates/` — starter script showing the metadata convention
- `tui/` — the Go/Bubble Tea launcher

## Metadata convention

Each script starts with a comment block between `@meta` and `@end`. See
`templates/example.sh`.

## Running the TUI

```
cd tui && go run .
```
```

- [ ] **Step 7: Verify and commit**

Run:
```bash
cd /Users/johanrunesson/Development/private/scripts
ls net templates tui/go.mod
```
Expected: `net/portkill.sh`, `templates/example.sh`, and `tui/go.mod` all exist.

```bash
git add -A
git commit -m "chore: scaffold script folders, migrate portkill, init go module"
```

---

### Task 1: `meta` package — parse the metadata block

**Files:**
- Create: `tui/internal/meta/meta.go`
- Test: `tui/internal/meta/meta_test.go`

**Interfaces:**
- Produces:
  - `type Arg struct { Name string; Required bool; Help string }`
  - `type ScriptMeta struct { Name, Description, Category string; Args []Arg; Path string; Documented bool }`
  - `func Parse(content []byte) (ScriptMeta, error)` — returns `ErrNoMetaBlock` when no `@meta` block is found within the first 20 lines; a YAML error on malformed content; otherwise the parsed meta with `Documented = true`.
  - `var ErrNoMetaBlock error`

- [ ] **Step 1: Write the failing test**

Create `tui/internal/meta/meta_test.go`:

```go
package meta

import (
	"errors"
	"testing"
)

func TestParseBashBlockWithArgs(t *testing.T) {
	content := []byte(`#!/bin/bash
# @meta
# name: portkill
# description: Kill the process on a TCP port
# category: net
# args:
#   - name: port
#     required: true
#     help: TCP port number
# @end

echo hi
`)
	m, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "portkill" || m.Category != "net" {
		t.Fatalf("bad meta: %+v", m)
	}
	if !m.Documented {
		t.Fatalf("expected Documented true")
	}
	if len(m.Args) != 1 || m.Args[0].Name != "port" || !m.Args[0].Required {
		t.Fatalf("bad args: %+v", m.Args)
	}
}

func TestParseNodeComments(t *testing.T) {
	content := []byte(`#!/usr/bin/env node
// @meta
// name: hello
// description: say hi
// category: dev
// @end
console.log("hi")
`)
	m, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "hello" || m.Category != "dev" {
		t.Fatalf("bad meta: %+v", m)
	}
	if len(m.Args) != 0 {
		t.Fatalf("expected no args, got %+v", m.Args)
	}
}

func TestParseNoBlock(t *testing.T) {
	_, err := Parse([]byte("#!/bin/bash\necho hi\n"))
	if !errors.Is(err, ErrNoMetaBlock) {
		t.Fatalf("expected ErrNoMetaBlock, got %v", err)
	}
}

func TestParseMalformedYAML(t *testing.T) {
	content := []byte(`# @meta
# name: bad
#   description: : : broken
# args: [unterminated
# @end
`)
	if _, err := Parse(content); err == nil {
		t.Fatalf("expected a parse error for malformed yaml")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd tui && go test ./internal/meta/`
Expected: FAIL — `undefined: Parse` / package does not compile.

- [ ] **Step 3: Write minimal implementation**

Create `tui/internal/meta/meta.go`:

```go
// Package meta parses the embedded @meta/@end metadata block from a script file.
package meta

import (
	"bufio"
	"bytes"
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

// Arg describes one positional argument a script accepts.
type Arg struct {
	Name     string `yaml:"name"`
	Required bool   `yaml:"required"`
	Help     string `yaml:"help"`
}

// ScriptMeta is the parsed metadata for a single script.
type ScriptMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`
	Args        []Arg  `yaml:"args"`

	Path       string `yaml:"-"` // resolved during scan
	Documented bool   `yaml:"-"` // false when no valid block was found
}

// ErrNoMetaBlock is returned when no @meta block is found in the scanned lines.
var ErrNoMetaBlock = errors.New("no metadata block found")

const maxScanLines = 20

// Parse extracts and parses the @meta/@end block from script content.
func Parse(content []byte) (ScriptMeta, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var (
		inBlock   bool
		yamlLines []string
		lineNum   int
	)
	for scanner.Scan() {
		lineNum++
		body := stripComment(scanner.Text())
		if !inBlock {
			if lineNum > maxScanLines {
				break
			}
			if strings.TrimSpace(body) == "@meta" {
				inBlock = true
			}
			continue
		}
		if strings.TrimSpace(body) == "@end" {
			var m ScriptMeta
			if err := yaml.Unmarshal([]byte(strings.Join(yamlLines, "\n")), &m); err != nil {
				return ScriptMeta{}, err
			}
			m.Documented = true
			return m, nil
		}
		yamlLines = append(yamlLines, body)
	}
	return ScriptMeta{}, ErrNoMetaBlock
}

// stripComment removes a leading # or // comment marker and one optional space,
// preserving the relative indentation of the remaining YAML content.
func stripComment(line string) string {
	s := strings.TrimLeft(line, " \t")
	switch {
	case strings.HasPrefix(s, "//"):
		s = s[2:]
	case strings.HasPrefix(s, "#"):
		s = s[1:]
	default:
		return line
	}
	return strings.TrimPrefix(s, " ")
}
```

- [ ] **Step 4: Add the dependency and run tests**

Run:
```bash
cd tui
go get gopkg.in/yaml.v3
go test ./internal/meta/
```
Expected: PASS (all four tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/johanrunesson/Development/private/scripts
git add tui/internal/meta tui/go.mod tui/go.sum
git commit -m "feat: add meta package to parse script metadata blocks"
```

---

### Task 2: `registry` package — discover scripts

**Files:**
- Create: `tui/internal/registry/registry.go`
- Test: `tui/internal/registry/registry_test.go`

**Interfaces:**
- Consumes: `meta.ScriptMeta`, `meta.Parse` from Task 1.
- Produces: `func Scan(root string) ([]meta.ScriptMeta, error)` — walks `root`, skipping `.git`, `tui`, `docs`, and `templates` directories; returns one `ScriptMeta` per `.sh`/`.py`/`.js` file with `Path` set. Undocumented files get `Name = filename`, `Documented = false`, and `Category` from the folder. Documented files with an empty category also fall back to the folder name.

- [ ] **Step 1: Write the failing test**

Create `tui/internal/registry/registry_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd tui && go test ./internal/registry/`
Expected: FAIL — `undefined: Scan`.

- [ ] **Step 3: Write minimal implementation**

Create `tui/internal/registry/registry.go`:

```go
// Package registry discovers scripts under a collection root.
package registry

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"scriptstui/internal/meta"
)

var scriptExts = map[string]bool{".sh": true, ".py": true, ".js": true}
var skipDirs = map[string]bool{".git": true, "tui": true, "docs": true, "templates": true}

// Scan walks root and returns metadata for every recognized script file.
func Scan(root string) ([]meta.ScriptMeta, error) {
	var out []meta.ScriptMeta
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && skipDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if !scriptExts[strings.ToLower(filepath.Ext(d.Name()))] {
			return nil
		}
		content, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		m, perr := meta.Parse(content)
		if perr != nil {
			m = meta.ScriptMeta{Name: d.Name()}
		}
		if m.Category == "" {
			m.Category = filepath.Base(filepath.Dir(path))
		}
		m.Path = path
		out = append(out, m)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd tui && go test ./internal/registry/`
Expected: PASS (both tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/johanrunesson/Development/private/scripts
git add tui/internal/registry
git commit -m "feat: add registry package to discover scripts"
```

---

### Task 3: `runner` package — execute a script and stream output

**Files:**
- Create: `tui/internal/runner/runner.go`
- Test: `tui/internal/runner/runner_test.go`

**Interfaces:**
- Produces:
  - `type Line struct { Text string; IsErr bool }`
  - `func Run(ctx context.Context, path string, args []string) (<-chan Line, <-chan int)` — streams stdout/stderr lines on the first channel (closed when the process ends), then sends the exit code on the second channel. A failure to start emits an error `Line` and exit code `1`.

- [ ] **Step 1: Write the failing test**

Create `tui/internal/runner/runner_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd tui && go test ./internal/runner/`
Expected: FAIL — `undefined: Run`.

- [ ] **Step 3: Write minimal implementation**

Create `tui/internal/runner/runner.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd tui && go test ./internal/runner/`
Expected: PASS (both tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/johanrunesson/Development/private/scripts
git add tui/internal/runner
git commit -m "feat: add runner package to execute and stream scripts"
```

---

### Task 4: TUI — list, detail pane, and root discovery

**Files:**
- Create: `tui/main.go`
- Create: `tui/items.go`
- Create: `tui/model.go`
- Test: `tui/items_test.go`

**Interfaces:**
- Consumes: `registry.Scan`, `meta.ScriptMeta`.
- Produces:
  - `func discoverRoot() string` (in `main.go`) — resolves the collection root from the `-root` flag, then `SCRIPTS_HOME` env, then the parent of the working directory's `tui` folder, defaulting to `..`.
  - `type scriptItem struct { m meta.ScriptMeta }` implementing `list.DefaultItem` (`Title() string`, `Description() string`, `FilterValue() string`).
  - `func newItems(metas []meta.ScriptMeta) []list.Item`.
  - `type model struct { ... }` with the Bubble Tea `Init/Update/View` methods; for this task it shows the list on the left and a detail pane on the right.

- [ ] **Step 1: Add TUI dependencies**

Run:
```bash
cd tui
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
```
Expected: dependencies added to `go.mod`/`go.sum` with no errors.

- [ ] **Step 2: Write the failing test for items**

Create `tui/items_test.go`:

```go
package main

import (
	"testing"

	"scriptstui/internal/meta"
)

func TestScriptItemFields(t *testing.T) {
	item := scriptItem{m: meta.ScriptMeta{
		Name:        "portkill",
		Description: "kill a port",
		Category:    "net",
	}}
	if item.Title() != "portkill" {
		t.Fatalf("Title = %q", item.Title())
	}
	if item.Description() != "kill a port" {
		t.Fatalf("Description = %q", item.Description())
	}
	if item.FilterValue() != "portkill net kill a port" {
		t.Fatalf("FilterValue = %q", item.FilterValue())
	}
}

func TestScriptItemUndocumented(t *testing.T) {
	item := scriptItem{m: meta.ScriptMeta{Name: "raw.py", Category: "dev"}}
	if item.Description() != "(undocumented)" {
		t.Fatalf("Description = %q", item.Description())
	}
}

func TestNewItemsCount(t *testing.T) {
	items := newItems([]meta.ScriptMeta{{Name: "a"}, {Name: "b"}})
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd tui && go test .`
Expected: FAIL — `undefined: scriptItem` / `undefined: newItems`.

- [ ] **Step 4: Implement items**

Create `tui/items.go`:

```go
package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"scriptstui/internal/meta"
)

// scriptItem adapts a ScriptMeta to the bubbles list.DefaultItem interface.
type scriptItem struct {
	m meta.ScriptMeta
}

func (i scriptItem) Title() string { return i.m.Name }

func (i scriptItem) Description() string {
	if i.m.Description == "" {
		return "(undocumented)"
	}
	return i.m.Description
}

func (i scriptItem) FilterValue() string {
	return strings.TrimSpace(i.m.Name + " " + i.m.Category + " " + i.m.Description)
}

func newItems(metas []meta.ScriptMeta) []list.Item {
	items := make([]list.Item, 0, len(metas))
	for _, m := range metas {
		items = append(items, scriptItem{m: m})
	}
	return items
}
```

- [ ] **Step 5: Implement the model (list + detail)**

Create `tui/model.go`:

```go
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
```

- [ ] **Step 6: Implement main + root discovery**

Create `tui/main.go`:

```go
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
```

- [ ] **Step 7: Run tests and build**

Run:
```bash
cd tui
go test .
go build ./...
```
Expected: tests PASS; build succeeds with no errors.

- [ ] **Step 8: Manual smoke test**

Run: `cd tui && go run .`
Expected: an alt-screen TUI listing `portkill` (and any other scripts). Arrow keys move the selection; the right pane updates to show description, category, path, and the `port` argument. `/` filters; `q` quits.

- [ ] **Step 9: Commit**

```bash
cd /Users/johanrunesson/Development/private/scripts
git add tui/main.go tui/items.go tui/model.go tui/items_test.go tui/go.mod tui/go.sum
git commit -m "feat: TUI list and detail pane with root discovery"
```

---

### Task 5: TUI — run the selected script and stream output

**Files:**
- Modify: `tui/model.go`
- Create: `tui/run.go`

**Interfaces:**
- Consumes: `runner.Run`, `runner.Line`.
- Produces (in `run.go`):
  - `type outputLineMsg runner.Line`
  - `type doneMsg struct { code int }`
  - `func startRun(path string, args []string) tea.Cmd` and the follow-up read commands that pump `runner`'s channels into Bubble Tea messages.
- Adds a `stateRunning` view to `model` showing streamed output and the final exit code; `esc` returns to the list.

- [ ] **Step 1: Implement the run commands**

Create `tui/run.go`:

```go
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

// startRun launches the script and returns a command that begins reading output.
func startRun(path string, args []string) tea.Cmd {
	lines, done := runner.Run(context.Background(), path, args)
	s := &runSession{lines: lines, done: done}
	return s.readNext()
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
```

Note: the model holds the `*runSession` so it can call `readNext()` again after each `outputLineMsg`.

- [ ] **Step 2: Extend the model with running state**

Modify `tui/model.go`. Add the state enum and fields, and update `newModel`, `Update`, and `View`.

Add near the top (after imports):

```go
type appState int

const (
	stateList appState = iota
	stateRunning
)
```

Change the `model` struct to:

```go
type model struct {
	list    list.Model
	width   int
	height  int
	state   appState
	session *runSession
	output  []string
	exited  bool
	code    int
}
```

In `Update`, replace the `tea.KeyMsg` handling and add the new message cases so the method reads:

```go
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
					m.state = stateRunning
					m.output = nil
					m.exited = false
					cmd := startRunFor(&m, sel.m.Path, nil)
					return m, cmd
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}
```

Add this helper to `model.go` so the model keeps a reference to the session:

```go
func startRunFor(m *model, path string, args []string) tea.Cmd {
	lines, done := runnerRun(path, args)
	m.session = &runSession{lines: lines, done: done}
	return m.session.readNext()
}
```

And in `run.go`, replace `startRun` with an exported-to-package seam so the model creates the session and holds it. Change `run.go` to expose:

```go
func runnerRun(path string, args []string) (<-chan runner.Line, <-chan int) {
	return runner.Run(context.Background(), path, args)
}
```

(Remove the now-unused `startRun` function from `run.go`; keep `runSession`, `readNext`, `outputLineMsg`, `doneMsg`.)

- [ ] **Step 3: Add the running view**

In `tui/model.go`, change `View` to dispatch on state:

```go
func (m model) View() string {
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
```

- [ ] **Step 4: Build and run tests**

Run:
```bash
cd tui
go test ./...
go build ./...
```
Expected: PASS and a clean build.

- [ ] **Step 5: Manual smoke test**

Run: `cd tui && go run .`
Steps: select `portkill`, press `enter`. Because portkill is run with no args here, it prints its "Must specify a port!" message and exits with code 1; the running view shows that output and `Exit code: 1 (failed)`. Press `esc` to return to the list.
(Full argument prompting is added in Task 6.)

- [ ] **Step 6: Commit**

```bash
cd /Users/johanrunesson/Development/private/scripts
git add tui/model.go tui/run.go
git commit -m "feat: run selected script and stream output in the TUI"
```

---

### Task 6: TUI — prompt for declared arguments before running

**Files:**
- Modify: `tui/model.go`
- Create: `tui/argform.go`
- Test: `tui/argform_test.go`

**Interfaces:**
- Consumes: `meta.Arg`.
- Produces (in `argform.go`):
  - `type argForm struct { args []meta.Arg; inputs []textinput.Model; focus int }`
  - `func newArgForm(args []meta.Arg) argForm`
  - `func (f argForm) values() []string`
  - `func missingRequired(args []meta.Arg, values []string) []string` — pure function returning the names of unfilled required args.
- Adds a `stateArgs` state to `model`: pressing `enter` on a script with args opens the form; submitting a valid form transitions to `stateRunning` with the collected values; a script with no args runs immediately (Task 5 behavior).

- [ ] **Step 1: Write the failing test for validation**

Create `tui/argform_test.go`:

```go
package main

import (
	"testing"

	"scriptstui/internal/meta"
)

func TestMissingRequired(t *testing.T) {
	args := []meta.Arg{
		{Name: "port", Required: true},
		{Name: "signal", Required: false},
	}
	missing := missingRequired(args, []string{"", "TERM"})
	if len(missing) != 1 || missing[0] != "port" {
		t.Fatalf("expected [port], got %v", missing)
	}

	none := missingRequired(args, []string{"8080", ""})
	if len(none) != 0 {
		t.Fatalf("expected no missing, got %v", none)
	}
}

func TestArgFormValues(t *testing.T) {
	f := newArgForm([]meta.Arg{{Name: "port"}, {Name: "signal"}})
	if len(f.values()) != 2 {
		t.Fatalf("expected 2 values, got %d", len(f.values()))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd tui && go test .`
Expected: FAIL — `undefined: missingRequired` / `undefined: newArgForm`.

- [ ] **Step 3: Add the textinput dependency**

Run: `cd tui && go get github.com/charmbracelet/bubbles/textinput`
Expected: dependency resolved (already present via bubbles; command is a no-op or updates go.sum).

- [ ] **Step 4: Implement the arg form**

Create `tui/argform.go`:

```go
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
```

- [ ] **Step 5: Wire the form into the model**

Modify `tui/model.go`:

Add `stateArgs` to the state enum:

```go
const (
	stateList appState = iota
	stateArgs
	stateRunning
)
```

Add fields to `model`:

```go
	form    argForm
	pending meta.ScriptMeta // script awaiting args
	formErr string
```

(Ensure `meta` is imported in `model.go`: `"scriptstui/internal/meta"`.)

Replace the `"enter"` case in the `stateList` key handling with:

```go
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
```

Add a `stateArgs` branch at the top of the `tea.KeyMsg` handling (before the `stateRunning` check):

```go
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
```

Add a `stateArgs` branch to `View`:

```go
	if m.state == stateArgs {
		v := m.form.view()
		if m.formErr != "" {
			v += "\n" + m.formErr + "\n"
		}
		return v
	}
```

- [ ] **Step 6: Run tests and build**

Run:
```bash
cd tui
go test ./...
go build ./...
```
Expected: PASS and clean build.

- [ ] **Step 7: Manual smoke test**

Run: `cd tui && go run .`
Steps: select `portkill`, press `enter` → the arg form appears prompting for `port`. Press `enter` with it empty → see `Required: port`. Type an unused high port (e.g. `59999`), press `enter` → the script runs and reports the port (nothing listening, so kill is a no-op) with an exit code. Press `esc` to return.

- [ ] **Step 8: Commit**

```bash
cd /Users/johanrunesson/Development/private/scripts
git add tui/model.go tui/argform.go tui/argform_test.go tui/go.mod tui/go.sum
git commit -m "feat: prompt for declared arguments before running scripts"
```

---

### Task 7: Documentation polish and final verification

**Files:**
- Modify: `README.md`
- Modify: `tui/README.md` (create)

**Interfaces:** none (documentation only).

- [ ] **Step 1: Add a TUI README**

Create `tui/README.md`:

```markdown
# scripts TUI

A Bubble Tea launcher that discovers scripts in the parent collection and runs them.

## Run

```
go run .            # scans the parent directory (..)
go run . -root /path/to/scripts
SCRIPTS_HOME=/path/to/scripts go run .
```

## Build a binary

```
go build -o scripts-tui .
./scripts-tui -root ..
```

## Keys

- `↑/↓` move, `/` filter, `enter` run, `q` quit
- In the argument form: `tab`/`↑↓` move between fields, `enter` submit, `esc` cancel
- In the output view: `esc` returns to the list once the script has finished
```

- [ ] **Step 2: Full test + vet + build pass**

Run:
```bash
cd tui
go vet ./...
go test ./...
go build ./...
```
Expected: no vet warnings, all tests PASS, clean build.

- [ ] **Step 3: End-to-end manual check**

Run: `cd tui && go run .`
Confirm: `portkill` and `example` (from `templates/`… note: templates is skipped by the scanner, so `example` will NOT appear — confirm only real category scripts appear). Selecting `portkill`, entering a port, and running works end to end.

- [ ] **Step 4: Commit**

```bash
cd /Users/johanrunesson/Development/private/scripts
git add README.md tui/README.md
git commit -m "docs: add TUI usage README"
```
