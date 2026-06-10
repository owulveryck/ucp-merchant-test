// Package competitive provides multi-agent competitive pricing.
package competitive

import (
	"fmt"
	"log"
	"time"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
)

// Orchestrator coordinates the multi-agent pricing system.
type Orchestrator struct {
	priceIntel    models.PriceIntelligencer
	marketAnalyst models.MarketAnalyzer
	strategyRec   models.StrategyRecommender
	marginVal     models.MarginValidator
}

// AgentDecisions holds the complete decision trail from all agents.
type AgentDecisions struct {
	Intel       models.PriceIntelligence
	Insight     models.MarketInsight
	Recommendation models.PricingRecommendation
	Validation  models.ValidationResult
}

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(
	priceIntel models.PriceIntelligencer,
	marketAnalyst models.MarketAnalyzer,
	strategyRec models.StrategyRecommender,
	marginVal models.MarginValidator,
) *Orchestrator {
	return &Orchestrator{
		priceIntel:    priceIntel,
		marketAnalyst: marketAnalyst,
		strategyRec:   strategyRec,
		marginVal:     marginVal,
	}
}

// CalculateDiscount orchestrates the 4 agents to calculate a competitive discount.
func (o *Orchestrator) CalculateDiscount(
	productID string,
	ourPrice int,
	context models.BusinessConfig,
) models.ValidationResult {

	log.Printf("[Orchestrator] Starting competitive pricing analysis for product %s", productID)

	// Agent 1: Price Intelligence
	intel, err := o.priceIntel.Analyze(productID, ourPrice)
	if err != nil {
		log.Printf("[Orchestrator] Price intelligence failed: %v", err)
		return models.ValidationResult{
			ProductID:       productID,
			Approved:        false,
			FinalPrice:      ourPrice,
			FinalDiscount:   0,
			Margin:          100,
			Warnings:        []string{fmt.Sprintf("Price intelligence failed: %v", err)},
			Rejected:        true,
			RejectionReason: "Could not analyze competitor prices",
		}
	}

	log.Printf("[Orchestrator] Price Intelligence: rank %d/%d, lowest: $%.2f (%s)",
		intel.OurRank, intel.TotalCount, float64(intel.LowestPrice)/100, intel.LowestBy)

	// Record current price for trend analysis
	_ = o.marketAnalyst.RecordPrice(productID, ourPrice, time.Now())

	// Agent 2: Market Analysis
	insight, err := o.marketAnalyst.Analyze(intel)
	if err != nil {
		log.Printf("[Orchestrator] Market analysis failed: %v", err)
		// Continue with empty insight
		insight = models.MarketInsight{
			ProductID:   productID,
			Position:    "unknown",
			Trend:       "stable",
			Opportunity: "optimize",
		}
	}

	log.Printf("[Orchestrator] Market Analysis: %s position, %s trend, opportunity: %s",
		insight.Position, insight.Trend, insight.Opportunity)

	// Agent 3: Strategy Recommender
	rec, err := o.strategyRec.Recommend(intel, insight, context)
	if err != nil {
		log.Printf("[Orchestrator] Strategy recommendation failed: %v", err)
		return models.ValidationResult{
			ProductID:       productID,
			Approved:        false,
			FinalPrice:      ourPrice,
			FinalDiscount:   0,
			Margin:          100,
			Warnings:        []string{fmt.Sprintf("Strategy recommendation failed: %v", err)},
			Rejected:        true,
			RejectionReason: "Could not determine pricing strategy",
		}
	}

	log.Printf("[Orchestrator] Strategy: %s, target: $%.2f, discount: $%.2f, confidence: %.0f%%",
		rec.Strategy, float64(rec.TargetPrice)/100, float64(rec.DiscountAmount)/100, rec.Confidence)
	log.Printf("[Orchestrator] Reasoning: %v", rec.Reasoning)

	// Agent 4: Margin Validator
	result, err := o.marginVal.Validate(rec, ourPrice)
	if err != nil {
		log.Printf("[Orchestrator] Margin validation failed: %v", err)
		return models.ValidationResult{
			ProductID:       productID,
			Approved:        false,
			FinalPrice:      ourPrice,
			FinalDiscount:   0,
			Margin:          100,
			Warnings:        []string{fmt.Sprintf("Margin validation failed: %v", err)},
			Rejected:        true,
			RejectionReason: "Could not validate margin constraints",
		}
	}

	if result.Rejected {
		log.Printf("[Orchestrator] ❌ Pricing REJECTED: %s", result.RejectionReason)
	} else if len(result.Warnings) > 0 {
		log.Printf("[Orchestrator] ⚠️  Pricing adjusted: %v", result.Warnings)
		log.Printf("[Orchestrator] ✅ Final: $%.2f (discount: $%.2f, margin: %d%%)",
			float64(result.FinalPrice)/100, float64(result.FinalDiscount)/100, result.Margin)
	} else {
		log.Printf("[Orchestrator] ✅ Pricing approved: $%.2f (discount: $%.2f, margin: %d%%)",
			float64(result.FinalPrice)/100, float64(result.FinalDiscount)/100, result.Margin)
	}

	return result
}

