package agents

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
)

// StrategyRecommenderAgent recommends pricing strategies based on business context.
type StrategyRecommenderAgent struct {
	config models.BusinessConfig
}

// NewStrategyRecommenderAgent creates a new strategy recommender agent.
func NewStrategyRecommenderAgent(config models.BusinessConfig) *StrategyRecommenderAgent {
	return &StrategyRecommenderAgent{
		config: config,
	}
}

// Recommend recommends a pricing strategy.
func (a *StrategyRecommenderAgent) Recommend(
	intel models.PriceIntelligence,
	insight models.MarketInsight,
	context models.BusinessConfig,
) (models.PricingRecommendation, error) {

	reasons := []string{}

	// Already cheapest? Consider premium strategy
	if intel.OurRank == 1 {
		return a.premiumStrategy(intel, insight, context, reasons)
	}

	// Low stock? Need to sell fast
	if context.StockLevel > 0 && context.StockLevel < context.StockThreshold {
		reasons = append(reasons, fmt.Sprintf("Low stock (%d units) - clear inventory quickly", context.StockLevel))
		return a.aggressiveStrategy(intel, insight, context, reasons)
	}

	// Price war detected? Match to stay competitive
	if insight.Opportunity == "price_war" {
		reasons = append(reasons, "Price war detected - match market to stay competitive")
		return a.matchStrategy(intel, insight, context, reasons)
	}

	// Volume objective + not competitive? Be aggressive
	if context.Objective == "volume" && intel.OurRank > 2 {
		reasons = append(reasons, "Volume objective - must be competitive")
		return a.aggressiveStrategy(intel, insight, context, reasons)
	}

	// Margin objective + already competitive? Keep premium
	if context.Objective == "margin" && intel.OurRank <= 2 {
		reasons = append(reasons, "Margin objective + already competitive - maintain position")
		return a.premiumStrategy(intel, insight, context, reasons)
	}

	// Premium brand + market allows it?
	if context.BrandPosition == "premium" && insight.Opportunity == "premium_position" {
		reasons = append(reasons, "Premium brand positioning + market spread allows it")
		return a.premiumStrategy(intel, insight, context, reasons)
	}

	// Rising market? Be defensive
	if insight.Opportunity == "rising_market" {
		reasons = append(reasons, "Rising market - moderate adjustment")
		return a.defensiveStrategy(intel, insight, context, reasons)
	}

	// Default: balanced approach
	reasons = append(reasons, "Standard competitive positioning")
	return a.balancedStrategy(intel, insight, context, reasons)
}

// aggressiveStrategy beats competitor by significant margin to WIN.
func (a *StrategyRecommenderAgent) aggressiveStrategy(
	intel models.PriceIntelligence,
	insight models.MarketInsight,
	context models.BusinessConfig,
	reasons []string,
) (models.PricingRecommendation, error) {

	// Beat lowest competitor by $1 minimum (100 cents)
	// This ensures we're ALWAYS cheaper than competition
	beatAmount := 100 // $1 in cents
	targetPrice := intel.LowestPrice - beatAmount

	// Make sure we beat by at least $1, even if that's > 10%
	minBeatPercent := intel.LowestPrice * 90 / 100 // 10% cheaper
	if targetPrice > minBeatPercent {
		targetPrice = minBeatPercent
	}

	discount := intel.OurPrice - targetPrice

	// Calculate expected margin (will be validated by Agent 4)
	costPrice := intel.OurPrice * context.CostPercent / 100
	expectedMargin := 0
	if targetPrice > 0 {
		expectedMargin = (targetPrice - costPrice) * 100 / targetPrice
	}

	reasons = append(reasons, fmt.Sprintf("Beat lowest competitor by $%.2f to guarantee win", float64(beatAmount)/100))

	return models.PricingRecommendation{
		ProductID:      intel.ProductID,
		Strategy:       "aggressive",
		TargetPrice:    targetPrice,
		DiscountAmount: discount,
		Confidence:     95, // Higher confidence - we WILL win
		Reasoning:      reasons,
		ExpectedMargin: expectedMargin,
	}, nil
}

