# PLAN: OmniDevX Ecosystem — Implementation Plan

**Scope:** Phased implementation across `omnidevx-core`, `omnidevx`, `omni-github`, and `devfolio`. Requirements in [PRD.md](PRD.md) and [TRD.md](TRD.md); version milestones in [ROADMAP.md](ROADMAP.md).

**Sequencing principle:** everything depends on the canonical IR, so `omnidevx-core` contracts come first; retrospective importers before live collection (zero-config value first); analytics before presentation polish.

## Phase 0 — Specs and Scaffolding

**Repos:** devfolio (this spec set), omnidevx-core (new)

- [x] Distill `IDEATION_CHAT_METRICS.md` into PRD/TRD/PLAN/ROADMAP (this directory).
- [x] **Data spike:** verify what the local session stores actually contain before finalizing the IR. Findings (2026-07-16): Claude Code keeps per-project session JSONL (`~/.claude/projects/<project>/<session>.jsonl`, 136 projects / 861 MB on the reference machine) with typed records carrying sessionId, cwd, gitBranch, model, and full token usage including cache tiers — sessions/prompts/models/tokens/cost/repo are all recoverable. Codex has **migrated from rollout JSONL to SQLite** (`state_5.sqlite`: threads, agent_jobs, thread_spawn_edges); only legacy JSONL remains in `~/.codex/sessions/`. The Codex importer must read both formats, and the SQLite dependency drives its placement in `omni-openai/omnidevx` (TRD §1).
- [x] Create `plexusone/omnidevx-core` (2026-07-16: repo created with README disambiguation vs. `omnidxi`, LICENSE, .gitignore; CHANGELOG/ROADMAP/mkdocs/CI scaffolding still pending — add before first push).
- [x] Add matching disambiguation line to `omnidxi-core` README (2026-07-16).
- [x] Add scope note to devfolio `TASKS.md` (same treatment as root `PRD.md`) pointing at this spec set.

## Phase 1 — Canonical Contracts (omnidevx-core)

**Goal:** the IR everything else depends on.

- [x] Root package `omnidevx` (2026-07-16): `Event`, `ai.*` `EventType`s (incl. `ai.usage.recorded`, added from spike data), shared attribute keys, `SubjectRef`, `Source`, `Period` (half-open, tested), `Provenance`, `Diagnostic`. Still pending: `devx.*` event types (arrive with git/GitHub providers), `Metric`/`Measurement` (arrive with aggregation).
- [x] `Collector` interface + `CollectRequest`/`CollectionResult` (2026-07-16); composition via constructor injection through the `omnidevx` Engine (created early) — no priority registry, per TRD §2. Compile-time `var _ Collector` assertions in both providers.
- [x] Local event store (2026-07-16, `omnidevx-core/store`, stdlib-only — JSONL confirmed over SQLite): daily JSONL per source under `~/.plexusone/omnidevx/data/events/YYYY/MM/DD/<product>.jsonl`, owner-only permissions, ID-dedup idempotent writes, period/product-scoped reads with damaged-line diagnostics. Verified on real data: 131,921 events persisted (70 MB), second import 100% deduplicated, 7-day read in ~230ms. Bonus finding: collectors emit ~3% duplicate IDs from Claude Code session resume/branch prefix copies — the store is the dedup boundary by design.
- [ ] Generate + embed JSON Schemas (`omnidevx.event/v1`) via invopop/jsonschema + schemago lint.
- [ ] Conformance test package (`providertest/`) analogous to omnillm's.

**Exit criteria:** a mock collector round-trips events through schema validation and the registry.

## Phase 2 — Retrospective Importers (omnidevx-core)

**Goal:** zero-config historical value for a solo developer.

- [x] `providers/claudecode/history.go` (2026-07-16, in core, stdlib-only) — sessions, prompts (sidechains excluded), assistant messages with model + token usage, tool completions with name attribution, diagnostics for unparseable lines. Verified against the real store: 134,259 events / 156 sessions / 0 diagnostics. Deferred: cost estimates (needs a maintained pricing table).
- [x] `omni-openai/omnidevx` package (2026-07-16) — Codex importer reading the SQLite `threads` index and rollout JSONL with session dedup, schema-drift fallback, and repo-URL normalization. Verified: 1,956 events / 5 sessions / 0 diagnostics.
- [x] `providers/git/` (2026-07-16, in core) — built on `grokify/gogit` (gitscan renamed to a generic git base library, CLI preserved at `gogit/cmd/gitscan`; adds trailer parsing, calendar-date log filtering, branch/origin, repo discovery). Emits `devx.change.committed` with `KnownAITools` AI-attribution ported from devfolio (canonical copy now lives here). Hash-keyed IDs dedup clones. Verified: 6,913 commits across 133 repos since 2026-01-01 in 22s, 0 diagnostics — **61.8% AI-assisted**.
- [x] All importers stamp `collectionMode: history` with confidence 0.9 (privacy rule enforced by tests: content-like attribute keys fail the build).
- [x] Events persist to the `~/.plexusone/omnidevx/data/` store (2026-07-16): Engine collect → `store.Write` verified end-to-end on real data with idempotent re-import.

