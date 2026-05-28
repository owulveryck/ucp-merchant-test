package competitive

import (
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/discount"
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
)

// PricingStrategy defines how the agent competes with other merchants.
type PricingStrategy string

const (
	// StrategyMatchPrice matches the lowest competitor price exactly.
	StrategyMatchPrice PricingStrategy = "match"

	// StrategyBeatPrice beats the competitor by a percentage (default 5%).
	StrategyBeatPrice PricingStrategy = "beat"

	// StrategyAutoDiscount automatically generates the optimal discount.
	StrategyAutoDiscount PricingStrategy = "auto"
)

// Special discount codes recognized by the competitive pricing agent.
const (
	// CodeAutoCompete triggers automatic competitive pricing calculation.
	CodeAutoCompete = "AUTO_COMPETE"

	// CodePrefixComp is a reserved prefix for competitor-specific codes.
	CodePrefixComp = "COMP_"
)

// CompetitorPriceSource provides access to competitor pricing data.
// Implementations typically query a Shopping Graph or price aggregator.
type CompetitorPriceSource interface {
	// GetLowestPrice returns the lowest competitor price for a product.
	// Returns the price in minor currency units (cents), the merchant ID
	// offering that price, and an error if the lookup fails.
	//
	// If no competitors sell this product, returns an error.
	// The returned merchantID should be excluded when making pricing decisions
	// (it's the merchant offering the lowest price, so comparing against self).
	GetLowestPrice(productID string) (price int, merchantID string, err error)

	// GetCompetitorPrices returns all competitor prices for a product.
	// Useful for advanced strategies that consider price distribution.
	GetCompetitorPrices(productID string) ([]CompetitorPrice, error)
}

// CompetitorPrice represents a single competitor's price for a product.
type CompetitorPrice struct {
	MerchantID   string // Unique merchant identifier
	MerchantName string // Human-readable merchant name
	ProductID    string // Product identifier at this merchant
	Price        int    // Price in minor currency units (cents)
	InStock      bool   // Whether the product is available
}

// SearchResult represents a product search result from the Shopping Graph.
type SearchResult struct {
	Rank          int      `json:"rank"`           // Search ranking position
	ProductID     string   `json:"product_id"`     // Product identifier
	Title         string   `json:"title"`          // Product title/name
	MerchantID    string   `json:"merchant_id"`    // Merchant offering this product
	MerchantName  string   `json:"merchant_name"`  // Merchant display name
	MerchantURL   string   `json:"merchant_url"`   // Merchant endpoint URL
	Price         int      `json:"price"`          // Price in minor currency units
	PriceDisplay  string   `json:"price_display"`  // Formatted price string
	InStock       bool     `json:"in_stock"`       // Availability status
	DiscountHints []string `json:"discount_hints"` // Available discount codes
	Sponsored     bool     `json:"sponsored"`      // True if result was promoted via CPC bidding
	ActualCPC     int      `json:"actual_cpc"`     // Actual cost-per-click paid
	QualityScore  float64  `json:"quality_score"`  // Ad quality score
}

// DiscountCalculation represents the result of a competitive pricing calculation.
type DiscountCalculation struct {
	// ProductID is the product being priced.
	ProductID string

	// OurPrice is the merchant's base price (before discount).
	OurPrice int

	// CompetitorPrice is the lowest competitor price found.
	CompetitorPrice int

	// CompetitorMerchantID identifies which competitor has the lowest price.
	CompetitorMerchantID string

	// DiscountAmount is the calculated discount in minor currency units.
	// This is the amount that will be subtracted from the line item total.
	DiscountAmount int

	// FinalPrice is the price after discount (OurPrice - DiscountAmount).
	FinalPrice int

	// MarginPercent is the profit margin after discount.
	// Calculated as: (FinalPrice - CostPrice) / FinalPrice * 100
	MarginPercent int

	// Applied indicates whether the discount was actually applied.
	// False if margin constraints prevented the discount.
	Applied bool

	// Reason explains why the discount was or wasn't applied.
	Reason string
}

