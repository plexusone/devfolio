# Getting Started

## Installation

```bash
go install github.com/plexusone/devfolio/cmd/devfolio@latest
```

Or build from source:

```bash
git clone https://github.com/plexusone/devfolio.git
cd devfolio
go build ./cmd/devfolio
```

## Authentication (contributor profile only)

`devfolio contributor profile` requires a GitHub personal access token as
`GITHUB_TOKEN`. `devfolio team velocity` and `devfolio devx dashboard` do
not need one — team velocity reads a portfolio file, and devx dashboard
reads the local OmniDevX event store.

```bash
export GITHUB_TOKEN=your_token_here
```

### Fine-grained token (recommended)

Create at: <https://github.com/settings/personal-access-tokens/new>

**Repository access:** select "Public repositories (read-only)", or
specific repos if you need private repo data.

**Repository permissions:**

| Permission | Access | Purpose |
|------------|--------|---------|
| Contents | Read-only | Read commit data |
| Pull requests | Read-only | Count PRs |
| Issues | Read-only | Count issues |
| Metadata | Read-only | Repository info (auto-included) |

**Account permissions:**

| Permission | Access | Purpose |
|------------|--------|---------|
| Profile | Read-only | User info (name, bio, etc.) |

### Classic token

Create at: <https://github.com/settings/tokens/new>

**Required scopes:**

| Scope | Purpose |
|-------|---------|
| `public_repo` | Access public repository data |
| `read:user` | Read user profile information |

Add `repo` scope instead of `public_repo` for private repository access.

## Generate a contributor profile

```bash
devfolio contributor profile --user grokify -o profile.json
```

The profile includes repository breakdown, language stats, an activity
heatmap, and AI collaboration metrics (which tools, how often, since when).
See [Contributor Commands](commands/contributor.md) for every flag.

## Generate a team velocity dashboard

Team velocity is built from a `structured-changelog` portfolio, not raw
GitHub data directly:

```bash
schangelog portfolio discover --org plexusone -o manifest.json
schangelog portfolio aggregate manifest.json -o portfolio.json
devfolio team velocity portfolio.json -o velocity.json
```

See [Team Commands](commands/team.md).

## Generate a DevX usage dashboard

This one is different: it doesn't call the GitHub API at all. It reads
events that [omnidevx-core](https://github.com/plexusone/omnidevx-core)'s
providers already collected into the local store
(`~/.plexusone/omnidevx/data/`), and it's the only command that produces a
[dashforge](https://github.com/plexusone/dashforge) dashboard by default
(contributor profile only does this with `--dashboard`):

```bash
devfolio devx dashboard --person person:jane --days 30 -o dashboard.json
```

Open the result in dashforge's static viewer, validate it with
`dashforge validate dashboard.json`, or serve it through
[VisionStudio](https://github.com/ProductBuildersHQ/visionstudio)'s DevX
panel by writing it to `~/.plexusone/omnidevx/dashboard.json`. See
[DevX Commands](commands/devx.md).
