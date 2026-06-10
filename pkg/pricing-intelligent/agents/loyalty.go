// Package agents contains specialized pricing agents.
package agents

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-intelligent/models"
)

// LoyaltyAgentImpl implements the loyalty analysis agent.
type LoyaltyAgentImpl struct {
	customerData models.CustomerDataSource
	vipThreshold VIPThreshold
}

// VIPThreshold defines criteria for VIP status.
type VIPThreshold struct {
	MinTotalSpent      int // Minimum total spent (cents)
	MinPurchaseCount   int // Minimum number of purchases
	MaxDaysSinceLastPurchase int // Maximum days since last purchase
}

// DefaultVIPThreshold provides default VIP criteria.
var DefaultVIPThreshold = VIPThreshold{
	MinTotalSpent:      50000, // $500
	MinPurchaseCount:   5,
	MaxDaysSinceLastPurchase: 90,
}

// NewLoyaltyAgent creates a new loyalty agent.
func NewLoyaltyAgent(customerData models.CustomerDataSource, threshold VIPThreshold) *LoyaltyAgentImpl {
	return &LoyaltyAgentImpl{
		customerData: customerData,
		vipThreshold: threshold,
	}
}

// AnalyzeCustomer analyzes customer profile and determines VIP pricing.
//
// INTENTION: "Ce client mérite-t-il un prix préférentiel ?"
//
// ANALYSE:
// - Historique d'achats
// - Montant dépensé
// - Potentiel de fidélité
//
// DÉCISION:
// - Client VIP : Prix ajusté (discount 5-15%)
// - Client standard : Prix normal
func (a *LoyaltyAgentImpl) AnalyzeCustomer(customerID string) (models.LoyaltyDecision, error) {
	// Get customer profile
	profile, err := a.customerData.GetCustomerProfile(customerID)
	if err != nil {
		return models.LoyaltyDecision{}, fmt.Errorf("failed to get customer profile: %w", err)
	}

	decision := models.LoyaltyDecision{
		CustomerID: customerID,
		Reasoning:  []string{},
	}

	// Check VIP criteria
	isVIP := true
	confidence := 100

	// Criterion 1: Total spent
	if profile.TotalSpent < a.vipThreshold.MinTotalSpent {
		isVIP = false
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Total dépensé $%.2f < seuil VIP $%.2f",
				float64(profile.TotalSpent)/100,
				float64(a.vipThreshold.MinTotalSpent)/100))
		confidence -= 30
	} else {
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("✓ Total dépensé $%.2f (excellent)",
				float64(profile.TotalSpent)/100))
	}

	// Criterion 2: Purchase count
	if profile.PurchaseCount < a.vipThreshold.MinPurchaseCount {
		isVIP = false
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Nombre d'achats %d < seuil VIP %d",
				profile.PurchaseCount, a.vipThreshold.MinPurchaseCount))
		confidence -= 30
	} else {
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("✓ %d achats (client fidèle)", profile.PurchaseCount))
	}

	// Criterion 3: Recency
	if profile.LastPurchaseDays > a.vipThreshold.MaxDaysSinceLastPurchase {
		confidence -= 20
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("⚠ Dernier achat il y a %d jours (inactif)", profile.LastPurchaseDays))
	} else {
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("✓ Dernier achat il y a %d jours (actif)", profile.LastPurchaseDays))
	}

	decision.IsVIP = isVIP
	decision.Confidence = max(0, confidence)

	// Calculate suggested discount
	if isVIP {
		// VIP discount: 5-15% based on spending level
		if profile.TotalSpent >= 100000 { // $1000+
			decision.SuggestedDiscount = 15
			decision.Reasoning = append(decision.Reasoning, "💎 Client VIP Premium : 15% de réduction")
		} else if profile.TotalSpent >= 75000 { // $750+
			decision.SuggestedDiscount = 10
			decision.Reasoning = append(decision.Reasoning, "💎 Client VIP : 10% de réduction")
		} else {
			decision.SuggestedDiscount = 5
			decision.Reasoning = append(decision.Reasoning, "💎 Client VIP : 5% de réduction")
		}
	} else {
		decision.SuggestedDiscount = 0
		decision.Reasoning = append(decision.Reasoning, "Client standard : prix normal")
	}

	return decision, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
