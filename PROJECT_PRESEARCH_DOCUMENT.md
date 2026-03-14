# Project Presearch Document

## Document Metadata

- Project: Equinox
- Versioning: unversioned live document
- Status: Draft, expanded refinement pass
- Last Updated: 2026-03-14
- Scope of this document: pre-implementation research to refine PRD v001 and expose architectural risks before implementation planning
- Related documents:
  - `PROJECT_SPECIFICATION.md`
  - `PROJECT_REQUIREMENTS_DOCUMENT.md`

## 1. Purpose

This document exists to answer the research questions that sit underneath the Equinox spec before implementation planning begins.

The central question is not simply "which prediction market APIs can be called?" It is:

- how different venues encode the same real-world proposition,
- how settlement and governance vary,
- how quote models differ,
- which fields are trustworthy,
- how to classify near-matches safely,
- and what architecture will still work when the system grows from 2 venues to 3 and eventually to 10 or more.

This pass is intentionally more detailed than the first draft. The goal is to front-load ambiguity, not hide it.

## 2. Research Questions and Current Answers

### 2.1 Question: Which venues are realistic candidates for the Equinox prototype?

Current answer:

- The strongest primary pair remains `Polymarket + Kalshi`.
- The strongest additional candidates depend on the goal:
  - `Myriad` is the strongest third venue for architecture learning because it already separates `question` and `market`.
  - `PredictIt` is the strongest additional public real-money-like market-data candidate found in this pass, but its public API is sparse, rate-limited, and often exposes bucketed contracts rather than simple binary propositions.
  - `Manifold` remains valuable for semantic overlap research but is route-unsafe against real-money venues by default because it is play-money and creator-resolved.
  - `Metaculus` is useful as a semantic forecasting corpus, not as a routing peer.
  - `Zeitgeist` is a plausible future candidate but currently looks like a higher-integration-cost decentralized protocol candidate rather than a near-term MVP add.

### 2.2 Question: Which research areas need the most attention?

Current answer:

The highest-attention areas are:

1. Mapping the broader candidate landscape rather than reasoning only from the first two venues.
2. Describing how architecture should scale beyond pairwise matching.
3. Treating composite, bucketed, and combo-style contracts as a separate design problem.
4. Distinguishing "good semantic research venue" from "good routing peer."
5. Giving a best current answer for each forward-looking research question.

These are the areas this pass addresses most directly.

### 2.3 Question: What are the major architectural challenges?

Current answer:

The major challenges are:

1. Defining proposition identity independent of venue IDs.
2. Decomposing grouped, bucketed, combo, or dependent markets into routeable units safely.
3. Extracting the real semantic deadline from inconsistent structured fields and rules text.
4. Comparing resolution governance across oracle-, exchange-, creator-, and protocol-driven venues.
5. Normalizing very different quote surfaces such as direct books, bid-only reciprocal books, and AMM-derived probabilities.
6. Comparing liquidity and fillability when venues expose different depth and fee models.
7. Handling unsupported market families explicitly instead of coercing them into false binary equivalence.
8. Avoiding pairwise logic explosion as venue count grows.
9. Preserving enough provenance that every normalization, match, and route decision is auditable.

### 2.4 Question: What happens if we add a third market?

Current answer:

If the current architecture is pairwise and venue-name-driven, adding a third venue will expose design weaknesses immediately.

A third venue forces the system to answer:

- Is matching pairwise only, or can many venue contracts attach to one canonical proposition?
- Does the router compare venues by normalized capability and policy, or by hard-coded if-statements?
- Can unsupported market types be fenced off cleanly?
- Are route decisions restricted by comparability domain, or does the system pretend all venues are interchangeable?

The safest answer is:

- add a third venue only through an adapter plus capability-profile boundary,
- attach contracts to canonical propositions or proposition clusters,
- and route only within explicitly allowed comparability domains.

### 2.5 Question: Are there even viable candidates for a third market?

Current answer:

Yes, but the candidate pool is smaller and messier than it first appears.

If the requirements are:

- public or low-friction read access,
- live market metadata,
- current price signals,
- enough resolution detail to understand semantics,
- and sufficient documentation or inspectability,

then the pool is limited. There are candidates, but they are heterogeneous and often only satisfy part of that list.

### 2.6 Question: What should the long-range architecture look like in a world with 10+ venues?

Current answer:

It should not be built around permanent venue-pair comparison code.

It should instead have:

- a `venue adapter layer`,
- a `venue capability registry`,
- a `canonical proposition registry`,
- a `candidate retrieval and equivalence engine`,
- an `evidence graph` or provenance store,
- a `quote normalization service`,
- and a `policy-driven router` that operates on normalized risk/comparability attributes rather than venue names.

### 2.7 Question: Should all venues be routable against each other?

Current answer:

No.

The system should distinguish:

- venues that are safe to compare for routing,
- venues that are only safe to compare semantically,
- and venues that should be treated only as research or control data.

Default route-comparability domains should likely include at least:

- regulated or venue-operated real-money exchange style,
- tokenized/onchain real-money-like,
- play-money/social,
- and forecasting/non-market.

### 2.8 Question: Can two contracts have the same text and still be unsafe to treat as equivalent?

Current answer:

Yes. This is one of the main lessons of the research.

Examples:

- same text but different deadline,
- same text but creator resolution on one venue and oracle/exchange resolution on another,
- same text but one venue treats the contract as a final-state condition while another resolves on first occurrence,
- same text but one side is a bucket in a ranged market rather than a simple binary proposition.

