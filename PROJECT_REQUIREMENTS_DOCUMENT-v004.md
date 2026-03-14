# Project Requirements Document

## Document Metadata

- Project: Equinox
- Version: v004
- Status: Draft
- Last Updated: 2026-03-14
- Source of Truth for This Revision: `PROJECT_SPECIFICATION.md`
- Research companion: `PROJECT_PRESEARCH_DOCUMENT.md`
- Revision intent: reframe the MVP around canonical identity and clustering to better align the PRD with the specification and the next implementation-planning phase

## 1. Purpose

Project Equinox is an infrastructure prototype for cross-venue prediction market normalization and routing simulation.

The prototype should answer this question:

- can markets from different venues be attached to a shared canonical identity layer strongly enough to support explainable comparison and venue-agnostic routing decisions?

This is not a trading product. It is a prototype for evaluating whether cross-venue normalization is viable under real-world ambiguity.

## 2. Architectural Framing

The specification emphasizes architectural reasoning over feature breadth. This PRD therefore frames the problem around the abstractions the prototype needs to expose clearly.

This PRD therefore treats the central architectural problem as:

- discovering a shared identity for the same real-world event across venues,
- separating event identity from proposition identity,
- distinguishing semantic equivalence from route-safe equivalence,
- and carrying ambiguity forward explicitly instead of hiding it.

The submission should therefore optimize for:

- strong problem framing,
- clear system decomposition,
- explicit tradeoffs,
- honest ambiguity handling,
- readable and modular code,
- and a demo that makes the reasoning inspectable.

## 3. Core Thesis

The architectural center of Equinox should be a canonical clustering model, not a permanent set of venue-pair comparisons.

For the MVP, that means:

- create canonical event clusters,
- create canonical proposition clusters within those events,
- attach venue market instances to those proposition clusters,
- and route only between venue instances that are attached to a sufficiently trustworthy proposition cluster and pass route-safety checks.

Pairwise similarity scoring may still be used internally to form or evaluate clusters, but pairwise matching should not be the top-level product model.

## 4. Submission-Specific MVP Commitments

To keep the prototype finishable while still reflecting the real architectural problem, this submission makes the following commitments:

- Implemented venues:
  - `Polymarket`
  - `Kalshi`
- Primary identity model:
  - canonical event clusters with canonical proposition clusters beneath them
- Routeable market family:
  - simple binary yes or no propositions only
- Clustering scope:
  - lightweight clustering for the two implemented venues
  - not a production-grade many-venue clustering service
- Demo posture:
  - fixture-first by default
  - live mode proves ingestion and clustering inputs, not guaranteed live routeability
- Deployment posture:
  - local-first
  - containerized reviewer path preferred
  - CLI-first
- Persistence posture:
  - relational identity store preferred
  - local PostgreSQL is the preferred submission choice
- Preferred implementation posture for this submission:
  - `Go 1.22+`

These choices are submission preferences, not product requirements. They are intended to balance seriousness of architecture with reviewer convenience:

- Go aligns with the spec's note about common internal tooling.
- A containerized local run makes the demo reproducible.
- PostgreSQL gives the clustering model a serious relational persistence layer without requiring cloud infrastructure.
- Local deployment was chosen intentionally because the spec explicitly allows it, which lets the MVP focus on architectural clarity instead of environment complexity.

## 5. Scope

### In Scope

- integration with at least two public prediction market APIs
- ingestion of market metadata and current quote or order book data
- canonical event and proposition clustering
- heuristic or hybrid equivalence detection with documented methodology
- routing simulation for hypothetical orders
- a durable local identity store or registry boundary for clusters and assessments
- containerized local deployment for reproducible demo execution
- inspectable explanations or artifacts for normalization, clustering, and routing
- local setup and deterministic demo support

### Out of Scope

- real-money execution
- wallets, brokerage accounts, or settlement integration
- regulatory or compliance implementation
- production UI, production infrastructure, or ops hardening
- exhaustive support for every market family on every venue
- a distributed many-venue registry or graph service

## 6. Research-Informed Framing

The presearch document materially changes how this PRD frames the problem. The most important findings carried forward are:

