// Package main demonstrates the intelligent pricing system.
package main

import (
	"fmt"
	"log"

	pricing "github.com/owulveryck/ucp-merchant-test/pkg/pricing-intelligent"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-intelligent/agents"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-intelligent/models"
)

// MockCustomerDataSource provides mock customer data.
type MockCustomerDataSource struct {
	customers map[string]models.CustomerProfile
}

func NewMockCustomerData() *MockCustomerDataSource {
	return &MockCustomerDataSource{
		customers: map[string]models.CustomerProfile{
			"customer_vip": {
				CustomerID:      "customer_vip",
				TotalSpent:      120000, // $1200
				PurchaseCount:   15,
				AveragePurchase: 8000, // $80
				LastPurchaseDays: 10,
				IsVIP:           true,
			},
			"customer_standard": {
				CustomerID:      "customer_standard",
				TotalSpent:      20000, // $200
				PurchaseCount:   2,
				AveragePurchase: 10000, // $100
				LastPurchaseDays: 45,
				IsVIP:           false,
			},
		},
	}
}

func (m *MockCustomerDataSource) GetCustomerProfile(customerID string) (models.CustomerProfile, error) {
	if profile, ok := m.customers[customerID]; ok {
		return profile, nil
	}
	// Return default for unknown customers
	return models.CustomerProfile{
		CustomerID:      customerID,
		TotalSpent:      0,
		PurchaseCount:   0,
		AveragePurchase: 0,
		LastPurchaseDays: 999,
		IsVIP:           false,
	}, nil
}

// MockCompetitorDataSource provides mock competitor prices.
type MockCompetitorDataSource struct {
	prices map[string][]int
}

func NewMockCompetitorData() *MockCompetitorDataSource {
	return &MockCompetitorDataSource{
		prices: map[string][]int{
			"product_competitive": {6200, 6500, 5900}, // Competitors: $62, $65, $59
			"product_leader":      {7000, 7200},       // We're cheapest at $60
			"product_alone":       {},                 // No competitors
		},
	}
}

func (m *MockCompetitorDataSource) GetCompetitorPrices(productID string) ([]int, error) {
	if prices, ok := m.prices[productID]; ok {
		return prices, nil
	}
	return []int{}, nil
}

func main() {
	fmt.Println("=== SYSTÈME MULTI-AGENTS DE PRICING INTELLIGENT ===\n")

	// Create data sources
	customerData := NewMockCustomerData()
	competitorData := NewMockCompetitorData()

	// Create agents
	loyaltyAgent := agents.NewLoyaltyAgent(customerData, agents.DefaultVIPThreshold)
	compAgent := agents.NewCompetitivenessAgent(competitorData)

	// Create orchestrator
	orchestrator := pricing.NewOrchestrator(loyaltyAgent, compAgent, 10) // 10% min margin

	// Test scenarios
	scenarios := []struct {
		name       string
		request    models.PricingRequest
	}{
		{
			name: "Client VIP + Produit compétitif",
			request: models.PricingRequest{
				CustomerID: "customer_vip",
				ProductID:  "product_competitive",
				BasePrice:  6000, // $60
				CostPrice:  5000, // $50
			},
		},
		{
			name: "Client Standard + Produit compétitif",
			request: models.PricingRequest{
				CustomerID: "customer_standard",
				ProductID:  "product_competitive",
				BasePrice:  6000, // $60
				CostPrice:  5000, // $50
			},
		},
		{
			name: "Client VIP + Déjà leader",
			request: models.PricingRequest{
				CustomerID: "customer_vip",
				ProductID:  "product_leader",
				BasePrice:  6000, // $60 (already cheapest)
				CostPrice:  5000, // $50
			},
		},
		{
			name: "Client Standard + Sans concurrence",
			request: models.PricingRequest{
				CustomerID: "customer_standard",
				ProductID:  "product_alone",
				BasePrice:  8000, // $80
				CostPrice:  6000, // $60
			},
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("SCÉNARIO %d: %s\n", i+1, scenario.name)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		result, err := orchestrator.CalculateOptimalPrice(scenario.request)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		// Display result
		fmt.Printf("📋 RÉSULTAT:\n")
		fmt.Printf("   Prix de base : $%.2f\n", float64(result.BasePrice)/100)
		fmt.Printf("   Prix final   : $%.2f\n", float64(result.FinalPrice)/100)
		fmt.Printf("   Réduction    : $%.2f (%d%%)\n",
			float64(result.Discount)/100, result.DiscountPercent)
		fmt.Printf("   Marge        : %d%%\n", result.Margin)
		fmt.Printf("   Client VIP   : %v\n", result.IsVIP)
		fmt.Printf("   Compétitif   : %v\n\n", result.IsCompetitive)

		fmt.Printf("💡 RAISONNEMENT:\n")
		for _, reason := range result.Reasoning {
			fmt.Printf("   %s\n", reason)
		}
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Println("✓ Démonstration terminée")
}
