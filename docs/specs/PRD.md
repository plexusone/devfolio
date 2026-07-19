# PRD: OmniDevX Ecosystem — AI SPACE Developer Metrics

**Scope:** Cross-repo product requirements for the PlexusOne developer-experience (DevX) telemetry and analytics ecosystem: `omnidevx-core`, `omnidevx`, `omni-github/omnidevx`, and DevFolio's role as the presentation layer.

**Residency note:** These specs live in `devfolio/docs/specs/` as a bootstrap home because the source ideation (`IDEATION_CHAT_METRICS.md`) originated here. Once `omnidevx-core` exists, the ecosystem-scoped portions should move there (or to `plexusone-specs`). The root [`PRD.md`](../../PRD.md) remains the DevFolio application-scoped PRD; this document is the ecosystem-scoped superset.

**Companion documents:**

- [TRD.md](TRD.md) — technical requirements and architecture
- [PLAN.md](PLAN.md) — phased implementation plan across repos
- [ROADMAP.md](ROADMAP.md) — per-repo version milestones

## Problem Statement

AI coding agents (Claude Code, Codex CLI, Cursor) change how software is built, but measurement frameworks have not kept up:

1. **The unit of measurement shifts from developer output to developer leverage.** Commits and LOC are weak signals when AI writes a growing fraction of the code. The meaningful question changes from "how productive is the developer?" to "how effectively does the developer orchestrate AI?"
2. **No vendor-neutral telemetry standard exists** for AI-assisted development. Each tool (Claude Code, Codex CLI, Cursor) has its own native telemetry; nothing normalizes across them.
3. **Existing frameworks (DORA, SPACE) need extension, not replacement.** The DORA team itself recommends evolving existing frameworks by adding AI-specific metrics (suggestion acceptance, trust, review effort) while keeping the original dimensions intact.
4. **Activity is not value.** Millions of AI-generated LOC per quarter say nothing about what was accomplished. Measurement must progress from activity to output to delivered value to business outcome.

## Vision

**AI SPACE** = SPACE + AI Attribution + AI Economics + AI Quality — implemented as a vendor-neutral collection, normalization, and analytics domain (**OmniDevX**) that any coding agent, IDE, or engineering platform can feed, and any analytics or portfolio system can consume. Analogous to OpenTelemetry for AI developer productivity.

DevFolio becomes the flagship consumer: a developer portfolio and reporting application built on canonical OmniDevX data.

## Naming Disambiguation

**OmniDevX ≠ OmniDXI.** The existing `plexusone/omnidxi` + `omnidxi-core` repos are **Digital Experience Intelligence** — a vendor-neutral facade over product-analytics platforms (Amplitude, Mixpanel, Heap, Pendo). OmniDevX is a separate capability domain for developer and AI-agent work telemetry. They must remain separate repos and must not share types or collectors. Both READMEs should carry a one-line disambiguation once `omnidevx-core` exists.

## Target Users

| User | Primary Use Case |
|------|------------------|
| Solo / hobby developers | Understand personal AI leverage, cost, and flow; zero-config historical import |
| Individual contributors | Portfolio generation enriched with AI SPACE metrics |
| Engineering managers | Team scorecards derived from the same canonical events (later phase) |
| Open source maintainers | Adoption, engagement, and contribution outcomes across an ecosystem of repos |
| Tool builders | Emit or consume canonical DevX events without vendor coupling |

**Solo developers first, unifying into teams.** No organizational politics, no privacy negotiation, easy instrumentation, rapid iteration. Build and validate the telemetry at the individual level first; once individual AI SPACE is implemented and proven, **team velocity is derived by aggregating individual velocity** — the same canonical events and period reports roll up developer → team → organization. Team metrics are a unification of individual metrics, not a separate measurement system.

## Product Principles

