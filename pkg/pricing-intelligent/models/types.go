// Package models contains data types for intelligent pricing system.
package models

// CustomerProfile represents customer information for loyalty analysis.
type CustomerProfile struct {
	CustomerID      string  // Unique customer identifier
	TotalSpent      int     // Total amount spent (cents)
	PurchaseCount   int     // Number of purchases
	AveragePurchase int     // Average purchase amount (cents)
	LastPurchaseDays int    // Days since last purchase
	IsVIP           bool    // VIP status
}

// LoyaltyDecision represents the loyalty agent's decision.
type LoyaltyDecision struct {
	CustomerID       string   // Customer identifier
	IsVIP            bool     // Is this a VIP customer?
	SuggestedDiscount int     // Suggested discount percentage (0-100)
	Reasoning        []string // Explanation of decision
	Confidence       int      // Confidence score (0-100)
}

// CompetitivenessDecision represents the competitiveness agent's decision.
type CompetitivenessDecision struct {
	ProductID         string   // Product identifier
	LowestCompetitor  int      // Lowest competitor price (cents)
	OurPosition       int      // Our rank (1 = cheapest)
	TotalCompetitors  int      // Total number of competitors
	SuggestedPrice    int      // Suggested price to win (cents)
	Reasoning         []string // Explanation of decision
	Confidence        int      // Confidence score (0-100)
}

// PricingRequest represents a request for optimal price calculation.
type PricingRequest struct {
	CustomerID string // Customer identifier
	ProductID  string // Product identifier
	BasePrice  int    // Current base price (cents)
	CostPrice  int    // Cost price (cents)
}

// PricingResult represents the orchestrator's final decision.
type PricingResult struct {
	ProductID         string   // Product identifier
	CustomerID        string   // Customer identifier
	BasePrice         int      // Original base price (cents)
	FinalPrice        int      // Final optimized price (cents)
	Discount          int      // Applied discount (cents)
	DiscountPercent   int      // Discount percentage
	Margin            int      // Final margin percentage
	IsVIP             bool     // Was VIP pricing applied?
	IsCompetitive     bool     // Is price competitive?
	Reasoning         []string // Complete reasoning chain
	LoyaltyDecision   LoyaltyDecision
	CompetitivenessDecision CompetitivenessDecision
}
