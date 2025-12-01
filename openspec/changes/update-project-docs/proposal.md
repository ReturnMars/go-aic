# Change: Update Project Documentation

## Why
The current `README.md` references deprecated components (Bubble Tea) and outdated build scripts (`scripts/build.go`). It also lacks information about new features like the dedicated Chat Mode flag (`--chat`).

## What Changes
- Update `README.md` to reflect the switch to Native TUI + Glamour.
- Add documentation for the new `--chat` (`-c`) and `--quick` (`-q`) flags in `README.md`.
- Update build instructions in `README.md` to use standard `go build` or `goreleaser`.
- Clarify the AI configuration section in `README.md`.
- Update `AGENTS.md` to remove references to Bubble Tea and reflect the Native TUI architecture.
- Update State Machine diagram in `AGENTS.md` to match the new interaction flow.

## Impact
- Affected files: `README.md`, `AGENTS.md`
- No code changes, purely documentation.

