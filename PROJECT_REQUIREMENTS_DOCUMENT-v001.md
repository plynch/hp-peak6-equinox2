# Project Requirements Document

## Document Metadata

- Project: Equinox
- Version: v001
- Status: Draft
- Last Updated: 2026-03-14
- Source of Truth for This Revision: `PROJECT_SPECIFICATION.md`
- Research companion: `PROJECT_PRESEARCH_DOCUMENT.md`

## 1. Executive Summary

Project Equinox is an infrastructure prototype to test whether prediction markets from different venues can be ingested, normalized, compared, and used to simulate venue routing decisions for hypothetical orders.

The product is not a trading system. Its purpose is to answer a narrower question: can we define a venue-agnostic market model, identify materially equivalent propositions across venues, and explain why a simulated order would route to one venue instead of another?

This PRD defines an MVP that:

- Connects to at least two public prediction market APIs.
- Normalizes venue-specific schemas into a canonical internal representation.
- Attempts cross-venue equivalence detection at the event-plus-outcome level.
- Simulates routing for hypothetical orders using normalized pricing and liquidity data.
- Produces structured explanations for matches, non-matches, and routing decisions.

## 2. Product Goal

Determine whether cross-venue normalization and routing infrastructure is viable enough to justify deeper investment.

## 3. Problem Statement

Prediction market venues fragment the same real-world question across different schemas, naming conventions, resolution rules, expirations, outcome structures, and pricing formats. Two contracts may look similar but resolve under different conditions. Without normalization and equivalence logic, price comparison and routing are unreliable.

The core product problem is not only finding similar-looking markets. It is distinguishing:

- truly equivalent propositions,
- near-matches that should not be routed interchangeably,
- and venue-specific contracts with no valid cross-venue analog.

## 4. Goals and Non-Goals

### Goals

- Prove that at least two venues can be integrated behind a clean adapter boundary.
- Define a canonical market model that is independent of venue schemas.
- Detect and score candidate equivalent markets across venues.
- Simulate routing for hypothetical orders without venue-specific logic in the router.
- Preserve a path to adding a third venue without rewriting router logic or hard-coding venue-pair-specific matcher behavior.
- Explain every important decision in a way a reviewer can inspect.
- Handle missing, inconsistent, or ambiguous data without crashing or silently overconfident behavior.

### Non-Goals

- Real-money trading or order execution.
- Wallet, brokerage, or account integration.
- Regulatory or compliance implementation.
- Production-grade UI.
- Production-grade matching accuracy guarantees.
- Full support for every prediction market structure on every venue.

## 5. Primary Users

- Candidate or engineer building the prototype and validating architectural choices.
- Reviewer evaluating decomposition, tradeoff awareness, and decision quality.
- Research operator running ingestion, inspecting matches, and executing routing simulations locally.

## 6. Product Principles

- Explainability over raw optimization.
- Correct rejection over false-positive matching.
- Clear module boundaries over premature sophistication.
- Reproducible local execution over infrastructure complexity.
- Deterministic heuristics first; optional AI only as an explicitly logged extension.

## 7. MVP Scope

### In Scope

- Public API ingestion for at least two venues.
- Market metadata ingestion.
- Price and top-of-book or depth ingestion where available.
- Canonical event and outcome normalization.
- Cross-venue candidate generation and equivalence scoring.
- Hypothetical routing for supported order types.
- Human-readable and machine-readable explanations.
- Local execution and setup instructions.

### Out of Scope

- Live trading.
- Real-time streaming infrastructure unless a venue makes polling impractical.
- Portfolio management.
- Payout, settlement, or balance accounting.
- Frontend product polish.

## 8. v001 Product Decisions

### 8.1 Matching Unit

The canonical matching unit is `event + outcome`, not raw venue market ID.

Reason:

- One venue may represent a multi-outcome event as many binary markets.
- Another may represent the same event as a grouped market with multiple selectable outcomes.
- Routing only makes sense when the compared exposures represent the same truth condition.

