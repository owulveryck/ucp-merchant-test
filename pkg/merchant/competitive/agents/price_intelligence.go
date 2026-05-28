// Package agents contains the specialized agents for competitive pricing.
package agents

import (
	"fmt"
	"log"
	"sort"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
)

// PriceIntelligenceAgent gathers and analyzes competitor pricing data.
type PriceIntelligenceAgent struct {
	competitorAPI models.CompetitorPriceSource
	merchantID    string
}

// NewPriceIntelligenceAgent creates a new price intelligence agent.
func NewPriceIntelligenceAgent(competitorAPI models.CompetitorPriceSource, merchantID string) *PriceIntelligenceAgent {
	return &PriceIntelligenceAgent{
		competitorAPI: competitorAPI,
		merchantID:    merchantID,
	}
}

// Analyze gathers competitor prices and calculates market statistics.
func (a *PriceIntelligenceAgent) Analyze(productID string, ourPrice int) (models.PriceIntelligence, error) {
	// Get competitor prices
	competitors, err := a.competitorAPI.GetCompetitorPrices(productID)
	if err != nil {
		return models.PriceIntelligence{}, fmt.Errorf("failed to get competitor prices: %w", err)
	}

	log.Printf("[DEBUG PriceIntel] Got %d competitors from API for product %s", len(competitors), productID)
	log.Printf("[DEBUG PriceIntel] Our merchantID: %s", a.merchantID)
	for i, comp := range competitors {
		log.Printf("[DEBUG PriceIntel] Competitor %d: ID=%s, Name=%s, Price=%d, InStock=%v",
			i, comp.MerchantID, comp.MerchantName, comp.Price, comp.InStock)
	}

	// Filter out ourselves and out-of-stock
	var validCompetitors []models.CompetitorPrice
	for _, comp := range competitors {
		if comp.MerchantID == a.merchantID {
			log.Printf("[DEBUG PriceIntel] FILTERED OUT (ourselves): %s", comp.MerchantID)
			continue // Skip ourselves
		}
		if !comp.InStock {
			log.Printf("[DEBUG PriceIntel] FILTERED OUT (out of stock): %s", comp.MerchantID)
			continue // Skip out of stock
		}
		validCompetitors = append(validCompetitors, comp)
	}

	log.Printf("[DEBUG PriceIntel] After filtering: %d valid competitors", len(validCompetitors))

	// If no competitors, return basic intelligence
	if len(validCompetitors) == 0 {
		return models.PriceIntelligence{
			ProductID:   productID,
			OurPrice:    ourPrice,
			Competitors: []models.CompetitorPrice{},
			LowestPrice: ourPrice,
			LowestBy:    a.merchantID,
			AvgPrice:    ourPrice,
			MaxPrice:    ourPrice,
			PriceSpread: 0,
			OurRank:     1,
			TotalCount:  1,
		}, nil
	}

	// Calculate statistics
	allPrices := []int{ourPrice}
	for _, comp := range validCompetitors {
		allPrices = append(allPrices, comp.Price)
	}

	sort.Ints(allPrices)

	lowest := allPrices[0]
	highest := allPrices[len(allPrices)-1]
	spread := highest - lowest

	// Calculate average
	sum := 0
	for _, price := range allPrices {
		sum += price
	}
	avg := sum / len(allPrices)

	// Find who has the lowest price
	lowestBy := a.merchantID
	lowestPrice := ourPrice
	for _, comp := range validCompetitors {
		if comp.Price < lowestPrice {
			lowestPrice = comp.Price
			lowestBy = comp.MerchantID
		}
	}

	// Calculate our rank (1 = cheapest)
	ourRank := 1
	for _, price := range allPrices {
		if price < ourPrice {
			ourRank++
		}
	}

	return models.PriceIntelligence{
		ProductID:   productID,
		OurPrice:    ourPrice,
		Competitors: validCompetitors,
		LowestPrice: lowestPrice,
		LowestBy:    lowestBy,
		AvgPrice:    avg,
		MaxPrice:    highest,
		PriceSpread: spread,
		OurRank:     ourRank,
		TotalCount:  len(allPrices),
	}, nil
}