## 3. Executive Summary

The current best answer to the Equinox feasibility question is:

- cross-venue normalization is feasible,
- cross-venue semantic matching is feasible but ambiguity-heavy,
- safe routing is only feasible when capability and governance differences are first-class inputs,
- and exact live overlap is sparse enough that fixture mode is mandatory.

More specifically:

- `Polymarket + Kalshi` is still the best primary pair for the MVP routing problem.
- `Myriad` is the most architecturally interesting third venue candidate found in this pass because it exposes a `question` abstraction that already resembles the canonical proposition layer Equinox needs.
- `PredictIt` is a legitimate additional market-data candidate, but its low public rate limit and bucket-heavy market shapes increase normalization pressure.
- `Manifold` remains extremely useful for semantic research and negative examples, but it should not be treated as a default routing peer for real-money venues.
- The biggest architectural risk is not adapter code. It is semantic drift combined with pairwise scaling.

## 4. Architectural Pressure Points

### 4.1 Pairwise Thinking Is Insufficient

Why this matters:

- a two-venue view hides what happens when a third venue arrives,
- pairwise matching and proposition clustering are not the same thing,
- and long-range architecture can look fine at two venues while breaking at three.

Answer:

The architecture should evolve toward a proposition-registry model, not permanent venue-pair logic.

### 4.2 Candidate Landscape Must Be Explicit

Why this matters:

- venue selection is itself part of the architectural problem,
- different venues stress different parts of the normalization stack,
- and implementation planning depends on knowing whether candidate three is even worth supporting.

Answer:

This pass adds a candidate matrix and a recommendation by use case.

### 4.3 Composite Market Families Need First-Class Handling

Why this matters:

- `binary` as a venue field is not enough,
- a combo contract can still advertise itself as binary,
- and ranged/bucketed families create false positives if treated as simple propositions.

Answer:

Composition type must be normalized explicitly.

### 4.4 Route-Safety Policy Must Be Explicit

Answer:

The system needs route-comparability domains and explicit policy gates.

### 4.5 Open Research Questions Need Provisional Answers

Answer:

This version answers every question it raises. Where evidence is still incomplete, the document gives the best current answer and labels the uncertainty.

## 5. Venue Landscape and Candidate Assessment

### 5.1 Candidate Matrix

| Venue | Economic Domain | Typical Market Shape | Resolution Authority | Quote Surface | API / Access Posture | Equinox Value | Current Verdict |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Polymarket | Real-money tokenized / onchain-settled with offchain matching | Binary contracts grouped into events, including neg-risk structures | UMA Optimistic Oracle plus dispute path | Public CLOB bids and asks | Public docs and public market data | Core primary venue | `primary` |
| Kalshi | Regulated real-money exchange | Binary, scalar, and multivariate/combo markets under series and events | Venue/exchange rules and settlement | Public REST order book with reciprocal ask inference | Strong public docs, public REST reads, auth for WebSocket | Core primary venue | `primary` |
| Manifold | Play-money social market | Binary, multiple choice, numeric, dependent markets | Creator resolution with moderator override | Probability + AMM + limit-order surfaces | Public alpha API | Semantic stress-test and route-unsafe controls | `research_only` |
| Myriad | Tokenized / protocol-style public market | Questions containing one or more markets, binary and multi-outcome examples observed | Resolution source and resolution title exposed per market | Outcome prices in market payloads | Public REST API and docs | Strong third-venue architecture candidate | `secondary_candidate` |
| PredictIt | Real-money-like public political market data | Markets with multiple contracts, often bucketed/ranged | Venue-operated settlement and market rules | Contract quote fields such as best buy/sell yes/no | Public JSON endpoint, low public rate limit, sparse formal docs in this pass | Good additional market-data candidate, high composition pressure | `secondary_candidate` |
| Metaculus | Forecasting platform, not a trading venue | Questions and community forecasts | Admin or platform-driven resolution using cited sources | Forecast probabilities, not exchange quotes | API requires authentication token | Good semantic corpus, poor routing peer | `research_only` |
| Zeitgeist | Decentralized protocol candidate | Protocol-level prediction markets with chain/native infra | Protocol / oracle / chain-native mechanisms | Protocol-specific market data via SDK/indexing path | Official docs, higher integration cost | Plausible future venue, not near-term MVP | `future_candidate` |
| Futuur | Insufficiently verified in this pass | Not fully assessed here | Not fully assessed here | Not fully assessed here | Official docs exist, but access and detail were not fully verified from this environment | Possible future candidate after dedicated research | `insufficient_data` |

### 5.2 What This Matrix Means

The important conclusion is that "prediction market venue" is not one homogeneous category.

The current candidate pool naturally breaks into four classes:

1. real-money routing peers:
   - Polymarket
   - Kalshi
   - possibly PredictIt depending on scope
2. architecturally useful but not automatically route-safe:
   - Myriad
   - Manifold
3. semantic or forecasting corpora:
   - Metaculus
4. future or high-integration-cost protocol candidates:
   - Zeitgeist
   - other decentralized protocols

### 5.3 If Forced to Pick One Third Venue Today

Question: If Equinox adds exactly one more venue now, which should it be?

Current answer:

- If the goal is to learn the most about the long-range architecture, choose `Myriad`.
- If the goal is to stay as close as possible to additional public market-data routing candidates, choose `PredictIt`.
- If the goal is to build a richer semantic evaluation set with route-unsafe controls, keep `Manifold`.

Practical recommendation:

