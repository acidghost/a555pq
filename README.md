# a555pq

[![CI](https://github.com/acidghost/a555pq/actions/workflows/ci.yaml/badge.svg)](https://github.com/acidghost/a555pq/actions/workflows/ci.yaml)

A CLI tool to query package information from various package managers like PyPI, npm, and container registries.

## Installation

```bash
go install github.com/acidghost/a555pq@latest
```

Versioned archives for Linux (`amd64`, `arm64`) and macOS (`arm64`) are also
available from [GitHub Releases](https://github.com/acidghost/a555pq/releases).
Each release includes SHA-256 checksums and signed build provenance.

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

### Filtering by Release Age

The `versions`, `latest`, and `show` commands of registry-backed ecosystems
(npm, PyPI, cargo, gem, maven, nuget, golang, and most others) accept a
`--min-release-age` flag that filters out versions released within the given
timespan. This is useful for avoiding freshly published releases.

```bash
# List chalk versions older than 1 year
a555pq npm versions chalk --min-release-age 1y

# Get the newest version of lodash that is at least 30 days old
a555pq npm latest lodash --min-release-age 30d
```

Supported units (case-insensitive): `y` (years), `mo` (months), `w` (weeks),
`d` (days), `h` (hours), `m` (minutes), `s` (seconds), `ms` (milliseconds).
Units can be combined, e.g. `1w2d` or `1d6h`. Years and months are
approximated as 365 and 30 days respectively.

Some ecosystems do not publish release timestamps (e.g. deno, terraform,
haxelib, julia, luarocks, nimble). For those the flag has no effect and all
versions are returned.

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
- Lint GitHub Actions: `just actions-lint`
- List all commands: `just help`

## Releases and supply-chain verification

A GitHub-verified signed annotated tag is the release source of truth. Tags must
use strict SemVer with a leading `v`; prereleases such as `v0.2.0-rc.1` are
supported. Merge the intended changes to `main`, wait for required CI, then tag
the exact commit:

```sh
git tag -s v0.1.0 -m 'a555pq v0.1.0'
git push origin v0.1.0
```

The release workflow rejects lightweight or unverified tags, tags not reachable
from `main`, malformed versions, and `v0.0.0`. It builds with an exact version,
commit, and commit-derived timestamp, runs tests, creates deterministic archives
and checksums, and signs GitHub artifact attestations before the `release`
environment can publish the immutable GitHub Release. Prerelease tags produce
GitHub prereleases.

Configure these repository controls before the first release:

1. Protect `main`, require pull requests and the CI check, and require CODEOWNERS
   review for workflow changes.
2. Create a tag ruleset for `v*` that restricts creation to release maintainers
   and blocks tag updates and deletion.
3. Create a `release` environment restricted to `v*` tags, add required
   reviewers, and prevent administrators from bypassing its protection rules
   where practical.
4. Enable immutable GitHub Releases in the repository settings.
5. Keep the repository's default workflow token permission read-only. The
   workflow grants write access only to its attestation and publication jobs.

Release tags and attached artifacts are immutable. If a run fails after an
artifact is published, inspect the commit and attestation and complete only the
missing publication; never move a release tag or replace an existing artifact.

After downloading an archive and `SHA256SUMS`, verify both its checksum and its
keyless GitHub Actions provenance:

```sh
VERSION=v0.1.0
ARCHIVE=a555pq_0.1.0_linux-amd64.tar.gz

gh release download "$VERSION" \
  --repo acidghost/a555pq \
  --pattern "$ARCHIVE" \
  --pattern SHA256SUMS
# Linux:
sha256sum --check --ignore-missing SHA256SUMS
# macOS:
shasum -a 256 --check --ignore-missing SHA256SUMS

gh attestation verify "$ARCHIVE" \
  --repo acidghost/a555pq \
  --signer-workflow acidghost/a555pq/.github/workflows/release.yaml \
  --source-ref "refs/tags/$VERSION" \
  --deny-self-hosted-runners
```

The `.intoto.jsonl` file attached to each release is the corresponding
attestation bundle for offline retention.
