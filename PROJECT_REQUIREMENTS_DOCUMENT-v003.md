# Project Requirements Document

## Document Metadata

- Project: Equinox
- Version: v003
- Status: Draft
- Last Updated: 2026-03-14
- Source of Truth for This Revision: `PROJECT_SPECIFICATION.md`
- Research companion: `PROJECT_PRESEARCH_DOCUMENT.md`
- Revision intent: make MVP implementation scope and architectural decisions more explicit while keeping v001 and v002 unchanged

## 1. Executive Summary

Project Equinox is an infrastructure prototype. Its purpose is to test whether prediction markets from different venues can be ingested, normalized into a shared representation, compared for equivalence, and used to simulate routing decisions for hypothetical orders.

This PRD intentionally prioritizes:

- architectural clarity,
- tradeoff awareness,
- explainable reasoning,
- and disciplined prototype scope.

It does not aim to specify a production trading system.

## 2. Core Questions and Explicit MVP Commitments

The prototype should answer four questions clearly:

1. Can at least two prediction market venues be integrated behind clean adapter boundaries?
2. Can their contracts be mapped into a useful canonical representation without hiding important differences?
3. Can the system distinguish strong matches, weak matches, and unsafe matches with documented reasoning?
4. Can a venue-agnostic routing layer simulate a decision using normalized inputs and explain why it routed or refused to route?

If the final answer to one of these questions is "only partially" or "not reliably," that is acceptable as long as the reasoning is explicit and evidence-backed.

### 2.1 Explicit MVP Commitments

To avoid ambiguity during implementation planning, the MVP should be understood as making the following concrete commitments:

- Implemented venues:
  - `Polymarket`
  - `Kalshi`
- Routeable market family:
  - simple binary yes or no propositions only
- Matching scope:
  - pairwise proposition matching across the two implemented venues
- Event handling:
  - event metadata may be used for candidate generation and explanation
  - full event clustering is not a required MVP subsystem
- Default demo mode:
  - fixture-first
- Live mode purpose:
  - prove current public ingestion works
  - not guarantee a routeable live match on every run

## 3. Evaluation Posture

The project specification emphasizes architectural thinking rather than feature breadth. This PRD therefore treats the following as primary evaluation goals:

- clear system decomposition,
- explicit handling of ambiguity,
- well-defended design choices,
- readable and modular implementation,
- and a demo path that makes the architecture easy to inspect.

Functional completeness is secondary to those goals.

## 4. Scope

### In Scope

- Integration with at least two public prediction market APIs.
- Ingestion of market metadata and current pricing data.
- A canonical internal model sufficient for cross-venue comparison.
- Heuristic or hybrid equivalence detection.
- Routing simulation for hypothetical orders.
- Local setup and deterministic demo support.
- Written explanations of key architectural and matching decisions.

### Out of Scope

- Real-money order execution.
- Wallet, account, or brokerage integration.
- Regulatory or compliance implementation.
- Production infrastructure, monitoring, or ops hardening.
- Production UI polish.
- Exhaustive support for every market type on every venue.

### Explicit Venue Scope

The MVP implementation should support exactly these venues:

- `Polymarket`
- `Kalshi`

Other venues may inform research and future design, but they are not required for the implementation plan or MVP build.

### Explicit Routeable Market Scope

The MVP should treat a market as routeable only if all of the following are true:

- it is a binary proposition with a clear yes or no settlement condition,
- it can be interpreted as a single real-world assertion rather than a bundle of conditions,
- it has enough metadata and quote data to compare responsibly,
- and its timing and resolution semantics are understandable from available fields or rules text.

Supported routeable shapes for the MVP:

- Polymarket standard binary markets
- Polymarket binary markets inside grouped events when the individual proposition is independently interpretable
- Kalshi standard binary markets

### Explicit Ingest-Only or Unsupported Market Scope

The MVP may ingest but should not attempt to route or force-match the following:

- Kalshi scalar markets
- Kalshi multivariate or combo-style markets
- binary-looking markets that actually encode many joint conditions
- placeholder, `Other`, or similarly ambiguous outcomes
- markets with materially unclear resolution timing
- markets with materially unclear resolution semantics

If the system encounters these, the preferred behavior is to label them unsupported or insufficiently clear rather than coerce them into the canonical routeable set.

## 5. Key Decisions and Tradeoffs

This section makes the major design choices explicit and documents alternatives.

