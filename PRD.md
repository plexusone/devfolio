# Product Requirements Document: DevFolio

## Overview

DevFolio is a CLI tool for generating developer portfolios and team velocity metrics from multiple data sources including GitHub activity, git history, and structured changelogs.

## Problem Statement

Engineering teams and individual developers lack unified tools to:

1. **Measure AI-native development** - No standard way to track AI coding assistant adoption and impact
2. **Generate developer portfolios** - Manual process to aggregate contributions across repositories
3. **Track team velocity** - Disconnected metrics across changelogs, commits, and PRs
4. **Visualize development activity** - GitHub's contribution graph is limited to single users

## Target Users

| User | Primary Use Case |
|------|------------------|
| Individual Contributors | Portfolio generation for career development |
| Engineering Managers | Team velocity dashboards and metrics |
| Recruiters (Hiring) | Evaluate candidate contribution patterns |
| Job Seekers | Showcase development portfolio |
| Open Source Maintainers | Track contributor activity across projects |

## Core Features

### Phase 1: Individual Contributor Profiles (v0.1.0) ✅

#### 1.1 GitHub Profile Generation

Generate contributor profiles from GitHub API data.

**Input:**
- GitHub username
- Optional: organization filter, date range

**Output:**
- Repository breakdown with contribution counts
- Aggregate statistics (commits, PRs, issues, reviews)
- Language distribution
- Activity heatmap data

**CLI:**
```bash
devfolio contributor profile --user <username> [--org <org>...] [--since YYYY-MM-DD] [--until YYYY-MM-DD] -o profile.json
```

#### 1.2 AI Collaboration Tracking

Detect and track AI-assisted development via commit co-author signatures.

**Supported Tools:**
| Tool | Email Pattern | GitHub Recognized |
|------|---------------|-------------------|
| Claude Code | `noreply@anthropic.com` | Yes |
| GitHub Copilot | `noreply@github.com` | Yes |
| Gemini CLI | `218195315+gemini-cli@users.noreply.github.com` | Yes |
| Cursor | `ai@cursor.sh` | No |
| Aider | `aider@aider.chat` | No |

**Metrics:**
- Total AI-assisted commits
- AI commit percentage
- Per-tool breakdown with first/last used dates
- Most used tool
- AI activity heatmap

### Phase 2: Team Velocity (v0.2.0)

#### 2.1 Portfolio Integration

Consume structured-changelog portfolio data for team metrics.

**Input:**
- Portfolio JSON from `schangelog portfolio aggregate`

**Output:**
- Team-wide release velocity
- Category breakdown (features, fixes, improvements)
- Per-project contribution breakdown
- Time series data for trend analysis

**CLI:**
```bash
devfolio team velocity <portfolio.json> [--granularity day|week|month] [--since YYYY-MM-DD] [--until YYYY-MM-DD] -o velocity.json
```

#### 2.2 Multi-Source Aggregation

Combine data from multiple sources:

| Source | Data |
|--------|------|
| structured-changelog | Release entries, categories |
| GitHub API | Commits, PRs, issues, reviews |
| Git history | Commit frequency, file changes |

### Phase 3: Output Formats (v0.3.0)

#### 3.1 Dashboard Export

Export dashforge-compatible JSON for visualization.

**Widgets:**
- Summary metrics (total releases, commits, contributors)
- Activity heatmap (GitHub-style calendar)
- Velocity trend chart (bar/line)
- Category breakdown (pie chart)
- Project table with sortable columns

**CLI:**
```bash
devfolio team velocity portfolio.json --dashboard -o dashboard.json
devfolio contributor profile --user grokify --dashboard -o dashboard.json
```

#### 3.2 Markdown Export

Generate static markdown reports.

**CLI:**
```bash
devfolio contributor profile --user grokify --markdown -o PROFILE.md
devfolio team velocity portfolio.json --markdown -o VELOCITY.md
```

#### 3.3 Static Site Generation

Generate static HTML site with embedded visualizations.

**CLI:**
```bash
devfolio team velocity portfolio.json --site -o ./site/
devfolio contributor profile --user grokify --site -o ./site/
```

### Phase 4: Advanced Features (v0.4.0+)

#### 4.1 Comparative Analysis

Compare contributors or teams over time.

```bash
devfolio contributor compare --users grokify,johndoe --since 2024-01-01
devfolio team compare --portfolios team-a.json,team-b.json
```

#### 4.2 Trend Detection

Identify velocity trends and anomalies.

- Increasing/decreasing velocity
- Seasonal patterns
- Contributor churn detection

#### 4.3 GitHub Actions Integration

Automated profile/dashboard updates via scheduled workflows.

```yaml
- uses: plexusone/devfolio-action@v1
  with:
    command: contributor profile
    user: ${{ github.actor }}
    output: profile.json
```

## Data Sources

### Primary Sources

| Source | Package | Description |
|--------|---------|-------------|
| GitHub API | `datasource/github` | User activity, repos, PRs, issues |
| Git History | `datasource/git` | Local commit analysis |
| Changelogs | `datasource/changelog` | structured-changelog portfolio data |

### Data Flow

```
GitHub API ─────┐
                │
Git History ────┼──▶ Aggregator ──▶ Metrics ──▶ Output
                │
Changelog ──────┘
```

## Output Packages

| Package | Description |
|---------|-------------|
| `output/dashboard` | Dashforge-compatible JSON |
| `output/markdown` | Static markdown reports |
| `output/site` | Static HTML with visualizations |

## Non-Goals

- Real-time monitoring (this is a batch/CLI tool)
- Code review or quality metrics
- CI/CD integration beyond GitHub Actions
- Non-GitHub forges (GitLab, Bitbucket) in v1.x

## Success Metrics

| Metric | Target |
|--------|--------|
| Profile generation time | < 30s for 100 repos |
| Dashboard load time | < 2s for 1 year of data |
| AI tool detection accuracy | > 99% |

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| go-github | v84+ | GitHub API client |
| gogithub | v0.10.0+ | GitHub utilities |
| structured-changelog | v0.12.0+ | Portfolio data |
| cobra | v1.10+ | CLI framework |

## References

- [structured-changelog](https://github.com/grokify/structured-changelog) - Changelog aggregation
- [gogithub](https://github.com/grokify/gogithub) - GitHub API utilities
- [dashforge](https://github.com/grokify/dashforge) - Dashboard visualization