- For research and architecture, prioritize `Myriad`.
- For a later real-money-like expansion experiment, evaluate `PredictIt` next.

## 6. Detailed Venue Research

### 6.1 Polymarket

#### Why It Matters

Polymarket is a strong stress test because it combines:

- public event and market metadata,
- a direct order book,
- grouped event structures,
- negative-risk behavior,
- and decentralized resolution machinery.

#### Object Model

- `event` is the grouping object.
- `market` is the binary tradable proposition.
- Multi-outcome questions are often expressed as multiple binary markets under one event.

Architectural implication:

- Equivalence must happen at the proposition level, not at the event title alone.

#### Quote Model

- Public CLOB provides direct bids and asks.
- This is the cleanest order-book style quote surface among the venues studied here.

Architectural implication:

- Polymarket is the best baseline venue for direct top-of-book and depth normalization.

#### Resolution Model

- Uses UMA Optimistic Oracle with dispute flow.
- Resolution rules and source text matter.

Architectural implication:

- The router cannot treat Polymarket as governance-equivalent to venue-operated or creator-resolved venues just because prices are comparable.

#### Data-Quality Findings

Observed on 2026-03-14:

- Some raw `endDate` fields do not align with the plain-language deadline in the description.
- Example:
  - `OpenAI IPO before 2027?`
  - raw `endDate`: `2026-12-31T00:00:00Z`
  - description deadline: `December 31, 2026, 11:59 PM ET`
- Example:
  - `Aaron Taylor-Johnson announced as next James Bond?`
  - raw `endDate`: `2026-06-30T00:00:00Z`
  - description deadline: `June 30, 2026, 11:59 PM ET`
- Counterexample:
  - `MicroStrategy sells any Bitcoin by June 30, 2026?`
  - raw `endDate` aligned with the stated ET deadline.

Current answer to the field-trust question:

- Raw date fields are useful but not sufficient.
- Deadline extraction needs source ranking, conflict detection, and confidence.

#### Structural Risk

- Negative-risk and augmented negative-risk events can introduce placeholders and `Other`.

Current answer:

- V1 should ingest these flags and store them, but should route only on clearly named outcomes unless a later phase adds explicit support.

### 6.2 Kalshi

#### Why It Matters

Kalshi is the clearest exchange-style contrast to Polymarket.

It differs in:

- object hierarchy,
- quote representation,
- numeric precision,
- historical-data partitioning,
- and supported market families.

#### Object Model

- `series` groups related events.
- `event` is a primary user-facing grouping.
- `market` is the tradeable contract.

Architectural implication:

- Event titles are not enough; the meaningful tradable unit is still the market-level proposition after normalization.

#### Quote Model

- Public orderbook response exposes bids only.
- Asks must be inferred by reciprocity from the opposite side.

Current answer:

- Quote normalizers must carry both `native` and `implied` values.
- The router must know when an ask is observed versus inferred.

#### Numeric and Fee Model

- Fixed-point dollar strings and fixed-point counts are documented.
- Some markets allow subpenny pricing and fractional contracts.
- Fee rounding is documented.

Current answer:

- Numeric normalization must preserve precision and should avoid float sloppiness in the implementation.

#### Historical Split

- Older market data moves to `/historical/...` endpoints.

Current answer:

- Historical fixture generation and replay should be handled inside the Kalshi adapter boundary, not in generic router code.

#### Documentation / Behavior Inconsistency

Observed on 2026-03-14:

- The orderbook guide says the REST orderbook endpoint is public.
- The endpoint reference declares auth headers.
- Live unauthenticated GET requests succeeded.

Current answer:

- Engineering should trust live endpoint behavior enough for the prototype, but record the inconsistency and keep fixture mode ready.

#### Unsupported-Shape Signal from Live Data

Observed on 2026-03-14:

- A broad `GET /markets?limit=2&status=open` sample returned multi-game sports combo contracts.
- These contracts were labeled `market_type = binary`.
- Their titles encoded many joint conditions in one proposition string.
- `settlement_sources` was `null` for the sample markets inspected.

Architectural implication:

- Venue-provided `binary` does not mean "simple binary proposition."
- Composition type must be normalized independently.

#### Metadata Reliability Finding

Observed on 2026-03-14:

- `KXBOND-30` returned `mutually_exclusive: false` despite the outcome set appearing candidate-exclusive.

Current answer:

- Exclusivity flags should be treated as signals, not proof.

### 6.3 Manifold

#### Why It Matters

Manifold is useful precisely because it is not a clean routing peer.

It pressures the system to separate:

- semantic equivalence,
- route safety,
- and quote comparability.

#### Economic and Governance Model

- Play-money economy using `MANA`.
- Creator resolves markets.
- Moderators can override.

Current answer:

- Manifold should not be a default routing peer for real-money venues.
- It remains valuable as a semantic research venue and source of route-unsafe positives.

#### Quote Model

- Limit orders plus AMM.
- Public market payloads expose probabilities, pools, and liquidity.

Current answer:

- Manifold quote normalization should be treated as a distinct adapter family, not shoehorned into a CLOB abstraction.

#### API and Data Posture

- Official API is explicitly alpha.
- Public GET endpoints are available.
- Data docs note bulk dumps, but the published bulk-dump freshness observed in docs lags far behind live API needs.

Current answer:

- Use live API for research and fixtures, not old bulk dumps as a primary source.

### 6.4 Myriad

#### Why It Matters

Myriad is the most important new finding in this pass.

Why:

