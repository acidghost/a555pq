# a555pq

A CLI tool to query package information from various package managers like PyPI, npm, and container registries.

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

- `a555pq container <command> <image>` - Query container registries
- `a555pq github <command> <owner/repo>` - Query GitHub repositories
- `a555pq npm <command> <package>` - Query npm
- `a555pq pypi <command> <package>` - Query PyPI

Available commands:

- `show <package>` - Display detailed package information (supports `--raw` flag for full API response)
- `versions <package>` - List all versions with upload dates
- `latest <package>` - Show only the latest version
- `browse <package>` - Open package page in browser
- `version` - Display build metadata

### GitHub Authentication

The GitHub commands support two authentication methods:

1. **GITHUB_TOKEN environment variable**: Set this to use your GitHub token
2. **gh CLI**: If `gh` CLI is installed and authenticated, the token is automatically retrieved

Authentication enables:

- GraphQL API with proper tag dates in versions command
- Higher rate limits
- Access to private repositories

To force REST mode (for unauthenticated requests), use the `--rest` flag with the `versions` command:

```bash
a555pq github versions owner/repo --rest
```

### Container Registry Support

The container command supports multiple public registries:

**Supported Registries:**

- **Docker Hub** - Default registry (e.g., `nginx`, `library/nginx`)
- **GitHub Container Registry** - `ghcr.io/owner/image`
- **Google Container Registry** - `gcr.io/project/image`
- **Azure Container Registry** - `registry.azurecr.io/image`
- **Amazon ECR Public** - `public.ecr.aws/alias/image`
- **Quay.io** - `quay.io/organization/image`

### Output Formats

All commands support JSON output: `-o json` or `--output json`

Example:

```bash
a555pq container show nginx --output json
a555pq github show facebook/react
a555pq github versions golang/go
a555pq npm versions express --output json
a555pq pypi show requests -o json
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
just run container show nginx
just run github show facebook/react
just run npm show express
just run pypi show requests
```

### Other Commands

- Format code: `just fmt`
- Run linter: `just lint`
- Vendor dependencies: `just vendor` (runs `go mod tidy` and `go mod vendor`)
- Clean build artifacts: `just clean`
- Install binary to GOPATH/bin: `just install`
- List all commands: `just help`
