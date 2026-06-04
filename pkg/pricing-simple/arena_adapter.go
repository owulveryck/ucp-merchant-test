// Package pricing provides Arena integration for the 3-agent system.
package pricing

import (
	"log"
	"strings"

	"github.com/owulveryck/ucp-merchant-test/pkg/model"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple/models"
)

// ArenaAdapter adapts the 3-agent system to Arena's discount interface.
type ArenaAdapter struct {
	orchestrator  *Orchestrator
	costPrice     int
	customerStore CustomerDataSource
}

// CustomerDataSource interface for customer profile lookup.
type CustomerDataSource interface {
	GetCustomerProfile(customerID string) (models.CustomerProfile, error)
}

// CompetitorDataSource interface for competitor price lookup.
type CompetitorDataSource interface {
	GetCompetitorPrices(productID string) ([]models.CompetitorPrice, error)
}

// NewArenaAdapter creates an adapter for Arena integration.
func NewArenaAdapter(
	orchestrator *Orchestrator,
	costPrice int,
	customerStore CustomerDataSource,
) *ArenaAdapter {
	return &ArenaAdapter{
		orchestrator:  orchestrator,
		costPrice:     costPrice,
		customerStore: customerStore,
	}
}

// CalculateDiscount calculates discount using the 3-agent system.
//
// This method is called by Arena's discount_adapter.go when AUTO_COMPETE code is detected.
func (a *ArenaAdapter) CalculateDiscount(productID string, ourPrice int, context map[string]interface{}) (int, int, error) {
	// Extract customer ID from context (if available)
	customerID := "customer_standard" // default
	if ctx, ok := context["customer_id"].(string); ok && ctx != "" {
		customerID = ctx
	}

	// Build pricing request
	request := models.PricingRequest{
		ProductID:  productID,
		CustomerID: customerID,
		BasePrice:  ourPrice,
		CostPrice:  a.costPrice,
	}

	// Call 3-agent orchestrator
	decision, decisions, err := a.orchestrator.CalculateOptimalPrice(request)
	if err != nil {
		log.Printf("[ArenaAdapter] Orchestrator error: %v", err)
		return 0, ourPrice, err
	}

	// Log agent decisions for Arena dashboard
	log.Printf("[ArenaAdapter] 3-AGENT PRICING COMPLETE")
	log.Printf("  Agent 1 (Market Intelligence): Position=%d/%d, Lowest=$%.2f",
		decisions.Market.OurPosition,
		decisions.Market.TotalCompetitors,
		float64(decisions.Market.LowestCompetitor)/100)
	log.Printf("  Agent 2 (Customer Retention): Tier=%s, VIP=%v, Discount=%d%%",
		decisions.Retention.CustomerTier,
		decisions.Retention.IsVIP,
		decisions.Retention.SuggestedDiscount)
	log.Printf("  Agent 3 (Final Decision): Strategy=%s, Price=$%.2f, Margin=%d%%",
		decision.Strategy,
		float64(decision.FinalPrice)/100,
		decision.Margin)

	// Calculate discount amount
	discount := ourPrice - decision.FinalPrice

	// Return discount and final price
	return discount, decision.FinalPrice, nil
}

// CalculateDiscountWithTrace calculates discount and returns full agent decisions.
//
// This method is used when Arena needs the full decision trail for the dashboard.
func (a *ArenaAdapter) CalculateDiscountWithTrace(productID string, ourPrice int, context map[string]interface{}) (Decision, *AgentDecisions, error) {
	// Extract customer ID from context (if available)
	customerID := "customer_standard" // default
	if ctx, ok := context["customer_id"].(string); ok && ctx != "" {
		customerID = ctx
	}

	// Build pricing request
	request := models.PricingRequest{
		ProductID:  productID,
		CustomerID: customerID,
		BasePrice:  ourPrice,
		CostPrice:  a.costPrice,
	}

	// Call 3-agent orchestrator
	decision, decisions, err := a.orchestrator.CalculateOptimalPrice(request)
	if err != nil {
		log.Printf("[ArenaAdapter] Orchestrator error: %v", err)
		return Decision{}, nil, err
	}

	// Convert to Arena's Decision format
	arenaDecision := Decision{
		Approved:      decision.Approved,
		FinalPrice:    decision.FinalPrice,
		FinalDiscount: ourPrice - decision.FinalPrice,
		Margin:        decision.Margin,
		Warnings:      decision.Warnings,
		Strategy:      decision.Strategy,
	}

	return arenaDecision, decisions, nil
}

