# DevFolio

Developer portfolio and team velocity metrics generator.

## Overview

DevFolio generates developer portfolios and team velocity dashboards from:

- 📋 Changelog data (via [structured-changelog](https://github.com/grokify/structured-changelog))
- 📜 Git history
- 🐙 GitHub activity (commits, PRs, issues, reviews)
- 🤖 [OmniDevX](https://github.com/plexusone/omnidevx-core) telemetry (Claude Code, Codex CLI, git, GitHub — normalized into one canonical event model)

## Use Cases

- 📊 **Team velocity dashboards** — engineering managers track team output
- 👤 **Individual contributor portfolios** — track your own contributions over time
- 🤖 **AI usage dashboards** — see your own AI-assisted development activity (sessions, tokens, cost, AI-assisted commits)
- 🔍 **Recruiting (hiring)** — evaluate candidate contribution patterns
- 💼 **Recruiting (job seeking)** — showcase your development portfolio

## Quick Start

=== "Contributor Profile"

    ```bash
    export GITHUB_TOKEN=your_token
    devfolio contributor profile --user grokify -o profile.json
    ```

=== "Team Velocity"

    ```bash
    # 1. Create a portfolio from changelogs (using structured-changelog)
    schangelog portfolio discover --org plexusone -o manifest.json
    schangelog portfolio aggregate manifest.json -o portfolio.json

    # 2. Generate team velocity dashboard
    devfolio team velocity portfolio.json -o velocity.json
    ```

=== "DevX Dashboard"

    ```bash
    # Requires events already collected into the local OmniDevX store
    # via the omnidevx-core providers (this command only reads/reports)
    devfolio devx dashboard --person person:jane -o dashboard.json

    # Or a calendar period report (weekly/monthly/quarterly)
    devfolio devx dashboard --person person:jane --period monthly
    ```

=== "Quarterly Report"

    ```bash
    devfolio quarterly report \
      --username grokify \
      --since 2026-04-01 --until 2026-06-30 \
      --repos ~/go/src/github.com/grokify \
      --events ~/.plexusone/omnidevx/data/events \
      --output Q2-2026.html
    ```

See [Commands](commands/contributor.md) for the full flag reference on each.

## Features

### Contributor Profile

- Repository breakdown with contribution counts
- Language statistics
- Daily activity data for heatmap visualization
- Aggregate statistics (commits, PRs, issues, reviews)
- AI collaboration tracking — detects `Co-Authored-By:` trailers from Claude Code, GitHub Copilot, Gemini CLI, Cursor, and Aider
- Optional [uiforge](https://github.com/plexusone/uiforge)-compatible dashboard export (`--dashboard`)

### Team Velocity

- Total releases and changelog entries
- Breakdown by category (features, fixes, improvements, etc.)
- Time series data for velocity trends
- Activity heatmap data (GitHub-style)
- Per-project contribution breakdown

### DevX Dashboard

- Built from the [OmniDevX](https://github.com/plexusone/omnidevx-core) local event store: Claude Code, Codex CLI, git, and GitHub activity in one canonical model
- Headline metric tiles (sessions, prompts, commits, AI-assisted %, tool calls, cost, coverage), daily activity/cost charts, and a source-coverage table
- Exports as a portable [uiforge](https://github.com/plexusone/uiforge) dashboard JSON — open it in uiforge's static viewer, or serve it through [VisionStudio](https://github.com/ProductBuildersHQ/visionstudio)'s DevX panel
- DevFolio only reads and reports; it never collects events itself — that's `omnidevx-core`'s job

### Quarterly Report

- Joins GitHub stats, git commit analytics, changelog highlights, and (with `--events`) OmniDevX token spend into one report
- Self-contained HTML by default (inline SVG charts), or interactive ECharts via CDN with `--chart-engine echarts`
- `--format dashboard` exports the same data as a [uiforge](https://github.com/plexusone/uiforge) Dashboard IR JSON; `--format json` exports the raw report data
- Project highlights are pulled automatically from any `CHANGELOG.md`/`CHANGELOG.json` found under `--repos`

## Installation

```bash
go install github.com/plexusone/devfolio/cmd/devfolio@latest
```

## Requirements

- Go 1.26 or later
- `GITHUB_TOKEN` environment variable (for `contributor profile`; not needed for `devx dashboard`)

## Related Projects

- [omnidevx-core](https://github.com/plexusone/omnidevx-core) — canonical event model, local store, period-report aggregation, embedded pricing
- [uiforge](https://github.com/plexusone/uiforge) — the dashboard-IR format `devx dashboard` and `quarterly report --format dashboard` export to
- [structured-changelog](https://github.com/grokify/structured-changelog) — JSON changelog format and portfolio aggregation
- [gogit](https://github.com/grokify/gogit) — git history parsing, commit stats
- [gogithub](https://github.com/grokify/gogithub) — GitHub API utilities, profile stats

## License

MIT
