# Demo Script

## 1) Run checks
```bash
make dev
```

`make dev` runs tests and the fixture demo in one step. The fixture demo now prints:

- the routeable proposition clusters
- example `route-order` commands
- the current routing outcomes for the default hypothetical orders

If you want a quick list of supported commands first:

```bash
make
```

## 2) Route a specific hypothetical order
```bash
make route-order
```

You can also try:

```bash
make route-order CLUSTER=prop-001 SIDE=sell_yes LIMIT=0.55 SIZE=1000
```

## 3) Inspect artifact
```bash
LATEST=$(ls -1 artifacts | tail -n 1)
cat artifacts/$LATEST/bundle.json
```

While presenting, call out:
- Normalization derives routeability-relevant signals from source-style fields (outcomes, market_type, rules text, deadline parseability).
- Event clusters include mixed routeability members.
- Proposition clusters show explicit classifications and refusal reasons.
- `evaluation_labels.clear_non_match_case` points to an `explicit_non_match` assessment (paired cross-venue rejection), not just a single-member cluster fallback.
- `route-order` only works for proposition clusters marked `routeable`.
- In the current fixture corpus, `prop-001` is the routeable proposition cluster.
- The router currently supports hypothetical `buy_yes` and `sell_yes` orders only.
- `buy_yes` uses `yes_ask <= limit`; `sell_yes` uses `yes_bid >= limit`.
- With current fixture quotes, the default `make route-order` call routes to `Polymarket`, while `SIDE=sell_yes LIMIT=0.55` routes to `Kalshi`.

## 4) Optional live inspect
```bash
make live-inspect LIVE_LIMIT=3
```
