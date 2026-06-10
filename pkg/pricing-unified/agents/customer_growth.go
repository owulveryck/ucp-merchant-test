// Package agents contains the specialized agents for the unified pricing system.
package agents

import (
	"fmt"
	"log"

	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/models"
)

// CustomerProfile represents customer data for analysis.
type CustomerProfile struct {
	CustomerID       string
	TotalSpent       int // Total dépensé en centimes
	PurchaseCount    int // Nombre d'achats
	LastPurchaseDays int // Jours depuis dernier achat
}

// CustomerDataSource provides customer data.
type CustomerDataSource interface {
	GetCustomerProfile(customerID string) (CustomerProfile, error)
}

// CustomerGrowthAgent is Agent 2: Analyzes customer retention value.
type CustomerGrowthAgent struct {
	dataSource CustomerDataSource
}

// NewCustomerGrowthAgent creates Agent 2.
func NewCustomerGrowthAgent(dataSource CustomerDataSource) *CustomerGrowthAgent {
	return &CustomerGrowthAgent{
		dataSource: dataSource,
	}
}

// Analyze answers the question: "Est-ce que lambda est un client que je peux garder ?"
//
// INTENTION: Identifier les clients de valeur pour leur offrir des avantages
// DÉCISION: OUI/NON + niveau de réduction selon la valeur client
func (a *CustomerGrowthAgent) Analyze(customerID string) (models.CustomerGrowthDecision, error) {
	log.Printf("[Agent Customer Growth] Analyzing customer: %s", customerID)

	// Get customer profile
	profile, err := a.dataSource.GetCustomerProfile(customerID)
	if err != nil {
		return models.CustomerGrowthDecision{}, err
	}

	decision := models.CustomerGrowthDecision{
		RetentionReasoning: []string{},
	}

	// Determine customer tier based on lifetime value
	totalSpent := profile.TotalSpent
	purchaseCount := profile.PurchaseCount

	decision.LifetimeValue = totalSpent

	// Tier classification with retention logic
	switch {
	case totalSpent >= 100000: // $1000+
		decision.ShouldRetain = true
		decision.CustomerTier = "premium"
		decision.SuggestedDiscount = 15
		decision.RetentionReasoning = append(decision.RetentionReasoning,
			fmt.Sprintf("Client PREMIUM: $%.2f dépensés - ABSOLUMENT à garder", float64(totalSpent)/100))

	case totalSpent >= 50000: // $500-$1000
		decision.ShouldRetain = true
		decision.CustomerTier = "gold"
		decision.SuggestedDiscount = 10
		decision.RetentionReasoning = append(decision.RetentionReasoning,
			fmt.Sprintf("Client GOLD: $%.2f dépensés - Important de garder", float64(totalSpent)/100))

	case totalSpent >= 20000: // $200-$500
		decision.ShouldRetain = true
		decision.CustomerTier = "silver"
		decision.SuggestedDiscount = 5
		decision.RetentionReasoning = append(decision.RetentionReasoning,
			fmt.Sprintf("Client SILVER: $%.2f dépensés - Bon à garder", float64(totalSpent)/100))

	default: // < $200
		decision.CustomerTier = "standard"
		decision.SuggestedDiscount = 0

		// Decide if worth retaining based on potential
		if purchaseCount >= 3 {
			decision.ShouldRetain = true
			decision.RetentionReasoning = append(decision.RetentionReasoning,
				fmt.Sprintf("Client régulier (%d achats) - Potentiel de croissance", purchaseCount))
		} else {
			decision.ShouldRetain = false
			decision.RetentionReasoning = append(decision.RetentionReasoning,
				fmt.Sprintf("Nouveau client ($%.2f dépensés) - Pas prioritaire pour réduction", float64(totalSpent)/100))
		}
	}

	// Add frequency insight
	if decision.ShouldRetain {
		decision.RetentionReasoning = append(decision.RetentionReasoning,
			fmt.Sprintf("Fréquence: %d achats - Fidélité confirmée", purchaseCount))
	}

	log.Printf("[Agent Customer Growth] Decision: ShouldRetain=%v, Tier=%s, Discount=%d%%",
		decision.ShouldRetain, decision.CustomerTier, decision.SuggestedDiscount)

	return decision, nil
}
