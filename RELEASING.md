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

1. **semantic-release** decides whether to publish **`vX.Y.Z`** / a GitHub Release; it is **no-op** when **`main`** has only **`chore`** since the last tag (idempotent **`semantic-release`** run).
2. **[GoReleaser](https://goreleaser.com/)** runs in the **`goreleaser`** job of **`Release`** ([`.github/workflows/release.yml`](.github/workflows/release.yml)) after **`resolve-release-tag`** picks a semver **`v*`** tag. Routine **`main`** merges use the tag semantic-release **just produced** (`GITHUB_TOKEN` tag pushes cannot start other workflows reliably—see **[chained workflows](https://docs.github.com/en/actions/using-workflows/triggering-a-workflow#triggering-a-workflow-from-a-workflow)**).
3. **Re-runs** (**Actions → Re-run workflow / jobs** on the same **`push`** run): when **`run_attempt`** is greater than **`1`**, GoReleaser runs only if a semver **`v*`** tag points at the workflow’s **`HEAD`** (recovery after **`goreleaser`** failed once **`semantic-release`** already tagged that commit; no-op chores without a release stay skipped).
4. **`workflow_dispatch`**: pick **`main`** (full **`semantic-release` + conditional GoReleaser**) or any **`refs/tags/v*`** tag (GoReleaser-only). GitHub still shows every branch in the picker, but **`assert-dispatch-ref`** fails fast if you choose anything besides **`main`** or a **`v*`** release tag—other branches/tags are unsupported.

GoReleaser must attach artifacts to the **existing** GitHub Release for that tag (**`release --clean`**).

## Homebrew tap

Formula repo: **`github.com/jahnkelabs/homebrew-tap`**.

Users install with:

```bash
brew tap jahnkelabs/tap
brew install sage
brew upgrade sage
```

### Organization migration (completed when these land on GitHub)

Repositories live under **`jahnkelabs`**: **`jahnkelabs/sage`** and **`jahnkelabs/homebrew-tap`**. After a transfer, update your local clone:

```bash
git remote set-url origin https://github.com/jahnkelabs/sage.git
```

### One-time setup: shared GitHub App **HomebrewTap**

Cross-repo formula commits use a **GitHub App installation token** minted in Actions ([`actions/create-github-app-token`](https://github.com/actions/create-github-app-token)), not the default **`GITHUB_TOKEN`**, and not a long-lived PAT.

The **`jahnkelabs`** org operates one GitHub App named **`HomebrewTap`**, installed on **`homebrew-tap`** only. Credentials are stored once at the organization:

| Kind | Name |
|------|------|
| Organization **variable** | **`HOMEBREW_TAP_GH_APP_ID`** |
| Organization **secret** | **`HOMEBREW_TAP_GH_APP_SECRET_KEY`** (full PEM private key) |

Org admins must grant **`jahnkelabs/sage`** (and any other publishing repo) **repository access** to both the variable and the secret under **Organization → Settings → Secrets and variables → Actions**, or widen visibility per policy.

Example CLI for **selected** repos (`repo` is the short name inside **`jahnkelabs`**):

```bash
gh variable set HOMEBREW_TAP_GH_APP_ID --org jahnkelabs --repos sage --body "$HOMEBREW_TAP_GH_APP_ID"
gh secret set HOMEBREW_TAP_GH_APP_SECRET_KEY --org jahnkelabs --repos sage < path/to/homebrew-tap-app.pem
```

See **[`packaging/homebrew-tap/README.md`](packaging/homebrew-tap/README.md)** for publisher-facing docs shared with **`github.com/jahnkelabs/homebrew-tap`**.

The **`Release`** job passes the minted tap token into GoReleaser as **`TAP_GITHUB_TOKEN`** (see [`.goreleaser.yml`](.goreleaser.yml)).

If **`main`** uses **branch protection**, allow **`GITHUB_TOKEN`/Actions** (or an appropriate bypass actor) to **push tags** and publish releases—or semantic-release will fail when creating **`v*`** tags.

### Legacy PAT (discouraged)

If you temporarily used a **`TAP_GITHUB_TOKEN`** repository secret (fine-grained or classic PAT), remove it once the GitHub App path works—the workflow no longer reads **`secrets.TAP_GITHUB_TOKEN`**.

### First formula

The first successful tagged release publishes **`Formula/sage.rb`** automatically. Until then, `brew install` will fail—document that for early adopters if needed.

## Local checks

```bash
docker run --rm -v "$PWD:/src" -w /src goreleaser/goreleaser:v2.5.0 release --snapshot --clean --skip=publish
```
