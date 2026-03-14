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

Explain while presenting:
- Event clusters include mixed routeability members.
- Proposition clusters show explicit classifications and refusal reasons.
- Routing decisions either select best venue or provide justified refusal.

## 4) Optional live inspect
```bash
go run ./cmd/equinox live-inspect --limit 3
```
