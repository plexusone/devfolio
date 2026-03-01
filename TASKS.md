# DevFolio Tasks

Prioritized implementation tasks organized by phase.

## Phase 1: Individual Contributor Profiles (v0.1.0) ✅

### Completed

- [x] CLI scaffolding with cobra (root, version, contributor, team commands)
- [x] `contributor profile` command implementation
- [x] GitHub API client for user activity
- [x] Profile types (Profile, RepoContrib, ContributorStats)
- [x] AI collaboration tracking types (AICollabStats, AIToolStat, KnownAITool)
- [x] Co-author parsing from commit messages
- [x] Known AI tools registry (Claude Code, Copilot, Gemini CLI, Cursor, Aider)
- [x] Activity heatmap data aggregation
- [x] CI/CD workflows (go-ci, go-lint, go-sast-codeql)
- [x] Dependabot configuration
- [x] CHANGELOG.json and CHANGELOG.md
- [x] README with usage documentation

---

## Phase 2: Team Velocity (v0.2.0)

### 2.1 Core Team Package

- [ ] Create `team/types.go` with team metrics types
  - [ ] `TeamVelocity` - aggregate team metrics
  - [ ] `ProjectMetrics` - per-project breakdown
  - [ ] `TimeSeries` - time-based data points
  - [ ] `CategoryBreakdown` - features/fixes/improvements

- [ ] Create `team/client.go` with velocity calculation
  - [ ] `LoadPortfolio()` - load structured-changelog portfolio
  - [ ] `CalculateVelocity()` - compute team metrics
  - [ ] `GroupByPeriod()` - day/week/month aggregation

### 2.2 Team Velocity Command

- [ ] Implement `cmd/devfolio/team_velocity.go`
  - [ ] Load portfolio JSON input
  - [ ] Parse `--granularity` flag (day/week/month)
  - [ ] Parse `--since` and `--until` date filters
  - [ ] Output velocity JSON

### 2.3 Testing

- [ ] `team/types_test.go` - type tests
- [ ] `team/client_test.go` - calculation tests
- [ ] Integration test with sample portfolio

### 2.4 Documentation

- [ ] Update README with team velocity examples
- [ ] Update CHANGELOG.json for v0.2.0

---

## Phase 3: Output Formats (v0.3.0)

### 3.1 Dashboard Export

- [ ] Create `output/dashboard/types.go`
  - [ ] Widget types (metric, chart, table, heatmap)
  - [ ] Dashboard layout types
  - [ ] Dashforge-compatible JSON structure

- [ ] Create `output/dashboard/export.go`
  - [ ] `ExportContributorDashboard()` - from Profile
  - [ ] `ExportTeamDashboard()` - from TeamVelocity
  - [ ] Widget generation helpers

- [ ] Add `--dashboard` flag to commands
  - [ ] `contributor profile --dashboard`
  - [ ] `team velocity --dashboard`

### 3.2 Markdown Export

- [ ] Create `output/markdown/types.go`
  - [ ] Template types
  - [ ] Section configuration

- [ ] Create `output/markdown/render.go`
  - [ ] `RenderContributorProfile()` - profile to markdown
  - [ ] `RenderTeamVelocity()` - velocity to markdown
  - [ ] Table generation helpers

- [ ] Add `--markdown` flag to commands

### 3.3 Static Site Generation

- [ ] Create `output/site/types.go`
  - [ ] Site configuration
  - [ ] Page templates

- [ ] Create `output/site/generate.go`
  - [ ] HTML template embedding
  - [ ] Chart.js/ECharts integration
  - [ ] Static asset bundling

- [ ] Add `--site` flag to commands

### 3.4 Testing

- [ ] Dashboard export tests
- [ ] Markdown rendering tests
- [ ] Site generation tests

---

## Phase 4: Data Sources (v0.4.0)

### 4.1 Git History Source

- [ ] Create `datasource/git/types.go`
  - [ ] Commit analysis types
  - [ ] File change statistics

- [ ] Create `datasource/git/parser.go`
  - [ ] Parse local git history
  - [ ] Extract commit metadata
  - [ ] Calculate file churn metrics

### 4.2 Changelog Source

- [ ] Create `datasource/changelog/types.go`
  - [ ] Changelog entry types
  - [ ] Category mapping

- [ ] Create `datasource/changelog/loader.go`
  - [ ] Load CHANGELOG.json files
  - [ ] Parse portfolio manifests
  - [ ] Aggregate changelog data

### 4.3 GitHub Source Enhancements

- [ ] Create `datasource/github/types.go`
  - [ ] Enhanced activity types
  - [ ] Rate limit handling

- [ ] Move contributor client to `datasource/github/`
- [ ] Add GraphQL support for contribution calendar

---

## Phase 5: Advanced Features (v0.5.0+)

### 5.1 Comparative Analysis

- [ ] `contributor compare` command
- [ ] `team compare` command
- [ ] Side-by-side metrics
- [ ] Diff visualization

### 5.2 Trend Detection

- [ ] Velocity trend analysis
- [ ] Anomaly detection
- [ ] Seasonal pattern identification

### 5.3 GitHub Actions

- [ ] Create `plexusone/devfolio-action`
- [ ] Scheduled profile updates
- [ ] PR comment integration

---

## Backlog (Unprioritized)

- [ ] Configuration file support (`.devfolio.yaml`)
- [ ] Multiple output formats in single run
- [ ] Custom AI tool registry via config
- [ ] Export to CSV/Excel
- [ ] Webhook notifications
- [ ] Plugin system for custom data sources
- [ ] GitLab/Bitbucket support
- [ ] MkDocs documentation site

---

## Priority Legend

| Priority | Description |
|----------|-------------|
| P0 | Critical for release |
| P1 | Important, include if time permits |
| P2 | Nice to have |
| P3 | Future consideration |

## Current Focus

**Next milestone:** v0.2.0 (Team Velocity)

**Immediate tasks:**
1. `team/types.go` - Define team metrics types
2. `team/client.go` - Implement velocity calculation
3. `team_velocity.go` - Complete CLI command
4. Tests and documentation
