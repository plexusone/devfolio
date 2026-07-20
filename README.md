# DevFolio

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

Developer portfolio and team velocity metrics generator.

## Overview

DevFolio generates developer portfolios and team velocity dashboards from:

- 📋 Changelog data (via [structured-changelog](https://github.com/grokify/structured-changelog))
- 📜 Git history
- 🐙 GitHub activity (commits, PRs, issues, reviews)

## Use Cases

- 📊 **Team velocity dashboards** - Engineering managers track team output
- 👤 **Individual contributor portfolios** - Track your own contributions over time
- 🔍 **Recruiting (hiring)** - Evaluate candidate contribution patterns
- 💼 **Recruiting (job seeking)** - Showcase your development portfolio

## Installation

```bash
go install github.com/plexusone/devfolio/cmd/devfolio@latest
```

## Quick Start

### Team Velocity

```bash
# 1. Create a portfolio from changelogs (using structured-changelog)
schangelog portfolio discover --org plexusone -o manifest.json
schangelog portfolio aggregate manifest.json -o portfolio.json

# 2. Generate team velocity dashboard
devfolio team velocity portfolio.json -o velocity.json
```

### Individual Contributor Profile

```bash
# Generate contributor profile from GitHub
export GITHUB_TOKEN=your_token
devfolio contributor profile --user grokify -o profile.json
```

### DevX Usage Dashboard

```bash
# Requires events already collected into the local OmniDevX store
# (via omnidevx-core providers — this command only reads/reports)
devfolio devx dashboard --person person:jane -o dashboard.json
```

## Commands

### Team Commands

```bash
# Generate team velocity dashboard
devfolio team velocity <portfolio.json> [flags]

Flags:
  -o, --output string      Output file (default: stdout)
      --granularity string Time granularity: day, week, month (default "week")
      --since string       Start date (YYYY-MM-DD)
      --until string       End date (YYYY-MM-DD)
```

### Contributor Commands

```bash
# Generate contributor profile
devfolio contributor profile [flags]

Flags:
      --user string         GitHub username (required)
  -o, --output string       Output file (default: stdout)
      --org strings         Filter to specific organizations
      --since string        Start date (YYYY-MM-DD)
      --until string        End date (YYYY-MM-DD)
      --api-only            Force API-only mode, skip local repo detection
      --local-path string   Additional local path to search for repos
      --dashboard           Output dashforge-compatible dashboard JSON
```

### DevX Commands

```bash
# Export a dashforge dashboard from the local OmniDevX store
devfolio devx dashboard [flags]

Flags:
      --person string      Canonical personId to report on (required)
      --days int            Number of days ending today to report on (default 30)
      --store-dir string    OmniDevX store directory (default: ~/.plexusone/omnidevx/data)
  -o, --output string       Output file (default: stdout)
```

## Output Formats

### Team Velocity Dashboard

The velocity dashboard includes:

- Total releases and changelog entries
- Breakdown by category (features, fixes, improvements, etc.)
- Time series data for velocity trends
- Activity heatmap data (GitHub-style)
- Per-project contribution breakdown

Compatible with [dashforge](https://github.com/plexusone/dashforge) static dashboards.

### DevX Usage Dashboard

Built from the [OmniDevX](https://github.com/plexusone/omnidevx-core)
local event store (Claude Code, Codex CLI, git, and GitHub activity in one
canonical model), exported as a [dashforge](https://github.com/plexusone/dashforge)
dashboard: headline metric tiles (sessions, prompts, commits, AI-assisted
%, tool calls, cost, coverage), daily activity/cost charts, and a
source-coverage table. Unlike `contributor profile --dashboard`, this
export is built against dashforge's `dashboardir` package directly, so
its chart widgets render correctly in dashforge's current viewer.

Can also be served through [VisionStudio](https://github.com/ProductBuildersHQ/visionstudio)'s
DevX panel by writing the output to `~/.plexusone/omnidevx/dashboard.json`.

### Contributor Profile

The contributor profile includes:

- User information (name, bio, location, etc.)
- Repository breakdown with contribution counts
- Language statistics
- Daily activity data for heatmap visualization
- Aggregate statistics (commits, PRs, issues, reviews)
- **AI collaboration metrics** (see below)

### AI Collaboration Tracking

devfolio tracks AI-assisted development by detecting co-author signatures in commits. This measures how "AI-native" a developer is.

**Supported AI Tools:**

| Tool | Detection Method | Status |
|------|-----------------|--------|
| Claude Code | `Co-Authored-By: Claude <noreply@anthropic.com>` | Recognized by GitHub |
| GitHub Copilot | `Co-Authored-By: ... <noreply@github.com>` | Recognized by GitHub |
| Gemini CLI | `Co-Authored-By: gemini-cli ... <218195315+gemini-cli@users.noreply.github.com>` | Recognized by GitHub |
| Cursor | `Co-Authored-By: ... <ai@cursor.sh>` | Detection via message parsing |
| Aider | `Co-Authored-By: ... <aider@aider.chat>` | Detection via message parsing |

All tools are detected by parsing commit messages for `Co-Authored-By:` trailers.

**AI Stats Output:**

```json
{
  "aiStats": {
    "totalAiCommits": 42,
    "aiCommitPercent": 23.5,
    "byTool": {
      "Claude Code": {
        "name": "Claude Code",
        "commits": 35,
        "firstUsed": "2024-06-15",
        "lastUsed": "2025-02-26",
        "recognized": true
      }
    },
    "mostUsedTool": "Claude Code",
    "firstAiCommit": "2024-06-15",
    "aiActivity": [
      {"date": "2025-02-25", "count": 3},
      {"date": "2025-02-26", "count": 5}
    ]
  }
}
```

This data can be used to:

- Showcase AI-native development practices in portfolios
- Track adoption of AI tools across a team
- Measure productivity impact of AI assistance

## Requirements

- Go 1.26 or later
- `GITHUB_TOKEN` environment variable (for `contributor profile` only —
  `team velocity` and `devx dashboard` don't call the GitHub API)

## Authentication

DevFolio requires a GitHub personal access token set as `GITHUB_TOKEN`:

```bash
export GITHUB_TOKEN=your_token_here
```

### Fine-Grained Token (Recommended)

Create at: https://github.com/settings/personal-access-tokens/new

**Repository access:**
- Select "Public repositories (read-only)" for public repos
- Or select specific repos if you need private repo data

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

### Classic Token

Create at: https://github.com/settings/tokens/new

**Required scopes:**

| Scope | Purpose |
|-------|---------|
| `public_repo` | Access public repository data |
| `read:user` | Read user profile information |

Add `repo` scope instead of `public_repo` if you need access to private repositories.

## Documentation

Full documentation at [plexusone.github.io/devfolio](https://plexusone.github.io/devfolio)

## Related Projects

- [omnidevx-core](https://github.com/plexusone/omnidevx-core) - Canonical event model, local store, period-report aggregation
- [structured-changelog](https://github.com/grokify/structured-changelog) - JSON changelog format and aggregation
- [gogithub](https://github.com/grokify/gogithub) - GitHub API utilities
- [dashforge](https://github.com/plexusone/dashforge) - Static dashboard generation (the `devx dashboard` export format)
- [VisionStudio](https://github.com/ProductBuildersHQ/visionstudio) - renders `devx dashboard` output in its DevX panel

## License

MIT

 [go-ci-svg]: https://github.com/plexusone/devfolio/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/devfolio/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/devfolio/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/devfolio/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/devfolio/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/devfolio/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/devfolio
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/devfolio
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/devfolio
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/devfolio
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.github.io/devfolio
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fdevfolio
 [loc-svg]: https://tokei.rs/b1/github/plexusone/devfolio
 [repo-url]: https://github.com/plexusone/devfolio
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/devfolio/blob/main/LICENSE
