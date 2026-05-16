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

### One-time setup: GitHub App + organization secrets

Cross-repo formula commits use a **GitHub App installation token** minted in Actions ([`actions/create-github-app-token`](https://github.com/actions/create-github-app-token)), not the default **`GITHUB_TOKEN`**, and not a long-lived PAT.

1. **Create a GitHub App** (recommended: owned by **`jahnkelabs`** org; follow org policy).
2. Grant the app repository permission **Contents: Read and write** (and **Metadata** read-only as required by GitHub).
3. **Install** the app on **`jahnkelabs`** and allow it to access **`homebrew-tap`** only (least privilege).
4. Note the **App ID** and generate a **private key** (PEM) from the App settings.
5. In the org: **Settings → Secrets and variables → Actions** (organization level), create:
   - **`GH_APP_ID`** — the numeric App ID (stored as an organization secret).
   - **`GH_APP_PRIVATE_KEY`** — full PEM contents.

   Under **Repository access** for each secret, choose **Selected repositories** and include **`jahnkelabs/sage`** so workflow runs there can read them.

   Org owners can also use the CLI (repo slug is **`sage`** within **`jahnkelabs`**):

   ```bash
   gh secret set GH_APP_ID --org jahnkelabs --repos sage --body "$GH_APP_ID"
   gh secret set GH_APP_PRIVATE_KEY --org jahnkelabs --repos sage < path/to/app-private-key.pem
   ```

The GoReleaser workflow passes the minted token to GoReleaser as **`TAP_GITHUB_TOKEN`** (see [`.goreleaser.yml`](.goreleaser.yml)).

If **`main`** uses **branch protection**, allow **`GITHUB_TOKEN`/Actions** (or an appropriate bypass actor) to **push tags** and publish releases—or semantic-release will fail when creating **`v*`** tags.

### Legacy PAT (discouraged)

If you temporarily used a **`TAP_GITHUB_TOKEN`** repository secret (fine-grained or classic PAT), remove it once the GitHub App path works—the workflow no longer reads **`secrets.TAP_GITHUB_TOKEN`**.

### First formula

The first successful tagged release publishes **`Formula/sage.rb`** automatically. Until then, `brew install` will fail—document that for early adopters if needed.

## Local checks

```bash
docker run --rm -v "$PWD:/src" -w /src goreleaser/goreleaser:v2.5.0 release --snapshot --clean --skip=publish
```
