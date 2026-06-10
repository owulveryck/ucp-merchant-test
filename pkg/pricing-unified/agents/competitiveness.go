// Package agents contains the specialized agents for the unified pricing system.
package agents

import (
	"fmt"
	"log"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive"
	compModels "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/models"
)

// CompletivenessAgent is Agent 3: Analyzes market competitiveness.
// It wraps the existing 4-agent system for backwards compatibility.
type CompletivenessAgent struct {
	orchestrator   *competitive.Orchestrator
	merchantID     string
	costPrice      int
	businessConfig compModels.BusinessConfig
}

// NewCompletivenessAgent creates Agent 3.
func NewCompletivenessAgent(
	orchestrator *competitive.Orchestrator,
	merchantID string,
	costPrice int,
	businessConfig compModels.BusinessConfig,
) *CompletivenessAgent {
	return &CompletivenessAgent{
		orchestrator:   orchestrator,
		merchantID:     merchantID,
		costPrice:      costPrice,
		businessConfig: businessConfig,
	}
}

// Analyze answers the question: "Est-ce que je suis compétitif sur cet item ?"
//
// INTENTION: Déterminer si notre prix est compétitif et quel prix recommander
// DÉCISION: Prix compétitif + stratégie à adopter
//
// This agent delegates to the existing 4-agent system:
// - Agent 1: Price Intelligence (gets competitor prices)
// - Agent 2: Market Analysis (analyzes market position)
// - Agent 3: Strategy Recommender (recommends pricing strategy)
// - Agent 4: Margin Validator (validates margins)
func (a *CompletivenessAgent) Analyze(productID string, basePrice int) (models.CompetitivenessDecision, error) {
	log.Printf("[Agent Compétitivité] Analyzing competitiveness for product: %s at $%.2f",
		productID, float64(basePrice)/100)

	// Delegate to the existing 4-agent orchestrator
	validationResult, agentDecisions := a.orchestrator.CalculateDiscountWithTrace(
		productID,
		basePrice,
		a.businessConfig,
	)

	// Map the 4-agent results to our simplified decision
	decision := models.CompetitivenessDecision{
		CompetitiveReasoning: []string{},
	}

	// Extract intelligence from Agent 1 (Price Intelligence)
	intel := agentDecisions.Intel
	decision.MarketPosition = intel.OurRank
	decision.TotalCompetitors = intel.TotalCount
	decision.LowestCompetitor = intel.LowestPrice

	// Determine if we're competitive
	decision.IsCompetitive = (intel.OurRank <= 2) // Top 2 = competitive

	decision.CompetitiveReasoning = append(decision.CompetitiveReasoning,
		fmt.Sprintf("Agent 1 (Price Intelligence): Position %d/%d, Concurrent le moins cher: $%.2f",
			intel.OurRank, intel.TotalCount, float64(intel.LowestPrice)/100))

	// Extract market insight from Agent 2 (Market Analysis)
	insight := agentDecisions.Insight
	if len(insight.Reasoning) > 0 {
		decision.CompetitiveReasoning = append(decision.CompetitiveReasoning,
			fmt.Sprintf("Agent 2 (Market Analysis): %s", insight.Reasoning[0]))
	}

	// Extract strategy from Agent 3 (Strategy Recommender)
	rec := agentDecisions.Recommendation
	decision.Strategy = rec.Strategy
	if len(rec.Reasoning) > 0 {
		decision.CompetitiveReasoning = append(decision.CompetitiveReasoning,
			fmt.Sprintf("Agent 3 (Strategy): %s", rec.Reasoning[0]))
	}

	// Extract validation from Agent 4 (Margin Validator)
	val := agentDecisions.Validation
	decision.RecommendedPrice = val.FinalPrice
	decision.Margin = val.Margin

	if val.Approved {
		decision.CompetitiveReasoning = append(decision.CompetitiveReasoning,
			"Agent 4 (Margin Validator): Approuvé")
	} else if val.Rejected {
		decision.CompetitiveReasoning = append(decision.CompetitiveReasoning,
			fmt.Sprintf("Agent 4 (Margin Validator): %s", val.RejectionReason))
	}

	// Also store the validation result for reference
	_ = validationResult

	log.Printf("[Agent Compétitivité] Decision: IsCompetitive=%v, Position=%d/%d, Price=$%.2f",
		decision.IsCompetitive, decision.MarketPosition, decision.TotalCompetitors,
		float64(decision.RecommendedPrice)/100)

	return decision, nil
}
