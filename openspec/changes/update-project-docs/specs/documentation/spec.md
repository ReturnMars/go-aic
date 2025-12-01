## ADDED Requirements
### Requirement: Project README
The project MUST have a `README.md` file that accurately reflects the current features and usage.

#### Scenario: Build Instructions
- **WHEN** a developer reads the "Build" section
- **THEN** it SHOULD describe valid build commands (e.g., `go build`, `goreleaser`)
- **AND** it MUST NOT reference non-existent scripts (e.g., `scripts/build.go`)

#### Scenario: Feature Listing
- **WHEN** a user checks the feature list
- **THEN** it SHOULD mention Native TUI and Markdown support
- **AND** it SHOULD mention the Chat Mode capability

#### Scenario: Usage Flags
- **WHEN** a user looks for usage examples
- **THEN** it SHOULD document the `--chat` (`-c`) flag for direct chat entry
- **AND** it SHOULD document the `--quick` (`-q`) flag for skipping confirmations

### Requirement: Architecture Documentation (AGENTS.md)
The `AGENTS.md` file MUST accurately describe the system's technical architecture for future AI agents.

#### Scenario: TUI Architecture
- **WHEN** an AI agent reads the Architecture Guide
- **THEN** it SHOULD NOT see references to "Bubble Tea" as the core UI framework
- **AND** it SHOULD describe the UI as "Native TUI" or "Native Terminal" with "Glamour" for rendering

#### Scenario: State Machine
- **WHEN** an AI agent reviews the State Flow
- **THEN** it SHOULD reflect the Chat Mode loop and the simplified Commit flow