1. **Augment, don't replace.** SPACE → AI SPACE and DORA → AI DORA are extensions (`AIExtension` fields), never forks. Base metrics stay comparable across the industry.
2. **Frameworks are profiles over one event store.** SPACE, DORA, and future frameworks (flow, cost, quality) are analytical views over the same canonical events — collectors are never coupled to a framework.
3. **Generation ≠ acceptance ≠ retention ≠ quality ≠ outcome.** Each stage of the AI-code lifecycle (generated → applied → retained → committed → merged → verified → operated → delivered) is measured separately.
4. **Observed vs. inferred, always labeled.** Every metric carries provenance (collection mode, method, confidence). AI commit attribution is a signal, not proof of AI productivity.
5. **Scope discipline.** SPACE may be scoped to a person; DORA must be scoped to a delivery system (repo, service, product) — never a person.
6. **Collect broadly, publish narrowly.** Private work may contribute to aggregate metrics without disclosing repository identities. Publication is an explicit, profile-driven projection, never the raw store.
7. **Lead with narrative and value, not volume.** Portfolios foreground delivered capabilities and value density, not LOC leaderboards.

## Core Capabilities (Ecosystem Scope)

### 1. Collection

- **Retrospective importers** (zero-config first-run value): Claude Code and Codex CLI local session histories, git history, GitHub API.
- **Live collection:** native OpenTelemetry export from Claude Code and Codex CLI as the preferred continuous mechanism.
- **Enrichment:** a small, selective Claude Code hooks package for events OTel and git cannot provide (prompt submit, tool outcome, permission, task completion, session end).
- **Micro-surveys:** one optional question per session/day to capture Satisfaction, which logs cannot infer.
- **Value inputs:** `grokify/structured-changelog` portfolios as a first-class provider bridging activity → delivered change → value.

### 2. Canonical Model

- One event IR across all sources, with subject identity, source descriptor, context, attributes, and provenance/confidence.
- Person-level identity resolution (multiple GitHub accounts, git emails, devices) — GitHub username is never the canonical person ID.
- Period aggregation: raw events → daily summaries → weekly reports → monthly (canonical reporting cadence) → quarterly/annual synthesis.

### 3. Analytics

- `space` and `dora` engines, each with `Traditional` and `AIAugmented` profiles.
- AI extensions: attribution, economics (tokens, cost, cost-per-outcome), quality (acceptance, rewrite, verification failure, rollback), autonomy, human-control signals.
- Cohort comparisons (AI-assisted vs. not) preferred over synthetic composite "AI scores."
- Value analytics: structured-changelog categories mapped to value classes; value-density metrics (e.g., capability additions per 100 commits) that remain stable as AI inflates raw activity.
- Adoption/outcome analytics for maintainers: snapshot + delta metrics for stars, forks, contributors, traffic; relationship-level funnel (awareness → contribution → retention).

### 4. Presentation (DevFolio)

- Contributor profiles embedding concise AI SPACE summaries with references to full period reports.
- Publication profiles (`public-minimal`, `public-portfolio` (default), `public-transparent`, `private-share`, `private-full`) with a two-dimensional disclosure model: which repos contribute to stats vs. which repo identities are disclosed.
- Static portfolio publishing to `{username}/devfolio` GitHub Pages; GitHub profile README sections via managed markers. No automatic pushes by default.
- Report cadence: write monthly, summarize quarterly, reflect annually.

## Out of Scope (for now)

- Team and enterprise aggregation — deferred, not separate: team velocity will be a rollup of proven individual period reports (see PRD "Solo developers first, unifying into teams"). The identity model and period-report IR are designed for this from day one; only the rollup layer is deferred.
- A universal shell wrapper / proxy around coding agents (native OTel makes this unnecessary).
- GitLab, Jira/Atlassian, and CI-vendor thick providers (planned repo slots, not initial deliverables).
- Real-time dashboards / hosted service — initial deliverables are CLI + JSON artifacts + static sites.
- Using metrics for individual performance evaluation. Provenance labeling is mandatory precisely because recruiting/evaluation misuse is a known risk.

## Success Criteria

1. A solo developer can run a zero-config historical import and see sessions, prompts, tool calls, tokens, cost, and AI-assisted commits across Claude Code + Codex CLI + git within minutes.
2. The same weekly `DeveloperPeriodReport` is reproducible from stored events when metric formulas change (reprocessability).
3. Adding a new coding agent requires only a new collector — no changes to the analytics engines or DevFolio.
4. SPACE and DORA reports validate against their published schemas and clearly separate base metrics from AI extensions.
5. DevFolio publishes a portfolio in which no private repository identity leaks under the default publication profile.
