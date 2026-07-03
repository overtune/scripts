# Scripts Collection + TUI Launcher — Design

**Date:** 2026-07-03
**Status:** Approved (pending spec review)

## Summary

Turn this repo into a well-organized home for reusable scripts (bash, node, python)
and build a terminal UI (TUI) that discovers, browses, and runs them. Each script
carries its own metadata so the TUI can present it and prompt for arguments before
running. The TUI is written in Go using the Bubble Tea framework, compiling to a
single static binary.

## Goals

- A clear, conventional directory structure for scripts, organized by domain.
- Every script is self-documenting via an embedded metadata block.
- A TUI that scans the collection, lists scripts grouped by category, supports
  fuzzy search, shows details, prompts for declared arguments, and runs the script
  with live-streamed output and a visible exit code.
- No registration step: adding a properly-formatted script to a folder makes it
  appear in the TUI automatically.

## Non-Goals (v1)

- Favorites, run history, tags beyond a single category.
- Editing/creating scripts from within the TUI.
- Remote script sources, sync, or sharing.

## Directory Structure

```
scripts/
├── net/            # networking scripts (portkill.sh lives here)
├── git/            # git helpers
├── files/          # filesystem utilities
├── dev/            # development helpers
├── media/          # media/file conversion, etc.
├── templates/      # starter script + metadata block to copy
├── tui/            # the Go TUI application (Go module)
├── docs/           # design docs and specs
└── README.md       # catalog (can be regenerated from metadata later)
```

- Categories are just folders. A script's category is declared in its metadata
  (which should match its folder); the folder is the physical home, the metadata
  is the source of truth the TUI reads.
- New categories = new folders. No central list to update.

## Script Metadata Convention

Each script begins with a structured comment block delimited by `@meta` / `@end`.
The TUI strips the leading comment prefix (`#` for bash/python, `//` for node) and
parses the remaining YAML.

Example (`net/portkill.sh`):

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
...
```

### Metadata fields

| Field         | Required | Description                                              |
|---------------|----------|----------------------------------------------------------|
| `name`        | yes      | Display name in the TUI                                   |
| `description` | yes      | One-line summary                                          |
| `category`    | yes      | Grouping key (should match containing folder)            |
| `args`        | no       | Ordered list of positional arguments                     |

Each entry in `args`:

| Field      | Required | Description                                      |
|------------|----------|--------------------------------------------------|
| `name`     | yes      | Argument name (shown as the prompt label)        |
| `required` | no       | Defaults to `false`                              |
| `help`     | no       | Hint text shown next to the prompt               |

### Parsing rules

- The block must appear within the first ~20 lines of the file.
- Lines between `@meta` and `@end` have their leading comment prefix and one
  optional following space stripped, then are parsed as YAML.
- A script with no valid metadata block is still listed, but shown as
  "undocumented" (name = filename, no description, no declared args). This keeps
  legacy scripts visible while nudging toward documentation.

## TUI Behavior

Framework: **Go + Bubble Tea** (with Lip Gloss for styling, Bubbles for list/input
components). Builds to a single static binary.

### Layout

- **Left pane:** scripts grouped by category, with a fuzzy-search/filter input.
- **Right pane:** detail view of the selected script — name, description, resolved
  file path, declared args, and a source preview.

### Flow

1. On launch, the TUI scans the collection root, walking category folders and
   parsing metadata from each script file.
2. User filters/selects a script.
3. On "run":
   - If the script declares `args`, show a form prompting for each in order.
     `required` args must be filled before proceeding.
   - Execute the script as a subprocess, passing the collected args positionally.
   - Stream stdout/stderr live into an output view.
   - On completion, show the exit code (success/failure styling).
4. User can return to the list and pick another script.

### Script execution

- Scripts are executed directly (they carry their own shebang and are marked
  executable), with collected arguments appended positionally.
- The collection root is discovered relative to the binary or via an env var /
  flag (e.g. `SCRIPTS_HOME`), defaulting to the repo root when run from within it.

## Components (for isolation/testability)

- **`meta` package:** parse a file's metadata block → `ScriptMeta` struct.
  Pure function over file contents; independently testable.
- **`registry` package:** walk the collection root, produce `[]ScriptMeta` with
  resolved paths. Depends on `meta`.
- **`runner` package:** given a `ScriptMeta` + arg values, spawn the process and
  expose a stream of output lines + final exit code. No UI dependency.
- **`tui` package (main):** Bubble Tea models wiring list, detail, arg form, and
  output views over `registry` + `runner`.

Each package has one clear purpose and a narrow interface, so the parser and runner
can be unit-tested without the TUI.

## Error Handling

- Malformed metadata block → script listed as "undocumented"; a parse warning is
  available in the detail pane rather than crashing the scan.
- Missing/inaccessible collection root → clear startup error with the resolved path.
- Non-executable or missing script at run time → surfaced in the output view, not a
  crash.
- Non-zero exit code → shown plainly with failure styling; not treated as a TUI
  error.

## Testing

- `meta`: table-driven tests over sample comment blocks (bash `#`, node `//`,
  valid, malformed, missing, args present/absent).
- `registry`: scan a fixture directory tree, assert discovered scripts + categories.
- `runner`: run small fixture scripts, assert streamed output and exit codes
  (including a deliberately failing script).
- TUI models: unit-test state transitions (filter, select, run, arg validation)
  where practical.

## Migration / First Deliverable

- Create the folder structure.
- Move `portkill.sh` → `net/portkill.sh` and add its metadata block.
- Add a `templates/` starter script demonstrating the metadata convention.
- Implement the Go TUI (`tui/`) covering: scan, grouped list, fuzzy search, detail
  pane, arg form, run + streamed output + exit code.