**Exit criteria:** one command collects Claude Code + Codex + git events for a date range on this machine, persists them to the local store, and emits valid canonical event JSON. *Status: **MET** (2026-07-16) — all three sources collect via the `omnidevx` Engine and persist to `~/.plexusone/omnidevx/data/` with idempotent dedup, verified on real data.*

## Phase 3 — Aggregation and Period Reports (omnidevx-core)

- [ ] Daily summary builder; weekly/monthly rollups derived from days (never month-only).
- [ ] Identity resolution (`personId` + identities; hashed git emails; device-scoped local accounts).
- [ ] `DeveloperPeriodReport` (`omnidevx.developer-period/v1`) with `combined` + `bySource` metrics, coverage scoring, and safe-to-combine rules.
- [ ] Session-to-commit correlation (events near commits, AI-assisted commit linkage).

**Exit criteria:** reproducible weekly report from stored events; reprocessing with changed formulas yields updated reports without recollection.

## Phase 4 — SPACE Engine (incubated in omnidevx-core)

- [ ] `space/` package in `omnidevx-core`: dimensions, metric formulas, required-signal declarations, `Traditional`/`AIAugmented` profiles, `SPACEReport{Dimensions, AIExtension}`.
- [ ] Scorecard output (solo-developer headline metrics) + report schemas.
- [ ] DORA is deliberately deferred to Phase 5 — its delivery/verification signals (workflows, releases, deploys) don't exist until the GitHub provider lands, and a DORA report at this stage would be mostly coverage warnings.

**Exit criteria:** `space.Calculate` over Phase 3 reports produces a validated AI SPACE report with explicit coverage gaps. **This starts the dogfooding clock (see Dogfooding Gate below).**

## Phase 5 — Value, GitHub, and DORA

- [ ] `providers/structuredchangelog/` in core: wrap `grokify/structured-changelog/changelog`; category → value-class mapping; emit `devx.change.delivered`; correlate entries with commits (conventional-commits bridge).
- [ ] `gogithub`: add `RepositoryAdoptionSnapshot` helpers.
- [x] `omni-github/omnidevx` package, first increment (2026-07-16, pulled forward from M5): built on `gogithub/profile.GetUserProfile`, emits `devx.profile.snapshot`, `devx.contribution.snapshot` (per-repo), and `devx.contribution.recorded` (daily) with api-mode provenance. Live-verified against the published June 2026 stats: additions/deletions reproduce exactly (1,716,920/251,531); commit variance (2,068 vs 1,754) is all-repos vs public-only scope. Fidelity note: month-scale collects yield monthly-granularity calendar data, not daily. Remaining GitHub work below.
- [ ] `omni-github/omnidevx` remaining: per-item PR/review/issue/workflow → `devx.*` events (review.requested/completed, change.integrated, verification.completed), adoption snapshot + delta, contributor dedup across repos.
- [ ] Reconcile with the live `grokify/grokify/stats` monthly pipeline (`gogithub/profile/monthly_output` → `grokify_github_public_YYYY-MM.json` + quarterly rollups + SVG/README renders): `DeveloperPeriodReport` generalizes these GitHub-scoped proto-period-reports to multi-source; the `profile/svg` + `profile/readme` renderers become consumers of period reports (Phase 8), not a parallel pipeline.
- [ ] Optional cheap win: import `grokify/releaselog` JSON IR (already generated for plexusone.dev `/releases/`) as delivery events.
- [ ] `dora/` package (moved from Phase 4): four keys over GitHub workflow/release/deploy events, delivery-system subject enforcement, `AIDORAExtension`, cohort comparison (AI-assisted vs. not).
- [ ] Publish the batteries-included `plexusone/omnidevx` repo (created early on 2026-07-16 with Engine + re-exports, local replaces): move `space/`/`dora/` out of core, compose thick providers (`omni-openai/omnidevx`, `omni-github/omnidevx`), drop replace directives.
- [ ] Value-density metrics in the analytics layer.

**Exit criteria:** period reports include `deliveredChange` and `outcomes{adoption, engagement, contribution}` sections for the plexusone/grokify/agentplexus orgs, and `dora.Calculate` produces reports with real delivery data.

## Phase 6 — DevFolio Integration (devfolio v0.2.x)

- [ ] Depend on `omnidevx`; add `devfolio collect`, `devfolio aggregate --period week|month`, `devfolio space report` commands.
- [ ] `contributor.Profile` gains `AISpace *AISpaceProfileSummary` (headline metrics + `reportRef` + trends); AI stats computation moves out of `contributor/client.go` into collectors.
- [ ] Remove the empty `datasource/` stub directories — collection is omnidevx's job; empty dirs duplicating an external dependency's role are pure confusion (decided 2026-07-16).
- [ ] Reconcile root `PRD.md` Phase 2–4 roadmap with this plan.

