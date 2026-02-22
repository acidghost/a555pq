# Agent Guidelines for a555pq

This file provides essential guidelines for agentic coding tools working in this repository.

## Build and Development Commands

Use Justfile for all operations (run `just help` for complete list):
- `just build` / `just b` - Build binary for current platform
- `just run <args>` - Build and execute with arguments
- `just fmt lint` - Format and lint code
- `just test` - Test code
- `just clean` - Remove build directory
- if a common opeation is not in the Justfile, consider adding it

## Write Tests

Use tests to solidify what you want to build before starting to build new features.

## Adding New Commands

1. Create new package in `cmd/` (e.g., `cmd/npm/`)
2. Implement `internal/<package>/client.go` and `types.go`
3. Add command structs in `cmd/<package>/cmd.go`
4. Register in `cmd/root.go` init()
5. Add output types to `internal/formatter/output.go`
6. Implement formatter methods for new output types
