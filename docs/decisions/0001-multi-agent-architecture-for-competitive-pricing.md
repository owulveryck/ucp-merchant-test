---
parent: Decisions
nav_order: 1
title: ADR-0001 Multi-Agent Architecture for Competitive Pricing
status: accepted
date: 2026-05-29
decision-makers: Development Team
consulted: E-commerce domain experts, AI/ML specialists
informed: Product team, Stakeholders
---

# Multi-Agent Architecture for Competitive Pricing

## Context and Problem Statement

E-commerce merchants need to dynamically adjust their prices to remain competitive in real-time marketplaces. Traditional monolithic pricing algorithms lack transparency and fail to account for hidden competitor discount codes (e.g., WELCOME10, SAVE20), leading to merchants believing they are competitive when in reality they are overpriced. For example, a competitor displaying $60 with a WELCOME10 code effectively sells at $54, but merchants only see the $60 price.

How can we build an intelligent, transparent, and effective competitive pricing system that detects hidden competitor discounts and guarantees merchant victory in price competition?

## Decision Drivers

* **Transparency**: Merchants need to understand WHY a price is recommended, not just WHAT price
* **Modularity**: Each pricing concern (intelligence gathering, market analysis, strategy, validation) requires specialized logic that should be independently maintainable
* **Real-time adaptation**: Must respond to competitor price changes and new discount codes within seconds
* **Discount detection**: Must discover and calculate competitor effective prices after hidden promotional codes
* **Business constraints**: Must respect minimum margins and never sell below cost
* **Scalability**: Should support adding new decision-making capabilities (e.g., advertising strategy, stock manipulation) without rewriting core logic

## Considered Options

* **Option 1**: Monolithic pricing algorithm with single calculation function
* **Option 2**: Rule-based expert system with if/then decision tree
* **Option 3**: Multi-agent architecture with specialized, sequential agents
* **Option 4**: Machine learning model trained on historical pricing data

## Decision Outcome

Chosen option: "**Multi-agent architecture with specialized, sequential agents**", because it provides the best balance of transparency, modularity, and real-time performance while allowing domain experts to understand and validate each decision step.

### Consequences

* **Good**, because merchants can see the reasoning of each agent (e.g., "Agent 1 found lowest competitor at $54", "Agent 4 accepted 6% margin to win")
* **Good**, because each agent can be modified independently (e.g., change margin validation logic without touching price intelligence)
* **Good**, because new capabilities can be added as new agents (e.g., Agent 5 for advertising bidding) without refactoring existing code
* **Good**, because domain experts can audit and validate agent behavior without understanding AI/ML internals
* **Bad**, because sequential processing adds latency vs. a single calculation (mitigated: total execution time <2 seconds)
* **Bad**, because increased code complexity with 4 separate agent implementations vs. 1 monolithic function

### Confirmation

Implementation confirmed through:
1. **Code review**: Each agent implements a single-responsibility interface (`PriceIntelligencer`, `MarketAnalyzer`, `StrategyRecommender`, `MarginValidator`)
2. **Integration tests**: End-to-end tests verify all 4 agents collaborate correctly
3. **User acceptance**: Dashboard displays reasoning from all 4 agents, validated by non-technical users
4. **Performance tests**: Full agent orchestration completes in <2 seconds for realistic competitive scenarios

## Pros and Cons of the Options

### Option 1: Monolithic Pricing Algorithm

Single function that inputs competitor prices and outputs recommended price.

* **Good**, because simple to implement and understand for small teams
* **Good**, because fastest execution (no inter-agent communication)
* **Bad**, because opaque decision-making ("why did it recommend $53?")
* **Bad**, because difficult to modify one aspect (e.g., margin calculation) without risking side effects
* **Bad**, because cannot explain reasoning to merchants

### Option 2: Rule-Based Expert System

Decision tree with if/then rules (e.g., "IF rank > 2 THEN apply aggressive strategy").

