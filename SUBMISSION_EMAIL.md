# Submission Email Draft

Hi,

I'm submitting my solution for Project Equinox.

The repository contains a local-first prototype that integrates Polymarket and Kalshi, normalizes venue data into a canonical identity model, clusters markets at both the event and proposition levels, and simulates venue-agnostic routing decisions for hypothetical orders. The core engine is Go, and the primary demo surface is a thin local web UI backed by the same fixture-first pipeline as the CLI. In addition to the deterministic fixture demo, the repo now includes an ongoing live Premier League scan that fetches the current upcoming slate from both venues and simulates routing across every routeable proposition cluster it can match.

The primary reviewer path is fixture-first for determinism and ease of review. The prototype also includes an optional live EPL batch command and a live-inspect command to validate current public ingestion availability.

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
make live-epl
make live-inspect LIVE_LIMIT=1
```

The implementation intentionally prioritizes architectural clarity, ambiguity handling, and explainability over production polish.

Thanks for your time.

Best,

Patrick
