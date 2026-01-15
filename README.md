# a555pq

A CLI tool to query package information from various package managers like PyPI, npm, and more.

## Installation

```bash
go install github.com/acidghost/a555pq@latest
```

Or build from source:

```bash
git clone https://github.com/acidghost/a555pq.git
cd a555pq
just build
```

## Usage

### Commands

All package index commands follow the same pattern:

- `a555pq pypi <command> <package>` - Query PyPI
- `a555pq npm <command> <package>` - Query npm

Available commands:
- `show <package>` - Display detailed package information (supports `--raw` flag for full API response)
- `versions <package>` - List all versions with upload dates
- `latest <package>` - Show only the latest version
- `browse <package>` - Open package page in browser
- `version` - Display build metadata

### Output Formats

All commands support JSON output: `-o json` or `--output json`

Example:
```bash
a555pq pypi show requests -o json
a555pq npm versions express --output json
```

The `show` command supports `--raw` for complete API responses.

## Development

### Building

We recommend using [just](https://github.com/casey/just) for all build operations:

```bash
just build        # Build for current platform
just build-all    # Cross-compile for darwin-arm64, linux-arm64, linux-amd64
```

### Running

Build and execute with arguments:

```bash
just run pypi show requests
just run npm show express
```

### Other Commands

- Format code: `just fmt`
- Run linter: `just lint`
- Vendor dependencies: `just vendor` (runs `go mod tidy` and `go mod vendor`)
- Clean build artifacts: `just clean`
- Install binary to GOPATH/bin: `just install`
- List all commands: `just help`
