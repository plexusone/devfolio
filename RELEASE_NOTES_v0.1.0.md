# Release Notes v0.1.0

This is the initial release of devfolio, a CLI tool for generating developer portfolios and team velocity metrics from GitHub activity. A key feature is AI collaboration tracking to measure how "AI-native" developers are by detecting AI coding assistant co-authors in commit messages.

## Features

### Contributor Profile Generation

Generate comprehensive contributor profiles from GitHub activity:

```bash
devfolio contributor profile --user grokify -o profile.json
```

#### Profile Data

| Field | Description |
|-------|-------------|
| **Repositories** | List of repos with commits, PRs, issues, reviews |
| **Stats** | Aggregate totals across all repositories |
| **Languages** | Programming language breakdown |
| **Activity** | Daily activity data for heatmap visualization |
| **AI Stats** | AI collaboration metrics (see below) |

### AI Collaboration Tracking

Track AI-assisted development by parsing `Co-Authored-By:` trailers in commit messages:

```bash
devfolio contributor profile --user grokify -o profile.json
# profile.json includes aiStats section
```

#### Supported AI Tools

| Tool | Email Pattern | GitHub Recognized |
|------|---------------|-------------------|
| Claude Code | `noreply@anthropic.com` | Yes |
| GitHub Copilot | `noreply@github.com`, `copilot@github.com` | Yes |
| Gemini CLI | `218195315+gemini-cli@users.noreply.github.com` | Yes |
| Cursor | `ai@cursor.sh` | No |
| Aider | `aider@aider.chat` | No |

#### AI Stats Metrics

| Metric | Description |
|--------|-------------|
| `totalAiCommits` | Number of commits with AI co-authors |
| `aiCommitPercent` | Percentage of total commits that are AI-assisted |
| `byTool` | Breakdown by AI tool with first/last used dates |
| `mostUsedTool` | The AI tool used most frequently |
| `firstAiCommit` | Date of first AI-assisted commit |
| `aiActivity` | Daily AI commit counts for heatmap |

### Filtering Options

```bash
# Filter by date range
devfolio contributor profile --user grokify --since 2024-01-01 --until 2024-12-31

# Filter by organizations
devfolio contributor profile --user grokify --org plexusone --org agentplexus
```

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

## Requirements

- Go 1.25+
- `GITHUB_TOKEN` environment variable with appropriate scopes

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `go-github` | v84 | GitHub API client |
| `gogithub` | v0.10.0 | GitHub authentication utilities |
| `structured-changelog` | v0.12.0 | Changelog tooling |
| `cobra` | v1.10.2 | CLI framework |

## Coming Soon

- `team velocity` command for team-level metrics
- Dashboard export for visualization
- Multi-project portfolio aggregation

## Links

- [GitHub Repository](https://github.com/plexusone/devfolio)
- [Changelog](https://github.com/plexusone/devfolio/blob/main/CHANGELOG.md)
