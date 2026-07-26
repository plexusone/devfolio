# TRD: DevFolio DoltDB Analytics Store

## Overview

Technical design for persisting omnidevx telemetry events into a Dolt database.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  ~/.plexusone/omnidevx/data/events/                            │
│  └── {year}/{month}/{day}/                                      │
│      ├── claude-code.jsonl  (token events)                      │
│      └── git.jsonl          (commit events)                     │
└─────────────────────┬───────────────────────────────────────────┘
                      │ devfolio db import
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│  ~/.plexusone/omnidevx/doltdb/                                  │
│  └── devx/                    (Dolt database)                   │
│      ├── token_events         (table)                           │
│      ├── git_events           (table)                           │
│      ├── import_watermarks    (table - tracks last import)      │
│      └── .dolt/               (version control)                 │
└─────────────────────┬───────────────────────────────────────────┘
                      │ dolt push
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│  DoltHub or self-hosted remote                                  │
│  └── plexusone/devx                                             │
└─────────────────────────────────────────────────────────────────┘
```

## Database Schema

### Table: token_events

Stores `ai.message.completed` events from Claude Code.

```sql
CREATE TABLE token_events (
    id VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(64) NOT NULL,
    timestamp DATETIME NOT NULL,
    
    -- Subject
    person_id VARCHAR(64) NOT NULL,
    
    -- Context
    session_id VARCHAR(64),
    workspace VARCHAR(512),
    
    -- Attributes
    model VARCHAR(64) NOT NULL,
    input_tokens BIGINT NOT NULL DEFAULT 0,
    output_tokens BIGINT NOT NULL DEFAULT 0,
    cache_read_tokens BIGINT NOT NULL DEFAULT 0,
    cache_creation_tokens BIGINT NOT NULL DEFAULT 0,
    
    -- Computed (denormalized for query performance)
    cost_usd DECIMAL(10, 4),
    
    -- Provenance
    collection_mode VARCHAR(32),
    confidence DECIMAL(3, 2),
    
    -- Import metadata
    imported_at DATETIME NOT NULL,
    source_file VARCHAR(512),
    
    INDEX idx_timestamp (timestamp),
    INDEX idx_model (model),
    INDEX idx_workspace (workspace),
    INDEX idx_person_id (person_id)
);
```

### Table: git_events

Stores `devx.change.committed` events from git history.

```sql
CREATE TABLE git_events (
    id VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(64) NOT NULL,
    timestamp DATETIME NOT NULL,
    
    -- Subject
    person_id VARCHAR(64) NOT NULL,
    
    -- Context
    repository VARCHAR(256) NOT NULL,
    workspace VARCHAR(512),
    git_branch VARCHAR(256),
    
    -- Attributes
    commit_hash VARCHAR(40) NOT NULL,
    author_email VARCHAR(256),
    insertions INT NOT NULL DEFAULT 0,
    deletions INT NOT NULL DEFAULT 0,
    files_changed INT NOT NULL DEFAULT 0,
    ai_assisted BOOLEAN NOT NULL DEFAULT FALSE,
    ai_model VARCHAR(64),
    ai_product VARCHAR(64),
    
    -- Provenance
    collection_mode VARCHAR(32),
    confidence DECIMAL(3, 2),
    
    -- Import metadata
    imported_at DATETIME NOT NULL,
    source_file VARCHAR(512),
    
    INDEX idx_timestamp (timestamp),
    INDEX idx_repository (repository),
    INDEX idx_commit_hash (commit_hash),
    INDEX idx_ai_assisted (ai_assisted)
);
```

### Table: import_watermarks

Tracks import progress for incremental imports.

```sql
CREATE TABLE import_watermarks (
    source VARCHAR(64) PRIMARY KEY,  -- 'claude-code' or 'git'
    last_file VARCHAR(512) NOT NULL, -- last imported file path
    last_timestamp DATETIME NOT NULL,
    event_count BIGINT NOT NULL,
    updated_at DATETIME NOT NULL
);
```

## Data Flow

### Historical Import

1. Scan all JSONL files in date order
2. Parse each event, map to table schema
3. Compute `cost_usd` using `omnidevx-core/report.LookupPricing()`
4. Batch INSERT (1000 rows per transaction)
5. Update watermark after each file

```go
type Importer struct {
    db        *sql.DB
    pricing   map[string]report.ModelPricing
    batchSize int
}

