# Sage — agent instructions

Context for AI coding agents working in this repository.

## Go toolchain: use the `dev` container

This repo ships a **`dev`** Compose service ([`docker-compose.yml`](docker-compose.yml)) with **Go 1.22** (`golang:1.22-bookworm`). The project tree is mounted at `/src`.

**Prefer the `dev` container** when the host has no `go` on `PATH`, or when you need a consistent toolchain. Do not assume `go` is installed locally.

### Build and test

From the repository root:

```bash
docker compose run --rm dev go test ./...
docker compose run --rm dev go build -o sage .
docker compose run --rm dev go mod tidy
```

Makefile wrappers (same underlying command):

```bash
make test
make build
make tidy
make shell    # interactive shell in the dev container
```

### Contract tests

CLI behavior is validated via subprocess against a built binary in [`cmd/cli_contract_test.go`](cmd/cli_contract_test.go). Run the full suite with `make test` or `docker compose run --rm dev go test ./...`.

When changing passthrough verbs, flags, or `sage install`, extend contract tests rather than calling unexported helpers directly.

## `sage` CLI in this repo

- **Compose file**: `docker-compose.yml` (service name `dev`).
- **User-facing docs**: [`README.md`](README.md), [`CONTRIBUTING.md`](CONTRIBUTING.md).
- **Agent integrations**: `sage install --targets cursor` installs a global Cursor rule; source lives under [`integrations/cursor/rules/`](integrations/cursor/rules/).

For Docker Compose work in *other* projects, follow the installed `sage-compose` user rule (prefer `sage` over raw `docker compose`, discover aliases with `sage aliases --json`).