- It exposes `questions` as first-class objects.
- Questions contain one or more markets.
- Market payloads expose `resolutionSource` and `resolutionTitle`.
- The model is already close to the proposition-registry design Equinox needs.

#### Observed Model from Public API

Observed on 2026-03-14:

- `questions` endpoint returns `title`, `expiresAt`, `marketCount`, and nested `markets`.
- Nested markets expose:
  - `id`
  - `title`
  - `expiresAt`
  - `resolvesAt`
  - `resolutionSource`
  - `resolutionTitle`
  - `state`
  - `networkId`
  - `outcomes`

Architectural implication:

- Myriad provides a useful reference model for separating `question identity` from `market instances`.

#### Quote / Pricing Model

- Market payloads expose per-outcome `price`.
- Fees are exposed in the market response.

Current answer:

- Myriad is easier to normalize than Manifold for proposition identity, but still differs from direct-book venues.

#### Content Distribution

Observed on 2026-03-14:

- Sample open questions included crypto micro-markets and multi-outcome sports futures.
- An `OpenAI` keyword search returned a resolved binary question with explicit `Associated Press` resolution title and source link.

Current answer:

- Myriad looks promising structurally, but live proposition overlap with Polymarket or Kalshi may still be sparse depending on the topic slice.

#### Third-Venue Verdict

Current answer:

- Myriad is the best third venue if the goal is to harden the canonical proposition architecture.

### 6.5 PredictIt

#### Why It Matters

PredictIt is a more meaningful candidate than it looked in the first pass because it does expose public market data with current price fields.

#### Observed Public API Behavior

Observed on 2026-03-14:

- `GET https://www.predictit.org/api/marketdata/all/` returned public JSON.
- Response headers exposed rate-limit headers.
- Observed public limit in headers:
  - `X-Rate-Limit-Limit-Minute: 5`

Current answer:

- PredictIt is accessible enough for research and perhaps low-rate polling, but the rate budget is tight.

#### Market Shape

Observed sample market:

- `How many House seats will Republicans win in the 2026 midterm election?`
- Contracts included:
  - `192 or fewer`
  - `193 to 197`
  - `198 to 202`

Architectural implication:

- PredictIt often encodes bucketed or ranged markets rather than simple binary propositions.
- These are not safe to treat as generic binary equivalents.

#### Field Reliability Finding

Observed on 2026-03-14:

- Sample contract-level `dateEnd` field returned `NA`.

Current answer:

- PredictIt would require stronger deadline extraction from market-level metadata and rules text.

#### Quote Model

- Contract payloads expose fields such as:
  - `lastTradePrice`
  - `bestBuyYesCost`
  - `bestBuyNoCost`
  - `bestSellYesCost`
  - `bestSellNoCost`

Current answer:

- PredictIt has a usable quote surface for routing research, but composition type and rate limits are first-order issues.

#### Third-Venue Verdict

Current answer:

- PredictIt is the best additional public market-data candidate if the goal is to stay near the real-money routing problem.
- It is not the best choice if the immediate goal is to simplify canonical proposition architecture.

### 6.6 Metaculus

#### Why It Matters

Metaculus is not a routing venue, but it is still relevant because it is a structured forecasting system with resolution mechanics and a large semantic question corpus.

#### Access Posture

Observed on 2026-03-14:

- Public API requests without authentication returned an error indicating that API access requires an authenticated token.

Current answer:

- Metaculus is not a low-friction MVP integration target for routing research from the current environment.

#### Governance Model

- Official help pages indicate that admins resolve questions and may rely on cited sources.

Current answer:

- Metaculus belongs in the semantic or forecasting class, not the routing-peer class.

### 6.7 Zeitgeist

#### Why It Matters

Zeitgeist is useful as a signal of what future decentralized integrations may look like.

#### What We Can Say Confidently from Official Docs

- Official docs exist.
- The integration path appears SDK- and protocol-oriented rather than simple public REST ingestion.
- Market metadata and protocol-level concerns are more chain-native than the primary venues in this research.

Current answer:

- Zeitgeist is a plausible future candidate for a later, more protocol-heavy Equinox phase.
- It is not the best next step for the MVP.

### 6.8 Futuur

Question: Is Futuur a serious candidate?

Current answer:

- Official docs exist, but this pass did not gather enough accessible official detail from the current environment to recommend Futuur for the MVP.
- It should remain in the "dedicated future research" bucket until public market-data posture, auth requirements, and resolution model are verified.

## 7. Cross-Venue Mismatch Taxonomy

### 7.1 Economic Domain Mismatch

Examples:

- Polymarket and Kalshi are economically comparable in a way that Manifold is not.
- Myriad may be closer to tokenized prediction markets than to exchange-style venues.

Current answer:

- Economic domain must be part of route-safety policy, not a footnote.

### 7.2 Resolution Authority Mismatch

Examples:

- Oracle/dispute driven
- exchange/venue rule driven
- creator driven
- admin/platform driven

Current answer:

- Resolution authority is a core semantic field and a route-safety field.

### 7.3 Market Shape Mismatch

Examples:

- Polymarket grouped binary events
- Kalshi scalar and combo events
- PredictIt bucketed contract families
- Manifold non-binary markets
- Myriad multi-outcome futures

Current answer:

- Market shape must be normalized beyond a shallow binary/non-binary flag.

### 7.4 Composition-Type Mismatch

Examples:

- simple binary proposition
- bucket inside a ranged market
- combo contract that encodes many conditions
- placeholder or `Other` outcome

