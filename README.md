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
- 🤖 [OmniDevX](https://github.com/plexusone/omnidevx-core) telemetry (Claude Code, Codex CLI, git, GitHub — normalized into one canonical event model)

## Use Cases

- 📊 **Team velocity dashboards** - Engineering managers track team output
- 👤 **Individual contributor portfolios** - Track your own contributions over time
- 🤖 **AI usage dashboards** - See your own AI-assisted development activity (sessions, tokens, cost, AI-assisted commits)
- 📈 **Quarterly reports** - Combine GitHub stats, git commit analytics, changelog highlights, and token spend into one HTML report
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

# Calendar-month report, written to
# ~/.plexusone/omnidevx/reports/monthly/2026-08.json
devfolio devx dashboard --person person:jane --period monthly
```

### Quarterly Report

```bash
devfolio quarterly report \
  --username grokify \
  --since 2026-04-01 --until 2026-06-30 \
  --repos ~/go/src/github.com/grokify \
  --events ~/.plexusone/omnidevx/data/events \
  --output Q2-2026.html
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
      --dashboard           Output uiforge-compatible dashboard JSON
```

### DevX Commands

```bash
# Export a uiforge dashboard from the local OmniDevX store
devfolio devx dashboard [flags]

Flags:
      --person string      Canonical personId to report on (required)
      --days int            Number of days ending today to report on (default 30, ignored with --period)
      --store-dir string    OmniDevX store directory (default: ~/.plexusone/omnidevx/data)
  -o, --output string       Output file (default: stdout, or the standard reports path with --period)
      --period string       Generate a calendar period report instead of a rolling window: weekly, monthly, or quarterly
      --for string          Anchor date for --period, YYYY-MM-DD (default: today)
```

### Quarterly Commands

```bash
# Generate a quarterly developer report
devfolio quarterly report [flags]

Flags:
      --username string      GitHub username (required)
      --year int              Year (e.g., 2026)
      --quarter int           Quarter (1-4)
      --since string          Start date (YYYY-MM-DD), overrides --year/--quarter
      --until string          End date (YYYY-MM-DD), overrides --year/--quarter
      --repos string          Root directory containing git repos
      --stats string          Directory with gogithub profile stats (report.json)
      --stats-file string     Direct path to a quarterly stats JSON file
      --events string         omnidevx events directory for token spend
  -o, --output string         Output file path (default "quarterly-report.html")
      --format string         Output format: html, json, dashboard (default "html")
      --chart-engine string   Chart engine: svg (self-contained), echarts (CDN) (default "svg")
```

## Output Formats

### Team Velocity Dashboard

The velocity dashboard includes:

- Total releases and changelog entries
- Breakdown by category (features, fixes, improvements, etc.)
- Time series data for velocity trends
- Activity heatmap data (GitHub-style)
- Per-project contribution breakdown

Compatible with [uiforge](https://github.com/plexusone/uiforge) static dashboards.

### DevX Usage Dashboard

Built from the [OmniDevX](https://github.com/plexusone/omnidevx-core)
local event store (Claude Code, Codex CLI, git, and GitHub activity in one
canonical model), exported as a [uiforge](https://github.com/plexusone/uiforge)
dashboard: headline metric tiles (sessions, prompts, commits, AI-assisted
%, tool calls, cost, coverage), daily activity/cost charts, and a
source-coverage table. Unlike `contributor profile --dashboard`, this
export is built against uiforge's `dashboardir` package directly, so
its chart widgets render correctly in uiforge's current viewer.

Can also be served through [VisionStudio](https://github.com/ProductBuildersHQ/visionstudio)'s
DevX panel by writing the output to `~/.plexusone/omnidevx/dashboard.json`.

**`--period weekly|monthly|quarterly`** builds a calendar-aligned report
instead — weeks are always Monday-Sunday, months/quarters add donut and
stacked-bar model-breakdown charts (monthly gets a weekly breakdown,
quarterly gets both weekly and monthly). Written by default to
`~/.plexusone/omnidevx/reports/{type}/{label}.json`, which VisionStudio's
period selector reads via `GET /api/devx/periods` and
`GET /api/devx/reports/{periodType}/{label}`.

### Quarterly Report

Joins four data sources into one report: GitHub stats (via `gogithub/profile`),
git commit analytics (via `gogit`), changelog highlights (via
`structured-changelog`), and token spend (via `omnidevx-core`, when
`--events` is passed).

- **`--format html`** (default) — self-contained HTML report: summary
  metrics, commit-category breakdown, LOC distribution, project
  highlights, and (when `--events` is set) a token-spend section with
  per-model/per-category donut and stacked-bar charts. Charts render as
  inline SVG by default, or via CDN-hosted ECharts with `--chart-engine echarts`.
- **`--format dashboard`** — exports to [uiforge](https://github.com/plexusone/uiforge)
  Dashboard IR (JSON) for rendering in other tools.
- **`--format json`** — the raw report data as JSON, for custom processing.

Any repo with a `CHANGELOG.md`/`CHANGELOG.json` under `--repos` also
contributes project highlights automatically — no separate flag needed.

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
- [gogit](https://github.com/grokify/gogit) - Git history parsing, commit stats
- [gogithub](https://github.com/grokify/gogithub) - GitHub API utilities
- [uiforge](https://github.com/plexusone/uiforge) - Static dashboard generation (the `devx dashboard`/`quarterly report --format dashboard` export format)
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
