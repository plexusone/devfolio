# TRD: OmniDevX Ecosystem — Technical Requirements

**Scope:** Technical architecture for the OmniDevX domain across `omnidevx-core`, `omnidevx`, `omni-github`, and `devfolio`. Product context in [PRD.md](PRD.md); sequencing in [PLAN.md](PLAN.md) and [ROADMAP.md](ROADMAP.md).

## 1. Repository and Module Layout

One repository = one Go module = multiple packages. Repositories mark ownership, dependency, and release boundaries; packages mark code and domain boundaries.

```text
plexusone/omnidevx-core     Canonical IR, interfaces, aggregation, thin
                            providers (claudecode, git, survey,
                            structuredchangelog, genericotel); space/ and
                            dora/ analytics incubate here pre-extraction
plexusone/omni-openai       Existing thick provider repo; adds omnidevx/
                            package for the Codex CLI collector (needs a
                            SQLite driver — too heavy for core)
plexusone/omni-github       Thick GitHub provider repo; adds omnidevx/ package
                            alongside existing omnistorage/, omniskill/
plexusone/omnidevx          Batteries-included bundle; created at M5 when
                            there are thick providers to compose; space/ and
                            dora/ move here from core at that point
plexusone/devfolio          Portfolio/profile application (consumer)
```

Placement rule (per 2026-07-16 decision): a collector lives in `omnidevx-core/providers/` when it uses no vendor SDK and no heavy dependency (we build the structs ourselves); it lives in the vendor's existing `omni-<vendor>/omnidevx/` package when it needs an official SDK or a large dependency. Applied: Claude Code parsing is stdlib-only JSONL → core; Codex parsing requires SQLite → `omni-openai/omnidevx`. `omni-anthropic/omnidevx` is the promotion target if the Claude Code collector ever needs `anthropic-sdk-go` (e.g., API-side usage/cost lookups).

Future slots (not initial deliverables): `omni-gitlab/omnidevx`, `omni-atlassian/omnidevx`, CI-vendor providers case-by-case.

### Dependency rules

- Provider domain packages depend inward on a shared provider client layer (e.g., `omni-github/githubclient`) and outward on their domain core (`omnidevx-core`) — never sideways on sibling domain packages.
- Analytics packages (`omnidevx/space`, `omnidevx/dora`) depend only on `omnidevx-core` — never on any provider.
- DevFolio depends on `omnidevx` (batteries-included) and orchestrates; it owns no collection logic.
- `omnidevx-core` stays dependency-light (stdlib + minimal utilities). A provider is promoted out of core into its own repo only when it needs a substantial third-party SDK, heavy auth/API lifecycle, or independent releases.

### GitHub layering

```text
google/go-github → grokify/gogithub → omni-github/githubclient → omni-github/omnidevx → omnidevx-core events
```

`gogithub` keeps generic GitHub ergonomics (and gains `RepositoryAdoptionSnapshot`); `omni-github/omnidevx` does GitHub-to-DevX semantic normalization only.

## 2. Omni Pattern Conventions (from omnillm-core)

`omnidevx-core` follows the established registration pattern:

- Root package `omnidevx` at repo root; interfaces in a `provider/` (or equivalent) sub-package with type aliases re-exported from the root.
- **Constructor injection is the primary composition mechanism** (`omnidevx.New(claudecode.New(...), codex.New(...), ...)`). Unlike omnillm, no OmniDevX source currently has both a thin and thick implementation, so the priority-override registry (`PriorityThin`/`PriorityThick`) is deferred until a real override need appears. If added later, follow `omnillm-core/registry.go` exactly.
- Typed `ProviderName` string constants; `ProviderConfig` / functional options.
- Thin implementations under `providers/<name>/`.
- The batteries-included `omnidevx` repo only re-exports core aliases and composes providers; its creation is deferred until there is more than core to bundle (see PLAN).
- Standard repo scaffolding: `CHANGELOG.{md,json}`, `ROADMAP.{md,json}`, `mkdocs.yml`, `docs/`, go-ci/go-lint/go-sast badges.

Reference implementations: `omnillm-core/registry.go`, `omnillm-core/provider/interface.go`, `omnillm/omnillm.go`.

## 3. Core Contracts (omnidevx-core)

### Collector

```go
type Collector interface {
    Source() SourceDescriptor
    Collect(ctx context.Context, req CollectRequest) (*CollectionResult, error)
}

type CollectionResult struct {
    Source      SourceDescriptor `json:"source"`
    Subject     SubjectRef       `json:"subject"`
    Period      Period           `json:"period"`
    Events      []Event          `json:"events"`
    Diagnostics []Diagnostic     `json:"diagnostics,omitempty"`
    CollectedAt time.Time        `json:"collectedAt"`
}
```

