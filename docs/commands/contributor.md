# Contributor Commands

Generate individual contributor portfolios from GitHub activity. Requires
`GITHUB_TOKEN` — see [Getting Started](../getting-started.md#authentication-contributor-profile-only).

## `devfolio contributor profile`

```bash
devfolio contributor profile --user grokify -o profile.json
```

| Flag | Description |
|------|-------------|
| `--user` | GitHub username (required) |
| `-o`, `--output` | Output file (default: stdout) |
| `--since` | Start date (`YYYY-MM-DD`) |
| `--until` | End date (`YYYY-MM-DD`) |
| `--org` | Filter to specific organizations (repeatable) |
| `--api-only` | Force API-only mode, skip local repo detection |
| `--local-path` | Additional local path to search for repos |
| `--dashboard` | Output [dashforge](https://github.com/plexusone/dashforge)-compatible dashboard JSON instead of the raw profile |

### Examples

```bash
# Generate profile for a user
devfolio contributor profile --user grokify -o profile.json

# Generate a dashforge-compatible dashboard
devfolio contributor profile --user grokify --dashboard -o dashboard.json

# Limit to specific organizations
devfolio contributor profile --user grokify --org fleet-ops --org agentplexus

# Filter by date range
devfolio contributor profile --user grokify --since 2024-01-01
```

### Local repo detection

By default, for each repository DevFolio checks whether a local clone
exists under `~/go/src/github.com` (or an additional path passed via
`--local-path`) and, if so, reads contribution data from `git log`
locally instead of the GitHub API — faster, avoids rate limits, and sees
full commit messages for AI co-author trailer detection. It falls back to
the API automatically if no local clone is found or the local read fails.
Pass `--api-only` to skip local detection entirely and always use the API.

Progress (repository fetch, per-repo processing) is reported to stderr as
profile generation runs.

!!! note "`--dashboard` chart rendering"
    `--dashboard` uses an older chart-widget shape (`output/dashboard`
    package) that predates dashforge's current viewer format. Metric and
    table widgets render correctly; chart widgets (the language pie chart,
    AI-tools bar chart) do not currently render in dashforge's viewer. The
    `devx dashboard` command's export does not have this issue — it's built
    against dashforge's `dashboardir` package directly.

## Profile Contents

| Field | Description |
|-------|-------------|
| **Repositories** | List of repos with commits, PRs, issues, reviews |
| **Stats** | Aggregate totals across all repositories |
| **Languages** | Programming language breakdown |
| **Activity** | Daily activity data for heatmap visualization |
| **AI Stats** | AI collaboration metrics (see below) |

## AI Collaboration Tracking

DevFolio tracks AI-assisted development by parsing `Co-Authored-By:`
trailers in commit messages.

| Tool | Email Pattern | Recognized by GitHub |
|------|---------------|----------------------|
| Claude Code | `noreply@anthropic.com` | Yes |
| GitHub Copilot | `noreply@github.com`, `copilot@github.com` | Yes |
| Gemini CLI | `218195315+gemini-cli@users.noreply.github.com` | Yes |
| Cursor | `ai@cursor.sh` | No (message parsing) |
| Aider | `aider@aider.chat` | No (message parsing) |

**AI stats output:**

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

This measures co-author *signal*, not a verified AI-contribution
percentage — a commit is counted if its trailer names a known AI tool,
not by analyzing which lines the tool actually wrote.