### 8.2 Market Types for MVP

The MVP must support binary outcomes first-class.

Grouped multi-outcome venues are allowed if they can be decomposed into binary outcome propositions such as:

- "Aaron Taylor-Johnson announced as next James Bond?"
- "Hakeem Jeffries is next Speaker of the House?"

Unsupported market structures may be ingested but must be labeled unsupported for matching or routing rather than forced into a bad mapping.

### 8.3 Venue Strategy

Preferred initial venue pair:

- Polymarket
- Kalshi

Why:

- Both expose public market metadata and current pricing data.
- Both create a meaningful normalization challenge.
- Both are strong signals for architectural thinking because they differ in event modeling and contract grouping.

Risk:

- Live exact overlap appears limited and many tempting pairs are only near-equivalent.

Research-backed decision:

- Manifold should be treated as a secondary research/control venue, not the default routing fallback.

Fallback:

- If live overlap between Polymarket and Kalshi is too thin for a convincing demo, the preferred fallback is fixture mode using saved snapshots from the same primary pair.
- Manifold may still be used to expand the labeled evaluation set or to test semantic matching behavior across a very different venue model.

### 8.4 Live and Fixture Modes

The product must support:

- `live mode`: pull current data from public APIs.
- `fixture mode`: run against saved snapshots and labeled evaluation cases.

Reason:

- live overlap changes over time,
- APIs evolve,
- reviewers need deterministic demos and tests,
- and primary-venue exact matches may need to be demonstrated from saved snapshots rather than whatever is live that day.

### 8.5 AI Usage Policy

The default MVP path should not require AI in the critical path.

If AI is used, it must be:

- optional,
- behind a distinct interface,
- invoked only for low-confidence disambiguation or explanation enrichment,
- fully logged,
- and disclosed in the final deliverables.

### 8.6 Technical Flexibility

This PRD does not prescribe a specific language, framework, database, or cloud environment.

Constraints:

- local deployment must be supported,
- architectural boundaries must stay clear,
- and any chosen stack must be justified by simplicity, readability, and speed of execution for a prototype.

## 9. Definition of "Equivalent"

Equinox should distinguish **semantic equivalence** from **route-safe equivalence**.

Two venue contracts are semantically equivalent only if a unit long exposure to each resolves on the same underlying truth condition with materially aligned timing and resolution semantics.

Two venue contracts are route-safe equivalent only if they are semantically equivalent **and** their venue-level risk differences are acceptable for routing comparison in this prototype.

For MVP purposes, semantic equivalence requires all of the following:

1. Same underlying real-world subject or event.
2. Same asserted outcome or proposition.
3. Same effective resolution window or an acceptable bounded difference that does not change truth conditions.
4. No conflict in resolution semantics that could produce different outcomes for the same real-world facts.
5. Compatible outcome polarity and contract direction after normalization.

For MVP purposes, route-safe equivalence additionally requires review of:

1. Resolution authority and governance model.
2. Economic domain comparability.
3. Quote-model comparability and data confidence.
4. Fee-model visibility sufficient for the requested routing simulation.

The system must classify comparisons into:

- `semantic_equivalent_and_route_safe`
- `semantic_equivalent_but_route_unsafe`
- `near_equivalent_but_not_safe`
- `not_equivalent`
- `insufficient_data`

Examples:

- Same person, same office, same deadline: potentially equivalent.
- Same proposition text but creator-resolved play-money venue versus oracle-resolved real-money venue: potentially semantically equivalent but route-unsafe.
- Same topic but different deadline: near-equivalent, not safe.
- Same event but party-control versus named-speaker outcome: not equivalent.

## 10. Functional Requirements

### Venue Integration

- `FR-1`: The system must integrate with at least two public prediction market venue APIs.
- `FR-2`: Each venue integration must be isolated behind an adapter interface that handles fetching, parsing, and venue-specific field mapping.
- `FR-3`: Adapters must ingest both metadata and pricing data for supported markets.
- `FR-4`: Adapters must tolerate missing optional fields, temporary API failures, and pagination differences.