Current answer:

- Composition type is separate from market type and must be normalized explicitly.

### 7.5 Deadline and Settlement-Window Mismatch

Examples:

- one venue resolves by a 2026 deadline,
- another resolves by a 2030 deadline,
- one venue closes trading before the real semantic cutoff,
- another resolves later.

Current answer:

- Semantic deadline must come from a trust-ranked extraction process, not blindly from a raw timestamp.

### 7.6 Field-Reliability Mismatch

Examples:

- Polymarket raw `endDate` conflicts with description.
- PredictIt contract `dateEnd` can be `NA`.
- Kalshi metadata flags can understate exclusivity.
- Kalshi broad market feeds can mark combo contracts as `binary`.

Current answer:

- Canonicalization is not field renaming. It is field interpretation with confidence.

### 7.7 Quote-Model Mismatch

Examples:

- direct bid/ask book
- bid-only reciprocal book
- AMM probability surface
- per-outcome price without full displayed book

Current answer:

- Quote normalization must preserve quote model and confidence, not just emit one generic price field.

### 7.8 Fee-Model Mismatch

Examples:

- venue fees visible in docs or payloads,
- partial fee visibility,
- play-money where dollar fee comparability is meaningless.

Current answer:

- Net-price routing requires fee visibility or a routing penalty/block.

### 7.9 Freshness and Access Mismatch

Examples:

- public WebSocket versus auth-only WebSocket
- public REST versus token-gated API
- low public rate limits versus generous public reads

Current answer:

- Ingestion strategy must be venue-specific while outputs remain normalized.

### 7.10 Licensing and Usage Mismatch

Examples:

- Manifold bulk data licensing constraints for some use cases
- varying public API expectations across venues

Current answer:

- Prototype scope is still fine, but long-range data strategy must track licensing and operational access rules.

## 8. Major Architectural Challenges

### 8.1 Canonical Proposition Identity

Problem:

- venue IDs are local,
- titles drift,
- event groupings differ,
- and one venue may represent a proposition as a simple binary while another represents it as one member of a family.

Current answer:

- Equinox needs a canonical proposition layer with many-to-one attachment from venue contracts.

### 8.2 Pairwise Matching Does Not Scale

Problem:

- with 2 venues, pairwise comparison feels natural;
- with 3 venues, pairwise edges become harder to reason about;
- with 10+ venues, all-versus-all comparison becomes operationally noisy and conceptually weak.

Current answer:

- matching should evolve from pairwise comparison toward proposition clustering over a canonical registry.

### 8.3 Market Decomposition and Composition

Problem:

- one venue may group outcomes,
- another may bucket a range,
- another may expose a compound combo as a binary contract.

Current answer:

- the system needs explicit `composition_type` and unsupported-case handling.

### 8.4 Temporal Semantics

Problem:

- close time, end date, resolution deadline, and real semantic cutoff are not always the same thing.

Current answer:

- canonical models should distinguish:
  - trade close,
  - semantic deadline,
  - resolution time if known,
  - and deadline confidence.

### 8.5 Resolution Rule Parsing and Versioning

Problem:

- resolution semantics often live in prose,
- rules can change,
- and multiple fields may conflict.

Current answer:

- store raw rules, extracted summaries, and provenance.
- if implementation later introduces AI or NLP extraction, keep heuristics as the baseline and log every extraction source.

### 8.6 Quote Normalization

Problem:

- a "best ask" may be observed directly, inferred, or unavailable.
- depth may be rich, sparse, or absent.

Current answer:

- quote objects need typed metadata such as:
  - `quote_model`
  - `best_ask_is_implied`
  - `depth_confidence`
  - `fee_visibility`

### 8.7 Fill Simulation

Problem:

- displayed liquidity is not equally informative across venues,
- fee models vary,
- and some venues expose probability more readily than executable depth.

Current answer:

- MVP routing should remain modest: compare normalized entry cost, visible depth where available, freshness, and capability confidence.

### 8.8 Route Comparability Policy

Problem:

- even semantically equivalent propositions may not be good routing peers.

Current answer:

- add route-comparability domains and policy gates.
- examples:
  - `strict_real_money_only`
  - `include_tokenized`
  - `semantic_only`

### 8.9 Venue Capability Registry

Problem:

- if the router knows venue names, it will accumulate special cases.

Current answer:

- the router should consume normalized capability profiles and policy flags, not switch on venue strings.

### 8.10 Backfill, Polling, and Snapshotting

Problem:

- venues differ in rate limits, history availability, and stream access.

Current answer:

- fixture generation should be a first-class subsystem, not an afterthought.

### 8.11 Provenance and Explainability

Problem:

- reviewers need to understand why the system normalized or rejected something.

Current answer:

- every normalized field that matters should point back to a source payload or rule text origin.

### 8.12 Evaluation and Human Review

Problem:

- ambiguity will not disappear.

Current answer:

- Equinox needs a labeled set containing:
  - exact-equivalent and route-safe examples,
  - semantic-but-route-unsafe examples,
  - near-match unsafe examples,
  - clear non-matches,
  - and field-conflict examples.

### 8.13 Schema Evolution

Problem:

- public APIs change.

Current answer:

- store raw payloads, adapter versions, and fixture snapshots so model changes remain debuggable.

## 9. Adding a Third Market Today

### 9.1 Question: What changes technically if a third market is added?

Current answer:

At minimum, Equinox must add:

- a new adapter,
- a venue capability profile,
- outcome mapping rules,
- composition-type handling,
- and fixture coverage for that venue.

