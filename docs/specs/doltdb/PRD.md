# PRD: DevFolio DoltDB Analytics Store

## Overview

Persist developer telemetry events (token spend, git commits) from JSONL files into a Dolt database for SQL-based analytics, versioned backups, and cross-machine synchronization.

## Problem Statement

Current state:
- 13 months of telemetry data in `~/.plexusone/omnidevx/data/events/`
- 250 JSONL files, ~99MB total
- Date-partitioned: `{year}/{month}/{day}/{source}.jsonl`
- Sources: `claude-code.jsonl` (token events), `git.jsonl` (commit events)

Pain points:
1. **No SQL analytics** — aggregations require custom Go code per query
2. **No backup strategy** — local-only, no versioning, no sync
3. **Schema evolution** — adding fields requires migration tooling
4. **Cross-machine** — no way to merge data from multiple workstations

## Goals

1. **SQL analytics** — query token spend by model/day/project, commit velocity, cost trends
2. **Version-controlled data** — Dolt branches for schema changes, rollback bad imports
3. **Incremental sync** — import new events without re-processing history
4. **Remote backup** — push to DoltHub or self-hosted remote

## Non-Goals

- Real-time streaming (batch import is sufficient)
- Multi-tenant (single developer use case)
- Web UI (CLI + SQL is sufficient)

## User Stories

### US-1: Import historical data
As a developer, I want to import my 13 months of JSONL events into Dolt so I can query them with SQL.

### US-2: Daily incremental import
As a developer, I want new events to be imported daily without re-processing historical data.

### US-3: Query token spend
As a developer, I want to run SQL queries like:
```sql
SELECT model, SUM(input_tokens + output_tokens) as total_tokens, 
       SUM(cost_usd) as total_cost
FROM token_events
WHERE timestamp >= '2026-01-01'
GROUP BY model
ORDER BY total_cost DESC;
```

### US-4: Query by project
As a developer, I want to see token spend per workspace/repository:
```sql
SELECT workspace, SUM(cost_usd) as cost
FROM token_events
WHERE timestamp BETWEEN '2026-07-01' AND '2026-07-31'
GROUP BY workspace
ORDER BY cost DESC;
```

### US-5: Backup to remote
As a developer, I want to push my Dolt database to a remote for backup and cross-machine sync.

### US-6: Schema evolution
As a developer, I want to add new columns (e.g., `reasoning_tokens`) without breaking existing queries.

## Success Metrics

| Metric | Target |
|--------|--------|
| Historical import time | < 5 minutes for 99MB |
| Daily incremental import | < 30 seconds |
| Query latency (aggregations) | < 1 second |
| Storage overhead vs JSONL | < 2x |

## Dependencies

- Dolt CLI (`dolt`) installed
- omnidevx-core for event schema definitions
- devfolio CLI for import commands

## Risks

| Risk | Mitigation |
|------|------------|
| Dolt not installed | Document install, check in CLI |
| Schema changes break imports | Use Dolt branching for migrations |
| Large data growth | Partition by year, archive old data |

## Timeline

- Phase 1: Schema design + historical import (1 week)
- Phase 2: Incremental import + CLI commands (1 week)
- Phase 3: Remote sync + backup automation (1 week)