### Normalization

- `FR-5`: The system must normalize venue data into a canonical internal representation that is independent of venue schemas.
- `FR-6`: The canonical model must preserve source provenance so every normalized record can be traced back to original venue identifiers and raw payloads.
- `FR-7`: The normalizer must convert venue-specific pricing into a shared probability-oriented representation where possible.
- `FR-8`: The normalizer must identify unsupported market structures explicitly instead of silently coercing them.
- `FR-8a`: The normalizer must capture semantic deadline, deadline source, and deadline confidence rather than relying on a single raw timestamp field.
- `FR-8b`: The normalizer must capture resolution authority, governance model, and economic domain for each supported venue contract.
- `FR-8c`: The system must maintain a venue capability profile describing quote model, fee visibility, supported market shapes, and historical-data behavior.
- `FR-8d`: The normalizer must preserve contract composition type so simple binary, bucketed-range, combo/parlay-like, scalar-derived, and placeholder-driven contracts are not conflated.

### Equivalence Detection

- `FR-9`: The system must generate cross-venue candidate matches using normalized data.
- `FR-9a`: The canonical matching model must not assume that only two venue contracts can map to the same proposition, even if the MVP internally evaluates candidate pairs.
- `FR-10`: The system must score candidate pairs using deterministic signals such as text similarity, entity overlap, event category, outcome mapping, deadline alignment, and resolution-rule compatibility.
- `FR-11`: The system must emit a match classification and confidence score for each evaluated pair.
- `FR-12`: The system must produce explanations for both accepted matches and rejected near-matches.
- `FR-12a`: The matcher must be able to classify a pair as semantically equivalent but route-unsafe when governance, economic domain, or quote comparability makes routing unsafe.

### Routing Simulation

- `FR-13`: The system must accept a hypothetical order input against a canonical proposition.
- `FR-14`: The router must compare all matched venue options using normalized pricing, available liquidity or depth, and data confidence.
- `FR-15`: The router must not contain hard-coded venue-specific decision logic.
- `FR-16`: The router must produce a chosen venue or a `do_not_route` result when no safe match exists.
- `FR-17`: The router must explain the decision in structured and human-readable form.
- `FR-17a`: The router must apply venue capability differences through normalized penalties or blockers, including quote-model limits, governance risk, and fee visibility limits.

### Logging and Inspection

- `FR-18`: The system must log raw ingestion outcomes, normalization output, equivalence decisions, and routing decisions.
- `FR-19`: The system must preserve enough evidence to reconstruct why a decision was made.
- `FR-20`: The prototype must expose outputs through a local developer-friendly interface such as CLI commands, JSON artifacts, or a minimal local API.

### Reproducibility

- `FR-21`: The project must include fixture snapshots from supported venues for deterministic demo and test runs.
- `FR-22`: The project must include a labeled evaluation set with at least exact-match, near-match, and non-match examples.
- `FR-23`: Fixture mode must be runnable without external credentials or live API availability.
- `FR-24`: The repository must define a single clear reviewer demo path that exercises ingestion, normalization, matching, and routing in a deterministic way.

## 11. Canonical Data Model Requirements

The product must define and use the following logical entities.

### Venue Capability Profile

Required fields:

- `venue`
- `economic_domain`
- `resolution_authority`
- `resolution_governance`
- `quote_model`
- `supports_direct_asks`
- `supports_depth`
- `supports_historical_backfill`
- `supported_market_shapes`
- `fee_model_visibility`

### Canonical Event

Required fields:

- `canonical_event_id`
- `title`
- `category`
- `entities`
- `event_time_window`
- `resolution_summary`
- `source_venues`

### Canonical Proposition

A proposition is the routeable unit.

Required fields:

- `canonical_proposition_id`
- `canonical_event_id`
- `proposition_text`
- `outcome_label`
- `polarity`
- `semantic_deadline`
- `semantic_deadline_source`
- `semantic_deadline_confidence`
- `resolution_criteria_summary`
- `resolution_authority`
- `resolution_governance`
- `economic_domain`
- `exclusivity_type`
- `market_structure`
- `composition_type`
- `supported_for_routing`
- `source_contracts`

