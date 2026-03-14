# Equinox MVP (Go 1.22+)

Equinox is a CLI-first, local-first prototype for cross-venue prediction market clustering and routing simulation.

Implemented venues:
- Polymarket
- Kalshi

## What this MVP demonstrates
- Fixture-backed ingestion of metadata and quote/order-book-like fields.
- Canonical event clusters from cross-venue heuristic similarity (event family, token overlap, category, deadline proximity).
- Canonical proposition clusters inside each event from proposition-text similarity + deadline checks (not exact fixture key equality).
- Explicit routeability classifications (routeable, event-only, unsupported, ambiguous).
- Venue-agnostic routing simulation over normalized inputs with limit-price feasibility enforcement.
- Inspectable artifacts and durable relational persistence (SQLite fallback relational store).

## Quickstart
```bash
go test ./...
go run ./cmd/equinox fixture-demo
```

Outputs:
- SQLite DB: `equinox.db`
- Artifacts: `artifacts/<timestamp>/bundle.json`

## Commands
- `go run ./cmd/equinox fixture-demo`
  - Deterministic, secret-free reviewer path.
- `go run ./cmd/equinox live-inspect --limit 3`
  - Optional public API check for current ingestion viability.

## Evaluation set labels
The fixture artifacts include labels for:
- strong route-safe proposition cluster
- near-match or event-only case
- clear non-match case
- unsupported-shape case
- ambiguity case

## Scope boundaries
- No third venue.
- No real-money execution.
- No production UI.
- Routeable family restricted to simple binary yes/no only.
- Unsupported/ambiguous markets are surfaced, not forced into routeable clusters.
