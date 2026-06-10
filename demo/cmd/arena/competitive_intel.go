package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
)

// CompetitorInfo represents a competitor's pricing information.
type CompetitorInfo struct {
	MerchantID   string  `json:"merchant_id"`
	MerchantName string  `json:"merchant_name"`
	Price        int     `json:"price"`
	PriceDisplay string  `json:"price_display"`
	InStock      bool    `json:"in_stock"`
	IsUs         bool    `json:"is_us"`
}

// PricingRecommendation contains pricing intelligence and recommendations.
type PricingRecommendation struct {
	OurPrice          int              `json:"our_price"`
	OurPriceDisplay   string           `json:"our_price_display"`
	LowestPrice       int              `json:"lowest_price"`
	LowestPriceBy     string           `json:"lowest_price_by"`
	Competitors       []CompetitorInfo `json:"competitors"`
	RecommendedPrice  int              `json:"recommended_price"`
	RecommendedPriceDisplay string     `json:"recommended_price_display"`
	PriceDifference   int              `json:"price_difference"`
	WouldWin          bool             `json:"would_win"`
	MarginPercent     int              `json:"margin_percent"`
	Message           string           `json:"message"`
}

// handleCompetitiveIntel returns competitive pricing intelligence.
func handleCompetitiveIntel(w http.ResponseWriter, r *http.Request, m *arenaMerchant, graphURL string, merchantID string, costPrice int) {
	w.Header().Set("Content-Type", "application/json")

	if graphURL == "" {
		json.NewEncoder(w).Encode(PricingRecommendation{
			Message: "Shopping Graph not configured",
		})
		return
	}

	// Get our current price
	m.config.mu.RLock()
	ourPrice := m.config.SellingPrice
	m.config.mu.RUnlock()

	// Search for our product in the Shopping Graph
	searchBody, _ := json.Marshal(map[string]interface{}{
		"query": "casque audio",
		"limit": 20,
	})
	searchResp, err := httpClient.Post(
		graphURL+"/search",
		"application/json",
		bytes.NewReader(searchBody),
	)
	if err != nil {
		log.Printf("competitive intel: shopping graph search failed: %v", err)
		json.NewEncoder(w).Encode(PricingRecommendation{
			Message: fmt.Sprintf("Shopping Graph unavailable: %v", err),
		})
		return
	}
	defer searchResp.Body.Close()

	var searchResult struct {
		Results []struct {
			MerchantID   string `json:"merchant_id"`
			MerchantName string `json:"merchant_name"`
			Price        int    `json:"price"`
			InStock      bool   `json:"in_stock"`
		} `json:"results"`
	}

	if err := json.NewDecoder(searchResp.Body).Decode(&searchResult); err != nil {
		log.Printf("competitive intel: failed to decode search results: %v", err)
		json.NewEncoder(w).Encode(PricingRecommendation{
			Message: "Failed to parse competitor data",
		})
		return
	}

	// Build competitor list
	competitors := []CompetitorInfo{}
	lowestPrice := -1
	lowestPriceBy := ""

	for _, result := range searchResult.Results {
		if !result.InStock {
			continue // Skip out-of-stock
		}

		isUs := result.MerchantID == merchantID
		competitors = append(competitors, CompetitorInfo{
			MerchantID:   result.MerchantID,
			MerchantName: result.MerchantName,
			Price:        result.Price,
			PriceDisplay: formatPrice(result.Price),
			InStock:      result.InStock,
			IsUs:         isUs,
		})

		// Track lowest competitor price (excluding us)
		if !isUs && (lowestPrice == -1 || result.Price < lowestPrice) {
			lowestPrice = result.Price
			lowestPriceBy = result.MerchantName
		}
	}

	// Sort competitors by price
	sort.Slice(competitors, func(i, j int) bool {
		return competitors[i].Price < competitors[j].Price
	})

	// Calculate recommendation
	recommendation := PricingRecommendation{
		OurPrice:        ourPrice,
		OurPriceDisplay: formatPrice(ourPrice),
		Competitors:     competitors,
	}

	if lowestPrice == -1 {
		// No competitors found
		recommendation.Message = "✅ You're the only merchant! No price adjustment needed."
		recommendation.WouldWin = true
		recommendation.RecommendedPrice = ourPrice
		recommendation.RecommendedPriceDisplay = formatPrice(ourPrice)
	} else {
		recommendation.LowestPrice = lowestPrice
		recommendation.LowestPriceBy = lowestPriceBy

		// Calculate recommended price: beat lowest by 5% or $0.50, whichever is greater
		beatByPercent := lowestPrice * 5 / 100
		beatByAmount := max(beatByPercent, 50) // At least $0.50
		recommendedPrice := lowestPrice - beatByAmount

		// Ensure we maintain minimum margin
		minPrice := costPrice * 110 / 100 // Cost + 10% margin
		if recommendedPrice < minPrice {
			recommendedPrice = minPrice
			recommendation.Message = fmt.Sprintf("⚠️ Cannot beat %s ($%.2f) while maintaining 10%% margin. Lowest viable price: $%.2f",
				lowestPriceBy,
				float64(lowestPrice)/100,
				float64(recommendedPrice)/100,
			)
		} else {
			priceDiff := ourPrice - recommendedPrice
			recommendation.Message = fmt.Sprintf("💡 Lower to $%.2f to beat %s and win sales!",
				float64(recommendedPrice)/100,
				lowestPriceBy,
			)
			recommendation.PriceDifference = priceDiff
		}

		recommendation.RecommendedPrice = recommendedPrice
		recommendation.RecommendedPriceDisplay = formatPrice(recommendedPrice)
		recommendation.MarginPercent = (recommendedPrice - costPrice) * 100 / recommendedPrice
		recommendation.WouldWin = recommendedPrice < lowestPrice

		// Check if we're already winning
		if ourPrice <= lowestPrice {
			recommendation.Message = "🏆 You have the best price! Clients will choose you."
			recommendation.WouldWin = true
		}
	}

	json.NewEncoder(w).Encode(recommendation)
}

// formatPrice formats cents to dollar string.
func formatPrice(cents int) string {
	return fmt.Sprintf("$%.2f", float64(cents)/100)
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