func (i *Importer) ImportFile(path string) error {
    // Parse JSONL
    // Map to rows
    // Batch insert
    // Update watermark
}
```

### Incremental Import

1. Read watermark for source
2. Find files newer than watermark
3. Import only new files
4. Update watermark

### Cost Calculation

Uses `omnidevx-core/report.LookupPricing()` at import time:

```go
func calculateCost(model string, input, output, cacheRead, cacheWrite int64) float64 {
    pricing, ok := report.LookupPricing(model)
    if !ok {
        return 0
    }
    return report.EstimateCost(pricing, input, output, cacheRead, cacheWrite)
}
```

## CLI Commands

### devfolio db init

Initialize Dolt database.

```bash
devfolio db init [--path ~/.plexusone/omnidevx/doltdb/devx]
```

### devfolio db import

Import events from JSONL files.

```bash
# Full historical import
devfolio db import --source ~/.plexusone/omnidevx/data/events

# Incremental (default)
devfolio db import --incremental

# Specific date range
devfolio db import --since 2026-07-01 --until 2026-07-31
```

### devfolio db query

Run SQL queries.

```bash
# Interactive SQL shell
devfolio db query

# Single query
devfolio db query "SELECT model, SUM(cost_usd) FROM token_events GROUP BY model"
```

### devfolio db sync

Push/pull from remote.

```bash
devfolio db sync push
devfolio db sync pull
```

## Dolt Integration

### Database Location

```
~/.plexusone/omnidevx/doltdb/devx/
├── .dolt/
│   ├── config.json
│   └── noms/
├── token_events.sql
└── git_events.sql
```

### Branching Strategy

- `main` — production data
- `schema/*` — schema migrations
- `import/*` — test imports before merging

### Remote Configuration

```bash
dolt remote add origin https://doltremoteapi.dolthub.com/plexusone/devx
# or
dolt remote add origin file:///backup/doltdb/devx
```

## Go Implementation

### Package Structure

```
devfolio/
├── cmd/devfolio/
│   └── db.go              # CLI commands
├── doltdb/
│   ├── schema.go          # Table definitions
│   ├── importer.go        # JSONL → Dolt import
│   ├── watermark.go       # Incremental import tracking
│   └── queries.go         # Common analytics queries
```

### Dependencies

```go
import (
    "database/sql"
    _ "github.com/dolthub/go-mysql-server" // or use dolt CLI
)
```

Options for Dolt access:
1. **Shell out to `dolt sql`** — simplest, no driver needed
2. **go-mysql-server** — embedded MySQL-compatible server
3. **dolt CLI as subprocess** — balance of simplicity and capability

Recommendation: Start with shelling out to `dolt sql` for MVP.

## Testing

### Unit Tests

- Schema validation
- JSONL parsing
- Cost calculation

### Integration Tests

- Full import of sample data
- Incremental import
- Query correctness

### Test Data

Create `testdata/events/` with sample JSONL covering:
- Multiple models
- Edge cases (missing fields, unknown models)
- Date boundaries

## Migration Strategy

### Initial Schema (v1)

Deploy tables as specified above.

### Future Migrations

Use Dolt branches:

```bash
dolt checkout -b schema/add-reasoning-tokens
dolt sql -q "ALTER TABLE token_events ADD COLUMN reasoning_tokens BIGINT DEFAULT 0"
dolt commit -m "Add reasoning_tokens column"
dolt checkout main
dolt merge schema/add-reasoning-tokens
```

## Performance Considerations

### Batch Size

- 1000 rows per INSERT for historical import
- Single-row INSERT for real-time (future)

### Indexes

Indexes on:
- `timestamp` — time-range queries
- `model` — model aggregations
- `workspace` — project-level analytics
- `repository` — per-repo queries

### Query Optimization

For large date ranges, use:
```sql
SELECT ... WHERE timestamp >= ? AND timestamp < ?
```

Not:
```sql
SELECT ... WHERE DATE(timestamp) = ?
```

## Security

- Database is local-only by default
- Remote push requires authentication (DoltHub or SSH)
- No secrets stored in database (workspace paths only)
