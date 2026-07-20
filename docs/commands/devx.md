# DevX Commands

Read collected [OmniDevX](https://github.com/plexusone/omnidevx-core)
events (Claude Code, Codex CLI, git, GitHub) from the local store and
build period reports and dashboards from them.

Requires events already collected into the local store
(`~/.plexusone/omnidevx/data/` by default) via the `omnidevx-core`
providers — this command group only reads and reports, it does not
collect. See `omnidevx-core`'s
[Getting Started](https://plexusone.github.io/omnidevx-core/getting-started/)
for collecting events into the store first.

## `devfolio devx dashboard`

```bash
devfolio devx dashboard --person person:jane [flags]
```

Builds a `DeveloperPeriodReport` via `omnidevx-core/report`, then exports
it as a [dashforge](https://github.com/plexusone/dashforge) dashboard: 8
headline metric tiles (sessions, prompts, commits + AI-assisted %, tool
calls + failure rate, cost, coverage), a daily commits/prompts chart, a
daily cost chart, and a source-coverage table.

| Flag | Description |
|------|-------------|
| `--person` | Canonical personId to report on (required) |
| `--days` | Number of days ending today to report on (default `30`) |
| `--store-dir` | OmniDevX store directory (default: `~/.plexusone/omnidevx/data`) |
| `-o`, `--output` | Output file (default: stdout) |

### Examples

```bash
# Last 30 days, written to stdout
devfolio devx dashboard --person person:jane

# Last 7 days, written to a file
devfolio devx dashboard --person person:jane --days 7 -o dashboard.json
```

## Viewing the dashboard

The output is a single portable JSON file — three ways to view it:

1. **dashforge's static viewer** — `viewer/index.html?dashboard=<file>` in
   [dashforge](https://github.com/plexusone/dashforge).
2. **Validate it** — `dashforge validate dashboard.json` (via dashforge's
   CLI) checks it against the `dashboardir` schema.
3. **VisionStudio's DevX panel** — write the output to
   `~/.plexusone/omnidevx/dashboard.json` and
   [VisionStudio](https://github.com/ProductBuildersHQ/visionstudio)'s
   daemon serves it at `GET /api/devx/dashboard`, rendered in its sidebar's
   DevX → Usage Dashboard view:

   ```bash
   devfolio devx dashboard --person person:jane -o ~/.plexusone/omnidevx/dashboard.json
   ```

   VisionStudio only ever reads this already-generated file — it never
   queries the OmniDevX event store directly. DevFolio decides what's
   safe to show; VisionStudio is a read-only consumer.

## Data quality

If the report period includes GitHub `devx.profile.snapshot` or
`devx.contribution.snapshot` events (period-total snapshots, not daily
deltas), the dashboard's `quality.warnings` will note they aren't yet
merged into the report — a known, documented gap
(see `omnidevx-core`'s [Period Reports](https://plexusone.github.io/omnidevx-core/concepts/reports/)),
not a silent undercount.
