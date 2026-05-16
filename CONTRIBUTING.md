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

```bash
go test ./...
go build -o sage .
```
