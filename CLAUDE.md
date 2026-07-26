# CLAUDE.md — devfolio

DevFolio is a developer portfolio and team velocity metrics generator. It produces HTML reports and dashboards from git history, GitHub activity, changelog data, and AI token spend.

## Quick Start

```bash
# Generate a quarterly report
devfolio quarterly report \
  --username grokify \
  --since 2026-04-01 \
  --until 2026-06-30 \
  --repos ~/go/src/github.com/grokify \
  --events ~/.plexusone/omnidevx/data/events \
  --output ~/stats/Q2-2026.html

# Use ECharts instead of SVG for interactive charts
devfolio quarterly report ... --chart-engine echarts
```

## Data Sources

### 1. Git Repositories (`--repos`)

Scans local git repos for commit history. Extracts:

- Commit counts, insertions, deletions
- Conventional commit categories (feat, fix, docs, chore, etc.)
- AI-assisted commits (detects Claude Code co-author trailers)
- Lines of code by category

**Flag:** `--repos <directory>`

**Example:**
```bash
devfolio quarterly report --repos ~/go/src/github.com/grokify ...
```

**Path:** Directory containing git repos (scans recursively for `.git` directories)

### 2. GitHub Stats (`--stats` or `--stats-file`)

Pre-generated GitHub stats from [gogithub/profile](https://github.com/grokify/gogithub). Provides:

- PRs opened/merged, reviews given
- Releases published
- Repository contributions
- Additions/deletions from GitHub API

**Flags:**

- `--stats <directory>` — Directory containing `report.json` from gogithub profile
- `--stats-file <path>` — Direct path to a quarterly stats JSON file

**How to generate:**
```bash
# Using gogithub profile CLI
gogithub profile stats --username grokify --since 2026-04-01 --until 2026-06-30 \
  --output ~/stats/Q2-2026/report.json

# Then pass to devfolio
devfolio quarterly report --stats ~/stats/Q2-2026 ...
# or
devfolio quarterly report --stats-file ~/stats/Q2-2026/report.json ...
```

**Note:** If neither `--stats` nor `--stats-file` is provided, GitHub stats section will be empty.

### 3. Token Spend Events (`--events`)

Reads Claude Code telemetry from omnidevx event files:

- Input/output tokens per model
- Cache read/write tokens
- Cost calculation using embedded pricing

**Flag:** `--events <directory>`

**Example:**
```bash
devfolio quarterly report --events ~/.plexusone/omnidevx/data/events ...
```

**Path:** `~/.plexusone/omnidevx/data/events/` (date-partitioned JSONL)

**Structure:**
```
events/
└── {year}/{month}/{day}/
    ├── claude-code.jsonl  # Token events (ai.message.completed)
    └── git.jsonl          # Commit events (devx.change.committed)
```

**How to generate:** Token events are collected by omnidevx-core from Claude Code session logs. Run `omnidevx collect` or configure automatic collection.

**Note:** If `--events` is not provided, Token Spend section will be empty.

### 4. Changelog Data

Parses structured changelogs for release highlights (auto-detected from repos):

- Version releases with dates
- Categorized changes (Added, Changed, Fixed, etc.)
- Project descriptions

**No flag required** — extracted automatically from repos specified in `--repos`.

Looks for `CHANGELOG.md` or `CHANGELOG.json` in each repository.

## Output Formats

### HTML Report (`--format html`, default)

```bash
devfolio quarterly report \
  --username grokify \
  --since 2026-04-01 --until 2026-06-30 \
  --repos ~/go/src/github.com/grokify \
  --stats ~/stats/Q2-2026 \
  --events ~/.plexusone/omnidevx/data/events \
  --output ~/reports/Q2-2026.html

# With interactive ECharts (requires CDN)
devfolio quarterly report ... --chart-engine echarts --output ~/reports/Q2-2026-echarts.html
```

Self-contained HTML with:

- Summary metrics (commits, LOC, releases, repos, AI-assisted %)
- Commit category breakdown (bar chart + donut)
- LOC by category distribution
- Project highlights grid
- Token spend section:
  - Donut charts: tokens by model, cost by model
  - Donut charts: tokens by category, cost by category
  - Summary cards: input/output/cache tokens, total cost
  - Stacked bar: tokens and cost breakdown per model
  - Detailed table with CSV download

**Chart engines:**

- `svg` (default): Self-contained, no external dependencies
- `echarts`: Interactive charts via CDN, tooltips, hover effects

### Dashboard IR (`--format dashboard`)

```bash
devfolio quarterly report \
  --username grokify \
  --since 2026-04-01 --until 2026-06-30 \
  --repos ~/go/src/github.com/grokify \
  --format dashboard \
  --output ~/reports/Q2-2026-dashboard.json
```

Exports to uiforge Dashboard IR (JSON) for rendering in other tools.

### JSON (`--format json`)

```bash
devfolio quarterly report \
  --username grokify \
  --since 2026-04-01 --until 2026-06-30 \
  --repos ~/go/src/github.com/grokify \
  --format json \
  --output ~/reports/Q2-2026.json
```

Raw report data as JSON for custom processing.

## Metrics Generated

### Commit Metrics

| Metric | Description |
|--------|-------------|
| Total Commits | Count of commits in period |
| Insertions | Lines added |
| Deletions | Lines removed |
| Net Lines | Insertions - Deletions |
| By Category | Breakdown by conventional commit type |
| AI-Assisted | Commits with Claude Code co-author |
| AI % | Percentage of AI-assisted commits |

### Token Metrics

| Metric | Description |
|--------|-------------|
| Input Tokens | Tokens sent to model |
| Output Tokens | Tokens generated by model |
| Cache Read Tokens | Tokens read from prompt cache |
| Cache Write Tokens | Tokens written to prompt cache |
| Total Tokens | Sum of all token types |
| Cost USD | Calculated from model pricing |

### Model Breakdown

Per-model stats for: Fable 5, Opus 4.x, Sonnet 5, Haiku 4.5

Sorted by tier (Fable > Opus > Sonnet > Haiku), then version descending.

## Pricing

Token costs are calculated using embedded pricing from `omnidevx-core/report/pricing.json`:

```
Cost = (Input × $/M) + (Output × $/M) + (CacheRead × $/M) + (CacheWrite × $/M)
```

Current pricing (2026-07):

| Model | Input | Output | Cache Read | Cache Write |
|-------|-------|--------|------------|-------------|
| Fable 5 | $10.00/M | $50.00/M | $1.00/M | $12.50/M |
| Opus 4.x | $5.00/M | $25.00/M | $0.50/M | $6.25/M |
| Sonnet 5 | $2.00/M | $10.00/M | $0.20/M | $2.50/M |
| Haiku 4.5 | $1.00/M | $5.00/M | $0.10/M | $1.25/M |

## Common Tasks

### Add a new chart to HTML report

1. Edit `cmd/devfolio/quarterly_report.go` (main HTML generation, ~1400 lines)
2. Add SVG version first (self-contained), then ECharts version
3. Use existing patterns: `donut-chart`, `stacked-chart` CSS classes
4. Test with: `go run ./cmd/devfolio quarterly report --repos ~/go/src/github.com/grokify --events ~/.plexusone/omnidevx/data/events --output /tmp/test.html && open /tmp/test.html`

### Add a new metric

1. Add field to `quarterly/report.go` structs (`Report`, `TokenSpendSummary`, etc.)
2. Populate in `cmd/devfolio/quarterly_report.go` data collection
3. Display in HTML generation section

### Update model pricing

1. Edit `~/go/src/github.com/plexusone/omnidevx-core/report/pricing.json`
2. Run `go mod tidy` in devfolio to pick up changes
3. Costs are calculated at report generation time via `report.LookupPricing()`

### Test HTML output

```bash
# Generate and open in browser
go run ./cmd/devfolio quarterly report \
  --repos ~/go/src/github.com/grokify \
  --events ~/.plexusone/omnidevx/data/events \
  --output /tmp/test.html && open /tmp/test.html

# Test both chart engines
go run ./cmd/devfolio quarterly report ... --chart-engine svg --output /tmp/test-svg.html
go run ./cmd/devfolio quarterly report ... --chart-engine echarts --output /tmp/test-echarts.html

# Test print styles: open HTML, Cmd+P to preview print
```

## Key Files

| File | Purpose | Lines |
|------|---------|-------|
| `cmd/devfolio/quarterly_report.go` | HTML generation, charts, CSS | ~1400 |
| `quarterly/report.go` | Report struct, data types | ~100 |
| `quarterly/dashboard.go` | Dashboard IR export | ~300 |

Most feature work happens in `quarterly_report.go`.

## Known Issues

- `changelog.RankedRelease` and `changelog.ExtractMultiRepoHighlights` are undefined — changelog integration is incomplete, ignore these compiler warnings for now
- Project Highlights section may be empty if changelogs aren't found

## Architecture Decisions

| Decision | Rationale |
|----------|-----------|
| SVG charts as default | Self-contained HTML, no CDN dependency, works offline |
| ECharts as option | Interactive tooltips, hover effects for presentation |
| Pricing in omnidevx-core | Shared across tools, single source of truth |
| Embedded via `//go:embed` | No runtime file dependencies, pricing baked into binary |
| Cost calculated at report time | Allows re-running old data with updated pricing |
| Model sort by tier then version | Fable > Opus > Sonnet > Haiku, newest versions first |

## Code Style

### HTML Generation Pattern

All HTML is built via `strings.Builder` with `sb.WriteString()`:

```go
var sb strings.Builder
sb.WriteString(`<div class="card">`)
sb.WriteString(fmt.Sprintf(`<div class="card-value">%d</div>`, value))
sb.WriteString(`</div>`)
```

- Use backticks for HTML strings (avoids escaping quotes)
- Use `fmt.Sprintf()` for interpolation
- Close tags explicitly, don't rely on browser auto-closing

### Adding Charts

1. **SVG version first** — always implement self-contained SVG
2. **ECharts second** — wrap in `if useECharts { ... } else { ... }`
3. **Both share same container classes** — `.donut-chart`, `.stacked-chart`

### Number Formatting

```go
formatNumber(12345)      // "12,345" — with commas
formatTokensK(1234567)   // "1,235" — divide by 1000, for token columns
formatWithCommas(12345)  // "12,345" — recursive comma insertion
```

## CSS Class Naming

### Layout

| Class | Purpose |
|-------|---------|
| `.grid` | 3-column responsive grid for metric cards |
| `.card` | Container with background, border, padding |
| `.card-title` | Metric label (uppercase, muted) |
| `.card-value` | Large metric number |

### Charts

| Class | Purpose |
|-------|---------|
| `.donut-row` | Flex container for side-by-side donuts |
| `.donut-chart` | Single donut with title, legend |
| `.donut-container` | SVG wrapper with aspect-ratio |
| `.donut-svg` | The SVG element (rotated -90deg) |
| `.donut-segment` | Each pie slice (circle with stroke-dasharray) |
| `.donut-center` | Centered label overlay |
| `.donut-legend` | Legend below chart |
| `.echarts-container` | ECharts div (340px height) |
| `.stacked-chart` | Stacked bar chart container |
| `.stacked-bar` | Single bar column |
| `.stacked-bar-segment` | One segment of a stacked bar |

### Lists

| Class | Purpose |
|-------|---------|
| `.category-list` | Horizontal bar chart as list |
| `.category-item` | Row with name, bar, percentage, count |
| `.category-bar` | Filled bar (width set inline) |
| `.model-table` | Table for token/cost breakdown |
| `.model-table .num` | Right-aligned numeric columns |

### CSS Variables

```css
:root {
  --bg: #0f172a;           /* Page background */
  --bg-card: #1e293b;      /* Card background */
  --text: #f1f5f9;         /* Primary text */
  --text-muted: #94a3b8;   /* Secondary text */
  --accent: #3b82f6;       /* Primary accent (blue) */
  --accent-light: rgba(59, 130, 246, 0.1);  /* Hover state */
  --border: #334155;       /* Borders */
}
```

Dark theme by default. Print styles preserve dark theme with `print-color-adjust: exact`.

## Debugging Tips

### Inspect Generated HTML

```bash
# Generate to temp file and open
go run ./cmd/devfolio quarterly report ... --output /tmp/test.html && open /tmp/test.html

# View raw HTML
cat /tmp/test.html | less

# Check for malformed HTML
tidy -errors /tmp/test.html 2>&1 | head -20
```

### Debug Token Data

```bash
# Check what token files exist
find ~/.plexusone/omnidevx/data/events -name "claude-code.jsonl" | wc -l

# Sample token event
head -1 ~/.plexusone/omnidevx/data/events/2026/07/03/claude-code.jsonl | jq .

# Sum tokens manually
jq -s '[.[].attributes.input_tokens // 0] | add' \
  ~/.plexusone/omnidevx/data/events/2026/07/*/claude-code.jsonl
```

### Debug Chart Rendering

- **SVG not showing**: Check `stroke-dasharray` calculations (circumference = 439.82 for r=70)
- **ECharts blank**: Open browser console, check for JS errors
- **Colors wrong**: Check gradient IDs match between `<defs>` and `stroke="url(#...)"`
- **Donut percentages off**: Verify total calculation before computing individual %

### Common Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| Empty Token Spend section | `--events` not provided | Add `--events ~/.plexusone/omnidevx/data/events` |
| $0.00 cost for model | Model not in pricing.json | Add to omnidevx-core/report/pricing.json |
| Chart segments don't add to 100% | Gap percentage too high | Reduce `gapPct` constant |
| Table row with empty model | Zero-token entries | Check skip condition: `model == "" \|\| tokens == 0` |
| Print cuts off charts | Missing `break-inside: avoid` | Add to `.donut-row`, `.stacked-chart` |

## Development

### Build & Test

```bash
# Build
go build ./...

# Run tests
go test -v ./...

# Lint
golangci-lint run

# Run locally
go run ./cmd/devfolio quarterly report --help
```

### Dependencies

Key dependencies (check latest versions before updating):

| Module | Purpose |
|--------|---------|
| `github.com/grokify/gogit` | Git history parsing, commit stats |
| `github.com/grokify/gogithub` | GitHub API, profile stats |
| `github.com/grokify/structured-changelog` | Changelog parsing |
| `github.com/plexusone/omnidevx-core` | Token event parsing, model pricing |
| `github.com/plexusone/uiforge` | Dashboard IR types |
| `github.com/spf13/cobra` | CLI framework |

### Related Repositories

| Repo | Location | Purpose |
|------|----------|---------|
| omnidevx-core | `~/go/src/github.com/plexusone/omnidevx-core` | Event schemas, pricing data (`report/pricing.json`) |
| gogit | `~/go/src/github.com/grokify/gogit` | Git parsing library |
| gogithub | `~/go/src/github.com/grokify/gogithub` | GitHub stats, profile module |
| structured-changelog | `~/go/src/github.com/grokify/structured-changelog` | Changelog parsing |

**Updating pricing:** Edit `omnidevx-core/report/pricing.json`, then `go mod tidy` in devfolio to pick up changes.

### Sample Data Locations

| Data | Path | Notes |
|------|------|-------|
| Token events | `~/.plexusone/omnidevx/data/events/` | 13 months of JSONL |
| Git repos | `~/go/src/github.com/grokify/` | 400+ repos |
| Output reports | `~/go/src/github.com/grokify/grokify.github.io/src/stats/` | Published HTML |

## Conventions

- **Go module:** `github.com/plexusone/devfolio`
- **Datasources:** pluggable via `datasource/` packages
- **Output:** HTML reports and structured JSON
- **Pricing:** Embedded via `//go:embed` in omnidevx-core

## Project Structure

```
devfolio/
├── cmd/devfolio/
│   ├── main.go
│   ├── quarterly.go        # quarterly command group
│   └── quarterly_report.go # HTML report generation
├── quarterly/
│   ├── report.go           # Report struct, data types
│   └── dashboard.go        # Dashboard IR export
├── datasource/
│   ├── git/                # Git history parsing
│   ├── github/             # GitHub API client
│   └── changelog/          # Changelog parsing
└── docs/specs/
    ├── PRD.md, TRD.md, PLAN.md, ROADMAP.md
    └── doltdb/             # DoltDB analytics spec
```

## Future: DoltDB Analytics

Spec in `docs/specs/doltdb/` for SQL-based analytics:

- Import JSONL events to Dolt database
- SQL queries for aggregations
- Version-controlled data with branching
- Remote backup to DoltHub

## PRISM Control

This repo is registered in [prism-control](https://github.com/ProductBuildersHQ/prism-control). Use `prismctl work ready --repo github.com/plexusone/devfolio` to find claimable work, and carry the `Refs: RMI-DEVFOLIO-<NNN>` trailer on every commit.
