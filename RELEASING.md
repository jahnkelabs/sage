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

GoReleaser must attach artifacts to the **existing** GitHub Release for that tag (**`release --clean`**). semantic-release creates that release as a **draft** ([`release.config.cjs`](release.config.cjs) `draftRelease: true`) so GoReleaser can upload assets before the release is published.

## Recovery

### semantic-release OK, GoReleaser failed

Typical symptom: tag and GitHub Release exist, but the Release workflow failed in **`goreleaser`** (e.g. `422 Cannot upload assets to an immutable release`).

1. **Do not delete the git tag** while a published release still references it.
2. **Re-run** via **Actions → Release → Run workflow** and select the existing **`refs/tags/vX.Y.Z`** ref (GoReleaser-only). Do not re-run on **`main`** unless you intend a new semver bump.
3. If the release is published/immutable and uploads still fail, delete the **GitHub Release** (not necessarily the tag), fix config, and dispatch on the tag again.

### Skipping a burned tag (e.g. v1.0.3)

If a tag name cannot be recreated (GH013 / [immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases) after a published release was deleted), ship the same commit under the next patch:

```bash
git tag v1.0.4 <commit-sha>
git push origin v1.0.4
gh workflow run release.yml --repo jahnkelabs/sage --ref v1.0.4
```

### Broken state: release without tag (or vice versa)

Avoid re-running **`main`** while a release exists for a version whose git tag was deleted—that often fails on `git push --tags`. Delete stray draft releases, restore or bump the tag, then use **tag dispatch** as above.

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
