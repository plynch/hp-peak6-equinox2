# Submission Notes

## Overview

Equinox is a local-first infrastructure prototype for cross-venue prediction market normalization, clustering, and routing simulation. The implementation remains Go-first and scriptable from the CLI, but the primary demo surface is now a thin local web UI backed by the same engine.

This submission implements:

- `Polymarket` and `Kalshi` adapters
- canonical event clustering
- canonical proposition clustering inside events
- venue-agnostic routing simulation
- deterministic fixture-first demo support
- ongoing live Fed decision scan support
- ongoing live Premier League scan support
- durable local persistence through a relational store boundary
- a terminal-first operator CLI plus a thin local demo UI

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

For the terminal-first all-sources demo:

```bash
make demo-cli
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

Optional live Fed scan:

```bash
make live-fed
```

Optional live Premier League scan:

```bash
make live-epl
```

To route a specific hypothetical order against the fixture state:

```bash
make route-order EVENT_QUERY='liverpool vs tottenham' PROP_QUERY='liverpool win' LIMIT=0.76
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

### Terminal-first operator CLI

The CLI is treated as a first-class demo surface, not just a verification harness. `make demo-cli` walks fixture, live Fed, and live EPL in one pass, while `make scan` and selector-based `make route-order` make it possible to demonstrate discovery and routing from the terminal without relying on internal cluster IDs.

### SQLite fallback

The PRD originally preferred PostgreSQL as the local relational store. The implemented MVP uses SQLite as an embedded relational fallback to preserve a clean relational persistence boundary while keeping the reviewer path deterministic and low-friction.

### Fixture-first review path

The primary reviewer path is fixture-backed because live cross-venue routeable overlaps are unstable and because the assignment is primarily evaluating architecture, decomposition, and ambiguity handling.

### Live operator paths

The repository includes both a live Fed scan path (`make live-fed` / `make dev-live-fed`) and a live EPL scan path (`make live-epl` / `make dev-live-epl`). These fetch currently open public markets from both venues, cluster the overlapping events and propositions, and simulate routing across every routeable proposition cluster they find. These paths are secondary to the deterministic fixture path, but they demonstrate that the same architecture can operate continuously on current public data.

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
- Live routeability depends on current public overlap and market availability between Polymarket and Kalshi.
- Live venue APIs can rate-limit or change their open slate during the demo window; the CLI handles this more gracefully now, but the deterministic fixture path remains the primary reviewer path.
- The public APIs do not expose a stable official EPL matchweek field, so the live scan infers matchweek-style windows from fixture dates.
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