If those changes require router rewrites, the architecture is too pairwise.

### 9.2 Question: Which third market is best right now?

Current answer:

- `Myriad` is the best architecture-expanding third venue.
- `PredictIt` is the best additional public market-data candidate if the goal is to stay closer to routable real-money-style comparisons.
- `Manifold` is still the best route-unsafe semantic-control venue.

### 9.3 Question: Should the third venue be included in routing immediately?

Current answer:

Not necessarily.

The third venue can be added in one of three modes:

1. `routing_peer`
2. `semantic_only`
3. `research_only`

Recommended answer:

- Add Myriad or Manifold initially as `semantic_only` or `research_only` unless a dedicated pass validates routing comparability.

### 9.4 Question: What breaks first if the design is weak?

Current answer:

The first failure modes will be:

- venue-pair branching in matcher code,
- unsupported contract families being misclassified as simple binaries,
- and router logic pretending price fields are equally executable across venues.

## 10. Long-Range Architecture for 10+ Venues

### 10.1 High-Level Answer

If Equinox eventually wants to route between 10+ venues, it cannot remain:

- adapter-driven only,
- pairwise only,
- or title-similarity driven.

It needs a layered architecture with explicit identity, policy, and evidence models.

### 10.2 Recommended Layers

1. `Venue Adapter Layer`
   - Pull raw venue data.
   - Handle auth, rate limits, paging, history, retries, and transport quirks.

2. `Normalized Raw Warehouse`
   - Store raw payloads and minimally normalized records.
   - Keep source provenance intact.

3. `Venue Capability Registry`
   - Record quote model, economic domain, governance model, supported structures, fee visibility, and history posture.

4. `Canonical Proposition Registry`
   - Maintain event and proposition identities independent of venue.
   - Allow many venue contracts to attach to one canonical proposition.

5. `Candidate Retrieval Engine`
   - Use blocking signals like entity, category, time window, and lexical similarity.
   - Avoid full all-versus-all scans where possible.

6. `Equivalence Engine`
   - Score semantic compatibility.
   - Attach evidence and blocking reasons.
   - Output cluster membership or pairwise decisions.

7. `Quote Normalization Service`
   - Normalize top-of-book, implied asks, AMM probabilities, depth confidence, and fee visibility into a typed quote model.

8. `Routing Policy Engine`
   - Enforce comparability domains and route-safety rules.
   - Produce `route_to_<venue>` or `do_not_route`.

9. `Fixture and Evaluation System`
   - Freeze snapshots.
   - Run labeled tests and regression suites.

10. `Analyst Review Loop`
   - Collect hard cases and feed them back into heuristics and fixtures.

### 10.3 Data Model Direction

The long-range model should look more like this:

- `venue_contract`
- `normalized_contract`
- `canonical_event`
- `canonical_proposition`
- `proposition_membership`
- `equivalence_evidence`
- `quote_snapshot`
- `routing_policy_result`

Why:

- a pure left-right match record is too limited once three or more venues can point to the same proposition.

### 10.4 Candidate Retrieval at Scale

In a 10+ venue world, pairwise comparison explodes.

Current answer:

- do not compare every contract to every other contract by default.
- first reduce the search space with:
  - category,
  - entity extraction,
  - date bucket,
  - venue domain,
  - and market family filters.

### 10.5 Equivalence at Scale

Current answer:

- use a scoring pipeline that can either:
  - attach a contract to an existing canonical proposition,
  - create a new canonical proposition,
  - or reject attachment and log why.

This is better than permanently recomputing every pair independently.

### 10.6 Routing Policy at Scale

Current answer:

- route policy should not ask "is venue A better than venue B?"
- it should ask:
  - which candidate contracts are semantically attached to the proposition,
  - which are route-safe under the active policy,
  - and which has the best normalized execution profile among the route-safe set.

### 10.7 New Venue Onboarding Checklist

If Equinox adds a new venue later, the onboarding checklist should include:

1. access posture:
   - public REST
   - token-gated REST
   - WebSocket
   - SDK only
2. market family inventory:
   - simple binary
   - bucketed range
   - scalar
   - combo
   - multi-outcome
   - placeholder / other
3. resolution model:
   - authority
   - governance
   - dispute model
   - rule availability
4. quote model:
   - direct book
   - implied book
   - AMM
   - outcome price only
5. fee visibility:
   - clear
   - partial
   - opaque
6. historical posture:
   - accessible
   - split live/history
   - fixture only
7. route domain:
   - routing peer
   - semantic only
   - research only

### 10.8 What Fails Without This Architecture

If Equinox does not evolve this way, a 10+ venue version will likely suffer from:

- duplicated venue-pair heuristics,
- inconsistent equivalence standards,
- fragile router branching,
- opaque unsupported-case handling,
- and a testing matrix that becomes impossible to trust.

## 11. Live Empirical Examples

All observations in this section were made on 2026-03-14.

### 11.1 Strong Semantic Match, Route-Unsafe Pair

Pair:

- Polymarket: `OpenAI IPO before 2027?`
- Manifold: `OpenAI IPO before 2027?`

Observed answer:

- Strong semantic alignment.
- Material governance and economic mismatch.

Classification answer:

- `semantic_equivalent_but_route_unsafe`

### 11.2 High-Lexical-Overlap Near-Match

Pair:

- Polymarket: `Aaron Taylor-Johnson announced as next James Bond?`
- Kalshi: `Will Aaron Taylor-Johnson be the next James Bond?`

