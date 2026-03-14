# Submission Notes

## Overview

Equinox is a local-first infrastructure prototype for cross-venue prediction market normalization, clustering, and routing simulation. The implementation remains Go-first and scriptable from the CLI, but the primary demo surface is now a thin local web UI backed by the same engine.

This submission implements:

- `Polymarket` and `Kalshi` adapters
- canonical event clustering
- canonical proposition clustering inside events
- venue-agnostic routing simulation
- deterministic fixture-first demo support
- durable local persistence through a relational store boundary

The current fixture corpus includes a live-style Premier League event for Liverpool vs Tottenham on Sunday, March 15, 2026 at 11:30 AM Central, modeled with three routeable cross-venue match-outcome propositions.

This is intentionally a prototype, not a trading product.

## What To Review First

Recommended read order:

1. `README.md`
2. `docs/ARCHITECTURE.md`
3. `docs/DEMO.md`

Then run:

```bash
make dev
```

Then open [http://127.0.0.1:8080](http://127.0.0.1:8080).

To see the supported operator commands:

```bash
make
```

To run verification from the CLI:

```bash
make verify
```

To inspect routeable proposition clusters before routing a specific CLI order:

```bash
make list-clusters ROUTEABLE_ONLY=1
```

Optional live ingest check:

```bash
make live-inspect LIVE_LIMIT=1
```

To route a specific hypothetical order against the fixture state:

```bash
make route-order CLUSTER=prop-008 SIDE=buy_yes LIMIT=0.76
```

## Main Architectural Decisions

- Event clustering is the architectural center.
- Proposition clustering exists inside event clusters.
- Venue market instances attach beneath proposition clusters.
- Routing only occurs for route-safe proposition clusters.
- Binary-only is a routeability constraint, not an event-clustering constraint.
- Unsupported and ambiguous contracts are surfaced explicitly rather than coerced into routing.

## Tradeoffs

### Local-first deployment

Local deployment was chosen deliberately because the project specification explicitly allows it. This keeps the prototype focused on canonical identity, ambiguity handling, and routing structure rather than deployment overhead.

### Thin demo UI

The local web UI is intentionally thin. It is a presentation layer over the same fixture-backed Go pipeline already exercised by the CLI. This makes the demo easier to operate without introducing a second application architecture.

### SQLite fallback

The PRD originally preferred PostgreSQL as the local relational store. The implemented MVP uses SQLite as an embedded relational fallback to preserve a clean relational persistence boundary while keeping the reviewer path deterministic and low-friction.

### Fixture-first review path

The primary reviewer path is fixture-backed because live cross-venue routeable overlaps are unstable and because the assignment is primarily evaluating architecture, decomposition, and ambiguity handling.

### Curated versus derived fixture behavior

Fixtures remain curated for deterministic demo quality, but the primary fixture path derives meaningful normalization behavior in code, including:

- normalized proposition text
- binary and `Other` inference
- unsupported-shape inference
- ambiguity cue inference
- deadline provenance inference

Still curated in fixtures:

- event family and category hints
- source-style rules text and deadline strings
- quote and depth snapshots

## Known Limitations

- Clustering remains heuristic and fixture-calibrated.
- The live-inspect path validates public ingestion availability only; it does not guarantee live routeable overlaps.
- The prototype does not model fees, settlement economics, execution risk, or real-money trading.
- Only simple binary yes or no propositions are treated as routeable in the MVP.

## AI Usage Disclosure

AI tools were used during planning and implementation.

Specifically:

- ChatGPT/Codex was used to help research venue behavior, refine the PRD, and generate implementation drafts.
- Long-running Codex cloud tasks were used to produce implementation PRs.
- Those PRs were reviewed critically, rejected when necessary, and iterated on before merge.
- Final architectural direction, scope decisions, and merge decisions were made under human supervision.

## Final Reviewer Takeaway

The intended value of this submission is not execution polish. It is a clear demonstration that:

- cross-venue event identity can be modeled explicitly
- proposition-level routeability should be stricter than event similarity
- ambiguity and unsupported structures can be surfaced honestly
- routing can remain venue-agnostic when it consumes normalized signals only