### 5.1 Initial Venue Pair

Chosen approach:

- `Polymarket + Kalshi`

Why:

- both expose public market data,
- both present real normalization challenges,
- and both are strong representatives of different venue models.

Alternatives considered:

- `Manifold` as a second venue:
  - useful for semantic overlap research,
  - weaker as a routing peer because of play-money economics and creator resolution.
- adding a third venue immediately:
  - valuable for long-range thinking,
  - unnecessary for meeting the spec on time.

Tradeoff:

- this keeps the implementation focused,
- but exact live overlap may be sparse.

### 5.2 Matching Unit

Chosen approach:

- the canonical matching unit is the proposition, meaning the real-world event plus the specific outcome being asserted.

Alternatives considered:

- event-level matching:
  - too coarse for routing.
- venue market ID matching:
  - not cross-venue meaningful.

Tradeoff:

- proposition-level matching is more work than simple title comparison,
- but it is the most defensible unit for cross-venue reasoning.

### 5.3 Binary-First Support

Chosen approach:

- support binary contracts first,
- and explicitly mark unsupported or partially supported market structures.

Alternatives considered:

- scalar and broad multi-outcome support in the MVP.

Tradeoff:

- narrower product coverage,
- higher confidence in what the prototype does support.

### 5.4 Event Clustering Decision

Chosen approach:

- do not build a standalone event-clustering subsystem for the MVP
- use pairwise proposition matching between supported venue contracts
- allow lightweight canonical event labels only as explanation aids if useful

Alternatives considered:

- full event clustering or proposition-registry-first matching

Tradeoff:

- pairwise matching is simpler and easier to finish well for two venues,
- but it is not the long-range architecture for a many-venue system

### 5.5 Heuristics-First Matching

Chosen approach:

- deterministic heuristics are the default matching method.

Alternatives considered:

- AI-assisted matching in the core path.

Tradeoff:

- heuristics may miss some borderline matches,
- but they are easier to explain, test, and defend in a prototype setting.

AI note:

- AI remains optional.
- If used, it should be clearly documented and treated as an assistive component, not a hidden dependency.

### 5.6 Fixture-First Demo Strategy

Chosen approach:

- support both live mode and fixture mode,
- with fixture mode as the default reviewer path.

Alternatives considered:

- live-only demo.

Tradeoff:

- fixtures add setup work,
- but they remove demo fragility and make the prototype reproducible.

## 6. Prototype Architecture

The prototype should be decomposed into the following logical parts:

1. `Venue adapters`
   - fetch venue data,
   - handle paging and endpoint quirks,
   - and convert raw responses into an internal source-contract form.

2. `Normalization layer`
   - extracts a canonical proposition representation,
   - preserves provenance,
   - and records ambiguity or unsupported cases.

3. `Matcher`
   - generates candidate cross-venue proposition pairs,
   - evaluates equivalence strength,
   - and produces explanations and confidence notes.

4. `Router`
   - consumes normalized matches and normalized quote data,
   - simulates venue selection for a hypothetical order,
   - and may return either a chosen venue or `do_not_route`.

5. `Artifacts and docs`
   - persist sample outputs,
   - make reasoning inspectable,
   - and support deterministic review.

### Architectural Boundary Rule

Venue-specific behavior belongs in the adapters and normalization layer.

The router may consume normalized attributes such as:

- quote model,
- fee visibility,
- resolution model category,
- confidence level,
- and supported-for-routing flags.

The router should not branch directly on venue names.

Unacceptable example:

- `if venue == "kalshi" then ...`

Acceptable example:

- `if quote_model == "bid_only_reciprocal" then ...`

### Event Handling Rule

For the MVP, event-level data is supporting context, not the primary matching target.

That means:

- event titles and metadata may help narrow candidate pairs,
- matched propositions may optionally be grouped under a lightweight canonical event label for explanation,
- but the MVP does not need a durable cross-venue event clustering system.

## 7. Minimal Canonical Model

The prototype needs a canonical model, but it does not need a production-grade universal schema. The canonical model should be only as detailed as required to support matching, routing, and explanation.

Minimum logical records:

### 7.1 Source Contract

Should preserve:

- venue,
- source identifier,
- source title,
- raw timing fields,
- raw resolution text or references if available,
- source quote data,
- and raw payload provenance.

### 7.2 Canonical Proposition

Should preserve:

