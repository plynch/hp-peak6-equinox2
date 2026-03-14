# Equinox MVP (Go 1.22+)

Equinox is a local-first prototype for cross-venue prediction market clustering and routing simulation. The core engine is Go, and the primary demo surface is now a thin local web UI backed by the same fixture-first pipeline as the CLI.

Implemented venues:
- Polymarket
- Kalshi

Additional docs:
- `SUBMISSION.md`
- `docs/ARCHITECTURE.md`
- `docs/DEMO.md`

## What this MVP demonstrates
- Fixture-backed ingestion of metadata and quote/order-book-like fields.
- Canonical event clusters from cross-venue heuristic similarity (event family, token overlap, category, deadline proximity).
- Canonical proposition clusters inside each event from proposition-text similarity + deadline checks.
- Explicit routeability classifications (routeable, event-only, unsupported, ambiguous).
- Venue-agnostic routing simulation over normalized inputs with limit-price feasibility enforcement.
- Inspectable artifacts and durable relational persistence (SQLite fallback relational store).

## Fixture posture: what is derived vs curated
Fixtures are still curated for stable demos, but they now look more like venue-style source records and less like pre-normalized canonical rows.

Derived in code during normalization:
- normalized proposition text from market question/title,
- binary/non-binary and `Other` detection from outcomes and market type,
- unsupported-shape inference from market type/outcomes/rules text,
- ambiguity notes from semantics cues in market/rules text,
- deadline provenance (`explicit_market_deadline`, `rules_text_only`, `missing`).

Still curated in fixtures:
- event family/category labels,
- source-style rules text and deadline fields,
- quote/depth snapshot values.

## Quickstart
```bash
make dev
```

Then open [http://127.0.0.1:8080](http://127.0.0.1:8080).

To see the available operator commands:

```bash
make
```

Outputs:
- SQLite DB: `equinox.db`
- Artifacts: `artifacts/<timestamp>/bundle.json`

## Commands
- `make dev`
  - Starts the local web UI demo on `http://127.0.0.1:8080`.
- `make verify`
  - Runs `go test ./...` plus the deterministic fixture CLI flow.
- `make`
  - Prints the supported demo and operator targets.
- `go run ./cmd/equinox fixture-demo`
  - Deterministic CLI artifact/materialization path.
- `make route-order`
  - Routes the default hypothetical `buy_yes` order against the default fixture cluster `prop-001`.
- `make route-order SIDE=sell_yes LIMIT=0.55`
  - Routes a `sell_yes` order against the same fixture cluster.
- `make live-inspect LIVE_LIMIT=3`
  - Optional public API check for current ingestion viability.

## What Orders Can Be Routed
- Only proposition clusters marked `routeable` can be routed.
- The current CLI supports hypothetical `buy_yes` and `sell_yes` orders.
- Orders must specify:
  - proposition cluster id
  - side
  - limit probability
  - size notional
- In the current fixture corpus, the routeable cluster is:
  - `prop-001`: `fed hike rate march 2026 meeting`
- The router refuses:
  - unsupported clusters
  - ambiguous clusters
  - event-only clusters
  - executable quotes that violate the order limit

## How Routing Works
- `make dev` loads fixture state, normalizes venue records, builds event clusters, then proposition clusters, writes the local artifact and SQLite state, and starts the web UI.
- The web UI uses that same fixture snapshot to show routeable clusters, default routing outcomes, and a browser-based order simulator.
- `make verify` exercises the CLI path without needing the browser.
- `make route-order ...` reloads that same fixture state and looks up the requested proposition cluster.
- For `buy_yes`, the executable price is `yes_ask`, and it must be less than or equal to the order limit.
- For `sell_yes`, the executable price is `yes_bid`, and it must be greater than or equal to the order limit.
- The router discards non-executable venues, then ranks feasible venues by price closeness to the limit plus available depth.
- With the current fixture quotes:
  - `make route-order` routes `buy_yes` on `prop-001` to `Polymarket`
  - `make route-order SIDE=sell_yes LIMIT=0.55` routes `sell_yes` on `prop-001` to `Kalshi`

## Evaluation set labels
The fixture artifacts include labels for:
- strong route-safe proposition cluster
- near-match or event-only case
- clear non-match case (explicit paired non-match assessment)
- unsupported-shape case
- ambiguity case

## Scope boundaries
- No third venue.
- No real-money execution.
- No production-grade UI.
- Routeable family restricted to simple binary yes/no only.
- Unsupported/ambiguous markets are surfaced, not forced into routeable clusters.
