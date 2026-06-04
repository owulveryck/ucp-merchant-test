---
status: accepted
date: 2026-05-29
---

# ADR-0001: Multi-Agent Architecture for Competitive Pricing

## Problem

Merchants lose sales without understanding why. Competitors display $60 but actually sell at $54 (hidden WELCOME10 code). Traditional algorithms don't detect these hidden discounts.

## Decision

Architecture with 4 specialized agents running in sequence:
- **Agent 1**: Detects competitor promo codes, calculates real prices
- **Agent 2**: Analyzes market position
- **Agent 3**: Recommends pricing strategy
- **Agent 4**: Validates margin and cost constraints

## Why

- Transparent: Each agent explains its reasoning
- Modular: Change one agent without touching others
- Extensible: Add Agent 5 (advertising) or Agent 6 (inventory) easily

## Consequences

**Positive**
- Merchants see complete reasoning
- Independent modification of each agent
- New capabilities without rewriting code

**Negative**
- 4 files instead of 1 function
- Sequential latency (<2s, acceptable)

## Validation

- Performance: <2s end-to-end
- Dashboard shows reasoning from all 4 agents
- Unit tests per agent + integration tests

## Implementation

`pkg/merchant/competitive/orchestrator.go`
`pkg/merchant/competitive/agents/`