- normalized subject or event,
- normalized asserted outcome,
- relevant timing or deadline notes,
- polarity or direction,
- routing support status,
- and ambiguity notes or assumption notes.

Implementation note:

- a canonical proposition may be created for explanation and comparison purposes without implying that the system supports general multi-venue clustering beyond the current pairwise MVP.

### 7.3 Quote Snapshot

Should preserve:

- venue,
- timestamp,
- normalized price representation,
- liquidity or depth summary if available,
- quote caveats such as inferred ask or limited depth,
- and confidence notes.

### 7.4 Match Assessment

Should preserve:

- compared contracts,
- classification,
- confidence,
- supporting signals,
- blocking signals,
- and explanation.

### 7.5 Routing Decision

Should preserve:

- input order request,
- venue options considered,
- selected venue or `do_not_route`,
- decision rationale,
- and important caveats.

### Optional Helper Model

A lightweight venue traits or capability configuration is allowed if useful. It should stay small and serve the architecture, not become a second product.

## 8. Equivalence Logic

### 8.1 Working Definition

Two contracts should be treated as equivalent only when they appear to resolve on the same underlying truth condition with materially aligned timing and resolution semantics.

This definition should be treated as a working prototype rule, not a claim of perfect market ontology.

### 8.2 Expected Output Classes

The matcher should support at least these outcomes:

- strong equivalent candidate suitable for routing review,
- semantically similar but route-unsafe,
- near-match for analysis only,
- not equivalent,
- insufficient data.

The exact labels may vary in implementation as long as the distinctions are preserved.

For the MVP, these classes are expected to be produced from pairwise proposition comparisons between supported venue contracts rather than from a global clustering engine.

### 8.3 Ambiguity Policy

The system must not force a positive equivalence result when evidence is weak.

When evidence is incomplete or conflicting, the matcher should:

- lower confidence,
- record assumptions,
- and prefer `near-match` or `insufficient_data` over overconfident matching.

### 8.4 Matching Method Guidance

The prototype may use:

- rules,
- heuristics,
- embeddings,
- LLM assistance,
- or a hybrid approach.

Requirement:

- whatever method is chosen must be documented,
- and the reasoning behind that choice must be explained.

## 9. Routing Logic

The router exists to test whether venue-agnostic decision logic is plausible. It does not need to optimize execution perfectly.

### 9.1 Minimum Supported Routing Input

The MVP should support a small order surface, preferably:

- `BUY_YES`
- `BUY_NO`

`SELL` support is optional.

### 9.2 Routing Inputs

The router may consider:

- normalized price,
- visible liquidity if available,
- quote freshness,
- match confidence,
- route-safety status,
- and fees if they are visible enough to use responsibly.

The router should only consider venue options that came through the supported pairwise matching pipeline. It should not independently search raw venue inventories or perform new venue-specific interpretation.

### 9.3 Routing Output

The router should produce:

- `route_to_<venue>`
- or `do_not_route`

A justified `do_not_route` outcome is a successful prototype result.

### 9.4 Routing Abstraction Requirement

The router should operate on normalized inputs, not raw venue payloads.

The PRD intentionally leaves implementation freedom here, but the submission should clearly document:

- what normalized inputs the router receives,
- what venue-specific details are abstracted away,
- and which normalized differences are still allowed to influence the decision.

## 10. Ambiguity, Assumptions, and Confidence

This is a core prototype concern, not a side note.

The system should explicitly record:

- missing fields,
- conflicting fields,
- unsupported market types,
- inferred values,
- confidence notes,
- and assumptions made during normalization or matching.

At least one example in the final submission should show how the system behaves when data is ambiguous or incomplete.

## 11. Local Setup, Deployment, and Demo

The prototype should be local-first.

### 11.1 Setup and Deployment Posture

- A reviewer should be able to run the default demo from the repository root.
- The default reviewer path should require no cloud deployment.
- Fixture mode should require no secrets.
- Live mode may use optional environment configuration only if needed.

Cloud deployment is allowed but unnecessary unless it clearly simplifies the reviewer experience.

### 11.2 Demo Modes

The submission should support two demo paths:

1. `fixture demo`
   - deterministic,
   - secret-free,
   - and the default first-run path.

2. `live demo`
   - proves current public ingestion works,
   - but does not need to guarantee a routeable match on every run.

### 11.3 Reviewer Experience

