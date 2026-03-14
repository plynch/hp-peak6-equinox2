# AGENTS.md

## Repository Purpose

This repository is currently planning-first. The main source-of-truth artifacts are the project specification, the current PRD, and the presearch document.

Use this file as practical guidance for Codex runs in this repo. Keep it short, accurate, and aligned with the latest planning docs.

## Read First

Before proposing architecture, implementation, or review conclusions, read these files in order:

1. `PROJECT_SPECIFICATION.md`
2. `PROJECT_REQUIREMENTS_DOCUMENT.md`
3. `PROJECT_PRESEARCH_DOCUMENT.md`

If the task is a substantial implementation task, also read:

4. `PLANS.md`

If these documents conflict with older notes or deleted planning artifacts, trust the three files above.

## Current Planning Posture

The current PRD frames Equinox around canonical identity and clustering.

Assume the following unless the task explicitly changes planning:

- implemented venues:
  - `Polymarket`
  - `Kalshi`
- identity model:
  - canonical event clusters
  - canonical proposition clusters
  - venue market instances attached beneath them
- supported routeable proposition family:
  - simple binary yes or no propositions only
- event clustering posture:
  - event clustering is central to the MVP
  - event clusters may contain routeable, non-routeable, ambiguous, or unsupported market instances
  - `binary-only` narrows what may be routed, not what may be clustered at the event level
- demo posture:
  - fixture-first by default
  - live mode proves ingestion but may not prove live routeability every run
- packaging and persistence posture:
  - containerized reviewer path preferred
  - local PostgreSQL preferred
  - local deployment chosen intentionally because the spec explicitly allows it
  - preserve a clean path to hosted deployment later
- preferred implementation posture for this submission:
  - `Go 1.22+`
  - local-first
  - CLI-first

## Constraints

Do not silently widen scope beyond the PRD. In particular:

- do not add extra venues without an explicit planning change
- do not collapse the architecture into permanent venue-pair matching only
- do not treat scalar, combo, bucketed, or ambiguous `Other`-style contracts as supported routeable markets
- do not hide ambiguity behind aggressive heuristics
- do not rewrite older locked PRD versions

## Planning and Implementation Expectations

When asked to create or revise an implementation plan:

- preserve clear boundaries between adapters, normalization, clustering, routing, and artifact output
- include a durable local identity-store or registry boundary rather than assuming process-local state only
- make the main reviewer path easy to demo live, preferably via one-command containerized startup
- keep local deployment as the MVP default, but avoid choices that would block a later hosted deployment path
- keep routing venue-agnostic
- include the PRD's named target evaluation set in the implementation and demo plan
- include inspectable outputs for clustering and routing decisions
- prefer a deterministic fixture-first reviewer path
- state what is intentionally unsupported

When asked to implement the MVP or run a long Codex task:

- use `PLANS.md` as the execution plan
- keep `PLANS.md` aligned with reality if material implementation decisions change
- do not let Docker or PostgreSQL setup block the core prototype if the environment makes them unreliable
- preserve a clean relational store boundary even if an embedded relational fallback is used for execution reliability
- do not stop before running the relevant checks, preferably `go test ./...`, unless the repo state makes that impossible

## Review Guidelines

- treat spec drift, hidden ambiguity, and venue-specific router logic as high-priority issues
- prioritize bugs, risks, regressions, and missing evidence over style commentary
- flag cases where contracts appear clustered or routed without enough supporting evidence

## Documentation Hygiene

- Keep `PROJECT_REQUIREMENTS_DOCUMENT.md` synced to the newest active PRD version.
- If creating a new PRD version, leave older versioned PRDs unchanged unless explicitly asked.
- Keep repo instructions concise. If guidance grows large, move detail into the PRD or a task-specific doc instead of bloating `AGENTS.md`.
- Keep long-running implementation detail in `PLANS.md`, not in `AGENTS.md`.
