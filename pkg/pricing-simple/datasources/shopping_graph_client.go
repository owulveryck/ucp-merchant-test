// Package datasources provides data sources for the pricing system.
package datasources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-simple/models"
)

// ShoppingGraphClient fetches competitor prices from the Shopping Graph.
type ShoppingGraphClient struct {
	baseURL      string
	ourMerchantID string
	httpClient   *http.Client
}

// NewShoppingGraphClient creates a new Shopping Graph client.
func NewShoppingGraphClient(baseURL, ourMerchantID string) *ShoppingGraphClient {
	return &ShoppingGraphClient{
		baseURL:      baseURL,
		ourMerchantID: ourMerchantID,
		httpClient:   &http.Client{Timeout: 5 * time.Second},
	}
}

// GetCompetitorPrices fetches competitor prices for a product from the Shopping Graph.
func (c *ShoppingGraphClient) GetCompetitorPrices(productID string) ([]models.CompetitorPrice, error) {
	// Search the shopping graph for this product
	searchBody := map[string]interface{}{
		"query": productID,
		"limit": 100,
	}

	bodyBytes, _ := json.Marshal(searchBody)
	resp, err := c.httpClient.Post(
		c.baseURL+"/search",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("shopping graph search failed: %w", err)
	}
	defer resp.Body.Close()

	var searchResult struct {
		Results []struct {
			MerchantID   string `json:"merchant_id"`
			MerchantName string `json:"merchant_name"`
			ProductID    string `json:"product_id"`
			ProductName  string `json:"product_name"`
			Price        int    `json:"price"`
			Rank         int    `json:"rank"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode shopping graph response: %w", err)
	}

	// Convert to CompetitorPrice format, excluding ourselves
	competitors := []models.CompetitorPrice{}
	for _, result := range searchResult.Results {
		if result.MerchantID == c.ourMerchantID {
			// Skip our own listing
			continue
		}

		competitors = append(competitors, models.CompetitorPrice{
			MerchantID:     result.MerchantID,
			MerchantName:   result.MerchantName,
			Price:          result.Price,
			EffectivePrice: result.Price, // Will be adjusted by agent if discount codes found
			DiscountHints:  []string{},
			InStock:        true, // Assume in stock if listed
		})
	}

	log.Printf("[ShoppingGraphClient] Found %d competitors for product %s", len(competitors), productID)

	return competitors, nil
}
