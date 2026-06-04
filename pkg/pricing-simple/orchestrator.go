// Package pricing provides simplified 3-agent pricing system.
package pricing

import (
	"fmt"
	"log"

	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple/models"
)

// Orchestrator coordinates the 3-agent pricing system.
type Orchestrator struct {
	marketAgent   models.MarketIntelligenceAgent
	retentionAgent models.LoyaltyAgent
	decisionAgent models.FinalDecisionAgent
}

// AgentDecisions holds the complete decision trail from all 3 agents.
type AgentDecisions struct {
	Market    models.MarketIntelligenceDecision
	Retention models.LoyaltyDecision
	Final     models.FinalPricingDecision
}

// NewOrchestrator creates a new 3-agent orchestrator.
func NewOrchestrator(
	marketAgent models.MarketIntelligenceAgent,
	retentionAgent models.LoyaltyAgent,
	decisionAgent models.FinalDecisionAgent,
) *Orchestrator {
	return &Orchestrator{
		marketAgent:    marketAgent,
		retentionAgent: retentionAgent,
		decisionAgent:  decisionAgent,
	}
}

// CalculateOptimalPrice orchestrates the 3 agents to calculate optimal price.
//
// FLUX:
// 1. Agent 1 (Market Intelligence) + Agent 2 (Customer Retention) en parallèle
// 2. Agent 3 (Final Decision) synthétise les 2
func (o *Orchestrator) CalculateOptimalPrice(request models.PricingRequest) (models.FinalPricingDecision, *AgentDecisions, error) {
	log.Printf("[Orchestrator] Starting 3-agent pricing for customer=%s product=%s",
		request.CustomerID, request.ProductID)

	decisions := &AgentDecisions{}

	// PARALLEL: Agent 1 + Agent 2
	log.Printf("[Orchestrator] Step 1: Consulting Market Intelligence + Customer Retention in parallel")

	// Agent 1: Market Intelligence
	marketDecision, err := o.marketAgent.Analyze(request.ProductID, request.BasePrice)
	if err != nil {
		return models.FinalPricingDecision{}, nil, fmt.Errorf("market intelligence failed: %w", err)
	}
	decisions.Market = marketDecision

	log.Printf("[Orchestrator] Market: Position=%d/%d, Lowest=$%.2f, Gap=$%.2f",
		marketDecision.OurPosition,
		marketDecision.TotalCompetitors,
		float64(marketDecision.LowestCompetitor)/100,
		float64(marketDecision.CompetitiveGap)/100)

	// Agent 2: Customer Retention
	retentionDecision, err := o.retentionAgent.AnalyzeCustomer(request.CustomerID)
	if err != nil {
		return models.FinalPricingDecision{}, nil, fmt.Errorf("customer retention failed: %w", err)
	}
	decisions.Retention = retentionDecision

	log.Printf("[Orchestrator] Retention: VIP=%v, Tier=%s, Discount=%d%%",
		retentionDecision.IsVIP,
		retentionDecision.CustomerTier,
		retentionDecision.SuggestedDiscount)

	// SEQUENTIAL: Agent 3
	log.Printf("[Orchestrator] Step 2: Final Decision orchestration")

	finalDecision, err := o.decisionAgent.Decide(marketDecision, retentionDecision, request)
	if err != nil {
		return models.FinalPricingDecision{}, nil, fmt.Errorf("final decision failed: %w", err)
	}
	decisions.Final = finalDecision

	log.Printf("[Orchestrator] Final: Price=$%.2f, Discount=$%.2f (%d%%), Margin=%d%%, Strategy=%s",
		float64(finalDecision.FinalPrice)/100,
		float64(finalDecision.TotalDiscount)/100,
		finalDecision.DiscountPercent,
		finalDecision.Margin,
		finalDecision.Strategy)

	if len(finalDecision.Warnings) > 0 {
		log.Printf("[Orchestrator] Warnings: %v", finalDecision.Warnings)
	}

	return finalDecision, decisions, nil
}
