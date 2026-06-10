// Package models defines interfaces for the multi-agent competitive pricing system.
package models

import "time"

// PriceIntelligencer analyzes market prices and returns intelligence.
type PriceIntelligencer interface {
	Analyze(productID string, ourPrice int) (PriceIntelligence, error)
}

// MarketAnalyzer analyzes market conditions and trends.
type MarketAnalyzer interface {
	Analyze(intel PriceIntelligence) (MarketInsight, error)
	RecordPrice(productID string, price int, timestamp time.Time) error
}

// StrategyRecommender recommends pricing strategies based on context.
type StrategyRecommender interface {
	Recommend(intel PriceIntelligence, insight MarketInsight, context BusinessConfig) (PricingRecommendation, error)
}

// MarginValidator validates pricing decisions against margin constraints.
type MarginValidator interface {
	Validate(rec PricingRecommendation, ourPrice int) (ValidationResult, error)
}

// CompetitorPriceSource provides competitor price data.
type CompetitorPriceSource interface {
	GetLowestPrice(productID string) (price int, merchantID string, err error)
	GetCompetitorPrices(productID string) ([]CompetitorPrice, error)
}

// HistoryStore stores and retrieves price history.
type HistoryStore interface {
	RecordPrice(productID string, price int, timestamp time.Time) error
	GetTrend(productID string, duration time.Duration) (Trend, error)
	GetPriceHistory(productID string, limit int) ([]PricePoint, error)
}

// PricePoint represents a single price observation.
type PricePoint struct {
	ProductID string
	Price     int
	Timestamp time.Time
}