### Canonical Quote

Required fields:

- `venue`
- `source_contract_id`
- `timestamp`
- `best_bid`
- `best_ask`
- `mid_probability`
- `depth_summary`
- `quote_model`
- `best_ask_is_implied`
- `fee_model_reference`
- `quote_confidence`

### Match Record

Required fields:

- `left_proposition_id`
- `right_proposition_id`
- `classification`
- `score`
- `supporting_signals`
- `blocking_signals`
- `explanation`

## 12. Local Deployment and Setup Requirements

The MVP must be packaged for local execution from the project root.

Required setup characteristics:

- A reviewer must be able to run the fixture-backed demo with no secrets.
- Live mode may require optional environment variables only if a chosen venue later requires them.
- The repository must document installation, bootstrap, and run steps in a single primary entry document, preferably `README.md`.
- The project should prefer a low-friction local workflow such as direct CLI commands, a task runner, or a lightweight container wrapper.

Recommended posture:

- fixture mode is the default first-run experience,
- live mode is a secondary run path,
- and cloud deployment should be treated as unnecessary unless it materially simplifies local execution.

Recommended interface posture:

- prefer a CLI-first demo surface,
- emit machine-readable artifacts such as JSON files for match and route outputs,
- and keep the reviewer path short enough that a clean fixture demo can be run in a few commands from the project root.

## 13. Equivalence Detection Methodology

The MVP matcher should use a staged pipeline.

### Stage 1: Candidate Generation

Generate a manageable set of possible matches using:

- category alignment,
- date-window overlap,
- named entity overlap,
- normalized token similarity,
- market structure compatibility,
- and venue capability compatibility.

Implementation note:

- The matcher should be designed so a new venue can attach contracts to an existing canonical proposition registry or cluster, rather than forcing the system into permanent venue-pair-specific comparison paths.
- Simple pairwise evaluation is acceptable for MVP-scale data if the data model and explanations do not lock the architecture into pair-specific assumptions.

### Stage 2: Outcome Mapping

Try to map venue-specific outcomes into canonical propositions.

Examples:

- grouped outcome on one venue to binary actor-specific contract on another,
- binary yes contract to complement-aware yes or no orientation on another venue.

### Stage 3: Semantic Safety Checks

Reject or downgrade pairs when:

- deadlines differ materially,
- resolution sources imply materially different rules,
- field conflicts make semantic deadline uncertain,
- one contract represents party control while the other represents a named person,
- one contract includes "other" or fallback semantics that the other venue lacks,
- one venue resolves on first occurrence and the other resolves at final state,
- or governance / economic-domain differences make routing unsafe even if the proposition is semantically aligned.

### Stage 4: Final Classification

Apply thresholds to classify as:

- semantically equivalent and route-safe,
- semantically equivalent but route-unsafe,
- near-match for analyst review only,
- not equivalent,
- insufficient data.

### Key Requirement

The system is allowed to find that many similar-looking markets are not safely equivalent. That is a successful outcome if the reasoning is clear and defensible.

## 14. Routing Simulation Methodology

### Supported MVP Order Inputs

The routing simulator must support at least:

- canonical proposition identifier,
- order side,
- target size or notional,
- quote timestamp or latest available quote.

Recommended MVP order scope:

- `BUY_YES`
- `BUY_NO`

`SELL` may be deferred unless needed for the final prototype demo.

### Routing Decision Factors

The router should evaluate:

- effective entry price for the requested exposure,
- available displayed liquidity or estimated fill capacity,
- quote freshness,
- venue data quality,
- match confidence,
- fee-model visibility,
- and economic / governance comparability.

For this prototype, explainability and defensible structure matter more than execution-quality optimization.

### Routing Outcomes

The router must return one of:

- `route_to_<venue>`
- `do_not_route`

