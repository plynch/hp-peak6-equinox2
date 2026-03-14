# Demo Script

## 1) Run checks
```bash
go test ./...
```

## 2) Run fixture demo
```bash
go run ./cmd/equinox fixture-demo
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

## 4) Optional live inspect
```bash
go run ./cmd/equinox live-inspect --limit 3
```
