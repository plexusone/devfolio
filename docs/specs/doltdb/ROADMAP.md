# ROADMAP: DevFolio DoltDB Analytics Store

## Phase 1: Foundation (Week 1)

Schema design, historical import, basic CLI.

| RMI | Title | Status | Description |
|-----|-------|--------|-------------|
| RMI-DEVFOLIO-050 | Create doltdb package | Not Started | Package structure, dolt CLI wrapper |
| RMI-DEVFOLIO-051 | Define token_events schema | Not Started | CREATE TABLE with indexes |
| RMI-DEVFOLIO-052 | Define git_events schema | Not Started | CREATE TABLE with indexes |
| RMI-DEVFOLIO-053 | Implement JSONL parser | Not Started | Parse claude-code.jsonl, git.jsonl |
| RMI-DEVFOLIO-054 | Implement batch importer | Not Started | Batch INSERT with transactions |
| RMI-DEVFOLIO-055 | Add db init command | Not Started | `devfolio db init` |
| RMI-DEVFOLIO-056 | Add db import command | Not Started | `devfolio db import --source` |
| RMI-DEVFOLIO-057 | Historical import validation | Not Started | Import 13 months, verify counts |

## Phase 2: Incremental & Queries (Week 2)

Incremental import, SQL queries, analytics.

| RMI | Title | Status | Description |
|-----|-------|--------|-------------|
| RMI-DEVFOLIO-058 | Implement watermark tracking | Not Started | Track last imported file/timestamp |
| RMI-DEVFOLIO-059 | Incremental import logic | Not Started | Import only new files since watermark |
| RMI-DEVFOLIO-060 | Add db query command | Not Started | `devfolio db query [sql]` |
| RMI-DEVFOLIO-061 | Pre-built analytics: by model | Not Started | TokenSpendByModel() |
| RMI-DEVFOLIO-062 | Pre-built analytics: by workspace | Not Started | TokenSpendByWorkspace() |
| RMI-DEVFOLIO-063 | Pre-built analytics: daily trend | Not Started | DailyTokenTrend() |
| RMI-DEVFOLIO-064 | Quarterly report from DoltDB | Not Started | Alternative data source for reports |

## Phase 3: Sync & Automation (Week 3)

Remote backup, cross-machine sync, automation.

| RMI | Title | Status | Description |
|-----|-------|--------|-------------|
| RMI-DEVFOLIO-065 | Add db remote command | Not Started | `devfolio db remote add` |
| RMI-DEVFOLIO-066 | Add db sync push command | Not Started | Push to DoltHub/remote |
| RMI-DEVFOLIO-067 | Add db sync pull command | Not Started | Pull from remote |
| RMI-DEVFOLIO-068 | Document automated import | Not Started | launchd/cron setup guide |
| RMI-DEVFOLIO-069 | Multi-machine merge strategy | Not Started | Handle conflicts from multiple sources |

## Phase 4: Advanced Analytics (Future)

Extended analytics, visualization integration.

| RMI | Title | Status | Description |
|-----|-------|--------|-------------|
| RMI-DEVFOLIO-070 | Cost forecasting | Not Started | Project monthly spend from trends |
| RMI-DEVFOLIO-071 | Model efficiency metrics | Not Started | Output/input ratio, cache hit rate |
| RMI-DEVFOLIO-072 | Project attribution | Not Started | Map workspaces to project names |
| RMI-DEVFOLIO-073 | DoltDB → quarterly report | Not Started | Use DoltDB as primary data source |
| RMI-DEVFOLIO-074 | Export to dashboard IR | Not Started | Generate uiforge dashboards from SQL |

## Dependencies

### External

- Dolt CLI (`brew install dolt`)
- DoltHub account (optional, for remote backup)

### Internal

- omnidevx-core: `report.LookupPricing()` for cost calculation
- devfolio: existing CLI infrastructure (cobra)

## Milestones

| Milestone | Target Date | Deliverable |
|-----------|-------------|-------------|
| M1: Historical Import | Week 1 | All 13 months imported, queryable |
| M2: Daily Automation | Week 2 | Incremental import working |
| M3: Remote Backup | Week 3 | Data backed up to DoltHub |
| M4: Quarterly Integration | Week 4 | Reports pull from DoltDB |

## Notes

### Why Dolt?

1. **Git-like versioning** — branch, commit, merge data changes
2. **SQL interface** — standard MySQL-compatible queries
3. **Schema evolution** — ALTER TABLE with version history
4. **Remote sync** — push/pull like git for backups
5. **Offline-first** — works locally, syncs when connected

### Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| SQLite | Simple, embedded | No versioning, no remote sync |
| PostgreSQL | Full-featured | Heavyweight for single-user |
| DuckDB | Fast analytics | No versioning, no remote |
| Plain files | Already have them | No SQL, hard to query |

### Storage Estimate

| Data | Current | 1 Year | 3 Years |
|------|---------|--------|---------|
| JSONL | 99 MB | ~100 MB | ~300 MB |
| DoltDB | ~150 MB | ~200 MB | ~500 MB |

Dolt adds ~50% overhead for versioning metadata — acceptable for the benefits.