### Canonical Event

```go
type Event struct {
    ID         string         `json:"id"`
    Type       EventType      `json:"type"`
    Timestamp  time.Time      `json:"timestamp"`
    Subject    SubjectRef     `json:"subject"`
    Source     Source         `json:"source"`     // provider, product, version
    Context    EventContext   `json:"context"`    // sessionId, promptId, repository, workspace
    Attributes map[string]any `json:"attributes,omitempty"`
    Provenance Provenance     `json:"provenance"` // collectionMode, confidence
}
```

- Schema ID: `omnidevx.event/v1`.
- `Provenance.CollectionMode` ∈ `history | otel | hooks | api | survey`; `Confidence` ∈ [0,1]. Historical reconstruction is inherently less certain than observed events and must say so.
- Event-type namespaces: `ai.*` for agent-session events (`ai.session.started`, `ai.prompt.submitted`, `ai.tool.completed`, `ai.patch.generated/applied`), `devx.*` for work semantics (`devx.change.committed`, `devx.change.integrated`, `devx.review.requested/completed`, `devx.verification.completed`, `devx.delivery.deployed`, `devx.outcome.completed`, `devx.change.delivered`, `devx.adoption.*`, `devx.contribution.*`).
- Source identifiers stay canonical in the IR (`{"provider": "anthropic", "product": "claude-code"}`) even though Go package names are compressed (`providers/claudecode`).

### Identity

Canonical person identity is a `personId`, never a GitHub username:

```json
{
  "personId": "person:01J...",
  "identities": [
    { "type": "github", "value": "grokify" },
    { "type": "git_email", "valueHash": "sha256:..." },
    { "type": "local_account", "deviceId": "device:mac-studio", "value": "john" }
  ]
}
```

This prepares for team aggregation without conflating accounts with people. Git emails may be stored hashed.

### Period Aggregation and DeveloperPeriodReport

```go
type DeveloperPeriodReport struct {
    SchemaVersion string           // omnidevx.developer-period/v1
    Subject       Subject
    Period        Period           // start, end, timezone, granularity
    Sources       []SourceCoverage // product, sessions, coverage, collectionModes
    Metrics       MetricSet        // combined + bySource
    Quality       DataQuality      // coverageScore, warnings
}
```

Rules:

- Raw events / daily summaries are the durable store; **week** is the default user-visible period; **month** is the canonical reporting cadence; quarterly/annual are synthesized from monthly. Never aggregate into months and discard daily resolution.
- `Metrics` retains both `combined` and `bySource` views. Some metrics combine safely (sessions, cost, tokens with model retained); others do not (acceptance rate only with matching definitions, AI LOC not reliably). The safe-to-combine table from the ideation doc governs.
- Every metric value carries `Measurement{Kind: observed|estimated, Source/Method, Confidence}`.
- **Rollup path:** the same report shape scales upward — `DeveloperPeriodReport` values aggregate into a future `TeamPeriodReport` (and organization reports) over the same period grid. Team velocity is computed from individual period reports, never collected separately; this is why person-level identity resolution and daily-resolution storage are mandatory from v0.1.

### AI-Code Lifecycle

Measure each stage separately; never collapse them:

```text
generated → applied → retained → committed → merged → verified → operated → delivered
```

Historical logs recover stages 4–8 well; hooks/OTel are required for 1–3.

## 3a. Local Event Store

Raw canonical events are durable and reprocessable — metric formulas change; recollection should never be required.

- **Location:** `~/.plexusone/omnidevx/data/` (single PlexusOne home; room for sibling domains later).
- **Layout:** daily JSONL per source: `events/YYYY/MM/DD/<product>.jsonl`, plus `reports/<subject>/<period>.json` for generated period reports and `identity/` for the person/identity map. Plain files first; add a SQLite index only when query patterns demand it.
- **Privacy rule (collection side):** events capture **metadata only by default** — event types, durations, counts, models, token/cost figures, repo/branch identifiers, file paths. Never prompt text, response text, or file contents unless a user explicitly opts in per collector. This mirrors the publication-side disclosure model: collect broadly *in metadata*, never hoard content.
- **Team storage:** individual raw events never leave the developer's machine. Team rollups consume shared `DeveloperPeriodReport` artifacts (already metadata-level), subject to team privacy rules (aggregation minimums, member-detail visibility) defined in Plan Phase 9.

