# scripts

A collection of reusable scripts (bash, node, python), organized by category,
each carrying an embedded metadata block. A terminal UI in `tui/` discovers and
runs them.

## Layout

- `net/`, `git/`, `files/`, `dev/`, `media/`, `sys/` — scripts grouped by domain
- `templates/` — starter script showing the metadata convention
- `tui/` — the Go/Bubble Tea launcher

## Metadata convention

Each script starts with a comment block between `@meta` and `@end`. See
`templates/example.sh`.

## Running the TUI

```
cd tui && go run .
```