* **Good**, because human-readable rules
* **Good**, because easy to add new rules
* **Neutral**, because moderate transparency (can trace which rule fired)
* **Bad**, because rule conflicts as system grows (Rule 47 contradicts Rule 12)
* **Bad**, because brittle when adding new decision dimensions (discount codes, advertising bids)

### Option 3: Multi-Agent Architecture (CHOSEN)

Four specialized agents execute sequentially:
1. **Agent 1 (Price Intelligence)**: Gathers competitor prices, detects discount codes, calculates effective prices
2. **Agent 2 (Market Analysis)**: Analyzes market position, trends, opportunities
3. **Agent 3 (Strategy Recommender)**: Recommends pricing strategy based on business context
4. **Agent 4 (Margin Validator)**: Validates profitability constraints

* **Good**, because transparent (each agent explains its reasoning)
* **Good**, because modular (change Agent 4 without touching Agent 1)
* **Good**, because extensible (add Agent 5 for advertising without rewriting)
* **Good**, because testable (unit test each agent independently)
* **Good**, because auditable (domain experts can validate each agent's logic)
* **Neutral**, because sequential latency (mitigated: <2 second total)
* **Bad**, because more code to maintain (4 agent files vs 1 monolithic file)

### Option 4: Machine Learning Model

Train a neural network on historical pricing data to predict optimal price.

* **Good**, because can discover non-obvious pricing patterns
* **Good**, because improves over time with more data
* **Bad**, because black box (cannot explain why $53 is recommended)
* **Bad**, because requires large training dataset (not available for new merchants)
* **Bad**, because difficult to encode business constraints (never sell below cost)
* **Bad**, because slow to adapt to new competitor strategies (requires retraining)

## More Information

### Agent Implementation Details

Each agent implements a specific interface:

```go
// Agent 1
type PriceIntelligencer interface {
    Analyze(productID string, ourPrice int) (PriceIntelligence, error)
}

// Agent 2
type MarketAnalyzer interface {
    Analyze(intel PriceIntelligence) (MarketInsight, error)
}

// Agent 3
type StrategyRecommender interface {
    Recommend(intel PriceIntelligence, insight MarketInsight, context BusinessConfig) (PricingRecommendation, error)
}

// Agent 4
type MarginValidator interface {
    Validate(rec PricingRecommendation, ourPrice int) (ValidationResult, error)
}
```

### Key Innovation: Discount Code Detection

Agent 1 queries the Shopping Graph and extracts `discount_hints` field:

```json
{
  "merchant_id": "abc123",
  "price": 6000,
  "discount_hints": ["WELCOME10", "SAVE20"]
}
```

The agent parses these codes using heuristics:
- `WELCOME10` → 10% discount
- `SAVE20` → 20% discount
- `FIXED500` → $5 fixed discount

This allows calculating **effective price** (price after best discount), which is what smart buyer agents actually pay.

### Future Extensions

This architecture supports planned enhancements:
- **Agent 5 (Advertising Strategy)**: Recommend CPC bid adjustments when price alone cannot win
- **Agent 6 (Stock Optimizer)**: Recommend stock levels to create urgency (e.g., "Only 3 left!")
- **Agent 2 Enhancement**: Add time-series analysis for price trend prediction

### Re-evaluation Criteria

Re-evaluate this decision if:
1. Agent orchestration latency exceeds 5 seconds (user experience degrades)
2. Adding a new agent requires changes to >2 existing agents (modularity breaks down)
3. Merchants cannot understand agent reasoning (transparency fails)
4. 50%+ of pricing decisions are overridden manually by merchants (trust breaks down)

### References

- Source code: `pkg/merchant/competitive/orchestrator.go`
- Agent implementations: `pkg/merchant/competitive/agents/`
- Dashboard UI: `demo/cmd/arena/dashboard.go` (section "Intelligence de Prix")
- Test results: `DEMO_SCENARIOS.md`