- `Polymarket + Kalshi` remains the best primary pair for the MVP
- the real challenge is canonical identity and semantic safety, not raw schema translation
- same-looking contracts can still differ materially in deadline, settlement semantics, governance, or economic domain
- venue-provided labels such as `binary` are not reliable proof of routeable simplicity
- quote models differ enough that route safety must consider observed versus inferred values
- exact live overlap may be sparse enough that fixture-backed examples are necessary
- a third venue would quickly expose pairwise-only architecture as inadequate

As a result, the MVP should visibly model clustering now, with a lightweight but durable local identity layer rather than ephemeral process-only state.

## 7. Venue and Market Scope

### 7.1 Implemented Venues

The MVP should support exactly these venues:

- `Polymarket`
- `Kalshi`

Other venues may remain in research and future design discussion, but they are not required for implementation.

### 7.2 Supported Routeable Market Family

The MVP should treat a venue market instance as routeable only if all of the following are true:

- it expresses a binary yes or no settlement condition
- it represents a single proposition rather than a bundle of dependent conditions
- it has enough metadata and quote data to support responsible comparison
- it can be placed into an event and proposition cluster with adequate confidence
- its deadline and resolution semantics are sufficiently clear

Supported MVP shapes:

- Polymarket standard binary markets
- Polymarket binary markets inside grouped events when the proposition is independently interpretable
- Kalshi standard binary markets

### 7.3 Ingest-Only or Unsupported Families

The MVP may ingest but should not force into routeable proposition clusters:

- Kalshi scalar markets
- Kalshi multivariate or combo-style markets
- bucketed or range-like contracts that are not simple binary propositions
- placeholder, `Other`, or similarly ambiguous outcomes
- markets with materially unclear deadlines
- markets with materially unclear resolution semantics

The preferred behavior is explicit unsupported classification, not coercion.

### 7.4 Priority Routing-Candidate Characteristics

Based on the presearch, the best routing candidates for this MVP are markets with all of the following characteristics:

- simple binary yes or no structure
- one clearly stated proposition rather than a bundle of dependent conditions
- a directly comparable subject and trigger condition across venues
- a clearly extractable semantic deadline
- sufficient quote visibility for both venues
- no placeholder or `Other` outcomes
- no range-bucket or combo semantics hiding behind a `binary` label

In practice, the most promising routing candidates are narrow person, company, or event-outcome propositions with explicit deadlines, not broad topic-adjacent political or sports contracts.

### 7.5 Required Target Evaluation Set

The implementation and demo should intentionally include a small labeled set of target cases drawn from the presearch.

Required categories:

1. strong route-safe proposition cluster
   - at least one `Polymarket + Kalshi` simple binary pair that the system treats as route-safe
   - because the presearch did not confirm a reliable named live pair, this may be fixture-backed rather than dependent on current live overlap

2. near-match or event-only case
   - include:
     - Polymarket: `Aaron Taylor-Johnson announced as next James Bond?`
     - Kalshi: `Will Aaron Taylor-Johnson be the next James Bond?`
   - this should demonstrate that high lexical and topical overlap is not enough when trigger semantics or timing differ

3. clear non-match case
   - include:
     - Polymarket: Democratic control of the House after the 2026 midterms
     - Kalshi: Hakeem Jeffries as next Speaker
   - this should demonstrate that topic adjacency is not proposition identity

4. unsupported-shape case
   - include:
     - a Kalshi combo-style sports or multi-condition contract labeled as `binary`
   - this should demonstrate that venue-reported market type is not sufficient for routeability

5. ambiguity or field-conflict case
   - include at least one Polymarket deadline-conflict example where structured time fields and plain-language description disagree materially
   - preferred named example from presearch:
     - `OpenAI IPO before 2027?`

The goal of this set is not just testing. It is to make the architectural reasoning legible in the live demo.

### 7.6 Priority Event Families for MVP Research and Demo

The MVP should prioritize event families in this order.

1. `Fed / FOMC decisions and closely related macro decisions`
   - this is the preferred first domain for cross-venue routing
   - why:
     - both `Polymarket` and `Kalshi` clearly list Fed-related markets
     - resolution sources are explicit and authoritative
     - deadlines are scheduled
     - semantics are narrower and easier to normalize than many sports props
   - preferred target events:
     - the next scheduled FOMC decision during implementation
     - one or more adjacent scheduled 2026 FOMC decisions if needed for fixtures