### Explanation Requirements

Each routing result must explain:

- which venue options were considered,
- what quote and liquidity data was used,
- what penalties or adjustments were applied,
- and why the winner was selected or why routing was refused.

## 15. Explainability Requirements

Every major decision must be inspectable without reading source code.

Required explanation surfaces:

- why a source market was normalized in a specific way,
- why a specific deadline source was trusted over conflicting raw fields,
- why two propositions were considered equivalent or not,
- why a venue won the routing simulation,
- and what missing or ambiguous data reduced confidence.

Required explanation format:

- concise narrative summary,
- structured machine-readable fields,
- and raw source references.

## 16. Non-Functional Requirements

- `NFR-1`: The prototype must run locally with documented setup instructions.
- `NFR-2`: The codebase must maintain strict separation between adapters, normalizers, matcher, and router.
- `NFR-3`: The system must degrade gracefully under partial API failure.
- `NFR-4`: The prototype must favor readability and modularity over throughput optimization.
- `NFR-5`: Time assumptions, venue assumptions, and unsupported cases must be documented.
- `NFR-6`: Tests should cover normalization, equivalence classification, and routing logic using fixtures.
- `NFR-7`: The system should support adding another venue with limited impact outside the adapter and mapping layers.
- `NFR-8`: Conflicting venue fields must be surfaced explicitly in logs or output artifacts rather than silently overwritten.
- `NFR-9`: Adding a third venue should primarily require a new adapter, capability profile, and mapping logic rather than router changes or pair-specific branching.
- `NFR-10`: Candidate generation should preserve a path beyond a naive all-versus-all pairwise scan; the MVP may use simpler pairwise heuristics if sampled data volume remains small.
- `NFR-11`: The default reviewer path should work entirely in local execution from the repository root without requiring cloud infrastructure.

## 17. Success Criteria

The MVP is successful if it can:

- ingest current data from at least two venues,
- normalize supported contracts into a shared model,
- produce a defensible set of exact matches, near-matches, and non-matches,
- simulate at least one meaningful cross-venue routing decision for an exact-equivalent proposition,
- and clearly explain both positive and negative decisions.

The MVP is still considered successful if the final conclusion is that exact cross-venue routing opportunities are sparse, provided the system demonstrates that conclusion with evidence.

## 18. Delivery and Demo Plan

The final submission should be demoable in two modes.

### Fixture Demo

Purpose:

- deterministic reviewer walkthrough,
- no dependency on live market overlap,
- no credential requirement.

Minimum expected flow:

1. Run one documented command or short command sequence from the project root.
2. Load saved venue snapshots.
3. Show normalized propositions and a small labeled match set.
4. Execute at least one routing simulation.
5. Emit a human-readable explanation and machine-readable artifact for the result.

### Live Demo

Purpose:

- show that the adapters can ingest current public venue data,
- and demonstrate that the architecture is not fixture-only.

Minimum expected flow:

1. Run a documented live-ingestion command from the project root.
2. Fetch current metadata and pricing from the supported venues.
3. Show either:
   - one live exact-equivalent routeable proposition, or
   - evidence that the current live sample mostly yields near-matches or non-matches.

Success note:

- The live demo does not need to guarantee a route every time.
- It does need to prove the ingestion and evaluation pipeline works on current public data.

### Reviewer-Facing Artifacts

The final repository should include:

- one primary setup and run document, preferably `README.md`,
- one brief architecture overview, preferably a dedicated markdown file,
- a short justification of the chosen stack and local execution approach,
- one demo walkthrough section or script reference,
- fixture snapshots and labeled evaluation cases,
- and example output artifacts for fixture or live runs.

The architecture overview should minimally explain:

- component boundaries,
- end-to-end data flow from ingestion to routing decision,
- and why the chosen stack is appropriate for a local prototype.

## 19. Deliverables Required from the Final Implementation

