# Equinox MVP Execution Plan

## Purpose

This file is the execution plan for a long-running Codex implementation task in this repository.

Use it when the task is to build the Equinox MVP, not when the task is only document review or small edits.

This plan is intentionally concrete because the repository is still planning-heavy and code-light.

## Read Order

Before implementing, read these files in order:

1. `PROJECT_SPECIFICATION.md`
2. `PROJECT_REQUIREMENTS_DOCUMENT.md`
3. `PROJECT_PRESEARCH_DOCUMENT.md`
4. `AGENTS.md`
5. `PLANS.md`

If the plan conflicts with the specification or current PRD, the specification and PRD win.

## Goal

Produce a PR-ready implementation of the Equinox MVP that:

- integrates `Polymarket` and `Kalshi`
- ingests market metadata and pricing or order-book data
- builds canonical event clusters and canonical proposition clusters
- simulates venue-agnostic routing for hypothetical orders
- emits inspectable artifacts for clustering and routing decisions
- includes a deterministic fixture-first demo path
- includes concise reviewer-facing setup and demo instructions

## Delivery Standard

The implementation should optimize for:

- architectural clarity
- clean decomposition
- explicit ambiguity handling
- honest unsupported-case handling
- easy reviewer demo execution
- high probability of successful completion in a single Codex cloud run

Do not optimize for production polish or feature breadth.

## Non-Goals

Do not add any of the following unless the user explicitly changes scope:

- third-venue implementation
- real-money execution
- wallets or settlement integration
- production UI
- distributed clustering services
- exhaustive sports or market-family support
- ML-heavy or graph-heavy clustering infrastructure

## Key Architectural Decisions

Implement the MVP around these core abstractions:

- `event_cluster`
- `proposition_cluster`
- `venue_market_instance`
- `equivalence_assessment`
- `routing_decision`

Maintain clean boundaries between:

- adapters
- normalization
- clustering
- storage
- routing
- artifact output

The router must be venue-agnostic:

- acceptable inputs are normalized cluster-level signals
- unacceptable inputs are raw venue payloads and venue-name branches

## Store Strategy

The PRD prefers a relational store and names PostgreSQL as the preferred submission choice. For a long-running Codex cloud task, completion reliability matters more than strict adherence to that preference.

Execution rule:

- preserve a clean relational store boundary no matter what
- if Docker plus PostgreSQL is straightforward and does not threaten completion, use it
- if Docker or PostgreSQL becomes a material blocker in the cloud environment, use an embedded relational fallback such as SQLite
- if the fallback is used, document it clearly and preserve a clean path to PostgreSQL later

Never block the full implementation on Docker-in-Docker assumptions.

## Packaging Strategy

Primary objective:

- a working Go CLI prototype with deterministic fixture execution and passing tests

Secondary objective:

- containerized reviewer packaging

Execution rule:

- do not let container packaging block the core prototype
- build the core app and tests first
- add Docker packaging only after the core system works
- if full multi-container verification is not possible in the cloud environment, still write the packaging files if they are straightforward and document any unverified parts honestly

## Preferred Implementation Shape

Language and posture:

- `Go 1.22+`
- stdlib-first unless a dependency materially improves correctness or speed
- CLI-first
- local-first

Suggested repository shape:

- `go.mod`
- `cmd/equinox/`
- `internal/adapters/polymarket/`
- `internal/adapters/kalshi/`
- `internal/model/`
- `internal/store/`
- `internal/normalize/`
- `internal/cluster/`
- `internal/router/`
- `internal/artifacts/`
- `testdata/fixtures/`
- `docs/` only if needed for reviewer guidance

## Functional Scope

Supported implementation scope:

- `Polymarket`
- `Kalshi`
- event-level clustering
- proposition-level clustering
- simple binary yes or no propositions only for routeable proposition clusters
- routing simulation for supported proposition clusters

Important clarification:

- event clustering is broader than routing scope
- the system should still form event clusters around semantically related market instances even when some members are non-routeable, unsupported, ambiguous, or only event-level matches
- `binary-only` is a routing and proposition-support constraint, not a rejection of clustering as the architectural center