2. `Professional soccer / association football`
   - this is the preferred second domain for sports clustering and conditional routing research
   - why:
     - both `Polymarket` and `Kalshi` clearly support soccer markets
     - soccer is globally important and especially relevant in Europe
     - scheduled fixtures create strong event-identity anchors
     - tournament, advancement, and team-specific yes or no markets can fit the clustering model well
   - caveat:
     - soccer match semantics can be tricky
     - the system must distinguish carefully between:
       - win in regulation
       - draw
       - to advance
       - to win tournament
       - exact score
     - generic match-result markets are not automatically route-safe just because the teams align
   - preferred target event families:
     - next week's league fixtures in the Big Five European leagues:
       - Premier League
       - La Liga
       - Bundesliga
       - Serie A
       - Ligue 1
     - Champions League winner
     - Champions League to-advance markets
     - league-winner markets
     - team-specific yes or no tournament outcome markets when proposition wording aligns cleanly

3. `Professional golf / PGA Tour`
   - this is the preferred third domain for clustering and market-structure stress-testing
   - why:
     - both `Polymarket` and `Kalshi` clearly support PGA or golf-related markets
     - golf provides useful variation in event, proposition, and quote structure
   - caveat:
     - golf should not be the first sports routing domain unless exact proposition overlap is confirmed
     - many golf markets are event winners, head-to-heads, top-X finishes, make-cut markets, or specialty props, which increases normalization pressure
   - preferred target event families:
     - Masters Tournament
     - PGA Championship
     - PGA Tour tournament winner or make-cut style markets only when exact proposition overlap is confirmed

Practical recommendation:

- use Fed decisions as the first serious routing domain
- use next week's Big Five league fixtures as the first serious sports cadence
- use professional soccer more broadly as the second domain to prove the clustering model on globally common sports events with rich but tricky semantics
- use PGA or golf as a third domain if additional structure stress-testing is useful

## 8. Canonical Identity Model

The MVP should make three levels of identity explicit.

### 8.1 Canonical Event Cluster

An event cluster represents the shared real-world event or question family that multiple venue markets may refer to.

Examples:

- "Who will be the next James Bond actor?"
- "Will OpenAI IPO before 2027?"

The event cluster exists to group semantically related propositions. It is not yet enough for routing.

### 8.2 Canonical Proposition Cluster

A proposition cluster represents a specific assertable outcome inside an event cluster.

Examples:

- "Aaron Taylor-Johnson will be announced as the next James Bond"
- "OpenAI will IPO before the relevant deadline"

Routing should happen at this level, not at the event level.

### 8.3 Venue Market Instance

A venue market instance is the venue-specific tradable contract attached to a proposition cluster.

It retains venue-native details such as:

- source identifiers
- native wording
- quote model
- timing fields
- governance or settlement characteristics

## 9. Required System Decomposition

The prototype should keep these components distinct:

1. Venue adapters
   - fetch and parse venue-specific data
   - preserve source fields and raw provenance

2. Identity and normalization layer
   - normalize venue markets into candidate event and proposition representations
   - preserve ambiguity, unsupported labels, and assumption notes

3. Clustering layer
   - group venue market instances into canonical event clusters
   - group compatible instances into canonical proposition clusters
   - record confidence and reasons

4. Identity store or registry boundary
   - persist cluster records, assessments, and stable references
   - support local durability without requiring distributed infrastructure
   - model the canonical identity layer as a real relational system, not a process-local cache

5. Routing layer
   - evaluate routeable venue instances attached to a proposition cluster
   - remain venue-agnostic by consuming normalized inputs only

6. Explanation and artifact layer
   - emit inspectable outputs for clustering decisions, route decisions, and refusals

The router must not inspect raw venue payloads directly and must not branch on venue names.

## 10. Minimal Canonical Model

This section defines the minimum distinctions that should exist in the implementation. Exact field names may vary.

The preferred storage posture for this submission is a relational database, with PostgreSQL as the default choice. An alternative store is acceptable only if it still preserves durable identities, cluster relationships, and inspectable reassessment history without increasing reviewer friction.

### 10.1 Event Cluster

Each event cluster should capture at least:

- stable cluster identifier
- canonical event label
- member venue market instances or proposition candidates
- cluster confidence
- explanation or evidence notes

