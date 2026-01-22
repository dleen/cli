# Netflix gh CLI

A patched version of the GitHub CLI that works automatically with Netflix's git proxy setup.

## Why?

Netflix uses `git.netflix.net` as a git proxy, but the GitHub Enterprise API is at `github.netflix.net`. The standard `gh` CLI doesn't handle this - it tries to make API calls to `git.netflix.net` which fails.

This patched version automatically detects Netflix git proxy remotes and routes API calls to the correct host.

## Install

### 1. Uninstall brew version (if installed)

```bash
brew uninstall gh
```

### 2. Download the patched binary

**Option A: Using existing gh (before uninstalling brew version)**

```bash
# Download using gh (requires GH_HOST since you're not in a Netflix repo)
GH_HOST=github.netflix.net gh release download v2.83.2-proxy-fix --repo dleen/cli --pattern "*.tar.gz" --dir /tmp
tar -xzf /tmp/gh-darwin-arm64.tar.gz -C /tmp
mv /tmp/gh-darwin-arm64 ~/.local/bin/gh
chmod +x ~/.local/bin/gh
```

**Option B: Download from web browser**

1. Go to: https://github.netflix.net/dleen/cli/releases/tag/v2.83.2-proxy-fix
2. Download `gh-darwin-arm64.tar.gz`
3. Extract and install:

```bash
tar -xzf ~/Downloads/gh-darwin-arm64.tar.gz -C /tmp
mv /tmp/gh-darwin-arm64 ~/.local/bin/gh
chmod +x ~/.local/bin/gh
```

### 3. Ensure ~/.local/bin is in your PATH

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then reload:

```bash
source ~/.zshrc  # or source ~/.bashrc
```

### 4. Authenticate to Netflix GHE

```bash
gh auth login -p https -h github.netflix.net
```

### 5. Verify

```bash
gh --version
# Should show: gh version 2.83.2-...

cd your-netflix-repo
gh pr list
# Should work without setting GH_HOST!
```

## Usage

Just use `gh` normally in any Netflix repo:

```bash
gh pr list
gh pr create --title "My PR" --body "Description"
gh pr view 123
gh pr checkout 123
gh repo view
```

No need to set `GH_HOST` or run `ghe-fix-proxy`!

## How it works

When the CLI detects a git remote pointing to `git.netflix.net`, it automatically maps API calls to `github.netflix.net`. Git operations still use the original remote.

## Compatibility

- Works with Netflix GHE repos (git.netflix.net remotes)
- Works with github.com repos
- Works with other enterprise setups via `GH_HOST` environment variable

## Source

- Branch: https://github.netflix.net/dleen/cli/tree/fix-git-proxy-remotes
- Based on upstream PR: https://github.com/cli/cli/pull/12180
