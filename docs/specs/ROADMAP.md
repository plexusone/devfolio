# ROADMAP: OmniDevX Ecosystem — Version Milestones

**Scope:** Per-repo version targets mapped to the phases in [PLAN.md](PLAN.md). Versions are targets, not commitments; each repo keeps its own `CHANGELOG.json`/`ROADMAP.json` as the authoritative record once created.

## Milestone Map

| Milestone | omnidevx-core | omni-openai | omnidevx | omni-github | gogithub | devfolio | Plan phase |
|-----------|---------------|-------------|----------|-------------|----------|----------|------------|
| M1 Contracts + store | v0.1.0 ◐ | — | — | — | — | — | 1 |
| M2 Historical import | v0.2.0 ✅ | omnidevx pkg (Codex) ✅ | — | — | — | — | 2 |
| M3 Period reports | v0.3.0 | — | — | — | — | — | 3 |
| M4 SPACE engine 🐕 | v0.4.0 (space/ incubated) | — | — | — | — | — | 4 |
| M5 Value + GitHub + DORA | v0.5.0 | — | v0.1.0 (created; space/dora move here) | omnidevx pkg ◐ (profile collector done 2026-07-16) | adoption snapshot | — | 5 |
| M6 DevFolio AI SPACE | — | — | — | — | — | v0.2.0 | 6 |
| M7 Live collection | v0.6.0 | otel support | v0.2.0 | — | — | — | 7 |
| M8 Publication | — | — | — | — | — | v0.3.0 | 8 |
| M9 Team unification 🚧 | v0.7.0 | — | v0.3.0 | — | — | v0.4.0 | 9 |

🐕 M4 starts the continuous **dogfooding gate** (weekly self-reports on real data). 🚧 M9 is **blocked by the dogfooding gate**: ~8+ weeks of reviewed self-reports; noise metrics cut before team rollups multiply them.

**Status (2026-07-16):** ✅ done · ◐ in progress. M2 is complete: all three importers (claudecode, Codex, git) plus the JSONL event store are implemented, tested (39 tests across four repos, lint-clean), and verified end-to-end on real data — ~139k events in `~/.plexusone/omnidevx/data/`, including 6,807 commits (61.8% AI-assisted since 2026-01-01). The git provider is built on `grokify/gogit` (gitscan renamed; see below). Remaining for M1: JSON Schemas, `providertest/`. Everything is local-only — no repo published yet; publish order is gogit (GitHub rename gitscan→gogit) → omnidevx-core → omni-openai → omnidevx (dropping replace directives).

## Per-Repo Detail

### plexusone/omnidevx-core

- **v0.1.0 — Contracts + store.** Canonical `Event` IR (`omnidevx.event/v1`) shaped by the Phase 0 spike data; `Collector` interface with constructor-injection composition (no priority registry yet); identity types; local event store at `~/.plexusone/omnidevx/data/` (daily JSONL per source, metadata-only privacy rule enforced at the type level); embedded generated schemas; `providertest/` conformance suite.
- **v0.2.0 — Importers.** `providers/claudecode` (stdlib-only session-JSONL reader) and `providers/git` with AI co-author attribution (ported from devfolio).
- **v0.3.0 — Aggregation.** Daily/weekly/monthly rollups, identity resolution, `DeveloperPeriodReport` (`omnidevx.developer-period/v1`), session↔commit correlation, reprocessing from the stored events.
- **v0.4.0 — SPACE engine (incubated).** `space/` package with `Traditional`/`AIAugmented` profiles, required-signal declarations, coverage scoring, solo-developer scorecard.
- **v0.5.0 — Value provider + extraction.** `providers/structuredchangelog` with category → value-class mapping; `space/` (and new `dora/`) extracted to the `omnidevx` batteries repo.
- **v0.6.0 — Live collection.** `providers/claudecode/otel.go` + hooks package, `providers/genericotel`, `providers/survey`.
- **v0.7.0 — Team rollup.** `TeamPeriodReport` derived from member `DeveloperPeriodReport`s; team identity mapping.

### plexusone/omni-openai (existing repo)

- **With M2 — omnidevx package.** Codex CLI collector alongside the existing `omnillm/` and `omnivoice/` domain packages: reads the current SQLite store (`threads`, `agent_jobs`, ...) and legacy rollout JSONL; the SQLite driver dependency is why this lives here rather than in core (TRD §1 placement rule).
- **With M7.** Codex OTel ingestion.

### plexusone/omni-anthropic (existing repo)

- **No initial work.** Named as the promotion target for the Claude Code collector only if it ever needs `anthropic-sdk-go` (e.g., API-side usage/cost reconciliation). Until then Claude Code stays thin in core.