- Working prototype.
- Setup and run instructions.
- Brief architecture overview.
- Short justification of chosen stack and local execution approach.
- Demo walkthrough or demo command sequence.
- Written explanation of normalization logic.
- Written explanation of equivalence logic.
- Written explanation of routing logic.
- Disclosure of any AI tooling used during development or runtime.

## 20. Risks and Open Questions

### R1. Low Exact-Overlap Risk

Current live research suggests many cross-venue pairs are thematically similar but not safely equivalent.

Mitigation:

- require a curated evaluation set,
- require fixture mode,
- and prioritize primary-pair snapshots over changing the routing pair.

### R2. Resolution Semantics Drift

Two contracts can share nouns and still resolve differently.

Mitigation:

- resolution criteria must be parsed into the canonical model,
- and semantic mismatches must block routing.

### R3. Incomplete Liquidity Data

Some venues expose richer depth than others.

Mitigation:

- normalize what is available,
- attach quote-confidence metadata,
- and refuse overconfident routing.

### R4. API Instability

Public endpoints, auth expectations, and pagination behavior may change.

Mitigation:

- isolate adapters,
- support fixture mode,
- and store raw payloads for debugging.

### R5. Optional AI Scope Creep

An AI matcher could make the system less reproducible.

Mitigation:

- keep heuristics as the source of truth,
- and treat AI as an optional shadow scorer or explainer only.

### R6. Field Reliability Risk

Important venue fields may be incomplete, inconsistent, or weaker than the natural-language rules.

Mitigation:

- parse title and description alongside structured fields,
- store deadline source and confidence,
- and block route-safe matching when key fields conflict.

### R7. Economic and Governance Mismatch

Two venues can carry the same question text while still being poor routing peers because of different economic models or resolution authorities.

Mitigation:

- separate semantic equivalence from route safety,
- and require explicit venue capability review before routing across materially different venues.

### R8. Pairwise Scaling Risk

If the matcher is implemented only as venue-A versus venue-B logic, complexity and inconsistency will grow quickly when a third or fourth venue is added.

Mitigation:

- use a canonical proposition registry,
- keep venue capability profiles normalized,
- and make candidate generation cluster-friendly rather than pair-specific.

### R9. Demo Fragility Risk

A prototype can satisfy the architecture goals and still demo poorly if the reviewer path depends on unstable live overlap, undocumented setup steps, or too many commands.

Mitigation:

- make fixture mode the default first-run path,
- keep the demo command sequence short,
- and include saved example outputs in the repository.

## 21. Recommended Research Before v002

- Run a targeted overlap study across Polymarket, Kalshi, and Manifold to quantify exact versus near-equivalent match yield.
- Decide the minimum acceptable number of exact-equivalent propositions for the routing demo from the primary venue pair.
- Confirm which venues expose enough quote depth for size-aware simulation.
- Define the initial labeled evaluation set and error taxonomy.
- Validate fee handling assumptions for any venue used in routing comparison.
- Evaluate whether Myriad or PredictIt is the best third-venue research candidate after the primary Polymarket plus Kalshi pair.
- Define route-comparability domains for any future multi-venue expansion, such as real-money exchange, onchain/tokenized, and play-money/social.

## 22. Acceptance Criteria

The implementation will be accepted against this PRD when all of the following are true:

1. Two venue adapters fetch live public data successfully in local execution.
2. The normalized model can represent supported contracts from both venues without leaking venue-specific logic into the router.
3. The system outputs at least one labeled example each of `semantic_equivalent_and_route_safe` or fixture-based equivalent, `semantic_equivalent_but_route_unsafe`, `near_equivalent_but_not_safe`, and `not_equivalent`.
4. At least one exact-equivalent proposition can be routed across two venues in either live or fixture mode.
5. Every accepted or rejected match includes a readable explanation.
6. Every routing decision includes the compared venues, quote inputs, and selection reason.
7. The repository includes setup instructions and fixtures that allow deterministic re-runs.
8. At least one example demonstrates conflict handling when structured venue fields and natural-language rules disagree.
9. Unsupported composite, bucketed, scalar-derived, or placeholder-driven contracts are surfaced explicitly rather than treated as simple binary equivalents.
10. A reviewer can run the default fixture-backed demo from the project root without secrets.
11. The repository documents a live-data demo path even if the live sample yields no safe route.

