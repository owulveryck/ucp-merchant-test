// Package models defines the data types used by the multi-agent competitive pricing system.
package models

import "time"

// PriceIntelligence contains market price data for a product.
// This is the output of the Price Intelligence Agent.
type PriceIntelligence struct {
	ProductID   string
	OurPrice    int // in cents
	Competitors []CompetitorPrice
	LowestPrice int    // Lowest EFFECTIVE price (after discounts)
	LowestBy    string // merchant ID
	AvgPrice    int
	MaxPrice    int
	PriceSpread int // max - min
	OurRank     int // 1 = cheapest, higher = more expensive
	TotalCount  int // total number of merchants (including us)
}

// CompetitorPrice represents a single competitor's price.
type CompetitorPrice struct {
	MerchantID     string
	MerchantName   string
	Price          int       // Displayed price (before discounts)
	InStock        bool
	Timestamp      time.Time
	DiscountHints  []string  // Available discount codes
	EffectivePrice int       // Estimated price after best discount
}

// MarketInsight contains the market analysis for a product.
// This is the output of the Market Analysis Agent.
type MarketInsight struct {
	ProductID          string
	Position           string  // "leader" | "follower" | "premium" | "budget"
	Trend              string  // "stable" | "rising" | "falling"
	TrendPercent       float64 // % change over period
	Competitiveness    float64 // 0-100 score (100 = most competitive)
	Opportunity        string  // "price_war" | "premium_position" | "match_market" | "optimize"
	Reasoning          string  // human-readable explanation
	PriceVolatility    float64 // measure of price stability
	MarketConcentration string // "competitive" | "moderate" | "concentrated"
}

// PricingRecommendation contains the recommended pricing strategy.
// This is the output of the Strategy Recommender Agent.
type PricingRecommendation struct {
	ProductID      string
	Strategy       string   // "aggressive" | "balanced" | "match" | "premium" | "defensive"
	TargetPrice    int      // recommended price in cents
	DiscountAmount int      // discount needed to reach target
	Confidence     float64  // 0-100 confidence in recommendation
	Reasoning      []string // list of reasons for this strategy
	ExpectedMargin int      // expected margin % at target price
}

// ValidationResult contains the final validated pricing decision.
// This is the output of the Margin Validator Agent.
type ValidationResult struct {
	ProductID      string
	Approved       bool
	FinalPrice     int      // final approved price
	FinalDiscount  int      // final approved discount
	Margin         int      // margin % at final price
	MarginDollars  int      // margin in cents
	Warnings       []string // any warnings or adjustments made
	Rejected       bool     // true if discount was rejected
	RejectionReason string  // reason for rejection
}

// BusinessConfig contains business context for pricing decisions.
type BusinessConfig struct {
	Objective      string // "volume" | "margin" | "balanced"
	StockLevel     int    // current stock quantity
	StockThreshold int    // low stock threshold
	BrandPosition  string // "budget" | "mid" | "premium"
	MinMargin      int    // minimum margin % required
	CostPercent    int    // cost as % of base price (e.g., 60)
}

// MarginConfig contains margin validation constraints.
type MarginConfig struct {
	MinMarginPercent int  // minimum margin % (e.g., 10)
	CostPercent      int  // cost as % of price (e.g., 60) - DEPRECATED, use ActualCost instead
	ActualCost       int  // actual cost in cents (takes precedence over CostPercent)
	HardFloor        bool // never sell below cost
}

// Trend represents price trend data over time.
type Trend struct {
	Direction     string    // "up" | "down" | "stable"
	PercentChange float64   // % change
	Period        time.Duration // period analyzed
	DataPoints    int       // number of price points
	Volatility    float64   // measure of variance
}
