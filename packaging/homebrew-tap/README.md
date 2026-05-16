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

Useful when another CLI repo wants its own formula on **this** tap.

### Tap layout

- **`Formula/<formula>.rb`** — Homebrew formula files (committed manually for bootstrap or by automation).

### GoReleaser

Configure **[brews](https://goreleaser.com/customization/homebrew-taps/)** with:

- **`repository.owner`** / **`repository.name`** pointing at **`jahnkelabs/homebrew-tap`** (and **`repository.branch`** if not **`main`**).
- **`repository.token`** set from CI (installation token below)—not **`GITHUB_TOKEN`** from the upstream repo alone.

### Cross-repo authentication (recommended)

1. Create a **GitHub App** with **Contents: Read and write** on **this tap repo** only; install it on **`jahnkelabs`** with access **only** to **`homebrew-tap`**.
2. Store **`GH_APP_ID`** and **`GH_APP_PRIVATE_KEY`** as **organization secrets** on **`jahnkelabs`**, with **repository access** granted to the **upstream** repo that runs CI (for sage: **`jahnkelabs/sage`**).
3. In that upstream workflow, use **`actions/create-github-app-token`** with **`owner: jahnkelabs`** and **`repositories: homebrew-tap`**, then pass **`steps.<id>.outputs.token`** into GoReleaser as **`TAP_GITHUB_TOKEN`** (or whatever env your **`brews[].repository.token`** reads).

See **[`RELEASING.md`](https://github.com/jahnkelabs/sage/blob/main/RELEASING.md)** on **`jahnkelabs/sage`** for the concrete sage checklist.

### Alternatives

Fine-grained or classic PATs can push here but rotate poorly compared with installation tokens—prefer a GitHub App.

### Branch protection

If **`main`** is protected, ensure automation can push formula commits (rules bypass or merge queues—whatever matches your org policy).

---

Most **`Formula/*.rb`** updates for **sage** come from **GoReleaser** when a **`v*`** tag is pushed on **`github.com/jahnkelabs/sage`**; manual edits may be overwritten by the next release.