Observed answer:

- Real-world subject aligns.
- Deadline and likely trigger semantics do not align.

Classification answer:

- `near_equivalent_but_not_safe`

### 11.3 Related Topic, Different Proposition

Pair:

- Polymarket: Democratic control of the House after the 2026 midterms
- Kalshi: Hakeem Jeffries as next Speaker

Observed answer:

- Topic adjacency is not proposition identity.

Classification answer:

- `not_equivalent`

### 11.4 PredictIt Range-Bucket Example

Observed example:

- PredictIt market: `How many House seats will Republicans win in the 2026 midterm election?`
- Contracts:
  - `192 or fewer`
  - `193 to 197`
  - `198 to 202`

Architectural answer:

- These are bucket members of a range family, not interchangeable with simple single-threshold binaries.

### 11.5 Kalshi Combo Example

Observed example:

- Open Kalshi sample returned multi-game sports contracts with many conditions embedded in one `binary` market title.

Architectural answer:

- Venue-reported `binary` must not be treated as proof of simple proposition structure.

### 11.6 Myriad Question-Market Example

Observed example:

- Myriad `questions` endpoint returns a question-level object with nested markets and explicit resolution source/title.

Architectural answer:

- This is strong evidence that Equinox should separate canonical question or event identity from market-instance identity.

### 11.7 Polymarket Deadline-Conflict Example

Observed example:

- raw structured field and plain-language description can disagree materially about the semantic deadline.

Architectural answer:

- deadline provenance and confidence are mandatory canonical fields.

## 12. Concrete Architectural Recommendations

### 12.1 Recommendation: Treat Equinox as a Semantic Safety System

Current answer:

- do not frame the system as simple schema normalization.
- frame it as semantic normalization plus route-safety evaluation.

### 12.2 Recommendation: Separate Semantic Equivalence from Route Safety

Current answer:

- this distinction is necessary, not optional.

### 12.3 Recommendation: Normalize Composition Type

Current answer:

- add explicit composition categories such as:
  - `simple_binary`
  - `bucket_member`
  - `combo_binary`
  - `scalar_derived`
  - `multi_outcome_member`
  - `placeholder_or_other`

### 12.4 Recommendation: Build Around Canonical Proposition Membership

Current answer:

- the data model should allow many venue contracts to attach to one proposition.

### 12.5 Recommendation: Introduce Route Domains

Current answer:

- not every semantic match should be routable.

### 12.6 Recommendation: Invest Early in Fixtures

Current answer:

- fixture mode is mandatory because live overlap is sparse and venue behavior changes.

### 12.7 Recommendation: Preserve Evidence for Every Important Decision

Current answer:

- raw payload link,
- extracted rule summary,
- deadline source,
- and blocker reasons should be preserved for each decision.

### 12.8 Recommendation: Use a Local-First, Fixture-First Demo Posture

Current answer:

- the prototype should be demoed locally from the project root,
- fixture mode should be the default reviewer path,
- and live mode should be a secondary proof that current public APIs can still be ingested.

Why:

- live overlap is sparse,
- public APIs and docs drift,
- some venues have low rate limits or inconsistent behavior,
- and the spec explicitly values clarity and reasoning over infrastructure sophistication.

## 13. Recommended Refinements to PRD v001

This section captures the main PRD changes suggested by the research.

Current answer:

1. Keep `Polymarket + Kalshi` as the primary routing pair.
2. Add explicit support for `composition_type` so binary-looking combos and bucket members are not misclassified.
3. Add the ability for more than two source contracts to attach to one canonical proposition.
4. Add scalability guidance so a third venue can be added without router rewrites.
5. Add route-comparability language for future multi-venue growth.
6. Preserve deadline source and conflict handling as first-class canonical fields.
7. Keep Manifold in the research or semantic-only lane by default.
8. Evaluate Myriad and PredictIt as next-step venue candidates for future phases.
9. Define an explicit local reviewer demo path with fixture-first execution and documented live-mode fallback.

## 14. Remaining Research Gaps and Current Best Answers

This section captures the remaining open questions and the best current answers.

### 14.1 Question: Is there enough live exact overlap to demonstrate routing convincingly?

Current best answer:

- Probably not reliably enough to depend on live overlap alone.
- Fixture mode remains mandatory.

### 14.2 Question: Is Myriad a better third venue than PredictIt?

Current best answer:

- For architecture learning, yes.
- For public real-money-like quote comparison, not necessarily.

### 14.3 Question: Can PredictIt be treated as a clean routing peer?

Current best answer:

- Not cleanly yet.
- It is promising, but composition type, low public rate limits, and sparse deadline fields need a dedicated follow-up pass.

### 14.4 Question: Should Metaculus be integrated at all?

Current best answer:

- Only if Equinox later wants a semantic or forecasting comparison mode.
- It is not needed for the MVP routing problem.

### 14.5 Question: Should decentralized protocol venues be part of the MVP?

Current best answer:

- No.
- They are valuable future stress tests, but they add integration cost before the core proposition and routing architecture is proven.

### 14.6 Remaining Submission Risks

Current best answer:

The main remaining submission risks are:

1. Some multi-venue and scaling language may still look more ambitious than the MVP strictly requires.
2. The prototype can still fail in practice if the final repository does not make the fixture demo extremely easy to run.
3. Exact-equivalent live overlap may remain too sparse, so the final implementation must prove that fixture mode is not hiding architectural weakness.
4. Quote normalization and fee handling may still need to stay conservative if one venue exposes weaker execution data than another.
5. The chosen stack must be justified briefly and concretely, or the implementation may look more complex than the spec requires.

