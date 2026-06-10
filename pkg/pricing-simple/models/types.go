// Package models contains data types for the simplified 3-agent pricing system.
package models

// MarketIntelligenceDecision represents Agent 1's market analysis.
type MarketIntelligenceDecision struct {
	ProductID          string   // Product being analyzed
	LowestCompetitor   int      // Lowest competitor price (effective, after discounts)
	OurPosition        int      // Our rank (1=cheapest, 2=second...)
	TotalCompetitors   int      // Total number of competitors
	MarketTrend        string   // "rising", "stable", "falling"
	CompetitiveGap     int      // Gap with cheapest (negative = we're more expensive)
	DiscountCodesFound []string // Competitor discount codes detected
	Reasoning          []string // Explanation of findings
}

// LoyaltyDecision represents Agent 2's customer value analysis.
type LoyaltyDecision struct {
	CustomerID        string   // Customer being analyzed
	IsVIP             bool     // VIP status
	CustomerTier      string   // "premium", "gold", "silver", "standard"
	SuggestedDiscount int      // Suggested discount percentage (0-15%)
	LifetimeValue     int      // Total spent (cents)
	PurchaseCount     int      // Number of purchases
	Reasoning         []string // Explanation of decision
}

// FinalPricingDecision represents Agent 3's orchestrated decision.
type FinalPricingDecision struct {
	ProductID      string   // Product ID
	CustomerID     string   // Customer ID
	OriginalPrice  int      // Original product price (cents)
	FinalPrice     int      // Final recommended price (cents)
	TotalDiscount  int      // Total discount amount (cents)
	DiscountPercent int     // Discount percentage
	Margin         int      // Final margin percentage
	IsVIP          bool     // Was VIP pricing applied?
	IsCompetitive  bool     // Is price competitive?
	Strategy       string   // "vip_priority", "market_match", "minimum_viable"
	Approved       bool     // Approved for application?
	Warnings       []string // Warnings if any
	Reasoning      []string // Complete reasoning chain
}

// PricingRequest represents a request for price calculation.
type PricingRequest struct {
	ProductID  string // Product to price
	CustomerID string // Customer requesting
	BasePrice  int    // Current price (cents)
	CostPrice  int    // Cost price (cents)
}

// CustomerProfile represents customer data.
type CustomerProfile struct {
	CustomerID      string
	TotalSpent      int // Total lifetime value (cents)
	PurchaseCount   int
	LastPurchaseDays int
}

// CompetitorPrice represents a competitor's pricing.
type CompetitorPrice struct {
	MerchantID      string
	MerchantName    string
	Price           int      // Displayed price
	EffectivePrice  int      // Price after discount codes
	DiscountHints   []string // Discount codes
	InStock         bool
}

// BusinessConfig represents business constraints.
type BusinessConfig struct {
	MinMarginPercent int    // Minimum acceptable margin (e.g., 10)
	CostPercent      int    // Cost as percentage of price (e.g., 80 = 80%)
	HardFloor        bool   // Never sell below cost?
	MerchantID       string // Our merchant ID
}
