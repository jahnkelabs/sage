# `devenjahnke/homebrew-tap`

Homebrew formulas published automatically from **[sage releases](https://github.com/devenjahnke/sage/releases)**.

## Install sage

```bash
brew tap devenjahnke/tap
brew install sage
brew upgrade sage
```

This repository is normally updated by **GoReleaser** when a new **`v*`** tag is pushed on `devenjahnke/sage`. Manual edits to `Formula/sage.rb` may be overwritten by the next release.

## Maintainer bootstrap

Create this GitHub repo **`homebrew-tap`** under user/org **`devenjahnke`**, then add the **`TAP_GITHUB_TOKEN`** secret to the **`sage`** repo as described in [`RELEASING.md`](https://github.com/devenjahnke/sage/blob/main/RELEASING.md) on the sage project.
