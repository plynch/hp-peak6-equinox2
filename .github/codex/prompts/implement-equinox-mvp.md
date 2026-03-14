Implement the Equinox MVP in this repository.

Read these files first:

1. `PROJECT_SPECIFICATION.md`
2. `PROJECT_REQUIREMENTS_DOCUMENT.md`
3. `PROJECT_PRESEARCH_DOCUMENT.md`
4. `AGENTS.md`
5. `PLANS.md`

Follow `PLANS.md` unless the specification, PRD, repository state, or cloud environment requires a small justified deviation.

Target outcome:

- a PR-ready Go implementation
- clean adapter, normalization, clustering, storage, routing, and artifact boundaries
- deterministic fixture-first demo path
- explicit ambiguity and unsupported handling
- reviewer-facing setup and demo instructions

Important clarification:

- event clustering is the core architecture
- `binary-only` applies to routeable proposition clusters, not to event clustering in general

Before finishing:

- run the relevant checks, ideally `go test ./...`
- review the diff for bugs, regressions, venue-specific routing logic, and scope drift
- update docs so they match the implementation