### plexusone/omnidevx (batteries-included — created 2026-07-16, ahead of schedule)

Created early (local-only, on replace directives) to validate the composition
story: type re-exports, provider constructor re-exports, and an `Engine`
(constructor injection, per-collector failure isolation). Publishes once
`omnidevx-core` and `omni-openai` land on GitHub and the replaces are dropped.

- **v0.1.0 — Analytics home.** Created once there are thick providers to compose. Receives `space/` and `dora/` from core; `dora/` ships here first (four keys over GitHub workflow/release/deploy events, delivery-system subject enforcement, `AIDORAExtension`, AI-assisted vs. not cohorts); re-exports + composition of `omni-openai/omnidevx` and `omni-github/omnidevx`; value-density metrics.
- **v0.2.0 — Observed metrics.** Acceptance/rework/autonomy metrics upgraded to observed provenance when live collectors are present.
- **v0.3.0 — Team profiles.** Team-scoped SPACE profile over rolled-up reports.

### plexusone/omni-github

- **omnidevx package (with M5).** GitHub → `devx.*` event normalization (commits, PRs, reviews, workflows, issues), adoption snapshots + deltas, contributor dedup; follows the existing `omnistorage/`/`omniskill/` package conventions over a shared client layer.

### grokify/gogit (renamed from gitscan, 2026-07-16)

- **With M2 (done locally).** Generic git base library — repo discovery, `git log` parsing with trailers and numstat, branch/origin, remote-URL normalization — mirroring gogithub's role for GitHub. The gitscan CLI is preserved at `cmd/gitscan` (module path changed to `github.com/grokify/gogit`; GitHub-side rename pending at publish). `structured-changelog/gitlog` and the scanner package can migrate onto it later.

### grokify/gogithub

- **Existing basis (noted 2026-07-16):** `gogithub/profile` already collects GitHub contribution stats (`UserProfile` via GraphQL/REST: commits, PRs, issues, reviews, per-repo stats, contribution calendar, activity timeline) with monthly/quarterly outputs feeding `grokify/grokify/stats` and SVG/README renderers — the generic layer for M5 and the Phase 8 publishing surface.
- **With M5.** `RepositoryAdoptionSnapshot` (stars, forks, watchers, issues, contributors, traffic) helper — GitHub-generic, no DevX semantics. `omni-github/omnidevx` consumes `gogithub/profile` for contribution data.

### grokify/releaselog

- **With M5 (optional, cheap).** Its multi-org releases JSON IR (already generated for plexusone.dev `/releases/`) imported as delivery events. Also the reference pattern for the dogfooding usability bar: Go CLI → JSON IR → embeddable JS viewer.

### plexusone/devfolio

- **v0.2.0 — AI SPACE profiles (Plan Phase 6).** Depend on `omnidevx`; `collect`/`aggregate`/`space report` commands; `Profile.AISpace` summary with `reportRef`s; AI stats move from `contributor/client.go` into collectors; empty `datasource/` stubs removed. Reorders the root PRD's "Phase 2: team velocity" — individual velocity ships first because team velocity will be derived from it (see v0.4.0).
- **v0.3.0 — Publication (Plan Phase 8).** Disclosure profiles + projections, `publish`/`site build`/`site deploy`, `{username}/devfolio` Pages, profile-README managed markers, monthly report cadence. Dashboard-IR export (TRD §7) feeds `dashforge`'s viewer and, downstream, `ProductBuildersHQ/visionstudio`'s dashboard panel. Dogfooding target: an AI SPACE scorecard on plexusone.dev fed the way releaselog feeds `/releases/` today.
- **v0.4.0 — Team unification (Plan Phase 9).** Team velocity rebased onto rollups of individual `DeveloperPeriodReport`s — the unification of the solo-first strategy; existing changelog-based `team velocity` command is superseded. Team privacy/aggregation rules.
- **v0.5.0+ — Cadence + ecosystem.** Quarterly/annual synthesis, capability/ecosystem graph rollups, multi-audience report projections, organization-level rollups.

## Deferred / Unscheduled

- `omni-gitlab/omnidevx`, `omni-atlassian/omnidevx`, CI-vendor providers — when demand and dependency weight justify a thick provider.
- Priority-override registry (`PriorityThin`/`PriorityThick`) — only if a source ever gains both thin and thick implementations.
- Enterprise/organization scorecards — rollups of team reports, after M9 team unification proves out.
- Extraction of these specs to `omnidevx-core` or `plexusone-specs` — once `omnidevx-core` v0.1.0 exists.
- Capability Sandwich Architecture write-up (potential arXiv paper) — independent of code milestones.
