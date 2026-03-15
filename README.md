# Equinox MVP (Go 1.22+)

Equinox is a local-first prototype for cross-venue prediction market clustering and routing simulation. The core engine is Go, and the demo surface is deliberately split into:
- a thin local web UI for fast reviewer walkthroughs
- a polished CLI for terminal-first scanning, selection, and routing across every supported source

Implemented venues:
- Polymarket
- Kalshi

Additional docs:
- `SUBMISSION.md`
- `docs/ARCHITECTURE.md`
- `docs/DEMO.md`

## What this MVP demonstrates
- Fixture-backed ingestion of metadata and quote/order-book-like fields.
- Ongoing live Premier League ingestion from Polymarket and Kalshi public APIs.
- Ongoing live Fed decision ingestion from Polymarket and Kalshi public APIs.
- Canonical event clusters from cross-venue heuristic similarity (event family, token overlap, category, deadline proximity).
- Canonical proposition clusters inside each event from proposition-text similarity + deadline checks.
- A live-style Premier League event cluster for Liverpool vs Tottenham on Sunday, March 15, 2026 at 11:30 AM Central, with three routeable cross-venue propositions (`liverpool win`, `draw`, `tottenham win`).
- A live EPL batch scan that fetches the current upcoming slate and simulates routing across every routeable match-outcome proposition it can match.
- A live Fed batch scan that fetches the current plus next few open meetings and simulates routing across every exact, route-safe cross-venue rate-decision proposition it can match.
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

For the terminal showcase across all supported sources:

```bash
make demo-cli
```

For the live Premier League browser demo:

```bash
make dev-live-epl
```

By default, the live EPL path uses the current plus next 4 matchweek-style windows that can be inferred from the public APIs.

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
- `make demo-cli`
  - Runs the full terminal showcase across the fixture snapshot, live Fed meetings, and live Premier League matchweek windows.
- `make dev-live-epl`
  - Starts the same web UI backed by a live Premier League snapshot from the public Polymarket and Kalshi APIs.
- `make dev-live-fed`
  - Starts the same web UI backed by a live Fed decision snapshot from the public Polymarket and Kalshi APIs.
- `make verify`
  - Runs `go test ./...` plus the deterministic fixture CLI flow.
- `make`
  - Prints the supported demo and operator targets.
- `make scan SOURCE=live-fed`
  - Runs the source-aware terminal scan for a single source (`fixture`, `live-fed`, `live-epl`, or `all-live`).
- `make scan SOURCE=all-live FED_MEETINGS=2 LIVE_MATCHWEEKS=1`
  - Prints every currently routeable live event across the supported Fed and EPL domains, then shows the underlying proposition clusters and selector-ready route commands.
- `make showcase`
  - Alias-equivalent to the terminal demo flow used by `make demo-cli`.
- `make live-fed`
  - Fetches the current plus next 4 open Fed meetings from both venues, builds live event/proposition clusters, and simulates routing across every routeable proposition cluster it finds.
- `make live-epl`
  - Fetches the current plus next 4 upcoming Premier League matchweek-style windows from both venues, builds live event/proposition clusters, and simulates routing across every routeable proposition cluster it finds.
- `make list-clusters ROUTEABLE_ONLY=1`
  - Lists the currently routeable proposition clusters so you can choose an explicit target for routing.
- `go run ./cmd/equinox fixture-demo`
  - Deterministic CLI artifact/materialization path.
- `make route-order CLUSTER=prop-001`
  - Routes a `buy_yes` hypothetical order against the Fed proposition cluster.
- `make route-order EVENT_QUERY='fomc march 2026' PROP_QUERY='fed hike rate march meeting' LIMIT=0.60`
  - Routes by human-readable selector instead of internal cluster id.
- `make route-order CLUSTER=prop-008 SIDE=buy_yes LIMIT=0.76`
  - Routes a `buy_yes` hypothetical order against the Liverpool–Tottenham `liverpool win` proposition.
- `make route-order CLUSTER=prop-007 SIDE=buy_yes LIMIT=0.15`
  - Routes a `buy_yes` hypothetical order against the Liverpool–Tottenham `draw` proposition.
- `make live-inspect LIVE_LIMIT=3`
  - Optional public API check for current ingestion viability.

## What Orders Can Be Routed
- Only proposition clusters marked `routeable` can be routed.
- The current CLI supports hypothetical `buy_yes` and `sell_yes` orders.
- Orders can specify either:
  - a proposition cluster id, or
  - a human-readable event selector plus proposition selector
- Orders also specify:
  - side
  - limit probability
  - size notional