### 10.2 Proposition Cluster

Each proposition cluster should capture at least:

- stable cluster identifier
- canonical proposition statement
- event-cluster reference
- asserted outcome side or outcome meaning
- composition type, such as simple binary vs combo-like or bucketed
- semantic deadline or close condition
- deadline provenance and confidence
- economic domain or venue class
- resolution authority or governance model
- routeability status
- ambiguity or assumption notes

### 10.3 Venue Market Instance

Each venue market instance should capture at least:

- source venue
- source market identifier
- native title or wording
- market family
- quote model
- normalized yes or no pricing view for the hypothetical order
- observed versus inferred quote indicators where relevant
- liquidity or depth signal if available
- quote freshness
- instance-level ambiguity notes

### 10.4 Assessment Records

The system should produce inspectable records for:

- cluster membership decisions
- proposition equivalence assessments
- route decisions or refusals

## 11. Equivalence Model

The implementation must define what "equivalent" means. For this prototype, equivalence should be expressed at more than one level.

### 11.1 Event-Level Equivalence

Two venue markets may be event-equivalent when they refer to the same broader real-world event or question family, even if they do not represent the same routeable proposition.

### 11.2 Proposition-Level Equivalence

Two venue markets are proposition-equivalent only when they represent the same assertable outcome strongly enough to be placed in the same proposition cluster.

### 11.3 Route-Safe Equivalence

Two proposition-equivalent venue markets are route-safe equivalents only when their differences in timing, composition, quote quality, governance, and economic domain do not make routing misleading.

This distinction is important. The presearch found that semantic similarity alone is not enough.

### 11.4 Required Classifications

The system should distinguish at least:

- `event_match_only`
- `strong_proposition_match`
- `near_match`
- `unsupported_or_unsafe`
- `insufficient_data`

### 11.5 Evidence Requirement

For every important equivalence or non-equivalence decision, the system should record:

- supporting evidence
- blocking evidence
- assumptions
- confidence
- and any unresolved ambiguity

## 12. Clustering Methodology Requirements

The implementation does not need a production clustering engine, but it should make the methodology explicit.

Acceptable MVP posture:

- use pairwise similarity scoring as one input to clustering
- build lightweight event and proposition clusters backed by durable local relational storage
- allow clusters to remain uncertain or unresolved
- refuse clustering when semantics are too weak

Unacceptable posture:

- flatten the architecture into permanent venue-pair logic with no canonical identity layer
- treat event titles as sufficient proof of proposition identity
- silently force every market into a cluster

## 13. Routing Requirements

The routing engine is a simulation layer, not an execution layer.

### 13.1 Router Inputs

The router may consume normalized inputs such as:

- proposition-cluster membership and confidence
- normalized price view for the hypothetical order
- quote model
- observed versus inferred quote indicators
- quote freshness
- liquidity or depth signal
- economic-domain comparability flags
- governance comparability flags
- route-safety flags

### 13.2 Router Boundaries

The router should not consume:

- raw venue payloads
- venue-name branches
- ad hoc assumptions that only hold for one venue schema

### 13.3 Router Outputs

For a supported hypothetical order, the router should produce:

- `route_to_<venue>`
- or `do_not_route`

A justified refusal is a successful outcome.

### 13.4 Router Explanation

Every routing outcome should explain:

- which proposition cluster it operated on
- which normalized inputs mattered
- what route-safety checks mattered
- and why the selected venue or refusal was appropriate

## 14. Ambiguity and Unsupported Handling

This is a core requirement.

The system should explicitly surface:

- missing fields
- conflicting fields
- inferred deadlines
- unsupported market families
- uncertain cluster assignments
- unclear governance or settlement semantics
- quote-confidence limitations

The preferred behavior is:

- explicit ambiguity over hidden ambiguity
- safe refusal over unsafe normalization
- inspectable evidence over opaque scoring

At least one example in the final submission should show a market that is related enough to cluster at the event level but not safe enough to cluster at the proposition or routing level.

## 15. Setup, Deployment, and Demo

### 15.1 Setup and Deployment Posture

- the prototype should run locally from the repository root
- the default reviewer path should require no cloud deployment
- the default reviewer path should ideally be a one-command containerized startup
- the default reviewer path should require no secrets
- live ingestion may use optional environment configuration if needed, but the primary demo should not depend on it
- the preferred local packaging model is:
  - application container
  - local PostgreSQL container
  - mounted fixtures and output artifacts