The reviewer path should be short and explicit.

Preferred posture:

- CLI-first commands,
- documented expected outputs,
- and saved example artifacts for quick inspection.

## 12. Required Documentation and Deliverables

The final submission should include:

- a working prototype,
- setup and run instructions,
- a brief architecture overview,
- a short design-decisions or tradeoffs section,
- a short statement of supported versus unsupported market types,
- a short statement of the MVP matching scope, including whether event clustering is implemented or intentionally deferred,
- a written explanation of equivalence logic,
- a written explanation of routing logic,
- a demo walkthrough or command sequence,
- and AI disclosure only if AI was used.

The architecture overview should explain:

- component boundaries,
- end-to-end data flow,
- and why the chosen implementation approach is appropriate for a local prototype.

## 13. Prototype Risks

### 13.1 Sparse Exact Overlap

Live exact matches may be uncommon.

Implication:

- fixture support is necessary,
- and the prototype should be prepared to show that many live candidates are only near-matches.

### 13.2 Over-Normalization

Too much normalization can hide meaningful venue differences.

Implication:

- provenance and caveats must be preserved,
- and unsupported or unclear cases should stay visible.

### 13.3 Venue-Agnostic Router Drift

The architecture may claim to be venue-agnostic while quietly reintroducing venue-specific logic in the router.

Implication:

- the implementation should document router inputs and boundaries clearly.

### 13.4 Event-Clustering Scope Drift

The prototype may become harder to ship if pairwise matching quietly expands into a broader clustering system without clear need.

Implication:

- keep the MVP focused on pairwise proposition matching unless a clustering layer becomes necessary to support the chosen demo.

### 13.5 Ambiguity Overconfidence

The system may overstate equivalence when timing or resolution semantics are unclear.

Implication:

- confidence and assumptions must be part of the outputs.

### 13.6 Demo Fragility

The system may be architecturally sound but still hard to review if the setup is brittle or the demo depends on live luck.

Implication:

- keep the fixture demo path short and deterministic.

### 13.7 Scope Creep

A prototype can fail by trying to solve too much.

Implication:

- prioritize the narrowest implementation that convincingly answers the spec’s core questions.

## 14. Acceptance Criteria

The PRD is satisfied if the final implementation demonstrates all of the following:

1. Two public venue adapters work in local execution.
2. The system produces a canonical proposition representation that enables cross-venue comparison for supported binary markets.
3. The matcher performs pairwise proposition comparison across the supported venues and can show at least one strong match or fixture-backed equivalent, one near-match, and one non-match or insufficient-data case.
4. The router can simulate at least one decision using normalized inputs derived from matched supported markets and explain why it routed or refused to route.
5. The implementation surfaces at least one unsupported market example instead of incorrectly forcing it into the routeable set.
6. The repository includes a deterministic fixture-backed demo path from the project root.
7. The submission documents the chosen architecture, key tradeoffs, supported market scope, and handling of ambiguity.
8. The implementation keeps venue-specific logic out of the router.

## 15. Specification Traceability

| Spec Requirement | PRD Coverage |
| --- | --- |
| Connect to at least two venues | Sections 4, 6, 14 |
| Define canonical internal market model | Sections 6, 7, 14 |
| Attempt equivalence detection | Sections 8, 10, 14 |
| Simulate routing decisions | Sections 6, 9, 14 |
| Log and explain decisions | Sections 6, 8, 9, 10, 12 |
| Handle imperfect data gracefully | Sections 8, 10, 13, 14 |
| Clear separation of concerns | Sections 6, 9, 12 |
| Routing logic without venue-specific assumptions | Sections 6, 9, 14 |
| Document what "equivalent" means | Section 8 |
| AI optional but disclosed if used | Sections 5.5, 12 |
| Local deployment acceptable | Section 11 |
| Technical flexibility in tooling and infrastructure | Sections 4, 11, 12 |
| Tradeoff awareness around ambiguity and real-world mismatch | Sections 3, 5, 8, 10, 13 |

## 16. Summary of v003 Changes

Compared with v002, this revision deliberately:

- makes the supported venue set explicit,
- makes the supported and unsupported market families explicit,
- states that the MVP matcher is pairwise proposition matching,
- states that full event clustering is intentionally out of scope for the MVP,
- clarifies that event metadata is supporting context rather than the primary matching target,
- and makes the implementation-facing scope boundaries harder to misread.
