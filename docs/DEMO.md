# Demo Script

## Goal

Use one concrete story:

> It is the morning of Sunday, March 15, 2026. Liverpool vs Tottenham is later today at 11:30 AM Central. I want to place a hypothetical $77 bet on Liverpool to win, and I’m using Equinox to decide where that order should route.

## Recommended 4-5 minute flow

1. Start with live discovery in the terminal:

```bash
make scan SOURCE=live-epl LIVE_MATCHWEEKS=1
```

Say:
- The system is pulling current public Polymarket and Kalshi EPL markets.
- It is clustering same-event markets across venues.
- Liverpool vs Tottenham appears in today’s live routeable slate.

2. Make the live routing decision in the terminal:

```bash
make route-order SOURCE=live-epl LIVE_MATCHWEEKS=1 EVENT_QUERY='liverpool vs tottenham' PROP_QUERY='liverpool win' LIMIT=0.76 SIZE=77
```

Say:
- This is a hypothetical `buy_yes` order.
- Size is `$77`.
- Limit is `0.76`.
- The router is choosing between normalized cross-venue candidates, not branching on venue-specific logic.

3. Switch to the deterministic app demo:

```bash
make dev
```

Then open [http://127.0.0.1:8080](http://127.0.0.1:8080).

4. In the app:
- Select the Liverpool vs Tottenham event.
- Select `liverpool win`.
- Set size to `77`.
- Keep side as `buy_yes`.
- Set limit to `0.76`.
- Run the routing simulation.
- Call out that the app and CLI use the same Go engine.

## Fallback

If the live EPL APIs change or the overlap is thin, skip the live step and go straight to:

```bash
make dev
```

Then say:
- The fixture path is the primary reviewer path.
- The live path is only there to show the same architecture operating on current public data.

## If you need one extra command

```bash
make list-clusters ROUTEABLE_ONLY=1 SOURCE=live-epl LIVE_MATCHWEEKS=1
```

Use that only if you want to explicitly show Liverpool vs Tottenham in the live cluster list before routing.