**Exit criteria:** `devfolio contributor profile --include-ai-space` produces a profile with an AI SPACE summary sourced from period reports.

## Phase 7 — Live Collection and Satisfaction

- [ ] `providers/claudecode/otel.go`, `providers/codex/otel.go`, `providers/genericotel/` — OTLP ingestion, prompt↔tool correlation, approval tracking.
- [ ] Small Claude Code hooks package (only: UserPromptSubmit, PostToolUse/Failure, PermissionRequest, TaskCompleted, Stop, SessionEnd) feeding `collectionMode: hooks` events.
- [ ] `providers/survey/` — one-question session/day micro-surveys for Satisfaction and rework.
- [ ] Patch lifecycle instrumentation: generated → applied → retained (the stages history cannot recover).

**Exit criteria:** acceptance-rate and rework metrics flip from `estimated` to `observed` provenance when live collection is enabled.

## Phase 8 — Publication and Cadence (devfolio v0.3.x+)

- [ ] Disclosure engine: canonical private IR → profile-driven projections (`public-minimal|public-portfolio|public-transparent|private-share|private-full`); leak tests for private repo identities.
- [ ] `devfolio publish` / `site build` / `site deploy --provider github-pages` targeting `{username}/devfolio`; profile-README managed markers; Go templates + embedded assets.
- [ ] Dashboard-IR export (2026-07-19 architecture decision, TRD §7): project SPACE/AI-SPACE/DORA/AI-DORA period reports into two artifacts — a disclosure-safe data JSON and a `plexusone/dashforge` `dashboardir.Dashboard` definition. `ProductBuildersHQ/productbuildershq-frameworks` used only here (metric ID → level-threshold lookup for widget config), never inside the `omnidevx-core` compute engine. Consumed by dashforge's static viewer and, downstream, by `ProductBuildersHQ/visionstudio`'s daemon — projection-only, never canonical private IR, preserving the org boundary.
- [ ] Monthly capability report generation (canonical cadence), quarterly/annual synthesis from monthlies.
- [ ] Later: ecosystem/capability graph rollups (repository → capability → application → ecosystem) and multi-audience projections.

**Exit criteria:** `devfolio update` runs collect → aggregate → redact → generate → commit (push explicit) end-to-end.

## Phase 9 — Team Unification

**Goal:** team velocity derived from proven individual velocity — the unification of the solo-first strategy. Gated on individual AI SPACE metrics being validated in Phases 1–7.

- [ ] `TeamPeriodReport`: rollup of member `DeveloperPeriodReport`s over the same period grid (identity resolution from Phase 3 maps accounts → people → team).
- [ ] Team SPACE profile (team-scoped Satisfaction/Communication semantics); DORA stays delivery-system-scoped and now gets its natural subject.
- [ ] Rebase `devfolio team velocity` onto canonical events/rollups, replacing the changelog-only implementation.
- [ ] Privacy model for team contexts: individual detail visibility rules, aggregation minimums.

**Exit criteria:** a team scorecard is reproducible purely from its members' period reports — no team-specific collectors exist.

## Dogfooding Gate (continuous, from Phase 4)

The ideation doc's own bar: prove the metrics correlate with perceived productivity before extending to teams. This gate runs continuously and **blocks Phase 9**.

- [ ] From the first `space.Calculate` (Phase 4), generate weekly reports on John's own data every week and check them against the lived experience of that week — which of the ~25–30 candidate metrics are signal vs. noise decides what calcifies into v1 schemas.
- [ ] Usability bar: the `grokify/releaselog` → plexusone.dev pattern (Go CLI → JSON IR → embeddable JS viewer on the site's `/releases/` page). OmniDevX/DevFolio output should reach the same bar: a scorecard JSON that plexusone.dev (or a personal DevFolio site) can render with an embeddable widget, generated by one CLI command.
- [ ] Target artifact: an AI SPACE scorecard section on plexusone.dev (or `{username}/devfolio` Pages) fed the same way `/releases/` is fed by releaselog today.
- [ ] Gate check before Phase 9: at least ~8 weeks of self-reports reviewed; metrics that never informed a real decision get cut or demoted before team rollups multiply their noise.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Codex/Claude local session formats change without notice | Treat as internal formats: defensive parsers, diagnostics not failures, confidence < 1.0, prefer OTel prospectively |
| Metric definitions drift across tools (acceptance ≠ acceptance) | `bySource` always retained; combine only per the safe-to-combine table |
| Scope creep (ecosystem graph, multi-audience reports) before core value | Phases 1–4 ship a working solo-developer scorecard before any Phase 8 work starts |
| `omnidxi` name confusion | README disambiguation both sides (Phase 0); memory note recorded |
| Metrics misused for evaluation | Provenance mandatory on every metric; docs state observed-vs-inferred limits |
