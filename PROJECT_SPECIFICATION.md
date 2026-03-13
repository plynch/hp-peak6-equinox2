# Project Spec

## Project Title: Equinox

Cross-Venue Prediction Market Aggregation & Routing Simulation

### Overview

Prediction markets trade real-world events across multiple venues. The same event may exist on
different platforms under different naming conventions, structures, and liquidity conditions.

There is no unified way to detect equivalence across venues or compare pricing in a normalized
manner.

Project Equinox explores whether it is technically feasible to identify equivalent markets across
prediction platforms, normalize them into a shared internal representation, and simulate routing
decisions between venues. This is an infrastructure prototype, not a trading product.

### Problem Statement

Prediction markets are fragmented. Markets representing the same real-world outcome may
differ across venues in naming, expiration logic, contract design, or pricing format. This
fragmentation makes cross-venue comparison difficult and complicates any attempt at intelligent
routing or unified access.

The core challenge is determining whether market equivalence can be programmatically
inferred and whether a venue-agnostic routing layer can be built on top of that normalization.

### Objectives

The objective of this project is to build a working prototype that connects to at least two
prediction market venues, defines a canonical internal market model, attempts to detect
equivalent markets across venues, and simulates routing decisions for hypothetical trades. The
system should log and explain its matching and routing decisions.
We are not looking for production polish. We are evaluating architectural thinking, tradeoff
awareness, and clarity of reasoning.

### Scope

The prototype should integrate with public APIs from two prediction market venues and ingest
both market metadata and pricing data. It should define an internal representation of a market
that is independent of venue-specific schemas. It should attempt to match markets that refer to
the same underlying real-world event and simulate a routing decision for a hypothetical order.

Real-money trading, wallet integration, regulatory implementation, and production UI are
explicitly out of scope.

### Core Expectations

The system should include a clear separation between venue integration, normalization,
equivalence detection, and routing logic. Routing logic should not contain venue-specific
assumptions.

Candidates must define what “equivalent” means and justify their methodology. Matching may
be rule-based, heuristic, AI-assisted, or hybrid. The approach must be documented.
The routing engine should evaluate available venues for a hypothetical order, produce a
decision, and clearly explain why that venue was selected. We are less concerned with
optimizing execution quality than with understanding the reasoning and structure behind the
decision.

The system should handle imperfect data gracefully and document assumptions made where
information is incomplete or ambiguous.

### Technical Flexibility

Our internal stack frequently uses Go and Google Cloud Platform, but candidates may choose
their preferred tools if justified. AI usage is optional. If used, it must be clearly documented along
with the reasoning behind its application.
Local deployment is acceptable. Clean structure and clarity are more important than
infrastructure sophistication.

### Deliverables

Candidates should provide a working prototype, setup instructions, a brief architecture overview,
and written explanations of their equivalence logic and routing logic. If AI tools were used during
development, that usage should be disclosed.

### Evaluation Criteria

We are evaluating how the problem is framed, how the system is decomposed, how ambiguity is
handled, and how decisions are justified. Code readability, modularity, and documentation
quality matter. UI polish does not.

The goal of Project Equinox is to assess whether cross-venue normalization and routing
infrastructure is viable, and how a candidate approaches designing such a system under
real-world ambiguity.