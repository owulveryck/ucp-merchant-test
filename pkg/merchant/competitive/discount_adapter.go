package competitive

import (
	"log"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/discount"
	ucpmodel "github.com/owulveryck/ucp-merchant-test/pkg/model"
)

// DiscountAdapter adapts the Orchestrator to implement discount.DiscountLookup.
// It handles both static discount codes (via baseData) and dynamic competitive
// pricing (via orchestrator when code is "AUTO_COMPETE").
type DiscountAdapter struct {
	baseData         discount.DiscountLookup
	orchestrator     *Orchestrator
	config           models.BusinessConfig
	onAgentDecisions func(*AgentDecisions) // Callback to notify of agent decisions
	lastDecisions    *AgentDecisions       // Last agent decisions for retrieval
}

// NewDiscountAdapter creates a new discount adapter.
func NewDiscountAdapter(
	baseData discount.DiscountLookup,
	orchestrator *Orchestrator,
	config models.BusinessConfig,
) *DiscountAdapter {
	return &DiscountAdapter{
		baseData:     baseData,
		orchestrator: orchestrator,
		config:       config,
	}
}

// FindDiscountByCode implements discount.DiscountLookup.
func (a *DiscountAdapter) FindDiscountByCode(code string) *discount.Discount {
	if code == "AUTO_COMPETE" {
		return &discount.Discount{
			Code:        "AUTO_COMPETE",
			Type:        "competitive",
			Value:       0, // Dynamic, calculated later
			Description: "Dynamic competitive pricing",
		}
	}
	return a.baseData.FindDiscountByCode(code)
}

// ApplyCompetitiveDiscounts applies competitive pricing discounts.
// This is used instead of the standard discount.ApplyDiscounts when AUTO_COMPETE is detected.
func (a *DiscountAdapter) ApplyCompetitiveDiscounts(
	codes []string,
	lineItems []ucpmodel.LineItem,
) *ucpmodel.Discounts {

	// Check if AUTO_COMPETE is in the codes
	hasAutoCompete := false
	for _, code := range codes {
		if code == "AUTO_COMPETE" {
			hasAutoCompete = true
			break
		}
	}

	// If no AUTO_COMPETE, delegate to standard discount logic
	if !hasAutoCompete {
		req := &ucpmodel.DiscountsRequest{Codes: codes}
		return discount.ApplyDiscounts(req, lineItems, a.baseData)
	}

	log.Printf("[DiscountAdapter] AUTO_COMPETE detected, using multi-agent pricing")

	// Calculate competitive discount for each line item
	totalDiscount := 0

	for _, item := range lineItems {
		// Get product price from line item
		productID := item.Item.ID
		ourPrice := item.Item.Price

		// Use configured context
		context := a.config

		// Orchestrator calculates the discount (with trace for dashboard)
		result, decisions := a.orchestrator.CalculateDiscountWithTrace(productID, ourPrice, context)

		// Store decisions for later retrieval
		a.lastDecisions = decisions

		// Notify dashboard if callback is set
		log.Printf("[DiscountAdapter] DEBUG: callback=%v, decisions=%v", a.onAgentDecisions != nil, decisions != nil)
		if a.onAgentDecisions != nil && decisions != nil {
			log.Printf("[DiscountAdapter] Calling agent decisions callback...")
			a.onAgentDecisions(decisions)
			log.Printf("[DiscountAdapter] Callback completed")
		} else {
			log.Printf("[DiscountAdapter] Callback NOT called (callback nil=%v, decisions nil=%v)", a.onAgentDecisions == nil, decisions == nil)
		}

		if result.Approved && !result.Rejected {
			// Apply discount for this item's quantity
			itemDiscount := result.FinalDiscount * item.Quantity
			totalDiscount += itemDiscount

			log.Printf("[DiscountAdapter] Product %s: discount $%.2f x %d = $%.2f",
				productID,
				float64(result.FinalDiscount)/100,
				item.Quantity,
				float64(itemDiscount)/100)
		}
	}

	if totalDiscount == 0 {
		log.Printf("[DiscountAdapter] No competitive discount applied")
		return nil
	}

	log.Printf("[DiscountAdapter] Total competitive discount: $%.2f", float64(totalDiscount)/100)

	return &ucpmodel.Discounts{
		Codes: codes,
		Applied: []ucpmodel.AppliedDiscount{
			{
				Code:   "AUTO_COMPETE",
				Title:  "Competitive Pricing",
				Amount: totalDiscount,
			},
		},
	}
}

// UpdateConfig updates the business configuration.
func (a *DiscountAdapter) UpdateConfig(config models.BusinessConfig) {
	a.config = config
}

// SetAgentDecisionsCallback sets the callback for agent decisions notifications.
func (a *DiscountAdapter) SetAgentDecisionsCallback(callback func(*AgentDecisions)) {
	a.onAgentDecisions = callback
}

// ApplyDiscountsWithContext is an alias for ApplyCompetitiveDiscounts.
// This method exists for compatibility with the arena merchant interface.
func (a *DiscountAdapter) ApplyDiscountsWithContext(
	codes []string,
	lineItems []ucpmodel.LineItem,
) *ucpmodel.Discounts {
	return a.ApplyCompetitiveDiscounts(codes, lineItems)
}

// GetLastDecisions returns the most recent agent decisions.
// Returns nil if no pricing calculation has been performed yet.
func (a *DiscountAdapter) GetLastDecisions() *AgentDecisions {
	return a.lastDecisions
}
