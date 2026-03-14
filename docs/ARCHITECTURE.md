# Architecture and Tradeoffs

## Boundaries
- `internal/adapters/*`: venue-specific payload loading (fixture + optional live inspect).
- `internal/normalize`: venue payload -> canonical `VenueMarketInstance`.
- `internal/cluster`: canonical event and proposition clustering + equivalence assessments.
- `internal/router`: route simulation from normalized proposition clusters only.
- `internal/store`: relational persistence boundary (SQLite file in this MVP environment).
- `internal/artifacts`: inspectable JSON artifact emission.

## Clustering-first identity model
1. Build event clusters by heuristic similarity across normalized venue instances (event-family/category checks, token-overlap in titles, deadline proximity).
2. Build proposition clusters inside each event cluster by proposition text similarity and semantic guardrails.
3. Attach venue market instances beneath proposition clusters.
4. Route only if proposition cluster is route-safe.

## Routeability policy
Routeability is denied if any member is:
- unsupported shape,
- non-binary family,
- Other/bucket/combo style,
- or ambiguous on deadline/resolution semantics.

Event clustering can still include these members. Binary-only applies to routing/proposition safety, not event clustering breadth.

## Persistence choice
PRD prefers PostgreSQL. In this execution environment, the MVP uses SQLite as an embedded relational fallback so the prototype remains deterministic and local-first while preserving a clean relational boundary (`internal/store`).

## Venue-agnostic router guarantee
Router scoring uses normalized quote/depth fields from proposition clusters and does not branch on venue names or venue-native payload shapes. Router also hard-rejects non-executable candidates that violate the hypothetical order limit.

## Known limitations
- Canonical key derivation is heuristic and fixture-calibrated.
- Live inspect path validates ingestion availability only; it does not guarantee live routeable overlaps.
- No real execution, account logic, or fee-accurate settlement simulation.
