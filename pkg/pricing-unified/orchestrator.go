// Package pricing provides the unified multi-agent pricing orchestrator.
package pricing

import (
	"fmt"
	"log"

	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/agents"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/models"
)

// VendorOrchestrator is Agent 1: The main orchestrator (Agent Vendeur).
//
// FLUX:
// 1. Acheteur Lambda → Agent Vendeur (demande de prix)
// 2. Agent Vendeur → Agent Customer Growth (vérifier si client à garder)
// 3. Agent Vendeur → Agent Compétitivité (vérifier compétitivité)
// 4. Agent Customer Growth → Agent Vendeur (réponse OUI/NON)
// 5. Agent Compétitivité → Agent Vendeur (prix compétitif)
// 6. Agent Vendeur → Acheteur Lambda (prix final)
type VendorOrchestrator struct {
	customerGrowth  *agents.CustomerGrowthAgent
	competitiveness *agents.CompletivenessAgent
}

// NewVendorOrchestrator creates the main orchestrator (Agent Vendeur).
func NewVendorOrchestrator(
	customerGrowth *agents.CustomerGrowthAgent,
	competitiveness *agents.CompletivenessAgent,
) *VendorOrchestrator {
	return &VendorOrchestrator{
		customerGrowth:  customerGrowth,
		competitiveness: competitiveness,
	}
}

// DeterminePricing orchestrates the multi-agent pricing decision.
//
// AGENT 1 (VENDEUR) INTENTION:
// "Je suis acheteur lambda et je veux tel item - quel prix lui donner ?"
//
// DÉCISION PROCESS:
// 1. Consulter Agent 2 (Customer Growth): Client à garder ?
// 2. Consulter Agent 3 (Compétitivité): Prix compétitif ?
// 3. Synthétiser les réponses pour déterminer le prix final
func (o *VendorOrchestrator) DeterminePricing(request models.PricingRequest) (models.VendorDecision, error) {
	log.Printf("[Agent Vendeur] Demande de prix pour %s par client %s - Prix de base: $%.2f",
		request.ProductID, request.CustomerID, float64(request.BasePrice)/100)

	decision := models.VendorDecision{
		OriginalPrice:     request.BasePrice,
		DecisionReasoning: []string{},
	}

	// Step 1: Consulter Agent 2 (Customer Growth)
	log.Printf("[Agent Vendeur] → Consultation Agent 2 (Customer Growth)")
	customerDecision, err := o.customerGrowth.Analyze(request.CustomerID)
	if err != nil {
		return decision, fmt.Errorf("customer growth analysis failed: %w", err)
	}
	decision.CustomerGrowth = customerDecision

	decision.DecisionReasoning = append(decision.DecisionReasoning,
		fmt.Sprintf("Agent 2 (Customer Growth): %s - %s",
			map[bool]string{true: "OUI, garder ce client", false: "NON, client non prioritaire"}[customerDecision.ShouldRetain],
			customerDecision.CustomerTier))

	// Step 2: Consulter Agent 3 (Compétitivité)
	log.Printf("[Agent Vendeur] → Consultation Agent 3 (Compétitivité)")
	compDecision, err := o.competitiveness.Analyze(request.ProductID, request.BasePrice)
	if err != nil {
		return decision, fmt.Errorf("competitiveness analysis failed: %w", err)
	}
	decision.Competitiveness = compDecision

	decision.DecisionReasoning = append(decision.DecisionReasoning,
		fmt.Sprintf("Agent 3 (Compétitivité): Position %d/%d - Prix recommandé: $%.2f",
			compDecision.MarketPosition, compDecision.TotalCompetitors,
			float64(compDecision.RecommendedPrice)/100))

	// Step 3: SYNTHÈSE - Agent Vendeur décide du prix final
	log.Printf("[Agent Vendeur] → Synthèse des décisions")

	// Start with competitive price, or base price if no competitive data
	finalPrice := compDecision.RecommendedPrice
	if finalPrice == 0 {
		// No competitive data - use base price as starting point
		finalPrice = request.BasePrice
		decision.DecisionReasoning = append(decision.DecisionReasoning,
			"Pas de données de marché - utilisation du prix de base")
	}

	// Apply VIP discount if customer should be retained
	if customerDecision.ShouldRetain && customerDecision.SuggestedDiscount > 0 {
		vipDiscount := finalPrice * customerDecision.SuggestedDiscount / 100
		finalPrice = finalPrice - vipDiscount

		decision.DecisionReasoning = append(decision.DecisionReasoning,
			fmt.Sprintf("Bonus fidélité appliqué: -%d%% (client %s)",
				customerDecision.SuggestedDiscount, customerDecision.CustomerTier))

		// But never go below cost
		if finalPrice < request.CostPrice {
			minPrice := request.CostPrice * 110 / 100 // Cost + 10% minimum
			decision.DecisionReasoning = append(decision.DecisionReasoning,
				"⚠ Ajustement: prix ne peut être sous le coût")
			finalPrice = minPrice
		}
	}

	// Calculate final metrics
	decision.FinalPrice = finalPrice
	decision.TotalDiscount = request.BasePrice - finalPrice
	if request.BasePrice > 0 {
		decision.DiscountPercent = decision.TotalDiscount * 100 / request.BasePrice
	}
	if finalPrice > 0 {
		decision.Margin = (finalPrice - request.CostPrice) * 100 / finalPrice
	}

	// Determine strategy
	if customerDecision.ShouldRetain && customerDecision.CustomerTier != "standard" {
		decision.Strategy = "vip_retention"
	} else if compDecision.IsCompetitive {
		decision.Strategy = "competitive_pricing"
	} else {
		decision.Strategy = "market_alignment"
	}

	decision.DecisionReasoning = append(decision.DecisionReasoning, "")
	decision.DecisionReasoning = append(decision.DecisionReasoning, "=== DÉCISION VENDEUR ===")
	decision.DecisionReasoning = append(decision.DecisionReasoning,
		fmt.Sprintf("Prix final: $%.2f (-%d%%)", float64(finalPrice)/100, decision.DiscountPercent))
	decision.DecisionReasoning = append(decision.DecisionReasoning,
		fmt.Sprintf("Marge: %d%%", decision.Margin))
	decision.DecisionReasoning = append(decision.DecisionReasoning,
		fmt.Sprintf("Stratégie: %s", decision.Strategy))

	log.Printf("[Agent Vendeur] ✓ Prix final décidé: $%.2f (marge %d%%)",
		float64(finalPrice)/100, decision.Margin)

	return decision, nil
}
