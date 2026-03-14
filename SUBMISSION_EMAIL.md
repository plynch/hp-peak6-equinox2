# Submission Email Draft

Hi,

I'm submitting my solution for Project Equinox.

The repository contains a local-first prototype that integrates Polymarket and Kalshi, normalizes venue data into a canonical identity model, clusters markets at both the event and proposition levels, and simulates venue-agnostic routing decisions for hypothetical orders. The core engine is Go, and the primary demo surface is a thin local web UI backed by the same fixture-first pipeline as the CLI. The fixture corpus includes a live-style Liverpool vs Tottenham Premier League event for Sunday, March 15, 2026 at 11:30 AM Central with multiple routeable match-outcome propositions across both venues.

The primary reviewer path is fixture-first for determinism and ease of review. The prototype also includes an optional live-inspect command to validate current public ingestion availability.

Suggested review order:

1. `SUBMISSION.md`
2. `README.md`
3. `docs/ARCHITECTURE.md`
4. `docs/DEMO.md`

Suggested commands:

```bash
make dev
make verify
make list-clusters ROUTEABLE_ONLY=1
make live-inspect LIVE_LIMIT=1
```

The implementation intentionally prioritizes architectural clarity, ambiguity handling, and explainability over production polish.

Thanks for your time.

Best,

Patrick
