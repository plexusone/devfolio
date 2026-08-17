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
it as a [uiforge](https://github.com/plexusone/uiforge) dashboard: 8
headline metric tiles (sessions, prompts, commits + AI-assisted %, tool
calls + failure rate, cost, coverage), a daily commits/prompts chart, a
daily cost chart, and a source-coverage table.

| Flag | Description |
|------|-------------|
| `--person` | Canonical personId to report on (required) |
| `--days` | Number of days ending today to report on (default `30`, ignored with `--period`) |
| `--store-dir` | OmniDevX store directory (default: `~/.plexusone/omnidevx/data`) |
| `-o`, `--output` | Output file (default: stdout, or the standard reports path with `--period`) |
| `--period` | Generate a calendar period report instead of a rolling window: `weekly`, `monthly`, or `quarterly` |
| `--for` | Anchor date for `--period`, `YYYY-MM-DD` (default: today) |

### Examples

```bash
# Last 30 days, written to stdout
devfolio devx dashboard --person person:jane

# Last 7 days, written to a file
devfolio devx dashboard --person person:jane --days 7 -o dashboard.json
```

## Period reports (`--period`)

`--period weekly|monthly|quarterly` builds a calendar-aligned report instead
of a rolling window — the period containing `--for` (default: today), with
weeks always Monday-Sunday. Monthly reports add a weekly per-model token/cost
breakdown; quarterly reports add both weekly and monthly per-model
breakdowns, each rendered as donut and stacked-bar chart widgets in the
resulting dashboard.

```bash
# Current calendar month, written to
# ~/.plexusone/omnidevx/reports/monthly/2026-08.json
devfolio devx dashboard --person person:jane --period monthly

# A specific past month
devfolio devx dashboard --person person:jane --period monthly --for 2026-07-15

# Current calendar quarter (embeds monthly and weekly model breakdowns)
devfolio devx dashboard --person person:jane --period quarterly
```

Unless `-o`/`--output` is given, period reports are written to
`~/.plexusone/omnidevx/reports/{weekly,monthly,quarterly}/{label}.json` —
the exact path VisionStudio's daemon reads from (see below).

## Viewing the dashboard

The output is a single portable JSON file:

1. **uiforge's static viewer** — `viewer/index.html?dashboard=<file>` in
   [uiforge](https://github.com/plexusone/uiforge) (v0.5.0+ has a light/dark
   theme toggle and caps numeric display in tooltips, axis labels, and
   metric tiles to 2 decimal places).
2. **Validate it** — `uiforge validate dashboard.json` (via uiforge's
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
4. **VisionStudio's period selector** — period reports written to the
   standard reports path (the `--period` default) are listed at
   `GET /api/devx/periods` and served individually at
   `GET /api/devx/reports/{periodType}/{label}` — VisionStudio's UI has a
   period dropdown that switches between them, with bar and donut chart
   renderers for the model breakdowns.

## Data quality

If the report period includes GitHub `devx.profile.snapshot` or
`devx.contribution.snapshot` events (period-total snapshots, not daily
deltas), the dashboard's `quality.warnings` will note they aren't yet
merged into the report — a known, documented gap
(see `omnidevx-core`'s [Period Reports](https://plexusone.github.io/omnidevx-core/concepts/reports/)),
not a silent undercount.
