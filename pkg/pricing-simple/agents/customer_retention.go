// Package agents contains the 3 specialized pricing agents.
package agents

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple/models"
)

// CustomerRetentionAgentImpl implements customer value analysis (Agent 2).
type CustomerRetentionAgentImpl struct {
	customerData models.CustomerDataSource
}

// VIPThresholds defines criteria for customer tiers.
type VIPThresholds struct {
	PremiumMinSpent int // $1000+
	GoldMinSpent    int // $500+
	SilverMinSpent  int // $200+
	MinPurchases    int // 2+ purchases for any tier
}

// DefaultVIPThresholds provides default tier criteria.
var DefaultVIPThresholds = VIPThresholds{
	PremiumMinSpent: 100000, // $1000
	GoldMinSpent:    50000,  // $500
	SilverMinSpent:  20000,  // $200
	MinPurchases:    2,
}

// NewCustomerRetentionAgent creates Agent 2.
func NewCustomerRetentionAgent(customerData models.CustomerDataSource) *CustomerRetentionAgentImpl {
	return &CustomerRetentionAgentImpl{
		customerData: customerData,
	}
}

// AnalyzeCustomer analyzes customer value and determines VIP pricing.
//
// INTENTION: "Ce client mérite-t-il un prix préférentiel ?"
//
// ANALYSE:
// - Historique d'achats
// - Montant total dépensé (Lifetime Value)
// - Potentiel de fidélité
//
// DÉCISION:
// - Tier: Premium (15%), Gold (10%), Silver (5%), Standard (0%)
func (a *CustomerRetentionAgentImpl) AnalyzeCustomer(customerID string) (models.LoyaltyDecision, error) {
	decision := models.LoyaltyDecision{
		CustomerID: customerID,
		Reasoning:  []string{},
	}

	// Get customer profile
	profile, err := a.customerData.GetCustomerProfile(customerID)
	if err != nil {
		return decision, fmt.Errorf("failed to get customer profile: %w", err)
	}

	decision.LifetimeValue = profile.TotalSpent
	decision.PurchaseCount = profile.PurchaseCount

	// Determine tier and discount
	thresholds := DefaultVIPThresholds

	if profile.TotalSpent >= thresholds.PremiumMinSpent && profile.PurchaseCount >= 10 {
		decision.CustomerTier = "premium"
		decision.SuggestedDiscount = 15
		decision.IsVIP = true
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("💎 Client VIP Premium : $%.2f dépensé, %d achats",
				float64(profile.TotalSpent)/100, profile.PurchaseCount))
		decision.Reasoning = append(decision.Reasoning, "Réduction : 15%")
	} else if profile.TotalSpent >= thresholds.GoldMinSpent && profile.PurchaseCount >= 5 {
		decision.CustomerTier = "gold"
		decision.SuggestedDiscount = 10
		decision.IsVIP = true
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("🥇 Client Gold : $%.2f dépensé, %d achats",
				float64(profile.TotalSpent)/100, profile.PurchaseCount))
		decision.Reasoning = append(decision.Reasoning, "Réduction : 10%")
	} else if profile.TotalSpent >= thresholds.SilverMinSpent && profile.PurchaseCount >= thresholds.MinPurchases {
		decision.CustomerTier = "silver"
		decision.SuggestedDiscount = 5
		decision.IsVIP = true
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("🥈 Client Silver : $%.2f dépensé, %d achats",
				float64(profile.TotalSpent)/100, profile.PurchaseCount))
		decision.Reasoning = append(decision.Reasoning, "Réduction : 5%")
	} else {
		decision.CustomerTier = "standard"
		decision.SuggestedDiscount = 0
		decision.IsVIP = false
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Client Standard : $%.2f dépensé, %d achats",
				float64(profile.TotalSpent)/100, profile.PurchaseCount))
		decision.Reasoning = append(decision.Reasoning, "Pas de réduction fidélité")
	}

	// Add recency insight
	if profile.LastPurchaseDays <= 30 {
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("✓ Client actif (dernier achat : %d jours)", profile.LastPurchaseDays))
	} else if profile.LastPurchaseDays <= 90 {
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Client régulier (dernier achat : %d jours)", profile.LastPurchaseDays))
	} else {
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("⚠ Client inactif (dernier achat : %d jours) - à réactiver", profile.LastPurchaseDays))
	}

	return decision, nil
}
