// Package agents contains the 3 specialized pricing agents.
package agents

import (
	"fmt"
	"sort"
	"strings"

	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple/models"
)

// MarketIntelligenceAgentImpl implements market analysis (fusion of Price Intelligence + Market Analysis).
type MarketIntelligenceAgentImpl struct {
	competitorData models.CompetitorDataSource
	merchantID     string
}

// NewMarketIntelligenceAgent creates Agent 1.
func NewMarketIntelligenceAgent(competitorData models.CompetitorDataSource, merchantID string) *MarketIntelligenceAgentImpl {
	return &MarketIntelligenceAgentImpl{
		competitorData: competitorData,
		merchantID:     merchantID,
	}
}

// Analyze performs complete market intelligence analysis.
//
// INTENTION: "Sommes-nous compétitifs sur ce produit ?"
//
// ANALYSE:
// - Prix des concurrents (avec détection codes promo)
// - Prix effectifs après réductions
// - Notre position marché
// - Tendance marché
func (a *MarketIntelligenceAgentImpl) Analyze(productID string, ourPrice int) (models.MarketIntelligenceDecision, error) {
	decision := models.MarketIntelligenceDecision{
		ProductID: productID,
		Reasoning: []string{},
	}

	// Get competitor prices
	competitors, err := a.competitorData.GetCompetitorPrices(productID)
	if err != nil {
		return decision, fmt.Errorf("failed to get competitor prices: %w", err)
	}

	// Filter out ourselves and out-of-stock
	var validCompetitors []models.CompetitorPrice
	var allDiscountCodes []string
	for _, comp := range competitors {
		if comp.MerchantID == a.merchantID {
			continue // Skip ourselves
		}
		if !comp.InStock {
			continue // Skip out of stock
		}
		validCompetitors = append(validCompetitors, comp)
		allDiscountCodes = append(allDiscountCodes, comp.DiscountHints...)
	}

	decision.DiscountCodesFound = allDiscountCodes
	decision.TotalCompetitors = len(validCompetitors) + 1 // +1 for us

	// No competitors - we set the market
	if len(validCompetitors) == 0 {
		decision.LowestCompetitor = ourPrice
		decision.OurPosition = 1
		decision.CompetitiveGap = 0
		decision.MarketTrend = "stable"
		decision.Reasoning = append(decision.Reasoning,
			"Aucun concurrent : vous définissez le prix du marché")
		return decision, nil
	}

	// Build price list (use effective prices)
	allPrices := []int{ourPrice}
	for _, comp := range validCompetitors {
		priceToUse := comp.EffectivePrice
		if priceToUse == 0 {
			priceToUse = comp.Price // Fallback
		}
		allPrices = append(allPrices, priceToUse)
	}

	sort.Ints(allPrices)

	// Find lowest and our position
	lowest := allPrices[0]
	decision.LowestCompetitor = lowest

	ourPosition := 1
	for i, price := range allPrices {
		if price == ourPrice {
			ourPosition = i + 1
			break
		}
	}
	decision.OurPosition = ourPosition
	decision.CompetitiveGap = ourPrice - lowest

	// Determine market trend (simplified: based on price spread)
	highest := allPrices[len(allPrices)-1]
	spread := highest - lowest
	avgPrice := 0
	for _, p := range allPrices {
		avgPrice += p
	}
	avgPrice = avgPrice / len(allPrices)

	if spread > avgPrice*20/100 { // Spread > 20% of average
		decision.MarketTrend = "volatile"
	} else {
		decision.MarketTrend = "stable"
	}

	// Generate reasoning
	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Marché : %d concurrents analysés", len(validCompetitors)))

	if len(allDiscountCodes) > 0 {
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Codes promo détectés : %s", strings.Join(allDiscountCodes, ", ")))
	}

	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Prix concurrent le plus bas : $%.2f", float64(lowest)/100))

	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Notre position : %d/%d", ourPosition, decision.TotalCompetitors))

	if decision.CompetitiveGap > 0 {
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Écart compétitif : -$%.2f (nous sommes plus chers)", float64(decision.CompetitiveGap)/100))
	} else if decision.CompetitiveGap < 0 {
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Écart compétitif : +$%.2f (nous sommes moins chers)", float64(-decision.CompetitiveGap)/100))
	} else {
		decision.Reasoning = append(decision.Reasoning, "Prix identique au concurrent le moins cher")
	}

	decision.Reasoning = append(decision.Reasoning,
		fmt.Sprintf("Tendance marché : %s", decision.MarketTrend))

	return decision, nil
}
