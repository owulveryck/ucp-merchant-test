// Package main demonstrates the unified multi-agent pricing system.
package main

import (
	"fmt"
	"log"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive"
	compAgents "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/agents"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/history"
	compModels "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
	pricing "github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/agents"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/models"
)

// MockCustomerDataSource for testing
type MockCustomerDataSource struct{}

func (m *MockCustomerDataSource) GetCustomerProfile(customerID string) (agents.CustomerProfile, error) {
	profiles := map[string]agents.CustomerProfile{
		"premium_customer": {
			CustomerID:       "premium_customer",
			TotalSpent:       150000, // $1500
			PurchaseCount:    15,
			LastPurchaseDays: 10,
		},
		"standard_customer": {
			CustomerID:       "standard_customer",
			TotalSpent:       10000, // $100
			PurchaseCount:    1,
			LastPurchaseDays: 90,
		},
	}

	if profile, ok := profiles[customerID]; ok {
		return profile, nil
	}

	return agents.CustomerProfile{
		CustomerID:       customerID,
		TotalSpent:       0,
		PurchaseCount:    0,
		LastPurchaseDays: 999,
	}, nil
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║  SYSTÈME MULTI-AGENTS UNIFIÉ - ARCHITECTURE HYBRIDE     ║")
	fmt.Println("║  Agent Vendeur → Customer Growth + Compétitivité        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Setup: Create the existing 4-agent system (for Agent 3 Compétitivité)
	sgClient := competitive.NewShoppingGraphClient("http://localhost:9000")
	priceIntel := compAgents.NewPriceIntelligenceAgent(sgClient, "test_merchant")
	historyStore := history.NewInMemoryHistoryStore()
	marketAnalyst := compAgents.NewMarketAnalysisAgent(historyStore)
	businessConfig := compModels.BusinessConfig{
		Objective:      "volume",
		MinMargin:      10,
		CostPercent:    60,
		StockThreshold: 20,
		BrandPosition:  "mid",
	}
	strategyRec := compAgents.NewStrategyRecommenderAgent(businessConfig)
	marginConfig := compModels.MarginConfig{
		MinMarginPercent: 10,
		CostPercent:      60,
		ActualCost:       5000,
		HardFloor:        true,
	}
	marginVal := compAgents.NewMarginValidatorAgent(marginConfig)

	orchestrator4Agents := competitive.NewOrchestrator(
		priceIntel,
		marketAnalyst,
		strategyRec,
		marginVal,
	)

	// Create the 3 unified agents
	customerData := &MockCustomerDataSource{}

	agent2CustomerGrowth := agents.NewCustomerGrowthAgent(customerData)
	agent3Competitiveness := agents.NewCompletivenessAgent(orchestrator4Agents, "test_merchant", 5000, businessConfig)

	// Create Agent 1 Vendeur (orchestrator)
	agent1Vendeur := pricing.NewVendorOrchestrator(agent2CustomerGrowth, agent3Competitiveness)

	// Test scenarios
	scenarios := []struct {
		name    string
		request models.PricingRequest
	}{
		{
			name: "Client Premium demande un produit",
			request: models.PricingRequest{
				ProductID:  "casque_audio",
				CustomerID: "premium_customer",
				BasePrice:  6000, // $60
				CostPrice:  5000, // $50
			},
		},
		{
			name: "Client Standard demande un produit",
			request: models.PricingRequest{
				ProductID:  "casque_audio",
				CustomerID: "standard_customer",
				BasePrice:  6000, // $60
				CostPrice:  5000, // $50
			},
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("SCÉNARIO %d: %s\n", i+1, scenario.name)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		decision, err := agent1Vendeur.DeterminePricing(scenario.request)
		if err != nil {
			log.Printf("Erreur: %v", err)
			continue
		}

		fmt.Printf("👤 AGENT 2: CUSTOMER GROWTH\n")
		fmt.Printf("   Garder ce client ? %s\n", map[bool]string{true: "✅ OUI", false: "❌ NON"}[decision.CustomerGrowth.ShouldRetain])
		fmt.Printf("   Tier: %s\n", decision.CustomerGrowth.CustomerTier)
		fmt.Printf("   Réduction suggérée: %d%%\n", decision.CustomerGrowth.SuggestedDiscount)
		fmt.Printf("   Lifetime Value: $%.2f\n\n", float64(decision.CustomerGrowth.LifetimeValue)/100)

		fmt.Printf("📊 AGENT 3: COMPÉTITIVITÉ\n")
		fmt.Printf("   Compétitif ? %s\n", map[bool]string{true: "✅ OUI", false: "❌ NON"}[decision.Competitiveness.IsCompetitive])
		fmt.Printf("   Position: %d/%d\n", decision.Competitiveness.MarketPosition, decision.Competitiveness.TotalCompetitors)
		if decision.Competitiveness.LowestCompetitor > 0 {
			fmt.Printf("   Concurrent le moins cher: $%.2f\n", float64(decision.Competitiveness.LowestCompetitor)/100)
		}
		fmt.Printf("   Prix recommandé: $%.2f\n\n", float64(decision.Competitiveness.RecommendedPrice)/100)

		fmt.Printf("🎯 AGENT 1: VENDEUR (DÉCISION FINALE)\n")
		fmt.Printf("   Prix de base: $%.2f\n", float64(decision.OriginalPrice)/100)
		fmt.Printf("   Prix final offert: $%.2f\n", float64(decision.FinalPrice)/100)
		fmt.Printf("   Réduction: $%.2f (%d%%)\n", float64(decision.TotalDiscount)/100, decision.DiscountPercent)
		fmt.Printf("   Marge: %d%%\n", decision.Margin)
		fmt.Printf("   Stratégie: %s\n\n", decision.Strategy)

		fmt.Printf("💡 RAISONNEMENT:\n")
		for _, reason := range decision.DecisionReasoning {
			if reason != "" {
				fmt.Printf("   %s\n", reason)
			}
		}
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Println("✓ Démonstration terminée")
}
