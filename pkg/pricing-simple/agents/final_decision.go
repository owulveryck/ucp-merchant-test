// Package agents contains the 3 specialized pricing agents.
package agents

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple/models"
)

// FinalDecisionAgentImpl orchestrates market + retention decisions (Agent 3).
type FinalDecisionAgentImpl struct {
	minMarginPercent int
	hardFloor        bool
}

// NewFinalDecisionAgent creates Agent 3.
func NewFinalDecisionAgent(minMarginPercent int, hardFloor bool) *FinalDecisionAgentImpl {
	return &FinalDecisionAgentImpl{
		minMarginPercent: minMarginPercent,
		hardFloor:        hardFloor,
	}
}

// Decide orchestrates the final pricing decision.
//
// INTENTION: "Quel prix optimiser pour maximiser profit ET vente ?"
//
// LOGIQUE:
// 1. Prix de base = battre concurrent le plus bas de $1
// 2. Appliquer bonus VIP si applicable
// 3. Valider marge minimum (ne jamais vendre à perte)
func (a *FinalDecisionAgentImpl) Decide(
	marketDecision models.MarketIntelligenceDecision,
	retentionDecision models.LoyaltyDecision,
	request models.PricingRequest,
) (models.FinalPricingDecision, error) {

	decision := models.FinalPricingDecision{
		ProductID:     request.ProductID,
		CustomerID:    request.CustomerID,
		OriginalPrice: request.BasePrice,
		Reasoning:     []string{},
	}

	decision.Reasoning = append(decision.Reasoning, "=== SYNTHÈSE DES AGENTS ===")

	// Step 1: Determine base competitive price
	var baseCompetitivePrice int
	var strategy string

	if marketDecision.OurPosition == 1 {
		// Already cheapest - keep current price
		baseCompetitivePrice = request.BasePrice
		strategy = "market_leader"
		decision.Reasoning = append(decision.Reasoning,
			"Stratégie : Leader de marché (déjà le moins cher)")
	} else {
		// Beat lowest competitor by $1
		baseCompetitivePrice = marketDecision.LowestCompetitor - 100
		strategy = "market_match"
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Stratégie : Battre concurrent ($%.2f - $1.00 = $%.2f)",
				float64(marketDecision.LowestCompetitor)/100,
				float64(baseCompetitivePrice)/100))
	}

	// Step 2: Apply VIP discount if applicable
	finalPrice := baseCompetitivePrice
	vipDiscount := 0

	if retentionDecision.IsVIP && retentionDecision.SuggestedDiscount > 0 {
		vipDiscount = baseCompetitivePrice * retentionDecision.SuggestedDiscount / 100
		finalPrice = baseCompetitivePrice - vipDiscount

		decision.IsVIP = true
		strategy = "vip_priority"

		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Client %s : réduction VIP %d%% (-$%.2f)",
				retentionDecision.CustomerTier,
				retentionDecision.SuggestedDiscount,
				float64(vipDiscount)/100))
	} else {
		decision.IsVIP = false
	}

	// Step 3: Validate margin constraints
	costPrice := request.CostPrice
	margin := 0
	if finalPrice > 0 {
		margin = (finalPrice - costPrice) * 100 / finalPrice
	}

	warnings := []string{}

	// Hard floor: never sell below cost
	if a.hardFloor && finalPrice < costPrice {
		// Adjust to minimum viable price (cost + min margin)
		minViablePrice := costPrice * (100 + a.minMarginPercent) / 100

		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("⚠ ALERTE : Prix $%.2f < coût $%.2f",
				float64(finalPrice)/100,
				float64(costPrice)/100))

		finalPrice = minViablePrice
		margin = a.minMarginPercent
		strategy = "minimum_viable"

		warnings = append(warnings,
			fmt.Sprintf("Prix ajusté au minimum viable : $%.2f (coût + %d%% marge)",
				float64(minViablePrice)/100, a.minMarginPercent))
		warnings = append(warnings, "Impossible de concurrencer sans vendre à perte")

		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Prix ajusté : $%.2f (coût + %d%%)",
				float64(finalPrice)/100, a.minMarginPercent))
	} else if margin < a.minMarginPercent {
		// Margin below target but above cost - accept to WIN
		warnings = append(warnings,
			fmt.Sprintf("Marge réduite : %d%% (cible : %d%%) pour GAGNER", margin, a.minMarginPercent))

		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("✓ Marge réduite acceptée : %d%% pour gagner la vente", margin))
	} else {
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("✓ Marge acceptable : %d%%", margin))
	}

	// Calculate totals
	totalDiscount := request.BasePrice - finalPrice
	discountPercent := 0
	if request.BasePrice > 0 {
		discountPercent = totalDiscount * 100 / request.BasePrice
	}

	// Set final decision
	decision.FinalPrice = finalPrice
	decision.TotalDiscount = totalDiscount
	decision.DiscountPercent = discountPercent
	decision.Margin = margin
	decision.Strategy = strategy
	decision.Warnings = warnings
	decision.Approved = true // Always approved (we adjust if needed)
	decision.IsCompetitive = marketDecision.OurPosition <= 2 || finalPrice <= marketDecision.LowestCompetitor

	// Final summary
	decision.Reasoning = append(decision.Reasoning, "")
	decision.Reasoning = append(decision.Reasoning, "=== DÉCISION FINALE ===")
	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Prix de base : $%.2f", float64(request.BasePrice)/100))
	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Prix final : $%.2f", float64(finalPrice)/100))
	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Réduction totale : $%.2f (%d%%)",
			float64(totalDiscount)/100, discountPercent))
	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Marge finale : %d%%", margin))
	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Stratégie appliquée : %s", strategy))

	return decision, nil
}
