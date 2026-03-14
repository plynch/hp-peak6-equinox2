# Demo Script

## 1) Start the local demo
```bash
make dev
```

Then open [http://127.0.0.1:8080](http://127.0.0.1:8080).

`make dev` now:

- loads the current fixture snapshot
- materializes `equinox.db`
- writes `artifacts/<timestamp>/bundle.json`
- starts the local browser demo

If you want a quick list of supported commands first:

```bash
make
```

## 2) Inspect routeable clusters from the CLI
```bash
make list-clusters ROUTEABLE_ONLY=1
```

This is the explicit selection step before routing a CLI order.

## 3) Use the web UI first
- Review the routeable cluster card.
- Use the order simulator form with the default values.
- Confirm that the default `buy_yes` order routes to `Polymarket`.
- Change to `sell_yes` with limit `0.58` on the Fed cluster and confirm it routes to `Kalshi`.
- Switch to the Liverpool vs Tottenham event, scheduled for Sunday, March 15, 2026 at 11:30 AM Central.
- Confirm that the `draw` proposition at limit `0.15` routes to `Kalshi`.
- Confirm that the `liverpool win` proposition at limit `0.76` routes to `Polymarket`.
- Confirm that the `tottenham win` proposition at limit `0.10` routes to `Polymarket`.

## 4) Run checks from the CLI
```bash
make verify
```

`make verify` runs tests plus the fixture CLI path.

## 5) Route specific hypothetical orders from the CLI
```bash
make route-order CLUSTER=prop-001
```

You can also try:

```bash
make route-order CLUSTER=prop-001 SIDE=sell_yes LIMIT=0.58 SIZE=1000
make route-order CLUSTER=prop-007 SIDE=buy_yes LIMIT=0.15 SIZE=1000
make route-order CLUSTER=prop-008 SIDE=buy_yes LIMIT=0.76 SIZE=1000
make route-order CLUSTER=prop-009 SIDE=buy_yes LIMIT=0.10 SIZE=1000
```

## 6) Inspect artifact
```bash
LATEST=$(ls -1 artifacts | tail -n 1)
cat artifacts/$LATEST/bundle.json
```

While presenting, call out:
- The web UI is thin and local-only. It sits on top of the same Go engine as the CLI and does not change the architecture.
- Normalization derives routeability-relevant signals from source-style fields (outcomes, market_type, rules text, deadline parseability).
- Event clusters include mixed routeability members.
- Proposition clusters show explicit classifications and refusal reasons.
- `evaluation_labels.clear_non_match_case` points to an `explicit_non_match` assessment (paired cross-venue rejection), not just a single-member cluster fallback.
- `route-order` only works for proposition clusters marked `routeable`.
- In the current fixture corpus, there are currently five routeable proposition clusters:
  - `prop-001` for the Fed hike proposition
  - `prop-004` for the Liverpool-Arsenal both-teams-to-score proposition
  - `prop-007`, `prop-008`, and `prop-009` for the Liverpool vs Tottenham match outcome propositions (`draw`, `liverpool win`, `tottenham win`)
- The router currently supports hypothetical `buy_yes` and `sell_yes` orders only.
- `buy_yes` uses `yes_ask <= limit`; `sell_yes` uses `yes_bid >= limit`.
- With current fixture quotes, `prop-001 buy_yes` routes to `Polymarket`, `prop-001 sell_yes LIMIT=0.58` routes to `Kalshi`, `prop-007 buy_yes LIMIT=0.15` routes to `Kalshi`, and `prop-008` / `prop-009 buy_yes` route to `Polymarket`.

## 7) Optional live inspect
```bash
make live-inspect LIVE_LIMIT=3
```
