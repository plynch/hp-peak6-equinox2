# Submission Email Draft

Hi,

I'm submitting my solution for Project Equinox.

The repository contains a CLI-first, local-first prototype that integrates Polymarket and Kalshi, normalizes venue data into a canonical identity model, clusters markets at both the event and proposition levels, and simulates venue-agnostic routing decisions for hypothetical orders.

The primary reviewer path is fixture-first for determinism and ease of review. The prototype also includes an optional live-inspect command to validate current public ingestion availability.

Suggested review order:

1. `SUBMISSION.md`
2. `README.md`
3. `docs/ARCHITECTURE.md`
4. `docs/DEMO.md`

Suggested commands:

```bash
go test ./...
go run ./cmd/equinox fixture-demo
go run ./cmd/equinox live-inspect --limit 1
```

The implementation intentionally prioritizes architectural clarity, ambiguity handling, and explainability over production polish.

Thanks for your time.

Best,

Patrick
