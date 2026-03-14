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

## 2) Use the web UI first
- Review the routeable cluster card.
- Use the order simulator form with the default values.
- Confirm that the default `buy_yes` order routes to `Polymarket`.
- Change to `sell_yes` with limit `0.55` and confirm it routes to `Kalshi`.

## 3) Run checks from the CLI
```bash
make verify
```

`make verify` runs tests plus the fixture CLI path.

## 4) Route a specific hypothetical order from the CLI
```bash
make route-order
```

You can also try:

```bash
make route-order CLUSTER=prop-001 SIDE=sell_yes LIMIT=0.55 SIZE=1000
```

## 5) Inspect artifact
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
- In the current fixture corpus, `prop-001` is the routeable proposition cluster.
- The router currently supports hypothetical `buy_yes` and `sell_yes` orders only.
- `buy_yes` uses `yes_ask <= limit`; `sell_yes` uses `yes_bid >= limit`.
- With current fixture quotes, the default `make route-order` call routes to `Polymarket`, while `SIDE=sell_yes LIMIT=0.55` routes to `Kalshi`.

## 6) Optional live inspect
```bash
make live-inspect LIVE_LIMIT=3
```
