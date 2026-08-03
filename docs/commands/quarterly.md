# Quarterly Commands

Generate a quarterly developer report joining GitHub stats, git commit
analytics, changelog highlights, and token spend data into one report.
Each data source is optional and independent — omit a flag and that
section of the report is simply empty.

## `devfolio quarterly report`

```bash
devfolio quarterly report \
  --username grokify \
  --since 2026-04-01 --until 2026-06-30 \
  --repos ~/go/src/github.com/grokify \
  --events ~/.plexusone/omnidevx/data/events \
  --output Q2-2026.html
```

| Flag | Description |
|------|-------------|
| `--username` | GitHub username (required) |
| `--year` | Year (e.g., `2026`) |
| `--quarter` | Quarter (1-4) |
| `--since` | Start date (`YYYY-MM-DD`), overrides `--year`/`--quarter` |
| `--until` | End date (`YYYY-MM-DD`), overrides `--year`/`--quarter` |
| `--repos` | Root directory containing git repos |
| `--stats` | Directory with `gogithub` profile stats (`report.json`) |
| `--stats-file` | Direct path to a quarterly stats JSON file |
| `--events` | `omnidevx-core` events directory for token spend |
| `-o`, `--output` | Output file path (default `quarterly-report.html`) |
| `--format` | Output format: `html`, `json`, `dashboard` (default `html`) |
| `--chart-engine` | Chart engine: `svg` (self-contained), `echarts` (CDN) (default `svg`) |

## Data sources

### Git repositories (`--repos`)

Scans local git repos for commit history: commit counts, insertions,
deletions, conventional-commit categories, AI-assisted commits (Claude
Code co-author trailer detection), and LOC by category.

### GitHub stats (`--stats` / `--stats-file`)

Pre-generated GitHub stats from
[gogithub/profile](https://github.com/grokify/gogithub), covering PRs
opened/merged, reviews given, releases published, and repo contributions.
Generate it separately, then pass it in:

```bash
gogithub profile stats --username grokify --since 2026-04-01 --until 2026-06-30 \
  --output ~/stats/Q2-2026/report.json

devfolio quarterly report --stats ~/stats/Q2-2026 ...
# or
devfolio quarterly report --stats-file ~/stats/Q2-2026/report.json ...
```

If neither flag is set, the GitHub stats section is empty.

### Token spend events (`--events`)

Reads Claude Code telemetry from
[omnidevx-core](https://github.com/plexusone/omnidevx-core) event files
under `~/.plexusone/omnidevx/data/events/{year}/{month}/{day}/` — input/
output/cache tokens per model, with cost calculated from embedded
pricing. If `--events` isn't provided, the token spend section is empty.

### Changelog data

Extracted automatically from repos passed via `--repos` — no separate
flag. Looks for `CHANGELOG.md` or `CHANGELOG.json` in each repository and
pulls release highlights via
[structured-changelog](https://github.com/grokify/structured-changelog).

## Output formats

### `--format html` (default)

Self-contained HTML report with:

- Summary metrics (commits, LOC, releases, repos, AI-assisted %)
- Commit category breakdown (bar chart + donut)
- LOC by category distribution
- Project highlights grid
- Token spend section (when `--events` is set): donut charts by model and
  category, summary cards, stacked bar chart, and a detailed table with
  CSV download

Charts render as self-contained SVG by default. Pass
`--chart-engine echarts` for interactive ECharts (loaded from a CDN —
requires network access to view).

### `--format dashboard`

Exports to [uiforge](https://github.com/plexusone/uiforge) Dashboard IR
(JSON) for rendering in other tools.

### `--format json`

The raw report data as JSON, for custom processing.

## Examples

```bash
# Quarter shorthand instead of --since/--until
devfolio quarterly report --username grokify --year 2026 --quarter 2 \
  --repos ~/go/src/github.com/grokify --output Q2-2026.html

# Interactive charts
devfolio quarterly report --username grokify --since 2026-04-01 --until 2026-06-30 \
  --repos ~/go/src/github.com/grokify --chart-engine echarts --output Q2-2026-echarts.html

# Dashboard IR export
devfolio quarterly report --username grokify --since 2026-04-01 --until 2026-06-30 \
  --repos ~/go/src/github.com/grokify --format dashboard --output Q2-2026-dashboard.json
```

## Pricing

Token costs are calculated from embedded pricing in
`omnidevx-core/report/pricing.json`:

```text
Cost = (Input × $/M) + (Output × $/M) + (CacheRead × $/M) + (CacheWrite × $/M)
```

Models are grouped by tier — Fable > Opus > Sonnet > Haiku — sorted
newest version first within each tier.
