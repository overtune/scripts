# Follow-ups

Non-blocking follow-ups deferred during the initial build of the scripts collection + TUI.
Ordered roughly by value. Each notes the concrete files/locations to touch.

Design spec: `docs/superpowers/specs/2026-07-03-scripts-tui-design.md`
Implementation plan: `docs/superpowers/plans/2026-07-03-scripts-tui.md`

## 1. Wire runner cancellation into the TUI (highest value)

The `runner` package already supports context cancellation (streaming and start-failure
sends both honor `ctx.Done()`, and it's race-tested). But the TUI doesn't use it:
`runnerRun` in `tui/run.go` calls `runner.Run(context.Background(), ...)`, so there is no
way to abort a running/hung script — `q`/`ctrl+c` are swallowed while running
(`tui/model.go`, `stateRunning` key handling), and Ctrl+C at the terminal orphans the
subprocess (Bubble Tea quits but the child keeps running).

Fix sketch:
- Thread a real `context.Context` + `cancel` through `runSession` (store `cancel` on the
  session in `tui/run.go`).
- Add an abort key while running (e.g. `esc`/`ctrl+c` before exit) that calls `cancel()`.
- Call `cancel()` on program teardown so quitting never orphans a child.

## 2. Detail pane height clamp / scroll

`model.height` is stored on the model but never consulted in `View()`/`detailView()`
(`tui/model.go`). The detail pane's source preview has no height constraint, so a long
preview can overflow a short terminal with no scroll. Consider a `viewport` (bubbles) or
truncating to the available height.

## 3. Make `discoverRoot()` unit-testable

`discoverRoot()` in `tui/main.go` registers a `-root` flag on the global
`flag.CommandLine` and calls `flag.Parse()` inline, so it can't be unit-tested (a second
call panics with "flag redefined"). Refactor to accept args / use a dedicated `flag.FlagSet`
so the precedence logic (`-root` > `SCRIPTS_HOME` > `..`) can be tested.

## 4. `Esc` quits while browsing (undocumented)

While browsing the list (not filtering), `Esc` falls through to bubbles' default keymap
and quits the program (`tui/model.go` only intercepts `q`/`ctrl+c`). Either document it or
rebind so `Esc` doesn't quit from the list.

## 5. Close registry test-coverage gaps

`tui/internal/registry/registry_test.go` doesn't directly assert several explicit
constraints from the spec (the code is correct by inspection, but untested):
- skip-list covers `.git` and `templates` (only `tui`/`docs` are currently exercised)
- case-insensitive extensions (`.SH`, `.PY`, `.JS`)
- the root directory itself is never skipped even if its basename matches a skip entry
- a *documented* script with an empty `category:` falls back to the folder name
