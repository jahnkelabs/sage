# `jahnkelabs/homebrew-tap`

Organization **[Homebrew tap](https://docs.brew.sh/Taps)** for formulas maintained here—often updated automatically by upstream release tooling.

## Consumers

### sage

Published from **[sage releases](https://github.com/jahnkelabs/sage/releases)**:

```bash
brew tap jahnkelabs/tap
brew install sage
brew upgrade sage
```

Because the repo name is **`homebrew-tap`**, Homebrew exposes it as **`jahnkelabs/tap`**.

## Publishers (generic pattern)

Useful when another project wants its own formula committed to **this** tap from CI.

### Shared GitHub App: **HomebrewTap**

The **`jahnkelabs`** organization already runs a GitHub App named **`HomebrewTap`** that is installed with access **only** to **`jahnkelabs/homebrew-tap`** (least privilege). Upstream repos do **not** create a separate app—they reuse this installation and mint short-lived tokens in Actions.

Org-wide credentials for Actions:

| Kind | Name | Purpose |
|------|------|--------|
| **Organization variable** | **`HOMEBREW_TAP_GH_APP_ID`** | Numeric GitHub App ID for **HomebrewTap** |
| **Organization secret** | **`HOMEBREW_TAP_GH_APP_SECRET_KEY`** | PEM private key for **HomebrewTap** (`BEGIN … END` block) |

**Repository access:** For each variable and secret, org admins set visibility under **Organization → Settings → Secrets and variables → Actions** so **every repository that publishes formulas here** can read them—typically **Selected repositories** (add each upstream repo) or **All repositories**, depending on policy. Without access, workflows see empty values and token minting fails.

### Tap layout

- **`Formula/<formula>.rb`** — Homebrew formula files (committed manually for bootstrap or by automation).

### GoReleaser

Configure **[brews](https://goreleaser.com/customization/homebrew-taps/)** with:

- **`repository.owner`** / **`repository.name`** pointing at **`jahnkelabs/homebrew-tap`** (and **`repository.branch`** if not **`main`**).
- **`repository.token`** set from CI (installation token below)—not **`GITHUB_TOKEN`** from the upstream repo alone.

### Minting a tap token in Actions

Use **`actions/create-github-app-token`** with the shared org credentials, then pass **`steps.<id>.outputs.token`** into GoReleaser as **`TAP_GITHUB_TOKEN`** (or whatever env **`brews[].repository.token`** reads):

```yaml
- uses: actions/create-github-app-token@v2
  id: tap-token
  with:
    app-id: ${{ vars.HOMEBREW_TAP_GH_APP_ID }}
    private-key: ${{ secrets.HOMEBREW_TAP_GH_APP_SECRET_KEY }}
    owner: jahnkelabs
    repositories: |
      homebrew-tap
    permission-contents: write
```

### Sage-specific notes

See **[`RELEASING.md`](https://github.com/jahnkelabs/sage/blob/main/RELEASING.md)** on **`jahnkelabs/sage`** for how **sage** wires GoReleaser on tag pushes.

### Alternatives

Fine-grained or classic PATs can push here but rotate poorly compared with installation tokens—prefer the shared app flow above.

### Branch protection

If **`main`** is protected, ensure automation can push formula commits (rules bypass or merge queues—whatever matches your org policy).

---

Most **`Formula/*.rb`** updates for **sage** come from **GoReleaser** when a **`v*`** tag is pushed on **`github.com/jahnkelabs/sage`**; manual edits may be overwritten by the next release.
