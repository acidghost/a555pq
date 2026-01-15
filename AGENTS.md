# Agent Guidelines for a555pq

This file provides essential guidelines for agentic coding tools working in this repository.

## Build and Development Commands

Use Justfile for all operations (run `just help` for complete list):
- `just build` / `just b` - Build binary for current platform
- `just build-all` - Cross-compile (darwin-arm64, linux-arm64, linux-amd64)
- `just run <args>` / `just r <args>` - Build and execute with arguments
- `just fmt` - Format code using gofmt
- `just lint` - Run golangci-lint (must pass before commits)
- `just install` - Copy binary to GOPATH/bin
- `just clean` - Remove build directory

**Testing**: No test command in justfile - use standard Go:
- `go test ./...` - Run all tests
- `go test -v ./...` - Run tests with verbose output
- `go test -run TestName ./path/to/package` - Run single test
- `go test -cover ./...` - Run tests with coverage report

## Code Style Guidelines

### Imports and Formatting
- Standard library first, then internal packages, then external dependencies
- Group imports logically with blank lines between groups
- Always run `just fmt` before committing (uses gofmt)
- Never add comments unless explicitly requested
- Keep responses concise (fewer than 4 lines unless detail requested)

### Types and Enums
- Use typed enums for fixed sets of values (not strings) - see `cmd/shared/shared.go:3`
- Custom types for special JSON handling - see `internal/pypi/types.go:43` (Keywords)
- Interface for polymorphism - see `internal/formatter/output.go:10` (OutputFormatter)

### Naming Conventions
- Acronyms: JSON, URL, PyPI, API (all uppercase)
- Package names: lowercase single words (pypi, formatter, cmd, shared)
- Build-time variables: SCREAMING_SNAKE_CASE (buildVersion, buildCommit)
- Constants: PascalCase (Table, JSON)
- Functions: PascalCase for exported, camelCase for unexported
- Unused parameters: prefix with `_` (not used in cmd handlers)

### Error Handling
- Use `fmt.Errorf` with `%w` verb for error wrapping - see `internal/pypi/client.go:29`
- CLI commands return errors directly; Cobra handles display
- For non-fatal errors (e.g., browser launch): print warning to stderr and continue
- Don't use `os.Exit()` in command handlers - return error instead
- Always check error returns from functions like `client.GetPackageInfo()`

### CLI Commands (Cobra)
- Use `RunE` for commands that can fail (returns error)
- Use `ExactArgs(n)` for fixed argument counts
- Unused command parameters: `func(_ *cobra.Command, args []string)`
- Register subcommands in `init()` functions
- Set `Use`, `Short`, `Args`, and `RunE` fields
- Output formatters: Use factory pattern (see cmd/pypi/show.go:38)

### Linting Configuration
Linters enabled (see `.golangci.yml`):
- **copyloopvar** - Prevent loop variable shadowing
- **gosec** - Security checks (use `//nolint:gosec` sparingly)
- **nilerr** - Catch errors that are never checked
- **predeclared** - Avoid shadowing predeclared identifiers
- **revive** - Code quality (replaces golint)
- **tparallel** - Proper test parallelism
- **unconvert** - Remove unnecessary type conversions
- **unparam** - Detect unused function parameters
- Formatter: **gofmt**

Lint exclusions:
- `exhaustruct` is disabled (Cobra commands don't need full struct init)
- Go 1.25+ `copyloopvar` is enabled

### Important Patterns

**HTTP Client** - Use custom client with timeout:
```go
httpClient: &http.Client{Timeout: 30 * time.Second}
```

**Table Output** - Use tabwriter with specific parameters:
```go
tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
```

**Output Format** - Check enum value directly:
```go
if shared.OutputFormat == shared.JSON { ... }
```

**Formatter Creation** - Use factory methods:
```go
var f formatter.OutputFormatter
if shared.OutputFormat == shared.JSON {
    f = formatter.NewJSONFormatter()
} else {
    f = formatter.NewTableFormatter()
}
return f.Format(output)
```

**Build Metadata** - Always include version, commit, date in root command

### Security Notes
- Whitelist subprocess commands when using `exec.Command()` - see `cmd/pypi/browse.go`
- Use `//nolint:gosec` only after reviewing the security implication
- Never commit secrets or credentials

### Common Gotchas
- Don't use `var-naming` for enum constants (disabled in linter config)
- Build commands use git; if no git repo, fallback to 'unknown'
- Binary is ~8.7MB due to Go runtime
- Platform detection uses `runtime.GOOS` for OS-specific commands

### Adding New Commands
1. Create new package in `cmd/` (e.g., `cmd/npm/`)
2. Implement `internal/<package>/client.go` and `types.go`
3. Add command structs in `cmd/<package>/cmd.go`
4. Register in `cmd/root.go` init()
5. Add output types to `internal/formatter/output.go`
6. Implement formatter methods for new output types

### Project Structure
```
a555pq/
├── main.go                    # Entry point with build variables
├── justfile                   # Build commands
├── .golangci.yml              # Linting configuration
├── cmd/                       # CLI commands (Cobra)
│   ├── root.go               # Root command and version
│   ├── shared/               # Shared types and constants
│   └── pypi/                 # PyPI package commands
├── internal/                  # Internal packages (not exported)
│   ├── pypi/                 # PyPI client and types
│   └── formatter/            # Output formatters (table/JSON)
└── build/                     # Compiled binaries (gitignored)
```

### Module Information
- Module: `github.com/acidghost/a555pq`
- Go version: 1.25.3
- CLI framework: Cobra v1.10.2