## 23. Specification Traceability

| Spec Requirement | PRD Coverage |
| --- | --- |
| Connect to at least two venues | `FR-1`, `FR-2`, `FR-3` |
| Define canonical internal market model | `FR-5`, `FR-6`, Section 11 |
| Attempt equivalence detection | `FR-9` to `FR-12`, Section 13 |
| Simulate routing decisions | `FR-13` to `FR-17`, Section 14 |
| Log and explain decisions | `FR-18` to `FR-20`, Section 15 |
| Handle imperfect data gracefully | `FR-4`, `FR-8`, `NFR-3`, Section 20 |
| Clear separation of concerns | `NFR-2`, Sections 10 to 14 |
| Routing logic without venue-specific assumptions | `FR-15` |
| Document what "equivalent" means | Section 9 |
| AI optional but disclosed if used | Section 8.5, Section 19 |
| Local deployment acceptable | `FR-23`, `NFR-1`, Section 12, Section 18 |
| Working prototype plus setup, architecture, and logic writeups | Sections 18 and 19 |
| Technical flexibility in tooling and infrastructure | Section 8.6 |
| Tradeoff awareness around ambiguity and real-world mismatch | Sections 9, 13, 20 |

## 24. Research Appendix

### 24.1 Observed Venue Patterns on 2026-03-14

- Polymarket exposes public event and market metadata plus CLOB order book data.
- Kalshi exposes public event, market, and order book endpoints through its trade API.
- Manifold exposes broad public market search and market detail APIs but uses a play-money, creator-resolved model.
- Live inspection showed that similar themes often differ in resolution deadline, governance model, or event semantics.

### 24.2 Important Example from Research

Observed near-match:

- Polymarket event: `Next James Bond actor?`
- Polymarket proposition example: `Aaron Taylor-Johnson announced as next James Bond?`
- Kalshi event: `Who will be the next James Bond?`
- Kalshi proposition example: `Will Aaron Taylor-Johnson be the next James Bond?`

Why this matters:

- These look strongly related.
- They are not automatically equivalent.
- The live contracts use different end dates and may encode different announcement windows.

This example should become a labeled near-match test case in fixture mode.

### 24.3 Additional Research-Driven Notes

- Manifold should be treated as a secondary research venue rather than the default routing fallback.
- Some Polymarket markets exposed raw end-date fields that did not obviously align with the natural-language deadline in the market description.
- Kalshi documentation currently contains an orderbook-auth inconsistency: one guide says the REST orderbook endpoint is public, while the endpoint OpenAPI declares auth; live unauthenticated requests succeeded during research.

### 24.4 Candidate Venue APIs Researched

- Polymarket event and market metadata:
  - `https://gamma-api.polymarket.com/events`
  - `https://gamma-api.polymarket.com/markets`
- Polymarket order book:
  - `https://clob.polymarket.com/book`
- Kalshi events, markets, and order book:
  - `https://api.elections.kalshi.com/trade-api/v2/events`
  - `https://api.elections.kalshi.com/trade-api/v2/markets`
  - `https://api.elections.kalshi.com/trade-api/v2/markets/{ticker}/orderbook`
- Manifold search API researched as a fallback candidate:
  - `https://api.manifold.markets/v0/search-markets`

## 25. Summary of v001 Decisions

This PRD recommends an explainable, heuristics-first MVP that treats event-plus-outcome as the canonical matching unit, supports both live and fixture modes, and prioritizes safe equivalence classification over aggressive routing.

The largest unresolved product question is still venue-pair suitability for a convincing live routing demo, but the research now sharpens the answer: keep Polymarket plus Kalshi as the primary routing pair, use fixtures aggressively, and treat governance or economic-model mismatches as explicit safety dimensions rather than accidental details.
