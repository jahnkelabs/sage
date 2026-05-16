# sage

`sage` is a small CLI around **Docker Compose** that forwards familiar compose verbs unchanged and adds **dynamic command aliases** declared as compose labels (`sage.alias.<name>`).

## Install

### Homebrew (recommended)

Requires the **[`jahnkelabs/homebrew-tap`](https://github.com/jahnkelabs/homebrew-tap)** repo (see bootstrap notes in [`RELEASING.md`](RELEASING.md) if it does not exist yet).

```bash
brew tap jahnkelabs/tap
brew install sage
brew upgrade sage
```

### Go toolchain

Requires Go **1.22+**:

```bash
go install github.com/jahnkelabs/sage@latest
```

Ensure **`$(go env GOPATH)/bin`** is on your `PATH`.

### GitHub Releases

Download a **`tar.gz`** for your OS/arch from the **[Releases](https://github.com/jahnkelabs/sage/releases)** page, unpack, and move `sage` onto your `PATH`.

### Build from source

```bash
git clone https://github.com/jahnkelabs/sage.git
cd sage
go build -o sage .
```

## Contributing & releases

See [`CONTRIBUTING.md`](CONTRIBUTING.md) (PR titles: **`feat`**, **`fix`**, **`chore`** only; squash merge). Maintainer release mechanics (semantic versioning, GitHub App / Homebrew tap): [`RELEASING.md`](RELEASING.md).

## Usage

```text
sage [flags] <compose verb | alias> [args...]
```

- **Passthrough**: `sage up -d`, `sage logs -f api`, etc. behave like `docker compose …` (or `docker-compose` when that binary is present).
- **Aliases**: labels on a service define extra “top-level” commands, e.g. `sage.alias.migrate: bundle exec rake db:migrate` → `sage migrate` runs `docker compose run --rm -it <service> bundle exec rake db:migrate` (profiles from the service are passed as `--profile` flags when present).
- **`--dry-run`**: prints the full `docker compose` / `docker-compose` argv that would run.
- **`--file`**: repeat to pass multiple compose files. A **short `-f` is intentionally not used** so flags like `logs -f` keep working as in stock compose.
- **`SAGE_NO_TTY`**: set to `1` / `true` to drop interactive TTY allocation (useful for CI and agents).

### Subcommands

- `sage aliases` — tabular list; `--json` for machines / agents.
- `sage completion <bash|zsh|fish>` — shell completion scripts.

### Agent / CI

For tests or sandboxes without a working Docker daemon, you can point sage at a **pre-merged** compose YAML (the shape `docker compose config` emits) with:

```bash
export SAGE_COMPOSE_CONFIG_DUMP=/absolute/path/to/merged-compose.yml
```

## Label schema

```yaml
services:
  api:
    profiles: [dev]
    labels:
      sage.alias.migrate: bundle exec rake db:migrate
      sage.alias.test: bundle exec rspec
```

Alias names must be unique across all services.

Environment variables in the label value expand as `$VAR` / `${VAR}` from the host environment before splitting into argv (missing variables warn and substitute empty strings).

## Name collision

`sage` may conflict with [SageMath](https://www.sagemath.org/) or other tools named `sage` on `PATH`. Check `which sage` after install.

## License

MIT — see [LICENSE](LICENSE).
