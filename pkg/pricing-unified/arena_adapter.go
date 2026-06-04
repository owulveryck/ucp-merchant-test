// Package pricing provides Arena integration for the unified multi-agent pricing system.
package pricing

import (
	"log"

	"github.com/owulveryck/ucp-merchant-test/pkg/model"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/models"
)

// ArenaAdapter adapts the unified pricing system to Arena's discount interface.
type ArenaAdapter struct {
	vendorOrchestrator *VendorOrchestrator
	lastDecision       *models.VendorDecision
	decisionCallback   func(*models.VendorDecision)
}

// NewArenaAdapter creates an adapter for Arena integration.
func NewArenaAdapter(vendorOrchestrator *VendorOrchestrator) *ArenaAdapter {
	return &ArenaAdapter{
		vendorOrchestrator: vendorOrchestrator,
	}
}

// SetDecisionCallback sets a callback to receive agent decisions.
func (a *ArenaAdapter) SetDecisionCallback(callback func(*models.VendorDecision)) {
	a.decisionCallback = callback
}

// GetLastDecision returns the last pricing decision.
func (a *ArenaAdapter) GetLastDecision() *models.VendorDecision {
	return a.lastDecision
}

// ApplyDiscounts implements the Arena discount interface.
// This is called by legacy code that doesn't know about AUTO_COMPETE.
func (a *ArenaAdapter) ApplyDiscounts(codes []string) *model.Discounts {
	// Legacy interface - not used with competitive pricing
	return nil
}

// ApplyDiscountsWithContext implements the competitive pricing interface.
// This is called when AUTO_COMPETE code is detected.
func (a *ArenaAdapter) ApplyDiscountsWithContext(codes []string, lineItems []model.LineItem) *model.Discounts {
	// Check if AUTO_COMPETE is in the codes
	hasAutoCompete := false
	for _, code := range codes {
		if code == "AUTO_COMPETE" {
			hasAutoCompete = true
			break
		}
	}

	if !hasAutoCompete {
		// No AUTO_COMPETE, return no discount
		return nil
	}

	log.Printf("[ArenaAdapter] AUTO_COMPETE detected, using 3-agent pricing system")

	if len(lineItems) == 0 {
		log.Printf("[ArenaAdapter] No line items")
		return nil
	}

	// Get the first product (simplified for now)
	firstItem := lineItems[0]
	productID := firstItem.Item.ID
	basePrice := firstItem.Item.Price

	// For now, use a default customer ID
	// In a real implementation, this would come from the checkout context
	customerID := "default_customer"

	// Create pricing request
	request := models.PricingRequest{
		ProductID:  productID,
		CustomerID: customerID,
		BasePrice:  basePrice,
		CostPrice:  basePrice * 80 / 100, // Assume 80% cost (20% base margin)
	}

	// Call the 3-agent system
	decision, err := a.vendorOrchestrator.DeterminePricing(request)
	if err != nil {
		log.Printf("[ArenaAdapter] Error determining pricing: %v", err)
		return nil
	}

	// Store the decision
	a.lastDecision = &decision

	// Call the callback if set
	if a.decisionCallback != nil {
		a.decisionCallback(&decision)
	}

	// Calculate total discount amount
	discountAmount := decision.TotalDiscount

	log.Printf("[ArenaAdapter] 3-agent system decision: FinalPrice=$%.2f, Discount=$%.2f, Strategy=%s",
		float64(decision.FinalPrice)/100, float64(discountAmount)/100, decision.Strategy)

	// Build the discount result
	result := &model.Discounts{
		Codes: codes,
		Applied: []model.AppliedDiscount{
			{
				Code:   "AUTO_COMPETE",
				Amount: discountAmount,
			},
		},
	}

	log.Printf("[ArenaAdapter] Strategy: %s, Reasoning: %v", decision.Strategy, decision.DecisionReasoning)

	return result
}
