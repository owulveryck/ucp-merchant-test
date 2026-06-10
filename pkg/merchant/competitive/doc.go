// Package competitive provides dynamic competitive pricing capabilities for UCP merchants.
//
// The competitive package enables merchants to automatically adjust their discount codes
// in real-time based on competitor pricing. It integrates with the Shopping Graph to
// discover competitor prices and applies configurable pricing strategies to remain competitive
// while maintaining minimum profit margins.
//
// # Architecture
//
// The package implements the discount.DiscountLookup interface, allowing it to be injected
// into the merchant implementation as a drop-in replacement for static discount lookups.
//
// Key components:
//   - CompetitivePricingAgent: Core agent that calculates dynamic discounts
//   - CompetitorPriceSource: Interface for querying competitor prices
//   - ShoppingGraphClient: HTTP client implementing CompetitorPriceSource
//   - PricingStrategy: Configurable strategies (match, beat, auto)
//
// # Usage
//
// Create a competitive pricing agent and inject it into the merchant:
//
//	sgClient := competitive.NewShoppingGraphClient("http://localhost:9000")
//	agent := competitive.NewCompetitivePricingAgent(
//	    baseDiscountLookup,
//	    sgClient,
//	    competitive.StrategyBeatPrice,
//	    10, // 10% minimum margin
//	)
//	merchant := newSimpleMerchant(catalog, shopData, agent, ...)
//
// # Pricing Strategies
//
// - StrategyMatchPrice: Match the lowest competitor price exactly
// - StrategyBeatPrice: Beat competitor by a percentage (default 5%)
// - StrategyAutoDiscount: Automatically generate discount to undercut competition
//
// # Special Discount Codes
//
// The agent recognizes special discount codes that trigger competitive pricing:
//   - "AUTO_COMPETE": Calculate optimal discount to beat all competitors
//   - "COMP_*": Reserved prefix for future competitor-specific codes
//
// # Safety Features
//
// - Minimum margin validation: Never discount below configured margin
// - Timeout protection: Fall back to static codes if Shopping Graph is slow
// - Cache support: Optional caching of competitor prices (5-10s TTL)
// - Error handling: Graceful degradation when competitor data unavailable
//
// # Integration Points
//
// The agent integrates at checkout update time:
//  1. Client submits checkout with "AUTO_COMPETE" discount code
//  2. Agent queries Shopping Graph for competitor prices of line items
//  3. Agent calculates minimum discount needed to beat competition
//  4. Agent validates margin requirements
//  5. Agent returns dynamic discount with calculated amount
//  6. Standard pricing.CalculateTotals() applies the discount
//
package competitive
