# Multi-Agent Competitive Pricing - Integration Guide

## Architecture Overview

The multi-agent competitive pricing system consists of 4 specialized agents coordinated by an orchestrator:

1. **Price Intelligence Agent** - Gathers competitor prices and calculates market statistics
2. **Market Analysis Agent** - Analyzes market conditions, trends, and opportunities
3. **Strategy Recommender Agent** - Recommends pricing strategy based on business context
4. **Margin Validator Agent** - Validates pricing decisions against margin constraints

## Integration Steps

### 1. Create the Shopping Graph Client

```go
import (
    "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive"
)

sgClient := competitive.NewShoppingGraphClient("http://localhost:9000")
```

### 2. Create the 4 Agents

```go
import (
    "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/agents"
    "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/history"
    "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
)

// Agent 1: Price Intelligence
priceIntel := agents.NewPriceIntelligenceAgent(sgClient, "merchant_a")

// Agent 2: Market Analysis (with history)
historyStore := history.NewInMemoryHistoryStore()
marketAnalyst := agents.NewMarketAnalysisAgent(historyStore)

// Agent 3: Strategy Recommender
businessConfig := models.BusinessConfig{
    Objective:      "volume",  // or "margin" or "balanced"
    StockThreshold: 20,        // Low stock threshold
    BrandPosition:  "mid",     // or "budget" or "premium"
    MinMargin:      10,        // Minimum margin %
    CostPercent:    60,        // Cost as % of price
}
strategyRec := agents.NewStrategyRecommenderAgent(businessConfig)

// Agent 4: Margin Validator
marginConfig := models.MarginConfig{
    MinMarginPercent: 10,   // 10% minimum margin
    CostPercent:      60,   // Cost is 60% of price
    HardFloor:        true, // Never sell below cost
}
marginVal := agents.NewMarginValidatorAgent(marginConfig)
```

### 3. Create the Orchestrator

```go
orchestrator := competitive.NewOrchestrator(
    priceIntel,
    marketAnalyst,
    strategyRec,
    marginVal,
)
```

### 4. Create the Discount Adapter

```go
import (
    "github.com/owulveryck/ucp-merchant-test/pkg/merchant/discount"
)

// Assuming you have static discount data
// var shopData discount.DiscountLookup

discountAdapter := competitive.NewDiscountAdapter(
    shopData,        // Fallback for normal discount codes
    orchestrator,    // For AUTO_COMPETE
    businessConfig,  // Business context
)
```

### 5. Integrate with Merchant

```go
import (
    "github.com/owulveryck/ucp-merchant-test/pkg/merchant"
)

// Create merchant with competitive pricing adapter
merchant := merchant.NewSimpleMerchant(
    catalog,
    shopData,
    discountAdapter,  // Uses multi-agent system for AUTO_COMPETE
    fulfillmentData,
    paymentProcessor,
)
```

## Usage

### Triggering Competitive Pricing

Use the special discount code `AUTO_COMPETE`:

```json
POST /checkout/:id
{
  "discount_codes": ["AUTO_COMPETE"]
}
```

### What Happens

1. **Price Intelligence Agent** queries Shopping Graph for competitor prices
2. **Market Analysis Agent** analyzes market position, trends, opportunities
3. **Strategy Recommender Agent** recommends strategy based on:
   - Stock level (low stock → aggressive)
   - Business objective (volume → aggressive, margin → premium)
   - Market conditions (price war → match, stable → balanced)
   - Brand position (premium → keep high, budget → aggressive)
4. **Margin Validator Agent** validates the recommended price:
   - Ensures minimum margin is met
   - Adjusts if necessary
   - Rejects if below cost

### Example Scenarios

#### Scenario 1: Normal Stock, Volume Objective

```
Context:
  - Our price: $89.99
  - Competitor: $84.99
  - Stock: 150 units (normal)
  - Objective: volume

Result:
  - Strategy: balanced
  - Target: $80.75 (beat by 5%)
  - Reasoning: "Standard competitive positioning"
```

#### Scenario 2: Low Stock, Need to Clear

```
Context:
  - Our price: $89.99
  - Competitor: $84.99
  - Stock: 15 units (LOW)
  - Objective: volume

Result:
  - Strategy: aggressive
  - Target: $76.50 (beat by 10%)
  - Reasoning: "Low stock (15 units) - clear inventory quickly"
```

#### Scenario 3: Already Leader, Margin Objective

```
Context:
  - Our price: $89.99
  - Competitor: $95.00 (we're cheaper!)
  - Objective: margin

Result:
  - Strategy: premium
  - Target: $88.19 (reduce by 2% max)
  - Reasoning: "Already market leader - maximize margin"
```

#### Scenario 4: Price War Detected

```
Context:
  - Our price: $89.99
  - Competitors: $84.99 → $83.50 → $82.00 (falling)
  - Trend: down -8%

Result:
  - Strategy: match
  - Target: $82.00 (match lowest)
  - Reasoning: "Price war detected - match market to stay competitive"
```

## Monitoring

The orchestrator logs detailed information at each step:

```
[Orchestrator] Starting competitive pricing analysis for product prod_123
[Orchestrator] Price Intelligence: rank 2/3, lowest: $84.99 (merchant_b)
[Orchestrator] Market Analysis: follower position, stable trend, opportunity: optimize
[Orchestrator] Strategy: balanced, target: $80.75, discount: $9.24, confidence: 80%
[Orchestrator] Reasoning: ["Standard competitive positioning"]
[Orchestrator] ✅ Pricing approved: $80.75 (discount: $9.24, margin: 25%)
```

## Configuration

### Business Context

Update business context dynamically:

```go
discountAdapter.UpdateConfig(models.BusinessConfig{
    Objective:      "margin",  // Changed from "volume"
    StockLevel:     50,
    StockThreshold: 20,
    BrandPosition:  "premium",
    MinMargin:      15,  // Increased minimum margin
    CostPercent:    60,
})
```

### Strategies

The Strategy Recommender uses 5 strategies:

- **aggressive**: Beat by 10% (low stock, price war)
- **balanced**: Beat by 5% (standard competition)
- **match**: Match exactly (defensive, price war)
- **premium**: Reduce minimally or keep (already leader, margin focus)
- **defensive**: Beat by 3% (rising market)

## Files Structure

```
pkg/merchant/competitive/
├── orchestrator.go           # Orchestrates the 4 agents
├── discount_adapter.go       # Adapts to discount.DiscountLookup
├── shoppinggraph.go          # Shopping Graph HTTP client
├── agents/
│   ├── price_intelligence.go    # Agent 1
│   ├── market_analysis.go       # Agent 2
│   ├── strategy_recommender.go  # Agent 3
│   └── margin_validator.go      # Agent 4
├── models/
│   ├── types.go             # Data types
│   └── interfaces.go        # Agent interfaces
└── history/
    └── store.go             # Price history storage
```

## Testing

To test the integration:

1. Start Shopping Graph: `./shopping-graph --port 9000`
2. Start merchant with AUTO_COMPETE enabled
3. Create checkout with products
4. Apply discount code: `AUTO_COMPETE`
5. Observe logs to see agent decisions

## Benefits Over Single-Agent Approach

| Aspect | Single Agent | Multi-Agent |
|--------|-------------|-------------|
| Context-aware | No | Yes (stock, objective, trend) |
| Adaptability | Fixed (always 5%) | Dynamic (0-10% based on context) |
| Explainability | None | Detailed reasoning |
| Testability | Difficult | Easy (test agents separately) |
| Extensibility | Hard | Easy (add new agents) |
| Intelligence | Rules-based | Context-driven decisions |
