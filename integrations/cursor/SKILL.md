---
name: sage-compose
description: Use when working with the sage CLI for Docker Compose label-based aliases (compose.alias.*), sage --dry-run, or sage aliases --json.
---

# sage (Docker Compose aliases)

`sage` wraps Docker Compose: familiar verbs pass through unchanged; shortcuts come from compose **service labels** `compose.alias.<name>`.

## Discovery (do this first in a repo)

- Run `sage --help` for globals plus an **Aliases** table (when compose config loads).
- Run `sage aliases --json` for a stable machine-readable list (`alias`, `service`, `command`).

If Docker / compose config is unavailable, help still prints but aliases show as unavailable—do not invent alias names.

## How routing works

- First positional token:
  - Known compose verbs (`up`, `down`, `ps`, `logs`, `run`, …) → forwarded to `docker compose`.
  - Else if it matches a discovered alias → `docker compose run --rm` (+ `-it` unless `SAGE_NO_TTY`) for the labeled service with the label command argv (after env substitution).
  - Else → forwarded to compose as-is (compose may error).

## Safe inspection

- Prefer `sage --dry-run <verb-or-alias> …` to print the exact downstream argv before suggesting destructive compose commands.

## Env substitution

Label commands expand `$VAR` / `${VAR}` from the **host** environment before splitting into argv; missing vars warn and become empty strings.

## Flags

- **`--file`**: repeatable compose files (no short `-f` so `logs -f` keeps working).
- **`--project-name` / `-p`**: compose project name.
- **`SAGE_COMPOSE_CONFIG_DUMP`**: optional path to merged compose YAML for environments without a working compose daemon (tests / CI).

## Collision note

Some systems already ship a `sage` binary (e.g. SageMath). Use `which sage` if commands behave unexpectedly.
