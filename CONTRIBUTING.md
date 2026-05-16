# Contributing

## Pull requests

- Target **`main`**.
- Use **squash merge** when merging so each PR produces **one commit** on `main` with a clear conventional subject line.
- PR titles must follow **[Conventional Commits](https://www.conventionalcommits.org/)** with one of these types only:

  | Type | When |
  |------|------|
  | `feat` | User-visible behavior or new capability |
  | `fix` | Bug fix |
  | `chore` | Maintenance (docs-only, CI, refactors that don’t change behavior, etc.) |

  Examples: `feat: add sage aliases JSON flag`, `fix: handle empty compose profiles`, `chore: tighten CI timeouts`.

CI validates PR titles via [`.github/workflows/semantic-pr.yml`](.github/workflows/semantic-pr.yml).

## Releases (maintainers)

See [`RELEASING.md`](RELEASING.md). **`feat`** and **`fix`** bumps trigger semver tags and GitHub Releases; **`chore`** does **not** publish a release—bundle releasable work with a `feat` or `fix` when users need a new version.

## Development

See [`AGENTS.md`](AGENTS.md) for agent-oriented notes (including using the `dev` container when Go is not on the host).

With Go installed locally:

```bash
go test ./...
go build -o sage .
```

Without local Go (Docker only):

```bash
docker compose run --rm dev go test ./...
docker compose run --rm dev go build -o sage .
```

Or use the Makefile wrappers: `make test`, `make build`, `make tidy`, `make shell`.

## Testing

CLI behavior is covered by **contract tests** in [`cmd/cli_contract_test.go`](cmd/cli_contract_test.go): they build the real `sage` binary and invoke it via subprocess, asserting exit code and stdout/stderr—the same path users get from the shell. Other tests in `cmd/` cover pure helpers (alias YAML parsing, env substitution) and help output.

When you change compose passthrough, flag parsing, or dry-run output, add or extend a table row in `cli_contract_test.go` rather than calling internal functions like `forwardCompose` directly.
