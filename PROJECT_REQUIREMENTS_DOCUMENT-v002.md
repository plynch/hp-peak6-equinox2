# Project Requirements Document

## Document Metadata

- Project: Equinox
- Version: v002
- Status: Draft
- Last Updated: 2026-03-14
- Source of Truth for This Revision: `PROJECT_SPECIFICATION.md`
- Research companion: `PROJECT_PRESEARCH_DOCUMENT.md`
- Revision intent: refine PRD v001 using external review feedback while keeping v001 unchanged

## 1. Executive Summary

Project Equinox is an infrastructure prototype. Its purpose is to test whether prediction markets from different venues can be ingested, normalized into a shared representation, compared for equivalence, and used to simulate routing decisions for hypothetical orders.

This PRD intentionally prioritizes:

- architectural clarity,
- tradeoff awareness,
- explainable reasoning,
- and disciplined prototype scope.

It does not aim to specify a production trading system.

## 2. Core Questions the Prototype Must Answer

The prototype should answer four questions clearly:

1. Can at least two prediction market venues be integrated behind clean adapter boundaries?
2. Can their contracts be mapped into a useful canonical representation without hiding important differences?
3. Can the system distinguish strong matches, weak matches, and unsafe matches with documented reasoning?
4. Can a venue-agnostic routing layer simulate a decision using normalized inputs and explain why it routed or refused to route?

If the final answer to one of these questions is "only partially" or "not reliably," that is acceptable as long as the reasoning is explicit and evidence-backed.

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

### 5.4 Heuristics-First Matching

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

### 5.5 Fixture-First Demo Strategy

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
   - generates candidate cross-venue pairs,
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

### 13.4 Ambiguity Overconfidence

The system may overstate equivalence when timing or resolution semantics are unclear.

Implication:

- confidence and assumptions must be part of the outputs.

### 13.5 Demo Fragility

The system may be architecturally sound but still hard to review if the setup is brittle or the demo depends on live luck.

Implication:

- keep the fixture demo path short and deterministic.

### 13.6 Scope Creep

A prototype can fail by trying to solve too much.

Implication:

- prioritize the narrowest implementation that convincingly answers the spec’s core questions.

## 14. Acceptance Criteria

The PRD is satisfied if the final implementation demonstrates all of the following:

1. Two public venue adapters work in local execution.
2. The system produces a canonical representation that enables cross-venue comparison.
3. The matcher can show at least one strong match or fixture-backed equivalent, one near-match, and one non-match or insufficient-data case.
4. The router can simulate at least one decision using normalized inputs and explain why it routed or refused to route.
5. The repository includes a deterministic fixture-backed demo path from the project root.
6. The submission documents the chosen architecture, key tradeoffs, and handling of ambiguity.
7. The implementation keeps venue-specific logic out of the router.

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
| AI optional but disclosed if used | Sections 5.4, 12 |
| Local deployment acceptable | Section 11 |
| Technical flexibility in tooling and infrastructure | Sections 4, 11, 12 |
| Tradeoff awareness around ambiguity and real-world mismatch | Sections 3, 5, 8, 10, 13 |

## 16. Summary of v002 Changes

Compared with v001, this revision deliberately:

- reduces implementation detail that reads like production specification,
- elevates architectural reasoning and tradeoff documentation,
- softens rigid schema requirements into a minimal canonical model,
- clarifies ambiguity and assumption handling,
- clarifies what venue-agnostic routing means in practice,
- simplifies the AI guidance,
- and makes the local demo and reviewer experience a first-class requirement.
