# Team Commands

Generate team velocity metrics from aggregated changelog data. No GitHub
token required — this reads a portfolio file, not the GitHub API directly.

## `devfolio team velocity`

```bash
devfolio team velocity <portfolio.json> [flags]
```

The portfolio file is generated with
[`schangelog portfolio aggregate`](https://github.com/grokify/structured-changelog),
not by DevFolio itself:

```bash
schangelog portfolio discover --org plexusone -o manifest.json
schangelog portfolio aggregate manifest.json -o portfolio.json
devfolio team velocity portfolio.json -o velocity.json
```

| Flag | Description |
|------|-------------|
| `-o`, `--output` | Output file (default: stdout) |
| `--granularity` | Time granularity: `day`, `week`, `month` (default `week`) |
| `--since` | Start date (`YYYY-MM-DD`) |
| `--until` | End date (`YYYY-MM-DD`) |

### Examples

```bash
# Generate velocity dashboard
devfolio team velocity portfolio.json -o velocity.json

# Filter by date range
devfolio team velocity portfolio.json --since 2024-01-01 --until 2024-06-30
```

## Output

The velocity dashboard includes:

- Total releases and changelog entries
- Breakdown by category (features, fixes, improvements, etc.)
- Time series data for velocity trends
- Activity heatmap data (GitHub-style)
- Per-project contribution breakdown

Compatible with [dashforge](https://github.com/plexusone/dashforge) static
dashboards.

!!! note "Team velocity today vs. the OmniDevX rollup"
    This command computes velocity from changelog entries only. The
    ecosystem plan is to rebase team velocity onto rollups of individual
    [`DeveloperPeriodReport`](https://plexusone.github.io/omnidevx-core/concepts/reports/)s
    once identity resolution and multi-source aggregation are proven at
    the individual level — see the
    [OmniDevX ecosystem plan](https://github.com/plexusone/devfolio/blob/main/docs/specs/PLAN.md)
    (Phase 9). This command is not yet superseded; it's the current,
    working implementation.