// Config holds configuration for the competitive pricing agent.
type Config struct {
	// Strategy determines how we compete (match, beat, auto).
	Strategy PricingStrategy

	// MinMarginPercent is the minimum acceptable profit margin (0-100).
	// Discounts that would reduce margin below this threshold are rejected.
	MinMarginPercent int

	// BeatByPercent is the percentage to beat competitor prices (for StrategyBeatPrice).
	// Default: 5 means beat by 5% or $0.50, whichever is greater.
	BeatByPercent int

	// BeatByMinAmount is the minimum amount in cents to beat competitor (for StrategyBeatPrice).
	// Default: 50 cents.
	BeatByMinAmount int

	// CostPricePercent is the percentage of retail price that is cost (for margin calculation).
	// Default: 60 means cost is 60% of retail, so 40% margin.
	// Used when actual cost data is unavailable.
	CostPricePercent int

	// Timeout is the maximum duration for Shopping Graph queries.
	// Default: 500ms.
	TimeoutMs int

	// EnableCache enables caching of competitor price lookups.
	// Default: true with 10s TTL.
	EnableCache bool

	// CacheTTLSeconds is the cache time-to-live in seconds.
	// Default: 10.
	CacheTTLSeconds int
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Strategy:         StrategyBeatPrice,
		MinMarginPercent: 10,
		BeatByPercent:    5,
		BeatByMinAmount:  50, // $0.50
		CostPricePercent: 60, // 40% margin
		TimeoutMs:        500,
		EnableCache:      true,
		CacheTTLSeconds:  10,
	}
}

// CompetitivePricingAgent calculates dynamic discounts based on competitor prices.
type CompetitivePricingAgent struct {
	baseData      discount.DiscountLookup // Static discount codes (fallback)
	competitorAPI CompetitorPriceSource   // Source of competitor pricing data
	config        Config                  // Agent configuration
	merchantID    string                  // This merchant's ID (to exclude from comparisons)
}

// Ensure CompetitivePricingAgent implements discount.DiscountLookup.
var _ discount.DiscountLookup = (*CompetitivePricingAgent)(nil)

// NewCompetitivePricingAgent creates a new competitive pricing agent.
//
// Parameters:
//   - baseData: Fallback for static discount codes (e.g., "10OFF", "WELCOME20")
//   - competitorAPI: Source of competitor prices (typically Shopping Graph client)
//   - merchantID: This merchant's unique ID (excluded from competitor comparisons)
//   - config: Agent configuration (use DefaultConfig() for defaults)
func NewCompetitivePricingAgent(
	baseData discount.DiscountLookup,
	competitorAPI CompetitorPriceSource,
	merchantID string,
	config Config,
) *CompetitivePricingAgent {
	return &CompetitivePricingAgent{
		baseData:      baseData,
		competitorAPI: competitorAPI,
		config:        config,
		merchantID:    merchantID,
	}
}

// ApplyDiscountsWithContext is an extended version of ApplyDiscounts that has access
// to line items for calculating competitive discounts.
//
// This method should be called instead of the standard discount.ApplyDiscounts when
// competitive pricing is enabled.
func (a *CompetitivePricingAgent) ApplyDiscountsWithContext(
	codes []string,
	lineItems []model.LineItem,
) *model.Discounts {
	result := &model.Discounts{
		Codes:   codes,
		Applied: []model.AppliedDiscount{},
	}

	if len(codes) == 0 {
		return nil // No codes submitted
	}

	// Apply static codes first
	subtotal := calculateSubtotal(lineItems)
	remainingSubtotal := subtotal

	for _, code := range codes {
		// Check if it's a static discount
		if disc := a.baseData.FindDiscountByCode(code); disc != nil {
			var amount int
			if disc.Type == "percentage" {
				amount = remainingSubtotal * disc.Value / 100
			} else {
				amount = disc.Value
			}

			remainingSubtotal -= amount

			result.Applied = append(result.Applied, model.AppliedDiscount{
				Code:   disc.Code,
				Title:  disc.Description,
				Amount: amount,
			})
		}
	}

	// Check for AUTO_COMPETE code
	for _, code := range codes {
		if code == CodeAutoCompete {
			competitiveDiscount := a.calculateCompetitiveDiscount(lineItems)
			if competitiveDiscount > 0 {
				result.Applied = append(result.Applied, model.AppliedDiscount{
					Code:   CodeAutoCompete,
					Title:  "Competitive Price Match",
					Amount: competitiveDiscount,
				})
			}
		}
	}

	return result
}

// calculateSubtotal sums the subtotals from all line items.
func calculateSubtotal(lineItems []model.LineItem) int {
	subtotal := 0
	for _, item := range lineItems {
		if st := findTotal(item.Totals, "subtotal"); st != nil {
			subtotal += st.Amount
		}
	}
	return subtotal
}

// findTotal searches for a total with the given type in a slice of totals.
// Returns nil if not found.
func findTotal(totals []model.Total, totalType string) *model.Total {
	for i := range totals {
		if totals[i].Type == totalType {
			return &totals[i]
		}
	}
	return nil
}