Mitigation direction:

- keep the code path narrow,
- keep the demo path short,
- and make the reasoning artifacts explicit.

## 15. Source Index

### 15.1 Polymarket

- [Overview](https://docs.polymarket.com/market-data/overview)
- [Markets & Events](https://docs.polymarket.com/concepts/markets-events)
- [Prices & Orderbook](https://docs.polymarket.com/concepts/prices-orderbook)
- [Resolution](https://docs.polymarket.com/concepts/resolution)
- [Negative Risk Markets](https://docs.polymarket.com/advanced/neg-risk)
- [Fetching Markets](https://docs.polymarket.com/market-data/fetching-markets)
- [Rate Limits](https://docs.polymarket.com/api-reference/rate-limits)
- [List Events API](https://docs.polymarket.com/api-reference/events/list-events.md)
- [Get Order Book API](https://docs.polymarket.com/api-reference/market-data/get-order-book.md)
- [Fees](https://docs.polymarket.com/trading/fees)
- [Market Channel WebSocket](https://docs.polymarket.com/market-data/websocket/market-channel)
- [Geographic Restrictions](https://docs.polymarket.com/api-reference/geoblock)
- [Live Events Endpoint](https://gamma-api.polymarket.com/events?closed=false&limit=1)
- [Live Markets Endpoint](https://gamma-api.polymarket.com/markets?closed=false&limit=1)
- [Live CLOB Order Book](https://clob.polymarket.com/book)

### 15.2 Kalshi

- [Kalshi Glossary](https://docs.kalshi.com/getting_started/terms)
- [Get Events API](https://docs.kalshi.com/api-reference/events/get-events.md)
- [Get Event API](https://docs.kalshi.com/api-reference/events/get-event.md)
- [Get Markets API](https://docs.kalshi.com/api-reference/market/get-markets.md)
- [Get Market Orderbook API](https://docs.kalshi.com/api-reference/market/get-market-orderbook.md)
- [Orderbook Responses](https://docs.kalshi.com/getting_started/orderbook_responses)
- [Fixed-Point Migration](https://docs.kalshi.com/getting_started/fixed_point_migration)
- [Fee Rounding](https://docs.kalshi.com/getting_started/fee_rounding)
- [Understanding Pagination](https://docs.kalshi.com/getting_started/pagination)
- [Historical Data](https://docs.kalshi.com/getting_started/historical_data)
- [Rate Limits and Tiers](https://docs.kalshi.com/getting_started/rate_limits)
- [Get Multivariate Events API](https://docs.kalshi.com/api-reference/events/get-multivariate-events.md)
- [Market Ticker WebSocket](https://docs.kalshi.com/websockets/market-ticker)
- [Get Series Fee Changes API](https://docs.kalshi.com/api-reference/exchange/get-series-fee-changes.md)
- [Live Events Endpoint](https://api.elections.kalshi.com/trade-api/v2/events?limit=1&status=open)
- [Live Markets Endpoint](https://api.elections.kalshi.com/trade-api/v2/markets?limit=2&status=open)

### 15.3 Manifold

- [Manifold API](https://docs.manifold.markets/api)
- [Raw API Docs Source](https://raw.githubusercontent.com/manifoldmarkets/manifold/main/docs/docs/api.md)
- [Manifold FAQ](https://docs.manifold.markets/faq)
- [Raw FAQ Source](https://raw.githubusercontent.com/manifoldmarkets/manifold/main/docs/docs/faq.md)
- [Manifold Data](https://docs.manifold.markets/data)
- [Raw Data Docs Source](https://raw.githubusercontent.com/manifoldmarkets/manifold/main/docs/docs/data.md)
- [Live Search API](https://api.manifold.markets/v0/search-markets)

### 15.4 Myriad

- [Myriad Docs](https://docs.myriad.markets/)
- [Myriad API Reference](https://docs.myriad.markets/myriad-api-spec)
- [Live Questions Endpoint](https://api-v2.myriadprotocol.com/questions?page=1&limit=2)
- [Live Markets Endpoint](https://api-v2.myriadprotocol.com/markets?page=1&limit=2&state=open)

### 15.5 PredictIt

- [Public Market Data Endpoint](https://www.predictit.org/api/marketdata/all/)

### 15.6 Metaculus

- [Metaculus FAQ](https://www.metaculus.com/faq/)
- [Metaculus About](https://www.metaculus.com/about/)
- [API Endpoint Requiring Auth](https://www.metaculus.com/api2/questions/?limit=2)

### 15.7 Zeitgeist

- [Zeitgeist Docs](https://docs.zeitgeist.pm/)

### 15.8 Futuur

- [Futuur Docs](https://docs.futuur.com/)

## 16. Bottom Line

The most important conclusion from this expanded presearch is that Equinox should be designed as a proposition identity and route-safety system, not as a thin adapter layer plus text matching.

If the MVP proves anything, it should prove these claims:

- normalized proposition identity is possible,
- unsafe lookalikes can be rejected for principled reasons,
- routing can be simulated without venue-specific router logic,
- and the architecture has a credible path from 2 venues to 3 and eventually to many more.

The best next-step implementation posture remains:

- primary routing pair: `Polymarket + Kalshi`
- mandatory fixture mode
- third-venue research candidate: `Myriad`
- additional public market-data candidate for later follow-up: `PredictIt`
- semantic control venue: `Manifold`