Containerization in this project is for reproducibility and demo simplicity, not for infrastructure sophistication.

Local deployment is the chosen MVP posture because the specification explicitly allows it and because it keeps the project focused on identity, clustering, routing, and explainability instead of deployment overhead.

### 15.2 Beyond-Local Deployment Path

Although local deployment is the chosen MVP scope boundary, the architecture should still preserve a credible path to hosted deployment later.

The intended evolution path is:

1. keep the application packaged as a containerized service
2. keep PostgreSQL as the system of record for clusters, venue market instances, assessments, and routing decisions
3. move from local container orchestration to a hosted container runtime
4. move from local PostgreSQL to a managed PostgreSQL deployment
5. externalize configuration, scheduled ingestion, and artifact storage without changing the core clustering and routing boundaries

The important architectural point is:

- local deployment was chosen to limit MVP scope, not because the system is meant to remain local-only
- the application and data boundaries should be designed so they can be lifted into a hosted environment with limited structural change
- this future path should remain secondary to the current requirement of a simple, reviewer-friendly local demo

### 15.3 Demo Modes

The submission should support two demo paths:

1. fixture demo
   - deterministic
   - secret-free
   - primary reviewer path

2. live inspect or live ingest demo
   - proves current public ingestion works
   - does not need to guarantee a routeable live cluster every run

### 15.4 Demo Script Requirements

The demo should be easy for the submitter to run live in front of a reviewer.

The preferred demo flow is:

1. start the local stack with one command
2. run the fixture demo that:
   - ingests fixture-backed venue data
   - builds event and proposition clusters
   - shows at least one strong cluster, one event-only or near-match case, one unsupported case, and one ambiguity case
   - produces a routing decision or justified refusal
3. optionally run a live inspect command to show that current public ingestion works

The demo should avoid requiring the presenter to manually stitch together raw JSON files or ad hoc shell steps.

### 15.5 Reviewer Experience

The demo should make it easy to inspect:

- raw venue inputs at a high level
- event clusters
- proposition clusters
- routeability decisions
- and routing outcomes or refusals

CLI-first operation inside a containerized local stack is preferred.

## 16. Required Deliverables

The final submission should include:

- a working prototype
- setup and run instructions
- Docker packaging for the primary reviewer path
- a brief architecture overview
- a short design-decisions or tradeoffs section
- a written explanation of the clustering and equivalence methodology
- a written explanation of the routing methodology
- a short statement of supported versus unsupported market families
- a short statement of ambiguity-handling behavior
- a demo walkthrough or command sequence
- a short presenter-oriented demo script
- and AI disclosure only if AI was used

The documentation should let a reviewer understand the design without inferring it from source code alone.

## 17. Success Criteria

This PRD is satisfied when the implementation demonstrates all of the following:

1. `Polymarket` and `Kalshi` are integrated behind clean adapter boundaries.
2. The system ingests both market metadata and pricing or order book data.
3. The system produces canonical event clusters and canonical proposition clusters.
4. Supported binary propositions can be attached to proposition clusters with inspectable reasoning.
5. Unsupported or unclear market types are surfaced explicitly rather than forced into routeable clusters.
6. The system can show at least:
   - one event cluster with multiple venue instances
   - one strong proposition cluster
   - one near-match or event-only match
   - one unsupported example
   - and one ambiguity example
7. The router consumes normalized cluster-level inputs and remains venue-agnostic.
8. The system emits inspectable explanations or artifacts for clustering and routing.
9. Cluster and assessment records survive beyond a single process run through a durable local storage or artifact mechanism.
10. The repository includes a deterministic local demo path that is easy to present live.
11. The primary reviewer path is reproducible through containerized local execution and concise reviewer-facing documentation.

## 18. Deferred Work

The following are intentionally deferred:

- a distributed or service-backed many-venue proposition registry
- full graph-based or ML-heavy clustering infrastructure
- third-venue implementation
- scalar or multi-outcome routing
- real-money execution
- managed cloud infrastructure and UI-heavy workflows

The long-range architecture discussed in the presearch remains important, but this MVP only needs to prove the clustering-first framing in a narrow, defensible form.