Required unsupported handling:

- scalar markets
- combo or multivariate markets
- bucketed or range-like contracts
- placeholder or `Other` outcomes
- materially unclear deadline or resolution semantics

## Required Demo and Evaluation Set

The implementation should include a labeled set that demonstrates:

- one strong route-safe proposition cluster
- one event-only or near-match case
- one clear non-match case
- one unsupported-shape case
- one ambiguity case

Prioritize domains in this order:

1. `Fed / FOMC`
2. next week's Big Five soccer fixtures and closely related soccer event families
3. golf only if additional structure stress-testing is useful

The strong route-safe example may be fixture-backed.

## Implementation Phases

### Phase 1. Scaffold the repo

Create the initial Go project structure and the minimal developer workflow.

Expected outputs:

- `go.mod`
- CLI entrypoint
- package layout
- README or run instructions only if needed
- test command that runs cleanly even before all features are complete

### Phase 2. Define the core model

Implement types for:

- event clusters
- proposition clusters
- venue market instances
- quote views
- equivalence assessments
- routing decisions
- artifact payloads

Preserve explicit fields for:

- confidence
- ambiguity notes
- routeability status
- deadline provenance
- quote freshness
- observed versus inferred quote signals

### Phase 3. Implement adapters and fixture ingestion

Implement:

- `Polymarket` adapter
- `Kalshi` adapter

Each adapter should support:

- fixture-backed ingestion
- live inspect or live ingest mode if public access is available

Fixture ingestion is required and should be the default demo path.

### Phase 4. Implement storage

Implement the relational persistence boundary for:

- cluster records
- venue market instances
- assessment records
- routing decisions

Keep the persistence boundary clean enough that store choice can vary without changing clustering or routing logic.

### Phase 5. Implement normalization and clustering

Implement:

- normalization from venue-native payloads into candidate identities
- event clustering
- proposition clustering
- explicit classification for:
  - strong proposition match
  - event-only match
  - near-match
  - unsupported or unsafe
  - insufficient data

The system must preserve refusals and uncertainty instead of forcing matches.

### Phase 6. Implement routing

Implement routing simulation for hypothetical orders using only normalized inputs.

The router should:

- choose a venue or refuse to route
- explain the decision
- never branch on venue names

### Phase 7. Emit artifacts

Emit inspectable artifacts for:

- normalized market views
- event clusters
- proposition clusters
- equivalence assessments
- routing decisions

Artifacts should make the demo legible without requiring the reviewer to read source code.

### Phase 8. Add tests and reviewer docs

Before finishing:

- add focused tests for the core logic
- run the relevant checks
- review the diff for scope drift and risky assumptions
- write concise setup and demo instructions

## Verification Gates

Do not declare the task complete unless these are true or explicitly documented as blocked:

1. `go test ./...` succeeds.
2. The fixture demo path runs end to end.
3. The labeled evaluation set is present and inspectable.
4. At least one routing decision or justified refusal is produced.
5. The output explains why clustering or routing happened.
6. Unsupported and ambiguous cases are surfaced explicitly.
7. Documentation matches the implemented behavior.

If Docker packaging is added, validate it if the environment supports it. If it cannot be validated, say so clearly in the final docs.

## Review Checklist

Before finishing, review the implementation for:

- accidental venue-specific router logic
- markets forced into proposition clusters without enough evidence
- hidden ambiguity
- unsupported market types treated as supported
- drift from the current PRD
- unnecessary framework or dependency sprawl

## Allowed Deviations

You may deviate from this plan only when:

- the repository state makes the planned step impossible or clearly inferior
- the specification or PRD requires a different choice
- the cloud environment makes a preferred packaging or store choice materially unreliable

When deviating:

- make the smallest viable change
- document the reason in the relevant docs
- preserve the architectural boundaries described in the PRD

## Definition of Done

The task is done when the repository contains a coherent, PR-ready MVP with:

- working code
- tests or other concrete verification
- fixture-backed demo support
- live ingest or live inspect support if feasible
- reviewer-facing setup instructions
- reviewer-facing demo instructions
- implementation artifacts that make the clustering-first architecture easy to inspect