// CalculateDiscountWithTrace orchestrates the 4 agents and returns both the result and the full decision trail.
func (o *Orchestrator) CalculateDiscountWithTrace(
	productID string,
	ourPrice int,
	context models.BusinessConfig,
) (models.ValidationResult, *AgentDecisions) {

	decisions := &AgentDecisions{}

	log.Printf("[Orchestrator] Starting competitive pricing analysis for product %s", productID)

	// Agent 1: Price Intelligence
	intel, err := o.priceIntel.Analyze(productID, ourPrice)
	if err != nil {
		log.Printf("[Orchestrator] Price intelligence failed: %v", err)
		result := models.ValidationResult{
			ProductID:       productID,
			Approved:        false,
			FinalPrice:      ourPrice,
			FinalDiscount:   0,
			Margin:          100,
			Warnings:        []string{fmt.Sprintf("Price intelligence failed: %v", err)},
			Rejected:        true,
			RejectionReason: "Could not analyze competitor prices",
		}
		return result, decisions
	}
	decisions.Intel = intel

	log.Printf("[Orchestrator] Price Intelligence: rank %d/%d, lowest: $%.2f (%s)",
		intel.OurRank, intel.TotalCount, float64(intel.LowestPrice)/100, intel.LowestBy)

	// Record current price for trend analysis
	_ = o.marketAnalyst.RecordPrice(productID, ourPrice, time.Now())

	// Agent 2: Market Analysis
	insight, err := o.marketAnalyst.Analyze(intel)
	if err != nil {
		log.Printf("[Orchestrator] Market analysis failed: %v", err)
		// Continue with empty insight
		insight = models.MarketInsight{
			ProductID:   productID,
			Position:    "unknown",
			Trend:       "stable",
			Opportunity: "optimize",
		}
	}
	decisions.Insight = insight

	log.Printf("[Orchestrator] Market Analysis: %s position, %s trend, opportunity: %s",
		insight.Position, insight.Trend, insight.Opportunity)

	// Agent 3: Strategy Recommender
	rec, err := o.strategyRec.Recommend(intel, insight, context)
	if err != nil {
		log.Printf("[Orchestrator] Strategy recommendation failed: %v", err)
		result := models.ValidationResult{
			ProductID:       productID,
			Approved:        false,
			FinalPrice:      ourPrice,
			FinalDiscount:   0,
			Margin:          100,
			Warnings:        []string{fmt.Sprintf("Strategy recommendation failed: %v", err)},
			Rejected:        true,
			RejectionReason: "Could not determine pricing strategy",
		}
		return result, decisions
	}
	decisions.Recommendation = rec

	log.Printf("[Orchestrator] Strategy: %s, target: $%.2f, discount: $%.2f, confidence: %.0f%%",
		rec.Strategy, float64(rec.TargetPrice)/100, float64(rec.DiscountAmount)/100, rec.Confidence)
	log.Printf("[Orchestrator] Reasoning: %v", rec.Reasoning)

	// Agent 4: Margin Validator
	result, err := o.marginVal.Validate(rec, ourPrice)
	if err != nil {
		log.Printf("[Orchestrator] Margin validation failed: %v", err)
		result = models.ValidationResult{
			ProductID:       productID,
			Approved:        false,
			FinalPrice:      ourPrice,
			FinalDiscount:   0,
			Margin:          100,
			Warnings:        []string{fmt.Sprintf("Margin validation failed: %v", err)},
			Rejected:        true,
			RejectionReason: "Could not validate margin constraints",
		}
		return result, decisions
	}
	decisions.Validation = result

	if result.Rejected {
		log.Printf("[Orchestrator] ❌ Pricing REJECTED: %s", result.RejectionReason)
	} else if len(result.Warnings) > 0 {
		log.Printf("[Orchestrator] ⚠️  Pricing adjusted: %v", result.Warnings)
		log.Printf("[Orchestrator] ✅ Final: $%.2f (discount: $%.2f, margin: %d%%)",
			float64(result.FinalPrice)/100, float64(result.FinalDiscount)/100, result.Margin)
	} else {
		log.Printf("[Orchestrator] ✅ Pricing approved: $%.2f (discount: $%.2f, margin: %d%%)",
			float64(result.FinalPrice)/100, float64(result.FinalDiscount)/100, result.Margin)
	}

	return result, decisions
}
