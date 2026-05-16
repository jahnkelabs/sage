# Releasing sage

Releases are automated with **[semantic-release](https://semantic-release.gitbook.io/)** on every push to **`main`** (see [`.github/workflows/release.yml`](.github/workflows/release.yml)).

## Version bumps

Configured in [`release.config.cjs`](release.config.cjs) (`@semantic-release/commit-analyzer` **releaseRules**):

| Commit type | Effect |
|-------------|--------|
| `feat` | **Minor** semver bump |
| `fix` | **Patch** semver bump |
| `chore` | **No release** (no tag, no GitHub Release) |

Use **`feat`** or **`fix`** when you want Homebrew / `go install …@latest` consumers to pick up a build.

If **`main`** only receives **`chore`** commits since the last tag, semantic-release does nothing—expected.

## GitHub Releases and binaries

1. **semantic-release** creates **`vX.Y.Z`** and a GitHub Release with notes.
2. **[GoReleaser](https://goreleaser.com/)** runs on **`push` tags `v*`** ([`.github/workflows/goreleaser.yml`](.github/workflows/goreleaser.yml)), attaches archives/checksums to that release, and updates the Homebrew formula.

GoReleaser must attach artifacts to the **existing** release for the tag (default behavior when the release already exists).

## Homebrew tap

Formula repo: **`github.com/devenjahnke/homebrew-tap`**.

Users install with:

```bash
brew tap devenjahnke/tap
brew install sage
brew upgrade sage
```

### One-time setup

1. Create the GitHub repo **`devenjahnke/homebrew-tap`** (see [`packaging/homebrew-tap/README.md`](packaging/homebrew-tap/README.md) for suggested README contents).
2. In **`devenjahnke/sage`** GitHub settings → **Secrets and variables → Actions**, add **`TAP_GITHUB_TOKEN`**:

   - Fine-grained PAT: **Contents: Read and write** on **`homebrew-tap`** only (and metadata read on **`sage`** if needed for cross-repo workflows—or use classic PAT with `repo` scope on both).

The default **`GITHUB_TOKEN`** in Actions **cannot** push to another repository; GoReleaser needs **`TAP_GITHUB_TOKEN`** to commit `Formula/sage.rb`.

If **`main`** uses **branch protection**, allow **`GITHUB_TOKEN`/Actions** (or an appropriate bypass actor) to **push tags** and publish releases—or semantic-release will fail when creating **`v*`** tags.

### First formula

The first successful tagged release publishes **`Formula/sage.rb`** automatically. Until then, `brew install` will fail—document that for early adopters if needed.

## Local checks

```bash
docker run --rm -v "$PWD:/src" -w /src goreleaser/goreleaser:v2.5.0 release --snapshot --clean --skip=publish
```
