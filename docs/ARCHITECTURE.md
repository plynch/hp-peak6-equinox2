# Architecture and Tradeoffs

## Boundaries
- `internal/adapters/*`: venue-specific payload loading (fixture, optional live inspect, and live Premier League scan).
- `internal/normalize`: venue payload -> canonical `VenueMarketInstance`.
- `internal/cluster`: canonical event and proposition clustering + equivalence assessments.
- `internal/router`: route simulation from normalized proposition clusters only.
- `internal/demo`: shared fixture snapshot construction, artifact materialization, and order simulation helpers.
- `internal/web`: thin local demo UI over the shared demo snapshot.
- `internal/store`: relational persistence boundary (SQLite file in this MVP environment).
- `internal/artifacts`: inspectable JSON artifact emission.

## Clustering-first identity model
1. Build event clusters by heuristic similarity across normalized venue instances (event-family/category checks, token-overlap in titles, deadline proximity).
2. Build proposition clusters inside each event cluster by proposition text similarity and semantic guardrails.
3. Attach venue market instances beneath proposition clusters.
4. Route only if proposition cluster is route-safe.

## Normalization posture (focused repair pass)
The primary fixture path now derives key semantic fields in code from venue-style source fields instead of copying pre-labeled normalized fields:
- proposition text normalization from question/title,
- binary and `Other` inference from outcomes/market type,
- unsupported-shape inference from outcomes/market type/rules text,
- ambiguity cue inference from rules/title text,
- deadline provenance inference from parseable deadline fields.

Remaining fixture curation is explicit and intentional (event family/category and source-like rules/quote snapshots) to preserve deterministic demo behavior.

## Non-match evidence
Artifacts include explicit `explicit_non_match` assessments for paired cross-venue markets in the same event context when proposition similarity is below threshold. This provides a legible clear non-match case beyond single-member fallback clusters.

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

## Demo surface choice
The repository now has both:
- a CLI path for deterministic checks and artifact inspection
- a thin local web UI for reviewer/demo usability
- a live EPL CLI/UI mode that exercises the same architecture against current public data

The web UI is intentionally not a separate product architecture. It is a thin layer over the same shared demo snapshot used by the CLI.

## Known limitations
- Canonical key derivation remains heuristic and fixture-calibrated.
- Live EPL routeability depends on current public overlap between Polymarket and Kalshi; the fixture path remains the deterministic fallback.
- No real execution, account logic, or fee-accurate settlement simulation.