// balancedStrategy beats competitor by 5%.
func (a *StrategyRecommenderAgent) balancedStrategy(
	intel models.PriceIntelligence,
	insight models.MarketInsight,
	context models.BusinessConfig,
	reasons []string,
) (models.PricingRecommendation, error) {

	// Beat lowest by 5%
	targetPrice := intel.LowestPrice * 95 / 100
	discount := intel.OurPrice - targetPrice

	costPrice := intel.OurPrice * context.CostPercent / 100
	expectedMargin := 0
	if targetPrice > 0 {
		expectedMargin = (targetPrice - costPrice) * 100 / targetPrice
	}

	return models.PricingRecommendation{
		ProductID:      intel.ProductID,
		Strategy:       "balanced",
		TargetPrice:    targetPrice,
		DiscountAmount: discount,
		Confidence:     80,
		Reasoning:      reasons,
		ExpectedMargin: expectedMargin,
	}, nil
}

// matchStrategy matches the lowest competitor price.
func (a *StrategyRecommenderAgent) matchStrategy(
	intel models.PriceIntelligence,
	insight models.MarketInsight,
	context models.BusinessConfig,
	reasons []string,
) (models.PricingRecommendation, error) {

	targetPrice := intel.LowestPrice
	discount := intel.OurPrice - targetPrice

	costPrice := intel.OurPrice * context.CostPercent / 100
	expectedMargin := 0
	if targetPrice > 0 {
		expectedMargin = (targetPrice - costPrice) * 100 / targetPrice
	}

	return models.PricingRecommendation{
		ProductID:      intel.ProductID,
		Strategy:       "match",
		TargetPrice:    targetPrice,
		DiscountAmount: discount,
		Confidence:     90,
		Reasoning:      reasons,
		ExpectedMargin: expectedMargin,
	}, nil
}

// premiumStrategy keeps current price or reduces minimally.
func (a *StrategyRecommenderAgent) premiumStrategy(
	intel models.PriceIntelligence,
	insight models.MarketInsight,
	context models.BusinessConfig,
	reasons []string,
) (models.PricingRecommendation, error) {

	// Keep current price or reduce by 2% max
	targetPrice := intel.OurPrice * 98 / 100
	discount := intel.OurPrice - targetPrice

	costPrice := intel.OurPrice * context.CostPercent / 100
	expectedMargin := 0
	if targetPrice > 0 {
		expectedMargin = (targetPrice - costPrice) * 100 / targetPrice
	}

	return models.PricingRecommendation{
		ProductID:      intel.ProductID,
		Strategy:       "premium",
		TargetPrice:    targetPrice,
		DiscountAmount: discount,
		Confidence:     70,
		Reasoning:      reasons,
		ExpectedMargin: expectedMargin,
	}, nil
}

// defensiveStrategy makes moderate adjustment.
func (a *StrategyRecommenderAgent) defensiveStrategy(
	intel models.PriceIntelligence,
	insight models.MarketInsight,
	context models.BusinessConfig,
	reasons []string,
) (models.PricingRecommendation, error) {

	// Beat by 3%
	targetPrice := intel.LowestPrice * 97 / 100
	discount := intel.OurPrice - targetPrice

	costPrice := intel.OurPrice * context.CostPercent / 100
	expectedMargin := 0
	if targetPrice > 0 {
		expectedMargin = (targetPrice - costPrice) * 100 / targetPrice
	}

	return models.PricingRecommendation{
		ProductID:      intel.ProductID,
		Strategy:       "defensive",
		TargetPrice:    targetPrice,
		DiscountAmount: discount,
		Confidence:     75,
		Reasoning:      reasons,
		ExpectedMargin: expectedMargin,
	}, nil
}
