package main

import (
	"encoding/json"
	"log"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive"
	compAgents "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/agents"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/history"
	compModels "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
	pricing "github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/agents"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/datasources"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/models"
)

// setup3AgentPricing configures the new 3-agent system for a merchant.
func (s *ArenaServer) setup3AgentPricing(m *arenaMerchant, name, id string, notifier *Notifier) {
	log.Printf("Enabling 3-AGENT UNIFIED pricing for tenant %s (minMargin=%d%%)", name, s.minMargin)

	// Create customer data source
	customerData := datasources.NewMockCustomerDataSource()

	// Agent 2: Customer Growth
	agent2 := agents.NewCustomerGrowthAgent(customerData)
	log.Printf("[%s] Agent 2 (Customer Growth) initialized", name)

	// Agent 3: Competitiveness (wraps existing 4-agent system)
	// We need to create the 4-agent system for Agent 3 to wrap
	sgClient := competitive.NewShoppingGraphClient(s.graphURL)
	priceIntel := compAgents.NewPriceIntelligenceAgent(sgClient, id)
	historyStore := history.NewInMemoryHistoryStore()
	marketAnalyst := compAgents.NewMarketAnalysisAgent(historyStore)

	businessConfig := compModels.BusinessConfig{
		Objective:      "volume",
		StockThreshold: 20,
		BrandPosition:  "mid",
		MinMargin:      s.minMargin,
		CostPercent:    60,
	}
	strategyRec := compAgents.NewStrategyRecommenderAgent(businessConfig)

	marginConfig := compModels.MarginConfig{
		MinMarginPercent: s.minMargin,
		CostPercent:      60,
		ActualCost:       s.costPrice,
		HardFloor:        true,
	}
	marginVal := compAgents.NewMarginValidatorAgent(marginConfig)

	orchestrator4Agents := competitive.NewOrchestrator(
		priceIntel,
		marketAnalyst,
		strategyRec,
		marginVal,
	)

	agent3 := agents.NewCompletivenessAgent(orchestrator4Agents, id, s.costPrice, businessConfig)
	log.Printf("[%s] Agent 3 (Competitiveness) initialized - wraps 4-agent system", name)

	// Agent 1: Vendor Orchestrator
	agent1 := pricing.NewVendorOrchestrator(agent2, agent3)
	log.Printf("[%s] Agent 1 (Vendor Orchestrator) initialized", name)

	// Create Arena adapter
	adapter := pricing.NewArenaAdapter(agent1)

	// Set callback to send decisions to dashboard via SSE
	adapter.SetDecisionCallback(func(decision *models.VendorDecision) {
		log.Printf("[Tenant %s] 3-Agent decisions callback triggered!", name)

		event := map[string]interface{}{
			"type": "vendor_decision",
			"agent1": map[string]interface{}{
				"final_price":     decision.FinalPrice,
				"strategy":        decision.Strategy,
				"margin":          decision.Margin,
				"discount_pct":    decision.DiscountPercent,
				"total_discount":  decision.TotalDiscount,
				"reasoning":       decision.DecisionReasoning,
			},
			"agent2": map[string]interface{}{
				"should_retain":      decision.CustomerGrowth.ShouldRetain,
				"tier":               decision.CustomerGrowth.CustomerTier,
				"suggested_discount": decision.CustomerGrowth.SuggestedDiscount,
				"lifetime_value":     decision.CustomerGrowth.LifetimeValue,
				"reasoning":          decision.CustomerGrowth.RetentionReasoning,
			},
			"agent3": map[string]interface{}{
				"is_competitive":    decision.Competitiveness.IsCompetitive,
				"market_position":   decision.Competitiveness.MarketPosition,
				"total_competitors": decision.Competitiveness.TotalCompetitors,
				"lowest_competitor": decision.Competitiveness.LowestCompetitor,
				"recommended_price": decision.Competitiveness.RecommendedPrice,
				"strategy":          decision.Competitiveness.Strategy,
				"margin":            decision.Competitiveness.Margin,
				"reasoning":         decision.Competitiveness.CompetitiveReasoning,
			},
		}

		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("[Tenant %s] ERROR marshaling JSON: %v", name, err)
			return
		}

		log.Printf("[Tenant %s] Sending vendor_decision SSE event", name)
		notifier.SendRaw(data)
	})

	// Inject into merchant
	m.pricingAgent = adapter

	log.Printf("✅ 3-AGENT UNIFIED pricing configured for %s (Vendor → Customer Growth + Competitiveness)", name)
}
