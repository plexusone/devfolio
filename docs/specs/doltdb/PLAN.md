# PLAN: DevFolio DoltDB Analytics Store

## Implementation Plan

### Phase 1: Schema & Historical Import (Week 1)

#### Step 1.1: Create doltdb package structure

```
devfolio/
├── doltdb/
│   ├── doltdb.go          # DB connection, init
│   ├── schema.go          # CREATE TABLE statements
│   ├── schema_test.go
│   ├── importer.go        # JSONL parser, batch insert
│   ├── importer_test.go
│   └── testdata/
│       └── events/        # Sample JSONL for tests
```

#### Step 1.2: Implement schema.go

Define SQL schema as Go constants:

```go
package doltdb

const SchemaVersion = "1.0.0"

const CreateTokenEvents = `
CREATE TABLE IF NOT EXISTS token_events (
    id VARCHAR(255) PRIMARY KEY,
    ...
);`

const CreateGitEvents = `...`

const CreateWatermarks = `...`

func InitSchema(dbPath string) error {
    // Shell out to: dolt sql -q "CREATE TABLE ..."
}
```

#### Step 1.3: Implement importer.go

```go
type TokenEvent struct {
    ID                   string
    Timestamp            time.Time
    PersonID             string
    SessionID            string
    Workspace            string
    Model                string
    InputTokens          int64
    OutputTokens         int64
    CacheReadTokens      int64
    CacheCreationTokens  int64
    CostUSD              float64
}

func ParseClaudeCodeJSONL(path string) ([]TokenEvent, error)
func ImportTokenEvents(dbPath string, events []TokenEvent) error
```

#### Step 1.4: Add CLI commands

```go
// cmd/devfolio/db.go
var dbCmd = &cobra.Command{Use: "db"}
var dbInitCmd = &cobra.Command{Use: "init"}
var dbImportCmd = &cobra.Command{Use: "import"}
```

#### Step 1.5: Historical import

```bash
devfolio db init
devfolio db import --source ~/.plexusone/omnidevx/data/events
```

Acceptance criteria:
- [ ] All 250 JSONL files imported
- [ ] ~99MB data loads in < 5 minutes
- [ ] `dolt sql -q "SELECT COUNT(*) FROM token_events"` returns expected count

### Phase 2: Incremental Import & Queries (Week 2)

#### Step 2.1: Watermark tracking

```go
type Watermark struct {
    Source        string    // "claude-code" or "git"
    LastFile      string
    LastTimestamp time.Time
    EventCount    int64
}

func GetWatermark(dbPath, source string) (*Watermark, error)
func UpdateWatermark(dbPath string, w *Watermark) error
```

#### Step 2.2: Incremental import logic

```go
func ImportIncremental(dbPath, eventsDir string) error {
    // 1. Get watermark
    // 2. Find files newer than watermark
    // 3. Import only new files
    // 4. Update watermark
}
```

#### Step 2.3: Query command

```go
var dbQueryCmd = &cobra.Command{
    Use:   "query [sql]",
    Short: "Run SQL query against DoltDB",
    Run: func(cmd *cobra.Command, args []string) {
        if len(args) == 0 {
            // Interactive mode
            runInteractiveSQL(dbPath)
        } else {
            // Single query
            runQuery(dbPath, args[0])
        }
    },
}
```

#### Step 2.4: Pre-built analytics queries

```go
// doltdb/queries.go
func TokenSpendByModel(dbPath string, since, until time.Time) ([]ModelSpend, error)
func TokenSpendByWorkspace(dbPath string, since, until time.Time) ([]WorkspaceSpend, error)
func DailyTokenTrend(dbPath string, since, until time.Time) ([]DailySpend, error)
```

### Phase 3: Remote Sync & Automation (Week 3)

#### Step 3.1: Remote configuration

```go
var dbRemoteCmd = &cobra.Command{Use: "remote"}
var dbRemoteAddCmd = &cobra.Command{Use: "add [name] [url]"}
```

#### Step 3.2: Sync commands

```go
var dbSyncCmd = &cobra.Command{Use: "sync"}
var dbPushCmd = &cobra.Command{Use: "push"}
var dbPullCmd = &cobra.Command{Use: "pull"}
```

#### Step 3.3: Automated daily import (optional)

Create launchd plist or cron job:

```bash
# Daily at midnight
0 0 * * * devfolio db import --incremental
```

## File-by-File Implementation Order

1. `doltdb/doltdb.go` — DB path resolution, dolt CLI wrapper
2. `doltdb/schema.go` — Table DDL, InitSchema()
3. `doltdb/schema_test.go` — Schema creation tests
4. `doltdb/importer.go` — JSONL parsing, batch insert
5. `doltdb/importer_test.go` — Import tests with sample data
6. `doltdb/watermark.go` — Incremental import tracking
7. `cmd/devfolio/db.go` — CLI commands (init, import, query)
8. `doltdb/queries.go` — Pre-built analytics queries
9. `cmd/devfolio/db_sync.go` — Remote sync commands

## Testing Strategy

### Unit Tests

| Test | File | Coverage |
|------|------|----------|
| Schema DDL valid | schema_test.go | All CREATE statements |
| JSONL parsing | importer_test.go | Token events, git events |
| Cost calculation | importer_test.go | All model tiers |
| Watermark CRUD | watermark_test.go | Get, Update, edge cases |

### Integration Tests

| Test | Description |
|------|-------------|
| Full import | Import testdata/events/, verify counts |
| Incremental | Import, add file, re-import, verify delta |
| Query accuracy | Run queries, compare to manual calculation |

### Test Data

Create `doltdb/testdata/events/2026/07/01/`:
- `claude-code.jsonl` — 100 token events, multiple models
- `git.jsonl` — 50 commit events, mix of AI-assisted

## Dependencies

### Required

- `dolt` CLI (document install: `brew install dolt`)
- `omnidevx-core` for pricing lookup

### Optional

- DoltHub account for remote backup
- `github.com/dolthub/go-mysql-server` for embedded mode (future)

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Dolt not installed | Check in `db init`, print install instructions |
| Import fails midway | Use transactions per file, resume from watermark |
| Schema changes | Branch before ALTER, test, merge |
| Disk space | Monitor growth, archive old data to separate DB |

## Success Criteria

### Phase 1 Complete

- [ ] `devfolio db init` creates Dolt database with schema
- [ ] `devfolio db import` loads all historical JSONL
- [ ] `dolt sql -q "SELECT ..."` returns valid results
- [ ] Import completes in < 5 minutes

### Phase 2 Complete

- [ ] Incremental import only processes new files
- [ ] `devfolio db query` runs SQL queries
- [ ] Pre-built queries return correct aggregations

### Phase 3 Complete

- [ ] `devfolio db sync push` backs up to remote
- [ ] `devfolio db sync pull` restores from remote
- [ ] Documentation for automated daily import
