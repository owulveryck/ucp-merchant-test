package agents

import (
	"fmt"
	"math"
	"time"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
)

// MarketAnalysisAgent analyzes market conditions and trends.
type MarketAnalysisAgent struct {
	historyStore models.HistoryStore
}

// NewMarketAnalysisAgent creates a new market analysis agent.
func NewMarketAnalysisAgent(historyStore models.HistoryStore) *MarketAnalysisAgent {
	return &MarketAnalysisAgent{
		historyStore: historyStore,
	}
}

// RecordPrice records a price observation in history.
func (a *MarketAnalysisAgent) RecordPrice(productID string, price int, timestamp time.Time) error {
	return a.historyStore.RecordPrice(productID, price, timestamp)
}

// Analyze analyzes market conditions based on price intelligence.
func (a *MarketAnalysisAgent) Analyze(intel models.PriceIntelligence) (models.MarketInsight, error) {
	// Get trend over last 7 days
	trend, err := a.historyStore.GetTrend(intel.ProductID, 7*24*time.Hour)
	if err != nil {
		return models.MarketInsight{}, fmt.Errorf("failed to get trend: %w", err)
	}

	// Determine position
	position := a.determinePosition(intel)

	// Calculate competitiveness score (0-100)
	competitiveness := a.calculateCompetitiveness(intel)

	// Identify opportunity
	opportunity := a.identifyOpportunity(intel, trend)

	// Market concentration
	concentration := a.calculateMarketConcentration(intel)

	// Generate reasoning
	reasoning := a.generateReasoning(intel, trend, position, opportunity)

	return models.MarketInsight{
		ProductID:           intel.ProductID,
		Position:            position,
		Trend:               trend.Direction,
		TrendPercent:        trend.PercentChange,
		Competitiveness:     competitiveness,
		Opportunity:         opportunity,
		Reasoning:           reasoning,
		PriceVolatility:     trend.Volatility,
		MarketConcentration: concentration,
	}, nil
}

// determinePosition determines our market position.
func (a *MarketAnalysisAgent) determinePosition(intel models.PriceIntelligence) string {
	if intel.TotalCount == 1 {
		return "leader" // Only us
	}

	// Calculate position based on rank and price relative to average
	priceVsAvg := float64(intel.OurPrice-intel.AvgPrice) / float64(intel.AvgPrice) * 100

	switch intel.OurRank {
	case 1:
		return "leader" // Cheapest
	case 2:
		if priceVsAvg < 5 {
			return "follower" // Close to leader
		}
		return "mid-market"
	default:
		if priceVsAvg > 15 {
			return "premium" // Significantly more expensive
		} else if priceVsAvg > 5 {
			return "mid-market"
		}
		return "follower"
	}
}

// calculateCompetitiveness calculates a 0-100 competitiveness score.
func (a *MarketAnalysisAgent) calculateCompetitiveness(intel models.PriceIntelligence) float64 {
	if intel.TotalCount == 1 {
		return 100 // Only us = 100% competitive
	}

	// Base score on rank (rank 1 = 100, last = 0)
	rankScore := float64(intel.TotalCount-intel.OurRank) / float64(intel.TotalCount-1) * 100

	// Adjust based on price vs average
	if intel.AvgPrice > 0 {
		priceVsAvg := float64(intel.OurPrice) / float64(intel.AvgPrice)
		// If we're cheaper than average, boost score
		if priceVsAvg < 1.0 {
			rankScore *= 1.1 // 10% boost
		}
		// If we're more expensive than average, reduce score
		if priceVsAvg > 1.1 {
			rankScore *= 0.9 // 10% penalty
		}
	}

	return math.Min(100, math.Max(0, rankScore))
}

// identifyOpportunity identifies market opportunities.
func (a *MarketAnalysisAgent) identifyOpportunity(intel models.PriceIntelligence, trend models.Trend) string {
	// Price war detection: falling prices + high volatility
	if trend.Direction == "down" && trend.PercentChange < -5 && trend.Volatility > 10 {
		return "price_war"
	}

	// Premium position: large spread + we're expensive but stable
	if intel.PriceSpread > intel.AvgPrice*30/100 && intel.OurRank > intel.TotalCount/2 {
		return "premium_position"
	}

	// Match market: stable prices, we're near average
	priceVsAvg := math.Abs(float64(intel.OurPrice-intel.AvgPrice)) / float64(intel.AvgPrice) * 100
	if trend.Direction == "stable" && priceVsAvg < 10 {
		return "match_market"
	}

	// Rising market: prices going up
	if trend.Direction == "up" && trend.PercentChange > 3 {
		return "rising_market"
	}

	// Default: optimize
	return "optimize"
}

// calculateMarketConcentration determines if market is competitive.
func (a *MarketAnalysisAgent) calculateMarketConcentration(intel models.PriceIntelligence) string {
	if intel.TotalCount <= 2 {
		return "concentrated"
	}

	// Look at price spread relative to average
	if intel.AvgPrice > 0 {
		spreadPercent := float64(intel.PriceSpread) / float64(intel.AvgPrice) * 100
		if spreadPercent < 10 {
			return "competitive" // Tight pricing
		} else if spreadPercent < 25 {
			return "moderate"
		}
	}

	return "concentrated" // Large price differences
}

// generateReasoning creates human-readable explanation.
func (a *MarketAnalysisAgent) generateReasoning(
	intel models.PriceIntelligence,
	trend models.Trend,
	position string,
	opportunity string,
) string {
	reasoning := fmt.Sprintf("Market position: %s (rank %d of %d). ", position, intel.OurRank, intel.TotalCount)

	if trend.DataPoints > 0 {
		reasoning += fmt.Sprintf("Price trend: %s (%.1f%% over %v). ", trend.Direction, trend.PercentChange, trend.Period)
	}

	switch opportunity {
	case "price_war":
		reasoning += "Price war detected - competitors aggressively lowering prices."
	case "premium_position":
		reasoning += "Market allows premium positioning with current spread."
	case "match_market":
		reasoning += "Stable market conditions, maintain competitive position."
	case "rising_market":
		reasoning += "Market prices trending upward."
	default:
		reasoning += "Opportunity to optimize pricing strategy."
	}

	return reasoning
}
