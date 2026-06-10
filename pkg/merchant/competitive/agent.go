package competitive

import (
	"fmt"
	"log"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/discount"
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
)

// FindDiscountByCode looks up a discount by code.
// Implements discount.DiscountLookup interface.
//
// Static codes are looked up via the base discount lookup.
// Special codes like "AUTO_COMPETE" return nil here; they're handled
// in ApplyDiscountsWithContext() which has access to line items.
func (a *CompetitivePricingAgent) FindDiscountByCode(code string) *discount.Discount {
	// First check base data for static codes
	if a.baseData != nil {
		if disc := a.baseData.FindDiscountByCode(code); disc != nil {
			return disc
		}
	}

	// AUTO_COMPETE and COMP_* codes are handled specially in ApplyDiscountsWithContext
	// Return nil here to indicate "not a static code"
	return nil
}

// calculateCompetitiveDiscount computes the total discount needed to beat competitors.
//
// For each line item:
//  1. Query Shopping Graph for lowest competitor price
//  2. Compare with our price
//  3. Calculate discount needed based on strategy
//  4. Validate margin constraints
//  5. Sum discounts across all items
func (a *CompetitivePricingAgent) calculateCompetitiveDiscount(lineItems []model.LineItem) int {
	totalDiscount := 0

	for _, item := range lineItems {
		calculation := a.calculateItemDiscount(item)

		if calculation.Applied {
			totalDiscount += calculation.DiscountAmount
			log.Printf("[CompetitivePricing] %s: our=$%.2f competitor=$%.2f discount=$%.2f margin=%d%%",
				calculation.ProductID,
				float64(calculation.OurPrice)/100,
				float64(calculation.CompetitorPrice)/100,
				float64(calculation.DiscountAmount)/100,
				calculation.MarginPercent,
			)
		} else {
			log.Printf("[CompetitivePricing] %s: %s",
				calculation.ProductID,
				calculation.Reason,
			)
		}
	}

	return totalDiscount
}

// calculateItemDiscount calculates competitive discount for a single line item.
func (a *CompetitivePricingAgent) calculateItemDiscount(item model.LineItem) DiscountCalculation {
	productID := item.Item.ID

	// Get our unit price from line item totals
	ourUnitPrice := 0
	if total := findTotal(item.Totals, "total"); total != nil {
		ourUnitPrice = total.Amount / item.Quantity
	}

	if ourUnitPrice == 0 {
		return DiscountCalculation{
			ProductID: productID,
			Applied:   false,
			Reason:    "unable to determine our price",
		}
	}

	// Query Shopping Graph for competitor prices
	competitorPrice, competitorID, err := a.competitorAPI.GetLowestPrice(productID)
	if err != nil {
		return DiscountCalculation{
			ProductID: productID,
			OurPrice:  ourUnitPrice,
			Applied:   false,
			Reason:    fmt.Sprintf("no competitor data: %v", err),
		}
	}

	// Skip if competitor is us
	if competitorID == a.merchantID {
		return DiscountCalculation{
			ProductID:       productID,
			OurPrice:        ourUnitPrice,
			CompetitorPrice: competitorPrice,
			Applied:         false,
			Reason:          "we already have lowest price",
		}
	}

	// Skip if we're already cheaper
	if ourUnitPrice <= competitorPrice {
		return DiscountCalculation{
			ProductID:            productID,
			OurPrice:             ourUnitPrice,
			CompetitorPrice:      competitorPrice,
			CompetitorMerchantID: competitorID,
			Applied:              false,
			Reason:               "already cheaper than competitor",
		}
	}

	// Calculate discount based on strategy
	discountAmount := a.calculateStrategyDiscount(ourUnitPrice, competitorPrice)

	// Apply to all quantities
	totalDiscountAmount := discountAmount * item.Quantity

	// Calculate final price and margin
	finalUnitPrice := ourUnitPrice - discountAmount
	costPrice := ourUnitPrice * a.config.CostPricePercent / 100
	marginPercent := 0
	if finalUnitPrice > 0 {
		marginPercent = (finalUnitPrice - costPrice) * 100 / finalUnitPrice
	}

	// Validate margin constraint
	if marginPercent < a.config.MinMarginPercent {
		return DiscountCalculation{
			ProductID:            productID,
			OurPrice:             ourUnitPrice,
			CompetitorPrice:      competitorPrice,
			CompetitorMerchantID: competitorID,
			DiscountAmount:       totalDiscountAmount,
			FinalPrice:           finalUnitPrice,
			MarginPercent:        marginPercent,
			Applied:              false,
			Reason:               fmt.Sprintf("margin %d%% below minimum %d%%", marginPercent, a.config.MinMarginPercent),
		}
	}

	return DiscountCalculation{
		ProductID:            productID,
		OurPrice:             ourUnitPrice,
		CompetitorPrice:      competitorPrice,
		CompetitorMerchantID: competitorID,
		DiscountAmount:       totalDiscountAmount,
		FinalPrice:           finalUnitPrice,
		MarginPercent:        marginPercent,
		Applied:              true,
		Reason:               "competitive discount applied",
	}
}

// calculateStrategyDiscount calculates discount per unit based on pricing strategy.
func (a *CompetitivePricingAgent) calculateStrategyDiscount(ourPrice, competitorPrice int) int {
	priceDiff := ourPrice - competitorPrice

	switch a.config.Strategy {
	case StrategyMatchPrice:
		// Match competitor exactly
		return priceDiff

	case StrategyBeatPrice:
		// Beat by percentage or minimum amount, whichever is greater
		beatByPercent := competitorPrice * a.config.BeatByPercent / 100
		beatAmount := max(beatByPercent, a.config.BeatByMinAmount)
		return priceDiff + beatAmount

	case StrategyAutoDiscount:
		// Auto mode: beat by 1% or $0.25, whichever is greater
		beatByPercent := competitorPrice * 1 / 100
		beatAmount := max(beatByPercent, 25)
		return priceDiff + beatAmount

	default:
		return priceDiff
	}
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