## 4. Collection Modes per Provider

Each provider supports up to three modes:

```text
providers/<name>/
├── collector.go     Collector implementation + Options
├── history.go       Retrospective importer (local session files)
├── otel.go          Native OTel ingestion
├── hooks.go         (claudecode only) lifecycle-hook enrichment
├── normalize.go     Native records → canonical events
└── types.go         Provider-native types (internal where possible)
```

| Provider | history | otel | hooks | Notes |
|----------|---------|------|-------|-------|
| claudecode | yes | yes | yes (selective) | In core. Per-project session JSONL under `~/.claude/projects/` — typed records (user/assistant/progress/system) with sessionId, cwd, gitBranch, model, and full token usage incl. cache tiers (verified 2026-07-16). Hooks limited to: UserPromptSubmit, PostToolUse(+Failure), PermissionRequest, TaskCompleted, Stop, SessionEnd |
| codex | yes | yes | no | In `omni-openai/omnidevx`. Two historical formats (verified 2026-07-16): legacy rollout JSONL under `~/.codex/sessions/YYYY/MM/DD/` and current SQLite (`state_5.sqlite`/`logs_2.sqlite`: threads, agent_jobs, thread_spawn_edges) — importer must read both. Internal formats — treat as unstable; prefer OTel prospectively |
| git | yes | — | — | In core, built on `grokify/gogit` (gitscan renamed into a generic git base library — discovery, log+trailer parsing, branch/origin; CLI kept at `cmd/gitscan`). Emits `devx.change.committed`; canonical `KnownAITools` AI-attribution registry lives here (ported from devfolio) |
| survey | — | — | — | 1-question session/day micro-surveys; the only reliable Satisfaction source |
| structuredchangelog | yes | — | — | Wraps `grokify/structured-changelog/changelog`; emits `devx.change.delivered` with category → value-class mapping |
| genericotel | — | yes | — | Generic OTLP ingestion for tools without a dedicated provider |

Relationship to OmniObserve: `omniobserve` owns operational telemetry (traces/metrics/logs/GenAI spans); `omnidevx` owns developer-work semantics. OmniDevX may consume OTel data but must not make OTel its core abstraction.

## 5. Analytics Engines (`space`/`dora` packages)

Frameworks are profiles over canonical events, packaged as `space` and `dora` packages (not separate repos, not `ai-space`/`ai-dora` packages). They incubate in `omnidevx-core` while the IR is still churning and move to the batteries-included `omnidevx` repo at M5:

```go
report, err := space.Calculate(ctx, space.Request{
    Profile: space.AIAugmented, // or space.Traditional
    Subject: person,
    Period:  week,
    Events:  events,
})

type SPACEReport struct {
    Dimensions  SPACEDimensions
    AIExtension *AISPACEExtension `json:"aiExtension,omitempty"`
}

type DORAReport struct {
    Metrics     DORAMetrics
    AIExtension *AIDORAExtension `json:"aiExtension,omitempty"`
}
```

Constraints:

- SPACE subject: person or delivery system. DORA subject: delivery system only (repo/service/product) — reject person-scoped DORA requests.
- AI extensions add attribution, human-control, economics, and quality dimensions around unchanged base metrics. Prefer cohort comparisons (AI-assisted vs. not) over composite AI scores.
- Engines consume canonical observations only; they never call provider APIs.
- Each engine declares required input signals and computes a coverage score so missing collectors degrade gracefully (report what could not be computed, never silently zero).

### Value analytics

