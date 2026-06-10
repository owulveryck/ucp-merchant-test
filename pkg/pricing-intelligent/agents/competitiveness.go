// Package agents contains specialized pricing agents.
package agents

import (
	"fmt"
	"sort"

	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-intelligent/models"
)

// CompetitivenessAgentImpl implements the market competitiveness analysis agent.
type CompetitivenessAgentImpl struct {
	competitorData models.CompetitorDataSource
}

// NewCompetitivenessAgent creates a new competitiveness agent.
func NewCompetitivenessAgent(competitorData models.CompetitorDataSource) *CompetitivenessAgentImpl {
	return &CompetitivenessAgentImpl{
		competitorData: competitorData,
	}
}

// AnalyzeMarket analyzes competitor prices and determines competitive pricing.
//
// INTENTION: "Sommes-nous compétitifs sur ce produit ?"
//
// ANALYSE:
// - Prix des concurrents
// - Codes promo actifs (future)
// - Notre position marché
//
// DÉCISION:
// - Prix pour gagner : $XX.XX
// - Position actuelle : Xème/Y
func (a *CompetitivenessAgentImpl) AnalyzeMarket(productID string, basePrice int) (models.CompetitivenessDecision, error) {
	// Get competitor prices
	prices, err := a.competitorData.GetCompetitorPrices(productID)
	if err != nil {
		return models.CompetitivenessDecision{}, fmt.Errorf("failed to get competitor prices: %w", err)
	}

	decision := models.CompetitivenessDecision{
		ProductID: productID,
		Reasoning: []string{},
	}

	// No competitors - we set the market price
	if len(prices) == 0 {
		decision.LowestCompetitor = basePrice
		decision.OurPosition = 1
		decision.TotalCompetitors = 1
		decision.SuggestedPrice = basePrice
		decision.Confidence = 100
		decision.Reasoning = append(decision.Reasoning,
			"Aucun concurrent : vous définissez le prix du marché")
		return decision, nil
	}

	// Sort prices to find lowest
	allPrices := append([]int{basePrice}, prices...)
	sort.Ints(allPrices)

	lowest := allPrices[0]
	decision.LowestCompetitor = lowest
	decision.TotalCompetitors = len(allPrices)

	// Find our position
	ourPosition := 1
	for i, price := range allPrices {
		if price == basePrice {
			ourPosition = i + 1
			break
		}
	}
	decision.OurPosition = ourPosition

	// Analysis
	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Marché : %d concurrents analysés", len(prices)))
	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Prix concurrent le plus bas : $%.2f", float64(lowest)/100))
	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Notre position actuelle : %d/%d", ourPosition, decision.TotalCompetitors))

	// Decision: calculate competitive price
	if ourPosition == 1 {
		// Already cheapest - keep price or increase slightly
		decision.SuggestedPrice = basePrice
		decision.Confidence = 100
		decision.Reasoning = append(decision.Reasoning,
			"✓ Vous êtes déjà le moins cher : maintenir le prix")
	} else {
		// Not cheapest - beat lowest by $1
		beatAmount := 100 // $1 in cents
		decision.SuggestedPrice = lowest - beatAmount

		decision.Confidence = 90
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("📊 Prix suggéré pour gagner : $%.2f (battre concurrent de $1)",
				float64(decision.SuggestedPrice)/100))
	}

	return decision, nil
}
