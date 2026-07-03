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