- In the current fixture corpus, the routeable clusters are:
  - `prop-001`: `fed hike rate march meeting`
  - `prop-004`: `both teams score`
  - `prop-007`: `draw` for `Liverpool vs Tottenham`
  - `prop-008`: `liverpool win` for `Liverpool vs Tottenham`
  - `prop-009`: `tottenham win` for `Liverpool vs Tottenham`
- In the live EPL scan, routeable clusters are discovered dynamically from the current upcoming slate and can include many matched match-outcome propositions across both venues.
- In the live Fed scan, routeable clusters are discovered dynamically from the current plus next few open meetings and currently include exact overlaps like `fed no change`, `fed cut exactly 25bps`, and `fed cut more than 25bps`.
- Because live identifiers and prices move, the preferred live operator flow is `make scan SOURCE=live-fed ...`, `make scan SOURCE=live-epl ...`, or `make scan SOURCE=all-live ...`, then copy the selector-ready `make route-order ...` command it prints.
- If the public APIs expose fewer than 4 upcoming matchweek-style windows, the live scan returns whatever overlap is currently available.
- The router refuses:
  - unsupported clusters
  - ambiguous clusters
  - event-only clusters
  - executable quotes that violate the order limit

## How Routing Works
- `make dev` loads fixture state, normalizes venue records, builds event clusters, then proposition clusters, writes the local artifact and SQLite state, and starts the web UI.
- `make dev-live-epl` runs that same pipeline against live upcoming Premier League data instead of the curated fixture set.
- `make dev-live-fed` runs that same pipeline against live current/upcoming Fed decision data instead of the curated fixture set.
- For EPL, the live path groups fixtures into matchweek-style windows by date gaps because the public APIs expose match dates reliably but do not expose a stable official matchweek field.
- For Fed, the live path fetches the current plus next few open meetings, then matches only exact route-safe bucket semantics across venues. It will cluster broader event context even when proposition overlap is only partial.
- The web UI uses that same fixture snapshot to show routeable clusters, default routing outcomes, and a browser-based order simulator.
- `make verify` exercises the CLI path without needing the browser.
- `make list-clusters ROUTEABLE_ONLY=1` is the CLI discovery step before routing a specific order.
- `make route-order ...` reloads that same source state and resolves either the requested proposition cluster id or a human-readable event/proposition selector.
- `make live-epl` is the batch live operator path. It fetches the current EPL slate, builds clusters, and emits one `buy_yes` plus one `sell_yes` marketable routing simulation for each live routeable proposition cluster.
- `make live-fed` is the batch live operator path for FOMC decisions. It fetches the current plus next few open meetings, builds clusters, and emits one `buy_yes` plus one `sell_yes` marketable routing simulation for each live routeable proposition cluster.
- Use `LIVE_MATCHWEEKS=<n>` to widen or narrow the live EPL lookahead if you want fewer or more upcoming windows.
- Use `FED_MEETINGS=<n>` to widen or narrow the live Fed lookahead if you want fewer or more upcoming meetings.
- For `buy_yes`, the executable price is `yes_ask`, and it must be less than or equal to the order limit.
- For `sell_yes`, the executable price is `yes_bid`, and it must be greater than or equal to the order limit.
- The router discards non-executable venues, then ranks feasible venues by price closeness to the limit plus available depth.
- With the current fixture quotes:
  - `make route-order CLUSTER=prop-001` routes `buy_yes` on the Fed cluster to `Polymarket`
  - `make route-order CLUSTER=prop-001 SIDE=sell_yes LIMIT=0.58` routes `sell_yes` on the Fed cluster to `Kalshi`
  - `make route-order CLUSTER=prop-007 SIDE=buy_yes LIMIT=0.15` routes the Liverpool–Tottenham `draw` proposition to `Kalshi`
  - `make route-order CLUSTER=prop-008 SIDE=buy_yes LIMIT=0.76` routes the Liverpool–Tottenham `liverpool win` proposition to `Polymarket`
  - `make route-order CLUSTER=prop-009 SIDE=buy_yes LIMIT=0.10` routes the Liverpool–Tottenham `tottenham win` proposition to `Polymarket`

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

## Suggested demo posture
- Use `make dev` or `make verify` for the deterministic submission demo path.
- Use `make demo-cli` if you want the terminal to tell the whole story across fixture, Fed, and EPL in one run.
- Use `make live-fed` / `make dev-live-fed` and `make live-epl` / `make dev-live-epl` to show that the same architecture can continuously ingest and route across current open markets in both supported domains.
- The default live EPL lookahead is the current plus next 4 matchweek-style windows, bounded by what the current public APIs actually have open.
- The default live Fed lookahead is the current plus next 4 open meetings, bounded by what the current public APIs actually have open.
- If live public APIs change or thin out during the demo window, fall back to the fixture-backed path and say so explicitly.
For the live Fed browser demo:

```bash
make dev-live-fed
```