- structured-changelog's 20 categories map to 9 value classes (Capability, Customer Quality, Trust & Risk, Lifecycle, Operability, Engineering Quality, Maintainability, Enablement, Communication).
- `grokify/releaselog` is a complementary delivery-signal source: its multi-org GitHub-releases JSON IR (already generated for plexusone.dev's `/releases/` page) can be imported cheaply as `devx.change.delivered`/release events, alongside (not instead of) the structured-changelog and omni-github collectors.
- Value-density metric family (e.g., capability additions per 100 commits, released changes per AI session) as AI-stable alternatives to LOC.
- Three-stage value model: delivered change → verified value → realized outcome.

### Adoption analytics (maintainers)

- `gogithub.RepositoryAdoptionSnapshot` (stars, forks, watchers, issues, contributors, traffic); prefer snapshot + delta over raw counts; deduplicate contributors across repos; separate demand signals from workload signals.

## 6. DevFolio Integration

- `contributor.Profile` gains `AISpace *AISpaceProfileSummary` — a concise summary with `reportRef` pointers to full period reports. The profile is a presentation document; period reports are the analytical source of truth.
- Existing `AICollabStats` co-author detection migrates into the git/github collectors' normalize step; `contributor/client.go` stops computing AI stats directly.
- Disclosure model: canonical private IR (`devfolio/v1`) with derived projections (`devfolio.public-profile/v1`) — metric scope (which repos contribute) is independent of repository disclosure (which identities appear). Five publication profiles; `public-portfolio` default. Projections are derived, never destructive edits of the canonical store.
- Publishing: `{username}/devfolio` repo → GitHub Pages; profile-README sections via `<!-- DEVFOLIO:START/END -->` managed markers; Go templates + embedded assets; `--push` always explicit.

## 7. Dashboard Projection and Visualization

Visualization is a **projection consumer**, not a new collection or analytics layer — it reads already-computed period reports and never touches provider APIs or raw events.

### Two-artifact export

Every visualized report exports as two separate JSON artifacts, not one:

1. **Data** — the disclosure-safe projection of a `DeveloperPeriodReport`/`SPACEReport`/`DORAReport` (redacted per the active publication profile from §6). Generated per period by DevFolio.
2. **Dashboard definition** — a [dashforge](https://github.com/plexusone/dashforge) `dashboardir.Dashboard` JSON (`Layout`, `DataSources[]` pointing at the data artifact by URL, `Widgets[]`: `metric` widgets for SPACE's five dimensions + AI extension fields, `metric`/`chart` widgets for DORA's four keys, `table` for the `bySource` breakdown). Mostly static — built once per report type, not regenerated per period.

Dashforge is chosen over a Grafana/OTel stack (evaluated 2026-07-19) because it is JSON-IR-first (matches the ecosystem's Go-first schema convention), starts static-file-only with zero infrastructure, and already has a working precedent for this exact pattern (`pipelineconductor check -o data.json` → `viewer/?dashboard=...`). This keeps visualization inside the existing "CLI + JSON artifacts + static sites" scope (PRD §Out of Scope) rather than introducing a hosted/real-time service.

### Metric scoring and thresholds

[`ProductBuildersHQ/productbuildershq-frameworks`](https://github.com/ProductBuildersHQ/productbuildershq-frameworks) is a **taxonomy/threshold catalog** (`AISpaceFramework`, `AIDoraFramework`: dimension → metric ID → `MetricLevels{Elite,High,Medium,Low}`), not a report schema — it carries no subject, period, or computed value. It is used **only at the projection step**, to look up a computed metric's level/tier and label for display (feeding dashforge's `MetricConfig.Thresholds`/`Icon`), never inside `omnidevx-core`'s `space`/`dora` compute engines. Two reasons: `omnidevx-core` is committed to staying dependency-light (§1 Dependency rules), and this avoids a hard cross-org dependency (`plexusone` → `ProductBuildersHQ`) inside the compute path — if published thresholds change, only presentation shifts, not stored metric values. `omnidevx`'s emitted metric IDs should align with `frameworks.SpaceMetric.ID`/`DoraMetric.ID` where the two vocabularies overlap, so the lookup is a direct key match rather than a translation table.

### Visionstudio consumption boundary

[`ProductBuildersHQ/visionstudio`](https://github.com/ProductBuildersHQ/visionstudio) consumes the exported **projection only** (data + dashboard-definition JSON), never the canonical private IR or raw period reports. This mirrors the disclosure model in §6: visionstudio is a separate product/org (ProductBuildersHQ, not PlexusOne), so the org boundary gets the same treatment as any other public consumer. Integration is read-only file/URL consumption by visionstudio's Go daemon (rendered via dashforge's static viewer or natively alongside its existing `maturity-model/` dashboard component) — no shared database, no live query path into DevFolio's local store.

## 8. Schema Workflow

Per global Go-first convention: Go structs are the source of truth; JSON Schemas are generated (`invopop/jsonschema`), linted with `schemago lint`, embedded via `//go:embed`, and committed alongside the types. Applies to `omnidevx.event/v1`, `omnidevx.developer-period/v1`, SPACE/DORA report schemas, and DevFolio projection schemas.

## 9. Non-Requirements

- No universal shell wrapper/proxy around coding agents.
- No nested Go modules within a repo without demonstrated need.
- No collector may compute framework metrics (SPACE/DORA/value) — collectors normalize only.
- No silent metric combination across sources with mismatched definitions.