// Decision represents the final pricing decision for Arena.
type Decision struct {
	Approved      bool
	FinalPrice    int
	FinalDiscount int
	Margin        int
	Warnings      []string
	Strategy      string
}

// ApplyDiscountsWithContext is the main entry point called by Arena when AUTO_COMPETE is used.
//
// This method implements the interface expected by arenaMerchant.pricingAgent.
func (a *ArenaAdapter) ApplyDiscountsWithContext(codes []string, lineItems []model.LineItem) *model.Discounts {
	// Check if AUTO_COMPETE code is present
	hasAutoCompete := false
	for _, code := range codes {
		if strings.EqualFold(code, "AUTO_COMPETE") {
			hasAutoCompete = true
			break
		}
	}

	if !hasAutoCompete {
		log.Printf("[ArenaAdapter] No AUTO_COMPETE code - returning nil")
		return nil
	}

	log.Printf("[ArenaAdapter] AUTO_COMPETE detected - activating 3-agent pricing system")

	// Process each line item
	totalDiscount := 0
	appliedDiscounts := []model.AppliedDiscount{}

	for _, item := range lineItems {
		// Extract product ID and price from item
		productID := item.Item.ID
		basePrice := item.Item.Price

		log.Printf("[ArenaAdapter] Processing item: %s, basePrice=$%.2f",
			productID, float64(basePrice)/100)

		// Build pricing request
		request := models.PricingRequest{
			ProductID:  productID,
			CustomerID: "customer_standard", // Default for now
			BasePrice:  basePrice,
			CostPrice:  a.costPrice,
		}

		// Call 3-agent orchestrator
		decision, decisions, err := a.orchestrator.CalculateOptimalPrice(request)
		if err != nil {
			log.Printf("[ArenaAdapter] Orchestrator error: %v", err)
			continue
		}

		// Log agent decisions
		log.Printf("[ArenaAdapter] 3-AGENT PRICING COMPLETE for %s", productID)
		log.Printf("  Agent 1 (Market Intelligence): Position=%d/%d, Lowest=$%.2f",
			decisions.Market.OurPosition,
			decisions.Market.TotalCompetitors,
			float64(decisions.Market.LowestCompetitor)/100)
		log.Printf("  Agent 2 (Customer Retention): Tier=%s, VIP=%v, Discount=%d%%",
			decisions.Retention.CustomerTier,
			decisions.Retention.IsVIP,
			decisions.Retention.SuggestedDiscount)
		log.Printf("  Agent 3 (Final Decision): Strategy=%s, Price=$%.2f, Margin=%d%%",
			decision.Strategy,
			float64(decision.FinalPrice)/100,
			decision.Margin)

		// Calculate discount for this item
		itemDiscount := basePrice - decision.FinalPrice
		if itemDiscount > 0 {
			itemTotalDiscount := itemDiscount * item.Quantity
			totalDiscount += itemTotalDiscount

			appliedDiscounts = append(appliedDiscounts, model.AppliedDiscount{
				Code:   "AUTO_COMPETE",
				Title:  "Multi-Agent Competitive Pricing",
				Amount: itemTotalDiscount,
			})

			log.Printf("[ArenaAdapter] Item discount: $%.2f x %d = $%.2f total",
				float64(itemDiscount)/100,
				item.Quantity,
				float64(itemTotalDiscount)/100)
		}
	}

	// Return discount structure
	if totalDiscount <= 0 {
		log.Printf("[ArenaAdapter] No discount applied (totalDiscount=$%.2f)", float64(totalDiscount)/100)
		return nil
	}

	log.Printf("[ArenaAdapter] FINAL DISCOUNT: $%.2f", float64(totalDiscount)/100)

	return &model.Discounts{
		Codes:   codes,
		Applied: appliedDiscounts,
	}
}
