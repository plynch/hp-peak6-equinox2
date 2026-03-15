# Demo Script

## Recommended filmed flow for Sunday, March 15, 2026

If you are filming on the morning of Sunday, March 15, 2026 before the Liverpool vs Tottenham Premier League match at 11:30 AM Central, use this order:

1. `make scan SOURCE=live-epl LIVE_MATCHWEEKS=1`
2. say the concrete user story out loud:
   - "It is Sunday morning, March 15, 2026."
   - "Liverpool vs Tottenham is later today at 11:30 AM Central."
   - "I want to place a hypothetical $77 bet on Liverpool to win."
   - "I am going to use Equinox to make a routing decision for that order."
3. call out that the live public APIs currently show a routeable Liverpool vs Tottenham event cluster for the same-day match
4. optionally run:
   - `make list-clusters ROUTEABLE_ONLY=1 SOURCE=live-epl LIVE_MATCHWEEKS=1`
   - `make route-order SOURCE=live-epl LIVE_MATCHWEEKS=1 EVENT_QUERY='liverpool vs tottenham' PROP_QUERY='liverpool win' LIMIT=0.76 SIZE=77`
5. then switch to the deterministic path with `make dev`

That gives you an organic live discovery moment first, then a stable demo path second.

If the live slate or venue overlap changes before filming, fall back to `make dev` immediately and say the deterministic fixture path is the primary reviewer path.

## Recommended one-sentence framing

> "I want to place a hypothetical $77 Liverpool win order on the morning of Sunday, March 15, 2026, and I’m using Equinox to figure out which venue currently offers the best executable routing outcome."

## 1) Start the deterministic local demo
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

If you want the terminal-first full showcase across every supported source:

```bash
make demo-cli
```

## 2) Optional: run the live Fed scan
```bash
make live-fed
```

This fetches the current plus next 4 open Fed meetings from Polymarket and Kalshi, builds live event/proposition clusters, and simulates routing across every routeable rate-decision proposition it can match.

For the browser version of that same live path:

```bash
make dev-live-fed
```

Use the live Fed path to show that the same architecture can operate continuously on FOMC markets, not just sports.
If you want fewer or more meetings, use `FED_MEETINGS=<n>`.

## 3) Optional: run the live Premier League scan
```bash
make live-epl
```

This fetches the current plus next 4 Premier League matchweek-style windows from Polymarket and Kalshi, builds live event/proposition clusters, and simulates routing across every routeable proposition cluster it finds.

For the browser version of that same live path:

```bash
make dev-live-epl
```

Use the live path when the public APIs are stable and you want to show the ongoing operating model, not just the deterministic fixture demo.
If you want to widen or narrow the window, use `LIVE_MATCHWEEKS=<n>`.

For the Sunday, March 15, 2026 filmed demo, prefer this narrower live-discovery pass:

```bash
make scan SOURCE=live-epl LIVE_MATCHWEEKS=1
```

On the morning of Sunday, March 15, 2026, this should organically surface the same-day Liverpool vs Tottenham match in the live routeable cluster list if both venue APIs still expose it.

## 4) Inspect routeable clusters from the CLI
```bash
make list-clusters ROUTEABLE_ONLY=1
```

This is the explicit selection step before routing a CLI order.

For live Fed:

```bash
make list-clusters ROUTEABLE_ONLY=1 SOURCE=live-fed FED_MEETINGS=2
```

For live EPL:

```bash
make list-clusters ROUTEABLE_ONLY=1 SOURCE=live-epl LIVE_MATCHWEEKS=1
```

That command is useful right after `make scan SOURCE=live-epl LIVE_MATCHWEEKS=1` if you want to explicitly show that Liverpool vs Tottenham has three routeable proposition clusters:
- `draw`
- `liverpool win`
- `tottenham win`

## 5) Use the web UI first
- Review the routeable cluster card.
- Use the order simulator form with the default values.
- Confirm that the default `buy_yes` order routes to `Polymarket`.
- Change to `sell_yes` with limit `0.58` on the Fed cluster and confirm it routes to `Kalshi`.
- Switch to the Liverpool vs Tottenham event, scheduled for Sunday, March 15, 2026 at 11:30 AM Central.
- In the order form, set the order size to `77`.
- Keep the side as `buy_yes`.
- Set the limit to `0.76`.
- Choose the `liverpool win` proposition.
- Use the app to make the routing decision for that hypothetical $77 order.
- Confirm that the current routing decision goes to `Polymarket`.
- Optionally switch to `draw` at limit `0.15` and show that the current routing decision goes to `Kalshi`.

## 6) Run checks from the CLI
```bash
make verify
```

`make verify` runs tests plus the fixture CLI path.

## 7) Route specific hypothetical orders from the CLI
```bash
make route-order CLUSTER=prop-001
```

You can also try:

```bash
make route-order CLUSTER=prop-001 SIDE=sell_yes LIMIT=0.58 SIZE=1000
make route-order CLUSTER=prop-007 SIDE=buy_yes LIMIT=0.15 SIZE=1000
make route-order CLUSTER=prop-008 SIDE=buy_yes LIMIT=0.76 SIZE=77
make route-order CLUSTER=prop-009 SIDE=buy_yes LIMIT=0.10 SIZE=1000
make route-order EVENT_QUERY='fomc march 2026' PROP_QUERY='fed hike rate march meeting' LIMIT=0.60 SIZE=1000
make route-order EVENT_QUERY='liverpool vs tottenham' PROP_QUERY='liverpool win' LIMIT=0.76 SIZE=77
make route-order SOURCE=live-epl LIVE_MATCHWEEKS=1 EVENT_QUERY='liverpool vs tottenham' PROP_QUERY='draw' LIMIT=0.15 SIZE=77
make route-order SOURCE=live-epl LIVE_MATCHWEEKS=1 EVENT_QUERY='liverpool vs tottenham' PROP_QUERY='liverpool win' LIMIT=0.76 SIZE=77
```

The selector-based form is better for the terminal demo because it does not rely on internal `prop-00x` ids.
For live data, prefer `make scan SOURCE=live-fed ...` or `make scan SOURCE=live-epl ...` first, then copy the selector-ready routing command printed by the scan itself.

For the Sunday, March 15, 2026 filmed demo, the strongest live terminal sequence is:

```bash
make scan SOURCE=live-epl LIVE_MATCHWEEKS=1
make route-order SOURCE=live-epl LIVE_MATCHWEEKS=1 EVENT_QUERY='liverpool vs tottenham' PROP_QUERY='liverpool win' LIMIT=0.76 SIZE=77
```

That sequence looks organic because the event is discovered from current public data first, then the $77 Liverpool win order is routed from the terminal second.

## 8) Inspect artifact
```bash
LATEST=$(ls -1 artifacts | tail -n 1)
cat artifacts/$LATEST/bundle.json
```

While presenting, call out:
- The web UI is thin and local-only. It sits on top of the same Go engine as the CLI and does not change the architecture.
- `make live-fed`, `make dev-live-fed`, `make live-epl`, and `make dev-live-epl` use the same normalization/clustering/routing engine as the fixture path; only the data source changes.
- The live EPL path approximates matchweeks from fixture dates because the public APIs expose dates reliably but do not expose a stable official matchweek field.
- The live Fed path fetches the current plus next few open meetings, then routes only exact bucket semantics it can justify as route-safe across venues.
- On the morning of Sunday, March 15, 2026, `make scan SOURCE=live-epl LIVE_MATCHWEEKS=1` should surface Liverpool vs Tottenham as a same-day routeable event cluster if both venue APIs still have those markets open.
- Normalization derives routeability-relevant signals from source-style fields (outcomes, market_type, rules text, deadline parseability).
- Event clusters include mixed routeability members.
- Proposition clusters show explicit classifications and refusal reasons.
- `evaluation_labels.clear_non_match_case` points to an `explicit_non_match` assessment (paired cross-venue rejection), not just a single-member cluster fallback.
- `route-order` only works for proposition clusters marked `routeable`.
- `route-order` can resolve by human-readable event and proposition selectors, which is the cleaner demo path in the terminal.
- For the filmed sports story, the concrete hypothetical order is: `buy_yes`, `liverpool win`, `limit=0.76`, `size=77`, on Sunday, March 15, 2026 before kickoff.
- In the current fixture corpus, there are currently five routeable proposition clusters:
  - `prop-001` for the Fed hike proposition
  - `prop-004` for the Liverpool-Arsenal both-teams-to-score proposition
  - `prop-007`, `prop-008`, and `prop-009` for the Liverpool vs Tottenham match outcome propositions (`draw`, `liverpool win`, `tottenham win`)
- In the live Fed and live EPL scans, the number of routeable proposition clusters is dynamic and depends on the current overlapping open slate. The live commands print them explicitly each run.
- For the filmed demo, use exact absolute dates when talking about the sports example: Sunday, March 15, 2026 at 11:30 AM Central for Liverpool vs Tottenham.
- The router currently supports hypothetical `buy_yes` and `sell_yes` orders only.
- `buy_yes` uses `yes_ask <= limit`; `sell_yes` uses `yes_bid >= limit`.
- With current fixture quotes, `prop-001 buy_yes` routes to `Polymarket`, `prop-001 sell_yes LIMIT=0.58` routes to `Kalshi`, `prop-007 buy_yes LIMIT=0.15` routes to `Kalshi`, and `prop-008` / `prop-009 buy_yes` route to `Polymarket`.

## 9) Optional live inspect
```bash
make live-inspect LIVE_LIMIT=3
```
