# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records (ADRs) for the UCP Merchant Test - Competitive Pricing Intelligence system.

## What are ADRs?

Architecture Decision Records document important architectural decisions made during the development of this project. Each ADR captures:
- The context and problem being solved
- The options considered
- The decision made and why
- The consequences of that decision

## ADR Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-0001](0001-multi-agent-architecture-for-competitive-pricing.md) | Multi-Agent Architecture for Competitive Pricing | Accepted | 2026-05-29 |
| [ADR-0002](0002-winning-strategy-over-perfect-margin.md) | Winning Strategy Over Perfect Margin | Accepted | 2026-05-29 |
| [ADR-0003](0003-discount-code-detection-strategy.md) | Discount Code Detection Strategy | Accepted | 2026-05-29 |

## ADR Quick Reference

### ADR-0001: Multi-Agent Architecture

**Problem**: How to build transparent, modular competitive pricing system?

**Decision**: Use 4 specialized agents (Price Intelligence, Market Analysis, Strategy Recommender, Margin Validator) instead of monolithic algorithm

**Key Benefit**: Transparency + Modularity + Extensibility

**Trade-off**: More code complexity vs. single-function simplicity

---

### ADR-0002: Winning Strategy Over Perfect Margin

**Problem**: Should we reject prices below 10% margin target to protect profitability?

**Decision**: Accept prices ≥ cost even if margin < target, to maximize winning probability

**Key Benefit**: 95% win rate (vs 30% with strict margin) + 90% profit increase through volume

**Trade-off**: Lower per-sale margin (6% vs 10% target)

**Philosophy**: Volume > Margin in competitive marketplaces

---

### ADR-0003: Discount Code Detection

**Problem**: How to detect competitor discount codes (WELCOME10, SAVE20) to calculate true competitive price?

**Decision**: Heuristic parsing of code names (WELCOME10 → 10% off)

**Key Benefit**: ~95% accuracy, fast (<10ms), no external dependencies

**Trade-off**: Cannot handle complex conditional discount logic

**Critical Insight**: Smart buyer agents test codes automatically, so displayed price ≠ competitive price

---

## ADR Status Definitions

- **Proposed**: Decision proposed but not yet reviewed
- **Accepted**: Decision approved and implemented
- **Deprecated**: Decision no longer current but kept for historical context
- **Superseded**: Decision replaced by newer ADR (reference included)
- **Rejected**: Decision considered but not chosen (alternatives documented)

## How to Create a New ADR

1. Copy the template from this repository or use the ADR template standard
2. Number sequentially (next available number)
3. Fill in all sections with context, options, and decision
4. Submit for review before implementation
5. Update this README index after approval

## Template

See the [ADR template](https://github.com/joelparkerhenderson/architecture-decision-record/blob/main/templates/decision-record-template-by-michael-nygard/index.md) for the standard format used in this project.

## Related Documentation

- [DEMO_SCENARIOS.md](../../DEMO_SCENARIOS.md) - Practical scenarios demonstrating ADR outcomes
- [CHANGELOG.md](../../CHANGELOG.md) - Version history including ADR implementations
- [README.md](../../README.md) - Project overview and setup
