// Package pricing provides intelligent multi-agent pricing system.
package pricing

import (
	"fmt"
	"log"

	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-intelligent/models"
)

// Orchestrator coordinates loyalty and competitiveness agents.
// Acts as "Agent Vendeur" - the central brain.
type Orchestrator struct {
	loyaltyAgent         models.LoyaltyAgent
	competitivenessAgent models.CompetitivenessAgent
	minMarginPercent     int // Minimum acceptable margin (%)
}

// NewOrchestrator creates a new pricing orchestrator.
func NewOrchestrator(
	loyaltyAgent models.LoyaltyAgent,
	competitivenessAgent models.CompetitivenessAgent,
	minMarginPercent int,
) *Orchestrator {
	return &Orchestrator{
		loyaltyAgent:         loyaltyAgent,
		competitivenessAgent: competitivenessAgent,
		minMarginPercent:     minMarginPercent,
	}
}

// CalculateOptimalPrice orchestrates agents to calculate optimal price.
//
// INTENTION: "Quel prix optimiser pour maximiser profit et vente ?"
//
// ORCHESTRATION:
// 1. Agent Fidélité : Ce client mérite-t-il un prix préférentiel ?
// 2. Agent Compétitivité : Sommes-nous compétitifs sur ce produit ?
// 3. Décision finale : Synthèse des 2 agents
//
// DÉCISION:
// - Prix final calculé selon valeur client + position marché
func (o *Orchestrator) CalculateOptimalPrice(request models.PricingRequest) (models.PricingResult, error) {
	log.Printf("[Orchestrator] Starting price calculation for customer=%s product=%s",
		request.CustomerID, request.ProductID)

	result := models.PricingResult{
		ProductID:  request.ProductID,
		CustomerID: request.CustomerID,
		BasePrice:  request.BasePrice,
		Reasoning:  []string{},
	}

	// Step 1: Consult Agent Fidélité
	log.Printf("[Orchestrator] Step 1: Consulting loyalty agent for customer %s", request.CustomerID)
	loyaltyDecision, err := o.loyaltyAgent.AnalyzeCustomer(request.CustomerID)
	if err != nil {
		return result, fmt.Errorf("loyalty agent failed: %w", err)
	}
	result.LoyaltyDecision = loyaltyDecision
	result.IsVIP = loyaltyDecision.IsVIP

	log.Printf("[Orchestrator] Loyalty: VIP=%v, Discount=%d%%",
		loyaltyDecision.IsVIP, loyaltyDecision.SuggestedDiscount)

	// Step 2: Consult Agent Compétitivité
	log.Printf("[Orchestrator] Step 2: Consulting competitiveness agent for product %s", request.ProductID)
	compDecision, err := o.competitivenessAgent.AnalyzeMarket(request.ProductID, request.BasePrice)
	if err != nil {
		return result, fmt.Errorf("competitiveness agent failed: %w", err)
	}
	result.CompetitivenessDecision = compDecision

	log.Printf("[Orchestrator] Competitiveness: Position=%d/%d, Suggested=$%.2f",
		compDecision.OurPosition, compDecision.TotalCompetitors, float64(compDecision.SuggestedPrice)/100)

	// Step 3: Orchestrated decision
	log.Printf("[Orchestrator] Step 3: Making final orchestrated decision")

	// Start with competitive price
	finalPrice := compDecision.SuggestedPrice

	// Apply VIP discount if applicable
	if loyaltyDecision.IsVIP && loyaltyDecision.SuggestedDiscount > 0 {
		vipDiscount := finalPrice * loyaltyDecision.SuggestedDiscount / 100
		finalPrice = finalPrice - vipDiscount

		result.Reasoning = append(result.Reasoning,
			fmt.Sprintf("💎 Client VIP : réduction additionnelle de %d%% ($%.2f)",
				loyaltyDecision.SuggestedDiscount, float64(vipDiscount)/100))
	}

	// Validate margin
	margin := 0
	if finalPrice > 0 {
		margin = (finalPrice - request.CostPrice) * 100 / finalPrice
	}

	// Hard floor: never sell below cost
	if finalPrice < request.CostPrice {
		result.Reasoning = append(result.Reasoning,
			fmt.Sprintf("❌ Prix suggéré $%.2f < coût $%.2f : REJETÉ",
				float64(finalPrice)/100, float64(request.CostPrice)/100))

		// Fallback: minimum viable price = cost + min margin
		finalPrice = request.CostPrice * (100 + o.minMarginPercent) / 100
		margin = o.minMarginPercent

		result.Reasoning = append(result.Reasoning,
			fmt.Sprintf("Prix ajusté au minimum viable : $%.2f (marge %d%%)",
				float64(finalPrice)/100, margin))
	} else if margin < o.minMarginPercent {
		result.Reasoning = append(result.Reasoning,
			fmt.Sprintf("⚠ Marge réduite %d%% < cible %d%% pour GAGNER",
				margin, o.minMarginPercent))
	} else {
		result.Reasoning = append(result.Reasoning,
			fmt.Sprintf("✓ Marge acceptable : %d%%", margin))
	}

	// Calculate discount
	discount := request.BasePrice - finalPrice
	discountPercent := 0
	if request.BasePrice > 0 {
		discountPercent = discount * 100 / request.BasePrice
	}

	// Set final result
	result.FinalPrice = finalPrice
	result.Discount = discount
	result.DiscountPercent = discountPercent
	result.Margin = margin
	result.IsCompetitive = compDecision.OurPosition <= 2

	// Add agent reasoning to final reasoning
	result.Reasoning = append([]string{
		fmt.Sprintf("=== AGENT FIDÉLITÉ ==="),
	}, loyaltyDecision.Reasoning...)

	result.Reasoning = append(result.Reasoning, fmt.Sprintf("\n=== AGENT COMPÉTITIVITÉ ==="))
	result.Reasoning = append(result.Reasoning, compDecision.Reasoning...)

	result.Reasoning = append(result.Reasoning, fmt.Sprintf("\n=== DÉCISION FINALE ==="))
	result.Reasoning = append(result.Reasoning,
		fmt.Sprintf("Prix de base : $%.2f", float64(request.BasePrice)/100))
	result.Reasoning = append(result.Reasoning,
		fmt.Sprintf("Prix final : $%.2f", float64(finalPrice)/100))
	result.Reasoning = append(result.Reasoning,
		fmt.Sprintf("Réduction : $%.2f (%d%%)", float64(discount)/100, discountPercent))
	result.Reasoning = append(result.Reasoning,
		fmt.Sprintf("Marge : %d%%", margin))

	log.Printf("[Orchestrator] Final decision: $%.2f (discount=%d%%, margin=%d%%)",
		float64(finalPrice)/100, discountPercent, margin)

	return result, nil
}
