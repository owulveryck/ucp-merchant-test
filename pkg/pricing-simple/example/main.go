// Package main demonstrates the simplified 3-agent pricing system.
package main

import (
	"fmt"

	pricing "github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple/agents"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple/datasources"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple/models"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║  SYSTÈME MULTI-AGENTS DE PRICING INTELLIGENT (3)    ║")
	fmt.Println("║  Le bon prix, au bon client, selon le marché        ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create data sources
	customerData := datasources.NewMockCustomerData()
	competitorData := datasources.NewMockCompetitorData()

	// Create the 3 agents
	agent1 := agents.NewMarketIntelligenceAgent(competitorData, "our_store")
	agent2 := agents.NewCustomerRetentionAgent(customerData)
	agent3 := agents.NewFinalDecisionAgent(10, true) // 10% min margin, hard floor

	// Create orchestrator
	orchestrator := pricing.NewOrchestrator(agent1, agent2, agent3)

	// Test scenarios
	scenarios := []struct {
		name     string
		request  models.PricingRequest
	}{
		{
			name: "Client Premium + Marché compétitif",
			request: models.PricingRequest{
				ProductID:  "headphones",
				CustomerID: "customer_premium",
				BasePrice:  6000, // $60
				CostPrice:  5000, // $50
			},
		},
		{
			name: "Client Standard + Marché compétitif",
			request: models.PricingRequest{
				ProductID:  "headphones",
				CustomerID: "customer_standard",
				BasePrice:  6000, // $60
				CostPrice:  5000, // $50
			},
		},
		{
			name: "Client Gold + Leader marché",
			request: models.PricingRequest{
				ProductID:  "laptop",
				CustomerID: "customer_gold",
				BasePrice:  75000, // $750
				CostPrice:  65000, // $650
			},
		},
		{
			name: "Nouveau client + Monopole",
			request: models.PricingRequest{
				ProductID:  "phone",
				CustomerID: "customer_new",
				BasePrice:  50000, // $500
				CostPrice:  40000, // $400
			},
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("SCÉNARIO %d: %s\n", i+1, scenario.name)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		result, decisions, err := orchestrator.CalculateOptimalPrice(scenario.request)
		if err != nil {
			fmt.Printf("❌ Erreur: %v\n", err)
			continue
		}

		// Display results
		fmt.Printf("📊 AGENT 1 : INTELLIGENCE MARCHÉ\n")
		fmt.Printf("   Position : %d/%d\n", decisions.Market.OurPosition, decisions.Market.TotalCompetitors)
		fmt.Printf("   Prix concurrent le plus bas : $%.2f\n", float64(decisions.Market.LowestCompetitor)/100)
		if len(decisions.Market.DiscountCodesFound) > 0 {
			fmt.Printf("   Codes promo détectés : %v\n", decisions.Market.DiscountCodesFound)
		}
		fmt.Printf("   Tendance : %s\n\n", decisions.Market.MarketTrend)

		fmt.Printf("💎 AGENT 2 : CUSTOMER RETENTION\n")
		fmt.Printf("   Tier : %s\n", decisions.Retention.CustomerTier)
		fmt.Printf("   VIP : %v\n", decisions.Retention.IsVIP)
		fmt.Printf("   Réduction suggérée : %d%%\n", decisions.Retention.SuggestedDiscount)
		fmt.Printf("   Lifetime Value : $%.2f\n\n", float64(decisions.Retention.LifetimeValue)/100)

		fmt.Printf("⚙️  AGENT 3 : DÉCISION FINALE\n")
		fmt.Printf("   Stratégie : %s\n", result.Strategy)
		fmt.Printf("   Prix de base : $%.2f\n", float64(result.OriginalPrice)/100)
		fmt.Printf("   Prix final : $%.2f\n", float64(result.FinalPrice)/100)
		fmt.Printf("   Réduction totale : $%.2f (%d%%)\n",
			float64(result.TotalDiscount)/100, result.DiscountPercent)
		fmt.Printf("   Marge : %d%%\n\n", result.Margin)

		if len(result.Warnings) > 0 {
			fmt.Printf("⚠️  AVERTISSEMENTS:\n")
			for _, warning := range result.Warnings {
				fmt.Printf("   - %s\n", warning)
			}
			fmt.Println()
		}

		fmt.Printf("💡 RAISONNEMENT COMPLET:\n")
		for _, reason := range result.Reasoning {
			if reason != "" {
				fmt.Printf("   %s\n", reason)
			}
		}
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Println("✓ Démonstration terminée")
}
